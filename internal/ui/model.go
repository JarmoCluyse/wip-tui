package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/wip-tui/internal/config"
	"github.com/jarmocluyse/wip-tui/internal/explorer"
	"github.com/jarmocluyse/wip-tui/internal/git"
	"github.com/jarmocluyse/wip-tui/internal/repository"
	"github.com/jarmocluyse/wip-tui/internal/theme"
	"github.com/jarmocluyse/wip-tui/internal/ui/types"
)

type ViewState int

const (
	ListView ViewState = iota
	RepoManagementView
	ExplorerView
	DetailsView
	ActionConfigView
)

// Dependencies interface defines what the UI needs from the application layer
type Dependencies interface {
	GetConfigService() config.ConfigService
	GetStatusUpdater() *repository.StatusUpdater
	GetGitChecker() git.StatusChecker
}

type Model struct {
	Dependencies     Dependencies
	Config           *config.Config
	RepoHandler      *repository.Handler
	State            ViewState
	PreviousState    ViewState // Track the previous state for navigation
	Cursor           int
	ScrollOffset     int // New field for scrolling
	InputField       string
	InputPrompt      string
	ExplorerPath     string
	ExplorerItems    []explorer.Item
	ExplorerCursor   int
	ShowHelpModal    bool
	Width            int
	Height           int
	Err              error
	CachedNavItems   []types.NavigableItem // Cache for navigable items
	NavItemsNeedSync bool                  // Flag to indicate cache needs update
	SelectedNavItem  *types.NavigableItem  // Currently selected item for details view

	// Action configuration fields
	ActionConfigCursor   int            // Cursor for action list
	ActionConfigEditMode bool           // Whether we're editing an action
	ActionConfigFieldIdx int            // Current field being edited
	ActionConfigAction   *config.Action // Action being edited
	ActionConfigIsNew    bool           // Whether we're creating a new action

	// Handler instances for separated concerns
	KeyHandler        *KeyHandler
	NavigationHandler *NavigationHandler
	RepositoryHandler *RepositoryOperationHandler
	ExplorerHandler   *ExplorerHandler
}

type StatusMessage struct {
	Repositories []repository.Repository
}

type StyleConfig struct {
	Item              lipgloss.Style
	SelectedItem      lipgloss.Style
	StatusUncommitted lipgloss.Style
	StatusUnpushed    lipgloss.Style
	StatusUntracked   lipgloss.Style
	StatusError       lipgloss.Style
	StatusClean       lipgloss.Style
	StatusNotAdded    lipgloss.Style
	Input             lipgloss.Style
	Help              lipgloss.Style
	HelpModal         lipgloss.Style
	HelpModalTitle    lipgloss.Style
	HelpModalContent  lipgloss.Style
	HelpModalFooter   lipgloss.Style
	Branch            lipgloss.Style
	Border            lipgloss.Style
	IconRegular       lipgloss.Style
	IconBare          lipgloss.Style
	IconWorktree      lipgloss.Style
}

// CreateStyleConfig creates a new StyleConfig using the provided theme configuration.
func CreateStyleConfig(themeConfig theme.Theme) StyleConfig {
	return StyleConfig{
		Item: lipgloss.NewStyle(),
		SelectedItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.Selected)).
			Bold(true),
		StatusUncommitted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.StatusDirty)).
			Bold(true),
		StatusUnpushed: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.StatusUnpushed)).
			Bold(true),
		StatusUntracked: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.StatusUntracked)).
			Bold(true),
		StatusError: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.StatusError)).
			Bold(true),
		StatusClean: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.StatusClean)).
			Bold(true),
		StatusNotAdded: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.StatusNotAdded)),
		Input: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1).
			Width(50),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.Help)).
			Margin(1, 0),
		HelpModal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(themeConfig.Colors.Border)).
			Background(lipgloss.Color(themeConfig.Colors.ModalBackground)).
			Padding(1).
			Width(80).
			Height(30),
		HelpModalTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(themeConfig.Colors.Title)).
			Background(lipgloss.Color(themeConfig.Colors.TitleBackground)).
			Padding(0, 1).
			Width(76).
			Align(lipgloss.Center),
		HelpModalContent: lipgloss.NewStyle().
			Padding(1, 2).
			Width(76),
		HelpModalFooter: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.Help)).
			Padding(0, 2).
			Width(76).
			Align(lipgloss.Center),
		Branch: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.Branch)).
			Bold(true),
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(themeConfig.Colors.Border)).
			Padding(0, 1).
			Margin(0, 0, 0, 0),
		IconRegular: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.IconRegular)).
			Bold(true),
		IconBare: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.IconBare)).
			Bold(true),
		IconWorktree: lipgloss.NewStyle().
			Foreground(lipgloss.Color(themeConfig.Colors.IconWorktree)).
			Bold(true),
	}
}
