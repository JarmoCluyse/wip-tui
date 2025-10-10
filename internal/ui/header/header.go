package header

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/wip-tui/internal/theme"
)

// Renderer handles rendering of application headers/titles
type Renderer struct {
	titleStyle lipgloss.Style
}

// NewRenderer creates a new header renderer with the given theme
func NewRenderer(themeConfig theme.Theme) *Renderer {
	return &Renderer{
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(themeConfig.Colors.Title)).
			Background(lipgloss.Color(themeConfig.Colors.TitleBackground)).
			Padding(0, 1),
	}
}

// Render renders a header with the given title text and width
func (h *Renderer) Render(title string, width int) string {
	if width <= 0 {
		width = 80 // Default width
	}
	return h.titleStyle.Width(width).Render(title)
}

// RenderWithSpacing renders a header with the given title and width, adds spacing below it
func (h *Renderer) RenderWithSpacing(title string, width int) string {
	return h.Render(title, width) + "\n\n"
}
