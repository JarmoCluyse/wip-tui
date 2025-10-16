package main

import (
	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/repomanager"
	themeService "github.com/jarmocluyse/git-dash/internal/services/theme"
	"github.com/jarmocluyse/git-dash/ui"
)

// AppDependencies implements the ui.Dependencies interface.
type AppDependencies struct {
	configService config.ConfigService
	repoManager   *repomanager.RepoManager
	themeService  themeService.Service
}

// NewAppDependencies creates a new dependency container with services.
func NewAppDependencies(configPath string) *AppDependencies {
	var configService config.ConfigService
	if configPath != "" {
		configService = config.NewFileConfigServiceWithPath(configPath)
	} else {
		configService = config.NewFileConfigService()
	}

	// Create services
	repoManager := repomanager.NewRepoManager(configService)
	themeManager := themeService.NewManager(configService)

	// Initialize the repo manager
	if err := repoManager.Init(); err != nil {
		// Handle error gracefully, but continue
		// The UI can handle an empty repo manager
	}

	return &AppDependencies{
		configService: configService,
		repoManager:   repoManager,
		themeService:  themeManager,
	}
}

// GetConfigService returns the configuration service.
func (d *AppDependencies) GetConfigService() config.ConfigService {
	return d.configService
}

// GetRepoManager returns the repository manager.
func (d *AppDependencies) GetRepoManager() *repomanager.RepoManager {
	return d.repoManager
}

// GetThemeService returns the theme service.
func (d *AppDependencies) GetThemeService() themeService.Service {
	return d.themeService
}

// GetExplorerService returns the explorer service.
func (d *AppDependencies) GetExplorerService() interface{} {
	// TODO: Implement explorer service with new architecture
	return nil
}

// Verify at compile time that AppDependencies implements ui.Dependencies
var _ ui.Dependencies = (*AppDependencies)(nil)
