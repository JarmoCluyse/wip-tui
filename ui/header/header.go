// Package header provides application header rendering functionality.
package header

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/git-dash/internal/theme/types"
)

// Renderer handles rendering of application headers/titles.
type Renderer struct {
	titleStyle lipgloss.Style
}

// NewRenderer creates a new header renderer with the given theme.
func NewRenderer(themeConfig types.Theme) *Renderer {
	return &Renderer{
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(themeConfig.Colors.Title)).
			Background(lipgloss.Color(themeConfig.Colors.TitleBackground)).
			Padding(0, 1),
	}
}

// Render renders a header with the given title text and width.
func (h *Renderer) Render(title string, width int) string {
	if width <= 0 {
		width = 80 // Default width
	}
	return h.titleStyle.Width(width).Render(title)
}

// RenderWithCount renders a header with title on left and count on right.
func (h *Renderer) RenderWithCount(appName, configTitle string, count int, width int) string {
	if width <= 0 {
		width = 80 // Default width
	}

	// Build left side: "git-dash" and optional config title
	leftContent := appName
	if configTitle != "" {
		leftContent = fmt.Sprintf("%s - %s", appName, configTitle)
	}

	// Build right side: repo count
	rightContent := fmt.Sprintf("(%d)", count)

	// Calculate available space for spacing
	totalContentWidth := len(leftContent) + len(rightContent)
	availableWidth := width - 2 // Account for padding

	var content string
	if totalContentWidth >= availableWidth {
		// Not enough space, just use left content
		content = leftContent
	} else {
		// Add spacing between left and right content
		spacing := availableWidth - totalContentWidth
		content = leftContent + strings.Repeat(" ", spacing) + rightContent
	}

	return h.titleStyle.Width(width).Render(content)
}

// RenderWithSpacing renders a header with title and adds 2 newlines below.
// Returns 3 total lines: 1 for header + 2 for spacing.
func (h *Renderer) RenderWithSpacing(title string, width int) string {
	return h.Render(title, width) + "\n\n"
}

// RenderWithCountAndSpacing renders a header with count and adds 1 newline below.
func (h *Renderer) RenderWithCountAndSpacing(appName, configTitle string, count int, width int) string {
	return h.RenderWithCount(appName, configTitle, count, width) + "\n"
}
