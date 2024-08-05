package server

import (
	"errors"
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	apiServerDefinition "github.com/SovereignCloudStack/status-page-openapi/pkg/api/server"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// GetSeverities retrieves a list of all severities.
func (i *Implementation) GetSeverities(ctx echo.Context) error {
	var severities []*DbDef.Severity

	logger := i.logger.With().Str("handler", "GetSeverities").Logger()
	logger.Debug().Send()

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.Find(&severities)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error loading severites")

		return echo.ErrInternalServerError
	}

	data := make([]apiServerDefinition.Severity, len(severities))
	for severityIndex, severity := range severities {
		data[severityIndex] = severity.ToAPIResponse()
	}

	return ctx.JSON(http.StatusOK, apiServerDefinition.SeverityListResponse{ //nolint:wrapcheck
		Data: data,
	})
}

// CreateSeverity handles creation of severities.
func (i *Implementation) CreateSeverity(ctx echo.Context) error {
	var request apiServerDefinition.CreateSeverityJSONRequestBody

	logger := i.logger.With().Str("handler", "CreateSeverity").Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	if request == (apiServerDefinition.CreateSeverityJSONRequestBody{}) { //nolint: exhaustruct
		logger.Warn().Msg("empty request")

		return echo.ErrBadRequest
	}

	logger.Debug().Interface("request", request).Send()

	severity, err := DbDef.SeverityFromAPI(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing request")

		return echo.ErrBadRequest
	}

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.Create(&severity)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error creating severity")

		if errors.Is(res.Error, gorm.ErrDuplicatedKey) {
			return echo.ErrBadRequest
		}

		return echo.ErrInternalServerError
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}

// DeleteSeverity handles deletion of severities.
func (i *Implementation) DeleteSeverity(
	ctx echo.Context,
	severityName apiServerDefinition.SeverityNamePathParameter,
) error {
	logger := i.logger.With().Str("handler", "DeleteSeverity").Str("name", severityName).Logger()
	logger.Debug().Send()

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.Where("display_name = ?", severityName).Delete(&DbDef.Severity{}) //nolint: exhaustruct
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error deleting severity")

		return echo.ErrInternalServerError
	}

	if res.RowsAffected == 0 {
		logger.Warn().Msg("severity not found")

		return echo.ErrNotFound
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}

// GetSeverity retrieves a specific incident by it's name.
func (i *Implementation) GetSeverity( //nolint:dupl
	ctx echo.Context,
	severityName apiServerDefinition.SeverityNamePathParameter,
) error {
	var severity DbDef.Severity

	logger := i.logger.With().Str("handler", "GetSeverity").Str("name", severityName).Logger()
	logger.Debug().Send()

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.Where("display_name = ?", severityName).First(&severity)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			logger.Warn().Msg("severity not found")

			return echo.ErrNotFound
		}

		logger.Error().Err(res.Error).Msg("error loading severity")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusOK, apiServerDefinition.SeverityResponse{ //nolint:wrapcheck
		Data: severity.ToAPIResponse(),
	})
}

// UpdateSeverity handles updates of severities.
func (i *Implementation) UpdateSeverity(
	ctx echo.Context,
	severityName apiServerDefinition.SeverityNamePathParameter,
) error {
	var request apiServerDefinition.UpdateSeverityJSONRequestBody

	logger := i.logger.With().Str("handler", "UpdateSeverity").Str("name", severityName).Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding update severity request")

		return echo.ErrInternalServerError
	}

	if request == (apiServerDefinition.UpdateSeverityJSONRequestBody{}) { //nolint:exhaustruct
		logger.Warn().Msg("empty request")

		return echo.ErrBadRequest
	}

	logger.Debug().Interface("request", request).Send()

	severity, err := DbDef.SeverityFromAPI(&request)
	if err != nil {
		logger.Warn().Err(err).Msg("error parsing request")

		return echo.ErrBadRequest
	}

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.Where("display_name = ?", severityName).Updates(severity)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error updating severity")

		return echo.ErrInternalServerError
	}

	if res.RowsAffected == 0 {
		logger.Warn().Msg("severity not found")

		return echo.ErrNotFound
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}
