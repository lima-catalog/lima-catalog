package cache

import (
	"sync"
	"time"
)

// Entry represents a cached item with expiration
type Entry struct {
	Value      interface{}
	ExpiresAt  time.Time
	AccessedAt time.Time
}

// Cache is a thread-safe in-memory cache with TTL support
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*Entry
	ttl     time.Duration
}

// New creates a new cache with the specified TTL
func New(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]*Entry),
		ttl:     ttl,
	}
}

// Get retrieves a value from the cache
// Returns (value, true) if found and not expired, (nil, false) otherwise
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		c.Delete(key)
		return nil, false
	}

	// Update access time
	c.mu.Lock()
	entry.AccessedAt = time.Now()
	c.mu.Unlock()

	return entry.Value, true
}

// Set stores a value in the cache with the default TTL
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL stores a value in the cache with a custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	now := time.Now()
	entry := &Entry{
		Value:      value,
		ExpiresAt:  now.Add(ttl),
		AccessedAt: now,
	}

	c.mu.Lock()
	c.entries[key] = entry
	c.mu.Unlock()
}

// Delete removes a key from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	c.entries = make(map[string]*Entry)
	c.mu.Unlock()
}

// Size returns the number of entries in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Cleanup removes expired entries from the cache
func (c *Cache) Cleanup() int {
	now := time.Now()
	removed := 0

	c.mu.Lock()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
			removed++
		}
	}
	c.mu.Unlock()

	return removed
}

// StartCleanupTimer starts a background goroutine that periodically cleans up expired entries
func (c *Cache) StartCleanupTimer(interval time.Duration) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			c.Cleanup()
		}
	}()
	return ticker
}
