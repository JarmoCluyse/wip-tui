package repo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/git-dash/internal/git"
	"github.com/jarmocluyse/git-dash/internal/repository"
	"github.com/jarmocluyse/git-dash/internal/theme"
)

// ItemType represents the type of item being rendered
type ItemType int

const (
	ItemTypeRepository ItemType = iota
	ItemTypeWorktree
	ItemTypeWorktreeInTree
)

// RenderableItem represents any item that can be rendered (repository or worktree)
type RenderableItem struct {
	Type         ItemType
	Repository   *repository.Repository
	Worktree     *git.WorktreeInfo
	ParentName   string // For worktrees, the name of the parent repository
	BareRepoPath string // For worktrees, the path to the bare repository
	IsLast       bool   // For tree rendering, whether this is the last item
	RelativePath string // For worktrees, relative path to display
}

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
	IconRegular       lipgloss.Style
	IconBare          lipgloss.Style
	IconWorktree      lipgloss.Style
}

// NewRenderer creates a new repo renderer with the given styles and theme
func NewRenderer(styles StyleConfig, themeConfig theme.Theme) *Renderer {
	return &Renderer{
		styles: styles,
		theme:  themeConfig,
	}
}

// RenderItem renders any type of item (repository or worktree) with consistent formatting
func (r *Renderer) RenderItem(item RenderableItem, isSelected bool, width int, gitChecker git.StatusChecker) string {
	style := r.getItemStyle(isSelected)

	var iconRendered, name, branchInfo, status, treePrefix string

	switch item.Type {
	case ItemTypeRepository:
		repo := item.Repository
		// Get repository icon (don't apply color styling if selected - let the main style handle it)
		var iconText string
		if repo.IsBare {
			iconText = r.theme.Icons.Repository.Bare
			if !isSelected {
				iconRendered = r.styles.IconBare.Render(iconText)
			} else {
				iconRendered = iconText
			}
		} else if repo.IsWorktree {
			iconText = r.theme.Icons.Repository.Worktree
			if !isSelected {
				iconRendered = r.styles.IconWorktree.Render(iconText)
			} else {
				iconRendered = iconText
			}
		} else {
			iconText = r.theme.Icons.Repository.Regular
			if !isSelected {
				iconRendered = r.styles.IconRegular.Render(iconText)
			} else {
				iconRendered = iconText
			}
		}

		name = repo.Name
		status = r.getStatusIndicator(*repo)

		// Get branch info for non-bare repositories
		if !repo.IsBare {
			if gitChecker == nil {
				gitChecker = git.NewChecker()
			}
			branch := gitChecker.GetCurrentBranch(repo.Path)
			if branch != "" && branch != "unknown" {
				branchText := " " + r.theme.Icons.Branch.Icon + " " + branch
				branchInfo = r.styles.Branch.Render(branchText)
			}
		}

	case ItemTypeWorktree, ItemTypeWorktreeInTree:
		wt := item.Worktree
		iconText := r.theme.Icons.Repository.Worktree
		if !isSelected {
			iconRendered = r.styles.IconWorktree.Render(iconText)
		} else {
			iconRendered = iconText
		}

		if item.RelativePath != "" {
			name = item.RelativePath
		} else {
			name = r.getRelativePathToBareRepo(wt.Path, item.BareRepoPath)
		}

		branchText := " " + r.theme.Icons.Branch.Icon + " " + wt.Branch
		branchInfo = r.styles.Branch.Render(branchText)
		status = r.getWorktreeStatusIndicators(wt.Path, gitChecker)

		// Add tree structure for worktrees in tree view
		if item.Type == ItemTypeWorktreeInTree {
			if item.IsLast {
				treePrefix = r.theme.Icons.Tree.Last + " "
			} else {
				treePrefix = r.theme.Icons.Tree.Branch + " "
			}
		}
	}

	// Create the main content with consistent spacing
	var frontIndicator string
	if isSelected {
		frontIndicator = r.theme.Indicators.Selected
	} else {
		frontIndicator = strings.Repeat(" ", lipgloss.Width(r.theme.Indicators.Selected))
	}

	leftContent := fmt.Sprintf(" %s%s%s %s%s", frontIndicator, treePrefix, iconRendered, name, branchInfo)

	// Calculate padding to right-align status, always reserving space for end indicator
	totalWidth := width
	if totalWidth <= 0 {
		totalWidth = 140 // fallback
	}
	leftWidth := lipgloss.Width(leftContent)
	statusWidth := lipgloss.Width(status)

	// Always reserve space for end indicator and spacing, even when not selected
	endIndicatorWidth := lipgloss.Width(r.theme.Indicators.SelectedEnd)
	spacingWidth := 1 // One space before end indicator
	reservedEndWidth := endIndicatorWidth + spacingWidth

	// Ensure we don't exceed the terminal width - leave some margin
	maxContentWidth := totalWidth - 2 // margin for safety
	requiredWidth := leftWidth + statusWidth + reservedEndWidth

	var padding int
	if requiredWidth >= maxContentWidth {
		// Content too wide, use minimal spacing
		padding = 1
	} else {
		padding = maxContentWidth - leftWidth - statusWidth - reservedEndWidth
	}

	if padding < 1 {
		padding = 1
	}

	// Always add the reserved space, but only show the indicator when selected
	var endIndicator string
	if isSelected {
		styledEndIndicator := lipgloss.NewStyle().Foreground(lipgloss.Color(r.theme.Colors.Selected)).Render(r.theme.Indicators.SelectedEnd)
		endIndicator = " " + styledEndIndicator
	} else {
		// Reserve the space with spaces to prevent layout shift
		endIndicator = strings.Repeat(" ", endIndicatorWidth+1)
	}

	line := leftContent + strings.Repeat(" ", padding) + status + endIndicator
	return style.Render(line) + "\n"
}

