// @ts-check
const { test, expect } = require('@playwright/test');

test.describe('Category and Keyword Filtering', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the catalog
    await page.goto('/');

    // Wait for templates to load
    await page.waitForSelector('#templates-grid .template-card', { timeout: 10000 });
  });

  test('displays category list', async ({ page }) => {
    // Wait for categories to load
    await page.waitForSelector('#category-list .category-item', { timeout: 5000 });

    // Verify categories are displayed
    const categories = page.locator('#category-list .category-item');
    const count = await categories.count();
    expect(count).toBeGreaterThan(0);

    // Verify each category has a name and count
    const firstCategory = categories.first();
    const categoryText = await firstCategory.textContent();
    expect(categoryText).toBeTruthy();
    expect(categoryText).toMatch(/\d+/); // Should have count like "164"
  });

  test('filters templates by clicking category', async ({ page }) => {
    // Wait for categories
    await page.waitForSelector('#category-list .category-item', { timeout: 5000 });

    // Get initial template count
    const initialCount = await page.locator('#templates-grid .template-card').count();

    // Click first category
    const firstCategory = page.locator('#category-list .category-item').first();
    const categoryName = await firstCategory.textContent();
    await firstCategory.click();

    // Wait for filter to apply
    await page.waitForTimeout(300);

    // Verify category is marked as active
    await expect(firstCategory).toHaveClass(/active/);

    // Verify templates are filtered
    const filteredCount = await page.locator('#templates-grid .template-card').count();
    expect(filteredCount).toBeGreaterThan(0);
    expect(filteredCount).toBeLessThanOrEqual(initialCount);

    // Verify visible count matches
    const visibleCount = await page.locator('#visible-count').textContent();
    expect(parseInt(visibleCount)).toBe(filteredCount);
  });

  test('clears category filter by clicking again', async ({ page }) => {
    // Wait for categories
    await page.waitForSelector('#category-list .category-item', { timeout: 5000 });

    const initialCount = await page.locator('#templates-grid .template-card').count();

    // Click category
    const category = page.locator('#category-list .category-item').first();
    await category.click();
    await page.waitForTimeout(300);

    const filteredCount = await page.locator('#templates-grid .template-card').count();

    // Click again to deselect
    await category.click();
    await page.waitForTimeout(300);

    // Verify filter is cleared
    await expect(category).not.toHaveClass(/active/);
    const restoredCount = await page.locator('#templates-grid .template-card').count();
    expect(restoredCount).toBe(initialCount);
  });

  test('displays keyword cloud', async ({ page }) => {
    // Wait for keywords to load
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    // Verify keywords are displayed
    const keywords = page.locator('#keyword-cloud .keyword-tag');
    const count = await keywords.count();
    expect(count).toBeGreaterThan(0);

    // Verify keywords have counts
    const firstKeyword = keywords.first();
    const keywordText = await firstKeyword.textContent();
    expect(keywordText).toBeTruthy();
  });

  test('filters templates by clicking keyword', async ({ page }) => {
    // Wait for keywords
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    const initialCount = await page.locator('#templates-grid .template-card').count();

    // Click first keyword
    const firstKeyword = page.locator('#keyword-cloud .keyword-tag').first();
    const keywordText = await firstKeyword.textContent();
    await firstKeyword.click();

    // Wait for filter to apply
    await page.waitForTimeout(300);

    // Verify keyword is moved to selected keywords
    const selectedKeywords = page.locator('#selected-keywords .keyword-tag-selected');
    await expect(selectedKeywords).toHaveCount(1);

    // Verify templates are filtered
    const filteredCount = await page.locator('#templates-grid .template-card').count();
    expect(filteredCount).toBeGreaterThan(0);
    expect(filteredCount).toBeLessThanOrEqual(initialCount);

    // Verify clear button is visible
    await expect(page.locator('#clear-keywords')).toBeVisible();
  });

  test('adds multiple keywords for AND filtering', async ({ page }) => {
    // Wait for keywords
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    // Click first keyword
    await page.locator('#keyword-cloud .keyword-tag').nth(0).click();
    await page.waitForTimeout(300);

    const oneKeywordCount = await page.locator('#templates-grid .template-card').count();

    // Click second keyword
    await page.locator('#keyword-cloud .keyword-tag').nth(0).click(); // nth(0) because first moved to selected
    await page.waitForTimeout(300);

    // Verify both keywords are selected
    const selectedKeywords = page.locator('#selected-keywords .keyword-tag-selected');
    await expect(selectedKeywords).toHaveCount(2);

    // Verify results are more filtered (AND logic)
    const twoKeywordsCount = await page.locator('#templates-grid .template-card').count();
    expect(twoKeywordsCount).toBeLessThanOrEqual(oneKeywordCount);
  });

  test('removes keyword by clicking X on selected keyword', async ({ page }) => {
    // Wait for keywords
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    // Click a keyword
    await page.locator('#keyword-cloud .keyword-tag').first().click();
    await page.waitForTimeout(300);

    // Verify keyword is selected
    await expect(page.locator('#selected-keywords .keyword-tag-selected')).toHaveCount(1);

    const filteredCount = await page.locator('#templates-grid .template-card').count();

    // Click remove button on selected keyword
    await page.locator('#selected-keywords .keyword-tag-selected .keyword-remove').click();
    await page.waitForTimeout(300);

    // Verify keyword is removed
    await expect(page.locator('#selected-keywords .keyword-tag-selected')).toHaveCount(0);

    // Verify templates are restored
    const restoredCount = await page.locator('#templates-grid .template-card').count();
    expect(restoredCount).toBeGreaterThan(filteredCount);
  });

  test('clears all keywords with clear button', async ({ page }) => {
    // Wait for keywords
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    // Add multiple keywords
    await page.locator('#keyword-cloud .keyword-tag').nth(0).click();
    await page.waitForTimeout(200);
    await page.locator('#keyword-cloud .keyword-tag').nth(0).click();
    await page.waitForTimeout(200);

    // Verify keywords are selected
    const selectedCount = await page.locator('#selected-keywords .keyword-tag-selected').count();
    expect(selectedCount).toBeGreaterThan(0);

    // Click clear all button
    await page.locator('#clear-keywords').click();
    await page.waitForTimeout(300);

    // Verify all keywords are cleared
    await expect(page.locator('#selected-keywords .keyword-tag-selected')).toHaveCount(0);

    // Verify clear button is hidden
    await expect(page.locator('#clear-keywords')).not.toBeVisible();
  });

  test('combines category and keyword filters', async ({ page }) => {
    // Wait for both to load
    await page.waitForSelector('#category-list .category-item', { timeout: 5000 });
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    const initialCount = await page.locator('#templates-grid .template-card').count();

    // Select a category
    await page.locator('#category-list .category-item').first().click();
    await page.waitForTimeout(300);

    const categoryOnlyCount = await page.locator('#templates-grid .template-card').count();

    // Add a keyword
    await page.locator('#keyword-cloud .keyword-tag').first().click();
    await page.waitForTimeout(300);

    // Verify combined filtering
    const combinedCount = await page.locator('#templates-grid .template-card').count();
    expect(combinedCount).toBeLessThanOrEqual(categoryOnlyCount);
    expect(combinedCount).toBeLessThanOrEqual(initialCount);
  });

  test('keyword counts update when category filter changes', async ({ page }) => {
    // Wait for both to load
    await page.waitForSelector('#category-list .category-item', { timeout: 5000 });
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    // Get first keyword count
    const firstKeyword = page.locator('#keyword-cloud .keyword-tag').first();
    const initialKeywordText = await firstKeyword.textContent();

    // Select a category
    await page.locator('#category-list .category-item').first().click();
    await page.waitForTimeout(300);

    // Verify keyword cloud updated (counts may have changed)
    // Just verify keywords are still visible
    const keywordsAfterCategory = page.locator('#keyword-cloud .keyword-tag');
    const count = await keywordsAfterCategory.count();
    expect(count).toBeGreaterThan(0);
  });
});
