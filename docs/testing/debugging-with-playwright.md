# Debugging UI Issues with Playwright

**Last Updated:** 2025-11-23
**Status:** Operational Guide

## Quick Links

- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Playwright API Reference](https://playwright.dev/docs/api/class-playwright)
- [Project playwright.config.js](/home/user/lima-catalog/playwright.config.js)
- [E2E Test Directory](/home/user/lima-catalog/tests/e2e/)
- [E2E Testing Options Analysis](./e2e-integration-testing-options.md)

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Environment Setup](#environment-setup)
3. [Running Tests](#running-tests)
4. [Interactive Debugging](#interactive-debugging)
5. [Playwright MCP Limitations](#playwright-mcp-limitations)
6. [Common Issues and Solutions](#common-issues-and-solutions)
7. [Writing Debug Scripts](#writing-debug-scripts)

---

## Executive Summary

Playwright is fully functional in the lima-catalog VM environment and can be used for:
- ✅ Running automated E2E tests
- ✅ Interactive UI debugging via custom scripts
- ✅ Taking screenshots and videos of UI states
- ✅ Programmatic browser automation for investigation
- ❌ **NOT available:** Playwright MCP tools (see [limitations](#playwright-mcp-limitations))

**Key Finding:** While Playwright MCP doesn't work in this environment due to session initialization constraints, Playwright itself works perfectly via npm scripts and custom Node.js scripts, providing full programmatic browser control.

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
- `--single-process` - Runs browser in single process to avoid IPC permission issues
- `TMPDIR: process.env.HOME + '/tmp'` - Sets custom temp directory

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

### Run with Debug Mode

```bash
npx playwright test --debug
```

### Current Test Status (as of 2025-11-23)

- ✅ 16 tests passing
- ❌ 8 tests failing (mostly due to missing dynamic data)
- ⏭️ 1 skipped

**Passing tests include:**
- Category filtering
- Keyword filtering
- Modal functionality (partial)
- Search filtering
- Checkbox filters

**Failing tests:**
- Tests that depend on `/js/data.js` having full catalog data
- Console shows: `[getDynamicKeywords] Returning empty - missing data`

---

## Interactive Debugging

Since Playwright MCP tools are not available, use custom Node.js scripts for interactive debugging.

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
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
- [Playwright Debugging Guide](https://playwright.dev/docs/debug)
- [Playwright Trace Viewer](https://playwright.dev/docs/trace-viewer)
- [Playwright Codegen](https://playwright.dev/docs/codegen) - Generate test code by recording actions

### Test Files Reference
- [`tests/e2e/search.spec.js`](/home/user/lima-catalog/tests/e2e/search.spec.js) - Search and filtering tests
- [`tests/e2e/categories.spec.js`](/home/user/lima-catalog/tests/e2e/categories.spec.js) - Category and keyword tests
- [`tests/e2e/modal.spec.js`](/home/user/lima-catalog/tests/e2e/modal.spec.js) - Modal interaction tests
- [`tests/e2e/helpers.js`](/home/user/lima-catalog/tests/e2e/helpers.js) - Shared test utilities

---

**Note:** This guide will be updated as the Playwright testing infrastructure evolves. Last verified working on 2025-11-23 in lima-catalog VM environment.
