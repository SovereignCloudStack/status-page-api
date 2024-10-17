package db

import (
	"errors"
	"fmt"

	"github.com/SovereignCloudStack/status-page-api/pkg/api"
	apiServerDefinition "github.com/SovereignCloudStack/status-page-openapi/pkg/api/server"
	"github.com/google/uuid"
)

// Incident represents an incident happening to one or more [Component].
type Incident struct {
	DisplayName     *apiServerDefinition.DisplayName
	Description     *apiServerDefinition.Description
	Affects         *[]Impact `gorm:"foreignKey:IncidentID;constraint:OnDelete:CASCADE"`
	BeganAt         *apiServerDefinition.Date
	EndedAt         *apiServerDefinition.Date
	PhaseGeneration *apiServerDefinition.Incremental
	PhaseOrder      *apiServerDefinition.Incremental
	Phase           *Phase            `gorm:"foreignKey:PhaseGeneration,PhaseOrder;References:Generation,Order"`
	Updates         *[]IncidentUpdate `gorm:"foreignKey:IncidentID;constraint:OnDelete:CASCADE"`
	Model           `gorm:"embedded"`
}

// ToAPIResponse converts to API response.
func (i *Incident) ToAPIResponse() apiServerDefinition.IncidentResponseData {
	return apiServerDefinition.IncidentResponseData{
		Id:          i.ID,
		DisplayName: i.DisplayName,
		Description: i.Description,
		BeganAt:     i.BeganAt,
		EndedAt:     i.EndedAt,
		Phase: &apiServerDefinition.PhaseReference{
			Generation: *i.PhaseGeneration,
			Order:      *i.PhaseOrder,
		},
		Affects: i.GetImpactComponentList(),
		Updates: i.GetIncidentUpdates(),
	}
}

// GetImpactComponentList converts the Affects list to an [apiServerDefinition.ImpactComponentList].
func (i *Incident) GetImpactComponentList() *apiServerDefinition.ImpactComponentList {
	impacts := make(apiServerDefinition.ImpactComponentList, len(*i.Affects))

	for impactIndex, impact := range *i.Affects {
		impacts[impactIndex].Reference = impact.ComponentID
		impacts[impactIndex].Type = impact.ImpactTypeID
		impacts[impactIndex].Severity = impact.Severity
	}

	return &impacts
}

// GetIncidentUpdates converts the Updates list to an [apiServerDefinition.IncrementalList].
func (i *Incident) GetIncidentUpdates() *apiServerDefinition.IncrementalList {
	updates := make(apiServerDefinition.IncrementalList, len(*i.Updates))

	for updateIndex, update := range *i.Updates {
		updates[updateIndex] = *update.Order
	}

	return &updates
}

func isMaintenance(impacts *[]Impact) bool {
	if impacts == nil {
		return false
	}

	for _, impact := range *impacts {
		if impact.Severity == nil {
			continue
		}

		if *impact.Severity == api.MaintenanceSeverity {
			return true
		}
	}

	return false
}

// IncidentFromAPI creates an [Incident] from an API request.
func IncidentFromAPI(incidentRequest *apiServerDefinition.Incident) (*Incident, error) {
	if incidentRequest == nil {
		return nil, ErrEmptyValue
	}

	if incidentRequest.BeganAt != nil &&
		incidentRequest.EndedAt != nil &&
		incidentRequest.EndedAt.Before(*incidentRequest.BeganAt) {
		return nil, ErrEndsBeforeStart
	}

	affects, err := AffectsFromImpactComponentList(incidentRequest.Affects)
	if err != nil {
		if !errors.Is(err, ErrEmptyValue) {
			return nil, fmt.Errorf("error parsing affects: %w", err)
		}
	}

	if isMaintenance(affects) && incidentRequest.EndedAt == nil {
		return nil, ErrMaintenanceNeedsEnd
	}

	phase, err := PhaseReferenceFromAPI(incidentRequest.Phase)
	if err != nil {
		if !errors.Is(err, ErrEmptyValue) {
			return nil, fmt.Errorf("error parsing phase: %w", err)
		}
	}

	return &Incident{ //nolint:exhaustruct
		DisplayName: incidentRequest.DisplayName,
		Description: incidentRequest.Description,
		BeganAt:     incidentRequest.BeganAt,
		EndedAt:     incidentRequest.EndedAt,
		Phase:       phase,
		Affects:     affects,
	}, nil
}

// IncidentUpdate describes a action that changes the incident.
type IncidentUpdate struct {
	IncidentID  *ID                              `gorm:"primaryKey"`
	Order       *apiServerDefinition.Incremental `gorm:"primaryKey"`
	DisplayName *apiServerDefinition.DisplayName
	Description *apiServerDefinition.Description
	CreatedAt   *apiServerDefinition.Date
}

// ToAPIResponse converts to API response.
func (iu *IncidentUpdate) ToAPIResponse() apiServerDefinition.IncidentUpdateResponseData {
	return apiServerDefinition.IncidentUpdateResponseData{
		Order:       *iu.Order,
		DisplayName: iu.DisplayName,
		Description: iu.Description,
		CreatedAt:   iu.CreatedAt,
	}
}

// IncidentUpdateFromAPI creates an [IncidentUpdate] from an API request.
func IncidentUpdateFromAPI(
	incidentUpdateRequest *apiServerDefinition.IncidentUpdate,
	incidentID uuid.UUID,
	order int,
) (*IncidentUpdate, error) {
	if incidentUpdateRequest == nil {
		return nil, ErrEmptyValue
	}

	incidentUpdate := IncidentUpdate{
		IncidentID:  &incidentID,
		Order:       &order,
		DisplayName: incidentUpdateRequest.DisplayName,
		Description: incidentUpdateRequest.Description,
		CreatedAt:   incidentUpdateRequest.CreatedAt,
	}

	return &incidentUpdate, nil
}
