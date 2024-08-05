package db

import apiServerDefinition "github.com/SovereignCloudStack/status-page-openapi/pkg/api/server"

// Severity represents a severity of a incident affecting a component.
type Severity struct {
	DisplayName *apiServerDefinition.DisplayName   `yaml:"name"`
	Value       *apiServerDefinition.SeverityValue `gorm:"type:smallint;unique" yaml:"value"`
}

// ToAPIResponse converts to API response.
func (s *Severity) ToAPIResponse() apiServerDefinition.Severity {
	return apiServerDefinition.Severity{
		DisplayName: s.DisplayName,
		Value:       s.Value,
	}
}

// SeverityFromAPI creates a [Severity] from an API request.
func SeverityFromAPI(severityRequest *apiServerDefinition.SeverityRequest) (*Severity, error) {
	if severityRequest == nil {
		return nil, ErrEmptyValue
	}

	return &Severity{
		DisplayName: severityRequest.DisplayName,
		Value:       severityRequest.Value,
	}, nil
}

// NewSeverity checks the value and creates a new severity.
func NewSeverity(
	displayName apiServerDefinition.DisplayName,
	value apiServerDefinition.SeverityValue,
) (*Severity, error) {
	if value < 0 || value > 100 {
		return nil, ErrSeverityValueOutOfRange
	}

	return &Severity{
		DisplayName: &displayName,
		Value:       &value,
	}, nil
}
