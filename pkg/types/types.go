// Package types defines core data structures for the Lima catalog.
//
// The package provides types for:
//   - Template: Lima VM template metadata
//   - Repository: GitHub repository information
//   - Organization: GitHub organization/user information
//   - Progress: Collection state tracking
//   - Blocklist: Template filtering rules
//   - NotabilityMetrics: Template quality scoring
//
// JSON Tags:
//
// All exported fields have JSON tags for serialization to JSON Lines format.
// Time fields use RFC3339 format for consistent parsing.
//
// Validation:
//
// The Blocklist type includes pattern compilation methods for performance.
// Call CompilePatterns() after loading from YAML to pre-compile regex patterns.
//
// This package has no external dependencies except the standard library.
package types

import (
	"fmt"
	"regexp"
	"time"
)

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
	LastChecked      time.Time `json:"last_checked"`               // Last time we verified it exists
	LastUpdated      time.Time `json:"last_updated"`               // Last time content changed (SHA changed)
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
	Notability       *NotabilityMetrics `json:"notability,omitempty"` // Metrics for calculating notability score
	MinHashSignature []uint32  `json:"minhash_signature,omitempty"` // MinHash signature for duplicate detection (128 hash values)
	SimilarTemplates []SimilarTemplate `json:"similar_templates,omitempty"` // Templates similar to this one (populated during duplicate detection)
	OriginalID       string    `json:"original_id,omitempty"`       // If this is a copy, the ID of the original template
}

// SimilarTemplate represents a template that is similar to another template
type SimilarTemplate struct {
	ID          string  `json:"id"`                     // Template ID (owner/repo/path)
	Similarity  float64 `json:"similarity"`             // Jaccard similarity (0.0-1.0)
	SharedBands int     `json:"shared_bands,omitempty"` // Number of LSH bands shared (for debugging)
	IsOriginal  bool    `json:"is_original,omitempty"`  // True if this is the original among exact duplicates
}

// IsExactDuplicate returns true if similarity indicates an exact duplicate (>90%)
func (s SimilarTemplate) IsExactDuplicate() bool {
	return s.Similarity > 0.9
}

// NotabilityMetrics contains raw observations used to calculate template notability score
// These are stored as raw data to allow weight adjustments without re-analysis
type NotabilityMetrics struct {
	MessageLength            int      `json:"message_length"`                      // Length of message in characters (>0 means template has message)
	MessageLineCount         int      `json:"message_line_count,omitempty"`        // Number of lines in message (for line-based bonus)
	ProvisionCount           int      `json:"provision_count"`                     // Number of provision scripts (total)
	ProvisionSubstantial     int      `json:"provision_substantial,omitempty"`     // Number of substantial provision scripts (>10 lines, capped at 3, min 1)
	ProvisionTotalLines      int      `json:"provision_total_lines"`               // Total unique lines across all provision scripts
	ProbeCount               int      `json:"probe_count"`                         // Number of probe scripts (total)
	ProbeSubstantial         int      `json:"probe_substantial,omitempty"`         // Number of substantial probe scripts (>10 lines, capped at 3, min 1)
	ProbeTotalLines          int      `json:"probe_total_lines"`                   // Total unique lines across all probe scripts
	ParamCount               int      `json:"param_count"`                         // Number of configurable parameters
	EnvCount                 int      `json:"env_count"`                           // Number of environment variables
	CommentLineCount         int      `json:"comment_line_count"`                  // Number of unique comment lines in template
	UnusualImages            []string `json:"unusual_images,omitempty"`            // Images not in official templates
	AllImages                []string `json:"all_images,omitempty"`                // All image locations (for org/repo name matching)
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

// Blocklist represents filtering rules for excluding templates
type Blocklist struct {
	// Paths contains regex patterns matched against file path within repo
	// Example: "^\.github/workflows/" blocks all GitHub Actions files
	Paths []string `yaml:"paths"`

	// Repos contains regex patterns matched against full org/repo/path
	// Example: "^spamorg/" blocks entire org, "^org/repo$" blocks specific repo
	Repos []string `yaml:"repos"`

	// Compiled regex patterns for performance (populated after loading)
	compiledPaths []*regexp.Regexp
	compiledRepos []*regexp.Regexp
}

// Progress tracks the state of data collection for resumability
type Progress struct {
	Phase               string    `json:"phase"`                // "discovery", "metadata", "complete"
	TemplatesDiscovered int       `json:"templates_discovered"` // total templates found
	ReposFetched        int       `json:"repos_fetched"`        // repos metadata collected
	OrgsFetched         int       `json:"orgs_fetched"`         // orgs metadata collected
	LastUpdated         time.Time `json:"last_updated"`         // last progress update
	RateLimitRemaining  int       `json:"rate_limit_remaining"` // remaining API calls
	RateLimitReset      time.Time `json:"rate_limit_reset"`     // when rate limit resets
	OfficialTemplates   int       `json:"official_templates"`   // official lima-vm/lima templates
	CommunityTemplates  int       `json:"community_templates"`  // community templates
}

// CompilePatterns pre-compiles all regex patterns in the blocklist for performance
// Returns an error if any pattern is invalid
func (b *Blocklist) CompilePatterns() error {
	// Compile path patterns
	b.compiledPaths = make([]*regexp.Regexp, 0, len(b.Paths))
	for _, pattern := range b.Paths {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid path pattern %q: %w", pattern, err)
		}
		b.compiledPaths = append(b.compiledPaths, re)
	}

	// Compile repo patterns
	b.compiledRepos = make([]*regexp.Regexp, 0, len(b.Repos))
	for _, pattern := range b.Repos {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid repo pattern %q: %w", pattern, err)
		}
		b.compiledRepos = append(b.compiledRepos, re)
	}

	return nil
}

// GetCompiledPaths returns the pre-compiled path patterns
func (b *Blocklist) GetCompiledPaths() []*regexp.Regexp {
	return b.compiledPaths
}

// GetCompiledRepos returns the pre-compiled repo patterns
func (b *Blocklist) GetCompiledRepos() []*regexp.Regexp {
	return b.compiledRepos
}
