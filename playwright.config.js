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

  // Global timeout for the entire test suite
  globalTimeout: process.env.CI ? 15 * 60 * 1000 : undefined, // 15 minutes max in CI

  // Global setup to run before all tests
  globalSetup: require.resolve('./tests/e2e/global-setup.js'),

  // Reporter to use
  // 'line' reporter shows one line per completed test (better for streaming in CI, saved to test-output.txt)
  // 'github' reporter adds annotations to GitHub Actions
  // 'json' reporter creates structured data for automated analysis (saved to test-results/results.json)
  reporter: process.env.CI
    ? [['line'], ['github'], ['json', { outputFile: 'test-results/results.json' }]]
    : 'list',

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
            '--disable-gpu',
            '--disable-dev-shm-usage',
            // Use --single-process only in local development VM (not in GitHub CI)
            // GitHub CI has better IPC support but --single-process causes resource exhaustion
            ...(process.env.GITHUB_ACTIONS ? [] : ['--single-process']),
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
    // Redirect web server logs to file to avoid flooding CI output
    // Logs are saved to test-results/webserver.log for debugging
    command: 'mkdir -p test-results && cd web && python3 -m http.server 8000 2> ../test-results/webserver.log',
    url: 'http://localhost:8000',
    reuseExistingServer: !process.env.CI,
    timeout: 30 * 1000, // Increased timeout for web server startup
  },
});
