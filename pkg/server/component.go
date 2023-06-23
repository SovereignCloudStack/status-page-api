package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// GetComponents retrieves a list of all components.
func (i *Implementation) GetComponents(ctx echo.Context) error {
	var components []*DbDef.Component

	logger := i.logger.With().Str("handler", "GetComponents").Logger()
	logger.Debug().Send()

	res := i.dbCon.Preload("ActivelyAffectedBy", func(db *gorm.DB) *gorm.DB {
		return db.Joins("Incident").Where("ended_at IS NULL")
	}).Find(&components)

	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error loading components")

		return echo.ErrInternalServerError
	}

	data := make([]api.ComponentResponseData, len(components))
	for componentIndex, component := range components {
		data[componentIndex] = *component.ToAPIResponse()
	}

	return ctx.JSON(http.StatusOK, api.ComponentListResponse{ //nolint:wrapcheck
		Data: &data,
	})
}

// CreateComponent handles creation of components.
func (i *Implementation) CreateComponent(ctx echo.Context) error { //nolint:dupl
	var request api.CreateComponentJSONRequestBody

	logger := i.logger.With().Str("handler", "CreateComponent").Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	logger.Debug().Interface("request", request).Send()

	component, err := DbDef.ComponentFromAPI(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing request")

		return echo.ErrInternalServerError
	}

	res := i.dbCon.Create(&component)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error creating component")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusCreated, api.IdResponse{ //nolint:wrapcheck
		Id: component.ID.String(),
	})
}

// DeleteComponent handles deletion of components.
func (i *Implementation) DeleteComponent(ctx echo.Context, componentID api.ComponentIdPathParameter) error {
	logger := i.logger.With().Str("handler", "DeleteComponent").Str("id", componentID).Logger()
	logger.Debug().Send()

	res := i.dbCon.Where("id = ?", componentID).Delete(&DbDef.Component{}) //nolint: exhaustruct
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error deleting component")

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

	logger := i.logger.With().Str("handler", "GetComponent").Str("id", componentID).Logger()
	logger.Debug().Send()

	res := i.dbCon.Preload("ActivelyAffectedBy", func(db *gorm.DB) *gorm.DB {
		return db.Joins("Incident").Where("ended_at IS NULL")
	}).Where("id = ?", componentID).First(&component)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error loading component")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusOK, api.ComponentResponse{ //nolint:wrapcheck
		Data: component.ToAPIResponse(),
	})
}

// UpdateComponent handles updates of components.
func (i *Implementation) UpdateComponent(ctx echo.Context, componentID api.ComponentIdPathParameter) error {
	var request api.UpdateComponentJSONRequestBody

	logger := i.logger.With().Str("handler", "UpdateComponent").Str("id", componentID).Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding update component request")

		return echo.ErrInternalServerError
	}

	logger.Debug().Interface("request", request).Send()

	component, err := DbDef.ComponentFromAPI(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing request")

		return echo.ErrInternalServerError
	}

	componentUUID, err := uuid.Parse(componentID)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing id")
	}

	component.ID = &componentUUID

	res := i.dbCon.Updates(component)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error updating component")

		return echo.ErrInternalServerError
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}
