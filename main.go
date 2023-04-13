package main

import (
	"flag"
	"strings"

	"github.com/SovereignCloudStack/status-page-api/internal/app/swagger"
	DbDef "github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-api/pkg/server"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
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
	echoServer.Use(middleware.Logger())
	echoServer.Use(middleware.Recover())
	echoServer.Use(middleware.RemoveTrailingSlash())
	echoServer.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: strings.Split(*corsOrigins, ","),
	}))

	api.RegisterHandlers(echoServer, &server.Implementation{})

	echoServer.GET("/openapi.json", swagger.ServeOpenAPISpec)
	echoServer.GET("/swagger", swagger.ServeSwagger)

	// DB setup
	dbCon, err := gorm.Open(postgres.Open(*dbDsn), &gorm.Config{})
	if err != nil {
		echoServer.Logger.Fatal(err)
	}

	err = dbCon.AutoMigrate(
		&DbDef.Label{},
		&DbDef.ImpactType{},
		&DbDef.Phase{},
		&DbDef.Incident{},
		&DbDef.IncidentUpdate{},
		&DbDef.Component{},
	)
	if err != nil {
		echoServer.Logger.Fatal(err)
	}

	// Initialize "static" DB contents
	err = DbDef.Provision(*provisioningFile, dbCon)
	if err != nil {
		echoServer.Logger.Fatal(err)
	}

	// Starting server
	echoServer.Logger.Fatal(echoServer.Start(*addr))
}
