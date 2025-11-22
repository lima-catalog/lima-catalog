/**
 * Tests for sidebar.js - Sidebar rendering and interaction
 *
 * Note: sidebar.js has extensive keyboard navigation logic that relies on
 * DOM offsetTop calculations and complex focus management. These tests focus
 * on core rendering logic that innerHTML is being set with content.
 */

import { runner, assert } from './test-framework.js';
import {
    renderKeywordCloud,
    renderSelectedKeywords,
    renderCategoryList,
    updateSidebar
} from './sidebar.js';
import * as State from './state.js';

// Sample test data
const sampleTemplates = [
    { id: '1', name: 'alpine', keywords: ['linux', 'docker'], category: 'containers' },
    { id: '2', name: 'ubuntu', keywords: ['linux', 'desktop'], category: 'development' },
    { id: '3', name: 'fedora', keywords: ['linux', 'rpm'], category: 'development' }
];

// Mock DOM element creator
function createMockElement(tag, id = null) {
    const children = [];
    const eventListeners = {};

    return {
        tagName: tag.toUpperCase(),
        id: id || '',
        innerHTML: '',
        textContent: '',
        children,
        eventListeners,
        dataset: {},
        classList: {
            contains: (className) => false
        },
        offsetTop: 0,

        appendChild(child) {
            children.push(child);
        },

        querySelectorAll(selector) {
            return [];
        },

        addEventListener(event, handler) {
            if (!eventListeners[event]) eventListeners[event] = [];
            eventListeners[event].push(handler);
        },

        focus() {
            if (global.document) global.document.activeElement = this;
        }
    };
}

function setupSidebarDOM() {
    const elements = {
        selectedKeywords: createMockElement('div', 'selected-keywords'),
        keywordCloud: createMockElement('div', 'keyword-cloud'),
        categoryList: createMockElement('div', 'category-list'),
        search: createMockElement('input', 'search'),
        showOfficial: createMockElement('input', 'show-official'),
        showCommunity: createMockElement('input', 'show-community'),
        showDuplicates: createMockElement('input', 'show-duplicates'),
        showSimilars: createMockElement('input', 'show-similars'),
        sort: createMockElement('select', 'sort')
    };

    const originalGetElementById = global.document.getElementById;
    const originalQuerySelector = global.document.querySelector;
    const originalActiveElement = global.document.activeElement;

    global.document.getElementById = (id) => {
        const camelId = id.replace(/-([a-z])/g, (g) => g[1].toUpperCase());
        return elements[camelId] || null;
    };

    global.document.querySelector = (selector) => {
        if (selector === '.sidebar') {
            return createMockElement('div');
        }
        return null;
    };

    global.document.querySelectorAll = () => [];
    global.document.activeElement = null;

    return {
        elements,
        cleanup: () => {
            global.document.getElementById = originalGetElementById;
            global.document.querySelector = originalQuerySelector;
            global.document.activeElement = originalActiveElement;
        }
    };
}

// ============================================================================
// renderKeywordCloud() tests
// ============================================================================

runner.test('sidebar.js: renderKeywordCloud renders keywords with counts', () => {
    const { elements, cleanup } = setupSidebarDOM();

    State.setTemplates(sampleTemplates);
    State.clearAllSelections();

    const selectedKeywords = new Set();
    const cloudElement = elements.keywordCloud;

    renderKeywordCloud(sampleTemplates, selectedKeywords, cloudElement, () => {}, false);

    assert.ok(cloudElement.innerHTML.length > 0, 'Cloud should have content');
    assert.ok(cloudElement.innerHTML.includes('keyword-tag'), 'Should contain keyword-tag class');

    cleanup();
});

runner.test('sidebar.js: renderKeywordCloud shows empty message when no keywords', () => {
    const { elements, cleanup } = setupSidebarDOM();

    State.setTemplates([]);
    const selectedKeywords = new Set();
    const cloudElement = elements.keywordCloud;

    renderKeywordCloud([], selectedKeywords, cloudElement, () => {}, false);

    assert.ok(cloudElement.innerHTML.includes('No additional keywords available'),
        'Should show empty message');

    cleanup();
});

runner.test('sidebar.js: renderKeywordCloud includes keyword counts', () => {
    const { elements, cleanup } = setupSidebarDOM();

    State.setTemplates(sampleTemplates);
    State.clearAllSelections();

    const cloudElement = elements.keywordCloud;

    renderKeywordCloud(sampleTemplates, new Set(), cloudElement, () => {}, false);

    assert.ok(cloudElement.innerHTML.includes('keyword-count'), 'Should include count spans');

    cleanup();
});

// ============================================================================
// renderSelectedKeywords() tests
// ============================================================================

