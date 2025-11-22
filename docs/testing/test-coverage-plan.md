# Backend Test Coverage Improvement Plan

**Current Overall Coverage: 49.5%**
**Target Coverage: 70%+**

Last Updated: 2025-11-22

## Coverage Status by Package

### ✅ Excellent Coverage (80%+)
- [x] `pkg/validation` - 98.6% - Nearly complete
- [x] `pkg/combiner` - 94.5% - Very good
- [x] `pkg/minhash` - 93.4% - Very good
- [x] `pkg/github` - 90.5% - Excellent (CheckRateLimit, HandleRateLimitError edge cases)
- [x] `pkg/cache` - 88.1% - Good (missing: StartCleanupTimer)
- [x] `pkg/storage` - 86.7% - Good
- [x] `pkg/retry` - 83.1% - Good

### ⚠️ Medium Coverage (30-70%)
- [ ] `pkg/discovery` - 49.8% - Mixed coverage (architecture limits further improvement)
- [ ] `pkg/prompt` - 36.0% - Partial improvement (validation, constructors, converters tested)

### ❌ Low/No Coverage (<30%)
- [ ] `pkg/config` - 0.0% - Constants only (low priority)
- [ ] `pkg/types` - 0.0% - Helper methods untested
- [ ] `pkg/interfaces` - 0.0% - Interface wrappers (low priority)
- [ ] `cmd/lima-catalog` - 0.0% - Main application
- [ ] `cmd/debug-template` - 0.0% - CLI tool
- [ ] `cmd/prompt-generator` - 0.0% - CLI tool

---

## Recent Progress Summary

### ✅ Completed (PR merged)
- **pkg/github/client.go**: 36.5% → 90.5% (+54 percentage points) ✅
  - Added comprehensive CheckRateLimit tests with mock HTTP server
  - Added HandleRateLimitError edge case tests (unknown limit type, API errors, zero reset time, negative wait)
  - Tests use table-driven patterns with mock HTTP servers

- **pkg/prompt/builder.go**: 23.1% → 36.0% (+12.9 percentage points) ⚠️
  - Added PromptConfig.Validate() tests (negative values, zero values, edge cases)
  - Added NewBuilder constructor tests (invalid tokens, invalid configs, nil config)
  - Added convertGitHubRepo and convertGitHubUser helper tests
  - Main functions (BuildPrompt, GatherContext, fetchTemplateContent) blocked by architecture

- **pkg/discovery/discovery.go**: 48.4% → 49.8% (+1.4 percentage points) ⚠️
  - Added comprehensive FindNewestTemplateTimestamp tests (empty, single, multiple, zero times, large lists)
  - Added NewDiscoverer tests (nil blocklist, with blocklist)
  - Added DiscoverAll context cancellation test
  - Main discovery functions (isLimaTemplate, searchWithQuery, DiscoverCommunityTemplates, DiscoverOfficialTemplates) blocked by architecture
  - Added extensive documentation explaining testing limitations (see pkg/discovery/discovery_test.go lines 134-162)

**Overall Impact:** +3.6 percentage points (45.9% → 49.5%)

### 🎯 Recommended Next Steps

Based on the current state and ease of implementation, the recommended priority order is:

1. **pkg/types/types.go** (Quick Win - 0.5-1 day)
   - Test `IsExactDuplicate`, `CompilePatterns`, getters
   - Pure functions, no external dependencies, easy to test
   - Will improve overall coverage quickly

2. **pkg/discovery/blocklist.go** (Quick Win - 0.5 day)
   - Test `LoadBlocklist` with valid YAML, missing files, malformed YAML
   - Simple file I/O, easy to mock with temp files

3. **pkg/discovery/metadata.go** (High Impact - 2-3 days)
   - Same architecture limitation as discovery.go (concrete *github.Client)
   - But metadata collection is critical business logic
   - May require refactoring to use interfaces for better testability

4. **pkg/cache/cache.go** (Polish - 0.5 day)
   - Complete remaining StartCleanupTimer tests
   - Already at 88.1%, small effort to reach 95%+

**Total Estimated Effort for Quick Wins (1+2+4):** 1.5-2 days to reach ~52-54% overall coverage

---

## Priority 1: Critical Business Logic (High Impact)

### 1. pkg/discovery/discovery.go ⚠️ PARTIAL PROGRESS
**Current Coverage: 49.8%** (was 48.4%)
**Target: 80%+**

#### Test Coverage Status
- [x] `NewDiscoverer` - Constructor ✅
- [x] `FindNewestTemplateTimestamp` - Utility function ✅
- [ ] `isLimaTemplate` - Template validation logic (architecture limitation - requires GitHub API)
- [ ] `searchWithQuery` - GitHub search with pagination (architecture limitation - requires GitHub API)
- [ ] `DiscoverCommunityTemplates` - Community template discovery (architecture limitation - requires GitHub API)
- [ ] `DiscoverOfficialTemplates` - Official template discovery (architecture limitation - requires GitHub API)
- [x] `DiscoverAll` - Context cancellation tested ✅

