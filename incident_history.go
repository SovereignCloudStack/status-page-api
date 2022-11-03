package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/wI2L/jsondiff"
)

type IncidentHistory []jsondiff.Patch

func (h *IncidentHistory) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("data is of invalid type: %v", value)
	}
	return json.Unmarshal(data, h)
}

func (h IncidentHistory) Value() (driver.Value, error) {
	return json.Marshal(h)
}
