package copilot

import (
	"time"
)

// ChatRequest represents a request to the GitHub Copilot Chat API
type ChatRequest struct {
	// Model specifies which model to use (e.g., "gpt-4", "gpt-3.5-turbo")
	Model string `json:"model"`

	// Messages contains the conversation history
	Messages []Message `json:"messages"`

	// Temperature controls randomness (0.0-2.0)
	Temperature float32 `json:"temperature,omitempty"`

	// MaxTokens limits the response length
	MaxTokens int `json:"max_tokens,omitempty"`

	// TopP controls nucleus sampling
	TopP float32 `json:"top_p,omitempty"`

	// Stream enables streaming responses
	Stream bool `json:"stream,omitempty"`

	// N specifies how many completions to generate
	N int `json:"n,omitempty"`

	// Stop sequences where the API will stop generating
	Stop []string `json:"stop,omitempty"`
}

// Message represents a single message in the conversation
type Message struct {
	// Role is either "system", "user", or "assistant"
	Role string `json:"role"`

	// Content is the message text
	Content string `json:"content"`

	// Name is an optional identifier for the message author
	Name string `json:"name,omitempty"`
}

// ChatResponse represents a response from the GitHub Copilot Chat API
type ChatResponse struct {
	// ID is the unique identifier for this completion
	ID string `json:"id"`

	// Object is the type of object returned
	Object string `json:"object"`

	// Created is the Unix timestamp of when the completion was created
	Created int64 `json:"created"`

	// Model is the model used for this completion
	Model string `json:"model"`

	// Choices contains the generated completions
	Choices []Choice `json:"choices"`

	// Usage contains token usage information
	Usage Usage `json:"usage"`
}

// Choice represents a single completion choice
type Choice struct {
	// Index is the choice index
	Index int `json:"index"`

	// Message is the generated message (for non-streaming responses)
	Message *Message `json:"message,omitempty"`

	// Delta contains the message delta (for streaming responses)
	Delta *Message `json:"delta,omitempty"`

	// FinishReason indicates why the completion finished
	FinishReason string `json:"finish_reason"`
}

// Usage represents token usage statistics
type Usage struct {
	// PromptTokens is the number of tokens in the prompt
	PromptTokens int `json:"prompt_tokens"`

	// CompletionTokens is the number of tokens in the completion
	CompletionTokens int `json:"completion_tokens"`

	// TotalTokens is the total number of tokens used
	TotalTokens int `json:"total_tokens"`
}

// StreamChunk represents a chunk from a streaming response
type StreamChunk struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

// ErrorResponse represents an error from the API
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Config represents configuration for the Copilot provider
type Config struct {
	// Token is the GitHub authentication token
	Token string

	// BaseURL is the API base URL
	BaseURL string

	// Model is the default model to use
	Model string

	// Temperature is the default temperature setting
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
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		BaseURL:            "https://api.githubcopilot.com",
		Model:              "gpt-4",
		Temperature:        0.3,
		MaxTokens:          4096,
		Timeout:            60 * time.Second,
		MaxRetries:         3,
		RetryDelay:         time.Second,
		RateLimitPerMinute: 60,
	}
}
