package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/SovereignCloudStack/status-page-api/internal/app/config"
	"github.com/SovereignCloudStack/status-page-api/internal/app/logging"
	"github.com/SovereignCloudStack/status-page-api/internal/app/swagger"
	apiServerDefinition "github.com/SovereignCloudStack/status-page-openapi/pkg/api/server"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

// Server wraps the echo http server as API Server.
type Server struct {
	echo   *echo.Echo
	conf   *config.Server
	logger *zerolog.Logger
}

// New creates a new wrapped server.
func New(conf *config.Server, logger *zerolog.Logger, promMiddlewareConfig echoprometheus.MiddlewareConfig) *Server {
	// general server settings
	echoServer := echo.New()
	echoServer.HideBanner = true
	echoServer.HidePort = true

	// middlewares
	echoServer.Use(logging.NewEchoZerlogLogger(logger))
	echoServer.Use(middleware.Recover())
	echoServer.Use(middleware.RemoveTrailingSlash())

	if conf.CORS.Enabled {
		echoServer.Use(middleware.CORSWithConfig(middleware.CORSConfig{ //nolint:exhaustruct
			AllowOrigins: conf.CORS.AllowedOrigins,
		}))
	}

	echoServer.Use(echoprometheus.NewMiddlewareWithConfig(promMiddlewareConfig))

	// open api spec and swagger
	echoServer.GET("/openapi.json", swagger.ServeOpenAPISpec)

	if conf.SwaggerEnabled {
		echoServer.GET("/swagger", swagger.ServeSwagger)
	}

	return &Server{
		echo:   echoServer,
		conf:   conf,
		logger: logger,
	}
}

// RegisterAPI registers api spec and api implementation to the echo server.
func (s *Server) RegisterAPI(apiImplementation apiServerDefinition.ServerInterface) {
	apiServerDefinition.RegisterHandlers(s.echo, apiImplementation)
}

// Start starts the wrapped echo server.
func (s *Server) Start() error {
	s.logger.Log().Str("address", s.conf.Address).Msg("api server start listening")

	err := s.echo.Start(s.conf.Address)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return fmt.Errorf("error running api server: %w", err)
}

// Shutdown gracefully stops the api server.
func (s *Server) Shutdown(ctx context.Context) error {
	err := s.echo.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("error shutting down api server: %w", err)
	}

	return nil
}
