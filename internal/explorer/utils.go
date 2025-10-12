package explorer

import (
	"path/filepath"
	"strings"

	"github.com/jarmocluyse/wip-tui/internal/repository"
)

// shouldSkipHiddenEntry determines if an entry should be skipped because it's hidden.
func shouldSkipHiddenEntry(entryName string) bool {
	return strings.HasPrefix(entryName, ".")
}

// isPathAlreadyManaged checks if a path is already managed by any repository.
func isPathAlreadyManaged(targetPath string, managedRepositories []repository.Repository) bool {
	cleanTargetPath := filepath.Clean(targetPath)

	for _, repo := range managedRepositories {
		cleanRepoPath := filepath.Clean(repo.Path)
		if cleanRepoPath == cleanTargetPath {
			return true
		}
	}
	return false
}

// createParentDirectoryPath returns the parent directory path for a given path.
func createParentDirectoryPath(currentPath string) string {
	parent := filepath.Dir(currentPath)
	if parent == currentPath {
		return "/"
	}
	return parent
}

// createWorktreeDisplayName creates a display name for a git worktree entry.
func createWorktreeDisplayName(bareRepoPath, branchName string) string {
	baseName := filepath.Base(bareRepoPath)
	return "  ├─ " + baseName + "-" + branchName
}
