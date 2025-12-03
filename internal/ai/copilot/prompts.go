package copilot

import (
	"fmt"
	"strings"
	"time"

	"github.com/marccoxall/helmchecker/internal/ai"
)

// PromptTemplate defines a structured template for generating prompts
type PromptTemplate struct {
	// Name identifies the template
	Name string

	// Description explains the template's purpose
	Description string

	// SystemPrompt is the system message that sets context
	SystemPrompt string

	// UserPromptTemplate is a template string with placeholders
	UserPromptTemplate string

	// RequiredContext lists required context fields
	RequiredContext []string

	// MaxTokens is the recommended max tokens for this prompt
	MaxTokens int

	// Temperature is the recommended temperature setting
	Temperature float64
}

// PromptBuilder constructs prompts from templates and context
type PromptBuilder struct {
	templates map[string]*PromptTemplate
}

// NewPromptBuilder creates a new prompt builder with default templates
func NewPromptBuilder() *PromptBuilder {
	pb := &PromptBuilder{
		templates: make(map[string]*PromptTemplate),
	}

	// Register default templates
	pb.registerDefaultTemplates()

	return pb
}

// registerDefaultTemplates registers all built-in prompt templates
func (pb *PromptBuilder) registerDefaultTemplates() {
	pb.RegisterTemplate(helmChartAnalysisTemplate())
	pb.RegisterTemplate(kubernetesValidationTemplate())
	pb.RegisterTemplate(gitOpsPatternDetectionTemplate())
	pb.RegisterTemplate(upgradeRiskAssessmentTemplate())
	pb.RegisterTemplate(bestPracticesReviewTemplate())
	pb.RegisterTemplate(compatibilityCheckTemplate())
	pb.RegisterTemplate(dependencyAnalysisTemplate())
	pb.RegisterTemplate(securityAuditTemplate())
}

// RegisterTemplate adds a new template to the builder
func (pb *PromptBuilder) RegisterTemplate(template *PromptTemplate) {
	pb.templates[template.Name] = template
}

// GetTemplate retrieves a template by name
func (pb *PromptBuilder) GetTemplate(name string) (*PromptTemplate, bool) {
	template, ok := pb.templates[name]
	return template, ok
}

// Build constructs a complete AI request from a template and context
func (pb *PromptBuilder) Build(templateName string, ctx *ai.AnalysisContext) (*ai.Request, error) {
	template, ok := pb.GetTemplate(templateName)
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}

	// Validate required context
	if err := pb.validateContext(template, ctx); err != nil {
		return nil, fmt.Errorf("context validation failed: %w", err)
	}

	// Build the user prompt from template and context
	userPrompt := pb.buildUserPrompt(template, ctx)

	// Create the AI request
	req := &ai.Request{
		Query:       userPrompt,
		Type:        getAnalysisType(templateName),
		Context:     ctx,
		MaxTokens:   template.MaxTokens,
		Temperature: template.Temperature,
		Options: ai.RequestOptions{
			ResponseFormat: "markdown",
		},
	}

	return req, nil
}

// validateContext checks if all required context is present
func (pb *PromptBuilder) validateContext(template *PromptTemplate, ctx *ai.AnalysisContext) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}

	for _, required := range template.RequiredContext {
		switch required {
		case "repository":
			if ctx.RepositoryInfo == nil {
				return fmt.Errorf("missing required context: repository info")
			}
		case "patterns":
			if len(ctx.DetectedPatterns) == 0 {
				return fmt.Errorf("missing required context: detected patterns")
			}
		case "charts":
			if len(ctx.HelmCharts) == 0 {
				return fmt.Errorf("missing required context: helm charts")
			}
		}
	}

	return nil
}

// buildUserPrompt constructs the user prompt from template and context
func (pb *PromptBuilder) buildUserPrompt(template *PromptTemplate, ctx *ai.AnalysisContext) string {
	var buf strings.Builder

	// Start with template
	buf.WriteString(template.UserPromptTemplate)
	buf.WriteString("\n\n")

	// Add context sections
	buf.WriteString(buildContextSection(ctx))

	return buf.String()
}

