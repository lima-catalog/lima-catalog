package discovery

import (
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
		},
		Repos: []string{
			`^spamorg/`,
			`^someorg/spam-repo/`,
			`^someorg/repo/bad-template\.yaml$`,
			`^someorg/repo/subdir/`,
		},
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

	// Should not panic on invalid regex, just skip it
	result := IsBlocklisted("goodorg", "repo", "test.yaml", blocklist)
	if result != true {
		t.Errorf("should still match valid repo pattern")
	}

	// Invalid regex should be skipped
	result = IsBlocklisted("otherorg", "repo", "test.yaml", blocklist)
	if result != false {
		t.Errorf("invalid path regex should be skipped, not block")
	}
}
