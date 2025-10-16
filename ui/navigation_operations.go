package ui

import (
	"github.com/jarmocluyse/git-dash/ui/layout"
	"github.com/jarmocluyse/git-dash/ui/types"
)

// getVisibleItemCount calculates how many items can be visible based on terminal height.
func (m Model) getVisibleItemCount() int {
	// Use the height calculator to determine content area
	calc := layout.NewHeightCalculator()

	// Reserve space for help (1 line) and any header lines (varies by page)
	// For home page, typically 4 header lines
	headerLines := 4
	helpLines := 1

	contentHeight, _ := calc.CalculateContentAreaHeight(m.Height, headerLines+helpLines)
	if contentHeight < 3 {
		contentHeight = 15 // Fallback minimum for reasonable viewing
	}

	itemsPerScreen := contentHeight // Each borderless item takes ~1 line
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

// rebuildNavigableItems rebuilds the cached navigable items using the repository manager.
func (m *Model) rebuildNavigableItems() {
	if err := m.Dependencies.GetRepoManager().ReloadWorktrees(); err != nil {
		// Handle error if needed, for now just continue
		return
	}

	// Get the repository items and build navigable items
	repoItems := m.Dependencies.GetRepoManager().GetItems()

	var items []types.NavigableItem
	for _, repoItem := range repoItems {
		// Add main repository as navigable item
		repoNavItem := types.NavigableItem{
			Type:       "repository",
			Repository: repoItem,
		}
		items = append(items, repoNavItem)

		// Add worktrees as separate navigable items
		for i := range repoItem.SubItems {
			subItem := repoItem.SubItems[i]

			worktreeNavItem := types.NavigableItem{
				Type:         "worktree",
				WorktreeInfo: subItem,
				ParentRepo:   repoItem,
			}
			items = append(items, worktreeNavItem)
		}
	}

	m.CachedNavItems = items
}
