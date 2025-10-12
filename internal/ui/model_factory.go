package ui

import (
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/jarmocluyse/wip-tui/internal/config"
	"github.com/jarmocluyse/wip-tui/internal/repository"
)

// ModelFactory handles creation and initialization of UI models.
type ModelFactory struct{}

// NewModelFactory creates a new ModelFactory instance.
func NewModelFactory() *ModelFactory {
	return &ModelFactory{}
}

// CreateInitialModel creates and returns an initial Model with default configuration.
func (f *ModelFactory) CreateInitialModel(deps Dependencies) Model {
	cfg := f.loadConfiguration(deps)
	repoHandler := f.createRepositoryHandler(cfg)
	homeDir := f.getHomeDirectory()

	return Model{
		Dependencies:     deps,
		Config:           cfg,
		RepoHandler:      repoHandler,
		State:            ListView,
		Cursor:           0,
		ExplorerPath:     homeDir,
		ExplorerCursor:   0,
		NavItemsNeedSync: true,

		// Initialize handler instances
		KeyHandler:        NewKeyHandler(),
		NavigationHandler: NewNavigationHandler(),
		RepositoryHandler: NewRepositoryOperationHandler(),
		ExplorerHandler:   NewExplorerHandler(),
	}
}

// InitializeModel initializes the Model and returns commands to run on startup.
func (f *ModelFactory) InitializeModel(m Model) tea.Cmd {
	return m.updateRepositoryStatuses()
}

// loadConfiguration loads the application configuration or returns default config.
func (f *ModelFactory) loadConfiguration(deps Dependencies) *config.Config {
	cfg, err := deps.GetConfigService().Load()
	if err != nil {
		return &config.Config{RepositoryPaths: []string{}}
	}
	return cfg
}

// createRepositoryHandler creates and initializes a repository handler.
func (f *ModelFactory) createRepositoryHandler(cfg *config.Config) *repository.Handler {
	repoHandler := repository.NewHandler()
	repoHandler.SetRepositories(cfg.RepositoryPaths)
	return repoHandler
}

// getHomeDirectory returns the user's home directory or "/" as fallback.
func (f *ModelFactory) getHomeDirectory() string {
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		return "/"
	}
	return homeDir
}
