package discovery

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/config"
	"github.com/lima-catalog/lima-catalog/pkg/github"
	"github.com/lima-catalog/lima-catalog/pkg/types"
	"github.com/lima-catalog/lima-catalog/pkg/validation"
)

// MetadataCollector collects metadata for repositories and organizations
type MetadataCollector struct {
	client *github.Client
}

// NewMetadataCollector creates a new metadata collector with the provided GitHub client.
//
// The collector uses the client to fetch repository and organization information
// from the GitHub API with caching and rate limit management.
//
// Parameters:
//   - client: GitHub API client with authentication and caching
//
// Returns a configured MetadataCollector ready to fetch metadata.
func NewMetadataCollector(client *github.Client) *MetadataCollector {
	return &MetadataCollector{
		client: client,
	}
}

// CollectRepositoryMetadata fetches metadata for a single repository from the GitHub API.
//
// Fetches comprehensive repository information including:
//   - Basic info: description, topics, default branch
//   - Statistics: stars, forks, watchers
//   - Timestamps: created, updated, pushed dates
//   - License and language information
//   - Fork relationships (parent repo if applicable)
//
// Parameters:
//   - repoFullName: Repository identifier in "owner/repo" format
//
// Returns:
//   - Repository metadata populated from GitHub API response
//   - Error if repository name is invalid or API call fails
//
// The function validates the repository name format and wraps any API errors
// with additional context for debugging.
func (m *MetadataCollector) CollectRepositoryMetadata(repoFullName string) (*types.Repository, error) {
	owner, repo, err := validation.ParseRepoID(repoFullName)
	if err != nil {
		return nil, fmt.Errorf("invalid repository name: %s", repoFullName)
	}

	ghRepo, err := m.client.GetRepository(owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository %s: %w", repoFullName, err)
	}

	repository := &types.Repository{
		ID:            repoFullName,
		Owner:         owner,
		Name:          repo,
		Description:   ghRepo.GetDescription(),
		Topics:        ghRepo.Topics,
		Stars:         ghRepo.GetStargazersCount(),
		Forks:         ghRepo.GetForksCount(),
		Watchers:      ghRepo.GetWatchersCount(),
		Language:      ghRepo.GetLanguage(),
		DefaultBranch: ghRepo.GetDefaultBranch(),
		Homepage:      ghRepo.GetHomepage(),
		LastFetched:   time.Now(),
	}

	if ghRepo.License != nil {
		repository.License = ghRepo.License.GetSPDXID()
	}

	if ghRepo.CreatedAt != nil {
		repository.CreatedAt = ghRepo.CreatedAt.Time
	}

	if ghRepo.UpdatedAt != nil {
		repository.UpdatedAt = ghRepo.UpdatedAt.Time
	}

	if ghRepo.PushedAt != nil {
		repository.PushedAt = ghRepo.PushedAt.Time
	}

	return repository, nil
}

// CollectOrganizationMetadata fetches metadata for a GitHub user or organization.
//
// Fetches information including:
//   - Login name and display name
//   - Type (User or Organization)
//   - Bio/description
//   - Location, blog URL, and email
//
// Parameters:
//   - login: GitHub username or organization name
//
// Returns:
//   - Organization metadata populated from GitHub API response
//   - Error if API call fails
//
// The function works for both individual users and organization accounts,
// as the GitHub API treats them similarly.
func (m *MetadataCollector) CollectOrganizationMetadata(login string) (*types.Organization, error) {
	user, err := m.client.GetUser(login)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user/org %s: %w", login, err)
	}

	org := &types.Organization{
		ID:          login,
		Login:       login,
		Type:        user.GetType(),
		Name:        user.GetName(),
		Description: user.GetBio(),
		Location:    user.GetLocation(),
		Blog:        user.GetBlog(),
		Email:       user.GetEmail(),
		LastFetched: time.Now(),
	}

	return org, nil
}

