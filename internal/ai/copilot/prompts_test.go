package copilot

import (
	"strings"
	"testing"
	"time"

	"github.com/marccoxall/helmchecker/internal/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPromptBuilder(t *testing.T) {
	pb := NewPromptBuilder()
	require.NotNil(t, pb)

	// Verify default templates are registered
	expectedTemplates := []string{
		"helm-chart-analysis",
		"kubernetes-validation",
		"gitops-pattern-detection",
		"upgrade-risk-assessment",
		"best-practices-review",
		"compatibility-check",
		"dependency-analysis",
		"security-audit",
	}

	for _, name := range expectedTemplates {
		template, ok := pb.GetTemplate(name)
		assert.True(t, ok, "template %s should be registered", name)
		assert.NotNil(t, template, "template %s should not be nil", name)
		assert.Equal(t, name, template.Name)
		assert.NotEmpty(t, template.SystemPrompt)
		assert.NotEmpty(t, template.UserPromptTemplate)
		assert.Greater(t, template.MaxTokens, 0)
	}
}

func TestPromptBuilder_RegisterTemplate(t *testing.T) {
	pb := NewPromptBuilder()

	customTemplate := &PromptTemplate{
		Name:               "custom-test",
		Description:        "Test template",
		SystemPrompt:       "You are a test assistant",
		UserPromptTemplate: "Test prompt",
		RequiredContext:    []string{},
		MaxTokens:          1000,
		Temperature:        0.5,
	}

	pb.RegisterTemplate(customTemplate)

	retrieved, ok := pb.GetTemplate("custom-test")
	assert.True(t, ok)
	assert.Equal(t, customTemplate.Name, retrieved.Name)
	assert.Equal(t, customTemplate.SystemPrompt, retrieved.SystemPrompt)
}

func TestPromptBuilder_GetTemplate(t *testing.T) {
	pb := NewPromptBuilder()

	t.Run("existing template", func(t *testing.T) {
		template, ok := pb.GetTemplate("helm-chart-analysis")
		assert.True(t, ok)
		assert.NotNil(t, template)
	})

	t.Run("non-existing template", func(t *testing.T) {
		template, ok := pb.GetTemplate("non-existent")
		assert.False(t, ok)
		assert.Nil(t, template)
	})
}

func TestPromptBuilder_Build(t *testing.T) {
	pb := NewPromptBuilder()

	t.Run("successful build with valid context", func(t *testing.T) {
		ctx := &ai.AnalysisContext{
			RepositoryInfo: &ai.RepositoryInfo{
				Owner:      "testorg",
				Name:       "testrepo",
				URL:        "https://github.com/testorg/testrepo",
				Branch:     "main",
				CommitSHA:  "abc123",
				LastUpdate: time.Now(),
			},
			HelmCharts: []ai.HelmChartInfo{
				{
					Name:       "nginx",
					Version:    "1.0.0",
					AppVersion: "1.21.0",
					Path:       "/charts/nginx",
				},
			},
		}

		req, err := pb.Build("helm-chart-analysis", ctx)
		require.NoError(t, err)
		require.NotNil(t, req)

		assert.NotEmpty(t, req.Query)
		assert.Equal(t, ai.AnalysisTypeGeneral, req.Type)
		assert.Equal(t, ctx, req.Context)
		assert.Equal(t, 2000, req.MaxTokens)
		assert.Equal(t, 0.3, req.Temperature)

		// Verify query contains context information
		assert.Contains(t, req.Query, "testorg/testrepo")
		assert.Contains(t, req.Query, "nginx")
	})

	t.Run("build with non-existent template", func(t *testing.T) {
		ctx := &ai.AnalysisContext{}
		req, err := pb.Build("non-existent", ctx)
		assert.Error(t, err)
		assert.Nil(t, req)
		assert.Contains(t, err.Error(), "template not found")
	})

	t.Run("build with nil context", func(t *testing.T) {
		req, err := pb.Build("helm-chart-analysis", nil)
		assert.Error(t, err)
		assert.Nil(t, req)
		assert.Contains(t, err.Error(), "context is nil")
	})

	t.Run("build with missing required context", func(t *testing.T) {
		ctx := &ai.AnalysisContext{
			RepositoryInfo: &ai.RepositoryInfo{
				Owner: "test",
				Name:  "test",
			},
		}

		req, err := pb.Build("helm-chart-analysis", ctx)
		assert.Error(t, err)
		assert.Nil(t, req)
		assert.Contains(t, err.Error(), "helm charts")
	})
}

