package home

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/git"
	"github.com/jarmocluyse/git-dash/internal/repository"
	"github.com/jarmocluyse/git-dash/internal/theme"
	"github.com/jarmocluyse/git-dash/ui/components/help"
	"github.com/jarmocluyse/git-dash/ui/header"
	"github.com/jarmocluyse/git-dash/ui/repo"
	"github.com/jarmocluyse/git-dash/ui/types"
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
func (r *Renderer) RenderRepositoryList(repositories []repository.Repository, summaryData repository.SummaryData, cursor int, width, height int, actions []config.Action, configTitle string) string {
	content := r.header.RenderWithCountAndSpacing("git-dash", configTitle, len(repositories), width)

	// Add summary header for repository list
	content += r.renderSummary(summaryData, width)

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

	return helpBuilder.RenderWithBottomHelpAndHeader(content, bindings, width, height, 4) // Increased header count
}

// RenderNavigableList renders the navigable repository list (with worktrees as separate items)
func (r *Renderer) RenderNavigableList(items []types.NavigableItem, summaryData repository.SummaryData, cursor int, width, height int, gitChecker git.StatusChecker, actions []config.Action, configTitle string) string {
	content := r.header.RenderWithCountAndSpacing("git-dash", configTitle, len(items), width)

	// Add summary header
	content += r.renderSummaryHeader(summaryData, width)

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

	return helpBuilder.RenderWithBottomHelpAndHeader(content, bindings, width, height, 4) // Increased header count
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

// renderSummary renders a summary line showing repo/branch info and aggregated change counts
func (r *Renderer) renderSummary(summaryData repository.SummaryData, width int) string {
	// Build the summary line as a static table header with icons and padding
	frontIndicatorWidth := lipgloss.Width(r.theme.Indicators.Selected)
	frontPadding := strings.Repeat(" ", frontIndicatorWidth)
	leftPart := fmt.Sprintf("%s%s repo | %s branch", frontPadding, r.theme.Icons.Repository.Regular, r.theme.Icons.Branch.Icon)

	// Build summary counts
	var summaryParts []string
	if summaryData.TotalUncommitted > 0 {
		summaryParts = append(summaryParts, r.styles.StatusUncommitted.Render(fmt.Sprintf("%d uncommitted", summaryData.TotalUncommitted)))
	}
	if summaryData.TotalUnpushed > 0 {
		summaryParts = append(summaryParts, r.styles.StatusUnpushed.Render(fmt.Sprintf("%d unpushed", summaryData.TotalUnpushed)))
	}
	if summaryData.TotalUntracked > 0 {
		summaryParts = append(summaryParts, r.styles.StatusUntracked.Render(fmt.Sprintf("%d untracked", summaryData.TotalUntracked)))
	}
	if summaryData.TotalErrors > 0 {
		summaryParts = append(summaryParts, r.styles.StatusError.Render(fmt.Sprintf("%d errors", summaryData.TotalErrors)))
	}

	var rightPart string
	if len(summaryParts) > 0 {
		rightPart = strings.Join(summaryParts, " ")
	} else {
		rightPart = r.styles.StatusClean.Render("all clean")
	}

	// Calculate spacing to right-align the summary
	leftWidth := lipgloss.Width(leftPart)
	rightWidth := lipgloss.Width(rightPart)
	availableWidth := width - 2 // margin for safety
	padding := availableWidth - leftWidth - rightWidth

	if padding < 1 {
		padding = 1
	}

	summaryLine := leftPart + strings.Repeat(" ", padding) + rightPart
	return r.styles.Help.Render(summaryLine) + "\n\n"
}

// renderRepositorySummaryHeader renders a summary line for the repository list view
func (r *Renderer) renderRepositorySummaryHeader(summaryData repository.SummaryData, width int) string {
	return r.renderSummary(summaryData, width)
}

// renderSummaryHeader renders a summary line showing current repo/branch and total change counts
func (r *Renderer) renderSummaryHeader(summaryData repository.SummaryData, width int) string {
	return r.renderSummary(summaryData, width)
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
