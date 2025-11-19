package discovery

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/go-github/v57/github"
	"github.com/lima-catalog/lima-catalog/pkg/config"
	"github.com/lima-catalog/lima-catalog/pkg/types"
	"gopkg.in/yaml.v3"
)

// extractDomain extracts the top-level domain from a URL or location string
// e.g., "downloads.ubuntu.com" -> "ubuntu.com", "nixos.org" -> "nixos.org"
func extractDomain(location string) string {
	// Handle template:// references (these are references to local templates)
	if strings.HasPrefix(location, "template://") {
		return "" // Not a real domain, skip
	}

	var host string

	// Try to parse as URL
	if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") {
		parsedURL, err := url.Parse(location)
		if err == nil && parsedURL.Host != "" {
			host = parsedURL.Host
		}
	} else if strings.Contains(location, ".") {
		// If it looks like a domain (contains dots but no slashes before first dot)
		parts := strings.Split(location, "/")
		if len(parts) > 0 && strings.Contains(parts[0], ".") {
			host = parts[0]
		}
	}

	if host == "" {
		return ""
	}

	// Remove port if present
	host = strings.Split(host, ":")[0]
	host = strings.ToLower(host)

	// Extract registrable domain (last 2 parts)
	domainParts := strings.Split(host, ".")
	if len(domainParts) < 2 {
		return host // Already a simple domain
	}

	// Return last 2 parts (e.g., ubuntu.com from downloads.ubuntu.com)
	return strings.Join(domainParts[len(domainParts)-2:], ".")
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

// NotabilityScoreBreakdown contains the individual components of the notability score
type NotabilityScoreBreakdown struct {
	Message       float64 `json:"message"`        // 100 if has message
	Provision     float64 `json:"provision"`      // 10 per script + 1 per 10 lines
	Parameters    float64 `json:"parameters"`     // 20 per param
	EnvVars       float64 `json:"env_vars"`       // 10 per var
	Probes        float64 `json:"probes"`         // 5 per probe + 1 per 10 lines
	UnusualImages float64 `json:"unusual_images"` // 30 if any unusual domains
	Comments      float64 `json:"comments"`       // 2 per comment line
	Stars         float64 `json:"stars"`          // 1 per 10 stars (capped at 50)
	Total         float64 `json:"total"`          // Sum of all components
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
	breakdown := CalculateNotabilityScoreWithBreakdown(metrics, repoStars)
	return breakdown.Total
}

// CalculateNotabilityScoreWithBreakdown computes the score and returns the breakdown
func CalculateNotabilityScoreWithBreakdown(metrics *types.NotabilityMetrics, repoStars int) NotabilityScoreBreakdown {
	breakdown := NotabilityScoreBreakdown{}

	if metrics == nil {
		return breakdown
	}

	// Get default weights (can be made configurable in the future)
	weights := config.DefaultNotabilityWeights()

	// Message presence (strong signal for reusability)
	if metrics.MessageLength > 0 {
		breakdown.Message = weights.Message
	}

	// Provision scripts (indicates customization/setup)
	breakdown.Provision = float64(metrics.ProvisionCount)*weights.ProvisionBase +
		float64(metrics.ProvisionTotalLines)*weights.ProvisionLine

	// Parameters (indicates configurability)
	breakdown.Parameters = float64(metrics.ParamCount) * weights.Parameter

	// Environment variables (shows configuration effort)
	breakdown.EnvVars = float64(metrics.EnvCount) * weights.EnvVar

	// Probes (less important than provision)
	breakdown.Probes = float64(metrics.ProbeCount)*weights.ProbeBase +
		float64(metrics.ProbeTotalLines)*weights.ProbeLine

	// Unusual images (indicates specialized use case)
	// Award bonus once if any unusual domains (Lima uses first available image)
	if len(metrics.UnusualImages) > 0 {
		breakdown.UnusualImages = weights.UnusualImage
	}

	// Comment lines (indicates documentation quality)
	breakdown.Comments = float64(metrics.CommentLineCount) * weights.CommentLine

	// Repository stars (capped to avoid dominating other factors)
	starsScore := float64(repoStars) / weights.StarsPerPoint
	if starsScore > weights.MaxStarsPoints {
		starsScore = weights.MaxStarsPoints
	}
	breakdown.Stars = starsScore

	// Calculate total
	breakdown.Total = breakdown.Message + breakdown.Provision + breakdown.Parameters +
		breakdown.EnvVars + breakdown.Probes + breakdown.UnusualImages +
		breakdown.Comments + breakdown.Stars

	return breakdown
}

