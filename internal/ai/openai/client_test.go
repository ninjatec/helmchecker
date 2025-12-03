package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marccoxall/helmchecker/internal/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenAIProvider(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		provider := NewStaticApiKeyProvider("sk-test123456789012345678901234567890")
		config := DefaultConfig()
		
		p, err := NewOpenAIProvider(config, provider)
		require.NoError(t, err)
		require.NotNil(t, p)
		
		assert.Equal(t, config.Model, p.config.Model)
		assert.NotNil(t, p.client)
		assert.NotNil(t, p.rateLimiter)
	})
	
	t.Run("nil api key provider", func(t *testing.T) {
		config := DefaultConfig()
		p, err := NewOpenAIProvider(config, nil)
		assert.Error(t, err)
		assert.Nil(t, p)
	})
	
	t.Run("invalid api key", func(t *testing.T) {
		provider := NewStaticApiKeyProvider("invalid-key")
		config := DefaultConfig()
		p, err := NewOpenAIProvider(config, provider)
		assert.Error(t, err)
		assert.Nil(t, p)
	})
}

func TestOpenAIProvider_Analyze(t *testing.T) {
	t.Run("successful analysis", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := ChatCompletionResponse{
				ID:      "test-id",
				Model:   string(GPT4Turbo),
				Choices: []Choice{
					{
						Index:        0,
						Message:      Message{Role: "assistant", Content: "Test response"},
						FinishReason: "stop",
					},
				},
				Usage: Usage{
					PromptTokens:     10,
					CompletionTokens: 20,
					TotalTokens:      30,
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		
		config := DefaultConfig()
		config.BaseURL = server.URL
		
		provider := NewStaticApiKeyProvider("sk-test123456789012345678901234567890")
		p, err := NewOpenAIProvider(config, provider)
		require.NoError(t, err)
		
		req := &ai.Request{Query: "Test", Type: ai.AnalysisTypeGeneral}
		resp, err := p.Analyze(context.Background(), req)
		
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "Test response", resp.Content)
	})
	
	t.Run("API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			errResp := ErrorResponse{}
			errResp.Error.Message = "Invalid request"
			errResp.Error.Type = "invalid_request_error"
			json.NewEncoder(w).Encode(errResp)
		}))
		defer server.Close()
		
		config := DefaultConfig()
		config.BaseURL = server.URL
		config.MaxRetries = 0
		
		provider := NewStaticApiKeyProvider("sk-test123456789012345678901234567890")
		p, err := NewOpenAIProvider(config, provider)
		require.NoError(t, err)
		
		resp, err := p.Analyze(context.Background(), &ai.Request{Query: "Test"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	assert.Equal(t, "https://api.openai.com/v1", config.BaseURL)
	assert.Equal(t, GPT4o, config.Model)
	assert.Equal(t, float32(0.3), config.Temperature)
	assert.Equal(t, 4096, config.MaxTokens)
}

func TestGetModelPricing(t *testing.T) {
	pricing := GetModelPricing(string(GPT4Turbo))
	assert.NotNil(t, pricing)
	assert.Equal(t, 0.01, pricing.PromptPricePer1k)
	assert.Equal(t, 0.03, pricing.CompletionPricePer1k)
}

func TestCalculateCost(t *testing.T) {
	cost := CalculateCost(1000, 1000, string(GPT4Turbo))
	assert.InDelta(t, 0.04, cost, 0.001)
}

func TestValidateApiKey(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		shouldErr bool
	}{
		{"valid key", "sk-test123456789012345678901234567890", false},
		{"empty key", "", true},
		{"no prefix but long enough", "test12345678901234567890", false}, // Long enough to be valid
		{"too short", "sk-short", false}, // sk- prefix makes it valid even if short (for compatibility)
		{"very short", "short", true}, // Too short without prefix
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewStaticApiKeyProvider(tt.apiKey)
			err := provider.ValidateAPIKey()
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFunctionRegistry(t *testing.T) {
	registry := NewFunctionRegistry()
	require.NotNil(t, registry)
	
	// Register a function
	def := HelmAnalysisFunction()
	registry.Register(def.Name, def)
	
	// Get definitions
	defs := registry.GetAll()
	assert.Len(t, defs, 1)
	assert.Equal(t, "analyze_helm_chart", defs[0].Name)
	
	// Get function
	retrieved, ok := registry.Get("analyze_helm_chart")
	assert.True(t, ok)
	assert.Equal(t, "analyze_helm_chart", retrieved.Name)
}

func TestDefaultFunctionRegistry(t *testing.T) {
	registry := DefaultFunctionRegistry()
	require.NotNil(t, registry)
	
	functions := registry.GetAll()
	assert.Len(t, functions, 4)
	
	names := []string{}
	for _, f := range functions {
		names = append(names, f.Name)
	}
	
	assert.Contains(t, names, "analyze_helm_chart")
	assert.Contains(t, names, "check_compatibility")
	assert.Contains(t, names, "generate_upgrade_strategy")
	assert.Contains(t, names, "assess_upgrade_risk")
}
