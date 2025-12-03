package ai

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Cache defines the interface for caching AI responses
type Cache interface {
	// Get retrieves a cached response
	Get(key string) (*Response, bool)

	// Set stores a response in the cache
	Set(key string, response *Response, ttl time.Duration) error

	// Delete removes a response from the cache
	Delete(key string) error

	// Clear removes all cached responses
	Clear() error

	// Stats returns cache statistics
	Stats() CacheStats

	// Size returns the current cache size in bytes
	Size() int64

	// Count returns the number of cached items
	Count() int
}

// CacheStats contains cache statistics
type CacheStats struct {
	Hits            int64
	Misses          int64
	Evictions       int64
	Size            int64
	Count           int
	HitRate         float64
	AverageItemSize int64
}

// MemoryCache implements an in-memory LRU cache with TTL
type MemoryCache struct {
	mu          sync.RWMutex
	maxSize     int64
	currentSize int64
	items       map[string]*cacheEntry
	lru         *list.List
	stats       CacheStats
}

// cacheEntry represents a cached item
type cacheEntry struct {
	key        string
	response   *Response
	size       int64
	expiresAt  time.Time
	lruElement *list.Element
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(maxSize int64) *MemoryCache {
	return &MemoryCache{
		maxSize: maxSize,
		items:   make(map[string]*cacheEntry),
		lru:     list.New(),
	}
}

// Get retrieves a cached response
func (c *MemoryCache) Get(key string) (*Response, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.items[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		c.deleteEntry(entry)
		c.stats.Misses++
		return nil, false
	}

	// Move to front (most recently used)
	c.lru.MoveToFront(entry.lruElement)
	c.stats.Hits++

	// Mark response as cached
	responseCopy := *entry.response
	responseCopy.Cached = true

	return &responseCopy, true
}

// Set stores a response in the cache
func (c *MemoryCache) Set(key string, response *Response, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate size
	size := c.calculateSize(response)

	// Check if item already exists
	if existing, exists := c.items[key]; exists {
		// Update existing entry
		c.currentSize -= existing.size
		c.deleteEntry(existing)
	}

	// Evict items if necessary to make space
	for c.currentSize+size > c.maxSize && c.lru.Len() > 0 {
		c.evictOldest()
	}

	// If item is still too large, don't cache it
	if size > c.maxSize {
		return &ErrCacheFailed{
			Operation: "set",
			Reason:    fmt.Sprintf("item size %d exceeds max cache size %d", size, c.maxSize),
		}
	}

	// Create new entry
	entry := &cacheEntry{
		key:       key,
		response:  response,
		size:      size,
		expiresAt: time.Now().Add(ttl),
	}

	// Add to LRU list
	entry.lruElement = c.lru.PushFront(entry)
	c.items[key] = entry
	c.currentSize += size
	c.stats.Count = len(c.items)
	c.stats.Size = c.currentSize

	return nil
}

// Delete removes a response from the cache
func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.items[key]
	if !exists {
		return nil
	}

	c.deleteEntry(entry)
	c.stats.Count = len(c.items)
	c.stats.Size = c.currentSize

	return nil
}

// Clear removes all cached responses
func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheEntry)
	c.lru = list.New()
	c.currentSize = 0
	c.stats.Count = 0
	c.stats.Size = 0

	return nil
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Count = len(c.items)
	stats.Size = c.currentSize

	// Calculate hit rate
	totalRequests := stats.Hits + stats.Misses
	if totalRequests > 0 {
		stats.HitRate = float64(stats.Hits) / float64(totalRequests) * 100
	}

	// Calculate average item size
	if stats.Count > 0 {
		stats.AverageItemSize = stats.Size / int64(stats.Count)
	}

	return stats
}

// Size returns the current cache size in bytes
func (c *MemoryCache) Size() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentSize
}

// Count returns the number of cached items
func (c *MemoryCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// CleanupExpired removes expired entries
func (c *MemoryCache) CleanupExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0

	for key, entry := range c.items {
		if now.After(entry.expiresAt) {
			c.deleteEntry(entry)
			delete(c.items, key)
			removed++
		}
	}

	c.stats.Count = len(c.items)
	c.stats.Size = c.currentSize

	return removed
}

