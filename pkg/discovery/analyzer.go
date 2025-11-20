package discovery

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/config"
	"github.com/lima-catalog/lima-catalog/pkg/interfaces"
	"github.com/lima-catalog/lima-catalog/pkg/minhash"
	"github.com/lima-catalog/lima-catalog/pkg/types"
)

// Analyzer handles template analysis and categorization
type Analyzer struct {
	// OfficialKnowledge contains known lines and images from lima-vm/lima templates
	OfficialKnowledge *OfficialKnowledge
	// ForceAnalyze forces re-analysis of all templates, even if already analyzed
	ForceAnalyze bool
	// HTTPClient for making HTTP requests (allows mocking in tests)
	HTTPClient interfaces.HTTPClient
	// Clock for time operations (allows mocking in tests)
	Clock interfaces.Clock
	// MinHash for generating duplicate detection signatures
	MinHash *minhash.MinHash
}

// AnalyzerOption configures an Analyzer
type AnalyzerOption func(*Analyzer)

// WithForceAnalyze sets whether to force re-analysis of already analyzed templates
func WithForceAnalyze(force bool) AnalyzerOption {
	return func(a *Analyzer) {
		a.ForceAnalyze = force
	}
}

// WithHTTPClient sets a custom HTTP client for template fetching
func WithHTTPClient(client interfaces.HTTPClient) AnalyzerOption {
	return func(a *Analyzer) {
		a.HTTPClient = client
	}
}

// WithClock sets a custom clock for time operations
func WithClock(clock interfaces.Clock) AnalyzerOption {
	return func(a *Analyzer) {
		a.Clock = clock
	}
}

// WithOfficialKnowledge sets the official knowledge for filtering known lines
func WithOfficialKnowledge(ok *OfficialKnowledge) AnalyzerOption {
	return func(a *Analyzer) {
		a.OfficialKnowledge = ok
	}
}

// NewAnalyzer creates a new template analyzer with optional configuration.
//
// The analyzer is responsible for extracting metadata from Lima templates including
// OS images, categories, keywords, and notability metrics.
//
// Options:
//   - WithForceAnalyze(bool): Force re-analysis of already analyzed templates
//   - WithHTTPClient(client): Use custom HTTP client (for testing)
//   - WithClock(clock): Use custom clock (for testing)
//   - WithOfficialKnowledge(ok): Set official knowledge for filtering (for testing)
//
// Returns a configured Analyzer with empty OfficialKnowledge.
// In production, load OfficialKnowledge from file or update it before analyzing templates.
//
// Example:
//
//	// Production code
//	analyzer := NewAnalyzer(WithForceAnalyze(true))
//
//	// Test code with mocks
//	analyzer := NewAnalyzer(
//	    WithHTTPClient(mockClient),
//	    WithClock(mockClock),
//	)
func NewAnalyzer(opts ...AnalyzerOption) *Analyzer {
	a := &Analyzer{
		OfficialKnowledge: &OfficialKnowledge{
			LastUpdate: time.Time{}, // Zero time
			KnownLines: OfficialKnownLines{
				Comments:  []string{},
				Provision: []string{},
				Probes:    []string{},
				Messages:  []string{},
			},
			Images: []string{},
		},
		ForceAnalyze: false, // default
		HTTPClient:   interfaces.NewDefaultHTTPClient(),
		Clock:        interfaces.NewDefaultClock(),
		MinHash:      minhash.New(), // Default MinHash with 128 hashes, 5-word shingles
	}

	// Apply options
	for _, opt := range opts {
		opt(a)
	}

	return a
}

// AnalyzeTemplate performs comprehensive analysis on a single template.
//
// The analysis process:
//  1. Derives template name from path (strips generic names like "lima.yaml")
//  2. Generates human-readable display name
//  3. Parses template YAML to extract images, arch, provision scripts, etc.
//  4. Populates notability metrics (provision count, params, unusual images, etc.)
//  5. Infers category and use case from template content and repo info
//  6. Generates basic description combining OS, category, and technologies
//  7. Sets AnalyzedAt timestamp to current time
//
// Parameters:
//   - ctx: Context for cancellation support
//   - template: Template to analyze (will be modified in place with analysis results)
//   - repoInfo: Repository metadata (used for topic-based categorization, can be nil)
//
// Returns error only for critical failures. Parse failures are logged as warnings
// and the template is populated with empty/fallback values.
//
// The template's existing fields (ID, Repo, Path, URL, etc.) are preserved.
// Only analysis-related fields are populated or updated.
func (a *Analyzer) AnalyzeTemplate(ctx context.Context, template *types.Template, repoInfo *types.Repository) error {
	// Check if context is cancelled before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Step 1: Derive name
	template.Name = DeriveTemplateName(template.Path, template.Repo)
	template.DisplayName = GenerateDisplayName(template.Name)

	// Step 2: Parse template content
	templateInfo, err := ParseTemplate(template.URL, a.HTTPClient)
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

	// Populate notability metrics (filtering out known lines from official templates)
	template.Notability = PopulateNotabilityMetrics(templateInfo, a.OfficialKnowledge)

	// Step 2.5: Generate MinHash signature for duplicate detection
	// Download template content for MinHash (may be cached by HTTPClient)
	rawContent, err := a.downloadTemplateContent(template.URL)
	if err != nil {
		fmt.Printf("Warning: failed to download template for MinHash %s: %v\n", template.ID, err)
		// Continue without MinHash signature - not critical
	} else {
		template.MinHashSignature = a.MinHash.Signature(rawContent)
	}

	// Step 3: Infer basic category and description
	category, useCase := a.inferCategory(templateInfo, repoInfo)
	template.Category = category
	template.UseCase = useCase

	// Generate basic description
	template.ShortDescription = a.generateBasicDescription(template, templateInfo, repoInfo)

	template.AnalyzedAt = a.Clock.Now()

	return nil
}

