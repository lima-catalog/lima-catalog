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
	"context"
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
	ID                       string                              `json:"id"`
	Name                     string                              `json:"name"`
	Description              string                              `json:"description"`
	Keywords                 []string                            `json:"keywords"`
	Category                 string                              `json:"category"`
	Repo                     string                              `json:"repo"`
	Org                      string                              `json:"org"`
	Path                     string                              `json:"path"`
	Stars                    int                                 `json:"stars"`
	UpdatedAt                string                              `json:"updated_at"`
	Official                 bool                                `json:"official"`
	GithubURL                string                              `json:"github_url"`                           // github: scheme URL
	RawURL                   string                              `json:"raw_url"`                              // https: raw.githubusercontent.com URL
	NotabilityScore          float64                             `json:"notability_score"`                     // Weighted score for sorting by "interestingness"
	NotabilityScoreBreakdown *discovery.NotabilityScoreBreakdown `json:"notability_score_breakdown,omitempty"` // Debug: score components
	NotabilityScoreRanks     map[string]int                      `json:"notability_score_ranks,omitempty"`     // Rank for each component (1-based, with ties)
	SimilarTemplates         []types.SimilarTemplate             `json:"similar_templates,omitempty"`          // Similar/duplicate templates detected by MinHash+LSH
	OriginalID               string                              `json:"original_id,omitempty"`                // If this is a copy, the ID of the original template
	ValidationWarnings       []string                            `json:"validation_warnings,omitempty"`        // Validation warnings (macOS-specific settings, deprecated syntax, etc.)
}

// Combiner combines templates with repo/org metadata for frontend consumption
type Combiner struct {
	blocklist      *types.Blocklist
	fs             interfaces.FileSystem
	urlTransformer interfaces.URLTransformer
}

// NewCombiner creates a new combiner with blocklist and default FileSystem and URLTransformer
func NewCombiner(blocklist *types.Blocklist) *Combiner {
	return NewCombinerWithFS(blocklist, interfaces.NewDefaultFileSystem(), interfaces.NewDefaultURLTransformer())
}

// NewCombinerWithFS creates a new combiner with blocklist and custom FileSystem and URLTransformer
// This allows mocking file I/O and URL transformation for testing
func NewCombinerWithFS(blocklist *types.Blocklist, fs interfaces.FileSystem, urlTransformer interfaces.URLTransformer) *Combiner {
	return &Combiner{
		blocklist:      blocklist,
		fs:             fs,
		urlTransformer: urlTransformer,
	}
}

// calculateRanks computes ranks for all score components across all templates
// Returns a map from template ID to component ranks
func calculateRanks(templates []CombinedTemplate) map[string]map[string]int {
	if len(templates) == 0 {
		return nil
	}

	// Get all component keys from registry
	componentKeys := discovery.GetScoreComponentKeys()
	componentKeys = append(componentKeys, "total") // Add total as well

	// Initialize rank maps
	ranksByTemplateID := make(map[string]map[string]int)

	// For each component, calculate ranks
	for _, componentKey := range componentKeys {
		// Collect all scores for this component
		type scoreWithID struct {
			templateID string
			score      float64
		}
		var scores []scoreWithID

		for _, t := range templates {
			if t.NotabilityScoreBreakdown == nil {
				continue
			}

			var score float64
			if componentKey == "total" {
				score = t.NotabilityScoreBreakdown.Total
			} else {
				// Use the breakdown's ToMap to access component dynamically
				breakdownMap := t.NotabilityScoreBreakdown.ToMap()
				score = breakdownMap[componentKey]
			}

			scores = append(scores, scoreWithID{
				templateID: t.ID,
				score:      score,
			})
		}

		// Sort by score descending
		slices.SortFunc(scores, func(a, b scoreWithID) int {
			if a.score > b.score {
				return -1
			}
			if a.score < b.score {
				return 1
			}
			return 0
		})

		// Assign ranks (1-based, with ties)
		rank := 1
		for i, item := range scores {
			// Update rank if score changed (handle ties)
			if i > 0 && scores[i].score != scores[i-1].score {
				rank = i + 1
			}

			// Initialize map if needed
			if ranksByTemplateID[item.templateID] == nil {
				ranksByTemplateID[item.templateID] = make(map[string]int)
			}

			ranksByTemplateID[item.templateID][componentKey] = rank
		}
	}

	return ranksByTemplateID
}

