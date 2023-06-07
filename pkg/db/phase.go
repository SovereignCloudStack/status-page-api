package db

import "github.com/SovereignCloudStack/status-page-openapi/pkg/api"

// Phase represents a state of an incident on a movin scale to resolution of the incident.
type Phase struct {
	api.Phase
	Generation api.Incremental `gorm:"primaryKey"`
	Order      api.Incremental `gorm:"primaryKey"`
}
