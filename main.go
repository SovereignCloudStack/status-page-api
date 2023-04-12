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

var db *gorm.DB

func main() {
	// Reading config
	dbDsn := flag.String("postgres-dsn", "host=127.0.0.1 user=postgres dbname=postgres port=5432 password=debug sslmode=disable", "DB dsn")
	provisioningFile := flag.String("provisioning-file", "./provisioning.yaml", "YAML file containing components etc. to be provisioned on startup")
	addr := flag.String("addr", ":3000", "Address to listen on")
	corsOrigins := flag.String("cors-origins", "127.0.0.1,localhost", "Allowed CORS origins, seperated by ','")
	flag.Parse()

	// HTTP setup
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: strings.Split(*corsOrigins, ","),
	}))

	api.RegisterHandlers(e, &server.ServerImplementation{})

	e.GET("/openapi.json", func(c echo.Context) error {
		swagger, err := api.GetSwagger()
		if err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(500)
		}
		return c.JSON(200, swagger)
	})

	e.GET("/swagger", swagger.ServeSwagger)

	// Setup DB
	db, err := gorm.Open(postgres.Open(*dbDsn), &gorm.Config{})
	if err != nil {
		e.Logger.Fatal(err)
	}
	err = db.AutoMigrate(
		&DbDef.Label{},
		&DbDef.ImpactType{},
		&DbDef.Phase{},
		&DbDef.Incident{},
		&DbDef.IncidentUpdate{},
		&DbDef.Component{},
	)

	if err != nil {
		e.Logger.Fatal(err)
	}

	// Initialize "static" DB contents
	err = DbDef.Provision(*provisioningFile, db)
	if err != nil {
		e.Logger.Fatal(err)
	}

	// Starting server
	e.Logger.Fatal(e.Start(*addr))
}
