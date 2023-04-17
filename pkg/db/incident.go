package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Incident represents an incident happening to one or more [Component].
type Incident struct {
	ID             ID               `gorm:"primaryKey" json:"id"`
	Affects        []Component      `gorm:"many2many:component_incidents" json:"affects"`
	BeganAt        *time.Time       `json:"beganAt,omitempty"`
	Description    string           `json:"description"`
	EndedAt        *time.Time       `json:"endedAt"`
	ImpactTypeSlug string           `json:"-"`
	ImpactType     ImpactType       `gorm:"foreignKey:ImpactTypeSlug" json:"impactType"`
	PhaseSlug      string           `json:"-"`
	Phase          Phase            `gorm:"foreignKey:PhaseSlug" json:"phase"`
	Title          string           `json:"title"`
	Updates        []IncidentUpdate `gorm:"foreignKey:IncidentID" json:"updates"`
}

// IncidentUpdate describes a action that changes the incident.
type IncidentUpdate struct {
	ID         ID        `gorm:"primaryKey"`
	CreatedAt  time.Time `json:"createdAt"`
	Text       string    `json:"text"`
	IncidentID ID        `json:"-"`
}

// BeforeCreate implements the behavior before a database insertion. This adds an UUID as ID.
func (i *Incident) BeforeCreate(_ *gorm.DB) error {
	i.ID = ID(uuid.NewString())

	return nil
}

// BeforeCreate implements the behavior before a database insertion. This adds an UUID as ID.
func (iu *IncidentUpdate) BeforeCreate(_ *gorm.DB) error {
	iu.ID = ID(uuid.NewString())

	return nil
}

// GetAffectsIds is a helper function to convert the affected components to a list of [Component.ID]s.
func (i *Incident) GetAffectsIds() []string {
	componentIds := make([]string, len(i.Affects))

	for componentIndex, component := range i.Affects {
		componentIds[componentIndex] = string(component.ID)
	}

	return componentIds
}
