package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/SovereignCloudStack/status-page-api/internal/app/logger"
	"github.com/SovereignCloudStack/status-page-api/internal/app/swagger"
	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-api/pkg/server"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() { //nolint:funlen
	// setup logging
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Logger = log.Output(zerolog.ConsoleWriter{ //nolint:exhaustruct
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	echoLogger := log.With().Str("component", "echo").Logger()
	gormLogger := log.With().Str("component", "gorm").Logger()
	handlerLogger := log.With().Str("component", "handler").Logger()

	// Reading config
	dbDsn := flag.String(
		"postgres-dsn",
		"host=127.0.0.1 user=postgres dbname=postgres port=5432 password=debug sslmode=disable",
		"DB dsn",
	)
	provisioningFile := flag.String(
		"provisioning-file",
		"./provisioning.yaml",
		"YAML file containing components etc. to be provisioned on startup",
	)
	addr := flag.String(
		"addr",
		":3000",
		"Address to listen on",
	)
	corsOrigins := flag.String(
		"cors-origins",
		"127.0.0.1,localhost",
		"Allowed CORS origins, separated by ','",
	)

	flag.Parse()

	// HTTP setup
	echoServer := echo.New()
	echoServer.HideBanner = true
	echoServer.HidePort = true

	echoServer.Use(logger.NewEchoLoggerConfig(&echoLogger))
	echoServer.Use(middleware.Recover())
	echoServer.Use(middleware.RemoveTrailingSlash())
	echoServer.Use(middleware.CORSWithConfig(middleware.CORSConfig{ //nolint:exhaustruct
		AllowOrigins: strings.Split(*corsOrigins, ","),
	}))

	echoServer.GET("/openapi.json", swagger.ServeOpenAPISpec)
	echoServer.GET("/swagger", swagger.ServeSwagger)

	// DB setup
	dbCon, err := gorm.Open(postgres.Open(*dbDsn), &gorm.Config{ //nolint:exhaustruct
		Logger: logger.NewGormLogger(&gormLogger),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("error opening database connection")
	}

	err = dbCon.AutoMigrate(
		&DbDef.Component{},      //nolint:exhaustruct
		&DbDef.Phase{},          //nolint:exhaustruct
		&DbDef.IncidentUpdate{}, //nolint:exhaustruct
		&DbDef.Incident{},       //nolint:exhaustruct
		&DbDef.ImpactType{},     //nolint:exhaustruct
		&DbDef.Impact{},         //nolint:exhaustruct
	)
	if err != nil {
		log.Fatal().Err(err).Msg("error migrating structures")
	}

	// Initialize "static" DB contents
	err = DbDef.Provision(*provisioningFile, dbCon)
	if err != nil {
		log.Fatal().Err(err).Msg("error provisioning data")
	}

	// register api
	api.RegisterHandlers(echoServer, server.New(dbCon, &handlerLogger))

	// Starting server
	log.Info().Str("address", *addr).Msg("server start listening")
	log.Fatal().Err(echoServer.Start(*addr)).Msg("error running server")
}
