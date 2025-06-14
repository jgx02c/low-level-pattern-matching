package main

import (
	"hash/fnv"
	"sync"
	"sync/atomic"
	"time"
)

// CacheEntry represents a cached search result
type CacheEntry struct {
	Input    string
	Results  []MatchResult
	Duration time.Duration
	Created  time.Time
	Hits     int64
}

// Cache provides ultra-fast pattern matching result caching
type Cache struct {
	entries map[uint64]*CacheEntry
	mutex   sync.RWMutex
	maxSize int
	stats   CacheStats
}

// CacheStats tracks cache performance
type CacheStats struct {
	Hits         int64
	Misses       int64
	Evictions    int64
	TotalEntries int64
}

// NewCache creates a new cache with specified maximum size
func NewCache(maxSize int) *Cache {
	return &Cache{
		entries: make(map[uint64]*CacheEntry),
		maxSize: maxSize,
	}
}

// hash generates a fast hash for input text
func (c *Cache) hash(input string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(input))
	return h.Sum64()
}

// Get retrieves cached results for input text
func (c *Cache) Get(input string) ([]MatchResult, time.Duration, bool) {
	key := c.hash(input)

	c.mutex.RLock()
	entry, exists := c.entries[key]
	c.mutex.RUnlock()

	if exists {
		atomic.AddInt64(&entry.Hits, 1)
		atomic.AddInt64(&c.stats.Hits, 1)
		return entry.Results, entry.Duration, true
	}

	atomic.AddInt64(&c.stats.Misses, 1)
	return nil, 0, false
}

// Put stores search results in cache
func (c *Cache) Put(input string, results []MatchResult, duration time.Duration) {
	key := c.hash(input)

	entry := &CacheEntry{
		Input:    input,
		Results:  results,
		Duration: duration,
		Created:  time.Now(),
		Hits:     0,
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if we need to evict entries
	if len(c.entries) >= c.maxSize {
		c.evictLRU()
	}

	c.entries[key] = entry
	atomic.AddInt64(&c.stats.TotalEntries, 1)
}

// evictLRU removes the least recently used entry
func (c *Cache) evictLRU() {
	var oldestKey uint64
	var oldestTime time.Time

	first := true
	for key, entry := range c.entries {
		if first || entry.Created.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Created
			first = false
		}
	}

	if !first {
		delete(c.entries, oldestKey)
		atomic.AddInt64(&c.stats.Evictions, 1)
	}
}

// GetStats returns cache performance statistics
func (c *Cache) GetStats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return CacheStats{
		Hits:         atomic.LoadInt64(&c.stats.Hits),
		Misses:       atomic.LoadInt64(&c.stats.Misses),
		Evictions:    atomic.LoadInt64(&c.stats.Evictions),
		TotalEntries: int64(len(c.entries)),
	}
}

// Clear removes all cached entries
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries = make(map[uint64]*CacheEntry)
	atomic.StoreInt64(&c.stats.Hits, 0)
	atomic.StoreInt64(&c.stats.Misses, 0)
	atomic.StoreInt64(&c.stats.Evictions, 0)
	atomic.StoreInt64(&c.stats.TotalEntries, 0)
}

// HitRatio returns the cache hit ratio as a percentage
func (c *Cache) HitRatio() float64 {
	hits := atomic.LoadInt64(&c.stats.Hits)
	misses := atomic.LoadInt64(&c.stats.Misses)
	total := hits + misses

	if total == 0 {
		return 0.0
	}

	return float64(hits) / float64(total) * 100.0
}
