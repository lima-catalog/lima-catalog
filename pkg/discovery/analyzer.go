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
	// DetectDuplicates enables duplicate detection (default: true)
	DetectDuplicates bool
	// DuplicateSimilarityThreshold is the minimum similarity to report (default: 0.5)
	DuplicateSimilarityThreshold float64
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

// WithDetectDuplicates enables or disables duplicate detection
func WithDetectDuplicates(enabled bool) AnalyzerOption {
	return func(a *Analyzer) {
		a.DetectDuplicates = enabled
	}
}

// WithDuplicateSimilarityThreshold sets the minimum similarity for duplicate detection (0.0-1.0)
func WithDuplicateSimilarityThreshold(threshold float64) AnalyzerOption {
	return func(a *Analyzer) {
		a.DuplicateSimilarityThreshold = threshold
	}
}

// NewAnalyzer creates a new template analyzer with optional configuration.
//
// The analyzer is responsible for extracting metadata from Lima templates including
// OS images, categories, keywords, notability metrics, and duplicate detection.
//
// Options:
//   - WithForceAnalyze(bool): Force re-analysis of already analyzed templates
//   - WithHTTPClient(client): Use custom HTTP client (for testing)
//   - WithClock(clock): Use custom clock (for testing)
//   - WithOfficialKnowledge(ok): Set official knowledge for filtering (for testing)
//   - WithDetectDuplicates(bool): Enable/disable duplicate detection (default: true)
//   - WithDuplicateSimilarityThreshold(float64): Set similarity threshold 0.0-1.0 (default: 0.5)
//
// Returns a configured Analyzer with empty OfficialKnowledge.
// In production, load OfficialKnowledge from file or update it before analyzing templates.
//
// Duplicate Detection:
// By default, the analyzer detects duplicate templates using MinHash + LSH with a 50%
// similarity threshold. Templates with >50% similarity will be linked in SimilarTemplates.
//
// Example:
//
//	// Production code with default duplicate detection
//	analyzer := NewAnalyzer(WithForceAnalyze(true))
//
//	// Production code with custom threshold (70% similarity)
//	analyzer := NewAnalyzer(
//	    WithDuplicateSimilarityThreshold(0.7),
//	)
//
//	// Disable duplicate detection
//	analyzer := NewAnalyzer(
//	    WithDetectDuplicates(false),
//	)
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
		ForceAnalyze:                 false, // default
		HTTPClient:                   interfaces.NewDefaultHTTPClient(),
		Clock:                        interfaces.NewDefaultClock(),
		MinHash:                      minhash.New(), // Default MinHash with 128 hashes, 5-word shingles
		DetectDuplicates:             true,          // Enable duplicate detection by default
		DuplicateSimilarityThreshold: 0.5,           // Default 50% similarity threshold
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
		// Return error for invalid templates (e.g., missing required 'images' field)
		return fmt.Errorf("failed to parse template: %w", err)
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
// Duplicate Detection:
// After all templates are analyzed, duplicate detection runs automatically (unless disabled).
// This populates the SimilarTemplates field for each template with similar templates above
// the configured similarity threshold (default: 50%).
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

	// Detect duplicates if enabled
	if a.DetectDuplicates && len(analyzed) > 0 {
		fmt.Printf("Detecting duplicates with threshold %.2f...\n", a.DuplicateSimilarityThreshold)

		dd, err := minhash.NewDuplicateDetector(a.MinHash, a.DuplicateSimilarityThreshold)
		if err != nil {
			fmt.Printf("Warning: failed to create duplicate detector: %v\n", err)
			return analyzed, nil
		}

		analyzed, err = dd.DetectDuplicates(analyzed)
		if err != nil {
			fmt.Printf("Warning: failed to detect duplicates: %v\n", err)
			// Continue without duplicate detection - not critical
		} else {
			// Count templates with duplicates
			count := 0
			for _, t := range analyzed {
				if len(t.SimilarTemplates) > 0 {
					count++
				}
			}
			fmt.Printf("Found duplicates for %d templates\n", count)

			// Identify originals among exact duplicates
			identifyOriginals(analyzed, repoMap)

			// Count templates marked as copies
			copies := 0
			for _, t := range analyzed {
				if t.OriginalID != "" {
					copies++
				}
			}
			if copies > 0 {
				fmt.Printf("Identified %d templates as copies of originals\n", copies)
			}
		}
	}

	return analyzed, nil
}

