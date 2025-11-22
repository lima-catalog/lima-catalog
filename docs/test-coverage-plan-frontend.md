# Frontend Test Coverage Plan

## Executive Summary

**Current State:**
- ✅ 76 tests passing across 5 modules
- 📊 ~56% module coverage (5 of 9 core modules)
- 🎯 100% coverage on tested modules
- ❌ 4 largest/most complex modules untested (81% of codebase by size)

**Technology Stack:**
- Vanilla JavaScript (ES6 modules, no framework)
- Custom test framework (lightweight, browser + Node.js compatible)
- No build step or external dependencies

**Coverage Gaps:**
- UI interaction modules (app.js, appActions.js, modal.js, sidebar.js, keyboard.js)
- Integration testing
- DOM manipulation testing
- Async operations testing

---

## Current Test Coverage Analysis

### ✅ Fully Tested Modules (100% Coverage)

| Module | Size | Tests | Status | Coverage Areas |
|--------|------|-------|--------|----------------|
| `data.js` | 1.1KB | 7 | ✅ Complete | JSONL parsing, error handling, edge cases |
| `filters.js` | 11KB | 24+ | ✅ Complete | Filtering logic, sorting, keyword/category counts |
| `templateCard.js` | 20KB | 18+ | ✅ Complete | Display name derivation, HTML escaping, formatting |
| `theme.js` | 1.6KB | 11 | ✅ Complete | Theme management, localStorage, system preferences |
| `urlHelpers.js` | 2KB | 4 | ✅ Complete | URL generation, GitHub URL parsing |

### ❌ Untested Modules (No Coverage)

| Module | Size | Complexity | Priority | Testing Effort |
|--------|------|------------|----------|----------------|
| `modal.js` | 45KB | Very High | **High** | 🔴 High |
| `keyboard.js` | 43KB | Very High | **High** | 🔴 High |
| `sidebar.js` | 28KB | High | **High** | 🟡 Medium |
| `appActions.js` | 10KB | High | **High** | 🟡 Medium |
| `app.js` | 4KB | Medium | **Medium** | 🟡 Medium |
| `state.js` | 1.6KB | Low | **Low** | 🟢 Low |
| `utils.js` | 2.3KB | Low | **Low** | 🟢 Low |
| `config.js` | 0.2KB | None | **N/A** | - |

---

## Comprehensive Test Coverage Plan

### Phase 1: Foundation & Infrastructure (Week 1-2)

**Goal:** Establish DOM testing infrastructure and test simple modules

#### 1.1 Enhance Test Framework
- [ ] Add DOM testing support (jsdom or happy-dom for Node.js)
- [ ] Add async test support (`async/await` in tests)
- [ ] Add timer mocking utilities (for `debounce` tests)
- [ ] Add test coverage reporting (basic line/function coverage)
- [ ] Document testing patterns and best practices

**Estimated Effort:** 3-4 days
**Risk:** Low
**Value:** High (enables all future testing)

#### 1.2 Test Simple Modules
- [ ] **Test `state.js`** (getter/setter pattern)
  - [ ] Template storage and retrieval
  - [ ] Keyword selection (toggle, clear, has)
  - [ ] Category selection
  - [ ] Focused template tracking
  - [ ] Debug mode toggle
  - [ ] Clear all selections
  - **Estimated Tests:** 15-20
  - **Effort:** 0.5 day

- [ ] **Test `utils.js`** (utility functions)
  - [ ] `debounce()` - timing, cancellation, immediate execution
  - [ ] `trapFocus()` - focus cycling in modals
  - **Estimated Tests:** 8-10
  - **Effort:** 1 day

**Phase 1 Total:** ~5 days, +25-30 tests

---

### Phase 2: Core UI Actions (Week 3-4)

**Goal:** Test critical UI interaction logic

#### 2.1 Test `appActions.js` (High Priority)
**Why this matters:** Core filtering and rendering orchestration

