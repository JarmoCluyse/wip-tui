package git

import (
	"fmt"
	"time"
)

// CachedChecker wraps the git checker with caching capabilities
type CachedChecker struct {
	checker StatusChecker
	cache   *Cache
}

// NewCachedChecker creates a new cached git checker with 5 second TTL
func NewCachedChecker() StatusChecker {
	return &CachedChecker{
		checker: NewChecker(),
		cache:   NewCache(5 * time.Second), // 5 second cache
	}
}

// NewCachedCheckerWithTTL creates a new cached git checker with custom TTL
func NewCachedCheckerWithTTL(ttl time.Duration) StatusChecker {
	return &CachedChecker{
		checker: NewChecker(),
		cache:   NewCache(ttl),
	}
}

// ClearCache clears all cached values
func (c *CachedChecker) ClearCache() {
	c.cache.Clear()
}

func (c *CachedChecker) IsGitRepository(path string) bool {
	key := fmt.Sprintf("IsGitRepository:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(bool)
	}

	result := c.checker.IsGitRepository(path)
	c.cache.Set(key, result)
	return result
}

func (c *CachedChecker) IsBareRepository(path string) bool {
	key := fmt.Sprintf("IsBareRepository:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(bool)
	}

	result := c.checker.IsBareRepository(path)
	c.cache.Set(key, result)
	return result
}

func (c *CachedChecker) IsWorktree(path string) bool {
	key := fmt.Sprintf("IsWorktree:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(bool)
	}

	result := c.checker.IsWorktree(path)
	c.cache.Set(key, result)
	return result
}

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

func (c *CachedChecker) HasUncommittedChanges(path string) bool {
	key := fmt.Sprintf("HasUncommittedChanges:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(bool)
	}

	result := c.checker.HasUncommittedChanges(path)
	c.cache.Set(key, result)
	return result
}

func (c *CachedChecker) HasUnpushedCommits(path string) bool {
	key := fmt.Sprintf("HasUnpushedCommits:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(bool)
	}

	result := c.checker.HasUnpushedCommits(path)
	c.cache.Set(key, result)
	return result
}

func (c *CachedChecker) HasUntrackedFiles(path string) bool {
	key := fmt.Sprintf("HasUntrackedFiles:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(bool)
	}

	result := c.checker.HasUntrackedFiles(path)
	c.cache.Set(key, result)
	return result
}

func (c *CachedChecker) GetCurrentBranch(path string) string {
	key := fmt.Sprintf("GetCurrentBranch:%s", path)
	if value, ok := c.cache.Get(key); ok {
		return value.(string)
	}

	result := c.checker.GetCurrentBranch(path)
	c.cache.Set(key, result)
	return result
}

// cachedWorktreeResult is used to cache both worktrees and error results
type cachedWorktreeResult struct {
	Worktrees []WorktreeInfo
	Error     error
}
