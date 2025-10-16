package details

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jarmocluyse/git-dash/internal/git"
	"github.com/jarmocluyse/git-dash/internal/repository"
	"github.com/jarmocluyse/git-dash/internal/theme"
	"github.com/jarmocluyse/git-dash/ui/components/help"
	"github.com/jarmocluyse/git-dash/ui/header"
	"github.com/jarmocluyse/git-dash/ui/types"
)

// Renderer handles rendering of the repository details page
type Renderer struct {
	styles StyleConfig
	theme  theme.Theme
	header *header.Renderer
}

// NewRenderer creates a new details page renderer
func NewRenderer(styles StyleConfig, themeConfig theme.Theme) *Renderer {
	return &Renderer{
		styles: styles,
		theme:  themeConfig,
		header: header.NewRenderer(themeConfig),
	}
}

// Render renders the repository details view
func (r *Renderer) Render(item types.NavigableItem, width, height int) string {
	var title string
	var content string

	if item.Type == "repository" {
		title = fmt.Sprintf("Repository Details: %s", item.Repository.Name)
		content = r.renderRepositoryDetails(*item.Repository)
	} else if item.Type == "worktree" {
		title = fmt.Sprintf("Worktree Details: %s", filepath.Base(item.WorktreeInfo.Path))
		content = r.renderWorktreeDetails(*item.WorktreeInfo, item.ParentRepo)
	}

	headerContent := r.header.Render(title, width)

	// Calculate content area (reserve space for help at bottom)
	headerLines := strings.Count(headerContent, "\n") + 1
	contentHeight := height - headerLines - 3 // -3 for help line and spacing

	// Create bordered content
	borderedContent := r.styles.Border.
		Width(width - 4).
		Height(contentHeight).
		Render(content)

	// Combine header and bordered content
	mainContent := lipgloss.JoinVertical(lipgloss.Left,
		headerContent,
		"",
		borderedContent,
	)

	// Use help component to render with bottom-aligned help
	helpBuilder := help.NewBuilder(r.styles.Help)
	bindings := []help.KeyBinding{
		{Key: "b", Description: "back"},
		{Key: "Esc", Description: "back"},
	}

	return helpBuilder.RenderWithBottomHelp(mainContent, bindings, width, height)
}

// renderRepositoryDetails renders detailed information for a repository.
func (r *Renderer) renderRepositoryDetails(repo repository.Repository) string {
	var details []string

	// Basic info
	details = append(details, r.renderField("Name", repo.Name))
	details = append(details, r.renderField("Path", repo.Path))
	details = append(details, r.renderField("Type", r.getRepoType(repo)))

	// Show cached status information
	if repo.HasError {
		details = append(details, r.renderField("Status", "Error"))
	} else {
		var statusParts []string
		if repo.HasUncommitted {
			statusParts = append(statusParts, "Uncommitted changes")
		}
		if repo.HasUntracked {
			statusParts = append(statusParts, "Untracked files")
		}
		if repo.HasUnpushed {
			statusParts = append(statusParts, "Unpushed commits")
		}

		if len(statusParts) > 0 {
			details = append(details, r.renderField("Status", strings.Join(statusParts, ", ")))
		} else {
			details = append(details, r.renderField("Status", "Clean"))
		}
	}

	// Additional info
	if repo.AutoDiscover {
		details = append(details, r.renderField("Auto-discover", "Enabled"))
	}

	if repo.IsWorktree {
		details = append(details, r.renderField("Is Worktree", "Yes"))
	}

	return strings.Join(details, "\n")
}

// renderWorktreeDetails renders detailed information for a worktree.
func (r *Renderer) renderWorktreeDetails(worktree git.WorktreeInfo, parentRepo *repository.Repository) string {
	var details []string

	// Basic info
	details = append(details, r.renderField("Name", filepath.Base(worktree.Path)))
	details = append(details, r.renderField("Path", worktree.Path))
	details = append(details, r.renderField("Parent Repository", parentRepo.Name))
	details = append(details, r.renderField("Parent Path", parentRepo.Path))

	// Branch info if available
	if worktree.Branch != "" {
		details = append(details, r.renderField("Branch", worktree.Branch))
	}

	if worktree.Bare {
		details = append(details, r.renderField("Type", "Bare worktree"))
	} else {
		details = append(details, r.renderField("Type", "Regular worktree"))
	}

	return strings.Join(details, "\n")
}

// renderField renders a labeled field with consistent formatting.
func (r *Renderer) renderField(label, value string) string {
	labelStyled := r.styles.Label.Render(label + ":")
	valueStyled := r.styles.Value.Render(value)
	return fmt.Sprintf("%-20s %s", labelStyled, valueStyled)
}

// getRepoType returns a human-readable repository type description.
func (r *Renderer) getRepoType(repo repository.Repository) string {
	if repo.IsBare {
		return "Bare Repository"
	}
	return "Regular Repository"
}
