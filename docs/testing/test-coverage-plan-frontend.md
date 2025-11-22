# Frontend Test Coverage Improvement Plan

**Current Overall Coverage: ~56% (5 of 9 modules)**
**Target Coverage: 95%+**

Last Updated: 2025-11-22

## Coverage Status by Module

### ✅ Excellent Coverage (100%)
- [x] `data.js` - 100% - Complete (7 tests: JSONL parsing, error handling)
- [x] `filters.js` - 100% - Complete (24+ tests: filtering, sorting, keyword counts)
- [x] `templateCard.js` - 100% - Complete (18+ tests: display names, HTML escaping, formatting)
- [x] `theme.js` - 100% - Complete (11 tests: theme management, localStorage, system preferences)
- [x] `urlHelpers.js` - 100% - Complete (4 tests: URL generation, GitHub URL parsing)

**Total Tested:** 76 tests across 5 modules

### ❌ No Coverage (0%)
- [ ] `modal.js` - 45KB - Very high complexity - CRITICAL (preview, diff algorithm, YAML loading)
- [ ] `keyboard.js` - 43KB - Very high complexity - CRITICAL (navigation, accessibility)
- [ ] `sidebar.js` - 28KB - High complexity - HIGH PRIORITY (keyword cloud, category rendering)
- [ ] `appActions.js` - 10KB - High complexity - HIGH PRIORITY (filtering orchestration, event handling)
- [ ] `app.js` - 4KB - Medium complexity - MEDIUM PRIORITY (initialization, setup)
- [ ] `state.js` - 1.6KB - Low complexity - LOW PRIORITY (simple getters/setters)
- [ ] `utils.js` - 2.3KB - Low complexity - LOW PRIORITY (debounce, focus trap)
- [ ] `config.js` - 0.2KB - No logic - N/A (constants only)

**Total Untested:** 133.9KB (81% of codebase by size)

---

## Recent Progress Summary

### ✅ Initial Implementation (Current State)
- **Core Logic Modules:** 100% coverage on data parsing, filtering, and formatting
- **Test Framework:** Custom lightweight framework working in Node.js and browser
- **76 tests passing** with excellent patterns established

### 🎯 Recommended Next Steps
**Phase 1 (Foundation)** - Infrastructure + simple modules to enable all future testing

---

## Implementation Roadmap

### Phase 1: Foundation & Infrastructure ✅ TARGET: 65% COVERAGE
**Timeline:** 1-2 weeks | **Effort:** 5 days | **Priority:** CRITICAL

#### 1.1 Enhance Test Framework
- [ ] Add DOM testing support (jsdom or happy-dom for Node.js)
- [ ] Add async/await test support
- [ ] Add timer mocking utilities (for debounce tests)
- [ ] Add fetch mocking for network requests
- [ ] Add basic coverage reporting (line/function coverage)
- [ ] Document testing patterns and best practices

**Estimated Effort:** 3-4 days

#### 1.2 Test Simple Modules

##### state.js (Low Complexity)
- [ ] Test `getTemplates` / `setTemplates`
- [ ] Test `getFilteredTemplates` / `setFilteredTemplates`
- [ ] Test `toggleKeywordSelection` (add/remove/check)
- [ ] Test `clearKeywordSelection`
- [ ] Test `hasKeywordSelection`
- [ ] Test `setCategorySelection` / `getCategorySelection`
- [ ] Test `toggleCategorySelection`
- [ ] Test `setFocusedTemplate` / `getFocusedTemplate`
- [ ] Test `toggleDebugMode` / `isDebugMode`
- [ ] Test `clearAllSelections`
- [ ] Test state persistence across operations
- [ ] Test multiple keyword selections (Set behavior)

**Estimated Tests:** 15-20
**Estimated Effort:** 0.5 day

