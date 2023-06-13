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

	incidentList := make([]*api.Incident, len(incidents))
	for incidentIndex := range incidentList {
		incidentList[incidentIndex] = IncidentFromDB(incidents[incidentIndex])
	}

	return ctx.JSON(http.StatusOK, incidentList)
}

func (i *Implementation) CreateIncident(ctx echo.Context) error
func (i *Implementation) DeleteIncident(ctx echo.Context, incidentId api.IncidentIdPathParameter) error

// GetIncident retrieves a specific incident by ID.
func (i *Implementation) GetIncident(ctx echo.Context, incidentID string) error {
	var incident DbDef.Incident

	res := i.dbCon.Preload(clause.Associations).Where("id = ?", incidentID).First(&incident)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, IncidentFromDB(&incident))
}

func (i *Implementation) UpdateIncident(ctx echo.Context, incidentId api.IncidentIdPathParameter) error

func (i *Implementation) GetIncidentUpdates(ctx echo.Context, incidentId api.IncidentIdPathParameter) error
func (i *Implementation) CreateIncidentUpdate(ctx echo.Context, incidentId api.IncidentIdPathParameter) error
func (i *Implementation) DeleteIncidentUpdate(ctx echo.Context, incidentId api.IncidentIdPathParameter, updateOrder api.IncidentUpdateOrderPathParameter) error
func (i *Implementation) GetIncidentUpdate(ctx echo.Context, incidentId api.IncidentIdPathParameter, updateOrder api.IncidentUpdateOrderPathParameter) error
func (i *Implementation) UpdateIncidentUpdate(ctx echo.Context, incidentId api.IncidentIdPathParameter, updateOrder api.IncidentUpdateOrderPathParameter) error

// IncidentFromDB is a helper function, converting a [db.Incident] to an [api.Incident].
func IncidentFromDB(incident *DbDef.Incident) *api.Incident {
	return &api.Incident{}
}

// IncidentUpdatesFromDB is a helper function, converting a list of [db.IncidentUpdate]s to a list of [api.IncidentUpdate]s.
func IncidentUpdatesFromDB(updates []DbDef.IncidentUpdate) []api.IncidentUpdate {
	updateList := make([]api.IncidentUpdate, len(updates))
	return updateList
}
