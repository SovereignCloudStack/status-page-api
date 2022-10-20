package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ProvisionedResources struct {
	Components []*Component `json:"components"`
}

func provisionResources(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	resources := ProvisionedResources{}
	err = yaml.NewDecoder(file).Decode(&resources)
	if err != nil {
		return err
	}
	for _, component := range resources.Components {
		err := db.Save(&component).Error
		if err != nil {
			return err
		}
	}
	return nil
}
