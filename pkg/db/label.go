package db

import (
	"encoding/json"
	"fmt"

	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
)

type Labels api.Labels

func (l *Labels) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("%w: %v", ErrInvalidLabelData, value)
	}

	return json.Unmarshal(data, l) //nolint:wrapcheck
}
