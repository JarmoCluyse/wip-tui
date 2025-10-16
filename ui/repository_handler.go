package ui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/jarmocluyse/git-dash/internal/logging"
)

// RepositoryOperationHandler manages repository operations like add, remove, and toggle.
// Note: Explorer functionality temporarily disabled during refactoring
type RepositoryOperationHandler struct{}

// NewRepositoryOperationHandler creates a new RepositoryOperationHandler instance.
func NewRepositoryOperationHandler() *RepositoryOperationHandler {
	return &RepositoryOperationHandler{}
}

// ToggleRepositorySelection toggles the selected repository's managed status.
// Note: Explorer functionality temporarily disabled
func (h *RepositoryOperationHandler) ToggleRepositorySelection(m Model) (Model, tea.Cmd) {
	logging.Get().Debug("toggleRepositorySelection temporarily disabled")
	return m, nil
}

// DeleteSelectedRepository removes the currently selected repository.
func (h *RepositoryOperationHandler) DeleteSelectedRepository(m Model) (Model, tea.Cmd) {
	items := m.Dependencies.GetRepoManager().GetItems()

	if len(items) == 0 || m.Cursor >= len(items) {
		return m, nil
	}

	// Remove by path instead of index
	selectedPath := items[m.Cursor].Path
	m.Dependencies.GetRepoManager().RemoveRepo(selectedPath)

	navigationHandler := NewNavigationHandler()
	m.Cursor = navigationHandler.AdjustCursorAfterDeletion(m)

	return m, m.updateRepositoryStatuses()
}

// AddRepository adds a new repository from the given path.
func (h *RepositoryOperationHandler) AddRepository(m Model, path string) error {
	// Add the repository using the repository manager
	m.Dependencies.GetRepoManager().AddRepo(path)
	m.NavItemsNeedSync = true
	return nil
}

// RemoveRepositoryByPath removes a repository by its path.
func (h *RepositoryOperationHandler) RemoveRepositoryByPath(m Model, path string) {
	m.Dependencies.GetRepoManager().RemoveRepo(path)
	m.NavItemsNeedSync = true
}

// isValidSelection checks if the current selection is valid.
// Note: Explorer functionality temporarily disabled
func (h *RepositoryOperationHandler) isValidSelection(m Model) bool {
	return false
}

// processRepositoryToggle handles the actual toggle logic for a repository.
// Note: Explorer functionality temporarily disabled
func (h *RepositoryOperationHandler) processRepositoryToggle(m Model, selectedPath string) (Model, tea.Cmd) {
	logging.Get().Debug("processRepositoryToggle temporarily disabled")
	return m, nil
}
