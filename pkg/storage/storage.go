// Package storage implements JSON Lines-based file storage for Lima catalog data.
//
// The package stores catalog data in JSON Lines format (one JSON object per line),
// which enables:
//   - Minimal git diffs (line-by-line changes)
//   - Streaming reads/writes for large datasets
//   - Easy processing with standard Unix tools (grep, awk, jq)
//
// Data Files:
//
//   - templates.jsonl: Template metadata (700+ templates)
//   - repos.jsonl: Repository information
//   - orgs.jsonl: Organization/user information
//   - progress.json: Collection state (single JSON object)
//
// The Storage type uses the FileSystem interface for testability, allowing
// file operations to be mocked in tests.
//
// Error Handling:
//
// Load functions return empty slices for non-existent files (first run).
// Save functions return errors for I/O failures or JSON marshaling errors.
// All errors are wrapped with context using fmt.Errorf with %w.
package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lima-catalog/lima-catalog/pkg/interfaces"
	"github.com/lima-catalog/lima-catalog/pkg/types"
)

// Storage handles reading and writing data in JSON Lines format
type Storage struct {
	dataDir string
	fs      interfaces.FileSystem
}

// NewStorage creates a new storage instance
func NewStorage(dataDir string) (*Storage, error) {
	return NewStorageWithFS(dataDir, interfaces.NewDefaultFileSystem())
}

// NewStorageWithFS creates a new storage instance with a custom FileSystem
func NewStorageWithFS(dataDir string, fs interfaces.FileSystem) (*Storage, error) {
	// Create data directory if it doesn't exist
	if err := fs.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &Storage{
		dataDir: dataDir,
		fs:      fs,
	}, nil
}

// LoadTemplates loads all templates from the JSON Lines file
func (s *Storage) LoadTemplates() ([]types.Template, error) {
	path := filepath.Join(s.dataDir, "templates.jsonl")
	return loadJSONLines[types.Template](s.fs, path)
}

// SaveTemplates saves templates to the JSON Lines file
func (s *Storage) SaveTemplates(templates []types.Template) error {
	path := filepath.Join(s.dataDir, "templates.jsonl")
	return saveJSONLines(s.fs, path, templates)
}

// LoadRepositories loads all repositories from the JSON Lines file
func (s *Storage) LoadRepositories() ([]types.Repository, error) {
	path := filepath.Join(s.dataDir, "repos.jsonl")
	return loadJSONLines[types.Repository](s.fs, path)
}

// SaveRepositories saves repositories to the JSON Lines file
func (s *Storage) SaveRepositories(repos []types.Repository) error {
	path := filepath.Join(s.dataDir, "repos.jsonl")
	return saveJSONLines(s.fs, path, repos)
}

// LoadOrganizations loads all organizations from the JSON Lines file
func (s *Storage) LoadOrganizations() ([]types.Organization, error) {
	path := filepath.Join(s.dataDir, "orgs.jsonl")
	return loadJSONLines[types.Organization](s.fs, path)
}

// SaveOrganizations saves organizations to the JSON Lines file
func (s *Storage) SaveOrganizations(orgs []types.Organization) error {
	path := filepath.Join(s.dataDir, "orgs.jsonl")
	return saveJSONLines(s.fs, path, orgs)
}

// LoadProgress loads the progress state
func (s *Storage) LoadProgress() (*types.Progress, error) {
	path := filepath.Join(s.dataDir, "progress.json")

	file, err := s.fs.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty progress if file doesn't exist
			return &types.Progress{
				Phase: "discovery",
			}, nil
		}
		return nil, fmt.Errorf("failed to open progress file: %w", err)
	}
	defer func() { _ = file.Close() }()

	var progress types.Progress
	if err := json.NewDecoder(file).Decode(&progress); err != nil {
		return nil, fmt.Errorf("failed to decode progress: %w", err)
	}

	return &progress, nil
}

// SaveProgress saves the progress state
func (s *Storage) SaveProgress(progress *types.Progress) error {
	path := filepath.Join(s.dataDir, "progress.json")

	file, err := s.fs.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create progress file: %w", err)
	}
	defer func() { _ = file.Close() }()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(progress); err != nil {
		return fmt.Errorf("failed to encode progress: %w", err)
	}

	return nil
}

// loadJSONLines loads data from a JSON Lines file
func loadJSONLines[T any](fs interfaces.FileSystem, path string) ([]T, error) {
	file, err := fs.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty slice if file doesn't exist
			return []T{}, nil
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	var items []T
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		var item T
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			return nil, fmt.Errorf("failed to decode line %d: %w", lineNum, err)
		}
		items = append(items, item)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return items, nil
}

// saveJSONLines saves data to a JSON Lines file
func saveJSONLines[T any](fs interfaces.FileSystem, path string, items []T) error {
	file, err := fs.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = file.Close() }()

	writer := bufio.NewWriter(file)
	defer func() { _ = writer.Flush() }()

	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to marshal item: %w", err)
		}

		if _, err := writer.Write(data); err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}

		if _, err := writer.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	return nil
}