// isCodeComment checks if a normalized comment line looks like commented-out code
// rather than actual documentation. Code comments typically contain:
// - Shell metacharacters: $ & @ [ ] { } | ; > < `
// - Assignment patterns: VAR=value
// - Common shell commands at the start
// - File paths: /path/to/something
func isCodeComment(line string) bool {
	// Skip empty lines
	if line == "" {
		return false
	}

	// Check for shell metacharacters (strong indicator of code)
	codeChars := []string{"$", "&", "@", "[", "]", "{", "}", "|", ";", ">", "<", "`", "\\"}
	for _, char := range codeChars {
		if strings.Contains(line, char) {
			return true
		}
	}

	// Check for assignment pattern (VAR=value)
	if strings.Contains(line, "=") && !strings.Contains(line, "==") {
		// Simple heuristic: if it has = but not ==, and no spaces before =, likely assignment
		parts := strings.Split(line, "=")
		if len(parts) >= 2 && !strings.Contains(parts[0], " ") {
			return true
		}
	}

	// Check for common shell commands at start of line
	commonCommands := []string{
		"apt", "yum", "dnf", "pacman", "brew", "apk",
		"curl", "wget", "git", "docker", "kubectl", "systemctl",
		"echo", "export", "cd", "mkdir", "chmod", "chown",
		"sudo", "source", ".", "ln", "cp", "mv", "rm",
	}
	firstWord := strings.Fields(line)
	if len(firstWord) > 0 {
		cmd := strings.ToLower(firstWord[0])
		for _, knownCmd := range commonCommands {
			if cmd == knownCmd {
				return true
			}
		}
	}

	// Check for file paths (starts with / or ~/)
	if strings.HasPrefix(line, "/") || strings.HasPrefix(line, "~/") {
		return true
	}

	return false
}

// FilterUniqueLines counts unique lines, excluding empty lines, known lines, and code-like comments
// For comment lines, also filters out commented-out code (e.g., "# apt-get install foo")
func FilterUniqueLines(lines []string, knownLines map[string]bool) int {
	uniqueCount := 0
	for _, line := range lines {
		// Skip empty lines
		if line == "" {
			continue
		}
		// Skip if this line exists in known lines
		if knownLines[line] {
			continue
		}
		uniqueCount++
	}
	return uniqueCount
}

// FilterUniqueComments counts unique comment lines, excluding empty lines, known lines, and code-like comments
func FilterUniqueComments(commentLines []string, knownLines map[string]bool) int {
	uniqueCount := 0
	for _, line := range commentLines {
		// Skip empty lines
		if line == "" {
			continue
		}
		// Skip if this line exists in known lines
		if knownLines[line] {
			continue
		}
		// Skip if this looks like commented-out code
		if isCodeComment(line) {
			continue
		}
		uniqueCount++
	}
	return uniqueCount
}

// PopulateNotabilityMetrics creates NotabilityMetrics from TemplateInfo
// Filters out known lines from official templates
func PopulateNotabilityMetrics(info *TemplateInfo, ok *OfficialKnowledge) *types.NotabilityMetrics {
	// Build lookup maps from official knowledge
	knownComments := make(map[string]bool)
	for _, line := range ok.KnownLines.Comments {
		knownComments[line] = true
	}

	knownProvision := make(map[string]bool)
	for _, line := range ok.KnownLines.Provision {
		knownProvision[line] = true
	}

	knownProbes := make(map[string]bool)
	for _, line := range ok.KnownLines.Probes {
		knownProbes[line] = true
	}

	knownMessages := make(map[string]bool)
	for _, line := range ok.KnownLines.Messages {
		knownMessages[line] = true
	}

	officialImages := make(map[string]bool)
	for _, domain := range ok.Images {
		officialImages[domain] = true
	}

	// Filter out known lines (and code-like comments for comment lines)
	uniqueCommentCount := FilterUniqueComments(info.CommentLines, knownComments) // Filters code comments too
	uniqueProvisionLines := FilterUniqueLines(info.ProvisionLines, knownProvision)
	uniqueProbeLines := FilterUniqueLines(info.ProbeLines, knownProbes)

	// Check if message contains any unique lines
	messageLength := 0
	if len(info.MessageLines) > 0 {
		uniqueMessageLines := FilterUniqueLines(info.MessageLines, knownMessages)
		if uniqueMessageLines > 0 {
			// If there are unique message lines, count total message length
			messageLength = info.MessageLength
		}
	}

	return &types.NotabilityMetrics{
		MessageLength:       messageLength,
		ProvisionCount:      info.ProvisionCount,
		ProvisionTotalLines: uniqueProvisionLines,
		ProbeCount:          info.ProbeCount,
		ProbeTotalLines:     uniqueProbeLines,
		ParamCount:          info.ParamCount,
		EnvCount:            info.EnvCount,
		CommentLineCount:    uniqueCommentCount,
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

// FetchDefaultTemplateComments fetches and extracts comment lines from lima-vm/lima default.yaml
func FetchDefaultTemplateComments(ctx context.Context, client *github.Client) (map[string]bool, error) {
	// Fetch the default.yaml template
	fileContent, _, _, err := client.Repositories.GetContents(
		ctx,
		"lima-vm",
		"lima",
		"templates/default.yaml",
		&github.RepositoryContentGetOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch default.yaml: %w", err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	// Extract comment lines
	commentMap := make(map[string]bool)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Store normalized comment line
			commentMap[trimmed] = true
		}
	}

	return commentMap, nil
}
