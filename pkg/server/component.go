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

	return ctx.JSON(http.StatusOK, api.Component{
		AffectedBy:  component.GetAffectedByIDs(),
		DisplayName: component.DisplayName,
		Id:          string(component.ID),
		Labels:      component.GetLabelMap(),
	})
}

func (i *Implementation) GetComponents(ctx echo.Context) error {
	var components []DbDef.Component

	res := i.dbCon.Preload(clause.Associations).Find(&components)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	componentList := make([]api.Component, len(components))
	for componentIndex, component := range components {
		componentList[componentIndex].AffectedBy = component.GetAffectedByIDs()
		componentList[componentIndex].DisplayName = component.DisplayName
		componentList[componentIndex].Id = string(component.ID)
		componentList[componentIndex].Labels = component.GetLabelMap()
	}

	return ctx.JSON(http.StatusOK, componentList)
}
