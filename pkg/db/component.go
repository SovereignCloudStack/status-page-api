package db

import (
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
)

// Component represents a single component that could be affected by many [Incident].
type Component struct {
	Model              `gorm:"embedded"`
	DisplayName        *api.DisplayName `yaml:"displayname"`
	Labels             *Labels          `gorm:"type:jsonb" yaml:"labels"`
	ActivelyAffectedBy []*Impact        `gorm:"foreignKey:ComponentID"`
}
