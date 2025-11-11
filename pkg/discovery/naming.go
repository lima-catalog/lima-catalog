package discovery

import (
	"path/filepath"
	"strings"
)

// DeriveTemplateName generates a descriptive name from the template path and repo
// Handles generic filenames like "lima.yaml" by using repository or path context
func DeriveTemplateName(path, repoFullName string) string {
	// Get the filename without extension
	filename := filepath.Base(path)
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	// If filename is descriptive (not generic), use it
	if !isGenericName(nameWithoutExt) {
		return sanitizeName(nameWithoutExt)
	}

	// For generic names like "lima", derive from context
	// Try parent directory name first
	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		parentDir := filepath.Base(dir)
		if !isGenericDir(parentDir) {
			return sanitizeName(parentDir)
		}
	}

	// Fall back to repository name
	parts := strings.Split(repoFullName, "/")
	if len(parts) == 2 {
		repoName := parts[1]
		return sanitizeName(repoName)
	}

	// Last resort: use full path
	return sanitizeName(strings.ReplaceAll(path, "/", "-"))
}

// isGenericName checks if a filename is too generic to be useful
func isGenericName(name string) bool {
	genericNames := map[string]bool{
		"lima":     true,
		"template": true,
		"config":   true,
		"default":  true,
		"example":  true,
		"test":     true,
	}
	return genericNames[strings.ToLower(name)]
}

// isGenericDir checks if a directory name is too generic
func isGenericDir(dir string) bool {
	genericDirs := map[string]bool{
		"templates": true,
		"configs":   true,
		"examples":  true,
		"lima":      true,
		".lima":     true,
		"vms":       true,
	}
	return genericDirs[strings.ToLower(dir)]
}

// sanitizeName cleans up a name for use as an identifier
func sanitizeName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace common separators with hyphens
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, ".", "-")

	// Remove leading/trailing hyphens
	name = strings.Trim(name, "-")

	// Collapse multiple hyphens
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}

	return name
}

// GenerateDisplayName creates a human-readable display name from a template name
func GenerateDisplayName(name string) string {
	// Replace hyphens with spaces
	displayName := strings.ReplaceAll(name, "-", " ")

	// Capitalize first letter of each word
	words := strings.Fields(displayName)
	for i, word := range words {
		if len(word) > 0 {
			// Keep acronyms uppercase (e.g., "k8s", "ci")
			if strings.ToUpper(word) == word || len(word) <= 3 {
				words[i] = strings.ToUpper(word)
			} else {
				words[i] = strings.ToUpper(string(word[0])) + word[1:]
			}
		}
	}

	return strings.Join(words, " ")
}
