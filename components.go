package main

import (
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
	"os"
)

var components map[string]string

func loadComponents(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return yaml.NewDecoder(file).Decode(&components)
}

func componentList(c echo.Context) error {
	return c.JSON(200, components)
}

func componentGet(c echo.Context) error {
	component, ok := components[c.Param("id")]
	if !ok {
		return c.JSON(404, nil)
	}
	return c.JSON(200, component)
}
