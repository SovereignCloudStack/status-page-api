package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Incident struct {
	Id             Id               `gorm:"primaryKey" json:"id"`
	Affects        []Component      `gorm:"many2many:component_incidents" json:"affects"`
	BeganAt        *time.Time       `json:"beganAt,omitempty"`
	Description    string           `json:"description"`
	EndedAt        *time.Time       `json:"endedAt"`
	ImpactTypeSlug string           `json:"-"`
	ImpactType     ImpactType       `gorm:"foreignKey:ImpactTypeSlug" json:"impactType"`
	PhaseSlug      string           `json:"-"`
	Phase          Phase            `gorm:"foreignKey:PhaseSlug" json:"phase"`
	Title          string           `json:"title"`
	Updates        []IncidentUpdate `json:"updates"`
}

type IncidentUpdate struct {
	Id         Id        `gorm:"primaryKey"`
	CreatedAt  time.Time `json:"createdAt"`
	Text       string    `json:"text"`
	IncidentId Id        `json:"-"`
}

func (i *Incident) BeforeCreate(tx *gorm.DB) error {
	i.Id = Id(uuid.NewString())

	return nil
}

func (iu *IncidentUpdate) BeforeCreate(tx *gorm.DB) error {
	iu.Id = Id(uuid.NewString())

	return nil
}
