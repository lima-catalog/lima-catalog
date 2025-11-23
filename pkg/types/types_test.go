// Unit Tests: Core Types and Data Structures (types package)
//
// High-level overview of what's being tested:
// - SimilarTemplate.IsExactDuplicate() logic
// - Similarity threshold detection (90%+ is exact duplicate)
// - Template similarity score calculations
// - Edge cases around similarity thresholds
// - Type conversions and serialization
// - Data structure integrity

package types

import (
	"testing"
)

func TestSimilarTemplate_IsExactDuplicate(t *testing.T) {
	tests := []struct {
		name       string
		similarity float64
		expected   bool
	}{
		{
			name:       "Exact duplicate (100% similarity)",
			similarity: 1.0,
			expected:   true,
		},
		{
			name:       "Very high similarity (95%)",
			similarity: 0.95,
			expected:   true,
		},
		{
			name:       "Just above threshold (90.1%)",
			similarity: 0.901,
			expected:   true,
		},
		{
			name:       "Exactly at threshold (90%)",
			similarity: 0.9,
			expected:   false,
		},
		{
			name:       "Just below threshold (89.9%)",
			similarity: 0.899,
			expected:   false,
		},
		{
			name:       "Medium similarity (70%)",
			similarity: 0.7,
			expected:   false,
		},
		{
			name:       "Low similarity (30%)",
			similarity: 0.3,
			expected:   false,
		},
		{
			name:       "No similarity (0%)",
			similarity: 0.0,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := SimilarTemplate{
				ID:         "test/repo/template.yaml",
				Similarity: tt.similarity,
			}

			result := st.IsExactDuplicate()
			if result != tt.expected {
				t.Errorf("IsExactDuplicate() with similarity %.3f = %v, expected %v",
					tt.similarity, result, tt.expected)
			}
		})
	}
}

func TestBlocklist_CompilePatterns(t *testing.T) {
	tests := []struct {
		name        string
		blocklist   Blocklist
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid path patterns",
			blocklist: Blocklist{
				Paths: []string{
					"^test/",
					`\.txt$`,
					`^examples/.*\.yaml$`,
				},
				Repos: []string{},
			},
			expectError: false,
		},
		{
			name: "Valid repo patterns",
			blocklist: Blocklist{
				Paths: []string{},
				Repos: []string{
					"^spamorg/",
					"^baduser/badrepo$",
					"test.*repo",
				},
			},
			expectError: false,
		},
		{
			name: "Mixed valid patterns",
			blocklist: Blocklist{
				Paths: []string{
					`^\.github/workflows/`,
					"test",
				},
				Repos: []string{
					"^org/",
					"repo$",
				},
			},
			expectError: false,
		},
		{
			name: "Empty blocklist",
			blocklist: Blocklist{
				Paths: []string{},
				Repos: []string{},
			},
			expectError: false,
		},
		{
			name: "Invalid path pattern",
			blocklist: Blocklist{
				Paths: []string{
					"^test/",
					"[invalid",  // Unclosed bracket
				},
				Repos: []string{},
			},
			expectError: true,
			errorMsg:    "invalid path pattern",
		},
		{
			name: "Invalid repo pattern",
			blocklist: Blocklist{
				Paths: []string{},
				Repos: []string{
					"^org/",
					"(unclosed",  // Unclosed parenthesis
				},
			},
			expectError: true,
			errorMsg:    "invalid repo pattern",
		},
		{
			name: "Multiple invalid patterns",
			blocklist: Blocklist{
				Paths: []string{
					"valid",
					"*invalid",  // Invalid regex
				},
				Repos: []string{},
			},
			expectError: true,
			errorMsg:    "invalid path pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.blocklist.CompilePatterns()

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Verify compiled patterns are created
				if len(tt.blocklist.Paths) > 0 && len(tt.blocklist.compiledPaths) != len(tt.blocklist.Paths) {
					t.Errorf("expected %d compiled path patterns, got %d",
						len(tt.blocklist.Paths), len(tt.blocklist.compiledPaths))
				}

				if len(tt.blocklist.Repos) > 0 && len(tt.blocklist.compiledRepos) != len(tt.blocklist.Repos) {
					t.Errorf("expected %d compiled repo patterns, got %d",
						len(tt.blocklist.Repos), len(tt.blocklist.compiledRepos))
				}
			}
		})
	}
}

