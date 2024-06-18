package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/SovereignCloudStack/status-page-api/internal/app/config"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// Server holds the echo server, registry and config to provide metrics.
type Server struct {
	echo   *echo.Echo
	conf   *config.Metrics
	logger *zerolog.Logger
}

// New creates a new metrics server.
func New(metricsConfig *config.Metrics, logger *zerolog.Logger) *Server {
	metricsServer := echo.New()
	metricsServer.HideBanner = true
	metricsServer.HidePort = true

	metricsServer.GET("/metrics", echoprometheus.NewHandler())

	return &Server{
		echo:   metricsServer,
		conf:   metricsConfig,
		logger: logger,
	}
}

// GetMiddlewareConfig provides the middle ware for echo to supply metrics.
func (s *Server) GetMiddlewareConfig() echoprometheus.MiddlewareConfig {
	return echoprometheus.MiddlewareConfig{ //nolint: exhaustruct
		Namespace: s.conf.Namespace,
		Subsystem: s.conf.Subsystem,
	}
}

// Start checks the config and starts the server if configured.
func (s *Server) Start() error {
	if s.conf.Address != "" {
		s.logger.Log().Str("address", s.conf.Address).Msg("metrics server start listening")

		err := s.echo.Start(s.conf.Address)

		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("error running metrics server: %w", err)
	}

	s.logger.Debug().Msg("metrics server not configured")

	return nil
}

// Shutdown gracefully stops the metrics server.
func (s *Server) Shutdown(ctx context.Context) error {
	err := s.echo.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("error shutting down metrics server: %w", err)
	}

	return nil
}
