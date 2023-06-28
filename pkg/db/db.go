package db

import (
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// ID is the generally used identifier type. String type for UUIDs.
type ID = uuid.UUID

// Model sets the basic Data for all true database resources.
type Model struct {
	ID *ID `gorm:"primaryKey;type:uuid;"`
}

// BeforeCreate is a gorm hook to fill the ID field with a new UUID,
// before an insert statement is send to the database.
func (m *Model) BeforeCreate(_ *gorm.DB) error {
	if m.ID == nil {
		// pointer to id is nil
		id := uuid.New()
		m.ID = &id
	} else if *m.ID == uuid.Nil {
		// points to id but is empty id
		id := uuid.New()
		m.ID = &id
	}

	return nil
}

// Provision initializes the database with the contents of the provision file.
func Provision(filename string, dbCon *gorm.DB) error { //nolint:funlen,cyclop
	type ProvisionedResources struct {
		Components  []Component  `yaml:"components"`
		ImpactTypes []ImpactType `yaml:"impactTypes"`
		Phases      []Phase      `yaml:"phases"`
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening provisioning file `%s`: %w", filename, err)
	}
	defer file.Close()

	resources := ProvisionedResources{} //nolint:exhaustruct

	err = yaml.NewDecoder(file).Decode(&resources)
	if err != nil {
		return fmt.Errorf("error decoding provisioning file `%s`: %w", filename, err)
	}

	initialPhaseGeneration := 1

	err = dbCon.Transaction(func(dbTx *gorm.DB) error {
		// check if already provisioned
		var lastPhase Phase

		res := dbTx.
			Where(
				"generation = ? AND name = ?",
				initialPhaseGeneration, resources.Phases[len(resources.Phases)-1].Name,
			).
			First(&lastPhase)
		if res.Error != nil {
			if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return fmt.Errorf("error getting last phase: %w", err)
			}
		}

		if lastPhase.Order != nil && *lastPhase.Order == len(resources.Phases)-1 {
			// db has been provisioned before
			return nil
		}

		res = dbTx.Create(&resources.Components)
		if res.Error != nil {
			return fmt.Errorf("error saving components: %w", res.Error)
		}

		res = dbTx.Create(&resources.ImpactTypes)
		if res.Error != nil {
			return fmt.Errorf("error saving impact types: %w", res.Error)
		}

		for phaseIndex := range resources.Phases {
			order := phaseIndex

			resources.Phases[phaseIndex].Order = &order
			resources.Phases[phaseIndex].Generation = &initialPhaseGeneration
		}

		res = dbTx.Create(&resources.Phases)
		if res.Error != nil {
			return fmt.Errorf("error saving phases: %w", res.Error)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error in database transaction: %w", err)
	}

	return nil
}
