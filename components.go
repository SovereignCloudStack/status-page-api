package main

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Component struct {
	Slug      string      `gorm:"primaryKey" json:"slug"`
	Incidents []*Incident `gorm:"many2many:incident_component;" json:"incidents,omitempty"`
}

func componentList(c echo.Context) error {
	out := []*Component{}
	err := db.Find(&out).Error
	switch err {
	case nil:
		return c.JSON(200, out)
	default:
		c.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
}

func componentGet(c echo.Context) error {
	out := &Component{Slug: c.Param("slug")}
	err := db.Preload("Incidents").Take(&out).Error
	switch err {
	case nil:
		return c.JSON(200, out)
	case gorm.ErrRecordNotFound:
		return echo.NewHTTPError(404)
	default:
		c.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
}
