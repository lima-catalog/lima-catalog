# Backend Test Coverage Improvement Plan

**Current Overall Coverage: 45.9%**
**Target Coverage: 70%+**

Last Updated: 2025-11-22

## Coverage Status by Package

### ✅ Excellent Coverage (80%+)
- [x] `pkg/validation` - 98.6% - Nearly complete
- [x] `pkg/combiner` - 94.5% - Very good
- [x] `pkg/minhash` - 93.4% - Very good
- [x] `pkg/cache` - 88.1% - Good (missing: StartCleanupTimer)
- [x] `pkg/storage` - 86.7% - Good
- [x] `pkg/retry` - 83.1% - Good

### ⚠️ Medium Coverage (30-70%)
- [ ] `pkg/discovery` - 48.4% - Mixed coverage
- [ ] `pkg/github` - 36.5% - Needs improvement

### ❌ Low/No Coverage (<30%)
- [ ] `pkg/prompt` - 23.1% - Needs significant work
- [ ] `pkg/config` - 0.0% - Constants only (low priority)
- [ ] `pkg/types` - 0.0% - Helper methods untested
- [ ] `pkg/interfaces` - 0.0% - Interface wrappers (low priority)
- [ ] `cmd/lima-catalog` - 0.0% - Main application
- [ ] `cmd/debug-template` - 0.0% - CLI tool
- [ ] `cmd/prompt-generator` - 0.0% - CLI tool

---

## Priority 1: Critical Business Logic (High Impact)

### 1. pkg/discovery/discovery.go
**Current Coverage: 0% for main functions**
**Target: 80%+**

#### Missing Test Coverage
- [ ] `NewDiscoverer` - Constructor
- [ ] `FindNewestTemplateTimestamp` - Utility function
- [ ] `isLimaTemplate` - Template validation logic
- [ ] `searchWithQuery` - GitHub search with pagination
- [ ] `DiscoverCommunityTemplates` - Community template discovery
- [ ] `DiscoverOfficialTemplates` - Official template discovery
- [ ] `DiscoverAll` - Main discovery orchestration

#### Test Cases Needed (15-20)
- [ ] Test NewDiscoverer creates valid instance
- [ ] Test FindNewestTemplateTimestamp with empty list
- [ ] Test FindNewestTemplateTimestamp finds newest
- [ ] Test isLimaTemplate validates templates correctly
- [ ] Test isLimaTemplate rejects non-Lima files
- [ ] Test searchWithQuery handles pagination
- [ ] Test searchWithQuery applies blocklist filtering
- [ ] Test searchWithQuery handles rate limits
- [ ] Test searchWithQuery handles context cancellation
- [ ] Test DiscoverCommunityTemplates deduplication
- [ ] Test DiscoverCommunityTemplates incremental mode
- [ ] Test DiscoverCommunityTemplates with date filter
- [ ] Test DiscoverOfficialTemplates lists templates
- [ ] Test DiscoverOfficialTemplates incremental mode
- [ ] Test DiscoverOfficialTemplates detects changes
- [ ] Test DiscoverAll orchestration
- [ ] Test error handling for API failures

**Estimated Effort:** 2-3 days

---

### 2. pkg/discovery/metadata.go
**Current Coverage: 0% for main functions (96%+ for helper functions)**
**Target: 80%+**

#### Missing Test Coverage
- [ ] `NewMetadataCollector` - Constructor
- [ ] `CollectRepositoryMetadata` - Single repo fetch
- [ ] `CollectOrganizationMetadata` - Single org fetch
- [ ] `fetchRepositoriesConcurrent` - Concurrent fetching
- [ ] `fetchOrganizationsConcurrent` - Concurrent fetching
- [ ] `CollectMetadataIncremental` - Incremental refresh
- [ ] `CollectAllMetadata` - Full metadata collection

#### Test Cases Needed (12-15)
- [ ] Test NewMetadataCollector initialization
- [ ] Test CollectRepositoryMetadata success
- [ ] Test CollectRepositoryMetadata invalid repo name
- [ ] Test CollectRepositoryMetadata API failure
- [ ] Test CollectOrganizationMetadata success
- [ ] Test CollectOrganizationMetadata API failure
- [ ] Test fetchRepositoriesConcurrent concurrency limit
- [ ] Test fetchRepositoriesConcurrent rate limiting
- [ ] Test fetchRepositoriesConcurrent error handling
- [ ] Test fetchOrganizationsConcurrent same scenarios
- [ ] Test CollectMetadataIncremental refresh strategy
- [ ] Test CollectAllMetadata from scratch
- [ ] Test thread safety of concurrent operations