- [ ] **`filterAndRender()`**
  - [ ] Applies filters correctly
  - [ ] Updates URL parameters
  - [ ] Renders filtered results
  - [ ] Preserves focus when requested
  - [ ] Updates sidebar counts

- [ ] **`applyFiltersFromURL()`**
  - [ ] Parses URL parameters correctly
  - [ ] Sets keyword selections
  - [ ] Sets category selection
  - [ ] Sets sort order
  - [ ] Handles malformed URLs

- [ ] **`updateSidebarOnly()`**
  - [ ] Updates keyword cloud
  - [ ] Updates category list
  - [ ] Preserves template cards

- [ ] **Toggle handlers**
  - [ ] `handleKeywordToggle()` - multi-select logic
  - [ ] `handleCategoryToggle()` - single-select logic
  - [ ] State synchronization

- [ ] **`updateSortDropdown()`**
  - [ ] Updates dropdown value
  - [ ] Handles missing sort option

**Estimated Tests:** 30-40
**Effort:** 4-5 days
**Dependencies:** Requires DOM mocking (from Phase 1)

#### 2.2 Test `app.js` (Medium Priority)
**Why this matters:** Application initialization and setup

- [ ] **`initialize()`**
  - [ ] Loads data successfully
  - [ ] Applies URL filters
  - [ ] Renders initial UI
  - [ ] Handles data loading errors

- [ ] **`setupEventListeners()`**
  - [ ] Search input event listener
  - [ ] Sort dropdown listener
  - [ ] Clear keywords button
  - [ ] Official/community toggles
  - [ ] Show duplicates toggle

- [ ] **Event handler tests**
  - [ ] Search debouncing works
  - [ ] Sort change triggers re-render
  - [ ] Clear keywords resets state

**Estimated Tests:** 20-25
**Effort:** 3-4 days
**Dependencies:** Requires async testing, DOM mocking

**Phase 2 Total:** ~8 days, +50-65 tests

---

### Phase 3: Sidebar & Rendering (Week 5-6)

**Goal:** Test sidebar rendering and keyword management

#### 3.1 Test `sidebar.js` (High Priority)
**Why this matters:** Complex DOM rendering with event handling

- [ ] **`renderKeywordCloud()`**
  - [ ] Renders keywords with correct counts
  - [ ] Sorts keywords by count descending
  - [ ] Limits keywords to MAX_KEYWORDS_DISPLAY
  - [ ] Shows "Show more" link when needed
  - [ ] Highlights selected keywords
  - [ ] Excludes specified keywords
  - [ ] Handles empty keyword list

- [ ] **`renderCategories()`**
  - [ ] Renders all categories alphabetically
  - [ ] Shows correct counts per category
  - [ ] Highlights selected category
  - [ ] Shows "All" option with total count
  - [ ] Handles empty category list

- [ ] **`updateSidebar()`**
  - [ ] Updates both keywords and categories
  - [ ] Preserves scroll position
  - [ ] Preserves focus when requested

- [ ] **Dynamic keywords**
  - [ ] Shows dynamic keywords for focused template
  - [ ] Includes org/repo keywords
  - [ ] Excludes already-selected keywords
  - [ ] Handles templates without dynamic keywords

- [ ] **Event handling**
  - [ ] Keyword click toggles selection
  - [ ] Category click updates selection
  - [ ] "Show more" expands keyword list
  - [ ] Enter key activates keywords/categories

**Estimated Tests:** 35-45
**Effort:** 5-6 days
**Dependencies:** DOM mocking, state.js tests

**Phase 3 Total:** ~6 days, +35-45 tests

---

### Phase 4: Modal & Preview (Week 7-9)

**Goal:** Test complex modal logic and YAML preview

#### 4.1 Test `modal.js` (High Priority, High Complexity)
**Why this matters:** Largest module with complex diff algorithm

##### Modal State Management (5-7 tests)
- [ ] `openPreviewModal()` opens modal correctly
- [ ] `closePreviewModal()` cleans up state
- [ ] Focus trap prevents tab outside modal
- [ ] Escape key closes modal
- [ ] Multiple modals don't interfere

