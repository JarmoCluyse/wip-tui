package help

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/wip-tui/internal/ui/layout"
	"golang.org/x/term"
)

// KeyBinding represents a single key binding with its description
type KeyBinding struct {
	Key         string
	Description string
}

// Builder helps construct help text consistently across all pages
type Builder struct {
	style lipgloss.Style
}

// NewBuilder creates a new help text builder with the given style
func NewBuilder(style lipgloss.Style) *Builder {
	return &Builder{
		style: style,
	}
}

// BuildCompactHelp creates a compact help line with the given key bindings
// Always includes "q: quit" and "?: help" at the end
func (b *Builder) BuildCompactHelp(bindings []KeyBinding) string {
	var helpParts []string

	// Add provided bindings
	for _, binding := range bindings {
		helpParts = append(helpParts, binding.Key+": "+binding.Description)
	}

	// Always add standard endings
	helpParts = append(helpParts, "q: quit", "?: help")

	return b.style.Render(strings.Join(helpParts, "  "))
}

// RenderWithBottomHelp renders content with help positioned at the bottom
func (b *Builder) RenderWithBottomHelp(content string, bindings []KeyBinding, width, height int) string {
	return b.RenderWithBottomHelpAndHeader(content, bindings, width, height, 0)
}

// RenderWithBottomHelpAndHeader renders content with help positioned at the bottom, accounting for header lines
func (b *Builder) RenderWithBottomHelpAndHeader(content string, bindings []KeyBinding, width, height, headerLines int) string {
	helpText := b.BuildCompactHelp(bindings)

	// If height is 0, try multiple methods to detect terminal size
	if height == 0 {
		// Method 1: Try stdout file descriptor
		if fd := int(os.Stdout.Fd()); term.IsTerminal(fd) {
			if _, h, err := term.GetSize(fd); err == nil && h > 0 {
				height = h
			}
		}

		// Method 2: Try stderr file descriptor if stdout failed
		if height == 0 {
			if fd := int(os.Stderr.Fd()); term.IsTerminal(fd) {
				if _, h, err := term.GetSize(fd); err == nil && h > 0 {
					height = h
				}
			}
		}

		// Method 3: Try stdin file descriptor as last resort
		if height == 0 {
			if fd := int(os.Stdin.Fd()); term.IsTerminal(fd) {
				if _, h, err := term.GetSize(fd); err == nil && h > 0 {
					height = h
				}
			}
		}
	}

	// Always use HeightCalculator for consistent height calculations
	// It handles the case where height=0 by using a sensible default
	calc := layout.NewHeightCalculator()

	// Validate terminal height (this will use default if height is 0)
	correctedHeight, _ := calc.ValidateTerminalHeight(height)

	// Calculate content area height (reserve 1 line for help + header lines)
	contentAreaHeight, _ := calc.CalculateContentAreaHeight(correctedHeight, 1+headerLines)

	// Count actual content lines
	contentLines := calc.CountContentLines(content)

	// Calculate padding needed to position help at bottom
	paddingLines := calc.CalculatePaddingLines(contentLines, contentAreaHeight)

	// Split content into lines and truncate if necessary
	lines := strings.Split(content, "\n")
	if len(lines) > contentAreaHeight {
		lines = lines[:contentAreaHeight]
		// When content is truncated, no padding is needed - help goes right after content
		paddingLines = 0
	}

	// Build result: content + padding + help
	result := strings.Join(lines, "\n")

	// Add padding newlines to push help to bottom (only if paddingLines > 0)
	if paddingLines > 0 {
		for i := 0; i < paddingLines; i++ {
			result += "\n"
		}
	}

	// Add help at the bottom
	result += helpText

	return result
}

// BuildDetailedHelp creates detailed help text (for help modal)
func (b *Builder) BuildDetailedHelp(sections []HelpSection) string {
	var content strings.Builder

	for i, section := range sections {
		if i > 0 {
			content.WriteString("\n\n")
		}

		content.WriteString(b.style.Bold(true).Render(section.Title))
		content.WriteString("\n")

		for _, binding := range section.Bindings {
			content.WriteString(b.style.Render("  " + binding.Key + ": " + binding.Description))
			content.WriteString("\n")
		}
	}

	return content.String()
}

// HelpSection represents a section in detailed help
type HelpSection struct {
	Title    string
	Bindings []KeyBinding
}

// Common key bindings that can be reused
var (
	NavigationBindings = []KeyBinding{
		{"↑/↓", "navigate"},
		{"k/j", "navigate"},
	}

	StandardBindings = []KeyBinding{
		{"Enter", "select"},
		{"Esc", "back"},
	}

	QuitBinding = KeyBinding{"q", "quit"}
	HelpBinding = KeyBinding{"?", "help"}
)
