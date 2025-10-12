package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/wip-tui/internal/config"
	"github.com/jarmocluyse/wip-tui/internal/explorer"
	"github.com/jarmocluyse/wip-tui/internal/git"
	"github.com/jarmocluyse/wip-tui/internal/logging"
	"github.com/jarmocluyse/wip-tui/internal/repository"
	"github.com/jarmocluyse/wip-tui/internal/theme"
	"github.com/jarmocluyse/wip-tui/internal/ui/pages/details"
	"github.com/jarmocluyse/wip-tui/internal/ui/types"
)

// CreateInitialModel creates and returns an initial Model with default configuration.
func CreateInitialModel(deps Dependencies) Model {
	cfg, err := deps.GetConfigService().Load()
	if err != nil {
		cfg = &config.Config{RepositoryPaths: []string{}}
	}

	repoHandler := repository.NewHandler()
	repoHandler.SetRepositories(cfg.RepositoryPaths)

	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = "/"
	}

	return Model{
		Dependencies:     deps,
		Config:           cfg,
		RepoHandler:      repoHandler,
		State:            ListView,
		Cursor:           0,
		ExplorerPath:     homeDir,
		ExplorerCursor:   0,
		NavItemsNeedSync: true, // Initialize cache as needing sync
	}
}

// Init initializes the Model and returns commands to run on startup.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.updateRepositoryStatuses(),
		tea.WindowSize(), // Explicitly request window size
	)
}

// updateRepositoryStatuses creates a command that updates the status of all repositories.
func (m Model) updateRepositoryStatuses() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		repositories := m.RepoHandler.GetRepositories()
		updatedRepos := make([]repository.Repository, len(repositories))
		copy(updatedRepos, repositories)

		// Use concurrent status updating
		m.Dependencies.GetStatusUpdater().UpdateRepositories(updatedRepos)

		return StatusMessage{Repositories: updatedRepos}
	})
}

// Update handles incoming messages and updates the model accordingly.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case StatusMessage:
		return m.handleStatusUpdate(msg)
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	}
	return m, nil
}

// handleKeyPress processes keyboard input and returns updated model and commands.
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Log current state and key press
	stateNames := []string{"ListView", "RepoManagementView", "ExplorerView", "DetailsView"}
	stateName := "Unknown"
	if int(m.State) < len(stateNames) {
		stateName = stateNames[m.State]
	}
	logging.Get().Debug("key pressed", "key", msg.String(), "state", stateName)

	// Global help modal toggle
	if msg.String() == "?" {
		m.ShowHelpModal = !m.ShowHelpModal
		return m, nil
	}

	// If help modal is open, handle its keys
	if m.ShowHelpModal {
		return m.handleHelpModalKeys(msg)
	}

	switch m.State {
	case ListView:
		return m.handleListViewKeys(msg)
	case RepoManagementView:
		return m.handleRepoManagementViewKeys(msg)
	case ExplorerView:
		return m.handleExplorerViewKeys(msg)
	case DetailsView:
		return m.handleDetailsViewKeys(msg)
	}
	return m, nil
}

// handleStatusUpdate processes repository status updates and updates the model.
func (m Model) handleStatusUpdate(msg StatusMessage) (tea.Model, tea.Cmd) {
	// Update the repository handler with the updated repositories
	updatedPaths := make([]string, len(msg.Repositories))
	for i, repo := range msg.Repositories {
		updatedPaths[i] = repo.Path
	}
	m.RepoHandler.SetRepositories(updatedPaths)

	// Copy the repository states back to the handler
	for i, repo := range msg.Repositories {
		if i < len(m.RepoHandler.GetRepositories()) {
			repos := m.RepoHandler.GetRepositories()
			repos[i] = repo
		}
	}

	// Mark navigable items cache as needing sync
	m.NavItemsNeedSync = true

	return m, nil
}

