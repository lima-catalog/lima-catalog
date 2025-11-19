/**
 * Tests for filters.js
 */

import { runner, assert } from './test-framework.js';
import { getKeywordCounts, getCategoryCounts, applyFilters, sortTemplates, getDynamicKeywords } from './filters.js';

// Sample test data
const sampleTemplates = [
    {
        name: 'alpine',
        path: 'alpine.yaml',
        repo: 'lima-vm/lima',
        org: 'lima-vm',
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
        org: 'lima-vm',
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
        org: 'user',
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

// Test getDynamicKeywords
const multiRepoTemplates = [
    { name: 't1', repo: 'lima-vm/lima', org: 'lima-vm' },
    { name: 't2', repo: 'lima-vm/lima', org: 'lima-vm' },
    { name: 't3', repo: 'lima-vm/other', org: 'lima-vm' },
    { name: 't4', repo: 'user/repo', org: 'user' }
];

runner.test('getDynamicKeywords: returns empty array when no focused template', () => {
    const result = getDynamicKeywords(multiRepoTemplates, null);
    assert.equal(result.length, 0);
});

runner.test('getDynamicKeywords: returns empty array when focused template missing org/repo', () => {
    const result = getDynamicKeywords(multiRepoTemplates, { name: 't1' });
    assert.equal(result.length, 0);
});

runner.test('getDynamicKeywords: returns org/repo keyword when only one repo with multiple templates', () => {
    const templates = [
        { name: 't1', repo: 'user/repo', org: 'user' },
        { name: 't2', repo: 'user/repo', org: 'user' }
    ];
    const focusedTemplate = templates[0];
    const result = getDynamicKeywords(templates, focusedTemplate);

    assert.equal(result.length, 1);
    assert.equal(result[0][0], 'user/repo');
    assert.equal(result[0][1], 2); // 2 templates in this repo
    assert.equal(result[0][2], true); // isDynamic flag
});

runner.test('getDynamicKeywords: does not return keyword when only one template in single repo', () => {
    const templates = [
        { name: 't1', repo: 'user/repo', org: 'user' }
    ];
    const focusedTemplate = templates[0];
    const result = getDynamicKeywords(templates, focusedTemplate);

    assert.equal(result.length, 0); // No keywords when only 1 template
});

runner.test('getDynamicKeywords: returns org and org/repo keywords when multiple repos from same org', () => {
    const focusedTemplate = multiRepoTemplates[0]; // lima-vm/lima
    const result = getDynamicKeywords(multiRepoTemplates, focusedTemplate);

    // Should have: lima-vm, lima-vm/lima, lima-vm/other
    assert.equal(result.length, 3);

    const keywords = result.map(r => r[0]);
    assert.ok(keywords.includes('lima-vm'));
    assert.ok(keywords.includes('lima-vm/lima'));
    assert.ok(keywords.includes('lima-vm/other'));

    // Check counts
    const resultMap = new Map(result.map(r => [r[0], r[1]]));
    assert.equal(resultMap.get('lima-vm'), 3); // 2 from lima + 1 from other
    assert.equal(resultMap.get('lima-vm/lima'), 2);
    assert.equal(resultMap.get('lima-vm/other'), 1);

    // Check isDynamic flags
    assert.ok(result.every(r => r[2] === true));
});

runner.test('getDynamicKeywords: works for templates from different org with single repo', () => {
    const focusedTemplate = multiRepoTemplates[3]; // user/repo (only repo from user org)
    const result = getDynamicKeywords(multiRepoTemplates, focusedTemplate);

    // Should have: org/repo:user/repo (only 1 repo from user org, but it has 1 template, so no keyword)
    assert.equal(result.length, 0); // Only 1 template in this repo
});

// Test applyFilters with dynamic keywords
runner.test('applyFilters: filters by org dynamic keyword', () => {
    const templates = [
        { name: 't1', repo: 'lima-vm/lima', org: 'lima-vm', keywords: [] },
        { name: 't2', repo: 'lima-vm/other', org: 'lima-vm', keywords: [] },
        { name: 't3', repo: 'user/repo', org: 'user', keywords: [] }
    ];
    const result = applyFilters(templates, { selectedKeywords: new Set(['lima-vm']) });

    assert.equal(result.length, 2);
    assert.ok(result.every(t => t.org === 'lima-vm'));
});

runner.test('applyFilters: filters by repo dynamic keyword', () => {
    const templates = [
        { name: 't1', repo: 'lima-vm/lima', org: 'lima-vm', keywords: [] },
        { name: 't2', repo: 'lima-vm/lima', org: 'lima-vm', keywords: [] },
        { name: 't3', repo: 'lima-vm/other', org: 'lima-vm', keywords: [] },
        { name: 't4', repo: 'user/repo', org: 'user', keywords: [] }
    ];
    const result = applyFilters(templates, { selectedKeywords: new Set(['lima-vm/lima']) });

    assert.equal(result.length, 2);
    assert.ok(result.every(t => t.repo === 'lima-vm/lima'));
});

runner.test('applyFilters: combines dynamic keyword with regular keyword', () => {
    const templates = [
        { name: 't1', repo: 'lima-vm/lima', org: 'lima-vm', keywords: ['docker'] },
        { name: 't2', repo: 'lima-vm/lima', org: 'lima-vm', keywords: ['k8s'] },
        { name: 't3', repo: 'user/repo', org: 'user', keywords: ['docker'] }
    ];
    const result = applyFilters(templates, {
        selectedKeywords: new Set(['lima-vm', 'docker'])
    });

    assert.equal(result.length, 1);
    assert.equal(result[0].name, 't1'); // Only t1 matches both lima-vm AND docker
});

runner.test('applyFilters: handles multiple dynamic keywords', () => {
    const templates = [
        { name: 't1', repo: 'lima-vm/lima', org: 'lima-vm', keywords: [] },
        { name: 't2', repo: 'lima-vm/other', org: 'lima-vm', keywords: [] },
        { name: 't3', repo: 'user/repo', org: 'user', keywords: [] }
    ];
    const result = applyFilters(templates, {
        selectedKeywords: new Set(['lima-vm', 'lima-vm/lima'])
    });

    // AND logic: must match BOTH lima-vm AND lima-vm/lima
    assert.equal(result.length, 1);
    assert.equal(result[0].repo, 'lima-vm/lima');
});