// identifyOriginals determines which template is the "original" among exact duplicates.
// For each group of exact duplicates (similarity > 0.9), the original is determined by:
//  1. Official templates (lima-vm/lima) are always originals
//  2. Oldest repo creation date
//  3. Tie-breaker: higher star count
//  4. Final tie-breaker: alphabetically first (deterministic)
//
// Non-original templates get their OriginalID field set.
// The IsOriginal flag is set on SimilarTemplate entries.
func identifyOriginals(templates []types.Template, repoMap map[string]*types.Repository) {
	// Build a map for quick lookup
	templateMap := make(map[string]*types.Template)
	for i := range templates {
		templateMap[templates[i].ID] = &templates[i]
	}

	// Use union-find to group exact duplicates
	groups := buildExactDuplicateGroups(templates)

	// For each group, identify the original
	for _, group := range groups {
		if len(group) <= 1 {
			continue
		}

		// Find the original using heuristics
		originalID := findOriginal(group, templateMap, repoMap)

		// Mark all non-originals with OriginalID and update IsOriginal flags
		for _, id := range group {
			t := templateMap[id]
			if t == nil {
				continue
			}

			if id != originalID {
				t.OriginalID = originalID
			}

			// Update IsOriginal flag in SimilarTemplates
			for i := range t.SimilarTemplates {
				if t.SimilarTemplates[i].IsExactDuplicate() {
					t.SimilarTemplates[i].IsOriginal = (t.SimilarTemplates[i].ID == originalID)
				}
			}
		}
	}
}

// buildExactDuplicateGroups uses union-find to group templates that are exact duplicates.
// Returns a list of groups, where each group is a list of template IDs.
func buildExactDuplicateGroups(templates []types.Template) [][]string {
	// Union-find data structure
	parent := make(map[string]string)

	// Find with path compression
	var find func(id string) string
	find = func(id string) string {
		if parent[id] == "" {
			parent[id] = id
		}
		if parent[id] != id {
			parent[id] = find(parent[id])
		}
		return parent[id]
	}

	// Union two sets
	union := func(a, b string) {
		rootA, rootB := find(a), find(b)
		if rootA != rootB {
			parent[rootA] = rootB
		}
	}

	// Initialize all templates
	for i := range templates {
		parent[templates[i].ID] = templates[i].ID
	}

	// Union exact duplicates
	for i := range templates {
		t := &templates[i]
		for _, similar := range t.SimilarTemplates {
			if similar.IsExactDuplicate() {
				union(t.ID, similar.ID)
			}
		}
	}

	// Group by root
	groups := make(map[string][]string)
	for id := range parent {
		root := find(id)
		groups[root] = append(groups[root], id)
	}

	// Convert to slice
	result := make([][]string, 0, len(groups))
	for _, group := range groups {
		if len(group) > 1 {
			result = append(result, group)
		}
	}

	return result
}

// findOriginal determines which template in the group is the original.
// Priority: IsOfficial > oldest repo CreatedAt > highest Stars > alphabetical
func findOriginal(group []string, templateMap map[string]*types.Template, repoMap map[string]*types.Repository) string {
	var best string
	var bestRepo *types.Repository

	for _, id := range group {
		t := templateMap[id]
		if t == nil {
			continue
		}

		// Official templates always win
		if t.IsOfficial {
			return id
		}

		var repo *types.Repository
		if repoMap != nil {
			repo = repoMap[t.Repo]
		}

		// First candidate
		if best == "" {
			best = id
			bestRepo = repo
			continue
		}

		// Compare by repo creation date (older is better)
		if repo != nil && bestRepo != nil {
			if repo.CreatedAt.Before(bestRepo.CreatedAt) {
				best = id
				bestRepo = repo
				continue
			} else if bestRepo.CreatedAt.Before(repo.CreatedAt) {
				continue
			}
			// Same creation date, compare by stars
			if repo.Stars > bestRepo.Stars {
				best = id
				bestRepo = repo
				continue
			} else if bestRepo.Stars > repo.Stars {
				continue
			}
		} else if repo != nil && bestRepo == nil {
			// Prefer templates with repo info
			best = id
			bestRepo = repo
			continue
		} else if repo == nil && bestRepo != nil {
			continue
		}

		// Final tie-breaker: alphabetical (smaller ID wins for consistency)
		if id < best {
			best = id
			bestRepo = repo
		}
	}

	return best
}
