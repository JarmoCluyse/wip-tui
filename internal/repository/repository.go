package repository

import "github.com/jarmocluyse/wip-tui/internal/git"

type Repository struct {
	Name           string
	Path           string
	AutoDiscover   bool
	HasUncommitted bool
	HasUnpushed    bool
	HasUntracked   bool
	HasError       bool
	IsWorktree     bool
	IsBare         bool
}

type NavigableItem struct {
	Type         string // "repository" or "worktree"
	Repository   *Repository
	WorktreeInfo *git.WorktreeInfo
	ParentRepo   *Repository // For worktrees, reference to parent bare repo
}

type Handler struct {
	repositories []Repository
}

func NewHandler() *Handler {
	return &Handler{
		repositories: make([]Repository, 0),
	}
}

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

func (h *Handler) GetRepositories() []Repository {
	return h.repositories
}

func (h *Handler) AddRepository(name, path string) {
	repo := Repository{
		Name: name,
		Path: path,
	}
	h.repositories = append(h.repositories, repo)
}

func (h *Handler) AddRepositoryWithAutoDiscover(name, path string, autoDiscover bool) {
	repo := Repository{
		Name:         name,
		Path:         path,
		AutoDiscover: autoDiscover,
	}
	h.repositories = append(h.repositories, repo)
}

func (h *Handler) RemoveRepository(index int) {
	if h.isValidIndex(index) {
		h.repositories = append(h.repositories[:index], h.repositories[index+1:]...)
	}
}

func (h *Handler) RemoveRepositoryByPath(path string) {
	for i, repo := range h.repositories {
		if repo.Path == path {
			h.RemoveRepository(i)
			break
		}
	}
}

func (h *Handler) GetPaths() []string {
	paths := make([]string, 0, len(h.repositories))
	for _, repo := range h.repositories {
		paths = append(paths, repo.Path)
	}
	return paths
}

func (h *Handler) isValidIndex(index int) bool {
	return index >= 0 && index < len(h.repositories)
}

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