**Estimated Effort:** 2-3 days

---

### 3. pkg/github/client.go
**Current Coverage: 36.5%**
**Target: 80%+**

#### Missing Test Coverage
- [ ] `RateLimit` - Get rate limit status
- [ ] `CheckRateLimit` - Verify sufficient quota
- [ ] `SearchCode` - Code search API
- [ ] `ListRepositoryContents` - Directory listing
- [ ] `GetRepositoryContent` - File content fetch

#### Test Cases Needed (8-10)
- [ ] Test RateLimit retrieves limits
- [ ] Test CheckRateLimit with sufficient quota
- [ ] Test CheckRateLimit below threshold
- [ ] Test CheckRateLimit error handling
- [ ] Test SearchCode with pagination
- [ ] Test SearchCode with query options
- [ ] Test ListRepositoryContents success
- [ ] Test ListRepositoryContents error handling
- [ ] Test GetRepositoryContent success
- [ ] Test GetRepositoryContent error handling

**Estimated Effort:** 1-2 days

---

### 4. pkg/prompt/builder.go
**Current Coverage: 23.1%**
**Target: 70%+**

#### Missing Test Coverage
- [ ] `NewBuilder` - Constructor with validation
- [ ] `BuildPrompt` - Main prompt generation
- [ ] `GatherContext` - Context collection
- [ ] `fetchTemplateContent` - Template download
- [ ] `fetchReadme` - README fetching
- [ ] `findTemplateReferences` - Git clone and grep
- [ ] `convertGitHubRepo` - Type conversion
- [ ] `convertGitHubUser` - Type conversion
- [ ] `Validate` (types.go) - Config validation

#### Test Cases Needed (12-15)
- [ ] Test NewBuilder with valid config
- [ ] Test NewBuilder with invalid token
- [ ] Test NewBuilder with nil config (uses default)
- [ ] Test BuildPrompt end-to-end
- [ ] Test GatherContext collects all data
- [ ] Test fetchTemplateContent success
- [ ] Test fetchReadme tries multiple filenames
- [ ] Test findTemplateReferences parsing
- [ ] Test template rendering with context
- [ ] Test error handling for missing files
- [ ] Test README length truncation
- [ ] Test type conversions
- [ ] Test config validation

**Estimated Effort:** 2-3 days

---

## Priority 2: Helper Functions (Medium Impact)

### 5. pkg/types/types.go
**Current Coverage: 0%**
**Target: 80%+**

#### Missing Test Coverage
- [ ] `IsExactDuplicate` - Similarity threshold
- [ ] `CompilePatterns` - Regex compilation
- [ ] `GetCompiledPaths` - Getter
- [ ] `GetCompiledRepos` - Getter

#### Test Cases Needed (6-8)
- [ ] Test IsExactDuplicate with >90% similarity
- [ ] Test IsExactDuplicate with <90% similarity
- [ ] Test CompilePatterns with valid regex
- [ ] Test CompilePatterns with invalid regex
- [ ] Test GetCompiledPaths returns patterns
- [ ] Test GetCompiledRepos returns patterns

**Estimated Effort:** 0.5-1 day

---

### 6. pkg/discovery/blocklist.go
**Current Coverage: 0% for LoadBlocklist**
**Target: 80%+**

#### Missing Test Coverage
- [ ] `LoadBlocklist` - Load from YAML file

#### Test Cases Needed (4-5)
- [ ] Test LoadBlocklist with valid YAML
- [ ] Test LoadBlocklist with missing file
- [ ] Test LoadBlocklist with malformed YAML
- [ ] Test LoadBlocklist compiles patterns
- [ ] Test LoadBlocklist error handling

**Estimated Effort:** 0.5 day

---

### 7. pkg/config/constants.go
**Current Coverage: 0%**
**Target: 80%+** (Low priority)

#### Missing Test Coverage
- [ ] `DefaultNotabilityWeights` - Returns defaults

#### Test Cases Needed (1-2)
- [ ] Test DefaultNotabilityWeights returns expected values
- [ ] Test weight values are reasonable

**Estimated Effort:** 0.25 day

---

### 8. pkg/cache/cache.go
**Current Coverage: 88.1%**
**Target: 95%+**

#### Missing Test Coverage
- [ ] `StartCleanupTimer` - Background cleanup