func TestPromptBuilder_ValidateContext(t *testing.T) {
	pb := NewPromptBuilder()

	tests := []struct {
		name        string
		template    *PromptTemplate
		context     *ai.AnalysisContext
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid context with all requirements",
			template: &PromptTemplate{
				RequiredContext: []string{"repository", "charts", "patterns"},
			},
			context: &ai.AnalysisContext{
				RepositoryInfo: &ai.RepositoryInfo{
					Owner: "test",
					Name:  "test",
				},
				HelmCharts: []ai.HelmChartInfo{
					{Name: "test"},
				},
				DetectedPatterns: []ai.PatternInfo{
					{Type: "flux"},
				},
			},
			shouldError: false,
		},
		{
			name: "missing repository",
			template: &PromptTemplate{
				RequiredContext: []string{"repository"},
			},
			context:     &ai.AnalysisContext{},
			shouldError: true,
			errorMsg:    "repository info",
		},
		{
			name: "missing charts",
			template: &PromptTemplate{
				RequiredContext: []string{"charts"},
			},
			context: &ai.AnalysisContext{
				HelmCharts: []ai.HelmChartInfo{},
			},
			shouldError: true,
			errorMsg:    "helm charts",
		},
		{
			name: "missing patterns",
			template: &PromptTemplate{
				RequiredContext: []string{"patterns"},
			},
			context: &ai.AnalysisContext{
				DetectedPatterns: []ai.PatternInfo{},
			},
			shouldError: true,
			errorMsg:    "detected patterns",
		},
		{
			name:        "nil context",
			template:    &PromptTemplate{},
			context:     nil,
			shouldError: true,
			errorMsg:    "context is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pb.validateContext(tt.template, tt.context)
			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildContextSection(t *testing.T) {
	t.Run("complete context", func(t *testing.T) {
		ctx := &ai.AnalysisContext{
			RepositoryInfo: &ai.RepositoryInfo{
				Owner:      "testorg",
				Name:       "testrepo",
				URL:        "https://github.com/testorg/testrepo",
				Branch:     "main",
				CommitSHA:  "abc123def456",
				LastUpdate: time.Date(2025, 12, 3, 10, 0, 0, 0, time.UTC),
			},
			DetectedPatterns: []ai.PatternInfo{
				{
					Type:       "flux",
					Version:    "2.0.0",
					Path:       "/flux-system",
					Confidence: 0.95,
					Resources:  []string{"gitrepo", "kustomization"},
				},
			},
			HelmCharts: []ai.HelmChartInfo{
				{
					Name:          "nginx",
					Version:       "1.0.0",
					AppVersion:    "1.21.0",
					LatestVersion: "1.2.0",
					Dependencies:  []string{"common"},
				},
				{
					Name:            "postgresql",
					Version:         "10.0.0",
					Deprecated:      true,
					BreakingChanges: []string{"change1", "change2"},
				},
			},
			GitHistory: []ai.CommitInfo{
				{
					SHA:     "abc123def",
					Author:  "John Doe",
					Date:    time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
					Message: "Update nginx chart",
				},
			},
			Constraints: []string{
				"Must maintain 99.9% uptime",
				"Zero downtime deployments required",
			},
		}

		result := buildContextSection(ctx)

		// Verify all sections are present
		assert.Contains(t, result, "## Analysis Context")
		assert.Contains(t, result, "### Repository")
		assert.Contains(t, result, "testorg/testrepo")
		assert.Contains(t, result, "main")
		assert.Contains(t, result, "abc123def456")

		assert.Contains(t, result, "### Detected GitOps Patterns")
		assert.Contains(t, result, "flux")
		assert.Contains(t, result, "v2.0.0")
		assert.Contains(t, result, "95.0%")

		assert.Contains(t, result, "### Helm Charts")
		assert.Contains(t, result, "nginx")
		assert.Contains(t, result, "v1.0.0")
		assert.Contains(t, result, "Latest Version: 1.2.0")
		assert.Contains(t, result, "postgresql")
		assert.Contains(t, result, "DEPRECATED")
		assert.Contains(t, result, "Breaking Changes: 2")

		assert.Contains(t, result, "### Recent Changes")
		assert.Contains(t, result, "abc123de")
		assert.Contains(t, result, "Update nginx chart")

		assert.Contains(t, result, "### Constraints")
		assert.Contains(t, result, "99.9% uptime")
	})

	t.Run("minimal context", func(t *testing.T) {
		ctx := &ai.AnalysisContext{}
		result := buildContextSection(ctx)

		assert.Contains(t, result, "## Analysis Context")
		// Should not crash with empty context
		assert.NotEmpty(t, result)
	})
}

func TestGetAnalysisType(t *testing.T) {
	tests := []struct {
		templateName string
		expectedType ai.AnalysisType
	}{
		{"helm-chart-analysis", ai.AnalysisTypeGeneral},
		{"kubernetes-validation", ai.AnalysisTypeCompatibility},
		{"gitops-pattern-detection", ai.AnalysisTypePatternDetection},
		{"upgrade-risk-assessment", ai.AnalysisTypeRiskAssessment},
		{"best-practices-review", ai.AnalysisTypeRecommendation},
		{"compatibility-check", ai.AnalysisTypeCompatibility},
		{"dependency-analysis", ai.AnalysisTypeImpact},
		{"security-audit", ai.AnalysisTypeRiskAssessment},
		{"unknown-template", ai.AnalysisTypeGeneral},
	}

	for _, tt := range tests {
		t.Run(tt.templateName, func(t *testing.T) {
			result := getAnalysisType(tt.templateName)
			assert.Equal(t, tt.expectedType, result)
		})
	}
}

func TestTemplateDefinitions(t *testing.T) {
	templates := []*PromptTemplate{
		helmChartAnalysisTemplate(),
		kubernetesValidationTemplate(),
		gitOpsPatternDetectionTemplate(),
		upgradeRiskAssessmentTemplate(),
		bestPracticesReviewTemplate(),
		compatibilityCheckTemplate(),
		dependencyAnalysisTemplate(),
		securityAuditTemplate(),
	}

	for _, template := range templates {
		t.Run(template.Name, func(t *testing.T) {
			// Verify basic fields
			assert.NotEmpty(t, template.Name)
			assert.NotEmpty(t, template.Description)
			assert.NotEmpty(t, template.SystemPrompt)
			assert.NotEmpty(t, template.UserPromptTemplate)
			assert.Greater(t, template.MaxTokens, 0)
			assert.GreaterOrEqual(t, template.Temperature, 0.0)
			assert.LessOrEqual(t, template.Temperature, 1.0)

			// Verify system prompt is descriptive
			assert.True(t, len(template.SystemPrompt) > 50,
				"System prompt should be descriptive")

			// Verify user prompt template has structure
			assert.True(t, strings.Contains(template.UserPromptTemplate, "#") ||
				strings.Contains(template.UserPromptTemplate, "##"),
				"User prompt should have markdown structure")

			// Verify reasonable token limits
			assert.LessOrEqual(t, template.MaxTokens, 3000,
				"Token limit should be reasonable")
		})
	}
}

func TestPromptBuilder_BuildWithDifferentTemplates(t *testing.T) {
	pb := NewPromptBuilder()

	// Create a comprehensive context
	ctx := &ai.AnalysisContext{
		RepositoryInfo: &ai.RepositoryInfo{
			Owner:      "testorg",
			Name:       "testrepo",
			URL:        "https://github.com/testorg/testrepo",
			Branch:     "main",
			CommitSHA:  "abc123",
			LastUpdate: time.Now(),
		},
		DetectedPatterns: []ai.PatternInfo{
			{Type: "flux", Version: "2.0.0", Path: "/", Confidence: 0.9},
		},
		HelmCharts: []ai.HelmChartInfo{
			{Name: "test-chart", Version: "1.0.0"},
		},
	}

	templates := []string{
		"helm-chart-analysis",
		"kubernetes-validation",
		"gitops-pattern-detection",
		"upgrade-risk-assessment",
		"best-practices-review",
		"compatibility-check",
		"dependency-analysis",
		"security-audit",
	}

	for _, templateName := range templates {
		t.Run(templateName, func(t *testing.T) {
			req, err := pb.Build(templateName, ctx)
			require.NoError(t, err)
			require.NotNil(t, req)

			assert.NotEmpty(t, req.Query)
			assert.NotEqual(t, ai.AnalysisType(""), req.Type)
			assert.Greater(t, req.MaxTokens, 0)
			assert.GreaterOrEqual(t, req.Temperature, 0.0)
		})
	}
}

func BenchmarkPromptBuilder_Build(b *testing.B) {
	pb := NewPromptBuilder()
	ctx := &ai.AnalysisContext{
		RepositoryInfo: &ai.RepositoryInfo{
			Owner:      "testorg",
			Name:       "testrepo",
			URL:        "https://github.com/testorg/testrepo",
			Branch:     "main",
			CommitSHA:  "abc123",
			LastUpdate: time.Now(),
		},
		HelmCharts: []ai.HelmChartInfo{
			{Name: "test", Version: "1.0.0"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pb.Build("helm-chart-analysis", ctx)
	}
}

func BenchmarkBuildContextSection(b *testing.B) {
	ctx := &ai.AnalysisContext{
		RepositoryInfo: &ai.RepositoryInfo{
			Owner:      "testorg",
			Name:       "testrepo",
			URL:        "https://github.com/testorg/testrepo",
			Branch:     "main",
			CommitSHA:  "abc123",
			LastUpdate: time.Now(),
		},
		DetectedPatterns: []ai.PatternInfo{
			{Type: "flux", Version: "2.0.0", Path: "/"},
		},
		HelmCharts: []ai.HelmChartInfo{
			{Name: "test", Version: "1.0.0"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildContextSection(ctx)
	}
}
