// @ts-check
const fs = require('fs');
const path = require('path');

/**
 * Mock the catalog data fetching for tests and block external CDN requests
 * @param {import('@playwright/test').Page} page - Playwright page object
 */
export async function mockCatalogData(page) {
  // Read the local catalog data
  const catalogPath = path.join(__dirname, '../../web/catalog.jsonl');
  const catalogData = fs.readFileSync(catalogPath, 'utf-8');

  // Block all external CDN requests first to prevent 403 errors
  await page.route('**/*', (route) => {
    const url = route.request().url();
    if (url.includes('cdnjs.cloudflare.com') || url.includes('cdn.')) {
      // Return empty response for CDN resources
      route.fulfill({
        status: 200,
        contentType: 'application/javascript',
        body: '/* CDN resource mocked for testing */',
      });
    } else if (url.includes('/data/catalog.jsonl')) {
      // Return local catalog data
      route.fulfill({
        status: 200,
        contentType: 'text/plain',
        body: catalogData,
      });
    } else {
      route.continue();
    }
  });
}
