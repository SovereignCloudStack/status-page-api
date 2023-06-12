package db

import "github.com/SovereignCloudStack/status-page-openapi/pkg/api"

// ImpactType represents the type of impact.
type ImpactType struct {
	api.ImpactType `gorm:"embedded"`
	Model          `gorm:"embedded"`
}

type Impact struct {
	Incident   *Incident   `gorm:"foreignKey:IncidentID"`
	Component  *Component  `gorm:"foreignKey:ComponentID"`
	ImpactType *ImpactType `gorm:"foreignKey:ImpactTypeID"`

	IncidentID   ID `gorm:"primaryKey"`
	ComponentID  ID `gorm:"primaryKey"`
	ImpactTypeID ID `gorm:"primaryKey"`
}
