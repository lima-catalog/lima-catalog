package discovery

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/github"
	"github.com/lima-catalog/lima-catalog/pkg/types"
)

// Discoverer handles template discovery
type Discoverer struct {
	client    *github.Client
	blocklist *types.Blocklist
}

// NewDiscoverer creates a new template discoverer
func NewDiscoverer(client *github.Client, blocklist *types.Blocklist) *Discoverer {
	return &Discoverer{
		client:    client,
		blocklist: blocklist,
	}
}

// FindNewestTemplateTimestamp finds the newest DiscoveredAt timestamp from existing templates
// Returns zero time if no templates exist
func FindNewestTemplateTimestamp(templates []types.Template) time.Time {
	var newest time.Time
	for _, t := range templates {
		if t.DiscoveredAt.After(newest) {
			newest = t.DiscoveredAt
		}
	}
	return newest
}

// isLimaTemplate checks if a file is a valid Lima template by checking for the required "images:" top-level key
func (d *Discoverer) isLimaTemplate(owner, repo, path string) bool {
	content, err := d.client.GetRepositoryContent(owner, repo, path)
	if err != nil {
		return false // If we can't fetch it, exclude it
	}

	// Decode the content (it's base64 encoded)
	contentStr, err := content.GetContent()
	if err != nil {
		return false
	}

	// Check for "images:" as a top-level YAML key (at the start of a line)
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "images:") {
			return true
		}
	}

	return false
}

// searchWithQuery performs a GitHub code search with pagination
func (d *Discoverer) searchWithQuery(query string) ([]types.Template, error) {
	var templates []types.Template
	excludedCount := 0
	blocklistedCount := 0

	page := 1
	for {
		fmt.Printf("  Searching page %d...\n", page)

		result, resp, err := d.client.SearchCode(query, page)
		if err != nil {
			// Check if it's a rate limit error
			if resp != nil && (resp.StatusCode == 403 || resp.StatusCode == 429) {
				// Check rate limit to get reset time
				limits, _ := d.client.RateLimit()
				if limits != nil && limits.Search != nil {
					resetTime := limits.Search.Reset.Time
					waitDuration := time.Until(resetTime)
					if waitDuration > 0 {
						fmt.Printf("  Rate limit exceeded, waiting %v until reset at %s\n",
							waitDuration.Round(time.Second), resetTime.Format(time.RFC3339))
						time.Sleep(waitDuration + 5*time.Second) // Add 5s buffer
						fmt.Println("  Retrying after rate limit reset...")
						continue // Retry the same page
					}
				}
			}
			return nil, fmt.Errorf("code search failed: %w", err)
		}

		if len(result.CodeResults) == 0 {
			break
		}

		fmt.Printf("  Found %d results on page %d\n", len(result.CodeResults), page)

		for _, item := range result.CodeResults {
			repoFullName := item.GetRepository().GetFullName()
			path := item.GetPath()

			// Parse owner and repo name
			parts := strings.SplitN(repoFullName, "/", 2)
			if len(parts) != 2 {
				continue
			}
			owner, repo := parts[0], parts[1]

			// Check blocklist BEFORE fetching content (saves API calls)
			if IsBlocklisted(owner, repo, path, d.blocklist) {
				blocklistedCount++
				continue
			}

			// Check if this is actually a Lima template by verifying it has "images:" key
			if !d.isLimaTemplate(owner, repo, path) {
				excludedCount++
				continue
			}

			template := types.Template{
				ID:           fmt.Sprintf("%s/%s", repoFullName, path),
				Repo:         repoFullName,
				Path:         path,
				SHA:          item.GetSHA(),
				URL:          item.GetHTMLURL(),
				DiscoveredAt: time.Now(),
				LastChecked:  time.Now(),
				IsOfficial:   false,
			}

			templates = append(templates, template)
		}

		// Check if we've reached the last page
		if len(result.CodeResults) < 100 {
			break
		}

		page++

		// Add delay between pagination requests to avoid hitting search rate limits
		// Search API has a limit of 30 requests/minute
		time.Sleep(3 * time.Second) // 3 seconds = max 20 requests/minute
	}

	if blocklistedCount > 0 {
		fmt.Printf("  Blocklisted %d files (matched blocklist rules)\n", blocklistedCount)
	}
	if excludedCount > 0 {
		fmt.Printf("  Excluded %d files that don't have 'images:' top-level key\n", excludedCount)
	}

	return templates, nil
}

