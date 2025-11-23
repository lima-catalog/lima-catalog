// Unit Tests: Template Naming (discovery package)
//
// High-level overview of what's being tested:
// - Deriving template names from file paths and repository context
// - Generic name detection (lima, template, config, default, etc.)
// - Generic directory detection (templates, examples, configs, etc.)
// - Fallback strategies: filename -> parent dir -> repo name
// - Name sanitization (lowercase, hyphen conversion, trimming)
// - Display name generation with smart capitalization
// - Acronym handling (short words stay uppercase)
// - Title case conversion for multi-word names
// - Special character handling (dots, underscores, spaces to hyphens)
// - Edge cases: empty paths, deep nesting, single-segment repos
// - Idempotent sanitization (running twice produces same result)

package discovery

import (
	"testing"
)

func TestDeriveTemplateName(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		repoFullName string
		expected     string
	}{
		{
			name:         "Descriptive filename",
			path:         "ubuntu.yaml",
			repoFullName: "owner/repo",
			expected:     "ubuntu",
		},
		{
			name:         "Descriptive filename with extension",
			path:         "fedora-dev.yaml",
			repoFullName: "owner/repo",
			expected:     "fedora-dev",
		},
		{
			name:         "Generic lima.yaml uses parent dir",
			path:         "kubernetes/lima.yaml",
			repoFullName: "owner/repo",
			expected:     "kubernetes",
		},
		{
			name:         "Generic template.yaml uses parent dir",
			path:         "docker/template.yaml",
			repoFullName: "owner/repo",
			expected:     "docker",
		},
		{
			name:         "Generic default.yaml uses parent dir",
			path:         "alpine/default.yaml",
			repoFullName: "owner/repo",
			expected:     "alpine",
		},
		{
			name:         "Generic name with generic parent uses repo name",
			path:         "templates/lima.yaml",
			repoFullName: "owner/myproject",
			expected:     "myproject",
		},
		{
			name:         "Generic name at root uses repo name",
			path:         "lima.yaml",
			repoFullName: "owner/lima-kubernetes",
			expected:     "lima-kubernetes",
		},
		{
			name:         "Nested path with generic parent",
			path:         "examples/test/lima.yaml",
			repoFullName: "owner/test-repo",
			expected:     "test",
		},
		{
			name:         "Deep nesting",
			path:         "vms/production/database/postgres.yaml",
			repoFullName: "owner/infra",
			expected:     "postgres",
		},
		{
			name:         "Path with underscores",
			path:         "my_custom_vm.yaml",
			repoFullName: "owner/repo",
			expected:     "my-custom-vm",
		},
		{
			name:         "Path with spaces (unlikely but possible)",
			path:         "my vm.yaml",
			repoFullName: "owner/repo",
			expected:     "my-vm",
		},
		{
			name:         "Path with dots",
			path:         "lima.ubuntu.22.04.yaml",
			repoFullName: "owner/repo",
			expected:     "lima-ubuntu-22-04",
		},
		{
			name:         "Generic config.yaml",
			path:         "k8s/config.yaml",
			repoFullName: "owner/kubernetes-setup",
			expected:     "k8s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeriveTemplateName(tt.path, tt.repoFullName)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIsGenericName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"lima", "lima", true},
		{"LIMA uppercase", "LIMA", true},
		{"template", "template", true},
		{"config", "config", true},
		{"default", "default", true},
		{"example", "example", true},
		{"test", "test", true},
		{"ubuntu", "ubuntu", false},
		{"fedora", "fedora", false},
		{"my-template", "my-template", false},
		{"k8s", "k8s", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGenericName(tt.input)
			if result != tt.expected {
				t.Errorf("isGenericName(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsGenericDir(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"templates", "templates", true},
		{"TEMPLATES uppercase", "TEMPLATES", true},
		{"configs", "configs", true},
		{"examples", "examples", true},
		{"lima", "lima", true},
		{".lima", ".lima", true},
		{"vms", "vms", true},
		{"kubernetes", "kubernetes", false},
		{"docker", "docker", false},
		{"my-app", "my-app", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGenericDir(tt.input)
			if result != tt.expected {
				t.Errorf("isGenericDir(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Lowercase conversion",
			input:    "MyTemplate",
			expected: "mytemplate",
		},
		{
			name:     "Underscores to hyphens",
			input:    "my_custom_template",
			expected: "my-custom-template",
		},
		{
			name:     "Spaces to hyphens",
			input:    "my custom template",
			expected: "my-custom-template",
		},
		{
			name:     "Dots to hyphens",
			input:    "ubuntu.22.04",
			expected: "ubuntu-22-04",
		},
		{
			name:     "Leading hyphens removed",
			input:    "-test",
			expected: "test",
		},
		{
			name:     "Trailing hyphens removed",
			input:    "test-",
			expected: "test",
		},
		{
			name:     "Multiple hyphens collapsed",
			input:    "my---template",
			expected: "my-template",
		},
		{
			name:     "Complex case",
			input:    "__My.Custom Template__",
			expected: "my-custom-template",
		},
		{
			name:     "Already clean",
			input:    "simple",
			expected: "simple",
		},
		{
			name:     "Mixed separators",
			input:    "my_template.v2 beta",
			expected: "my-template-v2-beta",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeName(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenerateDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple name",
			input:    "ubuntu",
			expected: "Ubuntu",
		},
		{
			name:     "Hyphenated name",
			input:    "fedora-dev",
			expected: "Fedora DEV", // "dev" is <=3 chars, treated as acronym
		},
		{
			name:     "Multiple words",
			input:    "my-custom-template",
			expected: "MY Custom Template", // "my" is <=3 chars, treated as acronym
		},
		{
			name:     "Acronym (short word)",
			input:    "k8s",
			expected: "K8S",
		},
		{
			name:     "CI/CD",
			input:    "ci-cd",
			expected: "CI CD",
		},
		{
			name:     "VM",
			input:    "vm",
			expected: "VM",
		},
		{
			name:     "API",
			input:    "api",
			expected: "API",
		},
		{
			name:     "Mixed short and long",
			input:    "api-gateway",
			expected: "API Gateway",
		},
		{
			name:     "Database template",
			input:    "postgresql-database",
			expected: "Postgresql Database",
		},
		{
			name:     "Already uppercase acronym",
			input:    "AWS",
			expected: "AWS",
		},
		{
			name:     "Numbers",
			input:    "ubuntu-22-04",
			expected: "Ubuntu 22 04",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateDisplayName(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDeriveTemplateNameEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		repoFullName string
		description  string
		allowEmpty   bool
	}{
		{
			name:         "Empty path fallback",
			path:         "",
			repoFullName: "owner/repo",
			description:  "Empty path returns empty (edge case)",
			allowEmpty:   true, // Current behavior - could be improved to fallback to repo name
		},
		{
			name:         "Single segment repo",
			path:         "lima.yaml",
			repoFullName: "standalone",
			description:  "Should handle repo without owner gracefully",
			allowEmpty:   false,
		},
		{
			name:         "Deep nesting with all generic",
			path:         "templates/examples/lima/lima.yaml",
			repoFullName: "owner/myrepo",
			description:  "Should fallback through generic directories to repo name",
			allowEmpty:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic
			result := DeriveTemplateName(tt.path, tt.repoFullName)
			if result == "" && !tt.allowEmpty {
				t.Errorf("expected non-empty result for %s", tt.description)
			}
		})
	}
}

func TestSanitizeNameIdempotent(t *testing.T) {
	// Sanitizing an already-sanitized name should produce the same result
	input := "my-clean-name"
	result1 := sanitizeName(input)
	result2 := sanitizeName(result1)

	if result1 != result2 {
		t.Errorf("sanitizeName is not idempotent: %q -> %q -> %q", input, result1, result2)
	}
}
