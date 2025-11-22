#!/usr/bin/env node
/**
 * Node.js test runner for lima-catalog
 * Runs all unit tests and reports results
 */

import { runner } from './web/js/test-framework.js';

// Mock minimal DOM for Node.js environment
// Mimics browser behavior: textContent escapes <, >, & but not quotes
global.document = {
    activeElement: null,
    createElement: (tag) => {
        return {
            textContent: '',
            innerHTML: '',
            set textContent(value) {
                // Simple HTML escaping matching browser textContent behavior
                this.innerHTML = String(value)
                    .replace(/&/g, '&amp;')
                    .replace(/</g, '&lt;')
                    .replace(/>/g, '&gt;');
            }
        };
    },
    getElementById: (id) => null,
    documentElement: {
        setAttribute: () => {},
        removeAttribute: () => {}
    },
    querySelector: (selector) => null,
    querySelectorAll: (selector) => [],
    addEventListener: (event, handler) => {},
    removeEventListener: (event, handler) => {},
    body: {
        style: {},
        appendChild: () => {}
    }
};

// Mock window object for URL and history testing
global.window = {
    location: {
        href: 'http://localhost:3000',
        search: '',
        pathname: '/',
        protocol: 'http:',
        host: 'localhost:3000',
        hostname: 'localhost',
        port: '3000'
    },
    history: {
        pushState: () => {},
        replaceState: () => {}
    },
    matchMedia: (query) => ({
        matches: false,
        media: query,
        addEventListener: () => {},
        removeEventListener: () => {}
    }),
    addEventListener: (event, handler) => {},
    removeEventListener: (event, handler) => {}
};

// Import all test files
await import('./web/js/urlHelpers.test.js');
await import('./web/js/data.test.js');
await import('./web/js/filters.test.js');
await import('./web/js/templateCard.test.js');
await import('./web/js/theme.test.js');
await import('./web/js/state.test.js');
await import('./web/js/utils.test.js');
await import('./web/js/appActions.test.js');
await import('./web/js/app.test.js');
await import('./web/js/sidebar.test.js');
await import('./web/js/modal.test.js');

// Run tests
console.log('🧪 Running lima-catalog test suite...\n');

const results = await runner.run();

console.log('\n' + '='.repeat(60));
console.log(`Tests: ${results.total} | Passed: ${results.passed} | Failed: ${results.failed}`);
console.log('='.repeat(60));

// Exit with appropriate code
process.exit(results.failed > 0 ? 1 : 0);
