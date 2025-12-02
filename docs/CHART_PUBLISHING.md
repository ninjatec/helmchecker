# Helm Chart Publishing Guide

## Publishing Options

### 1. GitHub Pages (Free & Recommended)

**Setup:**
1. Enable GitHub Pages in repository settings
2. Set source to "GitHub Actions"
3. Use the provided workflow

**Usage by consumers:**
```bash
helm repo add ninjatec https://ninjatec.github.io/helmchecker
helm install helmchecker ninjatec/helmchecker
```

### 2. OCI Registry (Docker Hub/GitHub Container Registry)

**Setup:**
```bash
# Enable OCI support in Helm
export HELM_EXPERIMENTAL_OCI=1

# Login to registry
helm registry login ghcr.io -u USERNAME

# Package and push
helm package helm-chart
helm push helmchecker-0.1.0.tgz oci://ghcr.io/ninjatec/charts
```

**Usage by consumers:**
```bash
helm install helmchecker oci://ghcr.io/ninjatec/charts/helmchecker --version 0.1.0
```

### 3. Artifact Hub

**Setup:**
1. Create account at [artifacthub.io](https://artifacthub.io)
2. Add your GitHub Pages repository
3. Add `artifacthub-repo.yml`:

```yaml
repositoryID: "ninjatec-helmchecker"
owners:
  - name: "ninjatec"
    email: "contact@ninjatec.com"
```

### 4. ChartMuseum (Self-hosted)

**Setup:**
```bash
# Deploy ChartMuseum
helm repo add chartmuseum https://chartmuseum.github.io/charts
helm install chartmuseum chartmuseum/chartmuseum

# Upload chart
curl --data-binary "@helmchecker-0.1.0.tgz" http://chartmuseum-url/api/charts
```

## Release Process

### Automated Release (Recommended)

1. **Create a new release:**
   ```bash
   ./scripts/release-chart.sh 0.1.1
   ```

2. **Review and commit:**
   ```bash
   git add .
   git commit -m "Release chart version 0.1.1"
   git tag v0.1.1
   git push origin main --tags
   ```

3. **GitHub Actions automatically:**
   - Packages the chart
   - Updates the repository index
   - Publishes to GitHub Pages

### Manual Release

```bash
# Package chart
make helm-package

# Update index
make helm-index

# Commit and push
git add charts/
git commit -m "Update chart repository"
git push origin main
```

## Chart Distribution

### For Public Charts
- **GitHub Pages**: Free, reliable, good for open source
- **Artifact Hub**: Great discoverability
- **OCI Registry**: Modern approach, integrates with container registries

### For Private Charts
- **ChartMuseum**: Full-featured private repository
- **OCI Private Registry**: Simple, integrates with existing infrastructure
- **AWS S3 + CloudFront**: Scalable, cost-effective

## Best Practices

1. **Semantic Versioning**: Use semver for chart versions
2. **Version Preservation**: All older chart versions are automatically preserved
3. **Changelog**: Maintain CHANGELOG.md for version history
4. **Security**: Sign charts with GPG keys for production
5. **Testing**: Validate charts before publishing
6. **Documentation**: Include comprehensive README and examples

## Version Management

The publishing process automatically preserves all older chart versions:
- The `helm repo index --merge` command ensures older versions remain available
- Both GitHub Actions and local scripts maintain version history
- Users can install any previously published version

Check available versions:
```bash
# Using make command
make helm-versions

# Using helm directly
helm search repo ninjatec/helmchecker --versions
```

## Consumer Documentation

Once published, users can install your chart:

```bash
# Add repository
helm repo add ninjatec https://ninjatec.github.io/helmchecker

# Search for charts
helm search repo helmchecker

# Install
helm install my-helmchecker ninjatec/helmchecker \
  --namespace helmchecker --create-namespace \
  --set config.git.repository=https://github.com/myorg/k8s-repo.git
```