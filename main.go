package main

import (
	"flag"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

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

	// Reading config
	componentsFile := flag.String("components-file", "./components.yaml", "YAML file containing components")
	addr := flag.String("addr", ":3000", "Address to listen on")
	flag.Parse()
	err := loadComponents(*componentsFile)
	if err != nil {
		e.Logger.Fatal(err)
	}

	// Starting server
	e.Logger.Fatal(e.Start(*addr))
}