// fetchRepositoriesConcurrent fetches multiple repositories concurrently with controlled concurrency.
//
// Uses goroutines with a semaphore pattern to limit concurrent API calls and
// respect GitHub rate limits. Requests are staggered with delays to avoid bursts.
//
// Parameters:
//   - repoNames: List of repository identifiers in "owner/repo" format
//   - maxConcurrent: Maximum number of simultaneous API calls (semaphore size)
//
// Returns a map of repository ID to Repository metadata. Failed fetches are logged
// as warnings but do not stop other fetches from completing.
//
// Thread Safety:
// The function is safe for concurrent use. Results are protected by a mutex
// to prevent race conditions when multiple goroutines write to the map.
//
// Rate Limiting:
// Each fetch (after the first) includes a configurable delay to respect GitHub
// API rate limits. The delay is defined in config.MetadataAPIDelay.
func (m *MetadataCollector) fetchRepositoriesConcurrent(repoNames []string, maxConcurrent int) map[string]types.Repository {
	results := make(map[string]types.Repository)
	var mu sync.Mutex // Protects results map
	var wg sync.WaitGroup

	// Semaphore to limit concurrency
	sem := make(chan struct{}, maxConcurrent)

	for i, repoName := range repoNames {
		wg.Add(1)
		go func(name string, index int) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Add delay before each fetch to respect rate limits
			// Stagger initial requests to avoid burst
			if index > 0 {
				time.Sleep(config.MetadataAPIDelay)
			}

			repo, err := m.CollectRepositoryMetadata(name)
			if err != nil {
				fmt.Printf("Warning: failed to fetch %s: %v\n", name, err)
				return
			}

			// Store result (thread-safe)
			mu.Lock()
			results[repo.ID] = *repo
			mu.Unlock()
		}(repoName, i)
	}

	wg.Wait()
	return results
}

// fetchOrganizationsConcurrent fetches multiple organizations concurrently with controlled concurrency.
//
// Uses goroutines with a semaphore pattern to limit concurrent API calls and
// respect GitHub rate limits. Requests are staggered with delays to avoid bursts.
//
// Parameters:
//   - orgNames: List of GitHub usernames or organization names
//   - maxConcurrent: Maximum number of simultaneous API calls (semaphore size)
//
// Returns a map of organization ID (login) to Organization metadata. Failed fetches
// are logged as warnings but do not stop other fetches from completing.
//
// Thread Safety:
// The function is safe for concurrent use. Results are protected by a mutex
// to prevent race conditions when multiple goroutines write to the map.
//
// Rate Limiting:
// Each fetch (after the first) includes a configurable delay to respect GitHub
// API rate limits. The delay is defined in config.MetadataAPIDelay.
func (m *MetadataCollector) fetchOrganizationsConcurrent(orgNames []string, maxConcurrent int) map[string]types.Organization {
	results := make(map[string]types.Organization)
	var mu sync.Mutex // Protects results map
	var wg sync.WaitGroup

	// Semaphore to limit concurrency
	sem := make(chan struct{}, maxConcurrent)

	for i, orgName := range orgNames {
		wg.Add(1)
		go func(name string, index int) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Add delay before each fetch to respect rate limits
			// Stagger initial requests to avoid burst
			if index > 0 {
				time.Sleep(config.MetadataAPIDelay)
			}

			org, err := m.CollectOrganizationMetadata(name)
			if err != nil {
				fmt.Printf("Warning: failed to fetch %s: %v\n", name, err)
				return
			}

			// Store result (thread-safe)
			mu.Lock()
			results[org.ID] = *org
			mu.Unlock()
		}(orgName, i)
	}

	wg.Wait()
	return results
}

