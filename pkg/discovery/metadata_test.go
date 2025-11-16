package discovery

import (
	"testing"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

func TestSelectReposToRefresh(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		newTemplates    []types.Template
		existingRepos   []types.Repository
		expectedMin     int
		expectedMax     int
		checkContains   []string
		checkNotContain []string
	}{
		{
			name: "New template repos are always included",
			newTemplates: []types.Template{
				{Repo: "owner1/repo1"},
				{Repo: "owner2/repo2"},
			},
			existingRepos: []types.Repository{},
			expectedMin:   2,
			expectedMax:   2,
			checkContains: []string{"owner1/repo1", "owner2/repo2"},
		},
		{
			name:         "No new templates, no stale repos",
			newTemplates: []types.Template{},
			existingRepos: []types.Repository{
				{ID: "owner1/repo1", LastFetched: now.Add(-10 * 24 * time.Hour)}, // 10 days old
			},
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name:         "One stale repo (>30 days), should be selected",
			newTemplates: []types.Template{},
			existingRepos: []types.Repository{
				{ID: "owner1/repo1", LastFetched: now.Add(-40 * 24 * time.Hour)}, // 40 days old
			},
			expectedMin:   1,
			expectedMax:   1,
			checkContains: []string{"owner1/repo1"},
		},
		{
			name:         "Mix of fresh and stale repos, select 5%",
			newTemplates: []types.Template{},
			existingRepos: func() []types.Repository {
				repos := make([]types.Repository, 100)
				for i := 0; i < 100; i++ {
					age := 35 * 24 * time.Hour // All stale (>30 days)
					if i < 50 {
						age = 10 * 24 * time.Hour // Half are fresh (<30 days)
					}
					repos[i] = types.Repository{
						ID:          "owner/repo" + string(rune('0'+i)),
						LastFetched: now.Add(-age),
					}
				}
				return repos
			}(),
			expectedMin: 5,  // 5% of 100 = 5
			expectedMax: 5,  // Maximum 5% should be selected
		},
		{
			name: "New template repo already in stale list - should not duplicate",
			newTemplates: []types.Template{
				{Repo: "owner1/repo1"},
			},
			existingRepos: []types.Repository{
				{ID: "owner1/repo1", LastFetched: now.Add(-40 * 24 * time.Hour)}, // Stale
			},
			expectedMin:   1,
			expectedMax:   1,
			checkContains: []string{"owner1/repo1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SelectReposToRefresh(tt.newTemplates, tt.existingRepos)

			if len(result) < tt.expectedMin || len(result) > tt.expectedMax {
				t.Errorf("Expected %d-%d repos to refresh, got %d", tt.expectedMin, tt.expectedMax, len(result))
			}

			// Check that expected repos are in result
			resultMap := make(map[string]bool)
			for _, repo := range result {
				resultMap[repo] = true
			}

			for _, expected := range tt.checkContains {
				if !resultMap[expected] {
					t.Errorf("Expected repo %s to be in result, but it wasn't", expected)
				}
			}

			for _, notExpected := range tt.checkNotContain {
				if resultMap[notExpected] {
					t.Errorf("Did not expect repo %s to be in result, but it was", notExpected)
				}
			}
		})
	}
}

func TestSelectOrgsToRefresh(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		newTemplates    []types.Template
		existingOrgs    []types.Organization
		expectedMin     int
		expectedMax     int
		checkContains   []string
		checkNotContain []string
	}{
		{
			name: "New template orgs are always included",
			newTemplates: []types.Template{
				{Repo: "owner1/repo1"},
				{Repo: "owner2/repo2"},
			},
			existingOrgs: []types.Organization{},
			expectedMin:  2,
			expectedMax:  2,
			checkContains: []string{"owner1", "owner2"},
		},
		{
			name:         "No new templates, no stale orgs",
			newTemplates: []types.Template{},
			existingOrgs: []types.Organization{
				{ID: "owner1", LastFetched: now.Add(-10 * 24 * time.Hour)}, // 10 days old
			},
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name:         "One stale org (>30 days), should be selected",
			newTemplates: []types.Template{},
			existingOrgs: []types.Organization{
				{ID: "owner1", LastFetched: now.Add(-40 * 24 * time.Hour)}, // 40 days old
			},
			expectedMin:   1,
			expectedMax:   1,
			checkContains: []string{"owner1"},
		},
		{
			name: "Deduplicate orgs from same owner",
			newTemplates: []types.Template{
				{Repo: "owner1/repo1"},
				{Repo: "owner1/repo2"},
				{Repo: "owner1/repo3"},
			},
			existingOrgs: []types.Organization{},
			expectedMin:  1,
			expectedMax:  1,
			checkContains: []string{"owner1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SelectOrgsToRefresh(tt.newTemplates, tt.existingOrgs)

			if len(result) < tt.expectedMin || len(result) > tt.expectedMax {
				t.Errorf("Expected %d-%d orgs to refresh, got %d", tt.expectedMin, tt.expectedMax, len(result))
			}

			// Check that expected orgs are in result
			resultMap := make(map[string]bool)
			for _, org := range result {
				resultMap[org] = true
			}

			for _, expected := range tt.checkContains {
				if !resultMap[expected] {
					t.Errorf("Expected org %s to be in result, but it wasn't", expected)
				}
			}

			for _, notExpected := range tt.checkNotContain {
				if resultMap[notExpected] {
					t.Errorf("Did not expect org %s to be in result, but it was", notExpected)
				}
			}
		})
	}
}
