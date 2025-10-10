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
func (r *Renderer) RenderRepository(repo repository.Repository, isSelected bool, cursorIndicator string) string {
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

	// Calculate padding to right-align status (assuming 110 char width to account for border)
	totalWidth := 110
	leftWidth := lipgloss.Width(leftContent)
	statusWidth := lipgloss.Width(statusIndicator)
	padding := totalWidth - leftWidth - statusWidth
	if padding < 1 {
		padding = 1
	}

	// Format: cursor icon name branch [padding] status
	line := leftContent + strings.Repeat(" ", padding) + statusIndicator
	content := style.Render(line)

	// If this is a bare repository, add its worktrees to the content
	if repo.IsBare {
		worktreeContent := r.renderWorktrees(repo)
		if worktreeContent != "" {
			content += "\n" + worktreeContent
		}
	}

	// Wrap all content in border
	borderedContent := r.styles.Border.Render(content)
	return borderedContent + "\n"
}

// RenderRepositoryOnly renders just the repository without worktrees
// Use this when worktrees are rendered separately as navigable items
func (r *Renderer) RenderRepositoryOnly(repo repository.Repository, isSelected bool, cursorIndicator string) string {
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

	// Calculate padding to right-align status (assuming 110 char width to account for border)
	totalWidth := 110
	leftWidth := lipgloss.Width(leftContent)
	statusWidth := lipgloss.Width(statusIndicator)
	padding := totalWidth - leftWidth - statusWidth
	if padding < 1 {
		padding = 1
	}

	// Format: cursor icon name branch [padding] status
	line := leftContent + strings.Repeat(" ", padding) + statusIndicator
	content := style.Render(line)

	// Don't wrap bare repositories in border - they will be grouped with their worktrees
	if repo.IsBare {
		return content
	}

	// Wrap content in border for regular repositories
	borderedContent := r.styles.Border.Render(content)
	return borderedContent
}

// RenderWorktree renders a single worktree item (without border - used in navigable mode)
func (r *Renderer) RenderWorktree(wt git.WorktreeInfo, parentName, bareRepoPath string, isSelected bool, cursorIndicator string, isLast bool) string {
	style := r.getItemStyle(isSelected)
	status := r.getWorktreeStatusIndicators(wt.Path)

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

	// Calculate padding to right-align status (assuming 110 char width to account for border)
	totalWidth := 110
	leftWidth := lipgloss.Width(leftContent)
	statusWidth := lipgloss.Width(status)
	padding := totalWidth - leftWidth - statusWidth
	if padding < 1 {
		padding = 1
	}

	// Format: cursor treeIcon 🌳 path branch [padding] status
	line := leftContent + strings.Repeat(" ", padding) + status
	content := style.Render(line)

	// Don't wrap in border - this makes worktrees appear as part of parent repo
	return content
}

func (r *Renderer) getItemStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return r.styles.SelectedItem
	}
	return r.styles.Item
}

func (r *Renderer) getStatusIndicator(repo repository.Repository) string {
	var indicators []string

	if repo.HasError {
		indicators = append(indicators, r.styles.StatusError.Render(r.theme.Indicators.Error))
		return strings.Join(indicators, " ")
	}

	// Don't show status icons for bare repos - they'll be shown on worktrees
	if repo.IsBare {
		return "📁"
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

func (r *Renderer) renderWorktrees(repo repository.Repository) string {
	gitChecker := git.NewChecker()
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
		worktreeLines = append(worktreeLines, r.renderWorktreeItem(wt, repo.Name, repo.Path, isLast))
	}

	return strings.Join(worktreeLines, "\n")
}

func (r *Renderer) renderWorktreeItem(wt git.WorktreeInfo, repoName string, bareRepoPath string, isLast bool) string {
	// Create worktree status
	status := r.getWorktreeStatusIndicators(wt.Path)

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
	leftContent := fmt.Sprintf("   %s 🌳 %s%s", treeIcon, relativePath, branchInfo)

	// Calculate padding to right-align status (assuming 110 char width to account for border)
	totalWidth := 110
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

func (r *Renderer) getRelativePathToBareRepo(worktreePath, bareRepoPath string) string {
	bareRepoDir := filepath.Dir(bareRepoPath)

	if relPath, err := filepath.Rel(bareRepoDir, worktreePath); err == nil {
		return relPath
	}

	return worktreePath
}

func (r *Renderer) getWorktreeStatusIndicators(path string) string {
	gitChecker := git.NewChecker()

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
