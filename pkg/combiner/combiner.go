// Package combiner merges catalog data into frontend-optimized JSON Lines format.
//
// The combiner takes templates, repositories, and organizations data and creates
// a single combined dataset optimized for the frontend GitHub Pages site.
//
// Responsibilities:
//
//   - Merge template data with repository metadata
//   - Apply blocklist filtering
//   - Generate display names and descriptions
//   - Create GitHub raw URLs for templates
//   - Sort templates by org/repo/path for stable output
//
// Output Format:
//
// The combined data is written in JSON Lines format (one JSON object per line)
// to catalog.jsonl. Each line contains a CombinedTemplate with all data needed
// by the frontend, minimizing client-side processing.
//
// Blocklist Filtering:
//
// Templates matching blocklist patterns (CI configs, test files, etc.) are
// excluded from the combined output. The number of filtered templates is logged.
//
// Error Handling:
//
// Templates without repository data are skipped (logged).
// Invalid repository formats are logged and skipped.
// All errors are wrapped with context using fmt.Errorf with %w.
package combiner

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/discovery"
	"github.com/lima-catalog/lima-catalog/pkg/interfaces"
	"github.com/lima-catalog/lima-catalog/pkg/types"
	"github.com/lima-catalog/lima-catalog/pkg/validation"
)

// CombinedTemplate represents the optimized template data for the frontend
type CombinedTemplate struct {
	ID                      string                                 `json:"id"`
	Name                    string                                 `json:"name"`
	Description             string                                 `json:"description"`
	Keywords                []string                               `json:"keywords"`
	Category                string                                 `json:"category"`
	Repo                    string                                 `json:"repo"`
	Org                     string                                 `json:"org"`
	Path                    string                                 `json:"path"`
	Stars                   int                                    `json:"stars"`
	UpdatedAt               string                                 `json:"updated_at"`
	Official                bool                                   `json:"official"`
	URL                     string                                 `json:"url"`
	RawURL                  string                                 `json:"raw_url"`
	NotabilityScore         float64                                `json:"notability_score"`           // Weighted score for sorting by "interestingness"
	NotabilityScoreBreakdown *discovery.NotabilityScoreBreakdown   `json:"notability_score_breakdown,omitempty"` // Debug: score components
	SimilarTemplates        []types.SimilarTemplate                `json:"similar_templates,omitempty"` // Similar/duplicate templates detected by MinHash+LSH
	OriginalID              string                                 `json:"original_id,omitempty"`      // If this is a copy, the ID of the original template
}

// Combiner combines templates with repo/org metadata for frontend consumption
type Combiner struct {
	blocklist *types.Blocklist
	fs        interfaces.FileSystem
}

// NewCombiner creates a new combiner with blocklist and default FileSystem
func NewCombiner(blocklist *types.Blocklist) *Combiner {
	return NewCombinerWithFS(blocklist, interfaces.NewDefaultFileSystem())
}

// NewCombinerWithFS creates a new combiner with blocklist and custom FileSystem
// This allows mocking file I/O for testing
func NewCombinerWithFS(blocklist *types.Blocklist, fs interfaces.FileSystem) *Combiner {
	return &Combiner{
		blocklist: blocklist,
		fs:        fs,
	}
}

