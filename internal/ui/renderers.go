package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/wip-tui/internal/explorer"
	"github.com/jarmocluyse/wip-tui/internal/repository"
	"github.com/jarmocluyse/wip-tui/internal/theme"
	"github.com/jarmocluyse/wip-tui/internal/ui/header"
	"github.com/jarmocluyse/wip-tui/internal/ui/repo"
)

type ExplorerItem = explorer.Item

type ListViewRenderer struct {
	styles StyleConfig
	theme  theme.Theme
	header *header.Renderer
	repo   *repo.Renderer
}

func NewListViewRenderer(styles StyleConfig, themeConfig theme.Theme) *ListViewRenderer {
	repoStyles := repo.StyleConfig{
		Item:              styles.Item,
		SelectedItem:      styles.SelectedItem,
		StatusUncommitted: styles.StatusUncommitted,
		StatusUnpushed:    styles.StatusUnpushed,
		StatusUntracked:   styles.StatusUntracked,
		StatusError:       styles.StatusError,
		StatusClean:       styles.StatusClean,
		StatusNotAdded:    styles.StatusNotAdded,
		Help:              styles.Help,
		Branch:            styles.Branch,
		Border:            styles.Border,
	}
	return &ListViewRenderer{
		styles: styles,
		theme:  themeConfig,
		header: header.NewRenderer(themeConfig),
		repo:   repo.NewRenderer(repoStyles, themeConfig),
	}
}

func (r *ListViewRenderer) Render(repositories []repository.Repository, cursor int, width int) string {
	content := r.header.RenderWithSpacing("Git Repository Status", width)

	if len(repositories) == 0 {
		content += r.renderEmptyState()
	} else {
		content += r.renderRepositoryList(repositories, cursor)
	}

	// Position help at bottom of terminal
	help := r.renderHelp()

	// Use lipgloss to place content and help
	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		lipgloss.Place(120, 1, lipgloss.Left, lipgloss.Bottom, help),
	)
}

func (r *ListViewRenderer) RenderNavigable(items []NavigableItem, cursor int, width, height int) string {
	title := r.header.Render("Git Repository Status", width)

	var mainContent string
	if len(items) == 0 {
		mainContent = r.renderEmptyState()
	} else {
		mainContent = r.renderNavigableItemList(items, cursor)
	}

	// Position help at bottom of terminal
	help := r.renderHelp()

	// Use actual terminal dimensions with fallbacks
	if height == 0 {
		height = 24
	}
	if width == 0 {
		width = 120
	}

	// Calculate lines used
	titleLines := strings.Count(title, "\n") + 1
	contentLines := strings.Count(mainContent, "\n") + 1
	helpLines := strings.Count(help, "\n") + 1
	usedLines := titleLines + 2 + contentLines + helpLines // +2 for spacing after title

	// Calculate empty lines needed to push help to very bottom
	emptyLines := height - usedLines + 1 // +1 to push it to the very last line
	if emptyLines < 0 {
		emptyLines = 0
	}

	// Build the full view
	fullContent := title + "\n\n" + mainContent
	if emptyLines > 0 {
		fullContent += strings.Repeat("\n", emptyLines)
	}
	fullContent += help

	return fullContent
}

func (r *ListViewRenderer) renderNavigableItemList(items []NavigableItem, cursor int) string {
	var content string
	i := 0

	for i < len(items) {
		item := items[i]

		if item.Type == "repository" && item.Repository.IsBare {
			// Start of bare repository group - collect all items in this group
			groupContent := r.renderNavigableItem(item, i, cursor)

			// Add all worktrees that belong to this bare repository
			j := i + 1
			for j < len(items) && items[j].Type == "worktree" && items[j].ParentRepo.Path == item.Repository.Path {
				groupContent += "\n" + r.renderNavigableItem(items[j], j, cursor)
				j++
			}

			// Wrap the entire group in a border
			borderedGroup := r.styles.Border.Render(groupContent)
			content += borderedGroup + "\n"

			// Move index to after the group
			i = j
		} else {
			// Regular item (non-bare repository or standalone worktree)
			content += r.renderNavigableItem(item, i, cursor) + "\n"
			i++
		}
	}

	return content
}

func (r *ListViewRenderer) renderNavigableItem(item NavigableItem, index, cursor int) string {
	isSelected := index == cursor
	cursorIndicator := r.getCursorIndicator(isSelected)

	if item.Type == "repository" {
		return r.repo.RenderRepositoryOnly(*item.Repository, isSelected, cursorIndicator)
	} else if item.Type == "worktree" {
		wt := item.WorktreeInfo
		parentName := item.ParentRepo.Name

		return r.repo.RenderWorktree(*wt, parentName, item.ParentRepo.Path, isSelected, cursorIndicator, item.IsLast)
	}

	return ""
}

