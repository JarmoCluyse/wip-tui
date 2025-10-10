package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewState int

const (
	ListView ViewState = iota
	AddRepoView
	ExplorerView
)

type Dependencies struct {
	configService ConfigService
	statusUpdater *RepositoryStatusUpdater
	gitChecker    GitStatusChecker
}

type Model struct {
	dependencies   Dependencies
	config         *Config
	state          ViewState
	cursor         int
	inputField     string
	inputPrompt    string
	explorerPath   string
	explorerItems  []ExplorerItem
	explorerCursor int
	err            error
}

type ExplorerItem struct {
	Name           string
	Path           string
	IsDirectory    bool
	IsGitRepo      bool
	IsAdded        bool
	IsWorktree     bool
	WorktreeInfo   *WorktreeInfo
	HasUncommitted bool
	HasUnpushed    bool
	HasUntracked   bool
	HasError       bool
}

type NavigableItem struct {
	Type         string // "repository" or "worktree"
	Repository   *Repository
	WorktreeInfo *WorktreeInfo
	ParentRepo   *Repository // For worktrees, reference to parent bare repo
}

type StatusMessage struct {
	repositories []Repository
}

type StyleConfig struct {
	title             lipgloss.Style
	item              lipgloss.Style
	selectedItem      lipgloss.Style
	statusUncommitted lipgloss.Style
	statusUnpushed    lipgloss.Style
	statusUntracked   lipgloss.Style
	statusError       lipgloss.Style
	statusClean       lipgloss.Style
	statusNotAdded    lipgloss.Style
	input             lipgloss.Style
	help              lipgloss.Style
}

func createStyleConfig() StyleConfig {
	return StyleConfig{
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1),
		item: lipgloss.NewStyle().
			PaddingLeft(2),
		selectedItem: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true),
		statusUncommitted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true),
		statusUnpushed: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD93D")).
			Bold(true),
		statusUntracked: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Bold(true),
		statusError: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true),
		statusClean: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6BCF7F")).
			Bold(true),
		statusNotAdded: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),
		input: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1).
			Width(50),
		help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Margin(1, 0),
	}
}

func createDependencies() Dependencies {
	configService := NewFileConfigService()
	gitChecker := NewGitChecker()
	statusUpdater := NewRepositoryStatusUpdater(gitChecker)

	return Dependencies{
		configService: configService,
		statusUpdater: statusUpdater,
		gitChecker:    gitChecker,
	}
}

func (m Model) createDirectoryExplorer() DirectoryExplorer {
	return NewDirectoryExplorer(m.dependencies.gitChecker, nil)
}

func createInitialModel() Model {
	deps := createDependencies()
	config, err := deps.configService.Load()
	if err != nil {
		config = &Config{Repositories: []Repository{}}
	}

	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = "/"
	}

	return Model{
		dependencies:   deps,
		config:         config,
		state:          ListView,
		cursor:         0,
		explorerPath:   homeDir,
		explorerCursor: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return m.updateRepositoryStatuses()
}

func (m Model) updateRepositoryStatuses() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		repositories := make([]Repository, len(m.config.Repositories))
		copy(repositories, m.config.Repositories)

		for i := range repositories {
			m.dependencies.statusUpdater.UpdateStatus(&repositories[i])
		}

		return StatusMessage{repositories: repositories}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case StatusMessage:
		return m.handleStatusUpdate(msg)
	case tea.WindowSizeMsg:
		return m, nil
	}
	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Log current state and key press
	stateNames := []string{"ListView", "AddRepoView", "ExplorerView"}
	stateName := "Unknown"
	if int(m.state) < len(stateNames) {
		stateName = stateNames[m.state]
	}
	logger.Debug("key pressed", "key", msg.String(), "state", stateName)

	switch m.state {
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
	m.config.Repositories = msg.repositories
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
	}
	return m, nil
}

func (m Model) enterExplorerMode() (Model, tea.Cmd) {
	m.state = ExplorerView
	return m.loadExplorerDirectory()
}