##### utils.js (Low Complexity)
- [ ] Test `debounce()` basic functionality
- [ ] Test `debounce()` cancellation behavior
- [ ] Test `debounce()` immediate execution option
- [ ] Test `debounce()` with multiple rapid calls
- [ ] Test `debounce()` with different delays
- [ ] Test `trapFocus()` cycles through elements
- [ ] Test `trapFocus()` wraps at end
- [ ] Test `trapFocus()` handles Shift+Tab backwards
- [ ] Test `trapFocus()` with single focusable element
- [ ] Test `trapFocus()` with no focusable elements

**Estimated Tests:** 10-12
**Estimated Effort:** 1 day

**Phase 1 Total:** ~5 days, +25-32 tests, **~65% module coverage**

---

### Phase 2: Core UI Actions ✅ TARGET: 75% COVERAGE
**Timeline:** 2 weeks | **Effort:** 8 days | **Priority:** HIGH

#### 2.1 Test appActions.js (High Priority - Core Orchestration)

##### filterAndRender()
- [ ] Test applies all active filters correctly
- [ ] Test updates URL parameters
- [ ] Test renders filtered template cards
- [ ] Test preserves focus when `preserveFocus=true`
- [ ] Test updates sidebar counts
- [ ] Test updates document title with result count
- [ ] Test handles empty filter results
- [ ] Test handles all templates filtered out

##### applyFiltersFromURL()
- [ ] Test parses keyword URL parameter (single)
- [ ] Test parses multiple keywords from URL
- [ ] Test parses category URL parameter
- [ ] Test parses sort URL parameter
- [ ] Test parses search URL parameter
- [ ] Test handles malformed URL parameters
- [ ] Test handles invalid keyword values
- [ ] Test handles invalid category values
- [ ] Test sets UI state correctly (checkboxes, dropdown)
- [ ] Test triggers re-render after applying

##### updateSidebarOnly()
- [ ] Test updates keyword cloud
- [ ] Test updates category list
- [ ] Test preserves template cards (no re-render)
- [ ] Test updates counts correctly

##### Toggle Handlers
- [ ] Test `handleKeywordToggle()` adds keyword
- [ ] Test `handleKeywordToggle()` removes keyword
- [ ] Test `handleKeywordToggle()` multi-select behavior
- [ ] Test `handleCategoryToggle()` sets category
- [ ] Test `handleCategoryToggle()` clears on re-select
- [ ] Test `handleCategoryToggle()` single-select behavior

##### updateSortDropdown()
- [ ] Test updates dropdown to match current sort
- [ ] Test handles missing sort option (fallback)
- [ ] Test sets default when sort is invalid

**Estimated Tests:** 30-40
**Estimated Effort:** 4-5 days

#### 2.2 Test app.js (Medium Priority - Initialization)

##### initialize()
- [ ] Test loads data successfully
- [ ] Test applies URL filters after data load
- [ ] Test renders initial UI
- [ ] Test handles data loading errors gracefully
- [ ] Test shows error message on failure
- [ ] Test async initialization order

##### setupEventListeners()
- [ ] Test search input listener registered
- [ ] Test sort dropdown listener registered
- [ ] Test clear keywords button listener
- [ ] Test official/community toggle listeners
- [ ] Test show duplicates toggle listener
- [ ] Test listeners trigger correct actions

##### Event Handlers
- [ ] Test search input triggers debounced filter
- [ ] Test search debounce timing (300ms)
- [ ] Test sort change triggers re-render
- [ ] Test sort maintains current filters
- [ ] Test clear keywords resets selections
- [ ] Test clear keywords triggers re-render
- [ ] Test official filter triggers re-render
- [ ] Test community filter triggers re-render
- [ ] Test show duplicates toggle works

**Estimated Tests:** 20-25
**Estimated Effort:** 3-4 days

**Phase 2 Total:** ~8 days, +50-65 tests, **~75% module coverage**

---

### Phase 3: Sidebar & Rendering ✅ TARGET: 82% COVERAGE
**Timeline:** 1-2 weeks | **Effort:** 6 days | **Priority:** HIGH

