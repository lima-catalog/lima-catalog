// @ts-check

/**
 * Global setup that runs before all tests
 * Ensures the web server is fully ready and responsive
 */
async function globalSetup() {
  const maxRetries = 10;
  const retryDelay = 1000;
  const baseURL = 'http://localhost:8000';

  console.log('Waiting for web server to be ready...');

  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch(baseURL);
      if (response.ok) {
        console.log('Web server is ready!');
        // Give it a bit more time to stabilize
        await new Promise(resolve => setTimeout(resolve, 2000));
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
