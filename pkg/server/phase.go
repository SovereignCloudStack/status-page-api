package server

import (
	"errors"
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// GetPhaseList retrieves a list of all phases.
func (i *Implementation) GetPhaseList(ctx echo.Context, params api.GetPhaseListParams) error { //nolint:funlen
	var (
		generation int
		data       []api.Phase
	)

	logger := i.logger.With().Str("handler", "GetPhaseList").Logger()

	err := i.dbCon.Transaction(func(dbTx *gorm.DB) error {
		var transactionErr error
		generation, transactionErr = DbDef.GetCurrentPhaseGeneration(dbTx)
		if transactionErr != nil {
			logger.Error().Err(transactionErr).Msg("error getting current generation")

			return echo.ErrInternalServerError
		}

		if params.Generation != nil {
			if *params.Generation > generation || *params.Generation < 1 {
				logger.Warn().Msg("phase generation not found")

				return echo.ErrNotFound
			}

			generation = *params.Generation
		}

		logger.Debug().Int("generation", generation).Send()

		var phases []*DbDef.Phase

		res := dbTx.Where("generation = ?", generation).Order("\"order\" asc").Find(&phases)
		if res.Error != nil {
			logger.Error().Err(res.Error).Msg("error loading phase list")

			return echo.ErrInternalServerError
		}

		data = make([]api.Phase, len(phases))
		for phaseIndex, phase := range phases {
			data[phaseIndex] = *phase.Name
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, &echo.HTTPError{}) { //nolint:exhaustruct
			// Echo errors are already defined and logged
			return err //nolint:wrapcheck
		}

		logger.Error().Err(err).Msg("error in database transaction")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusOK, api.PhaseListResponse{ //nolint:wrapcheck
		Data: api.PhaseListResponseData{
			Generation: generation,
			Phases:     data,
		},
	})
}

// CreatePhaseList handles creation of phase lists.
func (i *Implementation) CreatePhaseList(ctx echo.Context) error { //nolint:funlen
	var (
		generation int
		request    api.CreatePhaseListJSONRequestBody
	)

	logger := i.logger.With().Str("handler", "CreatePhaseList").Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	err = i.dbCon.Transaction(func(dbTx *gorm.DB) error {
		var transactionErr error

		generation, transactionErr = DbDef.GetCurrentPhaseGeneration(dbTx)
		if transactionErr != nil {
			logger.Error().Err(transactionErr).Msg("error getting current phase generation")

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

		res := dbTx.Create(phases)
		if res.Error != nil {
			logger.Error().Err(res.Error).Msg("error creating phase list")

			return echo.ErrInternalServerError
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, &echo.HTTPError{}) { //nolint:exhaustruct
			// Echo errors are already defined and logged
			return err //nolint:wrapcheck
		}

		logger.Error().Err(err).Msg("error in transaction")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusCreated, api.GenerationResponse{ //nolint:wrapcheck
		Generation: generation,
	})
}
