package db

import "github.com/SovereignCloudStack/status-page-openapi/pkg/api"

// ImpactType represents the type of impact.
type ImpactType struct {
	Model          `gorm:"embedded"`
	api.ImpactType `gorm:"embedded"`
}

type Impact struct {
	IncidentID   ID         `gorm:"primaryKey"`
	ComponentID  ID         `gorm:"primaryKey"`
	ImpactTypeID ID         `gorm:"primaryKey"`
	impactType   ImpactType `gorm:"foreignKey:ImpactTypeID"`
}
