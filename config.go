package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Repository struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	AutoDiscover   bool   `json:"auto_discover"`
	HasUncommitted bool   `json:"-"`
	HasUnpushed    bool   `json:"-"`
	HasUntracked   bool   `json:"-"`
	HasError       bool   `json:"-"`
	IsWorktree     bool   `json:"-"`
	IsBare         bool   `json:"-"`
}

type Config struct {
	Repositories []Repository `json:"repositories"`
}

type ConfigService interface {
	Load() (*Config, error)
	Save(config *Config) error
}

type FileConfigService struct{}

func NewFileConfigService() ConfigService {
	return &FileConfigService{}
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".git-tui.json"), nil
}

func (f *FileConfigService) configExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (f *FileConfigService) createEmptyConfig() *Config {
	return &Config{Repositories: []Repository{}}
}

func (f *FileConfigService) loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (f *FileConfigService) writeToFile(path string, config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c *Config) AddRepository(name, path string) {
	repo := Repository{
		Name: name,
		Path: path,
	}
	c.Repositories = append(c.Repositories, repo)
}

func (c *Config) AddRepositoryWithAutoDiscover(name, path string, autoDiscover bool) {
	repo := Repository{
		Name:         name,
		Path:         path,
		AutoDiscover: autoDiscover,
	}
	c.Repositories = append(c.Repositories, repo)
}

func (c *Config) RemoveRepository(index int) {
	if c.isValidIndex(index) {
		c.Repositories = append(c.Repositories[:index], c.Repositories[index+1:]...)
	}
}

func (c *Config) isValidIndex(index int) bool {
	return index >= 0 && index < len(c.Repositories)
}
