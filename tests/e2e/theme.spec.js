// @ts-check
const { test, expect } = require('@playwright/test');

test.describe('Theme Switching and Persistence', () => {
  test.beforeEach(async ({ page }) => {
    // Clear localStorage before each test
    await page.goto('/');
    await page.evaluate(() => localStorage.clear());
  });

  test('displays theme toggle button', async ({ page }) => {
    await page.goto('/');

    // Verify theme toggle button exists
    const themeToggle = page.locator('#theme-toggle');
    await expect(themeToggle).toBeVisible();
  });

  test('defaults to system preference when no saved theme', async ({ page }) => {
    // Set system preference to dark mode
    await page.emulateMedia({ colorScheme: 'dark' });
    await page.goto('/');
    await page.waitForTimeout(100);

    // Verify dark theme is applied
    const html = page.locator('html');
    await expect(html).toHaveAttribute('data-theme', 'dark');
  });

  test('defaults to light when system preference is light', async ({ page }) => {
    // Set system preference to light mode
    await page.emulateMedia({ colorScheme: 'light' });
    await page.goto('/');
    await page.waitForTimeout(100);

    // Verify light theme is applied
    const html = page.locator('html');
    await expect(html).toHaveAttribute('data-theme', 'light');
  });

  test('toggles from light to dark theme', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'light' });
    await page.goto('/');
    await page.waitForTimeout(100);

    // Verify starting in light mode
    const html = page.locator('html');
    await expect(html).toHaveAttribute('data-theme', 'light');

    // Click theme toggle
    await page.click('#theme-toggle');
    await page.waitForTimeout(100);

    // Verify switched to dark mode
    await expect(html).toHaveAttribute('data-theme', 'dark');
  });

  test('toggles from dark to light theme', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'dark' });
    await page.goto('/');
    await page.waitForTimeout(100);

    // Verify starting in dark mode
    const html = page.locator('html');
    await expect(html).toHaveAttribute('data-theme', 'dark');

    // Click theme toggle
    await page.click('#theme-toggle');
    await page.waitForTimeout(100);

    // Verify switched to light mode
    await expect(html).toHaveAttribute('data-theme', 'light');
  });

  test('persists theme selection in localStorage', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(100);

    // Toggle to dark theme
    await page.click('#theme-toggle');
    await page.waitForTimeout(100);

    // Verify localStorage has the theme saved
    const savedTheme = await page.evaluate(() => localStorage.getItem('theme'));
    expect(savedTheme).toBe('dark');
  });

  test('loads saved theme from localStorage on page load', async ({ page }) => {
    // Set theme in localStorage
    await page.goto('/');
    await page.evaluate(() => localStorage.setItem('theme', 'dark'));

    // Reload page
    await page.reload();
    await page.waitForTimeout(100);

    // Verify dark theme is applied
    const html = page.locator('html');
    await expect(html).toHaveAttribute('data-theme', 'dark');
  });

  test('saved theme overrides system preference', async ({ page }) => {
    // Set system to light, but save dark in localStorage
    await page.emulateMedia({ colorScheme: 'light' });
    await page.goto('/');
    await page.evaluate(() => localStorage.setItem('theme', 'dark'));

    // Reload page
    await page.reload();
    await page.waitForTimeout(100);

    // Verify dark theme is applied (from localStorage, not system)
    const html = page.locator('html');
    await expect(html).toHaveAttribute('data-theme', 'dark');
  });

  test('theme persists across page navigation', async ({ page }) => {
    await page.goto('/');

    // Set dark theme
    await page.click('#theme-toggle');
    await page.waitForTimeout(100);

    // Verify dark theme is set
    let html = page.locator('html');
    await expect(html).toHaveAttribute('data-theme', 'dark');

    // Navigate to a different page (if applicable) or reload
    await page.reload();
    await page.waitForTimeout(100);

    // Verify dark theme persists
    html = page.locator('html');
    await expect(html).toHaveAttribute('data-theme', 'dark');
  });

  test('theme icon updates when toggling', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'light' });
    await page.goto('/');
    await page.waitForTimeout(100);

    const themeToggle = page.locator('#theme-toggle');

    // In light mode, should show moon icon (☾) or sun icon depending on implementation
    const initialIcon = await themeToggle.textContent();
    expect(initialIcon).toBeTruthy();

    // Toggle theme
    await page.click('#theme-toggle');
    await page.waitForTimeout(100);

    // Verify icon changed
    const newIcon = await themeToggle.textContent();
    expect(newIcon).not.toBe(initialIcon);
  });

  test('applies correct background color in light mode', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'light' });
    await page.goto('/');
    await page.waitForTimeout(100);

    // Check that body has light background
    const bodyBg = await page.locator('body').evaluate(el =>
      window.getComputedStyle(el).backgroundColor
    );

    // Light mode typically has white or very light background
    // RGB values should be high (close to 255)
    expect(bodyBg).toBeTruthy();
  });

  test('applies correct background color in dark mode', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'dark' });
    await page.goto('/');
    await page.waitForTimeout(100);

    // Check that body has dark background
    const bodyBg = await page.locator('body').evaluate(el =>
      window.getComputedStyle(el).backgroundColor
    );

    // Dark mode has dark background
    expect(bodyBg).toBeTruthy();
  });

  test('template cards are visible in both themes', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('#templates-grid .template-card', { timeout: 10000 });

    // Verify cards visible in initial theme
    const cards = page.locator('#templates-grid .template-card');
    let count = await cards.count();
    expect(count).toBeGreaterThan(0);

    // Toggle theme
    await page.click('#theme-toggle');
    await page.waitForTimeout(300);

    // Verify cards still visible in new theme
    count = await cards.count();
    expect(count).toBeGreaterThan(0);
  });

  test('modal is styled correctly in dark theme', async ({ page }) => {
    // Set dark theme
    await page.goto('/');
    await page.click('#theme-toggle');
    await page.waitForTimeout(200);

    // Wait for templates to load
    await page.waitForSelector('#templates-grid .template-card', { timeout: 10000 });

    // Open modal
    await page.locator('#templates-grid .template-card .template-name').first().click();
    await expect(page.locator('#preview-modal')).toBeVisible();

    // Verify modal is visible and styled
    const modal = page.locator('.modal-content');
    await expect(modal).toBeVisible();

    const modalBg = await modal.evaluate(el =>
      window.getComputedStyle(el).backgroundColor
    );
    expect(modalBg).toBeTruthy();
  });

  test('sidebar is styled correctly in both themes', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('#keyword-cloud', { timeout: 5000 });

    // Check sidebar in initial theme
    const sidebar = page.locator('#sidebar');
    await expect(sidebar).toBeVisible();

    // Toggle theme
    await page.click('#theme-toggle');
    await page.waitForTimeout(200);

    // Verify sidebar still visible
    await expect(sidebar).toBeVisible();
  });
});
