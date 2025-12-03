# Troubleshooting Guide

This guide helps you resolve common issues when running the helmchecker application.

## Common Issues and Solutions

### 1. Missing Helm repositories.yaml file

**Error:**
```
Warning: failed to update repositories: failed to load repository file: couldn't load repositories file (/home/helmchecker/.config/helm/repositories.yaml): no such file or directory
```

**Solution:**
This error occurs when Helm hasn't been initialized or no repositories have been added. The application now automatically creates the necessary directory structure and file. However, you may want to add some repositories manually:

```bash
# Add stable repository
helm repo add stable https://charts.helm.sh/stable

# Add bitnami repository
helm repo add bitnami https://charts.bitnami.com/bitnami

# Update repositories
helm repo update
```

### 2. Git Authentication Failure

**Error:**
```
Chart check failed: failed to clone repository: failed to clone repository: authentication required: Repository not found.
```

**Causes and Solutions:**

#### Missing Git Token
Ensure you have set the `GIT_TOKEN` environment variable:

```bash
# For GitHub
export GIT_TOKEN="ghp_your_personal_access_token_here"

# For other Git providers, use appropriate token format
```

#### Incorrect Repository URL
Make sure the `GIT_REPOSITORY` environment variable points to a valid repository:

```bash
# Examples:
export GIT_REPOSITORY="https://github.com/your-org/your-charts-repo.git"
```

#### Token Permissions
Ensure your GitHub Personal Access Token has the following permissions:
- `repo` (for private repositories)
- `public_repo` (for public repositories)
- `write:packages` (if using GitHub Container Registry)

### 3. Environment Variable Configuration

Set all required environment variables:

```bash
# Git/GitHub Configuration
export GIT_REPOSITORY="https://github.com/your-org/your-charts-repo.git"
export GIT_TOKEN="your_git_token_here"
export GIT_USERNAME="helmchecker"
export GIT_EMAIL="helmchecker@yourcompany.com"
export GIT_BRANCH="main"

# GitHub API Configuration (if different from Git)
export GITHUB_TOKEN="your_github_token_here"  # Can be same as GIT_TOKEN
export GITHUB_OWNER="your-org"
export GITHUB_REPO="your-charts-repo"

# Kubernetes Configuration (optional)
export KUBERNETES_NAMESPACE="default"

# Checker Configuration
export CHECKER_DRY_RUN="true"  # Set to false for actual updates
```

### 4. Testing Configuration

To test your configuration without making changes:

```bash
# Enable dry-run mode
export CHECKER_DRY_RUN="true"

# Run the checker
./bin/helmchecker
```

### 5. Docker/Kubernetes Configuration

If running in a container, ensure environment variables are properly passed:

```yaml
# Example Kubernetes ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: helmchecker-config
data:
  GIT_REPOSITORY: "https://github.com/your-org/your-charts-repo.git"
  GIT_USERNAME: "helmchecker"
  GIT_EMAIL: "helmchecker@yourcompany.com"
  GIT_BRANCH: "main"
  GITHUB_OWNER: "your-org"
  GITHUB_REPO: "your-charts-repo"
  CHECKER_DRY_RUN: "false"

---
# Example Secret for sensitive data
apiVersion: v1
kind: Secret
metadata:
  name: helmchecker-secrets
type: Opaque
data:
  GIT_TOKEN: <base64-encoded-token>
  GITHUB_TOKEN: <base64-encoded-token>
```

### 6. Helm Client Issues

If you encounter issues with Helm client initialization:

```bash
# Initialize Helm (if needed)
helm init --client-only

# List current repositories
helm repo list

# Add required repositories
helm repo add your-repo https://your-repo-url.com
```

### 7. Chart Version Detection Issues

The application currently uses a simplified chart version detection. For production use, consider:

1. Adding proper Helm chart repositories
2. Implementing semver-based version comparison
3. Adding repository authentication if needed

### 8. Network and Connectivity Issues

If you're behind a corporate firewall:

```bash
# Set proxy if needed
export HTTP_PROXY="http://proxy.company.com:8080"
export HTTPS_PROXY="https://proxy.company.com:8080"
export NO_PROXY="localhost,127.0.0.1,.company.com"

# For Git operations
git config --global http.proxy http://proxy.company.com:8080
git config --global https.proxy https://proxy.company.com:8080
```

### 9. Debugging Tips

Enable verbose logging:

```bash
# Add debug output (modify main.go to add logging)
export HELM_DEBUG="true"
export DEBUG="true"
```

Check what releases are detected:

```bash
# List current Helm releases
helm list --all-namespaces
```

Verify repository configuration:

```bash
# Check repositories
helm repo list

# Update repositories
helm repo update
```

### 10. Development and Testing

For local development and testing:

```bash
# Build the application
make build

# Run in dry-run mode
export CHECKER_DRY_RUN="true"
./bin/helmchecker

# Check logs for issues
echo "Check the output for any warnings or errors"
```

## Getting Help

If you continue to experience issues:

1. Check the application logs for detailed error messages
2. Verify all environment variables are set correctly
3. Test Git/GitHub connectivity manually
4. Ensure Helm is properly configured
5. Check network connectivity and firewall settings

For additional support, please open an issue in the project repository with:
- Error messages
- Environment variable configuration (redact sensitive values)
- Steps to reproduce the issue