package settings

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/repomanager"
	"github.com/jarmocluyse/git-dash/internal/theme"
	"github.com/jarmocluyse/git-dash/ui/components/direxplorer"
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
func (r *Renderer) Render(data SettingsData, currentSection SettingsSection, cursor int, width, height int, themeEditMode bool, themeEditValue string, actionEditMode bool, actionEditValue string, actionEditFieldType string, actionEditItemIndex int, repoActiveSection string, repoExplorer *direxplorer.Explorer, repoPasteMode bool, repoPasteValue string) string {
	content := r.header.RenderWithCountAndSpacing("git-dash", "", 1, width)
	content += r.header.RenderWithSpacing("Settings", width)

	content += r.renderSectionNavigation(currentSection, width)
	content += "\n"

	switch currentSection {
	case RepositoriesSection:
		content += r.renderRepositoriesSection(data.Repositories, cursor, width, height, repoActiveSection, repoExplorer, repoPasteMode, repoPasteValue)
	case ActionsSection:
		content += r.renderActionsSection(data.Actions, cursor, width, actionEditMode, actionEditValue, actionEditFieldType, actionEditItemIndex)
	case ThemeSection:
		content += r.renderThemeSection(data.Theme, cursor, width, themeEditMode, themeEditValue)
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

// renderRepositoriesSection renders the repositories settings section with two-part layout
func (r *Renderer) renderRepositoriesSection(repositories []*repomanager.RepoItem, cursor int, width, height int, activeSection string, explorer *direxplorer.Explorer, pasteMode bool, pasteValue string) string {
	// Calculate layout - split the width roughly in half
	leftWidth := width / 2
	rightWidth := width - leftWidth - 3 // Account for separator

	// Create left side - repositories list
	leftContent := r.renderRepositoriesList(repositories, cursor, leftWidth, activeSection == "list")

	// Create right side - explorer and paste input
	rightContent := r.renderRepositoryAdder(explorer, pasteMode, pasteValue, rightWidth, height, activeSection)

	// Combine left and right with separator
	return r.combineSideBySide(leftContent, rightContent, leftWidth, rightWidth)
}

// renderActionsSection renders the actions settings section with edit support
func (r *Renderer) renderActionsSection(actions []config.Action, cursor int, width int, actionEditMode bool, actionEditValue string, actionEditFieldType string, actionEditItemIndex int) string {
	var content string
	content += r.styles.SectionTitle.Render(fmt.Sprintf("Custom Actions (%d):", len(actions))) + "\n\n"

	if len(actions) == 0 {
		content += r.styles.EmptyState.Render("No custom actions configured. Press 'a' to add a new action.") + "\n\n"
		return content
	}

	for i, action := range actions {
		isSelected := i == cursor
		if actionEditMode && isSelected && actionEditItemIndex == i {
			// Show edit mode for this action
			content += r.renderActionItemEdit(action, actionEditValue, actionEditFieldType, width)
		} else {
			content += r.renderActionItem(action, isSelected, width)
		}
	}

	return content
}

// renderThemeSection renders the theme settings section with full editor
func (r *Renderer) renderThemeSection(themeConfig theme.Theme, cursor int, width int, themeEditMode bool, themeEditValue string) string {
	var content string
	content += r.styles.SectionTitle.Render("Theme Editor:") + "\n\n"

	// Create all theme items in categories
	themeItems := r.getAllThemeItems(themeConfig)

	if len(themeItems) == 0 {
		content += r.styles.EmptyState.Render("No theme items available.") + "\n\n"
		return content
	}

	// Group items by category for display
	categories := r.groupThemeItemsByCategory(themeItems)

	itemIndex := 0
	for _, category := range []string{"Colors", "Status Indicators", "Repository Icons", "UI Icons"} {
		if items, exists := categories[category]; exists {
			// Category header
			content += r.styles.SectionTitle.Render(fmt.Sprintf("%s:", category)) + "\n"

			// Render items in this category
			for _, item := range items {
				isSelected := itemIndex == cursor
				if themeEditMode && isSelected {
					// Show edit mode for this item (2 lines)
					content += r.renderThemeItemEdit(item.name, themeEditValue, item.itemType, width)
				} else {
					// Show normal mode (1 line)
					content += r.renderThemeItem(item.name, item.value, item.itemType, isSelected, width)
				}
				itemIndex++
			}
			content += "\n"
		}
	}

	return content
}

// ThemeItem represents a single editable theme item
type ThemeItem struct {
	name     string
	value    string
	itemType string // "color", "icon", "indicator"
	category string
}

// getAllThemeItems returns all editable theme items
func (r *Renderer) getAllThemeItems(themeConfig theme.Theme) []ThemeItem {
	var items []ThemeItem

	// Colors
	items = append(items, []ThemeItem{
		{"Title", themeConfig.Colors.Title, "color", "Colors"},
		{"Title Background", themeConfig.Colors.TitleBackground, "color", "Colors"},
		{"Selected", themeConfig.Colors.Selected, "color", "Colors"},
		{"Selected Background", themeConfig.Colors.SelectedBackground, "color", "Colors"},
		{"Help Text", themeConfig.Colors.Help, "color", "Colors"},
		{"Border", themeConfig.Colors.Border, "color", "Colors"},
		{"Modal Background", themeConfig.Colors.ModalBackground, "color", "Colors"},
		{"Branch", themeConfig.Colors.Branch, "color", "Colors"},
		{"Regular Icon", themeConfig.Colors.IconRegular, "color", "Colors"},
		{"Bare Icon", themeConfig.Colors.IconBare, "color", "Colors"},
		{"Worktree Icon", themeConfig.Colors.IconWorktree, "color", "Colors"},
	}...)

	// Status colors and indicators
	items = append(items, []ThemeItem{
		{"Clean Status Color", themeConfig.Colors.StatusClean, "color", "Status Indicators"},
		{"Clean Status Icon", themeConfig.Indicators.Clean, "indicator", "Status Indicators"},
		{"Dirty Status Color", themeConfig.Colors.StatusDirty, "color", "Status Indicators"},
		{"Dirty Status Icon", themeConfig.Indicators.Dirty, "indicator", "Status Indicators"},
		{"Unpushed Status Color", themeConfig.Colors.StatusUnpushed, "color", "Status Indicators"},
		{"Unpushed Status Icon", themeConfig.Indicators.Unpushed, "indicator", "Status Indicators"},
		{"Untracked Status Color", themeConfig.Colors.StatusUntracked, "color", "Status Indicators"},
		{"Untracked Status Icon", themeConfig.Indicators.Untracked, "indicator", "Status Indicators"},
		{"Error Status Color", themeConfig.Colors.StatusError, "color", "Status Indicators"},
		{"Error Status Icon", themeConfig.Indicators.Error, "indicator", "Status Indicators"},
		{"Not Added Status Color", themeConfig.Colors.StatusNotAdded, "color", "Status Indicators"},
		{"Not Added Status Icon", themeConfig.Indicators.NotAdded, "indicator", "Status Indicators"},
	}...)

	// Repository icons
	items = append(items, []ThemeItem{
		{"Regular Repository", themeConfig.Icons.Repository.Regular, "icon", "Repository Icons"},
		{"Bare Repository", themeConfig.Icons.Repository.Bare, "icon", "Repository Icons"},
		{"Worktree Repository", themeConfig.Icons.Repository.Worktree, "icon", "Repository Icons"},
	}...)

	// UI icons
	items = append(items, []ThemeItem{
		{"Selected Indicator", themeConfig.Indicators.Selected, "indicator", "UI Icons"},
		{"Selected End", themeConfig.Indicators.SelectedEnd, "indicator", "UI Icons"},
		{"Branch Icon", themeConfig.Icons.Branch.Icon, "icon", "UI Icons"},
		{"Tree Branch", themeConfig.Icons.Tree.Branch, "icon", "UI Icons"},
		{"Tree Last", themeConfig.Icons.Tree.Last, "icon", "UI Icons"},
		{"Folder Icon", themeConfig.Icons.Folder.Icon, "icon", "UI Icons"},
	}...)

	return items
}

// groupThemeItemsByCategory groups theme items by their category
func (r *Renderer) groupThemeItemsByCategory(items []ThemeItem) map[string][]ThemeItem {
	categories := make(map[string][]ThemeItem)
	for _, item := range items {
		categories[item.category] = append(categories[item.category], item)
	}
	return categories
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

// renderRepositoriesList renders the left side repositories list
func (r *Renderer) renderRepositoriesList(repositories []*repomanager.RepoItem, cursor int, width int, isActive bool) string {
	var content string

	// Title with active indicator
	titleStyle := r.styles.SectionTitle
	if isActive {
		titleStyle = titleStyle.Foreground(lipgloss.Color(r.theme.Colors.Selected))
	}
	content += titleStyle.Render(fmt.Sprintf("Repositories (%d)", len(repositories))) + "\n\n"

	if len(repositories) == 0 {
		content += r.styles.EmptyState.Render("No repositories configured") + "\n"
		content += r.styles.EmptyState.Render("Use explorer to add repos →") + "\n\n"
		return content
	}

	for i, repo := range repositories {
		isSelected := i == cursor && isActive
		content += r.renderRepositoryItem(repo, isSelected, width)
	}

	return content
}

// renderRepositoryAdder renders the right side explorer and paste input
func (r *Renderer) renderRepositoryAdder(explorer *direxplorer.Explorer, pasteMode bool, pasteValue string, width, height int, activeSection string) string {
	var content string

	// Title with active indicator
	titleStyle := r.styles.SectionTitle
	if activeSection == "explorer" || activeSection == "paste" {
		titleStyle = titleStyle.Foreground(lipgloss.Color(r.theme.Colors.Selected))
	}
	content += titleStyle.Render("Add Repository") + "\n\n"

	// Directory Explorer section
	explorerHeight := height - 8 // Reserve space for title, paste section, and padding
	if explorerHeight < 5 {
		explorerHeight = 5
	}

	if explorer != nil {
		explorerContent := explorer.Render(width, explorerHeight)
		if activeSection == "explorer" {
			// Add border to show it's active
			explorerContent = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(r.theme.Colors.Selected)).
				Padding(0, 1).
				Render(explorerContent)
		}
		content += explorerContent + "\n"
	} else {
		content += r.styles.EmptyState.Render("Explorer not available") + "\n"
	}

	// Separator
	content += r.styles.Item.Render(strings.Repeat("─", width-2)) + "\n\n"

	// Paste input section
	pasteTitle := "Paste Path"
	if activeSection == "paste" {
		pasteTitle = "→ " + pasteTitle
	}
	content += r.styles.SectionTitle.Render(pasteTitle) + "\n"

	// Help text for paste mode
	if !pasteMode {
		helpText := "Press 'a' to add repository by path"
		content += r.styles.Help.Render(helpText) + "\n"
	}

	// Input field
	inputValue := pasteValue
	if pasteMode && activeSection == "paste" {
		inputValue += "│"
	} else if !pasteMode {
		inputValue = "[Enter path here]"
	}

	inputStyle := r.styles.Item
	if activeSection == "paste" {
		inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(r.theme.Colors.Selected)).
			Padding(0, 1).
			Width(width - 4)
	} else {
		inputStyle = inputStyle.Width(width - 4)
	}

	content += inputStyle.Render(inputValue) + "\n"

	return content
}

// combineSideBySide combines left and right content side by side
func (r *Renderer) combineSideBySide(left, right string, leftWidth, rightWidth int) string {
	leftLines := strings.Split(strings.TrimRight(left, "\n"), "\n")
	rightLines := strings.Split(strings.TrimRight(right, "\n"), "\n")

	// Ensure both sides have the same number of lines
	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	// Pad shorter side with empty lines
	for len(leftLines) < maxLines {
		leftLines = append(leftLines, "")
	}
	for len(rightLines) < maxLines {
		rightLines = append(rightLines, "")
	}

	var result []string
	separator := " │ "

	for i := 0; i < maxLines; i++ {
		// Ensure left side is exactly leftWidth
		leftLine := leftLines[i]
		if lipgloss.Width(leftLine) > leftWidth {
			leftLine = leftLine[:leftWidth-3] + "..."
		} else {
			leftLine = leftLine + strings.Repeat(" ", leftWidth-lipgloss.Width(leftLine))
		}

		combinedLine := leftLine + separator + rightLines[i]
		result = append(result, combinedLine)
	}

	return strings.Join(result, "\n") + "\n"
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

	// Format command with args for display
	commandDisplay := action.Command
	if len(action.Args) > 0 {
		commandDisplay += " " + strings.Join(action.Args, " ")
	}

	actionLine := fmt.Sprintf(" %s%s: %s - %s", frontIndicator, action.Key, action.Name, commandDisplay)

	return style.Render(actionLine) + "\n"
}

// renderActionItemEdit renders an action item in edit mode with field-specific input
func (r *Renderer) renderActionItemEdit(action config.Action, editValue string, fieldType string, width int) string {
	style := r.styles.SelectedItem
	frontIndicator := r.theme.Indicators.Selected

	// Create a multi-line layout for better editing experience
	var content strings.Builder

	// Header line with action indicator
	content.WriteString(style.Render(fmt.Sprintf(" %sEditing Action:", frontIndicator)) + "\n")

	// Field definitions with current values and edit indicators
	fields := []struct {
		label       string
		value       string
		isEditing   bool
		fieldType   string
		description string
	}{
		{"Name", action.Name, fieldType == "name", "name", "Display name for the action"},
		{"Key", action.Key, fieldType == "key", "key", "Keyboard shortcut (e.g., 'g', 'ctrl+r')"},
		{"Command", action.Command, fieldType == "command", "command", "Shell command to execute"},
		{"Args", strings.Join(action.Args, " "), fieldType == "args", "args", "Arguments for the command (space-separated)"},
	}

	for i, field := range fields {
		// Field label with progress indicator
		fieldNum := i + 1
		totalFields := len(fields)
		progressIndicator := fmt.Sprintf("[%d/%d]", fieldNum, totalFields)

		var fieldStyle lipgloss.Style
		var displayValue string

		if field.isEditing {
			// Currently editing this field
			fieldStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(r.theme.Colors.Selected)).
				Bold(true)

			// Show edit field with border and cursor
			editField := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(r.theme.Colors.Selected)).
				Background(lipgloss.Color(r.theme.Colors.ModalBackground)).
				Padding(0, 1).
				Width(30).
				Render(editValue + "│")

			displayValue = editField

			// Add field description below for the active field
			descStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(r.theme.Colors.Help)).
				Italic(true)
			content.WriteString(style.Render(fmt.Sprintf("   %s %s %s:", progressIndicator, fieldStyle.Render("→"), field.label)) + "\n")
			content.WriteString(displayValue + "\n")
			content.WriteString(style.Render(fmt.Sprintf("     %s", descStyle.Render(field.description))) + "\n")
		} else {
			// Not editing - show current value
			fieldStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(r.theme.Colors.Help))

			if field.value == "" {
				displayValue = fieldStyle.Render("(empty)")
			} else {
				// Truncate long values for display
				maxLen := 40
				if len(field.value) > maxLen {
					displayValue = field.value[:maxLen] + "..."
				} else {
					displayValue = field.value
				}
			}

			content.WriteString(style.Render(fmt.Sprintf("   %s   %s: %s", progressIndicator, fieldStyle.Render(field.label), displayValue)) + "\n")
		}
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(r.theme.Colors.Help)).
		Italic(true)
	content.WriteString(style.Render(fmt.Sprintf("   %s", helpStyle.Render("Enter: next field • Tab: skip • Esc: cancel • Ctrl+S: save"))) + "\n")

	return content.String()
}

