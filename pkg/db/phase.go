package db

import "github.com/SovereignCloudStack/status-page-openapi/pkg/api"

// Phase represents a state of an incident on a moving scale to resolution of the incident.
type Phase struct {
	Name       *api.Phase       `yaml:"name"`
	Generation *api.Incremental `gorm:"primaryKey"`
	Order      *api.Incremental `gorm:"primaryKey"`
}

// PhaseReferenceFromAPI creates a [Phase] from an API request.
func PhaseReferenceFromAPI(phase *api.PhaseReference) (*Phase, error) {
	if phase == nil {
		return nil, ErrEmptyValue
	}

	return &Phase{ //nolint:exhaustruct
		Generation: &phase.Generation,
		Order:      &phase.Order,
	}, nil
}
