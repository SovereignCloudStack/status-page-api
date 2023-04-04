package db

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type Id string

func Provision(filename string, db *gorm.DB) error {
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
		err = db.Save(component).Error
		if err != nil {
			return fmt.Errorf("error saving component `%s`: %w", component.DisplayName, err)
		}
	}

	for _, impactType := range resources.ImpactTypes {
		err := db.Save(impactType).Error
		if err != nil {
			return fmt.Errorf("error saving impact type `%s`: %w", impactType.Slug, err)
		}
	}

	for _, phase := range resources.Phases {
		err := db.Save(&phase).Error
		if err != nil {
			return fmt.Errorf("error saving phase `%s`: %w", phase.Slug, err)
		}
	}

	return nil
}

func (l *Labels) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("`labels` must contain YAML mapping, has %v", value.Kind)
	}

	*l = make(Labels, len(value.Content)/2)

	for i := 0; i < len(value.Content); i += 2 {
		var res = &(*l)[i/2]

		if err := value.Content[i].Decode(&res.Slug); err != nil {
			return err
		}

		if err := value.Content[i+1].Decode(&res.Value); err != nil {
			return err
		}
	}

	return nil
}
