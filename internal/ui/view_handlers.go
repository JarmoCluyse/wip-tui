package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/wip-tui/internal/theme"
	"github.com/jarmocluyse/wip-tui/internal/ui/pages/details"
	"github.com/jarmocluyse/wip-tui/internal/ui/types"
)

// View renders the current view based on the model's state.
func (m Model) View() string {
	// Get the main view content
	var mainView string
	switch m.State {
	case ListView:
		mainView = m.renderListView()
	case RepoManagementView:
		mainView = m.renderRepoManagementView()
	case ExplorerView:
		mainView = m.renderExplorerView()
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
	allItems := m.buildNavigableItems()

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

	// Get the cached git checker
	gitChecker := m.Dependencies.GetGitChecker()

	return renderer.RenderNavigable(visibleItems, relativeCursor, m.Width, m.Height, gitChecker, m.Config.Keybindings.Actions)
}

// renderRepoManagementView renders the repository management view.
func (m Model) renderRepoManagementView() string {
	styles := CreateStyleConfig(m.Config.Theme)
	renderer := NewRepoManagementViewRenderer(styles, m.Config.Theme)
	repositories := m.RepoHandler.GetRepositories()
	return renderer.Render(repositories, m.Cursor, m.Width, m.Height)
}

// renderExplorerView renders the directory explorer view.
func (m Model) renderExplorerView() string {
	styles := CreateStyleConfig(m.Config.Theme)
	renderer := NewExplorerViewRenderer(styles, m.Config.Theme)
	return renderer.Render(m.ExplorerPath, m.ExplorerItems, m.ExplorerCursor, m.Width, m.Height)
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

	// Help content with keybindings
	helpContent := strings.Builder{}

	// General navigation
	helpContent.WriteString("GENERAL NAVIGATION:\n")
	helpContent.WriteString("  ↑/k           Navigate up\n")
	helpContent.WriteString("  ↓/j           Navigate down\n")
	helpContent.WriteString("  Enter         Select/Confirm\n")
	helpContent.WriteString("  Esc/q         Go back/Cancel\n")
	helpContent.WriteString("  Ctrl+C        Quit application\n")
	helpContent.WriteString("  ?             Toggle this help\n\n")

	// Repository list view
	helpContent.WriteString("REPOSITORY LIST:\n")
	helpContent.WriteString("  a             Add new repository\n")
	helpContent.WriteString("  r/F5          Refresh statuses\n")
	helpContent.WriteString("  d             Remove repository\n")
	helpContent.WriteString("  l             Open in Lazygit\n")
	helpContent.WriteString("  e             Browse directories\n\n")

	// Explorer view
	helpContent.WriteString("DIRECTORY EXPLORER:\n")
	helpContent.WriteString("  Space         Toggle repository selection\n")
	helpContent.WriteString("  l             Open directory in Lazygit\n\n")

	// Status indicators
	helpContent.WriteString("STATUS INDICATORS:\n")
	helpContent.WriteString(fmt.Sprintf("  %s            Clean (no changes)\n", m.Config.Theme.Indicators.Clean))
	helpContent.WriteString(fmt.Sprintf("  %s            Uncommitted changes\n", m.Config.Theme.Indicators.Dirty))
	helpContent.WriteString(fmt.Sprintf("  %s            Unpushed commits\n", m.Config.Theme.Indicators.Unpushed))
	helpContent.WriteString(fmt.Sprintf("  %s            Untracked files\n", m.Config.Theme.Indicators.Untracked))
	helpContent.WriteString(fmt.Sprintf("  %s            Error accessing repository\n", m.Config.Theme.Indicators.Error))

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

	// Use lipgloss.Place to overlay the modal on top of the background
	// This creates a proper overlay effect where the background is preserved
	// and the modal is centered on top
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, styledModal, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceForeground(lipgloss.NoColor{}))
}

// renderActionConfigView renders the action configuration view.
func (m Model) renderActionConfigView() string {
	styles := CreateStyleConfig(m.Config.Theme)
	renderer := NewActionConfigViewRenderer(styles, m.Config.Theme)
	return renderer.Render(m.Config.Keybindings.Actions, m.ActionConfigCursor, m.ActionConfigEditMode, m.ActionConfigAction, m.ActionConfigFieldIdx, m.Width, m.Height)
}
