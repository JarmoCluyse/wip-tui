package ui

import (
	"github.com/jarmocluyse/git-dash/ui/types"
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

// rebuildNavigableItems rebuilds the cached navigable items using the repository service.
func (m *Model) rebuildNavigableItems() {
	if err := m.Dependencies.GetRepositoryService().RefreshNavigableItems(); err != nil {
		// Handle error if needed, for now just continue
		return
	}

	// Get the navigable items from the service and convert to UI types
	navItems, err := m.Dependencies.GetRepositoryService().GetNavigableItems()
	if err != nil {
		return
	}

	// Convert from repository.NavigableItem to types.NavigableItem
	var items []types.NavigableItem
	for _, item := range navItems {
		uiItem := types.NavigableItem{
			Type:         item.Type,
			Repository:   item.Repository,
			WorktreeInfo: item.WorktreeInfo,
			ParentRepo:   item.ParentRepo,
		}
		items = append(items, uiItem)
	}

	m.CachedNavItems = items
}