#### Test Cases Needed (2-3)
- [ ] Test cleanup timer runs
- [ ] Test cleanup removes expired entries
- [ ] Test timer doesn't leak goroutines

**Estimated Effort:** 0.5 day

---

## Priority 3: Integration & Main Functions (Lower Priority)

### 9. cmd/lima-catalog/main.go
**Current Coverage: 0%**
**Target: 50%+** (Focus on testable functions)

#### Missing Test Coverage
- [ ] `setupEnvironment` - Parse environment variables
- [ ] `initializeStorage` - Storage setup
- [ ] `checkRateLimits` - Rate limit verification
- [ ] Phase orchestration functions

#### Test Cases Needed (10-15)
- [ ] Test setupEnvironment with valid env vars
- [ ] Test setupEnvironment with missing token
- [ ] Test setupEnvironment with defaults
- [ ] Test initializeStorage success
- [ ] Test checkRateLimits sufficient quota
- [ ] Test checkRateLimits insufficient quota

**Estimated Effort:** 2-3 days

---

### 10. pkg/interfaces/interfaces.go
**Current Coverage: 0%**
**Target: 50%+** (Low priority - thin wrappers)

#### Test Cases Needed (3-5)
- [ ] Test DefaultHTTPClient creation and delegation
- [ ] Test DefaultFileSystem operations
- [ ] Test DefaultClock returns current time

**Estimated Effort:** 0.5 day

---

## Improvement Areas

### 11. Existing Test Enhancement

#### pkg/discovery improvements
- [ ] `findOriginal` - Improve from 44.7% to 80%+
- [ ] `downloadTemplateContent` - Improve from 76.9% to 90%+
- [ ] `AnalyzeTemplate` - Improve from 85.7% to 95%+
- [ ] `AnalyzeTemplates` - Improve from 76.2% to 90%+

#### Additional edge cases
- [ ] Test error paths in download/analysis
- [ ] Test various template formats
- [ ] Test boundary conditions

**Estimated Effort:** 1-2 days

---

## Implementation Roadmap

### Phase 1: Foundation (Current)
**Target: Bring discovery and github packages to 70%+**

- [ ] Complete pkg/discovery/discovery.go tests
- [ ] Complete pkg/github/client.go tests
- [ ] Complete pkg/types/types.go tests

**Timeline:** 1-1.5 weeks
**Target Coverage:** 55-60%

### Phase 2: Metadata & Analysis
**Target: Complete metadata and prompt packages**

- [ ] Complete pkg/discovery/metadata.go tests
- [ ] Complete pkg/prompt/builder.go tests
- [ ] Complete pkg/discovery/blocklist.go tests

**Timeline:** 1-1.5 weeks
**Target Coverage:** 65-70%

### Phase 3: Refinement
**Target: Polish and edge cases**

- [ ] Improve existing test coverage in discovery
- [ ] Complete pkg/config/constants.go tests
- [ ] Complete pkg/cache/cache.go tests
- [ ] Add integration test scenarios

**Timeline:** 0.5-1 week
**Target Coverage:** 70-75%

### Phase 4: CLI & Integration (Optional)
**Target: Main application coverage**

- [ ] Complete cmd/lima-catalog tests
- [ ] End-to-end integration tests
- [ ] Performance testing

**Timeline:** 1 week
**Target Coverage:** 75%+

---

## Testing Patterns

### Mock External Dependencies
```go
// Use mock HTTP clients for GitHub API
type mockHTTPClient struct {
    content string
}

// Use mock clocks for time-dependent logic
type mockClock struct {
    now time.Time
}
```

### Table-Driven Tests
```go
tests := []struct {
    name        string
    input       string
    expected    bool
    expectError bool
}{
    // test cases...
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test logic
    })
}
```

### Test Error Paths
- Network failures
- Invalid input
- Rate limiting
- Nil checks
- Context cancellation

---

## Success Metrics

- **Overall Coverage:** 70%+ (up from 45.9%)
- **Priority 1 Packages:** 80%+ coverage
- **Priority 2 Packages:** 70%+ coverage
- **Critical Business Logic:** 90%+ coverage
- **All Tests Pass:** No flaky tests
- **Fast Test Suite:** <30 seconds for full suite

---

## Notes

- Focus on critical business logic first
- Follow existing test patterns in the codebase
- Use mocks for external dependencies
- Test both success and error paths
- Add comments explaining complex test scenarios
- Keep tests maintainable and readable
