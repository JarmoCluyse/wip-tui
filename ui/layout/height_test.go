package layout

import (
	"strings"
	"testing"
)

func TestHeightCalculator_ValidateTerminalHeight(t *testing.T) {
	calc := NewHeightCalculator()

	tests := []struct {
		name           string
		input          int
		expectedHeight int
		expectError    bool
	}{
		{"valid height", 50, 50, false},
		{"zero height", 0, 40, true},
		{"negative height", -5, 40, true},
		{"too small height", 5, 10, true},
		{"minimum height", 10, 10, false},
		{"large height", 200, 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calc.ValidateTerminalHeight(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expectedHeight {
				t.Errorf("expected height %d, got %d", tt.expectedHeight, result)
			}
		})
	}
}

func TestHeightCalculator_CalculateContentAreaHeight(t *testing.T) {
	calc := NewHeightCalculator()

	tests := []struct {
		name          string
		totalHeight   int
		reservedLines int
		expected      int
		expectError   bool
	}{
		{"normal case", 30, 5, 25, false},
		{"minimal space", 10, 8, 2, false},
		{"too small result", 5, 6, 4, false},   // Uses minimum height (10) - 6 = 4
		{"zero total height", 0, 5, 35, false}, // Uses default height (40) - 5 = 35
		{"negative reserved", 20, -2, 22, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calc.CalculateContentAreaHeight(tt.totalHeight, tt.reservedLines)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestHeightCalculator_CountContentLines(t *testing.T) {
	calc := NewHeightCalculator()

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"empty content", "", 0},
		{"single line", "hello", 1},
		{"multiple lines", "line1\nline2\nline3", 3},
		{"trailing newline", "line1\nline2\n", 3},
		{"multiple newlines", "line1\n\n\nline2", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CountContentLines(tt.content)
			if result != tt.expected {
				t.Errorf("expected %d lines, got %d", tt.expected, result)
			}
		})
	}
}

func TestHeightCalculator_CalculatePaddingLines(t *testing.T) {
	calc := NewHeightCalculator()

	tests := []struct {
		name            string
		contentLines    int
		availableHeight int
		expected        int
	}{
		{"normal padding", 5, 10, 5},
		{"no padding needed", 10, 10, 0},
		{"content too large", 15, 10, 0},
		{"zero content", 0, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculatePaddingLines(tt.contentLines, tt.availableHeight)
			if result != tt.expected {
				t.Errorf("expected %d padding lines, got %d", tt.expected, result)
			}
		})
	}
}

func TestHeightCalculator_CalculateVisibleItemCount(t *testing.T) {
	calc := NewHeightCalculator()

	tests := []struct {
		name        string
		totalHeight int
		headerLines int
		helpLines   int
		itemHeight  int
		expected    int
		expectError bool
	}{
		{"normal case", 50, 3, 1, 2, 23, false},
		{"single line items", 30, 2, 1, 1, 27, false},
		{"large items", 20, 2, 1, 5, 3, false},
		{"too small terminal", 5, 2, 1, 1, 7, false}, // Uses minimum height (10) - 3 = 7
		{"zero item height", 30, 2, 1, 0, 27, false}, // Uses default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calc.CalculateVisibleItemCount(tt.totalHeight, tt.headerLines, tt.helpLines, tt.itemHeight)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if result != tt.expected {
				t.Errorf("expected %d visible items, got %d", tt.expected, result)
			}
		})
	}
}

func TestHeightCalculator_CalculateHelpAreaHeight(t *testing.T) {
	calc := NewHeightCalculator()

	result := calc.CalculateHelpAreaHeight()
	expected := 1

	if result != expected {
		t.Errorf("expected help area height %d, got %d", expected, result)
	}
}

// Benchmark tests for performance
func BenchmarkCalculateContentAreaHeight(b *testing.B) {
	calc := NewHeightCalculator()
	for i := 0; i < b.N; i++ {
		calc.CalculateContentAreaHeight(50, 5)
	}
}

func BenchmarkCountContentLines(b *testing.B) {
	calc := NewHeightCalculator()
	content := strings.Repeat("line\n", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calc.CountContentLines(content)
	}
}

func BenchmarkCalculatePaddingLines(b *testing.B) {
	calc := NewHeightCalculator()
	for i := 0; i < b.N; i++ {
		calc.CalculatePaddingLines(10, 20)
	}
}
