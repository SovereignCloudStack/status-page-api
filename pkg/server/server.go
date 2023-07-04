package server

import (
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
