package ui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/jarmocluyse/wip-tui/internal/repository"
)

// updateRepositoryStatuses initiates an asynchronous update of all repository statuses.
func (m Model) updateRepositoryStatuses() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		repositories := m.RepoHandler.GetRepositories()
		updatedRepos := make([]repository.Repository, len(repositories))
		copy(updatedRepos, repositories)

		// Use concurrent status updating
		m.Dependencies.GetStatusUpdater().UpdateRepositories(updatedRepos)

		return StatusMessage{Repositories: updatedRepos}
	})
}

// handleStatusUpdate processes repository status updates and updates the model.
func (m Model) handleStatusUpdate(msg StatusMessage) (tea.Model, tea.Cmd) {
	// Update the repository handler with the updated repositories
	updatedPaths := make([]string, len(msg.Repositories))
	for i, repo := range msg.Repositories {
		updatedPaths[i] = repo.Path
	}
	m.RepoHandler.SetRepositories(updatedPaths)

	// Copy the repository states back to the handler
	for i, repo := range msg.Repositories {
		if i < len(m.RepoHandler.GetRepositories()) {
			repos := m.RepoHandler.GetRepositories()
			repos[i] = repo
		}
	}

	// Mark navigable items cache as needing sync
	m.NavItemsNeedSync = true

	return m, nil
}
