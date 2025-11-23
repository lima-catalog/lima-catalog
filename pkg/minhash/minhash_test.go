// Unit Tests: MinHash Implementation (minhash package)
//
// High-level overview of what's being tested:
// - MinHash instance creation with default and custom configurations
// - Text tokenization (word extraction with punctuation handling)
// - Shingle extraction (n-gram generation from tokens)
// - Signature generation from text content
// - Signature determinism (same input produces same signature)
// - Similarity calculation between signatures
// - Similarity detection with real template YAML content
// - Hash function behavior with different seeds
// - Edge cases: empty text, short text, duplicate shingles
// - Case normalization in text processing
// - Performance benchmarks for signature generation and similarity
// - YAML template similarity scenarios (Ubuntu vs Alpine examples)

package minhash

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		opts        []Option
		wantHashes  int
		wantShingle int
	}{
		{
			name:        "default configuration",
			opts:        nil,
			wantHashes:  DefaultNumHashes,
			wantShingle: DefaultShingleSize,
		},
		{
			name:        "custom num hashes",
			opts:        []Option{WithNumHashes(256)},
			wantHashes:  256,
			wantShingle: DefaultShingleSize,
		},
		{
			name:        "custom shingle size",
			opts:        []Option{WithShingleSize(3)},
			wantHashes:  DefaultNumHashes,
			wantShingle: 3,
		},
		{
			name: "both custom",
			opts: []Option{
				WithNumHashes(64),
				WithShingleSize(7),
			},
			wantHashes:  64,
			wantShingle: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mh := New(tt.opts...)

			if mh.SignatureSize() != tt.wantHashes {
				t.Errorf("SignatureSize() = %d, want %d", mh.SignatureSize(), tt.wantHashes)
			}

			if mh.ShingleSize() != tt.wantShingle {
				t.Errorf("ShingleSize() = %d, want %d", mh.ShingleSize(), tt.wantShingle)
			}

			if len(mh.seeds) != tt.wantHashes {
				t.Errorf("seeds length = %d, want %d", len(mh.seeds), tt.wantHashes)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple words",
			input: "hello world test",
			want:  []string{"hello", "world", "test"},
		},
		{
			name:  "with punctuation",
			input: "hello, world! test?",
			want:  []string{"hello", "world", "test"},
		},
		{
			name:  "multiple spaces",
			input: "hello    world     test",
			want:  []string{"hello", "world", "test"},
		},
		{
			name:  "with numbers",
			input: "test123 abc456 xyz",
			want:  []string{"test123", "abc456", "xyz"},
		},
		{
			name:  "empty string",
			input: "",
			want:  []string{},
		},
		{
			name:  "only punctuation",
			input: ".,!?;:",
			want:  []string{},
		},
		{
			name:  "mixed case preserved (before normalization)",
			input: "Hello World",
			want:  []string{"Hello", "World"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenize(tt.input)

			if len(got) != len(tt.want) {
				t.Errorf("tokenize() returned %d words, want %d", len(got), len(tt.want))
				t.Errorf("  got: %v", got)
				t.Errorf("  want: %v", tt.want)
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("tokenize()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestExtractShingles(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		shingleSize int
		wantCount   int
		wantContain string
	}{
		{
			name:        "simple text with k=3",
			text:        "this is a test document",
			shingleSize: 3,
			wantCount:   3, // "this is a", "is a test", "a test document"
			wantContain: "this is a",
		},
		{
			name:        "simple text with k=2",
			text:        "hello world test",
			shingleSize: 2,
			wantCount:   2, // "hello world", "world test"
			wantContain: "hello world",
		},
		{
			name:        "text shorter than shingle size",
			text:        "hi there",
			shingleSize: 5,
			wantCount:   1, // Falls back to full text
			wantContain: "hi there",
		},
		{
			name:        "empty text",
			text:        "",
			shingleSize: 3,
			wantCount:   0,
			wantContain: "",
		},
		{
			name:        "single word",
			text:        "hello",
			shingleSize: 3,
			wantCount:   1,
			wantContain: "hello",
		},
		{
			name:        "duplicate shingles removed",
			text:        "the the the the",
			shingleSize: 2,
			wantCount:   1, // Only "the the" appears once (deduplicated)
			wantContain: "the the",
		},
		{
			name:        "case normalized",
			text:        "Hello World TEST",
			shingleSize: 2,
			wantCount:   2,
			wantContain: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mh := New(WithShingleSize(tt.shingleSize))
			got := mh.extractShingles(tt.text)

			if len(got) != tt.wantCount {
				t.Errorf("extractShingles() returned %d shingles, want %d", len(got), tt.wantCount)
				t.Errorf("  got: %v", got)
				return
			}

			if tt.wantContain != "" {
				found := false
				for _, shingle := range got {
					if shingle == tt.wantContain {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("extractShingles() missing expected shingle %q", tt.wantContain)
					t.Errorf("  got: %v", got)
				}
			}
		})
	}
}

func TestSignature(t *testing.T) {
	mh := New()

	tests := []struct {
		name       string
		text       string
		wantLength int
		checkEmpty bool
	}{
		{
			name:       "normal text",
			text:       "This is a test document for MinHash signature generation",
			wantLength: DefaultNumHashes,
			checkEmpty: false,
		},
		{
			name:       "empty text",
			text:       "",
			wantLength: DefaultNumHashes,
			checkEmpty: true, // Should be all MaxHashValue
		},
		{
			name:       "short text",
			text:       "hi",
			wantLength: DefaultNumHashes,
			checkEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := mh.Signature(tt.text)

			if len(sig) != tt.wantLength {
				t.Errorf("Signature() length = %d, want %d", len(sig), tt.wantLength)
			}

			if tt.checkEmpty {
				// For empty text, all hashes should be MaxHashValue
				for i, h := range sig {
					if h != MaxHashValue {
						t.Errorf("Signature()[%d] = %d, want %d (MaxHashValue)", i, h, MaxHashValue)
						break
					}
				}
			} else {
				// For non-empty text, not all hashes should be MaxHashValue
				allMax := true
				for _, h := range sig {
					if h != MaxHashValue {
						allMax = false
						break
					}
				}
				if allMax {
					t.Error("Signature() all values are MaxHashValue, expected some computed hashes")
				}
			}
		})
	}
}

func TestSignatureDeterministic(t *testing.T) {
	mh := New()
	text := "This is a test document"

	sig1 := mh.Signature(text)
	sig2 := mh.Signature(text)

	if len(sig1) != len(sig2) {
		t.Fatal("Signatures have different lengths")
	}

	for i := range sig1 {
		if sig1[i] != sig2[i] {
			t.Errorf("Signature not deterministic: sig1[%d] = %d, sig2[%d] = %d",
				i, sig1[i], i, sig2[i])
		}
	}
}

func TestSignatureDifferentTexts(t *testing.T) {
	mh := New()

	text1 := "This is a test document"
	text2 := "This is a completely different document"

	sig1 := mh.Signature(text1)
	sig2 := mh.Signature(text2)

	// Signatures should be different
	identical := true
	for i := range sig1 {
		if sig1[i] != sig2[i] {
			identical = false
			break
		}
	}

	if identical {
		t.Error("Different texts produced identical signatures")
	}
}

func TestSimilarity(t *testing.T) {
	tests := []struct {
		name  string
		sig1  []uint32
		sig2  []uint32
		want  float64
	}{
		{
			name:  "identical signatures",
			sig1:  []uint32{1, 2, 3, 4, 5},
			sig2:  []uint32{1, 2, 3, 4, 5},
			want:  1.0,
		},
		{
			name:  "no matches",
			sig1:  []uint32{1, 2, 3, 4, 5},
			sig2:  []uint32{6, 7, 8, 9, 10},
			want:  0.0,
		},
		{
			name:  "partial match (60%)",
			sig1:  []uint32{1, 2, 3, 4, 5},
			sig2:  []uint32{1, 2, 3, 99, 100},
			want:  0.6,
		},
		{
			name:  "different lengths",
			sig1:  []uint32{1, 2, 3},
			sig2:  []uint32{1, 2},
			want:  0.0,
		},
		{
			name:  "empty signatures",
			sig1:  []uint32{},
			sig2:  []uint32{},
			want:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Similarity(tt.sig1, tt.sig2)
			if got != tt.want {
				t.Errorf("Similarity() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestSimilarityWithRealTexts(t *testing.T) {
	mh := New()

	tests := []struct {
		name       string
		text1      string
		text2      string
		expectHigh bool // expect similarity > 0.5
	}{
		{
			name:       "identical texts",
			text1:      "This is a test document for similarity",
			text2:      "This is a test document for similarity",
			expectHigh: true,
		},
		{
			name:       "very similar texts",
			text1:      "This is a test document for similarity detection",
			text2:      "This is a test document for similarity checking",
			expectHigh: true,
		},
		{
			name:       "completely different texts",
			text1:      "The quick brown fox jumps",
			text2:      "Lima virtual machine templates",
			expectHigh: false,
		},
		{
			name: "minor word change in longer text",
			text1: `Ubuntu-based development environment with Docker container runtime.
					Includes git version control system and build tools for compilation.
					Configured with 4 CPUs and 4 GiB memory for optimal performance.`,
			text2: `Ubuntu-based development environment with Docker container runtime.
					Includes git version control system and build tools for compilation.
					Configured with 4 CPUs and 8 GiB memory for optimal performance.`,
			expectHigh: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig1 := mh.Signature(tt.text1)
			sig2 := mh.Signature(tt.text2)
			similarity := Similarity(sig1, sig2)

			if tt.expectHigh && similarity <= 0.5 {
				t.Errorf("Expected high similarity (>0.5), got %f", similarity)
			} else if !tt.expectHigh && similarity > 0.5 {
				t.Errorf("Expected low similarity (<=0.5), got %f", similarity)
			}

			t.Logf("Similarity: %f", similarity)
		})
	}
}

func TestHashWithSeed(t *testing.T) {
	text := "test string"

	// Same string with same seed should produce same hash
	h1 := hashWithSeed(text, 42)
	h2 := hashWithSeed(text, 42)
	if h1 != h2 {
		t.Error("Same string and seed produced different hashes")
	}

	// Same string with different seeds should produce different hashes
	h3 := hashWithSeed(text, 123)
	if h1 == h3 {
		t.Error("Same string with different seeds produced identical hashes")
	}

	// Different strings with same seed should (probably) produce different hashes
	h4 := hashWithSeed("different string", 42)
	if h1 == h4 {
		t.Error("Different strings with same seed produced identical hashes")
	}
}

func TestYAMLTemplateExample(t *testing.T) {
	// Test with realistic Lima template content
	template1 := `
# Ubuntu-based development environment
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
    arch: x86_64

cpus: 4
memory: 4GiB

provision:
  - mode: system
    script: |
      apt-get update
      apt-get install -y docker.io git
`

	template2 := `
# Ubuntu development environment
images:
  - location: https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img
    arch: x86_64

cpus: 4
memory: 4GiB

provision:
  - mode: system
    script: |
      apt-get update
      apt-get install -y docker.io git vim
`

	template3 := `
# Alpine Linux container runtime
images:
  - location: https://dl-cdn.alpinelinux.org/alpine/v3.18/releases/x86_64/alpine-virt-3.18.0-x86_64.iso
    arch: x86_64

cpus: 2
memory: 2GiB

provision:
  - mode: system
    script: |
      apk add podman
`

	mh := New()

	sig1 := mh.Signature(template1)
	sig2 := mh.Signature(template2)
	sig3 := mh.Signature(template3)

	sim12 := Similarity(sig1, sig2)
	sim13 := Similarity(sig1, sig3)
	sim23 := Similarity(sig2, sig3)

	t.Logf("Similarity(template1, template2) = %f (very similar Ubuntu templates)", sim12)
	t.Logf("Similarity(template1, template3) = %f (different templates)", sim13)
	t.Logf("Similarity(template2, template3) = %f (different templates)", sim23)

	// Templates 1 and 2 are very similar (only differ by "vim")
	if sim12 < 0.7 {
		t.Errorf("Expected high similarity between template1 and template2, got %f", sim12)
	}

	// Templates 1 and 3 are quite different
	if sim13 > 0.5 {
		t.Errorf("Expected low similarity between template1 and template3, got %f", sim13)
	}

	// Templates 2 and 3 are quite different
	if sim23 > 0.5 {
		t.Errorf("Expected low similarity between template2 and template3, got %f", sim23)
	}
}

func BenchmarkSignature(b *testing.B) {
	mh := New()
	text := strings.Repeat("This is a test document for benchmarking MinHash performance. ", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mh.Signature(text)
	}
}

func BenchmarkSimilarity(b *testing.B) {
	mh := New()
	text := "This is a test document for benchmarking"
	sig1 := mh.Signature(text)
	sig2 := mh.Signature(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Similarity(sig1, sig2)
	}
}