func (m Model) loadExplorerDirectory() (Model, tea.Cmd) {
	explorer := m.createDirectoryExplorer()
	items, err := explorer.ListDirectory(m.explorerPath, m.config.Repositories)
	if err != nil {
		return m, nil
	}

	// Debug: Write to debug file
	logger.Debug("loading explorer directory",
		"path", m.explorerPath,
		"repositories_count", len(m.config.Repositories),
		"items_found", len(items))

	for i, item := range items {
		if item.IsGitRepo {
			logger.Debug("git repository found",
				"index", i,
				"path", item.Path,
				"is_added", item.IsAdded)
		}
	}

	m.explorerItems = items
	m.explorerCursor = 0
	return m, nil
}

func (m Model) handleExplorerViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()
	logger.Debug("explorer key pressed",
		"key", keyStr,
		"length", len(keyStr),
		"bytes", []byte(keyStr))

	switch keyStr {
	case "ctrl+c", "esc", "q":
		m.state = ListView
		return m, nil
	case "up", "k":
		return m.moveExplorerCursorUp(), nil
	case "down", "j":
		return m.moveExplorerCursorDown(), nil
	case "enter":
		return m.handleExplorerSelection()
	case "space", " ":
		logger.Info("space key detected, calling toggleRepositorySelection", "key", keyStr)
		return m.toggleRepositorySelection()
	default:
		logger.Debug("unhandled key in explorer",
			"key", keyStr,
			"length", len(keyStr),
			"bytes", []byte(keyStr))
	}
	return m, nil
}

func (m Model) moveExplorerCursorUp() Model {
	if m.explorerCursor > 0 {
		m.explorerCursor--
	}
	return m
}

func (m Model) moveExplorerCursorDown() Model {
	if m.explorerCursor < len(m.explorerItems)-1 {
		m.explorerCursor++
	}
	return m
}

func (m Model) handleExplorerSelection() (Model, tea.Cmd) {
	if len(m.explorerItems) == 0 || m.explorerCursor >= len(m.explorerItems) {
		return m, nil
	}

	selected := m.explorerItems[m.explorerCursor]

	if selected.IsDirectory {
		m.explorerPath = selected.Path
		return m.loadExplorerDirectory()
	}

	return m, nil
}

func (m Model) toggleRepositorySelection() (Model, tea.Cmd) {
	logger.Debug("toggleRepositorySelection called",
		"items_count", len(m.explorerItems),
		"cursor", m.explorerCursor)

	if len(m.explorerItems) == 0 || m.explorerCursor >= len(m.explorerItems) {
		logger.Debug("early return: no items or cursor out of bounds")
		return m, nil
	}

	selected := m.explorerItems[m.explorerCursor]

	if !selected.IsGitRepo {
		logger.Debug("early return: selected item is not a git repo", "path", selected.Path)
		return m, nil
	}

	logger.Info("toggling repository selection",
		"path", selected.Path,
		"is_added", selected.IsAdded,
		"current_repos_count", len(m.config.Repositories))

	for i, repo := range m.config.Repositories {
		logger.Debug("current repository", "index", i, "path", repo.Path)
	}

	if selected.IsAdded {
		logger.Info("removing repository", "path", selected.Path)
		m.removeRepositoryByPath(selected.Path)
	} else {
		name := filepath.Base(selected.Path)
		logger.Info("adding repository", "path", selected.Path, "name", name)
		m.config.AddRepository(name, selected.Path)
	}

	logger.Debug("after operation", "repos_count", len(m.config.Repositories))
	for i, repo := range m.config.Repositories {
		logger.Debug("repository after operation", "index", i, "path", repo.Path)
	}

	// Save configuration
	err := m.dependencies.configService.Save(m.config)
	if err != nil {
		logger.Error("error saving config", "error", err)
		return m, nil
	}

	// Reload directory to reflect changes
	return m.loadExplorerDirectory()
}

func (m Model) removeRepositoryByPath(path string) {
	cleanPath := filepath.Clean(path)
	for i, repo := range m.config.Repositories {
		cleanRepoPath := filepath.Clean(repo.Path)
		if cleanRepoPath == cleanPath {
			m.config.RemoveRepository(i)
			break
		}
	}
}