// createDirectoryExplorer creates a new directory explorer instance.
func (m Model) createDirectoryExplorer() explorer.Explorer {
	return explorer.New(m.Dependencies.GetGitChecker(), nil)
}

// loadExplorerDirectory loads and displays the contents of the current explorer directory.
func (m Model) loadExplorerDirectory() (Model, tea.Cmd) {
	explorer := m.createDirectoryExplorer()
	repositories := m.RepoHandler.GetRepositories()
	items, err := explorer.ListDirectory(m.ExplorerPath, repositories)
	if err != nil {
		return m, nil
	}

	// Debug: Write to debug file
	logging.Get().Debug("loading explorer directory",
		"path", m.ExplorerPath,
		"repositories_count", len(repositories),
		"items_found", len(items))

	for i, item := range items {
		if item.IsGitRepo {
			logging.Get().Debug("git repository found",
				"index", i,
				"path", item.Path,
				"is_added", item.IsAdded)
		}
	}

	m.ExplorerItems = items
	m.ExplorerCursor = 0
	return m, nil
}

// handleListViewKeys handles keyboard input in the main repository list view.
func (m Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	switch keyStr {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		return m.moveCursorUp(), nil
	case "down", "j":
		return m.moveCursorDown(), nil
	case "enter":
		return m.navigateToSelected()
	case "m":
		return m.enterRepoManagementMode(), nil
	case "e":
		return m.enterExplorerMode()
	case "w":
		return m.discoverWorktrees()
	case "d":
		return m.deleteSelectedRepository()
	case "r":
		return m, m.updateRepositoryStatuses()
	default:
		// Check for configurable actions
		if action := m.Config.Keybindings.FindActionByKey(keyStr); action != nil {
			return m.executeConfiguredAction(*action)
		}
	}
	return m, nil
}

// handleRepoManagementViewKeys handles keyboard input in the repository management view.
func (m Model) handleRepoManagementViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		return m.exitRepoManagementMode(), nil
	case "up", "k":
		return m.moveCursorUp(), nil
	case "down", "j":
		return m.moveCursorDown(), nil
	case "enter":
		return m.navigateToSelected()
	case "e":
		return m.enterExplorerMode()
	case "d":
		return m.deleteSelectedRepository()
	case "r":
		return m, m.updateRepositoryStatuses()
	case "q":
		return m, tea.Quit
	default:
		return m, nil
	}
}

// handleExplorerViewKeys handles keyboard input in the directory explorer view.
func (m Model) handleExplorerViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()
	logging.Get().Debug("explorer key pressed",
		"key", keyStr,
		"length", len(keyStr),
		"bytes", []byte(keyStr))

	switch keyStr {
	case "ctrl+c", "esc", "q":
		m.State = m.PreviousState // Return to the previous state
		return m, nil
	case "up", "k":
		return m.moveExplorerCursorUp(), nil
	case "down", "j":
		return m.moveExplorerCursorDown(), nil
	case "enter":
		return m.handleExplorerSelection()
	case "space", " ":
		logging.Get().Info("space key detected, calling toggleRepositorySelection", "key", keyStr)
		return m.toggleRepositorySelection()
	default:
		// Check for configurable actions
		if action := m.Config.Keybindings.FindActionByKey(keyStr); action != nil {
			return m.executeConfiguredActionForExplorer(*action)
		}

		logging.Get().Debug("unhandled key in explorer",
			"key", keyStr,
			"length", len(keyStr),
			"bytes", []byte(keyStr))
	}
	return m, nil
}

// handleDetailsViewKeys handles keyboard input in the repository details view.
func (m Model) handleDetailsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "b", "esc":
		// Return to list view
		m.State = ListView
		m.SelectedNavItem = nil
		return m, nil
	}
	return m, nil
}

