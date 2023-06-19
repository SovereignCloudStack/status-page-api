package db

import (
	"encoding/json"
	"fmt"

	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
)

// Labels are metadata for components.
type Labels api.Labels

// Scan implements the [database/sql.Scanner] interface to correctly read data.
func (l *Labels) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("%w: %v", ErrInvalidLabelData, value)
	}

	return json.Unmarshal(data, l) //nolint:wrapcheck
}