func (r *ListViewRenderer) renderEmptyState() string {
	return r.styles.Item.Render("No repositories configured.") + "\n\n"
}

func (r *ListViewRenderer) renderRepositoryList(repositories []repository.Repository, cursor int) string {
	var content string
	for i, repo := range repositories {
		content += r.renderRepositoryItem(repo, i, cursor)
	}
	return content
}

func (r *ListViewRenderer) renderRepositoryItem(repo repository.Repository, index, cursor int) string {
	isSelected := index == cursor
	cursorIndicator := r.getCursorIndicator(isSelected)
	return r.repo.RenderRepository(repo, isSelected, cursorIndicator)
}

func (r *ListViewRenderer) getCursorIndicator(isSelected bool) string {
	if isSelected {
		return ">"
	}
	return " "
}

func (r *ListViewRenderer) getItemStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return r.styles.SelectedItem
	}
	return r.styles.Item
}

func (r *ListViewRenderer) renderHelp() string {
	return r.styles.Help.Render("a: add  e: explore  d: delete  r: refresh  l: lazygit  q: quit  ?: help")
}

type ExplorerViewRenderer struct {
	styles StyleConfig
	theme  theme.Theme
	header *header.Renderer
}

func NewExplorerViewRenderer(styles StyleConfig, themeConfig theme.Theme) *ExplorerViewRenderer {
	return &ExplorerViewRenderer{
		styles: styles,
		theme:  themeConfig,
		header: header.NewRenderer(themeConfig),
	}
}

func (r *ExplorerViewRenderer) Render(currentPath string, items []ExplorerItem, cursor int, width int) string {
	content := r.header.RenderWithSpacing("Repository Explorer", width)
	content += r.styles.Help.Render(fmt.Sprintf("Current: %s", currentPath)) + "\n\n"

	if len(items) == 0 {
		content += r.styles.Item.Render("Directory is empty or cannot be read.") + "\n\n"
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
		return r.styles.SelectedItem
	}
	return r.styles.Item
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
		return r.styles.StatusClean.Render(r.theme.Indicators.Clean)
	}

	return r.styles.StatusNotAdded.Render(r.theme.Indicators.NotAdded)
}

func (r *ExplorerViewRenderer) getWorktreeStatus(item ExplorerItem) string {
	if item.HasError {
		return r.styles.StatusError.Render(r.theme.Indicators.Error)
	}

	var status []string

	if item.HasUncommitted {
		status = append(status, r.styles.StatusUncommitted.Render(r.theme.Indicators.Dirty))
	}

	if item.HasUnpushed {
		status = append(status, r.styles.StatusUnpushed.Render(r.theme.Indicators.Unpushed))
	}

	if item.HasUntracked {
		status = append(status, r.styles.StatusUntracked.Render(r.theme.Indicators.Untracked))
	}

	if len(status) == 0 {
		if item.IsAdded {
			return r.styles.StatusClean.Render(r.theme.Indicators.Clean)
		}
		return r.styles.StatusNotAdded.Render(r.theme.Indicators.NotAdded)
	}

	return strings.Join(status, " ")
}

func (r *ExplorerViewRenderer) renderExplorerHelp() string {
	help := r.styles.Help.Render("\nControls:")
	help += r.styles.Help.Render("  ↑/k: move up   ↓/j: move down   Enter: navigate")
	help += r.styles.Help.Render("  Space: toggle Git repo   Esc/q: back to list")
	help += r.styles.Help.Render("\nIcons:")
	help += r.styles.Help.Render(fmt.Sprintf("  📁: directory   🔗: Git repo   🌳: worktree   📄: file"))
	help += r.styles.Help.Render(fmt.Sprintf("  %s: added   %s: not added   %s: uncommitted   %s: unpushed   %s: untracked   %s: error",
		r.theme.Indicators.Clean, r.theme.Indicators.NotAdded, r.theme.Indicators.Dirty, r.theme.Indicators.Unpushed, r.theme.Indicators.Untracked, r.theme.Indicators.Error))
	return help
}

type AddRepoViewRenderer struct {
	styles StyleConfig
	theme  theme.Theme
	header *header.Renderer
}

func NewAddRepoViewRenderer(styles StyleConfig, themeConfig theme.Theme) *AddRepoViewRenderer {
	return &AddRepoViewRenderer{
		styles: styles,
		theme:  themeConfig,
		header: header.NewRenderer(themeConfig),
	}
}

func (r *AddRepoViewRenderer) Render(prompt, input string, width int) string {
	content := r.header.RenderWithSpacing("Add Repository", width)
	content += prompt + "\n"
	content += r.styles.Input.Render(input+"█") + "\n\n"
	content += r.styles.Help.Render("Press Enter to add, Esc to cancel")
	return content
}
