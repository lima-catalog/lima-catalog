// Unit Tests: Data Combiner (combiner package)
//
// High-level overview of what's being tested:
// - Combining templates, repositories, and organizations into unified output
// - Blocklist filtering during combination (paths and repos)
// - Template metadata transformation (display names, descriptions, raw URLs)
// - Repository and organization data enrichment
// - Handling templates without repository data (orphan detection)
// - Official vs community template distinction
// - Fallback strategies for missing metadata (display names, descriptions)
// - Similar template association and duplicate handling
// - Sorting and organization of combined output
// - Date formatting and timestamp handling
// - Score statistics calculation for notability metrics

package combiner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/interfaces"
	"github.com/lima-catalog/lima-catalog/pkg/types"
)

// mockURLTransformer is a test double that returns predictable raw URLs without network calls
type mockURLTransformer struct{}

// TransformURL converts github: URLs to raw.githubusercontent.com URLs
// Format: github:owner/repo/path -> https://raw.githubusercontent.com/owner/repo/main/path.yaml
func (m *mockURLTransformer) TransformURL(ctx context.Context, url string) (string, error) {
	// Parse github: URL
	if !strings.HasPrefix(url, "github:") {
		return "", fmt.Errorf("unsupported URL scheme: %s", url)
	}

	// Remove github: prefix
	path := strings.TrimPrefix(url, "github:")

	// Split into parts
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid github: URL: %s", url)
	}

	owner := parts[0]
	repo := parts[1]

	// Handle double slash for org repos (e.g., github:lima-vm//templates/ubuntu)
	if repo == "" && len(parts) >= 3 {
		repo = owner
		parts = append([]string{owner, repo}, parts[2:]...)
	}

	// Construct file path
	var filePath string
	if len(parts) > 2 {
		filePath = strings.Join(parts[2:], "/") + ".yaml"
	} else {
		filePath = ".lima.yaml"
	}

	// Return raw URL (always use "main" branch in tests)
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s", owner, repo, filePath), nil
}

