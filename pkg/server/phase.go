package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
)

// GetPhases retrieves a list of all phases.
func (i *Implementation) GetPhaseList(ctx echo.Context, params api.GetPhaseListParams) error {
	var (
		generation int
		err        error
	)

	// TODO: catch incorrect generations - 404 or 400

	if params.Generation == nil {
		generation, err = GetCurrentPhaseGeneration(i.dbCon)
		if err != nil {
			return echo.ErrInternalServerError
		}
	} else {
		generation = *params.Generation
	}

	var phases []*DbDef.Phase

	res := i.dbCon.Where("generation = ?", generation).Order("\"order\" asc").Find(&phases)

	err = res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	data := make([]api.Phase, len(phases))
	for phaseIndex, phase := range phases {
		data[phaseIndex] = *phase.Name
	}

	response := api.PhaseListResponse{
		Data: &api.PhaseListResponseData{
			Generation: generation,
			Phases:     data,
		},
	}

	return ctx.JSON(http.StatusOK, response) //nolint:wrapcheck
}

func (i *Implementation) CreatePhaseList(_ echo.Context) error {
	return nil
}
