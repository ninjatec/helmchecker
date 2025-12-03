#!/bin/bash

# Helmchecker Setup Script
# This script helps you configure helmchecker with the necessary environment variables

set -e

echo "🚀 Helmchecker Configuration Setup"
echo "=================================="

# Check if running in interactive mode
if [[ -t 0 ]]; then
    INTERACTIVE=true
else
    INTERACTIVE=false
fi

# Function to prompt for input
prompt_for_value() {
    local var_name="$1"
    local prompt_text="$2"
    local default_value="$3"
    local current_value="${!var_name}"
    
    if [[ "$INTERACTIVE" == "true" ]]; then
        if [[ -n "$current_value" ]]; then
            echo "Current $var_name: $current_value"
            read -p "$prompt_text (press Enter to keep current): " input_value
            if [[ -z "$input_value" ]]; then
                input_value="$current_value"
            fi
        else
            read -p "$prompt_text: " input_value
            if [[ -z "$input_value" && -n "$default_value" ]]; then
                input_value="$default_value"
            fi
        fi
    else
        # Non-interactive mode, use current value or default
        if [[ -n "$current_value" ]]; then
            input_value="$current_value"
        else
            input_value="$default_value"
        fi
    fi
    
    export "$var_name"="$input_value"
}

# Function to validate GitHub token
validate_github_token() {
    local token="$1"
    if [[ -n "$token" ]]; then
        echo "Validating GitHub token..."
        response=$(curl -s -H "Authorization: token $token" https://api.github.com/user)
        if echo "$response" | grep -q '"login"'; then
            echo "✅ GitHub token is valid"
            return 0
        else
            echo "❌ GitHub token validation failed"
            return 1
        fi
    else
        echo "⚠️ No GitHub token provided"
        return 1
    fi
}

# Check prerequisites
echo ""
echo "🔍 Checking prerequisites..."

# Check if helm is installed
if command -v helm >/dev/null 2>&1; then
    echo "✅ Helm is installed ($(helm version --short))"
else
    echo "❌ Helm is not installed. Please install Helm first: https://helm.sh/docs/intro/install/"
    exit 1
fi

# Check if git is installed
if command -v git >/dev/null 2>&1; then
    echo "✅ Git is installed ($(git --version))"
else
    echo "❌ Git is not installed. Please install Git first"
    exit 1
fi

# Check if kubectl is available (optional)
if command -v kubectl >/dev/null 2>&1; then
    echo "✅ kubectl is available ($(kubectl version --client --short 2>/dev/null || echo 'version unknown'))"
else
    echo "⚠️ kubectl is not available. This is optional but recommended for Kubernetes deployments"
fi

echo ""
echo "📝 Configuration Setup"
echo "======================"

# Git Repository Configuration
echo ""
echo "🔧 Git Repository Configuration"
prompt_for_value "GIT_REPOSITORY" "Enter the Git repository URL (e.g., https://github.com/your-org/helm-charts.git)"
prompt_for_value "GIT_USERNAME" "Enter Git username" "helmchecker"
prompt_for_value "GIT_EMAIL" "Enter Git email" "helmchecker@example.com"
prompt_for_value "GIT_BRANCH" "Enter Git branch" "main"

# GitHub Configuration
echo ""
echo "🔧 GitHub Configuration"
prompt_for_value "GITHUB_TOKEN" "Enter GitHub Personal Access Token (needs repo permissions)"
prompt_for_value "GITHUB_OWNER" "Enter GitHub repository owner/organization"
prompt_for_value "GITHUB_REPO" "Enter GitHub repository name"

# Use GitHub token as Git token if not specified
if [[ -z "$GIT_TOKEN" && -n "$GITHUB_TOKEN" ]]; then
    export GIT_TOKEN="$GITHUB_TOKEN"
    echo "Using GitHub token for Git operations"
else
    prompt_for_value "GIT_TOKEN" "Enter Git token (or press Enter to use GitHub token)" "$GITHUB_TOKEN"
fi

# Kubernetes Configuration
echo ""
echo "🔧 Kubernetes Configuration"
prompt_for_value "KUBERNETES_NAMESPACE" "Enter Kubernetes namespace (optional)" "default"

# Checker Configuration
echo ""
echo "🔧 Checker Configuration"
prompt_for_value "CHECKER_DRY_RUN" "Enable dry-run mode? (true/false)" "true"

# Validation
echo ""
echo "🔍 Validating Configuration..."

# Validate GitHub token
if ! validate_github_token "$GITHUB_TOKEN"; then
    echo "⚠️ GitHub token validation failed. The application may not work correctly."
fi

# Test Git repository access
if [[ -n "$GIT_REPOSITORY" ]]; then
    echo "Testing Git repository access..."
    temp_dir=$(mktemp -d)
    if git ls-remote "$GIT_REPOSITORY" >/dev/null 2>&1; then
        echo "✅ Git repository is accessible"
    else
        echo "❌ Cannot access Git repository. Check URL and credentials."
    fi
    rm -rf "$temp_dir"
fi

# Generate configuration summary
echo ""
echo "📋 Configuration Summary"
echo "========================"
echo "Git Repository: $GIT_REPOSITORY"
echo "Git Username: $GIT_USERNAME"
echo "Git Email: $GIT_EMAIL"
echo "Git Branch: $GIT_BRANCH"
echo "GitHub Owner: $GITHUB_OWNER"
echo "GitHub Repo: $GITHUB_REPO"
echo "Kubernetes Namespace: $KUBERNETES_NAMESPACE"
echo "Dry Run Mode: $CHECKER_DRY_RUN"

# Generate environment file
ENV_FILE=".env.helmchecker"
echo ""
echo "💾 Generating environment file: $ENV_FILE"

cat > "$ENV_FILE" << EOF
# Helmchecker Configuration
# Generated on $(date)

# Git Configuration
export GIT_REPOSITORY="$GIT_REPOSITORY"
export GIT_TOKEN="$GIT_TOKEN"
export GIT_USERNAME="$GIT_USERNAME"
export GIT_EMAIL="$GIT_EMAIL"
export GIT_BRANCH="$GIT_BRANCH"

# GitHub Configuration
export GITHUB_TOKEN="$GITHUB_TOKEN"
export GITHUB_OWNER="$GITHUB_OWNER"
export GITHUB_REPO="$GITHUB_REPO"

# Kubernetes Configuration
export KUBERNETES_NAMESPACE="$KUBERNETES_NAMESPACE"

# Checker Configuration
export CHECKER_DRY_RUN="$CHECKER_DRY_RUN"
EOF

echo "✅ Environment file created: $ENV_FILE"

# Generate usage instructions
echo ""
echo "🚀 Usage Instructions"
echo "==================="
echo "1. Source the environment file:"
echo "   source $ENV_FILE"
echo ""
echo "2. Run helmchecker:"
echo "   ./bin/helmchecker"
echo ""
echo "3. Or run in one command:"
echo "   source $ENV_FILE && ./bin/helmchecker"
echo ""
echo "4. For Docker/Kubernetes deployment, use the values in $ENV_FILE"
echo "   to configure your ConfigMap and Secrets"

# Helm repository setup
echo ""
echo "📦 Helm Repository Setup"
echo "======================="
echo "To ensure helmchecker can find chart updates, add some common repositories:"
echo ""
echo "helm repo add stable https://charts.helm.sh/stable"
echo "helm repo add bitnami https://charts.bitnami.com/bitnami"
echo "helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx"
echo "helm repo update"

# Offer to add repositories automatically
if [[ "$INTERACTIVE" == "true" ]]; then
    echo ""
    read -p "Would you like to add common Helm repositories automatically? (y/N): " add_repos
    if [[ "$add_repos" =~ ^[Yy]$ ]]; then
        echo "Adding common Helm repositories..."
        helm repo add stable https://charts.helm.sh/stable || true
        helm repo add bitnami https://charts.bitnami.com/bitnami || true
        helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx || true
        helm repo update
        echo "✅ Helm repositories added and updated"
    fi
fi

echo ""
echo "🎉 Setup complete! You can now run helmchecker."
echo "For troubleshooting, see docs/TROUBLESHOOTING.md"