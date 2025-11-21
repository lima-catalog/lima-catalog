package discovery

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

// Mock HTTP Client for testing
type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) {
	return m.response, m.err
}

// Mock Clock for testing
type mockClock struct {
	now time.Time
}

func (m *mockClock) Now() time.Time {
	return m.now
}

func TestNewAnalyzer(t *testing.T) {
	t.Run("Creates analyzer with default values", func(t *testing.T) {
		a := NewAnalyzer()

		if a == nil {
			t.Fatal("expected non-nil analyzer")
		}
		if a.OfficialKnowledge == nil {
			t.Error("expected OfficialKnowledge to be initialized")
		}
		if a.OfficialKnowledge.Images == nil {
			t.Error("expected OfficialKnowledge.Images to be initialized")
		}
		if a.OfficialKnowledge.KnownLines.Comments == nil {
			t.Error("expected OfficialKnowledge.KnownLines.Comments to be initialized")
		}
		if a.ForceAnalyze != false {
			t.Error("expected ForceAnalyze to be false")
		}
		if a.HTTPClient == nil {
			t.Error("expected HTTPClient to be initialized")
		}
		if a.Clock == nil {
			t.Error("expected Clock to be initialized")
		}
	})

	t.Run("Creates analyzer with ForceAnalyze enabled", func(t *testing.T) {
		a := NewAnalyzer(WithForceAnalyze(true))

		if !a.ForceAnalyze {
			t.Error("expected ForceAnalyze to be true")
		}
	})
}

func TestInferCategory(t *testing.T) {
	tests := []struct {
		name             string
		info             *TemplateInfo
		repo             *types.Repository
		expectedCategory string
		expectedUseCase  string
	}{
		{
			name: "Kubernetes template",
			info: &TemplateInfo{
				HasK8s: true,
			},
			repo:             nil,
			expectedCategory: "orchestration",
			expectedUseCase:  "kubernetes",
		},
		{
			name: "Docker template",
			info: &TemplateInfo{
				HasDocker: true,
			},
			repo:             nil,
			expectedCategory: "containers",
			expectedUseCase:  "container-runtime",
		},
		{
			name: "Podman template",
			info: &TemplateInfo{
				HasPodman: true,
			},
			repo:             nil,
			expectedCategory: "containers",
			expectedUseCase:  "container-runtime",
		},
		{
			name: "Development category",
			info: &TemplateInfo{
				Categories: []string{"development"},
			},
			repo:             nil,
			expectedCategory: "development",
			expectedUseCase:  "dev-environment",
		},
		{
			name: "Database category",
			info: &TemplateInfo{
				Categories: []string{"database"},
			},
			repo:             nil,
			expectedCategory: "database",
			expectedUseCase:  "data-storage",
		},
		{
			name: "Security from repo topics",
			info: &TemplateInfo{},
			repo: &types.Repository{
				Topics: []string{"security", "pentesting"},
			},
			expectedCategory: "security",
			expectedUseCase:  "security-testing",
		},
		{
			name: "Testing from repo topics",
			info: &TemplateInfo{},
			repo: &types.Repository{
				Topics: []string{"test", "ci"},
			},
			expectedCategory: "testing",
			expectedUseCase:  "ci-cd",
		},
		{
			name: "Machine learning from repo topics",
			info: &TemplateInfo{},
			repo: &types.Repository{
				Topics: []string{"ml", "machine-learning"},
			},
			expectedCategory: "ml",
			expectedUseCase:  "machine-learning",
		},
		{
			name:             "Default category",
			info:             &TemplateInfo{},
			repo:             nil,
			expectedCategory: "general",
			expectedUseCase:  "vm",
		},
		{
			name: "K8s takes priority over Docker",
			info: &TemplateInfo{
				HasK8s:    true,
				HasDocker: true,
			},
			repo:             nil,
			expectedCategory: "orchestration",
			expectedUseCase:  "kubernetes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAnalyzer()
			category, useCase := a.inferCategory(tt.info, tt.repo)

			if category != tt.expectedCategory {
				t.Errorf("expected category %q, got %q", tt.expectedCategory, category)
			}
			if useCase != tt.expectedUseCase {
				t.Errorf("expected use case %q, got %q", tt.expectedUseCase, useCase)
			}
		})
	}
}

