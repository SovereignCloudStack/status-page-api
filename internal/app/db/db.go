package db

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/SovereignCloudStack/status-page-api/internal/app/logging"
	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Database wraps the database connection.
type Database struct {
	conn   *gorm.DB
	logger *zerolog.Logger
}

// New creates a new wrapper for the database and initialize it.
func New(connection string, logger *zerolog.Logger) (*Database, error) {
	conn, err := gorm.Open(postgres.Open(connection), &gorm.Config{ //nolint:exhaustruct
		Logger: logging.NewGormLogger(logger),
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting database: %w", err)
	}

	err = conn.AutoMigrate(
		&DbDef.Component{},      //nolint:exhaustruct
		&DbDef.Phase{},          //nolint:exhaustruct
		&DbDef.IncidentUpdate{}, //nolint:exhaustruct
		&DbDef.Incident{},       //nolint:exhaustruct
		&DbDef.ImpactType{},     //nolint:exhaustruct
		&DbDef.Impact{},         //nolint:exhaustruct
		&DbDef.Severity{},       //nolint:exhaustruct
	)
	if err != nil {
		return nil, fmt.Errorf("error migrating database structure: %w", err)
	}

	return &Database{
		conn:   conn,
		logger: logger,
	}, nil
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

func provisionPhases(phases []DbDef.Phase, dbTx *gorm.DB, logger *zerolog.Logger) error {
	initialPhaseGeneration := 1

	// set initial phase and orders.
	for phaseIndex := range phases {
		order := phaseIndex

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
func (db *Database) Provision(filename string) error {
	provisioningLogger := db.logger.With().Str("method", "Provisioning").Logger()

	type ProvisionedResources struct {
		Components  []DbDef.Component  `yaml:"components"`
		ImpactTypes []DbDef.ImpactType `yaml:"impactTypes"`
		Phases      []DbDef.Phase      `yaml:"phases"`
		Severities  []DbDef.Severity   `yaml:"severities"`
	}

	provisioningLogger.Debug().Str("provisioningFile", filename).Msg("opening provisioning file")

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

	provisioningLogger.Info().Msg("read resources from provisioning file")
	provisioningLogger.Debug().Interface("resources", resources).Send()

	err = db.conn.Transaction(func(dbTx *gorm.DB) error {
		var txErr error

		txErr = provision(resources.Components, dbTx, &provisioningLogger)
		if txErr != nil {
			return fmt.Errorf("error provisioning components: %w", err)
		}

		txErr = provision(resources.ImpactTypes, dbTx, &provisioningLogger)
		if txErr != nil {
			return fmt.Errorf("error provisioning impact types: %w", err)
		}

		txErr = provisionPhases(resources.Phases, dbTx, &provisioningLogger)
		if txErr != nil {
			return fmt.Errorf("error provisioning phases: %w", err)
		}

		txErr = provision(resources.Severities, dbTx, &provisioningLogger)
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

// GetDBCon gets the wrapped database connection.
func (db *Database) GetDBCon() *gorm.DB {
	return db.conn
}
