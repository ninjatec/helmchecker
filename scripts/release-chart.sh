#!/bin/bash
set -e

# Script to release a new version of the Helm chart

VERSION=${1:-"0.1.0"}
CHART_DIR="helm-chart"
CHARTS_DIR="charts"

echo "📦 Releasing Helm Chart version $VERSION"

# Update chart version
echo "🔄 Updating chart version to $VERSION"
sed -i.bak "s/^version: .*/version: $VERSION/" $CHART_DIR/Chart.yaml
sed -i.bak "s/^appVersion: .*/appVersion: \"$VERSION\"/" $CHART_DIR/Chart.yaml
rm -f $CHART_DIR/Chart.yaml.bak

# Lint the chart
echo "🔍 Linting Helm chart..."
helm lint $CHART_DIR

# Package the chart
echo "📦 Packaging Helm chart..."
mkdir -p $CHARTS_DIR
helm package $CHART_DIR --destination $CHARTS_DIR

# Generate index
echo "📋 Generating repository index..."
helm repo index $CHARTS_DIR --url https://ninjatec.github.io/helmchecker

echo "✅ Chart version $VERSION ready for release!"
echo ""
echo "Next steps:"
echo "1. Review changes: git diff"
echo "2. Commit changes: git add . && git commit -m 'Release chart version $VERSION'"
echo "3. Create tag: git tag v$VERSION"
echo "4. Push: git push origin main --tags"
echo "5. GitHub Pages will automatically publish the chart"