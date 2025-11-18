package prompt

import (
	"github.com/lima-catalog/lima-catalog/pkg/types"
	"github.com/lima-catalog/lima-catalog/pkg/validation"
)

// TemplateContext contains all context needed for LLM prompt generation
type TemplateContext struct {
	// Template information
	Template *types.Template

	// Repository information
	Repository *types.Repository

	// Organization information
	Organization *types.Organization

	// Template YAML content (raw)
	TemplateContent string

	// Template YAML comments (extracted)
	Comments []string

	// README content
	ReadmeContent string

	// References to this template file in the repository
	// Each reference includes the file path and surrounding context
	References []TemplateReference
}

// TemplateReference represents a reference to the template file in the repository
type TemplateReference struct {
	// File path relative to repo root
	FilePath string

	// Line number where the reference was found
	LineNumber int

	// Context lines before the reference
	BeforeContext []string

	// The line containing the reference
	MatchLine string

	// Context lines after the reference
	AfterContext []string
}

// PromptConfig controls how the prompt is generated
type PromptConfig struct {
	// ContextLines is the number of lines to include before/after references
	ContextLines int

	// IncludeComments controls whether YAML comments are included
	IncludeComments bool

	// IncludeReferences controls whether template file references are included
	IncludeReferences bool

	// IncludeReadme controls whether README is included
	IncludeReadme bool

	// MaxReadmeLength limits README content (0 = unlimited)
	MaxReadmeLength int

	// MaxReferenceFiles limits number of reference files to include (0 = unlimited)
	MaxReferenceFiles int
}

// DefaultPromptConfig returns sensible defaults for prompt generation
func DefaultPromptConfig() *PromptConfig {
	return &PromptConfig{
		ContextLines:      15,
		IncludeComments:   true,
		IncludeReferences: true,
		IncludeReadme:     true,
		MaxReadmeLength:   5000, // ~1000 tokens
		MaxReferenceFiles: 10,   // Limit to avoid prompt bloat
	}
}

// Validate validates the prompt configuration
func (c *PromptConfig) Validate() error {
	if err := validation.ValidateContextLines(c.ContextLines); err != nil {
		return err
	}

	if err := validation.ValidateMaxLength(c.MaxReadmeLength, "MaxReadmeLength"); err != nil {
		return err
	}

	if err := validation.ValidateMaxFiles(c.MaxReferenceFiles); err != nil {
		return err
	}

	return nil
}