// handleHelpModalKeys handles keyboard input when the help modal is open.
func (m Model) handleHelpModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "?":
		m.ShowHelpModal = false
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// Helper methods
// moveCursorUp moves the cursor up in the current list and adjusts scroll offset if needed.
func (m Model) moveCursorUp() Model {
	if m.Cursor > 0 {
		m.Cursor--
		// Update scroll offset if needed
		if m.Cursor < m.ScrollOffset {
			m.ScrollOffset = m.Cursor
		}
	}
	return m
}

// moveCursorDown moves the cursor down in the current list and adjusts scroll offset if needed.
func (m Model) moveCursorDown() Model {
	navigableItems := m.getNavigableItems()
	if m.Cursor < len(navigableItems)-1 {
		m.Cursor++
		// Update scroll offset if needed
		visibleItems := m.getVisibleItemCount()
		if m.Cursor >= m.ScrollOffset+visibleItems {
			m.ScrollOffset = m.Cursor - visibleItems + 1
		}
	}
	return m
}

// Calculate how many items can be visible based on terminal height
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

// moveExplorerCursorUp moves the explorer cursor up by one position.
func (m Model) moveExplorerCursorUp() Model {
	if m.ExplorerCursor > 0 {
		m.ExplorerCursor--
	}
	return m
}

// moveExplorerCursorDown moves the explorer cursor down by one position.
func (m Model) moveExplorerCursorDown() Model {
	if m.ExplorerCursor < len(m.ExplorerItems)-1 {
		m.ExplorerCursor++
	}
	return m
}

// enterRepoManagementMode switches the UI to repository management mode.
func (m Model) enterRepoManagementMode() Model {
	m.State = RepoManagementView
	m.Cursor = 0
	m.ScrollOffset = 0
	return m
}

// exitRepoManagementMode switches the UI back to list view from repository management mode.
func (m Model) exitRepoManagementMode() Model {
	m.State = ListView
	return m
}

// addRepository adds a new repository from the input field and updates the configuration.
func (m Model) addRepository() (Model, tea.Cmd) {
	path := strings.TrimSpace(m.InputField)
	if path != "" {
		name := filepath.Base(path)
		m.RepoHandler.AddRepository(name, path)
		m.Config.RepositoryPaths = m.RepoHandler.GetPaths()
		m.Dependencies.GetConfigService().Save(m.Config)
		m.State = ListView
		m.NavItemsNeedSync = true // Cache needs update
		return m, m.updateRepositoryStatuses()
	}
	return m, nil
}

// removeLastCharacter removes the last character from the input field.
func (m Model) removeLastCharacter() Model {
	if len(m.InputField) > 0 {
		m.InputField = m.InputField[:len(m.InputField)-1]
	}
	return m
}

// appendCharacter appends a single character to the input field.
func (m Model) appendCharacter(char string) Model {
	if len(char) == 1 {
		m.InputField += char
	}
	return m
}

// enterExplorerMode switches the UI to directory explorer mode and loads the current directory.
func (m Model) enterExplorerMode() (Model, tea.Cmd) {
	m.PreviousState = m.State
	m.State = ExplorerView
	return m.loadExplorerDirectory()
}

// handleExplorerSelection handles selection in the directory explorer by navigating into directories.
func (m Model) handleExplorerSelection() (Model, tea.Cmd) {
	if len(m.ExplorerItems) == 0 || m.ExplorerCursor >= len(m.ExplorerItems) {
		return m, nil
	}

	selected := m.ExplorerItems[m.ExplorerCursor]

	if selected.IsDirectory {
		m.ExplorerPath = selected.Path
		return m.loadExplorerDirectory()
	}

	return m, nil
}