##### YAML Content Loading (10-15 tests)
- [ ] `loadTemplateContent()` fetches YAML successfully
- [ ] Handles network errors gracefully
- [ ] Shows loading indicator
- [ ] Caches loaded content
- [ ] Displays syntax-highlighted YAML
- [ ] Handles empty YAML files
- [ ] Handles malformed YAML
- [ ] Timeout handling for slow networks

##### Diff Generation (15-20 tests)
- [ ] `generateDiff()` creates unified diff format
- [ ] LCS algorithm finds correct changes
- [ ] Handles identical files (no diff)
- [ ] Handles completely different files
- [ ] Handles added lines correctly
- [ ] Handles removed lines correctly
- [ ] Handles modified lines correctly
- [ ] Context lines (unchanged) displayed
- [ ] Line numbers accurate
- [ ] Handles large diffs (performance)
- [ ] Handles empty files

##### Duplicate Templates (8-12 tests)
- [ ] Shows duplicate templates in modal
- [ ] Generates diffs between duplicates
- [ ] Handles multiple duplicates
- [ ] Links to duplicate template sources
- [ ] Handles no duplicates case

##### Similar Templates (5-8 tests)
- [ ] Shows similar templates
- [ ] Compares YAML content
- [ ] Displays similarity score/indicator
- [ ] Handles no similar templates

##### URL Handling (5-7 tests)
- [ ] Updates URL with `?preview=<id>`
- [ ] Loads preview from URL on page load
- [ ] Handles invalid preview IDs
- [ ] Copy URL to clipboard functionality
- [ ] Browser back/forward navigation

##### Focus Management (8-10 tests)
- [ ] Focus trap within modal
- [ ] Focus returns to trigger element on close
- [ ] Tab cycles through modal elements
- [ ] Shift+Tab cycles backwards
- [ ] Escape closes and restores focus

**Estimated Tests:** 56-79
**Effort:** 10-12 days
**Dependencies:** Async testing, DOM mocking, fetch mocking, Highlight.js mocking

**Phase 4 Total:** ~12 days, +56-79 tests

---

### Phase 5: Keyboard Navigation (Week 10-12)

**Goal:** Test comprehensive keyboard navigation system

#### 5.1 Test `keyboard.js` (High Priority, Very High Complexity)
**Why this matters:** Critical accessibility feature, complex viewport calculations

##### Keyboard Shortcut Registration (5-8 tests)
- [ ] All shortcuts registered on init
- [ ] Global shortcuts work from any context
- [ ] Context-specific shortcuts work correctly
- [ ] Shortcuts don't interfere with input fields
- [ ] Modifier keys (Ctrl, Shift, Alt) handled

##### Arrow Key Navigation (15-20 tests)
- [ ] Arrow Right navigates to next template
- [ ] Arrow Left navigates to previous template
- [ ] Arrow Down navigates to template below
- [ ] Arrow Up navigates to template above
- [ ] Handles edge cases (first/last template)
- [ ] Wrapping behavior at grid edges
- [ ] Respects grid layout (rows/columns)
- [ ] Scrolls viewport when needed
- [ ] Updates focused template in state
- [ ] Visual focus indicator appears

##### Grid Calculations (10-12 tests)
- [ ] `getTemplatePosition()` calculates row/column correctly
- [ ] `getTemplateAtPosition()` finds correct template
- [ ] Handles variable card widths (responsive)
- [ ] Handles different screen sizes
- [ ] Handles filtered results (gaps in grid)

##### Page Navigation (8-10 tests)
- [ ] Page Down scrolls one screen
- [ ] Page Up scrolls one screen
- [ ] Home goes to first template
- [ ] End goes to last template
- [ ] Scroll position updated correctly

