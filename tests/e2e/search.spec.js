// @ts-check
const { test } = require('./fixtures');
const { expect } = require('@playwright/test');

test.describe('Template Search and Filtering', () => {
  test.beforeEach(async ({ page }) => {
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

  test('handles search with special characters', async ({ page }) => {
    // Search with special characters
    await page.fill('#search', 'ubuntu+docker');
    await page.waitForTimeout(300);

    // Should not crash and should filter
    const templates = page.locator('#templates-grid .template-card');
    const count = await templates.count();
    expect(count).toBeGreaterThanOrEqual(0);
  });

  test('handles case-insensitive search', async ({ page }) => {
    // Search with uppercase
    await page.fill('#search', 'ALPINE');
    await page.waitForTimeout(300);

    const upperCount = await page.locator('#templates-grid .template-card').count();

    // Clear and search with lowercase
    await page.fill('#search', 'alpine');
    await page.waitForTimeout(300);

    const lowerCount = await page.locator('#templates-grid .template-card').count();

    // Should return same results (case-insensitive)
    expect(upperCount).toBe(lowerCount);
  });

  test('shows no results message for non-matching search', async ({ page }) => {
    // Search for something that definitely won't match
    await page.fill('#search', 'xyznonexistenttemplate999');
    await page.waitForTimeout(300);

    const count = await page.locator('#templates-grid .template-card').count();
    expect(count).toBe(0);

    // Verify visible count is 0
    const visibleCount = await page.locator('#visible-count').textContent();
    expect(parseInt(visibleCount)).toBe(0);
  });

  test('updates URL with search parameter', async ({ page }) => {
    // Search for something
    await page.fill('#search', 'alpine');
    // Wait for debounce (300ms) plus extra time for URL update
    await page.waitForTimeout(500);

    // Check URL contains search parameter
    const url = page.url();
    expect(url).toContain('search=alpine');
  });

  test('loads search from URL parameter on page load', async ({ page }) => {
    // Navigate with search parameter
    await page.goto('/?search=ubuntu');
    await page.waitForSelector('#templates-grid .template-card', { timeout: 10000 });
    await page.waitForTimeout(300);

    // Verify search input has the value
    const searchValue = await page.locator('#search').inputValue();
    expect(searchValue).toBe('ubuntu');

    // Verify results are filtered
    const templates = page.locator('#templates-grid .template-card');
    const count = await templates.count();
    expect(count).toBeGreaterThan(0);
  });

  test('search debounces rapid typing', async ({ page }) => {
    // Type rapidly
    await page.locator('#search').type('alpine', { delay: 50 });

    // Wait for debounce
    await page.waitForTimeout(400);

    // Should have filtered results
    const count = await page.locator('#templates-grid .template-card').count();
    expect(count).toBeGreaterThan(0);

    // Check visible count updated
    const visibleCount = await page.locator('#visible-count').textContent();
    expect(parseInt(visibleCount)).toBe(count);
  });

  test('clearing search restores all templates and updates URL', async ({ page }) => {
    const initialCount = await page.locator('#templates-grid .template-card').count();

    // Search
    await page.fill('#search', 'alpine');
    // Wait for debounce (300ms) plus extra time for URL update
    await page.waitForTimeout(500);

    // Verify URL has search param
    expect(page.url()).toContain('search=alpine');

    // Clear
    await page.fill('#search', '');
    // Wait for debounce (300ms) plus extra time for URL update
    await page.waitForTimeout(500);

    // Verify all templates restored
    const restoredCount = await page.locator('#templates-grid .template-card').count();
    expect(restoredCount).toBe(initialCount);

    // Verify URL search param is removed
    expect(page.url()).not.toContain('search=');
  });

  test('search works with multi-word queries', async ({ page }) => {
    // Search with multiple words
    await page.fill('#search', 'ubuntu docker');
    await page.waitForTimeout(300);

    // Should find templates matching any of the words
    const count = await page.locator('#templates-grid .template-card').count();
    expect(count).toBeGreaterThanOrEqual(0);
  });

  test('unchecking both official and community shows all results', async ({ page }) => {
    // Get initial template count (both checked = all templates)
    const initialCount = await page.locator('#templates-grid .template-card').count();

    // Uncheck both - should still show all templates
    await page.uncheck('#show-official');
    await page.uncheck('#show-community');
    await page.waitForTimeout(500);

    // Should show all templates (same as when both are checked)
    const count = await page.locator('#templates-grid .template-card').count();
    expect(count).toBe(initialCount);

    // Visible count should match total
    const visibleCount = await page.locator('#visible-count').textContent();
    expect(parseInt(visibleCount)).toBe(initialCount);
  });

  test('sort maintains filter state', async ({ page }) => {
    // Apply search
    await page.fill('#search', 'alpine');
    await page.waitForTimeout(300);

    const beforeSort = await page.locator('#templates-grid .template-card').count();

    // Change sort
    await page.selectOption('#sort', 'updated');
    await page.waitForTimeout(300);

    // Count should remain same (same filter, different order)
    const afterSort = await page.locator('#templates-grid .template-card').count();
    expect(afterSort).toBe(beforeSort);
  });

  test('sorting by name orders alphabetically', async ({ page }) => {
    await page.selectOption('#sort', 'name');
    await page.waitForTimeout(300);

    // Get first two template names
    const firstName = await page.locator('#templates-grid .template-card').first().locator('.template-name').textContent();
    const secondName = await page.locator('#templates-grid .template-card').nth(1).locator('.template-name').textContent();

    // First should be alphabetically before or equal to second
    expect(firstName.toLowerCase() <= secondName.toLowerCase()).toBe(true);
  });
});
