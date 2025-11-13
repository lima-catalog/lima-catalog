/**
 * Tests for filters.js
 */

import { runner, assert } from './test-framework.js';
import { getKeywordCounts, getCategoryCounts, applyFilters, sortTemplates } from './filters.js';

// Sample test data
const sampleTemplates = [
    {
        name: 'alpine',
        path: 'alpine.yaml',
        repo: 'lima-vm/lima',
        category: 'containers',
        keywords: ['alpine', 'linux', 'docker'],
        is_official: true,
        last_checked: '2024-01-15'
    },
    {
        name: 'ubuntu',
        path: 'ubuntu.yaml',
        repo: 'lima-vm/lima',
        category: 'development',
        keywords: ['ubuntu', 'linux'],
        is_official: true,
        last_checked: '2024-01-20'
    },
    {
        name: 'custom',
        path: 'custom.yaml',
        repo: 'user/repo',
        category: 'containers',
        keywords: ['docker', 'k8s'],
        is_official: false,
        last_checked: '2024-01-10'
    }
];

// Test getKeywordCounts
runner.test('getKeywordCounts: counts keywords correctly', () => {
    const result = getKeywordCounts(sampleTemplates);
    const resultMap = new Map(result);
    assert.equal(resultMap.get('linux'), 2);
    assert.equal(resultMap.get('docker'), 2);
    assert.equal(resultMap.get('alpine'), 1);
    assert.equal(resultMap.get('ubuntu'), 1);
    assert.equal(resultMap.get('k8s'), 1);
});

runner.test('getKeywordCounts: sorts by count descending', () => {
    const result = getKeywordCounts(sampleTemplates);
    // First items should have higher counts
    assert.ok(result[0][1] >= result[1][1]);
    assert.ok(result[1][1] >= result[2][1]);
});

runner.test('getKeywordCounts: excludes specified keywords', () => {
    const excluded = new Set(['linux', 'docker']);
    const result = getKeywordCounts(sampleTemplates, excluded);
    const resultMap = new Map(result);
    assert.equal(resultMap.has('linux'), false);
    assert.equal(resultMap.has('docker'), false);
    assert.equal(resultMap.get('alpine'), 1);
});

runner.test('getKeywordCounts: handles templates without keywords', () => {
    const templates = [
        { name: 'test1' },
        { name: 'test2', keywords: ['foo'] }
    ];
    const result = getKeywordCounts(templates);
    assert.equal(result.length, 1);
    assert.equal(result[0][0], 'foo');
});

// Test getCategoryCounts
runner.test('getCategoryCounts: counts categories correctly', () => {
    const result = getCategoryCounts(sampleTemplates);
    const resultMap = new Map(result);
    assert.equal(resultMap.get('containers'), 2);
    assert.equal(resultMap.get('development'), 1);
});

runner.test('getCategoryCounts: sorts alphabetically', () => {
    const result = getCategoryCounts(sampleTemplates);
    assert.equal(result[0][0], 'containers');
    assert.equal(result[1][0], 'development');
});

runner.test('getCategoryCounts: handles templates without categories', () => {
    const templates = [
        { name: 'test1' },
        { name: 'test2', category: 'foo' }
    ];
    const result = getCategoryCounts(templates);
    assert.equal(result.length, 1);
    assert.equal(result[0][0], 'foo');
});

// Test applyFilters
runner.test('applyFilters: filters by search term', () => {
    const result = applyFilters(sampleTemplates, { searchTerm: 'alpine' });
    assert.equal(result.length, 1);
    assert.equal(result[0].name, 'alpine');
});

runner.test('applyFilters: search is case-insensitive', () => {
    const result = applyFilters(sampleTemplates, { searchTerm: 'ALPINE' });
    assert.equal(result.length, 1);
    assert.equal(result[0].name, 'alpine');
});

runner.test('applyFilters: searches across multiple fields', () => {
    const result = applyFilters(sampleTemplates, { searchTerm: 'lima-vm' });
    assert.equal(result.length, 2); // alpine and ubuntu both from lima-vm/lima
});

runner.test('applyFilters: filters by category', () => {
    const result = applyFilters(sampleTemplates, { selectedCategory: 'containers' });
    assert.equal(result.length, 2);
    assert.ok(result.every(t => t.category === 'containers'));
});

runner.test('applyFilters: filters by single keyword', () => {
    const result = applyFilters(sampleTemplates, { selectedKeywords: new Set(['docker']) });
    assert.equal(result.length, 2);
    assert.ok(result.every(t => t.keywords.includes('docker')));
});

runner.test('applyFilters: filters by multiple keywords (AND logic)', () => {
    const result = applyFilters(sampleTemplates, { selectedKeywords: new Set(['linux', 'docker']) });
    assert.equal(result.length, 1);
    assert.equal(result[0].name, 'alpine');
});

runner.test('applyFilters: filters by type (official)', () => {
    const result = applyFilters(sampleTemplates, { typeFilter: 'official' });
    assert.equal(result.length, 2);
    assert.ok(result.every(t => t.is_official));
});

runner.test('applyFilters: filters by type (community)', () => {
    const result = applyFilters(sampleTemplates, { typeFilter: 'community' });
    assert.equal(result.length, 1);
    assert.equal(result[0].is_official, false);
});

runner.test('applyFilters: combines multiple filters', () => {
    const result = applyFilters(sampleTemplates, {
        selectedCategory: 'containers',
        typeFilter: 'official'
    });
    assert.equal(result.length, 1);
    assert.equal(result[0].name, 'alpine');
});

runner.test('applyFilters: returns all templates with empty filters', () => {
    const result = applyFilters(sampleTemplates, {});
    assert.equal(result.length, 3);
});

// Test sortTemplates
runner.test('sortTemplates: sorts by name', () => {
    const templates = [...sampleTemplates];
    const result = sortTemplates(templates, 'name', new Map());
    assert.equal(result[0].name, 'alpine');
    assert.equal(result[1].name, 'custom');
    assert.equal(result[2].name, 'ubuntu');
});

runner.test('sortTemplates: sorts by stars', () => {
    const templates = [...sampleTemplates];
    const repositories = new Map([
        ['lima-vm/lima', { stars: 100 }],
        ['user/repo', { stars: 50 }]
    ]);
    const result = sortTemplates(templates, 'stars', repositories);
    // First two should be from lima-vm/lima (100 stars)
    assert.equal(result[0].repo, 'lima-vm/lima');
    assert.equal(result[1].repo, 'lima-vm/lima');
    assert.equal(result[2].repo, 'user/repo');
});

runner.test('sortTemplates: sorts by updated date', () => {
    const templates = [...sampleTemplates];
    const result = sortTemplates(templates, 'updated', new Map());
    assert.equal(result[0].name, 'ubuntu'); // 2024-01-20
    assert.equal(result[1].name, 'alpine'); // 2024-01-15
    assert.equal(result[2].name, 'custom'); // 2024-01-10
});

runner.test('sortTemplates: handles missing repository data', () => {
    const templates = [...sampleTemplates];
    const result = sortTemplates(templates, 'stars', new Map());
    // Should not throw error
    assert.equal(result.length, 3);
});
