package combiner

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/discovery"
	"github.com/lima-catalog/lima-catalog/pkg/types"
)

// CombinedTemplate represents the optimized template data for the frontend
type CombinedTemplate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	Category    string   `json:"category"`
	Repo        string   `json:"repo"`
	Org         string   `json:"org"`
	Path        string   `json:"path"`
	Stars       int      `json:"stars"`
	UpdatedAt   string   `json:"updated_at"`
	Official    bool     `json:"official"`
	URL         string   `json:"url"`
	RawURL      string   `json:"raw_url"`
}

// Combiner combines templates with repo/org metadata for frontend consumption
type Combiner struct {
	blocklist *types.Blocklist
}

// NewCombiner creates a new combiner with blocklist
func NewCombiner(blocklist *types.Blocklist) *Combiner {
	return &Combiner{
		blocklist: blocklist,
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
		parts := strings.Split(template.Repo, "/")
		if len(parts) != 2 {
			fmt.Printf("Warning: Invalid repo format for template %s: %s\n", template.ID, template.Repo)
			continue
		}
		owner := parts[0]
		repoName := parts[1]

		// Skip blocklisted templates
		if discovery.IsBlocklisted(owner, repoName, template.Path, c.blocklist) {
			filtered++
			continue
		}

		// Skip templates with meta.noindex (future enhancement)
		// if template.MetaNoindex {
		//     filtered++
		//     continue
		// }

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

		// Create combined record
		combined = append(combined, CombinedTemplate{
			ID:          template.ID,
			Name:        c.getDisplayName(template),
			Description: c.getDescription(template),
			Keywords:    template.Keywords,
			Category:    template.Category,
			Repo:        template.Repo,
			Org:         owner,
			Path:        template.Path,
			Stars:       repo.Stars,
			UpdatedAt:   c.formatDate(repo.UpdatedAt),
			Official:    template.IsOfficial,
			URL:         template.URL,
			RawURL:      c.getRawURL(template, repo),
		})
	}

	// Sort combined templates by org/repo/path for stable output
	sort.Slice(combined, func(i, j int) bool {
		if combined[i].Org != combined[j].Org {
			return combined[i].Org < combined[j].Org
		}
		if combined[i].Repo != combined[j].Repo {
			return combined[i].Repo < combined[j].Repo
		}
		return combined[i].Path < combined[j].Path
	})

	// Write to file
	file, err := os.Create(outputPath)
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
// Priority: meta.description (future) > short_description > joined keywords
func (c *Combiner) getDescription(template types.Template) string {
	// Future: Check template.MetaDescription first
	// if template.MetaDescription != "" {
	//     return template.MetaDescription
	// }

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
