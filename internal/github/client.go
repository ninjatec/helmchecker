package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

// Client represents a GitHub client
type Client struct {
	client *github.Client
}

// NewClient creates a new GitHub client
func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return &Client{
		client: client,
	}
}

// CreatePullRequest creates a new pull request
func (c *Client) CreatePullRequest(ctx context.Context, owner, repo, title, body, head, base string) (*github.PullRequest, error) {
	newPR := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head),
		Base:  github.String(base),
		Body:  github.String(body),
	}

	pr, _, err := c.client.PullRequests.Create(ctx, owner, repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return pr, nil
}

// GetPullRequest gets an existing pull request
func (c *Client) GetPullRequest(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	return pr, nil
}

// ListPullRequests lists pull requests
func (c *Client) ListPullRequests(ctx context.Context, owner, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, error) {
	prs, _, err := c.client.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	return prs, nil
}

// CheckIfPRExists checks if a pull request already exists for the given head branch
func (c *Client) CheckIfPRExists(ctx context.Context, owner, repo, head string) (*github.PullRequest, error) {
	opts := &github.PullRequestListOptions{
		State: "open",
		Head:  fmt.Sprintf("%s:%s", owner, head),
		Base:  "main",
	}

	prs, err := c.ListPullRequests(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}

	if len(prs) > 0 {
		return prs[0], nil
	}

	return nil, nil
}