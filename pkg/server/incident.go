package server

import (
	"errors"
	"fmt"
	"net/http"
	"slices"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	apiServerDefinition "github.com/SovereignCloudStack/status-page-openapi/pkg/api/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetIncidents retrieves a list of all active incidents between a start and end.
func (i *Implementation) GetIncidents(ctx echo.Context, params apiServerDefinition.GetIncidentsParams) error {
	var incidents []*DbDef.Incident

	logger := i.logger.With().Str("handler", "GetIncidents").Logger()
	logger.Debug().Time("start", params.Start).Time("end", params.End).Send()

	if params.Start.IsZero() || params.End.IsZero() {
		logger.Warn().Msg("missing time parameter")

		return echo.ErrBadRequest
	}

	if params.End.Before(params.Start) {
		logger.Warn().Msg("end paramater before start parameter")

		return echo.ErrBadRequest
	}

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.
		Preload("Affects.Component").
		Preload(clause.Associations).
		Where(dbSession.
			Not(dbSession.
				Where("began_at < ?", params.Start).
				Where("ended_at < ?", params.Start))).
		Where(dbSession.
			Not(dbSession.
				Where("began_at > ?", params.End).
				Where("ended_at > ?", params.End))).
		Or(dbSession.
			Where("ended_at IS NULL").
			Where("began_at <= ?", params.End)).
		Find(&incidents)

	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error loading incidents")

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	data := make([]apiServerDefinition.IncidentResponseData, len(incidents))
	for incidentIndex, incident := range incidents {
		data[incidentIndex] = incident.ToAPIResponse()
	}

	return ctx.JSON(http.StatusOK, apiServerDefinition.IncidentListResponse{ //nolint:wrapcheck
		Data: data,
	})
}

// CreateIncident handles creation of incidents.
func (i *Implementation) CreateIncident(ctx echo.Context) error {
	var request apiServerDefinition.CreateIncidentJSONRequestBody

	logger := i.logger.With().Str("handler", "CreateIncident").Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	if request == (apiServerDefinition.CreateIncidentJSONRequestBody{}) { //nolint: exhaustruct
		logger.Warn().Msg("empty request")

		return echo.ErrBadRequest
	}

	logger.Debug().Interface("request", request).Send()

	incident, err := DbDef.IncidentFromAPI(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing request")

		return echo.ErrBadRequest
	}

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.Create(incident)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error creating incident")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusCreated, apiServerDefinition.IdResponse{ //nolint:wrapcheck
		Id: incident.ID,
	})
}

// DeleteIncident handles deletion of incidents.
func (i *Implementation) DeleteIncident(
	ctx echo.Context,
	incidentID apiServerDefinition.IncidentIdPathParameter,
) error {
	logger := i.logger.With().Str("handler", "DeleteIncident").Interface("id", incidentID).Logger()
	logger.Debug().Send()

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.Where("id = ?", incidentID).Delete(&DbDef.Incident{}) //nolint: exhaustruct
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
func (i *Implementation) GetIncident(ctx echo.Context, incidentID apiServerDefinition.IncidentIdPathParameter) error {
	var incident DbDef.Incident

	logger := i.logger.With().Str("handler", "GetIncident").Interface("id", incidentID).Logger()
	logger.Debug().Send()

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.
		Preload("Affects.Component").
		Preload(clause.Associations).
		Where("id = ?", incidentID).
		First(&incident)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			logger.Warn().Msg("incident not found")

			return echo.ErrNotFound
		}

		logger.Error().Err(res.Error).Msg("error loading incident")

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, apiServerDefinition.IncidentResponse{ //nolint:wrapcheck
		Data: incident.ToAPIResponse(),
	})
}

