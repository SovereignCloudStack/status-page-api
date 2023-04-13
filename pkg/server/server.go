package server

import "gorm.io/gorm"

type Implementation struct {
	dbCon *gorm.DB
}

func New(dbCon *gorm.DB) *Implementation {
	return &Implementation{
		dbCon: dbCon,
	}
}