// buildContextSection creates formatted context information
func buildContextSection(ctx *ai.AnalysisContext) string {
	var buf strings.Builder

	buf.WriteString("## Analysis Context\n\n")

	// Repository information
	if ctx.RepositoryInfo != nil {
		buf.WriteString("### Repository\n")
		buf.WriteString(fmt.Sprintf("- **Name**: %s/%s\n", ctx.RepositoryInfo.Owner, ctx.RepositoryInfo.Name))
		buf.WriteString(fmt.Sprintf("- **URL**: %s\n", ctx.RepositoryInfo.URL))
		buf.WriteString(fmt.Sprintf("- **Branch**: %s\n", ctx.RepositoryInfo.Branch))
		buf.WriteString(fmt.Sprintf("- **Commit**: %s\n", ctx.RepositoryInfo.CommitSHA))
		buf.WriteString(fmt.Sprintf("- **Last Updated**: %s\n\n", ctx.RepositoryInfo.LastUpdate.Format(time.RFC3339)))
	}

	// Detected patterns
	if len(ctx.DetectedPatterns) > 0 {
		buf.WriteString("### Detected GitOps Patterns\n")
		for _, pattern := range ctx.DetectedPatterns {
			buf.WriteString(fmt.Sprintf("- **%s** (v%s)\n", pattern.Type, pattern.Version))
			buf.WriteString(fmt.Sprintf("  - Path: `%s`\n", pattern.Path))
			buf.WriteString(fmt.Sprintf("  - Confidence: %.1f%%\n", pattern.Confidence*100))
			if len(pattern.Resources) > 0 {
				buf.WriteString(fmt.Sprintf("  - Resources: %d\n", len(pattern.Resources)))
			}
		}
		buf.WriteString("\n")
	}

	// Helm charts
	if len(ctx.HelmCharts) > 0 {
		buf.WriteString("### Helm Charts\n")
		for _, chart := range ctx.HelmCharts {
			buf.WriteString(fmt.Sprintf("- **%s** (v%s)\n", chart.Name, chart.Version))
			if chart.AppVersion != "" {
				buf.WriteString(fmt.Sprintf("  - App Version: %s\n", chart.AppVersion))
			}
			if chart.LatestVersion != "" && chart.LatestVersion != chart.Version {
				buf.WriteString(fmt.Sprintf("  - Latest Version: %s ⚠️\n", chart.LatestVersion))
			}
			if len(chart.Dependencies) > 0 {
				buf.WriteString(fmt.Sprintf("  - Dependencies: %d\n", len(chart.Dependencies)))
			}
			if chart.Deprecated {
				buf.WriteString("  - ⚠️ **DEPRECATED**\n")
			}
			if len(chart.BreakingChanges) > 0 {
				buf.WriteString(fmt.Sprintf("  - ⚠️ Breaking Changes: %d\n", len(chart.BreakingChanges)))
			}
		}
		buf.WriteString("\n")
	}

	// Git history (if available)
	if len(ctx.GitHistory) > 0 {
		buf.WriteString("### Recent Changes\n")
		count := len(ctx.GitHistory)
		if count > 5 {
			count = 5 // Show only last 5 commits
		}
		for i := 0; i < count; i++ {
			commit := ctx.GitHistory[i]
			buf.WriteString(fmt.Sprintf("- `%s` - %s (%s)\n",
				commit.SHA[:8],
				commit.Message,
				commit.Date.Format("2006-01-02")))
		}
		buf.WriteString("\n")
	}

	// Current and target state comparison
	if ctx.CurrentState != nil || ctx.TargetState != nil {
		buf.WriteString("### State Information\n")
		if ctx.CurrentState != nil {
			buf.WriteString("- Current state provided\n")
		}
		if ctx.TargetState != nil {
			buf.WriteString("- Target state provided\n")
		}
		buf.WriteString("\n")
	}

	// Constraints
	if len(ctx.Constraints) > 0 {
		buf.WriteString("### Constraints\n")
		for _, constraint := range ctx.Constraints {
			buf.WriteString(fmt.Sprintf("- %s\n", constraint))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

// getAnalysisType maps template name to analysis type
func getAnalysisType(templateName string) ai.AnalysisType {
	mapping := map[string]ai.AnalysisType{
		"helm-chart-analysis":       ai.AnalysisTypeGeneral,
		"kubernetes-validation":     ai.AnalysisTypeCompatibility,
		"gitops-pattern-detection":  ai.AnalysisTypePatternDetection,
		"upgrade-risk-assessment":   ai.AnalysisTypeRiskAssessment,
		"best-practices-review":     ai.AnalysisTypeRecommendation,
		"compatibility-check":       ai.AnalysisTypeCompatibility,
		"dependency-analysis":       ai.AnalysisTypeImpact,
		"security-audit":            ai.AnalysisTypeRiskAssessment,
	}

	if analysisType, ok := mapping[templateName]; ok {
		return analysisType
	}

	return ai.AnalysisTypeGeneral
}

// Template definitions

func helmChartAnalysisTemplate() *PromptTemplate {
	return &PromptTemplate{
		Name:        "helm-chart-analysis",
		Description: "Analyzes Helm charts for issues and improvements",
		SystemPrompt: "You are an expert DevOps engineer specializing in Helm charts and Kubernetes deployments. " +
			"Provide detailed analysis of Helm charts, identifying potential issues, suggesting improvements, " +
			"and recommending best practices.",
		UserPromptTemplate: `# Helm Chart Analysis Request

Please analyze the provided Helm chart(s) and provide a comprehensive assessment including:

## Analysis Requirements

1. **Chart Structure**: Evaluate the overall structure and organization
2. **Values Configuration**: Review values.yaml for completeness and best practices
3. **Template Quality**: Assess template files for correctness and maintainability
4. **Dependencies**: Analyze chart dependencies and versions
5. **Resource Definitions**: Check Kubernetes resource definitions
6. **Security**: Identify any security concerns
7. **Upgrade Path**: Assess upgrade compatibility and risks

## Output Format

Please structure your response as follows:

### Summary
- Overall assessment (Good/Fair/Needs Improvement)
- Key findings (3-5 bullet points)

### Detailed Analysis
- Issues found (if any)
- Recommendations
- Best practice violations

### Upgrade Considerations
- Breaking changes
- Required actions
- Risk level (Low/Medium/High)

### Action Items
Prioritized list of recommended actions`,
		RequiredContext: []string{"charts"},
		MaxTokens:       2000,
		Temperature:     0.3,
	}
}

func kubernetesValidationTemplate() *PromptTemplate {
	return &PromptTemplate{
		Name:        "kubernetes-validation",
		Description: "Validates Kubernetes manifests for correctness and best practices",
		SystemPrompt: "You are a Kubernetes expert with deep knowledge of API versions, resource specifications, " +
			"and deployment best practices. Validate manifests for correctness, security, and production-readiness.",
		UserPromptTemplate: `# Kubernetes Manifest Validation Request

Please validate the Kubernetes manifests and provide feedback on:

## Validation Checks

1. **API Versions**: Check for deprecated or removed API versions
2. **Resource Specifications**: Validate resource requests and limits
3. **Security Context**: Review security settings and policies
4. **Networking**: Evaluate service and ingress configurations
5. **Storage**: Assess persistent volume configurations
6. **High Availability**: Check for HA configurations
7. **Labels and Annotations**: Verify proper labeling

## Output Format

### Validation Results
- ✅ Passed checks
- ⚠️ Warnings
- ❌ Errors

### Critical Issues
List any blocking issues that must be addressed

### Recommendations
Suggested improvements for production deployment`,
		RequiredContext: []string{"patterns"},
		MaxTokens:       1500,
		Temperature:     0.2,
	}
}

func gitOpsPatternDetectionTemplate() *PromptTemplate {
	return &PromptTemplate{
		Name:        "gitops-pattern-detection",
		Description: "Detects and analyzes GitOps patterns in repository",
		SystemPrompt: "You are a GitOps expert familiar with Flux CD, ArgoCD, and Kustomize patterns. " +
			"Analyze repositories to identify GitOps patterns and provide recommendations.",
		UserPromptTemplate: `# GitOps Pattern Analysis Request

Please analyze the detected GitOps patterns and provide insights on:

## Analysis Areas

1. **Pattern Identification**: Confirm detected patterns and identify any missed patterns
2. **Configuration Quality**: Assess the quality of GitOps configurations
3. **Best Practices**: Compare against GitOps best practices
4. **Integration**: Evaluate how patterns work together
5. **Automation**: Assess level of automation and CI/CD integration

## Output Format

### Pattern Summary
Overview of detected patterns and their purposes

### Configuration Review
Assessment of current configurations

### Best Practice Alignment
How well the setup follows GitOps principles

### Recommendations
- Improvements to pattern usage
- Additional patterns to consider
- Configuration optimizations`,
		RequiredContext: []string{"repository", "patterns"},
		MaxTokens:       1800,
		Temperature:     0.3,
	}
}

func upgradeRiskAssessmentTemplate() *PromptTemplate {
	return &PromptTemplate{
		Name:        "upgrade-risk-assessment",
		Description: "Assesses risks associated with Helm chart upgrades",
		SystemPrompt: "You are a DevOps expert specializing in change management and risk assessment for Kubernetes deployments. " +
			"Provide detailed risk analysis for upgrade operations.",
		UserPromptTemplate: `# Upgrade Risk Assessment Request

Please assess the risks associated with upgrading the identified Helm charts:

## Risk Assessment Areas

1. **Breaking Changes**: Identify breaking changes between versions
2. **Dependency Impact**: Analyze impact on dependent systems
3. **Data Migration**: Assess data migration requirements
4. **Rollback Complexity**: Evaluate ease of rollback
5. **Downtime Risk**: Estimate potential downtime
6. **Resource Requirements**: Check for new resource requirements
7. **Configuration Changes**: Identify required configuration updates

## Risk Scoring

For each identified risk, provide:
- **Severity**: Low / Medium / High / Critical
- **Likelihood**: Low / Medium / High
- **Impact**: Description of potential impact
- **Mitigation**: Steps to reduce risk

## Output Format

### Executive Summary
- Overall risk level
- Top 3 concerns
- Recommended approach

### Detailed Risk Analysis
Per-chart risk breakdown

### Mitigation Strategy
Recommended actions to reduce risk

### Testing Plan
Suggested testing approach before production deployment`,
		RequiredContext: []string{"charts"},
		MaxTokens:       2500,
		Temperature:     0.2,
	}
}

func bestPracticesReviewTemplate() *PromptTemplate {
	return &PromptTemplate{
		Name:        "best-practices-review",
		Description: "Reviews configuration against Kubernetes and Helm best practices",
		SystemPrompt: "You are a Kubernetes and Helm expert with extensive experience in production deployments. " +
			"Review configurations against industry best practices and provide actionable recommendations.",
		UserPromptTemplate: `# Best Practices Review Request

Please review the configuration against Kubernetes and Helm best practices:

## Review Areas

1. **Security**: Security best practices and hardening
2. **Reliability**: High availability and fault tolerance
3. **Performance**: Resource optimization and efficiency
4. **Observability**: Logging, monitoring, and tracing
5. **Maintainability**: Code organization and documentation
6. **Scalability**: Scaling capabilities and limits
7. **Cost Optimization**: Resource efficiency and cost management

## Output Format

### Best Practices Score
- Overall score (0-100)
- Category breakdown

### Adherence to Standards
- ✅ Following best practices
- ⚠️ Partial compliance
- ❌ Not following best practices

### Priority Recommendations
Top 5 improvements ordered by impact

### Implementation Guide
Step-by-step guide for key recommendations`,
		RequiredContext: []string{},
		MaxTokens:       2000,
		Temperature:     0.3,
	}
}

func compatibilityCheckTemplate() *PromptTemplate {
	return &PromptTemplate{
		Name:        "compatibility-check",
		Description: "Checks compatibility between chart versions and Kubernetes versions",
		SystemPrompt: "You are an expert in Kubernetes API versions and backward compatibility. " +
			"Analyze compatibility issues and provide migration guidance.",
		UserPromptTemplate: `# Compatibility Check Request

Please analyze compatibility between the Helm charts and target Kubernetes version:

## Compatibility Analysis

1. **API Version Compatibility**: Check for deprecated or removed APIs
2. **Feature Compatibility**: Verify feature availability
3. **Version Matrix**: Analyze version compatibility matrix
4. **Migration Path**: Identify required migrations
5. **Testing Requirements**: Determine necessary compatibility tests

## Output Format

### Compatibility Summary
- Compatible: Yes/No/Partial
- Blocking issues count
- Warning count

### Detailed Findings
Per-resource compatibility analysis

### Migration Requirements
- API version updates needed
- Configuration changes required
- Feature replacements

### Testing Strategy
Recommended compatibility testing approach`,
		RequiredContext: []string{"charts"},
		MaxTokens:       1500,
		Temperature:     0.2,
	}
}

func dependencyAnalysisTemplate() *PromptTemplate {
	return &PromptTemplate{
		Name:        "dependency-analysis",
		Description: "Analyzes chart dependencies and their impacts",
		SystemPrompt: "You are an expert in Helm chart dependencies and microservices architecture. " +
			"Analyze dependency chains and identify potential issues.",
		UserPromptTemplate: `# Dependency Analysis Request

Please analyze the Helm chart dependencies:

## Analysis Requirements

1. **Dependency Tree**: Map out the complete dependency tree
2. **Version Conflicts**: Identify version conflicts or incompatibilities
3. **Update Impact**: Assess impact of dependency updates
4. **Circular Dependencies**: Check for circular dependencies
5. **Transitive Dependencies**: Analyze indirect dependencies
6. **Security**: Check for known vulnerabilities in dependencies

## Output Format

### Dependency Graph
Visual representation of dependencies

### Identified Issues
- Conflicts
- Vulnerabilities
- Version mismatches

### Update Recommendations
Suggested dependency updates with rationale

### Impact Assessment
Impact of recommended changes on the system`,
		RequiredContext: []string{"charts"},
		MaxTokens:       1800,
		Temperature:     0.3,
	}
}

func securityAuditTemplate() *PromptTemplate {
	return &PromptTemplate{
		Name:        "security-audit",
		Description: "Performs security audit of Kubernetes configurations",
		SystemPrompt: "You are a Kubernetes security expert with deep knowledge of container security, " +
			"network policies, and security best practices. Perform thorough security audits.",
		UserPromptTemplate: `# Security Audit Request

Please perform a comprehensive security audit:

## Security Review Areas

1. **Container Security**: Image security, scanning, and policies
2. **RBAC**: Role-based access control configuration
3. **Network Policies**: Network segmentation and policies
4. **Secrets Management**: Secret handling and storage
5. **Pod Security**: Pod security standards and policies
6. **Resource Limits**: Resource quotas and limits
7. **Compliance**: Regulatory compliance considerations

## Output Format

### Security Posture
- Overall security rating
- Critical issues count
- High priority items

### Vulnerability Assessment
Identified security vulnerabilities

### Compliance Check
Alignment with security standards (CIS, PCI-DSS, etc.)

### Remediation Plan
Prioritized security improvements with implementation steps

### Security Recommendations
Long-term security enhancements`,
		RequiredContext: []string{},
		MaxTokens:       2500,
		Temperature:     0.2,
	}
}
