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

// filterRemoteImages returns only image URLs that start with http:// or https://
// Local references and template expressions are filtered out
func filterRemoteImages(images []string) []string {
	var remote []string
	for _, img := range images {
		if strings.HasPrefix(img, "http://") || strings.HasPrefix(img, "https://") {
			remote = append(remote, img)
		}
	}
	return remote
}

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
// Only considers URLs starting with http:// or https://, filtering out local references
func IdentifyUnusualImages(images []string, officialDomains map[string]bool) []string {
	// Filter to only remote images (http:// or https://)
	remoteImages := filterRemoteImages(images)

	seenDomains := make(map[string]bool)
	var unusual []string

	for _, img := range remoteImages {
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

// CalculateCustomImagesScore checks if images match the org or repo name
// Only considers URLs starting with http:// or https://, filtering out local references
// Scores:
// - 25 points for one word boundary match (\bNAME or NAME\b)
// - 35 points for both word boundaries (\bNAME\b)
// Returns the sum of highest org match + highest repo match (0-70 points)
func CalculateCustomImagesScore(images []string, orgName string, repoName string) float64 {
	// Filter to only remote images (http:// or https://)
	remoteImages := filterRemoteImages(images)

	if len(remoteImages) == 0 {
		return 0
	}

	// Helper function to check word boundary matches
	checkMatch := func(str, name string) int {
		if name == "" {
			return 0
		}

		// Convert to lowercase for case-insensitive matching
		strLower := strings.ToLower(str)
		nameLower := strings.ToLower(name)

		// Check if name appears in the string
		if !strings.Contains(strLower, nameLower) {
			return 0
		}

		// Find all occurrences and check word boundaries
		bestMatch := 0
		idx := 0
		for {
			pos := strings.Index(strLower[idx:], nameLower)
			if pos == -1 {
				break
			}
			pos += idx

			// Check word boundaries
			leftBoundary := pos == 0 || !isAlphanumericOrUnderscore(strLower[pos-1])
			rightBoundary := pos+len(nameLower) >= len(strLower) || !isAlphanumericOrUnderscore(strLower[pos+len(nameLower)])

			var score int
			if leftBoundary && rightBoundary {
				score = 35 // Both boundaries
			} else if leftBoundary || rightBoundary {
				score = 25 // One boundary
			}

			if score > bestMatch {
				bestMatch = score
			}

			idx = pos + 1
		}

		return bestMatch
	}

	// Check all images for org and repo matches
	highestOrgMatch := 0
	highestRepoMatch := 0

	for _, img := range remoteImages {
		orgMatch := checkMatch(img, orgName)
		if orgMatch > highestOrgMatch {
			highestOrgMatch = orgMatch
		}

		repoMatch := checkMatch(img, repoName)
		if repoMatch > highestRepoMatch {
			highestRepoMatch = repoMatch
		}
	}

	return float64(highestOrgMatch + highestRepoMatch)
}

// isAlphanumericOrUnderscore checks if a byte is alphanumeric or underscore
func isAlphanumericOrUnderscore(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

// NotabilityScoreBreakdown contains the individual components of the notability score
type NotabilityScoreBreakdown struct {
	Message    float64 `json:"message"`     // 100 if has message
	Provision  float64 `json:"provision"`   // 10 per script + 1 per 10 lines
	Parameters float64 `json:"parameters"`  // 20 per param
	EnvVars    float64 `json:"env_vars"`    // 10 per var
	Probes     float64 `json:"probes"`      // 5 per probe + 1 per 10 lines
	ImageName  float64 `json:"image_name"`  // -100 if no remote images, 30 if unusual, +0-70 if custom names
	Comments   float64 `json:"comments"`    // 2 per comment line
	Stars      float64 `json:"stars"`       // 1 per 10 stars (capped at 50)
	Total      float64 `json:"total"`       // Sum of all components
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
// - Image name: -100 if no remote images, 30 if unusual images, +0-70 if custom org/repo names
// - Comment lines: 2 points per comment line (indicates documentation quality)
// - Repository stars: 1 point per 10 stars (capped at 50 points)
func CalculateNotabilityScore(metrics *types.NotabilityMetrics, orgName, repoName string, repoStars int) float64 {
	breakdown := CalculateNotabilityScoreWithBreakdown(metrics, orgName, repoName, repoStars)
	return breakdown.Total
}

// CalculateNotabilityScoreWithBreakdown computes the score and returns the breakdown
func CalculateNotabilityScoreWithBreakdown(metrics *types.NotabilityMetrics, orgName, repoName string, repoStars int) NotabilityScoreBreakdown {
	breakdown := NotabilityScoreBreakdown{}

	if metrics == nil {
		return breakdown
	}

	// Get default weights (can be made configurable in the future)
	weights := config.DefaultNotabilityWeights()

	// Message presence (strong signal for reusability)
	// Base bonus for having a message, plus line-based bonus for length
	if metrics.MessageLength > 0 {
		breakdown.Message = weights.Message
		// Add 1 point per line for better filtering granularity
		// (allows sorting by message quality, not just presence)
		breakdown.Message += float64(metrics.MessageLineCount)
	}

	// Provision scripts (indicates customization/setup)
	// Use substantial script count (>10 lines, capped at 3, min 1) to avoid
	// rewarding templates that split scripts into many tiny pieces
	breakdown.Provision = float64(metrics.ProvisionSubstantial)*weights.ProvisionBase +
		float64(metrics.ProvisionTotalLines)*weights.ProvisionLine

	// Parameters (indicates configurability)
	breakdown.Parameters = float64(metrics.ParamCount) * weights.Parameter

	// Environment variables (shows configuration effort)
	breakdown.EnvVars = float64(metrics.EnvCount) * weights.EnvVar

	// Probes (less important than provision)
	// Use substantial script count (>10 lines, capped at 3, min 1) to avoid
	// rewarding templates that split scripts into many tiny pieces
	breakdown.Probes = float64(metrics.ProbeSubstantial)*weights.ProbeBase +
		float64(metrics.ProbeTotalLines)*weights.ProbeLine

	// Image name scoring (combined unusual + custom images)
	// Logic:
	// - If no remote images (http/https): -100 penalty
	// - If has unusual images: 30 base + optional custom name bonus (0-70)
	// - If has only usual/official images: 0 (neutral)
	remoteImages := filterRemoteImages(metrics.AllImages)
	if len(remoteImages) == 0 {
		// No remote images - won't work on other computers
		breakdown.ImageName = -100
	} else if len(metrics.UnusualImages) > 0 {
		// Has unusual images - award base bonus + custom name bonus
		breakdown.ImageName = weights.UnusualImage
		breakdown.ImageName += CalculateCustomImagesScore(metrics.AllImages, orgName, repoName)
	}
	// else: has only usual/official images, ImageName stays 0 (neutral)

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
		breakdown.EnvVars + breakdown.Probes + breakdown.ImageName +
		breakdown.Comments + breakdown.Stars

	return breakdown
}

// IsCodeComment checks if a normalized comment line looks like commented-out code
// rather than actual documentation. Uses conservative heuristics to avoid false positives.
// Strong indicators: shell variables ($VAR), pipes (|), redirects (>, <), command chaining (&&, ||)
func IsCodeComment(line string) bool {
	// Skip empty lines
	if line == "" {
		return false
	}

	// Check for shell keywords at start of line
	lowerLine := strings.ToLower(line)
	shellKeywords := []string{"export ", "source ", "alias ", "unset ", "cd ", "mkdir ", "chmod ", "chown ", "ln ", "cp ", "mv ", "rm "}
	for _, keyword := range shellKeywords {
		if strings.HasPrefix(lowerLine, keyword) {
			return true
		}
	}

	// Check for shell variable expansion (strong indicator)
	// $VAR, ${VAR}, $(...), or backticks
	if strings.Contains(line, "$") {
		// Make sure it's not just a price ($5) by checking for variable patterns
		if strings.Contains(line, "${") || strings.Contains(line, "$(") {
			return true
		}
		// Check for $WORD pattern (not just $ followed by number)
		for i := strings.Index(line, "$"); i >= 0; i = strings.Index(line[i+1:], "$") {
			if i+1 < len(line) && (line[i+1] >= 'A' && line[i+1] <= 'Z' || line[i+1] >= 'a' && line[i+1] <= 'z' || line[i+1] == '_') {
				return true
			}
			if i+1 >= len(line) {
				break
			}
		}
	}

	// Check for backticks (command substitution)
	if strings.Contains(line, "`") {
		return true
	}

	// Check for pipe (likely command chaining, not markdown table)
	// Markdown tables have multiple pipes in a pattern like "| cell | cell |"
	pipeCount := strings.Count(line, "|")
	if pipeCount == 1 || (pipeCount > 1 && !strings.Contains(line, "| ")) {
		return true
	}

	// Check for command chaining operators
	if strings.Contains(line, "&&") || strings.Contains(line, "||") {
		return true
	}

	// Check for shell redirects (strong indicator)
	if strings.Contains(line, " > ") || strings.Contains(line, " >> ") ||
		strings.Contains(line, " < ") || strings.Contains(line, " 2>&1") ||
		strings.HasSuffix(line, ">") || strings.HasSuffix(line, ">>") {
		return true
	}

	// Check for assignment pattern (VAR=value without spaces, but not ==)
	if strings.Contains(line, "=") && !strings.Contains(line, "==") && !strings.Contains(line, "!=") {
		// Look for pattern: WORD=value with no spaces around =
		// But avoid URLs (http://) and explanatory text (key = value)
		if !strings.Contains(line, "://") && !strings.Contains(line, " = ") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				before := strings.TrimSpace(parts[0])
				// Check if before looks like a variable name (no spaces, starts with letter/underscore)
				if len(before) > 0 && !strings.Contains(before, " ") {
					if (before[0] >= 'A' && before[0] <= 'Z') ||
						(before[0] >= 'a' && before[0] <= 'z') ||
						before[0] == '_' {
						return true
					}
				}
			}
		}
	}

	// Check if line starts with absolute file path
	if strings.HasPrefix(line, "/etc/") || strings.HasPrefix(line, "/usr/") ||
		strings.HasPrefix(line, "/var/") || strings.HasPrefix(line, "/opt/") ||
		strings.HasPrefix(line, "/tmp/") || strings.HasPrefix(line, "/home/") ||
		strings.HasPrefix(line, "~/") {
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
		if IsCodeComment(line) {
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
	messageLineCount := 0
	if len(info.MessageLines) > 0 {
		uniqueMessageLines := FilterUniqueLines(info.MessageLines, knownMessages)
		if uniqueMessageLines > 0 {
			// If there are unique message lines, count total message length and line count
			messageLength = info.MessageLength
			messageLineCount = len(info.MessageLines)
		}
	}

	// Count substantial scripts (>10 lines) for provision and probes
	// Cap at 3, minimum 1 if any scripts exist
	provisionSubstantial := countSubstantialScripts(info.ProvisionScriptLines, info.ProvisionCount)
	probeSubstantial := countSubstantialScripts(info.ProbeScriptLines, info.ProbeCount)

	return &types.NotabilityMetrics{
		MessageLength:        messageLength,
		MessageLineCount:     messageLineCount,
		ProvisionCount:       info.ProvisionCount,
		ProvisionSubstantial: provisionSubstantial,
		ProvisionTotalLines:  uniqueProvisionLines,
		ProbeCount:           info.ProbeCount,
		ProbeSubstantial:     probeSubstantial,
		ProbeTotalLines:      uniqueProbeLines,
		ParamCount:           info.ParamCount,
		EnvCount:             info.EnvCount,
		CommentLineCount:     uniqueCommentCount,
		UnusualImages:        IdentifyUnusualImages(info.Images, officialImages),
		AllImages:            info.Images,
	}
}

// countSubstantialScripts counts scripts with >10 lines, capped at 3, minimum 1 if any scripts exist
func countSubstantialScripts(scriptLineCounts []int, totalScripts int) int {
	if totalScripts == 0 {
		return 0
	}

	// Count scripts with >10 lines
	substantial := 0
	for _, lineCount := range scriptLineCounts {
		if lineCount > 10 {
			substantial++
		}
	}

	// Cap at 3
	if substantial > 3 {
		substantial = 3
	}

	// Minimum 1 if any scripts exist
	if substantial == 0 && totalScripts > 0 {
		substantial = 1
	}

	return substantial
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
