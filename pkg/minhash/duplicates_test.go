package minhash

import (
	"testing"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

func TestNewDuplicateDetector(t *testing.T) {
	mh := New()

	tests := []struct {
		name      string
		threshold float64
		wantErr   bool
	}{
		{
			name:      "valid threshold 0.5",
			threshold: 0.5,
			wantErr:   false,
		},
		{
			name:      "valid threshold 0.0",
			threshold: 0.0,
			wantErr:   false,
		},
		{
			name:      "valid threshold 1.0",
			threshold: 1.0,
			wantErr:   false,
		},
		{
			name:      "invalid threshold negative",
			threshold: -0.1,
			wantErr:   true,
		},
		{
			name:      "invalid threshold > 1",
			threshold: 1.1,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dd, err := NewDuplicateDetector(mh, tt.threshold)

			if tt.wantErr {
				if err == nil {
					t.Error("NewDuplicateDetector() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewDuplicateDetector() unexpected error: %v", err)
				return
			}

			if dd.similarityThreshold != tt.threshold {
				t.Errorf("similarityThreshold = %f, want %f", dd.similarityThreshold, tt.threshold)
			}
		})
	}
}

func TestChooseLSHConfig(t *testing.T) {
	tests := []struct {
		name          string
		signatureSize int
		threshold     float64
		wantBands     int
		wantRows      int
	}{
		{
			name:          "threshold 0.8 -> strict (16x8)",
			signatureSize: 128,
			threshold:     0.8,
			wantBands:     16,
			wantRows:      8,
		},
		{
			name:          "threshold 0.7 -> strict (16x8)",
			signatureSize: 128,
			threshold:     0.7,
			wantBands:     16,
			wantRows:      8,
		},
		{
			name:          "threshold 0.6 -> balanced (32x4)",
			signatureSize: 128,
			threshold:     0.6,
			wantBands:     32,
			wantRows:      4,
		},
		{
			name:          "threshold 0.5 -> balanced (32x4)",
			signatureSize: 128,
			threshold:     0.5,
			wantBands:     32,
			wantRows:      4,
		},
		{
			name:          "threshold 0.4 -> loose (64x2)",
			signatureSize: 128,
			threshold:     0.4,
			wantBands:     64,
			wantRows:      2,
		},
		{
			name:          "threshold 0.2 -> loose (64x2)",
			signatureSize: 128,
			threshold:     0.2,
			wantBands:     64,
			wantRows:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bands, rows := chooseLSHConfig(tt.signatureSize, tt.threshold)

			if bands != tt.wantBands {
				t.Errorf("numBands = %d, want %d", bands, tt.wantBands)
			}

			if rows != tt.wantRows {
				t.Errorf("rowsPerBand = %d, want %d", rows, tt.wantRows)
			}
		})
	}
}

func TestDuplicateDetector_AddAndFindSimilar(t *testing.T) {
	mh := New()
	dd, err := NewDuplicateDetector(mh, 0.5)
	if err != nil {
		t.Fatalf("NewDuplicateDetector() failed: %v", err)
	}

	// Create templates with signatures
	template1 := &types.Template{
		ID:               "owner1/repo1/template.yaml",
		MinHashSignature: mh.Signature("Ubuntu Docker development environment"),
	}

	template2 := &types.Template{
		ID:               "owner2/repo2/template.yaml",
		MinHashSignature: mh.Signature("Ubuntu Docker development environment"),
	}

	template3 := &types.Template{
		ID:               "owner3/repo3/template.yaml",
		MinHashSignature: mh.Signature("Alpine Podman container runtime"),
	}

	// Add templates
	if err := dd.Add(template1); err != nil {
		t.Fatalf("Add(template1) error: %v", err)
	}
	if err := dd.Add(template2); err != nil {
		t.Fatalf("Add(template2) error: %v", err)
	}
	if err := dd.Add(template3); err != nil {
		t.Fatalf("Add(template3) error: %v", err)
	}

	if dd.Size() != 3 {
		t.Errorf("Size() = %d, want 3", dd.Size())
	}

	// Find similar to template1
	similar, err := dd.FindSimilar(template1)
	if err != nil {
		t.Errorf("FindSimilar() error: %v", err)
	}

	// Should find template2 (identical text) but not template3 (different)
	if len(similar) < 1 {
		t.Errorf("FindSimilar() found %d similar templates, want at least 1", len(similar))
	}

	// Check that template2 is in results
	foundTemplate2 := false
	for _, s := range similar {
		if s.ID == "owner2/repo2/template.yaml" {
			foundTemplate2 = true
			if s.Similarity != 1.0 {
				t.Errorf("template2 similarity = %f, want 1.0 (identical)", s.Similarity)
			}
			if !s.IsExactDuplicate() {
				t.Errorf("template2 should be exact duplicate (similarity > 0.9)")
			}
		}
	}

	if !foundTemplate2 {
		t.Error("FindSimilar() didn't find template2 (identical content)")
	}

	t.Logf("Found %d similar templates: %+v", len(similar), similar)
}

