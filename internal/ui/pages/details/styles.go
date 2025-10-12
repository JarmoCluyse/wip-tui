package details

import "github.com/charmbracelet/lipgloss"

// StyleConfig holds the styling configuration for the details page
type StyleConfig struct {
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	Help         lipgloss.Style
	Border       lipgloss.Style
	Title        lipgloss.Style
}
