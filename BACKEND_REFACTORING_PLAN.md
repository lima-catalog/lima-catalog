# Backend Refactoring Plan

**Date**: 2025-11-18
**Status**: Phase 1 In Progress

This document outlines a comprehensive refactoring plan for the lima-catalog backend Go codebase based on a thorough analysis of all 21 Go files (~4,500 lines).

---

## 📊 Current State Analysis

**Backend Code**: 21 Go files
- 12 source files
- 6 test files (blocklist, metadata, update, notability, combiner, prompt builder)
- 3 entry point files

**Architecture**:
- **cmd/lima-catalog** - Main data collection orchestrator
- **cmd/prompt-generator** - LLM prompt generation CLI
- **pkg/types** - Core data structures
- **pkg/github** - GitHub API client wrapper
- **pkg/storage** - JSON Lines file I/O
- **pkg/discovery** - Discovery, metadata, analysis (8 files, ~2000 lines)
- **pkg/combiner** - Frontend data generation
- **pkg/prompt** - LLM prompt building

---

## 🔴 Critical Issues Identified

### 1. God Function Anti-Pattern
**Location**: `cmd/lima-catalog/main.go:24-424`

The `run()` function is 400+ lines and handles everything:
- Environment variable parsing
- Storage initialization
- Progress loading
- GitHub client creation
- Rate limit checking
- Template discovery
- Metadata collection
- Template analysis
- Data combination
- Progress updates
- Error handling for all of the above

### 2. Missing Test Coverage
No tests for critical code paths:
- ❌ `pkg/discovery/discovery.go` (core discovery logic!)
- ❌ `pkg/discovery/analyzer.go` (template analysis!)
- ❌ `pkg/discovery/parser.go` (template parsing!)
- ❌ `pkg/discovery/naming.go`
- ❌ `pkg/storage/storage.go`
- ❌ `pkg/github/client.go`

### 3. Duplicated Rate Limit Handling
Rate limit code is copy-pasted across:
- `pkg/discovery/discovery.go:76-91`
- `pkg/discovery/metadata.go` (similar patterns)

### 4. Magic Numbers Everywhere
**Sleep durations**:
- `3 * time.Second` - discovery.go:146
- `5 * time.Second` - discovery.go:186, 209, 232
- `500 * time.Millisecond` - metadata.go:250, 270, 326, 344; analyzer.go:236

**Rate limit thresholds** (main.go:97, 103):
- `if core.Remaining < 100`
- `if search.Remaining < 5`

**Notability weights** (notability.go:80-143): All scoring weights hard-coded

### 5. Silent Failures / Poor Error Tracking
- Template validation failures ignored (discovery.go:118-121)
- LLM enhancement failures logged but not tracked (analyzer.go:102-106)
- Metadata fetch failures lost (metadata.go:243-246, 264-267)
- No structured logging or failure metrics

### 6. Complex Functions Need Breakdown
- `DiscoverCommunityTemplates` (discovery.go:161-262) - 100+ lines
- `FormatPrompt` (builder.go:121-265) - 143 lines
- `MergeTemplates` (update.go:47-109) - Nested conditions

---

## ⚠️ Medium Priority Issues

### 7. Poor Testability
- Direct API calls in business logic (no interfaces for mocking)
- Hard-coded `time.Now()` calls (update.go:52, 53, 59, 79, 80)
- HTTP calls without dependency injection (parser.go:82)
- File I/O not abstracted (storage.go)
- Git clone in production code (builder.go:312)

### 8. Missing Input Validation
- No GITHUB_TOKEN validation beyond empty check (main.go:31-34)
- No path sanitization (potential path traversal)
- No bounds checking on config (ContextLines, MaxReadmeLength)
- Template ID format not validated before `strings.Split()`

### 9. Inconsistent Patterns
- Mixed naming: `Collect` vs `Fetch` vs `Get`
- Inconsistent error handling: Some return errors, others log and continue
- Mixed logging: `fmt.Printf` vs `fmt.Fprintf(os.Stderr, ...)`
- Inconsistent nil checks

---

## 🟢 Low Priority Issues

### 10. Dead Code
- `types.Progress.LastSearchCursor` (types.go:94) - never used
- `types.Progress.TemplatesFetched` (types.go:101) - set but never read
- Commented future code (combiner.go:81-84, 167-170)
- Empty stub: `enhanceWithLLM` (analyzer.go:193-202)

### 11. Performance Opportunities
- No batching of API calls
- No caching
- Full file rewrites on every storage save
- Inefficient grep (clones entire repo)

### 12. Security Concerns
- No timeout on git clone (builder.go:312)
- Token potentially in error messages
- No input sanitization on template paths