func (m Model) navigateToSelected() (Model, tea.Cmd) {
	navigableItems := m.buildNavigableItems()
	if m.cursor >= len(navigableItems) {
		return m, nil
	}

	selectedItem := navigableItems[m.cursor]

	if selectedItem.Type == "worktree" {
		// Set the explorer to the worktree path
		m.state = ExplorerView
		m.explorerPath = selectedItem.WorktreeInfo.Path
		m.explorerCursor = 0
		return m.loadExplorerDirectory()
	} else if selectedItem.Type == "repository" {
		// Navigate to repository path
		m.state = ExplorerView
		m.explorerPath = selectedItem.Repository.Path
		m.explorerCursor = 0
		return m.loadExplorerDirectory()
	}

	return m, nil
}

func (m Model) discoverWorktrees() (Model, tea.Cmd) {
	navigableItems := m.buildNavigableItems()
	if m.cursor >= len(navigableItems) {
		return m, nil
	}

	selectedItem := navigableItems[m.cursor]

	// Only discover worktrees from bare repositories
	if selectedItem.Type != "repository" {
		return m, nil
	}

	selectedRepo := selectedItem.Repository
	gitChecker := NewGitChecker()

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
			m.config.AddRepository(name, wt.Path)
		}
	}

	m.dependencies.configService.Save(m.config)
	return m, m.updateRepositoryStatuses()
}

func (m Model) buildNavigableItems() []NavigableItem {
	var items []NavigableItem
	gitChecker := NewGitChecker()

	for i := range m.config.Repositories {
		repo := &m.config.Repositories[i]

		// Add the repository itself
		items = append(items, NavigableItem{
			Type:       "repository",
			Repository: repo,
		})

		// If it's a bare repo, add its worktrees
		if repo.IsBare {
			worktrees, err := gitChecker.ListWorktrees(repo.Path)
			if err == nil {
				for _, wt := range worktrees {
					if !wt.Bare && wt.Path != repo.Path {
						items = append(items, NavigableItem{
							Type:         "worktree",
							WorktreeInfo: &wt,
							ParentRepo:   repo,
						})
					}
				}
			}
		}
	}

	return items
}

func (m Model) moveCursorUp() Model {
	if m.cursor > 0 {
		m.cursor--
	}
	return m
}

func (m Model) moveCursorDown() Model {
	navigableItems := m.buildNavigableItems()
	if m.cursor < len(navigableItems)-1 {
		m.cursor++
	}
	return m
}

func (m Model) enterAddRepoMode() Model {
	m.state = AddRepoView
	m.inputField = ""
	m.inputPrompt = "Enter repository path: "
	return m
}

func (m Model) deleteSelectedRepository() (Model, tea.Cmd) {
	navigableItems := m.buildNavigableItems()
	if m.cursor >= len(navigableItems) {
		return m, nil
	}

	selectedItem := navigableItems[m.cursor]

	// Only allow deletion of repositories, not worktrees
	if selectedItem.Type == "repository" {
		// Find the repository index in the original array
		for i, repo := range m.config.Repositories {
			if repo.Path == selectedItem.Repository.Path {
				m.config.RemoveRepository(i)
				m.cursor = m.adjustCursorAfterDeletion()
				m.dependencies.configService.Save(m.config)
				break
			}
		}
	}

	return m, nil
}

