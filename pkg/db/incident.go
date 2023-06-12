package db

import (
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
)

// Incident represents an incident happening to one or more [Component].
type Incident struct {
	Model           `gorm:"embedded"`
	DisplayName     *api.DisplayName
	Description     *api.Description
	Affects         []*Impact `gorm:"foreignKey:IncidentID"`
	BeganAt         *api.Date
	EndedAt         *api.Date
	PhaseGeneration *api.Incremental
	PhaseOrder      *api.Incremental
	Phase           *Phase            `gorm:"foreignKey:PhaseGeneration,PhaseOrder;References:Generation,Order"`
	Updates         []*IncidentUpdate `gorm:"foreignKey:IncidentID"`
}

// IncidentUpdate describes a action that changes the incident.
type IncidentUpdate struct {
	IncidentID  *ID              `gorm:"primaryKey"`
	Order       *api.Incremental `gorm:"primaryKey"`
	DisplayName *api.DisplayName
	Description *api.Description
	CreatedAt   *api.Date
}