##### Special Shortcuts (10-15 tests)
- [ ] "/" focuses search input
- [ ] "?" opens help modal
- [ ] "@" toggles debug mode
- [ ] Enter opens preview modal for focused template
- [ ] Escape closes modals
- [ ] Tab/Shift+Tab navigate sidebar

##### Modal Context Navigation (10-15 tests)
- [ ] Ctrl+Right opens next template in modal
- [ ] Ctrl+Left opens previous template in modal
- [ ] Arrow keys disabled in modal (except Ctrl+)
- [ ] Navigation updates modal content
- [ ] URL updated on modal navigation

##### Focus Management (8-10 tests)
- [ ] Focus restored after modal close
- [ ] Focus visible with outline
- [ ] Focus scrolled into viewport
- [ ] Focus preserved during filter changes
- [ ] Focus cleared when template filtered out

**Estimated Tests:** 66-90
**Effort:** 12-15 days
**Dependencies:** DOM mocking, viewport mocking, extensive event simulation

**Phase 5 Total:** ~15 days, +66-90 tests

---

## Phase Summary

| Phase | Duration | Focus Area | New Tests | Cumulative Coverage |
|-------|----------|------------|-----------|---------------------|
| **Current** | - | Core logic | 76 | ~56% |
| **Phase 1** | 1-2 weeks | Infrastructure + simple modules | +25-30 | ~65% |
| **Phase 2** | 2 weeks | Core UI actions | +50-65 | ~75% |
| **Phase 3** | 1-2 weeks | Sidebar rendering | +35-45 | ~82% |
| **Phase 4** | 2-3 weeks | Modal & preview | +56-79 | ~90% |
| **Phase 5** | 2-3 weeks | Keyboard navigation | +66-90 | ~95%+ |
| **Total** | **10-12 weeks** | Full coverage | **+232-309 tests** | **95%+** |

---

## Alternative Approaches & Discussion Points

### Option A: Incremental Approach (Recommended)
- Follow phases 1-5 sequentially
- Validate each phase before proceeding
- Adjust priorities based on findings
- **Pros:** Lower risk, continuous value delivery
- **Cons:** Longer timeline

### Option B: Parallel Approach
- Work on multiple phases simultaneously
- Requires multiple developers
- Faster completion (6-8 weeks with 2-3 devs)
- **Pros:** Faster time to completion
- **Cons:** Higher coordination overhead, potential rework

### Option C: Risk-Based Prioritization
- Focus only on high-risk modules (modal.js, keyboard.js)
- Skip low-value tests (state.js, config.js)
- Aim for 80% coverage instead of 95%
- **Pros:** Faster, focuses on critical paths
- **Cons:** Leaves gaps in coverage

### Option D: Migration to Modern Test Framework
- Migrate from custom framework to Vitest or Jest
- Gain better tooling, coverage reports, watch mode
- Rewrite existing tests in new framework
- **Pros:** Better developer experience, industry standard
- **Cons:** 2-3 weeks migration effort, learning curve

---

## Testing Infrastructure Recommendations

### Must-Have Improvements
1. **DOM Testing Library**
   - Option 1: jsdom (mature, well-supported)
   - Option 2: happy-dom (faster, lighter)
   - Recommendation: jsdom for compatibility

2. **Async Test Support**
   - Add `async/await` support to test runner
   - Add test timeout configuration
   - Mock `fetch()` for network requests

3. **Timer Mocking**
   - Mock `setTimeout`/`setInterval` for debounce tests
   - Control time in tests

4. **Coverage Reporting**
   - Integrate c8 or nyc for coverage
   - Generate HTML reports
   - Track coverage over time

### Nice-to-Have Improvements
5. **Visual Regression Testing**
   - Screenshot comparison for UI changes
   - Tools: Percy, Chromatic, or Playwright

6. **Integration Testing**
   - Test multi-module workflows
   - End-to-end user journeys

7. **Performance Testing**
   - Measure rendering performance
   - Test with large datasets (1000+ templates)

