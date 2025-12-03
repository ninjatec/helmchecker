# Read-Only Filesystem Fix

## Problem
The application was failing with the error:
```
Warning: failed to update repositories: failed to create helm config directory: mkdir /.config: read-only file system
```

This occurred because:
1. The container runs with `readOnlyRootFilesystem: true` (a security best practice)
2. Helm was trying to create configuration directories in the root filesystem
3. The `HOME` environment variable wasn't set, causing Helm to default to `/`

## Solution
The fix implements a secure solution using in-memory tmpfs mounts:

### Environment Variables Added
- `HOME=/tmp` - Sets the home directory to the writable /tmp mount
- `HELM_CACHE_HOME=/tmp/.cache/helm` - Helm cache directory
- `HELM_CONFIG_HOME=/tmp/.config/helm` - Helm configuration directory  
- `HELM_DATA_HOME=/tmp/.local/share/helm` - Helm data directory

### Volume Mounts Added
Three emptyDir volumes provide in-memory storage:
- `tmp` → `/tmp` - General temporary storage
- `helm-cache` → `/tmp/.cache/helm` - Helm cache
- `helm-config` → `/tmp/.config/helm` - Helm configuration

## Security Considerations

✅ **Maintains Security Posture**
- Read-only root filesystem remains enabled
- No privileged escalation
- Runs as non-root user (UID 1000)
- All capabilities dropped

✅ **Uses emptyDir (tmpfs)**
- In-memory storage (no persistent data)
- Automatically cleaned up when pod terminates
- No persistent volume claims needed
- Isolated per pod instance

✅ **Follows Best Practices**
- Explicit Helm directories prevent unpredictable behavior
- Minimal writable surface area
- Helm configuration is ephemeral (appropriate for this use case)

## Implementation
Changes made in:
- `helm-chart/templates/cronjob.yaml` - Added environment variables and volumes
- `helm-chart/Chart.yaml` - Bumped chart version to 0.0.4

## Testing
After upgrading to chart version 0.0.4, the application should run without filesystem errors and successfully complete Helm operations.