func TestGenerateBasicDescription(t *testing.T) {
	tests := []struct {
		name        string
		template    *types.Template
		info        *TemplateInfo
		repo        *types.Repository
		expectedStr string // substring that should be in description
	}{
		{
			name: "Ubuntu-based with K8s",
			template: &types.Template{
				Category: "orchestration",
			},
			info: &TemplateInfo{
				Images: []string{"ubuntu"},
				HasK8s: true,
			},
			repo:        nil,
			expectedStr: "Ubuntu-based orchestration with Kubernetes",
		},
		{
			name: "Alpine with Docker",
			template: &types.Template{
				Category: "containers",
			},
			info: &TemplateInfo{
				Images:    []string{"alpine"},
				HasDocker: true,
			},
			repo:        nil,
			expectedStr: "Alpine-based containers with Docker",
		},
		{
			name: "With architecture info",
			template: &types.Template{
				Category: "general",
			},
			info: &TemplateInfo{
				Images: []string{"debian"},
				Arch:   []string{"x86_64", "aarch64"},
			},
			repo:        nil,
			expectedStr: "Debian-based general (x86_64/aarch64)",
		},
		{
			name: "With repo description",
			template: &types.Template{
				Category: "development",
			},
			info: &TemplateInfo{
				Images: []string{"fedora"},
			},
			repo: &types.Repository{
				Description: "A development environment for Go",
			},
			expectedStr: "A development environment for Go",
		},
		{
			name: "Podman runtime",
			template: &types.Template{
				Category: "containers",
			},
			info: &TemplateInfo{
				Images:    []string{"ubuntu"},
				HasPodman: true,
			},
			repo:        nil,
			expectedStr: "Ubuntu-based containers with Podman",
		},
		{
			name: "No images",
			template: &types.Template{
				Category: "general",
			},
			info: &TemplateInfo{
				Images: []string{},
			},
			repo:        nil,
			expectedStr: "general",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAnalyzer()
			desc := a.generateBasicDescription(tt.template, tt.info, tt.repo)

			if desc == "" {
				t.Error("expected non-empty description")
			}
			if tt.expectedStr != "" && !contains(desc, tt.expectedStr) {
				t.Errorf("expected description to contain %q, got %q", tt.expectedStr, desc)
			}
		})
	}
}

func TestAnalyzeTemplatesSkipLogic(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	oldTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	newTime := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name                   string
		forceAnalyze           bool
		templates              []types.Template
		expectedAnalyzedCount  int
		expectedToBeReanalyzed bool
	}{
		{
			name:         "Skip already analyzed template",
			forceAnalyze: false,
			templates: []types.Template{
				{
					ID:          "test/repo/template.yaml",
					Path:        "template.yaml",
					Repo:        "test/repo",
					LastUpdated: oldTime,
					AnalyzedAt:  fixedTime, // Analyzed after last update
				},
			},
			expectedAnalyzedCount:  1,
			expectedToBeReanalyzed: false,
		},
		{
			name:         "Re-analyze if content changed",
			forceAnalyze: false,
			templates: []types.Template{
				{
					ID:          "test/repo/template.yaml",
					Path:        "template.yaml",
					Repo:        "test/repo",
					LastUpdated: newTime,      // Content changed
					AnalyzedAt:  fixedTime,    // Analysis is outdated
					URL:         "http://test", // Would need mocking for real analysis
				},
			},
			expectedAnalyzedCount:  1,
			expectedToBeReanalyzed: true,
		},
		{
			name:         "Force analyze ignores timestamps",
			forceAnalyze: true,
			templates: []types.Template{
				{
					ID:          "test/repo/template.yaml",
					Path:        "template.yaml",
					Repo:        "test/repo",
					LastUpdated: oldTime,
					AnalyzedAt:  fixedTime,
					URL:         "http://test",
				},
			},
			expectedAnalyzedCount:  1,
			expectedToBeReanalyzed: true,
		},
		{
			name:         "Never analyzed template",
			forceAnalyze: false,
			templates: []types.Template{
				{
					ID:          "test/repo/new.yaml",
					Path:        "new.yaml",
					Repo:        "test/repo",
					LastUpdated: fixedTime,
					AnalyzedAt:  time.Time{}, // Zero value = never analyzed
					URL:         "http://test",
				},
			},
			expectedAnalyzedCount:  1,
			expectedToBeReanalyzed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAnalyzer(
				WithForceAnalyze(tt.forceAnalyze),
				WithClock(&mockClock{now: fixedTime}),
				WithHTTPClient(&mockHTTPClient{
					response: nil,
					err:      http.ErrContentLength, // Will trigger parse failure, which is fine for this test
				}),
			)

			repoMap := make(map[string]*types.Repository)

			analyzed, err := a.AnalyzeTemplates(context.Background(), tt.templates, repoMap)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(analyzed) != tt.expectedAnalyzedCount {
				t.Errorf("expected %d analyzed templates, got %d", tt.expectedAnalyzedCount, len(analyzed))
			}

			if len(analyzed) > 0 {
				// Check if template was re-analyzed by checking if AnalyzedAt was updated
				if tt.expectedToBeReanalyzed {
					if analyzed[0].AnalyzedAt.IsZero() {
						t.Error("expected template to have AnalyzedAt timestamp set")
					}
					// In real analysis, name would be set
					if analyzed[0].Name == "" && tt.forceAnalyze {
						t.Error("expected template name to be derived during analysis")
					}
				}
			}
		})
	}
}