8. **Accessibility Testing**
   - Automated a11y checks (axe-core)
   - Screen reader testing
   - Keyboard-only navigation validation

---

## Risk Assessment

### High-Risk Untested Areas
1. **Modal Diff Algorithm** (modal.js)
   - Complex LCS implementation
   - **Risk:** Incorrect diffs, performance issues
   - **Impact:** User sees wrong template comparisons
   - **Mitigation:** Phase 4 testing

2. **Keyboard Navigation** (keyboard.js)
   - Viewport calculations, focus management
   - **Risk:** Navigation breaks, focus lost
   - **Impact:** Accessibility failure, user frustration
   - **Mitigation:** Phase 5 testing

3. **Filter & Render Orchestration** (appActions.js)
   - State synchronization, event handling
   - **Risk:** Incorrect filtering, UI desync
   - **Impact:** Wrong search results, broken UX
   - **Mitigation:** Phase 2 testing

### Medium-Risk Untested Areas
4. **Sidebar Rendering** (sidebar.js)
   - Dynamic keyword generation
   - **Risk:** Missing keywords, incorrect counts
   - **Impact:** Degraded filtering UX
   - **Mitigation:** Phase 3 testing

5. **App Initialization** (app.js)
   - Data loading, error handling
   - **Risk:** App fails to load
   - **Impact:** Complete application failure
   - **Mitigation:** Phase 2 testing

---

## Success Metrics

### Quantitative Goals
- [ ] **Test Count:** 300+ total tests (from current 76)
- [ ] **Module Coverage:** 95%+ modules with tests
- [ ] **Line Coverage:** 80%+ lines covered
- [ ] **Function Coverage:** 90%+ functions covered

### Qualitative Goals
- [ ] **Confidence:** Developers confident in refactoring
- [ ] **Documentation:** Testing patterns documented
- [ ] **CI/CD:** All tests run on every PR
- [ ] **Maintainability:** Tests are clear and maintainable
- [ ] **Speed:** Full test suite runs in <10 seconds

---

## Open Questions for Discussion

1. **Timeline:**
   - Is 10-12 weeks acceptable for full coverage?
   - Should we aim for 80% coverage instead (faster)?
   - Do we have bandwidth for this effort?

2. **Approach:**
   - Incremental (Option A) vs. Parallel (Option B)?
   - Should we migrate to modern framework (Option D)?
   - Skip testing simple modules (state.js, utils.js)?

3. **Infrastructure:**
   - Which DOM testing library (jsdom vs. happy-dom)?
   - Do we need visual regression testing?
   - Should we add integration/E2E tests?

4. **Priorities:**
   - Are the phase priorities correct?
   - Should we focus on modal.js and keyboard.js first?
   - Any modules we can skip entirely?

5. **Resources:**
   - Who will work on this?
   - Can we allocate 50% time or 100% time?
   - Do we need external help/consultation?

6. **Maintenance:**
   - How do we ensure new code has tests?
   - Do we need coverage thresholds in CI?
   - Should we require tests for all new features?

---

## Next Steps

1. **Review & Discuss** this plan with the team
2. **Decide on approach** (A, B, C, or D)
3. **Allocate resources** and set timeline
4. **Create tasks** in project management tool
5. **Start with Phase 1** (infrastructure setup)
6. **Iterate and adjust** based on learnings

---

## Conclusion

The Lima Catalog frontend has a solid foundation with 56% module coverage on core logic (data, filtering, formatting). The main gap is testing for UI interaction modules (modal, keyboard, sidebar, app actions), which represent 81% of the codebase by size.

This plan provides a **phased approach** to reach 95%+ coverage over 10-12 weeks, with clear priorities, effort estimates, and risk assessments. The recommended path is **Option A (Incremental)** with infrastructure improvements first, followed by systematic testing of untested modules in order of priority.

**Key Recommendation:** Start with Phase 1 (infrastructure) and Phase 2 (core UI actions) to get the biggest risk reduction in the shortest time (4 weeks, ~75% coverage).
