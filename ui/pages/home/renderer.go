package home

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/git-dash/internal/config"
	"github.com/jarmocluyse/git-dash/internal/repomanager"
	theme "github.com/jarmocluyse/git-dash/internal/theme/types"
	"github.com/jarmocluyse/git-dash/ui/components/help"
	"github.com/jarmocluyse/git-dash/ui/header"
	"github.com/jarmocluyse/git-dash/ui/types"
)

// Renderer handles rendering of the home page (repository list)
type Renderer struct {
	styles StyleConfig
	theme  theme.Theme
	header *header.Renderer
}

// NewRenderer creates a new home page renderer
func NewRenderer(styles StyleConfig, themeConfig theme.Theme) *Renderer {
	return &Renderer{
		styles: styles,
		theme:  themeConfig,
		header: header.NewRenderer(themeConfig),
	}
}

// RenderRepositoryList renders the main repository list view
func (r *Renderer) RenderRepositoryList(repositories []*repomanager.RepoItem, summaryData repomanager.SummaryData, cursor int, width, height int, actions []config.Action, configTitle string) string {
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
	bindings = append(bindings, help.KeyBinding{Key: "e", Description: "open in file manager"})
	bindings = append(bindings, help.KeyBinding{Key: "s", Description: "settings"})

	return helpBuilder.RenderWithBottomHelpAndHeader(content, bindings, width, height, 4) // Increased header count
}

