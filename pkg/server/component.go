package server

import (
	"errors"
	"net/http"
	"time"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	apiServerDefinition "github.com/SovereignCloudStack/status-page-openapi/pkg/api/server"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func incidentJoin(at *time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		join := db.Joins("Incident")

		if at != nil {
			join = join.Where("began_at < ? AND ended_at > ?", at, at).Or("began_at < ? AND ended_at IS NULL", at)
		} else {
			join = join.Where("ended_at IS NULL")
		}

		return join
	}
}

// GetComponents retrieves a list of all components.
func (i *Implementation) GetComponents(ctx echo.Context, params apiServerDefinition.GetComponentsParams) error {
	var components []*DbDef.Component

	logger := i.logger.With().Str("handler", "GetComponents").Logger()
	logger.Debug().Interface("at", params.At).Send()

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.Preload("ActivelyAffectedBy", incidentJoin(params.At)).Find(&components)

	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error loading components")

		return echo.ErrInternalServerError
	}

	data := make([]apiServerDefinition.ComponentResponseData, len(components))
	for componentIndex, component := range components {
		data[componentIndex] = component.ToAPIResponse()
	}

	return ctx.JSON(http.StatusOK, apiServerDefinition.ComponentListResponse{ //nolint:wrapcheck
		Data: data,
	})
}

// CreateComponent handles creation of components.
func (i *Implementation) CreateComponent(ctx echo.Context) error { //nolint:dupl
	var request apiServerDefinition.CreateComponentJSONRequestBody

	logger := i.logger.With().Str("handler", "CreateComponent").Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	if request == (apiServerDefinition.CreateComponentJSONRequestBody{}) { //nolint: exhaustruct
		logger.Warn().Msg("empty request")

		return echo.ErrBadRequest
	}

	logger.Debug().Interface("request", request).Send()

	component, err := DbDef.ComponentFromAPI(&request)
	if err != nil {
		logger.Warn().Err(err).Msg("error parsing request")

		return echo.ErrBadRequest
	}

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.Create(&component)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error creating component")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusCreated, apiServerDefinition.IdResponse{ //nolint:wrapcheck
		Id: component.ID,
	})
}

// DeleteComponent handles deletion of components.
func (i *Implementation) DeleteComponent(
	ctx echo.Context,
	componentID apiServerDefinition.ComponentIdPathParameter,
) error {
	logger := i.logger.With().Str("handler", "DeleteComponent").Interface("id", componentID).Logger()
	logger.Debug().Send()

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.Where("id = ?", componentID).Delete(&DbDef.Component{}) //nolint: exhaustruct
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error deleting component")

		return echo.ErrInternalServerError
	}

	if res.RowsAffected == 0 {
		logger.Warn().Msg("component not found")

		return echo.ErrNotFound
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}

// GetComponent retrieves a specific component by ID.
func (i *Implementation) GetComponent(
	ctx echo.Context,
	componentID apiServerDefinition.ComponentIdPathParameter,
	params apiServerDefinition.GetComponentParams,
) error {
	var component DbDef.Component

	logger := i.logger.With().Str("handler", "GetComponent").Interface("id", componentID).Logger()
	logger.Debug().Send()

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.Preload("ActivelyAffectedBy", incidentJoin(params.At)).Where("id = ?", componentID).First(&component)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			logger.Warn().Msg("component not found")

			return echo.ErrNotFound
		}

		logger.Error().Err(res.Error).Msg("error loading component")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusOK, apiServerDefinition.ComponentResponse{ //nolint:wrapcheck
		Data: component.ToAPIResponse(),
	})
}

// UpdateComponent handles updates of components.
func (i *Implementation) UpdateComponent(
	ctx echo.Context,
	componentID apiServerDefinition.ComponentIdPathParameter,
) error {
	var request apiServerDefinition.UpdateComponentJSONRequestBody

	logger := i.logger.With().Str("handler", "UpdateComponent").Interface("id", componentID).Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding update component request")

		return echo.ErrInternalServerError
	}

	if request == (apiServerDefinition.UpdateComponentJSONRequestBody{}) { //nolint:exhaustruct
		logger.Warn().Msg("empty request")

		return echo.ErrBadRequest
	}

	logger.Debug().Interface("request", request).Send()

	component, err := DbDef.ComponentFromAPI(&request)
	if err != nil {
		logger.Warn().Err(err).Msg("error parsing request")

		return echo.ErrBadRequest
	}

	component.ID = componentID

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.Updates(component)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error updating component")

		return echo.ErrInternalServerError
	}

	if res.RowsAffected == 0 {
		logger.Warn().Msg("component not found")

		return echo.ErrNotFound
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}