// RenderRepository renders a repository item with selection
// This is the old method that includes worktrees - kept for backward compatibility
func (r *Renderer) RenderRepository(repo repository.Repository, isSelected bool, cursorIndicator string, width int) string {
	item := RenderableItem{
		Type:       ItemTypeRepository,
		Repository: &repo,
	}

	result := r.RenderItem(item, isSelected, width, nil)

	// If this is a bare repository, add its worktrees to the content
	if repo.IsBare {
		gitChecker := git.NewChecker()
		worktrees := r.renderWorktrees(repo, width, gitChecker)
		result += worktrees
	}

	// Don't wrap bare repositories in border - they will be grouped with their worktrees
	if repo.IsBare {
		return result
	}

	// Wrap regular and worktree repositories in a border
	return r.styles.Border.Render(result)
}

// RenderRepositoryOnly renders just the repository without worktrees
// Use this when worktrees are rendered separately as navigable items
func (r *Renderer) RenderRepositoryOnly(repo repository.Repository, isSelected bool, cursorIndicator string, width int) string {
	item := RenderableItem{
		Type:       ItemTypeRepository,
		Repository: &repo,
	}

	// Use the unified renderer and remove the trailing newline
	result := r.RenderItem(item, isSelected, width, nil)
	return strings.TrimSuffix(result, "\n")
}

// RenderWorktree renders a single worktree item (without border - used in navigable mode)
func (r *Renderer) RenderWorktree(wt git.WorktreeInfo, parentName, bareRepoPath string, isSelected bool, cursorIndicator string, isLast bool, width int, gitChecker git.StatusChecker) string {
	item := RenderableItem{
		Type:         ItemTypeWorktreeInTree,
		Worktree:     &wt,
		ParentName:   parentName,
		BareRepoPath: bareRepoPath,
		IsLast:       isLast,
	}

	// Remove the trailing newline since the unified function adds it
	result := r.RenderItem(item, isSelected, width, gitChecker)
	return strings.TrimSuffix(result, "\n")
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
		indicators = append(indicators, r.styles.IconWorktree.Render(r.theme.Icons.Repository.Worktree))
	}

	if repo.HasUncommitted {
		if repo.UncommittedCount > 0 {
			indicators = append(indicators, r.styles.StatusUncommitted.Render(fmt.Sprintf("%s%d", r.theme.Indicators.Dirty, repo.UncommittedCount)))
		} else {
			indicators = append(indicators, r.styles.StatusUncommitted.Render(r.theme.Indicators.Dirty))
		}
	}

	if repo.HasUnpushed {
		if repo.UnpushedCount > 0 {
			indicators = append(indicators, r.styles.StatusUnpushed.Render(fmt.Sprintf("%s%d", r.theme.Indicators.Unpushed, repo.UnpushedCount)))
		} else {
			indicators = append(indicators, r.styles.StatusUnpushed.Render(r.theme.Indicators.Unpushed))
		}
	}

	if repo.HasUntracked {
		if repo.UntrackedCount > 0 {
			indicators = append(indicators, r.styles.StatusUntracked.Render(fmt.Sprintf("%s%d", r.theme.Indicators.Untracked, repo.UntrackedCount)))
		} else {
			indicators = append(indicators, r.styles.StatusUntracked.Render(r.theme.Indicators.Untracked))
		}
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
	branchInfo := r.styles.Branch.Render(" " + r.theme.Icons.Branch.Icon + " " + wt.Branch)

	// Use different icon for last worktree
	var treeIcon string
	if isLast {
		treeIcon = r.theme.Icons.Tree.Last
	} else {
		treeIcon = r.theme.Icons.Tree.Branch
	}

	// Create the main content without status (for left alignment) - increased indentation
	leftContent := fmt.Sprintf("     %s %s %s%s", treeIcon, r.styles.IconWorktree.Render(r.theme.Icons.Repository.Worktree), relativePath, branchInfo)

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
		count := gitChecker.CountUncommittedChanges(path)
		if count > 0 {
			indicators = append(indicators, r.styles.StatusUncommitted.Render(fmt.Sprintf("%s%d", r.theme.Indicators.Dirty, count)))
		} else {
			indicators = append(indicators, r.styles.StatusUncommitted.Render(r.theme.Indicators.Dirty))
		}
	}

	if gitChecker.HasUnpushedCommits(path) {
		count := gitChecker.CountUnpushedCommits(path)
		if count > 0 {
			indicators = append(indicators, r.styles.StatusUnpushed.Render(fmt.Sprintf("%s%d", r.theme.Indicators.Unpushed, count)))
		} else {
			indicators = append(indicators, r.styles.StatusUnpushed.Render(r.theme.Indicators.Unpushed))
		}
	}

	if gitChecker.HasUntrackedFiles(path) {
		count := gitChecker.CountUntrackedFiles(path)
		if count > 0 {
			indicators = append(indicators, r.styles.StatusUntracked.Render(fmt.Sprintf("%s%d", r.theme.Indicators.Untracked, count)))
		} else {
			indicators = append(indicators, r.styles.StatusUntracked.Render(r.theme.Indicators.Untracked))
		}
	}

	if len(indicators) == 0 {
		return r.styles.StatusClean.Render(r.theme.Indicators.Clean)
	}

	return strings.Join(indicators, " ")
}
