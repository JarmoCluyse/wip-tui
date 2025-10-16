// Package repomanager provides repository management with hierarchical items and worktrees.
package repomanager

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jarmocluyse/git-dash/internal/config"
)

// RepoManager manages repositories and their worktrees.
type RepoManager struct {
	configService config.ConfigService
	items         []*RepoItem
}

// NewRepoManager creates a new repository manager.
func NewRepoManager(configService config.ConfigService) *RepoManager {
	return &RepoManager{
		configService: configService,
		items:         make([]*RepoItem, 0),
	}
}

// Init initializes the repository manager by loading paths from config and building the item list.
func (rm *RepoManager) Init() error {
	config, err := rm.configService.Load()
	if err != nil {
		return err
	}

	// Clear existing items
	rm.items = make([]*RepoItem, 0, len(config.RepositoryPaths))

	// Load repositories from config paths
	for _, path := range config.RepositoryPaths {
		item := &RepoItem{
			Name:     extractNameFromPath(path),
			Path:     path,
			SubItems: make([]*SubItem, 0),
		}

		// Update status for this repository
		rm.updateRepoStatus(item)

		// Load worktrees if this is a bare repository
		if item.IsBare {
			rm.loadWorktrees(item)
		}

		rm.items = append(rm.items, item)
	}

	return nil
}

// GetItems returns all repository items.
func (rm *RepoManager) GetItems() []*RepoItem {
	items := make([]*RepoItem, len(rm.items))
	copy(items, rm.items)
	return items
}

// AddRepo adds a new repository by path.
func (rm *RepoManager) AddRepo(path string) error {
	// Check if already exists
	for _, item := range rm.items {
		if item.Path == path {
			return nil // Already exists
		}
	}

	// Create new repo item
	item := &RepoItem{
		Name:     extractNameFromPath(path),
		Path:     path,
		SubItems: make([]*SubItem, 0),
	}

	// Update status
	rm.updateRepoStatus(item)

	// Load worktrees if bare
	if item.IsBare {
		rm.loadWorktrees(item)
	}

	rm.items = append(rm.items, item)

	// Update config
	config, err := rm.configService.Load()
	if err != nil {
		return err
	}

	config.AddRepositoryPath(path)
	return rm.configService.Save(config)
}

// RemoveRepo removes a repository by path.
func (rm *RepoManager) RemoveRepo(path string) error {
	// Find and remove the item
	for i, item := range rm.items {
		if item.Path == path {
			rm.items = append(rm.items[:i], rm.items[i+1:]...)
			break
		}
	}

	// Update config
	config, err := rm.configService.Load()
	if err != nil {
		return err
	}

	config.RemoveRepositoryPathByValue(path)
	return rm.configService.Save(config)
}

// ReloadWorktrees reloads worktrees for all bare repositories.
func (rm *RepoManager) ReloadWorktrees() error {
	for _, item := range rm.items {
		if item.IsBare {
			rm.loadWorktrees(item)
		}
	}
	return nil
}

// ReloadStatus reloads status for all repositories and their worktrees.
func (rm *RepoManager) ReloadStatus() error {
	for _, item := range rm.items {
		rm.updateRepoStatus(item)

		// Update status for all worktrees
		for _, subItem := range item.SubItems {
			rm.updateSubItemStatus(subItem)
		}
	}
	return nil
}

// GetSummary calculates and returns summary data for all repositories and worktrees.
func (rm *RepoManager) GetSummary() SummaryData {
	var data SummaryData

	for _, item := range rm.items {
		if item.HasUncommitted {
			data.TotalUncommitted += item.UncommittedCount
		}
		if item.HasUnpushed {
			data.TotalUnpushed += item.UnpushedCount
		}
		if item.HasUntracked {
			data.TotalUntracked += item.UntrackedCount
		}
		if item.HasError {
			data.TotalErrors++
		}

		// Add sub-items (worktrees)
		for _, subItem := range item.SubItems {
			if subItem.HasUncommitted {
				data.TotalUncommitted += subItem.UncommittedCount
			}
			if subItem.HasUnpushed {
				data.TotalUnpushed += subItem.UnpushedCount
			}
			if subItem.HasUntracked {
				data.TotalUntracked += subItem.UntrackedCount
			}
			if subItem.HasError {
				data.TotalErrors++
			}
		}
	}

	return data
}

