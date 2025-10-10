package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/jarmocluyse/wip-tui/internal/theme"
)

type Config struct {
	RepositoryPaths []string    `toml:"repository_paths"`
	Theme           theme.Theme `toml:"theme"`
}

type ConfigService interface {
	Load() (*Config, error)
	Save(config *Config) error
}

type FileConfigService struct {
	customConfigPath string
}

func NewFileConfigService() ConfigService {
	return &FileConfigService{}
}

func NewFileConfigServiceWithPath(configPath string) ConfigService {
	return &FileConfigService{
		customConfigPath: configPath,
	}
}

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

func (f *FileConfigService) Save(config *Config) error {
	configPath, err := f.getConfigPath()
	if err != nil {
		return err
	}

	return f.writeToFile(configPath, config)
}

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

func (f *FileConfigService) configExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (f *FileConfigService) createEmptyConfig() *Config {
	// Use theme loader to get the proper theme with loading order
	themeLoader := theme.NewLoader()
	loadedTheme, _ := themeLoader.LoadTheme()

	return &Config{
		RepositoryPaths: []string{},
		Theme:           loadedTheme,
	}
}

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

func (f *FileConfigService) writeToFile(path string, config *Config) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(config)
}

func (c *Config) AddRepositoryPath(path string) {
	c.RepositoryPaths = append(c.RepositoryPaths, path)
}

func (c *Config) RemoveRepositoryPath(index int) {
	if c.isValidIndex(index) {
		c.RepositoryPaths = append(c.RepositoryPaths[:index], c.RepositoryPaths[index+1:]...)
	}
}

func (c *Config) RemoveRepositoryPathByValue(path string) {
	for i, p := range c.RepositoryPaths {
		if p == path {
			c.RemoveRepositoryPath(i)
			break
		}
	}
}

func (c *Config) isValidIndex(index int) bool {
	return index >= 0 && index < len(c.RepositoryPaths)
}
