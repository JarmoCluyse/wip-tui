// Package repomanager provides repository management with hierarchical items and worktrees.
package repomanager

// WorktreeInfo contains information about a Git worktree.
type WorktreeInfo struct {
	Path   string
	Branch string
	Bare   bool
}

// RepoItem represents a repository that can have sub-items (worktrees).
type RepoItem struct {
	Name             string
	Path             string
	HasUncommitted   bool
	HasUnpushed      bool
	HasUntracked     bool
	HasError         bool
	IsWorktree       bool
	IsBare           bool
	UncommittedCount int
	UnpushedCount    int
	UntrackedCount   int
	SubItems         []*SubItem // Worktrees for this repository
}

// SubItem represents a worktree or other sub-component of a repository.
type SubItem struct {
	Name             string
	Path             string
	Branch           string
	HasUncommitted   bool
	HasUnpushed      bool
	HasUntracked     bool
	HasError         bool
	UncommittedCount int
	UnpushedCount    int
	UntrackedCount   int
	ParentRepo       *RepoItem
}

// SummaryData holds aggregated summary information.
type SummaryData struct {
	TotalUncommitted int
	TotalUnpushed    int
	TotalUntracked   int
	TotalErrors      int
}
