package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func UpdateYAMLFiles(repoDir string, adjustments []Adjustment) error {
	for _, adj := range adjustments {
		filePath := filepath.Join(repoDir, adj.Path)
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}

		var resourceConfig map[string]interface{}
		if err := yaml.Unmarshal(data, &resourceConfig); err != nil {
			return err
		}

		// Logic to update resourceConfig with adjustments goes here

		updatedData, err := yaml.Marshal(resourceConfig)
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(filePath, updatedData, 0644); err != nil {
			return err
		}
	}

	return nil
}
