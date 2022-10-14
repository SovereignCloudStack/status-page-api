package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Update struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	Content    string    `json:"content"`
	IncidentID string    `json:"-"`
}

func (u *Update) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.NewString()
	u.CreatedAt = time.Now()
	return nil
}

func updatesGet(c echo.Context) error {
	updates := []Update{}
	err := db.Where(&Update{IncidentID: c.Param("id")}).Find(&updates).Error
	switch err {
	case nil:
		return c.JSON(200, updates)
	default:
		c.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
}

func updateAdd(c echo.Context) error {
	newUpdate := Update{}
	err := c.Bind(&newUpdate)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(400)
	}
	newUpdate.IncidentID = c.Param("id")
	err = db.Create(&newUpdate).Error
	switch err {
	case nil:
		return c.JSON(200, newUpdate)
	default:
		c.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
}

func updateUpdate(c echo.Context) error {
	newUpdate := Update{}
	err := c.Bind(&newUpdate)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(400)
	}
	result := db.Where(
		&Update{ID: c.Param("updateid"), IncidentID: c.Param("id")},
	).Select("Content").Updates(newUpdate)
	if result.Error != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(500)
	}
	if result.RowsAffected == 0 {
		return echo.NewHTTPError(404)
	}
	return echo.NewHTTPError(200)
}
