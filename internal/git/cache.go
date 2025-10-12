package git

import (
	"sync"
	"time"
)

// CacheEntry represents a cached git operation result.
type CacheEntry struct {
	Value     any
	ExpiresAt time.Time
}

// Cache provides caching for git operations to improve performance.
type Cache struct {
	entries map[string]CacheEntry
	mutex   sync.RWMutex
	ttl     time.Duration
}

// NewCache creates a new cache with specified TTL.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]CacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a value from cache if it exists and hasn't expired.
func (c *Cache) Get(key string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Value, true
}

// Set stores a value in cache with expiration.
func (c *Cache) Set(key string, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries[key] = CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Clear removes all entries from cache.
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entries = make(map[string]CacheEntry)
}

// CleanExpired removes expired entries from cache.
func (c *Cache) CleanExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}
