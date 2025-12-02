# Helm Repository

This repository provides the Helm chart for **helmchecker** - an automated tool for checking outdated Helm charts in Kubernetes clusters.

## Usage

### Add the Helm Repository

```bash
helm repo add ninjatec https://ninjatec.github.io/helmchecker
helm repo update
```

### Install the Chart

```bash
# Install with default values
helm install helmchecker ninjatec/helmchecker

# Install with custom configuration
helm install helmchecker ninjatec/helmchecker \
  --set config.git.repository=https://github.com/your-org/your-repo.git \
  --set config.github.owner=your-org \
  --set config.github.repo=your-repo \
  --set secrets.githubToken=your-token
```

### Configuration

See the [full documentation](https://github.com/ninjatec/helmchecker) for detailed configuration options.

### Chart Values

Key configuration parameters:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.git.repository` | Target Git repository URL | `""` |
| `config.github.owner` | GitHub repository owner | `""` |
| `config.github.repo` | GitHub repository name | `""` |
| `cronjob.schedule` | Cron schedule | `"0 2 * * *"` |
| `secrets.githubToken` | GitHub API token | `""` |

### Example Values

```yaml
config:
  git:
    repository: "https://github.com/your-org/k8s-manifests.git"
    username: "helmchecker-bot"
    email: "helmchecker-bot@your-org.com"

  github:
    owner: "your-org"
    repo: "k8s-manifests"

secrets:
  githubToken: "ghp_your_token_here"

cronjob:
  schedule: "0 2 * * *"  # Daily at 2 AM
```

## Chart Versions

| Version | App Version | Description |
|---------|-------------|-------------|
| 0.1.0   | 0.1.0      | Initial release |

## Support

- 📖 [Documentation](https://github.com/ninjatec/helmchecker)
- 🐛 [Issues](https://github.com/ninjatec/helmchecker/issues)
- 💬 [Discussions](https://github.com/ninjatec/helmchecker/discussions)