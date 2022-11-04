package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/wI2L/jsondiff"
	"gorm.io/gorm"
)

type Incident struct {
	IncidentState
	ID      string          `gorm:"primaryKey"`
	History IncidentHistory `gorm:"type:jsonb" json:"history"`
}

type IncidentState struct {
	RecordCreatedAt time.Time    `gorm:"<-:create" json:"recordCreatedAt"`
	BeganAt         *time.Time   `json:"beganAt"`
	EndedAt         *time.Time   `json:"endedAt"`
	Title           string       `json:"title"`
	Components      []*Component `gorm:"many2many:incident_component;" json:"components,omitempty"`
	ImpactTypeSlug  string       `json:"-"`
	ImpactType      ImpactType   `gorm:"foreignKey:ImpactTypeSlug" json:"impactType"`
	PhaseSlug       string       `json:"-"`
	Phase           Phase        `gorm:"foreignKey:PhaseSlug" json:"phase"`
}

func (i *Incident) BeforeCreate(tx *gorm.DB) error {
	i.ID = uuid.NewString()
	i.RecordCreatedAt = time.Now()
	patch, err := jsondiff.Compare(Incident{}.IncidentState, i.IncidentState)
	if err != nil {
		return err
	}
	i.History = IncidentHistory{patch}
	return nil
}

func incidentGet(c echo.Context) error {
	incident := &Incident{ID: c.Param("id")}
	err := db.Preload("Phase").Preload("ImpactType").Preload("Components").Take(&incident).Error
	switch err {
	case nil:
		return c.JSON(200, incident)
	case gorm.ErrRecordNotFound:
		return echo.NewHTTPError(404)
	default:
		c.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
}

func incidentList(c echo.Context) error {
	var incidents []Incident
	err := db.Preload("Phase").Preload("ImpactType").Find(&incidents).Error
	switch err {
	case nil:
		return c.JSON(200, incidents)
	default:
		c.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
}

func incidentAdd(c echo.Context) error {
	newIncident := Incident{}
	err := c.Bind(&newIncident)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(400)
	}
	err = db.Create(&newIncident).Error
	switch err {
	case nil:
		return c.JSON(200, newIncident)
	default:
		c.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
}

func (i *Incident) Merge(other *Incident) *Incident {
	product := &Incident{
		ID:      i.ID,
		History: i.History,
		IncidentState: IncidentState{
			RecordCreatedAt: i.RecordCreatedAt,
			Title:           i.Title,
			ImpactType:      i.ImpactType,
			Phase:           i.Phase,
			Components:      i.Components,
			BeganAt:         i.BeganAt,
			EndedAt:         i.EndedAt,
		},
	}
	if len(other.Title) != 0 {
		product.Title = other.Title
	}
	if len(other.ImpactType.Slug) != 0 {
		product.ImpactType = other.ImpactType
	}
	if len(other.Phase.Slug) != 0 {
		product.Phase = other.Phase
	}
	if other.BeganAt != nil {
		product.BeganAt = other.BeganAt
	}
	if other.EndedAt != nil {
		product.EndedAt = other.EndedAt
	}
	if len(other.Components) != 0 {
		product.Components = other.Components
	}
	return product
}

func incidentUpdate(c echo.Context) error {
	// Get incident from request body
	newIncident := &Incident{}
	err := c.Bind(newIncident)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(400)
	}
	newIncident.ID = c.Param("id")

	err = db.Transaction(func(tx *gorm.DB) error {
		// Get full incident from DB
		currentIncident := &Incident{ID: newIncident.ID}
		err = tx.Preload("Phase").Preload("ImpactType").Take(&currentIncident).Error
		if err != nil {
			return err
		}

		// Constructing new desired incident state.
		// Take currentIncident as base, override fields that are set in request body
		newIncident = currentIncident.Merge(newIncident)

		// Compare
		// - currentIncident (what the incident was up until now)
		// - newIncident (what the client wants it to be)
		patch, err := jsondiff.Compare(currentIncident.IncidentState, newIncident.IncidentState)
		if err != nil {
			return err
		}
		if len(patch) != 0 {
			newIncident.History = append(newIncident.History, patch)
		}
		err = tx.Model(currentIncident).Association("Components").Replace(newIncident.Components)
		if err != nil {
			return err
		}

		return tx.Save(newIncident).Error
	})

	switch err {
	case nil:
		return c.JSON(200, newIncident)
	default:
		c.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
}
