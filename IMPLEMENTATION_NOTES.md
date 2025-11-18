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

**Frontend Integration (2025-01-16):**
- Updated frontend to use `catalog.jsonl` instead of loading `templates.jsonl` and `repos.jsonl` separately
- Simplified data loading: single HTTP request instead of two
- Removed repository state management and lookup code
- Updated URL helpers to use embedded `raw_url` field
- All 76 frontend tests passing

**Changed files:**
- `docs/js/data.js` - Replaced `loadTemplates()` and `loadRepositories()` with `loadCatalog()`
- `docs/js/app.js` - Removed repositories parameter from rendering functions
- `docs/js/filters.js` - Use embedded `stars` and `updated_at` fields directly
- `docs/js/templateCard.js` - Use embedded `description` and `stars` fields
- `docs/js/modal.js` - Removed repo parameter, use `raw_url` directly
- `docs/js/urlHelpers.js` - Convert `raw_url` to display URL without repo lookup
- `docs/js/state.js` - Removed repositories state management
- Test files updated to match new data structure

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

---

## Code Quality Improvements

### Keyboard Navigation Refactoring (2025-01-16)

**Motivation:** External code review noted that keyboard navigation in `app.js` was imperative and hard to maintain. Modifier key handling was inconsistent across different shortcuts.

**Changes:**
- Converted 160+ lines of if/else chains to declarative `KEYBOARD_SHORTCUTS` configuration object
- Created `getKeyString()` helper for consistent key+modifier matching
- Made modifier key requirements explicit (`requiresNoModifiers`, `allowsShift`)
- Added `skipIfTyping`, `preventDefault`, and `condition` flags for clear behavior specification
- Added descriptive comments for each shortcut directly in config

**Benefits:**
- All 15+ shortcuts visible at a glance in single config object
- Modifier key handling is now explicit and consistent
- Easier to add/modify shortcuts - just add/edit config entry
- Self-documenting structure with description fields
- Reduces chance of bugs like the Ctrl+Up navigation issue we just fixed
- Config-based approach makes it clear that `Ctrl+ArrowUp` is separate from plain `ArrowUp`

**Example config entry:**
```javascript
'Ctrl+ArrowUp': {
    description: 'Focus header (theme switcher)',
    skipIfTyping: false,
    preventDefault: true,
    action: (e, ctx) => {
        const themeButton = document.querySelector('.theme-switcher button');
        if (themeButton) themeButton.focus();
    }
}
```

**Files changed:**
- `docs/js/app.js` - Refactored `setupKeyboardShortcuts()` function

**No functional changes** - all existing shortcuts work exactly the same way.

---

## Debug Mode for Notability Score Tuning (2025-01-18)

**Motivation:** Need a way to visualize notability scores on the frontend while tuning the scoring weights, without cluttering the normal user interface.

**Implementation:**
- Added hidden debug mode toggled by `@` keyboard shortcut
- Debug mode replaces "Official"/"Community" badge text with notability score
- Badge color remains unchanged (official=blue, community=gray)
- Hover over debug badge shows detailed score breakdown popup
- Added `notability_score_breakdown` field to catalog.jsonl with individual components

**Backend Changes:**
- Created `NotabilityScoreBreakdown` struct with individual score components (message, provision, parameters, env_vars, probes, unusual_images, comments, stars, total)
- Modified `CalculateNotabilityScoreWithBreakdown()` to return breakdown
- Added breakdown to `CombinedTemplate` struct in combiner
- Breakdown field uses `omitempty` JSON tag (optional field)

**Frontend Changes:**
- Added debug mode state management to `state.js`
- Added `@` keyboard shortcut to toggle debug mode
- Modified `templateCard.js` to replace badge text in debug mode
- Created `createDebugScorePopup()` to generate hover popup with breakdown
- Added CSS for debug popup styling and notification
- Added "Notability" option to sort dropdown

**Files Changed:**
- `pkg/discovery/notability.go` - Added NotabilityScoreBreakdown struct and calculation
- `pkg/combiner/combiner.go` - Added breakdown field to CombinedTemplate
- `docs/js/state.js` - Added debug mode state
- `docs/js/app.js` - Added @ keyboard shortcut and notification
- `docs/js/templateCard.js` - Badge replacement and popup rendering
- `docs/index.html` - Added "Notability" to sort dropdown
- `docs/style.css` - Debug mode styling

**Comment Filtering Enhancement (2025-01-18):**

**Problem:** default.yaml scored 1132 primarily from 541 comment lines. Derivative templates inherited these comments, getting artificially inflated scores.

**Solution:** Implemented comment filtering during analysis:
1. Fetch default.yaml from lima-vm/lima repository
2. Extract and normalize all comment lines
3. Filter out comments present in default.yaml
4. Filter out empty comments (just `#` with whitespace)
5. Only count unique, meaningful comments

**Implementation:**
- Added `DefaultComments map[string]bool` to Analyzer struct
- Created `FetchDefaultTemplateComments()` to fetch and parse default.yaml
- Modified parser to store `CommentLines []string` (not just count)
- Created `isEmptyComment()` to detect empty comments
- Created `FilterUniqueComments()` to count filtered comments
- Updated `PopulateNotabilityMetrics()` to accept defaultComments parameter

**Files Changed:**
- `pkg/discovery/notability.go` - Added FilterUniqueComments, isEmptyComment, FetchDefaultTemplateComments
- `pkg/discovery/analyzer.go` - Added DefaultComments field and method
- `pkg/discovery/parser.go` - Store CommentLines array
- `cmd/lima-catalog/main.go` - Call FetchDefaultTemplateComments during analysis
- `pkg/discovery/notability_test.go` - Updated test expectations

**Impact:** Templates will be re-analyzed with new comment filtering, significantly reducing scores for derivative templates while preserving scores for templates with unique documentation.

---

## Debug Mode Subscore Sorting (2025-01-18)

**Motivation:** When tuning notability scoring weights, need ability to sort by individual score components to identify which templates score highest in each category.

**Implementation:**
- Dynamic sort dropdown: adds 8 subscore options when debug mode enabled
- Sort by individual components: message, provision, parameters, env_vars, probes, unusual_images, comments, stars
- Badge shows active subscore: when sorting by a component, badge displays that component's score instead of total
- Visual indicator: 🔍 emoji prefix on debug sort options
- Sort options removed when debug mode disabled

**Smart Badge Display:**
- Default: Shows total notability score
- Sorting by breakdown component: Shows that specific component score
- Badge title updates to indicate which score is displayed (e.g., "Comments Score (hover for breakdown)")
- Hover popup always shows full breakdown regardless of active sort

**Example Use Cases:**
- Sort by "Comments Score" to find templates with best unique documentation
- Sort by "Provision Score" to find templates with most setup scripts
- Sort by "Parameters Score" to find most configurable templates
- Sort by "Unusual Images Score" to find templates using non-standard distributions

**Files Changed:**
- `docs/js/filters.js` - Added 8 breakdown sort cases to sortTemplates()
- `docs/js/app.js` - Added updateSortDropdown() function, called on debug toggle and initialization
- `docs/js/templateCard.js` - Modified getDebugBadgeText() to accept sortBy parameter and return appropriate score
- `docs/js/filters.test.js` - Added 6 new tests for breakdown sorting

**Testing:**
- 6 new test cases covering breakdown sorting for different components
- Tests verify correct descending sort order
- Tests verify graceful handling of missing breakdown data
- Total: 83 tests passing (77 original + 6 new)
