package discovery

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// OfficialKnowledge represents known content from lima-vm/lima templates
type OfficialKnowledge struct {
	LastUpdate time.Time         `json:"lastUpdate"`
	KnownLines OfficialKnownLines `json:"knownLines"`
	Images     []string          `json:"images"` // Top-level domains only
}

// OfficialKnownLines contains normalized lines from official templates
type OfficialKnownLines struct {
	Comments  []string `json:"comments"`
	Provision []string `json:"provision"`
	Probes    []string `json:"probes"`
	Messages  []string `json:"messages"`
}

// LoadOfficialKnowledge loads official knowledge from JSON file
func LoadOfficialKnowledge(path string) (*OfficialKnowledge, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty knowledge if file doesn't exist
			return &OfficialKnowledge{
				LastUpdate: time.Time{}, // Zero time
				KnownLines: OfficialKnownLines{
					Comments:  []string{},
					Provision: []string{},
					Probes:    []string{},
					Messages:  []string{},
				},
				Images: []string{},
			}, nil
		}
		return nil, fmt.Errorf("failed to open official knowledge file: %w", err)
	}
	defer file.Close()

	var ok OfficialKnowledge
	if err := json.NewDecoder(file).Decode(&ok); err != nil {
		return nil, fmt.Errorf("failed to decode official knowledge: %w", err)
	}

	return &ok, nil
}

// SaveOfficialKnowledge saves official knowledge to JSON file
func SaveOfficialKnowledge(path string, ok *OfficialKnowledge) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create official knowledge file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(ok); err != nil {
		return fmt.Errorf("failed to encode official knowledge: %w", err)
	}

	return nil
}

// UpdateOfficialKnowledge scans lima repo history and updates official knowledge
func UpdateOfficialKnowledge(ctx context.Context, repoPath, outputPath string) (*OfficialKnowledge, error) {
	// Load existing knowledge
	ok, err := LoadOfficialKnowledge(outputPath)
	if err != nil {
		return nil, err
	}

	// Verify repo path exists
	if _, err := os.Stat(repoPath); err != nil {
		return nil, fmt.Errorf("lima repo path does not exist: %w", err)
	}

	// Save original HEAD to restore later
	originalHead, err := gitCurrentCommit(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get current HEAD: %w", err)
	}
	defer func() {
		// Always restore original HEAD
		_ = gitCheckout(ctx, repoPath, originalHead)
	}()

	// Process both directories (handles cases where directory might not exist in all commits)
	for _, dir := range []string{"templates", "examples"} {
		if err := processDirectory(ctx, repoPath, dir, ok); err != nil {
			// Log error but continue with other directory
			// The directory might not exist in all commits (e.g., examples was renamed to templates)
			fmt.Fprintf(os.Stderr, "Warning: failed to process %s directory: %v\n", dir, err)
			continue
		}
	}

	// Update timestamp
	ok.LastUpdate = time.Now().UTC()

	// Save updated knowledge
	if err := SaveOfficialKnowledge(outputPath, ok); err != nil {
		return nil, err
	}

	return ok, nil
}

// processDirectory processes a single directory's git history
func processDirectory(ctx context.Context, repoPath, dir string, ok *OfficialKnowledge) error {
	// Get commits that touched this directory (chronological order)
	commits, err := gitLogDirectory(ctx, repoPath, dir)
	if err != nil {
		return err
	}

	// Filter commits newer than lastUpdate
	var newCommits []commitInfo
	for _, c := range commits {
		if c.Timestamp.After(ok.LastUpdate) {
			newCommits = append(newCommits, c)
		}
	}

	if len(newCommits) == 0 {
		return nil // No new commits to process
	}

	// Process each commit
	for _, commit := range newCommits {
		if err := ctx.Err(); err != nil {
			return err // Context cancelled
		}

		// Checkout this commit
		if err := gitCheckout(ctx, repoPath, commit.Hash); err != nil {
			return fmt.Errorf("failed to checkout %s: %w", commit.Hash, err)
		}

		// Scan all YAML files in directory (recursive)
		dirPath := filepath.Join(repoPath, dir)

		// Check if directory exists at this commit (it might not in older/newer commits)
		if _, err := os.Stat(dirPath); err != nil {
			if os.IsNotExist(err) {
				continue // Directory doesn't exist at this commit, skip
			}
			// Other stat errors - log but continue to next commit
			fmt.Fprintf(os.Stderr, "Warning: stat failed for %s at %s: %v\n", dirPath, commit.Hash, err)
			continue
		}

		if err := scanYAMLFiles(dirPath, ok); err != nil {
			return fmt.Errorf("failed to scan YAML files at %s: %w", commit.Hash, err)
		}
	}

	return nil
}

