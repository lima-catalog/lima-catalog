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

// NewMetadataCollector creates a new metadata collector
func NewMetadataCollector(client *github.Client) *MetadataCollector {
	return &MetadataCollector{
		client: client,
	}
}

// CollectRepositoryMetadata fetches metadata for a repository
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
		IsFork:        ghRepo.GetFork(),
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

	if ghRepo.Parent != nil {
		repository.Parent = ghRepo.Parent.GetFullName()
	}

	return repository, nil
}

// CollectOrganizationMetadata fetches metadata for a user or organization
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

// fetchRepositoriesConcurrent fetches multiple repositories concurrently with controlled concurrency
// maxConcurrent controls how many API calls run in parallel (respects rate limits)
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

// fetchOrganizationsConcurrent fetches multiple organizations concurrently with controlled concurrency
// maxConcurrent controls how many API calls run in parallel (respects rate limits)
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

// SelectReposToRefresh selects repositories that need metadata refresh
// Returns repos from new templates + up to 5% of stale repos (>30 days old)
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

// SelectOrgsToRefresh selects organizations that need metadata refresh
// Returns orgs from new templates + up to 5% of stale orgs (>30 days old)
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

// CollectMetadataIncremental collects metadata for new templates and refreshes stale metadata
// Uses intelligent refresh cycle: new templates + 5% of stale (>30 days) entries per run
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

// CollectAllMetadata collects metadata for all unique repositories and organizations
// Used in non-incremental mode to fetch everything from scratch
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
