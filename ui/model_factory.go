package ui

import (
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/jarmocluyse/git-dash/internal/config"
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
	homeDir := f.getHomeDirectory()

	// Initialize repository service with paths from config
	deps.GetRepositoryService().LoadRepositories(cfg.RepositoryPaths)

	return Model{
		Dependencies:     deps,
		Config:           cfg,
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

// getHomeDirectory returns the user's home directory or "/" as fallback.
func (f *ModelFactory) getHomeDirectory() string {
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		return "/"
	}
	return homeDir
}
