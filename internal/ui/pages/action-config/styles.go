package actionconfig

import "github.com/charmbracelet/lipgloss"

// StyleConfig holds all styling configuration for the action config page.
type StyleConfig struct {
	Item          lipgloss.Style
	SelectedItem  lipgloss.Style
	SectionTitle  lipgloss.Style
	Help          lipgloss.Style
	Border        lipgloss.Style
	EmptyState    lipgloss.Style
	Input         lipgloss.Style
	InputPrompt   lipgloss.Style
	ActionKey     lipgloss.Style
	ActionCommand lipgloss.Style
	ActionDesc    lipgloss.Style
}
