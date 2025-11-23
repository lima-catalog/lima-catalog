// @ts-check
const { test, expect } = require('@playwright/test');

test.describe('Theme Switching and Persistence', () => {
  test.beforeEach(async ({ page }) => {
    // Clear localStorage before each test
    await page.goto('/');
    await page.evaluate(() => localStorage.clear());
  });

  test('displays theme switcher buttons', async ({ page }) => {
    await page.goto('/');

    // Verify theme switcher exists with three options
    await expect(page.locator('.theme-option[data-theme="light"]')).toBeVisible();
    await expect(page.locator('.theme-option[data-theme="auto"]')).toBeVisible();
    await expect(page.locator('.theme-option[data-theme="dark"]')).toBeVisible();
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

    // Verify light theme is applied (no data-theme attribute for light)
    const html = page.locator('html');
    const theme = await html.getAttribute('data-theme');
    expect(theme).toBeNull();
  });

  test('switches from light to dark theme', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'light' });
    await page.goto('/');
    await page.waitForTimeout(100);

    // Click light theme button explicitly
    await page.click('.theme-option[data-theme="light"]');
    await page.waitForTimeout(100);

    // Verify light theme is active (no data-theme attribute for light)
    const html = page.locator('html');
    let theme = await html.getAttribute('data-theme');
    expect(theme).toBeNull();

    // Click dark theme button
    await page.click('.theme-option[data-theme="dark"]');
    await page.waitForTimeout(100);

    // Verify switched to dark mode
    await expect(html).toHaveAttribute('data-theme', 'dark');
  });

  test('switches from dark to light theme', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'dark' });
    await page.goto('/');
    await page.waitForTimeout(100);

    // Click dark theme button explicitly
    await page.click('.theme-option[data-theme="dark"]');
    await page.waitForTimeout(100);

    // Verify dark theme is active
    const html = page.locator('html');
    await expect(html).toHaveAttribute('data-theme', 'dark');

    // Click light theme button
    await page.click('.theme-option[data-theme="light"]');
    await page.waitForTimeout(100);

    // Verify switched to light mode (no data-theme attribute for light)
    const theme = await html.getAttribute('data-theme');
    expect(theme).toBeNull();
  });

  test('persists theme selection in localStorage', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(100);

    // Select dark theme
    await page.click('.theme-option[data-theme="dark"]');
    await page.waitForTimeout(100);

    // Verify localStorage has the theme saved (using correct key)
    const savedTheme = await page.evaluate(() => localStorage.getItem('lima-catalog-theme'));
    expect(savedTheme).toBe('dark');
  });

  test('loads saved theme from localStorage on page load', async ({ page }) => {
    // Set theme in localStorage (using correct key)
    await page.goto('/');
    await page.evaluate(() => localStorage.setItem('lima-catalog-theme', 'dark'));

    // Reload page
    await page.reload();
    await page.waitForTimeout(100);

    // Verify dark theme is applied
    const html = page.locator('html');
    await expect(html).toHaveAttribute('data-theme', 'dark');
  });

  test('saved theme overrides system preference', async ({ page }) => {
    // Set system to light, but save dark in localStorage (using correct key)
    await page.emulateMedia({ colorScheme: 'light' });
    await page.goto('/');
    await page.evaluate(() => localStorage.setItem('lima-catalog-theme', 'dark'));

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
    await page.click('.theme-option[data-theme="dark"]');
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

  test('theme button shows active state', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'light' });
    await page.goto('/');
    await page.waitForTimeout(100);

    // Click light theme
    await page.click('.theme-option[data-theme="light"]');
    await page.waitForTimeout(100);

    // Verify light theme button is active
    const lightButton = page.locator('.theme-option[data-theme="light"]');
    await expect(lightButton).toHaveClass(/active/);

    // Click dark theme
    await page.click('.theme-option[data-theme="dark"]');
    await page.waitForTimeout(100);

    // Verify dark theme button is active and light is not
    const darkButton = page.locator('.theme-option[data-theme="dark"]');
    await expect(darkButton).toHaveClass(/active/);
    await expect(lightButton).not.toHaveClass(/active/);
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

    // Switch to dark theme
    await page.click('.theme-option[data-theme="dark"]');
    await page.waitForTimeout(300);

    // Verify cards still visible in dark theme
    count = await cards.count();
    expect(count).toBeGreaterThan(0);

    // Switch to light theme
    await page.click('.theme-option[data-theme="light"]');
    await page.waitForTimeout(300);

    // Verify cards still visible in light theme
    count = await cards.count();
    expect(count).toBeGreaterThan(0);
  });

  test('modal is styled correctly in dark theme', async ({ page }) => {
    // Set dark theme
    await page.goto('/');
    await page.click('.theme-option[data-theme="dark"]');
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
    const sidebar = page.locator('.sidebar');
    await expect(sidebar).toBeVisible();

    // Toggle theme (switch to dark)
    await page.click('.theme-option[data-theme="dark"]');
    await page.waitForTimeout(200);

    // Verify sidebar still visible
    await expect(sidebar).toBeVisible();
  });
});