// SelectReposToRefresh selects repositories that need metadata refresh using incremental strategy.
//
// Implements intelligent refresh cycle to minimize API calls while keeping data fresh:
//  1. Always refresh: repositories from newly discovered templates
//  2. Incrementally refresh: up to 5% of stale repositories (>30 days old)
//     prioritizing oldest first
//
// Parameters:
//   - newTemplates: Templates discovered in current run (need fresh repo metadata)
//   - existingRepos: All repositories currently in the database
//
// Returns:
// List of repository IDs ("owner/repo") that should be refreshed in this run.
//
// Refresh Strategy:
// The 5% incremental refresh ensures all stale metadata is updated over ~20 runs
// (about 20 days with daily runs), preventing unbounded staleness while keeping
// API usage low. Minimum 1 stale repo is refreshed per run even if 5% rounds to zero.
//
// Example:
// - 100 existing repos, 30 stale (>30 days), 5 new templates with 3 unique repos
// - Returns: 3 repos (from new templates) + 5 repos (oldest 5% of stale) = 8 total
func SelectReposToRefresh(newTemplates []types.Template, existingRepos []types.Repository) []string {
	// Get repos from new templates
	newRepoSet := make(map[string]bool)
	for _, t := range newTemplates {
		newRepoSet[t.Repo] = true
	}

	// Find stale repos (>30 days old) and sort by age (oldest first)
	const staleThreshold = 30 * 24 * time.Hour
	var staleCandidates []types.Repository

	for _, repo := range existingRepos {
		// Skip if this repo is already in new templates
		if newRepoSet[repo.ID] {
			continue
		}

		// Check if stale
		if time.Since(repo.LastFetched) > staleThreshold {
			staleCandidates = append(staleCandidates, repo)
		}
	}

	// Sort stale repos by LastFetched (oldest first)
	slices.SortFunc(staleCandidates, func(a, b types.Repository) int {
		if a.LastFetched.Before(b.LastFetched) {
			return -1
		}
		if b.LastFetched.Before(a.LastFetched) {
			return 1
		}
		return 0
	})

	// Select up to 5% of stale repos (prioritize oldest)
	maxRefresh := max(1, len(existingRepos)/20) // At least 1, up to 5%

	var staleToRefresh []string
	refreshCount := min(maxRefresh, len(staleCandidates))

	for i := 0; i < refreshCount; i++ {
		staleToRefresh = append(staleToRefresh, staleCandidates[i].ID)
	}

	// Combine new repos + stale repos
	newRepos := slices.Collect(maps.Keys(newRepoSet))
	result := make([]string, 0, len(newRepos)+len(staleToRefresh))
	result = append(result, newRepos...)
	result = append(result, staleToRefresh...)

	return result
}

// SelectOrgsToRefresh selects organizations that need metadata refresh using incremental strategy.
//
// Implements intelligent refresh cycle to minimize API calls while keeping data fresh:
//  1. Always refresh: organizations from newly discovered templates
//  2. Incrementally refresh: up to 5% of stale organizations (>30 days old)
//     prioritizing oldest first
//
// Parameters:
//   - newTemplates: Templates discovered in current run (extracts unique owners)
//   - existingOrgs: All organizations currently in the database
//
// Returns:
// List of organization IDs (GitHub logins) that should be refreshed in this run.
//
// Refresh Strategy:
// The 5% incremental refresh ensures all stale metadata is updated over ~20 runs
// (about 20 days with daily runs), preventing unbounded staleness while keeping
// API usage low. Minimum 1 stale org is refreshed per run even if 5% rounds to zero.
//
// Example:
// - 50 existing orgs, 15 stale (>30 days), 5 new templates from 2 unique owners
// - Returns: 2 orgs (from new templates) + 2 orgs (oldest 5% of stale) = 4 total
func SelectOrgsToRefresh(newTemplates []types.Template, existingOrgs []types.Organization) []string {
	// Get orgs from new templates
	newOrgSet := make(map[string]bool)
	for _, t := range newTemplates {
		owner, _, err := validation.ParseRepoID(t.Repo)
		if err == nil {
			newOrgSet[owner] = true
		}
	}

	// Find stale orgs (>30 days old) and sort by age (oldest first)
	const staleThreshold = 30 * 24 * time.Hour
	var staleCandidates []types.Organization

	for _, org := range existingOrgs {
		// Skip if this org is already in new templates
		if newOrgSet[org.ID] {
			continue
		}

		// Check if stale
		if time.Since(org.LastFetched) > staleThreshold {
			staleCandidates = append(staleCandidates, org)
		}
	}

	// Sort stale orgs by LastFetched (oldest first)
	slices.SortFunc(staleCandidates, func(a, b types.Organization) int {
		if a.LastFetched.Before(b.LastFetched) {
			return -1
		}
		if b.LastFetched.Before(a.LastFetched) {
			return 1
		}
		return 0
	})

	// Select up to 5% of stale orgs (prioritize oldest)
	maxRefresh := max(1, len(existingOrgs)/20) // At least 1, up to 5%

	var staleToRefresh []string
	refreshCount := min(maxRefresh, len(staleCandidates))

	for i := 0; i < refreshCount; i++ {
		staleToRefresh = append(staleToRefresh, staleCandidates[i].ID)
	}

	// Combine new orgs + stale orgs
	newOrgs := slices.Collect(maps.Keys(newOrgSet))
	result := make([]string, 0, len(newOrgs)+len(staleToRefresh))
	result = append(result, newOrgs...)
	result = append(result, staleToRefresh...)

	return result
}

