package db

// Phase represents a state of an incident on a movin scale to resolution of the incident.
type Phase struct {
	Slug  string `gorm:"primaryKey" json:"slug"`
	Order uint   `gorm:"unique"`
}
