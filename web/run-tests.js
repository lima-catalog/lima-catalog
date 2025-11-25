#!/usr/bin/env node
/**
 * Node.js test runner for frontend unit tests
 *
 * This allows the browser-based tests to run in CI without a browser.
 * Sets up minimal DOM mocks required by the test files.
 */

import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// =============================================================================
// DOM ENVIRONMENT SETUP
// Must be done before importing any test files
// =============================================================================

// Mock window object
global.window = {
    location: {
        href: 'http://localhost:3000',
        origin: 'http://localhost:3000',
        protocol: 'http:',
        host: 'localhost:3000',
        hostname: 'localhost',
        port: '3000',
        pathname: '/',
        search: '',
        hash: '',
        toString: () => 'http://localhost:3000'
    },
    matchMedia: (query) => ({
        matches: false,
        media: query,
        addEventListener: () => {},
        removeEventListener: () => {}
    }),
    history: {
        pushState: () => {},
        replaceState: () => {}
    },
    scrollTo: () => {},
    requestAnimationFrame: (cb) => setTimeout(cb, 0),
    getComputedStyle: () => ({
        getPropertyValue: () => ''
    })
};

// Mock localStorage
global.localStorage = (() => {
    let store = {};
    return {
        getItem: (key) => store[key] || null,
        setItem: (key, value) => { store[key] = String(value); },
        removeItem: (key) => { delete store[key]; },
        clear: () => { store = {}; },
        get length() { return Object.keys(store).length; },
        key: (i) => Object.keys(store)[i] || null
    };
})();

// Mock CSS.escape (used in appActions.js)
global.CSS = {
    escape: (str) => str.replace(/([^\w-])/g, '\\$1')
};

// Helper to create mock elements
function createMockElement(tag = 'div') {
    let _innerHTML = '';
    let _textContent = '';
    const children = [];
    const eventListeners = {};
    const style = {};
    const dataset = {};
    const classList = {
        _classes: [],
        add(...classes) { this._classes.push(...classes); },
        remove(...classes) { this._classes = this._classes.filter(c => !classes.includes(c)); },
        contains(className) { return this._classes.includes(className); },
        toggle(className) {
            if (this.contains(className)) {
                this.remove(className);
                return false;
            }
            this.add(className);
            return true;
        }
    };

    const element = {
        tagName: tag.toUpperCase(),
        nodeName: tag.toUpperCase(),
        id: '',
        className: '',
        children,
        childNodes: children,
        style,
        dataset,
        classList,
        eventListeners,
        attributes: {},
        offsetParent: {},
        offsetTop: 0,
        offsetLeft: 0,
        scrollTop: 0,
        scrollLeft: 0,

        get innerHTML() { return _innerHTML; },
        set innerHTML(value) { _innerHTML = value; },
        get textContent() { return _textContent; },
        set textContent(value) { _textContent = value; _innerHTML = escapeHTMLContent(value); },
        get value() { return this._value || ''; },
        set value(v) { this._value = v; },
        get checked() { return this._checked || false; },
        set checked(v) { this._checked = v; },
        get options() { return this._options || []; },
        set options(v) { this._options = v; },

        appendChild(child) { children.push(child); return child; },
        removeChild(child) {
            const idx = children.indexOf(child);
            if (idx > -1) children.splice(idx, 1);
            return child;
        },
        insertBefore(newNode, refNode) {
            const idx = children.indexOf(refNode);
            if (idx > -1) children.splice(idx, 0, newNode);
            else children.push(newNode);
            return newNode;
        },

        querySelector(selector) { return null; },
        querySelectorAll(selector) { return []; },

        getAttribute(name) { return this.attributes[name] || null; },
        setAttribute(name, value) { this.attributes[name] = value; },
        removeAttribute(name) { delete this.attributes[name]; },
        hasAttribute(name) { return name in this.attributes; },

        addEventListener(event, handler, options) {
            if (!eventListeners[event]) eventListeners[event] = [];
            eventListeners[event].push(handler);
        },
        removeEventListener(event, handler) {
            if (eventListeners[event]) {
                eventListeners[event] = eventListeners[event].filter(h => h !== handler);
            }
        },
        dispatchEvent(event) {
            const handlers = eventListeners[event.type] || [];
            handlers.forEach(h => h(event));
            return true;
        },

        focus() { global.document.activeElement = this; },
        blur() { if (global.document.activeElement === this) global.document.activeElement = null; },
        click() { this.dispatchEvent({ type: 'click', target: this }); },

        matches(selector) {
            if (selector.startsWith('.')) return classList.contains(selector.slice(1));
            if (selector.startsWith('#')) return this.id === selector.slice(1);
            return this.tagName.toLowerCase() === selector.toLowerCase();
        },

        contains(node) { return children.includes(node); },
        closest(selector) { return null; },
        getBoundingClientRect() {
            return { top: 0, left: 0, bottom: 0, right: 0, width: 0, height: 0 };
        }
    };

    return element;
}

