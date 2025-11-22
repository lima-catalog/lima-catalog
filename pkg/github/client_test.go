package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-github/v57/github"
)

func TestNewClient(t *testing.T) {
	ctx := context.Background()
	token := "test-token"

	client := NewClient(ctx, token)

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.client == nil {
		t.Error("expected GitHub client to be initialized")
	}
	if client.cache == nil {
		t.Error("expected cache to be initialized")
	}
	if client.ctx != ctx {
		t.Error("expected context to be stored")
	}
}

func TestGetClient(t *testing.T) {
	ctx := context.Background()
	client := NewClient(ctx, "test-token")

	ghClient := client.GetClient()

	if ghClient == nil {
		t.Error("expected non-nil GitHub client")
	}
	if ghClient != client.client {
		t.Error("expected GetClient to return the same underlying client")
	}
}

func TestGetRepository_Caching(t *testing.T) {
	// This test verifies that GetRepository uses caching correctly
	ctx := context.Background()

	// Create a test server that counts requests
	requestCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return a minimal repository JSON
		_, _ = w.Write([]byte(`{
			"id": 123,
			"name": "test-repo",
			"full_name": "test-owner/test-repo",
			"owner": {"login": "test-owner"},
			"description": "Test repository",
			"stargazers_count": 42,
			"topics": ["test", "example"]
		}`))
	}))
	defer ts.Close()

	// Create client pointing to test server
	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache, // Use real cache
	}

	// First call - should hit the API
	repo1, err := ghClient.GetRepository("test-owner", "test-repo")
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}
	if repo1 == nil {
		t.Fatal("expected non-nil repository")
	}
	if requestCount != 1 {
		t.Errorf("expected 1 request after first call, got %d", requestCount)
	}

	// Second call - should use cache
	repo2, err := ghClient.GetRepository("test-owner", "test-repo")
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if repo2 == nil {
		t.Fatal("expected non-nil repository on cached call")
	}
	if requestCount != 1 {
		t.Errorf("expected 1 request after cached call, got %d (cache not working)", requestCount)
	}

	// Verify it's the same repository
	if repo1.GetID() != repo2.GetID() {
		t.Error("expected same repository from cache")
	}
}

func TestGetUser_Caching(t *testing.T) {
	ctx := context.Background()

	// Create a test server that counts requests
	requestCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return a minimal user JSON
		_, _ = w.Write([]byte(`{
			"id": 456,
			"login": "test-user",
			"name": "Test User",
			"email": "test@example.com",
			"bio": "Test bio",
			"blog": "https://example.com"
		}`))
	}))
	defer ts.Close()

	// Create client pointing to test server
	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache,
	}

	// First call - should hit the API
	user1, err := ghClient.GetUser("test-user")
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}
	if user1 == nil {
		t.Fatal("expected non-nil user")
	}
	if requestCount != 1 {
		t.Errorf("expected 1 request after first call, got %d", requestCount)
	}

	// Second call - should use cache
	user2, err := ghClient.GetUser("test-user")
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if user2 == nil {
		t.Fatal("expected non-nil user on cached call")
	}
	if requestCount != 1 {
		t.Errorf("expected 1 request after cached call, got %d (cache not working)", requestCount)
	}

	// Verify it's the same user
	if user1.GetID() != user2.GetID() {
		t.Error("expected same user from cache")
	}
}

func TestGetRepository_Error(t *testing.T) {
	ctx := context.Background()

	// Create a test server that returns an error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer ts.Close()

	// Create client pointing to test server
	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache,
	}

	// Call should return error
	repo, err := ghClient.GetRepository("nonexistent", "repo")
	if err == nil {
		t.Error("expected error for nonexistent repository")
	}
	if repo != nil {
		t.Error("expected nil repository on error")
	}
}

