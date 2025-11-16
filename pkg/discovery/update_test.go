package discovery

import (
	"testing"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

func TestMergeTemplates(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name                  string
		existing              []types.Template
		discovered            []types.Template
		expectedAllCount      int
		expectedNewCount      int
		expectedUpdatedCount  int
		expectedUnchangedCount int
		checkResults          func(*testing.T, UpdateResult)
	}{
		{
			name: "Zero discovered templates preserves all existing",
			existing: []types.Template{
				{
					ID:           "owner1/repo1/template1.yaml",
					SHA:          "abc123",
					DiscoveredAt: now.Add(-48 * time.Hour),
					LastChecked:  now.Add(-24 * time.Hour),
				},
				{
					ID:           "owner1/repo1/template2.yaml",
					SHA:          "def456",
					DiscoveredAt: now.Add(-48 * time.Hour),
					LastChecked:  now.Add(-24 * time.Hour),
				},
			},
			discovered:             []types.Template{}, // Empty - nothing discovered
			expectedAllCount:       2,                  // All existing templates preserved
			expectedNewCount:       0,
			expectedUpdatedCount:   0,
			expectedUnchangedCount: 2,
			checkResults: func(t *testing.T, result UpdateResult) {
				if len(result.AllTemplates) != 2 {
					t.Errorf("Expected 2 templates in AllTemplates, got %d", len(result.AllTemplates))
				}
				// Should NOT mark as removed
				if len(result.RemovedTemplates) != 0 {
					t.Errorf("Expected 0 removed templates, got %d", len(result.RemovedTemplates))
				}
			},
		},
		{
			name: "New template added",
			existing: []types.Template{
				{
					ID:           "owner1/repo1/template1.yaml",
					SHA:          "abc123",
					DiscoveredAt: now.Add(-48 * time.Hour),
				},
			},
			discovered: []types.Template{
				{
					ID:           "owner1/repo1/template1.yaml",
					SHA:          "abc123", // Unchanged
					DiscoveredAt: now,
				},
				{
					ID:           "owner1/repo1/template2.yaml",
					SHA:          "def456", // New
					DiscoveredAt: now,
				},
			},
			expectedAllCount:       2,
			expectedNewCount:       1,
			expectedUpdatedCount:   0,
			expectedUnchangedCount: 1,
			checkResults: func(t *testing.T, result UpdateResult) {
				if len(result.NewTemplates) != 1 {
					t.Errorf("Expected 1 new template, got %d", len(result.NewTemplates))
				}
				if len(result.NewTemplates) > 0 && result.NewTemplates[0].ID != "owner1/repo1/template2.yaml" {
					t.Errorf("Expected new template ID 'owner1/repo1/template2.yaml', got '%s'", result.NewTemplates[0].ID)
				}
			},
		},
		{
			name: "Template updated (SHA changed)",
			existing: []types.Template{
				{
					ID:           "owner1/repo1/template1.yaml",
					SHA:          "abc123",
					DiscoveredAt: now.Add(-48 * time.Hour),
				},
			},
			discovered: []types.Template{
				{
					ID:           "owner1/repo1/template1.yaml",
					SHA:          "xyz789", // Changed SHA
					DiscoveredAt: now,
				},
			},
			expectedAllCount:       1,
			expectedNewCount:       0,
			expectedUpdatedCount:   1,
			expectedUnchangedCount: 0,
			checkResults: func(t *testing.T, result UpdateResult) {
				if len(result.UpdatedTemplates) != 1 {
					t.Errorf("Expected 1 updated template, got %d", len(result.UpdatedTemplates))
				}
				if len(result.UpdatedTemplates) > 0 {
					// Should preserve original discovery time
					if result.UpdatedTemplates[0].DiscoveredAt == now {
						t.Error("Updated template should preserve original DiscoveredAt, not use new time")
					}
					// Should have new SHA
					if result.UpdatedTemplates[0].SHA != "xyz789" {
						t.Errorf("Expected updated SHA 'xyz789', got '%s'", result.UpdatedTemplates[0].SHA)
					}
				}
			},
		},
		{
			name: "Template unchanged (same SHA)",
			existing: []types.Template{
				{
					ID:           "owner1/repo1/template1.yaml",
					SHA:          "abc123",
					DiscoveredAt: now.Add(-48 * time.Hour),
					LastChecked:  now.Add(-24 * time.Hour),
				},
			},
			discovered: []types.Template{
				{
					ID:           "owner1/repo1/template1.yaml",
					SHA:          "abc123", // Same SHA
					DiscoveredAt: now,
				},
			},
			expectedAllCount:       1,
			expectedNewCount:       0,
			expectedUpdatedCount:   0,
			expectedUnchangedCount: 1,
			checkResults: func(t *testing.T, result UpdateResult) {
				if result.UnchangedCount != 1 {
					t.Errorf("Expected 1 unchanged template, got %d", result.UnchangedCount)
				}
			},
		},
		{
			name: "Mixed scenario: new, updated, unchanged, not-checked",
			existing: []types.Template{
				{
					ID:           "owner1/repo1/template1.yaml",
					SHA:          "abc123",
					DiscoveredAt: now.Add(-48 * time.Hour),
				},
				{
					ID:           "owner1/repo1/template2.yaml",
					SHA:          "def456",
					DiscoveredAt: now.Add(-48 * time.Hour),
				},
				{
					ID:           "owner1/repo1/template3.yaml",
					SHA:          "ghi789",
					DiscoveredAt: now.Add(-48 * time.Hour),
				},
			},
			discovered: []types.Template{
				{
					ID:           "owner1/repo1/template1.yaml",
					SHA:          "abc999", // Updated
					DiscoveredAt: now,
				},
				{
					ID:           "owner1/repo1/template4.yaml",
					SHA:          "new123", // New
					DiscoveredAt: now,
				},
				// template2 and template3 not in discovered (not checked this run)
			},
			expectedAllCount:       4, // All 4 templates should be in output
			expectedNewCount:       1, // template4
			expectedUpdatedCount:   1, // template1
			expectedUnchangedCount: 2, // template2, template3 (not checked but preserved)
			checkResults: func(t *testing.T, result UpdateResult) {
				if len(result.AllTemplates) != 4 {
					t.Errorf("Expected 4 templates total, got %d", len(result.AllTemplates))
				}
				// Verify template2 and template3 are preserved
				foundTemplate2 := false
				foundTemplate3 := false
				for _, tmpl := range result.AllTemplates {
					if tmpl.ID == "owner1/repo1/template2.yaml" {
						foundTemplate2 = true
					}
					if tmpl.ID == "owner1/repo1/template3.yaml" {
						foundTemplate3 = true
					}
				}
				if !foundTemplate2 {
					t.Error("template2 should be preserved even though not in discovered")
				}
				if !foundTemplate3 {
					t.Error("template3 should be preserved even though not in discovered")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeTemplates(tt.existing, tt.discovered)

			if len(result.AllTemplates) != tt.expectedAllCount {
				t.Errorf("Expected %d total templates, got %d", tt.expectedAllCount, len(result.AllTemplates))
			}

			if len(result.NewTemplates) != tt.expectedNewCount {
				t.Errorf("Expected %d new templates, got %d", tt.expectedNewCount, len(result.NewTemplates))
			}

			if len(result.UpdatedTemplates) != tt.expectedUpdatedCount {
				t.Errorf("Expected %d updated templates, got %d", tt.expectedUpdatedCount, len(result.UpdatedTemplates))
			}

			if result.UnchangedCount != tt.expectedUnchangedCount {
				t.Errorf("Expected %d unchanged templates, got %d", tt.expectedUnchangedCount, result.UnchangedCount)
			}

			if tt.checkResults != nil {
				tt.checkResults(t, result)
			}
		})
	}
}

func TestMergeRepositories_Sorting(t *testing.T) {
	existing := []types.Repository{
		{ID: "org2/repo1", Owner: "org2", Name: "repo1"},
		{ID: "org1/repo2", Owner: "org1", Name: "repo2"},
	}

	collected := []types.Repository{
		{ID: "org1/repo1", Owner: "org1", Name: "repo1"},
	}

	result := MergeRepositories(existing, collected)

	// Should be sorted by owner, then name
	if len(result) != 3 {
		t.Fatalf("Expected 3 repos, got %d", len(result))
	}

	if result[0].ID != "org1/repo1" {
		t.Errorf("Expected first repo 'org1/repo1', got '%s'", result[0].ID)
	}
	if result[1].ID != "org1/repo2" {
		t.Errorf("Expected second repo 'org1/repo2', got '%s'", result[1].ID)
	}
	if result[2].ID != "org2/repo1" {
		t.Errorf("Expected third repo 'org2/repo1', got '%s'", result[2].ID)
	}
}

func TestMergeOrganizations_Sorting(t *testing.T) {
	existing := []types.Organization{
		{ID: "org-c"},
		{ID: "org-a"},
	}

	collected := []types.Organization{
		{ID: "org-b"},
	}

	result := MergeOrganizations(existing, collected)

	// Should be sorted by ID
	if len(result) != 3 {
		t.Fatalf("Expected 3 orgs, got %d", len(result))
	}

	if result[0].ID != "org-a" {
		t.Errorf("Expected first org 'org-a', got '%s'", result[0].ID)
	}
	if result[1].ID != "org-b" {
		t.Errorf("Expected second org 'org-b', got '%s'", result[1].ID)
	}
	if result[2].ID != "org-c" {
		t.Errorf("Expected third org 'org-c', got '%s'", result[2].ID)
	}
}
