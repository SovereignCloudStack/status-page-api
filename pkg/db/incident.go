package db

import "time"

type Incident struct {
	Id          Id               `gorm:"primaryKey" json:"id"`
	Affects     []Component      `gorm:"many2many:component_incidents" json:"affects"`
	BeganAt     *time.Time       `json:"beganAt,omitempty"`
	Description string           `json:"description"`
	EndedAt     *time.Time       `json:"endedAt"`
	ImpactType  ImpactType       `gorm:"many2many:incident_impact_types;references:slug" json:"impactType"`
	Phase       Phase            `gorm:"foreignKey:slug" json:"phase"`
	Title       string           `json:"title"`
	Updates     []IncidentUpdate `gorm:"foreignKey:Id" json:"updates"`
}

type IncidentUpdate struct {
	Id        Id        `gorm:"primaryKey"`
	CreatedAt time.Time `json:"createdAt"`
	Text      string    `json:"text"`
}
