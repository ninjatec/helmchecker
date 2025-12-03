package ai

import (
	"time"
)

// Request represents an analysis request to an AI provider
type Request struct {
	// ID is a unique identifier for this request
	ID string

	// Context contains the analysis context (repository info, patterns, etc.)
	Context *AnalysisContext

	// Query is the specific question or analysis task
	Query string

	// Type specifies the type of analysis requested
	Type AnalysisType

	// Options contains provider-specific options
	Options RequestOptions

	// MaxTokens limits the response length
	MaxTokens int

	// Temperature controls randomness (0.0-1.0)
	Temperature float64

	// Metadata for tracking and logging
	Metadata map[string]string
}

// Response represents an AI provider's response
type Response struct {
	// ID matches the request ID
	ID string

	// Content is the main response text
	Content string

	// StructuredData contains parsed structured output (if requested)
	StructuredData interface{}

	// Confidence is the provider's confidence in the response (0.0-1.0)
	Confidence float64

	// TokensUsed tracks token consumption
	TokensUsed TokenUsage

	// Provider identifies which provider generated this response
	Provider string

	// Duration is how long the request took
	Duration time.Duration

	// Metadata contains additional response information
	Metadata map[string]string

	// Cached indicates if this response came from cache
	Cached bool
}

// AnalysisContext provides context for AI analysis
type AnalysisContext struct {
	// RepositoryInfo contains repository metadata
	RepositoryInfo *RepositoryInfo

	// DetectedPatterns lists the GitOps patterns found
	DetectedPatterns []PatternInfo

	// HelmCharts lists the Helm charts being analyzed
	HelmCharts []HelmChartInfo

	// GitHistory contains relevant commit history
	GitHistory []CommitInfo

	// CurrentState describes the current deployment state
	CurrentState map[string]interface{}

	// TargetState describes the desired deployment state
	TargetState map[string]interface{}

	// Constraints are limitations or requirements for the analysis
	Constraints []string

	// AdditionalContext for custom data
	AdditionalContext map[string]interface{}
}

// RepositoryInfo contains repository metadata
type RepositoryInfo struct {
	Owner      string
	Name       string
	URL        string
	Branch     string
	CommitSHA  string
	LastUpdate time.Time
}

// PatternInfo describes a detected GitOps pattern
type PatternInfo struct {
	Type       string // "flux", "argocd", "kustomize", "kubernetes"
	Version    string
	Path       string
	Confidence float64
	Resources  []string
}

// HelmChartInfo contains Helm chart information
type HelmChartInfo struct {
	Name            string
	Version         string
	AppVersion      string
	Path            string
	ValuesFiles     []string
	Dependencies    []string
	Deprecated      bool
	LatestVersion   string
	BreakingChanges []string
}

// CommitInfo represents a Git commit
type CommitInfo struct {
	SHA       string
	Author    string
	Date      time.Time
	Message   string
	FilesChanged []string
}

// AnalysisType specifies the type of analysis
type AnalysisType string

const (
	// AnalysisTypePatternDetection detects GitOps patterns
	AnalysisTypePatternDetection AnalysisType = "pattern_detection"

	// AnalysisTypeCompatibility checks version compatibility
	AnalysisTypeCompatibility AnalysisType = "compatibility"

	// AnalysisTypeRiskAssessment assesses upgrade risks
	AnalysisTypeRiskAssessment AnalysisType = "risk_assessment"

	// AnalysisTypeRecommendation generates upgrade recommendations
	AnalysisTypeRecommendation AnalysisType = "recommendation"

	// AnalysisTypeImpact analyzes upgrade impact
	AnalysisTypeImpact AnalysisType = "impact"

	// AnalysisTypeStrategy generates upgrade strategies
	AnalysisTypeStrategy AnalysisType = "strategy"

	// AnalysisTypeConflict detects conflicts
	AnalysisTypeConflict AnalysisType = "conflict"

	// AnalysisTypeGeneral for general analysis
	AnalysisTypeGeneral AnalysisType = "general"
)

// RequestOptions contains provider-specific options
type RequestOptions struct {
	// Stream enables streaming responses
	Stream bool

	// UseCache controls cache usage
	UseCache bool

	// CacheTTL sets cache time-to-live
	CacheTTL time.Duration

	// RetryCount sets the number of retries
	RetryCount int

	// RetryDelay sets the delay between retries
	RetryDelay time.Duration

	// Timeout sets the request timeout
	Timeout time.Duration

	// ResponseFormat specifies the desired response format
	ResponseFormat string // "text", "json", "markdown"

	// IncludeConfidence requests confidence scores
	IncludeConfidence bool

	// AdditionalOptions for provider-specific settings
	AdditionalOptions map[string]interface{}
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	EstimatedCost    float64 // in USD
}

// Add calculates total tokens
func (t *TokenUsage) Add(other TokenUsage) {
	t.PromptTokens += other.PromptTokens
	t.CompletionTokens += other.CompletionTokens
	t.TotalTokens += other.TotalTokens
	t.EstimatedCost += other.EstimatedCost
}

// ProviderConfig contains provider configuration
type ProviderConfig struct {
	// Name is a unique identifier for this provider instance
	Name string

	// Type identifies the provider type (e.g., "openai", "copilot")
	Type string

	// Enabled indicates if this provider is active
	Enabled bool

	// Priority determines the order in fallback chains (lower = higher priority)
	Priority int

	// Auth contains authentication credentials
	Auth AuthConfig

	// Config contains provider-specific configuration
	Config map[string]interface{}

	// RateLimits defines rate limiting rules
	RateLimits RateLimitConfig

	// Cache contains caching configuration
	Cache CacheConfig

	// Retry contains retry configuration
	Retry RetryConfig
}

// AuthConfig contains authentication credentials
type AuthConfig struct {
	// Type specifies the auth type (e.g., "bearer", "api_key")
	Type string

	// Token for bearer token authentication
	Token string

	// APIKey for API key authentication
	APIKey string

	// Username for basic auth
	Username string

	// Password for basic auth
	Password string

	// AdditionalAuth for custom auth methods
	AdditionalAuth map[string]string
}

// RateLimitConfig defines rate limiting rules
type RateLimitConfig struct {
	// RequestsPerMinute limits requests per minute
	RequestsPerMinute int

	// RequestsPerHour limits requests per hour
	RequestsPerHour int

	// TokensPerMinute limits tokens per minute
	TokensPerMinute int

	// ConcurrentRequests limits concurrent requests
	ConcurrentRequests int
}

// CacheConfig contains caching configuration
type CacheConfig struct {
	// Enabled controls caching
	Enabled bool

	// TTL is the default time-to-live
	TTL time.Duration

	// MaxSize limits cache size in bytes
	MaxSize int64

	// EvictionPolicy determines cache eviction ("lru", "lfu", "fifo")
	EvictionPolicy string
}

// RetryConfig contains retry configuration
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// InitialDelay is the initial retry delay
	InitialDelay time.Duration

	// MaxDelay is the maximum retry delay
	MaxDelay time.Duration

	// Multiplier for exponential backoff
	Multiplier float64

	// RetryableErrors lists error patterns that should trigger retries
	RetryableErrors []string
}
