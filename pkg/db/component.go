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

func (c *Component) GetAffectedByIDs() []string {
	incidentIds := make([]string, len(c.AffectedBy))

	for incidentIndex, incident := range c.AffectedBy {
		incidentIds[incidentIndex] = string(incident.ID)
	}

	return incidentIds
}

func (c *Component) GetLabelMap() map[string]string {
	labelMap := make(map[string]string)

	for _, label := range c.Labels {
		labelMap[label.Name] = label.Value
	}

	return labelMap
}
