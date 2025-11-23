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
        console.log('Global setup complete');
        return;
      }
    } catch (error) {
      if (i === maxRetries - 1) {
        throw new Error(`Web server not ready after ${maxRetries} attempts`);
      }
      await new Promise(resolve => setTimeout(resolve, retryDelay));
    }
  }
}

module.exports = globalSetup;
