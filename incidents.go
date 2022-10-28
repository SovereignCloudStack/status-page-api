package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Incident struct {
	ID             string       `gorm:"primaryKey"`
	CreatedAt      time.Time    `json:"createdAt"`
	Title          string       `json:"title"`
	Components     []*Component `gorm:"many2many:incident_component;" json:"components,omitempty"`
	Updates        []*Update    `json:"updates,omitempty"`
	ImpactTypeSlug string       `json:"-"`
	ImpactType     ImpactType   `gorm:"foreignKey:ImpactTypeSlug" json:"impactType"`
}

func (i *Incident) BeforeCreate(tx *gorm.DB) error {
	i.ID = uuid.NewString()
	i.CreatedAt = time.Now()
	return nil
}

func incidentGet(c echo.Context) error {
	incident := &Incident{ID: c.Param("id")}
	err := db.Preload("ImpactType").Preload("Components").Preload("Updates").Take(&incident).Error
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
	err := db.Preload("ImpactType").Find(&incidents).Error
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
