package main

import (
	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/git"
	"github.com/jarmocluyse/git-dash/internal/repository"
	explorerService "github.com/jarmocluyse/git-dash/internal/services/explorer"
	repositoryService "github.com/jarmocluyse/git-dash/internal/services/repository"
	themeService "github.com/jarmocluyse/git-dash/internal/services/theme"
	"github.com/jarmocluyse/git-dash/ui"
)

// AppDependencies implements the ui.Dependencies interface.
type AppDependencies struct {
	configService     config.ConfigService
	statusUpdater     *repository.StatusUpdater
	repositoryService repositoryService.Service
	themeService      themeService.Service
	explorerService   explorerService.Service
}

// NewAppDependencies creates a new dependency container with services.
func NewAppDependencies(configPath string) *AppDependencies {
	var configService config.ConfigService
	if configPath != "" {
		configService = config.NewFileConfigServiceWithPath(configPath)
	} else {
		configService = config.NewFileConfigService()
	}

	gitChecker := git.NewCachedChecker()
	statusUpdater := repository.NewStatusUpdater(gitChecker)

	// Create services
	repositoryManager := repositoryService.NewManager(configService, statusUpdater, gitChecker)
	themeManager := themeService.NewManager(configService)
	explorerManager := explorerService.NewManager(gitChecker)

	return &AppDependencies{
		configService:     configService,
		statusUpdater:     statusUpdater,
		repositoryService: repositoryManager,
		themeService:      themeManager,
		explorerService:   explorerManager,
	}
}

// GetConfigService returns the configuration service.
func (d *AppDependencies) GetConfigService() config.ConfigService {
	return d.configService
}

// GetRepositoryService returns the repository service.
func (d *AppDependencies) GetRepositoryService() repositoryService.Service {
	return d.repositoryService
}

// GetThemeService returns the theme service.
func (d *AppDependencies) GetThemeService() themeService.Service {
	return d.themeService
}

// GetExplorerService returns the explorer service.
func (d *AppDependencies) GetExplorerService() explorerService.Service {
	return d.explorerService
}

// Verify at compile time that AppDependencies implements ui.Dependencies
var _ ui.Dependencies = (*AppDependencies)(nil)
