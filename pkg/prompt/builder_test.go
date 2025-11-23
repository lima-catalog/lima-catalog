// Unit Tests: LLM Prompt Generation (prompt package)
//
// High-level overview of what's being tested:
// - Extracting YAML comments from template content
// - Filtering out divider lines (===, ---, empty lines)
// - Ignoring inline comments (only extracting line-start comments)
// - Default prompt configuration values
// - Building formatted markdown prompts for LLM consumption
// - Including/excluding optional sections (comments, references, README)
// - Template file path and repository context
// - Template YAML content display
// - Reference context with before/after lines
// - Parsing grep output for file references
// - Prompt configuration validation (context lines, max lengths)
// - Builder creation with token validation
// - Converting GitHub API objects to internal types
// - Handling nil config (defaults to default config)

package prompt

import (
	"strings"
	"testing"

	"github.com/google/go-github/v57/github"
	"github.com/lima-catalog/lima-catalog/pkg/types"
)

func TestExtractYAMLComments(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "simple comments",
			content: `# This is a comment
images:
  - location: ubuntu.qcow2
# Another comment
provision:
  - script: echo hello`,
			expected: []string{
				"This is a comment",
				"Another comment",
			},
		},
		{
			name: "comments with dividers",
			content: `# ===========================
# This is a real comment
# ---------------------------
images:
  - location: ubuntu.qcow2`,
			expected: []string{
				"This is a real comment",
			},
		},
		{
			name:     "no comments",
			content:  `images:\n  - location: ubuntu.qcow2`,
			expected: []string{},
		},
		{
			name: "inline comments ignored",
			content: `images:  # inline comment
  - location: ubuntu.qcow2  # another inline`,
			expected: []string{}, // Inline comments are not extracted (only line-start comments)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractYAMLComments(tt.content)
			if len(got) != len(tt.expected) {
				t.Errorf("extractYAMLComments() got %d comments, want %d", len(got), len(tt.expected))
				t.Errorf("Got: %v", got)
				t.Errorf("Want: %v", tt.expected)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("extractYAMLComments() comment[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestIsYAMLSectionDivider(t *testing.T) {
	tests := []struct {
		comment  string
		expected bool
	}{
		{"===========================", true},
		{"---------------------------", true},
		{"   ------   ", true},
		{"This is a real comment", false},
		{"", true}, // empty is considered a divider
		{"=== Section ===", false},
	}

	for _, tt := range tests {
		t.Run(tt.comment, func(t *testing.T) {
			got := isYAMLSectionDivider(tt.comment)
			if got != tt.expected {
				t.Errorf("isYAMLSectionDivider(%q) = %v, want %v", tt.comment, got, tt.expected)
			}
		})
	}
}

func TestDefaultPromptConfig(t *testing.T) {
	config := DefaultPromptConfig()

	if config.ContextLines != 15 {
		t.Errorf("DefaultPromptConfig() ContextLines = %d, want 15", config.ContextLines)
	}
	if !config.IncludeComments {
		t.Error("DefaultPromptConfig() IncludeComments = false, want true")
	}
	if !config.IncludeReferences {
		t.Error("DefaultPromptConfig() IncludeReferences = false, want true")
	}
	if !config.IncludeReadme {
		t.Error("DefaultPromptConfig() IncludeReadme = false, want true")
	}
	if config.MaxReadmeLength != 5000 {
		t.Errorf("DefaultPromptConfig() MaxReadmeLength = %d, want 5000", config.MaxReadmeLength)
	}
	if config.MaxReferenceFiles != 10 {
		t.Errorf("DefaultPromptConfig() MaxReferenceFiles = %d, want 10", config.MaxReferenceFiles)
	}
}

func TestFormatPrompt(t *testing.T) {
	builder := &Builder{
		config: DefaultPromptConfig(),
	}

	ctx := &TemplateContext{
		Template: &types.Template{
			ID:   "owner/repo/template.yaml",
			Repo: "owner/repo",
			Path: "template.yaml",
		},
		TemplateContent: `# Test template
images:
  - location: ubuntu.qcow2`,
		Comments: []string{"Test template"},
		References: []TemplateReference{
			{
				FilePath:      "docs/example.md",
				LineNumber:    42,
				BeforeContext: []string{"Before line 1", "Before line 2"},
				MatchLine:     "Reference to template.yaml",
				AfterContext:  []string{"After line 1", "After line 2"},
			},
		},
	}

	prompt := builder.FormatPrompt(ctx)

	// Check that key sections are present
	requiredSections := []string{
		"# Lima Template Analysis Request",
		"## Context",
		"### Template File",
		"### Template YAML Content",
		"### Template Comments",
		"### References to Template in Repository",
		"## Analysis Instructions",
	}

	for _, section := range requiredSections {
		if !strings.Contains(prompt, section) {
			t.Errorf("FormatPrompt() missing section: %q", section)
		}
	}

	// Check that template content is included
	if !strings.Contains(prompt, "images:") {
		t.Error("FormatPrompt() missing template content")
	}

	// Check that comments are included
	if !strings.Contains(prompt, "Test template") {
		t.Error("FormatPrompt() missing comments")
	}

	// Check that references are included
	if !strings.Contains(prompt, "docs/example.md") {
		t.Error("FormatPrompt() missing references")
	}
	if !strings.Contains(prompt, "Reference to template.yaml") {
		t.Error("FormatPrompt() missing reference match line")
	}
}

func TestFormatPrompt_WithoutOptionalSections(t *testing.T) {
	builder := &Builder{
		config: &PromptConfig{
			IncludeComments:   false,
			IncludeReferences: false,
			IncludeReadme:     false,
		},
	}

	ctx := &TemplateContext{
		Template: &types.Template{
			ID:   "owner/repo/template.yaml",
			Repo: "owner/repo",
			Path: "template.yaml",
		},
		TemplateContent: `images:\n  - location: ubuntu.qcow2`,
		// When config excludes these, GatherContext won't populate them
		Comments:      []string{},
		ReadmeContent: "",
		References:    []TemplateReference{},
	}

	prompt := builder.FormatPrompt(ctx)

	// Check that optional sections are NOT included (because they're empty)
	if strings.Contains(prompt, "### Template Comments") {
		t.Error("FormatPrompt() should not include comments section when empty")
	}
	if strings.Contains(prompt, "### References to Template") {
		t.Error("FormatPrompt() should not include references section when empty")
	}
	if strings.Contains(prompt, "### README Content") {
		t.Error("FormatPrompt() should not include README section when empty")
	}

	// But template content should still be there
	if !strings.Contains(prompt, "### Template YAML Content") {
		t.Error("FormatPrompt() missing template YAML content section")
	}
}

func TestParseGrepOutput_EmptyOutput(t *testing.T) {
	refs := parseGrepOutput("", "/tmp/test")
	if len(refs) != 0 {
		t.Errorf("parseGrepOutput(\"\") = %d refs, want 0", len(refs))
	}
}

func TestPromptConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *PromptConfig
		expectError bool
	}{
		{
			name:        "Valid default config",
			config:      DefaultPromptConfig(),
			expectError: false,
		},
		{
			name: "Valid custom config",
			config: &PromptConfig{
				ContextLines:      10,
				IncludeComments:   true,
				IncludeReferences: true,
				IncludeReadme:     true,
				MaxReadmeLength:   10000,
				MaxReferenceFiles: 5,
			},
			expectError: false,
		},
		{
			name: "Negative context lines",
			config: &PromptConfig{
				ContextLines:      -1,
				MaxReadmeLength:   5000,
				MaxReferenceFiles: 10,
			},
			expectError: true,
		},
		{
			name: "Zero context lines (valid)",
			config: &PromptConfig{
				ContextLines:      0,
				MaxReadmeLength:   5000,
				MaxReferenceFiles: 10,
			},
			expectError: false,
		},
		{
			name: "Negative MaxReadmeLength",
			config: &PromptConfig{
				ContextLines:      15,
				MaxReadmeLength:   -1,
				MaxReferenceFiles: 10,
			},
			expectError: true,
		},
		{
			name: "Zero MaxReadmeLength (valid - unlimited)",
			config: &PromptConfig{
				ContextLines:      15,
				MaxReadmeLength:   0,
				MaxReferenceFiles: 10,
			},
			expectError: false,
		},
		{
			name: "Negative MaxReferenceFiles",
			config: &PromptConfig{
				ContextLines:      15,
				MaxReadmeLength:   5000,
				MaxReferenceFiles: -1,
			},
			expectError: true,
		},
		{
			name: "Zero MaxReferenceFiles (valid - unlimited)",
			config: &PromptConfig{
				ContextLines:      15,
				MaxReadmeLength:   5000,
				MaxReferenceFiles: 0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestNewBuilder_InvalidToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "Empty token",
			token: "",
		},
		{
			name:  "Short token",
			token: "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := NewBuilder(t.Context(), tt.token, nil)
			if err == nil {
				t.Error("expected error for invalid token but got nil")
			}
			if builder != nil {
				t.Error("expected nil builder on error")
			}
		})
	}
}

