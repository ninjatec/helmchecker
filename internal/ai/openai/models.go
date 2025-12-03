package openai

import (
	"time"
)

// Model represents an OpenAI model identifier
type Model string

const (
	// GPT4 is the GPT-4 base model
	GPT4 Model = "gpt-4"

	// GPT4Turbo is the GPT-4 Turbo model
	GPT4Turbo Model = "gpt-4-turbo-preview"

	// GPT4o is the GPT-4o model (optimized)
	GPT4o Model = "gpt-4o"

	// GPT35Turbo is the GPT-3.5 Turbo model
	GPT35Turbo Model = "gpt-3.5-turbo"
)

// Config represents configuration for the OpenAI provider
type Config struct {
	// APIKey is the OpenAI API key
	APIKey string

	// BaseURL is the API base URL
	BaseURL string

	// Model is the default model to use
	Model Model

	// Temperature is the default temperature setting (0.0-2.0)
	Temperature float32

	// MaxTokens is the default max tokens setting
	MaxTokens int

	// Timeout is the request timeout
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// RetryDelay is the delay between retries
	RetryDelay time.Duration

	// RateLimitPerMinute is the rate limit for requests
	RateLimitPerMinute int

	// Organization is the optional organization ID
	Organization string
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		BaseURL:            "https://api.openai.com/v1",
		Model:              GPT4o,
		Temperature:        0.3,
		MaxTokens:          4096,
		Timeout:            60 * time.Second,
		MaxRetries:         3,
		RetryDelay:         time.Second,
		RateLimitPerMinute: 60,
	}
}

// ModelPricing represents pricing information for OpenAI models
type ModelPricing struct {
	Model                Model
	PromptPricePer1k     float64 // Price per 1,000 prompt tokens in USD
	CompletionPricePer1k float64 // Price per 1,000 completion tokens in USD
}

// GetModelPricing returns pricing information for a given model
func GetModelPricing(model string) *ModelPricing {
	pricingTable := map[string]ModelPricing{
		string(GPT4): {
			Model:                GPT4,
			PromptPricePer1k:     0.03,
			CompletionPricePer1k: 0.06,
		},
		string(GPT4Turbo): {
			Model:                GPT4Turbo,
			PromptPricePer1k:     0.01,
			CompletionPricePer1k: 0.03,
		},
		string(GPT4o): {
			Model:                GPT4o,
			PromptPricePer1k:     0.005,
			CompletionPricePer1k: 0.015,
		},
		string(GPT35Turbo): {
			Model:                GPT35Turbo,
			PromptPricePer1k:     0.0005,
			CompletionPricePer1k: 0.0015,
		},
	}

	if pricing, ok := pricingTable[model]; ok {
		return &pricing
	}

	// Default to GPT-4 pricing if model not found
	pricing := pricingTable[string(GPT4)]
	return &pricing
}

// CalculateCost calculates the cost of a request based on token usage
func CalculateCost(promptTokens, completionTokens int, model string) float64 {
	pricing := GetModelPricing(model)
	
	promptCost := float64(promptTokens) / 1000.0 * pricing.PromptPricePer1k
	completionCost := float64(completionTokens) / 1000.0 * pricing.CompletionPricePer1k
	
	return promptCost + completionCost
}
