package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/jarmocluyse/git-dash/internal/git"
)

// addRepository adds a new repository from the input field and updates the configuration.
func (m Model) addRepository() (Model, tea.Cmd) {
	path := strings.TrimSpace(m.InputField)
	if path != "" {
		name := filepath.Base(path)
		m.Dependencies.GetRepositoryService().AddRepository(name, path)
		m.Config.RepositoryPaths = m.Dependencies.GetRepositoryService().GetRepositoryPaths()
		m.Dependencies.GetConfigService().Save(m.Config)
		m.State = ListView
		m.NavItemsNeedSync = true // Cache needs update
		return m, m.updateRepositoryStatuses()
	}
	return m, nil
}

// removeRepositoryByPath removes a repository from the handler by its path.
func (m Model) removeRepositoryByPath(path string) {
	m.Dependencies.GetRepositoryService().RemoveRepositoryByPath(path)
	m.NavItemsNeedSync = true // Cache needs update after removing
}

// discoverWorktrees discovers and adds worktrees from the currently selected bare repository.
func (m Model) discoverWorktrees() (Model, tea.Cmd) {
	navigableItems := m.getNavigableItems()
	if m.Cursor >= len(navigableItems) {
		return m, nil
	}

	selectedItem := navigableItems[m.Cursor]

	// Only discover worktrees from bare repositories
	if selectedItem.Type != "repository" {
		return m, nil
	}

	selectedRepo := selectedItem.Repository
	gitChecker := git.NewChecker()

	if !gitChecker.IsBareRepository(selectedRepo.Path) {
		return m, nil
	}

	worktrees, err := gitChecker.ListWorktrees(selectedRepo.Path)
	if err != nil {
		return m, nil
	}

	for _, wt := range worktrees {
		if wt.Path != selectedRepo.Path && !wt.Bare {
			name := fmt.Sprintf("%s-%s", selectedRepo.Name, wt.Branch)
			m.Dependencies.GetRepositoryService().AddRepository(name, wt.Path)
		}
	}

	m.Config.RepositoryPaths = m.Dependencies.GetRepositoryService().GetRepositoryPaths()
	m.Dependencies.GetConfigService().Save(m.Config)
	m.NavItemsNeedSync = true // Cache needs update after adding worktrees
	return m, m.updateRepositoryStatuses()
}

// deleteSelectedRepository removes the currently selected repository from the configuration.
func (m Model) deleteSelectedRepository() (Model, tea.Cmd) {
	navigableItems := m.getNavigableItems()
	if m.Cursor >= len(navigableItems) {
		return m, nil
	}

	selectedItem := navigableItems[m.Cursor]

	// Only allow deletion of repositories, not worktrees
	if selectedItem.Type == "repository" {
		// Find the repository index in the original array
		repositories := m.Dependencies.GetRepositoryService().GetRepositories()
		for i, repo := range repositories {
			if repo.Path == selectedItem.Repository.Path {
				m.Dependencies.GetRepositoryService().RemoveRepository(i)
				m.Cursor = m.adjustCursorAfterDeletion()
				m.Config.RepositoryPaths = m.Dependencies.GetRepositoryService().GetRepositoryPaths()
				m.Dependencies.GetConfigService().Save(m.Config)
				m.NavItemsNeedSync = true // Cache needs update after deletion
				break
			}
		}
	}

	return m, nil
}

// adjustCursorAfterDeletion adjusts the cursor position after deleting a repository.
func (m Model) adjustCursorAfterDeletion() int {
	navigableItems := m.getNavigableItems()
	if m.Cursor >= len(navigableItems) && len(navigableItems) > 0 {
		newCursor := len(navigableItems) - 1
		// Adjust scroll offset if needed
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