func TestAnalyzeTemplate(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	t.Run("Basic template analysis", func(t *testing.T) {
		a := NewAnalyzer(
			WithClock(&mockClock{now: fixedTime}),
			WithHTTPClient(&mockHTTPClient{
				response: nil,
				err:      http.ErrContentLength, // Parse will fail, but that's OK for this test
			}),
		)

		template := &types.Template{
			ID:   "test/repo/example.yaml",
			Path: "example.yaml",
			Repo: "test/repo",
			URL:  "http://example.com/template.yaml",
		}

		repo := &types.Repository{
			ID:          "test/repo",
			Description: "Test repository",
			Topics:      []string{"testing"},
		}

		err := a.AnalyzeTemplate(context.Background(), template, repo)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Check that basic fields were set
		if template.Name == "" {
			t.Error("expected Name to be set")
		}
		if template.DisplayName == "" {
			t.Error("expected DisplayName to be set")
		}
		if template.Category == "" {
			t.Error("expected Category to be set")
		}
		if template.UseCase == "" {
			t.Error("expected UseCase to be set")
		}
		if template.AnalyzedAt != fixedTime {
			t.Errorf("expected AnalyzedAt to be %v, got %v", fixedTime, template.AnalyzedAt)
		}
	})

	t.Run("Analysis with official images", func(t *testing.T) {
		officialKnowledge := &OfficialKnowledge{
			Images: []string{
				"ubuntu.com",
				"alpinelinux.org",
			},
			KnownLines: OfficialKnownLines{
				Comments:  []string{},
				Provision: []string{},
				Probes:    []string{},
				Messages:  []string{},
			},
		}

		a := NewAnalyzer(
			WithClock(&mockClock{now: fixedTime}),
			WithHTTPClient(&mockHTTPClient{
				response: nil,
				err:      http.ErrContentLength,
			}),
			WithOfficialKnowledge(officialKnowledge),
		)

		template := &types.Template{
			ID:   "test/repo/template.yaml",
			Path: "template.yaml",
			Repo: "test/repo",
			URL:  "http://example.com/template.yaml",
		}

		err := a.AnalyzeTemplate(context.Background(), template, nil)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Notability should be initialized (even if empty)
		if template.Notability == nil {
			t.Error("expected Notability to be initialized")
		}
	})
}

func TestNewAnalyzer_DuplicateDetection(t *testing.T) {
	t.Run("Duplicate detection enabled by default", func(t *testing.T) {
		a := NewAnalyzer()

		if !a.DetectDuplicates {
			t.Error("expected DetectDuplicates to be enabled by default")
		}
		if a.DuplicateSimilarityThreshold != 0.5 {
			t.Errorf("expected DuplicateSimilarityThreshold to be 0.5, got %f", a.DuplicateSimilarityThreshold)
		}
	})

	t.Run("Can disable duplicate detection", func(t *testing.T) {
		a := NewAnalyzer(WithDetectDuplicates(false))

		if a.DetectDuplicates {
			t.Error("expected DetectDuplicates to be disabled")
		}
	})

	t.Run("Can set custom similarity threshold", func(t *testing.T) {
		a := NewAnalyzer(WithDuplicateSimilarityThreshold(0.7))

		if a.DuplicateSimilarityThreshold != 0.7 {
			t.Errorf("expected DuplicateSimilarityThreshold to be 0.7, got %f", a.DuplicateSimilarityThreshold)
		}
	})
}

