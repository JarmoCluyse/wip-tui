// Package explorer provides file system exploration services for the application.
package explorer

import (
	"github.com/jarmocluyse/git-dash/internal/explorer"
	"github.com/jarmocluyse/git-dash/internal/repository"
)

// Service defines the interface for explorer operations.
type Service interface {
	// Directory exploration
	ListDirectory(path string, managedRepositories []repository.Repository) ([]explorer.Item, error)
	GetParentDirectory(currentPath string) string

	// Explorer management
	GetExplorer() explorer.Explorer
}
