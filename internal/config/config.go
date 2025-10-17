// Package config provides configuration management for the application.
package config

import "github.com/jarmocluyse/git-dash/internal/theme/types"

// Config represents the application configuration.
type Config struct {
	Title           string      `yaml:"title"`
	RepositoryPaths []string    `yaml:"repository_paths"`
	Theme           types.Theme `yaml:"theme"`
	Keybindings     Keybindings `yaml:"keybindings"`
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

// isValidIndex checks if the index is within valid bounds.
func (c *Config) isValidIndex(index int) bool {
	return index >= 0 && index < len(c.RepositoryPaths)
}