func TestDuplicateDetector_DetectDuplicates(t *testing.T) {
	mh := New()
	dd, err := NewDuplicateDetector(mh, 0.5)
	if err != nil {
		t.Fatalf("NewDuplicateDetector() failed: %v", err)
	}

	// Create templates with real YAML content
	template1Content := `
# Ubuntu development template
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
provision:
  - script: apt-get update && apt-get install -y docker.io git vim
`

	template2Content := `
# Ubuntu development template
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
provision:
  - script: apt-get update && apt-get install -y docker.io git emacs
`

	template3Content := `
# Alpine container runtime
images:
  - location: https://dl-cdn.alpinelinux.org/alpine/v3.18/releases/x86_64/alpine-virt-3.18.0-x86_64.iso
provision:
  - script: apk add podman
`

	templates := []types.Template{
		{
			ID:               "owner1/repo1/ubuntu.yaml",
			MinHashSignature: mh.Signature(template1Content),
		},
		{
			ID:               "owner2/repo2/ubuntu-dev.yaml",
			MinHashSignature: mh.Signature(template2Content),
		},
		{
			ID:               "owner3/repo3/alpine.yaml",
			MinHashSignature: mh.Signature(template3Content),
		},
	}

	// Detect duplicates
	result, err := dd.DetectDuplicates(templates)
	if err != nil {
		t.Fatalf("DetectDuplicates() error: %v", err)
	}

	// Check results
	if len(result) != 3 {
		t.Errorf("DetectDuplicates() returned %d templates, want 3", len(result))
	}

	// Template 1 should find template 2 as similar
	template1Result := result[0]
	if len(template1Result.SimilarTemplates) == 0 {
		t.Error("Template 1 should have similar templates")
	}

	foundTemplate2 := false
	for _, similar := range template1Result.SimilarTemplates {
		if similar.ID == "owner2/repo2/ubuntu-dev.yaml" {
			foundTemplate2 = true
			t.Logf("Template1 -> Template2 similarity: %f (exact=%v)",
				similar.Similarity, similar.IsExactDuplicate())

			if similar.Similarity < 0.7 {
				t.Errorf("Template1-Template2 similarity = %f, want >= 0.7",
					similar.Similarity)
			}
		}
	}

	if !foundTemplate2 {
		t.Error("Template 1 didn't find template 2 as similar (they should be very similar)")
	}

	// Template 3 should not have template 1 or 2 as similar (threshold 0.5)
	template3Result := result[2]
	for _, similar := range template3Result.SimilarTemplates {
		if similar.ID == "owner1/repo1/ubuntu.yaml" || similar.ID == "owner2/repo2/ubuntu-dev.yaml" {
			t.Errorf("Template 3 found %s as similar (similarity %f), but they are very different",
				similar.ID, similar.Similarity)
		}
	}
}

func TestDuplicateDetector_NoSignature(t *testing.T) {
	mh := New()
	dd, err := NewDuplicateDetector(mh, 0.5)
	if err != nil {
		t.Fatalf("NewDuplicateDetector() failed: %v", err)
	}

	// Template without signature
	template := &types.Template{
		ID:               "owner/repo/template.yaml",
		MinHashSignature: nil,
	}

	// Add should not error
	err = dd.Add(template)
	if err != nil {
		t.Errorf("Add() with no signature error: %v", err)
	}

	// FindSimilar should return empty, not error
	similar, err := dd.FindSimilar(template)
	if err != nil {
		t.Errorf("FindSimilar() with no signature error: %v", err)
	}

	if len(similar) != 0 {
		t.Errorf("FindSimilar() with no signature returned %d results, want 0", len(similar))
	}
}

func TestDuplicateDetector_Clear(t *testing.T) {
	mh := New()
	dd, err := NewDuplicateDetector(mh, 0.5)
	if err != nil {
		t.Fatalf("NewDuplicateDetector() failed: %v", err)
	}

	// Add some templates
	template := &types.Template{
		ID:               "owner/repo/template.yaml",
		MinHashSignature: mh.Signature("test content"),
	}
	if err := dd.Add(template); err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if dd.Size() != 1 {
		t.Fatalf("Size() = %d, want 1 before Clear", dd.Size())
	}

	// Clear
	dd.Clear()

	if dd.Size() != 0 {
		t.Errorf("Size() = %d, want 0 after Clear", dd.Size())
	}
}

func BenchmarkDuplicateDetector_DetectDuplicates(b *testing.B) {
	mh := New()
	dd, _ := NewDuplicateDetector(mh, 0.5)

	// Create 100 templates
	templates := make([]types.Template, 100)
	for i := 0; i < 100; i++ {
		content := "Template " + string(rune(i)) + " with some unique content"
		templates[i] = types.Template{
			ID:               "owner/repo/template" + string(rune(i)) + ".yaml",
			MinHashSignature: mh.Signature(content),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = dd.DetectDuplicates(templates) // Ignore results in benchmark
	}
}
