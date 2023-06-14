package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm/clause"
)

// GetComponents retrieves a list of all components.
func (i *Implementation) GetComponents(ctx echo.Context) error {
	var components []*DbDef.Component

	res := i.dbCon.Preload(clause.Associations).Find(&components)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	data := make([]api.ComponentResponseData, len(components))
	for componentIndex, component := range components {
		data[componentIndex].Id = component.ID.String()
		data[componentIndex].DisplayName = component.DisplayName
		data[componentIndex].Labels = (*api.Labels)(component.Labels)
		data[componentIndex].ActivelyAffectedBy = component.GetImpactIncidentList()
	}

	response := api.ComponentListResponse{
		Data: &data,
	}

	return ctx.JSON(http.StatusOK, response) //nolint:wrapcheck
}

func (i *Implementation) CreateComponent(_ echo.Context) error {
	return nil
}

func (i *Implementation) DeleteComponent(_ echo.Context, _ api.ComponentIdPathParameter) error {
	return nil
}

// GetComponent retrieves a specific component by ID.
func (i *Implementation) GetComponent(ctx echo.Context, componentID string) error {
	var component DbDef.Component

	res := i.dbCon.Preload(clause.Associations).Where("id = ?", componentID).First(&component)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	response := api.ComponentResponse{
		Data: &api.ComponentResponseData{
			Id:                 component.ID.String(),
			DisplayName:        component.DisplayName,
			Labels:             (*api.Labels)(component.Labels),
			ActivelyAffectedBy: component.GetImpactIncidentList(),
		},
	}

	return ctx.JSON(http.StatusOK, response) //nolint:wrapcheck
}

func (i *Implementation) UpdateComponent(_ echo.Context, _ api.ComponentIdPathParameter) error {
	return nil
}
