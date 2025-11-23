// @ts-check
const base = require('@playwright/test');
const fs = require('fs');
const path = require('path');

// Load test fixture manifest
const manifestPath = path.join(__dirname, 'fixtures', 'manifest.json');
let manifest = null;

try {
  if (fs.existsSync(manifestPath)) {
    manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'));
  }
} catch (error) {
  console.warn('Warning: Could not load test fixtures manifest:', error.message);
}

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
        return;
      }

      // Try to serve template YAML from fixtures
      if (manifest && manifest.templates) {
        // Find matching template by raw_url
        const template = Object.values(manifest.templates).find(t => t.raw_url === url);

        if (template && template.filename) {
          const fixturePath = path.join(__dirname, 'fixtures', 'templates', template.filename);

          if (fs.existsSync(fixturePath)) {
            route.fulfill({
              status: 200,
              contentType: 'text/yaml',
              path: fixturePath
            });
            return;
          }
        }
      }

      // Abort other GitHub requests (YAML files not in fixtures) to prevent hanging
      route.abort('blockedbyclient');
    });

    await use(page);
  },
});

module.exports = { test };
