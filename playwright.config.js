// @ts-check
const { defineConfig, devices } = require('@playwright/test');

/**
 * Playwright configuration for lima-catalog e2e tests
 * @see https://playwright.dev/docs/test-configuration
 */
module.exports = defineConfig({
  testDir: './tests/e2e',

  // Maximum time one test can run for
  timeout: process.env.CI ? 60 * 1000 : 30 * 1000, // Longer timeout in CI for stability

  // Test execution settings
  fullyParallel: !process.env.CI, // Disable parallel execution in CI due to resource constraints
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,

  // Reporter to use
  reporter: process.env.CI ? 'github' : 'list',

  // Shared settings for all the projects below
  use: {
    // Base URL for tests
    baseURL: 'http://localhost:8000',

    // Collect trace on first retry of failed test
    trace: 'on-first-retry',

    // Screenshot on failure
    screenshot: 'only-on-failure',

    // Video on failure
    video: 'retain-on-failure',

    // Run in headless mode
    headless: true,

    // Timeout for each action (click, fill, etc.)
    actionTimeout: process.env.CI ? 15 * 1000 : 10 * 1000,

    // Timeout for navigation
    navigationTimeout: process.env.CI ? 30 * 1000 : 15 * 1000,
  },

  // Configure projects for major browsers
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        // Ignore HTTPS errors for fetching data from GitHub
        ignoreHTTPSErrors: true,
        // Use new headless mode (doesn't require X11)
        launchOptions: {
          headless: true,
          timeout: 60000, // Browser launch timeout
          args: [
            '--no-sandbox',
            '--disable-setuid-sandbox',
            '--disable-web-security',
            '--disable-features=IsolateOrigins,site-per-process',
            '--disable-gpu',
            '--disable-dev-shm-usage',
            '--disable-software-rasterizer',
            '--single-process',  // Run everything in one process to avoid IPC permission issues
          ],
          // TMPDIR is set via environment variable in CI workflow
        },
      },
    },

    // Uncomment to test on Firefox
    // {
    //   name: 'firefox',
    //   use: { ...devices['Desktop Firefox'] },
    // },

    // Uncomment to test on Safari
    // {
    //   name: 'webkit',
    //   use: { ...devices['Desktop Safari'] },
    // },
  ],

  // Run local dev server before starting tests
  webServer: {
    command: 'cd web && python3 -m http.server 8000',
    url: 'http://localhost:8000',
    reuseExistingServer: !process.env.CI,
    timeout: 30 * 1000, // Increased timeout for web server startup
  },
});
