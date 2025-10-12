package repo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/wip-tui/internal/git"
	"github.com/jarmocluyse/wip-tui/internal/repository"
	"github.com/jarmocluyse/wip-tui/internal/theme"
)

// Renderer handles rendering of repository items and worktrees
type Renderer struct {
	styles StyleConfig
	theme  theme.Theme
}

// StyleConfig contains all the styles needed for repo rendering
type StyleConfig struct {
	Item              lipgloss.Style
	SelectedItem      lipgloss.Style
	StatusUncommitted lipgloss.Style
	StatusUnpushed    lipgloss.Style
	StatusUntracked   lipgloss.Style
	StatusError       lipgloss.Style
	StatusClean       lipgloss.Style
	StatusNotAdded    lipgloss.Style
	Help              lipgloss.Style
	Branch            lipgloss.Style
	Border            lipgloss.Style
}

// NewRenderer creates a new repo renderer with the given styles and theme
func NewRenderer(styles StyleConfig, themeConfig theme.Theme) *Renderer {
	return &Renderer{
		styles: styles,
		theme:  themeConfig,
	}
}

// RenderRepository renders a repository item with cursor indication and selection
// This is the old method that includes worktrees - kept for backward compatibility
func (r *Renderer) RenderRepository(repo repository.Repository, isSelected bool, cursorIndicator string, width int) string {
	style := r.getItemStyle(isSelected)
	statusIndicator := r.getStatusIndicator(repo)

	// Get icon for repository
	var icon string
	if repo.IsBare {
		icon = "📁"
	} else if repo.IsWorktree {
		icon = "🌳"
	} else {
		icon = "🔗"
	}

	// Get branch information for non-bare repositories
	var branchInfo string
	if !repo.IsBare {
		gitChecker := git.NewChecker()
		branch := gitChecker.GetCurrentBranch(repo.Path)
		if branch != "" && branch != "unknown" {
			branchInfo = r.styles.Branch.Render(" 🌿 " + branch)
		}
	}

	// Create the main content without status (for left alignment)
	leftContent := fmt.Sprintf("%s %s %s%s", cursorIndicator, icon, repo.Name, branchInfo)

	// Calculate padding to right-align status (use actual terminal width)
	totalWidth := width
	if totalWidth <= 0 {
		totalWidth = 110 // fallback
	}
	leftWidth := lipgloss.Width(leftContent)
	statusWidth := lipgloss.Width(statusIndicator)

	// Create end-of-line selection indicator
	var endIndicator string
	if isSelected {
		endIndicator = r.styles.SelectedItem.Render("█")
	} else {
		endIndicator = " "
	}

	endIndicatorWidth := lipgloss.Width(endIndicator)
	padding := max(totalWidth-leftWidth-statusWidth-endIndicatorWidth-1, 1)

	// Format: cursor icon name branch [padding] status endIndicator
	line := leftContent + strings.Repeat(" ", padding) + statusIndicator + " " + endIndicator
	content := style.Render(line)

	// If this is a bare repository, add its worktrees to the content
	if repo.IsBare {
		// For backward compatibility, create a git checker for this old API
		gitChecker := git.NewChecker()
		worktreeContent := r.renderWorktrees(repo, width, gitChecker)
		if worktreeContent != "" {
			content += "\n" + worktreeContent
		}
	}

	// No border - return content directly
	return content + "\n"
}

