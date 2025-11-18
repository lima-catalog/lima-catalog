package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/lima-catalog/lima-catalog/pkg/cache"
	"github.com/lima-catalog/lima-catalog/pkg/config"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client with rate limit management and caching
type Client struct {
	client *github.Client
	ctx    context.Context
	cache  *cache.Cache
}

// NewClient creates a new GitHub API client with authentication and caching
func NewClient(ctx context.Context, token string) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
		ctx:    ctx,
		cache:  cache.New(1 * time.Hour), // Cache API responses for 1 hour
	}
}

// GetClient returns the underlying GitHub client for direct API access
func (c *Client) GetClient() *github.Client {
	return c.client
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

// HandleRateLimitError checks if a response indicates a rate limit error and waits if needed.
// Returns true if it handled a rate limit error (caller should retry), false otherwise.
func (c *Client) HandleRateLimitError(resp *github.Response, limitType string) bool {
	// Check if response indicates rate limiting (403 Forbidden or 429 Too Many Requests)
	if resp == nil || (resp.StatusCode != 403 && resp.StatusCode != 429) {
		return false
	}

	// Get rate limit information
	limits, err := c.RateLimit()
	if err != nil {
		fmt.Printf("  Warning: failed to check rate limit: %v\n", err)
		return false
	}

	// Determine which rate limit to check based on type
	var resetTime time.Time
	switch limitType {
	case "search":
		if limits.Search != nil {
			resetTime = limits.Search.Reset.Time
		}
	case "core":
		if limits.Core != nil {
			resetTime = limits.Core.Reset.Time
		}
	default:
		fmt.Printf("  Warning: unknown rate limit type: %s\n", limitType)
		return false
	}

	if resetTime.IsZero() {
		return false
	}

	// Wait until rate limit resets
	waitDuration := time.Until(resetTime)
	if waitDuration > 0 {
		fmt.Printf("  Rate limit exceeded, waiting %v until reset at %s\n",
			waitDuration.Round(time.Second), resetTime.Format(time.RFC3339))
		time.Sleep(waitDuration + config.SearchAPIQueryDelay) // Add buffer
		fmt.Println("  Retrying after rate limit reset...")
		return true
	}

	return false
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

// GetRepository fetches repository information (with caching)
func (c *Client) GetRepository(owner, repo string) (*github.Repository, error) {
	cacheKey := fmt.Sprintf("repo:%s/%s", owner, repo)

	// Check cache first
	if cached, ok := c.cache.Get(cacheKey); ok {
		if repository, ok := cached.(*github.Repository); ok {
			return repository, nil
		}
	}

	// Fetch from API
	repository, _, err := c.client.Repositories.Get(c.ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.Set(cacheKey, repository)

	return repository, nil
}

// GetUser fetches user or organization information (with caching)
func (c *Client) GetUser(login string) (*github.User, error) {
	cacheKey := fmt.Sprintf("user:%s", login)

	// Check cache first
	if cached, ok := c.cache.Get(cacheKey); ok {
		if user, ok := cached.(*github.User); ok {
			return user, nil
		}
	}

	// Fetch from API
	user, _, err := c.client.Users.Get(c.ctx, login)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cache.Set(cacheKey, user)

	return user, nil
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
