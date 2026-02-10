package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"repoctr/pkg/models"
)

const configFileName = ".repoctrconfig.yaml"

// LoadConfig loads the .repoctrconfig.yaml file from the given directory.
// Returns an empty config if the file doesn't exist.
func LoadConfig(rootDir string) (*models.RepoCtrConfig, error) {
	configPath := filepath.Join(rootDir, configFileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &models.RepoCtrConfig{}, nil
		}
		return nil, err
	}

	var cfg models.RepoCtrConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveConfig saves the configuration to .repoctrconfig.yaml.
func SaveConfig(rootDir string, cfg *models.RepoCtrConfig) error {
	configPath := filepath.Join(rootDir, configFileName)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// Add header comment
	header := `# .repoctrconfig.yaml - Repository configuration
# This file contains user-defined exclusion patterns and project overrides.
# It is NOT auto-generated and is safe to edit manually.

`
	content := header + string(data)

	return os.WriteFile(configPath, []byte(content), 0644)
}

// ConfigPath returns the path to the .repoctrconfig.yaml file.
func ConfigPath(rootDir string) string {
	return filepath.Join(rootDir, configFileName)
}