// commitInfo represents a git commit
type commitInfo struct {
	Hash      string
	Timestamp time.Time
}

// gitLogDirectory gets commits that modified a directory
func gitLogDirectory(ctx context.Context, repoPath, dir string) ([]commitInfo, error) {
	cmd := exec.CommandContext(ctx, "git", "log", "--format=%H:%ct", "--reverse", "--", dir)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	var commits []commitInfo
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		hash := parts[0]
		timestamp := parts[1]

		// Parse Unix timestamp
		var ts int64
		if _, err := fmt.Sscanf(timestamp, "%d", &ts); err != nil {
			continue
		}

		commits = append(commits, commitInfo{
			Hash:      hash,
			Timestamp: time.Unix(ts, 0).UTC(),
		})
	}

	return commits, nil
}

// gitCurrentCommit gets current HEAD commit hash
func gitCurrentCommit(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// gitCheckout checks out a specific commit
func gitCheckout(ctx context.Context, repoPath, ref string) error {
	cmd := exec.CommandContext(ctx, "git", "checkout", ref)
	cmd.Dir = repoPath
	cmd.Stdout = nil // Suppress output
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout %s failed: %w", ref, err)
	}

	return nil
}

// scanYAMLFiles recursively scans YAML files and extracts knowledge
func scanYAMLFiles(dirPath string, ok *OfficialKnowledge) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Directory might have been deleted/renamed by git - just skip
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .yaml files
		if !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		// Extract knowledge from this file
		if err := extractKnowledgeFromFile(path, ok); err != nil {
			// Log error but continue with other files
			fmt.Fprintf(os.Stderr, "Warning: failed to extract from %s: %v\n", path, err)
		}

		return nil
	})
}

// extractKnowledgeFromFile extracts lines and images from a YAML file
func extractKnowledgeFromFile(path string, knowledge *OfficialKnowledge) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// Parse YAML to extract structured data
	var doc map[string]interface{}
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return err
	}

	// Extract images
	if images, ok := doc["images"]; ok {
		extractImages(images, knowledge)
	}

	// Extract provision scripts
	if provision, ok := doc["provision"]; ok {
		extractProvisionLines(provision, knowledge)
	}

	// Extract probes
	if probes, ok := doc["probes"]; ok {
		extractProbeLines(probes, knowledge)
	}

	// Extract message
	if message, ok := doc["message"]; ok {
		if msg, ok := message.(string); ok {
			extractMessageLines(msg, knowledge)
		}
	}

	// Extract comments from raw content
	extractCommentLines(string(content), knowledge)

	return nil
}

// extractImages extracts image domains from images field
func extractImages(images interface{}, knowledge *OfficialKnowledge) {
	imageList, ok := images.([]interface{})
	if !ok {
		return
	}

	for _, img := range imageList {
		imgMap, ok := img.(map[string]interface{})
		if !ok {
			continue
		}

		// Get location field
		location, ok := imgMap["location"].(string)
		if !ok {
			continue
		}

		// Extract domain from URL
		domain := extractTopLevelDomain(location)
		if domain != "" {
			addUnique(&knowledge.Images, domain)
		}
	}
}

