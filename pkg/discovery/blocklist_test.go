package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

func TestIsBlocklisted(t *testing.T) {
	blocklist := &types.Blocklist{
		Paths: []string{
			`^\.github/workflows/`,
			`^\.gitlab-ci\.ya?ml$`,
			`(^|/)tests?/`,  // Match /test/ or /tests/ or starting with test/ or tests/
			`^kubernetes/`,
			`(^|/)lima\.REJECTED\.yaml$`,           // Rejected templates
			`(^|/)rancher-desktop/lima/0/lima\.yaml$`,  // Rancher Desktop old config
			`(^|/)rancher-desktop/lima/_config/0\.yaml$`,  // Rancher Desktop newer config
		},
		Repos: []string{
			`^spamorg/`,
			`^someorg/spam-repo/`,
			`^someorg/repo/bad-template\.yaml$`,
			`^someorg/repo/subdir/`,
		},
	}

	// Compile patterns for performance testing
	if err := blocklist.CompilePatterns(); err != nil {
		t.Fatalf("failed to compile blocklist patterns: %v", err)
	}

	tests := []struct {
		name     string
		owner    string
		repo     string
		path     string
		expected bool
		reason   string
	}{
		// Path pattern tests
		{
			name:     "GitHub Actions workflow",
			owner:    "goodorg",
			repo:     "goodrepo",
			path:     ".github/workflows/ci.yaml",
			expected: true,
			reason:   "should block GitHub Actions workflows",
		},
		{
			name:     "GitLab CI yaml",
			owner:    "goodorg",
			repo:     "goodrepo",
			path:     ".gitlab-ci.yaml",
			expected: true,
			reason:   "should block GitLab CI .yaml files",
		},
		{
			name:     "GitLab CI yml",
			owner:    "goodorg",
			repo:     "goodrepo",
			path:     ".gitlab-ci.yml",
			expected: true,
			reason:   "should block GitLab CI .yml files",
		},
		{
			name:     "Test directory",
			owner:    "goodorg",
			repo:     "goodrepo",
			path:     "tests/template.yaml",
			expected: true,
			reason:   "should block files in tests/ directory",
		},
		{
			name:     "Test directory singular",
			owner:    "goodorg",
			repo:     "goodrepo",
			path:     "test/template.yaml",
			expected: true,
			reason:   "should block files in test/ directory",
		},
		{
			name:     "Kubernetes directory",
			owner:    "goodorg",
			repo:     "goodrepo",
			path:     "kubernetes/config.yaml",
			expected: true,
			reason:   "should block kubernetes/ directory",
		},
		{
			name:     "Rejected template at root",
			owner:    "m11o",
			repo:     "lima-stamps",
			path:     "lima.REJECTED.yaml",
			expected: true,
			reason:   "should block lima.REJECTED.yaml at root",
		},
		{
			name:     "Rejected template in subdirectory",
			owner:    "someorg",
			repo:     "somerepo",
			path:     "subdir/lima.REJECTED.yaml",
			expected: true,
			reason:   "should block lima.REJECTED.yaml in subdirectory",
		},
		{
			name:     "Rancher Desktop old config",
			owner:    "someorg",
			repo:     "somerepo",
			path:     "rancher-desktop/lima/0/lima.yaml",
			expected: true,
			reason:   "should block rancher-desktop old config",
		},
		{
			name:     "Rancher Desktop newer config",
			owner:    "someorg",
			repo:     "somerepo",
			path:     "rancher-desktop/lima/_config/0.yaml",
			expected: true,
			reason:   "should block rancher-desktop newer config",
		},
		{
			name:     "Valid template",
			owner:    "goodorg",
			repo:     "goodrepo",
			path:     "template.yaml",
			expected: false,
			reason:   "should allow valid templates",
		},

		// Repo pattern tests
		{
			name:     "Block entire org",
			owner:    "spamorg",
			repo:     "anyrepo",
			path:     "template.yaml",
			expected: true,
			reason:   "should block entire spamorg",
		},
		{
			name:     "Block specific repo",
			owner:    "someorg",
			repo:     "spam-repo",
			path:     "template.yaml",
			expected: true,
			reason:   "should block someorg/spam-repo",
		},
		{
			name:     "Allow different repo in same org",
			owner:    "someorg",
			repo:     "good-repo",
			path:     "template.yaml",
			expected: false,
			reason:   "should allow someorg/good-repo",
		},
		{
			name:     "Block specific template",
			owner:    "someorg",
			repo:     "repo",
			path:     "bad-template.yaml",
			expected: true,
			reason:   "should block someorg/repo/bad-template.yaml",
		},
		{
			name:     "Allow different template in same repo",
			owner:    "someorg",
			repo:     "repo",
			path:     "good-template.yaml",
			expected: false,
			reason:   "should allow someorg/repo/good-template.yaml",
		},
		{
			name:     "Block entire subdirectory",
			owner:    "someorg",
			repo:     "repo",
			path:     "subdir/template.yaml",
			expected: true,
			reason:   "should block someorg/repo/subdir/",
		},
		{
			name:     "Allow different subdirectory",
			owner:    "someorg",
			repo:     "repo",
			path:     "gooddir/template.yaml",
			expected: false,
			reason:   "should allow someorg/repo/gooddir/",
		},

		// Edge cases
		{
			name:     "Empty blocklist",
			owner:    "anyorg",
			repo:     "anyrepo",
			path:     ".github/workflows/test.yaml",
			expected: false,
			reason:   "empty blocklist should allow all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use empty blocklist for the "Empty blocklist" test
			testBlocklist := blocklist
			if tt.name == "Empty blocklist" {
				testBlocklist = &types.Blocklist{
					Paths: []string{},
					Repos: []string{},
				}
				// Compile empty patterns
				if err := testBlocklist.CompilePatterns(); err != nil {
					t.Fatalf("failed to compile empty blocklist: %v", err)
				}
			}

			result := IsBlocklisted(tt.owner, tt.repo, tt.path, testBlocklist)
			if result != tt.expected {
				t.Errorf("%s: expected %v but got %v (%s)", tt.name, tt.expected, result, tt.reason)
			}
		})
	}
}

