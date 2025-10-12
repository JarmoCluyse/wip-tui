package repomanagement

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/wip-tui/internal/theme"
)

// StyleConfig contains all the styles needed for repo management page rendering
type StyleConfig struct {
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
	EmptyState   lipgloss.Style
	SectionTitle lipgloss.Style
	Help         lipgloss.Style
}

// CreateStyleConfig creates a style configuration for the repo management page
func CreateStyleConfig(themeConfig theme.Theme) StyleConfig {
	return StyleConfig{
		Item: lipgloss.NewStyle().
			PaddingLeft(2),
		SelectedItem: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color(themeConfig.Colors.Selected)).
			Bold(true),
		EmptyState: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.Help)).
			Italic(true).
			Margin(1, 0),
		SectionTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.Title)).
			Bold(true).
			Margin(0, 0, 1, 0),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.Help)).
			Margin(2, 0, 0, 0),
	}
}
