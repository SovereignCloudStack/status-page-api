package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
)

func (i *Implementation) GetImpacttypes(ctx echo.Context) error {
	var impactTypes []*DbDef.ImpactType

	res := i.dbCon.Find(&impactTypes)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	impactTypeList := make([]*api.IncidentImpactType, len(impactTypes))
	for impactTypeIndex, impactType := range impactTypes {
		impactTypeList[impactTypeIndex] = &impactType.Slug
	}

	return ctx.JSON(http.StatusOK, impactTypeList)
}