func TestNewBuilder_InvalidConfig(t *testing.T) {
	// Valid token (length check only)
	validToken := "ghp_0123456789abcdefghijklmnopqrstuvwxyzABC"

	tests := []struct {
		name   string
		config *PromptConfig
	}{
		{
			name: "Negative context lines",
			config: &PromptConfig{
				ContextLines:      -1,
				MaxReadmeLength:   5000,
				MaxReferenceFiles: 10,
			},
		},
		{
			name: "Negative MaxReadmeLength",
			config: &PromptConfig{
				ContextLines:      15,
				MaxReadmeLength:   -1,
				MaxReferenceFiles: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := NewBuilder(t.Context(), validToken, tt.config)
			if err == nil {
				t.Error("expected error for invalid config but got nil")
			}
			if builder != nil {
				t.Error("expected nil builder on error")
			}
		})
	}
}

func TestNewBuilder_NilConfig(t *testing.T) {
	// Valid token (length check only)
	validToken := "ghp_0123456789abcdefghijklmnopqrstuvwxyzABC"

	// Should use default config when nil is passed
	builder, err := NewBuilder(t.Context(), validToken, nil)
	if err != nil {
		t.Fatalf("unexpected error with nil config: %v", err)
	}
	if builder == nil {
		t.Fatal("expected non-nil builder")
	}
	if builder.config == nil {
		t.Error("expected default config to be set")
	}
	if builder.config.ContextLines != 15 {
		t.Errorf("expected default ContextLines=15, got %d", builder.config.ContextLines)
	}
}

