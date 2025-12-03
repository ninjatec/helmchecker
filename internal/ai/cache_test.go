package ai

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache(1024 * 1024) // 1MB

	response := &Response{
		ID:       "test-1",
		Content:  "This is a test response",
		Provider: "test-provider",
		TokensUsed: TokenUsage{
			TotalTokens: 100,
		},
	}

	// Set the response
	err := cache.Set("key1", response, 1*time.Hour)
	require.NoError(t, err)

	// Get the response
	cached, found := cache.Get("key1")
	require.True(t, found)
	assert.Equal(t, response.ID, cached.ID)
	assert.Equal(t, response.Content, cached.Content)
	assert.True(t, cached.Cached)

	// Verify stats
	stats := cache.Stats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
	assert.Equal(t, 1, stats.Count)
}

func TestMemoryCache_Miss(t *testing.T) {
	cache := NewMemoryCache(1024 * 1024)

	// Try to get non-existent key
	_, found := cache.Get("nonexistent")
	assert.False(t, found)

	// Verify stats
	stats := cache.Stats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
}

func TestMemoryCache_TTL(t *testing.T) {
	cache := NewMemoryCache(1024 * 1024)

	response := &Response{
		ID:      "test-ttl",
		Content: "TTL test",
	}

	// Set with short TTL
	err := cache.Set("ttl-key", response, 50*time.Millisecond)
	require.NoError(t, err)

	// Should be available immediately
	_, found := cache.Get("ttl-key")
	assert.True(t, found)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, found = cache.Get("ttl-key")
	assert.False(t, found)
}

func TestMemoryCache_LRUEviction(t *testing.T) {
	cache := NewMemoryCache(500) // Small cache

	// Add items
	for i := 0; i < 10; i++ {
		response := &Response{
			ID:      string(rune('a' + i)),
			Content: "Test content for item",
		}
		err := cache.Set(string(rune('a'+i)), response, 1*time.Hour)
		require.NoError(t, err)
	}

	// Cache should have evicted some items
	assert.Less(t, cache.Count(), 10)
	stats := cache.Stats()
	assert.Greater(t, stats.Evictions, int64(0))
}

func TestMemoryCache_Update(t *testing.T) {
	cache := NewMemoryCache(1024 * 1024)

	response1 := &Response{
		ID:      "test-1",
		Content: "Original content",
	}

	response2 := &Response{
		ID:      "test-1",
		Content: "Updated content",
	}

	// Set initial value
	err := cache.Set("key1", response1, 1*time.Hour)
	require.NoError(t, err)

	// Update with new value
	err = cache.Set("key1", response2, 1*time.Hour)
	require.NoError(t, err)

	// Verify updated value
	cached, found := cache.Get("key1")
	require.True(t, found)
	assert.Equal(t, "Updated content", cached.Content)

	// Should still have only 1 item
	assert.Equal(t, 1, cache.Count())
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache(1024 * 1024)

	response := &Response{
		ID:      "test-delete",
		Content: "To be deleted",
	}

	// Set and verify
	err := cache.Set("key1", response, 1*time.Hour)
	require.NoError(t, err)
	_, found := cache.Get("key1")
	assert.True(t, found)

	// Delete
	err = cache.Delete("key1")
	require.NoError(t, err)

	// Verify deleted
	_, found = cache.Get("key1")
	assert.False(t, found)
	assert.Equal(t, 0, cache.Count())
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache(1024 * 1024)

	// Add multiple items
	for i := 0; i < 5; i++ {
		response := &Response{
			ID:      string(rune('a' + i)),
			Content: "Test content",
		}
		err := cache.Set(string(rune('a'+i)), response, 1*time.Hour)
		require.NoError(t, err)
	}

	assert.Equal(t, 5, cache.Count())

	// Clear all
	err := cache.Clear()
	require.NoError(t, err)

	// Verify empty
	assert.Equal(t, 0, cache.Count())
	assert.Equal(t, int64(0), cache.Size())
}