// RenderNavigableList renders the navigable repository list (with worktrees as separate items)
func (r *Renderer) RenderNavigableList(items []types.NavigableItem, summaryData repomanager.SummaryData, cursor int, width, height int, actions []config.Action, configTitle string) string {
	content := r.header.RenderWithCountAndSpacing("git-dash", configTitle, len(items), width)

	// Add summary header
	content += r.renderSummaryHeader(summaryData, width)

	if len(items) == 0 {
		content += r.renderEmptyState()
	} else {
		content += r.renderNavigableItemList(items, cursor, width)
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
	bindings = append(bindings, help.KeyBinding{Key: "e", Description: "open in file manager"})
	bindings = append(bindings, help.KeyBinding{Key: "s", Description: "settings"})

	return helpBuilder.RenderWithBottomHelpAndHeader(content, bindings, width, height, 4) // Increased header count
}

// renderNavigableItemList renders a list of navigable items (repositories and worktrees).
func (r *Renderer) renderNavigableItemList(items []types.NavigableItem, cursor int, width int) string {
	var content string
	i := 0

	for i < len(items) {
		item := items[i]

		if item.Type == "repository" && item.Repository.IsBare {
			// Start of bare repository group - collect all items in this group
			groupContent := r.renderNavigableItem(item, i, cursor, width, false)

			// Add all worktrees that belong to this bare repository
			j := i + 1
			worktreeStart := j
			for j < len(items) && items[j].Type == "worktree" && items[j].ParentRepo.Path == item.Repository.Path {
				j++
			}
			worktreeEnd := j

			// Render worktrees with knowledge of which is last
			for k := worktreeStart; k < worktreeEnd; k++ {
				isLastWorktree := (k == worktreeEnd-1)
				groupContent += "\n" + r.renderNavigableItem(items[k], k, cursor, width, isLastWorktree)
			}

			// No border - just add the group content directly
			content += groupContent + "\n"

			// Move index to after the group
			i = j
		} else {
			// Regular item (non-bare repository or standalone worktree)
			content += r.renderNavigableItem(item, i, cursor, width, false) + "\n"
			i++
		}
	}

	return content
}

// renderNavigableItem renders a single navigable item.
func (r *Renderer) renderNavigableItem(item types.NavigableItem, index, cursor int, width int, isLastWorktree bool) string {
	isSelected := index == cursor

	var style = r.styles.Item
	if isSelected {
		style = r.styles.SelectedItem
	}

	// Build front indicator
	var frontIndicator string
	if isSelected {
		frontIndicator = r.theme.Indicators.Selected
	} else {
		frontIndicator = strings.Repeat(" ", lipgloss.Width(r.theme.Indicators.Selected))
	}

	switch item.Type {
	case "repository":
		repo := item.Repository

		// Get repository icon based on type
		var repoIcon string
		if repo.IsBare {
			repoIcon = lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.IconBare)).Render(r.theme.Icons.Repository.Bare)
		} else if repo.IsWorktree {
			repoIcon = lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.IconWorktree)).Render(r.theme.Icons.Repository.Worktree)
		} else {
			repoIcon = lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.IconRegular)).Render(r.theme.Icons.Repository.Regular)
		}

		// Build status summary for the end
		var statusParts []string
		if repo.HasError {
			statusParts = append(statusParts, r.styles.StatusError.Render(r.theme.Indicators.Error))
		} else if !repo.IsBare {
			// Only show status for non-bare repositories
			if repo.HasUncommitted {
				statusParts = append(statusParts, r.styles.StatusUncommitted.Render(fmt.Sprintf("%s%d", r.theme.Indicators.Dirty, repo.UncommittedCount)))
			}
			if repo.HasUnpushed {
				statusParts = append(statusParts, r.styles.StatusUnpushed.Render(fmt.Sprintf("%s%d", r.theme.Indicators.Unpushed, repo.UnpushedCount)))
			}
			if repo.HasUntracked {
				statusParts = append(statusParts, r.styles.StatusUntracked.Render(fmt.Sprintf("%s%d", r.theme.Indicators.Untracked, repo.UntrackedCount)))
			}
			if len(statusParts) == 0 {
				statusParts = append(statusParts, r.styles.StatusClean.Render(r.theme.Indicators.Clean))
			}
		}

		// Build the main line with styled name if selected
		var repoName string
		if isSelected {
			repoName = lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.Selected)).Render(repo.Name)
		} else {
			repoName = repo.Name
		}
		repoLine := fmt.Sprintf(" %s%s %s", frontIndicator, repoIcon, repoName)

		// Build status summary
		statusSummary := strings.Join(statusParts, " ")

		// Calculate padding to fit status and end indicator
		repoLineWidth := lipgloss.Width(repoLine)
		statusWidth := lipgloss.Width(statusSummary)
		endIndicatorWidth := lipgloss.Width(r.theme.Indicators.SelectedEnd)
		spacingWidth := 2 // Space before status and space before end indicator
		reservedEndWidth := statusWidth + spacingWidth + endIndicatorWidth

		maxContentWidth := width - 2
		requiredWidth := repoLineWidth + reservedEndWidth

		var padding int
		if requiredWidth >= maxContentWidth {
			padding = 1
		} else {
			padding = maxContentWidth - repoLineWidth - reservedEndWidth
		}

		if padding < 0 {
			padding = 0
		}

		// Add status and end indicator
		var endIndicator string
		if isSelected {
			styledEndIndicator := lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.Selected)).Render(r.theme.Indicators.SelectedEnd)
			endIndicator = " " + statusSummary + " " + styledEndIndicator
		} else {
			endIndicator = " " + statusSummary + strings.Repeat(" ", endIndicatorWidth+1)
		}

		repoLine = repoLine + strings.Repeat(" ", padding) + endIndicator
		return style.Render(repoLine)

	case "worktree":
		worktree := item.WorktreeInfo
		parentRepo := item.ParentRepo

		// Use worktree icon
		worktreeIcon := lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.IconWorktree)).Render(r.theme.Icons.Repository.Worktree)

		// Choose the correct tree line based on position
		var treeLine string
		if isLastWorktree {
			treeLine = r.theme.Icons.Tree.Last
		} else {
			treeLine = r.theme.Icons.Tree.Branch
		}

		// Build branch info
		branchIcon := lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.Branch)).Render(r.theme.Icons.Branch.Icon)
		branchName := lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.Branch)).Render(worktree.Branch)
		branchInfo := fmt.Sprintf("%s%s", branchIcon, branchName)

		// Find the corresponding SubItem for this worktree to get status and counts
		var hasError, hasUncommitted, hasUnpushed, hasUntracked bool
		var uncommittedCount, unpushedCount, untrackedCount int

		for _, subItem := range parentRepo.SubItems {
			if subItem.Path == worktree.Path {
				hasError = subItem.HasError
				hasUncommitted = subItem.HasUncommitted
				hasUnpushed = subItem.HasUnpushed
				hasUntracked = subItem.HasUntracked
				uncommittedCount = subItem.UncommittedCount
				unpushedCount = subItem.UnpushedCount
				untrackedCount = subItem.UntrackedCount
				break
			}
		}

		// Build status summary for the end
		var statusParts []string
		if hasError {
			statusParts = append(statusParts, r.styles.StatusError.Render(r.theme.Indicators.Error))
		} else {
			if hasUncommitted {
				statusParts = append(statusParts, r.styles.StatusUncommitted.Render(fmt.Sprintf("%s%d", r.theme.Indicators.Dirty, uncommittedCount)))
			}
			if hasUnpushed {
				statusParts = append(statusParts, r.styles.StatusUnpushed.Render(fmt.Sprintf("%s%d", r.theme.Indicators.Unpushed, unpushedCount)))
			}
			if hasUntracked {
				statusParts = append(statusParts, r.styles.StatusUntracked.Render(fmt.Sprintf("%s%d", r.theme.Indicators.Untracked, untrackedCount)))
			}
			if len(statusParts) == 0 {
				statusParts = append(statusParts, r.styles.StatusClean.Render(r.theme.Indicators.Clean))
			}
		}

		// Build the main line with indentation for worktree and styled name if selected
		var worktreeName string
		if isSelected {
			worktreeName = lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.Selected)).Render(worktree.Name)
		} else {
			worktreeName = worktree.Name
		}
		worktreeLine := fmt.Sprintf(" %s%s%s %s %s", frontIndicator, treeLine, worktreeIcon, worktreeName, branchInfo)

		// Build status summary
		statusSummary := strings.Join(statusParts, " ")

		// Calculate padding to fit status and end indicator
		worktreeLineWidth := lipgloss.Width(worktreeLine)
		statusWidth := lipgloss.Width(statusSummary)
		endIndicatorWidth := lipgloss.Width(r.theme.Indicators.SelectedEnd)
		spacingWidth := 2 // Space before status and space before end indicator
		reservedEndWidth := statusWidth + spacingWidth + endIndicatorWidth

		maxContentWidth := width - 2
		requiredWidth := worktreeLineWidth + reservedEndWidth

		var padding int
		if requiredWidth >= maxContentWidth {
			padding = 1
		} else {
			padding = maxContentWidth - worktreeLineWidth - reservedEndWidth
		}

		if padding < 0 {
			padding = 0
		}

		// Add status and end indicator
		var endIndicator string
		if isSelected {
			styledEndIndicator := lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.Selected)).Render(r.theme.Indicators.SelectedEnd)
			endIndicator = " " + statusSummary + " " + styledEndIndicator
		} else {
			endIndicator = " " + statusSummary + strings.Repeat(" ", endIndicatorWidth+1)
		}

		worktreeLine = worktreeLine + strings.Repeat(" ", padding) + endIndicator
		return style.Render(worktreeLine)

	default:
		return ""
	}
}

