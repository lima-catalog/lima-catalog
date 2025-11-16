package discovery

import (
	"fmt"
	"sort"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

// UpdateResult contains statistics about an incremental update
type UpdateResult struct {
	AllTemplates     []types.Template // All templates to save (new + updated + unchanged)
	NewTemplates     []types.Template // Newly discovered templates
	UpdatedTemplates []types.Template // Templates with changed SHAs
	UnchangedCount   int              // Count of templates without changes
	RemovedTemplates []string         // Template IDs that were removed
}

// MergeTemplates performs an incremental update by merging existing templates with newly discovered ones
func MergeTemplates(existing, discovered []types.Template) UpdateResult {
	result := UpdateResult{
		AllTemplates:     []types.Template{},
		NewTemplates:     []types.Template{},
		UpdatedTemplates: []types.Template{},
		RemovedTemplates: []string{},
	}

	// Create maps for quick lookup
	existingMap := make(map[string]types.Template)
	for _, t := range existing {
		existingMap[t.ID] = t
	}

	discoveredMap := make(map[string]types.Template)
	for _, t := range discovered {
		discoveredMap[t.ID] = t
	}

	// In incremental mode, templates not in discoveredMap are UNCHANGED (not removed)
	// We preserve all existing templates by default, then update with discovered changes
	preservedTemplates := make(map[string]bool)

	// Find new and updated templates
	for id, newTemplate := range discoveredMap {
		if oldTemplate, exists := existingMap[id]; exists {
			// Check if template changed (different SHA)
			if oldTemplate.SHA != newTemplate.SHA {
				// Template was updated
				newTemplate.DiscoveredAt = oldTemplate.DiscoveredAt // Preserve original discovery time
				newTemplate.LastChecked = time.Now()
				result.UpdatedTemplates = append(result.UpdatedTemplates, newTemplate)
				result.AllTemplates = append(result.AllTemplates, newTemplate)
			} else {
				// Template unchanged but re-discovered, update last checked time
				oldTemplate.LastChecked = time.Now()
				result.UnchangedCount++
				result.AllTemplates = append(result.AllTemplates, oldTemplate)
			}
			preservedTemplates[id] = true
		} else {
			// New template
			result.NewTemplates = append(result.NewTemplates, newTemplate)
			result.AllTemplates = append(result.AllTemplates, newTemplate)
			preservedTemplates[id] = true
		}
	}

	// Preserve existing templates that weren't in discoveredMap
	// In incremental mode, absence from discoveredMap means "not checked", not "removed"
	// (Template deletion detection is Stage 7, not implemented yet)
	for id, oldTemplate := range existingMap {
		if !preservedTemplates[id] {
			// Template wasn't checked this run - preserve it unchanged
			oldTemplate.LastChecked = time.Now() // Update last checked time
			result.UnchangedCount++
			result.AllTemplates = append(result.AllTemplates, oldTemplate)
		}
	}

	// Note: RemovedTemplates is currently unused
	// Template deletion detection will be implemented in Stage 7

	return result
}

// MergeRepositories merges existing repository data with newly collected data
func MergeRepositories(existing, collected []types.Repository) []types.Repository {
	repoMap := make(map[string]types.Repository)

	// Start with existing data
	for _, r := range existing {
		repoMap[r.ID] = r
	}

	// Update/add with newly collected data
	for _, r := range collected {
		repoMap[r.ID] = r
	}

	// Convert back to slice
	result := make([]types.Repository, 0, len(repoMap))
	for _, r := range repoMap {
		result = append(result, r)
	}

	// Sort by owner (org), then name for stable output
	sort.Slice(result, func(i, j int) bool {
		if result[i].Owner != result[j].Owner {
			return result[i].Owner < result[j].Owner
		}
		return result[i].Name < result[j].Name
	})

	return result
}

// MergeOrganizations merges existing organization data with newly collected data
func MergeOrganizations(existing, collected []types.Organization) []types.Organization {
	orgMap := make(map[string]types.Organization)

	// Start with existing data
	for _, o := range existing {
		orgMap[o.ID] = o
	}

	// Update/add with newly collected data
	for _, o := range collected {
		orgMap[o.ID] = o
	}

	// Convert back to slice
	result := make([]types.Organization, 0, len(orgMap))
	for _, o := range orgMap {
		result = append(result, o)
	}

	// Sort by ID (login) for stable output
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}

// PrintUpdateSummary prints a summary of the update result
func PrintUpdateSummary(result UpdateResult) {
	fmt.Println()
	fmt.Println("=== Update Summary ===")
	fmt.Printf("New templates discovered: %d\n", len(result.NewTemplates))
	fmt.Printf("Templates updated: %d\n", len(result.UpdatedTemplates))
	fmt.Printf("Templates unchanged: %d\n", result.UnchangedCount)
	fmt.Printf("Templates removed: %d\n", len(result.RemovedTemplates))
	fmt.Println()

	if len(result.NewTemplates) > 0 {
		fmt.Println("New templates:")
		for _, t := range result.NewTemplates {
			fmt.Printf("  + %s\n", t.ID)
		}
		fmt.Println()
	}

	if len(result.UpdatedTemplates) > 0 {
		fmt.Println("Updated templates:")
		for _, t := range result.UpdatedTemplates {
			fmt.Printf("  ~ %s\n", t.ID)
		}
		fmt.Println()
	}

	if len(result.RemovedTemplates) > 0 {
		fmt.Println("Removed templates:")
		for _, id := range result.RemovedTemplates {
			fmt.Printf("  - %s\n", id)
		}
		fmt.Println()
	}
}
