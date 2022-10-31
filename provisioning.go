package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ProvisionedResources struct {
	Components  []*Component  `yaml:"components"`
	ImpactTypes []*ImpactType `yaml:"impactTypes"`
	Phases      []*Phase      `yaml:"phases"`
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
	for _, impactType := range resources.ImpactTypes {
		err := db.Save(&impactType).Error
		if err != nil {
			return err
		}
	}
	for _, phase := range resources.Phases {
		phase.provisioned = true
		err := db.Save(&phase).Error
		if err != nil {
			return err
		}
	}
	return nil
}
