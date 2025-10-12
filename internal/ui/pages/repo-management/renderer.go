package repomanagement

import (
	"fmt"

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
		cursorIndicator := r.getCursorIndicator(isSelected)

		var style = r.styles.Item
		if isSelected {
			style = r.styles.SelectedItem
		}

		repoLine := fmt.Sprintf("%s %s", cursorIndicator, repo.Name)
		if repo.Path != repo.Name {
			repoLine += fmt.Sprintf(" (%s)", repo.Path)
		}

		content += style.Render(repoLine) + "\n"
	}

	content += "\n"
	return content
}

// getCursorIndicator returns the appropriate cursor indicator for selected/unselected items.
func (r *Renderer) getCursorIndicator(isSelected bool) string {
	if isSelected {
		return ">"
	}
	return " "
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