// RenderRepositoryOnly renders just the repository without worktrees
// Use this when worktrees are rendered separately as navigable items
func (r *Renderer) RenderRepositoryOnly(repo repository.Repository, isSelected bool, cursorIndicator string, width int) string {
	style := r.getItemStyle(isSelected)
	statusIndicator := r.getStatusIndicator(repo)

	// Get icon for repository
	var icon string
	if repo.IsBare {
		icon = "📁"
	} else if repo.IsWorktree {
		icon = "🌳"
	} else {
		icon = "🔗"
	}

	// Get branch information for non-bare repositories
	var branchInfo string
	if !repo.IsBare {
		gitChecker := git.NewChecker()
		branch := gitChecker.GetCurrentBranch(repo.Path)
		if branch != "" && branch != "unknown" {
			branchInfo = r.styles.Branch.Render(" 🌿 " + branch)
		}
	}

	// Create the main content without status (for left alignment)
	leftContent := fmt.Sprintf("%s %s %s%s", cursorIndicator, icon, repo.Name, branchInfo)

	// Calculate padding to right-align status (use actual terminal width)
	totalWidth := width
	if totalWidth <= 0 {
		totalWidth = 110 // fallback
	}
	leftWidth := lipgloss.Width(leftContent)
	statusWidth := lipgloss.Width(statusIndicator)

	// Create end-of-line selection indicator
	var endIndicator string
	if isSelected {
		endIndicator = r.styles.SelectedItem.Render("█")
	} else {
		endIndicator = " "
	}

	endIndicatorWidth := lipgloss.Width(endIndicator)
	padding := max(totalWidth-leftWidth-statusWidth-endIndicatorWidth-1, 1)

	// Format: cursor icon name branch [padding] status endIndicator
	line := leftContent + strings.Repeat(" ", padding) + statusIndicator + " " + endIndicator
	content := style.Render(line)

	// Don't wrap bare repositories in border - they will be grouped with their worktrees
	if repo.IsBare {
		return content
	}

	// No border for regular repositories either
	return content
}

// RenderWorktree renders a single worktree item (without border - used in navigable mode)
func (r *Renderer) RenderWorktree(wt git.WorktreeInfo, parentName, bareRepoPath string, isSelected bool, cursorIndicator string, isLast bool, width int, gitChecker git.StatusChecker) string {
	style := r.getItemStyle(isSelected)
	status := r.getWorktreeStatusIndicators(wt.Path, gitChecker)

	// Use relative path in the name instead of separate path line
	relativePath := r.getRelativePathToBareRepo(wt.Path, bareRepoPath)
	branchInfo := r.styles.Branch.Render(" 🌿 " + wt.Branch)

	// Use different icon for last worktree
	var treeIcon string
	if isLast {
		treeIcon = "└─"
	} else {
		treeIcon = "├─"
	}

	// Create the main content without status (for left alignment)
	leftContent := fmt.Sprintf("%s %s 🌳 %s%s", cursorIndicator, treeIcon, relativePath, branchInfo)

	// Calculate padding to right-align status (use actual terminal width)
	totalWidth := width
	if totalWidth <= 0 {
		totalWidth = 140 // fallback
	}
	leftWidth := lipgloss.Width(leftContent)
	statusWidth := lipgloss.Width(status)

	// Create end-of-line selection indicator
	var endIndicator string
	if isSelected {
		endIndicator = r.styles.SelectedItem.Render("█")
	} else {
		endIndicator = " "
	}

	endIndicatorWidth := lipgloss.Width(endIndicator)
	// -1 for space
	padding := max(totalWidth-leftWidth-statusWidth-endIndicatorWidth-1, 1)

	// Format: cursor treeIcon 🌳 path branch [padding] status endIndicator
	line := leftContent + strings.Repeat(" ", padding) + status + " " + endIndicator
	content := style.Render(line)

	// Don't wrap in border - this makes worktrees appear as part of parent repo
	return content
}

// getItemStyle returns the appropriate style based on selection state.
func (r *Renderer) getItemStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return r.styles.SelectedItem
	}
	return r.styles.Item
}

// getStatusIndicator returns formatted status indicators for a repository.
func (r *Renderer) getStatusIndicator(repo repository.Repository) string {
	var indicators []string

	if repo.HasError {
		indicators = append(indicators, r.styles.StatusError.Render(r.theme.Indicators.Error))
		return strings.Join(indicators, " ")
	}

	// Don't show status icons for bare repos - they'll be shown on worktrees
	if repo.IsBare {
		return ""
	}

	if repo.IsWorktree {
		indicators = append(indicators, r.styles.Help.Render("🌳"))
	}

	if repo.HasUncommitted {
		indicators = append(indicators, r.styles.StatusUncommitted.Render(r.theme.Indicators.Dirty))
	}

	if repo.HasUnpushed {
		indicators = append(indicators, r.styles.StatusUnpushed.Render(r.theme.Indicators.Unpushed))
	}

	if repo.HasUntracked {
		indicators = append(indicators, r.styles.StatusUntracked.Render(r.theme.Indicators.Untracked))
	}

	if !repo.HasUncommitted && !repo.HasUnpushed && !repo.HasUntracked && !repo.IsBare {
		indicators = append(indicators, r.styles.StatusClean.Render(r.theme.Indicators.Clean))
	}

	return strings.Join(indicators, " ")
}

