package ui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/jarmocluyse/git-dash/internal/explorer"
	"github.com/jarmocluyse/git-dash/internal/logging"
)

// ExplorerHandler manages filesystem exploration operations.
type ExplorerHandler struct{}

// NewExplorerHandler creates a new ExplorerHandler instance.
func NewExplorerHandler() *ExplorerHandler {
	return &ExplorerHandler{}
}

// LoadExplorerDirectory loads the current explorer directory contents.
func (h *ExplorerHandler) LoadExplorerDirectory(m Model) (Model, tea.Cmd) {
	explorer := h.createDirectoryExplorer(m)
	repositories := m.Dependencies.GetRepositoryService().GetRepositories()
	items, err := explorer.ListDirectory(m.ExplorerPath, repositories)
	if err != nil {
		return m, nil
	}

	logging.Get().Debug("loading explorer directory",
		"path", m.ExplorerPath,
		"repositories_count", len(repositories),
		"items_found", len(items))

	for i, item := range items {
		if item.IsGitRepo {
			logging.Get().Debug("found git repository",
				"index", i,
				"name", item.Name,
				"path", item.Path,
				"is_added", item.IsAdded)
		}
	}

	m.ExplorerItems = items
	if m.ExplorerCursor >= len(items) {
		m.ExplorerCursor = max(0, len(items)-1)
	}

	return m, nil
}

// HandleExplorerSelection processes selection in explorer mode.
func (h *ExplorerHandler) HandleExplorerSelection(m Model) (Model, tea.Cmd) {
	if len(m.ExplorerItems) == 0 || m.ExplorerCursor >= len(m.ExplorerItems) {
		return m, nil
	}

	selected := m.ExplorerItems[m.ExplorerCursor]

	if selected.IsDirectory {
		m.ExplorerPath = selected.Path
		return h.LoadExplorerDirectory(m)
	}

	return m, nil
}

// NavigateToSelected navigates to the selected repository from the main list.
func (h *ExplorerHandler) NavigateToSelected(m Model) (Model, tea.Cmd) {
	repositories := m.Dependencies.GetRepositoryService().GetRepositories()
	if len(repositories) == 0 || m.Cursor >= len(repositories) {
		return m, nil
	}

	selectedRepo := repositories[m.Cursor]
	m.ExplorerPath = selectedRepo.Path
	m.State = ExplorerView
	m.ExplorerCursor = 0

	return h.LoadExplorerDirectory(m)
}

// EnterExplorerMode switches to explorer mode.
func (h *ExplorerHandler) EnterExplorerMode(m Model) (Model, tea.Cmd) {
	m.State = ExplorerView
	return h.LoadExplorerDirectory(m)
}

// createDirectoryExplorer creates an explorer instance for directory operations.
func (h *ExplorerHandler) createDirectoryExplorer(m Model) explorer.Explorer {
	return m.Dependencies.GetExplorerService().GetExplorer()
}
