// @ts-check

/**
 * Global setup that runs before all tests
 * Ensures the web server is fully ready and pre-loads catalog data
 */
async function globalSetup() {
  const maxRetries = 10;
  const retryDelay = 1000;
  const baseURL = 'http://localhost:8000';

  console.log('Waiting for web server to be ready...');

  // Wait for web server to respond
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch(baseURL);
      if (response.ok) {
        console.log('Web server is ready!');
        break;
      }
    } catch (error) {
      if (i === maxRetries - 1) {
        throw new Error(`Web server not ready after ${maxRetries} attempts`);
      }
      await new Promise(resolve => setTimeout(resolve, retryDelay));
    }
  }

  // Pre-load catalog.jsonl to warm up the cache/server
  console.log('Pre-loading catalog data...');
  try {
    const catalogResponse = await fetch(`${baseURL}/js/catalog.jsonl`);
    if (!catalogResponse.ok) {
      console.warn('Warning: Could not pre-load catalog.jsonl');
    } else {
      await catalogResponse.text(); // Actually read the response
      console.log('Catalog data pre-loaded successfully');
    }
  } catch (error) {
    console.warn('Warning: Error pre-loading catalog:', error.message);
  }

  console.log('Global setup complete');
}

module.exports = globalSetup;
