package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
)

func (i *Implementation) GetPhases(ctx echo.Context) error {
	var phases []*DbDef.Phase

	res := i.dbCon.Find(&phases)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	phaseList := make([]*api.IncidentPhase, len(phases))
	for _, phase := range phases {
		phaseList[phase.Order] = &phase.Slug
	}

	return ctx.JSON(http.StatusOK, phaseList)
}
