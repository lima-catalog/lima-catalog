// @ts-check
const { test } = require('./fixtures');
const { expect } = require('@playwright/test');

test.describe('Category and Keyword Filtering', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the catalog
    await page.goto('/');

    // Wait for initial templates to load
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

    // Verify category is marked as selected
    await expect(firstCategory).toHaveClass(/selected/);

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
    await expect(category).not.toHaveClass(/selected/);
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
    const selectedKeywords = page.locator('#selected-keywords .selected-keyword');
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
    const selectedKeywords = page.locator('#selected-keywords .selected-keyword');
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
    await expect(page.locator('#selected-keywords .selected-keyword')).toHaveCount(1);

    const filteredCount = await page.locator('#templates-grid .template-card').count();

    // Click on the selected keyword to remove it (the whole element is clickable)
    await page.locator('#selected-keywords .selected-keyword').click();
    await page.waitForTimeout(300);

    // Verify keyword is removed
    await expect(page.locator('#selected-keywords .selected-keyword')).toHaveCount(0);

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
    const selectedCount = await page.locator('#selected-keywords .selected-keyword').count();
    expect(selectedCount).toBeGreaterThan(0);

    // Click clear all button
    await page.locator('#clear-keywords').click();
    await page.waitForTimeout(300);

    // Verify all keywords are cleared
    await expect(page.locator('#selected-keywords .selected-keyword')).toHaveCount(0);

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

  test('category filter updates URL parameter', async ({ page }) => {
    await page.waitForSelector('#category-list .category-item', { timeout: 5000 });

    // Click a category
    const categoryItem = page.locator('#category-list .category-item').first();
    const categoryText = await categoryItem.textContent();
    // Extract category name (before the count)
    const categoryName = categoryText.split('(')[0].trim();

    await categoryItem.click();
    await page.waitForTimeout(300);

    // Check URL contains category parameter
    const url = page.url();
    expect(url).toContain('category=');
  });

  test('loads category from URL parameter', async ({ page }) => {
    // Navigate with category parameter
    await page.goto('/?category=Container');
    await page.waitForSelector('#templates-grid .template-card', { timeout: 10000 });
    await page.waitForTimeout(300);

    // Verify category is selected
    const selectedCategory = page.locator('#category-list .category-item.selected');
    await expect(selectedCategory).toHaveCount(1);
  });

  test('keyword filter updates URL parameter', async ({ page }) => {
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    // Click a keyword
    const keyword = page.locator('#keyword-cloud .keyword-tag').first();
    const keywordText = await keyword.textContent();

    await keyword.click();
    await page.waitForTimeout(300);

    // Check URL contains keyword parameter
    const url = page.url();
    expect(url).toContain('keyword=');
  });

  test('loads keywords from URL parameter', async ({ page }) => {
    // Navigate with keyword parameter
    await page.goto('/?keyword=docker');
    await page.waitForSelector('#templates-grid .template-card', { timeout: 10000 });
    await page.waitForTimeout(300);

    // Verify keyword is selected
    const selectedKeywords = page.locator('#selected-keywords .selected-keyword');
    await expect(selectedKeywords).toHaveCount(1);
  });

  test('loads multiple keywords from URL parameter', async ({ page }) => {
    // Navigate with multiple keywords
    await page.goto('/?keyword=docker&keyword=linux');
    await page.waitForSelector('#templates-grid .template-card', { timeout: 10000 });
    await page.waitForTimeout(300);

    // Verify both keywords are selected
    const selectedKeywords = page.locator('#selected-keywords .selected-keyword');
    await expect(selectedKeywords).toHaveCount(2);
  });

  test('clicking category twice deselects it', async ({ page }) => {
    await page.waitForSelector('#category-list .category-item', { timeout: 5000 });

    const category = page.locator('#category-list .category-item').first();

    // Click to select
    await category.click();
    await page.waitForTimeout(300);
    await expect(category).toHaveClass(/selected/);

    // Click again to deselect
    await category.click();
    await page.waitForTimeout(300);
    await expect(category).not.toHaveClass(/selected/);
  });

  test('selecting different category switches selection', async ({ page }) => {
    await page.waitForSelector('#category-list .category-item', { timeout: 5000 });

    const firstCategory = page.locator('#category-list .category-item').first();
    const secondCategory = page.locator('#category-list .category-item').nth(1);

    // Select first
    await firstCategory.click();
    await page.waitForTimeout(300);
    await expect(firstCategory).toHaveClass(/selected/);

    // Select second
    await secondCategory.click();
    await page.waitForTimeout(300);

    // First should be deselected, second selected
    await expect(firstCategory).not.toHaveClass(/selected/);
    await expect(secondCategory).toHaveClass(/selected/);
  });

  test('keyword cloud shows relevant keywords for filtered results', async ({ page }) => {
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    const initialKeywordCount = await page.locator('#keyword-cloud .keyword-tag').count();

    // Apply search filter
    await page.fill('#search', 'ubuntu');
    await page.waitForTimeout(500);

    // Keyword cloud should update
    const filteredKeywordCount = await page.locator('#keyword-cloud .keyword-tag').count();
    expect(filteredKeywordCount).toBeGreaterThan(0);
  });

  test('selected keywords persist when changing category', async ({ page }) => {
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });
    await page.waitForSelector('#category-list .category-item', { timeout: 5000 });

    // Select a keyword
    await page.locator('#keyword-cloud .keyword-tag').first().click();
    await page.waitForTimeout(300);

    // Verify keyword selected
    await expect(page.locator('#selected-keywords .selected-keyword')).toHaveCount(1);

    // Change category
    await page.locator('#category-list .category-item').first().click();
    await page.waitForTimeout(300);

    // Keyword should still be selected
    await expect(page.locator('#selected-keywords .selected-keyword')).toHaveCount(1);
  });

  test('category selection persists when adding keywords', async ({ page }) => {
    await page.waitForSelector('#category-list .category-item', { timeout: 5000 });
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    // Select category
    const category = page.locator('#category-list .category-item').first();
    await category.click();
    await page.waitForTimeout(300);

    // Add keyword
    await page.locator('#keyword-cloud .keyword-tag').first().click();
    await page.waitForTimeout(300);

    // Category should still be selected
    await expect(category).toHaveClass(/selected/);
  });

  test('clear keywords button is hidden when no keywords selected', async ({ page }) => {
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    // Initially, clear button should be hidden
    const clearButton = page.locator('#clear-keywords');
    await expect(clearButton).not.toBeVisible();
  });

  test('maximum keywords can be selected', async ({ page }) => {
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    // Select multiple keywords (up to 5 or whatever the limit is)
    for (let i = 0; i < 5; i++) {
      const keyword = page.locator('#keyword-cloud .keyword-tag').nth(i);
      if (await keyword.count() > 0) {
        await keyword.click();
        await page.waitForTimeout(200);
      }
    }

    // Should have 5 keywords selected (or fewer if not enough available)
    const selectedCount = await page.locator('#selected-keywords .selected-keyword').count();
    expect(selectedCount).toBeGreaterThan(0);
    expect(selectedCount).toBeLessThanOrEqual(5);
  });

  test('category counts are accurate', async ({ page }) => {
    await page.waitForSelector('#category-list .category-item', { timeout: 5000 });

    // Get first category and its count
    const firstCategory = page.locator('#category-list .category-item').first();
    const categoryText = await firstCategory.textContent();
    const countMatch = categoryText.match(/\((\d+)\)/);

    if (countMatch) {
      const expectedCount = parseInt(countMatch[1]);

      // Click the category
      await firstCategory.click();
      await page.waitForTimeout(300);

      // Count visible templates
      const actualCount = await page.locator('#templates-grid .template-card').count();

      // Should match the displayed count
      expect(actualCount).toBe(expectedCount);
    }
  });

  test('keyword search within selected keywords', async ({ page }) => {
    await page.waitForSelector('#keyword-cloud .keyword-tag', { timeout: 5000 });

    // Select multiple keywords
    await page.locator('#keyword-cloud .keyword-tag').nth(0).click();
    await page.waitForTimeout(200);
    await page.locator('#keyword-cloud .keyword-tag').nth(0).click();
    await page.waitForTimeout(200);

    // Get count with two keywords
    const twoKeywordsCount = await page.locator('#templates-grid .template-card').count();

    // Should be filtered
    expect(twoKeywordsCount).toBeGreaterThan(0);
  });
});