// renderSummary renders a summary line showing repo/branch info and aggregated change counts
func (r *Renderer) renderSummary(summaryData repomanager.SummaryData, width int) string {
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
	return r.styles.Help.Render(summaryLine) + "\n"
}

// renderRepositorySummaryHeader renders a summary line for the repository list view
func (r *Renderer) renderRepositorySummaryHeader(summaryData repomanager.SummaryData, width int) string {
	return r.renderSummary(summaryData, width)
}

// renderSummaryHeader renders a summary line showing current repo/branch and total change counts
func (r *Renderer) renderSummaryHeader(summaryData repomanager.SummaryData, width int) string {
	return r.renderSummary(summaryData, width)
}

// renderEmptyState renders a message when no repositories are configured.
func (r *Renderer) renderEmptyState() string {
	return r.styles.Item.Render("No repositories configured.") + "\n\n"
}

// renderRepositoryList renders a list of repositories with cursor indication.
func (r *Renderer) renderRepositoryList(repositories []*repomanager.RepoItem, cursor int, width int) string {
	var content string
	for i, repo := range repositories {
		content += r.renderRepositoryItem(repo, i, cursor, width)
	}
	return content
}

// renderRepositoryItem renders a single repository item.
func (r *Renderer) renderRepositoryItem(repo *repomanager.RepoItem, index, cursor int, width int) string {
	isSelected := index == cursor

	var style = r.styles.Item
	if isSelected {
		style = r.styles.SelectedItem
	}

	// Build front indicator
	var frontIndicator string
	if isSelected {
		frontIndicator = r.theme.Indicators.Selected
	} else {
		frontIndicator = strings.Repeat(" ", lipgloss.Width(r.theme.Indicators.Selected))
	}

	// Get repository icon based on type
	var repoIcon string
	if repo.IsBare {
		repoIcon = lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.IconBare)).Render(r.theme.Icons.Repository.Bare)
	} else if repo.IsWorktree {
		repoIcon = lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.IconWorktree)).Render(r.theme.Icons.Repository.Worktree)
	} else {
		repoIcon = lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.IconRegular)).Render(r.theme.Icons.Repository.Regular)
	}

	// Build status summary for the end
	var statusParts []string
	if repo.HasError {
		statusParts = append(statusParts, r.styles.StatusError.Render("error"))
	} else if !repo.IsBare {
		// Only show status for non-bare repositories
		if repo.HasUncommitted {
			statusParts = append(statusParts, r.styles.StatusUncommitted.Render(fmt.Sprintf("%d uncommitted", repo.UncommittedCount)))
		}
		if repo.HasUnpushed {
			statusParts = append(statusParts, r.styles.StatusUnpushed.Render(fmt.Sprintf("%d unpushed", repo.UnpushedCount)))
		}
		if repo.HasUntracked {
			statusParts = append(statusParts, r.styles.StatusUntracked.Render(fmt.Sprintf("%d untracked", repo.UntrackedCount)))
		}
		if len(statusParts) == 0 {
			statusParts = append(statusParts, r.styles.StatusClean.Render("clean"))
		}
	}

	// Build the main line with styled name if selected
	var repoName string
	if isSelected {
		repoName = lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.Selected)).Render(repo.Name)
	} else {
		repoName = repo.Name
	}
	repoLine := fmt.Sprintf(" %s%s %s", frontIndicator, repoIcon, repoName)

	// Build status summary
	statusSummary := strings.Join(statusParts, " ")

	// Calculate padding to fit status and end indicator
	repoLineWidth := lipgloss.Width(repoLine)
	statusWidth := lipgloss.Width(statusSummary)
	endIndicatorWidth := lipgloss.Width(r.theme.Indicators.SelectedEnd)
	spacingWidth := 2 // Space before status and space before end indicator
	reservedEndWidth := statusWidth + spacingWidth + endIndicatorWidth

	maxContentWidth := width - 2
	requiredWidth := repoLineWidth + reservedEndWidth

	var padding int
	if requiredWidth >= maxContentWidth {
		padding = 1
	} else {
		padding = maxContentWidth - repoLineWidth - reservedEndWidth
	}

	if padding < 0 {
		padding = 0
	}

	// Add status and end indicator
	var endIndicator string
	if isSelected {
		styledEndIndicator := lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.Selected)).Render(r.theme.Indicators.SelectedEnd)
		endIndicator = " " + statusSummary + " " + styledEndIndicator
	} else {
		endIndicator = " " + statusSummary + strings.Repeat(" ", endIndicatorWidth+1)
	}

	repoLine = repoLine + strings.Repeat(" ", padding) + endIndicator

	return style.Render(repoLine) + "\n"
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
