package server

import (
	"fmt"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// Implementation holds all functions definded by the [api.ServerInterface] and other needed components.
type Implementation struct {
	dbCon  *gorm.DB
	logger *zerolog.Logger
}

// New creates a new [Implementation] Object with the setted dbCon.
func New(dbCon *gorm.DB, logger *zerolog.Logger) *Implementation {
	return &Implementation{
		dbCon:  dbCon,
		logger: logger,
	}
}

func GetCurrentPhaseGeneration(dbCon *gorm.DB) (int, error) {
	type Result struct {
		Generation int
	}

	var result Result

	res := dbCon.Table("phases").Select("MAX(generation) as Generation").Scan(&result)
	if res.Error != nil {
		return 0, fmt.Errorf("error getting current generation: %w", res.Error)
	}

	return result.Generation, nil
}
