package copilot

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/marccoxall/helmchecker/internal/ai"
	"golang.org/x/time/rate"
)

// CopilotProvider implements the ai.Provider interface for GitHub Copilot
type CopilotProvider struct {
	config        Config
	client        *http.Client
	tokenProvider TokenProvider
	rateLimiter   *rate.Limiter
	mu            sync.RWMutex
	metrics       *ai.UsageMetrics
}

// NewCopilotProvider creates a new GitHub Copilot provider
func NewCopilotProvider(config Config, tokenProvider TokenProvider) (*CopilotProvider, error) {
	if tokenProvider == nil {
		return nil, ErrNoToken
	}

	// Validate the token
	if err := tokenProvider.ValidateToken(); err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Create authenticated HTTP client
	client := NewAuthenticatedClient(tokenProvider)
	client.Timeout = config.Timeout

	// Create rate limiter (requests per minute)
	rps := float64(config.RateLimitPerMinute) / 60.0
	rateLimiter := rate.NewLimiter(rate.Limit(rps), config.RateLimitPerMinute)

	return &CopilotProvider{
		config:        config,
		client:        client,
		tokenProvider: tokenProvider,
		rateLimiter:   rateLimiter,
		metrics:       ai.NewUsageMetrics(),
	}, nil
}

// Name returns the provider name
func (p *CopilotProvider) Name() string {
	return "github-copilot"
}

// Analyze sends an analysis request to GitHub Copilot
func (p *CopilotProvider) Analyze(ctx context.Context, req *ai.Request) (*ai.Response, error) {
	startTime := time.Now()

	// Wait for rate limiter
	if err := p.rateLimiter.Wait(ctx); err != nil {
		p.metrics.RecordFailure(p.Name(), "rate_limit")
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	// Build the prompt from the request
	chatReq := p.buildChatRequest(req)

	// Make the API request with retries
	var chatResp *ChatResponse
	var err error

	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(p.config.RetryDelay * time.Duration(attempt)):
			}
		}

		chatResp, err = p.doRequest(ctx, chatReq)
		if err == nil {
			break
		}

		// Don't retry on context cancellation or certain errors
		if ctx.Err() != nil || !isRetryableError(err) {
			break
		}
	}

	// Record metrics
	duration := time.Since(startTime)

	if err != nil {
		p.metrics.RecordFailure(p.Name(), "request_failed")
		return nil, err
	}

	// Build AI response
	resp := p.buildAIResponse(req, chatResp, duration, false)
	
	// Record success
	p.metrics.RecordRequest(p.Name(), resp.TokensUsed)
	p.metrics.RecordLatency(p.Name(), duration)
	
	if req.Type != "" {
		p.metrics.RecordRequestType(req.Type)
	}

	return resp, nil
}

// AnalyzeStream sends a streaming analysis request to GitHub Copilot
func (p *CopilotProvider) AnalyzeStream(ctx context.Context, req *ai.Request) (<-chan ai.StreamChunk, error) {
	// Wait for rate limiter
	if err := p.rateLimiter.Wait(ctx); err != nil {
		p.metrics.RecordFailure(p.Name(), "rate_limit")
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	// Build the streaming request
	chatReq := p.buildChatRequest(req)
	chatReq.Stream = true

	// Make the streaming request
	chunks, err := p.doStreamingRequest(ctx, chatReq)
	if err != nil {
		p.metrics.RecordFailure(p.Name(), "streaming_failed")
		return nil, err
	}

	return chunks, nil
}

// Validate checks if the provider is properly configured and accessible
func (p *CopilotProvider) Validate(ctx context.Context) error {
	// Create a simple validation request
	req := &ai.Request{
		Query:     "ping",
		Type:      ai.AnalysisTypeGeneral,
		MaxTokens: 10,
	}

	_, err := p.Analyze(ctx, req)
	return err
}

// GetMetrics returns usage metrics for this provider
func (p *CopilotProvider) GetMetrics() *ai.UsageMetrics {
	return p.metrics
}

// Close cleans up resources
func (p *CopilotProvider) Close() error {
	// Nothing to close for HTTP client
	return nil
}

// buildChatRequest converts an AI request to a Copilot chat request
func (p *CopilotProvider) buildChatRequest(req *ai.Request) *ChatRequest {
	// Build the system message
	systemMessage := p.buildSystemMessage(req)
	
	// Build the user message
	userMessage := p.buildUserMessage(req)

	messages := []Message{
		{Role: "system", Content: systemMessage},
		{Role: "user", Content: userMessage},
	}

	// Set defaults
	model := p.config.Model
	temperature := float32(p.config.Temperature)
	if req.Temperature > 0 {
		temperature = float32(req.Temperature)
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.config.MaxTokens
	}

	return &ChatRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      req.Options.Stream,
	}
}

