package server

import "gorm.io/gorm"

// Implementation holds all functions definded by the [api.ServerInterface] and other needed components.
type Implementation struct {
	dbCon *gorm.DB
}

// New creates a new [Implementation] Object with the setted dbCon.
func New(dbCon *gorm.DB) *Implementation {
	return &Implementation{
		dbCon: dbCon,
	}
}
