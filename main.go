package main

import (
	"flag"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// Reading config
	dbDsn := flag.String("postgres-dsn", "host=127.0.0.1 user=postgres dbname=postgres port=5432 sslmode=disable", "DB dsn")
	componentsFile := flag.String("components-file", "./components.yaml", "YAML file containing components")
	addr := flag.String("addr", ":3000", "Address to listen on")
	var corsOrigins string
	flag.StringVar(&corsOrigins, "cors-origins", "127.0.0.1,localhost", "Allowed CORS origins, seperated by ','")
	flag.Parse()

	// HTTP setup
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: strings.Split(corsOrigins, ","),
	}))

	components := e.Group("/components")
	{
		components.GET("", componentList)
		components.GET("/:slug", componentGet)
	}

	incidents := e.Group("/incidents")
	{
		incidents.GET("", incidentList)
		incidents.POST("", incidentAdd)
		incident := incidents.Group("/:id")
		{
			incident.GET("", incidentGet)
			updates := incident.Group("/updates")
			{
				updates.POST("", updateAdd)
				updates.GET("", updatesGet)
				updates.PATCH("/:updateid", updateUpdate)
			}
		}

	}

	// Setup DB
	var err error
	db, err = gorm.Open(postgres.Open(*dbDsn), &gorm.Config{})
	if err != nil {
		e.Logger.Fatal(err)
	}
	db.AutoMigrate(&Incident{}, &Component{}, &Update{})

	// Initialize "static" DB contents
	err = loadComponents(*componentsFile)
	if err != nil {
		e.Logger.Fatal(err)
	}

	// Starting server
	e.Logger.Fatal(e.Start(*addr))
}
