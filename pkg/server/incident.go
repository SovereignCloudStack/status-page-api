package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm/clause"
)

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

// IncidentFromDB is a helper function, converting a [db.Incident] to an [api.Incident].
func IncidentFromDB(incident *DbDef.Incident) *api.Incident {
	return &api.Incident{
		Affects:     incident.GetAffectsIds(),
		BeganAt:     incident.BeganAt,
		Description: incident.Description,
		EndedAt:     incident.EndedAt,
		Id:          string(incident.ID),
		ImpactType:  incident.ImpactType.Slug,
		Phase:       incident.Phase.Slug,
		Title:       incident.Title,
		Updates:     IncidentUpdatesFromDB(incident.Updates),
	}
}

// IncidentUpdatesFromDB is a helper function, converting a list of [db.IncidentUpdate]s to a list of [api.IncidentUpdate]s.
func IncidentUpdatesFromDB(updates []DbDef.IncidentUpdate) []api.IncidentUpdate {
	updateList := make([]api.IncidentUpdate, len(updates))

	for updateIndex := range updateList {
		updateList[updateIndex].CreatedAt = updates[updateIndex].CreatedAt
		updateList[updateIndex].Text = updates[updateIndex].Text
	}

	return updateList
}