func TestMemoryCache_CleanupExpired(t *testing.T) {
	cache := NewMemoryCache(1024 * 1024)

	// Add items with different TTLs
	response1 := &Response{ID: "expire-soon", Content: "Expires soon"}
	response2 := &Response{ID: "expire-later", Content: "Expires later"}

	err := cache.Set("key1", response1, 50*time.Millisecond)
	require.NoError(t, err)
	err = cache.Set("key2", response2, 1*time.Hour)
	require.NoError(t, err)

	// Wait for first to expire
	time.Sleep(100 * time.Millisecond)

	// Cleanup
	removed := cache.CleanupExpired()
	assert.Equal(t, 1, removed)
	assert.Equal(t, 1, cache.Count())

	// Verify the right one remains
	_, found := cache.Get("key2")
	assert.True(t, found)
}

func TestMemoryCache_Stats(t *testing.T) {
	cache := NewMemoryCache(1024 * 1024)

	response := &Response{
		ID:      "test-stats",
		Content: "Statistics test",
	}

	// Set item
	err := cache.Set("key1", response, 1*time.Hour)
	require.NoError(t, err)

	// Generate hits and misses
	cache.Get("key1")
	cache.Get("key1")
	cache.Get("nonexistent")

	stats := cache.Stats()
	assert.Equal(t, int64(2), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, 1, stats.Count)
	assert.Greater(t, stats.Size, int64(0))
	assert.Greater(t, stats.HitRate, 0.0)
	assert.Greater(t, stats.AverageItemSize, int64(0))
}

func TestMemoryCache_Concurrent(t *testing.T) {
	cache := NewMemoryCache(1024 * 1024)

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			response := &Response{
				ID:      string(rune('a' + id)),
				Content: "Concurrent test",
			}
			cache.Set(string(rune('a'+id)), response, 1*time.Hour)
			cache.Get(string(rune('a' + id)))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify cache is in valid state
	assert.Equal(t, 10, cache.Count())
	stats := cache.Stats()
	assert.Greater(t, stats.Hits, int64(0))
}

func TestGenerateCacheKey(t *testing.T) {
	req1 := &Request{
		ID:          "req-1",
		Query:       "test query",
		Type:        AnalysisTypeCompatibility,
		MaxTokens:   1000,
		Temperature: 0.5,
	}

	req2 := &Request{
		ID:          "req-2",
		Query:       "test query",
		Type:        AnalysisTypeCompatibility,
		MaxTokens:   1000,
		Temperature: 0.5,
	}

	req3 := &Request{
		ID:          "req-3",
		Query:       "different query",
		Type:        AnalysisTypeCompatibility,
		MaxTokens:   1000,
		Temperature: 0.5,
	}

	// Same content should generate same key
	key1 := GenerateCacheKey(req1)
	key2 := GenerateCacheKey(req2)
	assert.Equal(t, key1, key2)

	// Different content should generate different key
	key3 := GenerateCacheKey(req3)
	assert.NotEqual(t, key1, key3)
}

func TestCachedProvider(t *testing.T) {
	cache := NewMemoryCache(1024 * 1024)
	mockProvider := &MockProvider{
		name: "mock-provider",
		analyzeFunc: func(ctx context.Context, req *Request) (*Response, error) {
			return &Response{
				ID:      req.ID,
				Content: "Mock response",
			}, nil
		},
	}

	cachedProvider := NewCachedProvider(mockProvider, cache, 1*time.Hour)

	req := &Request{
		ID:    "test-req",
		Query: "test query",
		Type:  AnalysisTypeGeneral,
		Options: RequestOptions{
			UseCache: true,
		},
	}

	// First call should hit the provider
	ctx := context.Background()
	resp1, err := cachedProvider.Analyze(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp1)
	assert.False(t, resp1.Cached)
	assert.Equal(t, 1, mockProvider.analyzeCalls)

	// Second call should hit cache
	resp2, err := cachedProvider.Analyze(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp2)
	assert.True(t, resp2.Cached)
	assert.Equal(t, 1, mockProvider.analyzeCalls) // Should not increment
}

