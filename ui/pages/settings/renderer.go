package settings

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/repomanager"
	"github.com/jarmocluyse/git-dash/internal/theme"
	"github.com/jarmocluyse/git-dash/ui/components/help"
	"github.com/jarmocluyse/git-dash/ui/header"
)

type SettingsSection string

const (
	RepositoriesSection SettingsSection = "repositories"
	ActionsSection      SettingsSection = "actions"
	ThemeSection        SettingsSection = "theme"
)

type SettingsData struct {
	Repositories []*repomanager.RepoItem
	Actions      []config.Action
	Theme        theme.Theme
	Keybindings  config.Keybindings
}

// Renderer handles rendering of the settings page
type Renderer struct {
	styles StyleConfig
	theme  theme.Theme
	header *header.Renderer
}

// NewRenderer creates a new settings page renderer
func NewRenderer(styles StyleConfig, themeConfig theme.Theme) *Renderer {
	return &Renderer{
		styles: styles,
		theme:  themeConfig,
		header: header.NewRenderer(themeConfig),
	}
}

// Render renders the settings page with all sections
func (r *Renderer) Render(data SettingsData, currentSection SettingsSection, cursor int, width, height int) string {
	content := r.header.RenderWithCountAndSpacing("git-dash", "", 1, width)
	content += r.header.RenderWithSpacing("Settings", width)

	content += r.renderSectionNavigation(currentSection, width)
	content += "\n"

	switch currentSection {
	case RepositoriesSection:
		content += r.renderRepositoriesSection(data.Repositories, cursor, width)
	case ActionsSection:
		content += r.renderActionsSection(data.Actions, cursor, width)
	case ThemeSection:
		content += r.renderThemeSection(data.Theme, cursor, width)
	}

	// Use help component to render with bottom-aligned help
	helpBuilder := help.NewBuilder(r.styles.Help)
	bindings := r.getHelpBindings(currentSection)

	return helpBuilder.RenderWithBottomHelpAndHeader(content, bindings, width, height, 4)
}

// renderSectionNavigation renders the section tabs
func (r *Renderer) renderSectionNavigation(currentSection SettingsSection, width int) string {
	sections := []struct {
		key     SettingsSection
		display string
	}{
		{RepositoriesSection, "Repositories"},
		{ActionsSection, "Actions"},
		{ThemeSection, "Theme"},
	}

	var tabs []string
	for _, section := range sections {
		var style lipgloss.Style
		if section.key == currentSection {
			style = r.styles.SelectedItem
		} else {
			style = r.styles.Item
		}
		tabs = append(tabs, style.Render(section.display))
	}

	return strings.Join(tabs, " | ")
}

// renderRepositoriesSection renders the repositories settings section
func (r *Renderer) renderRepositoriesSection(repositories []*repomanager.RepoItem, cursor int, width int) string {
	var content string
	content += r.styles.SectionTitle.Render(fmt.Sprintf("Repositories (%d):", len(repositories))) + "\n\n"

	if len(repositories) == 0 {
		content += r.styles.EmptyState.Render("No repositories configured. Press 'e' to explore and add repositories.") + "\n\n"
		return content
	}

	for i, repo := range repositories {
		isSelected := i == cursor
		content += r.renderRepositoryItem(repo, isSelected, width)
	}

	return content
}

// renderActionsSection renders the actions settings section
func (r *Renderer) renderActionsSection(actions []config.Action, cursor int, width int) string {
	var content string
	content += r.styles.SectionTitle.Render(fmt.Sprintf("Custom Actions (%d):", len(actions))) + "\n\n"

	if len(actions) == 0 {
		content += r.styles.EmptyState.Render("No custom actions configured. Press 'a' to add a new action.") + "\n\n"
		return content
	}

	for i, action := range actions {
		isSelected := i == cursor
		content += r.renderActionItem(action, isSelected, width)
	}

	return content
}

// renderThemeSection renders the theme settings section
func (r *Renderer) renderThemeSection(themeConfig theme.Theme, cursor int, width int) string {
	var content string
	content += r.styles.SectionTitle.Render("Theme Settings:") + "\n\n"

	themeItems := []struct {
		name  string
		value string
	}{
		{"Primary Color", themeConfig.Colors.Selected},
		{"Title Color", themeConfig.Colors.Title},
		{"Help Color", themeConfig.Colors.Help},
		{"Error Color", themeConfig.Colors.StatusError},
		{"Branch Color", themeConfig.Colors.Branch},
	}

	for i, item := range themeItems {
		isSelected := i == cursor
		content += r.renderThemeItem(item.name, item.value, isSelected, width)
	}

	return content
}

// renderRepositoryItem renders a single repository item
func (r *Renderer) renderRepositoryItem(repo *repomanager.RepoItem, isSelected bool, width int) string {
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

	return style.Render(repoLine) + "\n"
}

// renderActionItem renders a single action item
func (r *Renderer) renderActionItem(action config.Action, isSelected bool, width int) string {
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

	actionLine := fmt.Sprintf(" %s%s: %s - %s", frontIndicator, action.Key, action.Name, action.Command)

	return style.Render(actionLine) + "\n"
}

// renderThemeItem renders a single theme item
func (r *Renderer) renderThemeItem(name, value string, isSelected bool, width int) string {
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

	colorPreview := lipgloss.NewStyle().
		Background(lipgloss.Color(value)).
		Foreground(lipgloss.Color("#000000")).
		Render("  ")

	themeLine := fmt.Sprintf(" %s%s: %s %s", frontIndicator, name, colorPreview, value)

	return style.Render(themeLine) + "\n"
}

// getHelpBindings returns the help bindings for the current section
func (r *Renderer) getHelpBindings(currentSection SettingsSection) []help.KeyBinding {
	commonBindings := []help.KeyBinding{
		{Key: "[/]", Description: "switch tab"},
		{Key: "↑/↓", Description: "navigate"},
		{Key: "Esc", Description: "back"},
	}

	switch currentSection {
	case RepositoriesSection:
		return append(commonBindings, []help.KeyBinding{
			{Key: "e", Description: "explore"},
			{Key: "d", Description: "delete"},
			{Key: "r", Description: "refresh"},
			{Key: "Enter", Description: "details"},
		}...)
	case ActionsSection:
		return append(commonBindings, []help.KeyBinding{
			{Key: "a", Description: "add"},
			{Key: "e", Description: "edit"},
			{Key: "d", Description: "delete"},
		}...)
	case ThemeSection:
		return append(commonBindings, []help.KeyBinding{
			{Key: "e", Description: "edit"},
			{Key: "r", Description: "reset"},
		}...)
	default:
		return commonBindings
	}
}
