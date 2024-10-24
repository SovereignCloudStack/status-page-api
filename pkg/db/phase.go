package db

import apiServerDefinition "github.com/SovereignCloudStack/status-page-openapi/pkg/api/server"

// Phase represents a state of an incident on a moving scale to resolution of the incident.
type Phase struct {
	Name       *apiServerDefinition.Phase       `gorm:"not null"   yaml:"name"`
	Generation *apiServerDefinition.Incremental `gorm:"primaryKey"`
	Order      *apiServerDefinition.Incremental `gorm:"primaryKey"`
}

// PhaseReferenceFromAPI creates a [Phase] from an API request.
func PhaseReferenceFromAPI(phase *apiServerDefinition.PhaseReference) (*Phase, error) {
	if phase == nil {
		return nil, ErrEmptyValue
	}

	return &Phase{ //nolint:exhaustruct
		Generation: &phase.Generation,
		Order:      &phase.Order,
	}, nil
}