// updateRepoStatus updates the status of a repository item.
func (rm *RepoManager) updateRepoStatus(item *RepoItem) {
	if !rm.isGitRepository(item.Path) {
		item.HasError = true
		item.HasUncommitted = false
		item.HasUnpushed = false
		item.HasUntracked = false
		item.UncommittedCount = 0
		item.UnpushedCount = 0
		item.UntrackedCount = 0
		return
	}

	item.IsBare = rm.isBareRepository(item.Path)
	item.IsWorktree = rm.isWorktree(item.Path)
	item.HasError = false

	if item.IsBare {
		// For bare repositories, no status information is relevant
		item.HasUncommitted = false
		item.HasUnpushed = false
		item.HasUntracked = false
		item.UncommittedCount = 0
		item.UnpushedCount = 0
		item.UntrackedCount = 0
	} else {
		// For regular repositories, check normal git status
		item.HasUncommitted = rm.hasUncommittedChanges(item.Path)
		item.HasUnpushed = rm.hasUnpushedCommits(item.Path)
		item.HasUntracked = rm.hasUntrackedFiles(item.Path)
		item.UncommittedCount = rm.countUncommittedChanges(item.Path)
		item.UnpushedCount = rm.countUnpushedCommits(item.Path)
		item.UntrackedCount = rm.countUntrackedFiles(item.Path)
	}
}

// loadWorktrees loads worktrees for a bare repository.
func (rm *RepoManager) loadWorktrees(item *RepoItem) {
	if !item.IsBare {
		return
	}

	worktrees, err := rm.listWorktrees(item.Path)
	if err != nil {
		return
	}

	// Clear existing sub-items
	item.SubItems = make([]*SubItem, 0, len(worktrees))

	// Create sub-items for each worktree, excluding the main repository itself
	for _, wt := range worktrees {
		// Skip the main repository - it should not be listed as its own worktree
		if wt.Path == item.Path {
			continue
		}

		subItem := &SubItem{
			Name:       extractNameFromPath(wt.Path),
			Path:       wt.Path,
			Branch:     wt.Branch,
			ParentRepo: item,
		}

		// Update status for this worktree
		rm.updateSubItemStatus(subItem)

		item.SubItems = append(item.SubItems, subItem)
	}
}

// updateSubItemStatus updates the status of a worktree sub-item.
func (rm *RepoManager) updateSubItemStatus(subItem *SubItem) {
	if !rm.isGitRepository(subItem.Path) {
		subItem.HasError = true
		subItem.HasUncommitted = false
		subItem.HasUnpushed = false
		subItem.HasUntracked = false
		subItem.UncommittedCount = 0
		subItem.UnpushedCount = 0
		subItem.UntrackedCount = 0
		return
	}

	subItem.HasError = false
	subItem.HasUncommitted = rm.hasUncommittedChanges(subItem.Path)
	subItem.HasUnpushed = rm.hasUnpushedCommits(subItem.Path)
	subItem.HasUntracked = rm.hasUntrackedFiles(subItem.Path)
	subItem.UncommittedCount = rm.countUncommittedChanges(subItem.Path)
	subItem.UnpushedCount = rm.countUnpushedCommits(subItem.Path)
	subItem.UntrackedCount = rm.countUntrackedFiles(subItem.Path)
}

// Git command methods

// isGitRepository checks if the given path contains a Git repository.
func (rm *RepoManager) isGitRepository(path string) bool {
	if rm.hasGitDirectory(path) {
		return true
	}
	if rm.isBareRepository(path) {
		return true
	}
	return rm.canRunGitCommand(path)
}

