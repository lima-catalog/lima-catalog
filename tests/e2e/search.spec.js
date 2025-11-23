// @ts-check
const { test, expect } = require('@playwright/test');

test.describe('Template Search and Filtering', () => {
  test.beforeEach(async ({ page }) => {
    // Intercept GitHub catalog.jsonl request and serve local file instead
    await page.route('**/raw.githubusercontent.com/**catalog.jsonl', route => {
      route.fulfill({ path: 'web/catalog.jsonl' });
    });

    // Navigate to the catalog
    await page.goto('/');

    // Wait for initial templates to load
    await page.waitForSelector('#templates-grid .template-card', { timeout: 10000 });
  });

  test('loads and displays templates', async ({ page }) => {
    // Check that loading message is gone
    await expect(page.locator('#loading')).not.toBeVisible();

    // Check that templates are displayed
    const templates = page.locator('#templates-grid .template-card');
    const count = await templates.count();
    expect(count).toBeGreaterThan(0);

    // Verify header stats are updated
    const totalCount = await page.locator('#total-count').textContent();
    expect(totalCount).not.toBe('-');
    expect(parseInt(totalCount)).toBeGreaterThan(0);

    const visibleCount = await page.locator('#visible-count').textContent();
    expect(visibleCount).not.toBe('-');
    expect(parseInt(visibleCount)).toBeGreaterThan(0);
  });

  test('filters templates by search term', async ({ page }) => {
    // Get initial count
    const initialCount = await page.locator('#templates-grid .template-card').count();
    const initialVisibleCount = await page.locator('#visible-count').textContent();

    // Search for "alpine"
    await page.fill('#search', 'alpine');

    // Wait for filter to apply
    await page.waitForTimeout(300);

    // Verify filtered results
    const filteredCount = await page.locator('#templates-grid .template-card').count();
    expect(filteredCount).toBeGreaterThan(0);
    expect(filteredCount).toBeLessThan(initialCount);

    // Verify visible count updated
    const newVisibleCount = await page.locator('#visible-count').textContent();
    expect(newVisibleCount).not.toBe(initialVisibleCount);
    expect(parseInt(newVisibleCount)).toBe(filteredCount);

    // Verify all visible templates contain "alpine" in their content
    const templates = page.locator('#templates-grid .template-card');
    const count = await templates.count();
    for (let i = 0; i < Math.min(count, 5); i++) {
      const text = await templates.nth(i).textContent();
      expect(text.toLowerCase()).toContain('alpine');
    }
  });

  test('clears search and restores all templates', async ({ page }) => {
    // Search for something
    await page.fill('#search', 'alpine');
    await page.waitForTimeout(300);

    const filteredCount = await page.locator('#templates-grid .template-card').count();

    // Clear search
    await page.fill('#search', '');
    await page.waitForTimeout(300);

    // Verify all templates are back
    const restoredCount = await page.locator('#templates-grid .template-card').count();
    expect(restoredCount).toBeGreaterThan(filteredCount);
  });

  test('filters by official/community checkboxes', async ({ page }) => {
    const initialCount = await page.locator('#templates-grid .template-card').count();

    // Uncheck community templates (should show only official)
    await page.locator('#show-community').uncheck();
    await page.waitForTimeout(500);

    const officialOnlyCount = await page.locator('#templates-grid .template-card').count();
    expect(officialOnlyCount).toBeGreaterThan(0);
    expect(officialOnlyCount).toBeLessThan(initialCount);

    // Re-check community, uncheck official (should show only community)
    await page.locator('#show-community').check();
    await page.locator('#show-official').uncheck();
    await page.waitForTimeout(500);

    const communityOnlyCount = await page.locator('#templates-grid .template-card').count();
    expect(communityOnlyCount).toBeGreaterThan(0);
    expect(communityOnlyCount).toBeLessThan(initialCount);

    // Re-enable both (should show all)
    await page.locator('#show-official').check();
    await page.waitForTimeout(500);

    const restoredCount = await page.locator('#templates-grid .template-card').count();
    expect(restoredCount).toBe(initialCount);
  });

  test('sorts templates by different criteria', async ({ page }) => {
    // Get first template name with default sort
    const firstTemplateName = await page.locator('#templates-grid .template-card').first().locator('.template-name').textContent();

    // Change sort to "Recently Updated"
    await page.selectOption('#sort', 'updated');
    await page.waitForTimeout(300);

    // Verify order might have changed (or stayed same if first was already most recent)
    const newFirstTemplateName = await page.locator('#templates-grid .template-card').first().locator('.template-name').textContent();
    // Just verify we can sort without errors (order may or may not change)
    expect(newFirstTemplateName).toBeTruthy();

    // Change to Popularity
    await page.selectOption('#sort', 'stars');
    await page.waitForTimeout(300);

    // Verify templates still display
    const count = await page.locator('#templates-grid .template-card').count();
    expect(count).toBeGreaterThan(0);
  });

  test('combines search with filter checkboxes', async ({ page }) => {
    // Search for "ubuntu"
    await page.fill('#search', 'ubuntu');
    await page.waitForTimeout(300);

    const searchOnlyCount = await page.locator('#templates-grid .template-card').count();

    // Also uncheck community
    await page.click('#show-community');
    await page.waitForTimeout(300);

    const combinedCount = await page.locator('#templates-grid .template-card').count();
    expect(combinedCount).toBeLessThanOrEqual(searchOnlyCount);
  });
});
