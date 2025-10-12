package main

import (
	"github.com/jarmocluyse/wip-tui/internal/config"
	"github.com/jarmocluyse/wip-tui/internal/git"
	"github.com/jarmocluyse/wip-tui/internal/repository"
	"github.com/jarmocluyse/wip-tui/internal/ui"
)

// AppDependencies implements the ui.Dependencies interface.
type AppDependencies struct {
	configService config.ConfigService
	statusUpdater *repository.StatusUpdater
	gitChecker    git.StatusChecker
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

	return &AppDependencies{
		configService: configService,
		statusUpdater: statusUpdater,
		gitChecker:    gitChecker,
	}
}

// GetConfigService returns the configuration service.
func (d *AppDependencies) GetConfigService() config.ConfigService {
	return d.configService
}

// GetStatusUpdater returns the repository status updater.
func (d *AppDependencies) GetStatusUpdater() *repository.StatusUpdater {
	return d.statusUpdater
}

// GetGitChecker returns the git status checker.
func (d *AppDependencies) GetGitChecker() git.StatusChecker {
	return d.gitChecker
}

// Verify at compile time that AppDependencies implements ui.Dependencies
var _ ui.Dependencies = (*AppDependencies)(nil)
