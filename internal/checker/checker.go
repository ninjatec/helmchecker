package checker

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/marccoxall/helmchecker/internal/config"
	gitclient "github.com/marccoxall/helmchecker/internal/git"
	"github.com/marccoxall/helmchecker/internal/github"
	"github.com/marccoxall/helmchecker/internal/helm"
)

// Checker represents the main chart checker
type Checker struct {
	helmClient   *helm.Client
	gitClient    *gitclient.Client
	githubClient *github.Client
	config       *config.Config
}

// ChartUpdate represents a chart that needs to be updated
type ChartUpdate struct {
	Release        *helm.Release
	CurrentVersion string
	LatestVersion  string
	Repository     string
}

// New creates a new checker instance
func New(helmClient *helm.Client, gitClient *gitclient.Client, githubClient *github.Client, cfg *config.Config) *Checker {
	return &Checker{
		helmClient:   helmClient,
		gitClient:    gitClient,
		githubClient: githubClient,
		config:       cfg,
	}
}

// Run executes the chart checking process
func (c *Checker) Run(ctx context.Context) error {
	log.Println("Starting chart update check...")

	// Get all installed releases
	releases, err := c.helmClient.ListReleases(ctx)
	if err != nil {
		return fmt.Errorf("failed to list releases: %w", err)
	}

	log.Printf("Found %d installed releases", len(releases))

	// Check for updates
	updates, err := c.checkForUpdates(ctx, releases)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if len(updates) == 0 {
		log.Println("No chart updates found")
		return nil
	}

	log.Printf("Found %d chart updates", len(updates))

	// Process updates if not in dry run mode
	if !c.config.Checker.DryRun {
		return c.processUpdates(ctx, updates)
	}

	// In dry run mode, just log what would be updated
	for _, update := range updates {
		log.Printf("DRY RUN: Would update %s from %s to %s",
			update.Release.Chart,
			update.CurrentVersion,
			update.LatestVersion)
	}

	return nil
}

// checkForUpdates checks all releases for available updates
func (c *Checker) checkForUpdates(ctx context.Context, releases []*helm.Release) ([]*ChartUpdate, error) {
	var updates []*ChartUpdate

	// Update repository indexes
	if err := c.helmClient.UpdateRepositories(ctx); err != nil {
		log.Printf("Warning: failed to update repositories: %v", err)
	}

	for _, release := range releases {
		// Skip if chart is in exclude list
		if c.isExcluded(release.Chart) {
			continue
		}

		// Skip if include list is specified and chart is not in it
		if len(c.config.Checker.IncludeCharts) > 0 && !c.isIncluded(release.Chart) {
			continue
		}

		log.Printf("Checking chart %s (current: %s)", release.Chart, release.Version)

		// Get latest version from repository
		latest, err := c.helmClient.GetLatestChartVersion(ctx, release.Chart, release.Repository)
		if err != nil {
			log.Printf("Warning: failed to get latest version for %s: %v", release.Chart, err)
			continue
		}

		// Compare versions
		if c.isNewerVersion(latest.Version, release.Version) {
			updates = append(updates, &ChartUpdate{
				Release:        release,
				CurrentVersion: release.Version,
				LatestVersion:  latest.Version,
				Repository:     release.Repository,
			})
		}
	}

	return updates, nil
}

// processUpdates processes the chart updates by creating branches and PRs
func (c *Checker) processUpdates(ctx context.Context, updates []*ChartUpdate) error {
	// Clone the repository
	repoPath, repo, err := c.gitClient.CloneRepository(ctx)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(repoPath); err != nil {
			log.Printf("Warning: failed to clean up temp directory %s: %v", repoPath, err)
		}
	}()

	for _, update := range updates {
		if err := c.processUpdate(ctx, repoPath, repo, update); err != nil {
			log.Printf("Failed to process update for %s: %v", update.Release.Chart, err)
			continue
		}
	}

	return nil
}

// processUpdate processes a single chart update
func (c *Checker) processUpdate(ctx context.Context, repoPath string, repo *git.Repository, update *ChartUpdate) error {
	branchName := fmt.Sprintf("update-%s-%s", update.Release.Chart, update.LatestVersion)
	
	log.Printf("Processing update for %s: %s -> %s", 
		update.Release.Chart, 
		update.CurrentVersion, 
		update.LatestVersion)

	// Check if PR already exists
	existingPR, err := c.githubClient.CheckIfPRExists(ctx, 
		c.config.GitHub.Owner, 
		c.config.GitHub.Repo, 
		branchName)
	if err != nil {
		return fmt.Errorf("failed to check for existing PR: %w", err)
	}

	if existingPR != nil {
		log.Printf("PR already exists for %s: %s", update.Release.Chart, *existingPR.HTMLURL)
		return nil
	}

	// Create a new branch
	if err := c.gitClient.CreateBranch(repo, branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// Update the chart files
	if err := c.updateChartFiles(repoPath, update); err != nil {
		return fmt.Errorf("failed to update chart files: %w", err)
	}

	// Commit changes
	commitMsg := fmt.Sprintf(c.config.Checker.CommitMessage, 
		update.Release.Chart, 
		update.LatestVersion)
	
	if err := c.gitClient.CommitChanges(repo, commitMsg); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Push branch
	if err := c.gitClient.PushBranch(repo, branchName); err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}

	// Create pull request
	prTitle := fmt.Sprintf(c.config.Checker.PullRequestTitle, 
		update.Release.Chart, 
		update.LatestVersion)
	
	prBody := fmt.Sprintf(c.config.Checker.PullRequestBody, 
		update.Release.Chart, 
		update.CurrentVersion, 
		update.LatestVersion)

	pr, err := c.githubClient.CreatePullRequest(ctx,
		c.config.GitHub.Owner,
		c.config.GitHub.Repo,
		prTitle,
		prBody,
		branchName,
		c.config.Git.Branch)
	
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	log.Printf("Created pull request for %s: %s", update.Release.Chart, *pr.HTMLURL)
	return nil
}

// updateChartFiles updates the chart files with new version information
func (c *Checker) updateChartFiles(repoPath string, update *ChartUpdate) error {
	// This is a simplified implementation
	// In a real scenario, you would need to:
	// 1. Find the chart files (Chart.yaml, values.yaml, etc.)
	// 2. Parse and update the version fields
	// 3. Handle different chart structures and formats

	// For demonstration, we'll create a simple update file
	updateContent := fmt.Sprintf(`# Chart Update
Chart: %s
Current Version: %s
New Version: %s
Repository: %s
Timestamp: %s
`, update.Release.Chart, update.CurrentVersion, update.LatestVersion, update.Repository, "2024-12-02")

	filename := fmt.Sprintf("updates/%s-update.txt", update.Release.Chart)
	return c.gitClient.UpdateFile(repoPath, filename, updateContent)
}

// isExcluded checks if a chart is in the exclude list
func (c *Checker) isExcluded(chartName string) bool {
	for _, excluded := range c.config.Checker.ExcludeCharts {
		if excluded == chartName {
			return true
		}
	}
	return false
}

// isIncluded checks if a chart is in the include list
func (c *Checker) isIncluded(chartName string) bool {
	if len(c.config.Checker.IncludeCharts) == 0 {
		return true
	}
	
	for _, included := range c.config.Checker.IncludeCharts {
		if included == chartName {
			return true
		}
	}
	return false
}

// isNewerVersion compares two version strings
// This is a simplified implementation - in production you should use semver
func (c *Checker) isNewerVersion(latest, current string) bool {
	// Remove 'v' prefix if present
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")
	
	// Simple string comparison (not semver compliant)
	return latest != current && latest > current
}