// renderEditField renders an input field with cursor indicator
func (r *Renderer) renderEditField(value string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(r.theme.Colors.Selected)).
		Padding(0, 1).
		Render(value + "|")
}

// renderThemeItem renders a single theme item with appropriate preview (single line)
func (r *Renderer) renderThemeItem(name, value, itemType string, isSelected bool, width int) string {
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

	var preview string
	switch itemType {
	case "color":
		// Color preview with background
		preview = lipgloss.NewStyle().
			Background(lipgloss.Color(value)).
			Foreground(lipgloss.Color("#000000")).
			Render("  ")
	case "icon", "indicator":
		// Icon/indicator preview - show the actual symbol
		preview = lipgloss.NewStyle().
			Foreground(lipgloss.Color(r.theme.Colors.Selected)).
			Render(fmt.Sprintf("[%s]", value))
	default:
		preview = value
	}

	themeLine := fmt.Sprintf(" %s%s: %s %s", frontIndicator, name, preview, value)

	return style.Render(themeLine) + "\n"
}

// renderThemeItemEdit renders a theme item in edit mode with input field
func (r *Renderer) renderThemeItemEdit(name, editValue, itemType string, width int) string {
	style := r.styles.SelectedItem

	frontIndicator := r.theme.Indicators.Selected

	var preview string
	switch itemType {
	case "color":
		// Color preview with background using edit value
		preview = lipgloss.NewStyle().
			Background(lipgloss.Color(editValue)).
			Foreground(lipgloss.Color("#000000")).
			Render("  ")
	case "icon", "indicator":
		// Icon/indicator preview - show the actual symbol
		preview = lipgloss.NewStyle().
			Foreground(lipgloss.Color(r.theme.Colors.Selected)).
			Render(fmt.Sprintf("[%s]", editValue))
	default:
		preview = editValue
	}

	// Show input field with cursor
	editField := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(r.theme.Colors.Selected)).
		Padding(0, 1).
		Render(editValue + "|")

	// Show the theme item on first line, edit field on second line at the beginning
	baseLine := fmt.Sprintf(" %s%s: %s", frontIndicator, name, preview)

	// Put edit field at the beginning of the next line (no indentation)
	line1 := style.Render(baseLine)
	line2 := style.Render(editField)

	return line1 + "\n" + line2 + "\n"
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
			{Key: "Tab", Description: "switch section"},
			{Key: "Enter", Description: "select/add"},
			{Key: "d", Description: "delete"},
			{Key: "r", Description: "refresh"},
		}...)
	case ActionsSection:
		return append(commonBindings, []help.KeyBinding{
			{Key: "a", Description: "add"},
			{Key: "e", Description: "edit"},
			{Key: "d", Description: "delete"},
		}...)
	case ThemeSection:
		return append(commonBindings, []help.KeyBinding{
			{Key: "e", Description: "edit value"},
			{Key: "Enter", Description: "save (when editing)"},
			{Key: "Esc", Description: "cancel edit"},
		}...)
	default:
		return commonBindings
	}
}
