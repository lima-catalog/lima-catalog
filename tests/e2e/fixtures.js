// @ts-check
const base = require('@playwright/test');

/**
 * Extended test with automatic GitHub request routing
 * Intercepts all GitHub raw.githubusercontent.com requests and serves local files
 */
const test = base.test.extend({
  page: async ({ page }, use) => {
    // Intercept GitHub raw.githubusercontent.com requests
    await page.route('**/raw.githubusercontent.com/**', route => {
      const url = route.request().url();

      // Serve local catalog.jsonl
      if (url.includes('catalog.jsonl')) {
        route.fulfill({ path: 'web/catalog.jsonl' });
      } else {
        // Abort other GitHub requests (YAML files) to prevent hanging
        route.abort('blockedbyclient');
      }
    });

    await use(page);
  },
});

module.exports = { test };
