package copilot

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var (
	// ErrNoToken is returned when no GitHub token is provided
	ErrNoToken = errors.New("no GitHub token provided")

	// ErrInvalidToken is returned when the token format is invalid
	ErrInvalidToken = errors.New("invalid GitHub token format")
)

// TokenProvider defines an interface for providing authentication tokens
type TokenProvider interface {
	// GetToken returns the authentication token
	GetToken() (string, error)

	// ValidateToken validates the token format and optionally checks with the API
	ValidateToken() error
}

// StaticTokenProvider provides a static token
type StaticTokenProvider struct {
	token string
}

// NewStaticTokenProvider creates a new static token provider
func NewStaticTokenProvider(token string) *StaticTokenProvider {
	return &StaticTokenProvider{token: token}
}

// GetToken returns the static token
func (p *StaticTokenProvider) GetToken() (string, error) {
	if p.token == "" {
		return "", ErrNoToken
	}
	return p.token, nil
}

// ValidateToken validates the token format
func (p *StaticTokenProvider) ValidateToken() error {
	if p.token == "" {
		return ErrNoToken
	}

	// GitHub tokens typically start with specific prefixes
	// ghp_ for personal access tokens
	// ghu_ for user tokens
	// ghs_ for server-to-server tokens
	if !strings.HasPrefix(p.token, "ghp_") &&
		!strings.HasPrefix(p.token, "ghu_") &&
		!strings.HasPrefix(p.token, "ghs_") &&
		!strings.HasPrefix(p.token, "github_pat_") {
		// For development/testing, allow other formats
		if len(p.token) < 10 {
			return ErrInvalidToken
		}
	}

	return nil
}

// EnvTokenProvider retrieves tokens from environment variables
type EnvTokenProvider struct {
	envVar string
}

// NewEnvTokenProvider creates a new environment token provider
func NewEnvTokenProvider(envVar string) *EnvTokenProvider {
	if envVar == "" {
		envVar = "GITHUB_TOKEN"
	}
	return &EnvTokenProvider{envVar: envVar}
}

// GetToken retrieves the token from the environment
func (p *EnvTokenProvider) GetToken() (string, error) {
	token := os.Getenv(p.envVar)
	if token == "" {
		return "", fmt.Errorf("%w: environment variable %s not set", ErrNoToken, p.envVar)
	}
	return token, nil
}

// ValidateToken validates the token from the environment
func (p *EnvTokenProvider) ValidateToken() error {
	token, err := p.GetToken()
	if err != nil {
		return err
	}

	provider := NewStaticTokenProvider(token)
	return provider.ValidateToken()
}

// AuthTransport wraps an http.RoundTripper to add authentication
type AuthTransport struct {
	// Transport is the underlying HTTP transport
	Transport http.RoundTripper

	// TokenProvider provides the authentication token
	TokenProvider TokenProvider
}

// RoundTrip implements the http.RoundTripper interface
func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqCopy := req.Clone(req.Context())

	// Get the token
	token, err := t.TokenProvider.GetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get authentication token: %w", err)
	}

	// Add authorization header
	reqCopy.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Add other required headers
	reqCopy.Header.Set("Content-Type", "application/json")
	reqCopy.Header.Set("Accept", "application/json")
	reqCopy.Header.Set("User-Agent", "HelmChecker/1.0")

	// Use the underlying transport
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	return transport.RoundTrip(reqCopy)
}

// NewAuthenticatedClient creates an HTTP client with authentication
func NewAuthenticatedClient(tokenProvider TokenProvider) *http.Client {
	return &http.Client{
		Transport: &AuthTransport{
			Transport:     http.DefaultTransport,
			TokenProvider: tokenProvider,
		},
	}
}

// ValidateGitHubToken validates a token by making a test API request
func ValidateGitHubToken(client *http.Client, token string) error {
	// Create a simple request to validate the token
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid or expired token")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
