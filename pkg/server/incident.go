package server

import (
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
)

func (s *Implementation) GetIncident(ctx echo.Context, incidentId string) error {
	return nil
}

func (s *Implementation) GetIncidents(ctx echo.Context, params api.GetIncidentsParams) error {
	return nil
}
