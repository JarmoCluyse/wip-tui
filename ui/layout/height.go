package layout

import (
	"errors"
	"strings"
)

// HeightCalculator provides clean, testable height calculation functions
type HeightCalculator struct {
	minTerminalHeight int
	defaultHeight     int
}

// NewHeightCalculator creates a new height calculator with sensible defaults
func NewHeightCalculator() *HeightCalculator {
	return &HeightCalculator{
		minTerminalHeight: 10,
		defaultHeight:     40,
	}
}

// ValidateTerminalHeight ensures the terminal height is usable
func (h *HeightCalculator) ValidateTerminalHeight(height int) (int, error) {
	if height <= 0 {
		return h.defaultHeight, errors.New("invalid height: must be positive")
	}
	if height < h.minTerminalHeight {
		return h.minTerminalHeight, errors.New("height too small, using minimum")
	}
	return height, nil
}

// CalculateContentAreaHeight calculates available space for content
// Formula: totalHeight - reservedLines
func (h *HeightCalculator) CalculateContentAreaHeight(totalHeight, reservedLines int) (int, error) {
	validHeight, err := h.ValidateTerminalHeight(totalHeight)
	// Use the validated height even if there was an error (it's been corrected)

	contentHeight := validHeight - reservedLines
	if contentHeight < 1 {
		return 1, errors.New("content area too small")
	}
	return contentHeight, err
}

// CalculateHelpAreaHeight calculates space needed for help component
// Formula: 1 line for compact help
func (h *HeightCalculator) CalculateHelpAreaHeight() int {
	return 1 // Compact help always takes 1 line
}

// CountContentLines counts actual lines in content, handling ANSI escapes
func (h *HeightCalculator) CountContentLines(content string) int {
	if content == "" {
		return 0
	}

	// Split by newlines
	lines := strings.Split(content, "\n")

	// TODO: Handle ANSI escape sequences that might affect line counting
	// For now, return simple line count
	return len(lines)
}

// CalculatePaddingLines calculates padding needed to position help at bottom
func (h *HeightCalculator) CalculatePaddingLines(contentLines, availableHeight int) int {
	paddingNeeded := availableHeight - contentLines
	if paddingNeeded < 0 {
		return 0
	}
	return paddingNeeded
}

// CalculateVisibleItemCount calculates how many items fit in available height
func (h *HeightCalculator) CalculateVisibleItemCount(totalHeight, headerLines, helpLines, itemHeight int) (int, error) {
	contentHeight, err := h.CalculateContentAreaHeight(totalHeight, headerLines+helpLines)
	// Use the contentHeight even if there was an error (it's been corrected)

	if itemHeight <= 0 {
		itemHeight = 1 // Default item height
	}

	visibleItems := contentHeight / itemHeight
	if visibleItems < 1 {
		return 1, errors.New("not enough space for any items")
	}

	return visibleItems, err
}
