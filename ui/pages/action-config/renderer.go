package actionconfig

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/theme"
	"github.com/jarmocluyse/git-dash/ui/components/help"
	"github.com/jarmocluyse/git-dash/ui/header"
)

// Renderer handles rendering of the action configuration page
type Renderer struct {
	styles StyleConfig
	theme  theme.Theme
	header *header.Renderer
}

// NewRenderer creates a new action config page renderer
func NewRenderer(styles StyleConfig, themeConfig theme.Theme) *Renderer {
	return &Renderer{
		styles: styles,
		theme:  themeConfig,
		header: header.NewRenderer(themeConfig),
	}
}

// Render renders the action configuration page
func (r *Renderer) Render(actions []config.Action, cursor int, width, height int) string {
	content := r.header.RenderWithSpacing("Action Configuration", width)

	if len(actions) == 0 {
		content += r.styles.EmptyState.Render("No custom actions configured. Press 'a' to add a new action.") + "\n\n"
	} else {
		content += r.renderActionList(actions, cursor, width)
	}

	// Use help component to render with bottom-aligned help
	helpBuilder := help.NewBuilder(r.styles.Help)
	bindings := []help.KeyBinding{
		{Key: "↑/↓", Description: "navigate"},
		{Key: "a", Description: "add action"},
		{Key: "e", Description: "edit action"},
		{Key: "d", Description: "delete action"},
		{Key: "Enter", Description: "edit action"},
		{Key: "Esc", Description: "back"},
	}

	return helpBuilder.RenderWithBottomHelpAndHeader(content, bindings, width, height, 3)
}

// RenderActionEditor renders the action editor interface
func (r *Renderer) RenderActionEditor(action *config.Action, currentField int, width, height int, isNew bool) string {
	title := "Edit Action"
	if isNew {
		title = "Add New Action"
	}

	content := r.header.RenderWithSpacing(title, width)

	fields := []struct {
		label string
		value string
		hint  string
	}{
		{"Name", action.Name, "Display name for the action"},
		{"Key", action.Key, "Key binding (e.g. 'l', 'o', 'ctrl+o')"},
		{"Description", action.Description, "Brief description of what this action does"},
		{"Command", action.Command, "Command to execute (use {path} for repository path)"},
		{"Args", strings.Join(action.Args, " "), "Arguments (space-separated, use {path} for repository path)"},
	}

	for i, field := range fields {
		isSelected := i == currentField

		var style = r.styles.Input
		var promptStyle = r.styles.InputPrompt
		if isSelected {
			style = r.styles.SelectedItem
			promptStyle = r.styles.SelectedItem
		}

		content += promptStyle.Render(fmt.Sprintf("%s:", field.label)) + "\n"
		content += style.Render(fmt.Sprintf("  %s", field.value)) + "\n"
		content += r.styles.Help.Render(fmt.Sprintf("  %s", field.hint)) + "\n\n"
	}

	// Use help component to render with bottom-aligned help
	helpBuilder := help.NewBuilder(r.styles.Help)
	bindings := []help.KeyBinding{
		{Key: "↑/↓", Description: "navigate fields"},
		{Key: "Enter", Description: "edit field"},
		{Key: "Ctrl+S", Description: "save"},
		{Key: "Esc", Description: "cancel"},
	}

	return helpBuilder.RenderWithBottomHelpAndHeader(content, bindings, width, height, 3)
}

// renderActionList renders the list of configured actions
func (r *Renderer) renderActionList(actions []config.Action, cursor int, width int) string {
	var content string
	content += r.styles.SectionTitle.Render(fmt.Sprintf("Configured Actions (%d):", len(actions))) + "\n\n"

	for i, action := range actions {
		content += r.renderActionItem(action, i, cursor, width)
	}

	content += "\n"
	return content
}

// renderActionItem renders a single action item
func (r *Renderer) renderActionItem(action config.Action, index, cursor int, width int) string {
	isSelected := index == cursor

	var style = r.styles.Item
	if isSelected {
		style = r.styles.SelectedItem
	}

	// Format: "[key] name - description"
	var frontIndicator string
	if isSelected {
		frontIndicator = r.theme.Indicators.Selected
	} else {
		frontIndicator = strings.Repeat(" ", lipgloss.Width(r.theme.Indicators.Selected))
	}

	keyPart := r.styles.ActionKey.Render(fmt.Sprintf("[%s]", action.Key))
	namePart := action.Name
	descPart := ""
	if action.Description != "" {
		descPart = " - " + r.styles.ActionDesc.Render(action.Description)
	}

	actionLine := fmt.Sprintf(" %s%s %s%s", frontIndicator, keyPart, namePart, descPart)

	// Calculate padding to right-align the end indicator with proper width constraints
	actionWidth := len(actionLine) // Use basic length for action content

	// Always reserve space for end indicator and spacing, even when not selected
	endIndicatorWidth := len(r.theme.Indicators.SelectedEnd)
	spacingWidth := 1 // One space before end indicator
	reservedEndWidth := endIndicatorWidth + spacingWidth

	// Ensure we don't exceed the terminal width - leave some margin
	maxContentWidth := width - 2 // margin for safety
	requiredWidth := actionWidth + reservedEndWidth

	var padding int
	if requiredWidth >= maxContentWidth {
		// Content too wide, use minimal spacing
		padding = 1
	} else {
		padding = maxContentWidth - actionWidth - reservedEndWidth
	}

	if padding < 0 {
		padding = 0
	}

	// Always add the reserved space, but only show the indicator when selected
	var endIndicator string
	if isSelected {
		styledEndIndicator := lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.Selected)).Render(r.theme.Indicators.SelectedEnd)
		endIndicator = " " + styledEndIndicator
	} else {
		// Reserve the space with spaces to prevent layout shift
		endIndicator = strings.Repeat(" ", endIndicatorWidth+1)
	}

	actionLine = actionLine + strings.Repeat(" ", padding) + endIndicator

	// Show command on second line
	commandLine := fmt.Sprintf("   %s", r.styles.ActionCommand.Render(r.formatCommand(action)))

	content := style.Render(actionLine) + "\n"
	content += style.Render(commandLine) + "\n"

	return content
}

// formatCommand formats the command and args for display
func (r *Renderer) formatCommand(action config.Action) string {
	if len(action.Args) == 0 {
		return action.Command
	}
	return fmt.Sprintf("%s %s", action.Command, strings.Join(action.Args, " "))
}
