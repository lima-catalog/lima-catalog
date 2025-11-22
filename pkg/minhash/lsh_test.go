package minhash

import (
	"testing"
)

func TestNewLSH(t *testing.T) {
	tests := []struct {
		name        string
		numBands    int
		rowsPerBand int
		wantErr     bool
	}{
		{
			name:        "valid configuration 16x8",
			numBands:    16,
			rowsPerBand: 8,
			wantErr:     false,
		},
		{
			name:        "valid configuration 8x16",
			numBands:    8,
			rowsPerBand: 16,
			wantErr:     false,
		},
		{
			name:        "invalid numBands zero",
			numBands:    0,
			rowsPerBand: 8,
			wantErr:     true,
		},
		{
			name:        "invalid numBands negative",
			numBands:    -1,
			rowsPerBand: 8,
			wantErr:     true,
		},
		{
			name:        "invalid rowsPerBand zero",
			numBands:    16,
			rowsPerBand: 0,
			wantErr:     true,
		},
		{
			name:        "invalid rowsPerBand negative",
			numBands:    16,
			rowsPerBand: -1,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lsh, err := NewLSH(tt.numBands, tt.rowsPerBand)

			if tt.wantErr {
				if err == nil {
					t.Error("NewLSH() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewLSH() unexpected error: %v", err)
				return
			}

			if lsh.numBands != tt.numBands {
				t.Errorf("numBands = %d, want %d", lsh.numBands, tt.numBands)
			}

			if lsh.rowsPerBand != tt.rowsPerBand {
				t.Errorf("rowsPerBand = %d, want %d", lsh.rowsPerBand, tt.rowsPerBand)
			}

			if lsh.buckets == nil {
				t.Error("buckets map is nil")
			}
		})
	}
}

func TestLSH_Add(t *testing.T) {
	lsh, err := NewLSH(32, 4)
	if err != nil {
		t.Fatalf("NewLSH() failed: %v", err)
	}

	// Create a 128-element signature
	sig := make([]uint32, 128)
	for i := range sig {
		sig[i] = uint32(i)
	}

	err = lsh.Add("template1", sig)
	if err != nil {
		t.Errorf("Add() unexpected error: %v", err)
	}

	// Verify the template was indexed
	if lsh.Size() != 1 {
		t.Errorf("Size() = %d, want 1", lsh.Size())
	}

	// Verify buckets were created
	if lsh.NumBuckets() == 0 {
		t.Error("NumBuckets() = 0, expected some buckets")
	}

	// Should have up to 32 buckets (one per band, but some might collide)
	if lsh.NumBuckets() > 32 {
		t.Errorf("NumBuckets() = %d, expected <= 32", lsh.NumBuckets())
	}
}

func TestLSH_Add_WrongLength(t *testing.T) {
	lsh, err := NewLSH(32, 4)
	if err != nil {
		t.Fatalf("NewLSH() failed: %v", err)
	}

	// Wrong length signature
	sig := make([]uint32, 64) // Should be 128

	err = lsh.Add("template1", sig)
	if err == nil {
		t.Error("Add() expected error for wrong length signature, got nil")
	}
}

func TestLSH_Query(t *testing.T) {
	lsh, err := NewLSH(32, 4)
	if err != nil {
		t.Fatalf("NewLSH() failed: %v", err)
	}

	// Add some signatures
	sig1 := make([]uint32, 128)
	sig2 := make([]uint32, 128)
	sig3 := make([]uint32, 128)

	// sig1 and sig2 are identical
	for i := range sig1 {
		sig1[i] = uint32(i * 2)
		sig2[i] = uint32(i * 2) // Same as sig1
		sig3[i] = uint32(i * 3) // Different
	}

	if err := lsh.Add("template1", sig1); err != nil {
		t.Fatalf("Add(template1) error: %v", err)
	}
	if err := lsh.Add("template2", sig2); err != nil {
		t.Fatalf("Add(template2) error: %v", err)
	}
	if err := lsh.Add("template3", sig3); err != nil {
		t.Fatalf("Add(template3) error: %v", err)
	}

	// Query with sig1 should find template1 and template2 (identical)
	candidates, err := lsh.Query(sig1)
	if err != nil {
		t.Errorf("Query() unexpected error: %v", err)
	}

	// Should find template1 and template2
	if _, ok := candidates["template1"]; !ok {
		t.Error("Query() didn't find template1 (identical signature)")
	}

	if _, ok := candidates["template2"]; !ok {
		t.Error("Query() didn't find template2 (identical signature)")
	}

	// Identical signatures should match on all bands (32 for both)
	if count, ok := candidates["template1"]; ok && count != 32 {
		t.Errorf("template1 band matches = %d, want 32 (all bands)", count)
	}

	if count, ok := candidates["template2"]; ok && count != 32 {
		t.Errorf("template2 band matches = %d, want 32 (all bands)", count)
	}

	t.Logf("Query results: %+v", candidates)
}

