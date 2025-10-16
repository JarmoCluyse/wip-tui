package ui

import (
	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/repomanager"
	"github.com/jarmocluyse/git-dash/internal/theme"
	actionconfig "github.com/jarmocluyse/git-dash/ui/pages/action-config"
	"github.com/jarmocluyse/git-dash/ui/pages/home"
	settings "github.com/jarmocluyse/git-dash/ui/pages/settings"
	"github.com/jarmocluyse/git-dash/ui/types"
)

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
func (r *ListViewRenderer) Render(repositories []*repomanager.RepoItem, summaryData *repomanager.SummaryData, cursor int, width, height int, actions []config.Action, configTitle string) string {
	return r.homeRenderer.RenderRepositoryList(repositories, *summaryData, cursor, width, height, actions, configTitle)
}

// RenderNavigable renders the navigable items list with the given cursor position and dimensions.
func (r *ListViewRenderer) RenderNavigable(items []types.NavigableItem, summaryData *repomanager.SummaryData, cursor int, width, height int, actions []config.Action, configTitle string) string {
	return r.homeRenderer.RenderNavigableList(items, *summaryData, cursor, width, height, actions, configTitle)
}

// ActionConfigRenderer renders the action configuration view.
type ActionConfigRenderer struct {
	actionConfigRenderer *actionconfig.Renderer
}

// NewActionConfigRenderer creates a new action config renderer.
func NewActionConfigRenderer(styles StyleConfig, themeConfig theme.Theme) *ActionConfigRenderer {
	actionConfigStyles := actionconfig.StyleConfig{
		Item:         styles.Item,
		SelectedItem: styles.SelectedItem,
		Help:         styles.Help,
	}
	return &ActionConfigRenderer{
		actionConfigRenderer: actionconfig.NewRenderer(actionConfigStyles, themeConfig),
	}
}

// Render renders the action configuration view.
func (r *ActionConfigRenderer) Render(actions []config.Action, cursor int, width, height int, title string) string {
	return r.actionConfigRenderer.Render(actions, cursor, width, height)
}

// SettingsRenderer renders the settings view.
type SettingsRenderer struct {
	settingsRenderer *settings.Renderer
}

// NewSettingsRenderer creates a new settings renderer.
func NewSettingsRenderer(styles StyleConfig, themeConfig theme.Theme) *SettingsRenderer {
	settingsStyles := settings.StyleConfig{
		Item:         styles.Item,
		SelectedItem: styles.SelectedItem,
		Help:         styles.Help,
		EmptyState:   styles.StatusNotAdded,
		SectionTitle: styles.Item,
	}
	return &SettingsRenderer{
		settingsRenderer: settings.NewRenderer(settingsStyles, themeConfig),
	}
}

// Render renders the settings view.
func (r *SettingsRenderer) Render(data settings.SettingsData, currentSection settings.SettingsSection, cursor int, width, height int, themeEditMode bool, themeEditValue string, actionEditMode bool, actionEditValue string, actionEditFieldType string, actionEditItemIndex int) string {
	return r.settingsRenderer.Render(data, currentSection, cursor, width, height, themeEditMode, themeEditValue, actionEditMode, actionEditValue, actionEditFieldType, actionEditItemIndex)
}
