package interfaces

import (
	"io"
	"net/http"
	"os"
	"time"
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
	defer file.Close()
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
