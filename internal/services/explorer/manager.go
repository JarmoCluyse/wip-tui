// Package explorer provides file system exploration services for the application.
package explorer

import (
	"github.com/jarmocluyse/git-dash/internal/explorer"
	"github.com/jarmocluyse/git-dash/internal/git"
	"github.com/jarmocluyse/git-dash/internal/repository"
)

// Manager implements the explorer service interface.
type Manager struct {
	explorer   explorer.Explorer
	gitChecker git.StatusChecker
}

// NewManager creates a new explorer manager instance.
func NewManager(gitChecker git.StatusChecker) *Manager {
	explorerInstance := explorer.New(gitChecker, nil)

	return &Manager{
		explorer:   explorerInstance,
		gitChecker: gitChecker,
	}
}

// ListDirectory explores a directory and returns its filesystem entries.
func (m *Manager) ListDirectory(path string, managedRepositories []repository.Repository) ([]explorer.Item, error) {
	return m.explorer.ListDirectory(path, managedRepositories)
}

// GetParentDirectory returns the parent directory path.
func (m *Manager) GetParentDirectory(currentPath string) string {
	return m.explorer.GetParentDirectory(currentPath)
}

// GetExplorer returns the underlying explorer instance.
func (m *Manager) GetExplorer() explorer.Explorer {
	return m.explorer
}
