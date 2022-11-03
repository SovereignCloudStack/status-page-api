package main

import (
	"fmt"

	"gorm.io/gorm"
)

type ImpactType struct {
	Slug        string `gorm:"primaryKey" json:"slug"`
	provisioned bool   `json:""`
}

func (t *ImpactType) BeforeCreate(tx *gorm.DB) error {
	if tx.Find(t).RowsAffected == 0 && !t.provisioned {
		return fmt.Errorf("attempted to create non-provisioned phase %v", t.Slug)
	}
	return nil
}

func (t *ImpactType) BeforeUpdate(tx *gorm.DB) error {
	if tx.Find(t).RowsAffected == 0 && !t.provisioned {
		return fmt.Errorf("attempted to update non-provisioned phase %v", t.Slug)
	}
	return nil
}
