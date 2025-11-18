package discovery

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/go-github/v57/github"
	"github.com/lima-catalog/lima-catalog/pkg/types"
	"gopkg.in/yaml.v3"
)

// extractDomain extracts the domain from a URL or location string
func extractDomain(location string) string {
	// Handle template:// references (these are references to local templates)
	if strings.HasPrefix(location, "template://") {
		return "" // Not a real domain, skip
	}

	// Try to parse as URL
	if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") {
		parsedURL, err := url.Parse(location)
		if err == nil && parsedURL.Host != "" {
			return strings.ToLower(parsedURL.Host)
		}
	}

	// If it looks like a domain (contains dots but no slashes before first dot)
	if strings.Contains(location, ".") {
		parts := strings.Split(location, "/")
		domain := parts[0]
		if strings.Contains(domain, ".") {
			return strings.ToLower(domain)
		}
	}

	return ""
}

// IdentifyUnusualImages returns a list of image domains not in the official template set
func IdentifyUnusualImages(images []string, officialDomains map[string]bool) []string {
	seenDomains := make(map[string]bool)
	var unusual []string

	for _, img := range images {
		domain := extractDomain(img)
		if domain == "" {
			continue // Skip if we couldn't extract a domain
		}

		// Skip if already seen
		if seenDomains[domain] {
			continue
		}
		seenDomains[domain] = true

		// Check if this domain is in official domains
		if !officialDomains[domain] {
			unusual = append(unusual, domain)
		}
	}

	return unusual
}

// CalculateNotabilityScore computes a weighted score from notability metrics
// Higher score = more interesting/notable template
//
// Weights (in order of importance):
// - Message: 100 points (indicates template meant for reuse)
// - Provision scripts: 10 points per script + 1 point per 10 lines
// - Parameters: 20 points per param (indicates configurability)
// - Environment vars: 10 points per var
// - Probes: 5 points per probe + 1 point per 10 lines
// - Unusual images: 30 points if any unusual image domains (Lima uses first available)
// - Comment lines: 2 points per comment line (indicates documentation quality)
// - Repository stars: 1 point per 10 stars (capped at 50 points)
func CalculateNotabilityScore(metrics *types.NotabilityMetrics, repoStars int) float64 {
	if metrics == nil {
		return 0.0
	}

	score := 0.0

	// Message presence (strong signal for reusability)
	if metrics.MessageLength > 0 {
		score += 100.0
	}

	// Provision scripts (indicates customization/setup)
	score += float64(metrics.ProvisionCount) * 10.0
	score += float64(metrics.ProvisionTotalLines) / 10.0

	// Parameters (indicates configurability)
	score += float64(metrics.ParamCount) * 20.0

	// Environment variables (shows configuration effort)
	score += float64(metrics.EnvCount) * 10.0

	// Probes (less important than provision)
	score += float64(metrics.ProbeCount) * 5.0
	score += float64(metrics.ProbeTotalLines) / 10.0

	// Unusual images (indicates specialized use case)
	// Award bonus once if any unusual domains (Lima uses first available image)
	if len(metrics.UnusualImages) > 0 {
		score += 30.0
	}

	// Comment lines (indicates documentation quality)
	score += float64(metrics.CommentLineCount) * 2.0

	// Repository stars (capped to avoid dominating other factors)
	starsScore := float64(repoStars) / 10.0
	if starsScore > 50.0 {
		starsScore = 50.0
	}
	score += starsScore

	return score
}

// PopulateNotabilityMetrics creates NotabilityMetrics from TemplateInfo
func PopulateNotabilityMetrics(info *TemplateInfo, officialImages map[string]bool) *types.NotabilityMetrics {
	return &types.NotabilityMetrics{
		MessageLength:       info.MessageLength,
		ProvisionCount:      info.ProvisionCount,
		ProvisionTotalLines: info.ProvisionTotalLines,
		ProbeCount:          info.ProbeCount,
		ProbeTotalLines:     info.ProbeTotalLines,
		ParamCount:          info.ParamCount,
		EnvCount:            info.EnvCount,
		CommentLineCount:    info.CommentLineCount,
		UnusualImages:       IdentifyUnusualImages(info.Images, officialImages),
	}
}

// FetchOfficialImages retrieves the list of official image domains from lima-vm/lima repository
// by fetching and parsing all files in the templates/_images/ directory
func FetchOfficialImages(ctx context.Context, client *github.Client) (map[string]bool, error) {
	officialDomains := make(map[string]bool)

	// Fetch the contents of lima-vm/lima/templates/_images/ directory
	_, dirContents, _, err := client.Repositories.GetContents(
		ctx,
		"lima-vm",
		"lima",
		"templates/_images",
		&github.RepositoryContentGetOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch _images directory: %w", err)
	}

	// Process each .yaml file in the directory
	for _, item := range dirContents {
		if item.GetType() != "file" || !strings.HasSuffix(item.GetName(), ".yaml") {
			continue
		}

		// Fetch the file content
		fileContent, _, _, err := client.Repositories.GetContents(
			ctx,
			"lima-vm",
			"lima",
			item.GetPath(),
			&github.RepositoryContentGetOptions{},
		)
		if err != nil {
			// Log error but continue processing other files
			continue
		}

		content, err := fileContent.GetContent()
		if err != nil {
			continue
		}

		// Parse the YAML to extract image information
		var imageTemplate struct {
			Images []struct {
				Location string `yaml:"location"`
			} `yaml:"images"`
		}

		if err := yaml.Unmarshal([]byte(content), &imageTemplate); err != nil {
			continue
		}

		// Extract domains from image locations
		for _, img := range imageTemplate.Images {
			if img.Location != "" {
				domain := extractDomain(img.Location)
				if domain != "" {
					officialDomains[domain] = true
				}
			}
		}
	}

	return officialDomains, nil
}
