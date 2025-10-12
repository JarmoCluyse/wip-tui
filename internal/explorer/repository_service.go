package explorer

import (
	"github.com/jarmocluyse/wip-tui/internal/git"
	"github.com/jarmocluyse/wip-tui/internal/repository"
)

// RepositoryStatusService provides Git repository status checking operations.
type RepositoryStatusService struct {
	gitChecker git.StatusChecker
}

// NewRepositoryStatusService creates a new RepositoryStatusService with the given Git checker.
func NewRepositoryStatusService(gitChecker git.StatusChecker) *RepositoryStatusService {
	return &RepositoryStatusService{gitChecker: gitChecker}
}

// IsGitRepository returns true if the given path is a Git repository.
func (s *RepositoryStatusService) IsGitRepository(path string) bool {
	return s.gitChecker.IsGitRepository(path)
}

// IsAlreadyManaged returns true if the path is already in the managed repositories list.
func (s *RepositoryStatusService) IsAlreadyManaged(path string, managedRepositories []repository.Repository) bool {
	return isPathAlreadyManaged(path, managedRepositories)
}

// GetGitStatus returns Git status flags for uncommitted, unpushed, and untracked changes.
func (s *RepositoryStatusService) GetGitStatus(path string) (uncommitted, unpushed, untracked bool) {
	if !s.gitChecker.IsGitRepository(path) {
		return false, false, false
	}

	uncommitted = s.gitChecker.HasUncommittedChanges(path)
	unpushed = s.gitChecker.HasUnpushedCommits(path)
	untracked = s.gitChecker.HasUntrackedFiles(path)

	return uncommitted, unpushed, untracked
}

// IsBareRepository returns true if the repository at the given path is a bare repository.
func (s *RepositoryStatusService) IsBareRepository(path string) bool {
	return s.gitChecker.IsBareRepository(path)
}

// ListWorktrees returns a list of all worktrees for the given bare repository.
func (s *RepositoryStatusService) ListWorktrees(bareRepoPath string) ([]git.WorktreeInfo, error) {
	return s.gitChecker.ListWorktrees(bareRepoPath)
}