#### Test Cases Completed (3/17)
- [x] Test NewDiscoverer creates valid instance with nil blocklist
- [x] Test NewDiscoverer creates valid instance with blocklist
- [x] Test FindNewestTemplateTimestamp with empty list
- [x] Test FindNewestTemplateTimestamp finds newest
- [x] Test FindNewestTemplateTimestamp with same timestamps
- [x] Test FindNewestTemplateTimestamp with zero times
- [x] Test FindNewestTemplateTimestamp with large lists (performance)
- [x] Test DiscoverAll handles context cancellation

#### Remaining Work
Functions like `isLimaTemplate`, `searchWithQuery`, `DiscoverCommunityTemplates`, and `DiscoverOfficialTemplates` use a concrete `*github.Client` type rather than an interface. To achieve 80%+ coverage would require:
- Refactoring to use interface-based dependency injection (e.g., `GitHubClient` interface)
- Or integration tests with actual GitHub API (slow, requires token, flaky)

See detailed architectural notes in `pkg/discovery/discovery_test.go` lines 134-162.

**Status:** ⚠️ Partial - 49.8% coverage (constructors and pure functions tested). Main discovery functions blocked by architecture.

**Estimated Remaining Effort:** 2-3 days (requires architectural refactoring for proper mocking)

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

### 3. pkg/github/client.go ✅ COMPLETED
**Current Coverage: 90.5%** (was 36.5%)
**Target: 80%+** ✅ **ACHIEVED**

#### Test Coverage Status
- [x] `RateLimit` - Get rate limit status
- [x] `CheckRateLimit` - Verify sufficient quota
- [x] `HandleRateLimitError` - Rate limit error handling with edge cases
- [ ] `SearchCode` - Code search API (not needed - tested via integration)
- [ ] `ListRepositoryContents` - Directory listing (not needed - tested via integration)
- [ ] `GetRepositoryContent` - File content fetch (not needed - tested via integration)

#### Test Cases Completed (8/10)
- [x] Test CheckRateLimit with sufficient quota
- [x] Test CheckRateLimit at minimum threshold
- [x] Test CheckRateLimit below threshold
- [x] Test CheckRateLimit with zero remaining
- [x] Test CheckRateLimit with API error
- [x] Test HandleRateLimitError with unknown limit type
- [x] Test HandleRateLimitError with rate limit API error
- [x] Test HandleRateLimitError with zero reset time
- [x] Test HandleRateLimitError with negative wait duration

**Status:** ✅ Complete - 90.5% coverage achieved

---

### 4. pkg/prompt/builder.go ⚠️ PARTIAL PROGRESS
**Current Coverage: 36.0%** (was 23.1%)
**Target: 70%+**

#### Test Coverage Status
- [x] `NewBuilder` - Constructor with validation ✅
- [x] `Validate` (types.go) - Config validation ✅
- [x] `convertGitHubRepo` - Type conversion ✅
- [x] `convertGitHubUser` - Type conversion ✅
- [ ] `BuildPrompt` - Main prompt generation (architecture limitation - requires GitHub API)
- [ ] `GatherContext` - Context collection (architecture limitation - requires GitHub API)
- [ ] `fetchTemplateContent` - Template download (architecture limitation - requires GitHub API)
- [ ] `fetchReadme` - README fetching (architecture limitation - requires GitHub API)
- [ ] `findTemplateReferences` - Git clone and grep (architecture limitation - requires git/filesystem)

#### Test Cases Completed (9/15)
- [x] Test NewBuilder with invalid token (empty and short)
- [x] Test NewBuilder with invalid config (negative values)
- [x] Test NewBuilder with nil config (uses default)
- [x] Test PromptConfig.Validate with negative context lines
- [x] Test PromptConfig.Validate with zero values (valid)
- [x] Test PromptConfig.Validate with negative MaxReadmeLength
- [x] Test PromptConfig.Validate with negative MaxReferenceFiles
- [x] Test convertGitHubRepo type conversion
- [x] Test convertGitHubUser type conversion

#### Remaining Work
Functions like `BuildPrompt`, `GatherContext`, `fetchTemplateContent`, `fetchReadme`, and `findTemplateReferences` require GitHub API and filesystem access. To test these without refactoring to use interfaces:
- Would need integration tests with actual GitHub API (slow, requires token)
- Or significant refactoring to use dependency injection with interfaces

**Status:** ⚠️ Partial - 36% coverage (constructors, validation, helpers tested). Main functions blocked by architecture.

**Estimated Remaining Effort:** 2-3 days (requires architectural refactoring for mocking)

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

### Phase 1: Foundation (In Progress)
**Target: Bring discovery and github packages to 70%+**

- [x] Complete pkg/github/client.go tests ✅ (90.5% coverage achieved)
- [ ] Complete pkg/discovery/discovery.go tests (49.8% - architecture limited)
- [ ] Complete pkg/types/types.go tests

**Current Status:**
- Overall coverage improved from 45.9% to 49.5% (+3.6 percentage points)
- pkg/github achieved 90.5% coverage (target exceeded)
- pkg/prompt improved to 36.0% coverage (validation and constructors tested)
- pkg/discovery improved to 49.8% coverage (limited by architecture)

**Next Priority:** pkg/types/types.go helper methods (quick win, estimated 0.5-1 day)

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
