# Implementation Notes - Lima Catalog

This document contains detailed implementation notes for completed features. For current architecture and remaining work, see [PLAN.md](PLAN.md).

## Completed Stages (Incremental Update Redesign)

### Stage 1: Incremental Discovery ✅

**Implementation details:**
- Timestamp-based incremental discovery using `pushed:>DATE` qualifier
- `FindNewestTemplateTimestamp()` helper finds baseline for incremental search
- 24-hour lookback buffer to handle timezone/clock skew
- Automatic fallback to full discovery when no existing data
- Sanity check warns if incremental search returns 0 results

**Files:**
- `pkg/discovery/discovery.go` - Incremental discovery with sinceDate parameter
- `config/blocklist.yaml` - Regex patterns for filtering
- `pkg/discovery/blocklist.go` - Filtering logic

**Test coverage:**
- 17 unit test cases for blocklist filtering
- Integration test suite (`scripts/test-integration.sh`)
- Makefile for easy testing

---

### Stage 2: Content Validation ✅

Already implemented in Phase 1. Existing logic works well:
- Download template content
- Parse YAML
- Verify `images:` key exists at top level
- ~31% false positive rate remains acceptable

---

### Stage 3: Template Analysis ✅

Already implemented in Phase 2. Existing keyword/category extraction works well:
- Technology detection from provisioning scripts
- Keyword extraction
- Category assignment based on detected technologies
- Name derivation from paths

**Incremental optimization:**
- Only re-analyzes templates when SHA changes (see `analyzer.go:170`)
- Preserves existing analysis until template is updated
- Significantly reduces daily processing time

---

### Stage 4: Metadata Management ✅

**Implementation details:**
- `SelectReposToRefresh()` and `SelectOrgsToRefresh()` selection functions
- Intelligent refresh cycle: new templates + 5% of stale (>30 days) entries
- **Oldest-first selection** (not random) ensures stalest data refreshed first
- Spreads refresh load over ~20 days (100% / 5%)
- Prevents thundering herd problem

**Files:**
- `pkg/discovery/metadata.go` - Refresh selection logic
- `pkg/discovery/metadata_test.go` - 11 unit test cases

**Key algorithm:**
```go
// Find stale entries (>30 days old)
// Sort by LastFetched (oldest first)
// Select up to 5% of total entries
// Prioritize oldest for refresh
```

---

### Stage 5: Frontend Data Combination ✅

**Implementation details:**
- `pkg/combiner` package for frontend data generation
- `CombineData()` method merges templates + repos + orgs
- Blocklist integration (skips filtered templates)
- Automatic sorting by org/repo/path for stable diffs
- Raw URL generation from default branch

**Output file:** `catalog.jsonl`

**Data combination logic:**
- Description priority: short_description > first 3 keywords > "Lima VM template"
- Name priority: display_name > name > path
- Joins template data with repo metadata (stars, updated_at)
- Extracts org from repo for organization field

**Files:**
- `pkg/combiner/combiner.go` - Data combination logic (218 lines)
- `pkg/combiner/combiner_test.go` - 19 unit test cases

**Integration tests:**
- Test 4 in `scripts/test-integration.sh`
- Validates JSON format, required fields, sorting

---

## Testing Strategy

### Unit Tests

**No network calls:**
- Blocklist filter matching (path patterns and repo names)
- Metadata refresh selection (oldest-first logic)
- Date parsing and timestamp handling
- Template YAML parsing
- Data file sorting algorithms
- Frontend data combination

**Current coverage:** 28 Go tests

### Integration Tests

**With real GitHub API:**
- Uses test GitHub token
- Run discovery on small query (limit=10)
- Verify blocklist filtering works
- Test incremental discovery (48-hour window)
- Validate data file formats
- Check catalog.jsonl generation

**Test suite:** `scripts/test-integration.sh`
- Test 1: Full discovery baseline
- Test 2: Incremental mode with timestamp filtering
- Test 3: Blocklist filtering
- Test 4: Frontend catalog generation

**Makefile target:** `make test`

---

## Data File Sorting

**Rationale:**
- **Human browsing**: Grouped by org/repo makes files easier to navigate
- **Stable diffs**: New templates from existing repos appear near related templates
- **Pattern detection**: Issues and trends easier to spot when related entries adjacent
- **Minimal overhead**: Sorting 700-1000 items is negligible (<1ms)

**Sort orders:**

**templates.jsonl:**
- Primary: org (alphabetical)
- Secondary: repo (alphabetical)
- Tertiary: path (alphabetical)

**repos.jsonl:**
- Primary: org (alphabetical)
- Secondary: repo (alphabetical)

**orgs.jsonl:**
- Single key: id (alphabetical)

**catalog.jsonl:**
- Matches templates.jsonl order: org/repo/path (alphabetical)

**Implementation:**
```go
sort.Slice(templates, func(i, j int) bool {
    if templates[i].Org != templates[j].Org {
        return templates[i].Org < templates[j].Org
    }
    if templates[i].Repo != templates[j].Repo {
        return templates[i].Repo < templates[j].Repo
    }
    return templates[i].Path < templates[j].Path
})
```

---

## Rollout History

### Phase 1: Blocklist & Incremental Discovery
- Added `config/blocklist.yaml` with initial path patterns
- Implemented timestamp-based incremental search
- Added integration test suite
- PR #58 merged

### Phase 2: Metadata Refresh Cycle
- Implemented oldest-first refresh selection
- Added metadata refresh tests
- Integrated with main.go incremental mode
- PR #60 merged

### Phase 3: Frontend Data Preparation
- Created combiner package
- Generated catalog.jsonl for frontend
- Added catalog validation to integration tests
- Renamed from templates-combined.jsonl to catalog.jsonl
- PR #61 merged (current)

---

## Environment Variables

**Required:**
```bash
GITHUB_TOKEN=<token>  # Personal access token with public_repo scope
```

**Optional (for future LLM stage):**
```bash
LLM_API_KEY=<api-key>
LLM_PROVIDER=anthropic  # anthropic, openai, etc.
LLM_MODEL=claude-3-haiku-20240307
LLM_MAX_PER_RUN=1
```

**Configuration:**
```bash
DATA_DIR=./data  # Default
INCREMENTAL=1    # Enable incremental mode
```

---

## Migration Notes

### Initial Sort (One-time Large Diff)
- First commit with sorted data files created large diff
- Subsequent updates have minimal, localized diffs
- Separate commit: "Sort data files for human readability"

### Blocklist Migration
- Moved from code to `config/blocklist.yaml`
- Changed location from root to `config/` directory
- Simplified patterns using regex for flexibility

### Catalog File Migration
- Original: Multiple separate JSONL files
- New: Single `catalog.jsonl` for frontend
- Frontend only needs to load one file
- Reduces network requests and client-side processing
