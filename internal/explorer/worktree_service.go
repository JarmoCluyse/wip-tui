package explorer

import (
	"github.com/jarmocluyse/wip-tui/internal/git"
	"github.com/jarmocluyse/wip-tui/internal/repository"
)

// WorktreeService handles git worktree operations and entry creation.
type WorktreeService struct {
	repositoryService *RepositoryStatusService
}

// NewWorktreeService creates a new WorktreeService instance.
func NewWorktreeService(repositoryService *RepositoryStatusService) *WorktreeService {
	return &WorktreeService{repositoryService: repositoryService}
}

// CreateWorktreeEntries creates filesystem entries for git worktrees in a bare repository.
func (s *WorktreeService) CreateWorktreeEntries(bareRepoPath string, managedRepositories []repository.Repository) ([]FileSystemEntry, error) {
	worktrees, err := s.repositoryService.ListWorktrees(bareRepoPath)
	if err != nil {
		return nil, err
	}

	var entries []FileSystemEntry
	for _, worktree := range worktrees {
		if s.shouldIncludeWorktree(worktree, bareRepoPath) {
			entry := s.createWorktreeEntry(worktree, bareRepoPath, managedRepositories)
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// shouldIncludeWorktree determines if a worktree should be included in the results.
func (s *WorktreeService) shouldIncludeWorktree(worktree git.WorktreeInfo, bareRepoPath string) bool {
	return !worktree.Bare && worktree.Path != bareRepoPath
}

// createWorktreeEntry creates a FileSystemEntry for a git worktree.
func (s *WorktreeService) createWorktreeEntry(worktree git.WorktreeInfo, bareRepoPath string, managedRepositories []repository.Repository) FileSystemEntry {
	displayName := createWorktreeDisplayName(bareRepoPath, worktree.Branch)
	repoInfo := s.buildWorktreeRepositoryInfo(worktree, managedRepositories)

	return NewFileSystemEntry(displayName, worktree.Path, true, repoInfo)
}

// buildWorktreeRepositoryInfo creates a RepositoryInfo for a git worktree.
func (s *WorktreeService) buildWorktreeRepositoryInfo(worktree git.WorktreeInfo, managedRepositories []repository.Repository) RepositoryInfo {
	isManaged := s.repositoryService.IsAlreadyManaged(worktree.Path, managedRepositories)
	uncommitted, unpushed, untracked := s.repositoryService.GetGitStatus(worktree.Path)
	hasError := !s.repositoryService.IsGitRepository(worktree.Path)

	return NewRepositoryInfo().
		WithGitRepository(true).
		WithManagedStatus(isManaged).
		WithWorktree(true, &worktree).
		WithGitStatus(uncommitted, unpushed, untracked).
		WithError(hasError)
}
