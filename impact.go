package main

type ImpactType struct {
	Slug string `gorm:"primaryKey" json:"slug"`
}