// renderWorktrees renders all valid worktrees for a bare repository.
func (r *Renderer) renderWorktrees(repo repository.Repository, width int, gitChecker git.StatusChecker) string {
	worktrees, err := gitChecker.ListWorktrees(repo.Path)
	if err != nil {
		return ""
	}

	var validWorktrees []git.WorktreeInfo
	for _, wt := range worktrees {
		// Skip the bare repository itself
		if wt.Bare || wt.Path == repo.Path {
			continue
		}
		validWorktrees = append(validWorktrees, wt)
	}

	var worktreeLines []string
	for i, wt := range validWorktrees {
		isLast := i == len(validWorktrees)-1
		worktreeLines = append(worktreeLines, r.renderWorktreeItem(wt, repo.Name, repo.Path, isLast, width, gitChecker))
	}

	return strings.Join(worktreeLines, "\n")
}

// renderWorktreeItem renders a single worktree item with tree structure indicators.
func (r *Renderer) renderWorktreeItem(wt git.WorktreeInfo, repoName string, bareRepoPath string, isLast bool, width int, gitChecker git.StatusChecker) string {
	// Create worktree status
	status := r.getWorktreeStatusIndicators(wt.Path, gitChecker)

	// Use relative path in the name instead of separate path line
	relativePath := r.getRelativePathToBareRepo(wt.Path, bareRepoPath)
	branchInfo := r.styles.Branch.Render(" 🌿 " + wt.Branch)

	// Use different icon for last worktree
	var treeIcon string
	if isLast {
		treeIcon = "└─"
	} else {
		treeIcon = "├─"
	}

	// Create the main content without status (for left alignment) - increased indentation
	leftContent := fmt.Sprintf("     %s 🌳 %s%s", treeIcon, relativePath, branchInfo)

	// Calculate padding to right-align status (use actual terminal width)
	totalWidth := width
	if totalWidth <= 0 {
		totalWidth = 140 // fallback
	}
	leftWidth := lipgloss.Width(leftContent)
	statusWidth := lipgloss.Width(status)
	padding := totalWidth - leftWidth - statusWidth
	if padding < 1 {
		padding = 1
	}

	// Format: [indent] treeIcon 🌳 path branch [padding] status
	line := leftContent + strings.Repeat(" ", padding) + status
	return r.styles.Item.Render(line)
}

// getRelativePathToBareRepo returns the relative path from the bare repository to the worktree.
func (r *Renderer) getRelativePathToBareRepo(worktreePath, bareRepoPath string) string {
	bareRepoDir := filepath.Dir(bareRepoPath)

	if relPath, err := filepath.Rel(bareRepoDir, worktreePath); err == nil {
		return relPath
	}

	return worktreePath
}

// getWorktreeStatusIndicators returns formatted status indicators for a worktree.
func (r *Renderer) getWorktreeStatusIndicators(path string, gitChecker git.StatusChecker) string {
	if !gitChecker.IsGitRepository(path) {
		return r.styles.StatusError.Render(r.theme.Indicators.Error)
	}

	var indicators []string

	if gitChecker.HasUncommittedChanges(path) {
		indicators = append(indicators, r.styles.StatusUncommitted.Render(r.theme.Indicators.Dirty))
	}

	if gitChecker.HasUnpushedCommits(path) {
		indicators = append(indicators, r.styles.StatusUnpushed.Render(r.theme.Indicators.Unpushed))
	}

	if gitChecker.HasUntrackedFiles(path) {
		indicators = append(indicators, r.styles.StatusUntracked.Render(r.theme.Indicators.Untracked))
	}

	if len(indicators) == 0 {
		return r.styles.StatusClean.Render(r.theme.Indicators.Clean)
	}

	return strings.Join(indicators, " ")
}
