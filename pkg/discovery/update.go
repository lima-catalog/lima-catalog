package discovery

import (
	"cmp"
	"fmt"
	"slices"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/interfaces"
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

// backfillLastUpdated initializes LastUpdated for templates that don't have it set
// This handles migration for old data before the LastUpdated field was added
func backfillLastUpdated(template *types.Template) {
	if template.LastUpdated.IsZero() {
		// Set to a time definitely before AnalyzedAt to avoid re-analysis
		if !template.AnalyzedAt.IsZero() {
			// Set to 1 hour before AnalyzedAt (or DiscoveredAt if that's earlier)
			template.LastUpdated = template.DiscoveredAt
			if template.AnalyzedAt.Add(-1 * time.Hour).After(template.DiscoveredAt) {
				template.LastUpdated = template.AnalyzedAt.Add(-1 * time.Hour)
			}
		} else {
			template.LastUpdated = template.DiscoveredAt
		}
	}
}

// processUpdatedTemplate handles a template whose SHA has changed
func processUpdatedTemplate(oldTemplate types.Template, newTemplate types.Template, clock interfaces.Clock) types.Template {
	newTemplate.DiscoveredAt = oldTemplate.DiscoveredAt // Preserve original discovery time
	newTemplate.LastChecked = clock.Now()                // We checked it
	newTemplate.LastUpdated = clock.Now()                // Content changed
	return newTemplate
}

// processUnchangedTemplate handles a template whose SHA hasn't changed but was checked
func processUnchangedTemplate(template types.Template, clock interfaces.Clock) types.Template {
	template.LastChecked = clock.Now() // We checked it
	// Don't update LastUpdated - content didn't change
	backfillLastUpdated(&template)
	return template
}

// processNewTemplate handles a newly discovered template
func processNewTemplate(template types.Template, clock interfaces.Clock) types.Template {
	template.LastChecked = clock.Now() // First check
	template.LastUpdated = clock.Now() // New content
	return template
}

// processUncheckedTemplate handles a template that wasn't checked this run
func processUncheckedTemplate(template types.Template) types.Template {
	// Template wasn't checked this run - preserve unchanged, don't update LastChecked
	backfillLastUpdated(&template)
	return template
}

// MergeTemplates performs an incremental update by merging existing templates with newly discovered ones
func MergeTemplates(existing, discovered []types.Template, clock interfaces.Clock) UpdateResult {
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
			// Template already exists - check if it changed
			if oldTemplate.SHA != newTemplate.SHA {
				// SHA changed - template was updated
				updated := processUpdatedTemplate(oldTemplate, newTemplate, clock)
				result.UpdatedTemplates = append(result.UpdatedTemplates, updated)
				result.AllTemplates = append(result.AllTemplates, updated)
			} else {
				// SHA unchanged - template is the same but we checked it
				unchanged := processUnchangedTemplate(oldTemplate, clock)
				result.UnchangedCount++
				result.AllTemplates = append(result.AllTemplates, unchanged)
			}
			preservedTemplates[id] = true
		} else {
			// New template
			new := processNewTemplate(newTemplate, clock)
			result.NewTemplates = append(result.NewTemplates, new)
			result.AllTemplates = append(result.AllTemplates, new)
			preservedTemplates[id] = true
		}
	}

	// Preserve existing templates that weren't in discoveredMap
	// In incremental mode, absence from discoveredMap means "not checked", not "removed"
	// (Template deletion detection is Stage 7, not implemented yet)
	for id, oldTemplate := range existingMap {
		if !preservedTemplates[id] {
			// Template wasn't checked this run - preserve as-is
			unchecked := processUncheckedTemplate(oldTemplate)
			result.UnchangedCount++
			result.AllTemplates = append(result.AllTemplates, unchecked)
		}
	}

	// Note: RemovedTemplates is currently unused
	// Template deletion detection will be implemented in Stage 7

	// Sort templates by ID for stable output
	slices.SortFunc(result.AllTemplates, func(a, b types.Template) int {
		return cmp.Compare(a.ID, b.ID)
	})

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
	slices.SortFunc(result, func(a, b types.Repository) int {
		return cmp.Or(
			cmp.Compare(a.Owner, b.Owner),
			cmp.Compare(a.Name, b.Name),
		)
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
	slices.SortFunc(result, func(a, b types.Organization) int {
		return cmp.Compare(a.ID, b.ID)
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