func TestBlocklist_GetCompiledPaths(t *testing.T) {
	bl := Blocklist{
		Paths: []string{
			"^test/",
			`\.md$`,
		},
		Repos: []string{},
	}

	// Compile patterns
	err := bl.CompilePatterns()
	if err != nil {
		t.Fatalf("failed to compile patterns: %v", err)
	}

	// Get compiled patterns
	compiled := bl.GetCompiledPaths()

	if len(compiled) != 2 {
		t.Errorf("expected 2 compiled patterns, got %d", len(compiled))
	}

	// Test that they work
	if !compiled[0].MatchString("test/file.txt") {
		t.Error("expected first pattern to match 'test/file.txt'")
	}

	if !compiled[1].MatchString("README.md") {
		t.Error("expected second pattern to match 'README.md'")
	}

	if compiled[0].MatchString("other/file.txt") {
		t.Error("expected first pattern not to match 'other/file.txt'")
	}
}

func TestBlocklist_GetCompiledRepos(t *testing.T) {
	bl := Blocklist{
		Paths: []string{},
		Repos: []string{
			"^spamorg/",
			"^user/badrepo$",
		},
	}

	// Compile patterns
	err := bl.CompilePatterns()
	if err != nil {
		t.Fatalf("failed to compile patterns: %v", err)
	}

	// Get compiled patterns
	compiled := bl.GetCompiledRepos()

	if len(compiled) != 2 {
		t.Errorf("expected 2 compiled patterns, got %d", len(compiled))
	}

	// Test that they work
	if !compiled[0].MatchString("spamorg/anyrepo") {
		t.Error("expected first pattern to match 'spamorg/anyrepo'")
	}

	if !compiled[1].MatchString("user/badrepo") {
		t.Error("expected second pattern to match 'user/badrepo'")
	}

	if compiled[1].MatchString("user/goodrepo") {
		t.Error("expected second pattern not to match 'user/goodrepo'")
	}
}

func TestBlocklist_CompilePatterns_Integration(t *testing.T) {
	// Test that compiled patterns work correctly for blocking
	bl := Blocklist{
		Paths: []string{
			`^\.github/workflows/`,
			"^test/",
			`\.test\.yaml$`,
		},
		Repos: []string{
			"^spamorg/",
			"^testuser/testrepo$",
		},
	}

	err := bl.CompilePatterns()
	if err != nil {
		t.Fatalf("failed to compile patterns: %v", err)
	}

	// Test path matching
	pathTests := []struct {
		path     string
		expected bool
	}{
		{".github/workflows/ci.yml", true},
		{"test/example.yaml", true},
		{"examples/ubuntu.test.yaml", true},
		{"templates/ubuntu.yaml", false},
		{"README.md", false},
	}

	compiledPaths := bl.GetCompiledPaths()
	for _, tt := range pathTests {
		matched := false
		for _, pattern := range compiledPaths {
			if pattern.MatchString(tt.path) {
				matched = true
				break
			}
		}

		if matched != tt.expected {
			t.Errorf("path %q: expected matched=%v, got matched=%v", tt.path, tt.expected, matched)
		}
	}

	// Test repo matching
	repoTests := []struct {
		repo     string
		expected bool
	}{
		{"spamorg/anyrepo", true},
		{"testuser/testrepo", true},
		{"goodorg/goodrepo", false},
		{"testuser/otherrepo", false},
	}

	compiledRepos := bl.GetCompiledRepos()
	for _, tt := range repoTests {
		matched := false
		for _, pattern := range compiledRepos {
			if pattern.MatchString(tt.repo) {
				matched = true
				break
			}
		}

		if matched != tt.expected {
			t.Errorf("repo %q: expected matched=%v, got matched=%v", tt.repo, tt.expected, matched)
		}
	}
}

func TestBlocklist_EmptyPatterns(t *testing.T) {
	// Test that empty blocklist compiles without error
	bl := Blocklist{}

	err := bl.CompilePatterns()
	if err != nil {
		t.Errorf("unexpected error compiling empty blocklist: %v", err)
	}

	if len(bl.GetCompiledPaths()) != 0 {
		t.Error("expected 0 compiled path patterns for empty blocklist")
	}

	if len(bl.GetCompiledRepos()) != 0 {
		t.Error("expected 0 compiled repo patterns for empty blocklist")
	}
}

// Helper function for substring matching
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		containsInner(s, substr)))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