// evictOldest removes the least recently used item
func (c *MemoryCache) evictOldest() {
	oldest := c.lru.Back()
	if oldest == nil {
		return
	}

	entry := oldest.Value.(*cacheEntry)
	c.deleteEntry(entry)
	c.stats.Evictions++
}

// deleteEntry removes an entry from the cache (must be called with lock held)
func (c *MemoryCache) deleteEntry(entry *cacheEntry) {
	if entry.lruElement != nil {
		c.lru.Remove(entry.lruElement)
	}
	delete(c.items, entry.key)
	c.currentSize -= entry.size
}

// calculateSize estimates the size of a response in bytes
func (c *MemoryCache) calculateSize(response *Response) int64 {
	// Estimate based on string lengths and metadata
	size := int64(len(response.ID))
	size += int64(len(response.Content))
	size += int64(len(response.Provider))

	// Add metadata size
	for k, v := range response.Metadata {
		size += int64(len(k) + len(v))
	}

	// Add structured data size (approximate)
	if response.StructuredData != nil {
		if data, err := json.Marshal(response.StructuredData); err == nil {
			size += int64(len(data))
		}
	}

	// Add fixed overhead for other fields
	size += 128 // overhead for struct fields, pointers, etc.

	return size
}

// GenerateCacheKey generates a cache key from a request
func GenerateCacheKey(req *Request) string {
	// Create a deterministic key from request fields
	keyData := struct {
		Query       string
		Type        AnalysisType
		MaxTokens   int
		Temperature float64
		Context     *AnalysisContext
	}{
		Query:       req.Query,
		Type:        req.Type,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Context:     req.Context,
	}

	// Serialize to JSON for consistent hashing
	data, err := json.Marshal(keyData)
	if err != nil {
		// Fallback to simpler key
		return fmt.Sprintf("%s:%s", req.Type, req.Query)
	}

	// Generate SHA256 hash
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// CachedProvider wraps a provider with caching
type CachedProvider struct {
	provider Provider
	cache    Cache
	ttl      time.Duration
}

// NewCachedProvider creates a new cached provider
func NewCachedProvider(provider Provider, cache Cache, ttl time.Duration) *CachedProvider {
	return &CachedProvider{
		provider: provider,
		cache:    cache,
		ttl:      ttl,
	}
}

// Name returns the provider name
func (p *CachedProvider) Name() string {
	return p.provider.Name() + "-cached"
}

// Analyze sends an analysis request with caching
func (p *CachedProvider) Analyze(ctx context.Context, req *Request) (*Response, error) {
	// Check if caching is enabled
	if req.Options.UseCache {
		key := GenerateCacheKey(req)
		if cached, found := p.cache.Get(key); found {
			return cached, nil
		}
	}

	// Call underlying provider
	resp, err := p.provider.Analyze(ctx, req)
	if err != nil {
		return nil, err
	}

	// Cache the response
	if req.Options.UseCache && resp != nil {
		ttl := p.ttl
		if req.Options.CacheTTL > 0 {
			ttl = req.Options.CacheTTL
		}
		_ = p.cache.Set(GenerateCacheKey(req), resp, ttl)
	}

	return resp, nil
}

// AnalyzeStream sends a streaming request (no caching for streams)
func (p *CachedProvider) AnalyzeStream(ctx context.Context, req *Request) (<-chan StreamChunk, error) {
	return p.provider.AnalyzeStream(ctx, req)
}

// Validate validates the provider
func (p *CachedProvider) Validate(ctx context.Context) error {
	return p.provider.Validate(ctx)
}

// GetMetrics returns provider metrics
func (p *CachedProvider) GetMetrics() *UsageMetrics {
	return p.provider.GetMetrics()
}

// Close closes the provider
func (p *CachedProvider) Close() error {
	return p.provider.Close()
}

// GetCache returns the underlying cache
func (p *CachedProvider) GetCache() Cache {
	return p.cache
}

// StartCleanupTimer starts a background goroutine to clean up expired entries
func StartCleanupTimer(cache *MemoryCache, interval time.Duration) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			cache.CleanupExpired()
		}
	}()
	return ticker
}