---

## 📋 Three-Phase Refactoring Plan

### **Phase 1: Critical Fixes** (Estimated: 2-3 days)

**Goal**: Fix the most critical code smells that affect maintainability and reliability.

#### 1.1. Break up `main.go:run()` function
Extract responsibilities into focused functions:
```go
func setupEnvironment() (*Config, error)
func initializeStorage(cfg *Config) (*Storage, error)
func initializeClient(token string) (*github.Client, error)
func checkRateLimits(client *github.Client) error
func runDiscoveryPhase(client *github.Client, storage *Storage, progress *types.Progress) error
func runMetadataPhase(client *github.Client, storage *Storage, progress *types.Progress) error
func runAnalysisPhase(client *github.Client, storage *Storage, progress *types.Progress) error
func runCombinePhase(storage *Storage) error
```

#### 1.2. Move magic numbers to constants/config
Create `pkg/config/constants.go`:
```go
const (
    // API delay constants
    SearchAPIPaginationDelay = 3 * time.Second
    SearchAPIQueryDelay = 5 * time.Second
    MetadataAPIDelay = 500 * time.Millisecond

    // Rate limit thresholds
    MinCoreRateLimitRemaining = 100
    MinSearchRateLimitRemaining = 5
)

type NotabilityWeights struct {
    Message       float64
    ProvisionBase float64
    ProvisionLine float64
    // ...
}
```

#### 1.3. Extract rate limit handling
Add to `pkg/github/client.go`:
```go
func (c *Client) WaitForRateLimit(limitType string) error
func (c *Client) CheckRateLimits(minCore, minSearch int) error
```

#### 1.4. Add tests for `discovery.go`
Test cases:
- `searchWithQuery` pagination handling
- `isLimaTemplate` validation logic
- `DiscoverCommunityTemplates` query building
- Rate limit retry logic

#### 1.5. Add tests for `analyzer.go`
Test cases:
- `AnalyzeTemplate` field population
- `inferCategory` logic
- `generateDescription` output
- Error handling paths

#### 1.6. Add tests for `parser.go`
Test cases:
- `ParseYAML` with various template formats
- `extractImages` parsing
- `ParseMarkdown` header extraction
- Error handling for malformed input

#### 1.7. Add structured logging
Replace `fmt.Printf` with structured logging (e.g., `log/slog`):
- Consistent log levels (Debug, Info, Warn, Error)
- Structured fields for better filtering
- Centralized logger configuration

---

### **Phase 2: Improve Quality** (Estimated: 3-4 days)

**Goal**: Improve testability, maintainability, and code quality.

#### 2.1. Introduce interfaces for testability
```go
// pkg/interfaces/interfaces.go
type HTTPClient interface {
    Get(url string) (*http.Response, error)
}

type FileSystem interface {
    Open(name string) (io.ReadCloser, error)
    Create(name string) (io.WriteCloser, error)
    ReadDir(name string) ([]os.FileInfo, error)
}

type Clock interface {
    Now() time.Time
}
```

Inject dependencies instead of using globals.

#### 2.2. Simplify `MergeTemplates` logic
Extract helper functions:
```go
func shouldUpdateTemplate(existing, new *Template) bool
func backfillTimestamps(template *Template)
func migrateAnalyzedAt(template *Template)
```

#### 2.3. Extract `FormatPrompt` sections
Break into helper functions:
```go
func writeHeader(buf *bytes.Buffer)
func writeTemplateInfo(buf *bytes.Buffer, ctx *PromptContext)
func writeRepositoryInfo(buf *bytes.Buffer, ctx *PromptContext)
func writeReferences(buf *bytes.Buffer, ctx *PromptContext)
func writeInstructions(buf *bytes.Buffer)
```

#### 2.4. Add input validation
- Validate GITHUB_TOKEN format
- Sanitize file paths
- Validate config bounds (ContextLines >= 0, etc.)
- Validate template ID format before parsing

#### 2.5. Implement retry logic with exponential backoff
Create `pkg/retry/retry.go`:
```go
func WithExponentialBackoff(fn func() error, maxRetries int) error
```

Use for:
- API calls
- File I/O
- Git operations

---

### **Phase 3: Polish** (Estimated: 2-3 days)

**Goal**: Clean up dead code, optimize performance, and add security hardening.

#### 3.1. Remove dead code
- Remove `types.Progress.LastSearchCursor`
- Remove `types.Progress.TemplatesFetched` or use it
- Remove commented code in combiner.go
- Implement or remove `enhanceWithLLM` stub

#### 3.2. Optimize API batching
- Batch repository metadata fetches
- Cache API responses within a run
- Use GraphQL API for complex queries

