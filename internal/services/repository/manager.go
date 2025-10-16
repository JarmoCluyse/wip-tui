// Package repository provides repository management services for the application.
package repository

import (
	"path/filepath"

	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/git"
	"github.com/jarmocluyse/git-dash/internal/repository"
)

// Manager implements the repository service interface.
type Manager struct {
	repositories   []repository.Repository
	navigableItems []repository.NavigableItem
	statusUpdater  *repository.StatusUpdater
	configService  config.ConfigService
	gitChecker     git.StatusChecker
}

// NewManager creates a new repository manager instance.
func NewManager(configService config.ConfigService, statusUpdater *repository.StatusUpdater, gitChecker git.StatusChecker) *Manager {
	return &Manager{
		repositories:   make([]repository.Repository, 0),
		navigableItems: make([]repository.NavigableItem, 0),
		statusUpdater:  statusUpdater,
		configService:  configService,
		gitChecker:     gitChecker,
	}
}

// LoadRepositories loads repositories from the provided paths.
func (m *Manager) LoadRepositories(paths []string) error {
	m.repositories = make([]repository.Repository, 0, len(paths))

	for _, path := range paths {
		repo := repository.Repository{
			Name: extractNameFromPath(path),
			Path: path,
		}
		m.repositories = append(m.repositories, repo)
	}

	return nil
}

// GetRepositories returns all managed repositories.
func (m *Manager) GetRepositories() []repository.Repository {
	return append([]repository.Repository(nil), m.repositories...)
}

// AddRepository adds a new repository to the manager.
func (m *Manager) AddRepository(name, path string) error {
	// Check if repository already exists
	if m.ContainsRepository(path) {
		return nil // Already exists, no error
	}

	repo := repository.Repository{
		Name: name,
		Path: path,
	}

	m.repositories = append(m.repositories, repo)

	// Update configuration
	if m.configService != nil {
		config, err := m.configService.Load()
		if err != nil {
			return err
		}

		config.AddRepositoryPath(path)
		return m.configService.Save(config)
	}

	return nil
}

// RemoveRepository removes a repository by index.
func (m *Manager) RemoveRepository(index int) error {
	if !m.IsValidRepositoryIndex(index) {
		return nil // Invalid index, no error
	}

	// Get the path before removing
	path := m.repositories[index].Path

	// Remove from repositories slice
	m.repositories = append(m.repositories[:index], m.repositories[index+1:]...)

	// Update configuration
	if m.configService != nil {
		config, err := m.configService.Load()
		if err != nil {
			return err
		}

		config.RemoveRepositoryPathByValue(path)
		return m.configService.Save(config)
	}

	return nil
}

// RemoveRepositoryByPath removes a repository by its path.
func (m *Manager) RemoveRepositoryByPath(path string) error {
	for i, repo := range m.repositories {
		if repo.Path == path {
			return m.RemoveRepository(i)
		}
	}
	return nil // Not found, no error
}

// GetRepositoryPaths returns all repository paths.
func (m *Manager) GetRepositoryPaths() []string {
	paths := make([]string, len(m.repositories))
	for i, repo := range m.repositories {
		paths[i] = repo.Path
	}
	return paths
}

// UpdateRepositoryStatus updates the status of a specific repository.
func (m *Manager) UpdateRepositoryStatus(index int) error {
	if !m.IsValidRepositoryIndex(index) {
		return nil
	}

	if m.statusUpdater == nil {
		return nil
	}

	m.statusUpdater.UpdateStatus(&m.repositories[index])
	return nil
}

// UpdateAllRepositoryStatuses updates the status of all repositories.
func (m *Manager) UpdateAllRepositoryStatuses() error {
	if m.statusUpdater == nil {
		return nil
	}

	for i := range m.repositories {
		m.statusUpdater.UpdateStatus(&m.repositories[i])
	}
	return nil
}

// GetNavigableItems returns all navigable items (repositories and worktrees).
func (m *Manager) GetNavigableItems() ([]repository.NavigableItem, error) {
	if len(m.navigableItems) == 0 {
		if err := m.RefreshNavigableItems(); err != nil {
			return nil, err
		}
	}
	return append([]repository.NavigableItem(nil), m.navigableItems...), nil
}

