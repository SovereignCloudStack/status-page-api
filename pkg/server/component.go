package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm/clause"
)

// GetComponent retrieves a specific component by ID.
func (i *Implementation) GetComponent(ctx echo.Context, componentID string) error {
	var component DbDef.Component

	res := i.dbCon.Preload(clause.Associations).Where("id = ?", componentID).First(&component)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, componentFromDB(&component))
}

// GetComponents retrieves a list of all components.
func (i *Implementation) GetComponents(ctx echo.Context) error {
	var components []*DbDef.Component

	res := i.dbCon.Preload(clause.Associations).Find(&components)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	componentList := make([]*api.Component, len(components))
	for componentIndex := range componentList {
		componentList[componentIndex] = componentFromDB(components[componentIndex])
	}

	return ctx.JSON(http.StatusOK, componentList)
}

// componentFromDB is a helper function, converting a [db.Component] to an [api.Component].
func componentFromDB(component *DbDef.Component) *api.Component {
	return &api.Component{
		AffectedBy:  component.GetAffectedByIDs(),
		DisplayName: component.DisplayName,
		Id:          string(component.ID),
		Labels:      component.GetLabelMap(),
	}
}