// writeScoreMetadata writes score component registry metadata to a JSON file
// This allows the frontend to use proper display names and descriptions from the backend
func (c *Combiner) writeScoreMetadata(metadataPath string) error {
	// Export the registry
	metadata := struct {
		Components []discovery.ScoreComponentMetadata `json:"components"`
	}{
		Components: discovery.ScoreComponentRegistry,
	}

	file, err := c.fs.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer func() { _ = file.Close() }()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty-print for readability
	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	return nil
}

// CombineData creates the frontend-optimized templates-combined.jsonl file
func (c *Combiner) CombineData(ctx context.Context, templates []types.Template, repos []types.Repository, orgs []types.Organization, outputPath string) error {
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

		// Construct github: scheme URL
		githubURL := getGitHubSchemeURL(template)

		// Get raw URL using Lima's URL transformation
		rawURL, err := c.getRawURL(ctx, template)
		if err != nil {
			fmt.Printf("Warning: Failed to get raw URL for template %s: %v\n", template.ID, err)
			continue
		}

		// Create combined record
		// Get validation warnings from notability metrics
		var validationWarnings []string
		if template.Notability != nil && len(template.Notability.ValidationWarningMsgs) > 0 {
			validationWarnings = template.Notability.ValidationWarningMsgs
		}

		combined = append(combined, CombinedTemplate{
			ID:                       template.ID,
			Name:                     c.getDisplayName(template),
			Description:              c.getDescription(template),
			Keywords:                 template.Keywords,
			Category:                 template.Category,
			Repo:                     template.Repo,
			Org:                      owner,
			Path:                     template.Path,
			Stars:                    repo.Stars,
			UpdatedAt:                c.formatDate(repo.UpdatedAt),
			Official:                 template.IsOfficial,
			GithubURL:                githubURL,
			RawURL:                   rawURL,
			NotabilityScore:          breakdown.Total,
			NotabilityScoreBreakdown: &breakdown,
			SimilarTemplates:         template.SimilarTemplates,
			OriginalID:               template.OriginalID,
			ValidationWarnings:       validationWarnings,
		})
	}

	// Calculate ranks for all templates across all components
	ranksByTemplateID := calculateRanks(combined)

	// Assign ranks to templates
	for i := range combined {
		if ranks, hasRanks := ranksByTemplateID[combined[i].ID]; hasRanks {
			combined[i].NotabilityScoreRanks = ranks
		}
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
	fmt.Printf("Output file: %s\n", outputPath)

	// Write score metadata file (score_metadata.json in same directory as catalog)
	metadataPath := strings.TrimSuffix(outputPath, ".jsonl") + "_score_metadata.json"
	if err := c.writeScoreMetadata(metadataPath); err != nil {
		return fmt.Errorf("failed to write score metadata: %w", err)
	}
	fmt.Printf("Score metadata: %s\n\n", metadataPath)

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

// getGitHubSchemeURL constructs a github: scheme URL from a template
// This matches the frontend logic in web/js/urlHelpers.js
func getGitHubSchemeURL(template types.Template) string {
	// Parse owner/repo from template.Repo
	parts := strings.SplitN(template.Repo, "/", 2)
	if len(parts) != 2 {
		// Fallback for invalid format
		return fmt.Sprintf("github:%s/%s", template.Repo, template.Path)
	}
	owner := parts[0]
	repoName := parts[1]

	// Process path: remove .yaml extension and /.lima suffix
	path := template.Path
	path = strings.TrimSuffix(path, ".yaml")
	path = strings.TrimSuffix(path, "/.lima")

	// If path is just .lima or empty, can omit
	if path == ".lima" || path == "" {
		if owner == repoName {
			// org/org shorthand
			return fmt.Sprintf("github:%s", owner)
		}
		return fmt.Sprintf("github:%s/%s", owner, repoName)
	}

	// For org repos (owner == repo), use double slash
	if owner == repoName {
		return fmt.Sprintf("github:%s//%s", owner, path)
	}

	return fmt.Sprintf("github:%s/%s/%s", owner, repoName, path)
}

// getRawURL constructs the raw GitHub URL for template content using Lima's URL transformation
// This handles github: URLs and automatically resolves symlinks and redirects
func (c *Combiner) getRawURL(ctx context.Context, template types.Template) (string, error) {
	// Construct github: scheme URL
	githubURL := getGitHubSchemeURL(template)

	// Use URLTransformer to convert github: URL to https: URL
	// This automatically handles symlinks and redirects
	httpsURL, err := c.urlTransformer.TransformURL(ctx, githubURL)
	if err != nil {
		return "", fmt.Errorf("failed to transform github: URL %q: %w", githubURL, err)
	}

	return httpsURL, nil
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
	Name         string
	Min          float64
	Max          float64
	Median       float64
	Average      float64
	StdDeviation float64
	ZeroPercent  float64
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
