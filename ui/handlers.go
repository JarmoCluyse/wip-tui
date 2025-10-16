package ui

import (
	"github.com/charmbracelet/bubbletea"
)

// CreateInitialModel creates and returns an initial Model with default configuration.
func CreateInitialModel(deps Dependencies) Model {
	// Use the ModelFactory for clean separation of concerns
	factory := NewModelFactory()
	return factory.CreateInitialModel(deps)
}

// Init initializes the Model and returns commands to run on startup.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.updateRepositoryStatuses(),
		tea.WindowSize(), // Explicitly request window size
	)
}

// updateRepositoryStatuses creates a command that updates the status of all repositories.
// Update handles incoming messages and updates the model accordingly.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case StatusUpdateComplete:
		return m.handleStatusUpdate(msg)
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	}
	return m, nil
}

// handleKeyPress processes keyboard input and returns updated model and commands.
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Delegate to the KeyHandler for clean separation of concerns
	return m.KeyHandler.HandleKeyPress(m, msg)
}

// handleListViewKeys handles keyboard input in the main repository list view.
func (m Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Delegate to the KeyHandler for clean separation of concerns
	return m.KeyHandler.handleListViewKeys(m, msg)
}

// handleRepoManagementViewKeys handles keyboard input in the repository management view.
func (m Model) handleRepoManagementViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Delegate to the KeyHandler for clean separation of concerns
	return m.KeyHandler.handleRepoManagementViewKeys(m, msg)
}

// handleExplorerViewKeys handles keyboard input in the directory explorer view.
func (m Model) handleExplorerViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Delegate to the KeyHandler for clean separation of concerns
	return m.KeyHandler.handleExplorerViewKeys(m, msg)
}