// buildSystemMessage creates the system prompt
func (p *CopilotProvider) buildSystemMessage(req *ai.Request) string {
	return "You are an expert DevOps engineer specializing in Kubernetes, Helm, and GitOps patterns. " +
		"You provide detailed, accurate analysis of deployment configurations, identify potential issues, " +
		"and suggest best practices. Always structure your responses clearly and provide actionable recommendations."
}

// buildUserMessage creates the user prompt from the request
func (p *CopilotProvider) buildUserMessage(req *ai.Request) string {
	var buf strings.Builder

	// Add query
	buf.WriteString(req.Query)
	buf.WriteString("\n\n")

	// Add context if available
	if req.Context != nil {
		buf.WriteString("## Context\n\n")
		
		if req.Context.RepositoryInfo != nil {
			buf.WriteString(fmt.Sprintf("Repository: %s/%s\n", 
				req.Context.RepositoryInfo.Owner, 
				req.Context.RepositoryInfo.Name))
			buf.WriteString(fmt.Sprintf("Branch: %s\n", req.Context.RepositoryInfo.Branch))
			buf.WriteString(fmt.Sprintf("Commit: %s\n\n", req.Context.RepositoryInfo.CommitSHA))
		}

		if len(req.Context.DetectedPatterns) > 0 {
			buf.WriteString("### Detected GitOps Patterns:\n")
			for _, pattern := range req.Context.DetectedPatterns {
				buf.WriteString(fmt.Sprintf("- %s (v%s) at %s\n", pattern.Type, pattern.Version, pattern.Path))
			}
			buf.WriteString("\n")
		}

		if len(req.Context.HelmCharts) > 0 {
			buf.WriteString("### Helm Charts:\n")
			for _, chart := range req.Context.HelmCharts {
				buf.WriteString(fmt.Sprintf("- %s (v%s)\n", chart.Name, chart.Version))
				if chart.LatestVersion != "" && chart.LatestVersion != chart.Version {
					buf.WriteString(fmt.Sprintf("  Latest version: %s\n", chart.LatestVersion))
				}
			}
			buf.WriteString("\n")
		}

		if len(req.Context.Constraints) > 0 {
			buf.WriteString("### Constraints:\n")
			for _, constraint := range req.Context.Constraints {
				buf.WriteString(fmt.Sprintf("- %s\n", constraint))
			}
			buf.WriteString("\n")
		}
	}

	// Add response format hints
	if req.Options.ResponseFormat == "json" {
		buf.WriteString("\nPlease respond in JSON format.\n")
	} else if req.Options.ResponseFormat == "markdown" {
		buf.WriteString("\nPlease respond in Markdown format.\n")
	}

	return buf.String()
}

// buildAIResponse converts a Copilot response to an AI response
func (p *CopilotProvider) buildAIResponse(req *ai.Request, chatResp *ChatResponse, duration time.Duration, cached bool) *ai.Response {
	var content string
	if len(chatResp.Choices) > 0 && chatResp.Choices[0].Message != nil {
		content = chatResp.Choices[0].Message.Content
	}

	return &ai.Response{
		ID:       req.ID,
		Content:  content,
		Provider: p.Name(),
		Duration: duration,
		TokensUsed: ai.TokenUsage{
			PromptTokens:     chatResp.Usage.PromptTokens,
			CompletionTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:      chatResp.Usage.TotalTokens,
			EstimatedCost:    estimateCost(chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens, chatResp.Model),
		},
		Cached:   cached,
		Metadata: map[string]string{
			"model":         chatResp.Model,
			"finish_reason": getFinishReason(chatResp),
		},
	}
}

