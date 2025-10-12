package home

import (
	"strings"

	"github.com/jarmocluyse/wip-tui/internal/config"
	"github.com/jarmocluyse/wip-tui/internal/git"
	"github.com/jarmocluyse/wip-tui/internal/repository"
	"github.com/jarmocluyse/wip-tui/internal/theme"
	"github.com/jarmocluyse/wip-tui/internal/ui/components/help"
	"github.com/jarmocluyse/wip-tui/internal/ui/header"
	"github.com/jarmocluyse/wip-tui/internal/ui/repo"
	"github.com/jarmocluyse/wip-tui/internal/ui/types"
)

// Renderer handles rendering of the home page (repository list)
type Renderer struct {
	styles StyleConfig
	theme  theme.Theme
	header *header.Renderer
	repo   *repo.Renderer
}

// NewRenderer creates a new home page renderer
func NewRenderer(styles StyleConfig, themeConfig theme.Theme) *Renderer {
	repoStyles := repo.StyleConfig{
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
	return &Renderer{
		styles: styles,
		theme:  themeConfig,
		header: header.NewRenderer(themeConfig),
		repo:   repo.NewRenderer(repoStyles, themeConfig),
	}
}

// RenderRepositoryList renders the main repository list view
func (r *Renderer) RenderRepositoryList(repositories []repository.Repository, cursor int, width, height int, actions []config.Action) string {
	content := r.header.RenderWithSpacing("Git Repository Status", width)

	if len(repositories) == 0 {
		content += r.renderEmptyState()
	} else {
		content += r.renderRepositoryList(repositories, cursor, width)
	}

	// Use help component to render with bottom-aligned help
	helpBuilder := help.NewBuilder(r.styles.Help)

	// Build help bindings from actions
	var bindings []help.KeyBinding
	for _, action := range actions {
		bindings = append(bindings, help.KeyBinding{
			Key:         action.Key,
			Description: action.Description,
		})
	}
	bindings = append(bindings, help.KeyBinding{Key: "m", Description: "manage repos"})

	return helpBuilder.RenderWithBottomHelpAndHeader(content, bindings, width, height, 3)
}

// RenderNavigableList renders the navigable repository list (with worktrees as separate items)
func (r *Renderer) RenderNavigableList(items []types.NavigableItem, cursor int, width, height int, gitChecker git.StatusChecker, actions []config.Action) string {
	content := r.header.RenderWithSpacing("Git Repository Status", width)

	if len(items) == 0 {
		content += r.renderEmptyState()
	} else {
		content += r.renderNavigableItemList(items, cursor, width, gitChecker)
	}

	// Use help component to render with bottom-aligned help
	helpBuilder := help.NewBuilder(r.styles.Help)

	// Build help bindings from actions
	var bindings []help.KeyBinding
	for _, action := range actions {
		bindings = append(bindings, help.KeyBinding{
			Key:         action.Key,
			Description: action.Description,
		})
	}
	bindings = append(bindings, help.KeyBinding{Key: "m", Description: "manage repos"})

	return helpBuilder.RenderWithBottomHelpAndHeader(content, bindings, width, height, 3)
}

// renderNavigableItemList renders a list of navigable items (repositories and worktrees).
func (r *Renderer) renderNavigableItemList(items []types.NavigableItem, cursor int, width int, gitChecker git.StatusChecker) string {
	var content string
	i := 0

	for i < len(items) {
		item := items[i]

		if item.Type == "repository" && item.Repository.IsBare {
			// Start of bare repository group - collect all items in this group
			groupContent := r.renderNavigableItem(item, i, cursor, width, gitChecker)

			// Add all worktrees that belong to this bare repository
			j := i + 1
			for j < len(items) && items[j].Type == "worktree" && items[j].ParentRepo.Path == item.Repository.Path {
				groupContent += "\n" + r.renderNavigableItem(items[j], j, cursor, width, gitChecker)
				j++
			}

			// No border - just add the group content directly
			content += groupContent + "\n"

			// Move index to after the group
			i = j
		} else {
			// Regular item (non-bare repository or standalone worktree)
			content += r.renderNavigableItem(item, i, cursor, width, gitChecker) + "\n"
			i++
		}
	}

	return content
}

// renderNavigableItem renders a single navigable item.
func (r *Renderer) renderNavigableItem(item types.NavigableItem, index, cursor int, width int, gitChecker git.StatusChecker) string {
	isSelected := index == cursor

	switch item.Type {
	case "repository":
		return r.repo.RenderRepositoryOnly(*item.Repository, isSelected, "", width)
	case "worktree":
		wt := item.WorktreeInfo
		parentName := item.ParentRepo.Name
		return r.repo.RenderWorktree(*wt, parentName, item.ParentRepo.Path, isSelected, "", item.IsLast, width, gitChecker)
	default:
		return ""
	}
}

// renderEmptyState renders a message when no repositories are configured.
func (r *Renderer) renderEmptyState() string {
	return r.styles.Item.Render("No repositories configured.") + "\n\n"
}

// renderRepositoryList renders a list of repositories with cursor indication.
func (r *Renderer) renderRepositoryList(repositories []repository.Repository, cursor int, width int) string {
	var content string
	for i, repo := range repositories {
		content += r.renderRepositoryItem(repo, i, cursor, width)
	}
	return content
}

// renderRepositoryItem renders a single repository item.
func (r *Renderer) renderRepositoryItem(repo repository.Repository, index, cursor int, width int) string {
	isSelected := index == cursor
	return r.repo.RenderRepository(repo, isSelected, "", width)
}

// renderHelp renders the help section with available actions.
func (r *Renderer) renderHelp(actions []config.Action) string {
	helpBuilder := help.NewBuilder(r.styles.Help)

	// Build bindings list
	var bindings []help.KeyBinding

	// Static keybindings - removed "e" and "d" as they're now only in management page
	bindings = append(bindings,
		help.KeyBinding{Key: "m", Description: "manage repos"},
		help.KeyBinding{Key: "r", Description: "refresh"},
	)

	// Dynamic action keybindings
	for _, action := range actions {
		bindings = append(bindings, help.KeyBinding{
			Key:         action.Key,
			Description: strings.ToLower(action.Name),
		})
	}

	return helpBuilder.BuildCompactHelp(bindings)
}