// DiscoverCommunityTemplates discovers community templates
// If sinceDate is provided (non-zero), only searches for templates pushed since that date
func (d *Discoverer) DiscoverCommunityTemplates(sinceDate time.Time) ([]types.Template, error) {
	// Use a map to deduplicate templates found by multiple queries
	templateMap := make(map[string]types.Template)

	// Build date qualifier if in incremental mode
	dateQualifier := ""
	if !sinceDate.IsZero() {
		dateQualifier = fmt.Sprintf(" pushed:>%s", sinceDate.Format("2006-01-02"))
	}

	// Query 1: Search for files with minimumLimaVersion (original query)
	// Use simpler syntax - search for .yaml files only (most common)
	query1 := "minimumLimaVersion extension:yaml -repo:lima-vm/lima" + dateQualifier
	fmt.Printf("Query 1: %s\n", query1)
	templates1, err := d.searchWithQuery(query1)
	if err != nil {
		return nil, fmt.Errorf("query 1 failed: %w", err)
	}
	fmt.Printf("Query 1 found %d templates:\n", len(templates1))
	for _, t := range templates1 {
		fmt.Printf("  - %s\n", t.ID)
		templateMap[t.ID] = t
	}

	// Wait before next query to avoid rate limits
	fmt.Println("Waiting 5 seconds before next query...")
	time.Sleep(5 * time.Second)

	// Also search for .yml extension
	query1b := "minimumLimaVersion extension:yml -repo:lima-vm/lima" + dateQualifier
	fmt.Printf("\nQuery 1b: %s\n", query1b)
	templates1b, err := d.searchWithQuery(query1b)
	if err != nil {
		return nil, fmt.Errorf("query 1b failed: %w", err)
	}
	fmt.Printf("Query 1b found %d templates:\n", len(templates1b))
	newFromQuery1b := 0
	for _, t := range templates1b {
		if _, exists := templateMap[t.ID]; !exists {
			fmt.Printf("  - %s (new)\n", t.ID)
			templateMap[t.ID] = t
			newFromQuery1b++
		}
	}
	fmt.Printf("Query 1b added %d new templates (duplicates: %d)\n", newFromQuery1b, len(templates1b)-newFromQuery1b)

	// Wait before next query to avoid rate limits
	fmt.Println("Waiting 5 seconds before next query...")
	time.Sleep(5 * time.Second)

	// Query 2: Search for files with images: and provision: fields (supplementary query)
	query2 := "images: provision: extension:yaml -repo:lima-vm/lima" + dateQualifier
	fmt.Printf("\nQuery 2: %s\n", query2)
	templates2, err := d.searchWithQuery(query2)
	if err != nil {
		return nil, fmt.Errorf("query 2 failed: %w", err)
	}
	fmt.Printf("Query 2 found %d templates:\n", len(templates2))
	newFromQuery2 := 0
	for _, t := range templates2 {
		// Only add if not already found
		if _, exists := templateMap[t.ID]; !exists {
			fmt.Printf("  - %s (new)\n", t.ID)
			templateMap[t.ID] = t
			newFromQuery2++
		}
	}
	fmt.Printf("Query 2 added %d new templates (duplicates: %d)\n", newFromQuery2, len(templates2)-newFromQuery2)

	// Wait before next query to avoid rate limits
	fmt.Println("Waiting 5 seconds before next query...")
	time.Sleep(5 * time.Second)

	// Also search for .yml extension
	query2b := "images: provision: extension:yml -repo:lima-vm/lima" + dateQualifier
	fmt.Printf("\nQuery 2b: %s\n", query2b)
	templates2b, err := d.searchWithQuery(query2b)
	if err != nil {
		return nil, fmt.Errorf("query 2b failed: %w", err)
	}
	fmt.Printf("Query 2b found %d templates:\n", len(templates2b))
	newFromQuery2b := 0
	for _, t := range templates2b {
		// Only add if not already found
		if _, exists := templateMap[t.ID]; !exists {
			fmt.Printf("  - %s (new)\n", t.ID)
			templateMap[t.ID] = t
			newFromQuery2b++
		}
	}
	fmt.Printf("Query 2b added %d new templates (duplicates: %d)\n", newFromQuery2b, len(templates2b)-newFromQuery2b)

	// Convert map to slice
	var templates []types.Template
	for _, t := range templateMap {
		templates = append(templates, t)
	}

	fmt.Printf("\nTotal unique templates after deduplication: %d\n", len(templates))

	return templates, nil
}

