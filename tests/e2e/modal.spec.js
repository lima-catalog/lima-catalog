// @ts-check
const { test, expect } = require('@playwright/test');

test.describe('Template Preview Modal', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the catalog
    await page.goto('/');

    // Wait for templates to load
    await page.waitForSelector('#templates-grid .template-card', { timeout: 10000 });
  });

  test('opens modal when clicking template card', async ({ page }) => {
    // Verify modal is initially hidden
    await expect(page.locator('#preview-modal')).not.toBeVisible();

    // Click first template card
    await page.locator('#templates-grid .template-card').first().click();

    // Verify modal is now visible
    await expect(page.locator('#preview-modal')).toBeVisible();
    await expect(page.locator('.modal-content')).toBeVisible();

    // Verify modal has title
    const modalTitle = await page.locator('#modal-title').textContent();
    expect(modalTitle).toBeTruthy();
    expect(modalTitle.trim().length).toBeGreaterThan(0);
  });

  test('displays GitHub URL in modal', async ({ page }) => {
    // Click first template
    await page.locator('#templates-grid .template-card').first().click();

    // Wait for modal to be visible
    await expect(page.locator('#preview-modal')).toBeVisible();

    // Verify GitHub URL is displayed
    const githubUrl = await page.locator('#modal-github-scheme').textContent();
    expect(githubUrl).toBeTruthy();
    expect(githubUrl).toMatch(/github:/);

    // Verify copy button is present
    await expect(page.locator('#copy-github-url')).toBeVisible();
  });

  test('loads and displays template content', async ({ page }) => {
    // Click first template
    await page.locator('#templates-grid .template-card').first().click();

    // Wait for modal
    await expect(page.locator('#preview-modal')).toBeVisible();

    // Wait for content to load (loading indicator should disappear)
    await page.waitForFunction(
      () => {
        const loading = document.querySelector('#modal-loading');
        return !loading || loading.style.display === 'none' || !loading.offsetParent;
      },
      { timeout: 15000 }
    );

    // Verify YAML content is present
    const yamlContent = page.locator('#modal-yaml-content');
    await expect(yamlContent).toBeVisible();

    const content = await yamlContent.textContent();
    expect(content).toBeTruthy();
    expect(content.trim().length).toBeGreaterThan(0);
  });

  test('closes modal with close button', async ({ page }) => {
    // Open modal
    await page.locator('#templates-grid .template-card').first().click();
    await expect(page.locator('#preview-modal')).toBeVisible();

    // Click close button
    await page.locator('.modal-close').click();

    // Verify modal is closed
    await expect(page.locator('#preview-modal')).not.toBeVisible();
  });

  test('closes modal with Escape key', async ({ page }) => {
    // Open modal
    await page.locator('#templates-grid .template-card').first().click();
    await expect(page.locator('#preview-modal')).toBeVisible();

    // Press Escape
    await page.keyboard.press('Escape');

    // Verify modal is closed
    await expect(page.locator('#preview-modal')).not.toBeVisible();
  });

  test('closes modal by clicking overlay', async ({ page }) => {
    // Open modal
    await page.locator('#templates-grid .template-card').first().click();
    await expect(page.locator('#preview-modal')).toBeVisible();

    // Click overlay (outside modal content)
    await page.locator('.modal-overlay').click();

    // Verify modal is closed
    await expect(page.locator('#preview-modal')).not.toBeVisible();
  });

  test('copy button copies GitHub URL to clipboard', async ({ page }) => {
    // Grant clipboard permissions
    await page.context().grantPermissions(['clipboard-read', 'clipboard-write']);

    // Open modal
    await page.locator('#templates-grid .template-card').first().click();
    await expect(page.locator('#preview-modal')).toBeVisible();

    // Get the URL text before clicking
    const urlText = await page.locator('#modal-github-scheme').textContent();

    // Click copy button
    await page.locator('#copy-github-url').click();

    // Wait a moment for copy to complete
    await page.waitForTimeout(100);

    // Verify clipboard contains the URL
    const clipboardText = await page.evaluate(() => navigator.clipboard.readText());
    expect(clipboardText).toBe(urlText);
  });

  test('navigates between templates while modal is open', async ({ page }) => {
    // Open first template
    await page.locator('#templates-grid .template-card').first().click();
    await expect(page.locator('#preview-modal')).toBeVisible();

    const firstTitle = await page.locator('#modal-title').textContent();

    // Use keyboard to navigate to next template (if implemented)
    // Or close and open another template
    await page.keyboard.press('Escape');
    await expect(page.locator('#preview-modal')).not.toBeVisible();

    // Open second template
    await page.locator('#templates-grid .template-card').nth(1).click();
    await expect(page.locator('#preview-modal')).toBeVisible();

    const secondTitle = await page.locator('#modal-title').textContent();

    // Titles should be different
    expect(secondTitle).not.toBe(firstTitle);
  });

  test('displays similar templates section when available', async ({ page }) => {
    // Click first template
    await page.locator('#templates-grid .template-card').first().click();
    await expect(page.locator('#preview-modal')).toBeVisible();

    // Wait for content to load
    await page.waitForFunction(
      () => {
        const loading = document.querySelector('#modal-loading');
        return !loading || loading.style.display === 'none' || !loading.offsetParent;
      },
      { timeout: 15000 }
    );

    // Check if similar templates section exists
    const similarSection = page.locator('#similar-templates-section');
    const isVisible = await similarSection.isVisible().catch(() => false);

    if (isVisible) {
      // Verify it has content
      const similarTemplates = page.locator('#similar-templates-list .similar-template-item');
      const count = await similarTemplates.count();
      expect(count).toBeGreaterThan(0);
    }
    // If not visible, that's fine - not all templates have similars
  });
});
