package ui

// NavigationHandler manages cursor movement and navigation operations.
type NavigationHandler struct{}

// NewNavigationHandler creates a new NavigationHandler instance.
func NewNavigationHandler() *NavigationHandler {
	return &NavigationHandler{}
}

// MoveCursorUp moves the cursor up and updates scroll offset if needed.
func (h *NavigationHandler) MoveCursorUp(m Model) Model {
	if m.Cursor > 0 {
		m.Cursor--
		if m.Cursor < m.ScrollOffset {
			m.ScrollOffset = m.Cursor
		}
	}
	return m
}

// MoveCursorDown moves the cursor down and updates scroll offset if needed.
func (h *NavigationHandler) MoveCursorDown(m Model) Model {
	navigableItems := m.getNavigableItems()
	if m.Cursor < len(navigableItems)-1 {
		m.Cursor++
		visibleItems := h.GetVisibleItemCount(m)
		if m.Cursor >= m.ScrollOffset+visibleItems {
			m.ScrollOffset = m.Cursor - visibleItems + 1
		}
	}
	return m
}

// MoveExplorerCursorUp moves the explorer cursor up.
func (h *NavigationHandler) MoveExplorerCursorUp(m Model) Model {
	if m.ExplorerCursor > 0 {
		m.ExplorerCursor--
	}
	return m
}

// MoveExplorerCursorDown moves the explorer cursor down.
func (h *NavigationHandler) MoveExplorerCursorDown(m Model) Model {
	if m.ExplorerCursor < len(m.ExplorerItems)-1 {
		m.ExplorerCursor++
	}
	return m
}

// GetVisibleItemCount returns the number of visible items based on terminal height.
func (h *NavigationHandler) GetVisibleItemCount(m Model) int {
	availableHeight := m.Height - 6
	if availableHeight < 3 {
		availableHeight = 15
	}
	itemsPerScreen := availableHeight
	if itemsPerScreen < 5 {
		itemsPerScreen = 10
	}
	return itemsPerScreen
}

// AdjustCursorAfterDeletion adjusts cursor position after a repository deletion.
func (h *NavigationHandler) AdjustCursorAfterDeletion(m Model) int {
	repositories := m.RepoHandler.GetRepositories()
	maxValidIndex := len(repositories) - 1

	if maxValidIndex < 0 {
		return 0
	}

	if m.Cursor > maxValidIndex {
		return maxValidIndex
	}

	return m.Cursor
}
