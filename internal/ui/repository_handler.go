package ui

import (
	"path/filepath"

	"github.com/charmbracelet/bubbletea"
	"github.com/jarmocluyse/wip-tui/internal/explorer"
	"github.com/jarmocluyse/wip-tui/internal/logging"
)

// RepositoryOperationHandler manages repository operations like add, remove, and toggle.
type RepositoryOperationHandler struct{}

// NewRepositoryOperationHandler creates a new RepositoryOperationHandler instance.
func NewRepositoryOperationHandler() *RepositoryOperationHandler {
	return &RepositoryOperationHandler{}
}

// ToggleRepositorySelection toggles the selected repository's managed status.
func (h *RepositoryOperationHandler) ToggleRepositorySelection(m Model) (Model, tea.Cmd) {
	logging.Get().Debug("toggleRepositorySelection called",
		"items_count", len(m.ExplorerItems),
		"cursor", m.ExplorerCursor)

	if !h.isValidSelection(m) {
		logging.Get().Debug("early return: invalid selection")
		return m, nil
	}

	selected := m.ExplorerItems[m.ExplorerCursor]

	if !selected.IsGitRepo {
		logging.Get().Debug("early return: selected item is not a git repo", "path", selected.Path)
		return m, nil
	}

	return h.processRepositoryToggle(m, selected)
}

// DeleteSelectedRepository removes the currently selected repository.
func (h *RepositoryOperationHandler) DeleteSelectedRepository(m Model) (Model, tea.Cmd) {
	repositories := m.RepoHandler.GetRepositories()

	if len(repositories) == 0 || m.Cursor >= len(repositories) {
		return m, nil
	}

	m.RepoHandler.RemoveRepository(m.Cursor)

	navigationHandler := NewNavigationHandler()
	m.Cursor = navigationHandler.AdjustCursorAfterDeletion(m)

	err := m.Dependencies.GetConfigService().Save(m.Config)
	if err != nil {
		logging.Get().Error("Failed to save config after repository deletion", "error", err)
	}

	return m, m.updateRepositoryStatuses()
}

// RemoveRepositoryByPath removes a repository by its path.
func (h *RepositoryOperationHandler) RemoveRepositoryByPath(m Model, path string) {
	m.RepoHandler.RemoveRepositoryByPath(path)
	m.NavItemsNeedSync = true
}

// isValidSelection checks if the current selection is valid.
func (h *RepositoryOperationHandler) isValidSelection(m Model) bool {
	return len(m.ExplorerItems) > 0 && m.ExplorerCursor < len(m.ExplorerItems)
}

// processRepositoryToggle handles the actual toggle logic for a repository.
func (h *RepositoryOperationHandler) processRepositoryToggle(m Model, selected explorer.Item) (Model, tea.Cmd) {
	logging.Get().Info("toggling repository selection",
		"path", selected.Path,
		"is_added", selected.IsAdded,
		"current_repos_count", len(m.RepoHandler.GetRepositories()))

	if selected.IsAdded {
		h.RemoveRepositoryByPath(m, selected.Path)
		logging.Get().Info("repository removed", "path", selected.Path)
	} else {
		name := filepath.Base(selected.Path)
		m.Config.RepositoryPaths = append(m.Config.RepositoryPaths, selected.Path)
		m.RepoHandler.AddRepository(name, selected.Path)
		m.NavItemsNeedSync = true
		logging.Get().Info("repository added", "path", selected.Path, "name", name)
	}

	err := m.Dependencies.GetConfigService().Save(m.Config)
	if err != nil {
		logging.Get().Error("Failed to save config", "error", err)
	}

	updatedModel, cmd := m.ExplorerHandler.LoadExplorerDirectory(m)
	return updatedModel, tea.Batch(cmd, m.updateRepositoryStatuses())
}
