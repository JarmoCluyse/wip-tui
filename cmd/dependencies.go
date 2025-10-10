package main

import (
	"github.com/jarmocluyse/wip-tui/internal/config"
	"github.com/jarmocluyse/wip-tui/internal/git"
	"github.com/jarmocluyse/wip-tui/internal/repository"
	"github.com/jarmocluyse/wip-tui/internal/ui"
)

// AppDependencies implements the ui.Dependencies interface
type AppDependencies struct {
	configService config.ConfigService
	statusUpdater *repository.StatusUpdater
	gitChecker    git.StatusChecker
}

func NewAppDependencies(configPath string) *AppDependencies {
	var configService config.ConfigService
	if configPath != "" {
		configService = config.NewFileConfigServiceWithPath(configPath)
	} else {
		configService = config.NewFileConfigService()
	}

	gitChecker := git.NewChecker()
	statusUpdater := repository.NewStatusUpdater(gitChecker)

	return &AppDependencies{
		configService: configService,
		statusUpdater: statusUpdater,
		gitChecker:    gitChecker,
	}
}

func (d *AppDependencies) GetConfigService() config.ConfigService {
	return d.configService
}

func (d *AppDependencies) GetStatusUpdater() *repository.StatusUpdater {
	return d.statusUpdater
}

func (d *AppDependencies) GetGitChecker() git.StatusChecker {
	return d.gitChecker
}

// Verify at compile time that AppDependencies implements ui.Dependencies
var _ ui.Dependencies = (*AppDependencies)(nil)
