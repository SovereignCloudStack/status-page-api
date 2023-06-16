package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
)

// GetImpactTypes retrieves a list of all impact types.
func (i *Implementation) GetImpactTypes(ctx echo.Context) error {
	var impactTypes []*DbDef.ImpactType

	res := i.dbCon.Find(&impactTypes)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	data := make([]api.ImpactTypeResponseData, len(impactTypes))
	for impactTypeIndex, impactType := range impactTypes {
		data[impactTypeIndex].Id = impactType.ID.String()
		data[impactTypeIndex].DisplayName = impactType.DisplayName
		data[impactTypeIndex].Description = impactType.Description
	}

	response := api.ImpactTypeListResponse{
		Data: &data,
	}

	return ctx.JSON(http.StatusOK, response) //nolint:wrapcheck
}

// CreateImpactType handles creation of impact types.
func (i *Implementation) CreateImpactType(ctx echo.Context) error {
	var request api.CreateImpactTypeJSONRequestBody

	err := ctx.Bind(&request)
	if err != nil {
		i.logger.Error().Err(err).Msg("error binding create impact type request")

		return echo.ErrInternalServerError
	}

	i.logger.Debug().Interface("request", request).Msg("creating impact type")

	impactType := DbDef.ImpactType{ //nolint:exhaustruct
		DisplayName: request.DisplayName,
		Description: request.Description,
	}

	res := i.dbCon.Save(&impactType)

	err = res.Error
	if err != nil {
		i.logger.Error().Err(err).Msg("error creating impact type")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusOK, api.IdResponse{ //nolint:wrapcheck
		Id: impactType.ID.String(),
	})
}

// DeleteImpactType handles deletion of impact types.
func (i *Implementation) DeleteImpactType(ctx echo.Context, impactTypeID api.ImpactTypeIdPathParameter) error {
	i.logger.Debug().Str("id", impactTypeID).Msg("deleting impact type")
	res := i.dbCon.Where("id = ?", impactTypeID).Delete(&DbDef.ImpactType{}) //nolint: exhaustruct

	err := res.Error
	if err != nil {
		i.logger.Error().Err(err).Msg("error deleting impact type")

		return echo.ErrInternalServerError
	}

	if res.RowsAffected == 0 {
		return echo.ErrNotFound
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}

// GetImpactType retrieves a specific impact type by ID.
func (i *Implementation) GetImpactType(ctx echo.Context, impactTypeID api.ImpactTypeIdPathParameter) error {
	var impactType DbDef.ImpactType

	res := i.dbCon.Where("id = ?", impactTypeID).First(&impactType)

	err := res.Error
	if err != nil {
		i.logger.Error().Err(err).Msg("error retrieving impact type")

		return echo.ErrInternalServerError
	}

	response := api.ImpactTypeResponse{
		Data: &api.ImpactTypeResponseData{
			Id:          impactType.ID.String(),
			DisplayName: impactType.DisplayName,
			Description: impactType.Description,
		},
	}

	return ctx.JSON(http.StatusOK, response) //nolint:wrapcheck
}

// UpdateImpactType handles updates of impact types.
func (i *Implementation) UpdateImpactType(ctx echo.Context, impactTypeID api.ImpactTypeIdPathParameter) error {
	var (
		request    api.UpdateImpactTypeJSONRequestBody
		impactType DbDef.ImpactType
	)

	err := ctx.Bind(&request)
	if err != nil {
		i.logger.Error().Err(err).Msg("error binding update impact type request")

		return echo.ErrInternalServerError
	}

	res := i.dbCon.Where("id = ?", impactTypeID).First(&impactType)

	err = res.Error
	if err != nil {
		i.logger.Error().Err(err).Msg("error receiving impact type for update")

		return echo.ErrInternalServerError
	}

	i.logger.Debug().
		Interface("src", request).
		Interface("dst", impactType).
		Str("id", impactTypeID).
		Msg("updating impact type")

	impactType.Update(&request)

	res = i.dbCon.Save(&impactType)

	err = res.Error
	if err != nil {
		i.logger.Error().Err(err).Msg("error saving impact type after update")

		return echo.ErrInternalServerError
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}
