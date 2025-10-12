package explore

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/wip-tui/internal/theme"
)

// StyleConfig contains all the styles needed for explore page rendering
type StyleConfig struct {
	Item              lipgloss.Style
	SelectedItem      lipgloss.Style
	StatusUncommitted lipgloss.Style
	StatusUnpushed    lipgloss.Style
	StatusUntracked   lipgloss.Style
	StatusError       lipgloss.Style
	StatusClean       lipgloss.Style
	StatusNotAdded    lipgloss.Style
	Help              lipgloss.Style
	Branch            lipgloss.Style
	IconRegular       lipgloss.Style
	IconBare          lipgloss.Style
	IconWorktree      lipgloss.Style
}

// CreateStyleConfig creates a style configuration for the explore page
func CreateStyleConfig(themeConfig theme.Theme) StyleConfig {
	return StyleConfig{
		Item: lipgloss.NewStyle().
			PaddingLeft(2),
		SelectedItem: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color(themeConfig.Colors.Selected)).
			Bold(true),
		StatusUncommitted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.StatusDirty)).
			Bold(true),
		StatusUnpushed: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.StatusUnpushed)).
			Bold(true),
		StatusUntracked: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.StatusUntracked)).
			Bold(true),
		StatusError: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.StatusError)).
			Bold(true),
		StatusClean: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.StatusClean)).
			Bold(true),
		StatusNotAdded: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.StatusNotAdded)),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.Help)).
			Margin(1, 0),
		Branch: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.Branch)),
		IconRegular: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.IconRegular)).
			Bold(true),
		IconBare: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.IconBare)).
			Bold(true),
		IconWorktree: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.IconWorktree)).
			Bold(true),
	}
}
