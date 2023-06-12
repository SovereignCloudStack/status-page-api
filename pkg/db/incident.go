package db

import (
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
)

// Incident represents an incident happening to one or more [Component].
type Incident struct {
	Affects         []*Impact `gorm:"foreignKey:IncidentID"`
	BeganAt         *api.Date
	Description     *api.Description
	DisplayName     *api.DisplayName
	EndedAt         *api.Date
	PhaseGeneration *api.Incremental
	PhaseOrder      *api.Incremental
	Phase           *Phase            `gorm:"foreignKey:PhaseGeneration,PhaseOrder;References:Generation,Order"`
	Updates         []*IncidentUpdate `gorm:"foreignKey:IncidentID"`
	Model           `gorm:"embedded"`
}

// IncidentUpdate describes a action that changes the incident.
type IncidentUpdate struct {
	api.IncidentUpdate `gorm:"embedded"`
	Order              api.Incremental `gorm:"primaryKey"`
	IncidentID         ID              `gorm:"primaryKey"`
}