// toggleRepositorySelection toggles the selection state of a git repository in the explorer.
func (m Model) toggleRepositorySelection() (Model, tea.Cmd) {
	logging.Get().Debug("toggleRepositorySelection called",
		"items_count", len(m.ExplorerItems),
		"cursor", m.ExplorerCursor)

	if len(m.ExplorerItems) == 0 || m.ExplorerCursor >= len(m.ExplorerItems) {
		logging.Get().Debug("early return: no items or cursor out of bounds")
		return m, nil
	}

	selected := m.ExplorerItems[m.ExplorerCursor]

	if !selected.IsGitRepo {
		logging.Get().Debug("early return: selected item is not a git repo", "path", selected.Path)
		return m, nil
	}

	logging.Get().Info("toggling repository selection",
		"path", selected.Path,
		"is_added", selected.IsAdded,
		"current_repos_count", len(m.RepoHandler.GetRepositories()))

	repositories := m.RepoHandler.GetRepositories()
	for i, repo := range repositories {
		logging.Get().Debug("current repository", "index", i, "path", repo.Path)
	}

	if selected.IsAdded {
		logging.Get().Info("removing repository", "path", selected.Path)
		m.removeRepositoryByPath(selected.Path)
	} else {
		name := filepath.Base(selected.Path)
		logging.Get().Info("adding repository", "path", selected.Path, "name", name)
		m.RepoHandler.AddRepository(name, selected.Path)
	}

	m.NavItemsNeedSync = true // Cache needs update after adding/removing

	repositories = m.RepoHandler.GetRepositories()
	logging.Get().Debug("after operation", "repos_count", len(repositories))
	for i, repo := range repositories {
		logging.Get().Debug("repository after operation", "index", i, "path", repo.Path)
	}

	// Update config with new paths and save
	m.Config.RepositoryPaths = m.RepoHandler.GetPaths()
	err := m.Dependencies.GetConfigService().Save(m.Config)
	if err != nil {
		logging.Get().Error("error saving config", "error", err)
		return m, nil
	}

	// Reload directory to reflect changes
	return m.loadExplorerDirectory()
}

// removeRepositoryByPath removes a repository from the handler by its path.
func (m Model) removeRepositoryByPath(path string) {
	m.RepoHandler.RemoveRepositoryByPath(path)
	m.NavItemsNeedSync = true // Cache needs update after removing
}

// navigateToSelected navigates to the currently selected repository or worktree in details view.
func (m Model) navigateToSelected() (Model, tea.Cmd) {
	navigableItems := m.getNavigableItems()
	if m.Cursor >= len(navigableItems) {
		return m, nil
	}

	selectedItem := navigableItems[m.Cursor]

	// Store the selected item and switch to details view
	m.SelectedNavItem = &selectedItem
	m.State = DetailsView

	return m, nil
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
			m.RepoHandler.AddRepository(name, wt.Path)
		}
	}

	m.Config.RepositoryPaths = m.RepoHandler.GetPaths()
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
		repositories := m.RepoHandler.GetRepositories()
		for i, repo := range repositories {
			if repo.Path == selectedItem.Repository.Path {
				m.RepoHandler.RemoveRepository(i)
				m.Cursor = m.adjustCursorAfterDeletion()
				m.Config.RepositoryPaths = m.RepoHandler.GetPaths()
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

// executeConfiguredAction executes a user-configured action on the currently selected repository.
func (m Model) executeConfiguredAction(action config.Action) (tea.Model, tea.Cmd) {
	navigableItems := m.getNavigableItems()
	if m.Cursor >= len(navigableItems) {
		return m, nil
	}

	selectedItem := navigableItems[m.Cursor]
	var targetPath string

	if selectedItem.Type == "worktree" {
		targetPath = selectedItem.WorktreeInfo.Path
	} else if selectedItem.Type == "repository" {
		targetPath = selectedItem.Repository.Path
	} else {
		return m, nil
	}

	// Use the configured action
	cmd := action.ExecuteOpenAction(targetPath)

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			logging.Get().Error("failed to run configured action",
				"error", err,
				"path", targetPath,
				"action", action.Name,
				"key", action.Key)
		}
		return nil
	})
}

