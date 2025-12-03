package ai

import (
	"sync"
	"time"
)

// UsageMetrics tracks AI provider usage statistics
type UsageMetrics struct {
	mu sync.RWMutex

	// TotalRequests counts all requests made
	TotalRequests int64

	// SuccessfulRequests counts successful requests
	SuccessfulRequests int64

	// FailedRequests counts failed requests
	FailedRequests int64

	// CachedResponses counts responses served from cache
	CachedResponses int64

	// TotalTokensUsed tracks total token consumption
	TotalTokensUsed int64

	// TotalCost tracks total cost in USD
	TotalCost float64

	// AverageLatency tracks average response time
	AverageLatency time.Duration

	// ProviderMetrics tracks per-provider metrics
	ProviderMetrics map[string]*ProviderMetrics

	// RequestsByType tracks requests by analysis type
	RequestsByType map[AnalysisType]int64

	// ErrorsByType tracks errors by type
	ErrorsByType map[string]int64

	// StartTime records when metrics started being collected
	StartTime time.Time

	// LastRequestTime records the last request time
	LastRequestTime time.Time
}

// ProviderMetrics tracks metrics for a specific provider
type ProviderMetrics struct {
	Name               string
	Requests           int64
	SuccessfulRequests int64
	FailedRequests     int64
	TokensUsed         int64
	TotalCost          float64
	AverageLatency     time.Duration
	LastUsed           time.Time
}

// NewUsageMetrics creates a new UsageMetrics instance
func NewUsageMetrics() *UsageMetrics {
	return &UsageMetrics{
		ProviderMetrics: make(map[string]*ProviderMetrics),
		RequestsByType:  make(map[AnalysisType]int64),
		ErrorsByType:    make(map[string]int64),
		StartTime:       time.Now(),
	}
}

// RecordRequest records a successful request
func (m *UsageMetrics) RecordRequest(provider string, tokens TokenUsage) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	m.SuccessfulRequests++
	m.TotalTokensUsed += int64(tokens.TotalTokens)
	m.TotalCost += tokens.EstimatedCost
	m.LastRequestTime = time.Now()

	// Update provider-specific metrics
	pm := m.getOrCreateProviderMetrics(provider)
	pm.Requests++
	pm.SuccessfulRequests++
	pm.TokensUsed += int64(tokens.TotalTokens)
	pm.TotalCost += tokens.EstimatedCost
	pm.LastUsed = time.Now()
}

// RecordFailure records a failed request
func (m *UsageMetrics) RecordFailure(provider string, errType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	m.FailedRequests++
	m.LastRequestTime = time.Now()

	// Update provider-specific metrics
	pm := m.getOrCreateProviderMetrics(provider)
	pm.Requests++
	pm.FailedRequests++
	pm.LastUsed = time.Now()

	// Track error type
	m.ErrorsByType[errType]++
}

// RecordCacheHit records a cache hit
func (m *UsageMetrics) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CachedResponses++
}

// RecordLatency records request latency
func (m *UsageMetrics) RecordLatency(provider string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update overall average latency
	if m.SuccessfulRequests > 0 {
		m.AverageLatency = time.Duration(
			(int64(m.AverageLatency)*m.SuccessfulRequests + int64(duration)) /
				(m.SuccessfulRequests + 1),
		)
	} else {
		m.AverageLatency = duration
	}

	// Update provider-specific latency
	pm := m.getOrCreateProviderMetrics(provider)
	if pm.SuccessfulRequests > 0 {
		pm.AverageLatency = time.Duration(
			(int64(pm.AverageLatency)*pm.SuccessfulRequests + int64(duration)) /
				(pm.SuccessfulRequests + 1),
		)
	} else {
		pm.AverageLatency = duration
	}
}

// RecordRequestType records a request by analysis type
func (m *UsageMetrics) RecordRequestType(analysisType AnalysisType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RequestsByType[analysisType]++
}

// GetSuccessRate returns the success rate as a percentage
func (m *UsageMetrics) GetSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.TotalRequests == 0 {
		return 0.0
	}
	return float64(m.SuccessfulRequests) / float64(m.TotalRequests) * 100
}

// GetCacheHitRate returns the cache hit rate as a percentage
func (m *UsageMetrics) GetCacheHitRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.TotalRequests == 0 {
		return 0.0
	}
	return float64(m.CachedResponses) / float64(m.TotalRequests) * 100
}