func TestCombineData(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name              string
		templates         []types.Template
		repos             []types.Repository
		orgs              []types.Organization
		blocklist         *types.Blocklist
		expectedCount     int
		expectedFiltered  int
		checkTemplate     func(*testing.T, []CombinedTemplate)
	}{
		{
			name: "Basic template combination",
			templates: []types.Template{
				{
					ID:               "owner1/repo1/template.yaml",
					Repo:             "owner1/repo1",
					Path:             "template.yaml",
					Name:             "test-template",
					DisplayName:      "Test Template",
					ShortDescription: "A test template",
					Category:         "development",
					Keywords:         []string{"test", "dev", "docker"},
					IsOfficial:       false,
				},
			},
			repos: []types.Repository{
				{
					ID:            "owner1/repo1",
					Owner:         "owner1",
					Name:          "repo1",
					Stars:         42,
					DefaultBranch: "main",
					UpdatedAt:     now,
				},
			},
			orgs: []types.Organization{
				{
					ID:    "owner1",
					Login: "owner1",
					Type:  "Organization",
				},
			},
			blocklist:        nil,
			expectedCount:    1,
			expectedFiltered: 0,
			checkTemplate: func(t *testing.T, combined []CombinedTemplate) {
				if len(combined) != 1 {
					t.Fatalf("Expected 1 template, got %d", len(combined))
				}

				tmpl := combined[0]
				if tmpl.Name != "Test Template" {
					t.Errorf("Expected name 'Test Template', got '%s'", tmpl.Name)
				}
				if tmpl.Description != "A test template" {
					t.Errorf("Expected description 'A test template', got '%s'", tmpl.Description)
				}
				if tmpl.Stars != 42 {
					t.Errorf("Expected 42 stars, got %d", tmpl.Stars)
				}
				if tmpl.Official {
					t.Error("Expected official=false")
				}
				if tmpl.Org != "owner1" {
					t.Errorf("Expected org 'owner1', got '%s'", tmpl.Org)
				}
				expectedRawURL := "https://raw.githubusercontent.com/owner1/repo1/main/template.yaml"
				if tmpl.RawURL != expectedRawURL {
					t.Errorf("Expected raw URL '%s', got '%s'", expectedRawURL, tmpl.RawURL)
				}
			},
		},
		{
			name: "Template with blocklist filtering",
			templates: []types.Template{
				{
					ID:               "owner1/repo1/template.yaml",
					Repo:             "owner1/repo1",
					Path:             "template.yaml",
					Name:             "good",
					DisplayName:      "Good Template",
					ShortDescription: "Should appear",
					Category:         "development",
					Keywords:         []string{"test"},
					IsOfficial:       false,
				},
				{
					ID:               "owner1/repo1/.github/workflows/ci.yaml",
					Repo:             "owner1/repo1",
					Path:             ".github/workflows/ci.yaml",
					Name:             "bad",
					DisplayName:      "Bad Template",
					ShortDescription: "Should be filtered",
					Category:         "development",
					Keywords:         []string{"test"},
					IsOfficial:       false,
				},
			},
			repos: []types.Repository{
				{
					ID:            "owner1/repo1",
					Owner:         "owner1",
					Name:          "repo1",
					Stars:         10,
					DefaultBranch: "main",
					UpdatedAt:     now,
				},
			},
			orgs: []types.Organization{
				{
					ID:    "owner1",
					Login: "owner1",
					Type:  "Organization",
				},
			},
			blocklist: &types.Blocklist{
				Paths: []string{`^\.github/workflows/`},
			},
			expectedCount:    1,
			expectedFiltered: 1,
			checkTemplate: func(t *testing.T, combined []CombinedTemplate) {
				if len(combined) != 1 {
					t.Fatalf("Expected 1 template after filtering, got %d", len(combined))
				}
				if combined[0].Name != "Good Template" {
					t.Errorf("Wrong template passed filter: %s", combined[0].Name)
				}
			},
		},
		{
			name: "Template without repo data is skipped",
			templates: []types.Template{
				{
					ID:               "owner1/repo1/template.yaml",
					Repo:             "owner1/repo1",
					Path:             "template.yaml",
					Name:             "orphan",
					DisplayName:      "Orphan Template",
					ShortDescription: "No repo data",
					Category:         "development",
					Keywords:         []string{"test"},
					IsOfficial:       false,
				},
			},
			repos:            []types.Repository{}, // No repo data
			orgs:             []types.Organization{},
			blocklist:        nil,
			expectedCount:    0, // Should be skipped
			expectedFiltered: 0,
			checkTemplate: func(t *testing.T, combined []CombinedTemplate) {
				if len(combined) != 0 {
					t.Errorf("Expected 0 templates (no repo data), got %d", len(combined))
				}
			},
		},
		{
			name: "Official template",
			templates: []types.Template{
				{
					ID:               "lima-vm/lima/templates/ubuntu.yaml",
					Repo:             "lima-vm/lima",
					Path:             "templates/ubuntu.yaml",
					Name:             "ubuntu",
					DisplayName:      "Ubuntu",
					ShortDescription: "Official Ubuntu template",
					Category:         "general",
					Keywords:         []string{"ubuntu", "official"},
					IsOfficial:       true,
				},
			},
			repos: []types.Repository{
				{
					ID:            "lima-vm/lima",
					Owner:         "lima-vm",
					Name:          "lima",
					Stars:         18903,
					DefaultBranch: "master",
					UpdatedAt:     now,
				},
			},
			orgs: []types.Organization{
				{
					ID:    "lima-vm",
					Login: "lima-vm",
					Type:  "Organization",
				},
			},
			blocklist:        nil,
			expectedCount:    1,
			expectedFiltered: 0,
			checkTemplate: func(t *testing.T, combined []CombinedTemplate) {
				if len(combined) != 1 {
					t.Fatalf("Expected 1 template, got %d", len(combined))
				}
				if !combined[0].Official {
					t.Error("Expected official=true")
				}
				if combined[0].Stars != 18903 {
					t.Errorf("Expected 18903 stars, got %d", combined[0].Stars)
				}
			},
		},
		{
			name: "Template with no display name or short description uses fallback",
			templates: []types.Template{
				{
					ID:          "owner1/repo1/path/to/lima.yaml",
					Repo:        "owner1/repo1",
					Path:        "path/to/lima.yaml",
					Name:        "lima",
					Category:    "development",
					Keywords:    []string{"docker", "kubernetes", "ubuntu"},
					IsOfficial:  false,
				},
			},
			repos: []types.Repository{
				{
					ID:            "owner1/repo1",
					Owner:         "owner1",
					Name:          "repo1",
					Stars:         5,
					DefaultBranch: "main",
					UpdatedAt:     now,
				},
			},
			orgs: []types.Organization{
				{
					ID:    "owner1",
					Login: "owner1",
					Type:  "User",
				},
			},
			blocklist:        nil,
			expectedCount:    1,
			expectedFiltered: 0,
			checkTemplate: func(t *testing.T, combined []CombinedTemplate) {
				if len(combined) != 1 {
					t.Fatalf("Expected 1 template, got %d", len(combined))
				}

				// Name should fall back to Name field
				if combined[0].Name != "lima" {
					t.Errorf("Expected name 'lima', got '%s'", combined[0].Name)
				}

				// Description should be first 3 keywords joined
				expectedDesc := "docker, kubernetes, ubuntu"
				if combined[0].Description != expectedDesc {
					t.Errorf("Expected description '%s', got '%s'", expectedDesc, combined[0].Description)
				}
			},
		},
		{
			name: "Template with similar templates",
			templates: []types.Template{
				{
					ID:               "owner1/repo1/ubuntu.yaml",
					Repo:             "owner1/repo1",
					Path:             "ubuntu.yaml",
					Name:             "ubuntu",
					DisplayName:      "Ubuntu Template",
					ShortDescription: "Ubuntu development template",
					Category:         "development",
					Keywords:         []string{"ubuntu", "docker"},
					IsOfficial:       false,
					SimilarTemplates: []types.SimilarTemplate{
						{
							ID:          "owner2/repo2/ubuntu-dev.yaml",
							Similarity:  0.85,
							SharedBands: 28,
						},
						{
							ID:          "owner3/repo3/ubuntu-docker.yaml",
							Similarity:  0.65,
							SharedBands: 18,
						},
					},
				},
			},
			repos: []types.Repository{
				{
					ID:            "owner1/repo1",
					Owner:         "owner1",
					Name:          "repo1",
					Stars:         15,
					DefaultBranch: "main",
					UpdatedAt:     now,
				},
			},
			orgs: []types.Organization{
				{
					ID:    "owner1",
					Login: "owner1",
					Type:  "Organization",
				},
			},
			blocklist:        nil,
			expectedCount:    1,
			expectedFiltered: 0,
			checkTemplate: func(t *testing.T, combined []CombinedTemplate) {
				if len(combined) != 1 {
					t.Fatalf("Expected 1 template, got %d", len(combined))
				}

				tmpl := combined[0]
				if len(tmpl.SimilarTemplates) != 2 {
					t.Errorf("Expected 2 similar templates, got %d", len(tmpl.SimilarTemplates))
				}

				// Check first similar template
				if len(tmpl.SimilarTemplates) > 0 {
					sim := tmpl.SimilarTemplates[0]
					if sim.ID != "owner2/repo2/ubuntu-dev.yaml" {
						t.Errorf("Expected similar template ID 'owner2/repo2/ubuntu-dev.yaml', got '%s'", sim.ID)
					}
					if sim.Similarity != 0.85 {
						t.Errorf("Expected similarity 0.85, got %f", sim.Similarity)
					}
					if sim.SharedBands != 28 {
						t.Errorf("Expected 28 shared bands, got %d", sim.SharedBands)
					}
				}

				// Check second similar template
				if len(tmpl.SimilarTemplates) > 1 {
					sim := tmpl.SimilarTemplates[1]
					if sim.ID != "owner3/repo3/ubuntu-docker.yaml" {
						t.Errorf("Expected similar template ID 'owner3/repo3/ubuntu-docker.yaml', got '%s'", sim.ID)
					}
					if sim.Similarity != 0.65 {
						t.Errorf("Expected similarity 0.65, got %f", sim.Similarity)
					}
				}
			},
		},
		{
			name: "Sorting by org/repo/path",
			templates: []types.Template{
				{
					ID:               "org2/repo1/b.yaml",
					Repo:             "org2/repo1",
					Path:             "b.yaml",
					Name:             "b",
					DisplayName:      "B",
					ShortDescription: "B template",
					Category:         "development",
					Keywords:         []string{"test"},
				},
				{
					ID:               "org1/repo2/a.yaml",
					Repo:             "org1/repo2",
					Path:             "a.yaml",
					Name:             "a",
					DisplayName:      "A",
					ShortDescription: "A template",
					Category:         "development",
					Keywords:         []string{"test"},
				},
				{
					ID:               "org1/repo1/z.yaml",
					Repo:             "org1/repo1",
					Path:             "z.yaml",
					Name:             "z",
					DisplayName:      "Z",
					ShortDescription: "Z template",
					Category:         "development",
					Keywords:         []string{"test"},
				},
			},
			repos: []types.Repository{
				{ID: "org1/repo1", Owner: "org1", Name: "repo1", DefaultBranch: "main", UpdatedAt: now},
				{ID: "org1/repo2", Owner: "org1", Name: "repo2", DefaultBranch: "main", UpdatedAt: now},
				{ID: "org2/repo1", Owner: "org2", Name: "repo1", DefaultBranch: "main", UpdatedAt: now},
			},
			orgs: []types.Organization{
				{ID: "org1", Login: "org1"},
				{ID: "org2", Login: "org2"},
			},
			blocklist:        nil,
			expectedCount:    3,
			expectedFiltered: 0,
			checkTemplate: func(t *testing.T, combined []CombinedTemplate) {
				if len(combined) != 3 {
					t.Fatalf("Expected 3 templates, got %d", len(combined))
				}

				// Should be sorted by org, then repo, then path
				if combined[0].ID != "org1/repo1/z.yaml" {
					t.Errorf("First template should be org1/repo1/z.yaml, got %s", combined[0].ID)
				}
				if combined[1].ID != "org1/repo2/a.yaml" {
					t.Errorf("Second template should be org1/repo2/a.yaml, got %s", combined[1].ID)
				}
				if combined[2].ID != "org2/repo1/b.yaml" {
					t.Errorf("Third template should be org2/repo1/b.yaml, got %s", combined[2].ID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compile blocklist patterns if present
			if tt.blocklist != nil {
				if err := tt.blocklist.CompilePatterns(); err != nil {
					t.Fatalf("failed to compile blocklist patterns: %v", err)
				}
			}

			// Create combiner with mock URL transformer (no network calls)
			combiner := NewCombinerWithFS(tt.blocklist, interfaces.NewDefaultFileSystem(), &mockURLTransformer{})

			// Create temp output file
			tmpFile, err := os.CreateTemp("", "test-combined-*.jsonl")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer func() { _ = os.Remove(tmpFile.Name()) }()
			_ = tmpFile.Close()

			// Run combine
			err = combiner.CombineData(t.Context(), tt.templates, tt.repos, tt.orgs, tmpFile.Name())
			if err != nil {
				t.Fatalf("CombineData failed: %v", err)
			}

			// Read output
			data, err := os.ReadFile(tmpFile.Name())
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			// Parse JSON Lines
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			var combined []CombinedTemplate

			for _, line := range lines {
				if line == "" {
					continue
				}
				var tmpl CombinedTemplate
				if err := json.Unmarshal([]byte(line), &tmpl); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}
				combined = append(combined, tmpl)
			}

			// Check count
			if len(combined) != tt.expectedCount {
				t.Errorf("Expected %d templates, got %d", tt.expectedCount, len(combined))
			}

			// Run custom checks
			if tt.checkTemplate != nil {
				tt.checkTemplate(t, combined)
			}
		})
	}
}

