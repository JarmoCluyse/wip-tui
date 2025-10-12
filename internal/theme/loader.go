package theme

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Loader handles theme loading with a proper loading order.
type Loader struct{}

// NewLoader creates a new theme loader.
func NewLoader() *Loader {
	return &Loader{}
}

// LoadTheme loads theme configuration with the following priority order:
// 1. User config file (highest priority)
// 2. System-wide config
// 3. Built-in defaults (lowest priority)
func (l *Loader) LoadTheme() (Theme, error) {
	// Start with default theme
	theme := Default()

	// Try to load system-wide theme (if it exists)
	if systemTheme, err := l.loadSystemTheme(); err == nil {
		theme = MergeWithDefault(systemTheme)
	}

	// Try to load user theme (highest priority)
	if userTheme, err := l.loadUserTheme(); err == nil {
		theme = MergeWithDefault(userTheme)
	}

	return theme, nil
}

// loadUserTheme loads theme from user config file.
func (l *Loader) loadUserTheme() (Theme, error) {
	configPath, err := l.getUserConfigPath()
	if err != nil {
		return Theme{}, err
	}

	if !l.fileExists(configPath) {
		return Theme{}, os.ErrNotExist
	}

	return l.loadThemeFromFile(configPath)
}

// loadSystemTheme loads theme from system-wide config.
func (l *Loader) loadSystemTheme() (Theme, error) {
	systemPaths := []string{
		"/etc/git-tui/theme.toml",
		"/usr/local/etc/git-tui/theme.toml",
	}

	for _, path := range systemPaths {
		if l.fileExists(path) {
			return l.loadThemeFromFile(path)
		}
	}

	return Theme{}, os.ErrNotExist
}

// loadThemeFromFile loads theme configuration from a specific file.
func (l *Loader) loadThemeFromFile(path string) (Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Theme{}, err
	}

	var config struct {
		Theme Theme `toml:"theme"`
	}

	_, err = toml.Decode(string(data), &config)
	if err != nil {
		return Theme{}, err
	}

	return config.Theme, nil
}

// getUserConfigPath returns the path to user config file.
func (l *Loader) getUserConfigPath() (string, error) {
	// Check for environment variable override first
	if configPath := os.Getenv("GIT_TUI_CONFIG"); configPath != "" {
		return configPath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".git-tui.toml"), nil
}

// fileExists checks if a file exists.
func (l *Loader) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
