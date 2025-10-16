package config

import (
	"os"
	"path/filepath"

	"github.com/jarmocluyse/git-dash/internal/theme"
	"gopkg.in/yaml.v3"
)

// getConfigPath determines the configuration file path.
func (f *FileConfigService) getConfigPath() (string, error) {
	// Use custom config path if provided
	if f.customConfigPath != "" {
		return f.customConfigPath, nil
	}

	// Check for environment variable override
	if configPath := os.Getenv("GIT_DASH_CONFIG"); configPath != "" {
		return configPath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".git-dash.yaml"), nil
}

// configExists checks if configuration file exists.
func (f *FileConfigService) configExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// createEmptyConfig creates a default configuration.
func (f *FileConfigService) createEmptyConfig() *Config {
	// Use the default theme from code
	loadedTheme := theme.Default()

	// Default actions with keybindings
	defaultActions := []Action{
		{
			Name:        "Lazygit",
			Key:         "l",
			Command:     "lazygit",
			Args:        []string{"-p", "{path}"},
			Description: "Lazygit",
		},
		{
			Name:        "VS Code",
			Key:         "c",
			Command:     "code",
			Args:        []string{"{path}"},
			Description: "VS Code",
		},
	}

	return &Config{
		RepositoryPaths: []string{},
		Theme:           loadedTheme,
		Keybindings: Keybindings{
			Actions: defaultActions,
		},
	}
}

// loadFromFile loads configuration from YAML file.
func (f *FileConfigService) loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Start with default config
	config := f.createEmptyConfig()

	// Decode YAML over the defaults
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	// Merge user theme overrides with code defaults
	config.Theme = theme.MergeWithDefault(config.Theme)

	return config, nil
}

// writeToFile writes configuration to YAML file.
func (f *FileConfigService) writeToFile(path string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
