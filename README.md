# Helm Chart Checker

A Kubernetes application that automatically checks for outdated Helm charts and creates pull requests with updates.

## Features

- Checks currently installed Helm charts for newer versions
- Compares with upstream chart repositories
- Creates Git branches and pull requests for chart updates
- Runs as a Kubernetes CronJob
- Configurable check intervals
- Support for multiple Helm repositories

## Architecture

- **Go Application**: Core logic for chart checking and Git operations
- **Helm Chart**: Kubernetes deployment configuration
- **CronJob**: Scheduled execution
- **RBAC**: Proper permissions for cluster access

## Quick Start

### Option 1: Interactive Setup (Recommended)

Use the interactive setup script to configure your environment:

```bash
./scripts/setup.sh
```

This will guide you through configuring all necessary environment variables and create a `.env.helmchecker` file.

### Option 2: Manual Configuration

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your configuration:
   ```bash
   # Required: Git repository where charts are stored
   export GIT_REPOSITORY="https://github.com/your-org/helm-charts.git"
   
   # Required: GitHub token with repo permissions
   export GITHUB_TOKEN="ghp_your_github_token_here"
   export GITHUB_OWNER="your-org"
   export GITHUB_REPO="helm-charts"
   
   # Optional: Customize behavior
   export CHECKER_DRY_RUN="true"  # Set false to create actual PRs
   ```

3. Source the configuration and run:
   ```bash
   source .env
   ./bin/helmchecker
   ```

### Option 3: Docker/Kubernetes Deployment

1. Build and push the Docker image:
   ```bash
   docker build -t your-registry/helmchecker:latest .
   docker push your-registry/helmchecker:latest
   ```

2. Install the Helm chart:
   ```bash
   helm install helmchecker ./helm-chart \
     --set image.repository=your-registry/helmchecker \
     --set image.tag=latest \
     --set secrets.githubToken=your-github-token \
     --set config.git.repository=https://github.com/your-org/helm-charts.git \
     --set config.github.owner=your-org \
     --set config.github.repo=helm-charts
   ```

   Or use with External Secrets Operator:
   ```bash
   helm install helmchecker ./helm-chart \
     --set image.repository=your-registry/helmchecker \
     --set image.tag=latest \
     --set externalSecret.enabled=true \
     --set externalSecret.name=helmchecker-secrets
   ```

## Configuration

### Required Environment Variables

- `GIT_REPOSITORY`: URL of the Git repository containing your Helm charts
- `GITHUB_TOKEN`: GitHub Personal Access Token with `repo` permissions
- `GITHUB_OWNER`: GitHub organization or user name
- `GITHUB_REPO`: GitHub repository name

### Optional Environment Variables

- `GIT_TOKEN`: Git authentication token (defaults to `GITHUB_TOKEN`)
- `GIT_USERNAME`: Git username for commits (default: "helmchecker")
- `GIT_EMAIL`: Git email for commits (default: "helmchecker@example.com")
- `GIT_BRANCH`: Target branch for pull requests (default: "main")
- `KUBERNETES_NAMESPACE`: Kubernetes namespace to check (default: all namespaces)
- `CHECKER_DRY_RUN`: Enable dry-run mode (default: false)
- `CHECKER_CHECK_PRERELEASE`: Include prerelease versions (default: false)

## Troubleshooting

If you encounter issues, check the [troubleshooting guide](docs/TROUBLESHOOTING.md) for common problems and solutions:

### Common Issues:
- **Missing repositories.yaml**: Helm configuration not initialized
- **Authentication errors**: Invalid or missing Git/GitHub tokens
- **Repository not found**: Incorrect repository URL or access permissions

### Quick Fixes:
1. Run the setup script: `./scripts/setup.sh`
2. Check configuration: `source .env && ./bin/helmchecker`
3. Enable dry-run mode: `export CHECKER_DRY_RUN=true`

For detailed troubleshooting, see [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)

## Development

### Building

```bash
# Build the application
make build

# Or manually
go build -o bin/helmchecker ./cmd/helmchecker
```

### Testing

```bash
# Run in dry-run mode for testing
export CHECKER_DRY_RUN=true
./bin/helmchecker
```

### Adding Helm Repositories

To ensure chart updates are detected, add common Helm repositories:

```bash
helm repo add stable https://charts.helm.sh/stable
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.