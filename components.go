package main

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Component struct {
	Slug       string      `gorm:"primaryKey" json:"slug"`
	Incidents  []*Incident `gorm:"many2many:incident_component;" json:"incidents,omitempty"`
	Conditions []string    `gorm:"-" json:"conditions"` // computed field
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
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Take(&out).Error
		if err != nil {
			return err
		}
		currentIncidents := []*Incident{}
		err = tx.Where(&Incident{Components: []*Component{out}}).Find(&currentIncidents).Error
		if err != nil {
			return err
		}
		for _, incident := range currentIncidents {
			out.Conditions = append(out.Conditions, incident.ImpactTypeSlug)
		}
		return nil
	})
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
