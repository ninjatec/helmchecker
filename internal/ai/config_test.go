package ai

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configData := `
ai:
  providers:
    - name: test-provider
      type: openai
      enabled: true
      priority: 1
      auth:
        type: api_key
        api_key: test-key
      config:
        model: gpt-4
        temperature: 0.5
  caching:
    enabled: true
    ttl: 3600
    max_size: 100MB
  rate_limiting:
    requests_per_minute: 60
    tokens_per_minute: 100000
`

	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(configData))
	require.NoError(t, err)
	tmpfile.Close()

	// Load config
	config, err := LoadConfig(tmpfile.Name())
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Verify loaded values
	assert.Len(t, config.AI.Providers, 1)
	assert.Equal(t, "test-provider", config.AI.Providers[0].Name)
	assert.Equal(t, "openai", config.AI.Providers[0].Type)
	assert.True(t, config.AI.Providers[0].Enabled)
	assert.Equal(t, "test-key", config.AI.Providers[0].Auth.APIKey)
	assert.True(t, config.AI.Caching.Enabled)
	assert.Equal(t, 3600, config.AI.Caching.TTL)
	assert.Equal(t, 60, config.AI.RateLimiting.RequestsPerMinute)
}

func TestLoadConfigWithEnvVars(t *testing.T) {
	// Set environment variable
	os.Setenv("TEST_API_KEY", "secret-key-123")
	defer os.Unsetenv("TEST_API_KEY")

	configData := `
ai:
  providers:
    - name: test-provider
      type: openai
      enabled: true
      priority: 1
      auth:
        type: api_key
        api_key: ${TEST_API_KEY}
  caching:
    enabled: true
    ttl: 3600
    max_size: 100MB
  rate_limiting:
    requests_per_minute: 60
`

	config, err := LoadConfigFromBytes([]byte(configData))
	require.NoError(t, err)
	assert.Equal(t, "secret-key-123", config.AI.Providers[0].Auth.APIKey)
}

func TestValidateConfig_Success(t *testing.T) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{
					Name:     "valid-provider",
					Type:     "openai",
					Enabled:  true,
					Priority: 1,
					Auth: AuthConfig{
						APIKey: "test-key",
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
			},
		},
	}

	err := ValidateConfig(config)
	assert.NoError(t, err)
}

func TestValidateConfig_NoProviders(t *testing.T) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{},
		},
	}

	err := ValidateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one provider must be configured")
}

func TestValidateConfig_MissingProviderName(t *testing.T) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{
					Type:    "openai",
					Enabled: true,
					Auth: AuthConfig{
						APIKey: "test",
					},
				},
			},
		},
	}

	err := ValidateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider name is required")
}

func TestValidateConfig_DuplicateProviderName(t *testing.T) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{
					Name:    "duplicate",
					Type:    "openai",
					Enabled: true,
					Auth:    AuthConfig{APIKey: "key1"},
				},
				{
					Name:    "duplicate",
					Type:    "copilot",
					Enabled: true,
					Auth:    AuthConfig{Token: "token1"},
				},
			},
		},
	}

	err := ValidateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate provider name")
}

func TestValidateConfig_InvalidProviderType(t *testing.T) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{
					Name:    "test",
					Type:    "invalid-type",
					Enabled: true,
					Auth:    AuthConfig{APIKey: "key"},
				},
			},
		},
	}

	err := ValidateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported provider type")
}

func TestValidateConfig_MissingAuth(t *testing.T) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{
					Name:    "test",
					Type:    "openai",
					Enabled: true,
					Auth:    AuthConfig{},
				},
			},
		},
	}

	err := ValidateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one authentication method must be provided")
}