func TestLSH_Query_WrongLength(t *testing.T) {
	lsh, err := NewLSH(32, 4)
	if err != nil {
		t.Fatalf("NewLSH() failed: %v", err)
	}

	// Wrong length signature
	sig := make([]uint32, 64) // Should be 128

	_, err = lsh.Query(sig)
	if err == nil {
		t.Error("Query() expected error for wrong length signature, got nil")
	}
}

func TestLSH_SimilarSignatures(t *testing.T) {
	mh := New()
	lsh, err := NewLSH(32, 4)
	if err != nil {
		t.Fatalf("NewLSH() failed: %v", err)
	}

	// Create similar templates
	template1 := `
# Ubuntu template
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
provision:
  - script: apt-get update && apt-get install -y docker.io git vim
`

	template2 := `
# Ubuntu template
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
provision:
  - script: apt-get update && apt-get install -y docker.io git emacs
`

	template3 := `
# Alpine template
images:
  - location: https://dl-cdn.alpinelinux.org/alpine/v3.18/releases/x86_64/alpine-virt-3.18.0-x86_64.iso
provision:
  - script: apk add podman
`

	sig1 := mh.Signature(template1)
	sig2 := mh.Signature(template2)
	sig3 := mh.Signature(template3)

	if err := lsh.Add("template1", sig1); err != nil {
		t.Fatalf("Add(template1) failed: %v", err)
	}
	if err := lsh.Add("template2", sig2); err != nil {
		t.Fatalf("Add(template2) failed: %v", err)
	}
	if err := lsh.Add("template3", sig3); err != nil {
		t.Fatalf("Add(template3) failed: %v", err)
	}

	// Query with template1
	candidates, err := lsh.Query(sig1)
	if err != nil {
		t.Errorf("Query() unexpected error: %v", err)
	}

	t.Logf("Candidates for template1: %+v", candidates)

	// Should definitely find template1 (itself)
	if _, ok := candidates["template1"]; !ok {
		t.Error("Query() didn't find template1 (itself)")
	}

	// Should likely find template2 (very similar, ~84% similarity)
	if _, ok := candidates["template2"]; !ok {
		t.Error("Query() didn't find template2 (very similar template)")
	}

	// Check actual similarities
	sim12 := Similarity(sig1, sig2)
	sim13 := Similarity(sig1, sig3)

	t.Logf("Similarity(template1, template2) = %f", sim12)
	t.Logf("Similarity(template1, template3) = %f", sim13)

	// template1 and template2 are very similar (should be detected)
	if sim12 < 0.7 {
		t.Errorf("template1 and template2 similarity = %f, expected > 0.7", sim12)
	}

	// template3 might not be in candidates (different template)
	// This is expected behavior - LSH filters out dissimilar templates
}

func TestLSH_Size(t *testing.T) {
	lsh, err := NewLSH(32, 4)
	if err != nil {
		t.Fatalf("NewLSH() failed: %v", err)
	}

	if lsh.Size() != 0 {
		t.Errorf("Size() = %d, want 0 (empty)", lsh.Size())
	}

	sig := make([]uint32, 128)

	_ = lsh.Add("template1", sig)
	if lsh.Size() != 1 {
		t.Errorf("Size() = %d, want 1", lsh.Size())
	}

	_ = lsh.Add("template2", sig)
	if lsh.Size() != 2 {
		t.Errorf("Size() = %d, want 2", lsh.Size())
	}

	// Adding same template again shouldn't change size
	_ = lsh.Add("template1", sig)
	if lsh.Size() != 2 {
		t.Errorf("Size() = %d, want 2 (no duplicates counted)", lsh.Size())
	}
}