#### 3.1 Test sidebar.js (High Priority - Complex DOM Rendering)

##### renderKeywordCloud()
- [ ] Test renders keywords with correct counts
- [ ] Test sorts keywords by count descending
- [ ] Test limits to MAX_KEYWORDS_DISPLAY (50)
- [ ] Test shows "Show more" link when needed
- [ ] Test hides "Show more" when under limit
- [ ] Test expands keyword list on "Show more" click
- [ ] Test highlights selected keywords (CSS class)
- [ ] Test excludes specified keywords
- [ ] Test handles empty keyword list
- [ ] Test handles single keyword
- [ ] Test event listeners attached to keywords
- [ ] Test keyword click triggers toggle

##### renderCategories()
- [ ] Test renders all categories alphabetically
- [ ] Test shows correct count per category
- [ ] Test highlights selected category
- [ ] Test shows "All" option with total count
- [ ] Test "All" selected by default
- [ ] Test handles empty category list
- [ ] Test event listeners attached to categories
- [ ] Test category click triggers toggle

##### updateSidebar()
- [ ] Test updates both keywords and categories
- [ ] Test preserves scroll position
- [ ] Test preserves focus when requested
- [ ] Test handles rapid consecutive updates

##### Dynamic Keywords
- [ ] Test shows dynamic keywords for focused template
- [ ] Test includes org keywords (e.g., "lima-vm")
- [ ] Test includes repo keywords (e.g., "lima-vm/lima")
- [ ] Test excludes already-selected keywords
- [ ] Test handles templates without focused state
- [ ] Test handles templates without org/repo
- [ ] Test dynamic keyword section visibility

##### Event Handling & Focus
- [ ] Test keyword click updates selection
- [ ] Test category click updates selection
- [ ] Test Enter key activates keywords
- [ ] Test Enter key activates categories
- [ ] Test Space key activates keywords
- [ ] Test keyboard navigation with Tab
- [ ] Test focus indicators visible

**Estimated Tests:** 35-45
**Estimated Effort:** 5-6 days

**Phase 3 Total:** ~6 days, +35-45 tests, **~82% module coverage**

---

### Phase 4: Modal & Preview ✅ TARGET: 90% COVERAGE
**Timeline:** 2-3 weeks | **Effort:** 12 days | **Priority:** HIGH (Largest Module)

#### 4.1 Test modal.js (High Priority - Very High Complexity)

##### Modal State Management
- [ ] Test `openPreviewModal()` opens modal
- [ ] Test `openPreviewModal()` adds CSS class to body
- [ ] Test `openPreviewModal()` sets up focus trap
- [ ] Test `closePreviewModal()` closes modal
- [ ] Test `closePreviewModal()` cleans up state
- [ ] Test `closePreviewModal()` removes CSS class
- [ ] Test `closePreviewModal()` restores focus
- [ ] Test Escape key closes modal
- [ ] Test clicking backdrop closes modal
- [ ] Test multiple modals don't interfere

##### YAML Content Loading
- [ ] Test `loadTemplateContent()` fetches YAML successfully
- [ ] Test shows loading indicator during fetch
- [ ] Test displays YAML content after load
- [ ] Test syntax highlighting applied (Highlight.js)
- [ ] Test handles network errors gracefully
- [ ] Test shows error message on fetch failure
- [ ] Test caches loaded content
- [ ] Test reuses cached content on re-open
- [ ] Test handles empty YAML files
- [ ] Test handles malformed YAML
- [ ] Test handles very large YAML files
- [ ] Test timeout for slow networks
- [ ] Test handles 404 errors
- [ ] Test handles rate limit errors

