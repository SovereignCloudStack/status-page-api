package db

import (
	"github.com/SovereignCloudStack/status-page-api/pkg/api"
	apiServerDefinition "github.com/SovereignCloudStack/status-page-openapi/pkg/api/server"
)

// ImpactType represents the type of impact.
type ImpactType struct {
	DisplayName *apiServerDefinition.DisplayName `gorm:"not null"    yaml:"displayname"`
	Description *apiServerDefinition.Description `yaml:"description"`
	Model       `gorm:"embedded"`
}

// ToAPIResponse converts to API response.
func (it *ImpactType) ToAPIResponse() apiServerDefinition.ImpactTypeResponseData {
	return apiServerDefinition.ImpactTypeResponseData{
		Id:          it.ID,
		DisplayName: it.DisplayName,
		Description: it.Description,
	}
}

// ImpactTypeFromAPI creates an [ImpactType] from an API request.
func ImpactTypeFromAPI(impactTypeRequest *apiServerDefinition.ImpactType) (*ImpactType, error) {
	if impactTypeRequest == nil {
		return nil, ErrEmptyValue
	}

	return &ImpactType{ //nolint:exhaustruct
		DisplayName: impactTypeRequest.DisplayName,
		Description: impactTypeRequest.Description,
	}, nil
}

// Impact connect a [Incident] with a [Component] and [ImpactType].
type Impact struct {
	Incident   *Incident   `gorm:"foreignKey:IncidentID"`
	Component  *Component  `gorm:"foreignKey:ComponentID"`
	ImpactType *ImpactType `gorm:"foreignKey:ImpactTypeID"`

	IncidentID   *ID `gorm:"primaryKey"`
	ComponentID  *ID `gorm:"primaryKey"`
	ImpactTypeID *ID `gorm:"primaryKey"`

	Severity *apiServerDefinition.SeverityValue `gorm:"type:smallint"`
}

// AffectsFromImpactComponentList parses a [apiServerDefinition.ImpactComponentList] to an [Impact] list.
func AffectsFromImpactComponentList(componentImpacts *apiServerDefinition.ImpactComponentList) (*[]Impact, error) {
	if componentImpacts == nil {
		return nil, ErrEmptyValue
	}

	impacts := make([]Impact, len(*componentImpacts))

	for impactIndex, impact := range *componentImpacts {
		if impact.Severity != nil && (*impact.Severity < api.MaintenanceSeverity || *impact.Severity > api.MaxSeverity) {
			return nil, ErrSeverityValueOutOfRange
		}

		impacts[impactIndex].ComponentID = impact.Reference
		impacts[impactIndex].ImpactTypeID = impact.Type

		impacts[impactIndex].Severity = impact.Severity
	}

	return &impacts, nil
}
