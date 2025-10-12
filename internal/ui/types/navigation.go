package types

import (
	"github.com/jarmocluyse/wip-tui/internal/git"
	"github.com/jarmocluyse/wip-tui/internal/repository"
)

// NavigableItem represents an item that can be navigated in the UI
type NavigableItem struct {
	Type         string // "repository" or "worktree"
	Repository   *repository.Repository
	WorktreeInfo *git.WorktreeInfo
	ParentRepo   *repository.Repository // For worktrees, reference to parent bare repo
	IsLast       bool                   // For worktrees, indicates if this is the last worktree for the parent repo
}
