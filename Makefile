# Makefile for Helm Chart Checker

# Variables
DOCKER_REGISTRY ?= docker.io/ninjatec
IMAGE_NAME ?= helmchecker
IMAGE_TAG ?= latest
CHART_NAME = helmchecker
NAMESPACE ?= helmchecker

# Go variables
GOOS ?= linux
GOARCH ?= amd64

.PHONY: help build test docker-build docker-push helm-lint helm-install helm-uninstall clean

help: ## Display this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go application
	@echo "Building Go application..."
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a -installsuffix cgo -o bin/helmchecker ./cmd/helmchecker

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) .

docker-push: docker-build ## Build and push Docker image
	@echo "Pushing Docker image..."
	docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

helm-lint: ## Lint the Helm chart
	@echo "Linting Helm chart..."
	helm lint ./helm-chart

helm-template: ## Generate Helm templates
	@echo "Generating Helm templates..."
	helm template $(CHART_NAME) ./helm-chart --values ./helm-chart/values.yaml

helm-install: ## Install the Helm chart
	@echo "Installing Helm chart..."
	helm install $(CHART_NAME) ./helm-chart \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--set image.repository=$(DOCKER_REGISTRY)/$(IMAGE_NAME) \
		--set image.tag=$(IMAGE_TAG)

helm-upgrade: ## Upgrade the Helm chart
	@echo "Upgrading Helm chart..."
	helm upgrade $(CHART_NAME) ./helm-chart \
		--namespace $(NAMESPACE) \
		--set image.repository=$(DOCKER_REGISTRY)/$(IMAGE_NAME) \
		--set image.tag=$(IMAGE_TAG)

helm-uninstall: ## Uninstall the Helm chart
	@echo "Uninstalling Helm chart..."
	helm uninstall $(CHART_NAME) --namespace $(NAMESPACE)

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	docker rmi $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) 2>/dev/null || true

helm-package: ## Package the Helm chart
	@echo "Packaging Helm chart..."
	helm package helm-chart --destination charts/

helm-index: helm-package ## Generate Helm repository index
	@echo "Generating Helm repository index..."
	helm repo index charts/ --url https://ninjatec.github.io/helmchecker

helm-publish: helm-index ## Publish Helm chart to repository
	@echo "Publishing Helm chart..."
	@echo "Make sure to commit and push the changes in charts/ directory"
	@echo "GitHub Pages will automatically serve the repository"

# Development targets
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	go mod tidy
	go mod download

dev-build: build ## Build for development

dev-run: build ## Run the application locally
	@echo "Running application locally..."
	./bin/helmchecker