package main

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Component struct {
	Slug       string      `gorm:"primaryKey" json:"slug"`
	Labels     Labels      `gorm:"type:jsonb" json:"labels"`
	Incidents  []*Incident `gorm:"many2many:incident_component;" json:"incidents,omitempty"`
	Conditions []string    `gorm:"-" json:"conditions"` // computed field
}

func componentLoad(filter interface{}) ([]*Component, error) {
	out := []*Component{}
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Find(&out, filter).Error
		if err != nil {
			return err
		}

		// TODO: Filter out inactive incidents or rewrite this part completely,
		// offloading matching to database
		currentIncidents := []*Incident{}
		err = tx.Preload("Components").Find(&currentIncidents).Error
		if err != nil {
			return err
		}
		for componentIdx := range out {
			for incidentIdx := range currentIncidents {
				for componentInIncidentIdx := range currentIncidents[incidentIdx].Components {
					if currentIncidents[incidentIdx].Components[componentInIncidentIdx].Slug == out[componentIdx].Slug {
						out[componentIdx].Conditions = append(out[componentIdx].Conditions, currentIncidents[incidentIdx].ImpactTypeSlug)
					}
				}
			}
		}
		return nil
	})
	return out, err
}

func componentList(c echo.Context) error {
	out, err := componentLoad([]string{})
	switch err {
	case nil:
		return c.JSON(200, out)
	default:
		c.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
}

func componentGet(c echo.Context) error {
	out, err := componentLoad([]string{c.Param("slug")})
	if len(out) == 0 {
		return echo.NewHTTPError(404)
	}
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

func componentQueryByLabels(c echo.Context) error {
	labels := Labels{}
	err := c.Bind(&labels)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(400)
	}
	out, err := componentLoad(LabelFilter("labels").HasLabels(labels))
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
