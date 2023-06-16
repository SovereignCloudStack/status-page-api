package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm/clause"
)

// GetIncidents retrieves a list of all active incidents between a start and end.
func (i *Implementation) GetIncidents(ctx echo.Context, params api.GetIncidentsParams) error {
	var incidents []*DbDef.Incident

	res := i.dbCon.Preload(
		clause.Associations,
	).Where(
		i.dbCon.Not(i.dbCon.Where("began_at < ?", params.Start).Where("ended_at < ?", params.Start)),
	).Where(
		i.dbCon.Not(i.dbCon.Where("began_at > ?", params.End).Where("ended_at > ?", params.End)),
	).Or(
		i.dbCon.Where("ended_at IS NULL").Where("began_at <= ?", params.End),
	).Find(&incidents)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	data := make([]api.IncidentResponseData, len(incidents))
	for incidentIndex, incident := range incidents {
		data[incidentIndex].Id = incident.ID.String()
		data[incidentIndex].DisplayName = incident.DisplayName
		data[incidentIndex].Description = incident.Description
		data[incidentIndex].BeganAt = incident.BeganAt
		data[incidentIndex].EndedAt = incident.EndedAt
		data[incidentIndex].Phase.Generation = incident.Phase.Generation
		data[incidentIndex].Phase.Order = incident.Phase.Order
		data[incidentIndex].Affects = incident.GetImpactComponentList()
		data[incidentIndex].Updates = incident.GetIncidentUpdates()
	}

	response := api.IncidentListResponse{
		Data: &data,
	}

	return ctx.JSON(http.StatusOK, response) //nolint:wrapcheck
}

func (i *Implementation) CreateIncident(_ echo.Context) error {
	return nil
}

func (i *Implementation) DeleteIncident(_ echo.Context, _ api.IncidentIdPathParameter) error {
	return nil
}

// GetIncident retrieves a specific incident by ID.
func (i *Implementation) GetIncident(ctx echo.Context, incidentID string) error {
	var incident DbDef.Incident

	res := i.dbCon.Preload(clause.Associations).Where("id = ?", incidentID).First(&incident)

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
				Generation: incident.Phase.Generation,
				Order:      incident.Phase.Order,
			},
			Affects: incident.GetImpactComponentList(),
			Updates: incident.GetIncidentUpdates(),
		},
	}

	return ctx.JSON(http.StatusOK, response) //nolint:wrapcheck
}

func (i *Implementation) UpdateIncident(_ echo.Context, _ api.IncidentIdPathParameter) error {
	return nil
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
