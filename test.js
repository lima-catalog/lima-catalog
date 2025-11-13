#!/usr/bin/env node
/**
 * Node.js test runner for lima-catalog
 * Runs all unit tests and reports results
 */

import { runner } from './docs/js/test-framework.js';

// Mock minimal DOM for Node.js environment
// Mimics browser behavior: textContent escapes <, >, & but not quotes
global.document = {
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
    getElementById: () => null
};

// Import all test files
await import('./docs/js/urlHelpers.test.js');
await import('./docs/js/data.test.js');
await import('./docs/js/filters.test.js');
await import('./docs/js/templateCard.test.js');

// Run tests
console.log('🧪 Running lima-catalog test suite...\n');

const results = await runner.run();

console.log('\n' + '='.repeat(60));
console.log(`Tests: ${results.total} | Passed: ${results.passed} | Failed: ${results.failed}`);
console.log('='.repeat(60));

// Exit with appropriate code
process.exit(results.failed > 0 ? 1 : 0);