// DiscoverOfficialTemplates fetches templates from lima-vm/lima repository
// DiscoverOfficialTemplates discovers official templates from lima-vm/lima
// If sinceDate is provided and existingTemplates is not empty, only returns templates that are new or changed
func (d *Discoverer) DiscoverOfficialTemplates(sinceDate time.Time, existingTemplates []types.Template) ([]types.Template, error) {
	var templates []types.Template

	// If incremental mode, check if lima-vm/lima repo was updated since sinceDate
	if !sinceDate.IsZero() && len(existingTemplates) > 0 {
		fmt.Println("Checking if lima-vm/lima repository was updated...")

		// Get lima-vm/lima repo info to check last push time
		repo, err := d.client.GetRepository("lima-vm", "lima")
		if err != nil {
			return nil, fmt.Errorf("failed to get lima-vm/lima repository info: %w", err)
		}

		lastPush := repo.GetPushedAt().Time
		if !lastPush.After(sinceDate) {
			fmt.Printf("lima-vm/lima not updated since %s (last push: %s)\n",
				sinceDate.Format("2006-01-02"), lastPush.Format("2006-01-02"))
			fmt.Println("Skipping official template enumeration (no changes)")
			return templates, nil
		}

		fmt.Printf("lima-vm/lima was updated at %s (checking for template changes)\n", lastPush.Format("2006-01-02"))
	} else {
		fmt.Println("Fetching official templates from lima-vm/lima...")
	}

	// List contents of the templates directory
	contents, err := d.client.ListRepositoryContents("lima-vm", "lima", "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to list templates directory: %w", err)
	}

	// Create map of existing templates for SHA comparison
	existingMap := make(map[string]types.Template)
	for _, t := range existingTemplates {
		if t.IsOfficial {
			existingMap[t.ID] = t
		}
	}

	newCount := 0
	changedCount := 0
	unchangedCount := 0

	for _, item := range contents {
		// Only include YAML files
		ext := filepath.Ext(item.GetName())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		// Skip directories and files starting with underscore (internal templates)
		if item.GetType() != "file" || strings.HasPrefix(item.GetName(), "_") {
			continue
		}

		templateID := fmt.Sprintf("lima-vm/lima/%s", item.GetPath())

		template := types.Template{
			ID:           templateID,
			Repo:         "lima-vm/lima",
			Path:         item.GetPath(),
			SHA:          item.GetSHA(),
			Size:         item.GetSize(),
			URL:          item.GetHTMLURL(),
			DiscoveredAt: time.Now(),
			LastChecked:  time.Now(),
			IsOfficial:   true,
		}

		// In incremental mode, only include new or changed templates
		if !sinceDate.IsZero() && len(existingTemplates) > 0 {
			if existing, found := existingMap[templateID]; found {
				if existing.SHA == template.SHA {
					// Unchanged - skip
					unchangedCount++
					continue
				} else {
					// Changed SHA - include it
					template.DiscoveredAt = existing.DiscoveredAt // Preserve original discovery time
					changedCount++
				}
			} else {
				// New template
				newCount++
			}
		}

		templates = append(templates, template)
	}

	// Print summary based on mode
	if !sinceDate.IsZero() && len(existingTemplates) > 0 {
		fmt.Printf("Found %d new, %d changed, %d unchanged official templates\n",
			newCount, changedCount, unchangedCount)
	} else {
		fmt.Printf("Found %d official templates\n", len(templates))
	}

	return templates, nil
}

// DiscoverAll discovers all templates (community + official)
// If sinceDate is provided (non-zero), only discovers templates pushed since that date
// If existingTemplates is provided, uses incremental mode (only returns new/changed templates)
func (d *Discoverer) DiscoverAll(sinceDate time.Time, existingTemplates []types.Template) ([]types.Template, error) {
	var allTemplates []types.Template

	// Discover community templates
	fmt.Println("=== Discovering Community Templates ===")
	if !sinceDate.IsZero() {
		fmt.Printf("Incremental mode: searching for templates pushed since %s\n", sinceDate.Format("2006-01-02"))
	}
	communityTemplates, err := d.DiscoverCommunityTemplates(sinceDate)
	if err != nil {
		return nil, fmt.Errorf("failed to discover community templates: %w", err)
	}
	fmt.Printf("Discovered %d community templates\n\n", len(communityTemplates))
	allTemplates = append(allTemplates, communityTemplates...)

	// Discover official templates (with incremental mode if sinceDate provided)
	fmt.Println("=== Discovering Official Templates ===")
	officialTemplates, err := d.DiscoverOfficialTemplates(sinceDate, existingTemplates)
	if err != nil {
		return nil, fmt.Errorf("failed to discover official templates: %w", err)
	}
	fmt.Printf("Discovered %d official templates\n\n", len(officialTemplates))
	allTemplates = append(allTemplates, officialTemplates...)

	return allTemplates, nil
}