##### Diff Generation (Complex Algorithm)
- [ ] Test `generateDiff()` creates unified diff format
- [ ] Test LCS algorithm finds correct longest common subsequence
- [ ] Test diff shows added lines with "+" prefix
- [ ] Test diff shows removed lines with "-" prefix
- [ ] Test diff shows unchanged context lines
- [ ] Test diff shows modified lines correctly
- [ ] Test handles identical files (no differences)
- [ ] Test handles completely different files
- [ ] Test handles line ending differences
- [ ] Test handles whitespace-only changes
- [ ] Test line numbers accurate in diff
- [ ] Test context lines (3 lines before/after)
- [ ] Test diff performance with large files (1000+ lines)
- [ ] Test diff with empty file vs. populated file
- [ ] Test diff coloring (added/removed/context)

##### Duplicate Templates
- [ ] Test shows duplicate templates in modal
- [ ] Test generates diffs between duplicates
- [ ] Test shows all duplicates (not just first)
- [ ] Test links to duplicate template sources
- [ ] Test duplicate section visible when duplicates exist
- [ ] Test duplicate section hidden when no duplicates
- [ ] Test clicking duplicate template opens it

##### Similar Templates
- [ ] Test shows similar templates section
- [ ] Test compares YAML content with similar templates
- [ ] Test shows similarity indicator
- [ ] Test links to similar template sources
- [ ] Test similar section hidden when none found
- [ ] Test clicking similar template opens it

##### URL Handling
- [ ] Test opening modal updates URL with `?preview=<id>`
- [ ] Test URL parameter persists on reload
- [ ] Test loads preview from URL on page load
- [ ] Test handles invalid preview IDs gracefully
- [ ] Test "Copy URL" button copies correct URL
- [ ] Test clipboard API fallback for older browsers
- [ ] Test browser back button closes modal
- [ ] Test browser forward button reopens modal

##### Focus Management (Accessibility)
- [ ] Test focus trap prevents Tab outside modal
- [ ] Test Tab cycles through modal elements
- [ ] Test Shift+Tab cycles backwards
- [ ] Test first focusable element focused on open
- [ ] Test focus returns to trigger element on close
- [ ] Test focus visible with outline
- [ ] Test handles modal with no focusable elements
- [ ] Test Escape closes and restores focus

##### Keyboard Shortcuts in Modal
- [ ] Test Ctrl+Right opens next template
- [ ] Test Ctrl+Left opens previous template
- [ ] Test navigation wraps at start/end
- [ ] Test navigation updates modal content
- [ ] Test navigation updates URL
- [ ] Test arrow keys disabled in modal (except Ctrl+)

**Estimated Tests:** 56-79
**Estimated Effort:** 10-12 days

**Phase 4 Total:** ~12 days, +56-79 tests, **~90% module coverage**

---

### Phase 5: Keyboard Navigation ✅ TARGET: 95%+ COVERAGE
**Timeline:** 2-3 weeks | **Effort:** 15 days | **Priority:** HIGH (Accessibility Critical)

#### 5.1 Test keyboard.js (High Priority - Very High Complexity, A11y Critical)

##### Keyboard Shortcut Registration
- [ ] Test all shortcuts registered on init
- [ ] Test global shortcuts work from any context
- [ ] Test context-specific shortcuts (modal vs. main)
- [ ] Test shortcuts don't trigger in input fields
- [ ] Test shortcuts don't trigger in textareas
- [ ] Test modifier keys handled (Ctrl, Shift, Alt)
- [ ] Test case-insensitive key matching

##### Arrow Key Navigation (Grid Layout)
- [ ] Test Arrow Right navigates to next template
- [ ] Test Arrow Left navigates to previous template
- [ ] Test Arrow Down navigates to template below
- [ ] Test Arrow Up navigates to template above
- [ ] Test navigation at first template (wrapping)
- [ ] Test navigation at last template (wrapping)
- [ ] Test navigation at grid edges (row wrapping)
- [ ] Test navigation respects grid layout (columns)
- [ ] Test navigation scrolls viewport when needed
- [ ] Test navigation updates focused template in state
- [ ] Test visual focus indicator appears (CSS class)
- [ ] Test focus indicator removed from previous template

