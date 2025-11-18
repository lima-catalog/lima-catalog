package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/combiner"
	"github.com/lima-catalog/lima-catalog/pkg/config"
	"github.com/lima-catalog/lima-catalog/pkg/discovery"
	"github.com/lima-catalog/lima-catalog/pkg/github"
	"github.com/lima-catalog/lima-catalog/pkg/interfaces"
	"github.com/lima-catalog/lima-catalog/pkg/storage"
	"github.com/lima-catalog/lima-catalog/pkg/types"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// appConfig holds environment configuration
type appConfig struct {
	token        string
	dataDir      string
	incremental  bool
	analyze      bool
	forceAnalyze bool
}

// setupEnvironment reads and validates environment variables
func setupEnvironment() (*appConfig, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	cfg := &appConfig{
		token:        token,
		dataDir:      dataDir,
		incremental:  os.Getenv("INCREMENTAL") != "",
		analyze:      os.Getenv("ANALYZE") != "",
		forceAnalyze: os.Getenv("FORCE_ANALYZE") != "",
	}

	return cfg, nil
}

// printConfig displays the current configuration
func printConfig(cfg *appConfig) {
	fmt.Println("====================================================================")
	fmt.Println("Lima Template Catalog - Data Collection Tool")
	fmt.Println("====================================================================")
	fmt.Println()
	fmt.Printf("Data directory: %s\n", cfg.dataDir)
	fmt.Printf("Incremental mode: %v\n", cfg.incremental)
	fmt.Printf("Analysis mode: %v\n", cfg.analyze)
	if cfg.forceAnalyze {
		fmt.Printf("Force re-analyze: %v (will re-analyze ALL templates)\n", cfg.forceAnalyze)
	}
	fmt.Println()
}

