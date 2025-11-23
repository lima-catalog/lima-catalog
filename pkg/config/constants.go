// Package config provides configuration constants and default values for the Lima catalog.
//
// The package defines:
//
//   - API rate limiting thresholds for GitHub API calls
//   - Delay constants for search and metadata API operations
//   - Concurrency limits for parallel operations
//   - Notability scoring weights for template ranking
//
// Notability Scoring:
//
// Templates are ranked using weighted scores based on features like provisioning scripts,
// documentation, repository stars, and validation warnings. Higher scores indicate more
// notable/interesting templates. Use DefaultNotabilityWeights() to get the standard
// configuration.
package config

import "time"

// API delay constants
const (
	// SearchAPIPaginationDelay is the delay between pages when paginating search results
	SearchAPIPaginationDelay = 3 * time.Second

	// SearchAPIQueryDelay is the delay between different search queries
	SearchAPIQueryDelay = 5 * time.Second

	// MetadataAPIDelay is the delay between metadata API calls
	MetadataAPIDelay = 500 * time.Millisecond

	// MaxMetadataConcurrency is the maximum number of concurrent metadata API calls
	// Allows parallel fetching while respecting rate limits
	MaxMetadataConcurrency = 5
)

// Rate limit threshold constants
const (
	// MinCoreRateLimitRemaining is the minimum required core API calls before proceeding
	MinCoreRateLimitRemaining = 100

	// MinSearchRateLimitRemaining is the minimum required search API calls before proceeding
	MinSearchRateLimitRemaining = 5
)

// NotabilityWeights defines the scoring weights for template notability
type NotabilityWeights struct {
	// Message points if template has a user-facing message
	Message float64

	// ProvisionBase points per provision script
	ProvisionBase float64

	// ProvisionLine points per 10 lines in provision scripts
	ProvisionLine float64

	// ProbeBase points per probe script
	ProbeBase float64

	// ProbeLine points per 10 lines in probe scripts
	ProbeLine float64

	// Parameter points per configurable parameter
	Parameter float64

	// EnvVar points per environment variable
	EnvVar float64

	// UnusualImage points if template uses unusual image domains
	UnusualImage float64

	// CommentLine points per unique comment line
	CommentLine float64

	// ValidationWarningFirstPenalty penalty for the first validation warning
	ValidationWarningFirstPenalty float64

	// ValidationWarningExtraPenalty penalty per additional validation warning after the first
	ValidationWarningExtraPenalty float64

	// StarsPerPoint number of stars required for 1 point (capped at 50 points)
	StarsPerPoint float64

	// MaxStarsPoints maximum points from repository stars
	MaxStarsPoints float64

	// CustomDriver bonus for custom VM drivers
	CustomDriver float64
}

// DefaultNotabilityWeights returns the default scoring weights
func DefaultNotabilityWeights() NotabilityWeights {
	return NotabilityWeights{
		Message:                       50.0, // Base bonus for user-facing message
		ProvisionBase:                 10.0, // Points per provision script
		ProvisionLine:                 0.1,  // Points per line (10 lines = 1 point)
		ProbeBase:                     5.0,  // Points per probe script
		ProbeLine:                     0.1,  // Points per line (10 lines = 1 point)
		Parameter:                     20.0, // Indicates configurability
		EnvVar:                        10.0, // Shows configuration effort
		UnusualImage:                  30.0, // Interesting unusual images
		CommentLine:                   2.0,  // Documentation quality
		ValidationWarningFirstPenalty: 50.0, // Penalty for first warning (subtracted from score)
		ValidationWarningExtraPenalty: 10.0, // Penalty per additional warning (subtracted from score)
		StarsPerPoint:                 10.0, // 10 stars = 1 point
		MaxStarsPoints:                50.0, // Cap star contribution
		CustomDriver:                  70.0, // Custom VM driver bonus
	}
}