// executeConfiguredActionForExplorer executes a user-configured action on the currently selected explorer item.
func (m Model) executeConfiguredActionForExplorer(action config.Action) (tea.Model, tea.Cmd) {
	if m.ExplorerCursor >= len(m.ExplorerItems) {
		return m, nil
	}

	selectedItem := m.ExplorerItems[m.ExplorerCursor]
	targetPath := selectedItem.Path

	// If it's not a git repository, try to find the nearest git repository
	if !selectedItem.IsGitRepo {
		// Use the current explorer path as fallback
		targetPath = m.ExplorerPath
	}

	// Use the configured action
	cmd := action.ExecuteOpenAction(targetPath)

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			logging.Get().Error("failed to run configured action",
				"error", err,
				"path", targetPath,
				"action", action.Name,
				"key", action.Key)
		}
		return nil
	})
}

// getNavigableItems returns cached navigable items or rebuilds if needed
// getNavigableItems returns cached navigable items or rebuilds if needed.
func (m *Model) getNavigableItems() []types.NavigableItem {
	if m.CachedNavItems == nil || m.NavItemsNeedSync {
		m.rebuildNavigableItems()
		m.NavItemsNeedSync = false
	}
	return m.CachedNavItems
}

// rebuildNavigableItems rebuilds the cached navigable items with concurrency
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

// View renders the current view based on the model's state.
func (m Model) View() string {
	// Get the main view content
	var mainView string
	switch m.State {
	case ListView:
		mainView = m.renderListView()
	case RepoManagementView:
		mainView = m.renderRepoManagementView()
	case ExplorerView:
		mainView = m.renderExplorerView()
	case DetailsView:
		mainView = m.renderDetailsView()
	default:
		mainView = ""
	}

	// If help modal is open, overlay it on top
	if m.ShowHelpModal {
		return m.renderHelpModal(mainView)
	}

	return mainView
}

// renderListView renders the main repository list view.
func (m Model) renderListView() string {
	styles := CreateStyleConfig(m.Config.Theme)
	renderer := NewListViewRenderer(styles, m.Config.Theme)

	// Get all navigable items
	allItems := m.buildNavigableItems()

	// Calculate visible window
	visibleCount := m.getVisibleItemCount()
	start := m.ScrollOffset
	end := start + visibleCount

	// Clamp to actual item bounds
	if start >= len(allItems) {
		start = len(allItems) - 1
		if start < 0 {
			start = 0
		}
	}
	if end > len(allItems) {
		end = len(allItems)
	}

	// Get visible items
	var visibleItems []types.NavigableItem
	if len(allItems) > 0 && start < end {
		visibleItems = allItems[start:end]
	}

	// Adjust cursor to be relative to visible window
	relativeCursor := m.Cursor - m.ScrollOffset

	// Get the cached git checker
	gitChecker := m.Dependencies.GetGitChecker()

	return renderer.RenderNavigable(visibleItems, relativeCursor, m.Width, m.Height, gitChecker, m.Config.Keybindings.Actions)
}

// renderRepoManagementView renders the repository management view.
func (m Model) renderRepoManagementView() string {
	styles := CreateStyleConfig(m.Config.Theme)
	renderer := NewRepoManagementViewRenderer(styles, m.Config.Theme)
	repositories := m.RepoHandler.GetRepositories()
	return renderer.Render(repositories, m.Cursor, m.Width, m.Height)
}

// renderExplorerView renders the directory explorer view.
func (m Model) renderExplorerView() string {
	styles := CreateStyleConfig(m.Config.Theme)
	renderer := NewExplorerViewRenderer(styles, m.Config.Theme)
	return renderer.Render(m.ExplorerPath, m.ExplorerItems, m.ExplorerCursor, m.Width, m.Height)
}

