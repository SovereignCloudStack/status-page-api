package server

import (
	"errors"
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetIncidents retrieves a list of all active incidents between a start and end.
func (i *Implementation) GetIncidents(ctx echo.Context, params api.GetIncidentsParams) error {
	var incidents []*DbDef.Incident

	logger := i.logger.With().Str("handler", "GetIncidents").Logger()
	logger.Debug().Time("start", params.Start).Time("end", params.End).Send()

	if params.Start.IsZero() || params.End.IsZero() {
		return echo.ErrBadRequest
	}

	if params.End.Before(params.Start) {
		return echo.ErrBadRequest
	}

	res := i.dbCon.
		Preload("Affects.Component").
		Preload(clause.Associations).
		Where(i.dbCon.
			Not(i.dbCon.
				Where("began_at < ?", params.Start).
				Where("ended_at < ?", params.Start))).
		Where(i.dbCon.
			Not(i.dbCon.
				Where("began_at > ?", params.End).
				Where("ended_at > ?", params.End))).
		Or(i.dbCon.
			Where("ended_at IS NULL").
			Where("began_at <= ?", params.End)).
		Find(&incidents)

	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error loading incidents")

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	data := make([]api.IncidentResponseData, len(incidents))
	for incidentIndex, incident := range incidents {
		data[incidentIndex] = incident.ToAPIResponse()
	}

	return ctx.JSON(http.StatusOK, api.IncidentListResponse{ //nolint:wrapcheck
		Data: data,
	})
}

// CreateIncident handles creation of incidents.
func (i *Implementation) CreateIncident(ctx echo.Context) error {
	var request api.CreateIncidentJSONRequestBody

	logger := i.logger.With().Str("handler", "CreateIncident").Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	logger.Debug().Interface("request", request).Send()

	incident, err := DbDef.IncidentFromAPI(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing request")

		return echo.ErrInternalServerError
	}

	res := i.dbCon.Create(incident)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error creating incident")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusCreated, api.IdResponse{ //nolint:wrapcheck
		Id: incident.ID.String(),
	})
}

// DeleteIncident handles deletion of incidents.
func (i *Implementation) DeleteIncident(ctx echo.Context, incidentID api.IncidentIdPathParameter) error {
	logger := i.logger.With().Str("handler", "DeleteIncident").Str("id", incidentID).Logger()
	logger.Debug().Send()

	res := i.dbCon.Where("id = ?", incidentID).Delete(&DbDef.Incident{}) //nolint: exhaustruct
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error deleting incident")

		return echo.ErrInternalServerError
	}

	if res.RowsAffected == 0 {
		logger.Warn().Msg("incident not found")

		return echo.ErrNotFound
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}

// GetIncident retrieves a specific incident by ID.
func (i *Implementation) GetIncident(ctx echo.Context, incidentID string) error {
	var incident DbDef.Incident

	logger := i.logger.With().Str("handler", "GetIncident").Str("id", incidentID).Logger()
	logger.Debug().Send()

	res := i.dbCon.Preload("Affects.Component").Preload(clause.Associations).Where("id = ?", incidentID).First(&incident)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			logger.Warn().Msg("incident not found")

			return echo.ErrNotFound
		}

		logger.Error().Err(res.Error).Msg("error loading incident")

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, api.IncidentResponse{ //nolint:wrapcheck
		Data: incident.ToAPIResponse(),
	})
}

// UpdateIncident handles updates of incidents.
func (i *Implementation) UpdateIncident(ctx echo.Context, incidentID api.IncidentIdPathParameter) error {
	var request api.UpdateIncidentJSONRequestBody

	logger := i.logger.With().Str("handler", "UpdateIncident").Str("id", incidentID).Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	logger.Debug().Interface("request", request).Send()

	incident, err := DbDef.IncidentFromAPI(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing request")

		return echo.ErrInternalServerError
	}

	incidentUUID, err := uuid.Parse(incidentID)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing id")

		return echo.ErrInternalServerError
	}

	incident.ID = &incidentUUID

	res := i.dbCon.Updates(&incident)
	if res.Error != nil {
		logger.Error().Err(err).Msg("error updating incident")

		return echo.ErrInternalServerError
	}

	if res.RowsAffected == 0 {
		logger.Warn().Msg("incident not found")

		return echo.ErrNotFound
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}

// GetIncidentUpdates retrieves a list of all updates for one incident.
func (i *Implementation) GetIncidentUpdates(ctx echo.Context, incidentID api.IncidentIdPathParameter) error {
	var incidentUpdates []DbDef.IncidentUpdate

	logger := i.logger.With().Str("handler", "GetIncidentUpdates").Str("id", incidentID).Logger()
	logger.Debug().Send()

	res := i.dbCon.Where("incident_id = ?", incidentID).Find(&incidentUpdates)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error loading incident updates")

		return echo.ErrInternalServerError
	}

	data := make([]api.IncidentUpdateResponseData, len(incidentUpdates))
	for incidentUpdateIndex, incidentUpdate := range incidentUpdates {
		data[incidentUpdateIndex] = incidentUpdate.ToAPIResponse()
	}

	return ctx.JSON(http.StatusOK, api.IncidentUpdateListResponse{ //nolint:wrapcheck
		Data: data,
	})
}

