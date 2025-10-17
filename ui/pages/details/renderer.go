package details

import (
	"fmt"
	"strings"

	"github.com/jarmocluyse/git-dash/internal/repomanager"
	theme "github.com/jarmocluyse/git-dash/internal/theme/types"
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
	var detailsContent string

	switch item.Type {
	case "repository":
		detailsContent = r.renderRepositoryDetails(item.Repository)
	case "worktree":
		detailsContent = r.renderWorktreeDetails(item.WorktreeInfo, item.ParentRepo)
	}

	// Build content with git-dash title like home page, then repository details title
	content := r.header.RenderWithCountAndSpacing("git-dash", "", 1, width)
	content += "\n"

	// Add bordered details content
	borderedContent := r.styles.Border.
		Width(width - 2).
		Render(detailsContent)

	content += borderedContent

	// Use help component the same way as home page
	helpBuilder := help.NewBuilder(r.styles.Help)
	bindings := []help.KeyBinding{
		{Key: "b", Description: "back"},
		{Key: "Esc", Description: "back"},
	}

	// Use header count of 4 for git-dash title + details title
	return helpBuilder.RenderWithBottomHelpAndHeader(content, bindings, width, height, 4)
}

// renderRepositoryDetails renders detailed information for a repository.
func (r *Renderer) renderRepositoryDetails(repo *repomanager.RepoItem) string {
	var details []string

	// Basic info
	details = append(details, r.renderField("Name", repo.Name))
	details = append(details, r.renderField("Path", repo.Path))
	details = append(details, r.renderField("Type", r.getRepoType(repo)))

	// Show cached status information (only for non-bare repositories)
	if repo.HasError {
		details = append(details, r.renderField("Status", "Error"))
	} else if !repo.IsBare {
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

	if repo.IsWorktree {
		details = append(details, r.renderField("Is Worktree", "Yes"))
	}

	if repo.IsBare {
		details = append(details, r.renderField("Is Bare", "Yes"))
	}

	return strings.Join(details, "\n")
}

// renderWorktreeDetails renders detailed information for a worktree.
func (r *Renderer) renderWorktreeDetails(worktree *repomanager.SubItem, parentRepo *repomanager.RepoItem) string {
	var details []string

	// Basic info
	details = append(details, r.renderField("Name", worktree.Name))
	details = append(details, r.renderField("Path", worktree.Path))
	details = append(details, r.renderField("Branch", worktree.Branch))
	details = append(details, r.renderField("Parent Repository", parentRepo.Name))

	// Show cached status information
	if worktree.HasError {
		details = append(details, r.renderField("Status", "Error"))
	} else {
		var statusParts []string
		if worktree.HasUncommitted {
			statusParts = append(statusParts, "Uncommitted changes")
		}
		if worktree.HasUntracked {
			statusParts = append(statusParts, "Untracked files")
		}
		if worktree.HasUnpushed {
			statusParts = append(statusParts, "Unpushed commits")
		}

		if len(statusParts) > 0 {
			details = append(details, r.renderField("Status", strings.Join(statusParts, ", ")))
		} else {
			details = append(details, r.renderField("Status", "Clean"))
		}
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
func (r *Renderer) getRepoType(repo *repomanager.RepoItem) string {
	if repo.IsBare {
		return "Bare Repository"
	}
	return "Regular Repository"
}
