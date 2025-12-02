package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gitconfig "github.com/marccoxall/helmchecker/internal/config"
)

// Client represents a Git client
type Client struct {
	config gitconfig.GitConfig
}

// NewClient creates a new Git client
func NewClient(cfg gitconfig.GitConfig) *Client {
	return &Client{
		config: cfg,
	}
}

// CloneRepository clones a repository to a temporary directory
func (c *Client) CloneRepository(ctx context.Context) (string, *gogit.Repository, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "helmchecker-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Clone the repository
	repo, err := gogit.PlainCloneContext(ctx, tempDir, false, &gogit.CloneOptions{
		URL:      c.config.Repository,
		Progress: os.Stdout,
	})
	if err != nil {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Printf("Warning: failed to clean up temp directory: %v\n", removeErr)
		}
		return "", nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	return tempDir, repo, nil
}

// CreateBranch creates a new branch from the base branch
func (c *Client) CreateBranch(repo *gogit.Repository, branchName string) error {
	workTree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get the current HEAD
	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create and checkout the new branch
	branchRefName := fmt.Sprintf("refs/heads/%s", branchName)
	err = workTree.Checkout(&gogit.CheckoutOptions{
		Branch: headRef.Name(),
		Create: true,
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// Create the branch reference
	ref := plumbing.NewHashReference(plumbing.ReferenceName(branchRefName), headRef.Hash())
	err = repo.Storer.SetReference(ref)
	if err != nil {
		return fmt.Errorf("failed to set branch reference: %w", err)
	}

	return nil
}

// CommitChanges commits changes to the repository
func (c *Client) CommitChanges(repo *gogit.Repository, message string) error {
	workTree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Add all changes
	_, err = workTree.Add(".")
	if err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Commit the changes
	commit, err := workTree.Commit(message, &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  c.config.Username,
			Email: c.config.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Log the commit
	obj, err := repo.CommitObject(commit)
	if err != nil {
		return fmt.Errorf("failed to get commit object: %w", err)
	}

	fmt.Printf("Committed changes: %s\n", obj.Hash)
	return nil
}

// PushBranch pushes a branch to the remote repository
func (c *Client) PushBranch(repo *gogit.Repository, branchName string) error {
	// Configure authentication
	auth := &http.BasicAuth{
		Username: c.config.Username,
		Password: c.config.Token,
	}

	// Push the branch
	err := repo.Push(&gogit.PushOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branchName, branchName)),
		},
		Auth: auth,
	})
	if err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}

	return nil
}

// UpdateFile updates a file in the repository
func (c *Client) UpdateFile(repoPath, filePath, content string) error {
	fullPath := filepath.Join(repoPath, filePath)

	// Ensure the directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	return nil
}