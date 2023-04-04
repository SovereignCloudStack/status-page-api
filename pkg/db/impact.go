package db

type ImpactType struct {
	Slug string `gorm:"primaryKey" json:"slug"`
}