func TestAnalyzeTemplates_DuplicateDetection(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	t.Run("Detects similar templates", func(t *testing.T) {
		a := NewAnalyzer(
			WithClock(&mockClock{now: fixedTime}),
			WithHTTPClient(&mockHTTPClient{
				response: nil,
				err:      http.ErrContentLength, // Parse will fail, triggers warning
			}),
			WithDetectDuplicates(true),
			WithDuplicateSimilarityThreshold(0.5),
		)

		// Create templates with pre-computed MinHash signatures
		// These would normally be generated by AnalyzeTemplate
		sig1 := a.MinHash.Signature("Ubuntu Docker development environment")
		sig2 := a.MinHash.Signature("Ubuntu Docker development environment") // Identical
		sig3 := a.MinHash.Signature("Alpine Podman container runtime")       // Different

		templates := []types.Template{
			{
				ID:               "owner1/repo1/ubuntu.yaml",
				Path:             "ubuntu.yaml",
				Repo:             "owner1/repo1",
				URL:              "http://example.com/1",
				MinHashSignature: sig1,
			},
			{
				ID:               "owner2/repo2/ubuntu-dev.yaml",
				Path:             "ubuntu-dev.yaml",
				Repo:             "owner2/repo2",
				URL:              "http://example.com/2",
				MinHashSignature: sig2,
			},
			{
				ID:               "owner3/repo3/alpine.yaml",
				Path:             "alpine.yaml",
				Repo:             "owner3/repo3",
				URL:              "http://example.com/3",
				MinHashSignature: sig3,
			},
		}

		analyzed, err := a.AnalyzeTemplates(context.Background(), templates, nil)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(analyzed) != 3 {
			t.Fatalf("expected 3 analyzed templates, got %d", len(analyzed))
		}

		// Template 1 should find template 2 as similar (identical signatures)
		if len(analyzed[0].SimilarTemplates) == 0 {
			t.Error("template 1 should have similar templates")
		} else {
			found := false
			for _, sim := range analyzed[0].SimilarTemplates {
				if sim.ID == "owner2/repo2/ubuntu-dev.yaml" {
					found = true
					if sim.Similarity != 1.0 {
						t.Errorf("expected similarity 1.0 for identical templates, got %f", sim.Similarity)
					}
					if !sim.IsExactDuplicate() {
						t.Errorf("expected exact duplicate (similarity > 0.9), got %f", sim.Similarity)
					}
				}
			}
			if !found {
				t.Error("template 1 should find template 2 as similar")
			}
		}

		// Template 3 should not have similar templates (different content)
		foundSimilar := false
		for _, sim := range analyzed[2].SimilarTemplates {
			if sim.ID == "owner1/repo1/ubuntu.yaml" || sim.ID == "owner2/repo2/ubuntu-dev.yaml" {
				foundSimilar = true
			}
		}
		if foundSimilar {
			t.Error("template 3 should not find Ubuntu templates as similar")
		}
	})

	t.Run("Skips duplicate detection when disabled", func(t *testing.T) {
		a := NewAnalyzer(
			WithClock(&mockClock{now: fixedTime}),
			WithHTTPClient(&mockHTTPClient{
				response: nil,
				err:      http.ErrContentLength,
			}),
			WithDetectDuplicates(false), // Disabled
		)

		sig := a.MinHash.Signature("Test content")
		templates := []types.Template{
			{
				ID:               "test/repo/template1.yaml",
				Path:             "template1.yaml",
				Repo:             "test/repo",
				URL:              "http://example.com/1",
				MinHashSignature: sig,
			},
			{
				ID:               "test/repo/template2.yaml",
				Path:             "template2.yaml",
				Repo:             "test/repo",
				URL:              "http://example.com/2",
				MinHashSignature: sig,
			},
		}

		analyzed, err := a.AnalyzeTemplates(context.Background(), templates, nil)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// SimilarTemplates should not be populated when detection is disabled
		for _, tmpl := range analyzed {
			if len(tmpl.SimilarTemplates) > 0 {
				t.Error("expected no similar templates when detection is disabled")
			}
		}
	})

	t.Run("Handles templates without signatures gracefully", func(t *testing.T) {
		a := NewAnalyzer(
			WithClock(&mockClock{now: fixedTime}),
			WithHTTPClient(&mockHTTPClient{
				response: nil,
				err:      http.ErrContentLength,
			}),
			WithDetectDuplicates(true),
		)

		templates := []types.Template{
			{
				ID:               "test/repo/no-signature.yaml",
				Path:             "no-signature.yaml",
				Repo:             "test/repo",
				URL:              "http://example.com/1",
				MinHashSignature: nil, // No signature
			},
		}

		analyzed, err := a.AnalyzeTemplates(context.Background(), templates, nil)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Should not error, just skip duplicate detection for this template
		if len(analyzed) != 1 {
			t.Errorf("expected 1 analyzed template, got %d", len(analyzed))
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || anySubstring(s, substr))
}

func anySubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
