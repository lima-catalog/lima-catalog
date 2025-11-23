// @ts-check
const { test, expect } = require('@playwright/test');

/**
 * E2E Tests: Visual Regression
 *
 * High-level overview of what's being tested:
 * - Screenshot comparisons for the entire homepage (light and dark modes)
 * - Individual component rendering (cards, sidebar, keyword cloud, category list)
 * - Interactive states (hover, focus, selected) for visual consistency
 * - Modal appearance in both themes with syntax-highlighted content
 * - Responsive design across viewports (mobile 375px, tablet 768px, desktop 1920px)
 * - UI elements: search bar, filters, sort dropdown, theme toggle, badges
 * - Empty states and edge cases (no results, filtered views)
 *
 * NOTE: These tests are currently SKIPPED (.skip) until the UI is more stable.
 * To enable: remove .skip and run: npx playwright test visual.spec.js --update-snapshots
 */

// Visual regression tests are skipped until the UI is more stable
// These tests require baseline screenshots to be generated and reviewed
// To enable: remove .skip and run: npx playwright test visual.spec.js --update-snapshots
test.describe.skip('Visual Regression Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the catalog
    await page.goto('/');

    // Wait for templates to load
    await page.waitForSelector('#templates-grid .template-card', { timeout: 10000 });

    // Wait a bit for all content to settle
    await page.waitForTimeout(500);
  });

  test('homepage renders correctly in light mode', async ({ page }) => {
    // Set light theme
    await page.emulateMedia({ colorScheme: 'light' });
    await page.evaluate(() => {
      document.documentElement.setAttribute('data-theme', 'light');
    });
    await page.waitForTimeout(200);

    // Take screenshot
    await expect(page).toHaveScreenshot('homepage-light.png', {
      fullPage: true,
      animations: 'disabled',
    });
  });

  test('homepage renders correctly in dark mode', async ({ page }) => {
    // Set dark theme
    await page.emulateMedia({ colorScheme: 'dark' });
    await page.evaluate(() => {
      document.documentElement.setAttribute('data-theme', 'dark');
    });
    await page.waitForTimeout(200);

    // Take screenshot
    await expect(page).toHaveScreenshot('homepage-dark.png', {
      fullPage: true,
      animations: 'disabled',
    });
  });

  test('template card renders consistently', async ({ page }) => {
    // Get first template card
    const firstCard = page.locator('#templates-grid .template-card').first();

    // Take screenshot of just the card
    await expect(firstCard).toHaveScreenshot('template-card.png', {
      animations: 'disabled',
    });
  });

  test('template card hover state', async ({ page }) => {
    // Hover over first card
    const firstCard = page.locator('#templates-grid .template-card').first();
    await firstCard.hover();
    await page.waitForTimeout(100);

    // Take screenshot
    await expect(firstCard).toHaveScreenshot('template-card-hover.png', {
      animations: 'disabled',
    });
  });

  test('template card focused state', async ({ page }) => {
    // Focus first card
    const firstCard = page.locator('#templates-grid .template-card').first();
    await firstCard.focus();
    await page.waitForTimeout(100);

    // Take screenshot
    await expect(firstCard).toHaveScreenshot('template-card-focused.png', {
      animations: 'disabled',
    });
  });

  test('sidebar renders consistently', async ({ page }) => {
    // Wait for sidebar content
    await page.waitForSelector('#sidebar', { timeout: 5000 });
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    const sidebar = page.locator('#sidebar');

    // Take screenshot
    await expect(sidebar).toHaveScreenshot('sidebar.png', {
      animations: 'disabled',
    });
  });

  test('keyword cloud renders consistently', async ({ page }) => {
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    const keywordCloud = page.locator('#keyword-cloud');

    // Take screenshot
    await expect(keywordCloud).toHaveScreenshot('keyword-cloud.png', {
      animations: 'disabled',
    });
  });

  test('selected keywords display correctly', async ({ page }) => {
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    // Select a keyword
    await page.locator('#keyword-cloud .keyword-tag').first().click();
    await page.waitForTimeout(300);

    // Wait for selected keywords to appear
    await page.waitForSelector('#selected-keywords .selected-keyword', { timeout: 2000 });

    const selectedKeywords = page.locator('#selected-keywords');

    // Take screenshot
    await expect(selectedKeywords).toHaveScreenshot('selected-keywords.png', {
      animations: 'disabled',
    });
  });

  test('category list renders consistently', async ({ page }) => {
    await page.waitForSelector('#category-list .category-item', { timeout: 5000 });

    const categoryList = page.locator('#category-list');

    // Take screenshot
    await expect(categoryList).toHaveScreenshot('category-list.png', {
      animations: 'disabled',
    });
  });

  test('selected category highlights correctly', async ({ page }) => {
    await page.waitForSelector('#category-list .category-item', { timeout: 5000 });

    // Click first category
    await page.locator('#category-list .category-item').first().click();
    await page.waitForTimeout(300);

    const categoryList = page.locator('#category-list');

    // Take screenshot
    await expect(categoryList).toHaveScreenshot('category-selected.png', {
      animations: 'disabled',
    });
  });

  test('search bar with input renders correctly', async ({ page }) => {
    // Type in search
    await page.fill('#search', 'alpine');
    await page.waitForTimeout(300);

    const searchContainer = page.locator('.search-container');

    // Take screenshot
    await expect(searchContainer).toHaveScreenshot('search-with-input.png', {
      animations: 'disabled',
    });
  });

  test('filtered results display correctly', async ({ page }) => {
    // Apply filter
    await page.fill('#search', 'ubuntu');
    await page.waitForTimeout(500);

    const templatesGrid = page.locator('#templates-grid');

    // Take screenshot
    await expect(templatesGrid).toHaveScreenshot('filtered-results.png', {
      animations: 'disabled',
    });
  });

  test('modal renders correctly in light mode', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'light' });

    // Open modal
    await page.locator('#templates-grid .template-card .template-name').first().click();
    await expect(page.locator('#preview-modal')).toBeVisible();

    // Wait for content to load
    await page.waitForTimeout(1000);

    const modal = page.locator('#preview-modal');

    // Take screenshot
    await expect(modal).toHaveScreenshot('modal-light.png', {
      animations: 'disabled',
    });
  });

  test('modal renders correctly in dark mode', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'dark' });
    await page.evaluate(() => {
      document.documentElement.setAttribute('data-theme', 'dark');
    });

    // Open modal
    await page.locator('#templates-grid .template-card .template-name').first().click();
    await expect(page.locator('#preview-modal')).toBeVisible();

    // Wait for content to load
    await page.waitForTimeout(1000);

    const modal = page.locator('#preview-modal');

    // Take screenshot
    await expect(modal).toHaveScreenshot('modal-dark.png', {
      animations: 'disabled',
    });
  });

  test('modal code content renders with syntax highlighting', async ({ page }) => {
    // Mock highlight.js
    await page.addInitScript(() => {
      window.hljs = {
        highlightElement: () => {},
        highlight: (code) => ({ value: code }),
        configure: () => {},
      };
    });

    // Open modal
    await page.locator('#templates-grid .template-card .template-name').first().click();
    await expect(page.locator('#preview-modal')).toBeVisible();

    // Wait for YAML content to load
    await page.waitForFunction(
      () => {
        const loading = document.querySelector('#modal-loading');
        return !loading || loading.style.display === 'none' || !loading.offsetParent;
      },
      { timeout: 15000 }
    );

    await page.waitForTimeout(500);

    const codeContent = page.locator('#modal-code-content');

    // Take screenshot
    await expect(codeContent).toHaveScreenshot('modal-code-content.png', {
      animations: 'disabled',
    });
  });

  test('header renders consistently', async ({ page }) => {
    const header = page.locator('header');

    // Take screenshot
    await expect(header).toHaveScreenshot('header.png', {
      animations: 'disabled',
    });
  });

  test('template count displays correctly', async ({ page }) => {
    await page.waitForTimeout(300);

    const stats = page.locator('.stats-container');

    // Take screenshot
    await expect(stats).toHaveScreenshot('template-stats.png', {
      animations: 'disabled',
    });
  });

  test('empty search results display correctly', async ({ page }) => {
    // Search for something that won't match
    await page.fill('#search', 'xyznonexistenttemplate123');
    await page.waitForTimeout(500);

    const templatesGrid = page.locator('#templates-grid');

    // Take screenshot
    await expect(templatesGrid).toHaveScreenshot('empty-results.png', {
      animations: 'disabled',
    });
  });

  test('mobile viewport renders correctly', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    await page.waitForTimeout(300);

    // Take screenshot
    await expect(page).toHaveScreenshot('mobile-view.png', {
      fullPage: true,
      animations: 'disabled',
    });
  });

  test('tablet viewport renders correctly', async ({ page }) => {
    // Set tablet viewport
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(300);

    // Take screenshot
    await expect(page).toHaveScreenshot('tablet-view.png', {
      fullPage: true,
      animations: 'disabled',
    });
  });

  test('wide desktop viewport renders correctly', async ({ page }) => {
    // Set wide viewport
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(300);

    // Take screenshot
    await expect(page).toHaveScreenshot('wide-desktop-view.png', {
      fullPage: false, // Just above the fold
      animations: 'disabled',
    });
  });

  test('sort dropdown renders correctly', async ({ page }) => {
    const sortDropdown = page.locator('#sort');

    // Take screenshot
    await expect(sortDropdown).toHaveScreenshot('sort-dropdown.png', {
      animations: 'disabled',
    });
  });

  test('filter checkboxes render correctly', async ({ page }) => {
    const filterToggles = page.locator('.filter-toggles');

    // Take screenshot
    await expect(filterToggles).toHaveScreenshot('filter-checkboxes.png', {
      animations: 'disabled',
    });
  });

  test('official template badge renders correctly', async ({ page }) => {
    // Find a template card with official badge
    const officialCard = page.locator('#templates-grid .template-card').first();

    // Take screenshot
    await expect(officialCard).toHaveScreenshot('official-template.png', {
      animations: 'disabled',
    });
  });

  test('theme toggle button renders correctly', async ({ page }) => {
    const themeToggle = page.locator('#theme-toggle');

    // Take screenshot
    await expect(themeToggle).toHaveScreenshot('theme-toggle.png', {
      animations: 'disabled',
    });
  });
});
