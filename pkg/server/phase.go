package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
)

// GetPhaseList retrieves a list of all phases.
func (i *Implementation) GetPhaseList(ctx echo.Context, params api.GetPhaseListParams) error {
	logger := i.logger.With().Str("handler", "GetPhaseList").Logger()

	generation, err := i.getCurrentPhaseGeneration()
	if err != nil {
		logger.Error().Err(err).Msg("error getting current generation")

		return echo.ErrInternalServerError
	}

	if params.Generation != nil {
		if *params.Generation > generation || *params.Generation < 1 {
			return echo.ErrNotFound
		}

		generation = *params.Generation
	}

	logger.Debug().Int("generation", generation).Send()

	var phases []*DbDef.Phase

	res := i.dbCon.Where("generation = ?", generation).Order("\"order\" asc").Find(&phases)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error loading phase list")

		return echo.ErrInternalServerError
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

	logger := i.logger.With().Str("handler", "CreatePhaseList").Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	generation, err := i.getCurrentPhaseGeneration()
	if err != nil {
		logger.Error().Err(err).Msg("error getting current phase generation")

		return echo.ErrInternalServerError
	}

	generation++

	logger.Debug().Interface("request", request).Int("generation", generation).Send()

	phases := make([]DbDef.Phase, len(request.Phases))

	for phaseIndex, phase := range request.Phases {
		order := phaseIndex
		name := phase

		phases[phaseIndex] = DbDef.Phase{
			Generation: &generation,
			Order:      &order,
			Name:       &name,
		}
	}

	res := i.dbCon.Create(phases)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error creating phase list")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusCreated, api.GenerationResponse{ //nolint:wrapcheck
		Generation: generation,
	})
}

func (i *Implementation) getCurrentPhaseGeneration() (int, error) {
	var generation int
	res := i.dbCon.
		Model(&DbDef.Phase{}). //nolint:exhaustruct
		Select("COALESCE(MAX(generation), 0)").
		Find(&generation)

	return generation, res.Error
}
