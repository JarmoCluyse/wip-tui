package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	theme "github.com/jarmocluyse/git-dash/internal/theme/types"
	"github.com/jarmocluyse/git-dash/ui/pages/details"
	"github.com/jarmocluyse/git-dash/ui/pages/settings"
	"github.com/jarmocluyse/git-dash/ui/types"
)

// stripANSI removes ANSI escape sequences from a string to get the actual display length
func stripANSI(str string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(str, "")
}

// View renders the current view based on the model's state.
func (m Model) View() string {
	// Get the main view content
	var mainView string
	switch m.State {
	case ListView:
		mainView = m.renderListView()
	case SettingsView:
		mainView = m.renderSettingsView()
	case DetailsView:
		mainView = m.renderDetailsView()
	case ActionConfigView:
		mainView = m.renderActionConfigView()
	default:
		mainView = ""
	}

	// If help modal is open, overlay it on top
	if m.ShowHelpModal {
		return m.renderHelpModal(mainView)
	}

	return mainView
}

// renderListView renders the main repository list view.
func (m Model) renderListView() string {
	styles := CreateStyleConfig(m.Config.Theme)
	renderer := NewListViewRenderer(styles, m.Config.Theme)

	// Get all navigable items
	allItems := m.getNavigableItems()

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
	var visibleItems []types.NavigableItem
	if len(allItems) > 0 && start < end {
		visibleItems = allItems[start:end]
	}

	// Adjust cursor to be relative to visible window
	relativeCursor := m.Cursor - m.ScrollOffset

	// Get summary data from repo manager
	summaryData := m.Dependencies.GetRepoManager().GetSummary()
	configTitle := m.Config.Title

	return renderer.RenderNavigable(visibleItems, &summaryData, relativeCursor, m.Width, m.Height, m.Config.Keybindings.Actions, configTitle)
}

// renderSettingsView renders the settings view.
func (m Model) renderSettingsView() string {
	theme := m.Dependencies.GetThemeService().GetTheme()
	styles := CreateStyleConfig(*theme)
	renderer := NewSettingsRenderer(styles, *theme)

	// Prepare settings data
	data := settings.SettingsData{
		Repositories: m.Dependencies.GetRepoManager().GetItems(),
		Actions:      m.Config.Keybindings.Actions,
		Theme:        m.Config.Theme,
		Keybindings:  m.Config.Keybindings,
	}

	// Determine current section
	var currentSection settings.SettingsSection
	switch m.SettingsSection {
	case "actions":
		currentSection = settings.ActionsSection
	case "theme":
		currentSection = settings.ThemeSection
	default:
		currentSection = settings.RepositoriesSection
	}

	return renderer.Render(data, currentSection, m.SettingsCursor, m.Width, m.Height, m.ThemeEditMode, m.ThemeEditValue, m.ActionEditMode, m.ActionEditValue, m.ActionEditFieldType, m.ActionEditItemIndex, m.RepoActiveSection, m.RepoExplorer, m.RepoPasteMode, m.RepoPasteValue)
}

// NewDetailsViewRenderer creates a new details view renderer with the given styles and theme.
func NewDetailsViewRenderer(styles StyleConfig, themeConfig theme.Theme) *details.Renderer {
	detailsStyles := details.StyleConfig{
		Item:         styles.Item,
		SelectedItem: styles.SelectedItem,
		Label:        styles.Item.Foreground(lipgloss.Color(themeConfig.Colors.Selected)).Bold(true),
		Value:        styles.Item,
		Help:         styles.Help,
		Border:       styles.Border,
		Title:        styles.Item.Bold(true),
	}
	return details.NewRenderer(detailsStyles, themeConfig)
}

// renderDetailsView renders the repository details view.
func (m Model) renderDetailsView() string {
	if m.SelectedNavItem == nil {
		// Fallback to list view if no item is selected
		return m.renderListView()
	}

	styles := CreateStyleConfig(m.Config.Theme)
	renderer := NewDetailsViewRenderer(styles, m.Config.Theme)
	return renderer.Render(*m.SelectedNavItem, m.Width, m.Height)
}

