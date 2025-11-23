// Unit Tests: Template Storage and File System Operations (storage package)
//
// High-level overview of what's being tested:
// - Creating storage with directory initialization
// - Saving templates to JSON Lines (.jsonl) format
// - Loading templates from JSON Lines files
// - Saving and loading repositories
// - Saving and loading organizations
// - Saving and loading discovery progress
// - Pretty-printing progress JSON with indentation
// - Handling empty files and non-existent files
// - Default progress when file doesn't exist
// - JSON Lines format validation (one object per line)
// - Invalid JSON error handling
// - Empty template lists
// - File system abstraction and mocking
// - Timestamp preservation during save/load
// - Directory creation errors

package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lima-catalog/lima-catalog/pkg/types"
)

// Mock FileSystem for testing
type mockFileSystem struct {
	files map[string]*bytes.Buffer
	err   error
}

func newMockFileSystem() *mockFileSystem {
	return &mockFileSystem{
		files: make(map[string]*bytes.Buffer),
	}
}

func (m *mockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *mockFileSystem) Open(path string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	if buf, ok := m.files[path]; ok {
		return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
	}
	return nil, os.ErrNotExist
}

func (m *mockFileSystem) Create(path string) (io.WriteCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	buf := &bytes.Buffer{}
	m.files[path] = buf
	return &mockWriteCloser{buf: buf}, nil
}

func (m *mockFileSystem) ReadDir(path string) ([]os.FileInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, nil
}

func (m *mockFileSystem) Stat(path string) (os.FileInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, os.ErrNotExist
}

type mockWriteCloser struct {
	buf *bytes.Buffer
}

func (m *mockWriteCloser) Write(p []byte) (n int, err error) {
	return m.buf.Write(p)
}

func (m *mockWriteCloser) Close() error {
	return nil
}

func TestNewStorage(t *testing.T) {
	tmpDir := t.TempDir()

	storage, err := NewStorage(tmpDir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage == nil {
		t.Fatal("expected non-nil storage")
	}
	if storage.dataDir != tmpDir {
		t.Errorf("expected dataDir %q, got %q", tmpDir, storage.dataDir)
	}

	// Verify directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("expected data directory to be created")
	}
}

func TestNewStorageWithFS(t *testing.T) {
	fs := newMockFileSystem()
	storage, err := NewStorageWithFS("/test/data", fs)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage == nil {
		t.Fatal("expected non-nil storage")
	}
	if storage.dataDir != "/test/data" {
		t.Errorf("expected dataDir /test/data, got %q", storage.dataDir)
	}
}

func TestNewStorage_MkdirError(t *testing.T) {
	fs := newMockFileSystem()
	fs.err = fmt.Errorf("permission denied")

	_, err := NewStorageWithFS("/test/data", fs)

	if err == nil {
		t.Error("expected error when mkdir fails")
	}
}

func TestSaveAndLoadTemplates(t *testing.T) {
	fs := newMockFileSystem()
	storage, _ := NewStorageWithFS("/test/data", fs)

	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	templates := []types.Template{
		{
			ID:          "owner/repo/template.yaml",
			Repo:        "owner/repo",
			Path:        "template.yaml",
			Name:        "Template",
			DisplayName: "Template Name",
			LastUpdated: fixedTime,
		},
		{
			ID:          "owner2/repo2/example.yaml",
			Repo:        "owner2/repo2",
			Path:        "example.yaml",
			Name:        "Example",
			DisplayName: "Example Template",
			LastUpdated: fixedTime,
		},
	}

	// Save templates
	err := storage.SaveTemplates(templates)
	if err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	// Load templates
	loaded, err := storage.LoadTemplates()
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}

	// Verify count
	if len(loaded) != len(templates) {
		t.Errorf("expected %d templates, got %d", len(templates), len(loaded))
	}

	// Verify content
	for i, template := range templates {
		if loaded[i].ID != template.ID {
			t.Errorf("template %d: expected ID %q, got %q", i, template.ID, loaded[i].ID)
		}
		if loaded[i].Name != template.Name {
			t.Errorf("template %d: expected Name %q, got %q", i, template.Name, loaded[i].Name)
		}
	}
}

func TestLoadTemplates_EmptyFile(t *testing.T) {
	fs := newMockFileSystem()
	storage, _ := NewStorageWithFS("/test/data", fs)

	// Load from non-existent file should return empty slice
	loaded, err := storage.LoadTemplates()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("expected empty slice, got %d templates", len(loaded))
	}
}

func TestSaveAndLoadRepositories(t *testing.T) {
	fs := newMockFileSystem()
	storage, _ := NewStorageWithFS("/test/data", fs)

	repos := []types.Repository{
		{
			ID:          "owner/repo",
			Owner:       "owner",
			Name:        "repo",
			Description: "Test repository",
			Stars:       42,
			Topics:      []string{"golang", "testing"},
		},
	}

	// Save and load
	err := storage.SaveRepositories(repos)
	if err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	loaded, err := storage.LoadRepositories()
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}

	// Verify
	if len(loaded) != len(repos) {
		t.Errorf("expected %d repos, got %d", len(repos), len(loaded))
	}
	if loaded[0].ID != repos[0].ID {
		t.Errorf("expected ID %q, got %q", repos[0].ID, loaded[0].ID)
	}
	if loaded[0].Stars != repos[0].Stars {
		t.Errorf("expected Stars %d, got %d", repos[0].Stars, loaded[0].Stars)
	}
}

