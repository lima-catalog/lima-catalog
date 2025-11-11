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
	client *github.Client
}

// NewDiscoverer creates a new template discoverer
func NewDiscoverer(client *github.Client) *Discoverer {
	return &Discoverer{
		client: client,
	}
}

// searchWithQuery performs a GitHub code search with pagination
func (d *Discoverer) searchWithQuery(query string) ([]types.Template, error) {
	var templates []types.Template

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
			template := types.Template{
				ID:           fmt.Sprintf("%s/%s", item.GetRepository().GetFullName(), item.GetPath()),
				Repo:         item.GetRepository().GetFullName(),
				Path:         item.GetPath(),
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

	return templates, nil
}

// DiscoverCommunityTemplates searches GitHub for community Lima templates
func (d *Discoverer) DiscoverCommunityTemplates() ([]types.Template, error) {
	// Use a map to deduplicate templates found by multiple queries
	templateMap := make(map[string]types.Template)

	// Query 1: Search for files with minimumLimaVersion (original query)
	// Use simpler syntax - search for .yaml files only (most common)
	query1 := "minimumLimaVersion extension:yaml -repo:lima-vm/lima"
	fmt.Printf("Query 1: %s\n", query1)
	templates1, err := d.searchWithQuery(query1)
	if err != nil {
		return nil, fmt.Errorf("query 1 failed: %w", err)
	}
	fmt.Printf("Query 1 found %d templates\n", len(templates1))

	for _, t := range templates1 {
		templateMap[t.ID] = t
	}

	// Wait before next query to avoid rate limits
	fmt.Println("Waiting 5 seconds before next query...")
	time.Sleep(5 * time.Second)

	// Also search for .yml extension
	query1b := "minimumLimaVersion extension:yml -repo:lima-vm/lima"
	fmt.Printf("\nQuery 1b: %s\n", query1b)
	templates1b, err := d.searchWithQuery(query1b)
	if err != nil {
		return nil, fmt.Errorf("query 1b failed: %w", err)
	}
	fmt.Printf("Query 1b found %d templates\n", len(templates1b))

	for _, t := range templates1b {
		if _, exists := templateMap[t.ID]; !exists {
			templateMap[t.ID] = t
		}
	}

	// Wait before next query to avoid rate limits
	fmt.Println("Waiting 5 seconds before next query...")
	time.Sleep(5 * time.Second)

	// Query 2: Search for files with images: and provision: fields (supplementary query)
	query2 := "images: provision: extension:yaml -repo:lima-vm/lima"
	fmt.Printf("\nQuery 2: %s\n", query2)
	templates2, err := d.searchWithQuery(query2)
	if err != nil {
		return nil, fmt.Errorf("query 2 failed: %w", err)
	}
	fmt.Printf("Query 2 found %d templates\n", len(templates2))

	for _, t := range templates2 {
		// Only add if not already found
		if _, exists := templateMap[t.ID]; !exists {
			templateMap[t.ID] = t
		}
	}

	// Wait before next query to avoid rate limits
	fmt.Println("Waiting 5 seconds before next query...")
	time.Sleep(5 * time.Second)

	// Also search for .yml extension
	query2b := "images: provision: extension:yml -repo:lima-vm/lima"
	fmt.Printf("\nQuery 2b: %s\n", query2b)
	templates2b, err := d.searchWithQuery(query2b)
	if err != nil {
		return nil, fmt.Errorf("query 2b failed: %w", err)
	}
	fmt.Printf("Query 2b found %d templates\n", len(templates2b))

	for _, t := range templates2b {
		// Only add if not already found
		if _, exists := templateMap[t.ID]; !exists {
			templateMap[t.ID] = t
		}
	}

	// Convert map to slice
	var templates []types.Template
	for _, t := range templateMap {
		templates = append(templates, t)
	}

	fmt.Printf("\nTotal unique templates after deduplication: %d\n", len(templates))

	return templates, nil
}

// DiscoverOfficialTemplates fetches templates from lima-vm/lima repository
func (d *Discoverer) DiscoverOfficialTemplates() ([]types.Template, error) {
	var templates []types.Template

	fmt.Println("Fetching official templates from lima-vm/lima...")

	// List contents of the templates directory
	contents, err := d.client.ListRepositoryContents("lima-vm", "lima", "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to list templates directory: %w", err)
	}

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

		template := types.Template{
			ID:           fmt.Sprintf("lima-vm/lima/%s", item.GetPath()),
			Repo:         "lima-vm/lima",
			Path:         item.GetPath(),
			SHA:          item.GetSHA(),
			Size:         item.GetSize(),
			URL:          item.GetHTMLURL(),
			DiscoveredAt: time.Now(),
			LastChecked:  time.Now(),
			IsOfficial:   true,
		}

		templates = append(templates, template)
	}

	fmt.Printf("Found %d official templates\n", len(templates))

	return templates, nil
}

// DiscoverAll discovers both community and official templates
func (d *Discoverer) DiscoverAll() ([]types.Template, error) {
	var allTemplates []types.Template

	// Discover community templates
	fmt.Println("=== Discovering Community Templates ===")
	communityTemplates, err := d.DiscoverCommunityTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to discover community templates: %w", err)
	}
	fmt.Printf("Discovered %d community templates\n\n", len(communityTemplates))
	allTemplates = append(allTemplates, communityTemplates...)

	// Discover official templates
	fmt.Println("=== Discovering Official Templates ===")
	officialTemplates, err := d.DiscoverOfficialTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to discover official templates: %w", err)
	}
	fmt.Printf("Discovered %d official templates\n\n", len(officialTemplates))
	allTemplates = append(allTemplates, officialTemplates...)

	return allTemplates, nil
}
