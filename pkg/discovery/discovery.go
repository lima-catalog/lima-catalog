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

// DiscoverCommunityTemplates searches GitHub for community Lima templates
func (d *Discoverer) DiscoverCommunityTemplates() ([]types.Template, error) {
	var templates []types.Template

	// Search for YAML files containing minimumLimaVersion, excluding lima-vm/lima
	query := "minimumLimaVersion extension:yml OR extension:yaml -repo:lima-vm/lima"

	page := 1
	for {
		fmt.Printf("Searching page %d...\n", page)

		result, _, err := d.client.SearchCode(query, page)
		if err != nil {
			return nil, fmt.Errorf("code search failed: %w", err)
		}

		if len(result.CodeResults) == 0 {
			break
		}

		fmt.Printf("Found %d results on page %d\n", len(result.CodeResults), page)

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
		time.Sleep(2 * time.Second) // Respect rate limits
	}

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
