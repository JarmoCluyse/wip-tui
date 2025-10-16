package repository

import (
	"runtime"
	"sync"

	"github.com/jarmocluyse/git-dash/internal/git"
)

// StatusUpdater handles updating Git repository status information.
type StatusUpdater struct {
	gitChecker git.StatusChecker
}

// NewStatusUpdater creates a new status updater with the given Git checker.
func NewStatusUpdater(gitChecker git.StatusChecker) *StatusUpdater {
	return &StatusUpdater{
		gitChecker: gitChecker,
	}
}

// UpdateStatus updates the Git status information for a single repository.
func (s *StatusUpdater) UpdateStatus(repo *Repository) {
	if !s.gitChecker.IsGitRepository(repo.Path) {
		s.setErrorStatus(repo)
		return
	}

	repo.IsBare = s.gitChecker.IsBareRepository(repo.Path)
	repo.IsWorktree = s.gitChecker.IsWorktree(repo.Path)
	repo.HasError = false

	if repo.IsBare {
		s.updateBareRepositoryStatus(repo)
	} else {
		s.updateRegularRepositoryStatus(repo)
	}
}

// UpdateRepositories updates status for multiple repositories concurrently using a worker pool.
func (s *StatusUpdater) UpdateRepositories(repositories []Repository) {
	// Use a worker pool to process repositories concurrently
	numWorkers := runtime.NumCPU()
	if len(repositories) < numWorkers {
		numWorkers = len(repositories)
	}

	// Channel for sending repositories to workers
	repoChan := make(chan int, len(repositories))

	// WaitGroup to wait for all workers to complete
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repoIndex := range repoChan {
				s.UpdateStatus(&repositories[repoIndex])
			}
		}()
	}

	// Send repository indices to workers
	for i := range repositories {
		repoChan <- i
	}
	close(repoChan)

	// Wait for all workers to complete
	wg.Wait()
}

// updateBareRepositoryStatus updates status information for bare repositories.
func (s *StatusUpdater) updateBareRepositoryStatus(repo *Repository) {
	repo.HasUncommitted = false
	repo.HasUntracked = false
	repo.UncommittedCount = 0
	repo.UnpushedCount = 0
	repo.UntrackedCount = 0

	worktrees, err := s.gitChecker.ListWorktrees(repo.Path)
	if err != nil {
		repo.HasUnpushed = false
		return
	}

	repo.HasUnpushed = len(worktrees) > 0
}

// updateRegularRepositoryStatus updates status information for regular repositories.
func (s *StatusUpdater) updateRegularRepositoryStatus(repo *Repository) {
	repo.HasUncommitted = s.gitChecker.HasUncommittedChanges(repo.Path)
	repo.HasUnpushed = s.gitChecker.HasUnpushedCommits(repo.Path)
	repo.HasUntracked = s.gitChecker.HasUntrackedFiles(repo.Path)

	// Get counts for display
	repo.UncommittedCount = s.gitChecker.CountUncommittedChanges(repo.Path)
	repo.UnpushedCount = s.gitChecker.CountUnpushedCommits(repo.Path)
	repo.UntrackedCount = s.gitChecker.CountUntrackedFiles(repo.Path)
}

// setCleanStatus sets all status flags to false indicating a clean repository.
func (s *StatusUpdater) setCleanStatus(repo *Repository) {
	repo.HasUncommitted = false
	repo.HasUnpushed = false
	repo.HasUntracked = false
	repo.HasError = false
	repo.UncommittedCount = 0
	repo.UnpushedCount = 0
	repo.UntrackedCount = 0
}

// setErrorStatus sets the repository to error state, clearing other status flags.
func (s *StatusUpdater) setErrorStatus(repo *Repository) {
	repo.HasUncommitted = false
	repo.HasUnpushed = false
	repo.HasUntracked = false
	repo.HasError = true
	repo.UncommittedCount = 0
	repo.UnpushedCount = 0
	repo.UntrackedCount = 0
}
