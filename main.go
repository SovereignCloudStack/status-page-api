package main

import (
	"flag"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// HTTP setup
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RemoveTrailingSlash())

	components := e.Group("/components")
	{
		components.GET("", componentList)
		components.GET("/:id", componentGet)
	}

	incidents := e.Group("/incidents")
	{
		incidents.GET("", incidentList)
		incidents.POST("", incidentAdd)
		incidents.GET("/:id", incidentGet)
	}

	// Reading config
	dbDsn := flag.String("db.dsn", "host=127.0.0.1 user=root dbname=defaultdb port=26257 sslmode=disable", "DB dsn")
	componentsFile := flag.String("components-file", "./components.yaml", "YAML file containing components")
	addr := flag.String("addr", ":3000", "Address to listen on")
	flag.Parse()
	err := loadComponents(*componentsFile)
	if err != nil {
		e.Logger.Fatal(err)
	}

	db, err = gorm.Open(postgres.Open(*dbDsn), &gorm.Config{})
	if err != nil {
		e.Logger.Fatal(err)
	}
	db.AutoMigrate(&Incident{})

	// Starting server
	e.Logger.Fatal(e.Start(*addr))
}
