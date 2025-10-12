package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/jarmocluyse/wip-tui/internal/theme"
)

// getConfigPath determines the configuration file path.
func (f *FileConfigService) getConfigPath() (string, error) {
	// Use custom config path if provided
	if f.customConfigPath != "" {
		return f.customConfigPath, nil
	}

	// Check for environment variable override
	if configPath := os.Getenv("GIT_TUI_CONFIG"); configPath != "" {
		return configPath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".git-tui.toml"), nil
}

// configExists checks if configuration file exists.
func (f *FileConfigService) configExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// createEmptyConfig creates a default configuration.
func (f *FileConfigService) createEmptyConfig() *Config {
	// Use theme loader to get the proper theme with loading order
	themeLoader := theme.NewLoader()
	loadedTheme, _ := themeLoader.LoadTheme()

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

// loadFromFile loads configuration from TOML file.
func (f *FileConfigService) loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Start with default config
	config := f.createEmptyConfig()

	// Decode TOML over the defaults
	_, err = toml.Decode(string(data), config)
	if err != nil {
		return nil, err
	}

	// Ensure theme has all default values if missing
	config.Theme = theme.MergeWithDefault(config.Theme)

	return config, nil
}

// writeToFile writes configuration to TOML file.
func (f *FileConfigService) writeToFile(path string, config *Config) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(config)
}
