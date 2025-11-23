// Package interfaces defines abstractions for external dependencies to enable testing.
//
// The package provides interfaces for:
//
//   - HTTP operations (HTTPClient)
//   - File system operations (FileSystem)
//   - Time operations (Clock)
//   - URL transformation (URLTransformer for Lima's github: URLs)
//
// Each interface has a corresponding default implementation that wraps the standard
// library or external dependencies. Tests can provide mock implementations to avoid
// I/O operations and control time.
//
// Example usage in production code:
//
//	fs := interfaces.NewDefaultFileSystem()
//	storage, err := storage.NewStorageWithFS(dataDir, fs)
//
// Example usage in tests:
//
//	// Provide mock file system
//	mockFS := &MockFileSystem{
//	    files: make(map[string][]byte),
//	}
//	storage, err := storage.NewStorageWithFS("/test", mockFS)
package interfaces

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	limatmpl "github.com/lima-vm/lima/v2/pkg/limatmpl"
)

// HTTPClient provides an interface for making HTTP requests
// This allows mocking HTTP calls in tests
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// FileSystem provides an interface for file system operations
// This allows mocking file I/O in tests
type FileSystem interface {
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
	ReadDir(name string) ([]os.FileInfo, error)
	Stat(name string) (os.FileInfo, error)
	MkdirAll(path string, perm os.FileMode) error
}

// Clock provides an interface for time operations
// This allows controlling time in tests
type Clock interface {
	Now() time.Time
}

// DefaultHTTPClient wraps the standard http.Client
type DefaultHTTPClient struct {
	client *http.Client
}

// NewDefaultHTTPClient creates a new DefaultHTTPClient with a 30-second timeout
func NewDefaultHTTPClient() *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second, // Prevent hanging on unresponsive servers
		},
	}
}

// Get makes an HTTP GET request
func (c *DefaultHTTPClient) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}

// DefaultFileSystem wraps the standard os package
type DefaultFileSystem struct{}

// NewDefaultFileSystem creates a new DefaultFileSystem
func NewDefaultFileSystem() *DefaultFileSystem {
	return &DefaultFileSystem{}
}

// Open opens a file for reading
func (fs *DefaultFileSystem) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

// Create creates a file for writing
func (fs *DefaultFileSystem) Create(name string) (io.WriteCloser, error) {
	return os.Create(name)
}

// ReadDir reads directory contents
func (fs *DefaultFileSystem) ReadDir(name string) ([]os.FileInfo, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	return file.Readdir(-1)
}

// Stat returns file information
func (fs *DefaultFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// MkdirAll creates a directory and all parent directories
func (fs *DefaultFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// DefaultClock wraps the standard time package
type DefaultClock struct{}

// NewDefaultClock creates a new DefaultClock
func NewDefaultClock() *DefaultClock {
	return &DefaultClock{}
}

// Now returns the current time
func (c *DefaultClock) Now() time.Time {
	return time.Now()
}

// URLTransformer provides an interface for transforming github: URLs to https: URLs
// This allows mocking Lima's URL transformation in tests
type URLTransformer interface {
	TransformURL(ctx context.Context, url string) (string, error)
}

// DefaultURLTransformer wraps Lima's TransformCustomURL
type DefaultURLTransformer struct{}

// NewDefaultURLTransformer creates a new DefaultURLTransformer
func NewDefaultURLTransformer() *DefaultURLTransformer {
	return &DefaultURLTransformer{}
}

// TransformURL transforms a github: URL to an https: URL using Lima's library
func (t *DefaultURLTransformer) TransformURL(ctx context.Context, url string) (string, error) {
	return limatmpl.TransformCustomURL(ctx, url)
}