func (m Model) adjustCursorAfterDeletion() int {
	navigableItems := m.buildNavigableItems()
	if m.cursor >= len(navigableItems) && len(navigableItems) > 0 {
		return len(navigableItems) - 1
	}
	return m.cursor
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

func (m Model) exitAddRepoMode() Model {
	m.state = ListView
	return m
}

func (m Model) addRepository() (Model, tea.Cmd) {
	path := strings.TrimSpace(m.inputField)
	if path != "" {
		name := filepath.Base(path)
		m.config.AddRepository(name, path)
		m.dependencies.configService.Save(m.config)
		m.state = ListView
		return m, m.updateRepositoryStatuses()
	}
	return m, nil
}

func (m Model) removeLastCharacter() Model {
	if len(m.inputField) > 0 {
		m.inputField = m.inputField[:len(m.inputField)-1]
	}
	return m
}

func (m Model) appendCharacter(char string) Model {
	if len(char) == 1 {
		m.inputField += char
	}
	return m
}

func (m Model) View() string {
	switch m.state {
	case ListView:
		return m.renderListView()
	case AddRepoView:
		return m.renderAddRepoView()
	case ExplorerView:
		return m.renderExplorerView()
	}
	return ""
}

func (m Model) renderListView() string {
	styles := createStyleConfig()
	renderer := NewListViewRenderer(styles)
	navigableItems := m.buildNavigableItems()
	return renderer.RenderNavigable(navigableItems, m.cursor)
}

func (m Model) renderAddRepoView() string {
	styles := createStyleConfig()
	renderer := NewAddRepoViewRenderer(styles)
	return renderer.Render(m.inputPrompt, m.inputField)
}

func (m Model) renderExplorerView() string {
	styles := createStyleConfig()
	renderer := NewExplorerViewRenderer(styles)
	return renderer.Render(m.explorerPath, m.explorerItems, m.explorerCursor)
}

type ListViewRenderer struct {
	styles StyleConfig
}

func NewListViewRenderer(styles StyleConfig) *ListViewRenderer {
	return &ListViewRenderer{styles: styles}
}

func (r *ListViewRenderer) Render(repositories []Repository, cursor int) string {
	content := r.styles.title.Render("Git Repository Status") + "\n\n"

	if len(repositories) == 0 {
		content += r.renderEmptyState()
	} else {
		content += r.renderRepositoryList(repositories, cursor)
	}

	content += r.renderHelp()
	return content
}

func (r *ListViewRenderer) RenderNavigable(items []NavigableItem, cursor int) string {
	content := r.styles.title.Render("Git Repository Status") + "\n\n"

	if len(items) == 0 {
		content += r.renderEmptyState()
	} else {
		content += r.renderNavigableItemList(items, cursor)
	}

	content += r.renderHelp()
	return content
}

func (r *ListViewRenderer) renderNavigableItemList(items []NavigableItem, cursor int) string {
	var content string
	for i, item := range items {
		content += r.renderNavigableItem(item, i, cursor)
	}
	return content
}

func (r *ListViewRenderer) renderNavigableItem(item NavigableItem, index, cursor int) string {
	isSelected := index == cursor
	cursorIndicator := r.getCursorIndicator(isSelected)
	style := r.getItemStyle(isSelected)

	if item.Type == "repository" {
		repo := item.Repository
		statusIndicator := r.getStatusIndicator(*repo)
		nameWithStatus := fmt.Sprintf("%s %s %s", cursorIndicator, repo.Name, statusIndicator)
		content := style.Render(nameWithStatus) + "\n"
		content += style.Render(fmt.Sprintf("   Path: %s", repo.Path)) + "\n"
		return content
	} else if item.Type == "worktree" {
		wt := item.WorktreeInfo
		parentName := item.ParentRepo.Name
		status := r.getWorktreeStatusIndicators(wt.Path)
		wtName := fmt.Sprintf("%s-%s", parentName, wt.Branch)

		line := fmt.Sprintf("%s ├─ 🌳 %s %s", cursorIndicator, wtName, status)
		content := style.Render(line) + "\n"

		relativePath := r.getRelativePathToBareRepo(wt.Path, item.ParentRepo.Path)
		content += style.Render(fmt.Sprintf("      Path: %s", relativePath)) + "\n"
		return content
	}

	return ""
}

func (r *ListViewRenderer) renderEmptyState() string {
	return r.styles.item.Render("No repositories configured.") + "\n\n"
}

func (r *ListViewRenderer) renderRepositoryList(repositories []Repository, cursor int) string {
	var content string
	for i, repo := range repositories {
		content += r.renderRepositoryItem(repo, i, cursor)
	}
	return content
}

func (r *ListViewRenderer) renderRepositoryItem(repo Repository, index, cursor int) string {
	isSelected := index == cursor
	cursorIndicator := r.getCursorIndicator(isSelected)
	style := r.getItemStyle(isSelected)

	statusIndicator := r.getStatusIndicator(repo)
	nameWithStatus := fmt.Sprintf("%s %s %s", cursorIndicator, repo.Name, statusIndicator)
	content := style.Render(nameWithStatus) + "\n"
	content += style.Render(fmt.Sprintf("   Path: %s", repo.Path)) + "\n"

	// If this is a bare repository, always show its worktrees
	if repo.IsBare {
		worktreeContent := r.renderWorktrees(repo)
		if worktreeContent != "" {
			content += worktreeContent
		}
	}

	return content
}

func (r *ListViewRenderer) getCursorIndicator(isSelected bool) string {
	if isSelected {
		return ">"
	}
	return " "
}

func (r *ListViewRenderer) getItemStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return r.styles.selectedItem
	}
	return r.styles.item
}

