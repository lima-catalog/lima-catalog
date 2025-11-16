package combiner

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

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
					URL:              "https://github.com/owner1/repo1/blob/main/template.yaml",
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
					URL:              "https://github.com/owner1/repo1/blob/main/template.yaml",
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
					URL:              "https://github.com/owner1/repo1/blob/main/.github/workflows/ci.yaml",
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
					URL:              "https://github.com/owner1/repo1/blob/main/template.yaml",
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
					URL:              "https://github.com/lima-vm/lima/blob/master/templates/ubuntu.yaml",
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
					URL:         "https://github.com/owner1/repo1/blob/main/path/to/lima.yaml",
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
					URL:              "https://github.com/org2/repo1/blob/main/b.yaml",
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
					URL:              "https://github.com/org1/repo2/blob/main/a.yaml",
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
					URL:              "https://github.com/org1/repo1/blob/main/z.yaml",
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
			// Create combiner
			combiner := NewCombiner(tt.blocklist)

			// Create temp output file
			tmpFile, err := os.CreateTemp("", "test-combined-*.jsonl")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())
			tmpFile.Close()

			// Run combine
			err = combiner.CombineData(tt.templates, tt.repos, tt.orgs, tmpFile.Name())
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
	combiner := NewCombiner(nil)

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
	combiner := NewCombiner(nil)

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

func TestGetRawURL(t *testing.T) {
	combiner := NewCombiner(nil)

	tests := []struct {
		name     string
		template types.Template
		repo     types.Repository
		expected string
	}{
		{
			name: "Standard template",
			template: types.Template{
				Repo: "owner/repo",
				Path: "template.yaml",
			},
			repo: types.Repository{
				DefaultBranch: "main",
			},
			expected: "https://raw.githubusercontent.com/owner/repo/main/template.yaml",
		},
		{
			name: "Template with master branch",
			template: types.Template{
				Repo: "lima-vm/lima",
				Path: "templates/ubuntu.yaml",
			},
			repo: types.Repository{
				DefaultBranch: "master",
			},
			expected: "https://raw.githubusercontent.com/lima-vm/lima/master/templates/ubuntu.yaml",
		},
		{
			name: "Template with nested path",
			template: types.Template{
				Repo: "owner/repo",
				Path: "path/to/nested/template.yaml",
			},
			repo: types.Repository{
				DefaultBranch: "main",
			},
			expected: "https://raw.githubusercontent.com/owner/repo/main/path/to/nested/template.yaml",
		},
		{
			name: "Template with empty branch fallback",
			template: types.Template{
				Repo: "owner/repo",
				Path: "template.yaml",
			},
			repo: types.Repository{
				DefaultBranch: "", // Empty
			},
			expected: "https://raw.githubusercontent.com/owner/repo/main/template.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := combiner.getRawURL(tt.template, tt.repo)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	combiner := NewCombiner(nil)

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
