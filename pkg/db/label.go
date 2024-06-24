package db

import (
	"encoding/json"
	"fmt"

	apiServerDefinition "github.com/SovereignCloudStack/status-page-openapi/pkg/api/server"
)

// Labels are metadata for components.
type Labels apiServerDefinition.Labels

// Scan implements the [database/sql.Scanner] interface to correctly read data.
func (l *Labels) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("%w: %v", ErrInvalidLabelData, value)
	}

	return json.Unmarshal(data, l) //nolint:wrapcheck
}
