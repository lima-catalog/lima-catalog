package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/discovery"
	"github.com/lima-catalog/lima-catalog/pkg/github"
	"github.com/lima-catalog/lima-catalog/pkg/storage"
	"github.com/lima-catalog/lima-catalog/pkg/types"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	fmt.Println("====================================================================")
	fmt.Println("Lima Template Catalog - Data Collection Tool")
	fmt.Println("====================================================================")
	fmt.Println()

	// Get GitHub token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	// Get data directory from environment or use default
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	// Check if incremental mode is enabled
	incremental := os.Getenv("INCREMENTAL") != ""

	fmt.Printf("Data directory: %s\n", dataDir)
	fmt.Printf("Incremental mode: %v\n", incremental)
	fmt.Println()

	// Initialize storage
	store, err := storage.NewStorage(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Load or create progress
	progress, err := store.LoadProgress()
	if err != nil {
		return fmt.Errorf("failed to load progress: %w", err)
	}

	fmt.Printf("Current phase: %s\n", progress.Phase)
	fmt.Printf("Templates discovered: %d\n", progress.TemplatesDiscovered)
	fmt.Printf("Repos fetched: %d\n", progress.ReposFetched)
	fmt.Printf("Orgs fetched: %d\n", progress.OrgsFetched)
	fmt.Println()

	// Create GitHub client
	ctx := context.Background()
	client := github.NewClient(ctx, token)

	// Check rate limit
	fmt.Println("Checking GitHub API rate limit...")
	limits, err := client.RateLimit()
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	core := limits.Core
	search := limits.Search

	fmt.Printf("Core API: %d/%d remaining (resets at %s)\n",
		core.Remaining, core.Limit, core.Reset.Time.Format(time.RFC3339))
	fmt.Printf("Search API: %d/%d remaining (resets at %s)\n",
		search.Remaining, search.Limit, search.Reset.Time.Format(time.RFC3339))
	fmt.Println()

	// Check if we have enough quota
	if core.Remaining < 100 {
		return fmt.Errorf("insufficient rate limit: only %d core API calls remaining", core.Remaining)
	}

	if search.Remaining < 5 {
		return fmt.Errorf("insufficient rate limit: only %d search API calls remaining", search.Remaining)
	}

	// Phase 1: Discovery
	var templates []types.Template

	if progress.Phase == "discovery" {
		fmt.Println("=== Phase 1: Template Discovery ===")
		fmt.Println()

		discoverer := discovery.NewDiscoverer(client)

		discoveredTemplates, err := discoverer.DiscoverAll()
		if err != nil {
			return fmt.Errorf("discovery failed: %w", err)
		}

		// If incremental mode, load existing templates and merge
		if incremental {
			existingTemplates, err := store.LoadTemplates()
			if err != nil {
				fmt.Printf("Warning: failed to load existing templates: %v\n", err)
				fmt.Println("Continuing with full collection...")
				templates = discoveredTemplates
			} else {
				fmt.Printf("Loaded %d existing templates for incremental update\n", len(existingTemplates))
				updateResult := discovery.MergeTemplates(existingTemplates, discoveredTemplates)
				discovery.PrintUpdateSummary(updateResult)

				// Use all templates (new + updated + unchanged)
				templates = updateResult.AllTemplates
			}
		} else {
			templates = discoveredTemplates
		}

		// Save templates
		if err := store.SaveTemplates(templates); err != nil {
			return fmt.Errorf("failed to save templates: %w", err)
		}

		// Count official vs community
		officialCount := 0
		communityCount := 0
		for _, t := range templates {
			if t.IsOfficial {
				officialCount++
			} else {
				communityCount++
			}
		}

		// Update progress
		progress.Phase = "metadata"
		progress.TemplatesDiscovered = len(templates)
		progress.OfficialTemplates = officialCount
		progress.CommunityTemplates = communityCount
		progress.LastUpdated = time.Now()

		if err := store.SaveProgress(progress); err != nil {
			return fmt.Errorf("failed to save progress: %w", err)
		}

		fmt.Printf("✓ Total templates: %d (%d official, %d community)\n",
			len(templates), officialCount, communityCount)
		fmt.Println()
	} else {
		// Load existing templates
		templates, err = store.LoadTemplates()
		if err != nil {
			return fmt.Errorf("failed to load templates: %w", err)
		}
		fmt.Printf("Loaded %d existing templates\n\n", len(templates))
	}

	// Phase 2: Metadata Collection
	if progress.Phase == "metadata" {
		fmt.Println("=== Phase 2: Metadata Collection ===")
		fmt.Println()

		collector := discovery.NewMetadataCollector(client)

		newRepositories, newOrganizations, err := collector.CollectAllMetadata(templates)
		if err != nil {
			return fmt.Errorf("metadata collection failed: %w", err)
		}

		// If incremental mode, merge with existing data
		var repositories []types.Repository
		var organizations []types.Organization

		if incremental {
			existingRepos, err := store.LoadRepositories()
			if err != nil {
				fmt.Printf("Warning: failed to load existing repos: %v\n", err)
				repositories = newRepositories
			} else {
				repositories = discovery.MergeRepositories(existingRepos, newRepositories)
				fmt.Printf("Merged repository data: %d total\n", len(repositories))
			}

			existingOrgs, err := store.LoadOrganizations()
			if err != nil {
				fmt.Printf("Warning: failed to load existing orgs: %v\n", err)
				organizations = newOrganizations
			} else {
				organizations = discovery.MergeOrganizations(existingOrgs, newOrganizations)
				fmt.Printf("Merged organization data: %d total\n", len(organizations))
			}
		} else {
			repositories = newRepositories
			organizations = newOrganizations
		}

		// Save repositories
		if err := store.SaveRepositories(repositories); err != nil {
			return fmt.Errorf("failed to save repositories: %w", err)
		}

		// Save organizations
		if err := store.SaveOrganizations(organizations); err != nil {
			return fmt.Errorf("failed to save organizations: %w", err)
		}

		// Update progress
		progress.Phase = "complete"
		progress.ReposFetched = len(repositories)
		progress.OrgsFetched = len(organizations)
		progress.LastUpdated = time.Now()

		// Update rate limit info
		limits, _ := client.RateLimit()
		if limits != nil {
			progress.RateLimitRemaining = limits.Core.Remaining
			progress.RateLimitReset = limits.Core.Reset.Time
		}

		if err := store.SaveProgress(progress); err != nil {
			return fmt.Errorf("failed to save progress: %w", err)
		}

		fmt.Printf("✓ Collected metadata for %d repositories\n", len(repositories))
		fmt.Printf("✓ Collected metadata for %d organizations\n", len(organizations))
		fmt.Println()
	}

	// Final summary
	fmt.Println("====================================================================")
	fmt.Println("Collection Complete!")
	fmt.Println("====================================================================")
	fmt.Println()
	fmt.Printf("Total templates: %d\n", progress.TemplatesDiscovered)
	fmt.Printf("  Official: %d\n", progress.OfficialTemplates)
	fmt.Printf("  Community: %d\n", progress.CommunityTemplates)
	fmt.Printf("Repositories: %d\n", progress.ReposFetched)
	fmt.Printf("Organizations: %d\n", progress.OrgsFetched)
	fmt.Println()
	fmt.Printf("Data saved to: %s\n", dataDir)
	fmt.Printf("  - templates.jsonl\n")
	fmt.Printf("  - repos.jsonl\n")
	fmt.Printf("  - orgs.jsonl\n")
	fmt.Printf("  - progress.json\n")
	fmt.Println()

	// Final rate limit check
	limits, _ = client.RateLimit()
	if limits != nil {
		fmt.Printf("API calls used: %d core, %d search\n",
			5000-limits.Core.Remaining,
			30-limits.Search.Remaining)
	}

	return nil
}
