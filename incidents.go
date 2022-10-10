package main

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Incident struct {
	gorm.Model
	Title string `json:"title"`
}

func incidentGet(c echo.Context) error {
	var incident Incident
	err := db.First(&incident, c.Param("id")).Error
	switch err {
	case nil:
		return c.JSON(200, incident)
	case gorm.ErrRecordNotFound:
		return c.JSON(404, nil)
	default:
		c.Error(err)
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
		c.Error(err)
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
		c.Error(err)
		return c.JSON(500, nil)
	}
}