// renderHelpModal renders the help modal overlay on top of the background view.
func (m Model) renderHelpModal(backgroundView string) string {
	styles := CreateStyleConfig(m.Config.Theme)

	// Help content with keybindings - make it view-specific
	helpContent := strings.Builder{}

	// General navigation (always shown)
	helpContent.WriteString("GENERAL NAVIGATION:\n")
	helpContent.WriteString("  ↑/k           Navigate up\n")
	helpContent.WriteString("  ↓/j           Navigate down\n")
	helpContent.WriteString("  Enter         Select/Confirm\n")
	helpContent.WriteString("  Esc/q         Go back/Cancel\n")
	helpContent.WriteString("  Ctrl+C        Quit application\n")
	helpContent.WriteString("  ?             Toggle this help\n\n")

	// View-specific sections
	switch m.State {
	case ListView:
		helpContent.WriteString("REPOSITORY LIST:\n")
		// Dynamically add configured actions
		for _, action := range m.Config.Keybindings.Actions {
			helpContent.WriteString(fmt.Sprintf("  %-13s %s\n", action.Key, action.Description))
		}
		helpContent.WriteString("  e             Open in file manager\n")
		helpContent.WriteString("  s             Settings\n")
		helpContent.WriteString("  r/F5          Refresh statuses\n")
		helpContent.WriteString("  w             Discover worktrees\n\n")
	case DetailsView:
		helpContent.WriteString("DETAILS VIEW:\n")
		helpContent.WriteString("  b/Esc         Back to list\n\n")
	case SettingsView:
		helpContent.WriteString("SETTINGS:\n")
		helpContent.WriteString("  [/]           Switch tabs\n")
		helpContent.WriteString("  Enter         View details (repos)\n")
		helpContent.WriteString("  a             Add action\n")
		helpContent.WriteString("  d             Delete repository\n")
		helpContent.WriteString("  e             Edit/Explore\n")
		helpContent.WriteString("  r             Refresh\n")
		helpContent.WriteString("  Esc           Back to list\n\n")
	case ActionConfigView:
		helpContent.WriteString("ACTION CONFIG:\n")
		helpContent.WriteString("  Enter/e       Edit action\n")
		helpContent.WriteString("  a             Add new action\n")
		helpContent.WriteString("  d             Delete action\n")
		helpContent.WriteString("  Esc           Back to settings\n\n")
	}

	// Status indicators (always shown)
	helpContent.WriteString("STATUS INDICATORS:\n")

	// Create colored status indicators using the theme colors
	cleanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.Config.Theme.Colors.StatusClean))
	dirtyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.Config.Theme.Colors.StatusDirty))
	unpushedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.Config.Theme.Colors.StatusUnpushed))
	untrackedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.Config.Theme.Colors.StatusUntracked))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.Config.Theme.Colors.StatusError))
	notAddedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.Config.Theme.Colors.StatusNotAdded))

	helpContent.WriteString(fmt.Sprintf("  %s            Clean (no changes)\n",
		cleanStyle.Render(m.Config.Theme.Indicators.Clean)))
	helpContent.WriteString(fmt.Sprintf("  %s            Uncommitted changes\n",
		dirtyStyle.Render(m.Config.Theme.Indicators.Dirty)))
	helpContent.WriteString(fmt.Sprintf("  %s            Unpushed commits\n",
		unpushedStyle.Render(m.Config.Theme.Indicators.Unpushed)))
	helpContent.WriteString(fmt.Sprintf("  %s            Untracked files\n",
		untrackedStyle.Render(m.Config.Theme.Indicators.Untracked)))
	helpContent.WriteString(fmt.Sprintf("  %s            Error accessing repository\n",
		errorStyle.Render(m.Config.Theme.Indicators.Error)))
	helpContent.WriteString(fmt.Sprintf("  %s            Not added to git\n",
		notAddedStyle.Render(m.Config.Theme.Indicators.NotAdded)))

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

	// Create a proper modal overlay by combining background and modal
	backgroundLines := strings.Split(backgroundView, "\n")

	// Ensure we have exactly the right number of background lines
	for len(backgroundLines) < height {
		backgroundLines = append(backgroundLines, "")
	}
	if len(backgroundLines) > height {
		backgroundLines = backgroundLines[:height]
	}

	// Prepare background lines by padding them to terminal width
	for i, line := range backgroundLines {
		runeCount := len([]rune(stripANSI(line))) // Handle multi-byte characters and ANSI codes properly
		if runeCount < width {
			backgroundLines[i] = line + strings.Repeat(" ", width-runeCount)
		} else if runeCount > width {
			runes := []rune(stripANSI(line))
			if len(runes) > width {
				backgroundLines[i] = string(runes[:width])
			}
		}
	}

	// Get modal dimensions - need to measure actual display width, not including ANSI codes
	modalLines := strings.Split(styledModal, "\n")
	modalHeight := len(modalLines)
	modalWidth := 80 // This matches the width set in styles.HelpModal

	// Calculate modal position (center)
	startY := (height - modalHeight) / 2
	startX := (width - modalWidth) / 2

	// Ensure modal doesn't go off screen
	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}
	if startY+modalHeight > height {
		startY = height - modalHeight
		if startY < 0 {
			startY = 0
		}
	}
	if startX+modalWidth > width {
		startX = width - modalWidth
		if startX < 0 {
			startX = 0
		}
	}

	// Overlay modal onto background - preserve background content with simpler approach
	for i, modalLine := range modalLines {
		y := startY + i
		if y >= 0 && y < len(backgroundLines) {
			// Get original background line and its plain text
			bgLine := backgroundLines[y]
			bgPlainText := stripANSI(bgLine)
			bgRunes := []rune(bgPlainText)

			// Get modal content as plain text for positioning
			modalPlainText := stripANSI(modalLine)
			modalRunes := []rune(modalPlainText)
			modalLineWidth := len(modalRunes)

			// Calculate the end position for the modal
			endX := startX + modalLineWidth
			if endX > width {
				endX = width
				modalLineWidth = width - startX
			}

			// Build the new line character by character
			var newLineRunes []rune

			// Before modal: preserve background characters
			for x := 0; x < startX && x < width; x++ {
				if x < len(bgRunes) {
					newLineRunes = append(newLineRunes, bgRunes[x])
				} else {
					newLineRunes = append(newLineRunes, ' ')
				}
			}

			// Modal area: use modal content (plain text)
			for x := 0; x < modalLineWidth && (startX+x) < width; x++ {
				if x < len(modalRunes) {
					newLineRunes = append(newLineRunes, modalRunes[x])
				} else {
					newLineRunes = append(newLineRunes, ' ')
				}
			}

			// After modal: preserve background characters
			for x := endX; x < width; x++ {
				if x < len(bgRunes) {
					newLineRunes = append(newLineRunes, bgRunes[x])
				} else {
					newLineRunes = append(newLineRunes, ' ')
				}
			}

			// Convert back to string and apply modal styling to just the modal portion
			plainNewLine := string(newLineRunes)

			// Now apply the modal styling correctly
			if strings.Contains(modalLine, "\x1b[") {
				// We need to apply modal styling to just the modal content area
				var finalLine strings.Builder

				// Add the part before modal (plain background content)
				if startX > 0 && len(newLineRunes) > startX {
					finalLine.WriteString(string(newLineRunes[:startX]))
				}

				// Add the styled modal content
				finalLine.WriteString(modalLine)

				// Add the part after modal (plain background content)
				if endX < len(newLineRunes) {
					finalLine.WriteString(string(newLineRunes[endX:]))
				}

				backgroundLines[y] = finalLine.String()
			} else {
				// No styling in modal, just use the plain overlay
				backgroundLines[y] = plainNewLine
			}
		}
	}

	return strings.Join(backgroundLines, "\n")
}

// renderActionConfigView renders the action configuration view.
func (m Model) renderActionConfigView() string {
	styles := CreateStyleConfig(m.Config.Theme)
	renderer := NewActionConfigRenderer(styles, m.Config.Theme)
	return renderer.Render(m.Config.Keybindings.Actions, m.ActionConfigCursor, m.Width, m.Height, "Action Configuration")
}
