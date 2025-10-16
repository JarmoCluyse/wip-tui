// Package repository provides Git repository management and status tracking functionality.
package repository

import "github.com/jarmocluyse/git-dash/internal/git"

// Repository represents a Git repository with its status information.
type Repository struct {
	Name             string
	Path             string
	AutoDiscover     bool
	HasUncommitted   bool
	HasUnpushed      bool
	HasUntracked     bool
	HasError         bool
	IsWorktree       bool
	IsBare           bool
	UncommittedCount int // Number of files with uncommitted changes (tracked files)
	UnpushedCount    int // Number of unpushed commits
	UntrackedCount   int // Number of untracked files
}

// NavigableItem represents a repository or worktree that can be navigated to.
type NavigableItem struct {
	Type         string // "repository" or "worktree"
	Repository   *Repository
	WorktreeInfo *git.WorktreeInfo
	ParentRepo   *Repository // For worktrees, reference to parent bare repo
}

// SummaryData holds aggregated summary information for repositories and navigable items.
type SummaryData struct {
	TotalUncommitted int
	TotalUnpushed    int
	TotalUntracked   int
	TotalErrors      int
}

// Handler manages a collection of repositories.
type Handler struct {
	repositories []Repository
}

// NewHandler creates a new repository handler.
func NewHandler() *Handler {
	return &Handler{
		repositories: make([]Repository, 0),
	}
}

// SetRepositories sets the repositories from a list of paths.
func (h *Handler) SetRepositories(paths []string) {
	h.repositories = make([]Repository, 0, len(paths))
	for _, path := range paths {
		repo := Repository{
			Name: extractNameFromPath(path),
			Path: path,
		}
		h.repositories = append(h.repositories, repo)
	}
}

// GetRepositories returns the list of managed repositories.
func (h *Handler) GetRepositories() []Repository {
	return h.repositories
}

// AddRepository adds a new repository with the given name and path.
func (h *Handler) AddRepository(name, path string) {
	repo := Repository{
		Name: name,
		Path: path,
	}
	h.repositories = append(h.repositories, repo)
}

// AddRepositoryWithAutoDiscover adds a repository with auto-discovery settings.
func (h *Handler) AddRepositoryWithAutoDiscover(name, path string, autoDiscover bool) {
	repo := Repository{
		Name:         name,
		Path:         path,
		AutoDiscover: autoDiscover,
	}
	h.repositories = append(h.repositories, repo)
}

// RemoveRepository removes a repository by index.
func (h *Handler) RemoveRepository(index int) {
	if h.isValidIndex(index) {
		h.repositories = append(h.repositories[:index], h.repositories[index+1:]...)
	}
}

// RemoveRepositoryByPath removes a repository by its path.
func (h *Handler) RemoveRepositoryByPath(path string) {
	for i, repo := range h.repositories {
		if repo.Path == path {
			h.RemoveRepository(i)
			break
		}
	}
}

// GetPaths returns the paths of all managed repositories.
func (h *Handler) GetPaths() []string {
	paths := make([]string, 0, len(h.repositories))
	for _, repo := range h.repositories {
		paths = append(paths, repo.Path)
	}
	return paths
}

// isValidIndex checks if the given index is valid for the repositories slice.
func (h *Handler) isValidIndex(index int) bool {
	return index >= 0 && index < len(h.repositories)
}

// extractNameFromPath extracts a repository name from its path.
func extractNameFromPath(path string) string {
	// Extract the last component of the path as the name
	if path == "" {
		return ""
	}

	// Remove trailing slashes
	for len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	// Find the last slash
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}

	return path
}
