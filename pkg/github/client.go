// Package github provides a wrapper around the GitHub API client with
// rate limit management and response caching.
//
// Features:
//
//   - Automatic rate limit checking and waiting
//   - In-memory response caching with TTL (reduces API calls)
//   - Centralized error handling for rate limit errors
//   - OAuth2 authentication
//
// Usage:
//
//	client := github.NewClient(ctx, token)
//	repo, err := client.GetRepository("owner", "repo")
//	if err != nil {
//	    return err
//	}
//
// Rate Limiting:
//
// The client automatically checks rate limits before making requests.
// When rate limits are exhausted, operations will fail rather than waiting
// indefinitely. Use HandleRateLimitError to check responses for rate limit errors.
//
// Caching:
//
// Repository and user responses are cached for 1 hour by default.
// This significantly reduces API calls when processing multiple templates
// from the same repository.
//
// Error Handling:
//
// All errors are wrapped with context using fmt.Errorf with %w.
// Network errors and API errors are returned to the caller.
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

// NewClient creates a new GitHub API client with OAuth2 authentication and response caching.
//
// The client is configured with:
//   - OAuth2 authentication using the provided token
//   - 1-hour response cache for repository and user API calls
//   - Automatic rate limit checking (must be called explicitly via CheckRateLimit)
//
// Parameters:
//   - ctx: Context for all API requests (cancellation, timeouts)
//   - token: GitHub personal access token or OAuth token
//
// Returns a configured Client ready for making GitHub API calls.
//
// The context is stored and used for all subsequent API calls made through
// this client instance.
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

// GetClient returns the underlying GitHub API client for direct access.
//
// Use this when you need to call GitHub API methods that aren't wrapped
// by this Client type. The returned client uses the same authentication
// and context as the wrapper.
//
// Returns the underlying *github.Client instance.
func (c *Client) GetClient() *github.Client {
	return c.client
}

// RateLimit returns the current rate limit status for all GitHub API endpoints.
//
// Returns rate limit information including:
//   - Core: General API calls (5000/hour for authenticated users)
//   - Search: Code search API calls (30/minute)
//   - GraphQL: GraphQL API calls
//
// Returns:
//   - Rate limit status with remaining calls and reset times
//   - Error if the API call fails
//
// Use this to check rate limit status before making expensive API operations.
func (c *Client) RateLimit() (*github.RateLimits, error) {
	limits, _, err := c.client.RateLimit.Get(c.ctx)
	return limits, err
}

// CheckRateLimit verifies that sufficient API calls remain before proceeding.
//
// Checks the Core API rate limit against the specified minimum threshold.
// Use this before starting operations that will make many API calls to avoid
// hitting rate limits mid-operation.
//
// Parameters:
//   - minimumRemaining: Minimum number of API calls that must be available.
//     Operation should not proceed if remaining calls fall below this threshold.
//
// Returns:
//   - nil if sufficient API calls remain (remaining >= minimumRemaining)
//   - Error with details about current limit and reset time if threshold not met
//
// The error message includes:
//   - Current remaining/total calls
//   - Reset timestamp (when the limit will refresh)
//   - Wait duration until reset
//
// Example:
//
//	if err := client.CheckRateLimit(100); err != nil {
//	    log.Printf("Rate limit too low: %v", err)
//	    return err
//	}
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
//
// Detects rate limit errors (403 Forbidden or 429 Too Many Requests) and automatically
// waits until the rate limit resets before returning.
//
// Parameters:
//   - resp: GitHub API response to check (nil-safe)
//   - limitType: Type of rate limit to check ("search" or "core")
//
// Returns:
//   - true: Rate limit error detected and handled (caller should retry the request)
//   - false: Not a rate limit error or unable to handle (caller should NOT retry)
//
// Behavior when rate limit detected:
//  1. Fetches current rate limit status
//  2. Determines reset time based on limitType
//  3. Calculates wait duration until reset
//  4. Sleeps until reset time (plus small buffer)
//  5. Prints progress messages to stderr
//  6. Returns true to signal retry
//
// The function is fail-safe: if it can't determine rate limit status or reset time,
// it returns false rather than waiting indefinitely.
//
// Example usage:
//
//	result, resp, err := client.SearchCode(query, page)
//	if client.HandleRateLimitError(resp, "search") {
//	    // Retry the same request after waiting
//	    result, resp, err = client.SearchCode(query, page)
//	}
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