##### Grid Position Calculations
- [ ] Test `getTemplatePosition()` calculates row correctly
- [ ] Test `getTemplatePosition()` calculates column correctly
- [ ] Test `getTemplatePosition()` with 3-column grid
- [ ] Test `getTemplatePosition()` with 2-column grid (responsive)
- [ ] Test `getTemplatePosition()` with 1-column grid (mobile)
- [ ] Test `getTemplateAtPosition()` finds correct template
- [ ] Test `getTemplateAtPosition()` with invalid position
- [ ] Test grid calculations with filtered results (gaps)
- [ ] Test grid calculations with variable card widths
- [ ] Test grid calculations after window resize

##### Page Navigation
- [ ] Test Page Down scrolls one viewport height
- [ ] Test Page Up scrolls one viewport height
- [ ] Test Home goes to first template
- [ ] Test Home focuses first template
- [ ] Test End goes to last template
- [ ] Test End focuses last template
- [ ] Test Page Down at bottom (no overflow scroll)
- [ ] Test Page Up at top (no overflow scroll)

##### Special Keyboard Shortcuts
- [ ] Test "/" focuses search input
- [ ] Test "/" clears search input (optional)
- [ ] Test "/" from any context works
- [ ] Test "?" opens help modal
- [ ] Test "?" shows keyboard shortcuts
- [ ] Test "@" toggles debug mode
- [ ] Test "@" shows debug info in UI
- [ ] Test Enter opens preview for focused template
- [ ] Test Enter with no focused template (no action)
- [ ] Test Escape closes modals
- [ ] Test Escape returns focus correctly

##### Sidebar Navigation
- [ ] Test Tab navigates to sidebar
- [ ] Test Tab cycles through keywords
- [ ] Test Tab cycles through categories
- [ ] Test Shift+Tab navigates backwards
- [ ] Test Enter activates keyword
- [ ] Test Enter activates category
- [ ] Test Space activates keyword
- [ ] Test Space activates category

##### Modal Context Navigation
- [ ] Test Ctrl+Right opens next template in modal
- [ ] Test Ctrl+Left opens previous template in modal
- [ ] Test modal navigation wraps at start/end
- [ ] Test modal navigation updates content
- [ ] Test modal navigation updates URL
- [ ] Test modal navigation preserves scroll position
- [ ] Test arrow keys disabled in modal (except Ctrl+)
- [ ] Test Tab still works in modal (focus trap)

##### Focus Management & Viewport
- [ ] Test focused template scrolled into view
- [ ] Test scroll position centered on focused element
- [ ] Test scroll respects smooth scrolling
- [ ] Test focus preserved during filter changes
- [ ] Test focus cleared when template filtered out
- [ ] Test focus moves to nearest template when current removed
- [ ] Test focus visible with CSS outline
- [ ] Test focus indicator high contrast

##### Edge Cases & Performance
- [ ] Test navigation with 1 template
- [ ] Test navigation with 0 templates (empty state)
- [ ] Test navigation with 1000+ templates (performance)
- [ ] Test rapid key presses (debouncing if needed)
- [ ] Test keyboard shortcuts don't interfere with each other
- [ ] Test navigation state persists across renders

**Estimated Tests:** 66-90
**Estimated Effort:** 12-15 days

**Phase 5 Total:** ~15 days, +66-90 tests, **~95%+ module coverage**

---

## Alternative Approaches

### Option A: Incremental (Recommended)
- Follow Phases 1-5 sequentially
- Validate each phase before proceeding
- **Timeline:** 10-12 weeks
- **Pros:** Lower risk, continuous value delivery
- **Cons:** Longer timeline

### Option B: Parallel Development
- Work on multiple phases simultaneously (requires 2-3 developers)
- **Timeline:** 6-8 weeks
- **Pros:** Faster completion
- **Cons:** Higher coordination overhead

### Option C: Risk-Based Prioritization
- Focus only on modal.js and keyboard.js (highest risk)
- Skip state.js, utils.js (low value)
- **Timeline:** 6-8 weeks
- **Target:** 80% coverage (instead of 95%)
- **Pros:** Faster, focuses on critical paths
- **Cons:** Leaves gaps