// NewDetailsViewRenderer creates a new details view renderer with the given styles and theme.
func NewDetailsViewRenderer(styles StyleConfig, themeConfig theme.Theme) *details.Renderer {
	detailsStyles := details.StyleConfig{
		Item:         styles.Item,
		SelectedItem: styles.SelectedItem,
		Label:        styles.Item.Foreground(lipgloss.Color(themeConfig.Colors.Selected)).Bold(true),
		Value:        styles.Item,
		Help:         styles.Help,
		Border:       styles.Border,
		Title:        styles.Item.Bold(true),
	}
	return details.NewRenderer(detailsStyles, themeConfig)
}

// renderDetailsView renders the repository details view.
func (m Model) renderDetailsView() string {
	if m.SelectedNavItem == nil {
		// Fallback to list view if no item is selected
		return m.renderListView()
	}

	styles := CreateStyleConfig(m.Config.Theme)
	renderer := NewDetailsViewRenderer(styles, m.Config.Theme)
	return renderer.Render(*m.SelectedNavItem, m.Width, m.Height)
}

// renderHelpModal renders the help modal overlay on top of the background view.
func (m Model) renderHelpModal(backgroundView string) string {
	styles := CreateStyleConfig(m.Config.Theme)

	// Help content with keybindings
	helpContent := strings.Builder{}

	// General navigation
	helpContent.WriteString("GENERAL NAVIGATION:\n")
	helpContent.WriteString("  ↑/k           Navigate up\n")
	helpContent.WriteString("  ↓/j           Navigate down\n")
	helpContent.WriteString("  Enter         Select/Confirm\n")
	helpContent.WriteString("  Esc/q         Go back/Cancel\n")
	helpContent.WriteString("  Ctrl+C        Quit application\n")
	helpContent.WriteString("  ?             Toggle this help\n\n")

	// Repository list view
	helpContent.WriteString("REPOSITORY LIST:\n")
	helpContent.WriteString("  a             Add new repository\n")
	helpContent.WriteString("  r/F5          Refresh statuses\n")
	helpContent.WriteString("  d             Remove repository\n")
	helpContent.WriteString("  l             Open in Lazygit\n")
	helpContent.WriteString("  e             Browse directories\n\n")

	// Explorer view
	helpContent.WriteString("DIRECTORY EXPLORER:\n")
	helpContent.WriteString("  Space         Toggle repository selection\n")
	helpContent.WriteString("  l             Open directory in Lazygit\n\n")

	// Status indicators
	helpContent.WriteString("STATUS INDICATORS:\n")
	helpContent.WriteString(fmt.Sprintf("  %s            Clean (no changes)\n", m.Config.Theme.Indicators.Clean))
	helpContent.WriteString(fmt.Sprintf("  %s            Uncommitted changes\n", m.Config.Theme.Indicators.Dirty))
	helpContent.WriteString(fmt.Sprintf("  %s            Unpushed commits\n", m.Config.Theme.Indicators.Unpushed))
	helpContent.WriteString(fmt.Sprintf("  %s            Untracked files\n", m.Config.Theme.Indicators.Untracked))
	helpContent.WriteString(fmt.Sprintf("  %s            Error accessing repository\n", m.Config.Theme.Indicators.Error))

	// Create modal with title and content
	title := styles.HelpModalTitle.Render("Help & Keybindings")
	content := styles.HelpModalContent.Render(helpContent.String())
	footer := styles.HelpModalFooter.Render("Press ? or Esc to close")

	modal := lipgloss.JoinVertical(lipgloss.Left, title, content, footer)
	styledModal := styles.HelpModal.Render(modal)

	// Use terminal dimensions with fallbacks
	width := m.Width
	height := m.Height
	if width == 0 {
		width = 120
	}
	if height == 0 {
		height = 40
	}

	// Use lipgloss.Place to overlay the modal on top of the background
	// This creates a proper overlay effect where the background is preserved
	// and the modal is centered on top
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, styledModal, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceForeground(lipgloss.NoColor{}))
}
