/**
 * Tests for state.js - Application state management
 */

import { runner, assert } from './test-framework.js';
import * as State from './state.js';

// Test templates data
const sampleTemplates = [
    { name: 'alpine', keywords: ['linux', 'docker'], category: 'containers' },
    { name: 'ubuntu', keywords: ['linux'], category: 'development' },
    { name: 'fedora', keywords: ['linux', 'rpm'], category: 'development' }
];

// Helper to reset state between tests
function resetState() {
    State.setTemplates([]);
    State.setFilteredTemplates([]);
    State.clearKeywordSelection();
    State.clearCategorySelection();
    State.setFocusedTemplate(null);
    // Reset debug mode
    if (State.isDebugMode()) {
        State.toggleDebugMode();
    }
}

runner.test('state.js: getTemplates returns empty array initially', () => {
    resetState();
    const templates = State.getTemplates();
    assert.deepEqual(templates, []);
});

runner.test('state.js: setTemplates stores templates', () => {
    resetState();
    State.setTemplates(sampleTemplates);
    const templates = State.getTemplates();
    assert.equal(templates.length, 3);
    assert.equal(templates[0].name, 'alpine');
});

runner.test('state.js: setTemplates replaces existing templates', () => {
    resetState();
    State.setTemplates(sampleTemplates);
    State.setTemplates([sampleTemplates[0]]);
    const templates = State.getTemplates();
    assert.equal(templates.length, 1);
    assert.equal(templates[0].name, 'alpine');
});

runner.test('state.js: getFilteredTemplates returns empty array initially', () => {
    resetState();
    const filtered = State.getFilteredTemplates();
    assert.deepEqual(filtered, []);
});

runner.test('state.js: setFilteredTemplates stores filtered results', () => {
    resetState();
    State.setFilteredTemplates([sampleTemplates[0], sampleTemplates[1]]);
    const filtered = State.getFilteredTemplates();
    assert.equal(filtered.length, 2);
    assert.equal(filtered[0].name, 'alpine');
    assert.equal(filtered[1].name, 'ubuntu');
});

runner.test('state.js: toggleKeywordSelection adds keyword', () => {
    resetState();
    State.toggleKeywordSelection('docker');
    const keywords = State.getSelectedKeywords();
    assert.ok(keywords.has('docker'));
});

runner.test('state.js: toggleKeywordSelection removes keyword on second toggle', () => {
    resetState();
    State.toggleKeywordSelection('docker');
    State.toggleKeywordSelection('docker');
    const keywords = State.getSelectedKeywords();
    assert.ok(!keywords.has('docker'));
});

runner.test('state.js: toggleKeywordSelection supports multiple keywords', () => {
    resetState();
    State.toggleKeywordSelection('docker');
    State.toggleKeywordSelection('linux');
    State.toggleKeywordSelection('k8s');
    const keywords = State.getSelectedKeywords();
    assert.equal(keywords.size, 3);
    assert.ok(keywords.has('docker'));
    assert.ok(keywords.has('linux'));
    assert.ok(keywords.has('k8s'));
});

runner.test('state.js: clearKeywordSelection clears all keywords', () => {
    resetState();
    State.toggleKeywordSelection('docker');
    State.toggleKeywordSelection('linux');
    State.clearKeywordSelection();
    const keywords = State.getSelectedKeywords();
    assert.equal(keywords.size, 0);
});

runner.test('state.js: getSelectedKeywords returns Set', () => {
    resetState();
    const keywords = State.getSelectedKeywords();
    assert.ok(keywords instanceof Set);
});

runner.test('state.js: setCategorySelection sets category', () => {
    resetState();
    State.setCategorySelection('containers');
    assert.equal(State.getSelectedCategory(), 'containers');
});

runner.test('state.js: setCategorySelection replaces previous category', () => {
    resetState();
    State.setCategorySelection('containers');
    State.setCategorySelection('development');
    assert.equal(State.getSelectedCategory(), 'development');
});

