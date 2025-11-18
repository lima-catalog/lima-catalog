package discovery

import (
	"strings"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

// OfficialImages lists image names used in official lima-vm/lima templates
// Templates using other images get a notability boost for being "unusual"
var OfficialImages = map[string]bool{
	"ubuntu":     true,
	"debian":     true,
	"fedora":     true,
	"alpine":     true,
	"arch":       true,
	"almalinux":  true,
	"rocky":      true,
	"opensuse":   true,
	"centos":     true,
	"oracle":     true,
	"amazonlinux": true,
}

// IdentifyUnusualImages returns a list of images not in the official template set
func IdentifyUnusualImages(images []string) []string {
	var unusual []string
	for _, img := range images {
		imgLower := strings.ToLower(img)
		isOfficial := false
		for officialImg := range OfficialImages {
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

	// Repository stars (capped to avoid dominating other factors)
	starsScore := float64(repoStars) / 10.0
	if starsScore > 50.0 {
		starsScore = 50.0
	}
	score += starsScore

	return score
}

// PopulateNotabilityMetrics creates NotabilityMetrics from TemplateInfo
func PopulateNotabilityMetrics(info *TemplateInfo) *types.NotabilityMetrics {
	return &types.NotabilityMetrics{
		MessageLength:       info.MessageLength,
		ProvisionCount:      info.ProvisionCount,
		ProvisionTotalLines: info.ProvisionTotalLines,
		ProbeCount:          info.ProbeCount,
		ProbeTotalLines:     info.ProbeTotalLines,
		ParamCount:          info.ParamCount,
		EnvCount:            info.EnvCount,
		UnusualImages:       IdentifyUnusualImages(info.Images),
	}
}
