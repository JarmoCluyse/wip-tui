package config

// ConfigService defines the interface for configuration management.
type ConfigService interface {
	Load() (*Config, error)
	Save(config *Config) error
}

// FileConfigService implements ConfigService using file-based storage.
type FileConfigService struct {
	customConfigPath string
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