// GetProviderMetrics returns metrics for a specific provider
func (m *UsageMetrics) GetProviderMetrics(provider string) *ProviderMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if pm, ok := m.ProviderMetrics[provider]; ok {
		// Return a copy to prevent external modification
		pmCopy := *pm
		return &pmCopy
	}
	return nil
}

// GetAllProviderMetrics returns all provider metrics
func (m *UsageMetrics) GetAllProviderMetrics() map[string]*ProviderMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*ProviderMetrics, len(m.ProviderMetrics))
	for name, pm := range m.ProviderMetrics {
		pmCopy := *pm
		result[name] = &pmCopy
	}
	return result
}

// Merge combines metrics from another UsageMetrics instance
func (m *UsageMetrics) Merge(other *UsageMetrics) {
	if other == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	other.mu.RLock()
	defer other.mu.RUnlock()

	m.TotalRequests += other.TotalRequests
	m.SuccessfulRequests += other.SuccessfulRequests
	m.FailedRequests += other.FailedRequests
	m.CachedResponses += other.CachedResponses
	m.TotalTokensUsed += other.TotalTokensUsed
	m.TotalCost += other.TotalCost

	// Merge provider metrics
	for name, otherPM := range other.ProviderMetrics {
		pm := m.getOrCreateProviderMetrics(name)
		pm.Requests += otherPM.Requests
		pm.SuccessfulRequests += otherPM.SuccessfulRequests
		pm.FailedRequests += otherPM.FailedRequests
		pm.TokensUsed += otherPM.TokensUsed
		pm.TotalCost += otherPM.TotalCost
	}

	// Merge request types
	for analysisType, count := range other.RequestsByType {
		m.RequestsByType[analysisType] += count
	}

	// Merge error types
	for errType, count := range other.ErrorsByType {
		m.ErrorsByType[errType] += count
	}
}

// Reset resets all metrics
func (m *UsageMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests = 0
	m.SuccessfulRequests = 0
	m.FailedRequests = 0
	m.CachedResponses = 0
	m.TotalTokensUsed = 0
	m.TotalCost = 0
	m.AverageLatency = 0
	m.ProviderMetrics = make(map[string]*ProviderMetrics)
	m.RequestsByType = make(map[AnalysisType]int64)
	m.ErrorsByType = make(map[string]int64)
	m.StartTime = time.Now()
	m.LastRequestTime = time.Time{}
}

// getOrCreateProviderMetrics gets or creates provider metrics (must be called with lock held)
func (m *UsageMetrics) getOrCreateProviderMetrics(provider string) *ProviderMetrics {
	if pm, ok := m.ProviderMetrics[provider]; ok {
		return pm
	}
	pm := &ProviderMetrics{
		Name: provider,
	}
	m.ProviderMetrics[provider] = pm
	return pm
}

// Snapshot returns a snapshot of current metrics
type MetricsSnapshot struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	CachedResponses    int64
	TotalTokensUsed    int64
	TotalCost          float64
	AverageLatency     time.Duration
	SuccessRate        float64
	CacheHitRate       float64
	ProviderMetrics    map[string]*ProviderMetrics
	RequestsByType     map[AnalysisType]int64
	ErrorsByType       map[string]int64
	Uptime             time.Duration
	LastRequestTime    time.Time
}

// Snapshot creates a snapshot of current metrics
func (m *UsageMetrics) Snapshot() *MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &MetricsSnapshot{
		TotalRequests:      m.TotalRequests,
		SuccessfulRequests: m.SuccessfulRequests,
		FailedRequests:     m.FailedRequests,
		CachedResponses:    m.CachedResponses,
		TotalTokensUsed:    m.TotalTokensUsed,
		TotalCost:          m.TotalCost,
		AverageLatency:     m.AverageLatency,
		SuccessRate:        m.GetSuccessRate(),
		CacheHitRate:       m.GetCacheHitRate(),
		ProviderMetrics:    m.GetAllProviderMetrics(),
		RequestsByType:     m.copyRequestsByType(),
		ErrorsByType:       m.copyErrorsByType(),
		Uptime:             time.Since(m.StartTime),
		LastRequestTime:    m.LastRequestTime,
	}
}

func (m *UsageMetrics) copyRequestsByType() map[AnalysisType]int64 {
	result := make(map[AnalysisType]int64, len(m.RequestsByType))
	for k, v := range m.RequestsByType {
		result[k] = v
	}
	return result
}

func (m *UsageMetrics) copyErrorsByType() map[string]int64 {
	result := make(map[string]int64, len(m.ErrorsByType))
	for k, v := range m.ErrorsByType {
		result[k] = v
	}
	return result
}
