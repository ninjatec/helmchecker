# Helm Chart Checker - Deployment Guide

## Prerequisites

- Kubernetes cluster access
- Helm 3.x installed
- Docker registry access (for custom images)
- GitHub personal access token

## Quick Start

### 1. Configure the Application

Create a `values-override.yaml` file with your specific configuration:

```yaml
# Image configuration
image:
  repository: your-registry/helmchecker
  tag: latest

# Configuration
config:
  git:
    repository: "https://github.com/your-org/your-repo.git"
    username: "helmchecker-bot"
    email: "helmchecker-bot@your-org.com"
    branch: "main"

  github:
    owner: "your-org"
    repo: "your-repo"

  checker:
    dryRun: false
    excludeCharts:
      - "system-chart"
      - "critical-chart"
    commitMessage: "chore: update helm chart %s to version %s"
    pullRequestTitle: "🤖 Update Helm chart %s to version %s"

# Secrets (alternatively, create the secret manually)
secrets:
  githubToken: "ghp_your_github_token_here"

# CronJob schedule (every day at 2 AM UTC)
cronjob:
  schedule: "0 2 * * *"
  timeZone: "UTC"
```

### 2. Deploy the Application

```bash
# Install with custom values
helm install helmchecker ./helm-chart -f values-override.yaml

# Or install with inline values
helm install helmchecker ./helm-chart \
  --set image.repository=your-registry/helmchecker \
  --set image.tag=latest \
  --set config.git.repository=https://github.com/your-org/your-repo.git \
  --set config.github.owner=your-org \
  --set config.github.repo=your-repo \
  --set secrets.githubToken=ghp_your_token_here
```

## Configuration Options

### Git Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.git.repository` | Git repository URL | `""` |
| `config.git.username` | Git username for commits | `"helmchecker"` |
| `config.git.email` | Git email for commits | `"helmchecker@example.com"` |
| `config.git.branch` | Base branch for PRs | `"main"` |

### GitHub Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.github.owner` | GitHub repository owner | `""` |
| `config.github.repo` | GitHub repository name | `""` |

### Checker Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.checker.dryRun` | Only log what would be updated | `false` |
| `config.checker.excludeCharts` | Charts to exclude from checking | `[]` |
| `config.checker.includeCharts` | Charts to include (empty = all) | `[]` |
| `config.checker.checkPrerelease` | Include pre-release versions | `false` |

### CronJob Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `cronjob.schedule` | Cron schedule | `"0 2 * * *"` |
| `cronjob.timeZone` | Time zone | `"UTC"` |
| `cronjob.suspend` | Suspend the cron job | `false` |

### External Secret Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `externalSecret.enabled` | Use external secret instead of built-in | `false` |
| `externalSecret.name` | Name of the external secret | `""` |
| `externalSecret.githubTokenKey` | Key for GitHub token in external secret | `"GITHUB_TOKEN"` |
| `externalSecret.gitTokenKey` | Key for Git token in external secret | `"GIT_TOKEN"` |

## Security Configuration

### Using External Secrets

#### Option 1: External Secrets Operator (Recommended)

If you're using External Secrets Operator or similar tools, you can reference external secrets directly:

```yaml
# Disable built-in secrets
secrets:
  githubToken: ""
  gitToken: ""

# Enable external secret reference
externalSecret:
  enabled: true
  name: "helmchecker-secrets"      # Name of your external secret
  githubTokenKey: "GITHUB_TOKEN"    # Key in the external secret
  gitTokenKey: "GIT_TOKEN"          # Optional: separate git token key
```

Install with external secret:

```bash
helm install helmchecker ./helm-chart \
  --set externalSecret.enabled=true \
  --set externalSecret.name=helmchecker-secrets
```

#### Option 2: Manual Secret Creation

Alternatively, create the secret manually:

```bash
kubectl create secret generic helmchecker-secrets \
  --from-literal=github-token=ghp_your_token_here \
  --from-literal=git-token=ghp_your_token_here
```

Then install without the secrets in values:

```bash
helm install helmchecker ./helm-chart \
  --set secrets.githubToken="" \
  --set secrets.gitToken=""
```

### RBAC

The chart creates appropriate RBAC resources by default. You can add additional rules:

```yaml
rbac:
  create: true
  additionalRules:
    - apiGroups: ["custom.io"]
      resources: ["customresources"]
      verbs: ["get", "list", "watch"]
```

## Monitoring and Troubleshooting

### Check CronJob Status

```bash
# View the CronJob
kubectl get cronjob helmchecker

# View recent jobs
kubectl get jobs -l app.kubernetes.io/name=helmchecker

# View pod logs
kubectl logs -l app.kubernetes.io/name=helmchecker --tail=100
```

### Debug Mode

Enable dry run mode to see what would be updated without making changes:

```bash
helm upgrade helmchecker ./helm-chart \
  --set config.checker.dryRun=true
```

### Manual Job Execution

Create a manual job for testing:

```bash
kubectl create job helmchecker-manual \
  --from=cronjob/helmchecker
```

## Customization

### Custom Docker Image

Build and use your own Docker image:

```bash
# Build the image
make docker-build DOCKER_REGISTRY=your-registry

# Push the image
make docker-push DOCKER_REGISTRY=your-registry

# Install with custom image
helm install helmchecker ./helm-chart \
  --set image.repository=your-registry/helmchecker \
  --set image.tag=latest
```

### Environment-Specific Configuration

Create different values files for different environments:

```bash
# Development
helm install helmchecker-dev ./helm-chart -f values-dev.yaml

# Production
helm install helmchecker-prod ./helm-chart -f values-prod.yaml
```

## Uninstallation

```bash
helm uninstall helmchecker
```

This will remove all resources created by the chart, including the CronJob, ServiceAccount, and RBAC resources.