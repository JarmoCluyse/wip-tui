package theme

import "github.com/jarmocluyse/git-dash/internal/theme/types"

// LoadTheme loads the theme from configuration, merging with defaults.
func (m *Manager) LoadTheme() (*types.Theme, error) {
	// Start with default theme
	currentTheme := Default()

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
		currentTheme = MergeWithDefault(config.Theme)
	}

	m.currentTheme = &currentTheme
	return &currentTheme, nil
}

// GetTheme returns the current theme, loading it if necessary.
func (m *Manager) GetTheme() *types.Theme {
	if m.currentTheme == nil {
		// Load theme if not already loaded
		if loadedTheme, err := m.LoadTheme(); err == nil {
			return loadedTheme
		}
		// Fallback to default if loading fails
		defaultTheme := Default()
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
func (m *Manager) GetDefaultTheme() types.Theme {
	return Default()
}

// MergeWithDefault merges a user theme with the default theme.
func (m *Manager) MergeWithDefault(userTheme types.Theme) types.Theme {
	return MergeWithDefault(userTheme)
}
