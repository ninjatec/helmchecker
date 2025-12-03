package ai

import (
	"fmt"
	"strings"
)

// ErrProviderNotSupported indicates an unsupported provider type
type ErrProviderNotSupported struct {
	Type string
}

func (e *ErrProviderNotSupported) Error() string {
	return fmt.Sprintf("provider type '%s' is not supported", e.Type)
}

// ErrProviderNotConfigured indicates a provider is not properly configured
type ErrProviderNotConfigured struct {
	Provider string
	Reason   string
}

func (e *ErrProviderNotConfigured) Error() string {
	return fmt.Sprintf("provider '%s' is not configured: %s", e.Provider, e.Reason)
}

// ErrProviderUnavailable indicates a provider is temporarily unavailable
type ErrProviderUnavailable struct {
	Provider string
	Reason   string
}

func (e *ErrProviderUnavailable) Error() string {
	return fmt.Sprintf("provider '%s' is unavailable: %s", e.Provider, e.Reason)
}

// ErrRateLimitExceeded indicates rate limit has been exceeded
type ErrRateLimitExceeded struct {
	Provider  string
	Limit     string
	RetryAfter string
}

func (e *ErrRateLimitExceeded) Error() string {
	msg := fmt.Sprintf("rate limit exceeded for provider '%s': %s", e.Provider, e.Limit)
	if e.RetryAfter != "" {
		msg += fmt.Sprintf(", retry after %s", e.RetryAfter)
	}
	return msg
}

// ErrInvalidRequest indicates an invalid request
type ErrInvalidRequest struct {
	Field   string
	Reason  string
}

func (e *ErrInvalidRequest) Error() string {
	return fmt.Sprintf("invalid request field '%s': %s", e.Field, e.Reason)
}

// ErrInvalidResponse indicates an invalid provider response
type ErrInvalidResponse struct {
	Provider string
	Reason   string
}

func (e *ErrInvalidResponse) Error() string {
	return fmt.Sprintf("invalid response from provider '%s': %s", e.Provider, e.Reason)
}

// ErrAuthenticationFailed indicates authentication failure
type ErrAuthenticationFailed struct {
	Provider string
	Reason   string
}

func (e *ErrAuthenticationFailed) Error() string {
	return fmt.Sprintf("authentication failed for provider '%s': %s", e.Provider, e.Reason)
}

// ErrQuotaExceeded indicates quota has been exceeded
type ErrQuotaExceeded struct {
	Provider string
	Resource string
}

func (e *ErrQuotaExceeded) Error() string {
	return fmt.Sprintf("quota exceeded for provider '%s': %s", e.Provider, e.Resource)
}

// ErrTimeout indicates a request timeout
type ErrTimeout struct {
	Provider string
	Duration string
}

func (e *ErrTimeout) Error() string {
	return fmt.Sprintf("request timeout for provider '%s' after %s", e.Provider, e.Duration)
}

// ErrAllProvidersFailed indicates all providers in a chain failed
type ErrAllProvidersFailed struct {
	LastError error
}

func (e *ErrAllProvidersFailed) Error() string {
	return fmt.Sprintf("all providers failed, last error: %v", e.LastError)
}

func (e *ErrAllProvidersFailed) Unwrap() error {
	return e.LastError
}

// ErrMultipleProviderErrors aggregates multiple provider errors
type ErrMultipleProviderErrors struct {
	Errors []error
}

func (e *ErrMultipleProviderErrors) Error() string {
	messages := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		messages[i] = err.Error()
	}
	return fmt.Sprintf("multiple provider errors: [%s]", strings.Join(messages, "; "))
}

// ErrCacheFailed indicates a cache operation failure
type ErrCacheFailed struct {
	Operation string
	Reason    string
}

func (e *ErrCacheFailed) Error() string {
	return fmt.Sprintf("cache operation '%s' failed: %s", e.Operation, e.Reason)
}

// ErrInvalidConfiguration indicates invalid configuration
type ErrInvalidConfiguration struct {
	Field  string
	Reason string
}

func (e *ErrInvalidConfiguration) Error() string {
	return fmt.Sprintf("invalid configuration for '%s': %s", e.Field, e.Reason)
}

// ErrContextCanceled indicates context was canceled
type ErrContextCanceled struct {
	Provider string
}

func (e *ErrContextCanceled) Error() string {
	return fmt.Sprintf("request canceled for provider '%s'", e.Provider)
}

// ErrTokenLimitExceeded indicates token limit was exceeded
type ErrTokenLimitExceeded struct {
	Requested int
	Limit     int
}

func (e *ErrTokenLimitExceeded) Error() string {
	return fmt.Sprintf("token limit exceeded: requested %d, limit %d", e.Requested, e.Limit)
}

// IsRetryable determines if an error should trigger a retry
func IsRetryable(err error) bool {
	switch err.(type) {
	case *ErrProviderUnavailable,
		*ErrTimeout,
		*ErrRateLimitExceeded:
		return true
	default:
		return false
	}
}

// IsPermanent determines if an error is permanent
func IsPermanent(err error) bool {
	switch err.(type) {
	case *ErrProviderNotSupported,
		*ErrProviderNotConfigured,
		*ErrInvalidRequest,
		*ErrAuthenticationFailed,
		*ErrInvalidConfiguration:
		return true
	default:
		return false
	}
}
