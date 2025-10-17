package theme

import (
	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/theme/types"
)

// Service defines the interface for theme management operations.
type Service interface {
	// Theme Loading and Access
	LoadTheme() (*types.Theme, error)
	GetTheme() *types.Theme
	RefreshTheme() error

	// Theme Configuration
	GetDefaultTheme() types.Theme
	MergeWithDefault(userTheme types.Theme) types.Theme
}

// Manager implements the theme service interface.
type Manager struct {
	currentTheme  *types.Theme
	configService config.ConfigService
}

// NewManager creates a new theme manager instance.
func NewManager(configService config.ConfigService) *Manager {
	return &Manager{
		configService: configService,
	}
}