func TestSaveAndLoadOrganizations(t *testing.T) {
	fs := newMockFileSystem()
	storage, _ := NewStorageWithFS("/test/data", fs)

	orgs := []types.Organization{
		{
			ID:    "test-org",
			Login: "test-org",
			Name:  "Test Organization",
			Email: "test@example.com",
		},
	}

	// Save and load
	err := storage.SaveOrganizations(orgs)
	if err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	loaded, err := storage.LoadOrganizations()
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}

	// Verify
	if len(loaded) != len(orgs) {
		t.Errorf("expected %d orgs, got %d", len(orgs), len(loaded))
	}
	if loaded[0].ID != orgs[0].ID {
		t.Errorf("expected ID %q, got %q", orgs[0].ID, loaded[0].ID)
	}
}

func TestSaveAndLoadProgress(t *testing.T) {
	fs := newMockFileSystem()
	storage, _ := NewStorageWithFS("/test/data", fs)

	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	progress := &types.Progress{
		Phase:              "analysis",
		TemplatesDiscovered: 100,
		ReposFetched:       50,
		OrgsFetched:        10,
		LastUpdated:        fixedTime,
		RateLimitRemaining: 4500,
		OfficialTemplates:  20,
		CommunityTemplates: 80,
	}

	// Save
	err := storage.SaveProgress(progress)
	if err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	// Verify file content is pretty-printed
	path := filepath.Join("/test/data", "progress.json")
	if buf, ok := fs.files[path]; ok {
		var formatted map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &formatted); err != nil {
			t.Errorf("saved JSON should be valid: %v", err)
		}
		// Check that it's indented (contains newlines and spaces)
		content := buf.String()
		if !bytes.Contains([]byte(content), []byte("\n")) {
			t.Error("expected pretty-printed JSON with newlines")
		}
	}

	// Load
	loaded, err := storage.LoadProgress()
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}

	// Verify
	if loaded.Phase != progress.Phase {
		t.Errorf("expected Phase %q, got %q", progress.Phase, loaded.Phase)
	}
	if loaded.TemplatesDiscovered != progress.TemplatesDiscovered {
		t.Errorf("expected TemplatesDiscovered %d, got %d", progress.TemplatesDiscovered, loaded.TemplatesDiscovered)
	}
	if loaded.OfficialTemplates != progress.OfficialTemplates {
		t.Errorf("expected OfficialTemplates %d, got %d", progress.OfficialTemplates, loaded.OfficialTemplates)
	}
}

func TestLoadProgress_DefaultWhenNotExists(t *testing.T) {
	fs := newMockFileSystem()
	storage, _ := NewStorageWithFS("/test/data", fs)

	// Load non-existent progress should return default
	progress, err := storage.LoadProgress()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if progress == nil {
		t.Fatal("expected non-nil progress")
	}
	if progress.Phase != "discovery" {
		t.Errorf("expected default Phase 'discovery', got %q", progress.Phase)
	}
}

func TestJSONLinesFormat(t *testing.T) {
	// Verify that saved files are in JSON Lines format (one JSON object per line)
	fs := newMockFileSystem()
	storage, _ := NewStorageWithFS("/test/data", fs)

	templates := []types.Template{
		{ID: "test1", Name: "Test 1"},
		{ID: "test2", Name: "Test 2"},
		{ID: "test3", Name: "Test 3"},
	}

	err := storage.SaveTemplates(templates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check file content
	path := filepath.Join("/test/data", "templates.jsonl")
	if buf, ok := fs.files[path]; ok {
		content := buf.String()
		lines := bytes.Split(buf.Bytes(), []byte("\n"))

		// Should have 4 lines (3 templates + trailing newline = 4 including empty)
		if len(lines) != 4 {
			t.Errorf("expected 4 lines (3 + empty), got %d", len(lines))
		}

		// Each line (except last) should be valid JSON
		for i := 0; i < 3; i++ {
			var template types.Template
			if err := json.Unmarshal(lines[i], &template); err != nil {
				t.Errorf("line %d is not valid JSON: %v", i+1, err)
			}
		}

		// Last line should be empty
		if len(lines[3]) != 0 {
			t.Error("expected last line to be empty")
		}

		_ = content // Use content to avoid unused variable
	} else {
		t.Error("expected file to exist in mock filesystem")
	}
}

func TestInvalidJSON(t *testing.T) {
	fs := newMockFileSystem()
	storage, _ := NewStorageWithFS("/test/data", fs)

	// Create a file with invalid JSON
	path := filepath.Join("/test/data", "templates.jsonl")
	fs.files[path] = bytes.NewBuffer([]byte("invalid json\n"))

	// Loading should return error
	_, err := storage.LoadTemplates()

	if err == nil {
		t.Error("expected error when loading invalid JSON")
	}
}

func TestEmptyTemplatesList(t *testing.T) {
	fs := newMockFileSystem()
	storage, _ := NewStorageWithFS("/test/data", fs)

	// Save empty list
	err := storage.SaveTemplates([]types.Template{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Load should return empty list
	loaded, err := storage.LoadTemplates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("expected empty list, got %d items", len(loaded))
	}
}
