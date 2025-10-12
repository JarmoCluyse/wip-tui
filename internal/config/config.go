// Package config provides configuration management for the application.
package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jarmocluyse/wip-tui/internal/theme"
)

// Action represents a configurable action with key binding and command.
type Action struct {
	Name        string   `toml:"name"`        // Display name for the action
	Key         string   `toml:"key"`         // Key binding (e.g., "l", "o", "ctrl+o")
	Command     string   `toml:"command"`     // The command to execute
	Args        []string `toml:"args"`        // Arguments to pass to the command
	Description string   `toml:"description"` // Description of what this action does
}

// Keybindings holds configuration for key bindings.
type Keybindings struct {
	Actions []Action `toml:"actions"` // List of configurable actions
}

// Config represents the application configuration.
type Config struct {
	RepositoryPaths []string    `toml:"repository_paths"`
	Theme           theme.Theme `toml:"theme"`
	Keybindings     Keybindings `toml:"keybindings"`
}

// ConfigService defines the interface for configuration management.
type ConfigService interface {
	Load() (*Config, error)
	Save(config *Config) error
}

// FileConfigService implements ConfigService using file-based storage.
type FileConfigService struct {
	customConfigPath string
}

// NewFileConfigService creates a new file-based config service.
func NewFileConfigService() ConfigService {
	return &FileConfigService{}
}

// NewFileConfigServiceWithPath creates a config service with custom path.
func NewFileConfigServiceWithPath(configPath string) ConfigService {
	return &FileConfigService{
		customConfigPath: configPath,
	}
}

// Load loads configuration from file or creates default if not found.
func (f *FileConfigService) Load() (*Config, error) {
	configPath, err := f.getConfigPath()
	if err != nil {
		return nil, err
	}

	if !f.configExists(configPath) {
		return f.createEmptyConfig(), nil
	}

	return f.loadFromFile(configPath)
}

// Save saves configuration to file.
func (f *FileConfigService) Save(config *Config) error {
	configPath, err := f.getConfigPath()
	if err != nil {
		return err
	}

	return f.writeToFile(configPath, config)
}

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
			Description: "Open repository in Lazygit",
		},
		{
			Name:        "VS Code",
			Key:         "c",
			Command:     "code",
			Args:        []string{"{path}"},
			Description: "Open repository in VS Code",
		},
		{
			Name:        "Terminal",
			Key:         "t",
			Command:     "gnome-terminal",
			Args:        []string{"--working-directory={path}"},
			Description: "Open terminal in repository directory",
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

// AddRepositoryPath adds a repository path to the configuration.
func (c *Config) AddRepositoryPath(path string) {
	c.RepositoryPaths = append(c.RepositoryPaths, path)
}

// RemoveRepositoryPath removes a repository path by index.
func (c *Config) RemoveRepositoryPath(index int) {
	if c.isValidIndex(index) {
		c.RepositoryPaths = append(c.RepositoryPaths[:index], c.RepositoryPaths[index+1:]...)
	}
}

// RemoveRepositoryPathByValue removes a repository path by value.
func (c *Config) RemoveRepositoryPathByValue(path string) {
	for i, p := range c.RepositoryPaths {
		if p == path {
			c.RemoveRepositoryPath(i)
			break
		}
	}
}

// FindActionByKey finds an action by its key binding.
func (k *Keybindings) FindActionByKey(key string) *Action {
	for i, action := range k.Actions {
		if action.Key == key {
			return &k.Actions[i]
		}
	}
	return nil
}

// GetActionKeys returns all configured action keys for help display.
func (k *Keybindings) GetActionKeys() []string {
	var keys []string
	for _, action := range k.Actions {
		keys = append(keys, action.Key)
	}
	return keys
}

// ExecuteOpenAction executes the configured action with the given path.
func (a *Action) ExecuteOpenAction(path string) *exec.Cmd {
	// Replace {path} placeholder in command and args
	command := strings.ReplaceAll(a.Command, "{path}", path)

	var args []string
	for _, arg := range a.Args {
		args = append(args, strings.ReplaceAll(arg, "{path}", path))
	}

	return exec.Command(command, args...)
}

// isValidIndex checks if the index is within valid bounds.
func (c *Config) isValidIndex(index int) bool {
	return index >= 0 && index < len(c.RepositoryPaths)
}