func TestValidateConfig_InvalidCacheTTL(t *testing.T) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{
					Name:    "test",
					Type:    "openai",
					Enabled: true,
					Auth:    AuthConfig{APIKey: "key"},
				},
			},
			Caching: CachingConfig{
				Enabled: true,
				TTL:     -1,
			},
		},
	}

	err := ValidateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TTL must be non-negative")
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"100B", 100, false},
		{"1KB", 1024, false},
		{"10MB", 10 * 1024 * 1024, false},
		{"2GB", 2 * 1024 * 1024 * 1024, false},
		{"1024", 1024, false},
		{"1.5MB", int64(1.5 * 1024 * 1024), false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseSize(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCachingConfig_GetCacheTTL(t *testing.T) {
	config := CachingConfig{TTL: 3600}
	assert.Equal(t, 3600*time.Second, config.GetCacheTTL())
}

func TestCachingConfig_GetMaxCacheSize(t *testing.T) {
	tests := []struct {
		maxSize  string
		expected int64
	}{
		{"100MB", 100 * 1024 * 1024},
		{"1GB", 1024 * 1024 * 1024},
		{"", 100 * 1024 * 1024}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.maxSize, func(t *testing.T) {
			config := CachingConfig{MaxSize: tt.maxSize}
			size, err := config.GetMaxCacheSize()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, size)
		})
	}
}

func TestConfig_GetProviderByName(t *testing.T) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{Name: "provider1", Type: "openai"},
				{Name: "provider2", Type: "copilot"},
			},
		},
	}

	provider, err := config.GetProviderByName("provider1")
	assert.NoError(t, err)
	assert.Equal(t, "provider1", provider.Name)

	_, err = config.GetProviderByName("nonexistent")
	assert.Error(t, err)
}

func TestConfig_GetEnabledProviders(t *testing.T) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{Name: "p1", Type: "openai", Enabled: true, Priority: 2},
				{Name: "p2", Type: "copilot", Enabled: false, Priority: 1},
				{Name: "p3", Type: "openai", Enabled: true, Priority: 1},
			},
		},
	}

	enabled := config.GetEnabledProviders()
	assert.Len(t, enabled, 2)
	assert.Equal(t, "p3", enabled[0].Name) // Priority 1
	assert.Equal(t, "p1", enabled[1].Name) // Priority 2
}

func TestConfig_GetProvidersByType(t *testing.T) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{Name: "p1", Type: "openai"},
				{Name: "p2", Type: "copilot"},
				{Name: "p3", Type: "openai"},
			},
		},
	}

	openaiProviders := config.GetProvidersByType("openai")
	assert.Len(t, openaiProviders, 2)

	copilotProviders := config.GetProvidersByType("copilot")
	assert.Len(t, copilotProviders, 1)
}

func TestConfig_MergeWithDefaults(t *testing.T) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{
					Name:    "test",
					Type:    "openai",
					Enabled: true,
					Auth:    AuthConfig{APIKey: "key"},
				},
			},
			Caching: CachingConfig{
				Enabled: true,
			},
		},
	}

	config.MergeWithDefaults()

	// Check defaults were applied
	assert.Equal(t, 3600, config.AI.Caching.TTL)
	assert.Equal(t, "100MB", config.AI.Caching.MaxSize)
	assert.Equal(t, 60, config.AI.RateLimiting.RequestsPerMinute)
	assert.Equal(t, 3, config.AI.Providers[0].Retry.MaxRetries)
	assert.Equal(t, 1*time.Second, config.AI.Providers[0].Retry.InitialDelay)
}

