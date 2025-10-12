package repomanagement

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/wip-tui/internal/repository"
	"github.com/jarmocluyse/wip-tui/internal/theme"
	"github.com/jarmocluyse/wip-tui/internal/ui/components/help"
	"github.com/jarmocluyse/wip-tui/internal/ui/header"
)

// Renderer handles rendering of the repository management page
type Renderer struct {
	styles StyleConfig
	theme  theme.Theme
	header *header.Renderer
}

// NewRenderer creates a new repo management page renderer
func NewRenderer(styles StyleConfig, themeConfig theme.Theme) *Renderer {
	return &Renderer{
		styles: styles,
		theme:  themeConfig,
		header: header.NewRenderer(themeConfig),
	}
}

// Render renders the repository management page showing current repositories
func (r *Renderer) Render(repositories []repository.Repository, cursor int, width, height int) string {
	content := r.header.RenderWithSpacing("Repository Management", width)

	if len(repositories) == 0 {
		content += r.styles.EmptyState.Render("No repositories configured. Press 'e' to explore and add repositories.") + "\n\n"
	} else {
		content += r.renderRepositoryList(repositories, cursor, width)
	}

	// Use help component to render with bottom-aligned help
	helpBuilder := help.NewBuilder(r.styles.Help)
	bindings := []help.KeyBinding{
		{Key: "↑/↓", Description: "navigate"},
		{Key: "e", Description: "explore"},
		{Key: "c", Description: "configure actions"},
		{Key: "d", Description: "delete"},
		{Key: "r", Description: "refresh"},
		{Key: "Enter", Description: "details"},
		{Key: "Esc", Description: "back"},
	}

	return helpBuilder.RenderWithBottomHelpAndHeader(content, bindings, width, height, 3)
}

// renderRepositoryList renders the list of repositories with navigation highlighting.
func (r *Renderer) renderRepositoryList(repositories []repository.Repository, cursor int, width int) string {
	var content string
	content += r.styles.SectionTitle.Render(fmt.Sprintf("Current Repositories (%d):", len(repositories))) + "\n\n"

	for i, repo := range repositories {
		isSelected := i == cursor

		var style = r.styles.Item
		if isSelected {
			style = r.styles.SelectedItem
		}

		var frontIndicator string
		if isSelected {
			frontIndicator = r.theme.Indicators.Selected
		} else {
			frontIndicator = strings.Repeat(" ", lipgloss.Width(r.theme.Indicators.Selected))
		}

		repoLine := fmt.Sprintf(" %s%s", frontIndicator, repo.Name)
		if repo.Path != repo.Name {
			repoLine += fmt.Sprintf(" (%s)", repo.Path)
		}

		// Calculate padding to right-align the end indicator with proper width constraints
		repoLineWidth := len(repoLine)

		// Always reserve space for end indicator and spacing, even when not selected
		endIndicatorWidth := len(r.theme.Indicators.SelectedEnd)
		spacingWidth := 1 // One space before end indicator
		reservedEndWidth := endIndicatorWidth + spacingWidth

		// Ensure we don't exceed the terminal width - leave some margin
		maxContentWidth := width - 2 // margin for safety
		requiredWidth := repoLineWidth + reservedEndWidth

		var padding int
		if requiredWidth >= maxContentWidth {
			// Content too wide, use minimal spacing
			padding = 1
		} else {
			padding = maxContentWidth - repoLineWidth - reservedEndWidth
		}

		if padding < 0 {
			padding = 0
		}

		// Always add the reserved space, but only show the indicator when selected
		endIndicator := "" // Start empty
		if isSelected {
			styledEndIndicator := lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.Selected)).Render(r.theme.Indicators.SelectedEnd)
			endIndicator += " " + styledEndIndicator
		} else {
			// Reserve the space with spaces to prevent layout shift
			endIndicator += strings.Repeat(" ", endIndicatorWidth+1)
		}

		repoLine = repoLine + strings.Repeat(" ", padding) + endIndicator

		content += style.Render(repoLine) + "\n"
	}

	content += "\n"
	return content
}

// renderHelp renders the help section with available key bindings.
func (r *Renderer) renderHelp() string {
	helpBuilder := help.NewBuilder(r.styles.Help)

	bindings := []help.KeyBinding{
		{Key: "↑/↓", Description: "navigate"},
		{Key: "e", Description: "explore"},
		{Key: "d", Description: "delete"},
		{Key: "r", Description: "refresh"},
		{Key: "Enter", Description: "details"},
		{Key: "Esc", Description: "back"},
	}

	return helpBuilder.BuildCompactHelp(bindings)
}
