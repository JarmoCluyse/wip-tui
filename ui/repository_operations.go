package ui

import (
	"strings"

	"github.com/charmbracelet/bubbletea"
)

// addRepository adds a new repository from the input field.
func (m Model) addRepository() (Model, tea.Cmd) {
	path := strings.TrimSpace(m.InputField)
	if path != "" {
		m.Dependencies.GetRepoManager().AddRepo(path)
		m.State = ListView
		m.NavItemsNeedSync = true
		return m, m.updateRepositoryStatuses()
	}
	return m, nil
}

// removeRepositoryByPath removes a repository by its path.
func (m Model) removeRepositoryByPath(path string) {
	m.Dependencies.GetRepoManager().RemoveRepo(path)
	m.NavItemsNeedSync = true
}

// discoverWorktrees discovers and adds worktrees from the currently selected repository.
func (m Model) discoverWorktrees() (Model, tea.Cmd) {
	navigableItems := m.getNavigableItems()
	if m.Cursor >= len(navigableItems) {
		return m, nil
	}

	selectedItem := navigableItems[m.Cursor]
	if selectedItem.Type != "repository" || !selectedItem.Repository.IsBare {
		return m, nil
	}

	m.Dependencies.GetRepoManager().ReloadWorktrees()
	m.NavItemsNeedSync = true
	return m, m.updateRepositoryStatuses()
}

// deleteSelectedRepository removes the currently selected repository.
func (m Model) deleteSelectedRepository() (Model, tea.Cmd) {
	navigableItems := m.getNavigableItems()
	if m.Cursor >= len(navigableItems) {
		return m, nil
	}

	selectedItem := navigableItems[m.Cursor]
	if selectedItem.Type == "repository" {
		m.Dependencies.GetRepoManager().RemoveRepo(selectedItem.Repository.Path)
		m.Cursor = m.adjustCursorAfterDeletion()
		m.NavItemsNeedSync = true
	}

	return m, nil
}

// adjustCursorAfterDeletion adjusts the cursor position after deleting a repository.
func (m Model) adjustCursorAfterDeletion() int {
	navigableItems := m.getNavigableItems()
	if m.Cursor >= len(navigableItems) && len(navigableItems) > 0 {
		newCursor := len(navigableItems) - 1
		visibleItems := m.getVisibleItemCount()
		if newCursor < m.ScrollOffset {
			m.ScrollOffset = newCursor
		} else if newCursor >= m.ScrollOffset+visibleItems {
			m.ScrollOffset = newCursor - visibleItems + 1
		}
		return newCursor
	}
	return m.Cursor
}
