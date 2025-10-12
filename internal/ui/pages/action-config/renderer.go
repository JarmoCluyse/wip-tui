package actionconfig

import (
	"fmt"
	"strings"

	"github.com/jarmocluyse/wip-tui/internal/config"
	"github.com/jarmocluyse/wip-tui/internal/theme"
	"github.com/jarmocluyse/wip-tui/internal/ui/components/help"
	"github.com/jarmocluyse/wip-tui/internal/ui/header"
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
	cursorIndicator := r.getCursorIndicator(isSelected)

	var style = r.styles.Item
	if isSelected {
		style = r.styles.SelectedItem
	}

	// Format: "> [key] name - description"
	keyPart := r.styles.ActionKey.Render(fmt.Sprintf("[%s]", action.Key))
	namePart := action.Name
	descPart := ""
	if action.Description != "" {
		descPart = " - " + r.styles.ActionDesc.Render(action.Description)
	}

	actionLine := fmt.Sprintf("%s %s %s%s", cursorIndicator, keyPart, namePart, descPart)

	// Show command on second line
	commandLine := fmt.Sprintf("    %s", r.styles.ActionCommand.Render(r.formatCommand(action)))

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

// getCursorIndicator returns the appropriate cursor indicator
func (r *Renderer) getCursorIndicator(isSelected bool) string {
	if isSelected {
		return ">"
	}
	return " "
}
