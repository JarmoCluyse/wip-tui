package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/wip-tui/internal/config"
	"github.com/jarmocluyse/wip-tui/internal/explorer"
	"github.com/jarmocluyse/wip-tui/internal/git"
	"github.com/jarmocluyse/wip-tui/internal/logging"
	"github.com/jarmocluyse/wip-tui/internal/repository"
)

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
		Dependencies:   deps,
		Config:         cfg,
		RepoHandler:    repoHandler,
		State:          ListView,
		Cursor:         0,
		ExplorerPath:   homeDir,
		ExplorerCursor: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return m.updateRepositoryStatuses()
}

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

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Log current state and key press
	stateNames := []string{"ListView", "AddRepoView", "ExplorerView"}
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
	case AddRepoView:
		return m.handleAddRepoViewKeys(msg)
	case ExplorerView:
		return m.handleExplorerViewKeys(msg)
	}
	return m, nil
}

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

	return m, nil
}

func (m Model) createDirectoryExplorer() explorer.Explorer {
	return explorer.New(m.Dependencies.GetGitChecker(), nil)
}

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

func (m Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		return m.moveCursorUp(), nil
	case "down", "j":
		return m.moveCursorDown(), nil
	case "enter":
		return m.navigateToSelected()
	case "a":
		return m.enterAddRepoMode(), nil
	case "e":
		return m.enterExplorerMode()
	case "w":
		return m.discoverWorktrees()
	case "d":
		return m.deleteSelectedRepository()
	case "r":
		return m, m.updateRepositoryStatuses()
	case "l":
		return m.openInLazygit()
	}
	return m, nil
}

func (m Model) handleAddRepoViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		return m.exitAddRepoMode(), nil
	case "enter":
		return m.addRepository()
	case "backspace":
		return m.removeLastCharacter(), nil
	default:
		return m.appendCharacter(msg.String()), nil
	}
}

func (m Model) handleExplorerViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()
	logging.Get().Debug("explorer key pressed",
		"key", keyStr,
		"length", len(keyStr),
		"bytes", []byte(keyStr))

	switch keyStr {
	case "ctrl+c", "esc", "q":
		m.State = ListView
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
	case "l":
		return m.openExplorerInLazygit()
	default:
		logging.Get().Debug("unhandled key in explorer",
			"key", keyStr,
			"length", len(keyStr),
			"bytes", []byte(keyStr))
	}
	return m, nil
}

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

func (m Model) moveCursorDown() Model {
	navigableItems := m.buildNavigableItems()
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
func (m Model) getVisibleItemCount() int {
	// Reserve space for header, help text, and some padding
	// Each repository item takes approximately 3 lines (with border)
	availableHeight := m.Height - 6 // Reserve space for header and help
	if availableHeight < 3 {
		availableHeight = 15 // Fallback minimum for reasonable viewing
	}
	itemsPerScreen := availableHeight / 3 // Each bordered item takes ~3 lines
	if itemsPerScreen < 3 {
		itemsPerScreen = 5 // Minimum reasonable number of items
	}
	return itemsPerScreen
}

func (m Model) moveExplorerCursorUp() Model {
	if m.ExplorerCursor > 0 {
		m.ExplorerCursor--
	}
	return m
}

func (m Model) moveExplorerCursorDown() Model {
	if m.ExplorerCursor < len(m.ExplorerItems)-1 {
		m.ExplorerCursor++
	}
	return m
}

func (m Model) enterAddRepoMode() Model {
	m.State = AddRepoView
	m.InputField = ""
	m.InputPrompt = "Enter repository path: "
	return m
}

func (m Model) exitAddRepoMode() Model {
	m.State = ListView
	return m
}

func (m Model) addRepository() (Model, tea.Cmd) {
	path := strings.TrimSpace(m.InputField)
	if path != "" {
		name := filepath.Base(path)
		m.RepoHandler.AddRepository(name, path)
		m.Config.RepositoryPaths = m.RepoHandler.GetPaths()
		m.Dependencies.GetConfigService().Save(m.Config)
		m.State = ListView
		return m, m.updateRepositoryStatuses()
	}
	return m, nil
}

func (m Model) removeLastCharacter() Model {
	if len(m.InputField) > 0 {
		m.InputField = m.InputField[:len(m.InputField)-1]
	}
	return m
}

func (m Model) appendCharacter(char string) Model {
	if len(char) == 1 {
		m.InputField += char
	}
	return m
}

func (m Model) enterExplorerMode() (Model, tea.Cmd) {
	m.State = ExplorerView
	return m.loadExplorerDirectory()
}

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

func (m Model) removeRepositoryByPath(path string) {
	m.RepoHandler.RemoveRepositoryByPath(path)
}

func (m Model) navigateToSelected() (Model, tea.Cmd) {
	navigableItems := m.buildNavigableItems()
	if m.Cursor >= len(navigableItems) {
		return m, nil
	}

	selectedItem := navigableItems[m.Cursor]

	if selectedItem.Type == "worktree" {
		// Set the explorer to the worktree path
		m.State = ExplorerView
		m.ExplorerPath = selectedItem.WorktreeInfo.Path
		m.ExplorerCursor = 0
		return m.loadExplorerDirectory()
	} else if selectedItem.Type == "repository" {
		// Navigate to repository path
		m.State = ExplorerView
		m.ExplorerPath = selectedItem.Repository.Path
		m.ExplorerCursor = 0
		return m.loadExplorerDirectory()
	}

	return m, nil
}

func (m Model) discoverWorktrees() (Model, tea.Cmd) {
	navigableItems := m.buildNavigableItems()
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
	return m, m.updateRepositoryStatuses()
}

func (m Model) deleteSelectedRepository() (Model, tea.Cmd) {
	navigableItems := m.buildNavigableItems()
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
				break
			}
		}
	}

	return m, nil
}

