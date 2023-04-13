package db

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Component struct {
	ID          ID         `gorm:"primaryKey" json:"id"`
	AffectedBy  []Incident `gorm:"many2many:component_incidents" json:"affectedBy"`
	DisplayName string     `json:"displayName"`
	Labels      Labels     `gorm:"many2many:component_labels" json:"labels"`
}

func (c *Component) BeforeCreate(_ *gorm.DB) error {
	c.ID = ID(uuid.NewString())

	return nil
}
