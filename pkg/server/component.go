package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm/clause"
)

func (i *Implementation) GetComponent(ctx echo.Context, componentId string) error {
	var component DbDef.Component

	res := i.dbCon.Preload(clause.Associations).Where("id = ?", componentId).First(&component)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error)
	}

	return ctx.JSON(http.StatusOK, componentFromDB(&component))
}

func (i *Implementation) GetComponents(ctx echo.Context) error {
	var components []*DbDef.Component

	res := i.dbCon.Preload(clause.Associations).Find(&components)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	componentList := make([]*api.Component, len(components))
	for componentIndex := range componentList {
		componentList[componentIndex] = componentFromDB(components[componentIndex])
	}

	return ctx.JSON(http.StatusOK, componentList)
}

func componentFromDB(component *DbDef.Component) *api.Component {
	return &api.Component{
		AffectedBy:  component.GetAffectedByIDs(),
		DisplayName: component.DisplayName,
		Id:          string(component.ID),
		Labels:      component.GetLabelMap(),
	}
}
