package server

import (
	"errors"
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// GetImpactTypes retrieves a list of all impact types.
func (i *Implementation) GetImpactTypes(ctx echo.Context) error {
	var impactTypes []*DbDef.ImpactType

	logger := i.logger.With().Str("handler", "GetImpactTypes").Logger()
	logger.Debug().Send()

	res := i.dbCon.Find(&impactTypes)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error loading impact types")

		return echo.ErrInternalServerError
	}

	data := make([]api.ImpactTypeResponseData, len(impactTypes))
	for impactTypeIndex, impactType := range impactTypes {
		data[impactTypeIndex] = impactType.ToAPIResponse()
	}

	return ctx.JSON(http.StatusOK, api.ImpactTypeListResponse{ //nolint:wrapcheck
		Data: data,
	})
}

// CreateImpactType handles creation of impact types.
func (i *Implementation) CreateImpactType(ctx echo.Context) error { //nolint:dupl
	var request api.CreateImpactTypeJSONRequestBody

	logger := i.logger.With().Str("handler", "CreateImpactType").Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	logger.Debug().Interface("request", request).Send()

	impactType, err := DbDef.ImpactTypeFromAPI(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing request")

		return echo.ErrInternalServerError
	}

	res := i.dbCon.Create(&impactType)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error creating impact type")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusOK, api.IdResponse{ //nolint:wrapcheck
		Id: impactType.ID.String(),
	})
}

// DeleteImpactType handles deletion of impact types.
func (i *Implementation) DeleteImpactType(ctx echo.Context, impactTypeID api.ImpactTypeIdPathParameter) error {
	logger := i.logger.With().Str("handler", "DeleteImpactType").Str("id", impactTypeID).Logger()
	logger.Debug().Send()

	res := i.dbCon.Where("id = ?", impactTypeID).Delete(&DbDef.ImpactType{}) //nolint: exhaustruct
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error deleting impact type")

		return echo.ErrInternalServerError
	}

	if res.RowsAffected == 0 {
		logger.Warn().Msg("impact type not found")

		return echo.ErrNotFound
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}

// GetImpactType retrieves a specific impact type by ID.
func (i *Implementation) GetImpactType(ctx echo.Context, impactTypeID api.ImpactTypeIdPathParameter) error {
	var impactType DbDef.ImpactType

	logger := i.logger.With().Str("handler", "GetImpactType").Str("id", impactTypeID).Logger()
	logger.Debug().Send()

	res := i.dbCon.Where("id = ?", impactTypeID).First(&impactType)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			logger.Warn().Msg("impact type not found")

			return echo.ErrNotFound
		}

		logger.Error().Err(res.Error).Msg("error loading impact type")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusOK, api.ImpactTypeResponse{ //nolint:wrapcheck
		Data: impactType.ToAPIResponse(),
	})
}

// UpdateImpactType handles updates of impact types.
func (i *Implementation) UpdateImpactType(ctx echo.Context, impactTypeID api.ImpactTypeIdPathParameter) error {
	var request api.UpdateImpactTypeJSONRequestBody

	logger := i.logger.With().Str("handler", "UpdateImpactType").Str("id", impactTypeID).Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	logger.Debug().Interface("request", request).Send()

	impactType, err := DbDef.ImpactTypeFromAPI(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing request")
	}

	impactTypeUUID, err := uuid.Parse(impactTypeID)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing id")
	}

	impactType.ID = &impactTypeUUID

	res := i.dbCon.Updates(&impactType)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error updating impact type")

		return echo.ErrInternalServerError
	}

	if res.RowsAffected == 0 {
		logger.Warn().Msg("impact type not found")

		return echo.ErrNotFound
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}