// RefreshNavigableItems refreshes the cache of navigable items.
func (m *Manager) RefreshNavigableItems() error {
	m.navigableItems = make([]repository.NavigableItem, 0)

	for i := range m.repositories {
		repo := &m.repositories[i]

		// Update repository status
		if m.statusUpdater != nil {
			m.statusUpdater.UpdateStatus(repo)
		}

		// Add main repository
		m.navigableItems = append(m.navigableItems, repository.NavigableItem{
			Type:       "repository",
			Repository: repo,
		})

		// Add worktrees if this is a bare repository
		if repo.IsBare && m.gitChecker != nil {
			worktrees, err := m.gitChecker.ListWorktrees(repo.Path)
			if err == nil {
				for _, wt := range worktrees {
					m.navigableItems = append(m.navigableItems, repository.NavigableItem{
						Type:         "worktree",
						WorktreeInfo: &wt,
						ParentRepo:   repo,
					})
				}
			}
		}
	}

	return nil
}

// GetRepositoryByIndex returns a repository by its index.
func (m *Manager) GetRepositoryByIndex(index int) (*repository.Repository, error) {
	if !m.IsValidRepositoryIndex(index) {
		return nil, nil
	}
	return &m.repositories[index], nil
}

// GetRepositoryByPath returns a repository by its path.
func (m *Manager) GetRepositoryByPath(path string) (*repository.Repository, error) {
	for i := range m.repositories {
		if m.repositories[i].Path == path {
			return &m.repositories[i], nil
		}
	}
	return nil, nil
}

// GetRepositoryCount returns the number of managed repositories.
func (m *Manager) GetRepositoryCount() int {
	return len(m.repositories)
}

// IsValidRepositoryIndex checks if the index is valid.
func (m *Manager) IsValidRepositoryIndex(index int) bool {
	return index >= 0 && index < len(m.repositories)
}

// ContainsRepository checks if a repository with the given path exists.
func (m *Manager) ContainsRepository(path string) bool {
	for _, repo := range m.repositories {
		if repo.Path == path {
			return true
		}
	}
	return false
}

// GetGitChecker returns the git status checker for rendering purposes.
func (m *Manager) GetGitChecker() git.StatusChecker {
	return m.gitChecker
}

// GetRepositorySummary calculates and returns summary data for all repositories.
func (m *Manager) GetRepositorySummary() repository.SummaryData {
	var data repository.SummaryData

	for _, repo := range m.repositories {
		if repo.HasUncommitted {
			data.TotalUncommitted += repo.UncommittedCount
		}
		if repo.HasUnpushed {
			data.TotalUnpushed += repo.UnpushedCount
		}
		if repo.HasUntracked {
			data.TotalUntracked += repo.UntrackedCount
		}
		if repo.HasError {
			data.TotalErrors++
		}
	}

	return data
}

// GetNavigableItemSummary calculates and returns summary data for all navigable items (repositories + worktrees).
func (m *Manager) GetNavigableItemSummary() repository.SummaryData {
	var data repository.SummaryData

	for _, item := range m.navigableItems {
		switch item.Type {
		case "repository":
			if item.Repository.HasUncommitted {
				data.TotalUncommitted += item.Repository.UncommittedCount
			}
			if item.Repository.HasUnpushed {
				data.TotalUnpushed += item.Repository.UnpushedCount
			}
			if item.Repository.HasUntracked {
				data.TotalUntracked += item.Repository.UntrackedCount
			}
			if item.Repository.HasError {
				data.TotalErrors++
			}
		case "worktree":
			// Calculate worktree status using git checker
			if item.WorktreeInfo != nil {
				uncommittedCount := m.gitChecker.CountUncommittedChanges(item.WorktreeInfo.Path)
				if uncommittedCount > 0 {
					data.TotalUncommitted += uncommittedCount
				}

				unpushedCount := m.gitChecker.CountUnpushedCommits(item.WorktreeInfo.Path)
				if unpushedCount > 0 {
					data.TotalUnpushed += unpushedCount
				}

				untrackedCount := m.gitChecker.CountUntrackedFiles(item.WorktreeInfo.Path)
				if untrackedCount > 0 {
					data.TotalUntracked += untrackedCount
				}

				// Check if there are any git errors for this worktree
				if !m.gitChecker.IsGitRepository(item.WorktreeInfo.Path) {
					data.TotalErrors++
				}
			}
		}
	}

	return data
}

// extractNameFromPath extracts a repository name from its path.
func extractNameFromPath(path string) string {
	if path == "" {
		return ""
	}

	// Remove trailing slashes
	for len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	return filepath.Base(path)
}
