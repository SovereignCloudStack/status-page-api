package main

import (
	"os"

	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type Component struct {
	Slug      string      `gorm:"primaryKey" json:"slug"`
	Incidents []*Incident `gorm:"many2many:incident_component;" json:"incidents,omitempty"`
}

func loadComponents(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	configComponents := []Component{}
	err = yaml.NewDecoder(file).Decode(&configComponents)
	if err != nil {
		return err
	}
	for _, configComponent := range configComponents {
		err = db.Take(&configComponent).Error
		switch err {
		case gorm.ErrRecordNotFound:
			saveErr := db.Save(&configComponent).Error
			if saveErr != nil {
				return saveErr
			}
		case nil:
		default:
			return err
		}
	}
	return nil
}

func componentList(c echo.Context) error {
	out := []*Component{}
	err := db.Find(&out).Error
	switch err {
	case nil:
		return c.JSON(200, out)
	default:
		c.Logger().Error(err)
		return c.JSON(500, nil)
	}
}

func componentGet(c echo.Context) error {
	out := &Component{Slug: c.Param("slug")}
	err := db.Preload("Incidents").Take(&out).Error
	switch err {
	case nil:
		return c.JSON(200, out)
	case gorm.ErrRecordNotFound:
		return c.JSON(404, nil)
	default:
		c.Logger().Error(err)
		return c.JSON(500, nil)
	}
}