func (m Model) adjustCursorAfterDeletion() int {
	navigableItems := m.buildNavigableItems()
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

func (m Model) openInLazygit() (Model, tea.Cmd) {
	navigableItems := m.buildNavigableItems()
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

	// Check if lazygit is available
	if _, err := exec.LookPath("lazygit"); err != nil {
		logging.Get().Error("lazygit not found in PATH", "error", err)
		return m, nil
	}

	// Create command to run lazygit in the target directory
	cmd := exec.Command("lazygit", "-p", targetPath)

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			logging.Get().Error("failed to run lazygit", "error", err, "path", targetPath)
		}
		return nil
	})
}

func (m Model) openExplorerInLazygit() (Model, tea.Cmd) {
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

	// Check if lazygit is available
	if _, err := exec.LookPath("lazygit"); err != nil {
		logging.Get().Error("lazygit not found in PATH", "error", err)
		return m, nil
	}

	// Create command to run lazygit in the target directory
	cmd := exec.Command("lazygit", "-p", targetPath)

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			logging.Get().Error("failed to run lazygit", "error", err, "path", targetPath)
		}
		return nil
	})
}

func (m Model) buildNavigableItems() []NavigableItem {
	var items []NavigableItem
	gitChecker := git.NewChecker()
	repositories := m.RepoHandler.GetRepositories()

	for i := range repositories {
		repo := &repositories[i]

		// Add the repository itself
		items = append(items, NavigableItem{
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
					items = append(items, NavigableItem{
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

func (m Model) View() string {
	// Get the main view content
	var mainView string
	switch m.State {
	case ListView:
		mainView = m.renderListView()
	case AddRepoView:
		mainView = m.renderAddRepoView()
	case ExplorerView:
		mainView = m.renderExplorerView()
	default:
		mainView = ""
	}

	// If help modal is open, overlay it on top
	if m.ShowHelpModal {
		return m.renderHelpModal(mainView)
	}

	return mainView
}

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
	var visibleItems []NavigableItem
	if len(allItems) > 0 && start < end {
		visibleItems = allItems[start:end]
	}

	// Adjust cursor to be relative to visible window
	relativeCursor := m.Cursor - m.ScrollOffset

	return renderer.RenderNavigable(visibleItems, relativeCursor, m.Width, m.Height)
}

func (m Model) renderAddRepoView() string {
	styles := CreateStyleConfig(m.Config.Theme)
	renderer := NewAddRepoViewRenderer(styles, m.Config.Theme)
	return renderer.Render(m.InputPrompt, m.InputField, m.Width)
}

func (m Model) renderExplorerView() string {
	styles := CreateStyleConfig(m.Config.Theme)
	renderer := NewExplorerViewRenderer(styles, m.Config.Theme)
	return renderer.Render(m.ExplorerPath, m.ExplorerItems, m.ExplorerCursor, m.Width)
}

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
