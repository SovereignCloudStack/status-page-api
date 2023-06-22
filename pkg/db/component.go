package db

import (
	"errors"
	"fmt"

	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
)

// Component represents a single component that could be affected by many [Incident].
type Component struct {
	Model              `gorm:"embedded"`
	DisplayName        *api.DisplayName `yaml:"displayname"`
	Labels             *Labels          `gorm:"type:jsonb"             yaml:"labels"`
	ActivelyAffectedBy *[]Impact        `gorm:"foreignKey:ComponentID"`
}

// GetImpactIncidentList converts the impact list.
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

// ComponentFromAPI creates a [Component] from an API request.
func ComponentFromAPI(componentRequest *api.Component) (*Component, error) {
	if componentRequest == nil {
		return nil, ErrEmptyValue
	}

	activelyAffectedBy, err := ActivelyAffectedByFromImpactIncidentList(componentRequest.ActivelyAffectedBy)
	if err != nil {
		if !errors.Is(err, ErrEmptyValue) {
			return nil, fmt.Errorf("error parsing actively affected by: %w", err)
		}
	}

	component := Component{ //nolint:exhaustruct
		DisplayName:        componentRequest.DisplayName,
		Labels:             (*Labels)(componentRequest.Labels),
		ActivelyAffectedBy: activelyAffectedBy,
	}

	return &component, nil
}
