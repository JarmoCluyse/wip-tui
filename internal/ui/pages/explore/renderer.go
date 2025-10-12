package explore

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/wip-tui/internal/theme"
	"github.com/jarmocluyse/wip-tui/internal/ui/components/help"
	"github.com/jarmocluyse/wip-tui/internal/ui/header"
)

// Renderer handles rendering of the explore page (file/directory browser)
type Renderer struct {
	styles StyleConfig
	theme  theme.Theme
	header *header.Renderer
}

// NewRenderer creates a new explore page renderer
func NewRenderer(styles StyleConfig, themeConfig theme.Theme) *Renderer {
	return &Renderer{
		styles: styles,
		theme:  themeConfig,
		header: header.NewRenderer(themeConfig),
	}
}

// Render renders the explore page with current path, items, and cursor position
func (r *Renderer) Render(currentPath string, items []Item, cursor int, width, height int) string {
	content := r.header.RenderWithSpacing("Repository Explorer", width)
	content += r.styles.Help.Render(fmt.Sprintf("Current: %s", currentPath)) + "\n\n"

	if len(items) == 0 {
		content += r.styles.Item.Render("Directory is empty or cannot be read.") + "\n\n"
	} else {
		content += r.renderItemList(items, cursor)
	}

	// Use help component to render with bottom-aligned help
	helpBuilder := help.NewBuilder(r.styles.Help)
	bindings := []help.KeyBinding{
		{Key: "↑/↓", Description: "navigate"},
		{Key: "Enter", Description: "open"},
		{Key: "Space", Description: "toggle"},
		{Key: "Esc", Description: "back"},
	}

	return helpBuilder.RenderWithBottomHelpAndHeader(content, bindings, width, height, 3)
}

// renderItemList renders a list of items with the given cursor position.
func (r *Renderer) renderItemList(items []Item, cursor int) string {
	var content string
	for i, item := range items {
		content += r.renderItem(item, i, cursor)
	}
	return content
}

// renderItem renders a single item with selection state.
func (r *Renderer) renderItem(item Item, index, cursor int) string {
	isSelected := index == cursor
	style := r.getItemStyle(isSelected)

	var frontIndicator string
	if isSelected {
		frontIndicator = r.theme.Indicators.Selected
	} else {
		frontIndicator = strings.Repeat(" ", lipgloss.Width(r.theme.Indicators.Selected))
	}

	icon := r.getItemIcon(item)

	// Create the display name with branch info for worktrees
	displayName := item.Name
	if item.IsWorktree && item.WorktreeInfo != nil {
		branchText := " " + r.theme.Icons.Branch.Icon + " " + item.WorktreeInfo.Branch
		branchInfo := r.styles.Branch.Render(branchText)
		displayName += branchInfo
	}

	status := r.getItemStatus(item)

	// Create left content with consistent spacing
	leftContent := fmt.Sprintf(" %s%s%s", frontIndicator, icon, displayName)
	if status != "" {
		leftContent += " " + status
	}

	// Always reserve space for end indicator and spacing to prevent layout shift
	if isSelected {
		styledEndIndicator := lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.Selected)).Render(r.theme.Indicators.SelectedEnd)
		leftContent += " " + styledEndIndicator // One space before indicator
	} else {
		// Reserve the space with spaces to prevent layout shift
		endIndicatorWidth := len(r.theme.Indicators.SelectedEnd)
		leftContent += strings.Repeat(" ", 1+endIndicatorWidth) // Same total width
	}

	content := style.Render(leftContent) + "\n"
	return content
}

// getItemStyle returns the appropriate style based on selection state.
func (r *Renderer) getItemStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return r.styles.SelectedItem
	}
	return r.styles.Item
}

// getItemIcon returns the appropriate icon for an item based on its type.
func (r *Renderer) getItemIcon(item Item) string {
	if item.Name == ".." {
		return r.theme.Icons.Folder.Icon
	}

	if item.IsWorktree {
		return r.styles.IconWorktree.Render(r.theme.Icons.Repository.Worktree)
	}

	if item.IsGitRepo {
		// Check if it's a bare repository by checking if it has worktree info but is not itself a worktree
		// For bare repositories, we typically don't have WorktreeInfo set, but this is a heuristic
		// In the explorer context, we'll treat all git repos as regular repositories for now
		// since the explorer doesn't distinguish between bare and regular repos in the Item struct
		return r.styles.IconRegular.Render(r.theme.Icons.Repository.Regular)
	}

	if item.IsDirectory {
		return r.theme.Icons.Folder.Icon
	}

	return r.theme.Icons.Tree.Branch + " "
}

// getItemStatus returns formatted status indicators for an item.
func (r *Renderer) getItemStatus(item Item) string {
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

// getWorktreeStatus returns detailed status indicators for a worktree item.
func (r *Renderer) getWorktreeStatus(item Item) string {
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

// renderHelp renders the help text for the explore page.
func (r *Renderer) renderHelp() string {
	helpBuilder := help.NewBuilder(r.styles.Help)

	bindings := []help.KeyBinding{
		{Key: "↑/↓", Description: "navigate"},
		{Key: "Enter", Description: "open"},
		{Key: "Space", Description: "toggle"},
		{Key: "Esc", Description: "back"},
	}

	return helpBuilder.BuildCompactHelp(bindings)
}
