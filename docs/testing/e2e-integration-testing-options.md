# End-to-End and Integration Testing Options

**Last Updated:** 2025-11-22
**Status:** Analysis and Recommendations

## Executive Summary

This document analyzes available options for implementing end-to-end (e2e) and integration tests for the lima-catalog project, focusing on tools that work in both local development environments (including VM sandboxes) and CI/CD pipelines like GitHub Actions.

**Key Findings:**
- ✅ Playwright is **already available** in the VM sandbox (v1.56.1)
- ✅ Multiple viable options exist for both Node.js and Go
- ✅ All recommended tools work well in GitHub Actions CI
- ⚠️ Browser installation requires handling system dependencies

---

## Table of Contents

1. [Current Testing State](#current-testing-state)
2. [Testing Options Overview](#testing-options-overview)
3. [Option 1: Playwright (Node.js)](#option-1-playwright-nodejs)
4. [Option 2: Puppeteer (Node.js)](#option-2-puppeteer-nodejs)
5. [Option 3: Cypress (Node.js)](#option-3-cypress-nodejs)
6. [Option 4: chromedp (Go)](#option-4-chromedp-go)
7. [Option 5: rod (Go)](#option-5-rod-go)
8. [Option 6: Go httptest (No Browser)](#option-6-go-httptest-no-browser)
9. [Comparison Matrix](#comparison-matrix)
10. [Recommendations](#recommendations)
11. [Implementation Considerations](#implementation-considerations)
12. [Resources](#resources)

---

## Current Testing State

### Existing Test Infrastructure

**Go Backend:**
- Unit tests: `go test ./...`
- Current coverage: ~49.7% (see [test-coverage-plan.md](test-coverage-plan.md))
- Integration tests: `scripts/test-integration.sh` (CLI tool testing with real GitHub API)

**JavaScript Frontend:**
- Custom lightweight test framework (no external dependencies)
- Tests run in both Node.js and browser environments
- Test files: `web/js/*.test.js`
- Test runner: `test.js` (Node.js) or `web/tests.html` (browser)

**CI Pipeline (`.github/workflows/ci.yml`):**
- Go tests with race detection
- Go vet and golangci-lint
- Build verification
- No frontend e2e tests currently

### Testing Gap

**Missing:** End-to-end tests for the web UI that verify:
- Template search and filtering work correctly
- Template preview modal functions properly
- URL helpers generate correct GitHub URLs
- Keyboard navigation works as expected
- Theme switching persists across page loads
- Data loading and parsing from JSONL files
- Integration between all frontend components

---

## Testing Options Overview

| Tool | Language | Browser Support | VM Sandbox | CI Ready | Learning Curve |
|------|----------|----------------|------------|----------|----------------|
| **Playwright** | Node.js/TS | Chrome, Firefox, Safari | ✅ Installed | ✅ Official Action | Medium |
| **Puppeteer** | Node.js/TS | Chrome/Chromium | ✅ Easy install | ✅ Well supported | Low |
| **Cypress** | Node.js/TS | Chrome, Firefox, Edge | ✅ Easy install | ✅ Well supported | Medium |
| **chromedp** | Go | Chrome/Chromium | ✅ Easy install | ✅ Works well | Medium |
| **rod** | Go | Chrome/Chromium | ✅ Easy install | ✅ Works well | Low-Medium |
| **httptest** | Go | None (HTTP only) | ✅ Built-in | ✅ Built-in | Low |

---

## Option 1: Playwright (Node.js)

### Overview

[Playwright](https://playwright.dev/) is Microsoft's modern browser automation framework, explicitly designed for end-to-end testing. It's **already installed** in the VM sandbox (v1.56.1).

### Pros

✅ **Already available in VM sandbox** - no additional installation needed
✅ **Cross-browser support** - Chrome, Firefox, Safari (WebKit)
✅ **Excellent CI support** - Official [GitHub Action](https://github.com/marketplace/actions/setup-playwright)
✅ **Modern API** - async/await, auto-waiting for elements
✅ **Multiple languages** - JavaScript, TypeScript, Python, .NET, Java
✅ **Built-in features** - screenshots, videos, traces, code generation
✅ **Active development** - Regular updates from Microsoft
✅ **Best-in-class stability** - Outperforms Cypress in stability tests

### Cons

❌ **Larger installation** - Browsers download ~300MB per browser
❌ **System dependencies** - May need `--with-deps` for full functionality (can fail in sandboxed environments)
⚠️ **Learning curve** - More features = more to learn

### VM Sandbox Status

```bash
$ playwright --version
Version 1.56.1

$ playwright install chromium
# Downloads browser to ~/.cache/ms-playwright/chromium-1194/
# Note: --with-deps may fail in sandbox, but browser binaries work fine
```

**Verification:** Browser installation works, but system dependency installation (`--with-deps`) may fail due to sandbox restrictions. This is acceptable for headless testing.

### Basic Example

```javascript
// tests/e2e/basic.spec.js
const { test, expect } = require('@playwright/test');

test('template search filters results', async ({ page }) => {
  // Start local server
  await page.goto('http://localhost:8000');

  // Wait for templates to load
  await page.waitForSelector('.template-card');

  // Count initial templates
  const initialCount = await page.locator('.template-card').count();
  expect(initialCount).toBeGreaterThan(0);

  // Search for "alpine"
  await page.fill('#search-input', 'alpine');

  // Verify filtered results
  await page.waitForSelector('.template-card:has-text("alpine")');
  const filteredCount = await page.locator('.template-card').count();
  expect(filteredCount).toBeLessThan(initialCount);
  expect(filteredCount).toBeGreaterThan(0);
});

test('template preview modal opens', async ({ page }) => {
  await page.goto('http://localhost:8000');
  await page.waitForSelector('.template-card');

  // Click first template
  await page.click('.template-card:first-child .preview-button');

  // Verify modal opens
  await expect(page.locator('.modal')).toBeVisible();
  await expect(page.locator('.modal .yaml-content')).toBeVisible();
});
```

### CI Integration (GitHub Actions)

```yaml
# .github/workflows/e2e.yml
name: E2E Tests

on: [pull_request, push]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '22'

      - name: Install dependencies
        run: npm ci

      - name: Install Playwright browsers
        run: npx playwright install --with-deps chromium

      - name: Run E2E tests
        run: npx playwright test

      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: playwright-report
          path: playwright-report/
```

### Package.json Setup

```json
{
  "devDependencies": {
    "@playwright/test": "^1.56.1"
  },
  "scripts": {
    "test:e2e": "playwright test",
    "test:e2e:headed": "playwright test --headed",
    "test:e2e:debug": "playwright test --debug"
  }
}
```

### When to Choose Playwright

- ✅ Need cross-browser testing (Chrome, Firefox, Safari)
- ✅ Want comprehensive testing features (screenshots, videos, traces)
- ✅ Prefer staying in the Node.js ecosystem
- ✅ Want excellent TypeScript support
- ✅ Need to test complex user interactions
- ✅ Already familiar with modern async/await testing patterns

---

## Option 2: Puppeteer (Node.js)

### Overview

[Puppeteer](https://pptr.dev/) is Google's official Node.js library for controlling Chrome/Chromium. It's lighter and simpler than Playwright but Chrome-only.

### Pros

✅ **Simpler API** - Easier learning curve than Playwright
✅ **Lighter weight** - Only downloads Chromium (~170MB vs Playwright's 300MB+ per browser)
✅ **Official Google project** - Well-maintained and documented
✅ **Fast execution** - Optimized for Chrome DevTools Protocol
✅ **Good CI support** - Works well in GitHub Actions
✅ **Mature ecosystem** - Extensive examples and community resources

### Cons

❌ **Chrome-only** - Cannot test Firefox or Safari
❌ **Fewer built-in features** - No auto-waiting, less sophisticated selectors
❌ **Less stable** - More manual waiting/timing code needed

### Basic Example

```javascript
// tests/e2e/puppeteer-example.js
const puppeteer = require('puppeteer');

(async () => {
  const browser = await puppeteer.launch({ headless: true });
  const page = await browser.newPage();

  await page.goto('http://localhost:8000');

  // Wait for templates to load
  await page.waitForSelector('.template-card');

  // Search for templates
  await page.type('#search-input', 'alpine');
  await page.waitForTimeout(500); // Manual waiting

  // Get filtered results
  const templates = await page.$$('.template-card');
  console.log(`Found ${templates.length} templates`);

  await browser.close();
})();
```

### CI Integration

```yaml
- name: Install Puppeteer
  run: npm install puppeteer

- name: Run tests
  run: node tests/e2e/puppeteer-test.js
```

### When to Choose Puppeteer

- ✅ Only need Chrome/Chromium testing
- ✅ Want simpler, lighter setup
- ✅ Prefer Google's tooling ecosystem
- ✅ Need quick setup for basic testing
- ❌ Don't need cross-browser support

### Resources

- [Comparison: Playwright vs. Puppeteer (2025)](https://www.zenrows.com/blog/playwright-vs-puppeteer)
- [Playwright vs Puppeteer: Which Should You Choose](https://blog.apify.com/playwright-vs-puppeteer/)

---

## Option 3: Cypress (Node.js)

### Overview

[Cypress](https://www.cypress.io/) is a developer-friendly testing framework with an excellent debugging experience but more architectural limitations.

### Pros

✅ **Excellent DX** - Interactive test runner with time-travel debugging
✅ **Real-time reloading** - Tests auto-reload on changes
✅ **Network stubbing** - Easy request/response mocking
✅ **Great documentation** - Extensive guides and examples
✅ **Visual debugging** - See what your tests see

### Cons

❌ **Limited cross-domain support** - Cannot test cross-origin iframes
❌ **Shadow DOM issues** - Struggles with closed shadow DOM
❌ **Single browser per test** - Cannot test multi-tab scenarios
❌ **Slower than Playwright** - Generally underperforms in benchmarks
❌ **More complex CI setup** - Requires additional configuration

### Basic Example

```javascript
// cypress/e2e/templates.cy.js
describe('Template Catalog', () => {
  beforeEach(() => {
    cy.visit('http://localhost:8000');
  });

  it('filters templates by search term', () => {
    cy.get('.template-card').should('have.length.greaterThan', 0);

    cy.get('#search-input').type('alpine');

    cy.get('.template-card')
      .should('have.length.greaterThan', 0)
      .and('have.length.lessThan', 100);
  });

  it('opens template preview modal', () => {
    cy.get('.template-card').first().click();
    cy.get('.modal').should('be.visible');
    cy.get('.modal .yaml-content').should('exist');
  });
});
```

### When to Choose Cypress

- ✅ Want the best debugging experience
- ✅ Prefer interactive test development
- ✅ Team is already familiar with Cypress
- ❌ Don't have cross-domain or shadow DOM requirements
- ⚠️ Understand its architectural limitations

### Resources

- [Playwright vs Cypress: Key Differences (2025)](https://katalon.com/resources-center/blog/playwright-vs-cypress)
- [Cypress vs Playwright in 2025](https://bugbug.io/blog/test-automation-tools/cypress-vs-playwright/)

---

## Option 4: chromedp (Go)

### Overview

[chromedp](https://github.com/chromedp/chromedp) is a faster, simpler way to drive browsers supporting the Chrome DevTools Protocol in Go, with no external dependencies.

### Pros

✅ **Native Go** - Same language as your backend
✅ **No external dependencies** - Pure Go implementation
✅ **Good performance** - Direct DevTools Protocol communication
✅ **Works in CI** - Easy GitHub Actions integration
✅ **Headless by default** - Perfect for CI/CD
✅ **Type-safe** - Compile-time checking

### Cons

❌ **Chrome-only** - No Firefox or Safari support
❌ **Architecture limits** - DOM node ID-based (can be slower for complex operations)
❌ **JSON overhead** - Decodes every browser message (performance impact)
⚠️ **Less mature** - Smaller ecosystem than Playwright/Puppeteer

### Basic Example

```go
// pkg/e2e/catalog_test.go
package e2e

import (
    "context"
    "testing"
    "time"

    "github.com/chromedp/chromedp"
)

func TestTemplateSearch(t *testing.T) {
    // Create context
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()

    // Set timeout
    ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    var templateCount int

    // Run test
    err := chromedp.Run(ctx,
        chromedp.Navigate("http://localhost:8000"),
        chromedp.WaitVisible(".template-card"),
        chromedp.Evaluate(`document.querySelectorAll('.template-card').length`, &templateCount),
    )

    if err != nil {
        t.Fatal(err)
    }

    if templateCount == 0 {
        t.Errorf("Expected templates to be loaded, got 0")
    }
}

func TestTemplateFilter(t *testing.T) {
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()

    ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    var beforeCount, afterCount int

    err := chromedp.Run(ctx,
        chromedp.Navigate("http://localhost:8000"),
        chromedp.WaitVisible(".template-card"),
        chromedp.Evaluate(`document.querySelectorAll('.template-card').length`, &beforeCount),
        chromedp.SendKeys("#search-input", "alpine"),
        chromedp.Sleep(500*time.Millisecond),
        chromedp.Evaluate(`document.querySelectorAll('.template-card').length`, &afterCount),
    )

    if err != nil {
        t.Fatal(err)
    }

    if afterCount >= beforeCount {
        t.Errorf("Expected filtered results (%d) to be less than initial (%d)",
            afterCount, beforeCount)
    }
}
```

### CI Integration

```yaml
# .github/workflows/e2e-go.yml
name: E2E Tests (Go)

on: [pull_request, push]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Install Chrome
        run: |
          wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
          sudo sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list'
          sudo apt-get update
          sudo apt-get install -y google-chrome-stable

      - name: Run E2E tests
        run: go test ./pkg/e2e/... -v
```

### When to Choose chromedp

- ✅ Prefer staying in Go ecosystem
- ✅ Want type safety and compile-time checks
- ✅ Already using Go for backend tests
- ✅ Only need Chrome testing
- ❌ Don't need cross-browser support

### Resources

- [chromedp GitHub](https://github.com/chromedp/chromedp)
- [Running UI Automation Tests with Go and Chrome on GitHub Actions](https://pradappandiyan.medium.com/running-ui-automation-tests-with-go-and-chrome-on-github-actions-1f56d7c63405)

---

## Option 5: rod (Go)

### Overview

[rod](https://github.com/go-rod/rod) is a high-level DevTools Protocol driver for Go, designed for web automation and scraping with better performance than chromedp.

### Pros

✅ **Best Go performance** - Faster and uses less memory than chromedp
✅ **Decode-on-demand** - Only parses needed messages (unlike chromedp)
✅ **Stable architecture** - Remote object ID-based (more consistent)
✅ **High-level API** - Easier to use than chromedp
✅ **Good documentation** - Clear examples and guides
✅ **Active development** - Regular updates and improvements

### Cons

❌ **Chrome-only** - No Firefox or Safari support
⚠️ **Less adoption** - Smaller community than chromedp
⚠️ **Newer project** - Less mature than other options

### Basic Example

```go
// pkg/e2e/rod_test.go
package e2e

import (
    "testing"
    "github.com/go-rod/rod"
    "github.com/go-rod/rod/lib/launcher"
)

func TestTemplateSearchWithRod(t *testing.T) {
    // Launch browser
    l := launcher.New().Headless(true)
    defer l.Cleanup()
    url := l.MustLaunch()

    browser := rod.New().ControlURL(url).MustConnect()
    defer browser.MustClose()

    // Navigate and test
    page := browser.MustPage("http://localhost:8000")

    // Wait for templates
    page.MustWaitLoad()
    page.MustElement(".template-card")

    // Count templates
    templates := page.MustElements(".template-card")
    if len(templates) == 0 {
        t.Error("Expected templates to be loaded")
    }

    // Test search
    page.MustElement("#search-input").MustInput("alpine")
    page.MustWaitIdle()

    filteredTemplates := page.MustElements(".template-card")
    if len(filteredTemplates) >= len(templates) {
        t.Errorf("Expected filtered results (%d) to be less than initial (%d)",
            len(filteredTemplates), len(templates))
    }
}

func TestTemplateModal(t *testing.T) {
    l := launcher.New().Headless(true)
    defer l.Cleanup()
    browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()
    defer browser.MustClose()

    page := browser.MustPage("http://localhost:8000")
    page.MustWaitLoad()

    // Click first template preview button
    page.MustElement(".template-card .preview-button").MustClick()

    // Verify modal is visible
    modal := page.MustElement(".modal")
    if !modal.MustVisible() {
        t.Error("Expected modal to be visible")
    }

    yamlContent := page.MustElement(".modal .yaml-content")
    if yamlContent.MustText() == "" {
        t.Error("Expected YAML content in modal")
    }
}
```

### CI Integration

Same as chromedp - just needs Chrome installed in CI environment. Can also use the chromedp/headless-shell Docker image.

### When to Choose rod

- ✅ Want best Go performance
- ✅ Need high-level API that's easier than chromedp
- ✅ Prefer staying in Go ecosystem
- ✅ Performance is critical
- ❌ Don't need cross-browser support

### Resources

- [rod GitHub](https://github.com/go-rod/rod)
- [Why rod?](https://github.com/go-rod/go-rod.github.io/blob/main/why-rod.md)
- [Golang Headless Browser: Best Tools for Automation](https://latenode.com/blog/golang-headless-browser-best-tools-for-automation)

---

## Option 6: Go httptest (No Browser)

### Overview

Use Go's standard library `net/http/httptest` for integration testing without a browser. Good for testing backend API endpoints and static file serving.

### Pros

✅ **No dependencies** - Part of Go standard library
✅ **Fast execution** - No browser overhead
✅ **Simple setup** - Just write Go tests
✅ **Perfect for APIs** - Test HTTP handlers directly
✅ **Works anywhere** - No browser installation needed

### Cons

❌ **No JavaScript execution** - Cannot test frontend interactivity
❌ **No DOM testing** - Cannot verify UI behavior
❌ **Limited scope** - Only for backend/API testing

### Basic Example

```go
// pkg/server/server_test.go
package server

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestStaticFileServing(t *testing.T) {
    // Create test server
    handler := http.FileServer(http.Dir("../../web"))
    server := httptest.NewServer(handler)
    defer server.Close()

    // Test index.html loads
    resp, err := http.Get(server.URL + "/index.html")
    if err != nil {
        t.Fatalf("Failed to GET /index.html: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status 200, got %d", resp.StatusCode)
    }

    if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
        t.Errorf("Expected text/html content type, got %s", ct)
    }
}

func TestCatalogJSONL(t *testing.T) {
    handler := http.FileServer(http.Dir("../../data"))
    server := httptest.NewServer(handler)
    defer server.Close()

    resp, err := http.Get(server.URL + "/catalog.jsonl")
    if err != nil {
        t.Fatalf("Failed to GET /catalog.jsonl: %v", err)
    }
    defer resp.Body.Close()

    // Verify response
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status 200, got %d", resp.StatusCode)
    }

    // Read and validate JSONL format
    body, _ := io.ReadAll(resp.Body)
    lines := strings.Split(string(body), "\n")

    for i, line := range lines {
        if line == "" {
            continue
        }
        var entry map[string]interface{}
        if err := json.Unmarshal([]byte(line), &entry); err != nil {
            t.Errorf("Line %d is not valid JSON: %v", i+1, err)
        }
    }
}
```

### When to Choose httptest

- ✅ Only testing backend/API functionality
- ✅ Want fastest possible test execution
- ✅ Don't need JavaScript execution
- ✅ Testing static file serving
- ❌ Need to verify frontend UI behavior

---

## Comparison Matrix

### Feature Comparison

| Feature | Playwright | Puppeteer | Cypress | chromedp | rod | httptest |
|---------|-----------|-----------|---------|----------|-----|----------|
| **Cross-browser** | ✅ Chrome, Firefox, Safari | ❌ Chrome only | ⚠️ Chrome, Firefox, Edge | ❌ Chrome only | ❌ Chrome only | N/A |
| **JavaScript execution** | ✅ Full | ✅ Full | ✅ Full | ✅ Full | ✅ Full | ❌ None |
| **Auto-waiting** | ✅ Built-in | ❌ Manual | ✅ Built-in | ⚠️ Manual | ✅ Built-in | N/A |
| **Screenshots/Videos** | ✅ Yes | ⚠️ Screenshots only | ✅ Yes | ⚠️ Screenshots only | ✅ Yes | ❌ No |
| **Network mocking** | ✅ Built-in | ⚠️ Manual | ✅ Built-in | ⚠️ Manual | ✅ Built-in | ✅ Built-in |
| **Debugging** | ✅ Traces, inspector | ⚠️ DevTools | ✅✅ Best-in-class | ⚠️ DevTools | ✅ DevTools | ✅ Standard Go |
| **Speed** | ⚠️ Medium | ✅ Fast | ❌ Slower | ✅ Fast | ✅✅ Fastest | ✅✅ Fastest |
| **Installation size** | ❌ Large (~300MB/browser) | ⚠️ Medium (~170MB) | ❌ Large | ⚠️ Medium | ⚠️ Medium | ✅ None |
| **Language** | Node.js/TS | Node.js/TS | Node.js/TS | Go | Go | Go |

### Ecosystem & Support

| Aspect | Playwright | Puppeteer | Cypress | chromedp | rod | httptest |
|--------|-----------|-----------|---------|----------|-----|----------|
| **Maintainer** | Microsoft | Google | Cypress.io | Community | Community | Go team |
| **GitHub Stars** | 64k+ | 87k+ | 46k+ | 11k+ | 5k+ | N/A (stdlib) |
| **First Release** | 2020 | 2017 | 2015 | 2017 | 2020 | 2009 |
| **Update frequency** | High | High | High | Medium | High | Stable |
| **Community size** | Large | Very large | Very large | Medium | Growing | Very large |
| **Documentation** | Excellent | Excellent | Excellent | Good | Good | Excellent |
| **CI Examples** | Abundant | Abundant | Abundant | Some | Some | Abundant |

### CI/CD Integration

| Tool | GitHub Actions | Docker Support | Headless Support | Setup Complexity |
|------|---------------|----------------|------------------|------------------|
| **Playwright** | ✅ Official action | ✅ Official images | ✅ Default | ⭐⭐ Medium |
| **Puppeteer** | ✅ Easy setup | ✅ Many images | ✅ Default | ⭐ Low |
| **Cypress** | ✅ Official action | ✅ Official images | ✅ Default | ⭐⭐⭐ Higher |
| **chromedp** | ✅ Works well | ✅ headless-shell | ✅ Default | ⭐⭐ Medium |
| **rod** | ✅ Works well | ✅ Can use chromedp | ✅ Default | ⭐⭐ Medium |
| **httptest** | ✅ Built-in | ✅ Any Go image | ✅ N/A | ⭐ Trivial |

---

## Recommendations

### Recommended: Playwright (Primary) + httptest (Backend)

**Primary Choice: Playwright**

Playwright is recommended as the primary e2e testing solution because:

1. ✅ **Already installed in VM** - No setup overhead
2. ✅ **Cross-browser coverage** - Future-proof testing
3. ✅ **Best stability** - Proven to outperform alternatives
4. ✅ **Comprehensive features** - Screenshots, traces, videos for debugging
5. ✅ **Excellent CI support** - Official GitHub Action
6. ✅ **Active development** - Microsoft backing ensures longevity
7. ✅ **Matches frontend stack** - JavaScript/TypeScript like the web app

**Supplementary: Go httptest**

For backend integration testing:

1. ✅ **Zero dependencies** - Standard library
2. ✅ **Fast execution** - No browser overhead
3. ✅ **Perfect for API testing** - Test HTTP handlers directly
4. ✅ **Already familiar** - Team knows Go testing

### Implementation Priority

**Phase 1: Foundation (Week 1)**
1. Set up Playwright with basic configuration
2. Create 3-5 critical path tests (search, filter, modal)
3. Add to CI pipeline
4. Document test writing guidelines

**Phase 2: Coverage (Week 2-3)**
1. Expand to 20-30 tests covering all features
2. Add visual regression tests (screenshots)
3. Test keyboard navigation
4. Test theme switching and persistence

**Phase 3: Backend Integration (Week 3-4)**
1. Add httptest tests for static file serving
2. Test JSONL data loading
3. Test error handling
4. Integration tests for full data pipeline

### Alternative Recommendation: rod (If Staying in Go)

If the team strongly prefers staying in the Go ecosystem:

**Choose rod over chromedp because:**
- ✅ Better performance (decode-on-demand)
- ✅ More stable architecture (remote object ID vs DOM node ID)
- ✅ Higher-level API (easier to use)
- ✅ Better suited for modern web apps

However, this sacrifices:
- ❌ Cross-browser testing
- ❌ Richer debugging tools
- ❌ Larger ecosystem and community

---

## Implementation Considerations

### Local Development Setup

**Prerequisites:**
```bash
# For Playwright
npm install -D @playwright/test
npx playwright install chromium  # Or: firefox, webkit

# For chromedp/rod
go get github.com/chromedp/chromedp
# or
go get github.com/go-rod/rod
```

**Running Tests Locally:**

```bash
# Playwright
npm run test:e2e                 # Run all tests
npm run test:e2e:headed          # Run with visible browser
npm run test:e2e:debug           # Debug mode with inspector

# Go (chromedp/rod)
go test ./pkg/e2e/... -v         # Run all e2e tests
go test ./pkg/e2e/... -v -run TestTemplateSearch  # Run specific test
```

### VM Sandbox Considerations

**Playwright in VM:**
- ✅ Playwright CLI is already available (`/opt/node22/bin/playwright`)
- ✅ Browser installation works: `playwright install chromium`
- ⚠️ System dependencies (`--with-deps`) may fail - acceptable for headless
- ✅ Headless mode works fine without full system dependencies

**Browser Installation:**
```bash
# Install only Chromium (lightest option)
playwright install chromium

# Check installation
ls ~/.cache/ms-playwright/
# Expected: chromium-1194/, chromium_headless_shell-1194/, ffmpeg-1011/

# Test it works
node -e "const { chromium } = require('playwright'); (async () => { const b = await chromium.launch(); await b.close(); })();"
```

### CI Pipeline Updates

**Option A: Add Playwright to Existing CI (Recommended)**

```yaml
# Add to .github/workflows/ci.yml

  e2e-tests:
    name: E2E Tests
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '22'

      - name: Install dependencies
        run: npm ci

      - name: Install Playwright browsers
        run: npx playwright install --with-deps chromium

      - name: Start web server
        run: |
          cd web
          python3 -m http.server 8000 &
          sleep 2

      - name: Run E2E tests
        run: npx playwright test

      - uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: playwright-report
          path: playwright-report/
```

**Option B: Separate E2E Workflow**

Create `.github/workflows/e2e.yml` to keep e2e tests separate from unit tests. This allows:
- Different trigger conditions
- Longer timeout limits
- Separate artifact storage
- Easier debugging

### Project Structure

**Recommended structure:**

```
lima-catalog/
├── tests/
│   ├── e2e/
│   │   ├── playwright.config.js      # Playwright configuration
│   │   ├── fixtures/                 # Test data
│   │   ├── pages/                    # Page Object Models
│   │   └── specs/
│   │       ├── search.spec.js
│   │       ├── filters.spec.js
│   │       ├── modal.spec.js
│   │       ├── keyboard.spec.js
│   │       └── theme.spec.js
│   └── integration/
│       └── httptest/                 # Go integration tests
│           ├── server_test.go
│           └── data_test.go
├── pkg/
│   └── e2e/                         # Go e2e tests (if using chromedp/rod)
│       ├── catalog_test.go
│       └── helpers.go
└── package.json                      # Add Playwright dependency
```

### Test Data Management

**Strategy 1: Use Production Data Branch**
```javascript
// tests/e2e/specs/search.spec.js
test.beforeEach(async ({ page }) => {
  // Fetch from actual data branch
  await page.goto('http://localhost:8000');
  // Real data = real tests
});
```

**Strategy 2: Mock Data for Faster Tests**
```javascript
// tests/e2e/fixtures/templates.json
{
  "templates": [/* minimal test data */]
}

// tests/e2e/specs/search.spec.js
test.beforeEach(async ({ page }) => {
  await page.route('**/data/catalog.jsonl', route => {
    route.fulfill({
      status: 200,
      body: fs.readFileSync('tests/e2e/fixtures/catalog.jsonl')
    });
  });
  await page.goto('http://localhost:8000');
});
```

**Recommendation:** Start with production data (Strategy 1) for authenticity, add mocks (Strategy 2) later for speed if needed.

### Performance Optimization

**For Playwright:**
1. Only install browsers you need: `playwright install chromium` (not all 3)
2. Use `--workers=2` to parallelize tests
3. Reuse browser contexts where possible
4. Enable trace only on CI failure to save time

**For Go tests:**
1. Use `t.Parallel()` for parallel test execution
2. Reuse browser instances across tests
3. Use `testing.Short()` for quick vs full test suites

### Common Pitfalls & Solutions

**Issue: Browser installation fails in VM**
- **Solution:** Use `playwright install chromium` without `--with-deps`, headless mode works fine

**Issue: Tests are flaky**
- **Solution:** Use Playwright's auto-waiting, avoid `sleep()`, use `waitForSelector()`

**Issue: Tests are slow**
- **Solution:** Run in parallel (`--workers`), only test critical paths, use httptest for backend

**Issue: CI runs out of disk space**
- **Solution:** Only install Chromium, not all browsers, clean up artifacts after tests

**Issue: Local server not ready before tests**
- **Solution:** Add health check or wait for specific URL to respond before running tests

---

## Resources

### Official Documentation

- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Puppeteer Documentation](https://pptr.dev/)
- [Cypress Documentation](https://docs.cypress.io/)
- [chromedp GitHub](https://github.com/chromedp/chromedp)
- [rod GitHub](https://github.com/go-rod/rod)
- [Go httptest Package](https://pkg.go.dev/net/http/httptest)

### Comparisons & Guides (2025)

- [Playwright vs Cypress: Key Differences for 2025](https://katalon.com/resources-center/blog/playwright-vs-cypress) - BrowserStack
- [Cypress vs Playwright in 2025](https://bugbug.io/blog/test-automation-tools/cypress-vs-playwright/) - BugBug
- [Playwright vs. Puppeteer in 2025: Which Should You Choose](https://www.zenrows.com/blog/playwright-vs-puppeteer) - ZenRows
- [Playwright vs Puppeteer: which is better in 2025?](https://blog.apify.com/playwright-vs-puppeteer/) - Apify
- [Why rod?](https://github.com/go-rod/go-rod.github.io/blob/main/why-rod.md) - go-rod
- [Puppeteer, Selenium, Playwright, Cypress – how to choose?](https://www.testim.io/blog/puppeteer-selenium-playwright-cypress-how-to-choose/) - Testim

### CI/CD Integration

- [Playwright CI Documentation](https://playwright.dev/docs/ci) - Official
- [Setup Playwright GitHub Action](https://github.com/marketplace/actions/setup-playwright) - GitHub Marketplace
- [Installing Playwright In GitHub Actions](https://stevefenton.co.uk/blog/2025/09/playwright-insteall-github-actions/) - Steve Fenton (Sept 2025)
- [Running UI Automation Tests with Go and Chrome on GitHub Actions](https://pradappandiyan.medium.com/running-ui-automation-tests-with-go-and-chrome-on-github-actions-1f56d7c63405) - Medium
- [GitHub Actions with Playwright: Automate Browser Testing Like a Pro](https://peerlist.io/jagss/articles/github-actions-with-playwright-automate-browser-testing-like) - Peerlist

### Performance & Architecture

- [Golang Headless Browser: Best Tools for Automation](https://latenode.com/blog/golang-headless-browser-best-tools-for-automation) - Latenode
- [Faster way to install Playwright Browsers on GitHub Actions](https://github.com/microsoft/playwright/issues/23388) - GitHub Issue

---

## Next Steps

1. **Review this document** with the team
2. **Decide on primary tool** (recommended: Playwright)
3. **Create proof-of-concept** with 3-5 basic tests
4. **Set up CI integration**
5. **Document test patterns** for contributors
6. **Expand test coverage** iteratively

For questions or suggestions, see [DEVELOPMENT.md](../../DEVELOPMENT.md) or open an issue.

---

**Last Updated:** 2025-11-22
**Author:** Analysis based on lima-catalog requirements and 2025 tooling landscape
