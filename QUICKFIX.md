# Quick Reference - Fixing Your Helmchecker Errors

## 🚨 Your Errors (Now Fixed!)

### Error 1: Missing repositories.yaml
```
Warning: failed to update repositories: failed to load repository file
```
✅ **FIXED** - Now automatically creates the file if missing

### Error 2: Git authentication failure
```
Chart check failed: failed to clone repository: authentication required
```
✅ **FIXED** - Added proper authentication and better error messages

## 🚀 Quick Fix - Get Running in 3 Steps

### Step 1: Run Setup Script
```bash
./scripts/setup.sh
```
This interactive script will:
- Guide you through configuration
- Validate your GitHub token
- Test repository access
- Create `.env.helmchecker` file

### Step 2: Configure Environment
You'll need to provide:
- **Git Repository URL**: Where your Helm charts are stored
- **GitHub Token**: Personal access token with `repo` permissions
- **GitHub Owner/Repo**: Your organization and repository name

Example:
```bash
GIT_REPOSITORY="https://github.com/myorg/helm-charts.git"
GITHUB_TOKEN="ghp_abc123..."
GITHUB_OWNER="myorg"
GITHUB_REPO="helm-charts"
```

### Step 3: Run Helmchecker
```bash
source .env.helmchecker
./bin/helmchecker
```

## 🔧 Manual Setup (Alternative)

If you prefer manual configuration:

```bash
# 1. Copy example file
cp .env.example .env

# 2. Edit with your values
nano .env

# 3. Source and run
source .env
./bin/helmchecker
```

## 📋 Minimum Required Variables

```bash
export GIT_REPOSITORY="https://github.com/your-org/helm-charts.git"
export GITHUB_TOKEN="ghp_your_token_here"
export GITHUB_OWNER="your-org"
export GITHUB_REPO="helm-charts"
export CHECKER_DRY_RUN="true"  # Safe mode for testing
```

## 🔍 Troubleshooting

**Still getting errors?**

1. Check your token has correct permissions:
   - Go to GitHub → Settings → Developer settings → Personal access tokens
   - Ensure `repo` scope is enabled

2. Verify repository URL:
   ```bash
   git ls-remote $GIT_REPOSITORY
   ```

3. Read the full guide:
   ```bash
   cat docs/TROUBLESHOOTING.md
   ```

## 📚 Additional Resources

- **Full Troubleshooting Guide**: `docs/TROUBLESHOOTING.md`
- **Complete Fix Details**: `docs/ERROR_FIXES.md`
- **Configuration Example**: `.env.example`
- **Updated README**: `README.md`

## 💡 Pro Tips

1. **Always test with dry-run first:**
   ```bash
   export CHECKER_DRY_RUN=true
   ```

2. **Add Helm repositories before running:**
   ```bash
   helm repo add stable https://charts.helm.sh/stable
   helm repo add bitnami https://charts.bitnami.com/bitnami
   helm repo update
   ```

3. **Check Helm releases:**
   ```bash
   helm list --all-namespaces
   ```

## ✅ What Was Fixed

| Issue | Solution |
|-------|----------|
| Missing repositories.yaml | Auto-creates file and directory structure |
| Git authentication failure | Proper token handling and helpful errors |
| No configuration validation | Clear validation with detailed error messages |

## 🎯 Next Steps

1. ✅ Run `./scripts/setup.sh`
2. ✅ Configure your environment variables
3. ✅ Test with `CHECKER_DRY_RUN=true`
4. ✅ Review created PRs (in dry-run, just logs)
5. ✅ Set `CHECKER_DRY_RUN=false` when ready for production

---

**Need Help?** See `docs/TROUBLESHOOTING.md` for detailed solutions to common issues.