// CollectMetadataIncremental collects metadata for new templates and refreshes stale metadata.
//
// Optimized for incremental mode where only a subset of metadata needs updating.
// Uses intelligent refresh strategy to balance data freshness with API usage.
//
// Parameters:
//   - newTemplates: Templates discovered in current run (require fresh metadata)
//   - existingRepos: All repository metadata from previous runs
//   - existingOrgs: All organization metadata from previous runs
//
// Returns:
//   - Complete list of repository metadata (existing + refreshed)
//   - Complete list of organization metadata (existing + refreshed)
//   - Error if critical API failures occur (individual fetch failures are logged but not fatal)
//
// Refresh Strategy:
//  1. Identifies repos/orgs needing refresh (new templates + 5% of stale entries)
//  2. Fetches metadata concurrently with rate limit protection
//  3. Merges refreshed metadata with existing data
//  4. Sorts results for stable output
//
// The function prints progress information showing how many entries are being
// refreshed and the concurrency level being used.
//
// Performance:
// Uses concurrent fetching (controlled by config.MaxMetadataConcurrency) to
// significantly speed up metadata collection while respecting API rate limits.
func (m *MetadataCollector) CollectMetadataIncremental(newTemplates []types.Template, existingRepos []types.Repository, existingOrgs []types.Organization) ([]types.Repository, []types.Organization, error) {
	// Select which repos and orgs need refreshing
	reposToRefresh := SelectReposToRefresh(newTemplates, existingRepos)
	orgsToRefresh := SelectOrgsToRefresh(newTemplates, existingOrgs)

	// Create maps of existing metadata for merging
	repoMap := make(map[string]types.Repository)
	for _, repo := range existingRepos {
		repoMap[repo.ID] = repo
	}

	orgMap := make(map[string]types.Organization)
	for _, org := range existingOrgs {
		orgMap[org.ID] = org
	}

	// Collect repository metadata concurrently
	fmt.Printf("\n=== Collecting Repository Metadata (Incremental) ===\n")
	fmt.Printf("New templates: %d repos | Stale (>30 days): refreshing up to 5%%\n", len(newTemplates))
	fmt.Printf("Fetching %d repositories (concurrent: %d)...\n", len(reposToRefresh), config.MaxMetadataConcurrency)

	if len(reposToRefresh) > 0 {
		fetchedRepos := m.fetchRepositoriesConcurrent(reposToRefresh, config.MaxMetadataConcurrency)
		// Merge fetched repos into existing map
		for id, repo := range fetchedRepos {
			repoMap[id] = repo
		}
	}
	fmt.Printf("Refreshed %d repositories\n\n", len(reposToRefresh))

	// Collect organization metadata concurrently
	fmt.Printf("=== Collecting Organization Metadata (Incremental) ===\n")
	fmt.Printf("Fetching %d organizations (concurrent: %d)...\n", len(orgsToRefresh), config.MaxMetadataConcurrency)

	if len(orgsToRefresh) > 0 {
		fetchedOrgs := m.fetchOrganizationsConcurrent(orgsToRefresh, config.MaxMetadataConcurrency)
		// Merge fetched orgs into existing map
		for id, org := range fetchedOrgs {
			orgMap[id] = org
		}
	}
	fmt.Printf("Refreshed %d organizations\n\n", len(orgsToRefresh))

	// Convert maps back to slices for merge functions
	collectedRepos := make([]types.Repository, 0, len(repoMap))
	for _, repo := range repoMap {
		collectedRepos = append(collectedRepos, repo)
	}

	collectedOrgs := make([]types.Organization, 0, len(orgMap))
	for _, org := range orgMap {
		collectedOrgs = append(collectedOrgs, org)
	}

	// Use merge functions to ensure proper sorting
	repositories := MergeRepositories(existingRepos, collectedRepos)
	organizations := MergeOrganizations(existingOrgs, collectedOrgs)

	return repositories, organizations, nil
}

