package main

import (
	"context"
	"log"
	"time"

	"github.com/marccoxall/helmchecker/internal/checker"
	"github.com/marccoxall/helmchecker/internal/config"
	"github.com/marccoxall/helmchecker/internal/git"
	"github.com/marccoxall/helmchecker/internal/github"
	"github.com/marccoxall/helmchecker/internal/helm"
)

func main() {
	log.Println("Starting Helm Chart Checker...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Helm client
	helmClient, err := helm.NewClient(cfg.Kubernetes.Namespace)
	if err != nil {
		log.Fatalf("Failed to initialize Helm client: %v", err)
	}

	// Initialize Git client
	gitClient := git.NewClient(cfg.Git)

	// Initialize GitHub client
	githubClient := github.NewClient(cfg.GitHub.Token)

	// Initialize checker
	checker := checker.New(helmClient, gitClient, githubClient, cfg)

	// Run the check
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := checker.Run(ctx); err != nil {
		log.Fatalf("Chart check failed: %v", err)
	}

	log.Println("Helm Chart Checker completed successfully")
}