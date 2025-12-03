package openai

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var (
	// ErrNoAPIKey is returned when no API key is provided
	ErrNoAPIKey = errors.New("no OpenAI API key provided")

	// ErrInvalidAPIKey is returned when the API key format is invalid
	ErrInvalidAPIKey = errors.New("invalid OpenAI API key format")
)

// ApiKeyProvider defines an interface for providing API keys
type ApiKeyProvider interface {
	// GetAPIKey returns the API key
	GetAPIKey() (string, error)

	// ValidateAPIKey validates the API key format
	ValidateAPIKey() error
}

// StaticApiKeyProvider provides a static API key
type StaticApiKeyProvider struct {
	apiKey string
}

// NewStaticApiKeyProvider creates a new static API key provider
func NewStaticApiKeyProvider(apiKey string) *StaticApiKeyProvider {
	return &StaticApiKeyProvider{apiKey: apiKey}
}

// GetAPIKey returns the static API key
func (p *StaticApiKeyProvider) GetAPIKey() (string, error) {
	if p.apiKey == "" {
		return "", ErrNoAPIKey
	}
	return p.apiKey, nil
}

// ValidateAPIKey validates the API key format
func (p *StaticApiKeyProvider) ValidateAPIKey() error {
	if p.apiKey == "" {
		return ErrNoAPIKey
	}

	// OpenAI API keys typically start with "sk-"
	if !strings.HasPrefix(p.apiKey, "sk-") {
		// For development/testing, allow other formats if they're long enough
		if len(p.apiKey) < 20 {
			return ErrInvalidAPIKey
		}
	}

	return nil
}

// EnvApiKeyProvider retrieves API keys from environment variables
type EnvApiKeyProvider struct {
	envVar string
}

// NewEnvApiKeyProvider creates a new environment API key provider
func NewEnvApiKeyProvider(envVar string) *EnvApiKeyProvider {
	if envVar == "" {
		envVar = "OPENAI_API_KEY"
	}
	return &EnvApiKeyProvider{envVar: envVar}
}

// GetAPIKey retrieves the API key from the environment
func (p *EnvApiKeyProvider) GetAPIKey() (string, error) {
	apiKey := os.Getenv(p.envVar)
	if apiKey == "" {
		return "", fmt.Errorf("%w: environment variable %s not set", ErrNoAPIKey, p.envVar)
	}
	return apiKey, nil
}

// ValidateAPIKey validates the API key from the environment
func (p *EnvApiKeyProvider) ValidateAPIKey() error {
	apiKey, err := p.GetAPIKey()
	if err != nil {
		return err
	}

	provider := NewStaticApiKeyProvider(apiKey)
	return provider.ValidateAPIKey()
}

// AuthTransport wraps an http.RoundTripper to add authentication
type AuthTransport struct {
	// Transport is the underlying HTTP transport
	Transport http.RoundTripper

	// ApiKeyProvider provides the API key
	ApiKeyProvider ApiKeyProvider

	// Organization is the optional organization ID
	Organization string
}

// RoundTrip implements the http.RoundTripper interface
func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqCopy := req.Clone(req.Context())

	// Get the API key
	apiKey, err := t.ApiKeyProvider.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	// Add authorization header
	reqCopy.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Add content type and accept headers
	reqCopy.Header.Set("Content-Type", "application/json")
	reqCopy.Header.Set("Accept", "application/json")

	// Add organization header if provided
	if t.Organization != "" {
		reqCopy.Header.Set("OpenAI-Organization", t.Organization)
	}

	// Add user agent
	reqCopy.Header.Set("User-Agent", "HelmChecker/1.0")

	// Use the underlying transport
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	return transport.RoundTrip(reqCopy)
}

// NewAuthenticatedClient creates an HTTP client with authentication
func NewAuthenticatedClient(apiKeyProvider ApiKeyProvider, organization string) *http.Client {
	return &http.Client{
		Transport: &AuthTransport{
			Transport:      http.DefaultTransport,
			ApiKeyProvider: apiKeyProvider,
			Organization:   organization,
		},
	}
}

// ValidateApiKey validates an API key by making a test API request
func ValidateApiKey(client *http.Client, apiKey string) error {
	// Create a simple request to validate the API key
	req, err := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate API key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid or expired API key")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
