package types

import "time"

// Template represents a Lima template file discovered on GitHub
type Template struct {
	ID               string    `json:"id"`                         // owner/repo/path/to/template.yaml
	Repo             string    `json:"repo"`                       // owner/repo
	Path             string    `json:"path"`                       // path/to/template.yaml
	SHA              string    `json:"sha"`                        // Git blob SHA
	Size             int       `json:"size"`                       // File size in bytes
	LastModified     time.Time `json:"last_modified"`              // Last commit date
	URL              string    `json:"url"`                        // Raw content URL
	DiscoveredAt     time.Time `json:"discovered_at"`              // When we found it
	LastChecked      time.Time `json:"last_checked"`               // Last time we verified it
	IsOfficial       bool      `json:"is_official"`                // From lima-vm/lima
	Name             string    `json:"name,omitempty"`             // Short name (e.g., "ubuntu-dev")
	DisplayName      string    `json:"display_name,omitempty"`     // Human-readable name
	ShortDescription string    `json:"short_description,omitempty"`// 1-2 sentence summary
	Description      string    `json:"description,omitempty"`      // Detailed description
	Category         string    `json:"category,omitempty"`         // Primary category (development, testing, etc.)
	UseCase          string    `json:"use_case,omitempty"`         // Specific use case
	Keywords         []string  `json:"keywords,omitempty"`         // Tags for searching
	Images           []string  `json:"images,omitempty"`           // OS images used
	Arch             []string  `json:"arch,omitempty"`             // Architectures supported
	AnalyzedAt       time.Time `json:"analyzed_at,omitempty"`      // When analysis was performed
}

// Repository represents a GitHub repository containing templates
type Repository struct {
	ID            string    `json:"id"`             // owner/repo
	Owner         string    `json:"owner"`          // owner login
	Name          string    `json:"name"`           // repo name
	Description   string    `json:"description"`    // repo description
	Topics        []string  `json:"topics"`         // repo topics/keywords
	Stars         int       `json:"stars"`          // stargazers count
	Forks         int       `json:"forks"`          // forks count
	Watchers      int       `json:"watchers"`       // watchers count
	Language      string    `json:"language"`       // primary language
	License       string    `json:"license"`        // license SPDX ID
	DefaultBranch string    `json:"default_branch"` // default branch name (e.g., "main", "master")
	CreatedAt     time.Time `json:"created_at"`     // repo creation date
	UpdatedAt     time.Time `json:"updated_at"`     // last update date
	PushedAt      time.Time `json:"pushed_at"`      // last push date
	Homepage      string    `json:"homepage"`       // homepage URL
	IsFork        bool      `json:"is_fork"`        // is this a fork?
	Parent        string    `json:"parent"`         // parent repo if fork (owner/repo)
	LastFetched   time.Time `json:"last_fetched"`   // when we fetched this data
}

// Organization represents a GitHub user or organization
type Organization struct {
	ID          string    `json:"id"`           // login
	Login       string    `json:"login"`        // username/org name
	Type        string    `json:"type"`         // "User" or "Organization"
	Name        string    `json:"name"`         // display name
	Description string    `json:"description"`  // bio/description
	Location    string    `json:"location"`     // location
	Blog        string    `json:"blog"`         // website URL
	Email       string    `json:"email"`        // public email
	LastFetched time.Time `json:"last_fetched"` // when we fetched this data
}

// Progress tracks the state of data collection for resumability
type Progress struct {
	Phase                string    `json:"phase"`                  // "discovery", "metadata", "complete"
	LastSearchCursor     string    `json:"last_search_cursor"`     // pagination cursor
	TemplatesDiscovered  int       `json:"templates_discovered"`   // total templates found
	ReposFetched         int       `json:"repos_fetched"`          // repos metadata collected
	OrgsFetched          int       `json:"orgs_fetched"`           // orgs metadata collected
	LastUpdated          time.Time `json:"last_updated"`           // last progress update
	RateLimitRemaining   int       `json:"rate_limit_remaining"`   // remaining API calls
	RateLimitReset       time.Time `json:"rate_limit_reset"`       // when rate limit resets
	TemplatesFetched     int       `json:"templates_fetched"`      // templates with full metadata
	OfficialTemplates    int       `json:"official_templates"`     // official lima-vm/lima templates
	CommunityTemplates   int       `json:"community_templates"`    // community templates
}
