package main

import (
	"os"
	"time"

	"github.com/SovereignCloudStack/status-page-api/internal/app/config"
	"github.com/SovereignCloudStack/status-page-api/internal/app/logging"
	"github.com/SovereignCloudStack/status-page-api/internal/app/swagger"
	"github.com/SovereignCloudStack/status-page-api/internal/metrics"
	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-api/pkg/server"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() { //nolint:funlen
	// setup logging
	logger := log.Output(zerolog.ConsoleWriter{ //nolint:exhaustruct
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}).Level(zerolog.WarnLevel)

	// Reading config
	conf, err := config.New()
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading config")
	}

	err = conf.IsValid()
	if err != nil {
		logger.Fatal().Err(err).Msg("config is invalid")
	}

	// leveled logging
	switch conf.Verbose {
	case 1:
		logger = logger.Level(zerolog.InfoLevel)
	case 2: //nolint:gomnd
		logger = logger.Level(zerolog.DebugLevel)
	case 3: //nolint:gomnd
		logger = logger.Level(zerolog.TraceLevel)
	}

	logger.Trace().Interface("config", conf).Send()

	// named logging
	echoLogger := logger.With().Str("component", "echo").Logger()
	gormLogger := logger.With().Str("component", "gorm").Logger()
	handlerLogger := logger.With().Str("component", "handler").Logger()
	provisioningLogger := logger.With().Str("component", "provisioning").Logger()
	metricsLogger := logger.With().Str("component", "metrics").Logger()

	// metric server
	metricsServer := metrics.New(&conf.Metrics, &metricsLogger)
	go func() {
		logger.Fatal().Err(metricsServer.Start()).Msg("error running metrics server")
	}()

	// HTTP setup
	echoServer := echo.New()
	echoServer.HideBanner = true
	echoServer.HidePort = true

	echoServer.Use(logging.NewEchoZerlogLogger(&echoLogger))
	echoServer.Use(middleware.Recover())
	echoServer.Use(middleware.RemoveTrailingSlash())
	echoServer.Use(middleware.CORSWithConfig(middleware.CORSConfig{ //nolint:exhaustruct
		AllowOrigins: conf.AllowedOrigins,
	}))
	echoServer.Use(echoprometheus.NewMiddlewareWithConfig(metricsServer.GetMiddlewareConfig()))

	// open api spec and swagger
	echoServer.GET("/openapi.json", swagger.ServeOpenAPISpec)

	if conf.SwaggerEnabled {
		echoServer.GET("/swagger", swagger.ServeSwagger)
	}

	// DB setup
	dbCon, err := gorm.Open(postgres.Open(conf.Database.ConnectionString), &gorm.Config{ //nolint:exhaustruct
		Logger: logging.NewGormLogger(&gormLogger),
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("error opening database connection")
	}

	err = dbCon.AutoMigrate(
		&DbDef.Component{},      //nolint:exhaustruct
		&DbDef.Phase{},          //nolint:exhaustruct
		&DbDef.IncidentUpdate{}, //nolint:exhaustruct
		&DbDef.Incident{},       //nolint:exhaustruct
		&DbDef.ImpactType{},     //nolint:exhaustruct
		&DbDef.Impact{},         //nolint:exhaustruct
		&DbDef.Severity{},       //nolint:exhaustruct
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("error migrating structures")
	}

	// Initialize "static" DB contents
	err = DbDef.Provision(conf.ProvisioningFile, dbCon, &provisioningLogger)
	if err != nil {
		logger.Fatal().Err(err).Msg("error provisioning data")
	}

	// register api
	api.RegisterHandlers(echoServer, server.New(dbCon, &handlerLogger))

	// Starting server
	logger.Log().Str("address", conf.ListenAddress).Msg("server start listening")
	logger.Fatal().Err(echoServer.Start(conf.ListenAddress)).Msg("error running server")
}
