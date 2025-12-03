# Helmchecker Error Fixes - Summary

## Issues Addressed

This document summarizes the fixes applied to resolve the errors you encountered.

### 1. Missing Helm repositories.yaml File

**Original Error:**
```
Warning: failed to update repositories: failed to load repository file: couldn't load repositories file (/home/helmchecker/.config/helm/repositories.yaml): no such file or directory
```

**Root Cause:**
The Helm client was trying to load a repository configuration file that didn't exist because Helm had not been initialized in the container environment.

**Fix Applied:**
Modified `internal/helm/client.go` - `UpdateRepositories()` function:
- Added automatic directory creation for Helm config
- Added graceful handling for missing repositories file
- Creates an empty repositories file if none exists
- Continues operation without error if no repositories are configured yet

**Code Changes:**
```go
// Ensure the helm config directory exists
if err := os.MkdirAll(filepath.Dir(repoFile), 0755); err != nil {
    return fmt.Errorf("failed to create helm config directory: %w", err)
}

f, err := repo.LoadFile(repoFile)
if err != nil {
    // If file doesn't exist, create a new one
    if os.IsNotExist(err) {
        f = repo.NewFile()
        if err := f.WriteFile(repoFile, 0644); err != nil {
            return fmt.Errorf("failed to create repository file: %w", err)
        }
        return nil // No repositories to update yet
    }
    return fmt.Errorf("failed to load repository file: %w", err)
}
```

### 2. Git Authentication Failure

**Original Error:**
```
Chart check failed: failed to clone repository: failed to clone repository: authentication required: Repository not found.
```

**Root Cause:**
The Git client was attempting to clone a repository without providing authentication credentials, even when they were available via environment variables.

**Fix Applied:**
Modified `internal/git/client.go` - `CloneRepository()` function:
- Added proper authentication handling using provided tokens
- Added conditional authentication (only use if token is available)
- Improved error messages with helpful hints about missing credentials
- Better error context for troubleshooting

**Code Changes:**
```go
// Configure authentication if token is available
var auth *http.BasicAuth
if c.config.Token != "" {
    auth = &http.BasicAuth{
        Username: c.config.Username,
        Password: c.config.Token,
    }
}

// Clone the repository
cloneOptions := &gogit.CloneOptions{
    URL:      c.config.Repository,
    Progress: os.Stdout,
}

// Only set auth if we have credentials
if auth != nil {
    cloneOptions.Auth = auth
}

repo, err := gogit.PlainCloneContext(ctx, tempDir, false, cloneOptions)
if err != nil {
    // ... cleanup ...
    
    // Provide more helpful error message
    if c.config.Token == "" {
        return "", nil, fmt.Errorf("failed to clone repository: %w (hint: make sure GIT_TOKEN environment variable is set if the repository requires authentication)", err)
    }
    return "", nil, fmt.Errorf("failed to clone repository: %w (hint: check repository URL and credentials)", err)
}
```

### 3. Configuration Validation

**Enhancement:**
Added comprehensive configuration validation to catch missing or invalid environment variables before attempting operations.

**Fix Applied:**
Modified `internal/config/config.go` - Added `Validate()` method:
- Validates all required environment variables are set
- Provides clear error messages listing all missing configurations
- Automatically uses GitHub token for Git operations if Git token is not specified
- Ensures proper setup before any operations begin

**Code Changes:**
```go
// Validate validates the configuration
func (c *Config) Validate() error {
    var errors []string

    // Validate Git configuration
    if c.Git.Repository == "" {
        errors = append(errors, "GIT_REPOSITORY environment variable is required")
    }
    
    if c.Git.Token == "" && c.GitHub.Token == "" {
        errors = append(errors, "either GIT_TOKEN or GITHUB_TOKEN environment variable is required")
    }

    // If Git token is empty but GitHub token is set, use GitHub token for Git operations
    if c.Git.Token == "" && c.GitHub.Token != "" {
        c.Git.Token = c.GitHub.Token
    }

    // Validate GitHub configuration
    if c.GitHub.Token == "" {
        errors = append(errors, "GITHUB_TOKEN environment variable is required")
    }
    
    if c.GitHub.Owner == "" {
        errors = append(errors, "GITHUB_OWNER environment variable is required")
    }
    
    if c.GitHub.Repo == "" {
        errors = append(errors, "GITHUB_REPO environment variable is required")
    }

    if len(errors) > 0 {
        return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
    }

    return nil
}
```

## Additional Resources Created

### 1. Troubleshooting Guide
**File:** `docs/TROUBLESHOOTING.md`
- Comprehensive troubleshooting documentation
- Solutions for common issues
- Configuration examples
- Network and firewall considerations
- Debugging tips

### 2. Setup Script
**File:** `scripts/setup.sh`
- Interactive configuration wizard
- Validates GitHub tokens
- Tests repository access
- Creates environment file automatically
- Optionally adds common Helm repositories

### 3. Environment Example
**File:** `.env.example`
- Template for environment configuration
- All required and optional variables documented
- Usage instructions included

### 4. Updated README
**File:** `README.md`
- Added quick start options
- Configuration guide
- Troubleshooting section
- Development instructions

## How to Use the Fixes

### Quick Start (Easiest)
```bash
# Run the interactive setup
./scripts/setup.sh

# Source the generated environment file
source .env.helmchecker

# Run helmchecker
./bin/helmchecker
```

### Manual Setup
```bash
# Copy example environment file
cp .env.example .env

# Edit with your configuration
nano .env

# Source and run
source .env
./bin/helmchecker
```

### Required Environment Variables
At minimum, set these before running:
```bash
export GIT_REPOSITORY="https://github.com/your-org/helm-charts.git"
export GITHUB_TOKEN="ghp_your_token_here"
export GITHUB_OWNER="your-org"
export GITHUB_REPO="helm-charts"
```

## Testing the Fixes

1. **Build the application:**
   ```bash
   go build -o bin/helmchecker ./cmd/helmchecker
   ```

2. **Set up configuration:**
   ```bash
   ./scripts/setup.sh
   source .env.helmchecker
   ```

3. **Run in dry-run mode:**
   ```bash
   export CHECKER_DRY_RUN=true
   ./bin/helmchecker
   ```

4. **Check for errors:**
   - No "repositories.yaml" error should appear
   - Authentication should work if credentials are correct
   - Configuration validation errors should be clear and helpful

## Next Steps

1. Set up your environment variables
2. Run the setup script or manually configure
3. Test with dry-run mode enabled
4. Once working, disable dry-run mode to create actual PRs

## Files Modified

- `internal/helm/client.go` - Fixed repositories file handling
- `internal/git/client.go` - Fixed authentication
- `internal/config/config.go` - Added validation

## Files Created

- `docs/TROUBLESHOOTING.md` - Troubleshooting guide
- `scripts/setup.sh` - Interactive setup script
- `.env.example` - Environment configuration template
- `README.md` - Updated documentation

## Summary

All reported errors have been addressed:
✅ Missing repositories.yaml file is now handled gracefully
✅ Git authentication is properly configured
✅ Configuration validation provides clear error messages
✅ Comprehensive documentation and setup tools added