func (r *ListViewRenderer) getStatusIndicator(repo Repository) string {
	var indicators []string

	if repo.HasError {
		indicators = append(indicators, r.styles.statusError.Render("✗"))
		return strings.Join(indicators, " ")
	}

	// Don't show status icons for bare repos - they'll be shown on worktrees
	if repo.IsBare {
		return "📁"
	}

	if repo.IsWorktree {
		indicators = append(indicators, r.styles.help.Render("🌳"))
	}

	if repo.HasUncommitted {
		indicators = append(indicators, r.styles.statusUncommitted.Render("●"))
	}

	if repo.HasUnpushed {
		indicators = append(indicators, r.styles.statusUnpushed.Render("↑"))
	}

	if repo.HasUntracked {
		indicators = append(indicators, r.styles.statusUntracked.Render("?"))
	}

	if !repo.HasUncommitted && !repo.HasUnpushed && !repo.HasUntracked && !repo.IsBare {
		indicators = append(indicators, r.styles.statusClean.Render("✓"))
	}

	return strings.Join(indicators, " ")
}

func (r *ListViewRenderer) renderWorktrees(repo Repository) string {
	gitChecker := NewGitChecker()
	worktrees, err := gitChecker.ListWorktrees(repo.Path)
	if err != nil {
		return ""
	}

	var content string
	for _, wt := range worktrees {
		// Skip the bare repository itself
		if wt.Bare || wt.Path == repo.Path {
			continue
		}

		content += r.renderWorktreeItem(wt, repo.Name, repo.Path)
	}

	return content
}

func (r *ListViewRenderer) renderWorktreeItem(wt WorktreeInfo, repoName string, bareRepoPath string) string {
	// Create worktree status
	status := r.getWorktreeStatusIndicators(wt.Path)

	// Format the worktree display
	wtName := fmt.Sprintf("%s-%s", repoName, wt.Branch)
	line := r.styles.item.Render(fmt.Sprintf("   ├─ 🌳 %s %s", wtName, status))
	line += "\n"

	// Make path relative to the bare repo directory
	relativePath := r.getRelativePathToBareRepo(wt.Path, bareRepoPath)
	line += r.styles.item.Render(fmt.Sprintf("      Path: %s", relativePath))
	line += "\n"

	return line
}

func (r *ListViewRenderer) getRelativePathToBareRepo(worktreePath, bareRepoPath string) string {
	bareRepoDir := filepath.Dir(bareRepoPath)

	if relPath, err := filepath.Rel(bareRepoDir, worktreePath); err == nil {
		return relPath
	}

	return worktreePath
}

func (r *ListViewRenderer) getWorktreeStatusIndicators(path string) string {
	gitChecker := NewGitChecker()

	if !gitChecker.IsGitRepository(path) {
		return r.styles.statusError.Render("✗")
	}

	var indicators []string

	if gitChecker.HasUncommittedChanges(path) {
		indicators = append(indicators, r.styles.statusUncommitted.Render("●"))
	}

	if gitChecker.HasUnpushedCommits(path) {
		indicators = append(indicators, r.styles.statusUnpushed.Render("↑"))
	}

	if gitChecker.HasUntrackedFiles(path) {
		indicators = append(indicators, r.styles.statusUntracked.Render("?"))
	}

	if len(indicators) == 0 {
		return r.styles.statusClean.Render("✓")
	}

	return strings.Join(indicators, " ")
}

func (r *ListViewRenderer) renderHelp() string {
	help := r.styles.help.Render("\nControls:")
	help += r.styles.help.Render("  ↑/k: move up   ↓/j: move down")
	help += r.styles.help.Render("  a: add repo    e: explore    d: delete repo    r: refresh")
	help += r.styles.help.Render("  w: discover worktrees from bare repo    q: quit")
	help += r.styles.help.Render("\nIndicators:")
	help += r.styles.help.Render("  📁: bare repo   🌳: worktree   ●: uncommitted   ↑: unpushed   ?: untracked   ✗: error   ✓: clean")
	return help
}

type ExplorerViewRenderer struct {
	styles StyleConfig
}

func NewExplorerViewRenderer(styles StyleConfig) *ExplorerViewRenderer {
	return &ExplorerViewRenderer{styles: styles}
}

