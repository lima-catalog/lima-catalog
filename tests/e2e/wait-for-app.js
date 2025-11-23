// @ts-check

/**
 * Helper to ensure the app is fully initialized before tests run
 * Waits for catalog.jsonl to be loaded AND app JavaScript to be ready
 */
async function waitForApp(page) {
  // Wait for templates grid to exist
  await page.waitForSelector('#templates-grid .template-card', { timeout: 10000 });

  // Wait for the app's state to be initialized AND data to be loaded
  await page.waitForFunction(() => {
    // Check that core app functions are available
    const hasAppActions = window.appActions &&
           typeof window.appActions.applyFiltersAndRender === 'function';

    // Check that templates data is actually loaded (not empty)
    const hasTemplates = window.state &&
                        window.state.getTemplates &&
                        window.state.getTemplates().length > 0;

    return hasAppActions && hasTemplates;
  }, { timeout: 15000 }); // Longer timeout for data loading

  // Small additional delay to ensure DOM updates have completed
  await page.waitForTimeout(500);
}

module.exports = { waitForApp };
