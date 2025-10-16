package types

import (
	"github.com/jarmocluyse/git-dash/internal/repomanager"
)

// NavigableItem represents an item that can be navigated in the UI
type NavigableItem struct {
	Type         string // "repository" or "worktree"
	Repository   *repomanager.RepoItem
	WorktreeInfo *repomanager.SubItem
	ParentRepo   *repomanager.RepoItem // For worktrees, reference to parent bare repo
	IsLast       bool                  // For worktrees, indicates if this is the last worktree for the parent repo
}
