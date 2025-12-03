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
     --set secrets.githubToken=your-github-token
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

- **Built-in Secrets**: Configure via `secrets.githubToken` in values.yaml
- **External Secrets**: Use External Secrets Operator with `externalSecret.enabled=true`
- **Full Configuration**: See `helm-chart/values.yaml` for all configuration options