func TestLSH_Clear(t *testing.T) {
	lsh, err := NewLSH(32, 4)
	if err != nil {
		t.Fatalf("NewLSH() failed: %v", err)
	}

	sig := make([]uint32, 128)
	_ = lsh.Add("template1", sig)
	_ = lsh.Add("template2", sig)

	if lsh.Size() != 2 {
		t.Fatalf("Size() = %d, want 2 before Clear", lsh.Size())
	}

	lsh.Clear()

	if lsh.Size() != 0 {
		t.Errorf("Size() = %d, want 0 after Clear", lsh.Size())
	}

	if lsh.NumBuckets() != 0 {
		t.Errorf("NumBuckets() = %d, want 0 after Clear", lsh.NumBuckets())
	}
}

func TestLSH_EstimateThreshold(t *testing.T) {
	tests := []struct {
		name        string
		numBands    int
		rowsPerBand int
		wantMin     float64
		wantMax     float64
	}{
		{
			name:        "32 bands x 4 rows (default)",
			numBands:    32,
			rowsPerBand: 4,
			wantMin:     0.41,  // (1/32)^(1/4) ≈ 0.42
			wantMax:     0.46,
		},
		{
			name:        "16 bands x 8 rows (higher threshold)",
			numBands:    16,
			rowsPerBand: 8,
			wantMin:     0.68,  // (1/16)^(1/8) ≈ 0.71
			wantMax:     0.74,
		},
		{
			name:        "8 bands x 16 rows (very high threshold)",
			numBands:    8,
			rowsPerBand: 16,
			wantMin:     0.86,  // (1/8)^(1/16) ≈ 0.88
			wantMax:     0.90,
		},
		{
			name:        "64 bands x 2 rows (very low threshold)",
			numBands:    64,
			rowsPerBand: 2,
			wantMin:     0.12,  // (1/64)^(1/2) = 0.125 exactly
			wantMax:     0.13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lsh, err := NewLSH(tt.numBands, tt.rowsPerBand)
			if err != nil {
				t.Fatalf("NewLSH() failed: %v", err)
			}

			threshold := lsh.EstimateThreshold()
			t.Logf("Estimated threshold: %f", threshold)

			if threshold < tt.wantMin || threshold > tt.wantMax {
				t.Errorf("EstimateThreshold() = %f, want between %f and %f",
					threshold, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestHashBand(t *testing.T) {
	band1 := []uint32{1, 2, 3, 4, 5, 6, 7, 8}
	band2 := []uint32{1, 2, 3, 4, 5, 6, 7, 8}
	band3 := []uint32{9, 10, 11, 12, 13, 14, 15, 16}

	// Same band, same index -> same hash
	h1a := hashBand(band1, 0)
	h1b := hashBand(band2, 0)
	if h1a != h1b {
		t.Error("hashBand() produced different hashes for identical bands")
	}

	// Same band, different index -> different hash
	h1c := hashBand(band1, 1)
	if h1a == h1c {
		t.Error("hashBand() produced same hash for different band indices")
	}

	// Different bands -> different hash
	h2 := hashBand(band3, 0)
	if h1a == h2 {
		t.Error("hashBand() produced same hash for different bands")
	}
}

func BenchmarkLSH_Add(b *testing.B) {
	lsh, _ := NewLSH(32, 4)
	sig := make([]uint32, 128)
	for i := range sig {
		sig[i] = uint32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lsh.Add("template", sig)
	}
}

func BenchmarkLSH_Query(b *testing.B) {
	lsh, _ := NewLSH(32, 4)

	// Add 1000 templates
	for i := 0; i < 1000; i++ {
		sig := make([]uint32, 128)
		for j := range sig {
			sig[j] = uint32(i*j + j)
		}
		_ = lsh.Add(string(rune(i)), sig)
	}

	querySig := make([]uint32, 128)
	for i := range querySig {
		querySig[i] = uint32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lsh.Query(querySig) // Ignore error in benchmark
	}
}