// initializeStorage creates and initializes storage
func initializeStorage(dataDir string) (*storage.Storage, error) {
	store, err := storage.NewStorage(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	return store, nil
}

// loadAndPrintProgress loads progress and displays current status
func loadAndPrintProgress(store *storage.Storage) (*types.Progress, error) {
	progress, err := store.LoadProgress()
	if err != nil {
		return nil, fmt.Errorf("failed to load progress: %w", err)
	}

	fmt.Printf("Current phase: %s\n", progress.Phase)
	fmt.Printf("Templates discovered: %d\n", progress.TemplatesDiscovered)
	fmt.Printf("Repos fetched: %d\n", progress.ReposFetched)
	fmt.Printf("Orgs fetched: %d\n", progress.OrgsFetched)
	fmt.Println()

	return progress, nil
}

// checkRateLimits verifies we have sufficient API quota
func checkRateLimits(client *github.Client) error {
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
	if core.Remaining < config.MinCoreRateLimitRemaining {
		return fmt.Errorf("insufficient rate limit: only %d core API calls remaining (minimum: %d)",
			core.Remaining, config.MinCoreRateLimitRemaining)
	}

	if search.Remaining < config.MinSearchRateLimitRemaining {
		return fmt.Errorf("insufficient rate limit: only %d search API calls remaining (minimum: %d)",
			search.Remaining, config.MinSearchRateLimitRemaining)
	}

	return nil
}

// runAnalysisPhase analyzes templates for keywords, categories, and notability
func runAnalysisPhase(ctx context.Context, client *github.Client, store *storage.Storage, forceAnalyze bool) error {
	fmt.Println("=== Phase 3: Template Analysis ===")
	fmt.Println()

	// Load templates and repositories
	templates, err := store.LoadTemplates()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	repositories, err := store.LoadRepositories()
	if err != nil {
		return fmt.Errorf("failed to load repositories: %w", err)
	}

	// Create repository map for quick lookup
	repoMap := make(map[string]*types.Repository)
	for i := range repositories {
		repoMap[repositories[i].ID] = &repositories[i]
	}

	// Create analyzer
	analyzer := discovery.NewAnalyzer(forceAnalyze)

	// Fetch official images for notability scoring
	fmt.Println("Fetching official images from lima-vm/lima...")
	if err := analyzer.FetchOfficialImagesForAnalyzer(ctx, client.GetClient()); err != nil {
		fmt.Printf("Warning: failed to fetch official images: %v\n", err)
		fmt.Println("Continuing with analysis (all images will be considered unusual)...")
	}
	fmt.Println()

	// Fetch default template comments for filtering
	fmt.Println("Fetching default template comments from lima-vm/lima...")
	if err := analyzer.FetchDefaultTemplateComments(ctx, client.GetClient()); err != nil {
		fmt.Printf("Warning: failed to fetch default template comments: %v\n", err)
		fmt.Println("Continuing with analysis (default comments won't be filtered)...")
	}
	fmt.Println()

	// Analyze templates
	analyzedTemplates, err := analyzer.AnalyzeTemplates(templates, repoMap)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Save analyzed templates
	if err := store.SaveTemplates(analyzedTemplates); err != nil {
		return fmt.Errorf("failed to save analyzed templates: %w", err)
	}

	fmt.Printf("✓ Analyzed %d templates\n", len(analyzedTemplates))
	fmt.Println()

	return nil
}

// runCombinePhase combines all data for the frontend
func runCombinePhase(store *storage.Storage, dataDir string) error {
	fmt.Println("=== Phase 4: Frontend Data Combination ===")
	fmt.Println()

	// Load all data files
	combineTemplates, err := store.LoadTemplates()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	combineRepos, err := store.LoadRepositories()
	if err != nil {
		return fmt.Errorf("failed to load repositories: %w", err)
	}

	combineOrgs, err := store.LoadOrganizations()
	if err != nil {
		return fmt.Errorf("failed to load organizations: %w", err)
	}

	// Load blocklist
	blocklist, err := discovery.LoadBlocklist(filepath.Join("config", "blocklist.yaml"))
	if err != nil {
		fmt.Printf("Warning: Failed to load blocklist: %v\n", err)
		blocklist = &types.Blocklist{} // Use empty blocklist
	}

	// Create combiner
	dataCombiner := combiner.NewCombiner(blocklist)

	// Generate catalog file for frontend
	catalogPath := filepath.Join(dataDir, "catalog.jsonl")
	if err := dataCombiner.CombineData(combineTemplates, combineRepos, combineOrgs, catalogPath); err != nil {
		return fmt.Errorf("failed to combine data: %w", err)
	}

	fmt.Printf("✓ Created catalog.jsonl\n")
	fmt.Println()

	return nil
}

// printFinalSummary displays collection completion status
func printFinalSummary(progress *types.Progress, dataDir string, client *github.Client) {
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
	fmt.Printf("  - catalog.jsonl (frontend data)\n")
	fmt.Printf("  - progress.json\n")
	fmt.Println()

	// Final rate limit check
	limits, _ := client.RateLimit()
	if limits != nil {
		fmt.Printf("API calls used: %d core, %d search\n",
			5000-limits.Core.Remaining,
			30-limits.Search.Remaining)
	}
}

// runMetadataPhase collects repository and organization metadata
func runMetadataPhase(ctx context.Context, client *github.Client, store *storage.Storage, progress *types.Progress, templates []types.Template, updateResult discovery.UpdateResult, incremental bool) error {
	// In incremental mode, always update metadata
	// In non-incremental mode, only run if phase is "metadata"
	shouldCollectMetadata := incremental || progress.Phase == "metadata"

	if !shouldCollectMetadata {
		return nil
	}

	fmt.Println("=== Phase 2: Metadata Collection ===")
	fmt.Println()

	collector := discovery.NewMetadataCollector(client)

	var repositories []types.Repository
	var organizations []types.Organization
	var err error

	if incremental {
		// Load existing metadata
		existingRepos, err := store.LoadRepositories()
		if err != nil {
			fmt.Printf("Warning: failed to load existing repos: %v\n", err)
			existingRepos = []types.Repository{}
		}

		existingOrgs, err := store.LoadOrganizations()
		if err != nil {
			fmt.Printf("Warning: failed to load existing orgs: %v\n", err)
			existingOrgs = []types.Organization{}
		}

		// Use incremental metadata collection
		// Only fetch metadata for NEW templates (not all discovered, which may include unchanged official templates)
		// Also refreshes 5% of stale (>30 days) entries
		newTemplates := append(updateResult.NewTemplates, updateResult.UpdatedTemplates...)
		repositories, organizations, err = collector.CollectMetadataIncremental(newTemplates, existingRepos, existingOrgs)
		if err != nil {
			return fmt.Errorf("incremental metadata collection failed: %w", err)
		}
	} else {
		// Use full metadata collection
		repositories, organizations, err = collector.CollectAllMetadata(templates)
		if err != nil {
			return fmt.Errorf("metadata collection failed: %w", err)
		}
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
	if !incremental {
		// Only update phase if not in incremental mode
		progress.Phase = "complete"
	}
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

	return nil
}

// runDiscoveryPhase discovers templates (either full or incremental)
func runDiscoveryPhase(ctx context.Context, client *github.Client, store *storage.Storage, progress *types.Progress, incremental bool) ([]types.Template, discovery.UpdateResult, error) {
	var templates []types.Template
	var updateResult discovery.UpdateResult

	// In incremental mode, always check for new/updated templates
	// In non-incremental mode, only run if phase is "discovery"
	shouldDiscover := incremental || progress.Phase == "discovery"

	if shouldDiscover {
		fmt.Println("=== Phase 1: Template Discovery ===")
		fmt.Println()

		// Load blocklist
		blocklist, err := discovery.LoadBlocklist("config/blocklist.yaml")
		if err != nil {
			return nil, updateResult, fmt.Errorf("failed to load blocklist: %w", err)
		}
		fmt.Printf("Loaded blocklist: %d path patterns, %d repo patterns\n", len(blocklist.Paths), len(blocklist.Repos))

		discoverer := discovery.NewDiscoverer(client, blocklist)

		// Determine search date for incremental mode
		var sinceDate time.Time
		var existingTemplates []types.Template
		if incremental {
			// Load existing templates to find newest timestamp
			existingTemplates, err = store.LoadTemplates()
			if err != nil {
				fmt.Printf("Warning: failed to load existing templates: %v\n", err)
				fmt.Println("Continuing with full discovery (no incremental filtering)...")
				fmt.Println()
			} else {
				fmt.Printf("Loaded %d existing templates for incremental update\n", len(existingTemplates))

				// Find newest template and search 24 hours before it
				newestTimestamp := discovery.FindNewestTemplateTimestamp(existingTemplates)
				if !newestTimestamp.IsZero() {
					sinceDate = newestTimestamp.Add(-24 * time.Hour)
					fmt.Printf("Newest template discovered at: %s\n", newestTimestamp.Format(time.RFC3339))
					fmt.Printf("Searching for templates pushed since: %s (24h overlap)\n\n", sinceDate.Format(time.RFC3339))
				} else {
					fmt.Println("No existing templates found - running full discovery")
					fmt.Println()
				}
			}
		}

		discoveredTemplates, err := discoverer.DiscoverAll(sinceDate, existingTemplates)
		if err != nil {
			return nil, updateResult, fmt.Errorf("discovery failed: %w", err)
		}

		// Note: Finding 0 templates in incremental mode is normal if no repositories have been updated
		// The pushed:>DATE query only finds repositories that were pushed after that date
		if incremental && !sinceDate.IsZero() && len(discoveredTemplates) == 0 {
			fmt.Println("\nNote: Incremental search found 0 new/updated templates")
			fmt.Println("This is normal if no repositories containing templates have been updated since the search date.")
		}

		// If incremental mode, merge with existing templates
		if incremental && len(existingTemplates) > 0 {
			updateResult = discovery.MergeTemplates(existingTemplates, discoveredTemplates, interfaces.NewDefaultClock())
			discovery.PrintUpdateSummary(updateResult)

			// Use all templates (new + updated + unchanged)
			templates = updateResult.AllTemplates
		} else {
			// Non-incremental mode: all templates are "new"
			templates = discoveredTemplates
			// Populate updateResult for consistency (so metadata collection logic works the same)
			updateResult.NewTemplates = discoveredTemplates
		}

		// Save templates
		if err := store.SaveTemplates(templates); err != nil {
			return nil, updateResult, fmt.Errorf("failed to save templates: %w", err)
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
		if !incremental {
			// Only update phase if not in incremental mode
			progress.Phase = "metadata"
		}
		progress.TemplatesDiscovered = len(templates)
		progress.OfficialTemplates = officialCount
		progress.CommunityTemplates = communityCount
		progress.LastUpdated = time.Now()

		if err := store.SaveProgress(progress); err != nil {
			return nil, updateResult, fmt.Errorf("failed to save progress: %w", err)
		}

		fmt.Printf("✓ Total templates: %d (%d official, %d community)\n",
			len(templates), officialCount, communityCount)
		fmt.Println()
	} else {
		// Load existing templates
		var err error
		templates, err = store.LoadTemplates()
		if err != nil {
			return nil, updateResult, fmt.Errorf("failed to load templates: %w", err)
		}
		fmt.Printf("Loaded %d existing templates\n\n", len(templates))
	}

	return templates, updateResult, nil
}

func run() error {
	// Setup environment and configuration
	cfg, err := setupEnvironment()
	if err != nil {
		return err
	}
	printConfig(cfg)

	// Initialize storage
	store, err := initializeStorage(cfg.dataDir)
	if err != nil {
		return err
	}

	// Load progress
	progress, err := loadAndPrintProgress(store)
	if err != nil {
		return err
	}

	// Create GitHub client
	ctx := context.Background()
	client := github.NewClient(ctx, cfg.token)

	// Check rate limits
	if err := checkRateLimits(client); err != nil {
		return err
	}

	// Phase 1: Discovery
	templates, updateResult, err := runDiscoveryPhase(ctx, client, store, progress, cfg.incremental)
	if err != nil {
		return err
	}

	// Phase 2: Metadata Collection
	if err := runMetadataPhase(ctx, client, store, progress, templates, updateResult, cfg.incremental); err != nil {
		return err
	}

	// Phase 3: Template Analysis (optional)
	if cfg.analyze {
		if err := runAnalysisPhase(ctx, client, store, cfg.forceAnalyze); err != nil {
			return err
		}
	}

	// Phase 4: Frontend Data Combination
	if err := runCombinePhase(store, cfg.dataDir); err != nil {
		return err
	}

	// Final summary
	printFinalSummary(progress, cfg.dataDir, client)

	return nil
}
