package ui

import (
	"github.com/jarmocluyse/wip-tui/internal/git"
	"github.com/jarmocluyse/wip-tui/internal/repository"
	"github.com/jarmocluyse/wip-tui/internal/ui/types"
)

// getVisibleItemCount calculates how many items can be visible based on terminal height.
func (m Model) getVisibleItemCount() int {
	// Reserve space for header, help text, and some padding
	// Each repository item now takes approximately 1 line (without border)
	availableHeight := m.Height - 6 // Reserve space for header and help
	if availableHeight < 3 {
		availableHeight = 15 // Fallback minimum for reasonable viewing
	}
	itemsPerScreen := availableHeight // Each borderless item takes ~1 line
	if itemsPerScreen < 5 {
		itemsPerScreen = 10 // Minimum reasonable number of items
	}
	return itemsPerScreen
}

// getNavigableItems returns cached navigable items or rebuilds if needed.
func (m *Model) getNavigableItems() []types.NavigableItem {
	if m.CachedNavItems == nil || m.NavItemsNeedSync {
		m.rebuildNavigableItems()
		m.NavItemsNeedSync = false
	}
	return m.CachedNavItems
}

// rebuildNavigableItems rebuilds the cached navigable items with concurrency.
func (m *Model) rebuildNavigableItems() {
	repositories := m.RepoHandler.GetRepositories()
	gitChecker := m.Dependencies.GetGitChecker()

	// Use channels to collect results
	type repoResult struct {
		index int
		items []types.NavigableItem
	}

	resultChan := make(chan repoResult, len(repositories))
	semaphore := make(chan struct{}, 8) // Limit to 8 concurrent operations

	// Process each repository concurrently
	for i := range repositories {
		go func(idx int, repo *repository.Repository) {
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			var repoItems []types.NavigableItem

			// Add the repository itself
			repoItems = append(repoItems, types.NavigableItem{
				Type:       "repository",
				Repository: repo,
			})

			// If it's a bare repo, add its worktrees
			if repo.IsBare {
				worktrees, err := gitChecker.ListWorktrees(repo.Path)
				if err == nil {
					var validWorktrees []git.WorktreeInfo
					for _, wt := range worktrees {
						if !wt.Bare && wt.Path != repo.Path {
							validWorktrees = append(validWorktrees, wt)
						}
					}

					// Add worktrees with proper IsLast flag
					for j, wt := range validWorktrees {
						isLast := j == len(validWorktrees)-1
						repoItems = append(repoItems, types.NavigableItem{
							Type:         "worktree",
							WorktreeInfo: &wt,
							ParentRepo:   repo,
							IsLast:       isLast,
						})
					}
				}
			}

			resultChan <- repoResult{index: idx, items: repoItems}
		}(i, &repositories[i])
	}

	// Collect results and maintain order
	results := make([][]types.NavigableItem, len(repositories))
	for i := 0; i < len(repositories); i++ {
		result := <-resultChan
		results[result.index] = result.items
	}

	// Flatten results in correct order
	var items []types.NavigableItem
	for _, repoItems := range results {
		items = append(items, repoItems...)
	}

	m.CachedNavItems = items
}

// buildNavigableItems creates a list of navigable items from repositories and their worktrees.
func (m Model) buildNavigableItems() []types.NavigableItem {
	var items []types.NavigableItem
	gitChecker := m.Dependencies.GetGitChecker() // Use the cached git checker instead of creating new
	repositories := m.RepoHandler.GetRepositories()

	for i := range repositories {
		repo := &repositories[i]

		// Add the repository itself
		items = append(items, types.NavigableItem{
			Type:       "repository",
			Repository: repo,
		})

		// If it's a bare repo, add its worktrees
		if repo.IsBare {
			worktrees, err := gitChecker.ListWorktrees(repo.Path)
			if err == nil {
				var validWorktrees []git.WorktreeInfo
				for _, wt := range worktrees {
					if !wt.Bare && wt.Path != repo.Path {
						validWorktrees = append(validWorktrees, wt)
					}
				}

				// Add worktrees with proper IsLast flag
				for j, wt := range validWorktrees {
					isLast := j == len(validWorktrees)-1
					items = append(items, types.NavigableItem{
						Type:         "worktree",
						WorktreeInfo: &wt,
						ParentRepo:   repo,
						IsLast:       isLast,
					})
				}
			}
		}
	}

	return items
}
