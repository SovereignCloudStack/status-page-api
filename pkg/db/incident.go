package db

import (
	"errors"
	"fmt"

	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/google/uuid"
)

// Incident represents an incident happening to one or more [Component].
type Incident struct {
	Model           `gorm:"embedded"`
	DisplayName     *api.DisplayName
	Description     *api.Description
	Affects         *[]Impact `gorm:"foreignKey:IncidentID;constraint:OnDelete:CASCADE"`
	BeganAt         *api.Date
	EndedAt         *api.Date
	PhaseGeneration *api.Incremental
	PhaseOrder      *api.Incremental
	Phase           *Phase            `gorm:"foreignKey:PhaseGeneration,PhaseOrder;References:Generation,Order"`
	Updates         *[]IncidentUpdate `gorm:"foreignKey:IncidentID"`
}

// ToAPIResponse converts to API response.
func (i *Incident) ToAPIResponse() api.IncidentResponseData {
	return api.IncidentResponseData{
		Id:          i.ID.String(),
		DisplayName: i.DisplayName,
		Description: i.Description,
		BeganAt:     i.BeganAt,
		EndedAt:     i.EndedAt,
		Phase: &api.PhaseReference{
			Generation: *i.PhaseGeneration,
			Order:      *i.PhaseOrder,
		},
		Affects: i.GetImpactComponentList(),
		Updates: i.GetIncidentUpdates(),
	}
}

// GetImpactComponentList converts the Affects list to an [api.ImpactComponentList].
func (i *Incident) GetImpactComponentList() *api.ImpactComponentList {
	impacts := make(api.ImpactComponentList, len(*i.Affects))

	for impactIndex, impact := range *i.Affects {
		componentID := impact.ComponentID.String()
		typeID := impact.ImpactTypeID.String()
		impacts[impactIndex].Reference = &componentID
		impacts[impactIndex].Type = &typeID
	}

	return &impacts
}

// GetIncidentUpdates converts the Updates list to an [api.IncrementalList].
func (i *Incident) GetIncidentUpdates() *api.IncrementalList {
	updates := make(api.IncrementalList, len(*i.Updates))

	for updateIndex, update := range *i.Updates {
		updates[updateIndex] = *update.Order
	}

	return &updates
}

// IncidentFromAPI creates an [Incident] from an API request.
func IncidentFromAPI(incidentRequest *api.Incident) (*Incident, error) {
	if incidentRequest == nil {
		return nil, ErrEmptyValue
	}

	affects, err := AffectsFromImpactComponentList(incidentRequest.Affects)
	if err != nil {
		if !errors.Is(err, ErrEmptyValue) {
			return nil, fmt.Errorf("error parsing affects: %w", err)
		}
	}

	phase, err := PhaseReferenceFromAPI(incidentRequest.Phase)
	if err != nil {
		if !errors.Is(err, ErrEmptyValue) {
			return nil, fmt.Errorf("error parsing phase: %w", err)
		}
	}

	incident := Incident{ //nolint:exhaustruct
		DisplayName: incidentRequest.DisplayName,
		Description: incidentRequest.Description,
		BeganAt:     incidentRequest.BeganAt,
		EndedAt:     incidentRequest.EndedAt,
		Phase:       phase,
		Affects:     affects,
	}

	return &incident, nil
}

// IncidentUpdate describes a action that changes the incident.
type IncidentUpdate struct {
	IncidentID  *ID              `gorm:"primaryKey"`
	Order       *api.Incremental `gorm:"primaryKey"`
	DisplayName *api.DisplayName
	Description *api.Description
	CreatedAt   *api.Date
}

// ToAPIResponse converts to API response.
func (iu *IncidentUpdate) ToAPIResponse() api.IncidentUpdateResponseData {
	return api.IncidentUpdateResponseData{
		Order:       *iu.Order,
		DisplayName: iu.DisplayName,
		Description: iu.Description,
		CreatedAt:   iu.CreatedAt,
	}
}

// IncidentUpdateFromAPI creates an [IncidentUpdate] from an API request.
func IncidentUpdateFromAPI(
	incidentUpdateRequest *api.IncidentUpdate,
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
