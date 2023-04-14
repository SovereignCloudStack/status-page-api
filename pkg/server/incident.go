package server

import (
	"net/http"

	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm/clause"
)

func (i *Implementation) GetIncident(ctx echo.Context, incidentID string) error {
	var incident DbDef.Incident

	res := i.dbCon.Preload(clause.Associations).Where("id = ?", incidentID).First(&incident)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, IncidentFromDB(&incident))
}

func (i *Implementation) GetIncidents(ctx echo.Context, params api.GetIncidentsParams) error {
	var incidents []*DbDef.Incident

	res := i.dbCon.Preload(
		clause.Associations,
	).Where(
		"began_at > ?", params.Start,
	).Where(
		"ended_at < ?", params.End,
	).Find(&incidents)

	err := res.Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	incidentList := make([]*api.Incident, len(incidents))
	for incidentIndex := range incidentList {
		incidentList[incidentIndex] = IncidentFromDB(incidents[incidentIndex])
	}

	return ctx.JSON(http.StatusOK, incidentList)
}

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

func IncidentUpdatesFromDB(updates []DbDef.IncidentUpdate) []api.IncidentUpdate {
	updateList := make([]api.IncidentUpdate, len(updates))

	for updateIndex := range updateList {
		updateList[updateIndex].CreatedAt = updates[updateIndex].CreatedAt
		updateList[updateIndex].Text = updates[updateIndex].Text
	}

	return updateList
}