// downloadTemplateContent downloads the raw template content from URL.
//
// Converts GitHub blob URL to raw URL and downloads the content.
// This is a helper for MinHash signature generation.
func (a *Analyzer) downloadTemplateContent(url string) (string, error) {
	// Convert GitHub blob URL to raw URL
	// Pattern: https://github.com/owner/repo/blob/commit/path
	// Target: https://raw.githubusercontent.com/owner/repo/commit/path
	rawURL := strings.Replace(url, "github.com", "raw.githubusercontent.com", 1)
	rawURL = strings.Replace(rawURL, "/blob/", "/", 1)

	// Download template content
	resp, err := a.HTTPClient.Get(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to download template: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to download template: HTTP %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	return string(content), nil
}

// inferCategory infers the template's category and use case from content and repo metadata.
//
// Category inference follows priority order:
//  1. Container orchestration: Kubernetes presence (highest priority)
//  2. Container runtime: Docker or Podman presence
//  3. Template categories: From explicit categories in template parsing
//  4. Repository topics: Security, testing, ML keywords
//  5. Default: "general" / "vm" (fallback)
//
// Parameters:
//   - info: Parsed template information (images, runtimes, categories)
//   - repo: Repository metadata (optional, used for topic analysis)
//
// Returns:
//   - category: Primary category (e.g., "orchestration", "containers", "development")
//   - useCase: Specific use case (e.g., "kubernetes", "container-runtime", "dev-environment")
//
// The function prioritizes technical capabilities over metadata for more accurate categorization.
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

// generateBasicDescription creates a concise template description without using LLM.
//
// Constructs description from:
//   - Primary OS image (e.g., "Ubuntu-based")
//   - Category (e.g., "development")
//   - Key technologies (Kubernetes, Docker, Podman)
//   - Architecture (if specific, e.g., "arm64")
//   - Repository description (if available)
//
// Parameters:
//   - template: Template with category already populated
//   - info: Parsed template information (images, arch, runtime flags)
//   - repo: Repository metadata (optional, adds context from repo description)
//
// Returns a human-readable description string suitable for display.
//
// Example outputs:
//   - "Ubuntu-based development with Docker (amd64)"
//   - "Alpine-based orchestration with Kubernetes"
//   - "Fedora-based containers with Podman. Repository for Lima VM templates."
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

// AnalyzeTemplates analyzes multiple templates with intelligent skip logic.
//
// Implements incremental analysis to avoid redundant work:
//   - Skips templates where AnalyzedAt > LastUpdated (already analyzed, content unchanged)
//   - Unless ForceAnalyze is true, which forces re-analysis of all templates
//   - Templates that fail analysis are skipped (logged but not included in output)
//
// Parameters:
//   - ctx: Context for cancellation support
//   - templates: Templates to analyze
//   - repoMap: Map of repository ID to metadata (for topic-based categorization)
//
// Returns:
//   - List of successfully analyzed templates (may be shorter than input if failures occur)
//   - Error only for critical failures (individual failures are logged but not fatal)
//
// Progress is printed for each template: "Analyzing [1/100] owner/repo/template.yaml..."
//
// Rate Limiting:
// Adds a delay (config.MetadataAPIDelay) between templates to avoid overwhelming
// external services when fetching template content.
//
// Use Cases:
//   - Incremental mode: Only analyze new/changed templates (efficient)
//   - Full refresh: Set ForceAnalyze=true to re-analyze everything (after logic changes)
func (a *Analyzer) AnalyzeTemplates(ctx context.Context, templates []types.Template, repoMap map[string]*types.Repository) ([]types.Template, error) {
	analyzed := make([]types.Template, 0, len(templates))

	for i := range templates {
		// Check for context cancellation before processing each template
		select {
		case <-ctx.Done():
			return analyzed, ctx.Err()
		default:
		}

		template := &templates[i]

		// Skip if already analyzed and content hasn't changed since analysis
		// LastUpdated tracks when content (SHA) changed, so if we analyzed after last update, skip
		// Unless ForceAnalyze is enabled, which forces re-analysis of all templates
		if !a.ForceAnalyze && !template.LastUpdated.IsZero() && template.AnalyzedAt.After(template.LastUpdated) {
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
		if err := a.AnalyzeTemplate(ctx, template, repoInfo); err != nil {
			// Check if error is due to context cancellation
			if ctx.Err() != nil {
				return analyzed, ctx.Err()
			}
			fmt.Printf("Warning: failed to analyze %s: %v (skipping)\n", template.ID, err)
			// Skip templates that fail analysis to avoid saving incomplete data
			continue
		}

		analyzed = append(analyzed, *template)

		// Rate limiting - be nice to external services
		// Use select to allow context cancellation during sleep
		select {
		case <-ctx.Done():
			return analyzed, ctx.Err()
		case <-time.After(config.MetadataAPIDelay):
		}
	}

	return analyzed, nil
}
