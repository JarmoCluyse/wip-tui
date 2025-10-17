package settings

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/git-dash/internal/theme/types"
)

// StyleConfig contains all the styles needed for settings page rendering
type StyleConfig struct {
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
	EmptyState   lipgloss.Style
	SectionTitle lipgloss.Style
	Help         lipgloss.Style
}

// CreateStyleConfig creates a style configuration for the settings page
func CreateStyleConfig(themeConfig types.Theme) StyleConfig {
	return StyleConfig{
		Item: lipgloss.NewStyle(),
		SelectedItem: lipgloss.NewStyle().
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
