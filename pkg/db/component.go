package db

import (
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
)

// Component represents a single component that could be affected by many [Incident].
type Component struct {
	Model              `gorm:"embedded"`
	DisplayName        *api.DisplayName `yaml:"displayname"`
	Labels             *Labels          `gorm:"type:jsonb"             yaml:"labels"`
	ActivelyAffectedBy *[]Impact        `gorm:"foreignKey:ComponentID"`
}

func (c *Component) GetImpactIncidentList() *api.ImpactIncidentList {
	impacts := make(api.ImpactIncidentList, len(*c.ActivelyAffectedBy))

	for impactIndex, impact := range *c.ActivelyAffectedBy {
		incidentID := impact.IncidentID.String()
		typeID := impact.ImpactTypeID.String()
		impacts[impactIndex].Reference = &incidentID
		impacts[impactIndex].Type = &typeID
	}

	return &impacts
}
