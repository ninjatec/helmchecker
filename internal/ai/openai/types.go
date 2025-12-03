package openai

// ChatCompletionRequest represents a request to the OpenAI Chat Completions API
type ChatCompletionRequest struct {
	// Model specifies which model to use
	Model string `json:"model"`

	// Messages contains the conversation history
	Messages []Message `json:"messages"`

	// Temperature controls randomness (0.0-2.0)
	Temperature float32 `json:"temperature,omitempty"`

	// MaxTokens limits the response length
	MaxTokens int `json:"max_tokens,omitempty"`

	// TopP controls nucleus sampling (0.0-1.0)
	TopP float32 `json:"top_p,omitempty"`

	// N specifies how many completions to generate
	N int `json:"n,omitempty"`

	// Stream enables streaming responses
	Stream bool `json:"stream,omitempty"`

	// Stop sequences where the API will stop generating
	Stop []string `json:"stop,omitempty"`

	// PresencePenalty penalizes new tokens based on presence (-2.0 to 2.0)
	PresencePenalty float32 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty penalizes new tokens based on frequency (-2.0 to 2.0)
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`

	// User is a unique identifier for the end-user
	User string `json:"user,omitempty"`

	// ResponseFormat specifies the format of the response
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`

	// Tools is a list of tools the model may call
	Tools []Tool `json:"tools,omitempty"`

	// ToolChoice controls which tool is called
	ToolChoice interface{} `json:"tool_choice,omitempty"`
}

// Message represents a single message in the conversation
type Message struct {
	// Role is "system", "user", "assistant", or "tool"
	Role string `json:"role"`

	// Content is the message text
	Content string `json:"content"`

	// Name is an optional identifier for the message author
	Name string `json:"name,omitempty"`

	// ToolCalls contains tool calls made by the assistant
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID is the ID of the tool call this message is responding to
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// ChatCompletionResponse represents a response from the OpenAI Chat Completions API
type ChatCompletionResponse struct {
	// ID is the unique identifier for this completion
	ID string `json:"id"`

	// Object is the type of object returned (e.g., "chat.completion")
	Object string `json:"object"`

	// Created is the Unix timestamp of when the completion was created
	Created int64 `json:"created"`

	// Model is the model used for this completion
	Model string `json:"model"`

	// Choices contains the generated completions
	Choices []Choice `json:"choices"`

	// Usage contains token usage information
	Usage Usage `json:"usage"`

	// SystemFingerprint is a unique identifier for the model configuration
	SystemFingerprint string `json:"system_fingerprint,omitempty"`
}

// Choice represents a single completion choice
type Choice struct {
	// Index is the choice index
	Index int `json:"index"`

	// Message is the generated message
	Message Message `json:"message"`

	// FinishReason indicates why the completion finished
	// Possible values: "stop", "length", "tool_calls", "content_filter"
	FinishReason string `json:"finish_reason"`

	// LogProbs contains log probabilities (if requested)
	LogProbs interface{} `json:"logprobs,omitempty"`
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
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice represents a choice in a streaming response
type StreamChoice struct {
	// Index is the choice index
	Index int `json:"index"`

	// Delta contains the message delta
	Delta MessageDelta `json:"delta"`

	// FinishReason indicates why the completion finished
	FinishReason string `json:"finish_reason,omitempty"`
}

// MessageDelta represents incremental message content in a stream
type MessageDelta struct {
	// Role is the message role (only in first chunk)
	Role string `json:"role,omitempty"`

	// Content is the incremental content
	Content string `json:"content,omitempty"`

	// ToolCalls contains incremental tool calls
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// FunctionDefinition defines a function that can be called
type FunctionDefinition struct {
	// Name is the function name
	Name string `json:"name"`

	// Description explains what the function does
	Description string `json:"description"`

	// Parameters describes the function parameters (JSON Schema)
	Parameters interface{} `json:"parameters"`
}

// FunctionCall represents a function call made by the model
type FunctionCall struct {
	// Name is the function name
	Name string `json:"name"`

	// Arguments are the function arguments (JSON string)
	Arguments string `json:"arguments"`
}

// Tool represents a tool the model can use
type Tool struct {
	// Type is the tool type (currently only "function")
	Type string `json:"type"`

	// Function is the function definition
	Function FunctionDefinition `json:"function"`
}

// ToolCall represents a tool call made by the model
type ToolCall struct {
	// ID is the tool call ID
	ID string `json:"id"`

	// Type is the tool type (currently only "function")
	Type string `json:"type"`

	// Function is the function call
	Function FunctionCall `json:"function"`
}

// ResponseFormat specifies the format of the response
type ResponseFormat struct {
	// Type is the format type ("text" or "json_object")
	Type string `json:"type"`
}

// ErrorResponse represents an error from the API
type ErrorResponse struct {
	Error struct {
		// Message is the error message
		Message string `json:"message"`

		// Type is the error type
		Type string `json:"type"`

		// Param is the parameter that caused the error
		Param string `json:"param,omitempty"`

		// Code is the error code
		Code string `json:"code,omitempty"`
	} `json:"error"`
}
