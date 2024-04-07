package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"github.com/mycompany/resourceadjuster/api/v1"
)

func ApplyVPARecommendationsToYAML(repoDir string, adjustment v1.ResourceAdjustment) error {
	for _, path := range adjustment.Spec.Paths {
		fullPath := filepath.Join(repoDir, path)
		data, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return err
		}

		var resourceConfig map[string]interface{}
		if err := yaml.Unmarshal(data, &resourceConfig); err != nil {
			return err
		}

		// Logic to update resourceConfig based on VPA recommendations goes here

		updatedData, err := yaml.Marshal(resourceConfig)
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(fullPath, updatedData, 0644); err != nil {
			return err
		}
	}

	return nil
}
