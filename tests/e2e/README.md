# End-to-End Tests

This directory contains Playwright E2E tests for the lima-catalog web application.

## Quick Start

```bash
# Setup (first time only)
mkdir -p /root/tmp
npm install
npx playwright install chromium

# Run all tests
npm run test:e2e

# Run specific test file
npx playwright test tests/e2e/search.spec.js

# Run in headed mode (see browser)
npm run test:e2e:headed
```

## Test Files

- **search.spec.js** (17 tests) - Search, filtering, sorting, URL parameters
- **categories.spec.js** (24 tests) - Category and keyword filtering
- **modal.spec.js** (22 tests) - Modal interactions and content loading
- **keyboard.spec.js** (22 tests) - Keyboard navigation and shortcuts
- **theme.spec.js** (15 tests) - Theme switching and persistence
- **visual.spec.js** (25 tests) - Visual regression (currently skipped)

**Total:** 125 tests (100 active, 25 skipped)

## Documentation

For detailed debugging guides, troubleshooting, and writing tests, see:

📖 **[Debugging with Playwright](../../docs/testing/debugging-with-playwright.md)**

This guide covers:
- Environment setup and configuration
- Running and debugging tests
- Common issues and solutions
- Writing custom debug scripts
- Interactive browser automation

## Common Issues

### Browser crashes immediately

**Solution:** Create the required temp directory:
```bash
mkdir -p /root/tmp
```

### Tests timeout

**Solution:** Ensure web server is running and data is built:
```bash
make build
```

### Visual tests fail

Visual regression tests are currently skipped. To enable them, update `visual.spec.js` and generate baseline screenshots.

## Contributing

When adding new tests:
1. Follow existing patterns in test files
2. Use descriptive test names
3. Add comments for complex assertions
4. Update test counts in this README
5. Run tests locally before committing