func TestCheckRateLimit(t *testing.T) {
	tests := []struct {
		name              string
		remaining         int
		minimumRemaining  int
		expectError       bool
	}{
		{
			name:             "Sufficient rate limit",
			remaining:        500,
			minimumRemaining: 100,
			expectError:      false,
		},
		{
			name:             "Exactly at minimum",
			remaining:        100,
			minimumRemaining: 100,
			expectError:      false,
		},
		{
			name:             "Below minimum",
			remaining:        50,
			minimumRemaining: 100,
			expectError:      true,
		},
		{
			name:             "Zero remaining",
			remaining:        0,
			minimumRemaining: 100,
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Create a test server that returns rate limit info
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				resetTime := time.Now().Add(1 * time.Hour).Unix()
				response := fmt.Sprintf(`{
					"resources": {
						"core": {
							"limit": 5000,
							"remaining": %d,
							"reset": %d,
							"used": %d
						},
						"search": {
							"limit": 30,
							"remaining": 29,
							"reset": %d,
							"used": 1
						}
					}
				}`, tt.remaining, resetTime, 5000-tt.remaining, resetTime)
				_, _ = w.Write([]byte(response))
			}))
			defer ts.Close()

			// Create client pointing to test server
			client := github.NewClient(nil)
			client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

			ghClient := &Client{
				client: client,
				ctx:    ctx,
				cache:  NewClient(ctx, "token").cache,
			}

			err := ghClient.CheckRateLimit(tt.minimumRemaining)
			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestCheckRateLimit_APIError(t *testing.T) {
	ctx := context.Background()

	// Create a test server that returns an error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message": "Internal Server Error"}`))
	}))
	defer ts.Close()

	// Create client pointing to test server
	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache,
	}

	err := ghClient.CheckRateLimit(100)
	if err == nil {
		t.Error("expected error when API fails")
	}
}

func TestHandleRateLimitError(t *testing.T) {
	ctx := context.Background()
	client := NewClient(ctx, "test-token")

	tests := []struct {
		name        string
		resp        *github.Response
		limitType   string
		expectError error
	}{
		{
			name:        "Nil response",
			resp:        nil,
			limitType:   "core",
			expectError: nil,
		},
		{
			name: "Non-rate-limit error (200 OK)",
			resp: &github.Response{
				Response: &http.Response{
					StatusCode: 200,
				},
			},
			limitType:   "core",
			expectError: nil,
		},
		{
			name: "Non-rate-limit error (404)",
			resp: &github.Response{
				Response: &http.Response{
					StatusCode: 404,
				},
			},
			limitType:   "core",
			expectError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.HandleRateLimitError(tt.resp, tt.limitType)
			if err != tt.expectError {
				t.Errorf("expected HandleRateLimitError to return %v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestHandleRateLimitError_UnknownLimitType(t *testing.T) {
	ctx := context.Background()

	// Mock server that returns rate limit info
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resetTime := time.Now().Add(1 * time.Hour).Unix()
		response := fmt.Sprintf(`{
			"resources": {
				"core": {
					"limit": 5000,
					"remaining": 0,
					"reset": %d,
					"used": 5000
				}
			}
		}`, resetTime)
		_, _ = w.Write([]byte(response))
	}))
	defer ts.Close()

	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache,
	}

	// Test with unknown limit type - should return nil without waiting
	resp := &github.Response{
		Response: &http.Response{
			StatusCode: 403,
		},
	}
	err := ghClient.HandleRateLimitError(resp, "unknown")
	if err != nil {
		t.Errorf("expected nil for unknown limit type, got %v", err)
	}
}

func TestHandleRateLimitError_RateLimitAPIError(t *testing.T) {
	ctx := context.Background()

	// Mock server that returns error when checking rate limits
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message": "Internal Server Error"}`))
	}))
	defer ts.Close()

	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache,
	}

	// Test with 403 response but rate limit API fails
	resp := &github.Response{
		Response: &http.Response{
			StatusCode: 403,
		},
	}
	err := ghClient.HandleRateLimitError(resp, "core")
	if err != nil {
		t.Errorf("expected nil when rate limit API fails (fail-safe), got %v", err)
	}
}