### Option D: Migrate to Modern Framework First
- Migrate from custom framework to Vitest or Jest
- Gain better tooling and coverage reports
- **Timeline:** +2-3 weeks upfront, then follow phases
- **Pros:** Better developer experience, industry standard
- **Cons:** Migration effort, rewrite existing tests

---

## Testing Infrastructure

### Must-Have Improvements
1. **DOM Testing Library**
   - Options: jsdom (mature) or happy-dom (faster)
   - Recommendation: jsdom for compatibility

2. **Async Test Support**
   - Add `async/await` to test runner
   - Add test timeout configuration

3. **Mock Utilities**
   - Mock `fetch()` for network requests
   - Mock timers (`setTimeout`/`setInterval`)
   - Mock `window.matchMedia`
   - Mock localStorage (already done)

4. **Coverage Reporting**
   - Integrate c8 or nyc for coverage
   - Generate HTML reports
   - Track coverage over time

### Nice-to-Have Improvements
5. **Visual Regression Testing**
   - Screenshot comparison for UI changes
   - Tools: Percy, Chromatic, or Playwright

6. **Integration Testing**
   - End-to-end user journeys
   - Multi-module workflows

7. **Performance Testing**
   - Rendering performance metrics
   - Test with large datasets (1000+ templates)

8. **Accessibility Testing**
   - Automated a11y checks (axe-core)
   - Screen reader testing
   - Keyboard-only navigation validation

---

## Risk Assessment

### Critical Untested Areas (High Impact)
1. **modal.js** - Diff algorithm (45KB)
   - **Risk:** Incorrect diffs, performance issues with large files
   - **Impact:** Users see wrong template comparisons, slow UI
   - **Mitigation:** Phase 4 comprehensive testing

2. **keyboard.js** - Navigation & focus (43KB)
   - **Risk:** Navigation breaks, focus lost, accessibility failure
   - **Impact:** App unusable without mouse, fails WCAG standards
   - **Mitigation:** Phase 5 comprehensive testing

3. **appActions.js** - Filter orchestration (10KB)
   - **Risk:** Incorrect filtering, UI state desync
   - **Impact:** Wrong search results, broken user experience
   - **Mitigation:** Phase 2 comprehensive testing

### Important Untested Areas (Medium Impact)
4. **sidebar.js** - Keyword rendering (28KB)
   - **Risk:** Missing keywords, incorrect counts
   - **Impact:** Degraded filtering UX, confusion
   - **Mitigation:** Phase 3 comprehensive testing

5. **app.js** - Initialization (4KB)
   - **Risk:** App fails to load
   - **Impact:** Complete application failure
   - **Mitigation:** Phase 2 testing with error scenarios

---

## Success Metrics

### Quantitative Goals
- [ ] **Test Count:** 300+ total tests (from current 76)
- [ ] **Module Coverage:** 95%+ modules with tests (8 of 9 modules)
- [ ] **Line Coverage:** 80%+ lines covered (requires coverage tool)
- [ ] **Function Coverage:** 90%+ functions covered

### Qualitative Goals
- [ ] **Confidence:** Developers confident in refactoring any module
- [ ] **Documentation:** Testing patterns well-documented for new tests
- [ ] **CI/CD:** All tests run on every PR, failures block merge
- [ ] **Maintainability:** Tests clear, readable, easy to debug
- [ ] **Speed:** Full test suite runs in <10 seconds

### Regression Prevention
- [ ] No bugs reported in tested modules
- [ ] Refactoring doesn't break functionality
- [ ] New features don't break existing tests

---

## Testing Patterns

