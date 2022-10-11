package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Incident struct {
	ID         string       `gorm:"primaryKey"`
	CreatedAt  time.Time    `json:"createdAt"`
	Title      string       `json:"title"`
	Components []*Component `gorm:"many2many:incident_component;" json:"components,omitempty"`
	Updates    []*Update    `json:"updates,omitempty"`
}

func (i *Incident) BeforeCreate(tx *gorm.DB) error {
	i.ID = uuid.NewString()
	i.CreatedAt = time.Now()
	return nil
}

func incidentGet(c echo.Context) error {
	incident := &Incident{ID: c.Param("id")}
	err := db.Preload("Components").Preload("Updates").Take(&incident).Error
	switch err {
	case nil:
		return c.JSON(200, incident)
	case gorm.ErrRecordNotFound:
		return c.JSON(404, nil)
	default:
		c.Logger().Error(err)
		return c.JSON(500, nil)
	}
}

func incidentList(c echo.Context) error {
	var incidents []Incident
	err := db.Find(&incidents).Error
	switch err {
	case nil:
		return c.JSON(200, incidents)
	default:
		c.Logger().Error(err)
		return c.JSON(500, nil)
	}
}

func incidentAdd(c echo.Context) error {
	newIncident := Incident{}
	err := c.Bind(&newIncident)
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(400, nil)
	}
	err = db.Create(&newIncident).Error
	switch err {
	case nil:
		return c.JSON(200, newIncident)
	default:
		c.Logger().Error(err)
		return c.JSON(500, nil)
	}
}