func prepareAffects(oldAffects, newAffects *[]DbDef.Impact, incidentID uuid.UUID, dbTx *gorm.DB) error {
	// Check if any impacts need deletion.
	if oldAffects == nil || len(*oldAffects) == 0 {
		return nil
	}

	// Check if impacts are modified at all.
	if newAffects == nil {
		return nil
	}

	// Check if any impacts are left.
	if len(*newAffects) == 0 {
		// Delete all impacts of this incident.
		err := dbTx.Delete(oldAffects).Where("incident_id = ?", incidentID).Error
		if err != nil {
			return fmt.Errorf("error deleting all incident impacts: %w", err)
		}

		return nil
	}

	// Collect all impacts that are expected and give them the incident id.
	newImpacts := make([]DbDef.Impact, len(*newAffects))

	for incidentImpactIndex := range *newAffects {
		newImpacts[incidentImpactIndex].IncidentID = &incidentID
	}

	var impactsToBeDeleted []DbDef.Impact

	for _, oldImpact := range *oldAffects {
		if !slices.ContainsFunc(newImpacts, func(newImpact DbDef.Impact) bool {
			return newImpact.ComponentID == oldImpact.ComponentID &&
				newImpact.ImpactTypeID == oldImpact.ImpactTypeID &&
				newImpact.IncidentID == oldImpact.IncidentID
		}) {
			impactsToBeDeleted = append(impactsToBeDeleted, oldImpact)
		}
	}

	err := dbTx.Delete(&impactsToBeDeleted).Error
	if err != nil {
		return fmt.Errorf("error deleting not needed impacts: %w", err)
	}

	return nil
}

