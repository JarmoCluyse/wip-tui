package explorer

import "github.com/jarmocluyse/git-dash/internal/repository"

// DirectoryExplorer provides clean filesystem navigation capabilities.
type DirectoryExplorer interface {
	// ExploreDirectory returns all filesystem entries in the specified directory path.
	ExploreDirectory(path string, managedRepositories []repository.Repository) ([]FileSystemEntry, error)

	// NavigateToParent returns the parent directory path of the given current path.
	NavigateToParent(currentPath string) string
}
