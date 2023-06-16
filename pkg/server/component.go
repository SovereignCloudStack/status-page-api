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

	// TODO: only load active impacts for components.
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

// CreateComponent handles creation of components.
func (i *Implementation) CreateComponent(ctx echo.Context) error {
	var request api.CreateComponentJSONRequestBody

	err := ctx.Bind(&request)
	if err != nil {
		return echo.ErrInternalServerError
	}

	i.logger.Debug().Interface("request", request).Msg("creating component")

	component := DbDef.Component{ //nolint:exhaustruct
		DisplayName: request.DisplayName,
		Labels:      (*DbDef.Labels)(request.Labels),
	}

	res := i.dbCon.Save(&component)

	err = res.Error
	if err != nil {
		i.logger.Error().Err(err).Msg("error creating component")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusOK, api.IdResponse{ //nolint:wrapcheck
		Id: component.ID.String(),
	})
}

// DeleteComponent handles deletion of components.
func (i *Implementation) DeleteComponent(ctx echo.Context, componentID api.ComponentIdPathParameter) error {
	i.logger.Debug().Str("id", componentID).Msg("deleting component")

	res := i.dbCon.Where("id = ?", componentID).Delete(&DbDef.Component{}) //nolint: exhaustruct

	err := res.Error
	if err != nil {
		i.logger.Error().Err(err).Msg("error deleting component")

		return echo.ErrInternalServerError
	}

	if res.RowsAffected == 0 {
		return echo.ErrNotFound
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
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

// UpdateComponent handles updates of components.
func (i *Implementation) UpdateComponent(ctx echo.Context, componentID api.ComponentIdPathParameter) error {
	var (
		request   api.UpdateComponentJSONRequestBody
		component DbDef.Component
	)

	err := ctx.Bind(&request)
	if err != nil {
		i.logger.Error().Err(err).Msg("error binding update component request")

		return echo.ErrInternalServerError
	}

	res := i.dbCon.Where("id = ?", componentID).First(&component)

	err = res.Error
	if err != nil {
		i.logger.Error().Err(err).Msg("error receiving component for update")

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	i.logger.Debug().Interface("src", request).Interface("dst", component).Str("id", componentID).Msg("updating component")

	component.Update(&request)

	res = i.dbCon.Save(component)

	err = res.Error
	if err != nil {
		i.logger.Error().Err(err).Msg("error saving component after update")

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}
