package server

import (
	"errors"
	"fmt"
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

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	err := dbSession.Transaction(func(dbTx *gorm.DB) error {
		var transactionErr error

		generation, transactionErr = DbDef.GetCurrentPhaseGeneration(dbTx)
		if transactionErr != nil {
			return fmt.Errorf("error getting current generation: %w", transactionErr)
		}

		if params.Generation != nil {
			if *params.Generation < 1 {
				return fmt.Errorf("%w: %d", ErrInvalidPhaseGeneration, *params.Generation)
			}

			if *params.Generation > generation {
				return fmt.Errorf("%w: %d", ErrPhaseGenerationNotFound, *params.Generation)
			}

			generation = *params.Generation
		}

		logger.Debug().Int("generation", generation).Send()

		var phases []*DbDef.Phase

		res := dbTx.Where("generation = ?", generation).Order("\"order\" asc").Find(&phases)
		if res.Error != nil {
			return fmt.Errorf("error loading phase list: %w", res.Error)
		}

		data = make([]api.Phase, len(phases))
		for phaseIndex, phase := range phases {
			data[phaseIndex] = *phase.Name
		}

		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidPhaseGeneration):
			logger.Warn().Err(err).Send()

			return echo.ErrBadRequest
		case errors.Is(err, ErrPhaseGenerationNotFound):
			logger.Warn().Err(err).Send()

			return echo.ErrNotFound
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

	if len(request.Phases) == 0 {
		logger.Warn().Msg("empty request")

		return echo.ErrBadRequest
	}

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	err = dbSession.Transaction(func(dbTx *gorm.DB) error {
		var transactionErr error

		generation, transactionErr = DbDef.GetCurrentPhaseGeneration(dbTx)
		if transactionErr != nil {
			return fmt.Errorf("error getting current generation: %w", transactionErr)
		}

		generation++

		logger.Debug().Interface("request", request).Int("generation", generation).Send()

		phases := make([]DbDef.Phase, len(request.Phases))

		for phaseIndex, phase := range request.Phases {
			order := phaseIndex //nolint:copyloopvar
			name := phase       //nolint:copyloopvar

			phases[phaseIndex] = DbDef.Phase{
				Generation: &generation,
				Order:      &order,
				Name:       &name,
			}
		}

		res := dbTx.Create(phases)
		if res.Error != nil {
			return fmt.Errorf("error creating phase list: %w", res.Error)
		}

		return nil
	})
	if err != nil {
		logger.Error().Err(err).Msg("error in transaction")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusCreated, api.GenerationResponse{ //nolint:wrapcheck
		Generation: generation,
	})
}