runner.test('state.js: setCategorySelection allows null category', () => {
    resetState();
    State.setCategorySelection('containers');
    State.setCategorySelection(null);
    assert.equal(State.getSelectedCategory(), null);
});

runner.test('state.js: getSelectedCategory returns null initially', () => {
    resetState();
    assert.equal(State.getSelectedCategory(), null);
});

runner.test('state.js: toggleCategorySelection sets category when null', () => {
    resetState();
    State.toggleCategorySelection('containers');
    assert.equal(State.getSelectedCategory(), 'containers');
});

runner.test('state.js: toggleCategorySelection clears category on second toggle', () => {
    resetState();
    State.toggleCategorySelection('containers');
    State.toggleCategorySelection('containers');
    assert.equal(State.getSelectedCategory(), null);
});

runner.test('state.js: toggleCategorySelection switches to new category', () => {
    resetState();
    State.toggleCategorySelection('containers');
    State.toggleCategorySelection('development');
    assert.equal(State.getSelectedCategory(), 'development');
});

runner.test('state.js: clearCategorySelection clears category', () => {
    resetState();
    State.setCategorySelection('containers');
    State.clearCategorySelection();
    assert.equal(State.getSelectedCategory(), null);
});

runner.test('state.js: setFocusedTemplate sets focused template', () => {
    resetState();
    State.setFocusedTemplate(sampleTemplates[0]);
    assert.deepEqual(State.getFocusedTemplate(), sampleTemplates[0]);
});

runner.test('state.js: setFocusedTemplate can clear focused template', () => {
    resetState();
    State.setFocusedTemplate(sampleTemplates[0]);
    State.setFocusedTemplate(null);
    assert.equal(State.getFocusedTemplate(), null);
});

runner.test('state.js: getFocusedTemplate returns null initially', () => {
    resetState();
    assert.equal(State.getFocusedTemplate(), null);
});

runner.test('state.js: isDebugMode returns false initially', () => {
    resetState();
    assert.equal(State.isDebugMode(), false);
});

runner.test('state.js: toggleDebugMode enables debug mode', () => {
    resetState();
    const result = State.toggleDebugMode();
    assert.equal(result, true);
    assert.equal(State.isDebugMode(), true);
});

runner.test('state.js: toggleDebugMode disables debug mode on second toggle', () => {
    resetState();
    State.toggleDebugMode();
    const result = State.toggleDebugMode();
    assert.equal(result, false);
    assert.equal(State.isDebugMode(), false);
});

runner.test('state.js: clearAllSelections clears keywords and category', () => {
    resetState();
    State.toggleKeywordSelection('docker');
    State.toggleKeywordSelection('linux');
    State.setCategorySelection('containers');
    State.clearAllSelections();
    assert.equal(State.getSelectedKeywords().size, 0);
    assert.equal(State.getSelectedCategory(), null);
});

runner.test('state.js: clearAllSelections does not clear templates', () => {
    resetState();
    State.setTemplates(sampleTemplates);
    State.clearAllSelections();
    assert.equal(State.getTemplates().length, 3);
});

runner.test('state.js: clearAllSelections does not clear filtered templates', () => {
    resetState();
    State.setFilteredTemplates([sampleTemplates[0]]);
    State.clearAllSelections();
    assert.equal(State.getFilteredTemplates().length, 1);
});

runner.test('state.js: clearAllSelections does not clear focused template', () => {
    resetState();
    State.setFocusedTemplate(sampleTemplates[0]);
    State.clearAllSelections();
    assert.deepEqual(State.getFocusedTemplate(), sampleTemplates[0]);
});

runner.test('state.js: clearAllSelections does not affect debug mode', () => {
    resetState();
    State.toggleDebugMode();
    State.clearAllSelections();
    assert.equal(State.isDebugMode(), true);
});