func TestConvertGitHubRepo(t *testing.T) {
	// Import github package for creating test objects
	starCount := 100
	forkCount := 50
	description := "Test repository"
	topics := []string{"lima", "vm"}
	homepage := "https://lima-vm.io"
	createdAt := github.Timestamp{}
	updatedAt := github.Timestamp{}

	ghRepo := &github.Repository{
		Name:            github.String("repo"),
		FullName:        github.String("owner/repo"),
		StargazersCount: &starCount,
		ForksCount:      &forkCount,
		Description:     &description,
		Topics:          topics,
		Homepage:        &homepage,
		CreatedAt:       &createdAt,
		UpdatedAt:       &updatedAt,
	}

	result := convertGitHubRepo(ghRepo, "owner", "repo")

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ID != "owner/repo" {
		t.Errorf("expected ID=owner/repo, got %s", result.ID)
	}
	if result.Owner != "owner" {
		t.Errorf("expected Owner=owner, got %s", result.Owner)
	}
	if result.Name != "repo" {
		t.Errorf("expected Name=repo, got %s", result.Name)
	}
	if result.Stars != 100 {
		t.Errorf("expected Stars=100, got %d", result.Stars)
	}
	if result.Forks != 50 {
		t.Errorf("expected Forks=50, got %d", result.Forks)
	}
	if result.Description != "Test repository" {
		t.Errorf("expected Description='Test repository', got %s", result.Description)
	}
	if len(result.Topics) != 2 {
		t.Errorf("expected 2 topics, got %d", len(result.Topics))
	}
	if result.Homepage != "https://lima-vm.io" {
		t.Errorf("expected Homepage=https://lima-vm.io, got %s", result.Homepage)
	}
}

func TestConvertGitHubUser(t *testing.T) {
	// Import github package for creating test objects
	id := int64(12345)
	login := "test-org"
	name := "Test Organization"
	bio := "Test bio"
	blog := "https://example.com"
	email := "test@example.com"
	location := "San Francisco"
	htmlURL := "https://github.com/test-org"
	avatarURL := "https://avatars.githubusercontent.com/u/12345"

	ghUser := &github.User{
		ID:        &id,
		Login:     &login,
		Name:      &name,
		Bio:       &bio,
		Blog:      &blog,
		Email:     &email,
		Location:  &location,
		HTMLURL:   &htmlURL,
		AvatarURL: &avatarURL,
	}

	result := convertGitHubUser(ghUser)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Login != "test-org" {
		t.Errorf("expected Login=test-org, got %s", result.Login)
	}
	if result.Name != "Test Organization" {
		t.Errorf("expected Name='Test Organization', got %s", result.Name)
	}
	if result.Description != "Test bio" {
		t.Errorf("expected Description='Test bio', got %s", result.Description)
	}
	if result.Blog != "https://example.com" {
		t.Errorf("expected Blog=https://example.com, got %s", result.Blog)
	}
	if result.Email != "test@example.com" {
		t.Errorf("expected Email=test@example.com, got %s", result.Email)
	}
}
