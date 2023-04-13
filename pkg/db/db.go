package db

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type ID string

func Provision(filename string, dbCon *gorm.DB) error {
	type ProvisionedResources struct {
		Components  []*Component  `yaml:"components"`
		ImpactTypes []*ImpactType `yaml:"impactTypes"`
		Phases      []*Phase      `yaml:"phases"`
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening provisioning file `%s`: %w", filename, err)
	}
	defer file.Close()

	resources := ProvisionedResources{}

	err = yaml.NewDecoder(file).Decode(&resources)
	if err != nil {
		return fmt.Errorf("error decoding provisioning file `%s`: %w", filename, err)
	}

	for _, component := range resources.Components {
		err = dbCon.Save(component).Error
		if err != nil {
			return fmt.Errorf("error saving component `%s`: %w", component.DisplayName, err)
		}
	}

	for _, impactType := range resources.ImpactTypes {
		err = dbCon.Save(impactType).Error
		if err != nil {
			return fmt.Errorf("error saving impact type `%s`: %w", impactType.Slug, err)
		}
	}

	var phaseOrder uint
	for phaseIndex := range resources.Phases {
		resources.Phases[phaseIndex].Order = phaseOrder

		err = dbCon.Save(&resources.Phases[phaseIndex]).Error
		if err != nil {
			return fmt.Errorf("error saving phase `%s`: %w", resources.Phases[phaseIndex].Slug, err)
		}

		phaseOrder++
	}

	// always add done phase as last
	err = dbCon.Save(&Phase{
		Slug:  "done",
		Order: phaseOrder,
	}).Error
	if err != nil {
		return fmt.Errorf("error saving phase `done`: %w", err)
	}

	return nil
}

func (l *Labels) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("error unmarshaling: %w", ErrLabelFormat)
	}

	const contentFieldSize = 2

	*l = make(Labels, len(value.Content)/contentFieldSize)

	for contentIndex := 0; contentIndex < len(value.Content); contentIndex += contentFieldSize {
		res := &(*l)[contentIndex/contentFieldSize]

		if err := value.Content[contentIndex].Decode(&res.Name); err != nil {
			return fmt.Errorf("error decoding label name: %w", err)
		}

		if err := value.Content[contentIndex+1].Decode(&res.Value); err != nil {
			return fmt.Errorf("error decoding label value: %w", err)
		}
	}

	return nil
}
