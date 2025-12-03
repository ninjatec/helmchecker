package ai

import (
	"context"
	"io"
)

// Provider defines the interface for AI service providers
type Provider interface {
	// Name returns the provider name (e.g., "openai", "copilot")
	Name() string

	// Analyze sends an analysis request to the AI provider
	Analyze(ctx context.Context, req *Request) (*Response, error)

	// AnalyzeStream sends an analysis request and streams the response
	AnalyzeStream(ctx context.Context, req *Request) (<-chan StreamChunk, error)

	// Validate checks if the provider is properly configured and accessible
	Validate(ctx context.Context) error

	// GetMetrics returns usage metrics for this provider
	GetMetrics() *UsageMetrics

	// Close cleans up any resources held by the provider
	Close() error
}

// StreamChunk represents a chunk of streaming response data
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

// ProviderFactory creates provider instances based on configuration
type ProviderFactory interface {
	// Create creates a new provider instance from configuration
	Create(config *ProviderConfig) (Provider, error)

	// SupportedTypes returns the list of provider types this factory supports
	SupportedTypes() []string
}

// ProviderRegistry manages multiple provider factories
type ProviderRegistry struct {
	factories map[string]ProviderFactory
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		factories: make(map[string]ProviderFactory),
	}
}

// Register registers a provider factory for a given type
func (r *ProviderRegistry) Register(providerType string, factory ProviderFactory) {
	r.factories[providerType] = factory
}

// Create creates a provider instance from configuration
func (r *ProviderRegistry) Create(config *ProviderConfig) (Provider, error) {
	factory, ok := r.factories[config.Type]
	if !ok {
		return nil, &ErrProviderNotSupported{Type: config.Type}
	}
	return factory.Create(config)
}

// GetSupportedTypes returns all registered provider types
func (r *ProviderRegistry) GetSupportedTypes() []string {
	types := make([]string, 0, len(r.factories))
	for t := range r.factories {
		types = append(types, t)
	}
	return types
}

// ProviderChain implements a fallback chain of providers
type ProviderChain struct {
	providers []Provider
	metrics   *UsageMetrics
}

// NewProviderChain creates a new provider chain with fallback support
func NewProviderChain(providers ...Provider) *ProviderChain {
	return &ProviderChain{
		providers: providers,
		metrics:   NewUsageMetrics(),
	}
}

// Name returns a combined name of all providers in the chain
func (c *ProviderChain) Name() string {
	if len(c.providers) == 0 {
		return "empty-chain"
	}
	return c.providers[0].Name() + "-chain"
}

// Analyze tries each provider in order until one succeeds
func (c *ProviderChain) Analyze(ctx context.Context, req *Request) (*Response, error) {
	var lastErr error
	for _, provider := range c.providers {
		resp, err := provider.Analyze(ctx, req)
		if err == nil {
			c.metrics.RecordRequest(provider.Name(), resp.TokensUsed)
			return resp, nil
		}
		lastErr = err
	}
	return nil, &ErrAllProvidersFailed{LastError: lastErr}
}

// AnalyzeStream tries each provider in order until one succeeds
func (c *ProviderChain) AnalyzeStream(ctx context.Context, req *Request) (<-chan StreamChunk, error) {
	var lastErr error
	for _, provider := range c.providers {
		stream, err := provider.AnalyzeStream(ctx, req)
		if err == nil {
			return stream, nil
		}
		lastErr = err
	}
	return nil, &ErrAllProvidersFailed{LastError: lastErr}
}

// Validate validates all providers in the chain
func (c *ProviderChain) Validate(ctx context.Context) error {
	for _, provider := range c.providers {
		if err := provider.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

// GetMetrics returns combined metrics from all providers
func (c *ProviderChain) GetMetrics() *UsageMetrics {
	combined := NewUsageMetrics()
	for _, provider := range c.providers {
		metrics := provider.GetMetrics()
		combined.Merge(metrics)
	}
	return combined
}

// Close closes all providers in the chain
func (c *ProviderChain) Close() error {
	var errs []error
	for _, provider := range c.providers {
		if err := provider.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return &ErrMultipleProviderErrors{Errors: errs}
	}
	return nil
}

// ResponseWriter provides a writer interface for streaming responses
type ResponseWriter interface {
	io.Writer
	
	// Complete marks the response as complete
	Complete() error
	
	// Error marks the response with an error
	Error(err error) error
}
