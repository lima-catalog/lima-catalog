package discovery

import (
	"fmt"
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

// CollectAllMetadata collects metadata for all unique repositories and organizations
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
