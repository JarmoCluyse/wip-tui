package ui

import (
	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/git"
	"github.com/jarmocluyse/git-dash/internal/repository"
	"github.com/jarmocluyse/git-dash/internal/theme"
	actionconfig "github.com/jarmocluyse/git-dash/ui/pages/action-config"
	"github.com/jarmocluyse/git-dash/ui/pages/explore"
	"github.com/jarmocluyse/git-dash/ui/pages/home"
	repomanagement "github.com/jarmocluyse/git-dash/ui/pages/repo-management"
	"github.com/jarmocluyse/git-dash/ui/types"
)

type ExplorerItem = explore.Item

type ListViewRenderer struct {
	homeRenderer *home.Renderer
}

// NewListViewRenderer creates a new list view renderer with the given styles and theme.
func NewListViewRenderer(styles StyleConfig, themeConfig theme.Theme) *ListViewRenderer {
	homeStyles := home.StyleConfig{
		Item:              styles.Item,
		SelectedItem:      styles.SelectedItem,
		StatusUncommitted: styles.StatusUncommitted,
		StatusUnpushed:    styles.StatusUnpushed,
		StatusUntracked:   styles.StatusUntracked,
		StatusError:       styles.StatusError,
		StatusClean:       styles.StatusClean,
		StatusNotAdded:    styles.StatusNotAdded,
		Help:              styles.Help,
		Branch:            styles.Branch,
		Border:            styles.Border,
		IconRegular:       styles.IconRegular,
		IconBare:          styles.IconBare,
		IconWorktree:      styles.IconWorktree,
	}
	return &ListViewRenderer{
		homeRenderer: home.NewRenderer(homeStyles, themeConfig),
	}
}

// Render renders the repository list with the given cursor position and dimensions.
func (r *ListViewRenderer) Render(repositories []repository.Repository, summaryData repository.SummaryData, cursor int, width, height int, actions []config.Action, configTitle string) string {
	return r.homeRenderer.RenderRepositoryList(repositories, summaryData, cursor, width, height, actions, configTitle)
}

// RenderNavigable renders the navigable items list with the given cursor position and dimensions.
func (r *ListViewRenderer) RenderNavigable(items []types.NavigableItem, summaryData repository.SummaryData, cursor int, width, height int, gitChecker git.StatusChecker, actions []config.Action, configTitle string) string {
	return r.homeRenderer.RenderNavigableList(items, summaryData, cursor, width, height, gitChecker, actions, configTitle)
}

type ExplorerViewRenderer struct {
	exploreRenderer *explore.Renderer
}

// NewExplorerViewRenderer creates a new explorer view renderer with the given styles and theme.
func NewExplorerViewRenderer(styles StyleConfig, themeConfig theme.Theme) *ExplorerViewRenderer {
	exploreStyles := explore.StyleConfig{
		Item:              styles.Item,
		SelectedItem:      styles.SelectedItem,
		StatusUncommitted: styles.StatusUncommitted,
		StatusUnpushed:    styles.StatusUnpushed,
		StatusUntracked:   styles.StatusUntracked,
		StatusError:       styles.StatusError,
		StatusClean:       styles.StatusClean,
		StatusNotAdded:    styles.StatusNotAdded,
		Help:              styles.Help,
		Branch:            styles.Branch,
		IconRegular:       styles.IconRegular,
		IconBare:          styles.IconBare,
		IconWorktree:      styles.IconWorktree,
	}
	return &ExplorerViewRenderer{
		exploreRenderer: explore.NewRenderer(exploreStyles, themeConfig),
	}
}

// Render renders the directory explorer with the current path, items, and cursor position.
func (r *ExplorerViewRenderer) Render(currentPath string, items []ExplorerItem, cursor int, width, height int) string {
	return r.exploreRenderer.Render(currentPath, items, cursor, width, height)
}

type RepoManagementViewRenderer struct {
	repoManagementRenderer *repomanagement.Renderer
}

// NewRepoManagementViewRenderer creates a new repository management view renderer with the given styles and theme.
func NewRepoManagementViewRenderer(styles StyleConfig, themeConfig theme.Theme) *RepoManagementViewRenderer {
	repoManagementStyles := repomanagement.StyleConfig{
		Item:         styles.Item,
		SelectedItem: styles.SelectedItem,
		EmptyState:   styles.Help,   // Reuse help style for empty state
		SectionTitle: styles.Border, // Reuse border style for section title
		Help:         styles.Help,
	}
	return &RepoManagementViewRenderer{
		repoManagementRenderer: repomanagement.NewRenderer(repoManagementStyles, themeConfig),
	}
}

// Render renders the repository management view with the given repositories and cursor position.
func (r *RepoManagementViewRenderer) Render(repositories []repository.Repository, cursor int, width, height int) string {
	return r.repoManagementRenderer.Render(repositories, cursor, width, height)
}

type ActionConfigViewRenderer struct {
	actionConfigRenderer *actionconfig.Renderer
}

// NewActionConfigViewRenderer creates a new action config view renderer with the given styles and theme.
func NewActionConfigViewRenderer(styles StyleConfig, themeConfig theme.Theme) *ActionConfigViewRenderer {
	actionConfigStyles := actionconfig.StyleConfig{
		Item:          styles.Item,
		SelectedItem:  styles.SelectedItem,
		SectionTitle:  styles.Border, // Reuse border style for section title
		Help:          styles.Help,
		Border:        styles.Border,
		EmptyState:    styles.Help, // Reuse help style for empty state
		Input:         styles.Item,
		InputPrompt:   styles.Item.Bold(true),
		ActionKey:     styles.Item.Bold(true),
		ActionCommand: styles.Help, // Dimmed style for command
		ActionDesc:    styles.Help, // Dimmed style for description
	}
	return &ActionConfigViewRenderer{
		actionConfigRenderer: actionconfig.NewRenderer(actionConfigStyles, themeConfig),
	}
}

// Render renders the action configuration view.
func (r *ActionConfigViewRenderer) Render(actions []config.Action, cursor int, editMode bool, editingAction *config.Action, fieldIdx int, width, height int) string {
	if editMode && editingAction != nil {
		return r.actionConfigRenderer.RenderActionEditor(editingAction, fieldIdx, width, height, false)
	}
	return r.actionConfigRenderer.Render(actions, cursor, width, height)
}
