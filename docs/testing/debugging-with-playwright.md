# Debugging UI Issues with Playwright

**Last Updated:** 2025-11-23
**Status:** Operational Guide

## Quick Links

- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Playwright API Reference](https://playwright.dev/docs/api/class-playwright)
- [Project playwright.config.js](/home/user/lima-catalog/playwright.config.js)
- [E2E Test Directory](/home/user/lima-catalog/tests/e2e/)
- [E2E Test Fixtures](/home/user/lima-catalog/tests/e2e/fixtures/README.md) - Local YAML fixtures for modal testing
- [E2E Testing Options Analysis](./e2e-integration-testing-options.md)

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Environment Setup](#environment-setup)
3. [Running Tests](#running-tests)
4. [Modern Debugging Methods](#modern-debugging-methods)
   - [UI Mode (Recommended)](#ui-mode-recommended)
   - [Trace Viewer](#trace-viewer)
   - [Playwright Agents (v1.56+)](#playwright-agents-v156)
   - [Inspector and Debug Mode](#inspector-and-debug-mode)
5. [Interactive Debugging with Scripts](#interactive-debugging-with-scripts)
6. [Debugging Test Flakiness](#debugging-test-flakiness)
7. [Playwright MCP Limitations](#playwright-mcp-limitations)
8. [Common Issues and Solutions](#common-issues-and-solutions)
9. [Writing Debug Scripts](#writing-debug-scripts)

---

## Executive Summary

Playwright is fully functional in the lima-catalog VM environment and can be used for:
- ✅ Running automated E2E tests
- ✅ **UI Mode** for visual debugging with time-travel (Playwright 1.32+)
- ✅ **Trace Viewer** for post-mortem test debugging
- ✅ **Playwright Agents** (v1.56+) for AI-powered test healing
- ✅ Interactive UI debugging via custom scripts
- ✅ Taking screenshots and videos of UI states
- ✅ Programmatic browser automation for investigation
- ❌ **NOT available:** Playwright MCP tools (see [limitations](#playwright-mcp-limitations))

**Current Version:** Playwright 1.56.1 with AI Agents support

**Key Finding:** While Playwright MCP doesn't work in this environment due to session initialization constraints, Playwright itself works perfectly via npm scripts and custom Node.js scripts, providing full programmatic browser control plus modern visual debugging tools like UI Mode and Trace Viewer.

---

## Environment Setup

### Prerequisites

```bash
# Install npm dependencies (includes @playwright/test)
npm install

# Install Playwright browsers (Chromium)
npx playwright install chromium

# Create required temp directory for Chromium
mkdir -p /root/tmp
```

### Web Server

The Playwright configuration automatically starts a web server, but you can also run it manually:

```bash
# Start web server on port 8000
cd web && python3 -m http.server 8000

# Or run in background
cd web && python3 -m http.server 8000 &
```

The application will be available at: `http://localhost:8000`

### Configuration

The project uses special configuration for sandbox environments in [`playwright.config.js`](/home/user/lima-catalog/playwright.config.js):

**Key settings:**
- `--no-sandbox` - Required for VM environments
- `--disable-setuid-sandbox` - Disables sandbox isolation
- `--disable-gpu` - Required for headless mode
- `--disable-dev-shm-usage` - Avoids /dev/shm issues
- `TMPDIR` - Set via environment variable in CI workflow

**Important:** Browser flags differ between environments:
- **Local development VM:** Uses `--single-process` to avoid IPC permission issues
- **GitHub CI:** Skips `--single-process` to prevent resource exhaustion crashes
- Detection: Uses `process.env.GITHUB_ACTIONS` to determine environment

**Why this matters:**
The `--single-process` flag is necessary in local VMs due to IPC permission restrictions, but causes browser crashes in GitHub CI after several tests due to resource exhaustion. The configuration automatically selects the appropriate flags based on the environment.

---

## Running Tests

### Run All Tests

```bash
npm run test:e2e
```

### Run Specific Test File

```bash
npx playwright test tests/e2e/search.spec.js
```

### Run in Headed Mode (see browser)

```bash
npm run test:e2e:headed

# Or specific test
npx playwright test tests/e2e/modal.spec.js --headed
```

### Run with Debugging Tools

```bash
# UI Mode - Visual debugging with time-travel (recommended)
npx playwright test --ui

# Inspector - Step-by-step debugging
npx playwright test --debug

# With trace recording for later analysis
npx playwright test --trace on
```

See the [Modern Debugging Methods](#modern-debugging-methods) section for detailed guidance on each debugging approach.

### Current Test Status (as of 2025-11-23)

**Phase 2 Complete: 125 total tests**

- ✅ **search.spec.js** - 17 tests (search, filtering, sorting, URL parameters)
- ✅ **categories.spec.js** - 24 tests (category/keyword filtering, URL state)
- ✅ **modal.spec.js** - 22 tests (modal open/close, content loading, navigation)
- ✅ **keyboard.spec.js** - 22 tests (arrow keys, shortcuts, focus management)
- ✅ **theme.spec.js** - 15 tests (light/dark/auto theme switching)
- ⏭️ **visual.spec.js** - 25 tests (SKIPPED - visual regression baselines not yet stable)

**Critical Setup Required:**
```bash
# Create temp directory for Chromium shared memory
mkdir -p /root/tmp
```

Without this directory, all tests fail with browser crash errors.

**Known Issues:**
- Visual regression tests are skipped until UI is more stable
- Some tests may be flaky due to timing - increase timeouts if needed

### E2E Test Fixtures

The E2E tests use a local fixture system for template YAML files to avoid network dependencies. See [`tests/e2e/fixtures/README.md`](../../tests/e2e/fixtures/README.md) for complete documentation.

**How it works:**
1. **Generation**: Scripts analyze the catalog and select 20 representative templates
2. **Storage**: Sample YAML files are stored in `tests/e2e/fixtures/templates/`
3. **Manifest**: A `manifest.json` file maps template URLs to local fixture filenames
4. **Serving**: During tests, `tests/e2e/fixtures.js` intercepts GitHub requests and serves local fixtures

**Key features:**
- ✅ Offline testing - no network access required
- ✅ Stable test data - fixtures don't change unexpectedly
- ✅ Fast tests - no network latency
- ✅ Predictable - same fixtures on every run

**Regenerating fixtures:**
```bash
# Generate sample fixtures (recommended for local development)
node scripts/create-sample-fixtures.js web/catalog.jsonl

# Download real YAML files (requires network)
node scripts/create-test-fixtures.js web/catalog.jsonl
```

**Fixture selection criteria:**
- First 3 templates from catalog (for index-based tests)
- First 5 templates alphabetically by name (matches UI default sort)
- 2 official templates from lima-vm/lima
- 2 templates with similar templates (for diff testing)
- Diverse categories and high notability scores

This enables comprehensive modal content testing without external dependencies.

---

## Modern Debugging Methods

Playwright provides several powerful debugging tools. This section covers the recommended modern approaches for debugging tests in 2025.

### UI Mode (Recommended)

**UI Mode** is the most powerful debugging tool for Playwright, providing a visual interface with time-travel debugging capabilities.

#### What is UI Mode?

UI Mode lets you:
- 👁️ **Watch tests run** in real-time with live browser preview
- ⏮️ **Time-travel** through each action using the timeline
- 🔍 **Inspect DOM snapshots** at any point in test execution
- 🎯 **Pick locators** visually by clicking on elements
- 🔄 **Re-run tests** instantly with watch mode
- 📊 **See all test results** in an organized view

#### Running UI Mode

```bash
# Run all tests in UI Mode
npx playwright test --ui

# Run specific test file in UI Mode
npx playwright test tests/e2e/modal.spec.js --ui

# Note: UI Mode may not work in headless VM environments
# In this case, use headed mode or trace viewer instead
```

#### UI Mode Features

**Timeline View:**
- Visual timeline of all actions with color-coded navigation and interactions
- Hover over any point to see a snapshot of the page at that moment
- Click to jump to any action instantly

**Actions Tab:**
- See every action taken during the test
- View the exact locator used for each action
- See timing information for each step
- Hover to highlight the element in the DOM snapshot

**Pick Locator:**
- Click "Pick Locator" button in the toolbar
- Click on any element in the browser preview
- Get the recommended Playwright locator instantly
- Great for writing new tests or debugging selectors

**Watch Mode:**
- Tests automatically re-run when code changes
- Instant feedback during development
- See test results update in real-time

#### UI Mode Limitations in VM Environments

**Note:** UI Mode requires a graphical environment and may not work in headless VMs or SSH sessions. If UI Mode fails to start:

1. Use **Trace Viewer** for post-mortem debugging (see below)
2. Enable **trace recording** in playwright.config.js
3. View traces after test completion: `npx playwright show-trace trace.zip`

---

### Trace Viewer

**Trace Viewer** provides time-travel debugging for completed test runs, perfect for CI/CD debugging and post-mortem analysis.

#### What is Trace Viewer?

Trace Viewer is a GUI tool that lets you explore recorded Playwright traces after tests have run. It provides the same time-travel capabilities as UI Mode, but works with pre-recorded traces.

#### Recording Traces

**Configure in playwright.config.js:**

```javascript
module.exports = defineConfig({
  use: {
    // Record trace on first retry of failed tests
    trace: 'on-first-retry',

    // Or always record traces (larger file size)
    // trace: 'on',

    // Or record on CI only
    // trace: process.env.CI ? 'on-first-retry' : 'off',
  },
});
```

**Trace options:**
- `'on'` - Record trace for every test (large files)
- `'on-first-retry'` - Only record for failed tests (recommended)
- `'retain-on-failure'` - Keep traces only for failed tests
- `'off'` - Don't record traces

#### Viewing Traces

**After test run:**

```bash
# View trace from test results
npx playwright show-trace test-results/example-test/trace.zip

# Or open trace viewer and load files manually
npx playwright show-trace
```

**Trace Viewer features:**
- Full timeline of test execution
- DOM snapshots at each step
- Network activity and console logs
- Screenshots and videos
- Action metadata and timing
- Pop-out DOM snapshot for detailed inspection

#### Using Traces for Debugging

**Common debugging workflow:**

1. Run tests with trace enabled
2. Test fails in CI or locally
3. Download trace.zip from test-results
4. Open in Trace Viewer: `npx playwright show-trace trace.zip`
5. Scrub through timeline to find where it failed
6. Inspect DOM state, network requests, console
7. Identify root cause and fix

**Example:**

```bash
# Run tests with trace on failure
npx playwright test --trace on-first-retry

# If test fails, open the trace
npx playwright show-trace test-results/modal-should-open-Modal/trace.zip
```

---

### Playwright Agents (v1.56+)

**Playwright Agents** are AI-powered tools introduced in v1.56 for automated test generation, planning, and healing.

#### The Healer Agent

The **Healer Agent** automatically debugs and fixes failing tests:

**What it does:**
- 🔍 Analyzes test failures by examining console logs, network requests, and DOM snapshots
- 🧠 Identifies root causes using AI analysis of execution flow
- 🔧 Suggests multiple fixes following best practices
- 🔄 Iteratively attempts fixes until tests pass
- ⏭️ Skips tests if actual functionality is broken (not test code)

**How it works:**
1. Runs tests in debug mode
2. Captures all execution data (DOM, network, console)
3. Uses AI to analyze failure patterns
4. Generates fix suggestions
5. Applies fixes and re-runs tests
6. Repeats until success or determines app is broken

#### Setting Up Playwright Agents

**Initialize agents:**

```bash
# For VS Code integration
npx playwright init-agents --loop=vscode

# For Claude Code integration
npx playwright init-agents --loop=claude

# For other editors
npx playwright init-agents --loop=opencode
```

This creates agent configuration files in your project.

#### The Three Agents

**Planner Agent:**
- Explores your application
- Produces a Markdown test plan
- Identifies key user flows to test

**Generator Agent:**
- Transforms test plans into Playwright test files
- Writes idiomatic Playwright code
- Follows best practices automatically

**Healer Agent:**
- Executes test suite
- Identifies failures
- Automatically repairs broken tests
- Reduces debugging time from hours to minutes

#### Benefits of Playwright Agents

- **Self-healing tests:** Automatically adapt to UI changes
- **Faster debugging:** AI identifies root causes instantly
- **Best practices:** Agents follow Playwright conventions
- **Reduced maintenance:** Less time fixing flaky tests

#### Limitations

**Note:** Playwright Agents require:
- Playwright 1.56.0 or higher ✅ (you have 1.56.1)
- AI model API access (OpenAI, Anthropic, etc.)
- Configuration in your development environment

For this VM environment, agents may require additional setup for AI API access.

---

### Inspector and Debug Mode

**Playwright Inspector** provides step-by-step debugging with a visual interface.

#### Running Inspector

```bash
# Run tests with inspector
npx playwright test --debug

# Run specific test with inspector
npx playwright test tests/e2e/modal.spec.js --debug

# Run from specific line
npx playwright test tests/e2e/modal.spec.js:25 --debug
```

#### Inspector Features

**Debugger controls:**
- ▶️ Step over actions one at a time
- ⏸️ Pause execution
- ▶️ Resume test execution
- 🔄 Re-run test

**Code view:**
- See current test code
- Highlighted current action
- View source maps

**Locator picker:**
- Click "Record" to explore the page
- Click elements to generate locators
- Copy locators for use in tests

#### Debug Mode in Tests

**Add explicit breakpoints:**

```javascript
test('debug specific interaction', async ({ page }) => {
  await page.goto('/');

  // Pause here - inspector will open
  await page.pause();

  await page.click('.template-card');

  // Continue normally
  await expect(page.locator('#preview-modal')).toBeVisible();
});
```

#### When to Use Inspector

**Use Inspector when:**
- Writing new tests and exploring the app
- Debugging specific user interactions
- Need to step through tests manually
- Want to experiment with locators

**Use Trace Viewer instead when:**
- Debugging failed CI runs
- Analyzing test flakiness
- Need to review entire test execution
- Working in headless environment

---

### Recommended Debugging Workflow

Choose the right debugging approach based on your situation:

#### 🎯 Quick Decision Guide

| Situation | Recommended Tool | Why |
|-----------|------------------|-----|
| **Writing new tests** | UI Mode `--ui` | Live preview, pick locators, instant feedback |
| **Local test failure** | UI Mode `--ui` or Inspector `--debug` | Step through execution, see what's happening |
| **CI test failure** | Trace Viewer | Download trace.zip from artifacts, time-travel debug |
| **Flaky tests** | Trace Viewer with `trace: 'on-first-retry'` | Capture only failures, analyze patterns |
| **Headless VM/SSH** | Custom scripts | No GUI available, programmatic control |
| **Multiple test failures** | Healer Agent | Let AI identify patterns and suggest fixes |
| **Understanding app flow** | Planner Agent | AI explores app and documents flows |

#### 🔄 Debugging Process

**For local development:**

1. **First attempt:** Run with UI Mode
   ```bash
   npx playwright test --ui
   ```
   - Watch test execute
   - Use timeline to find failure point
   - Pick locators to fix selectors

2. **If UI Mode unavailable:** Use Inspector
   ```bash
   npx playwright test --debug
   ```
   - Step through test line by line
   - Pause at failure point
   - Experiment with fixes

3. **For headless environments:** Use custom scripts
   ```javascript
   // Quick debug script
   const { chromium } = require('@playwright/test');
   // ... see examples below
   ```

**For CI/CD debugging:**

1. **Enable traces in config:**
   ```javascript
   use: { trace: 'on-first-retry' }
   ```

2. **Download test artifacts** from CI run

3. **Open trace locally:**
   ```bash
   npx playwright show-trace trace.zip
   ```

4. **Analyze:**
   - Timeline shows exact failure point
   - DOM snapshot shows page state
   - Console logs show JavaScript errors
   - Network tab shows failed requests

**For recurring issues:**

1. **Use Healer Agent** to automatically fix
2. **Review AI suggestions** for patterns
3. **Apply best practice fixes**
4. **Let agent validate fixes**

#### 🛠️ Debugging Toolkit

**Essential commands:**

```bash
# Interactive debugging
npx playwright test --ui                    # Best: UI Mode with time-travel
npx playwright test --debug                 # Good: Step-by-step inspector
npx playwright test --headed                # See browser but no controls

# Recording and playback
npx playwright test --trace on              # Record all traces
npx playwright show-trace trace.zip         # View recorded trace

# Codegen - generate tests by recording
npx playwright codegen http://localhost:8000

# AI Agents (v1.56+)
npx playwright init-agents --loop=claude    # Set up agents
```

**Pro tips:**

- Use `page.pause()` to add breakpoints in test code
- Add `trace: 'on'` temporarily for specific debugging sessions
- Use Codegen to experiment with locators interactively
- Check console logs: `page.on('console', msg => console.log(msg.text()))`

---

## Interactive Debugging with Scripts

For environments where UI Mode and Inspector don't work (headless VMs, SSH sessions), use custom Node.js scripts for interactive debugging.

### Basic Debug Script Template

Create a file like `debug-ui.js`:

```javascript
// debug-ui.js
const { chromium } = require('@playwright/test');

(async () => {
  // Launch browser with sandbox configuration
  const browser = await chromium.launch({
    headless: true,
    args: [
      '--no-sandbox',
      '--disable-setuid-sandbox',
      '--single-process',
      '--disable-gpu',
      '--disable-dev-shm-usage'
    ],
    env: {
      ...process.env,
      TMPDIR: process.env.HOME + '/tmp'
    }
  });

  const context = await browser.newContext();
  const page = await context.newPage();

  // Navigate to the app
  await page.goto('http://localhost:8000');
  await page.waitForLoadState('networkidle');

  // --- Your debugging code here ---

  // Example: Check if templates are loaded
  const templateCount = await page.locator('.template-card').count();
  console.log('Templates found:', templateCount);

  // Example: Check category list
  const categories = await page.locator('.category-item').allTextContents();
  console.log('Categories:', categories);

  // Example: Take screenshot
  await page.screenshot({ path: 'debug-screenshot.png', fullPage: true });
  console.log('Screenshot saved to debug-screenshot.png');

  // Example: Get element state
  const searchBox = page.locator('#search');
  const isVisible = await searchBox.isVisible();
  console.log('Search box visible:', isVisible);

  // Example: Interact with UI
  await searchBox.fill('ubuntu');
  await page.waitForTimeout(500); // Wait for filtering
  const filteredCount = await page.locator('.template-card').count();
  console.log('Filtered templates:', filteredCount);

  // Close browser
  await browser.close();
})();
```

### Run Debug Script

```bash
node debug-ui.js
```

### Common Debug Patterns

**Inspect element state:**
```javascript
const element = page.locator('.my-element');
console.log('Visible:', await element.isVisible());
console.log('Enabled:', await element.isEnabled());
console.log('Text:', await element.textContent());
console.log('HTML:', await element.innerHTML());
console.log('Attributes:', await element.getAttribute('class'));
```

**Check console messages:**
```javascript
page.on('console', msg => console.log('Browser:', msg.text()));
page.on('pageerror', error => console.log('Error:', error));
```

**Wait for conditions:**
```javascript
// Wait for element
await page.waitForSelector('.template-card');

// Wait for network idle
await page.waitForLoadState('networkidle');

// Wait for specific condition
await page.waitForFunction(() => {
  return document.querySelectorAll('.template-card').length > 0;
});
```

**Take screenshots at different stages:**
```javascript
await page.screenshot({ path: 'step1-initial.png' });
await searchBox.fill('ubuntu');
await page.screenshot({ path: 'step2-filtered.png' });
```

---

## Debugging Test Flakiness

### Recognizing Data Loading Timing Issues

**Critical Pattern:** If ALL tests fail on first run but ALL pass on retry, this is a strong indicator of asynchronous data loading timing issues, often related to caching.

**Symptoms:**
- Test report shows many "flaky" tests (e.g., "23 flaky")
- Individual test runs show failures with timeouts or missing elements
- Retries consistently pass
- Local tests may pass while CI fails (or vice versa)

**Root Cause:**
Tests run before async data (like `catalog.jsonl`) has finished loading. On retry, the data is cached by the browser, so it loads instantly and tests pass.

### Waiting for Data to Load

**Problem:** Simply waiting for DOM elements with `waitForSelector` may not be enough if the data hasn't loaded yet.

**Solution:** Pre-load critical data in global setup (see below), then wait for template cards to appear:

```javascript
test.beforeEach(async ({ page }) => {
  await page.goto('/');

  // Wait for initial templates to load and render
  await page.waitForSelector('#templates-grid .template-card', { timeout: 10000 });
});
```

This works because:
1. Global setup pre-loads `catalog.jsonl` before any tests run
2. The browser caches this data for subsequent test navigations
3. Template cards only appear after data is loaded and rendered

**Note:** The app uses ES modules, so don't try to check for `window.appActions` or `window.state` - these are never exposed globally.

### Global Setup for Data Pre-loading

To reduce flakiness, pre-load critical data files before any tests run:

```javascript
// tests/e2e/global-setup.js
async function globalSetup() {
  const baseURL = 'http://localhost:8000';

  // Wait for web server to be ready
  for (let i = 0; i < 10; i++) {
    try {
      const response = await fetch(baseURL);
      if (response.ok) break;
    } catch (error) {
      if (i === 9) throw new Error('Web server not ready');
      await new Promise(resolve => setTimeout(resolve, 1000));
    }
  }

  // Pre-load catalog.jsonl to warm up the cache/server
  const catalogResponse = await fetch(`${baseURL}/js/catalog.jsonl`);
  await catalogResponse.text(); // Actually read the response

  // Give server time to stabilize
  await new Promise(resolve => setTimeout(resolve, 2000));
}
```

**Configure in playwright.config.js:**
```javascript
module.exports = defineConfig({
  globalSetup: require.resolve('./tests/e2e/global-setup.js'),
  // ... other config
});
```

### Identifying Cache-Related Issues

**How to diagnose:**
1. Run tests multiple times and note the pattern
2. If first run consistently fails and retries pass → likely caching issue
3. Check browser DevTools Network tab for resource loading times
4. Add console logging to track data loading:
   ```javascript
   page.on('console', msg => console.log('Browser:', msg.text()));
   ```

**Common cache-related failures:**
- Timeouts waiting for elements that depend on data
- Empty arrays or undefined values in app state
- Filters returning zero results when data isn't loaded

### CI-Specific Flakiness

**Why CI differs from local:**
- Different caching behavior
- Resource constraints (CPU, memory)
- Network latency to external resources
- Parallel test execution may cause resource contention

**CI-specific configurations:**
```javascript
// playwright.config.js
module.exports = defineConfig({
  timeout: process.env.CI ? 60 * 1000 : 30 * 1000,
  fullyParallel: !process.env.CI,  // Disable parallel in CI
  workers: process.env.CI ? 1 : undefined,  // Single worker in CI
  retries: process.env.CI ? 2 : 0,

  use: {
    actionTimeout: process.env.CI ? 15 * 1000 : 10 * 1000,
    navigationTimeout: process.env.CI ? 30 * 1000 : 15 * 1000,
  },
});
```

### Debugging Checklist for Flaky Tests

When debugging flaky tests, check:

1. ✅ **Data Loading**: Is `waitForFunction` used to verify data is loaded?
2. ✅ **Global Setup**: Are critical resources pre-loaded before tests?
3. ✅ **Network Idle**: Is `waitForLoadState('networkidle')` used where appropriate?
4. ✅ **Timeouts**: Are timeouts sufficient for slow CI environments?
5. ✅ **Parallel Execution**: Could resource contention be causing issues?
6. ✅ **Browser Cache**: Could caching affect test behavior between runs?
7. ✅ **Console Errors**: Are there JavaScript errors preventing proper initialization?

### Best Practices

**DO:**
- ✅ Wait for application state, not just DOM elements
- ✅ Verify data is loaded before running assertions
- ✅ Pre-load critical resources in global setup
- ✅ Use longer timeouts in CI environments
- ✅ Run tests sequentially in CI if resource-constrained
- ✅ Add retry logic for network-dependent operations

**DON'T:**
- ❌ Rely solely on `waitForSelector` for async data
- ❌ Assume data is loaded when elements appear
- ❌ Use fixed `waitForTimeout` as the primary synchronization method
- ❌ Run highly parallel tests in resource-constrained environments
- ❌ Ignore "flaky" test patterns - they indicate real issues

---

## Playwright MCP Limitations

### Why Playwright MCP Doesn't Work

**Chicken-and-egg problem:**
1. MCP servers must be configured in `/root/.claude.json` before the session starts
2. Commands to configure MCP can only be executed after the session has started
3. MCP tools are loaded at session initialization, not dynamically mid-session
4. Therefore, MCP tools are never available in web VM sessions

**Configuration exists but is inactive:**
- The Playwright MCP server is configured in `/root/.claude.json`
- Configuration: `npx -y @playwright/mcp@latest` with stdio transport
- This configuration would work if loaded at session start
- For web VM environments, this is not currently possible

### Alternative Approach

Instead of MCP tools, use programmatic Playwright via Node.js scripts (see [Interactive Debugging](#interactive-debugging) section). This provides:
- Full Playwright API access
- Scriptable and repeatable debugging
- Custom investigation workflows
- Screenshot and video capture
- Complete browser automation

**Trade-off:** Slightly less convenient than MCP tools, but provides equivalent functionality.

---

## Common Issues and Solutions

### Issue: Tests are "flaky" - fail on first run, pass on retry

**Symptom:**
```
Notice: 23 flaky tests
```

All or most tests fail initially but pass when retried.

**Solution:**
This is almost always a data loading timing issue. See the [Debugging Test Flakiness](#debugging-test-flakiness) section for comprehensive guidance. Quick fixes:

1. Add global setup to pre-load data (see [`tests/e2e/global-setup.js`](/home/user/lima-catalog/tests/e2e/global-setup.js))

2. Wait for template cards to appear in each test:
   ```javascript
   test.beforeEach(async ({ page }) => {
     await page.goto('/');
     await page.waitForSelector('#templates-grid .template-card', { timeout: 10000 });
   });
   ```

3. Ensure templates only appear after data is loaded (check the rendering logic)

### Issue: Browser crashes with "Target closed"

**Symptom:**
```
Error: browserContext.newPage: Target page, context or browser has been closed
```

**Solution:**
Ensure `/root/tmp` directory exists:
```bash
mkdir -p /root/tmp
```

This directory is required for Chromium shared memory in sandbox environments.

### Issue: Tests timeout waiting for elements

**Symptom:**
```
Timeout 30000ms exceeded waiting for selector
```

**Solutions:**
1. Check if web server is running: `curl http://localhost:8000`
2. Increase timeout in `playwright.config.js`: `timeout: 60 * 1000`
3. Add explicit wait: `await page.waitForLoadState('networkidle')`
4. Check browser console: `page.on('console', msg => console.log(msg.text()))`

### Issue: Missing data in tests

**Symptom:**
```
[getDynamicKeywords] Returning empty - missing data
```

**Cause:** Tests expect `/js/data.js` to have catalog data, but it might be empty or not generated.

**Solution:**
Ensure data is built before running tests:
```bash
make build  # Generates data.js
npm run test:e2e
```

### Issue: SSL/TLS errors

**Symptom:**
```
ERROR:net/socket/ssl_client_socket_impl.cc:902] handshake failed
```

**Solution:**
These are warnings from GitHub API calls and can be ignored. Configuration includes `ignoreHTTPSErrors: true` for this reason.

### Issue: Web server logs flooding CI output

**Symptom:**
Thousands of HTTP request logs appear in test output:
```
127.0.0.1 - - [23/Nov/2025 05:33:34] "GET /js/sidebar.js HTTP/1.1" 200 -
127.0.0.1 - - [23/Nov/2025 05:33:34] "GET /js/app.js HTTP/1.1" 200 -
...
```

**Solution:**
Web server logs are automatically redirected to `test-results/webserver.log` (configured in `playwright.config.js`). This file is:
- Created automatically when tests run
- Always included in CI artifacts, regardless of test outcome (see `.github/workflows/ci.yml`)
- Available for debugging without cluttering console output

**To view web server logs:**
```bash
# After running tests locally
cat test-results/webserver.log

# In CI: Download the "playwright-report" artifact from the failed workflow
# The artifact includes:
# - test-output.txt: Complete console output from the test run
# - results.json: Structured test data for automated analysis
# - webserver.log: Python HTTP server logs
# - screenshots, traces, and videos for failed tests
```

---

## Writing Debug Scripts

### Example: Investigate Keyword Filtering Issue

```javascript
// debug-keywords.js
const { chromium } = require('@playwright/test');

(async () => {
  const browser = await chromium.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--single-process']
  });
  const page = await browser.newPage();
  await page.goto('http://localhost:8000');
  await page.waitForLoadState('networkidle');

  // Check keyword cloud
  const keywords = await page.locator('.keyword-tag').allTextContents();
  console.log('Available keywords:', keywords);

  // Click first keyword
  if (keywords.length > 0) {
    await page.locator('.keyword-tag').first().click();
    await page.waitForTimeout(500);

    // Check selected state
    const selectedKeywords = await page.locator('.selected-keyword').allTextContents();
    console.log('Selected keywords:', selectedKeywords);

    // Check filtered results
    const visibleCount = page.locator('#visible-count');
    const count = await visibleCount.textContent();
    console.log('Visible templates after filter:', count);

    // Take screenshot
    await page.screenshot({ path: 'keyword-filter-debug.png' });
  }

  await browser.close();
})();
```

### Example: Test Modal Behavior

```javascript
// debug-modal.js
const { chromium } = require('@playwright/test');

(async () => {
  const browser = await chromium.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--single-process']
  });
  const page = await browser.newPage();
  await page.goto('http://localhost:8000');
  await page.waitForLoadState('networkidle');

  // Click first template card
  await page.locator('.template-card').first().click();
  await page.waitForTimeout(500);

  // Check modal state
  const modal = page.locator('#preview-modal');
  const isVisible = await modal.isVisible();
  console.log('Modal visible:', isVisible);

  // Check modal content
  const title = await page.locator('#modal-title').textContent();
  const githubUrl = await page.locator('#modal-github-scheme').textContent();
  console.log('Modal title:', title);
  console.log('GitHub URL:', githubUrl);

  // Check if template YAML loaded
  const codeContent = await page.locator('#modal-code-content').textContent();
  console.log('YAML loaded:', codeContent.length > 0);
  console.log('YAML preview:', codeContent.substring(0, 100));

  // Take screenshot
  await page.screenshot({ path: 'modal-debug.png' });

  await browser.close();
})();
```

### Example: Performance Investigation

```javascript
// debug-performance.js
const { chromium } = require('@playwright/test');

(async () => {
  const browser = await chromium.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--single-process']
  });
  const page = await browser.newPage();

  // Track performance metrics
  const startTime = Date.now();
  await page.goto('http://localhost:8000');
  const loadTime = Date.now() - startTime;
  console.log('Page load time:', loadTime, 'ms');

  // Wait for templates to render
  await page.waitForSelector('.template-card');
  const renderTime = Date.now() - startTime;
  console.log('Templates rendered in:', renderTime, 'ms');

  // Check resource loading
  const performance = await page.evaluate(() => {
    const perfData = window.performance.getEntriesByType('resource');
    return perfData.map(r => ({
      name: r.name,
      duration: r.duration,
      size: r.transferSize
    }));
  });

  console.log('Resource loading times:');
  performance.forEach(r => {
    console.log(`  ${r.name}: ${r.duration.toFixed(2)}ms (${r.size} bytes)`);
  });

  await browser.close();
})();
```

---

## Additional Resources

### Related Documentation
- [E2E Testing Options Analysis](./e2e-integration-testing-options.md) - Comparison of testing frameworks
- [Frontend Test Coverage Plan](./test-coverage-plan-frontend.md) - What needs testing
- [UI/UX Guidelines](../guides/ui-ux-guidelines.md) - Design patterns to test

### External Resources

**Official Playwright Documentation:**
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
- [Debugging Tests](https://playwright.dev/docs/debug) - Official debugging guide
- [UI Mode](https://playwright.dev/docs/test-ui-mode) - Visual debugging interface
- [Trace Viewer](https://playwright.dev/docs/trace-viewer) - Time-travel debugging
- [Running and Debugging Tests](https://playwright.dev/docs/running-tests)
- [Playwright Codegen](https://playwright.dev/docs/codegen) - Generate test code by recording actions
- [Playwright API Reference](https://playwright.dev/docs/api/class-playwright)

**Playwright 1.56 Features (2025):**
- [Playwright Agents Announcement](https://medium.com/@umitozdemirf/playwright-1-59-playwright-agents-at-a-glance-76aef05cbcee) - Overview of AI-powered testing
- [Playwright Agents in Action](https://dev.to/playwright/playwright-agents-planner-generator-and-healer-in-action-5ajh) - Practical examples
- [Playwright v1.56 – When AI Agents Meets Testing](https://medium.com/@lakindudesilva007/playwright-v1-56-when-ai-agents-meets-testing-144ca6973390)

**Advanced Debugging Resources:**
- [UI Mode for Running and Debugging Tests](https://www.lambdatest.com/blog/playwright-ui-mode/)
- [Debugging & Tracing Features](https://testomat.io/blog/debugging-tracing-playwright-features-tips-techniques-to-running-and-debugging-tests/)
- [The Ultimate Guide to Playwright Trace Viewer](https://momentic.ai/blog/the-ultimate-guide-to-playwright-trace-viewer-master-time-travel-debugging)
- [Playwright Debug: A Complete Guide](https://autify.com/blog/playwright-debug)

### Test Files Reference
- [`tests/e2e/search.spec.js`](/home/user/lima-catalog/tests/e2e/search.spec.js) - Search, filtering, and sorting tests (17 tests)
- [`tests/e2e/categories.spec.js`](/home/user/lima-catalog/tests/e2e/categories.spec.js) - Category and keyword filtering tests (24 tests)
- [`tests/e2e/modal.spec.js`](/home/user/lima-catalog/tests/e2e/modal.spec.js) - Modal interaction and navigation tests (22 tests)
- [`tests/e2e/keyboard.spec.js`](/home/user/lima-catalog/tests/e2e/keyboard.spec.js) - Keyboard navigation and shortcuts (22 tests)
- [`tests/e2e/theme.spec.js`](/home/user/lima-catalog/tests/e2e/theme.spec.js) - Theme switching and persistence (15 tests)
- [`tests/e2e/visual.spec.js`](/home/user/lima-catalog/tests/e2e/visual.spec.js) - Visual regression tests (25 tests, currently skipped)
- [`tests/e2e/global-setup.js`](/home/user/lima-catalog/tests/e2e/global-setup.js) - Pre-loads data before tests run
- [`tests/e2e/helpers.js`](/home/user/lima-catalog/tests/e2e/helpers.js) - Shared test utilities

---

**Note:** This guide will be updated as the Playwright testing infrastructure evolves. Last verified working on 2025-11-23 in lima-catalog VM environment.
