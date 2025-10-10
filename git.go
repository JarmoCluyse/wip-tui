package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitStatusChecker interface {
	IsGitRepository(path string) bool
	IsBareRepository(path string) bool
	IsWorktree(path string) bool
	ListWorktrees(path string) ([]WorktreeInfo, error)
	HasUncommittedChanges(path string) bool
	HasUnpushedCommits(path string) bool
	HasUntrackedFiles(path string) bool
	GetCurrentBranch(path string) string
}

type WorktreeInfo struct {
	Path   string
	Branch string
	Bare   bool
}

type CommandLineGitChecker struct{}

func NewGitChecker() GitStatusChecker {
	return &CommandLineGitChecker{}
}

func (g *CommandLineGitChecker) IsGitRepository(path string) bool {
	if g.hasGitDirectory(path) {
		return true
	}
	if g.isBareRepository(path) {
		return true
	}
	return g.canRunGitCommand(path)
}

func (g *CommandLineGitChecker) IsBareRepository(path string) bool {
	return g.isBareRepository(path)
}

func (g *CommandLineGitChecker) isBareRepository(path string) bool {
	output, err := g.runGitCommand(path, "rev-parse", "--is-bare-repository")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

func (g *CommandLineGitChecker) hasGitDirectory(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	return info.IsDir() || g.isWorktreeGitFile(gitPath)
}

func (g *CommandLineGitChecker) isWorktreeGitFile(gitPath string) bool {
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

func (g *CommandLineGitChecker) ListWorktrees(path string) ([]WorktreeInfo, error) {
	output, err := g.runGitCommand(path, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	return g.parseWorktreeList(string(output))
}

func (g *CommandLineGitChecker) parseWorktreeList(output string) ([]WorktreeInfo, error) {
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

func (g *CommandLineGitChecker) IsWorktree(path string) bool {
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

func (g *CommandLineGitChecker) canRunGitCommand(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	return cmd.Run() == nil
}

func (g *CommandLineGitChecker) HasUncommittedChanges(path string) bool {
	output, err := g.runGitCommand(path, "status", "--porcelain")
	if err != nil {
		return false
	}
	return g.hasOutput(output)
}

func (g *CommandLineGitChecker) HasUnpushedCommits(path string) bool {
	if g.branchIsAhead(path) {
		return true
	}
	return g.hasCommitsAheadOfUpstream(path)
}

func (g *CommandLineGitChecker) branchIsAhead(path string) bool {
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

func (g *CommandLineGitChecker) hasCommitsAheadOfUpstream(path string) bool {
	output, err := g.runGitCommand(path, "log", "--oneline", "@{u}..")
	if err != nil {
		return false
	}
	return g.hasOutput(output)
}

func (g *CommandLineGitChecker) HasUntrackedFiles(path string) bool {
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

func (g *CommandLineGitChecker) GetCurrentBranch(path string) string {
	output, err := g.runGitCommand(path, "branch", "--show-current")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func (g *CommandLineGitChecker) runGitCommand(path string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	return cmd.Output()
}

func (g *CommandLineGitChecker) hasOutput(output []byte) bool {
	return len(strings.TrimSpace(string(output))) > 0
}

type RepositoryStatusUpdater struct {
	gitChecker GitStatusChecker
}

func NewRepositoryStatusUpdater(gitChecker GitStatusChecker) *RepositoryStatusUpdater {
	return &RepositoryStatusUpdater{
		gitChecker: gitChecker,
	}
}

func (r *RepositoryStatusUpdater) UpdateStatus(repo *Repository) {
	if !r.gitChecker.IsGitRepository(repo.Path) {
		r.setErrorStatus(repo)
		return
	}

	repo.IsBare = r.gitChecker.IsBareRepository(repo.Path)
	repo.IsWorktree = r.gitChecker.IsWorktree(repo.Path)
	repo.HasError = false

	if repo.IsBare {
		r.updateBareRepositoryStatus(repo)
	} else {
		r.updateRegularRepositoryStatus(repo)
	}
}

func (r *RepositoryStatusUpdater) updateBareRepositoryStatus(repo *Repository) {
	repo.HasUncommitted = false
	repo.HasUntracked = false

	worktrees, err := r.gitChecker.ListWorktrees(repo.Path)
	if err != nil {
		repo.HasUnpushed = false
		return
	}

	repo.HasUnpushed = len(worktrees) > 0
}

func (r *RepositoryStatusUpdater) updateRegularRepositoryStatus(repo *Repository) {
	repo.HasUncommitted = r.gitChecker.HasUncommittedChanges(repo.Path)
	repo.HasUnpushed = r.gitChecker.HasUnpushedCommits(repo.Path)
	repo.HasUntracked = r.gitChecker.HasUntrackedFiles(repo.Path)
}

func (r *RepositoryStatusUpdater) setCleanStatus(repo *Repository) {
	repo.HasUncommitted = false
	repo.HasUnpushed = false
	repo.HasUntracked = false
	repo.HasError = false
	repo.IsWorktree = false
	repo.IsBare = false
}

func (r *RepositoryStatusUpdater) setErrorStatus(repo *Repository) {
	repo.HasUncommitted = false
	repo.HasUnpushed = false
	repo.HasUntracked = false
	repo.HasError = true
	repo.IsWorktree = false
	repo.IsBare = false
}