### DOM Testing Pattern
```javascript
// Mock DOM elements for testing
const mockDOM = {
  getElementById: (id) => ({ /* mock element */ }),
  querySelector: (selector) => ({ /* mock element */ }),
  // ...
};

// Test renders correctly
runner.test('renderKeywordCloud displays keywords', () => {
  const keywords = [{ name: 'linux', count: 10 }];
  const container = mockDOM.getElementById('keyword-cloud');
  renderKeywordCloud(keywords, container);

  assert.ok(container.innerHTML.includes('linux'));
  assert.ok(container.innerHTML.includes('10'));
});
```

### Async Testing Pattern
```javascript
// Test async operations
runner.test('loadTemplateContent fetches YAML', async () => {
  const mockFetch = (url) => Promise.resolve({
    ok: true,
    text: () => Promise.resolve('images:\n  - ubuntu')
  });
  global.fetch = mockFetch;

  const content = await loadTemplateContent('https://example.com/template.yaml');
  assert.ok(content.includes('ubuntu'));
});
```

### Event Testing Pattern
```javascript
// Test event handlers
runner.test('keyword click toggles selection', () => {
  const mockEvent = {
    target: { dataset: { keyword: 'docker' } },
    preventDefault: () => {}
  };

  handleKeywordToggle(mockEvent);
  assert.ok(State.hasKeywordSelection('docker'));
});
```

### Focus Management Pattern
```javascript
// Test focus behavior
runner.test('modal traps focus within container', () => {
  const modal = { querySelectorAll: () => [btn1, btn2, btn3] };
  const mockEvent = { shiftKey: false, target: btn3 };

  trapFocus(modal, mockEvent);
  assert.equal(document.activeElement, btn1); // wrapped to start
});
```

---

## Open Questions for Discussion

1. **Timeline & Resources:**
   - Is 10-12 weeks acceptable for full coverage?
   - Can we allocate 50% or 100% of developer time?
   - Should we hire/contract additional help?

2. **Approach Selection:**
   - Option A (Incremental) vs. Option B (Parallel)?
   - Should we migrate to Vitest/Jest first (Option D)?
   - Or focus on high-risk modules only (Option C)?

3. **Infrastructure Decisions:**
   - jsdom vs. happy-dom for DOM testing?
   - Which coverage tool (c8, nyc, or framework built-in)?
   - Do we need visual regression testing now or later?

4. **Priorities & Scope:**
   - Are phase priorities correct (foundation → actions → sidebar → modal → keyboard)?
   - Should we skip testing state.js and utils.js entirely?
   - Do we need integration/E2E tests in addition to unit tests?

5. **Quick Wins:**
   - Start with Phase 1+2 only (4 weeks, 75% coverage) then reassess?
   - Or commit to full roadmap upfront?

6. **Maintenance:**
   - Should we require tests for all new features going forward?
   - Do we need coverage thresholds in CI (e.g., "fail if coverage drops below 70%")?
   - How do we handle legacy code vs. new code?

---

## Next Steps

1. **Review & Discuss** this plan with stakeholders
2. **Decide on approach** (A, B, C, or D)
3. **Allocate resources** and confirm timeline
4. **Create project tasks** in tracking system
5. **Start Phase 1** (infrastructure setup)
6. **Iterate and adjust** based on learnings

---

## Summary

The Lima Catalog frontend has strong test coverage (100%) on core logic modules (data parsing, filtering, formatting, theming) representing **56% module coverage** across 76 tests.

**Main Gap:** The 4 largest and most complex modules (modal.js, keyboard.js, sidebar.js, appActions.js) representing **81% of the codebase by size** have zero test coverage.

**Recommended Path:**
- **Quick Win:** Phase 1+2 (4 weeks, 75% coverage, ~125 tests) validates infrastructure and covers core UI actions
- **Full Coverage:** All 5 phases (10-12 weeks, 95%+ coverage, 300+ tests) provides comprehensive protection

**Key Risks Addressed:**
- ✅ Modal diff algorithm correctness
- ✅ Keyboard navigation and accessibility
- ✅ Filter/render orchestration integrity
- ✅ Sidebar rendering accuracy

This plan provides a clear roadmap to industry-standard test coverage while prioritizing high-risk, high-impact modules first.
