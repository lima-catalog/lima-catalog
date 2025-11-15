package discovery

import (
	"fmt"
	"os"
	"regexp"

	"github.com/lima-catalog/lima-catalog/pkg/types"
	"gopkg.in/yaml.v3"
)

// LoadBlocklist loads the blocklist from a YAML file
func LoadBlocklist(path string) (*types.Blocklist, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// If blocklist doesn't exist, return empty blocklist (no filtering)
		if os.IsNotExist(err) {
			return &types.Blocklist{
				Paths: []string{},
				Repos: []string{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read blocklist: %w", err)
	}

	var blocklist types.Blocklist
	if err := yaml.Unmarshal(data, &blocklist); err != nil {
		return nil, fmt.Errorf("failed to parse blocklist: %w", err)
	}

	return &blocklist, nil
}

// IsBlocklisted checks if a template should be excluded based on blocklist rules
func IsBlocklisted(owner, repo, path string, blocklist *types.Blocklist) bool {
	if blocklist == nil {
		return false
	}

	fullPath := owner + "/" + repo + "/" + path

	// Check repo patterns (matches against full org/repo/path)
	for _, pattern := range blocklist.Repos {
		matched, err := regexp.MatchString(pattern, fullPath)
		if err != nil {
			// Invalid regex, skip this pattern
			continue
		}
		if matched {
			return true
		}
	}

	// Check path patterns (matches against path within repo)
	for _, pattern := range blocklist.Paths {
		matched, err := regexp.MatchString(pattern, path)
		if err != nil {
			// Invalid regex, skip this pattern
			continue
		}
		if matched {
			return true
		}
	}

	return false
}
