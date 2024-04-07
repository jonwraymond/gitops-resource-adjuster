package config

import (
	"io/ioutil"
	"path/filepath"

	v1alpha1 "github.com/jonwraymond/gitops-resource-adjuster/api/v1alpha1" // Example path
	"gopkg.in/yaml.v2"
)


func ApplyVPARecommendationsToYAML(repoDir string, adjustment v1alpha1.ResourceAdjustment) error {
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
