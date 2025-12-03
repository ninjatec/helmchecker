# Copilot Prompt Templates

This directory contains example prompts and templates for GitHub Copilot analysis.

## Template Categories

### 1. Helm Chart Analysis
**Purpose**: Comprehensive analysis of Helm charts
**Template**: `helm-chart-analysis`
**Use Cases**:
- Initial chart review
- Pre-upgrade assessment
- Quality audit

**Example Context**:
```yaml
Repository: myorg/myapp
Charts:
  - nginx-ingress v1.0.0 → v1.2.0
  - postgresql v11.0.0
```

**Expected Output**:
- Chart structure assessment
- Values configuration review
- Template quality evaluation
- Security findings
- Upgrade recommendations

---

### 2. Kubernetes Validation
**Purpose**: Validate Kubernetes manifests
**Template**: `kubernetes-validation`
**Use Cases**:
- API version validation
- Resource specification review
- Security posture check

**Example Context**:
```yaml
Pattern: Flux CD
Resources:
  - Deployments: 5
  - Services: 3
  - Ingress: 2
```

**Expected Output**:
- API version compatibility
- Resource configuration validation
- Security recommendations
- Production readiness assessment

---

### 3. GitOps Pattern Detection
**Purpose**: Analyze GitOps implementation patterns
**Template**: `gitops-pattern-detection`
**Use Cases**:
- Pattern identification
- Configuration quality assessment
- Best practice alignment

**Example Context**:
```yaml
Detected Patterns:
  - Flux CD v2.0.0 (95% confidence)
  - Kustomize overlays (90% confidence)
```

**Expected Output**:
- Pattern validation
- Configuration quality score
- Best practice recommendations
- Integration suggestions

---

### 4. Upgrade Risk Assessment
**Purpose**: Assess risks of chart upgrades
**Template**: `upgrade-risk-assessment`
**Use Cases**:
- Pre-upgrade planning
- Risk mitigation strategy
- Stakeholder communication

**Example Context**:
```yaml
Upgrade:
  From: nginx-ingress v1.0.0
  To: nginx-ingress v1.2.0
  Breaking Changes: 2
  Deprecated APIs: 1
```

**Expected Output**:
- Risk level (Low/Medium/High/Critical)
- Impact analysis
- Mitigation strategies
- Testing recommendations
- Rollback plan

---

### 5. Best Practices Review
**Purpose**: Review against industry best practices
**Template**: `best-practices-review`
**Use Cases**:
- Security audit
- Performance optimization
- Operational excellence

**Example Context**:
```yaml
Resources:
  - Deployments without resource limits: 3
  - Services without monitoring: 2
  - Missing health checks: 4
```

**Expected Output**:
- Best practice compliance score
- Security recommendations
- Reliability improvements
- Performance optimizations
- Cost savings opportunities

---

### 6. Compatibility Check
**Purpose**: Check version compatibility
**Template**: `compatibility-check`
**Use Cases**:
- Kubernetes version upgrades
- Chart version compatibility
- API deprecation planning

**Example Context**:
```yaml
Current K8s: v1.24
Target K8s: v1.28
Charts using deprecated APIs:
  - Ingress: networking.k8s.io/v1beta1
  - PodSecurityPolicy: policy/v1beta1
```

**Expected Output**:
- Compatibility matrix
- Required migrations
- Breaking changes
- Migration guide

---

### 7. Dependency Analysis
**Purpose**: Analyze chart dependencies
**Template**: `dependency-analysis`
**Use Cases**:
- Dependency tree mapping
- Version conflict resolution
- Security vulnerability scanning

**Example Context**:
```yaml
Charts:
  - app-chart v1.0.0
    Dependencies:
      - postgresql v11.0.0
      - redis v16.0.0
      - common v2.0.0
```

**Expected Output**:
- Dependency graph
- Version conflicts
- Security vulnerabilities
- Update recommendations

---

### 8. Security Audit
**Purpose**: Comprehensive security review
**Template**: `security-audit`
**Use Cases**:
- Security compliance
- Vulnerability assessment
- Hardening recommendations

**Example Context**:
```yaml
Security Findings:
  - Containers running as root: 2
  - Missing network policies: 5
  - Secrets in plain text: 1
  - No resource limits: 3
```

**Expected Output**:
- Security posture rating
- Critical vulnerabilities
- Compliance status
- Remediation priorities
- Security roadmap

---

## Using Templates

### Via Go API:
```go
import "github.com/marccoxall/helmchecker/internal/ai/copilot"

pb := copilot.NewPromptBuilder()
req, err := pb.Build("helm-chart-analysis", ctx)
if err != nil {
    log.Fatal(err)
}

provider := copilot.NewCopilotProvider(config, tokenProvider)
response, err := provider.Analyze(context.Background(), req)
```

### Customizing Templates:
```go
pb := copilot.NewPromptBuilder()

customTemplate := &copilot.PromptTemplate{
    Name: "custom-analysis",
    Description: "Custom analysis template",
    SystemPrompt: "You are an expert...",
    UserPromptTemplate: "Analyze the following...",
    RequiredContext: []string{"charts"},
    MaxTokens: 2000,
    Temperature: 0.3,
}

pb.RegisterTemplate(customTemplate)
```

## Template Best Practices

1. **Keep prompts focused**: Each template should have a clear, specific purpose
2. **Provide context**: Include relevant repository and configuration details
3. **Structure output**: Request structured, actionable output formats
4. **Set appropriate limits**: Use reasonable token limits for each template
5. **Tune temperature**: Lower for factual analysis, higher for creative suggestions
6. **Validate context**: Ensure required context is available before generating prompts
7. **Test thoroughly**: Validate templates with various input scenarios

## Token Optimization

- **Helm Chart Analysis**: 2000 tokens (comprehensive)
- **Kubernetes Validation**: 1500 tokens (focused)
- **GitOps Patterns**: 1800 tokens (moderate)
- **Risk Assessment**: 2500 tokens (detailed)
- **Best Practices**: 2000 tokens (comprehensive)
- **Compatibility**: 1500 tokens (focused)
- **Dependencies**: 1800 tokens (moderate)
- **Security Audit**: 2500 tokens (detailed)

## Example Outputs

See the `examples/` subdirectory for sample outputs from each template type.