func TestIsBlocklistedNilBlocklist(t *testing.T) {
	result := IsBlocklisted("anyorg", "anyrepo", "any/path.yaml", nil)
	if result != false {
		t.Errorf("nil blocklist should not block anything")
	}
}

func TestIsBlocklistedInvalidRegex(t *testing.T) {
	blocklist := &types.Blocklist{
		Paths: []string{
			`[invalid(regex`,  // Invalid regex
		},
		Repos: []string{
			`^goodorg/`,  // Valid regex
		},
	}

	// CompilePatterns should now fail on invalid regex
	err := blocklist.CompilePatterns()
	if err == nil {
		t.Error("expected error when compiling invalid regex")
	}
}

func TestLoadBlocklist(t *testing.T) {
	t.Run("Load valid blocklist", func(t *testing.T) {
		// Create a temporary blocklist file
		tmpDir := t.TempDir()
		blocklistPath := filepath.Join(tmpDir, "blocklist.yaml")

		content := `paths:
  - ^\.github/workflows/
  - ^test/
  - \.test\.yaml$
repos:
  - ^spamorg/
  - ^baduser/badrepo/
`
		if err := os.WriteFile(blocklistPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test blocklist: %v", err)
		}

		// Load blocklist
		blocklist, err := LoadBlocklist(blocklistPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if blocklist == nil {
			t.Fatal("expected non-nil blocklist")
		}

		// Verify paths were loaded
		if len(blocklist.Paths) != 3 {
			t.Errorf("expected 3 path patterns, got %d", len(blocklist.Paths))
		}

		// Verify repos were loaded
		if len(blocklist.Repos) != 2 {
			t.Errorf("expected 2 repo patterns, got %d", len(blocklist.Repos))
		}

		// Verify patterns were compiled
		if len(blocklist.GetCompiledPaths()) != 3 {
			t.Errorf("expected 3 compiled path patterns, got %d", len(blocklist.GetCompiledPaths()))
		}

		if len(blocklist.GetCompiledRepos()) != 2 {
			t.Errorf("expected 2 compiled repo patterns, got %d", len(blocklist.GetCompiledRepos()))
		}

		// Verify patterns work correctly
		if !IsBlocklisted("baduser", "badrepo", "test.yaml", blocklist) {
			t.Error("expected baduser/badrepo to be blocklisted")
		}

		if IsBlocklisted("gooduser", "goodrepo", "template.yaml", blocklist) {
			t.Error("expected gooduser/goodrepo/template.yaml not to be blocklisted")
		}
	})

	t.Run("Missing file returns empty blocklist", func(t *testing.T) {
		// Try to load non-existent file
		blocklist, err := LoadBlocklist("/nonexistent/path/blocklist.yaml")
		if err != nil {
			t.Fatalf("expected no error for missing file, got: %v", err)
		}

		if blocklist == nil {
			t.Fatal("expected non-nil blocklist for missing file")
		}

		// Should return empty blocklist
		if len(blocklist.Paths) != 0 {
			t.Errorf("expected 0 path patterns for missing file, got %d", len(blocklist.Paths))
		}

		if len(blocklist.Repos) != 0 {
			t.Errorf("expected 0 repo patterns for missing file, got %d", len(blocklist.Repos))
		}
	})

	t.Run("Invalid YAML returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		blocklistPath := filepath.Join(tmpDir, "invalid.yaml")

		// Write invalid YAML
		content := `paths:
  - valid pattern
repos:
  [unclosed bracket
invalid yaml`
		if err := os.WriteFile(blocklistPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		// Should return error
		blocklist, err := LoadBlocklist(blocklistPath)
		if err == nil {
			t.Error("expected error for invalid YAML")
		}

		if blocklist != nil {
			t.Error("expected nil blocklist for invalid YAML")
		}
	})

	t.Run("Invalid regex pattern returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		blocklistPath := filepath.Join(tmpDir, "bad-regex.yaml")

		// Write blocklist with invalid regex
		content := `paths:
  - ^valid/path
  - "[invalid(regex"
repos:
  - ^goodorg/
`
		if err := os.WriteFile(blocklistPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		// Should return error when compiling patterns
		blocklist, err := LoadBlocklist(blocklistPath)
		if err == nil {
			t.Error("expected error for invalid regex pattern")
		}

		if blocklist != nil {
			t.Error("expected nil blocklist for invalid regex")
		}
	})

	t.Run("Empty blocklist file", func(t *testing.T) {
		tmpDir := t.TempDir()
		blocklistPath := filepath.Join(tmpDir, "empty.yaml")

		// Write empty file
		if err := os.WriteFile(blocklistPath, []byte(""), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		// Should succeed with empty blocklist
		blocklist, err := LoadBlocklist(blocklistPath)
		if err != nil {
			t.Fatalf("unexpected error for empty file: %v", err)
		}

		if blocklist == nil {
			t.Fatal("expected non-nil blocklist for empty file")
		}

		// Verify no patterns loaded
		if len(blocklist.Paths) != 0 {
			t.Errorf("expected 0 path patterns for empty file, got %d", len(blocklist.Paths))
		}

		if len(blocklist.Repos) != 0 {
			t.Errorf("expected 0 repo patterns for empty file, got %d", len(blocklist.Repos))
		}
	})

	t.Run("Blocklist with only paths", func(t *testing.T) {
		tmpDir := t.TempDir()
		blocklistPath := filepath.Join(tmpDir, "paths-only.yaml")

		content := `paths:
  - ^\.github/
  - ^test/
`
		if err := os.WriteFile(blocklistPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		blocklist, err := LoadBlocklist(blocklistPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(blocklist.Paths) != 2 {
			t.Errorf("expected 2 path patterns, got %d", len(blocklist.Paths))
		}

		if len(blocklist.Repos) != 0 {
			t.Errorf("expected 0 repo patterns, got %d", len(blocklist.Repos))
		}
	})

	t.Run("Blocklist with only repos", func(t *testing.T) {
		tmpDir := t.TempDir()
		blocklistPath := filepath.Join(tmpDir, "repos-only.yaml")

		content := `repos:
  - ^spamorg/
  - ^baduser/badrepo$
`
		if err := os.WriteFile(blocklistPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		blocklist, err := LoadBlocklist(blocklistPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(blocklist.Paths) != 0 {
			t.Errorf("expected 0 path patterns, got %d", len(blocklist.Paths))
		}

		if len(blocklist.Repos) != 2 {
			t.Errorf("expected 2 repo patterns, got %d", len(blocklist.Repos))
		}
	})

	t.Run("Permission denied error", func(t *testing.T) {
		// Skip this test if running as root (permissions don't apply)
		if os.Getuid() == 0 {
			t.Skip("skipping permission test when running as root")
		}

		tmpDir := t.TempDir()
		blocklistPath := filepath.Join(tmpDir, "forbidden.yaml")

		// Create file and remove read permissions
		if err := os.WriteFile(blocklistPath, []byte("paths:\n  - test"), 0000); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
		defer os.Chmod(blocklistPath, 0644) // Restore permissions for cleanup

		blocklist, err := LoadBlocklist(blocklistPath)

		// Should return error (not IsNotExist)
		if err == nil {
			t.Error("expected error for permission denied")
		} else if os.IsNotExist(err) {
			t.Error("expected permission error, not IsNotExist")
		}

		if blocklist != nil {
			t.Error("expected nil blocklist on error")
		}
	})
}