// HTML escape helper for textContent -> innerHTML
// Browser behavior: only escapes <, >, & (NOT quotes)
function escapeHTMLContent(str) {
    if (str == null) return '';
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');
}

// Mock document object
global.document = {
    activeElement: null,
    body: createMockElement('body'),
    documentElement: createMockElement('html'),
    title: '',

    getElementById(id) { return null; },
    querySelector(selector) { return null; },
    querySelectorAll(selector) { return []; },
    getElementsByClassName(className) { return []; },
    getElementsByTagName(tagName) { return []; },

    createElement(tag) { return createMockElement(tag); },
    createTextNode(text) { return { nodeType: 3, textContent: text }; },
    createDocumentFragment() { return createMockElement('fragment'); },

    addEventListener(event, handler) {
        if (!this._listeners) this._listeners = {};
        if (!this._listeners[event]) this._listeners[event] = [];
        this._listeners[event].push(handler);
    },
    removeEventListener(event, handler) {
        if (this._listeners && this._listeners[event]) {
            this._listeners[event] = this._listeners[event].filter(h => h !== handler);
        }
    },
    dispatchEvent(event) {
        if (this._listeners && this._listeners[event.type]) {
            this._listeners[event.type].forEach(h => h(event));
        }
        return true;
    }
};

// =============================================================================
// TEST RUNNER
// =============================================================================

async function main() {
    console.log('');
    console.log('='.repeat(60));
    console.log('Lima Catalog - Frontend Unit Tests (Node.js Runner)');
    console.log('='.repeat(60));
    console.log('');

    try {
        // Import test framework
        const { runner } = await import('./js/test-framework.js');

        // Import all test files (must match tests.html imports)
        const testFiles = [
            './js/urlHelpers.test.js',
            './js/data.test.js',
            './js/filters.test.js',
            './js/templateCard.test.js',
            './js/theme.test.js',
            './js/state.test.js',
            './js/utils.test.js',
            './js/sidebar.test.js',
            './js/modal.test.js',
            './js/app.test.js',
            './js/appActions.test.js'
        ];

        console.log(`Loading ${testFiles.length} test modules...`);
        console.log('');

        for (const file of testFiles) {
            await import(file);
        }

        // Run all tests
        const results = await runner.run();

        // Display summary
        console.log('');
        runner.displaySummary();

        // Output for CI parsing
        console.log('');
        console.log('TAP Version 14');
        console.log(`1..${results.total}`);

        let testNum = 0;
        for (const detail of results.details) {
            testNum++;
            if (detail.status === 'passed') {
                console.log(`ok ${testNum} - ${detail.name}`);
            } else {
                console.log(`not ok ${testNum} - ${detail.name}`);
                if (detail.error) {
                    console.log(`  ---`);
                    console.log(`  message: ${detail.error}`);
                    console.log(`  ---`);
                }
            }
        }

        // Exit with appropriate code
        if (results.failed > 0) {
            console.log('');
            console.error(`FAILED: ${results.failed} test(s) failed`);
            process.exit(1);
        } else {
            console.log('');
            console.log(`SUCCESS: All ${results.passed} tests passed`);
            process.exit(0);
        }

    } catch (error) {
        console.error('');
        console.error('Test runner failed with error:');
        console.error(error);
        process.exit(1);
    }
}

main();
