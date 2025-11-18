package prompt

import (
	"strings"
	"testing"

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
