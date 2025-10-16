// Package theme provides theme management services for the application.
package theme

import (
	"github.com/jarmocluyse/git-dash/internal/theme"
)

// Service defines the interface for theme management operations.
type Service interface {
	// Theme Loading and Access
	LoadTheme() (*theme.Theme, error)
	GetTheme() *theme.Theme
	RefreshTheme() error

	// Theme Configuration
	GetDefaultTheme() theme.Theme
	MergeWithDefault(userTheme theme.Theme) theme.Theme
}
