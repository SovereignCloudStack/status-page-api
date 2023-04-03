package db

type ImpactType struct {
	Slug string `gorm:"primaryKey;many2many:incident_impact_types" json:"slug"`
}
