package ui

import (
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

	// The repository manager is already initialized in dependencies

	return Model{
		Dependencies:     deps,
		Config:           cfg,
		State:            ListView,
		Cursor:           0,
		NavItemsNeedSync: true,

		// Initialize settings fields
		SettingsSection: "repositories",
		SettingsCursor:  0,

		// Initialize repository management fields
		RepoActiveSection: "list", // Start with repository list active
		RepoExplorer:      nil,    // Will be initialized when needed
		RepoPasteMode:     false,
		RepoPasteValue:    "",

		// Initialize handler instances
		KeyHandler:        NewKeyHandler(),
		NavigationHandler: NewNavigationHandler(),
		RepositoryHandler: NewRepositoryOperationHandler(),
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
