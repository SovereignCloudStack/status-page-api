package main

import (
	"fmt"

	"gorm.io/gorm"
)

type Phase struct {
	Slug        string `gorm:"primaryKey" json:"slug"`
	provisioned bool   `json:"-"`
}

func (p *Phase) BeforeCreate(tx *gorm.DB) error {
	if tx.Take(p).Error == gorm.ErrRecordNotFound && !p.provisioned {
		return fmt.Errorf("attempted to create non-provisioned phase %v", p.Slug)
	}
	return nil
}

func (p *Phase) BeforeUpdate(tx *gorm.DB) error {
	if tx.Take(p).Error == gorm.ErrRecordNotFound && !p.provisioned {
		return fmt.Errorf("attempted to update non-provisioned phase %v", p.Slug)
	}
	return nil
}
