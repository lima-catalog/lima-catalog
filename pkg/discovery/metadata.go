package discovery

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/github"
	"github.com/lima-catalog/lima-catalog/pkg/types"
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
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository name: %s", repoFullName)
	}

	owner, repo := parts[0], parts[1]

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

// SelectReposToRefresh selects repositories that need metadata refresh
// Returns repos from new templates + up to 5% of stale repos (>30 days old)
func SelectReposToRefresh(newTemplates []types.Template, existingRepos []types.Repository) []string {
	// Get repos from new templates
	newRepoSet := make(map[string]bool)
	for _, t := range newTemplates {
		newRepoSet[t.Repo] = true
	}

	// Find stale repos (>30 days old)
	const staleThreshold = 30 * 24 * time.Hour
	var staleCandidates []string
	existingRepoMap := make(map[string]types.Repository)

	for _, repo := range existingRepos {
		existingRepoMap[repo.ID] = repo

		// Skip if this repo is already in new templates
		if newRepoSet[repo.ID] {
			continue
		}

		// Check if stale
		if time.Since(repo.LastFetched) > staleThreshold {
			staleCandidates = append(staleCandidates, repo.ID)
		}
	}

	// Select up to 5% of stale repos
	maxRefresh := len(existingRepos) / 20 // 5%
	if maxRefresh < 1 {
		maxRefresh = 1
	}

	var staleToRefresh []string
	if len(staleCandidates) <= maxRefresh {
		staleToRefresh = staleCandidates
	} else {
		// Randomly sample to spread load evenly
		rand.Shuffle(len(staleCandidates), func(i, j int) {
			staleCandidates[i], staleCandidates[j] = staleCandidates[j], staleCandidates[i]
		})
		staleToRefresh = staleCandidates[:maxRefresh]
	}

	// Combine new repos + stale repos
	result := make([]string, 0, len(newRepoSet)+len(staleToRefresh))
	for repo := range newRepoSet {
		result = append(result, repo)
	}
	result = append(result, staleToRefresh...)

	return result
}

// SelectOrgsToRefresh selects organizations that need metadata refresh
// Returns orgs from new templates + up to 5% of stale orgs (>30 days old)
func SelectOrgsToRefresh(newTemplates []types.Template, existingOrgs []types.Organization) []string {
	// Get orgs from new templates
	newOrgSet := make(map[string]bool)
	for _, t := range newTemplates {
		parts := strings.Split(t.Repo, "/")
		if len(parts) == 2 {
			newOrgSet[parts[0]] = true
		}
	}

	// Find stale orgs (>30 days old)
	const staleThreshold = 30 * 24 * time.Hour
	var staleCandidates []string
	existingOrgMap := make(map[string]types.Organization)

	for _, org := range existingOrgs {
		existingOrgMap[org.ID] = org

		// Skip if this org is already in new templates
		if newOrgSet[org.ID] {
			continue
		}

		// Check if stale
		if time.Since(org.LastFetched) > staleThreshold {
			staleCandidates = append(staleCandidates, org.ID)
		}
	}

	// Select up to 5% of stale orgs
	maxRefresh := len(existingOrgs) / 20 // 5%
	if maxRefresh < 1 {
		maxRefresh = 1
	}

	var staleToRefresh []string
	if len(staleCandidates) <= maxRefresh {
		staleToRefresh = staleCandidates
	} else {
		// Randomly sample to spread load evenly
		rand.Shuffle(len(staleCandidates), func(i, j int) {
			staleCandidates[i], staleCandidates[j] = staleCandidates[j], staleCandidates[i]
		})
		staleToRefresh = staleCandidates[:maxRefresh]
	}

	// Combine new orgs + stale orgs
	result := make([]string, 0, len(newOrgSet)+len(staleToRefresh))
	for org := range newOrgSet {
		result = append(result, org)
	}
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

	// Collect repository metadata
	fmt.Printf("\n=== Collecting Repository Metadata (Incremental) ===\n")
	fmt.Printf("New templates: %d repos | Stale (>30 days): refreshing up to 5%%\n", len(newTemplates))
	fmt.Printf("Fetching %d repositories...\n", len(reposToRefresh))

	count := 0
	for _, repoName := range reposToRefresh {
		count++
		fmt.Printf("Fetching [%d/%d] %s...\n", count, len(reposToRefresh), repoName)

		repo, err := m.CollectRepositoryMetadata(repoName)
		if err != nil {
			fmt.Printf("Warning: failed to fetch %s: %v\n", repoName, err)
			continue
		}

		repoMap[repo.ID] = *repo
		time.Sleep(500 * time.Millisecond) // Be nice to the API
	}
	fmt.Printf("Refreshed %d repositories\n\n", len(reposToRefresh))

	// Collect organization metadata
	fmt.Printf("=== Collecting Organization Metadata (Incremental) ===\n")
	fmt.Printf("Fetching %d organizations...\n", len(orgsToRefresh))

	count = 0
	for _, orgName := range orgsToRefresh {
		count++
		fmt.Printf("Fetching [%d/%d] %s...\n", count, len(orgsToRefresh), orgName)

		org, err := m.CollectOrganizationMetadata(orgName)
		if err != nil {
			fmt.Printf("Warning: failed to fetch %s: %v\n", orgName, err)
			continue
		}

		orgMap[org.ID] = *org
		time.Sleep(500 * time.Millisecond) // Be nice to the API
	}
	fmt.Printf("Refreshed %d organizations\n\n", len(orgsToRefresh))

	// Convert maps back to slices
	repositories := make([]types.Repository, 0, len(repoMap))
	for _, repo := range repoMap {
		repositories = append(repositories, repo)
	}

	organizations := make([]types.Organization, 0, len(orgMap))
	for _, org := range orgMap {
		organizations = append(organizations, org)
	}

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
		parts := strings.Split(template.Repo, "/")
		if len(parts) == 2 {
			orgMap[parts[0]] = true
		}
	}

	var repositories []types.Repository
	var organizations []types.Organization

	// Collect repository metadata
	fmt.Printf("\n=== Collecting Repository Metadata ===\n")
	count := 0
	for repoName := range repoMap {
		count++
		fmt.Printf("Fetching [%d/%d] %s...\n", count, len(repoMap), repoName)

		repo, err := m.CollectRepositoryMetadata(repoName)
		if err != nil {
			fmt.Printf("Warning: failed to fetch %s: %v\n", repoName, err)
			continue
		}

		repositories = append(repositories, *repo)
		time.Sleep(500 * time.Millisecond) // Be nice to the API
	}
	fmt.Printf("Collected metadata for %d repositories\n\n", len(repositories))

	// Collect organization metadata
	fmt.Printf("=== Collecting Organization Metadata ===\n")
	count = 0
	for orgName := range orgMap {
		count++
		fmt.Printf("Fetching [%d/%d] %s...\n", count, len(orgMap), orgName)

		org, err := m.CollectOrganizationMetadata(orgName)
		if err != nil {
			fmt.Printf("Warning: failed to fetch %s: %v\n", orgName, err)
			continue
		}

		organizations = append(organizations, *org)
		time.Sleep(500 * time.Millisecond) // Be nice to the API
	}
	fmt.Printf("Collected metadata for %d organizations\n\n", len(organizations))

	return repositories, organizations, nil
}
