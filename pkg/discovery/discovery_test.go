package discovery

import (
	"testing"
	"time"

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

// Note: Additional tests for isLimaTemplate, searchWithQuery, DiscoverCommunityTemplates,
// DiscoverOfficialTemplates, and DiscoverAll would require:
//
// 1. Refactoring Discoverer to accept a GitHubClient interface instead of *github.Client
// 2. Creating a mock implementation of that interface for testing
//
// This is a common testing pattern but requires code changes to enable dependency injection.
// For now, these functions are tested indirectly through integration tests or would need
// interface-based refactoring to achieve unit test coverage.
//
// Recommended refactoring for future test coverage improvement:
//
//   type GitHubClient interface {
//       SearchCode(query string, page int) (*github.CodeSearchResult, *github.Response, error)
//       GetRepository(owner, repo string) (*github.Repository, error)
//       GetRepositoryContent(owner, repo, path string) (*github.RepositoryContent, error)
//       ListRepositoryContents(owner, repo, path string) ([]*github.RepositoryContent, error)
//   }
//
// Then Discoverer could accept GitHubClient interface, allowing mock implementations for testing.
