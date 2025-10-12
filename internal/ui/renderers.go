package ui

import (
	"github.com/jarmocluyse/wip-tui/internal/config"
	"github.com/jarmocluyse/wip-tui/internal/git"
	"github.com/jarmocluyse/wip-tui/internal/repository"
	"github.com/jarmocluyse/wip-tui/internal/theme"
	"github.com/jarmocluyse/wip-tui/internal/ui/pages/explore"
	"github.com/jarmocluyse/wip-tui/internal/ui/pages/home"
	repomanagement "github.com/jarmocluyse/wip-tui/internal/ui/pages/repo-management"
	"github.com/jarmocluyse/wip-tui/internal/ui/types"
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
	}
	return &ListViewRenderer{
		homeRenderer: home.NewRenderer(homeStyles, themeConfig),
	}
}

// Render renders the repository list with the given cursor position and dimensions.
func (r *ListViewRenderer) Render(repositories []repository.Repository, cursor int, width, height int, actions []config.Action) string {
	return r.homeRenderer.RenderRepositoryList(repositories, cursor, width, height, actions)
}

// RenderNavigable renders the navigable items list with the given cursor position and dimensions.
func (r *ListViewRenderer) RenderNavigable(items []types.NavigableItem, cursor int, width, height int, gitChecker git.StatusChecker, actions []config.Action) string {
	return r.homeRenderer.RenderNavigableList(items, cursor, width, height, gitChecker, actions)
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
