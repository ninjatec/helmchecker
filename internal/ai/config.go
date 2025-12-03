package ai

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete AI configuration
type Config struct {
	AI AIConfig `yaml:"ai"`
}

// AIConfig contains AI provider configuration
type AIConfig struct {
	Providers    []ProviderConfig   `yaml:"providers"`
	Caching      CachingConfig      `yaml:"caching"`
	RateLimiting RateLimitingConfig `yaml:"rate_limiting"`
}

// CachingConfig contains cache configuration
type CachingConfig struct {
	Enabled bool   `yaml:"enabled"`
	TTL     int    `yaml:"ttl"` // seconds
	MaxSize string `yaml:"max_size"`
}

// RateLimitingConfig contains rate limiting configuration
type RateLimitingConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	TokensPerMinute   int `yaml:"tokens_per_minute"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ErrInvalidConfiguration{
			Field:  "config_file",
			Reason: fmt.Sprintf("failed to read file: %v", err),
		}
	}

	// Expand environment variables in the YAML
	expanded := os.ExpandEnv(string(data))

	var config Config
	if err := yaml.Unmarshal([]byte(expanded), &config); err != nil {
		return nil, &ErrInvalidConfiguration{
			Field:  "yaml",
			Reason: fmt.Sprintf("failed to parse YAML: %v", err),
		}
	}

	// Validate the configuration
	if err := ValidateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadConfigFromBytes loads configuration from byte array
func LoadConfigFromBytes(data []byte) (*Config, error) {
	expanded := os.ExpandEnv(string(data))

	var config Config
	if err := yaml.Unmarshal([]byte(expanded), &config); err != nil {
		return nil, &ErrInvalidConfiguration{
			Field:  "yaml",
			Reason: fmt.Sprintf("failed to parse YAML: %v", err),
		}
	}

	if err := ValidateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ValidateConfig validates the configuration
func ValidateConfig(config *Config) error {
	if config == nil {
		return &ErrInvalidConfiguration{
			Field:  "config",
			Reason: "config is nil",
		}
	}

	// Validate providers
	if len(config.AI.Providers) == 0 {
		return &ErrInvalidConfiguration{
			Field:  "ai.providers",
			Reason: "at least one provider must be configured",
		}
	}

	providerNames := make(map[string]bool)
	for i, provider := range config.AI.Providers {
		// Validate provider name
		if provider.Name == "" {
			return &ErrInvalidConfiguration{
				Field:  fmt.Sprintf("ai.providers[%d].name", i),
				Reason: "provider name is required",
			}
		}

		// Check for duplicate names
		if providerNames[provider.Name] {
			return &ErrInvalidConfiguration{
				Field:  fmt.Sprintf("ai.providers[%d].name", i),
				Reason: fmt.Sprintf("duplicate provider name: %s", provider.Name),
			}
		}
		providerNames[provider.Name] = true

		// Validate provider type
		if provider.Type == "" {
			return &ErrInvalidConfiguration{
				Field:  fmt.Sprintf("ai.providers[%d].type", i),
				Reason: "provider type is required",
			}
		}

		// Validate supported types
		validTypes := map[string]bool{
			"openai":   true,
			"copilot":  true,
			"anthropic": true,
			"custom":   true,
		}
		if !validTypes[provider.Type] {
			return &ErrInvalidConfiguration{
				Field:  fmt.Sprintf("ai.providers[%d].type", i),
				Reason: fmt.Sprintf("unsupported provider type: %s", provider.Type),
			}
		}

		// Validate authentication
		if err := validateAuth(&provider.Auth, i); err != nil {
			return err
		}

		// Validate priority
		if provider.Priority < 0 {
			return &ErrInvalidConfiguration{
				Field:  fmt.Sprintf("ai.providers[%d].priority", i),
				Reason: "priority must be non-negative",
			}
		}
	}

	// Validate caching config
	if config.AI.Caching.Enabled {
		if config.AI.Caching.TTL < 0 {
			return &ErrInvalidConfiguration{
				Field:  "ai.caching.ttl",
				Reason: "TTL must be non-negative",
			}
		}

		if config.AI.Caching.MaxSize != "" {
			if _, err := ParseSize(config.AI.Caching.MaxSize); err != nil {
				return &ErrInvalidConfiguration{
					Field:  "ai.caching.max_size",
					Reason: fmt.Sprintf("invalid size format: %v", err),
				}
			}
		}
	}

	// Validate rate limiting
	if config.AI.RateLimiting.RequestsPerMinute < 0 {
		return &ErrInvalidConfiguration{
			Field:  "ai.rate_limiting.requests_per_minute",
			Reason: "requests_per_minute must be non-negative",
		}
	}

	if config.AI.RateLimiting.TokensPerMinute < 0 {
		return &ErrInvalidConfiguration{
			Field:  "ai.rate_limiting.tokens_per_minute",
			Reason: "tokens_per_minute must be non-negative",
		}
	}

	return nil
}

// validateAuth validates authentication configuration
func validateAuth(auth *AuthConfig, index int) error {
	if auth == nil {
		return &ErrInvalidConfiguration{
			Field:  fmt.Sprintf("ai.providers[%d].auth", index),
			Reason: "authentication configuration is required",
		}
	}

	hasAuth := false
	if auth.Token != "" {
		hasAuth = true
	}
	if auth.APIKey != "" {
		hasAuth = true
	}
	if auth.Username != "" && auth.Password != "" {
		hasAuth = true
	}
	if len(auth.AdditionalAuth) > 0 {
		hasAuth = true
	}

	if !hasAuth {
		return &ErrInvalidConfiguration{
			Field:  fmt.Sprintf("ai.providers[%d].auth", index),
			Reason: "at least one authentication method must be provided",
		}
	}

	return nil
}

// ParseSize parses a size string (e.g., "100MB", "1GB") into bytes
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))
	if sizeStr == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Order matters - check longer suffixes first
	suffixes := []struct {
		suffix     string
		multiplier int64
	}{
		{"TB", 1024 * 1024 * 1024 * 1024},
		{"GB", 1024 * 1024 * 1024},
		{"MB", 1024 * 1024},
		{"KB", 1024},
		{"B", 1},
	}

	for _, s := range suffixes {
		if strings.HasSuffix(sizeStr, s.suffix) {
			numStr := strings.TrimSuffix(sizeStr, s.suffix)
			numStr = strings.TrimSpace(numStr)

			var num float64
			_, err := fmt.Sscanf(numStr, "%f", &num)
			if err != nil {
				return 0, fmt.Errorf("invalid number: %v", err)
			}

			return int64(num * float64(s.multiplier)), nil
		}
	}

	// Try to parse as raw number (bytes)
	var num int64
	_, err := fmt.Sscanf(sizeStr, "%d", &num)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	return num, nil
}

// GetCacheTTL returns the cache TTL as a duration
func (c *CachingConfig) GetCacheTTL() time.Duration {
	return time.Duration(c.TTL) * time.Second
}

// GetMaxCacheSize returns the max cache size in bytes
func (c *CachingConfig) GetMaxCacheSize() (int64, error) {
	if c.MaxSize == "" {
		return 100 * 1024 * 1024, nil // Default 100MB
	}
	return ParseSize(c.MaxSize)
}

// GetProviderByName returns a provider config by name
func (c *Config) GetProviderByName(name string) (*ProviderConfig, error) {
	for _, provider := range c.AI.Providers {
		if provider.Name == name {
			return &provider, nil
		}
	}
	return nil, fmt.Errorf("provider not found: %s", name)
}

// GetEnabledProviders returns all enabled providers sorted by priority
func (c *Config) GetEnabledProviders() []ProviderConfig {
	var enabled []ProviderConfig
	for _, provider := range c.AI.Providers {
		if provider.Enabled {
			enabled = append(enabled, provider)
		}
	}

	// Sort by priority (lower number = higher priority)
	for i := 0; i < len(enabled); i++ {
		for j := i + 1; j < len(enabled); j++ {
			if enabled[j].Priority < enabled[i].Priority {
				enabled[i], enabled[j] = enabled[j], enabled[i]
			}
		}
	}

	return enabled
}

// GetProvidersByType returns all providers of a specific type
func (c *Config) GetProvidersByType(providerType string) []ProviderConfig {
	var providers []ProviderConfig
	for _, provider := range c.AI.Providers {
		if provider.Type == providerType {
			providers = append(providers, provider)
		}
	}
	return providers
}

// MergeWithDefaults merges config with default values
func (c *Config) MergeWithDefaults() {
	// Set default caching values if not specified
	if c.AI.Caching.TTL == 0 {
		c.AI.Caching.TTL = 3600 // 1 hour default
	}

	if c.AI.Caching.MaxSize == "" {
		c.AI.Caching.MaxSize = "100MB"
	}

	// Set default rate limiting if not specified
	if c.AI.RateLimiting.RequestsPerMinute == 0 {
		c.AI.RateLimiting.RequestsPerMinute = 60
	}

	// Set default retry config for each provider
	for i := range c.AI.Providers {
		provider := &c.AI.Providers[i]
		
		if provider.Retry.MaxRetries == 0 {
			provider.Retry.MaxRetries = 3
		}
		
		if provider.Retry.InitialDelay == 0 {
			provider.Retry.InitialDelay = 1 * time.Second
		}
		
		if provider.Retry.MaxDelay == 0 {
			provider.Retry.MaxDelay = 30 * time.Second
		}
		
		if provider.Retry.Multiplier == 0 {
			provider.Retry.Multiplier = 2.0
		}

		// Set default cache config
		if provider.Cache.TTL == 0 {
			provider.Cache.TTL = time.Duration(c.AI.Caching.TTL) * time.Second
		}

		if !provider.Cache.Enabled && c.AI.Caching.Enabled {
			provider.Cache.Enabled = true
		}
	}
}

// ToYAML converts the config back to YAML
func (c *Config) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}

// ApplyEnvironmentOverrides applies environment variable overrides
func (c *Config) ApplyEnvironmentOverrides() {
	// Override global settings
	if val := os.Getenv("AI_CACHE_ENABLED"); val != "" {
		c.AI.Caching.Enabled = val == "true" || val == "1"
	}

	if val := os.Getenv("AI_CACHE_TTL"); val != "" {
		var ttl int
		if _, err := fmt.Sscanf(val, "%d", &ttl); err == nil {
			c.AI.Caching.TTL = ttl
		}
	}

	if val := os.Getenv("AI_CACHE_MAX_SIZE"); val != "" {
		c.AI.Caching.MaxSize = val
	}

	if val := os.Getenv("AI_RATE_LIMIT_RPM"); val != "" {
		var rpm int
		if _, err := fmt.Sscanf(val, "%d", &rpm); err == nil {
			c.AI.RateLimiting.RequestsPerMinute = rpm
		}
	}

	// Override provider-specific settings
	for i := range c.AI.Providers {
		provider := &c.AI.Providers[i]
		prefix := "AI_PROVIDER_" + strings.ToUpper(provider.Name) + "_"

		if val := os.Getenv(prefix + "ENABLED"); val != "" {
			provider.Enabled = val == "true" || val == "1"
		}

		if val := os.Getenv(prefix + "PRIORITY"); val != "" {
			var priority int
			if _, err := fmt.Sscanf(val, "%d", &priority); err == nil {
				provider.Priority = priority
			}
		}

		// Override auth
		if val := os.Getenv(prefix + "TOKEN"); val != "" {
			provider.Auth.Token = val
		}

		if val := os.Getenv(prefix + "API_KEY"); val != "" {
			provider.Auth.APIKey = val
		}
	}
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{
					Name:     "openai-default",
					Type:     "openai",
					Enabled:  true,
					Priority: 1,
					Auth: AuthConfig{
						APIKey: "${OPENAI_API_KEY}",
					},
					Config: map[string]interface{}{
						"model":       "gpt-4-turbo",
						"temperature": 0.3,
						"max_tokens":  4096,
					},
					Cache: CacheConfig{
						Enabled: true,
						TTL:     3600 * time.Second,
						MaxSize: 100 * 1024 * 1024,
					},
					Retry: RetryConfig{
						MaxRetries:   3,
						InitialDelay: 1 * time.Second,
						MaxDelay:     30 * time.Second,
						Multiplier:   2.0,
					},
				},
			},
			Caching: CachingConfig{
				Enabled: true,
				TTL:     3600,
				MaxSize: "100MB",
			},
			RateLimiting: RateLimitingConfig{
				RequestsPerMinute: 60,
				TokensPerMinute:   100000,
			},
		},
	}
}
