package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Assuming ResourceAdjustment and other related types are defined elsewhere

func ApplyVPARecommendationsToYAML(repoDir string, adjustment ResourceAdjustment, resourceConfigurations map[string]ResourceSpecs) error {
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

		// Example logic to update resourceConfig based on VPA recommendations
		// This is a placeholder - you'll need to adjust paths and structure based on your actual YAML structure
		for containerName, specs := range resourceConfigurations {
			// Logic to navigate to the correct part of resourceConfig and apply updates
			// This is highly dependent on the structure of your YAML and what you're adjusting
			// For example, you might have a path like resourceConfig["spec"]["containers"]
			// and then loop through containers to find a match by name and adjust resources
		}

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