// CreateIncidentUpdate handles updates to an update for one incident.
func (i *Implementation) CreateIncidentUpdate(ctx echo.Context, incidentID api.IncidentIdPathParameter) error { //nolint:funlen,lll
	var (
		request api.CreateIncidentUpdateJSONRequestBody
		order   int
	)

	logger := i.logger.With().Str("handler", "CreateIncidentUpdate").Str("id", incidentID).Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	err = i.dbCon.Transaction(func(dbTx *gorm.DB) error {
		var (
			incidentUpdate *DbDef.IncidentUpdate
			transactionErr error
		)

		order, transactionErr = DbDef.GetHighestIncidentUpdateOrder(dbTx, incidentID)
		if transactionErr != nil {
			logger.Error().Err(transactionErr).Msg("error getting current highest order of incident")

			return echo.ErrInternalServerError
		}

		order++

		logger.Debug().Interface("request", request).Int("order", order).Send()

		incidentUpdate, transactionErr = DbDef.IncidentUpdateFromAPI(&request, incidentID, order)
		if transactionErr != nil {
			logger.Error().Err(transactionErr).Msg("error parsing request")

			return echo.ErrInternalServerError
		}

		res := dbTx.Create(&incidentUpdate)
		if res.Error != nil {
			logger.Error().Err(res.Error).Msg("error creating incident update")

			return echo.ErrInternalServerError
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

	return ctx.JSON(http.StatusCreated, api.OrderResponse{ //nolint:wrapcheck
		Order: order,
	})
}

// DeleteIncidentUpdate handles deletion of an update for one incident.
func (i *Implementation) DeleteIncidentUpdate(
	ctx echo.Context,
	incidentID api.IncidentIdPathParameter,
	incidentUpdateOrder api.IncidentUpdateOrderPathParameter,
) error {
	logger := i.logger.With().
		Str("handler", "DeleteIncidentUpdate").
		Str("id", incidentID).
		Int("order", incidentUpdateOrder).
		Logger()
	logger.Debug().Send()

	res := i.dbCon.
		Where("incident_id = ?", incidentID).
		Where("\"order\" = ?", incidentUpdateOrder).
		Delete(&DbDef.IncidentUpdate{}) //nolint: exhaustruct
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error deleting incident update")

		return echo.ErrInternalServerError
	}

	if res.RowsAffected == 0 {
		logger.Warn().Msg("incident update not found")

		return echo.ErrNotFound
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}

// GetIncidentUpdate retrieves a specific update for one incident.
func (i *Implementation) GetIncidentUpdate(
	ctx echo.Context,
	incidentID api.IncidentIdPathParameter,
	incidentUpdateOrder api.IncidentUpdateOrderPathParameter,
) error {
	var incidentUpdate DbDef.IncidentUpdate

	logger := i.logger.With().
		Str("handler", "GetIncidentUpdate").
		Str("id", incidentID).
		Int("order", incidentUpdateOrder).
		Logger()
	logger.Debug().Send()

	res := i.dbCon.
		Where("incident_id = ?", incidentID).
		Where("\"order\" = ?", incidentUpdateOrder).
		First(&incidentUpdate)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			logger.Warn().Msg("incident update not found")

			return echo.ErrNotFound
		}

		logger.Error().Err(res.Error).Msg("error loading incident update")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusOK, api.IncidentUpdateResponse{ //nolint:wrapcheck
		Data: incidentUpdate.ToAPIResponse(),
	})
}

// UpdateIncidentUpdate handles updates of updates for one incident.
func (i *Implementation) UpdateIncidentUpdate(
	ctx echo.Context,
	incidentID api.IncidentIdPathParameter,
	incidentUpdateOrder api.IncidentUpdateOrderPathParameter,
) error {
	var request api.UpdateIncidentUpdateJSONRequestBody

	logger := i.logger.With().
		Str("handler", "UpdateIncidentUpdate").
		Str("id", incidentID).
		Int("order", incidentUpdateOrder).
		Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	logger.Debug().
		Interface("request", request).
		Send()

	incidentUpdate, err := DbDef.IncidentUpdateFromAPI(&request, incidentID, incidentUpdateOrder)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing request")

		return echo.ErrInternalServerError
	}

	res := i.dbCon.Updates(&incidentUpdate)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error updating incident update")

		return echo.ErrInternalServerError
	}

	if res.RowsAffected == 0 {
		logger.Warn().Msg("incident update not found")

		return echo.ErrNotFound
	}

	return ctx.NoContent(http.StatusNoContent) //nolint: wrapcheck
}