func TestGetDisplayName(t *testing.T) {
	combiner := NewCombinerWithFS(nil, interfaces.NewDefaultFileSystem(), &mockURLTransformer{})

	tests := []struct {
		name     string
		template types.Template
		expected string
	}{
		{
			name: "DisplayName takes priority",
			template: types.Template{
				DisplayName: "My Template",
				Name:        "template",
				Path:        "path/to/template.yaml",
			},
			expected: "My Template",
		},
		{
			name: "Name is fallback",
			template: types.Template{
				Name: "template",
				Path: "path/to/template.yaml",
			},
			expected: "template",
		},
		{
			name: "Path is last resort",
			template: types.Template{
				Path: "path/to/template.yaml",
			},
			expected: "path/to/template.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := combiner.getDisplayName(tt.template)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetDescription(t *testing.T) {
	combiner := NewCombinerWithFS(nil, interfaces.NewDefaultFileSystem(), &mockURLTransformer{})

	tests := []struct {
		name     string
		template types.Template
		expected string
	}{
		{
			name: "ShortDescription takes priority",
			template: types.Template{
				ShortDescription: "A great template",
				Keywords:         []string{"docker", "kubernetes"},
			},
			expected: "A great template",
		},
		{
			name: "Keywords joined as fallback",
			template: types.Template{
				Keywords: []string{"docker", "kubernetes", "ubuntu", "test"},
			},
			expected: "docker, kubernetes, ubuntu",
		},
		{
			name: "Fewer than 3 keywords",
			template: types.Template{
				Keywords: []string{"docker", "test"},
			},
			expected: "docker, test",
		},
		{
			name: "No keywords fallback",
			template: types.Template{
				Keywords: []string{},
			},
			expected: "Lima VM template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := combiner.getDescription(tt.template)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestGetRawURL removed - getRawURL now uses Lima's TransformCustomURL which requires network access.
// The functionality is tested through TestCombineData integration tests.

func TestFormatDate(t *testing.T) {
	combiner := NewCombinerWithFS(nil, interfaces.NewDefaultFileSystem(), &mockURLTransformer{})

	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "Standard date",
			input:    time.Date(2024, 3, 15, 12, 30, 45, 0, time.UTC),
			expected: "2024-03-15",
		},
		{
			name:     "Zero time",
			input:    time.Time{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := combiner.formatDate(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestCalculateScoreStatistics(t *testing.T) {
	tests := []struct {
		name            string
		values          []float64
		expectedMin     float64
		expectedMax     float64
		expectedMed     float64
		expectedAvg     float64
		expectedZeroPct float64
	}{
		{
			name:            "Basic statistics",
			values:          []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expectedMin:     1.0,
			expectedMax:     5.0,
			expectedMed:     3.0,
			expectedAvg:     3.0,
			expectedZeroPct: 0.0,
		},
		{
			name:            "With zeros",
			values:          []float64{0, 0, 5.0, 10.0, 0},
			expectedMin:     0,
			expectedMax:     10.0,
			expectedMed:     0,
			expectedAvg:     3.0,
			expectedZeroPct: 60.0,
		},
		{
			name:            "Even number of values",
			values:          []float64{1.0, 2.0, 3.0, 4.0},
			expectedMin:     1.0,
			expectedMax:     4.0,
			expectedMed:     2.5, // Average of 2.0 and 3.0
			expectedAvg:     2.5,
			expectedZeroPct: 0.0,
		},
		{
			name:            "Single value",
			values:          []float64{42.0},
			expectedMin:     42.0,
			expectedMax:     42.0,
			expectedMed:     42.0,
			expectedAvg:     42.0,
			expectedZeroPct: 0.0,
		},
		{
			name:            "All zeros",
			values:          []float64{0, 0, 0},
			expectedMin:     0,
			expectedMax:     0,
			expectedMed:     0,
			expectedAvg:     0,
			expectedZeroPct: 100.0,
		},
		{
			name:            "Realistic notability scores",
			values:          []float64{25.5, 150.0, 0, 80.0, 200.5, 45.0, 0, 120.0},
			expectedMin:     0,
			expectedMax:     200.5,
			expectedMed:     62.5, // Average of 45.0 and 80.0
			expectedAvg:     77.625,
			expectedZeroPct: 25.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := calculateScoreStatistics("test", tt.values)

			if stats.Min != tt.expectedMin {
				t.Errorf("Min: expected %v, got %v", tt.expectedMin, stats.Min)
			}
			if stats.Max != tt.expectedMax {
				t.Errorf("Max: expected %v, got %v", tt.expectedMax, stats.Max)
			}
			if stats.Median != tt.expectedMed {
				t.Errorf("Median: expected %v, got %v", tt.expectedMed, stats.Median)
			}
			if stats.Average != tt.expectedAvg {
				t.Errorf("Average: expected %v, got %v", tt.expectedAvg, stats.Average)
			}
			if stats.ZeroPercent != tt.expectedZeroPct {
				t.Errorf("Zero percent: expected %v, got %v", tt.expectedZeroPct, stats.ZeroPercent)
			}
		})
	}
}
