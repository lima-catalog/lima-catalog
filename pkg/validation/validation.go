package validation

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidateGitHubToken validates that a token has a valid format
func ValidateGitHubToken(token string) error {
	if token == "" {
		return fmt.Errorf("GitHub token cannot be empty")
	}

	// GitHub tokens are typically 40+ characters
	if len(token) < 40 {
		return fmt.Errorf("GitHub token appears invalid (too short)")
	}

	// GitHub tokens should be alphanumeric with underscores
	// Classic tokens: ghp_... (40 chars after prefix)
	// Fine-grained tokens: github_pat_... (longer)
	validToken := regexp.MustCompile(`^(ghp_[a-zA-Z0-9]{36,}|github_pat_[a-zA-Z0-9_]{82}|[a-f0-9]{40})$`)
	if !validToken.MatchString(token) {
		return fmt.Errorf("GitHub token format appears invalid")
	}

	return nil
}

// ValidateTemplateID validates a template ID format (owner/repo/path)
func ValidateTemplateID(id string) error {
	if id == "" {
		return fmt.Errorf("template ID cannot be empty")
	}

	// Template ID should be owner/repo/path
	parts := strings.Split(id, "/")
	if len(parts) < 3 {
		return fmt.Errorf("template ID must be in format owner/repo/path, got: %s", id)
	}

	// Validate owner
	if parts[0] == "" {
		return fmt.Errorf("template owner cannot be empty")
	}

	// Validate repo
	if parts[1] == "" {
		return fmt.Errorf("template repo cannot be empty")
	}

	// Validate path (remaining parts)
	path := strings.Join(parts[2:], "/")
	if path == "" {
		return fmt.Errorf("template path cannot be empty")
	}

	// Path should end with .yaml or .yml
	if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
		return fmt.Errorf("template path must end with .yaml or .yml, got: %s", path)
	}

	return nil
}

// SanitizePath sanitizes a file path to prevent directory traversal
func SanitizePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Clean the path (removes .., ., etc.)
	cleaned := filepath.Clean(path)

	// Check for directory traversal attempts
	if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, "/../") {
		return "", fmt.Errorf("path contains directory traversal: %s", path)
	}

	// Check for absolute paths (we want relative paths only)
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("absolute paths not allowed: %s", path)
	}

	return cleaned, nil
}

// ValidateContextLines validates that context lines is non-negative
func ValidateContextLines(lines int) error {
	if lines < 0 {
		return fmt.Errorf("context lines must be non-negative, got: %d", lines)
	}

	// Reasonable upper bound to prevent excessive output
	if lines > 100 {
		return fmt.Errorf("context lines too large (max 100), got: %d", lines)
	}

	return nil
}

// ValidateMaxLength validates that a max length parameter is reasonable
func ValidateMaxLength(length int, name string) error {
	if length < 0 {
		return fmt.Errorf("%s must be non-negative, got: %d", name, length)
	}

	// Reasonable upper bound (10MB)
	const maxAllowed = 10 * 1024 * 1024
	if length > maxAllowed {
		return fmt.Errorf("%s too large (max %d), got: %d", name, maxAllowed, length)
	}

	return nil
}

// ValidateMaxFiles validates that a max files parameter is reasonable
func ValidateMaxFiles(count int) error {
	if count < 0 {
		return fmt.Errorf("max files must be non-negative, got: %d", count)
	}

	// Reasonable upper bound to prevent excessive processing
	if count > 1000 {
		return fmt.Errorf("max files too large (max 1000), got: %d", count)
	}

	return nil
}

// ValidateRepoIdentifier validates owner and repo name format
func ValidateRepoIdentifier(owner, repo string) error {
	if owner == "" {
		return fmt.Errorf("repository owner cannot be empty")
	}

	if repo == "" {
		return fmt.Errorf("repository name cannot be empty")
	}

	// GitHub allows alphanumeric, hyphens, and underscores
	validName := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	if !validName.MatchString(owner) {
		return fmt.Errorf("invalid owner name: %s", owner)
	}

	if !validName.MatchString(repo) {
		return fmt.Errorf("invalid repository name: %s", repo)
	}

	return nil
}