// CollectAllMetadata collects metadata for all unique repositories and organizations from scratch.
//
// Used in non-incremental mode (first run or full refresh) to fetch metadata for
// every repository and organization referenced by the template set.
//
// Parameters:
//   - templates: All templates in the catalog (discovers unique repos/orgs)
//
// Returns:
//   - Complete list of repository metadata (sorted by owner/name)
//   - Complete list of organization metadata (sorted by ID)
//   - Error if critical API failures occur (individual fetch failures are logged but not fatal)
//
// Process:
//  1. Extracts unique repository IDs and owner names from templates
//  2. Fetches all repository metadata concurrently
//  3. Fetches all organization metadata concurrently
//  4. Sorts results for stable, deterministic output
//
// The function prints progress information showing the total count of unique
// repos/orgs and the concurrency level being used.
//
// Performance:
// Uses concurrent fetching (controlled by config.MaxMetadataConcurrency) to
// significantly speed up metadata collection. A full run typically processes
// 100-200 repos and 50-100 orgs in a few minutes rather than hours.
//
// Use Cases:
//   - Initial catalog setup (no existing metadata)
//   - Full refresh after schema changes
//   - Rebuilding catalog from scratch
func (m *MetadataCollector) CollectAllMetadata(templates []types.Template) ([]types.Repository, []types.Organization, error) {
	// Track unique repos and orgs
	repoMap := make(map[string]bool)
	orgMap := make(map[string]bool)

	for _, template := range templates {
		repoMap[template.Repo] = true

		// Extract owner from repo
		owner, _, err := validation.ParseRepoID(template.Repo)
		if err == nil {
			orgMap[owner] = true
		}
	}

	// Collect repository metadata concurrently
	fmt.Printf("\n=== Collecting Repository Metadata ===\n")
	repoNames := slices.Collect(maps.Keys(repoMap))
	fmt.Printf("Fetching %d repositories (concurrent: %d)...\n", len(repoNames), config.MaxMetadataConcurrency)

	fetchedRepos := m.fetchRepositoriesConcurrent(repoNames, config.MaxMetadataConcurrency)
	repositories := slices.Collect(maps.Values(fetchedRepos))
	fmt.Printf("Collected metadata for %d repositories\n\n", len(repositories))

	// Collect organization metadata concurrently
	fmt.Printf("=== Collecting Organization Metadata ===\n")
	orgNames := slices.Collect(maps.Keys(orgMap))
	fmt.Printf("Fetching %d organizations (concurrent: %d)...\n", len(orgNames), config.MaxMetadataConcurrency)

	fetchedOrgs := m.fetchOrganizationsConcurrent(orgNames, config.MaxMetadataConcurrency)
	organizations := slices.Collect(maps.Values(fetchedOrgs))
	fmt.Printf("Collected metadata for %d organizations\n\n", len(organizations))

	// Sort for stable output
	slices.SortFunc(repositories, func(a, b types.Repository) int {
		return cmp.Or(
			cmp.Compare(a.Owner, b.Owner),
			cmp.Compare(a.Name, b.Name),
		)
	})

	slices.SortFunc(organizations, func(a, b types.Organization) int {
		return cmp.Compare(a.ID, b.ID)
	})

	return repositories, organizations, nil
}
