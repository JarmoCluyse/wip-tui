package ui

import (
	"github.com/charmbracelet/bubbletea"
)

// StatusUpdateComplete indicates that status updates have finished.
type StatusUpdateComplete struct{}

// updateRepositoryStatuses initiates an asynchronous update of all repository statuses.
func (m Model) updateRepositoryStatuses() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Use repository service to update all statuses
		m.Dependencies.GetRepositoryService().UpdateAllRepositoryStatuses()

		// Return a simple message indicating status update is complete
		return StatusUpdateComplete{}
	})
}

// handleStatusUpdate processes repository status updates and updates the model.
func (m Model) handleStatusUpdate(msg StatusUpdateComplete) (tea.Model, tea.Cmd) {
	// Repository service now handles all status updates internally
	// Just mark navigable items cache as needing sync
	m.NavItemsNeedSync = true

	return m, nil
}
