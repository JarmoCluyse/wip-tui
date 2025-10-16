package ui

import (
	"github.com/jarmocluyse/git-dash/ui/layout"
)

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

// GetVisibleItemCount returns the number of visible items based on terminal height.
func (h *NavigationHandler) GetVisibleItemCount(m Model) int {
	// Use the height calculator to determine content area
	calc := layout.NewHeightCalculator()

	// Reserve space for help (1 line) and any header lines (varies by page)
	// For home page, typically 4 header lines
	headerLines := 4
	helpLines := 1

	contentHeight, _ := calc.CalculateContentAreaHeight(m.Height, headerLines+helpLines)
	if contentHeight < 3 {
		contentHeight = 15
	}
	itemsPerScreen := contentHeight
	if itemsPerScreen < 5 {
		itemsPerScreen = 10
	}
	return itemsPerScreen
}

// AdjustCursorAfterDeletion adjusts cursor position after a repository deletion.
func (h *NavigationHandler) AdjustCursorAfterDeletion(m Model) int {
	items := m.Dependencies.GetRepoManager().GetItems()
	maxValidIndex := len(items) - 1

	if maxValidIndex < 0 {
		return 0
	}

	if m.Cursor > maxValidIndex {
		return maxValidIndex
	}

	return m.Cursor
}