// UpdateIncident handles updates of incidents.
func (i *Implementation) UpdateIncident( //nolint: funlen
	ctx echo.Context,
	incidentID apiServerDefinition.IncidentIdPathParameter,
) error {
	var request apiServerDefinition.UpdateIncidentJSONRequestBody

	logger := i.logger.With().
		Str("handler", "UpdateIncident").
		Interface("id", incidentID).
		Logger()

	// Check and validate request.
	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	if request == (apiServerDefinition.UpdateIncidentJSONRequestBody{}) { //nolint:exhaustruct
		logger.Warn().Msg("empty request")

		return echo.ErrBadRequest
	}

	logger.Debug().Interface("request", request).Send()

	// Prepare new incident.
	newIncident, err := DbDef.IncidentFromAPI(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing request")

		return echo.ErrBadRequest
	}

	newIncident.ID = incidentID

	// DB connection.
	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	err = dbSession.Transaction(func(dbTx *gorm.DB) error {
		// Check if incident exists.
		var (
			currentIncident DbDef.Incident
			transactionErr  error
		)

		currentIncident.ID = newIncident.ID

		transactionErr = dbTx.Preload("Affects").First(&currentIncident).Error
		if transactionErr != nil {
			if errors.Is(transactionErr, gorm.ErrRecordNotFound) {
				logger.Warn().Msg("incident not found")

				return echo.ErrNotFound
			}

			logger.Error().Err(transactionErr).Msg("error loading current incident from database")

			return echo.ErrInternalServerError
		}

		// Prepare for update.
		logger.Debug().Interface("currentIncident", currentIncident).Interface("newIncident", newIncident).Send()

		transactionErr = prepareAffects(currentIncident.Affects, newIncident.Affects, newIncident.ID, dbTx)
		if transactionErr != nil {
			logger.Error().Err(transactionErr).Msg("error updating affected components")

			return echo.ErrInternalServerError
		}

		transactionErr = dbTx.Updates(&newIncident).Error
		if transactionErr != nil {
			logger.Error().Err(transactionErr).Msg("error updating incident")

			return echo.ErrInternalServerError
		}

		return nil
	})
	if err != nil {
		logger.Error().Err(err).Msg("error in database transaction")

		// Don't wrap the echo errors.
		return err //nolint:wrapcheck
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}

// GetIncidentUpdates retrieves a list of all updates for one incident.
func (i *Implementation) GetIncidentUpdates(
	ctx echo.Context,
	incidentID apiServerDefinition.IncidentIdPathParameter,
) error {
	var incidentUpdates []DbDef.IncidentUpdate

	logger := i.logger.With().Str("handler", "GetIncidentUpdates").Interface("id", incidentID).Logger()
	logger.Debug().Send()

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.Where("incident_id = ?", incidentID).Find(&incidentUpdates)
	if res.Error != nil {
		logger.Error().Err(res.Error).Msg("error loading incident updates")

		return echo.ErrInternalServerError
	}

	data := make([]apiServerDefinition.IncidentUpdateResponseData, len(incidentUpdates))
	for incidentUpdateIndex, incidentUpdate := range incidentUpdates {
		data[incidentUpdateIndex] = incidentUpdate.ToAPIResponse()
	}

	return ctx.JSON(http.StatusOK, apiServerDefinition.IncidentUpdateListResponse{ //nolint:wrapcheck
		Data: data,
	})
}

// CreateIncidentUpdate handles updates to an update for one incident.
func (i *Implementation) CreateIncidentUpdate(
	ctx echo.Context,
	incidentID apiServerDefinition.IncidentIdPathParameter,
) error {
	var (
		request apiServerDefinition.CreateIncidentUpdateJSONRequestBody
		order   int
	)

	logger := i.logger.With().Str("handler", "CreateIncidentUpdate").Interface("id", incidentID).Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	if request == (apiServerDefinition.CreateIncidentUpdateJSONRequestBody{}) { //nolint: exhaustruct
		logger.Warn().Msg("empty request")

		return echo.ErrBadRequest
	}

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	err = dbSession.Transaction(func(dbTx *gorm.DB) error {
		var (
			incidentUpdate *DbDef.IncidentUpdate
			transactionErr error
		)

		order, transactionErr = DbDef.GetHighestIncidentUpdateOrder(dbTx, incidentID)
		if transactionErr != nil {
			return fmt.Errorf("error getting current highest order of incident: %w", transactionErr)
		}

		order++

		logger.Debug().Interface("request", request).Int("order", order).Send()

		incidentUpdate, transactionErr = DbDef.IncidentUpdateFromAPI(&request, incidentID, order)
		if transactionErr != nil {
			return fmt.Errorf("error parsing request: %w", transactionErr)
		}

		res := dbTx.Create(&incidentUpdate)
		if res.Error != nil {
			return fmt.Errorf("error creating incident update: %w", res.Error)
		}

		return nil
	})
	if err != nil {
		logger.Error().Err(err).Msg("error in transaction")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusCreated, apiServerDefinition.OrderResponse{ //nolint:wrapcheck
		Order: order,
	})
}

// DeleteIncidentUpdate handles deletion of an update for one incident.
func (i *Implementation) DeleteIncidentUpdate(
	ctx echo.Context,
	incidentID apiServerDefinition.IncidentIdPathParameter,
	incidentUpdateOrder apiServerDefinition.IncidentUpdateOrderPathParameter,
) error {
	logger := i.logger.With().
		Str("handler", "DeleteIncidentUpdate").
		Interface("id", incidentID).
		Int("order", incidentUpdateOrder).
		Logger()
	logger.Debug().Send()

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.
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
	incidentID apiServerDefinition.IncidentIdPathParameter,
	incidentUpdateOrder apiServerDefinition.IncidentUpdateOrderPathParameter,
) error {
	var incidentUpdate DbDef.IncidentUpdate

	logger := i.logger.With().
		Str("handler", "GetIncidentUpdate").
		Interface("id", incidentID).
		Int("order", incidentUpdateOrder).
		Logger()
	logger.Debug().Send()

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.
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

	return ctx.JSON(http.StatusOK, apiServerDefinition.IncidentUpdateResponse{ //nolint:wrapcheck
		Data: incidentUpdate.ToAPIResponse(),
	})
}

// UpdateIncidentUpdate handles updates of updates for one incident.
func (i *Implementation) UpdateIncidentUpdate(
	ctx echo.Context,
	incidentID apiServerDefinition.IncidentIdPathParameter,
	incidentUpdateOrder apiServerDefinition.IncidentUpdateOrderPathParameter,
) error {
	var request apiServerDefinition.UpdateIncidentUpdateJSONRequestBody

	logger := i.logger.With().
		Str("handler", "UpdateIncidentUpdate").
		Interface("id", incidentID).
		Int("order", incidentUpdateOrder).
		Logger()

	err := ctx.Bind(&request)
	if err != nil {
		logger.Error().Err(err).Msg("error binding request")

		return echo.ErrInternalServerError
	}

	if request == (apiServerDefinition.CreateIncidentUpdateJSONRequestBody{}) { //nolint: exhaustruct
		logger.Warn().Msg("empty request")

		return echo.ErrBadRequest
	}

	logger.Debug().
		Interface("request", request).
		Send()

	incidentUpdate, err := DbDef.IncidentUpdateFromAPI(&request, incidentID, incidentUpdateOrder)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing request")

		return echo.ErrInternalServerError
	}

	dbSession := i.dbCon.WithContext(ctx.Request().Context())

	res := dbSession.Updates(&incidentUpdate)
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
