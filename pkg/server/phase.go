package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
)

// GetPhaseList retrieves a list of all phases.
func (i *Implementation) GetPhaseList(ctx echo.Context, params api.GetPhaseListParams) error {
	generation, err := i.getCurrentPhaseGeneration()
	if err != nil {
		i.logger.Error().Err(err).Msg("error getting current phase generation")

		return echo.ErrInternalServerError
	}

	if params.Generation != nil {
		if *params.Generation > generation || *params.Generation < 1 {
			return echo.ErrNotFound
		}

		generation = *params.Generation
	}

	i.logger.Debug().Int("generation", generation).Msg("loading phase list")

	var phases []*DbDef.Phase

	res := i.dbCon.Where("generation = ?", generation).Order("\"order\" asc").Find(&phases)
	if res.Error != nil {
		i.logger.Error().Err(res.Error).Msg("error loading phase list")

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	data := make([]api.Phase, len(phases))
	for phaseIndex, phase := range phases {
		data[phaseIndex] = *phase.Name
	}

	return ctx.JSON(http.StatusOK, api.PhaseListResponse{ //nolint:wrapcheck
		Data: &api.PhaseListResponseData{
			Generation: generation,
			Phases:     data,
		},
	})
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

	for phaseIndex, phase := range request.Phases {
		order := phaseIndex
		name := phase
		phases[phaseIndex] = &DbDef.Phase{
			Generation: &generation,
			Order:      &order,
			Name:       &name,
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
	res := i.dbCon.Model(&DbDef.Phase{}).Select("COALESCE(MAX(generation), 0)").Find(&generation) //nolint:exhaustruct

	return generation, res.Error
}
