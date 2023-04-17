package db

// ImpactType represents the type of impact for an incident.
type ImpactType struct {
	Slug string `gorm:"primaryKey" json:"slug"`
}
