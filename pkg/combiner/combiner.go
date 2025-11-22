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
	"math"
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
		breakdown := discovery.CalculateNotabilityScoreWithBreakdown(template.Notability, owner, repoName, repo.Stars)

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
	defer func() { _ = file.Close() }()

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

	// Print notability score statistics
	printScoreStatistics(combined)

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

// ScoreStatistics holds statistical information for a score field
type ScoreStatistics struct {
	Name           string
	Min            float64
	Max            float64
	Median         float64
	Average        float64
	StdDeviation   float64
	ZeroPercent    float64
}

// calculateScoreStatistics computes statistics for a slice of score values
func calculateScoreStatistics(name string, values []float64) ScoreStatistics {
	if len(values) == 0 {
		return ScoreStatistics{Name: name}
	}

	// Sort for median calculation
	sorted := make([]float64, len(values))
	copy(sorted, values)
	slices.Sort(sorted)

	// Calculate min, max, median
	min := sorted[0]
	max := sorted[len(sorted)-1]
	var median float64
	if len(sorted)%2 == 0 {
		median = (sorted[len(sorted)/2-1] + sorted[len(sorted)/2]) / 2
	} else {
		median = sorted[len(sorted)/2]
	}

	// Calculate average
	sum := 0.0
	zeroCount := 0
	for _, v := range values {
		sum += v
		if v == 0 {
			zeroCount++
		}
	}
	avg := sum / float64(len(values))

	// Calculate standard deviation
	varianceSum := 0.0
	for _, v := range values {
		diff := v - avg
		varianceSum += diff * diff
	}
	stdDev := math.Sqrt(varianceSum / float64(len(values)))

	// Calculate zero percentage
	zeroPercent := (float64(zeroCount) / float64(len(values))) * 100.0

	return ScoreStatistics{
		Name:         name,
		Min:          min,
		Max:          max,
		Median:       median,
		Average:      avg,
		StdDeviation: stdDev,
		ZeroPercent:  zeroPercent,
	}
}

// printScoreStatistics prints statistics for all notability score fields
func printScoreStatistics(combined []CombinedTemplate) {
	if len(combined) == 0 {
		return
	}

	// Collect values for each score field
	var message, provision, parameters, envVars, probes, imageName, comments, stars, total []float64

	for _, t := range combined {
		if t.NotabilityScoreBreakdown != nil {
			message = append(message, t.NotabilityScoreBreakdown.Message)
			provision = append(provision, t.NotabilityScoreBreakdown.Provision)
			parameters = append(parameters, t.NotabilityScoreBreakdown.Parameters)
			envVars = append(envVars, t.NotabilityScoreBreakdown.EnvVars)
			probes = append(probes, t.NotabilityScoreBreakdown.Probes)
			imageName = append(imageName, t.NotabilityScoreBreakdown.ImageName)
			comments = append(comments, t.NotabilityScoreBreakdown.Comments)
			stars = append(stars, t.NotabilityScoreBreakdown.Stars)
			total = append(total, t.NotabilityScoreBreakdown.Total)
		}
	}

	// Calculate statistics for each field
	stats := []ScoreStatistics{
		calculateScoreStatistics("Message", message),
		calculateScoreStatistics("Provision", provision),
		calculateScoreStatistics("Parameters", parameters),
		calculateScoreStatistics("EnvVars", envVars),
		calculateScoreStatistics("Probes", probes),
		calculateScoreStatistics("ImageName", imageName),
		calculateScoreStatistics("Comments", comments),
		calculateScoreStatistics("Stars", stars),
		calculateScoreStatistics("Total", total),
	}

	// Print statistics table
	fmt.Printf("\n=== Notability Score Statistics ===\n")
	fmt.Printf("%-15s %8s %8s %8s %8s %8s %8s\n",
		"Name", "Min", "Max", "Median", "Avg", "StdDev", "Zero %")
	fmt.Printf("%-15s %8s %8s %8s %8s %8s %8s\n",
		"---------------", "--------", "--------", "--------", "--------", "--------", "--------")

	for _, s := range stats {
		fmt.Printf("%-15s %8.2f %8.2f %8.2f %8.2f %8.2f %7.1f%%\n",
			s.Name, s.Min, s.Max, s.Median, s.Average, s.StdDeviation, s.ZeroPercent)
	}
	fmt.Printf("\n")
}
