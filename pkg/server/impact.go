package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
)

// GetImpacttypes retrieves a list of all impact types.
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

func (i *Implementation) CreateImpactType(_ echo.Context) error {
	return nil
}

func (i *Implementation) DeleteImpactType(_ echo.Context, _ api.ImpactTypeIdPathParameter) error {
	return nil
}

func (i *Implementation) GetImpactType(_ echo.Context, _ api.ImpactTypeIdPathParameter) error {
	return nil
}

func (i *Implementation) UpdateImpactType(_ echo.Context, _ api.ImpactTypeIdPathParameter) error {
	return nil
}
