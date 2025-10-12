// Package git provides Git repository detection and status checking functionality.
package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WorktreeInfo contains information about a Git worktree.
type WorktreeInfo struct {
	Path   string
	Branch string
	Bare   bool
}

// StatusChecker provides Git repository status checking capabilities.
type StatusChecker interface {
	IsGitRepository(path string) bool
	IsBareRepository(path string) bool
	IsWorktree(path string) bool
	ListWorktrees(path string) ([]WorktreeInfo, error)
	HasUncommittedChanges(path string) bool
	HasUnpushedCommits(path string) bool
	HasUntrackedFiles(path string) bool
	GetCurrentBranch(path string) string
	CountUncommittedChanges(path string) int
	CountUnpushedCommits(path string) int
	CountUntrackedFiles(path string) int
}

// CommandLineChecker implements StatusChecker using Git command line.
type CommandLineChecker struct{}

// NewChecker creates a new command line based Git status checker.
func NewChecker() StatusChecker {
	return &CommandLineChecker{}
}

// IsGitRepository checks if the given path contains a Git repository.
func (g *CommandLineChecker) IsGitRepository(path string) bool {
	if g.hasGitDirectory(path) {
		return true
	}
	if g.isBareRepository(path) {
		return true
	}
	return g.canRunGitCommand(path)
}

// IsBareRepository checks if the repository at the given path is a bare repository.
func (g *CommandLineChecker) IsBareRepository(path string) bool {
	return g.isBareRepository(path)
}

// isBareRepository performs the actual bare repository check.
func (g *CommandLineChecker) isBareRepository(path string) bool {
	output, err := g.runGitCommand(path, "rev-parse", "--is-bare-repository")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

// hasGitDirectory checks if the path contains a .git directory or file.
func (g *CommandLineChecker) hasGitDirectory(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	return info.IsDir() || g.isWorktreeGitFile(gitPath)
}

// isWorktreeGitFile checks if the .git file is a worktree reference.
func (g *CommandLineChecker) isWorktreeGitFile(gitPath string) bool {
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	content, err := os.ReadFile(gitPath)
	if err != nil {
		return false
	}

	return strings.HasPrefix(string(content), "gitdir:")
}

// ListWorktrees returns all worktrees for the repository at the given path.
func (g *CommandLineChecker) ListWorktrees(path string) ([]WorktreeInfo, error) {
	output, err := g.runGitCommand(path, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	return g.parseWorktreeList(string(output))
}

// parseWorktreeList parses the output of git worktree list --porcelain.
func (g *CommandLineChecker) parseWorktreeList(output string) ([]WorktreeInfo, error) {
	var worktrees []WorktreeInfo
	var current WorktreeInfo

	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = WorktreeInfo{}
			}
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "worktree":
			current.Path = value
		case "branch":
			current.Branch = strings.TrimPrefix(value, "refs/heads/")
		case "bare":
			current.Bare = true
		}
	}

	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}

// IsWorktree checks if the given path is a Git worktree.
func (g *CommandLineChecker) IsWorktree(path string) bool {
	output, err := g.runGitCommand(path, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false
	}

	isInsideWorkTree := strings.TrimSpace(string(output)) == "true"
	if !isInsideWorkTree {
		return false
	}

	output, err = g.runGitCommand(path, "rev-parse", "--git-common-dir")
	if err != nil {
		return false
	}

	commonDir := strings.TrimSpace(string(output))
	output, err = g.runGitCommand(path, "rev-parse", "--git-dir")
	if err != nil {
		return false
	}

	gitDir := strings.TrimSpace(string(output))
	return commonDir != gitDir
}

// canRunGitCommand checks if git commands can be run in the given path.
func (g *CommandLineChecker) canRunGitCommand(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	return cmd.Run() == nil
}

// HasUncommittedChanges checks if the repository has uncommitted changes.
func (g *CommandLineChecker) HasUncommittedChanges(path string) bool {
	output, err := g.runGitCommand(path, "status", "--porcelain")
	if err != nil {
		return false
	}
	return g.hasOutput(output)
}

// HasUnpushedCommits checks if the repository has unpushed commits.
func (g *CommandLineChecker) HasUnpushedCommits(path string) bool {
	if g.branchIsAhead(path) {
		return true
	}
	return g.hasCommitsAheadOfUpstream(path)
}

// branchIsAhead checks if the current branch is ahead of its upstream.
func (g *CommandLineChecker) branchIsAhead(path string) bool {
	output, err := g.runGitCommand(path, "status", "--porcelain=v1", "--branch")
	if err != nil {
		return false
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		return false
	}

	return strings.Contains(lines[0], "[ahead")
}

// hasCommitsAheadOfUpstream checks for commits ahead of upstream using log.
func (g *CommandLineChecker) hasCommitsAheadOfUpstream(path string) bool {
	output, err := g.runGitCommand(path, "log", "--oneline", "@{u}..")
	if err != nil {
		return false
	}
	return g.hasOutput(output)
}

// HasUntrackedFiles checks if the repository has untracked files.
func (g *CommandLineChecker) HasUntrackedFiles(path string) bool {
	output, err := g.runGitCommand(path, "status", "--porcelain")
	if err != nil {
		return false
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "??") {
			return true
		}
	}
	return false
}

// GetCurrentBranch returns the current branch name.
func (g *CommandLineChecker) GetCurrentBranch(path string) string {
	output, err := g.runGitCommand(path, "branch", "--show-current")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// CountUncommittedChanges returns the number of files with uncommitted changes (tracked files only).
func (g *CommandLineChecker) CountUncommittedChanges(path string) int {
	output, err := g.runGitCommand(path, "status", "--porcelain")
	if err != nil {
		return 0
	}

	count := 0
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		// Skip untracked files (lines starting with ??)
		// Count modified files (M), added files (A), deleted files (D), renamed files (R), etc.
		if len(line) >= 2 && !strings.HasPrefix(line, "??") && strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

// CountUnpushedCommits returns the number of unpushed commits.
func (g *CommandLineChecker) CountUnpushedCommits(path string) int {
	// First try counting with log command
	output, err := g.runGitCommand(path, "log", "--oneline", "@{u}..")
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}

// CountUntrackedFiles returns the number of untracked files.
func (g *CommandLineChecker) CountUntrackedFiles(path string) int {
	output, err := g.runGitCommand(path, "status", "--porcelain")
	if err != nil {
		return 0
	}

	count := 0
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "??") {
			count++
		}
	}
	return count
}

// runGitCommand executes a git command in the specified directory.
func (g *CommandLineChecker) runGitCommand(path string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	return cmd.Output()
}

// hasOutput checks if the command output is non-empty.
func (g *CommandLineChecker) hasOutput(output []byte) bool {
	return len(strings.TrimSpace(string(output))) > 0
}
