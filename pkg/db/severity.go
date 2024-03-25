package db

import "github.com/SovereignCloudStack/status-page-openapi/pkg/api"

// Severity represents a severity of a incident affecting a component.
type Severity struct {
	DisplayName *api.DisplayName
	Value       *api.SeverityValue `grom:"type:smallint"`
}

// ToAPIResponse converts to API response.
func (s *Severity) ToAPIResponse() api.Severity {
	return api.Severity{
		DisplayName: s.DisplayName,
		Value:       s.Value,
	}
}

// SeverityFromAPI creates a [Severity] from an API request.
func SeverityFromAPI(severityRequest *api.SeverityRequest) (*Severity, error) {
	if severityRequest == nil {
		return nil, ErrEmptyValue
	}

	return &Severity{
		DisplayName: severityRequest.DisplayName,
		Value:       severityRequest.Value,
	}, nil
}

// NewSeverity checks the value and creates a new severity.
func NewSeverity(displayName api.DisplayName, value api.SeverityValue) (*Severity, error) {
	if value < 0 || value > 100 {
		return nil, ErrSeverityValueOutOfRange
	}

	return &Severity{
		DisplayName: &displayName,
		Value:       &value,
	}, nil
}
