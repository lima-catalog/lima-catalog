package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client with rate limit management
type Client struct {
	client *github.Client
	ctx    context.Context
}

// NewClient creates a new GitHub API client with authentication
func NewClient(ctx context.Context, token string) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
		ctx:    ctx,
	}
}

// RateLimit returns the current rate limit status
func (c *Client) RateLimit() (*github.RateLimits, error) {
	limits, _, err := c.client.RateLimit.Get(c.ctx)
	return limits, err
}

// CheckRateLimit checks if we have enough API calls remaining
func (c *Client) CheckRateLimit(minimumRemaining int) error {
	limits, err := c.RateLimit()
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	core := limits.Core
	if core.Remaining < minimumRemaining {
		resetTime := core.Reset.Time
		waitDuration := time.Until(resetTime)
		return fmt.Errorf("rate limit too low (%d/%d remaining), resets at %s (in %s)",
			core.Remaining, core.Limit, resetTime.Format(time.RFC3339), waitDuration)
	}

	return nil
}

// SearchCode searches for code on GitHub
func (c *Client) SearchCode(query string, page int) (*github.CodeSearchResult, *github.Response, error) {
	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: 100,
		},
	}

	result, resp, err := c.client.Search.Code(c.ctx, query, opts)
	return result, resp, err
}

// GetRepository fetches repository information
func (c *Client) GetRepository(owner, repo string) (*github.Repository, error) {
	repository, _, err := c.client.Repositories.Get(c.ctx, owner, repo)
	return repository, err
}

// GetUser fetches user or organization information
func (c *Client) GetUser(login string) (*github.User, error) {
	user, _, err := c.client.Users.Get(c.ctx, login)
	return user, err
}

// ListRepositoryContents lists contents of a directory in a repository
func (c *Client) ListRepositoryContents(owner, repo, path string) ([]*github.RepositoryContent, error) {
	_, contents, _, err := c.client.Repositories.GetContents(c.ctx, owner, repo, path, nil)
	return contents, err
}

// GetRepositoryContent gets a single file's content
func (c *Client) GetRepositoryContent(owner, repo, path string) (*github.RepositoryContent, error) {
	content, _, _, err := c.client.Repositories.GetContents(c.ctx, owner, repo, path, nil)
	return content, err
}
