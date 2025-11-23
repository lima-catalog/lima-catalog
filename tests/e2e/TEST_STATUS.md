# Playwright E2E Test Status

**Last Updated:** 2025-11-23
**Phase:** Phase 2 Implementation
**Total Tests:** 125

## Setup Requirements

### Critical Fix Required
The tests require `/root/tmp` directory to exist for Chromium's shared memory:

```bash
mkdir -p /root/tmp
```

Without this directory, all tests fail with:
```
Creating shared memory in /root/tmp/.org.chromium.Chromium.xxx failed: No such file or directory
```

This issue is documented in `docs/testing/PLAYWRIGHT_DEBUGGING.md`.

## Test Status by File

### ✅ search.spec.js (17 tests)
- **Status:** Most tests passing
- **Known Issues:** None confirmed yet
- **Notes:** Search, filtering, sorting, and URL parameter tests

### ✅ modal.spec.js (22 tests)
- **Status:** Most tests passing
- **Known Issues:** Need to verify network error handling
- **Notes:** Modal open/close, content loading, keyboard navigation

### ✅ categories.spec.js (24 tests)
- **Status:** Most tests passing
- **Known Issues:** Need to verify URL parameter tests
- **Notes:** Category/keyword filtering, URL state management

### ✅ keyboard.spec.js (22 tests)
- **Status:** Most tests passing
- **Known Issues:** Some navigation tests may fail depending on grid layout
- **Notes:** Arrow keys, Enter/Space, Escape, Tab navigation, shortcuts

### ⚠️ theme.spec.js (15 tests)
- **Status:** Tests need updates
- **Known Issues:**
  - Tests assume single `#theme-toggle` button
  - Actual implementation uses three `.theme-option` buttons (light/auto/dark)
- **Fix Required:** Update selectors to match actual HTML structure
- **Actual HTML:**
  ```html
  <button class="theme-option" data-theme="light">...</button>
  <button class="theme-option active" data-theme="auto">...</button>
  <button class="theme-option" data-theme="dark">...</button>
  ```

### ⚠️ visual.spec.js (25 tests)
- **Status:** Will fail on first run (expected)
- **Known Issues:**
  - No baseline screenshots exist yet
  - First run will create baselines
  - Subsequent runs will compare against baselines
- **Fix Required:**
  1. Run tests once to generate baselines
  2. Review baseline screenshots
  3. Commit baselines to git
- **Notes:** Visual regression requires manual review of initial baselines

## Running Tests Locally

```bash
# Ensure directory exists (run once)
mkdir -p /root/tmp

# Run all tests
make test-e2e

# Run specific file
npx playwright test tests/e2e/search.spec.js

# Run specific test
npx playwright test -g "loads and displays templates"

# Run with headed browser (for debugging)
npx playwright test --headed

# Update visual baselines (after reviewing)
npx playwright test --update-snapshots
```

## Known Test Issues

### Theme Tests Need Fixes
All 15 theme tests need to be updated to use the correct selectors:

**Change from:**
```javascript
const themeToggle = page.locator('#theme-toggle');
await themeToggle.click();
```

**Change to:**
```javascript
// Click specific theme button
await page.locator('.theme-option[data-theme="dark"]').click();

// Or check active theme
const activeTheme = page.locator('.theme-option.active');
const themeValue = await activeTheme.getAttribute('data-theme');
```

### Visual Tests - Baseline Generation
All 25 visual tests will fail on first run. This is expected behavior:

1. Run tests to generate baselines
2. Review screenshots in `tests/e2e/**/*.spec.js-snapshots/`
3. If screenshots look correct, commit them
4. Future runs will compare against these baselines

## Next Steps

1. **Fix theme.spec.js** - Update to match actual HTML structure
2. **Generate visual baselines** - Run visual tests once and review
3. **Verify all tests** - Run full suite and fix any remaining issues
4. **Document failures** - Update this file with any persistent issues
5. **CI Integration** - Ensure tests run reliably in GitHub Actions

## CI/CD Considerations

The GitHub Actions workflow may need updates:
- Ensure `/root/tmp` directory is created in CI
- Handle visual test baselines properly
- May need to disable visual tests in CI if screenshots differ between environments
