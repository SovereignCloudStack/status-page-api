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

type Model struct {
	ID *ID `gorm:"primaryKey;type:uuid;"`
}

func (m *Model) BeforeCreate(_ *gorm.DB) error {
	id := uuid.New()
	m.ID = &id

	return nil
}

// Provision initializes the database with the contents of the provision file.
func Provision(filename string, dbCon *gorm.DB) error { //nolint:funlen,cyclop
	type ProvisionedResources struct {
		Components  []*Component  `yaml:"components"`
		ImpactTypes []*ImpactType `yaml:"impactTypes"`
		Phases      []*Phase      `yaml:"phases"`
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

	// check if already provisioned
	var lastPhase Phase

	res := dbCon.
		Where("generation = ? AND name = ?", initialPhaseGeneration, resources.Phases[len(resources.Phases)-1].Name).
		First(&lastPhase)

	err = res.Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("error getting last phase: %w", err)
		}
	}

	if lastPhase.Order != nil && *lastPhase.Order == len(resources.Phases)-1 {
		// db has been provisioned before
		return nil
	}

	for _, component := range resources.Components {
		err = dbCon.Save(component).Error
		if err != nil {
			return fmt.Errorf("error saving component `%s`: %w", *component.DisplayName, err)
		}
	}

	for _, impactType := range resources.ImpactTypes {
		err = dbCon.Save(impactType).Error
		if err != nil {
			return fmt.Errorf("error saving impact type `%s`: %w", *impactType.DisplayName, err)
		}
	}

	var phaseOrder int

	for phaseIndex := range resources.Phases {
		phase := resources.Phases[phaseIndex]
		phase.Order = &phaseOrder
		phase.Generation = &initialPhaseGeneration

		err = dbCon.Save(&phase).Error
		if err != nil {
			return fmt.Errorf("error saving phase `%s`: %w", *phase.Name, err)
		}

		phaseOrder++
	}

	return nil
}