// estimateCost estimates the cost of a request (rough approximation)
func estimateCost(promptTokens, completionTokens int, model string) float64 {
	// Rough cost estimates (these would need to be updated based on actual pricing)
	var promptCostPer1k, completionCostPer1k float64

	switch {
	case strings.Contains(model, "gpt-4"):
		promptCostPer1k = 0.03
		completionCostPer1k = 0.06
	case strings.Contains(model, "gpt-3.5"):
		promptCostPer1k = 0.0015
		completionCostPer1k = 0.002
	default:
		promptCostPer1k = 0.01
		completionCostPer1k = 0.02
	}

	promptCost := float64(promptTokens) / 1000.0 * promptCostPer1k
	completionCost := float64(completionTokens) / 1000.0 * completionCostPer1k

	return promptCost + completionCost
}

// getFinishReason extracts the finish reason from the response
func getFinishReason(resp *ChatResponse) string {
	if len(resp.Choices) > 0 {
		return resp.Choices[0].FinishReason
	}
	return ""
}

// doRequest performs a non-streaming API request
func (p *CopilotProvider) doRequest(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/chat/completions", p.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		return nil, p.handleErrorResponse(httpResp)
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp, nil
}

// doStreamingRequest performs a streaming API request
func (p *CopilotProvider) doStreamingRequest(ctx context.Context, req *ChatRequest) (<-chan ai.StreamChunk, error) {
	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/chat/completions", p.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Accept", "text/event-stream")

	// Send request
	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		defer httpResp.Body.Close()
		return nil, p.handleErrorResponse(httpResp)
	}

	// Create output channel
	chunks := make(chan ai.StreamChunk, 10)

	// Start goroutine to read stream
	go p.readStream(ctx, httpResp.Body, chunks)

	return chunks, nil
}

// readStream reads and parses the streaming response
func (p *CopilotProvider) readStream(ctx context.Context, body io.ReadCloser, chunks chan<- ai.StreamChunk) {
	defer close(chunks)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	var totalTokens int

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// SSE format: "data: {...}"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Remove "data: " prefix
		data := strings.TrimPrefix(line, "data: ")

		// Check for end of stream
		if data == "[DONE]" {
			break
		}

		// Parse chunk
		var streamChunk StreamChunk
		if err := json.Unmarshal([]byte(data), &streamChunk); err != nil {
			// Send error chunk
			chunks <- ai.StreamChunk{
				Error: fmt.Errorf("failed to parse chunk: %w", err),
			}
			return
		}

		// Convert to AI chunk
		var content string
		var done bool
		if len(streamChunk.Choices) > 0 {
			if streamChunk.Choices[0].Delta != nil {
				content = streamChunk.Choices[0].Delta.Content
			}
			done = streamChunk.Choices[0].FinishReason != ""
		}

		// Estimate token count (rough approximation)
		if content != "" {
			totalTokens += len(strings.Fields(content))
		}

		chunks <- ai.StreamChunk{
			Content: content,
			Done:    done,
		}
	}

	// Record token usage estimate
	if totalTokens > 0 {
		tokenUsage := ai.TokenUsage{
			CompletionTokens: totalTokens,
			TotalTokens:      totalTokens,
		}
		p.metrics.RecordRequest(p.Name(), tokenUsage)
	}

	if err := scanner.Err(); err != nil {
		chunks <- ai.StreamChunk{
			Error: fmt.Errorf("stream read error: %w", err),
		}
	}
}

// handleErrorResponse processes error responses from the API
func (p *CopilotProvider) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP %d: failed to read error response", resp.StatusCode)
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return fmt.Errorf("API error (%s): %s", errResp.Error.Code, errResp.Error.Message)
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Retry on rate limit errors
	if strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "429") {
		return true
	}

	// Retry on temporary network errors
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "connection") {
		return true
	}

	// Retry on server errors (5xx)
	if strings.Contains(errStr, "500") || strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") || strings.Contains(errStr, "504") {
		return true
	}

	return false
}
