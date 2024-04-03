package db

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
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

func provision[S ~[]E, E any](data S, dbTx *gorm.DB, logger *zerolog.Logger) error {
	var (
		limit  = 5
		target = make(S, limit)
	)

	reflectType := reflect.TypeOf(data).Elem()
	provisioningLogger := logger.With().Str("function", "provision").Str("type", reflectType.Name()).Logger()

	// get from database if exist.
	res := dbTx.
		Limit(limit). // do not load all.
		Find(&target)
	if res.Error != nil {
		if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("error getting %s: %w", reflectType.Name(), res.Error)
		}
	}

	provisioningLogger.Debug().Interface("found", target).Send()

	// check if already provisioned.
	if len(target) != 0 {
		provisioningLogger.Info().Msg("already provisioned")

		return nil
	}

	provisioningLogger.Info().Msg("provisioning")
	provisioningLogger.Debug().Interface("data", data).Send()

	// create new data in db.
	res = dbTx.Create(data)
	if res.Error != nil {
		return fmt.Errorf("error creating data: %w", res.Error)
	}

	return nil
}

func provisionPhases(phases []Phase, dbTx *gorm.DB, logger *zerolog.Logger) error {
	initialPhaseGeneration := 1

	// set initial phase and orders.
	for phaseIndex := range phases {
		order := phaseIndex //nolint:copyloopvar

		phases[phaseIndex].Order = &order
		phases[phaseIndex].Generation = &initialPhaseGeneration
	}

	err := provision(phases, dbTx, logger)
	if err != nil {
		return err
	}

	return nil
}

// Provision initializes the database with the contents of the provision file.
func Provision(filename string, dbCon *gorm.DB, logger *zerolog.Logger) error {
	type ProvisionedResources struct {
		Components  []Component  `yaml:"components"`
		ImpactTypes []ImpactType `yaml:"impactTypes"`
		Phases      []Phase      `yaml:"phases"`
		Severities  []Severity   `yaml:"severities"`
	}

	logger.Debug().Str("provisioningFile", filename).Msg("opening provisioning file")

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening provisioning file `%s`: %w", filename, err)
	}
	defer file.Close()

	resources := ProvisionedResources{} //nolint:exhaustruct

	err = yaml.NewDecoder(file).Decode(&resources) //nolint:musttag // musstag has a false positive for ignored fields.
	if err != nil {
		return fmt.Errorf("error decoding provisioning file `%s`: %w", filename, err)
	}

	logger.Info().Msg("read resources from provisioning file")
	logger.Debug().Interface("resources", resources).Send()

	err = dbCon.Transaction(func(dbTx *gorm.DB) error {
		var txErr error

		txErr = provision(resources.Components, dbTx, logger)
		if txErr != nil {
			return fmt.Errorf("error provisioning components: %w", err)
		}

		txErr = provision(resources.ImpactTypes, dbTx, logger)
		if txErr != nil {
			return fmt.Errorf("error provisioning impact types: %w", err)
		}

		txErr = provisionPhases(resources.Phases, dbTx, logger)
		if txErr != nil {
			return fmt.Errorf("error provisioning phases: %w", err)
		}

		txErr = provision(resources.Severities, dbTx, logger)
		if txErr != nil {
			return fmt.Errorf("error provisioning severities: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error in provisioning transaction: %w", err)
	}

	return nil
}