func TestHandleRateLimitError_ZeroResetTime(t *testing.T) {
	ctx := context.Background()

	// Mock server that returns rate limit info with zero reset time
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"resources": {
				"core": {
					"limit": 5000,
					"remaining": 0,
					"reset": 0,
					"used": 5000
				}
			}
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer ts.Close()

	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache,
	}

	// Test with 403 response but zero reset time
	resp := &github.Response{
		Response: &http.Response{
			StatusCode: 403,
		},
	}
	err := ghClient.HandleRateLimitError(resp, "core")
	if err != nil {
		t.Errorf("expected nil for zero reset time, got %v", err)
	}
}

func TestHandleRateLimitError_NegativeWaitDuration(t *testing.T) {
	ctx := context.Background()

	// Mock server that returns rate limit info with reset time in the past
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resetTime := time.Now().Add(-1 * time.Hour).Unix()
		response := fmt.Sprintf(`{
			"resources": {
				"core": {
					"limit": 5000,
					"remaining": 0,
					"reset": %d,
					"used": 5000
				},
				"search": {
					"limit": 30,
					"remaining": 0,
					"reset": %d,
					"used": 30
				}
			}
		}`, resetTime, resetTime)
		_, _ = w.Write([]byte(response))
	}))
	defer ts.Close()

	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache,
	}

	tests := []struct {
		name       string
		statusCode int
		limitType  string
	}{
		{
			name:       "403 Forbidden with core limit",
			statusCode: 403,
			limitType:  "core",
		},
		{
			name:       "429 Too Many Requests with search limit",
			statusCode: 429,
			limitType:  "search",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &github.Response{
				Response: &http.Response{
					StatusCode: tt.statusCode,
				},
			}

			// Should return nil when wait duration is negative (reset already passed)
			err := ghClient.HandleRateLimitError(resp, tt.limitType)
			if err != nil {
				t.Errorf("expected nil for negative wait duration, got %v", err)
			}
		})
	}
}

func TestCacheKeyFormat(t *testing.T) {
	// This test verifies that cache keys are formatted correctly
	// by checking that different repos/users get different cache entries

	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Return different IDs based on the path
		switch r.URL.Path {
		case "/repos/owner1/repo1":
			_, _ = w.Write([]byte(`{"id": 1, "name": "repo1", "full_name": "owner1/repo1", "owner": {"login": "owner1"}}`))
		case "/repos/owner2/repo2":
			_, _ = w.Write([]byte(`{"id": 2, "name": "repo2", "full_name": "owner2/repo2", "owner": {"login": "owner2"}}`))
		case "/users/user1":
			_, _ = w.Write([]byte(`{"id": 10, "login": "user1"}`))
		case "/users/user2":
			_, _ = w.Write([]byte(`{"id": 20, "login": "user2"}`))
		}
	}))
	defer ts.Close()

	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache,
	}

	// Fetch two different repos
	repo1, _ := ghClient.GetRepository("owner1", "repo1")
	repo2, _ := ghClient.GetRepository("owner2", "repo2")

	if repo1 == nil || repo2 == nil {
		t.Fatal("expected both repos to be fetched")
	}

	if repo1.GetID() == repo2.GetID() {
		t.Error("expected different repos to have different IDs")
	}

	// Fetch two different users
	user1, _ := ghClient.GetUser("user1")
	user2, _ := ghClient.GetUser("user2")

	if user1 == nil || user2 == nil {
		t.Fatal("expected both users to be fetched")
	}

	if user1.GetID() == user2.GetID() {
		t.Error("expected different users to have different IDs")
	}
}