func (r *ExplorerViewRenderer) Render(currentPath string, items []ExplorerItem, cursor int) string {
	content := r.styles.title.Render("Repository Explorer") + "\n\n"
	content += r.styles.help.Render(fmt.Sprintf("Current: %s", currentPath)) + "\n\n"

	if len(items) == 0 {
		content += r.styles.item.Render("Directory is empty or cannot be read.") + "\n\n"
	} else {
		content += r.renderItemList(items, cursor)
	}

	content += r.renderExplorerHelp()
	return content
}

func (r *ExplorerViewRenderer) renderItemList(items []ExplorerItem, cursor int) string {
	var content string
	for i, item := range items {
		content += r.renderExplorerItem(item, i, cursor)
	}
	return content
}

func (r *ExplorerViewRenderer) renderExplorerItem(item ExplorerItem, index, cursor int) string {
	isSelected := index == cursor
	cursorIndicator := r.getCursorIndicator(isSelected)
	style := r.getItemStyle(isSelected)

	icon := r.getItemIcon(item)
	status := r.getItemStatus(item)

	line := fmt.Sprintf("%s %s%s", cursorIndicator, icon, item.Name)
	if status != "" {
		line += " " + status
	}

	content := style.Render(line) + "\n"
	return content
}

func (r *ExplorerViewRenderer) getCursorIndicator(isSelected bool) string {
	if isSelected {
		return ">"
	}
	return " "
}

func (r *ExplorerViewRenderer) getItemStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return r.styles.selectedItem
	}
	return r.styles.item
}

func (r *ExplorerViewRenderer) getItemIcon(item ExplorerItem) string {
	if item.Name == ".." {
		return "📁 "
	}
	if item.IsWorktree {
		return "🌳 "
	}
	if item.IsDirectory {
		return "📁 "
	}
	if item.IsGitRepo {
		return "🔗 "
	}
	return "📄 "
}

func (r *ExplorerViewRenderer) getItemStatus(item ExplorerItem) string {
	if !item.IsGitRepo {
		return ""
	}

	// For worktrees, show detailed status
	if item.IsWorktree {
		return r.getWorktreeStatus(item)
	}

	// For regular repos, show added/not added status
	if item.IsAdded {
		return r.styles.statusClean.Render("✓")
	}

	return r.styles.statusNotAdded.Render("○")
}

func (r *ExplorerViewRenderer) getWorktreeStatus(item ExplorerItem) string {
	if item.HasError {
		return r.styles.statusError.Render("✗")
	}

	var status []string

	if item.HasUncommitted {
		status = append(status, r.styles.statusUncommitted.Render("●"))
	}

	if item.HasUnpushed {
		status = append(status, r.styles.statusUnpushed.Render("↑"))
	}

	if item.HasUntracked {
		status = append(status, r.styles.statusUntracked.Render("?"))
	}

	if len(status) == 0 {
		if item.IsAdded {
			return r.styles.statusClean.Render("✓")
		}
		return r.styles.statusNotAdded.Render("○")
	}

	return strings.Join(status, " ")
}

func (r *ExplorerViewRenderer) renderExplorerHelp() string {
	help := r.styles.help.Render("\nControls:")
	help += r.styles.help.Render("  ↑/k: move up   ↓/j: move down   Enter: navigate")
	help += r.styles.help.Render("  Space: toggle Git repo   Esc/q: back to list")
	help += r.styles.help.Render("\nIcons:")
	help += r.styles.help.Render("  📁: directory   🔗: Git repo   🌳: worktree   📄: file")
	help += r.styles.help.Render("  ✓: added   ○: not added   ●: uncommitted   ↑: unpushed   ?: untracked   ✗: error")
	return help
}

type AddRepoViewRenderer struct {
	styles StyleConfig
}

func NewAddRepoViewRenderer(styles StyleConfig) *AddRepoViewRenderer {
	return &AddRepoViewRenderer{styles: styles}
}

func (r *AddRepoViewRenderer) Render(prompt, input string) string {
	content := r.styles.title.Render("Add Repository") + "\n\n"
	content += prompt + "\n"
	content += r.styles.input.Render(input+"█") + "\n\n"
	content += r.styles.help.Render("Press Enter to add, Esc to cancel")
	return content
}