// extractTopLevelDomain extracts registrable domain (e.g., ubuntu.com from downloads.ubuntu.com)
func extractTopLevelDomain(urlStr string) string {
	// Remove protocol
	urlStr = strings.TrimPrefix(urlStr, "https://")
	urlStr = strings.TrimPrefix(urlStr, "http://")

	// Get host part (before first /)
	parts := strings.Split(urlStr, "/")
	if len(parts) == 0 {
		return ""
	}
	host := parts[0]

	// Remove port if present
	host = strings.Split(host, ":")[0]

	// Extract last 2 parts (simple heuristic for registrable domain)
	domainParts := strings.Split(host, ".")
	if len(domainParts) < 2 {
		return host // Already a simple domain
	}

	// Return last 2 parts
	return strings.Join(domainParts[len(domainParts)-2:], ".")
}

// extractProvisionLines extracts lines from provision scripts
func extractProvisionLines(provision interface{}, knowledge *OfficialKnowledge) {
	provisionList, ok := provision.([]interface{})
	if !ok {
		return
	}

	for _, p := range provisionList {
		pMap, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		// Get script field
		script, ok := pMap["script"].(string)
		if !ok {
			continue
		}

		// Extract lines (ignore comments)
		lines := extractNonCommentLines(script)
		for _, line := range lines {
			addUnique(&knowledge.KnownLines.Provision, line)
		}
	}
}

// extractProbeLines extracts lines from probe scripts
func extractProbeLines(probes interface{}, knowledge *OfficialKnowledge) {
	probeList, ok := probes.([]interface{})
	if !ok {
		return
	}

	for _, p := range probeList {
		pMap, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		// Get script field
		script, ok := pMap["script"].(string)
		if !ok {
			continue
		}

		// Extract lines (ignore comments)
		lines := extractNonCommentLines(script)
		for _, line := range lines {
			addUnique(&knowledge.KnownLines.Probes, line)
		}
	}
}

// extractMessageLines extracts lines from message field
func extractMessageLines(message string, knowledge *OfficialKnowledge) {
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		normalized := normalizeLine(line)
		if normalized != "" {
			addUnique(&knowledge.KnownLines.Messages, normalized)
		}
	}
}

// extractCommentLines extracts comment lines from raw YAML content
func extractCommentLines(content string, knowledge *OfficialKnowledge) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()

		// Trim leading whitespace
		trimmed := strings.TrimLeft(line, " \t")

		// Check if it's a comment
		if strings.HasPrefix(trimmed, "#") {
			normalized := normalizeCommentLine(line)
			if normalized != "" {
				addUnique(&knowledge.KnownLines.Comments, normalized)
			}
		}
	}
}

// extractNonCommentLines extracts non-comment lines from script content
func extractNonCommentLines(script string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(script))
	for scanner.Scan() {
		line := scanner.Text()

		// Trim leading whitespace
		trimmed := strings.TrimLeft(line, " \t")

		// Skip comment lines (they're counted as YAML comments)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Normalize and add
		normalized := normalizeLine(line)
		if normalized != "" {
			lines = append(lines, normalized)
		}
	}
	return lines
}

// normalizeCommentLine normalizes a comment line (strip leading whitespace and #, trailing whitespace)
func normalizeCommentLine(line string) string {
	// Strip ALL leading whitespace and # characters (loop until nothing left to strip)
	for {
		before := line
		line = strings.TrimLeft(line, " \t#")
		if line == before {
			break // Nothing more to strip
		}
	}

	// Trim trailing whitespace
	line = strings.TrimRight(line, " \t")

	return line
}

// normalizeLine normalizes a regular line (strip leading and trailing whitespace)
func normalizeLine(line string) string {
	// Trim leading and trailing whitespace
	return strings.TrimSpace(line)
}

// addUnique adds a value to a slice if it's not already present
func addUnique(slice *[]string, value string) {
	for _, v := range *slice {
		if v == value {
			return // Already present
		}
	}
	*slice = append(*slice, value)
}

// SortOfficialKnowledge sorts all slices for consistent output
func SortOfficialKnowledge(ok *OfficialKnowledge) {
	sort.Strings(ok.KnownLines.Comments)
	sort.Strings(ok.KnownLines.Provision)
	sort.Strings(ok.KnownLines.Probes)
	sort.Strings(ok.KnownLines.Messages)
	sort.Strings(ok.Images)
}
