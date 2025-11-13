/**
 * Simple browser-based test framework for ES6 modules
 */

class TestRunner {
    constructor() {
        this.tests = [];
        this.results = {
            passed: 0,
            failed: 0,
            total: 0
        };
    }

    /**
     * Register a test
     * @param {string} name - Test name
     * @param {Function} fn - Test function
     */
    test(name, fn) {
        this.tests.push({ name, fn });
    }

    /**
     * Run all registered tests
     * @returns {Promise<Object>} Test results
     */
    async run() {
        this.results = { passed: 0, failed: 0, total: 0, details: [] };

        for (const test of this.tests) {
            this.results.total++;
            try {
                await test.fn();
                this.results.passed++;
                this.results.details.push({
                    name: test.name,
                    status: 'passed',
                    error: null
                });
                this.log('✓', test.name, 'color: green');
            } catch (error) {
                this.results.failed++;
                this.results.details.push({
                    name: test.name,
                    status: 'failed',
                    error: error.message
                });
                this.log('✗', test.name, 'color: red');
                this.log('  Error:', error.message, 'color: red; margin-left: 20px');
            }
        }

        return this.results;
    }

    /**
     * Log message with styling
     * @param {string} prefix - Log prefix
     * @param {string} message - Log message
     * @param {string} style - CSS style
     */
    log(prefix, message, style = '') {
        if (typeof document !== 'undefined') {
            const output = document.getElementById('test-output');
            if (output) {
                const line = document.createElement('div');
                line.textContent = `${prefix} ${message}`;
                line.style.cssText = style;
                output.appendChild(line);
            }
        }
        console.log(`${prefix} ${message}`);
    }

    /**
     * Display summary
     */
    displaySummary() {
        const summary = `\n${'='.repeat(50)}\nTests: ${this.results.total} | Passed: ${this.results.passed} | Failed: ${this.results.failed}\n${'='.repeat(50)}`;
        const style = this.results.failed === 0 ? 'color: green; font-weight: bold' : 'color: red; font-weight: bold';
        this.log('', summary, style);
    }
}

/**
 * Assertion functions
 */
export const assert = {
    /**
     * Assert that value is truthy
     */
    ok(value, message = 'Expected value to be truthy') {
        if (!value) {
            throw new Error(message);
        }
    },

    /**
     * Assert that two values are equal
     */
    equal(actual, expected, message) {
        if (actual !== expected) {
            throw new Error(message || `Expected ${JSON.stringify(expected)} but got ${JSON.stringify(actual)}`);
        }
    },

    /**
     * Assert that two values are deeply equal
     */
    deepEqual(actual, expected, message) {
        const actualStr = JSON.stringify(actual);
        const expectedStr = JSON.stringify(expected);
        if (actualStr !== expectedStr) {
            throw new Error(message || `Expected ${expectedStr} but got ${actualStr}`);
        }
    },

    /**
     * Assert that function throws an error
     */
    throws(fn, expectedError, message) {
        try {
            fn();
            throw new Error(message || 'Expected function to throw but it did not');
        } catch (error) {
            if (expectedError && !error.message.includes(expectedError)) {
                throw new Error(message || `Expected error containing "${expectedError}" but got "${error.message}"`);
            }
        }
    },

    /**
     * Assert that value includes substring or element
     */
    includes(haystack, needle, message) {
        const hasIt = typeof haystack === 'string'
            ? haystack.includes(needle)
            : Array.isArray(haystack) && haystack.includes(needle);

        if (!hasIt) {
            throw new Error(message || `Expected ${JSON.stringify(haystack)} to include ${JSON.stringify(needle)}`);
        }
    }
};

// Export singleton instance
export const runner = new TestRunner();
