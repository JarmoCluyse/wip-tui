package git

import (
	"fmt"
	"time"
)

// CachedChecker wraps the git checker with caching capabilities.
type CachedChecker struct {
	checker StatusChecker
	cache   *Cache
}

// NewCachedChecker creates a new cached git checker with 5 second TTL.
func NewCachedChecker() StatusChecker {
	// TODO: maybe make ttl configurable
	return &CachedChecker{
		checker: NewChecker(),
		cache:   NewCache(5 * time.Second), // 5 second cache
	}
}

// ClearCache clears all cached values.
func (c *CachedChecker) ClearCache() {
	c.cache.Clear()
}

// IsGitRepository checks if the given path contains a Git repository (cached).
func (c *CachedChecker) IsGitRepository(path string) bool {
	key := fmt.Sprintf("IsGitRepository:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(bool)
	}

	result := c.checker.IsGitRepository(path)
	c.cache.Set(key, result)
	return result
}

// IsBareRepository checks if the repository at the given path is a bare repository (cached).
func (c *CachedChecker) IsBareRepository(path string) bool {
	key := fmt.Sprintf("IsBareRepository:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(bool)
	}

	result := c.checker.IsBareRepository(path)
	c.cache.Set(key, result)
	return result
}

// IsWorktree checks if the given path is a Git worktree (cached).
func (c *CachedChecker) IsWorktree(path string) bool {
	key := fmt.Sprintf("IsWorktree:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(bool)
	}

	result := c.checker.IsWorktree(path)
	c.cache.Set(key, result)
	return result
}

// ListWorktrees returns all worktrees for the repository at the given path (cached).
func (c *CachedChecker) ListWorktrees(path string) ([]WorktreeInfo, error) {
	key := fmt.Sprintf("ListWorktrees:%s", path)
	if value, ok := c.cache.Get(key); ok {
		cachedResult := value.(cachedWorktreeResult)
		return cachedResult.Worktrees, cachedResult.Error
	}

	worktrees, err := c.checker.ListWorktrees(path)
	c.cache.Set(key, cachedWorktreeResult{Worktrees: worktrees, Error: err})
	return worktrees, err
}

// HasUncommittedChanges checks if the repository has uncommitted changes (cached).
func (c *CachedChecker) HasUncommittedChanges(path string) bool {
	key := fmt.Sprintf("HasUncommittedChanges:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(bool)
	}

	result := c.checker.HasUncommittedChanges(path)
	c.cache.Set(key, result)
	return result
}

// HasUnpushedCommits checks if the repository has unpushed commits (cached).
func (c *CachedChecker) HasUnpushedCommits(path string) bool {
	key := fmt.Sprintf("HasUnpushedCommits:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(bool)
	}

	result := c.checker.HasUnpushedCommits(path)
	c.cache.Set(key, result)
	return result
}

// HasUntrackedFiles checks if the repository has untracked files (cached).
func (c *CachedChecker) HasUntrackedFiles(path string) bool {
	key := fmt.Sprintf("HasUntrackedFiles:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(bool)
	}

	result := c.checker.HasUntrackedFiles(path)
	c.cache.Set(key, result)
	return result
}

// GetCurrentBranch returns the current branch name (cached).
func (c *CachedChecker) GetCurrentBranch(path string) string {
	key := fmt.Sprintf("GetCurrentBranch:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(string)
	}

	result := c.checker.GetCurrentBranch(path)
	c.cache.Set(key, result)
	return result
}

// CountUncommittedChanges returns the number of files with uncommitted changes (cached).
func (c *CachedChecker) CountUncommittedChanges(path string) int {
	key := fmt.Sprintf("CountUncommittedChanges:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(int)
	}

	result := c.checker.CountUncommittedChanges(path)
	c.cache.Set(key, result)
	return result
}

// CountUnpushedCommits returns the number of unpushed commits (cached).
func (c *CachedChecker) CountUnpushedCommits(path string) int {
	key := fmt.Sprintf("CountUnpushedCommits:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(int)
	}

	result := c.checker.CountUnpushedCommits(path)
	c.cache.Set(key, result)
	return result
}

// CountUntrackedFiles returns the number of untracked files (cached).
func (c *CachedChecker) CountUntrackedFiles(path string) int {
	key := fmt.Sprintf("CountUntrackedFiles:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(int)
	}

	result := c.checker.CountUntrackedFiles(path)
	c.cache.Set(key, result)
	return result
}

// cachedWorktreeResult is used to cache both worktrees and error results.
type cachedWorktreeResult struct {
	Worktrees []WorktreeInfo
	Error     error
}
