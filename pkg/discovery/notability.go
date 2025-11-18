package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v57/github"
	"github.com/lima-catalog/lima-catalog/pkg/types"
	"gopkg.in/yaml.v3"
)

// IdentifyUnusualImages returns a list of images not in the official template set
func IdentifyUnusualImages(images []string, officialImages map[string]bool) []string {
	var unusual []string
	for _, img := range images {
		imgLower := strings.ToLower(img)
		isOfficial := false
		for officialImg := range officialImages {
			if strings.Contains(imgLower, officialImg) {
				isOfficial = true
				break
			}
		}
		if !isOfficial {
			unusual = append(unusual, img)
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
// - Unusual images: 30 points per unusual image
// - Comment lines: 1 point per comment line (indicates documentation quality)
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
	score += float64(len(metrics.UnusualImages)) * 30.0

	// Comment lines (indicates documentation quality)
	score += float64(metrics.CommentLineCount)

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

// FetchOfficialImages retrieves the list of official image names from lima-vm/lima repository
// by fetching and parsing all files in the templates/_images/ directory
func FetchOfficialImages(ctx context.Context, client *github.Client) (map[string]bool, error) {
	officialImages := make(map[string]bool)

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

		// Extract image names from locations
		for _, img := range imageTemplate.Images {
			if img.Location != "" {
				imageName := extractImageName(img.Location)
				if imageName != "" && imageName != img.Location {
					// Only add if we successfully extracted a clean name
					officialImages[strings.ToLower(imageName)] = true
				}
			}
		}

		// Also add the base filename without extension as an image name
		// e.g., "_images/ubuntu.yaml" -> "ubuntu"
		baseName := strings.TrimSuffix(item.GetName(), ".yaml")
		if baseName != "" {
			officialImages[strings.ToLower(baseName)] = true
		}
	}

	return officialImages, nil
}
