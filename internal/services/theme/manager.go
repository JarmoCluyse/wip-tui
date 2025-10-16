// Package theme provides theme management services for the application.
package theme

import (
	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/theme"
)

// Manager implements the theme service interface.
type Manager struct {
	currentTheme  *theme.Theme
	configService config.ConfigService
}

// NewManager creates a new theme manager instance.
func NewManager(configService config.ConfigService) *Manager {
	return &Manager{
		configService: configService,
	}
}

// LoadTheme loads the theme from configuration, merging with defaults.
func (m *Manager) LoadTheme() (*theme.Theme, error) {
	// Start with default theme
	currentTheme := theme.Default()

	// Load configuration if available
	if m.configService != nil {
		config, err := m.configService.Load()
		if err != nil {
			// If config can't be loaded, just use defaults
			m.currentTheme = &currentTheme
			return &currentTheme, nil
		}

		// Merge user theme with defaults
		// Since Theme is a value type, we can directly merge it
		currentTheme = theme.MergeWithDefault(config.Theme)
	}

	m.currentTheme = &currentTheme
	return &currentTheme, nil
}

// GetTheme returns the current theme, loading it if necessary.
func (m *Manager) GetTheme() *theme.Theme {
	if m.currentTheme == nil {
		// Load theme if not already loaded
		if loadedTheme, err := m.LoadTheme(); err == nil {
			return loadedTheme
		}
		// Fallback to default if loading fails
		defaultTheme := theme.Default()
		m.currentTheme = &defaultTheme
	}
	return m.currentTheme
}

// RefreshTheme reloads the theme from configuration.
func (m *Manager) RefreshTheme() error {
	_, err := m.LoadTheme()
	return err
}

// GetDefaultTheme returns the default theme configuration.
func (m *Manager) GetDefaultTheme() theme.Theme {
	return theme.Default()
}

// MergeWithDefault merges a user theme with the default theme.
func (m *Manager) MergeWithDefault(userTheme theme.Theme) theme.Theme {
	return theme.MergeWithDefault(userTheme)
}