#### 3.3. Add caching layer
Create `pkg/cache/cache.go`:
- In-memory cache for API responses
- TTL-based expiration
- Cache key derivation

#### 3.4. Standardize naming conventions
- Use consistent verb prefixes: `Get`, `Fetch`, `Load`
- Standardize error handling patterns
- Unify logging approach

#### 3.5. Add timeouts and security hardening
- Add timeout to git clone operations
- Scrub tokens from error messages
- Add input sanitization for template paths
- Add rate limiting on outbound requests

---

## 📈 Success Metrics

After completing all phases:

✅ **Test Coverage**: >80% for critical paths
✅ **Function Length**: No functions >100 lines
✅ **Cyclomatic Complexity**: <10 per function
✅ **Code Duplication**: <5% duplicate code
✅ **Magic Numbers**: 0 (all constants defined)
✅ **Dead Code**: 0 unused fields/functions
✅ **Structured Logging**: 100% (no fmt.Printf)

---

## 🔄 Progress Tracking

### Phase 1: Critical Fixes
- [x] 1.1. Break up main.go run() function (400+ lines → 56 lines)
- [x] 1.2. Move magic numbers to constants (created pkg/config/constants.go)
- [x] 1.3. Extract rate limit handling (added Client.HandleRateLimitError() method)
- [x] 1.4. Add tests for parser.go (comprehensive coverage - 5 test functions, 30+ cases)
- [ ] 1.5. Add tests for discovery.go (deferred - requires complex mocking)
- [ ] 1.6. Add tests for analyzer.go (deferred - requires complex mocking)
- [ ] 1.7. Add structured logging (deferred to Phase 2 or later)

**Phase 1 Status**: 4/7 complete - Core refactoring done, deferred items moved to future phases

### Phase 2: Improve Quality
- [x] 2.1. Introduce interfaces for testability (HTTPClient, FileSystem, Clock)
  - Created `pkg/interfaces/interfaces.go` with HTTPClient, FileSystem, Clock interfaces
  - Added default implementations: DefaultHTTPClient, DefaultFileSystem, DefaultClock
  - Updated `ParseTemplate()` to accept HTTPClient parameter for testability
  - Updated `Analyzer` struct to include HTTPClient and Clock fields
  - Updated `Analyzer.AnalyzeTemplate()` to use Clock.Now() instead of time.Now()
  - Updated `MergeTemplates()` to accept Clock parameter
  - Updated `Storage` to use FileSystem interface for all file operations
  - Added `NewStorageWithFS()` constructor for dependency injection in tests
  - All tests pass (83/83)
- [x] 2.2. Simplify MergeTemplates logic
  - Extracted `backfillLastUpdated()` - eliminates duplicated timestamp migration logic
  - Extracted `processUpdatedTemplate()` - handles SHA-changed templates
  - Extracted `processUnchangedTemplate()` - handles SHA-same templates that were checked
  - Extracted `processNewTemplate()` - handles newly discovered templates
  - Extracted `processUncheckedTemplate()` - handles templates not checked this run
  - Main MergeTemplates function reduced from 89 lines to 59 lines (34% reduction)
  - Each processing path is now clear and testable in isolation
  - All tests pass (83/83)
- [ ] 2.3. Extract FormatPrompt sections
- [ ] 2.4. Add input validation
- [ ] 2.5. Implement retry logic

**Phase 2 Status**: 2/5 complete

### Phase 3: Polish
- [ ] 3.1. Remove dead code
- [ ] 3.2. Optimize API batching
- [ ] 3.3. Add caching layer
- [ ] 3.4. Standardize naming conventions
- [ ] 3.5. Add timeouts and security hardening

---

## 💡 Key Principles

Throughout refactoring, follow these principles:

1. **Single Responsibility**: Each function does one thing well
2. **Don't Repeat Yourself**: Extract common patterns
3. **Test First**: Add tests before refactoring
4. **Incremental Changes**: Small, reviewable commits
5. **Backward Compatibility**: Don't break existing workflows
6. **Performance Awareness**: Don't sacrifice performance for style

---

## 🚀 Getting Started

To begin Phase 1:

```bash
# Ensure all current tests pass
make test

# Create feature branch
git checkout -b refactor/phase-1-critical-fixes

# Start with 1.1 (breaking up main.go)
# Make small commits for each sub-task
# Run tests after each change
# Update this document's progress tracking
```

---

## 📚 References

- Go Code Review Comments: https://github.com/golang/go/wiki/CodeReviewComments
- Effective Go: https://go.dev/doc/effective_go
- Go Testing Best Practices: https://go.dev/blog/table-driven-tests
- Clean Code principles (adapted for Go)
