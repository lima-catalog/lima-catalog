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
        official: true,
        stars: 100,
        updated_at: '2024-01-15',
        notability_score: 150.5
    },
    {
        name: 'ubuntu',
        path: 'ubuntu.yaml',
        repo: 'lima-vm/lima',
        category: 'development',
        keywords: ['ubuntu', 'linux'],
        official: true,
        stars: 100,
        updated_at: '2024-01-20',
        notability_score: 75.3
    },
    {
        name: 'custom',
        path: 'custom.yaml',
        repo: 'user/repo',
        category: 'containers',
        keywords: ['docker', 'k8s'],
        official: false,
        stars: 50,
        updated_at: '2024-01-10',
        notability_score: 250.8
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
    assert.ok(result.every(t => t.official));
});

runner.test('applyFilters: filters by type (community)', () => {
    const result = applyFilters(sampleTemplates, { typeFilter: 'community' });
    assert.equal(result.length, 1);
    assert.equal(result[0].official, false);
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
    const result = sortTemplates(templates, 'name');
    assert.equal(result[0].name, 'alpine');
    assert.equal(result[1].name, 'custom');
    assert.equal(result[2].name, 'ubuntu');
});

runner.test('sortTemplates: sorts by stars', () => {
    const templates = [...sampleTemplates];
    const result = sortTemplates(templates, 'stars');
    // First two should have 100 stars (alpine, ubuntu)
    assert.equal(result[0].stars, 100);
    assert.equal(result[1].stars, 100);
    assert.equal(result[2].stars, 50);
});

runner.test('sortTemplates: sorts by updated date', () => {
    const templates = [...sampleTemplates];
    const result = sortTemplates(templates, 'updated');
    assert.equal(result[0].name, 'ubuntu'); // 2024-01-20
    assert.equal(result[1].name, 'alpine'); // 2024-01-15
    assert.equal(result[2].name, 'custom'); // 2024-01-10
});

runner.test('sortTemplates: sorts by notability score', () => {
    const templates = [...sampleTemplates];
    const result = sortTemplates(templates, 'notability');
    assert.equal(result[0].name, 'custom'); // 250.8
    assert.equal(result[1].name, 'alpine'); // 150.5
    assert.equal(result[2].name, 'ubuntu'); // 75.3
});

runner.test('sortTemplates: handles missing stars data', () => {
    const templates = [...sampleTemplates];
    const result = sortTemplates(templates, 'stars');
    // Should not throw error
    assert.equal(result.length, 3);
});

// Test debug mode breakdown sorting
const templatesWithBreakdown = [
    {
        name: 'template1',
        notability_score_breakdown: {
            message: 100,
            provision: 50,
            parameters: 40,
            env_vars: 30,
            probes: 20,
            unusual_images: 30,
            comments: 10,
            stars: 5
        }
    },
    {
        name: 'template2',
        notability_score_breakdown: {
            message: 0,
            provision: 100,
            parameters: 60,
            env_vars: 50,
            probes: 30,
            unusual_images: 0,
            comments: 20,
            stars: 10
        }
    },
    {
        name: 'template3',
        notability_score_breakdown: {
            message: 100,
            provision: 25,
            parameters: 80,
            env_vars: 10,
            probes: 15,
            unusual_images: 30,
            comments: 50,
            stars: 15
        }
    }
];

runner.test('sortTemplates: sorts by breakdown-message', () => {
    const templates = [...templatesWithBreakdown];
    const result = sortTemplates(templates, 'breakdown-message');
    // First two have 100, last one has 0
    assert.equal(result[0].notability_score_breakdown.message, 100);
    assert.equal(result[1].notability_score_breakdown.message, 100);
    assert.equal(result[2].notability_score_breakdown.message, 0);
});

runner.test('sortTemplates: sorts by breakdown-provision', () => {
    const templates = [...templatesWithBreakdown];
    const result = sortTemplates(templates, 'breakdown-provision');
    assert.equal(result[0].name, 'template2'); // 100
    assert.equal(result[1].name, 'template1'); // 50
    assert.equal(result[2].name, 'template3'); // 25
});

runner.test('sortTemplates: sorts by breakdown-parameters', () => {
    const templates = [...templatesWithBreakdown];
    const result = sortTemplates(templates, 'breakdown-parameters');
    assert.equal(result[0].name, 'template3'); // 80
    assert.equal(result[1].name, 'template2'); // 60
    assert.equal(result[2].name, 'template1'); // 40
});

runner.test('sortTemplates: sorts by breakdown-comments', () => {
    const templates = [...templatesWithBreakdown];
    const result = sortTemplates(templates, 'breakdown-comments');
    assert.equal(result[0].name, 'template3'); // 50
    assert.equal(result[1].name, 'template2'); // 20
    assert.equal(result[2].name, 'template1'); // 10
});

runner.test('sortTemplates: sorts by breakdown-unusual_images', () => {
    const templates = [...templatesWithBreakdown];
    const result = sortTemplates(templates, 'breakdown-unusual_images');
    // First two have 30, last one has 0
    assert.equal(result[0].notability_score_breakdown.unusual_images, 30);
    assert.equal(result[1].notability_score_breakdown.unusual_images, 30);
    assert.equal(result[2].notability_score_breakdown.unusual_images, 0);
});

runner.test('sortTemplates: handles missing breakdown data', () => {
    const templates = [
        { name: 'no-breakdown' },
        { name: 'with-breakdown', notability_score_breakdown: { message: 100 } }
    ];
    const result = sortTemplates(templates, 'breakdown-message');
    // Should not throw error, with-breakdown should be first
    assert.equal(result[0].name, 'with-breakdown');
    assert.equal(result[1].name, 'no-breakdown');
});