// CombineData creates the frontend-optimized templates-combined.jsonl file
func (c *Combiner) CombineData(templates []types.Template, repos []types.Repository, orgs []types.Organization, outputPath string) error {
	// Create lookup maps for efficient joining
	repoMap := make(map[string]types.Repository)
	for _, repo := range repos {
		repoMap[repo.ID] = repo
	}

	orgMap := make(map[string]types.Organization)
	for _, org := range orgs {
		orgMap[org.ID] = org
	}

	// Process templates
	var combined []CombinedTemplate
	filtered := 0

	for _, template := range templates {
		// Extract owner from repo for blocklist check
		owner, repoName, err := validation.ParseRepoID(template.Repo)
		if err != nil {
			fmt.Printf("Warning: Invalid repo format for template %s: %s\n", template.ID, template.Repo)
			continue
		}

		// Skip blocklisted templates
		if discovery.IsBlocklisted(owner, repoName, template.Path, c.blocklist) {
			filtered++
			continue
		}

		// Get repo data
		repo, hasRepo := repoMap[template.Repo]
		if !hasRepo {
			fmt.Printf("Warning: No repo data for template %s (repo: %s)\n", template.ID, template.Repo)
			continue
		}

		// Check org data exists (optional, just log warning)
		if _, hasOrg := orgMap[owner]; !hasOrg {
			fmt.Printf("Warning: No org data for template %s (org: %s)\n", template.ID, owner)
		}

		// Calculate notability score with breakdown
		breakdown := discovery.CalculateNotabilityScoreWithBreakdown(template.Notability, repo.Stars)

		// Create combined record
		combined = append(combined, CombinedTemplate{
			ID:                      template.ID,
			Name:                    c.getDisplayName(template),
			Description:             c.getDescription(template),
			Keywords:                template.Keywords,
			Category:                template.Category,
			Repo:                    template.Repo,
			Org:                     owner,
			Path:                    template.Path,
			Stars:                   repo.Stars,
			UpdatedAt:               c.formatDate(repo.UpdatedAt),
			Official:                template.IsOfficial,
			URL:                     template.URL,
			RawURL:                  c.getRawURL(template, repo),
			NotabilityScore:         breakdown.Total,
			NotabilityScoreBreakdown: &breakdown,
			SimilarTemplates:        template.SimilarTemplates,
			OriginalID:              template.OriginalID,
		})
	}

	// Sort combined templates by org/repo/path for stable output
	slices.SortFunc(combined, func(a, b CombinedTemplate) int {
		return cmp.Or(
			cmp.Compare(a.Org, b.Org),
			cmp.Compare(a.Repo, b.Repo),
			cmp.Compare(a.Path, b.Path),
		)
	})

	// Write to file using FileSystem interface
	file, err := c.fs.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, t := range combined {
		if err := encoder.Encode(t); err != nil {
			return fmt.Errorf("failed to encode template %s: %w", t.ID, err)
		}
	}

	fmt.Printf("\n=== Frontend Data Combination ===\n")
	fmt.Printf("Total templates: %d\n", len(templates))
	fmt.Printf("Filtered (blocklist): %d\n", filtered)
	fmt.Printf("Combined output: %d templates\n", len(combined))
	fmt.Printf("Output file: %s\n\n", outputPath)

	return nil
}

// getDisplayName returns the best display name for a template
func (c *Combiner) getDisplayName(template types.Template) string {
	if template.DisplayName != "" {
		return template.DisplayName
	}
	if template.Name != "" {
		return template.Name
	}
	return template.Path
}

// getDescription returns a description for the template
// Priority: short_description > joined keywords
func (c *Combiner) getDescription(template types.Template) string {
	if template.ShortDescription != "" {
		return template.ShortDescription
	}

	// Fallback: join first 3 keywords
	if len(template.Keywords) > 0 {
		count := 3
		if len(template.Keywords) < count {
			count = len(template.Keywords)
		}
		return strings.Join(template.Keywords[:count], ", ")
	}

	return "Lima VM template"
}

// getRawURL constructs the raw GitHub URL for template content
func (c *Combiner) getRawURL(template types.Template, repo types.Repository) string {
	// Convert GitHub blob URL to raw URL
	// From: https://github.com/owner/repo/blob/branch/path
	// To: https://raw.githubusercontent.com/owner/repo/branch/path

	branch := repo.DefaultBranch
	if branch == "" {
		branch = "main" // Fallback
	}

	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s",
		template.Repo,
		branch,
		template.Path)
}

// formatDate formats a time.Time to a simple date string
func (c *Combiner) formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}
