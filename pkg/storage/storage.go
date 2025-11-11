package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

// Storage handles reading and writing data in JSON Lines format
type Storage struct {
	dataDir string
}

// NewStorage creates a new storage instance
func NewStorage(dataDir string) (*Storage, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &Storage{
		dataDir: dataDir,
	}, nil
}

// LoadTemplates loads all templates from the JSON Lines file
func (s *Storage) LoadTemplates() ([]types.Template, error) {
	path := filepath.Join(s.dataDir, "templates.jsonl")
	return loadJSONLines[types.Template](path)
}

// SaveTemplates saves templates to the JSON Lines file
func (s *Storage) SaveTemplates(templates []types.Template) error {
	path := filepath.Join(s.dataDir, "templates.jsonl")
	return saveJSONLines(path, templates)
}

// LoadRepositories loads all repositories from the JSON Lines file
func (s *Storage) LoadRepositories() ([]types.Repository, error) {
	path := filepath.Join(s.dataDir, "repos.jsonl")
	return loadJSONLines[types.Repository](path)
}

// SaveRepositories saves repositories to the JSON Lines file
func (s *Storage) SaveRepositories(repos []types.Repository) error {
	path := filepath.Join(s.dataDir, "repos.jsonl")
	return saveJSONLines(path, repos)
}

// LoadOrganizations loads all organizations from the JSON Lines file
func (s *Storage) LoadOrganizations() ([]types.Organization, error) {
	path := filepath.Join(s.dataDir, "orgs.jsonl")
	return loadJSONLines[types.Organization](path)
}

// SaveOrganizations saves organizations to the JSON Lines file
func (s *Storage) SaveOrganizations(orgs []types.Organization) error {
	path := filepath.Join(s.dataDir, "orgs.jsonl")
	return saveJSONLines(path, orgs)
}

// LoadProgress loads the progress state
func (s *Storage) LoadProgress() (*types.Progress, error) {
	path := filepath.Join(s.dataDir, "progress.json")

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty progress if file doesn't exist
			return &types.Progress{
				Phase: "discovery",
			}, nil
		}
		return nil, fmt.Errorf("failed to open progress file: %w", err)
	}
	defer file.Close()

	var progress types.Progress
	if err := json.NewDecoder(file).Decode(&progress); err != nil {
		return nil, fmt.Errorf("failed to decode progress: %w", err)
	}

	return &progress, nil
}

// SaveProgress saves the progress state
func (s *Storage) SaveProgress(progress *types.Progress) error {
	path := filepath.Join(s.dataDir, "progress.json")

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create progress file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(progress); err != nil {
		return fmt.Errorf("failed to encode progress: %w", err)
	}

	return nil
}

// loadJSONLines loads data from a JSON Lines file
func loadJSONLines[T any](path string) ([]T, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty slice if file doesn't exist
			return []T{}, nil
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

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
func saveJSONLines[T any](path string, items []T) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

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
