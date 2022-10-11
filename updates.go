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
		return c.JSON(500, nil)
	}
}

func updateAdd(c echo.Context) error {
	newUpdate := Update{IncidentID: c.Param("id")}
	err := c.Bind(&newUpdate)
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(400, nil)
	}
	err = db.Create(&newUpdate).Error
	switch err {
	case nil:
		return c.JSON(200, newUpdate)
	default:
		c.Logger().Error(err)
		return c.JSON(500, nil)
	}
}