func TestConfig_ApplyEnvironmentOverrides(t *testing.T) {
	// Set environment variables
	os.Setenv("AI_CACHE_ENABLED", "false")
	os.Setenv("AI_CACHE_TTL", "7200")
	os.Setenv("AI_CACHE_MAX_SIZE", "200MB")
	os.Setenv("AI_RATE_LIMIT_RPM", "120")
	os.Setenv("AI_PROVIDER_TEST_ENABLED", "true")
	os.Setenv("AI_PROVIDER_TEST_PRIORITY", "5")
	
	defer func() {
		os.Unsetenv("AI_CACHE_ENABLED")
		os.Unsetenv("AI_CACHE_TTL")
		os.Unsetenv("AI_CACHE_MAX_SIZE")
		os.Unsetenv("AI_RATE_LIMIT_RPM")
		os.Unsetenv("AI_PROVIDER_TEST_ENABLED")
		os.Unsetenv("AI_PROVIDER_TEST_PRIORITY")
	}()

	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{
					Name:     "test",
					Type:     "openai",
					Enabled:  false,
					Priority: 1,
					Auth:     AuthConfig{APIKey: "key"},
				},
			},
			Caching: CachingConfig{
				Enabled: true,
				TTL:     3600,
				MaxSize: "100MB",
			},
			RateLimiting: RateLimitingConfig{
				RequestsPerMinute: 60,
			},
		},
	}

	config.ApplyEnvironmentOverrides()

	// Verify overrides
	assert.False(t, config.AI.Caching.Enabled)
	assert.Equal(t, 7200, config.AI.Caching.TTL)
	assert.Equal(t, "200MB", config.AI.Caching.MaxSize)
	assert.Equal(t, 120, config.AI.RateLimiting.RequestsPerMinute)
	assert.True(t, config.AI.Providers[0].Enabled)
	assert.Equal(t, 5, config.AI.Providers[0].Priority)
}

func TestConfig_ToYAML(t *testing.T) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{
					Name:     "test",
					Type:     "openai",
					Enabled:  true,
					Priority: 1,
					Auth:     AuthConfig{APIKey: "key"},
				},
			},
			Caching: CachingConfig{
				Enabled: true,
				TTL:     3600,
				MaxSize: "100MB",
			},
		},
	}

	yaml, err := config.ToYAML()
	assert.NoError(t, err)
	assert.NotEmpty(t, yaml)
	assert.Contains(t, string(yaml), "test")
	assert.Contains(t, string(yaml), "openai")
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	require.NotNil(t, config)

	// Verify default config is valid
	err := ValidateConfig(config)
	assert.NoError(t, err)

	// Check default values
	assert.Len(t, config.AI.Providers, 1)
	assert.Equal(t, "openai-default", config.AI.Providers[0].Name)
	assert.True(t, config.AI.Caching.Enabled)
	assert.Equal(t, 3600, config.AI.Caching.TTL)
	assert.Equal(t, 60, config.AI.RateLimiting.RequestsPerMinute)
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	assert.Error(t, err)
}

func TestLoadConfigFromBytes_InvalidYAML(t *testing.T) {
	invalidYAML := []byte(`
invalid: yaml: content:
  - this is
  - not: valid
`)

	_, err := LoadConfigFromBytes(invalidYAML)
	assert.Error(t, err)
}

func TestValidateConfig_NegativePriority(t *testing.T) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{
					Name:     "test",
					Type:     "openai",
					Priority: -1,
					Auth:     AuthConfig{APIKey: "key"},
				},
			},
		},
	}

	err := ValidateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "priority must be non-negative")
}

func TestValidateConfig_NegativeRateLimits(t *testing.T) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{
					Name: "test",
					Type: "openai",
					Auth: AuthConfig{APIKey: "key"},
				},
			},
			RateLimiting: RateLimitingConfig{
				RequestsPerMinute: -1,
			},
		},
	}

	err := ValidateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requests_per_minute must be non-negative")
}

func BenchmarkLoadConfig(b *testing.B) {
	configData := `
ai:
  providers:
    - name: test-provider
      type: openai
      enabled: true
      priority: 1
      auth:
        api_key: test-key
  caching:
    enabled: true
    ttl: 3600
    max_size: 100MB
  rate_limiting:
    requests_per_minute: 60
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadConfigFromBytes([]byte(configData))
	}
}

func BenchmarkValidateConfig(b *testing.B) {
	config := &Config{
		AI: AIConfig{
			Providers: []ProviderConfig{
				{
					Name:     "test",
					Type:     "openai",
					Enabled:  true,
					Priority: 1,
					Auth:     AuthConfig{APIKey: "key"},
				},
			},
			Caching: CachingConfig{
				Enabled: true,
				TTL:     3600,
				MaxSize: "100MB",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateConfig(config)
	}
}