func TestRateLimit(t *testing.T) {
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rate_limit" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resetTime := time.Now().Add(1 * time.Hour).Unix()
		response := `{
			"resources": {
				"core": {
					"limit": 5000,
					"remaining": 4999,
					"reset": ` + fmt.Sprintf("%d", resetTime) + `,
					"used": 1
				},
				"search": {
					"limit": 30,
					"remaining": 29,
					"reset": ` + fmt.Sprintf("%d", resetTime) + `,
					"used": 1
				}
			}
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer ts.Close()

	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache,
	}

	limits, err := ghClient.RateLimit()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if limits == nil {
		t.Fatal("expected non-nil rate limits")
	}

	if limits.Core == nil {
		t.Fatal("expected non-nil core rate limit")
	}

	if limits.Search == nil {
		t.Fatal("expected non-nil search rate limit")
	}
}

func TestSearchCode(t *testing.T) {
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/code" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query().Get("q")
		if query == "" {
			t.Error("expected query parameter")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"total_count": 1,
			"incomplete_results": false,
			"items": [
				{
					"name": "test.yaml",
					"path": "templates/test.yaml",
					"sha": "abc123",
					"html_url": "https://github.com/owner/repo/blob/main/templates/test.yaml",
					"repository": {
						"id": 123,
						"name": "repo",
						"full_name": "owner/repo",
						"owner": {"login": "owner"}
					}
				}
			]
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer ts.Close()

	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache,
	}

	result, resp, err := ghClient.SearchCode("test query", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil search result")
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if result.GetTotal() != 1 {
		t.Errorf("expected total count of 1, got %d", result.GetTotal())
	}

	if len(result.CodeResults) != 1 {
		t.Errorf("expected 1 code result, got %d", len(result.CodeResults))
	}
}

func TestListRepositoryContents(t *testing.T) {
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/contents/templates" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `[
			{
				"name": "ubuntu.yaml",
				"path": "templates/ubuntu.yaml",
				"type": "file",
				"size": 1024,
				"sha": "abc123",
				"html_url": "https://github.com/owner/repo/blob/main/templates/ubuntu.yaml"
			},
			{
				"name": "alpine.yaml",
				"path": "templates/alpine.yaml",
				"type": "file",
				"size": 512,
				"sha": "def456",
				"html_url": "https://github.com/owner/repo/blob/main/templates/alpine.yaml"
			}
		]`
		_, _ = w.Write([]byte(response))
	}))
	defer ts.Close()

	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache,
	}

	contents, err := ghClient.ListRepositoryContents("owner", "repo", "templates")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if contents == nil {
		t.Fatal("expected non-nil contents")
	}

	if len(contents) != 2 {
		t.Errorf("expected 2 contents, got %d", len(contents))
	}

	if contents[0].GetName() != "ubuntu.yaml" {
		t.Errorf("expected first file to be ubuntu.yaml, got %s", contents[0].GetName())
	}

	if contents[1].GetName() != "alpine.yaml" {
		t.Errorf("expected second file to be alpine.yaml, got %s", contents[1].GetName())
	}
}

func TestListRepositoryContents_Error(t *testing.T) {
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer ts.Close()

	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache,
	}

	contents, err := ghClient.ListRepositoryContents("owner", "repo", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}

	if contents != nil {
		t.Error("expected nil contents on error")
	}
}

func TestGetRepositoryContent(t *testing.T) {
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/contents/templates/test.yaml" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Base64 encoded "test content"
		response := `{
			"name": "test.yaml",
			"path": "templates/test.yaml",
			"type": "file",
			"size": 12,
			"sha": "abc123",
			"content": "dGVzdCBjb250ZW50",
			"encoding": "base64",
			"html_url": "https://github.com/owner/repo/blob/main/templates/test.yaml"
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer ts.Close()

	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache,
	}

	content, err := ghClient.GetRepositoryContent("owner", "repo", "templates/test.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if content == nil {
		t.Fatal("expected non-nil content")
	}

	if content.GetName() != "test.yaml" {
		t.Errorf("expected name test.yaml, got %s", content.GetName())
	}

	if content.GetPath() != "templates/test.yaml" {
		t.Errorf("expected path templates/test.yaml, got %s", content.GetPath())
	}

	// Decode content
	decoded, err := content.GetContent()
	if err != nil {
		t.Fatalf("failed to decode content: %v", err)
	}

	if decoded != "test content" {
		t.Errorf("expected decoded content 'test content', got %q", decoded)
	}
}

func TestGetRepositoryContent_Error(t *testing.T) {
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer ts.Close()

	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(ts.URL + "/")

	ghClient := &Client{
		client: client,
		ctx:    ctx,
		cache:  NewClient(ctx, "token").cache,
	}

	content, err := ghClient.GetRepositoryContent("owner", "repo", "nonexistent.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}

	if content != nil {
		t.Error("expected nil content on error")
	}
}
