package db

import (
	"fmt"

	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/google/uuid"
)

// ImpactType represents the type of impact.
type ImpactType struct {
	Model       `gorm:"embedded"`
	DisplayName *api.DisplayName `gorm:"not null"    yaml:"displayname"`
	Description *api.Description `yaml:"description"`
}

// Update updates the writables fields from an API request.
func (it *ImpactType) Update(impactType *api.ImpactType) {
	if impactType.DisplayName != nil {
		it.DisplayName = impactType.DisplayName
	}

	if impactType.Description != nil {
		it.Description = impactType.Description
	}
}

// Impact connect a [Incident] with a [Component] and [ImpactType].
type Impact struct {
	Incident   *Incident   `gorm:"foreignKey:IncidentID"`
	Component  *Component  `gorm:"foreignKey:ComponentID;constraint:OnDelete:CASCADE"`
	ImpactType *ImpactType `gorm:"foreignKey:ImpactTypeID"`

	IncidentID   *ID `gorm:"primaryKey"`
	ComponentID  *ID `gorm:"primaryKey"`
	ImpactTypeID *ID `gorm:"primaryKey"`
}

// ActivelyAffectedByFromImpactIncidentList parses a [api.ImpactIncidentList] to an [Impact] list.
func ActivelyAffectedByFromImpactIncidentList(incidentImpacts *api.ImpactIncidentList) (*[]Impact, error) {
	if incidentImpacts == nil {
		return nil, ErrEmptyValue
	}

	impacts := make([]Impact, len(*incidentImpacts))

	for impactIndex, impact := range *incidentImpacts {
		incidentID, err := uuid.Parse(*impact.Reference)
		if err != nil {
			return nil, fmt.Errorf("error parsing incident id: %w", err)
		}

		impactTypeID, err := uuid.Parse(*impact.Type)
		if err != nil {
			return nil, fmt.Errorf("error parsing impact type id: %w", err)
		}

		impacts[impactIndex].IncidentID = &incidentID
		impacts[impactIndex].ImpactTypeID = &impactTypeID
	}

	return &impacts, nil
}

// AffectsFromImpactComponentList parses a [api.ImpactComponentList] to an [Impact] list.
func AffectsFromImpactComponentList(componentImpacts *api.ImpactComponentList) (*[]Impact, error) {
	if componentImpacts == nil {
		return nil, ErrEmptyValue
	}

	impacts := make([]Impact, len(*componentImpacts))

	for impactIndex, impact := range *componentImpacts {
		componentID, err := uuid.Parse(*impact.Reference)
		if err != nil {
			return nil, fmt.Errorf("error parsing component id: %w", err)
		}

		impactTypeID, err := uuid.Parse(*impact.Type)
		if err != nil {
			return nil, fmt.Errorf("error parsing impact type id: %w", err)
		}

		impacts[impactIndex].ComponentID = &componentID
		impacts[impactIndex].ImpactTypeID = &impactTypeID
	}

	return &impacts, nil
}
