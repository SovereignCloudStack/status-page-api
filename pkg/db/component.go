package db

import (
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
)

// Component represents a single component that could be affected by many [Incident].
type Component struct {
	DisplayName        api.DisplayName
	Labels             api.Labels `gorm:"type:jsonb"`
	ActivelyAffectedBy []*Impact  `gorm:"foreignKey:ComponentID"`
	Model              `gorm:"embedded"`
}
