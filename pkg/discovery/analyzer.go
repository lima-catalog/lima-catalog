package discovery

import (
	"fmt"
	"strings"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

// Analyzer handles template analysis and categorization
type Analyzer struct {
	// LLMEnabled controls whether to use LLM for analysis
	LLMEnabled bool
	// LLMAPIKey is the API key for the LLM service
	LLMAPIKey string
}

// NewAnalyzer creates a new template analyzer
func NewAnalyzer(llmEnabled bool, apiKey string) *Analyzer {
	return &Analyzer{
		LLMEnabled: llmEnabled,
		LLMAPIKey:  apiKey,
	}
}

// AnalyzeTemplate performs full analysis on a template
func (a *Analyzer) AnalyzeTemplate(template *types.Template, repoInfo *types.Repository) error {
	// Step 1: Derive name
	template.Name = DeriveTemplateName(template.Path, template.Repo)
	template.DisplayName = GenerateDisplayName(template.Name)

	// Step 2: Parse template content
	templateInfo, err := ParseTemplate(template.URL)
	if err != nil {
		// If parsing fails, use basic info
		fmt.Printf("Warning: failed to parse template %s: %v\n", template.ID, err)
		templateInfo = &TemplateInfo{
			Images:     []string{},
			Keywords:   []string{},
			Categories: []string{},
		}
	}

	// Populate parsed fields
	template.Images = templateInfo.Images
	template.Arch = templateInfo.Arch
	template.Keywords = templateInfo.Keywords

	// Step 3: Infer basic category and description
	category, useCase := a.inferCategory(templateInfo, repoInfo)
	template.Category = category
	template.UseCase = useCase

	// Generate basic description
	template.ShortDescription = a.generateBasicDescription(template, templateInfo, repoInfo)

	// Step 4: Use LLM for enhanced analysis (if enabled)
	if a.LLMEnabled && a.LLMAPIKey != "" {
		if err := a.enhanceWithLLM(template, templateInfo, repoInfo); err != nil {
			fmt.Printf("Warning: LLM enhancement failed for %s: %v\n", template.ID, err)
			// Continue with basic analysis
		}
	}

	template.AnalyzedAt = time.Now()

	return nil
}

// inferCategory infers category and use case from parsed template info
func (a *Analyzer) inferCategory(info *TemplateInfo, repo *types.Repository) (string, string) {
	// Priority order for categories
	if info.HasK8s {
		return "orchestration", "kubernetes"
	}
	if info.HasDocker || info.HasPodman {
		return "containers", "container-runtime"
	}

	// Check categories from parsing
	if len(info.Categories) > 0 {
		primary := info.Categories[0]
		switch primary {
		case "development":
			return "development", "dev-environment"
		case "database":
			return "database", "data-storage"
		}
	}

	// Check repository topics
	if repo != nil {
		topics := strings.Join(repo.Topics, " ")
		topicsLower := strings.ToLower(topics)

		if strings.Contains(topicsLower, "security") || strings.Contains(topicsLower, "pentest") {
			return "security", "security-testing"
		}
		if strings.Contains(topicsLower, "test") || strings.Contains(topicsLower, "ci") {
			return "testing", "ci-cd"
		}
		if strings.Contains(topicsLower, "ml") || strings.Contains(topicsLower, "machine-learning") {
			return "ml", "machine-learning"
		}
	}

	// Default category
	return "general", "vm"
}

// generateBasicDescription creates a basic description without LLM
func (a *Analyzer) generateBasicDescription(template *types.Template, info *TemplateInfo, repo *types.Repository) string {
	parts := []string{}

	// Add OS information
	if len(info.Images) > 0 {
		parts = append(parts, fmt.Sprintf("%s-based", strings.Title(info.Images[0])))
	}

	// Add category
	parts = append(parts, template.Category)

	// Add key technologies
	if info.HasK8s {
		parts = append(parts, "with Kubernetes")
	} else if info.HasDocker {
		parts = append(parts, "with Docker")
	} else if info.HasPodman {
		parts = append(parts, "with Podman")
	}

	// Add architecture if specific
	if len(info.Arch) > 0 && len(info.Arch) < 3 {
		archStr := strings.Join(info.Arch, "/")
		parts = append(parts, fmt.Sprintf("(%s)", archStr))
	}

	description := strings.Join(parts, " ")

	// Add repository context if available
	if repo != nil && repo.Description != "" {
		description += ". " + repo.Description
	}

	return description
}

// enhanceWithLLM uses an LLM to generate better descriptions and categories
// This is a placeholder - will be implemented in the next step
func (a *Analyzer) enhanceWithLLM(template *types.Template, info *TemplateInfo, repo *types.Repository) error {
	// TODO: Implement LLM-based enhancement
	// This will call an LLM API (Claude, OpenAI, etc.) to generate:
	// - Better display name
	// - More detailed short description
	// - Full description
	// - Better keywords and categories

	return nil
}

// AnalyzeTemplates analyzes multiple templates
func (a *Analyzer) AnalyzeTemplates(templates []types.Template, repoMap map[string]*types.Repository) ([]types.Template, error) {
	analyzed := make([]types.Template, 0, len(templates))

	for i := range templates {
		template := &templates[i]

		// Skip if already analyzed and SHA hasn't changed
		if template.AnalyzedAt.After(template.LastChecked) {
			analyzed = append(analyzed, *template)
			continue
		}

		fmt.Printf("Analyzing [%d/%d] %s...\n", i+1, len(templates), template.ID)

		// Get repository info
		var repoInfo *types.Repository
		if repo, ok := repoMap[template.Repo]; ok {
			repoInfo = repo
		}

		// Analyze template
		if err := a.AnalyzeTemplate(template, repoInfo); err != nil {
			fmt.Printf("Warning: failed to analyze %s: %v\n", template.ID, err)
			// Continue with other templates
		}

		analyzed = append(analyzed, *template)

		// Rate limiting - be nice to external services
		time.Sleep(500 * time.Millisecond)
	}

	return analyzed, nil
}