// SearchCode searches for code on GitHub using the Code Search API.
//
// Executes a code search query with pagination support. Results include
// file paths, repository information, and content snippets.
//
// Parameters:
//   - query: GitHub code search query (supports all GitHub search qualifiers)
//   - page: Page number for pagination (1-indexed, max 10 pages per query)
//
// Returns:
//   - Search results with matched code files
//   - API response (includes rate limit info, useful for HandleRateLimitError)
//   - Error if search fails
//
// Rate Limits:
// The Code Search API has strict rate limits (30 requests/minute for authenticated users).
// Use HandleRateLimitError to detect and handle rate limit errors automatically.
//
// Results are limited to 100 items per page and 1000 total results per query.
//
// Example:
//
//	result, resp, err := client.SearchCode("minimumLimaVersion extension:yaml", 1)
//	if err != nil {
//	    if client.HandleRateLimitError(resp, "search") {
//	        // Retry after rate limit reset
//	    }
//	}
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

// GetRepository fetches repository information with automatic caching.
//
// Retrieves comprehensive repository metadata including stars, forks, topics,
// license, and timestamps. Responses are cached for 1 hour to minimize API calls.
//
// Parameters:
//   - owner: Repository owner (user or organization name)
//   - repo: Repository name
//
// Returns:
//   - Repository metadata from GitHub API
//   - Error if repository doesn't exist or API call fails
//
// Caching:
// On first call, fetches from GitHub API and stores in cache.
// Subsequent calls within 1 hour return cached data without API call.
// Cache key format: "repo:owner/name"
//
// This significantly reduces API usage when processing multiple templates
// from the same repository (common case in template catalogs).
//
// Example:
//
//	repo, err := client.GetRepository("lima-vm", "lima")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Stars: %d\n", repo.GetStargazersCount())
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

// GetUser fetches user or organization information with automatic caching.
//
// Retrieves user/organization metadata including name, bio, location, blog,
// and email. Works for both individual users and organization accounts.
// Responses are cached for 1 hour to minimize API calls.
//
// Parameters:
//   - login: GitHub username or organization name
//
// Returns:
//   - User/organization metadata from GitHub API
//   - Error if user doesn't exist or API call fails
//
// Caching:
// On first call, fetches from GitHub API and stores in cache.
// Subsequent calls within 1 hour return cached data without API call.
// Cache key format: "user:login"
//
// This significantly reduces API usage when processing multiple templates
// from repositories belonging to the same organization.
//
// Example:
//
//	user, err := client.GetUser("lima-vm")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Type: %s, Name: %s\n", user.GetType(), user.GetName())
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

// ListRepositoryContents lists the contents of a directory in a repository.
//
// Retrieves directory listing from the repository's default branch.
// Returns metadata for each file/subdirectory including name, type, size, and SHA.
//
// Parameters:
//   - owner: Repository owner (user or organization name)
//   - repo: Repository name
//   - path: Directory path within the repository (empty string for root)
//
// Returns:
//   - Slice of directory contents (files and subdirectories)
//   - Error if path doesn't exist, isn't a directory, or API call fails
//
// Note: This function does not use caching as directory contents change frequently.
//
// Example:
//
//	contents, err := client.ListRepositoryContents("lima-vm", "lima", "templates")
//	for _, item := range contents {
//	    if item.GetType() == "file" {
//	        fmt.Printf("File: %s (%d bytes)\n", item.GetName(), item.GetSize())
//	    }
//	}
func (c *Client) ListRepositoryContents(owner, repo, path string) ([]*github.RepositoryContent, error) {
	_, contents, _, err := c.client.Repositories.GetContents(c.ctx, owner, repo, path, nil)
	return contents, err
}

// GetRepositoryContent fetches the content of a single file from a repository.
//
// Retrieves file metadata and base64-encoded content from the repository's default branch.
// Use the returned RepositoryContent's GetContent() method to decode the file content.
//
// Parameters:
//   - owner: Repository owner (user or organization name)
//   - repo: Repository name
//   - path: File path within the repository
//
// Returns:
//   - File metadata and content (base64-encoded, call GetContent() to decode)
//   - Error if file doesn't exist, is a directory, or API call fails
//
// Note: This function does not use caching as file contents can change.
// Files larger than 1MB cannot be retrieved through this API.
//
// Example:
//
//	content, err := client.GetRepositoryContent("lima-vm", "lima", "templates/ubuntu.yaml")
//	if err != nil {
//	    return err
//	}
//	fileContent, err := content.GetContent()
//	if err != nil {
//	    return err
//	}
//	fmt.Println(fileContent)
func (c *Client) GetRepositoryContent(owner, repo, path string) (*github.RepositoryContent, error) {
	content, _, _, err := c.client.Repositories.GetContents(c.ctx, owner, repo, path, nil)
	return content, err
}