runner.test('sidebar.js: renderSelectedKeywords renders selected keywords', () => {
    const { elements, cleanup } = setupSidebarDOM();

    const selectedKeywords = new Set(['docker', 'linux']);
    const containerElement = elements.selectedKeywords;

    renderSelectedKeywords(selectedKeywords, containerElement, () => {}, false);

    assert.ok(containerElement.innerHTML.length > 0, 'Should have content');
    assert.ok(containerElement.innerHTML.includes('selected-keyword'),
        'Should contain selected-keyword class');

    cleanup();
});

runner.test('sidebar.js: renderSelectedKeywords shows empty when no selection', () => {
    const { elements, cleanup } = setupSidebarDOM();

    const selectedKeywords = new Set();
    const containerElement = elements.selectedKeywords;

    renderSelectedKeywords(selectedKeywords, containerElement, () => {}, false);

    assert.equal(containerElement.innerHTML, '', 'Should be empty with no selections');

    cleanup();
});

runner.test('sidebar.js: renderSelectedKeywords marks dynamic keywords', () => {
    const { elements, cleanup } = setupSidebarDOM();

    const selectedKeywords = new Set(['docker', 'org:lima-vm', 'org/repo:lima-vm/lima']);
    const containerElement = elements.selectedKeywords;

    renderSelectedKeywords(selectedKeywords, containerElement, () => {}, false);

    assert.ok(containerElement.innerHTML.length > 0, 'Should have content');
    assert.ok(containerElement.innerHTML.includes('selected-keyword-dynamic'),
        'Should mark dynamic keywords with special class');

    cleanup();
});

// ============================================================================
// renderCategoryList() tests
// ============================================================================

runner.test('sidebar.js: renderCategoryList renders categories with counts', () => {
    const { elements, cleanup } = setupSidebarDOM();

    const listElement = elements.categoryList;

    renderCategoryList(sampleTemplates, null, listElement, () => {});

    assert.ok(listElement.innerHTML.length > 0, 'Should have content');
    assert.ok(listElement.innerHTML.includes('category-item'),
        'Should contain category-item class');
    assert.ok(listElement.innerHTML.includes('category-count'), 'Should include counts');

    cleanup();
});

runner.test('sidebar.js: renderCategoryList marks selected category', () => {
    const { elements, cleanup } = setupSidebarDOM();

    const listElement = elements.categoryList;

    renderCategoryList(sampleTemplates, 'containers', listElement, () => {});

    assert.ok(listElement.innerHTML.includes('selected'),
        'Should mark selected category with selected class');
    assert.ok(listElement.innerHTML.includes('aria-pressed="true"'),
        'Should set aria-pressed for selected category');

    cleanup();
});

runner.test('sidebar.js: renderCategoryList shows empty message when no categories', () => {
    const { elements, cleanup } = setupSidebarDOM();

    const listElement = elements.categoryList;

    renderCategoryList([], null, listElement, () => {});

    assert.ok(listElement.innerHTML.includes('No categories available'),
        'Should show empty message');

    cleanup();
});

runner.test('sidebar.js: renderCategoryList renders multiple categories', () => {
    const { elements, cleanup } = setupSidebarDOM();

    const listElement = elements.categoryList;

    renderCategoryList(sampleTemplates, null, listElement, () => {});

    // Should have multiple category items
    const itemCount = (listElement.innerHTML.match(/category-item/g) || []).length;
    assert.ok(itemCount >= 2, 'Should have at least 2 categories');

    cleanup();
});

// ============================================================================
// updateSidebar() tests
// ============================================================================

runner.test('sidebar.js: updateSidebar calls all render functions', () => {
    const { elements, cleanup } = setupSidebarDOM();

    State.setTemplates(sampleTemplates);
    const selectedKeywords = new Set(['docker']);

    const state = {
        filteredTemplates: sampleTemplates,
        selectedKeywords: selectedKeywords,
        selectedCategory: 'containers'
    };

    updateSidebar(state, () => {}, () => {}, {});

    assert.ok(elements.keywordCloud.innerHTML.length > 0,
        'Keyword cloud should be updated');
    assert.ok(elements.selectedKeywords.innerHTML.length > 0,
        'Selected keywords should be updated');
    assert.ok(elements.categoryList.innerHTML.length > 0,
        'Category list should be updated');

    cleanup();
});

runner.test('sidebar.js: updateSidebar handles empty state', () => {
    const { elements, cleanup } = setupSidebarDOM();

    State.setTemplates([]);

    const state = {
        filteredTemplates: [],
        selectedKeywords: new Set(),
        selectedCategory: null
    };

    // Should not crash with empty state
    updateSidebar(state, () => {}, () => {}, {});

    assert.ok(true, 'Should handle empty state without errors');

    cleanup();
});

// Note: Full event handler and keyboard navigation testing would require
// extensive DOM mocking including offsetTop calculations. These complex
// interactions are covered through the HTML structure tests above and
// would benefit from integration testing.
