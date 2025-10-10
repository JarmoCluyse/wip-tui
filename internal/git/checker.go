package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type WorktreeInfo struct {
	Path   string
	Branch string
	Bare   bool
}

type StatusChecker interface {
	IsGitRepository(path string) bool
	IsBareRepository(path string) bool
	IsWorktree(path string) bool
	ListWorktrees(path string) ([]WorktreeInfo, error)
	HasUncommittedChanges(path string) bool
	HasUnpushedCommits(path string) bool
	HasUntrackedFiles(path string) bool
	GetCurrentBranch(path string) string
}

type CommandLineChecker struct{}

func NewChecker() StatusChecker {
	return &CommandLineChecker{}
}

func (g *CommandLineChecker) IsGitRepository(path string) bool {
	if g.hasGitDirectory(path) {
		return true
	}
	if g.isBareRepository(path) {
		return true
	}
	return g.canRunGitCommand(path)
}

func (g *CommandLineChecker) IsBareRepository(path string) bool {
	return g.isBareRepository(path)
}

func (g *CommandLineChecker) isBareRepository(path string) bool {
	output, err := g.runGitCommand(path, "rev-parse", "--is-bare-repository")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

func (g *CommandLineChecker) hasGitDirectory(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	return info.IsDir() || g.isWorktreeGitFile(gitPath)
}

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

func (g *CommandLineChecker) ListWorktrees(path string) ([]WorktreeInfo, error) {
	output, err := g.runGitCommand(path, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	return g.parseWorktreeList(string(output))
}

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

func (g *CommandLineChecker) canRunGitCommand(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	return cmd.Run() == nil
}

func (g *CommandLineChecker) HasUncommittedChanges(path string) bool {
	output, err := g.runGitCommand(path, "status", "--porcelain")
	if err != nil {
		return false
	}
	return g.hasOutput(output)
}

func (g *CommandLineChecker) HasUnpushedCommits(path string) bool {
	if g.branchIsAhead(path) {
		return true
	}
	return g.hasCommitsAheadOfUpstream(path)
}

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

func (g *CommandLineChecker) hasCommitsAheadOfUpstream(path string) bool {
	output, err := g.runGitCommand(path, "log", "--oneline", "@{u}..")
	if err != nil {
		return false
	}
	return g.hasOutput(output)
}

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

func (g *CommandLineChecker) GetCurrentBranch(path string) string {
	output, err := g.runGitCommand(path, "branch", "--show-current")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func (g *CommandLineChecker) runGitCommand(path string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	return cmd.Output()
}

func (g *CommandLineChecker) hasOutput(output []byte) bool {
	return len(strings.TrimSpace(string(output))) > 0
}
