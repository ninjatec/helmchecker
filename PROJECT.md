# Helm Chart Checker Project

## Overview

This project provides an automated solution for checking outdated Helm charts in your Kubernetes cluster and creating pull requests with updates. It runs as a Kubernetes CronJob and integrates with GitHub for automated PR creation.

## Project Structure

```
helmchecker/
├── cmd/
│   └── helmchecker/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── helm/                    # Helm client and operations
│   ├── git/                     # Git operations
│   ├── github/                  # GitHub API integration
│   └── checker/                 # Main checking logic
├── helm-chart/                  # Kubernetes deployment chart
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
├── docs/                        # Documentation
├── examples/                    # Configuration examples
├── .github/workflows/           # CI/CD pipelines
├── Dockerfile                   # Container image definition
├── Makefile                     # Build automation
└── go.mod                       # Go dependencies
```

## Features

- 🔍 **Automated Chart Discovery**: Scans all Helm releases in your cluster
- 📊 **Version Comparison**: Compares current versions with latest available versions
- 🌿 **Git Integration**: Creates branches and commits for each update
- 🔄 **PR Automation**: Automatically creates GitHub pull requests
- ⚙️ **Configurable**: Extensive configuration options via environment variables
- 🔒 **Security**: Runs with minimal RBAC permissions and security contexts
- 📅 **Scheduled**: Runs on configurable cron schedules
- 🚫 **Filtering**: Include/exclude specific charts from checking

## Quick Start

1. **Build and Push Docker Image**:
   ```bash
   make docker-build DOCKER_REGISTRY=your-registry
   make docker-push DOCKER_REGISTRY=your-registry
   ```

2. **Configure Values**:
   ```bash
   cp examples/values-dev.yaml values-local.yaml
   # Edit values-local.yaml with your configuration
   ```

3. **Deploy to Kubernetes**:
   ```bash
   helm install helmchecker ./helm-chart -f values-local.yaml
   ```

## Configuration

Key configuration parameters:

| Category | Parameter | Description |
|----------|-----------|-------------|
| **Git** | `config.git.repository` | Target repository for PRs |
| **GitHub** | `config.github.owner/repo` | GitHub repository details |
| **Schedule** | `cronjob.schedule` | Cron schedule (default: daily at 2 AM) |
| **Security** | `secrets.githubToken` | GitHub API token |
| **Behavior** | `config.checker.dryRun` | Test mode without creating PRs |

## Development

### Local Development

```bash
# Setup
make dev-setup

# Build
make build

# Test
make test

# Run locally
make dev-run
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Lint Helm chart
helm lint ./helm-chart
```

## Security Considerations

- Runs with non-root user (UID 65534)
- Read-only root filesystem
- Minimal RBAC permissions (read-only cluster access)
- Secrets stored separately from configuration
- Security scanning in CI/CD pipeline

## Monitoring

Monitor the application using:

```bash
# Check CronJob status
kubectl get cronjob helmchecker

# View recent job runs
kubectl get jobs -l app.kubernetes.io/name=helmchecker

# Check logs
kubectl logs -l app.kubernetes.io/name=helmchecker --tail=100
```

## Troubleshooting

1. **Enable Dry Run**: Set `config.checker.dryRun: true` to test without creating PRs
2. **Check Permissions**: Ensure GitHub token has repository write permissions
3. **Verify RBAC**: Confirm ServiceAccount has cluster read permissions
4. **Review Logs**: Check pod logs for detailed error information

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.