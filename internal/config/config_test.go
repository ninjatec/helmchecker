package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set some test environment variables
	os.Setenv("KUBERNETES_NAMESPACE", "test-namespace")
	os.Setenv("GIT_REPOSITORY", "https://github.com/test/repo.git")
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("CHECKER_DRY_RUN", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Kubernetes.Namespace != "test-namespace" {
		t.Errorf("Expected namespace 'test-namespace', got '%s'", cfg.Kubernetes.Namespace)
	}

	if cfg.Git.Repository != "https://github.com/test/repo.git" {
		t.Errorf("Expected repository 'https://github.com/test/repo.git', got '%s'", cfg.Git.Repository)
	}

	if cfg.GitHub.Token != "test-token" {
		t.Errorf("Expected token 'test-token', got '%s'", cfg.GitHub.Token)
	}

	if !cfg.Checker.DryRun {
		t.Errorf("Expected dry run to be true")
	}

	// Clean up
	os.Unsetenv("KUBERNETES_NAMESPACE")
	os.Unsetenv("GIT_REPOSITORY")
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("CHECKER_DRY_RUN")
}

func TestGetEnvOrDefault(t *testing.T) {
	// Test with existing env var
	os.Setenv("TEST_VAR", "test-value")
	result := getEnvOrDefault("TEST_VAR", "default")
	if result != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", result)
	}

	// Test with non-existing env var
	result = getEnvOrDefault("NON_EXISTING_VAR", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}

	os.Unsetenv("TEST_VAR")
}

func TestGetBoolEnvOrDefault(t *testing.T) {
	// Test with true
	os.Setenv("TEST_BOOL", "true")
	result := getBoolEnvOrDefault("TEST_BOOL", false)
	if !result {
		t.Errorf("Expected true, got false")
	}

	// Test with false
	os.Setenv("TEST_BOOL", "false")
	result = getBoolEnvOrDefault("TEST_BOOL", true)
	if result {
		t.Errorf("Expected false, got true")
	}

	// Test with invalid value (should return default)
	os.Setenv("TEST_BOOL", "invalid")
	result = getBoolEnvOrDefault("TEST_BOOL", true)
	if !result {
		t.Errorf("Expected true (default), got false")
	}

	// Test with non-existing env var
	result = getBoolEnvOrDefault("NON_EXISTING_BOOL", true)
	if !result {
		t.Errorf("Expected true (default), got false")
	}

	os.Unsetenv("TEST_BOOL")
}