func TestCachedProvider_CachingDisabled(t *testing.T) {
	cache := NewMemoryCache(1024 * 1024)
	mockProvider := &MockProvider{
		name: "mock-provider",
		analyzeFunc: func(ctx context.Context, req *Request) (*Response, error) {
			return &Response{
				ID:      req.ID,
				Content: "Mock response",
			}, nil
		},
	}

	cachedProvider := NewCachedProvider(mockProvider, cache, 1*time.Hour)

	req := &Request{
		ID:    "test-req",
		Query: "test query",
		Type:  AnalysisTypeGeneral,
		Options: RequestOptions{
			UseCache: false, // Caching disabled
		},
	}

	ctx := context.Background()

	// First call
	resp1, err := cachedProvider.Analyze(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp1)
	assert.Equal(t, 1, mockProvider.analyzeCalls)

	// Second call should still hit provider (no caching)
	resp2, err := cachedProvider.Analyze(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp2)
	assert.Equal(t, 2, mockProvider.analyzeCalls)
}

// MockProvider for testing
type MockProvider struct {
	name          string
	analyzeCalls  int
	analyzeFunc   func(ctx context.Context, req *Request) (*Response, error)
	streamFunc    func(ctx context.Context, req *Request) (<-chan StreamChunk, error)
	validateFunc  func(ctx context.Context) error
	metrics       *UsageMetrics
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Analyze(ctx context.Context, req *Request) (*Response, error) {
	m.analyzeCalls++
	if m.analyzeFunc != nil {
		return m.analyzeFunc(ctx, req)
	}
	return &Response{ID: req.ID, Content: "mock"}, nil
}

func (m *MockProvider) AnalyzeStream(ctx context.Context, req *Request) (<-chan StreamChunk, error) {
	if m.streamFunc != nil {
		return m.streamFunc(ctx, req)
	}
	ch := make(chan StreamChunk)
	close(ch)
	return ch, nil
}

func (m *MockProvider) Validate(ctx context.Context) error {
	if m.validateFunc != nil {
		return m.validateFunc(ctx)
	}
	return nil
}

func (m *MockProvider) GetMetrics() *UsageMetrics {
	if m.metrics != nil {
		return m.metrics
	}
	return NewUsageMetrics()
}

func (m *MockProvider) Close() error {
	return nil
}

func TestStartCleanupTimer(t *testing.T) {
	cache := NewMemoryCache(1024 * 1024)

	// Add expired item
	response := &Response{
		ID:      "expire-test",
		Content: "Will expire",
	}
	err := cache.Set("key1", response, 50*time.Millisecond)
	require.NoError(t, err)

	// Start cleanup timer
	ticker := StartCleanupTimer(cache, 100*time.Millisecond)
	defer ticker.Stop()

	// Wait for cleanup to run
	time.Sleep(200 * time.Millisecond)

	// Item should be cleaned up
	assert.Equal(t, 0, cache.Count())
}

func BenchmarkMemoryCache_Set(b *testing.B) {
	cache := NewMemoryCache(1024 * 1024 * 100) // 100MB
	response := &Response{
		ID:      "bench-test",
		Content: "Benchmark test content",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(string(rune(i)), response, 1*time.Hour)
	}
}

func BenchmarkMemoryCache_Get(b *testing.B) {
	cache := NewMemoryCache(1024 * 1024 * 100)
	response := &Response{
		ID:      "bench-test",
		Content: "Benchmark test content",
	}
	cache.Set("key1", response, 1*time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key1")
	}
}

func BenchmarkGenerateCacheKey(b *testing.B) {
	req := &Request{
		ID:          "bench-req",
		Query:       "benchmark query",
		Type:        AnalysisTypeCompatibility,
		MaxTokens:   1000,
		Temperature: 0.5,
		Context: &AnalysisContext{
			RepositoryInfo: &RepositoryInfo{
				Owner: "test",
				Name:  "repo",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateCacheKey(req)
	}
}
