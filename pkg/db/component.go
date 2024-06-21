package db

import apiServerDefinition "github.com/SovereignCloudStack/status-page-openapi/pkg/api/server"

// Component represents a single component that could be affected by many [Incident].
type Component struct {
	Model              `gorm:"embedded"`
	DisplayName        *apiServerDefinition.DisplayName `yaml:"displayname"`
	Labels             *Labels                          `gorm:"type:jsonb"             yaml:"labels"`
	ActivelyAffectedBy *[]Impact                        `gorm:"foreignKey:ComponentID"`
}

// ToAPIResponse converts to API response.
func (c *Component) ToAPIResponse() apiServerDefinition.ComponentResponseData {
	return apiServerDefinition.ComponentResponseData{
		Id:                 *c.ID,
		DisplayName:        c.DisplayName,
		Labels:             (*apiServerDefinition.Labels)(c.Labels),
		ActivelyAffectedBy: c.GetImpactIncidentList(),
	}
}

// GetImpactIncidentList converts the impact list.
func (c *Component) GetImpactIncidentList() *apiServerDefinition.ImpactIncidentList {
	impacts := make(apiServerDefinition.ImpactIncidentList, len(*c.ActivelyAffectedBy))

	for impactIndex, impact := range *c.ActivelyAffectedBy {
		impacts[impactIndex].Reference = impact.IncidentID
		impacts[impactIndex].Type = impact.ImpactTypeID
		impacts[impactIndex].Severity = impact.Severity
	}

	return &impacts
}

// ComponentFromAPI creates a [Component] from an API request.
func ComponentFromAPI(componentRequest *apiServerDefinition.Component) (*Component, error) {
	if componentRequest == nil {
		return nil, ErrEmptyValue
	}

	component := Component{ //nolint:exhaustruct
		DisplayName: componentRequest.DisplayName,
		Labels:      (*Labels)(componentRequest.Labels),
	}

	return &component, nil
}
