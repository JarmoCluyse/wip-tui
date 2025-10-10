package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/wip-tui/internal/config"
	"github.com/jarmocluyse/wip-tui/internal/explorer"
	"github.com/jarmocluyse/wip-tui/internal/git"
	"github.com/jarmocluyse/wip-tui/internal/repository"
	"github.com/jarmocluyse/wip-tui/internal/theme"
)

type ViewState int

const (
	ListView ViewState = iota
	AddRepoView
	ExplorerView
)

// Dependencies interface defines what the UI needs from the application layer
type Dependencies interface {
	GetConfigService() config.ConfigService
	GetStatusUpdater() *repository.StatusUpdater
	GetGitChecker() git.StatusChecker
}

type Model struct {
	Dependencies   Dependencies
	Config         *config.Config
	RepoHandler    *repository.Handler
	State          ViewState
	Cursor         int
	ScrollOffset   int // New field for scrolling
	InputField     string
	InputPrompt    string
	ExplorerPath   string
	ExplorerItems  []explorer.Item
	ExplorerCursor int
	ShowHelpModal  bool
	Width          int
	Height         int
	Err            error
}

type NavigableItem struct {
	Type         string // "repository" or "worktree"
	Repository   *repository.Repository
	WorktreeInfo *git.WorktreeInfo
	ParentRepo   *repository.Repository // For worktrees, reference to parent bare repo
	IsLast       bool                   // For worktrees, indicates if this is the last worktree for the parent repo
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
}

func CreateStyleConfig(themeConfig theme.Theme) StyleConfig {
	return StyleConfig{
		Item: lipgloss.NewStyle().
			PaddingLeft(2),
		SelectedItem: lipgloss.NewStyle().
			PaddingLeft(2).
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
	}
}
