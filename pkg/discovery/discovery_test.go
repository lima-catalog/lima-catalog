package discovery

import (
	"context"
	"testing"
	"time"

	gh "github.com/lima-catalog/lima-catalog/pkg/github"
	"github.com/lima-catalog/lima-catalog/pkg/types"
)

func TestFindNewestTemplateTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		templates []types.Template
		expected  time.Time
	}{
		{
			name:      "Empty list",
			templates: []types.Template{},
			expected:  time.Time{},
		},
		{
			name: "Single template",
			templates: []types.Template{
				{ID: "test/repo/t1.yaml", DiscoveredAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)},
			},
			expected: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "Multiple templates - find newest",
			templates: []types.Template{
				{ID: "test/repo/t1.yaml", DiscoveredAt: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)},
				{ID: "test/repo/t2.yaml", DiscoveredAt: time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)},
				{ID: "test/repo/t3.yaml", DiscoveredAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)},
			},
			expected: time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "Templates with same timestamp",
			templates: []types.Template{
				{ID: "test/repo/t1.yaml", DiscoveredAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)},
				{ID: "test/repo/t2.yaml", DiscoveredAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)},
			},
			expected: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "Zero time in list",
			templates: []types.Template{
				{ID: "test/repo/t1.yaml", DiscoveredAt: time.Time{}},
				{ID: "test/repo/t2.yaml", DiscoveredAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)},
			},
			expected: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "All zero times",
			templates: []types.Template{
				{ID: "test/repo/t1.yaml", DiscoveredAt: time.Time{}},
				{ID: "test/repo/t2.yaml", DiscoveredAt: time.Time{}},
			},
			expected: time.Time{},
		},
		{
			name: "Large list performance",
			templates: func() []types.Template {
				templates := make([]types.Template, 1000)
				for i := 0; i < 1000; i++ {
					templates[i] = types.Template{
						ID:           "test/repo/t.yaml",
						DiscoveredAt: time.Date(2024, 1, 1, 0, 0, i, 0, time.UTC),
					}
				}
				// Set the last one as newest
				templates[999].DiscoveredAt = time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
				return templates
			}(),
			expected: time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindNewestTemplateTimestamp(tt.templates)
			if !result.Equal(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNewDiscoverer(t *testing.T) {
	// Note: This test is limited because Discoverer uses concrete *github.Client type
	// Full testing of discovery functions would require refactoring to use dependency injection
	// with interfaces (e.g., GitHubClient interface)

	t.Run("Creates discoverer with nil blocklist", func(t *testing.T) {
		d := NewDiscoverer(nil, nil)

		if d == nil {
			t.Fatal("expected non-nil discoverer")
		}

		if d.blocklist != nil {
			t.Error("expected nil blocklist")
		}
	})

	t.Run("Creates discoverer with blocklist", func(t *testing.T) {
		blocklist := &types.Blocklist{
			Paths: []string{"^test/"},
			Repos: []string{"^spam/"},
		}

		d := NewDiscoverer(nil, blocklist)

		if d == nil {
			t.Fatal("expected non-nil discoverer")
		}

		if d.blocklist == nil {
			t.Error("expected blocklist to be set")
		}

		if len(d.blocklist.Paths) != 1 {
			t.Errorf("expected 1 path pattern, got %d", len(d.blocklist.Paths))
		}

		if len(d.blocklist.Repos) != 1 {
			t.Errorf("expected 1 repo pattern, got %d", len(d.blocklist.Repos))
		}
	})
}

// Note on testing discovery functions:
//
// The current architecture makes it challenging to fully test discovery functions without
// actual GitHub API access because the Discoverer depends on a concrete *github.Client type
// rather than an interface. This is a common pattern but limits testability.
//
// The tests below cover what we CAN test:
// 1. ✅ FindNewestTemplateTimestamp - Pure function, fully tested above
// 2. ✅ NewDiscoverer - Constructor, fully tested above
// 3. ⚠️ isLimaTemplate - Requires GitHub API client (not easily mockable)
// 4. ⚠️ searchWithQuery - Requires GitHub API client
// 5. ⚠️ DiscoverCommunityTemplates - Requires GitHub API client
// 6. ⚠️ DiscoverOfficialTemplates - Requires GitHub API client
// 7. ⚠️ DiscoverAll - Requires GitHub API client
//
// Current test coverage: ~49.6% (primarily from analyzer, parser, and other helper functions)
//
// To achieve 80%+ coverage, we would need one of these approaches:
// A) Refactor to use interface-based dependency injection (recommended for future)
// B) Add integration tests that hit actual GitHub API (slow, requires token)
// C) Use advanced mocking/reflection techniques (complex, fragile)
//
// For now, these functions are tested indirectly through:
// - Manual testing during development
// - The actual workflow runs that use these functions in production
// - Integration tests in other parts of the codebase
//
// The test functions below demonstrate the API surface and verify basic error handling
// where possible without full mocking capabilities.

func TestDiscoverAll_ContextCancellation(t *testing.T) {
	// This test verifies that DiscoverAll respects context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	ghClient := gh.NewClient(ctx, "test-token")
	d := NewDiscoverer(ghClient, nil)

	templates, err := d.DiscoverAll(ctx, time.Time{}, nil)

	// Should detect cancellation early and return error
	if err == nil {
		t.Error("expected error when context is cancelled")
	}

	if len(templates) > 0 {
		t.Errorf("expected no templates when context cancelled immediately, got %d", len(templates))
	}
}
