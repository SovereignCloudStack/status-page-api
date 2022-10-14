package main

import (
	"io"
	"log"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.POST("/", func(c echo.Context) error {
		data, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return echo.NewHTTPError(400)
		}
		log.Println("Got token!", string(data))
		return nil
	})

	log.Fatal(e.Start(":3002"))
}
