package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
)

// GetPhaseList retrieves a list of all phases.
func (i *Implementation) GetPhaseList(ctx echo.Context, params api.GetPhaseListParams) error {
	var (
		generation int
		err        error
	)

	// TODO: catch incorrect generations - 404 or 400

	if params.Generation == nil {
		generation, err = i.getCurrentPhaseGeneration()
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
		data[phaseIndex] = phase.Name
	}

	response := api.PhaseListResponse{
		Data: &api.PhaseListResponseData{
			Generation: generation,
			Phases:     data,
		},
	}

	return ctx.JSON(http.StatusOK, response) //nolint:wrapcheck
}

// CreatePhaseList handles creation of phase lists.
func (i *Implementation) CreatePhaseList(ctx echo.Context) error {
	var request api.CreatePhaseListJSONRequestBody

	err := ctx.Bind(&request)
	if err != nil {
		i.logger.Error().Err(err).Msg("error binding create phase list request")

		return echo.ErrInternalServerError
	}

	generation, err := i.getCurrentPhaseGeneration()
	if err != nil {
		i.logger.Error().Err(err).Msg("error getting current phase generation")
	}

	generation++

	i.logger.Debug().Interface("request", request).Int("generation", generation).Msg("creating phase list")

	phases := make([]*DbDef.Phase, len(request.Phases))
	for order, name := range request.Phases {
		phases[order] = &DbDef.Phase{
			Generation: generation,
			Order:      order,
			Name:       name,
		}
	}

	res := i.dbCon.Create(phases)
	if res.Error != nil {
		i.logger.Error().Err(res.Error).Msg("error creating phase list")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusCreated, api.GenerationResponse{ //nolint:wrapcheck
		Generation: generation,
	})
}

func (i *Implementation) getCurrentPhaseGeneration() (int, error) {
	var generation int
	res := i.dbCon.Model(&DbDef.Phase{}).Select("MAX(generation)").Find(&generation) //nolint:exhaustruct

	return generation, res.Error
}
