package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm/clause"
)

// GetIncidents retrieves a list of all active incidents between a start and end.
func (i *Implementation) GetIncidents(ctx echo.Context, params api.GetIncidentsParams) error {
	var incidents []*DbDef.Incident

	i.logger.Info().Time("start", params.Start).Time("end", params.End).Msg("retrieving incidents")

	res := i.dbCon.Preload("Affects.Component").Preload(clause.Associations).
		Where(i.dbCon.Not(i.dbCon.Where("began_at < ?", params.Start).Where("ended_at < ?", params.Start))).
		Where(i.dbCon.Not(i.dbCon.Where("began_at > ?", params.End).Where("ended_at > ?", params.End))).
		Or(i.dbCon.Where("ended_at IS NULL").Where("began_at <= ?", params.End)).
		Find(&incidents)

	if res.Error != nil {
		i.logger.Error().Err(res.Error).Msg("error loading incidents")

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	i.logger.Debug().Interface("incidents", incidents).Msg("raw database incidents")

	data := make([]api.IncidentResponseData, len(incidents))
	for incidentIndex, incident := range incidents {
		data[incidentIndex].Id = incident.ID.String()
		data[incidentIndex].DisplayName = incident.DisplayName
		data[incidentIndex].Description = incident.Description
		data[incidentIndex].BeganAt = incident.BeganAt
		data[incidentIndex].EndedAt = incident.EndedAt
		data[incidentIndex].Phase = &api.PhaseReference{
			Generation: *incident.Phase.Generation,
			Order:      *incident.Phase.Order,
		}
		data[incidentIndex].Affects = incident.GetImpactComponentList()
		data[incidentIndex].Updates = incident.GetIncidentUpdates()
	}

	response := api.IncidentListResponse{
		Data: &data,
	}

	return ctx.JSON(http.StatusOK, response) //nolint:wrapcheck
}

// CreateIncident handles creation of incidents.
func (i *Implementation) CreateIncident(ctx echo.Context) error {
	var request api.CreateIncidentJSONRequestBody

	err := ctx.Bind(&request)
	if err != nil {
		i.logger.Error().Err(err).Msg("error binding create incident request")

		return echo.ErrInternalServerError
	}

	i.logger.Debug().Interface("request", request).Msg("creating incident")

	incident, err := DbDef.IncidentFromAPI(&request)
	if err != nil {
		i.logger.Error().Err(err).Msg("error parsing incident creation request")
	}

	res := i.dbCon.Create(incident)
	if res.Error != nil {
		i.logger.Error().Err(res.Error).Msg("error creating incident")

		return echo.ErrInternalServerError
	}

	return ctx.JSON(http.StatusCreated, api.IdResponse{ //nolint:wrapcheck
		Id: incident.ID.String(),
	})
}

// DeleteIncident handles deletion of incidents.
func (i *Implementation) DeleteIncident(ctx echo.Context, incidentID api.IncidentIdPathParameter) error {
	i.logger.Debug().Str("id", incidentID).Msg("deleting incident")

	res := i.dbCon.Where("id = ?", incidentID).Delete(&DbDef.Incident{}) //nolint: exhaustruct
	if res.Error != nil {
		i.logger.Error().Err(res.Error).Msg("error deleting incident")

		return echo.ErrInternalServerError
	}

	if res.RowsAffected == 0 {
		return echo.ErrNotFound
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}

// GetIncident retrieves a specific incident by ID.
func (i *Implementation) GetIncident(ctx echo.Context, incidentID string) error {
	var incident DbDef.Incident

	res := i.dbCon.Preload("Affects.Component").Preload(clause.Associations).Where("id = ?", incidentID).First(&incident)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	response := api.IncidentResponse{
		Data: &api.IncidentResponseData{
			Id:          incident.ID.String(),
			DisplayName: incident.DisplayName,
			Description: incident.Description,
			BeganAt:     incident.BeganAt,
			EndedAt:     incident.EndedAt,
			Phase: &api.PhaseReference{
				Generation: *incident.Phase.Generation,
				Order:      *incident.Phase.Order,
			},
			Affects: incident.GetImpactComponentList(),
			Updates: incident.GetIncidentUpdates(),
		},
	}

	return ctx.JSON(http.StatusOK, response) //nolint:wrapcheck
}

// UpdateIncident handles updates of incidents.
func (i *Implementation) UpdateIncident(ctx echo.Context, incidentID api.IncidentIdPathParameter) error {
	var request api.UpdateIncidentJSONRequestBody

	err := ctx.Bind(&request)
	if err != nil {
		i.logger.Error().Err(err).Msg("error binding update incident request")
	}

	i.logger.Debug().Interface("request", request).Str("id", incidentID).Msg("updating incident")

	incident, err := DbDef.IncidentFromAPI(&request)
	if err != nil {
		i.logger.Error().Err(err).Msg("error parsing incident update request")

		return echo.ErrInternalServerError
	}

	incidentUUID, err := uuid.Parse(incidentID)
	if err != nil {
		i.logger.Error().Err(err).Msg("error prasing incident id for update")

		return echo.ErrInternalServerError
	}

	incident.ID = &incidentUUID

	res := i.dbCon.Updates(&incident)
	if res.Error != nil {
		i.logger.Error().Err(err).Msg("error saving incident after update")

		return echo.ErrInternalServerError
	}

	return ctx.NoContent(http.StatusNoContent) //nolint:wrapcheck
}

func (i *Implementation) GetIncidentUpdates(_ echo.Context, _ api.IncidentIdPathParameter) error {
	return nil
}

func (i *Implementation) CreateIncidentUpdate(_ echo.Context, _ api.IncidentIdPathParameter) error {
	return nil
}

func (i *Implementation) DeleteIncidentUpdate(
	_ echo.Context,
	_ api.IncidentIdPathParameter,
	_ api.IncidentUpdateOrderPathParameter,
) error {
	return nil
}

func (i *Implementation) GetIncidentUpdate(
	_ echo.Context,
	_ api.IncidentIdPathParameter,
	_ api.IncidentUpdateOrderPathParameter,
) error {
	return nil
}

func (i *Implementation) UpdateIncidentUpdate(
	_ echo.Context,
	_ api.IncidentIdPathParameter,
	_ api.IncidentUpdateOrderPathParameter,
) error {
	return nil
}
