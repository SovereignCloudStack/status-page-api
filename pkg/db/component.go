package db

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Component struct {
	Id          Id         `gorm:"primaryKey" json:"id"`
	AffectedBy  []Incident `gorm:"many2many:component_incidents" json:"affectedBy"`
	DisplayName string     `json:"displayName"`
	Labels      Labels     `gorm:"many2many:component_labels" json:"labels"`
}

func (c *Component) BeforeCreate(tx *gorm.DB) error {
	c.Id = Id(uuid.NewString())

	return nil
}