// isBareRepository checks if the repository at the given path is a bare repository.
func (rm *RepoManager) isBareRepository(path string) bool {
	output, err := rm.runGitCommand(path, "rev-parse", "--is-bare-repository")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

// hasGitDirectory checks if the path contains a .git directory or file.
func (rm *RepoManager) hasGitDirectory(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	return info.IsDir() || rm.isWorktreeGitFile(gitPath)
}

// isWorktreeGitFile checks if the .git file is a worktree reference.
func (rm *RepoManager) isWorktreeGitFile(gitPath string) bool {
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

// listWorktrees returns all worktrees for the repository at the given path.
func (rm *RepoManager) listWorktrees(path string) ([]WorktreeInfo, error) {
	output, err := rm.runGitCommand(path, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	return rm.parseWorktreeList(string(output))
}

// parseWorktreeList parses the output of git worktree list --porcelain.
func (rm *RepoManager) parseWorktreeList(output string) ([]WorktreeInfo, error) {
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

// isWorktree checks if the given path is a Git worktree.
func (rm *RepoManager) isWorktree(path string) bool {
	output, err := rm.runGitCommand(path, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false
	}

	isInsideWorkTree := strings.TrimSpace(string(output)) == "true"
	if !isInsideWorkTree {
		return false
	}

	output, err = rm.runGitCommand(path, "rev-parse", "--git-common-dir")
	if err != nil {
		return false
	}

	commonDir := strings.TrimSpace(string(output))
	output, err = rm.runGitCommand(path, "rev-parse", "--git-dir")
	if err != nil {
		return false
	}

	gitDir := strings.TrimSpace(string(output))
	return commonDir != gitDir
}

// canRunGitCommand checks if git commands can be run in the given path.
func (rm *RepoManager) canRunGitCommand(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	return cmd.Run() == nil
}

// hasUncommittedChanges checks if the repository has uncommitted changes.
func (rm *RepoManager) hasUncommittedChanges(path string) bool {
	output, err := rm.runGitCommand(path, "status", "--porcelain")
	if err != nil {
		return false
	}
	return rm.hasOutput(output)
}

// hasUnpushedCommits checks if the repository has unpushed commits.
func (rm *RepoManager) hasUnpushedCommits(path string) bool {
	if rm.branchIsAhead(path) {
		return true
	}
	return rm.hasCommitsAheadOfUpstream(path)
}

// branchIsAhead checks if the current branch is ahead of its upstream.
func (rm *RepoManager) branchIsAhead(path string) bool {
	output, err := rm.runGitCommand(path, "status", "--porcelain=v1", "--branch")
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
func (rm *RepoManager) hasCommitsAheadOfUpstream(path string) bool {
	output, err := rm.runGitCommand(path, "log", "--oneline", "@{u}..")
	if err != nil {
		return false
	}
	return rm.hasOutput(output)
}

// hasUntrackedFiles checks if the repository has untracked files.
func (rm *RepoManager) hasUntrackedFiles(path string) bool {
	output, err := rm.runGitCommand(path, "status", "--porcelain")
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

// countUncommittedChanges returns the number of files with uncommitted changes (tracked files only).
func (rm *RepoManager) countUncommittedChanges(path string) int {
	output, err := rm.runGitCommand(path, "status", "--porcelain")
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

// countUnpushedCommits returns the number of unpushed commits.
func (rm *RepoManager) countUnpushedCommits(path string) int {
	// First try counting with log command
	output, err := rm.runGitCommand(path, "log", "--oneline", "@{u}..")
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}

// countUntrackedFiles returns the number of untracked files.
func (rm *RepoManager) countUntrackedFiles(path string) int {
	output, err := rm.runGitCommand(path, "status", "--porcelain")
	if err != nil {
		return 0
	}

	count := 0
	lines := strings.SplitSeq(strings.TrimSpace(string(output)), "\n")
	for line := range lines {
		if strings.HasPrefix(line, "??") {
			count++
		}
	}
	return count
}

// runGitCommand executes a git command in the specified directory.
func (rm *RepoManager) runGitCommand(path string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	return cmd.Output()
}

// hasOutput checks if the command output is non-empty.
func (rm *RepoManager) hasOutput(output []byte) bool {
	return len(strings.TrimSpace(string(output))) > 0
}

// extractNameFromPath extracts a name from a file path.
func extractNameFromPath(path string) string {
	if path == "" {
		return ""
	}

	// Remove trailing slashes
	for len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	return filepath.Base(path)
}
