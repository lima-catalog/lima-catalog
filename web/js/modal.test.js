/**
 * Unit Tests: Modal and URL Management (modal.js)
 *
 * High-level overview of what's being tested:
 * - URL parameter parsing (filters, keywords, categories, sort, etc.)
 * - URL encoding/decoding of special characters in keywords
 * - Updating URL with filter state changes
 * - Using pushState vs replaceState for history management
 * - Updating URL when template is selected/deselected
 * - Opening template modal from URL parameter
 * - Modal event listener setup
 *
 * - Similarity badge generation:
 *   - "Original" badge for 100% similarity with isOriginal=true
 *   - "Exact" badge for 100% duplicates
 *   - "Near" badge for 90-99% similarity
 *   - "Similar" badge for <90% similarity
 *
 * - Diff generation:
 *   - Computing longest common subsequence (LCS) for diff algorithm
 *   - Generating unified diff format with +/- prefixes
 *   - Counting additions and deletions
 *   - Handling empty/identical texts
 *   - Hunk headers with line numbers
 *
 * - XSS prevention:
 *   - Escaping HTML special characters in user input
 *   - Preventing script injection in modal content
 */

import { runner, assert } from './test-framework.js';
import {
    getFiltersFromURL,
    updateURLWithFilters,
    updateURLForTemplateSelection,
    openPreviewModal,
    closePreviewModal,
    openTemplateFromURL,
    setupModalEventListeners,
    getSimilarityBadge,
    generateUnifiedDiff,
    computeLCS,
    escapeHtml
} from './modal.js';
import * as State from './state.js';

/**
 * Helper to create a mock URL with query parameters
 */
function setMockURL(search) {
    const baseURL = 'http://localhost:3000';
    const fullURL = search ? `${baseURL}${search}` : baseURL;
    const url = new URL(fullURL);

    // Create a proper location object that can be passed to new URL()
    global.window.location = {
        href: url.href,
        origin: url.origin,
        protocol: url.protocol,
        host: url.host,
        hostname: url.hostname,
        port: url.port,
        pathname: url.pathname,
        search: url.search,
        hash: url.hash,
        toString: () => url.href
    };
}

/**
 * Helper to get current URL search params
 */
function getCurrentURLParams() {
    return new URLSearchParams(global.window.location.search);
}

/**
 * Reset URL to clean state
 */
function resetURL() {
    setMockURL('');
}

// =============================================================================
// URL HANDLING TESTS
// =============================================================================

runner.test('modal.js: getFiltersFromURL returns defaults with empty URL', () => {
    resetURL();
    const filters = getFiltersFromURL();

    assert.equal(filters.search, '');
    assert.deepEqual(filters.keywords, []);
    assert.equal(filters.category, null);
    assert.equal(filters.official, true);
    assert.equal(filters.community, true);
    assert.equal(filters.duplicates, false);
    assert.equal(filters.similars, false);
    assert.equal(filters.sort, 'name');
});

runner.test('modal.js: getFiltersFromURL parses search parameter', () => {
    setMockURL('?search=docker');
    const filters = getFiltersFromURL();

    assert.equal(filters.search, 'docker');
});

runner.test('modal.js: getFiltersFromURL parses single keyword', () => {
    setMockURL('?keywords=linux');
    const filters = getFiltersFromURL();

    assert.deepEqual(filters.keywords, ['linux']);
});

runner.test('modal.js: getFiltersFromURL parses multiple keywords', () => {
    setMockURL('?keywords=linux,docker,kubernetes');
    const filters = getFiltersFromURL();

    assert.deepEqual(filters.keywords, ['linux', 'docker', 'kubernetes']);
});

runner.test('modal.js: getFiltersFromURL decodes URL-encoded keywords', () => {
    setMockURL('?keywords=lima-vm%2Flima,org%3Alima-vm');
    const filters = getFiltersFromURL();

    assert.deepEqual(filters.keywords, ['lima-vm/lima', 'org:lima-vm']);
});

runner.test('modal.js: getFiltersFromURL parses category', () => {
    setMockURL('?category=container');
    const filters = getFiltersFromURL();

    assert.equal(filters.category, 'container');
});

runner.test('modal.js: getFiltersFromURL parses official filter', () => {
    setMockURL('?official=false');
    const filters = getFiltersFromURL();

    assert.equal(filters.official, false);
    assert.equal(filters.community, true); // default
});

runner.test('modal.js: getFiltersFromURL parses community filter', () => {
    setMockURL('?community=false');
    const filters = getFiltersFromURL();

    assert.equal(filters.official, true); // default
    assert.equal(filters.community, false);
});

runner.test('modal.js: getFiltersFromURL parses duplicates filter', () => {
    setMockURL('?duplicates=true');
    const filters = getFiltersFromURL();

    assert.equal(filters.duplicates, true);
});

runner.test('modal.js: getFiltersFromURL parses similars filter', () => {
    setMockURL('?similars=true');
    const filters = getFiltersFromURL();

    assert.equal(filters.similars, true);
});

runner.test('modal.js: getFiltersFromURL parses sort parameter', () => {
    setMockURL('?sort=stars');
    const filters = getFiltersFromURL();

    assert.equal(filters.sort, 'stars');
});

runner.test('modal.js: getFiltersFromURL parses complex query string', () => {
    setMockURL('?search=ubuntu&keywords=linux,docker&category=os&sort=stars&official=true&duplicates=true');
    const filters = getFiltersFromURL();

    assert.equal(filters.search, 'ubuntu');
    assert.deepEqual(filters.keywords, ['linux', 'docker']);
    assert.equal(filters.category, 'os');
    assert.equal(filters.sort, 'stars');
    assert.equal(filters.official, true);
    assert.equal(filters.duplicates, true);
});

runner.test('modal.js: getFiltersFromURL handles empty keywords parameter', () => {
    setMockURL('?keywords=');
    const filters = getFiltersFromURL();

    assert.deepEqual(filters.keywords, []);
});

runner.test('modal.js: getFiltersFromURL filters out empty keyword values', () => {
    setMockURL('?keywords=linux,,docker');
    const filters = getFiltersFromURL();

    assert.deepEqual(filters.keywords, ['linux', 'docker']);
});

runner.test('modal.js: updateURLWithFilters sets search parameter', () => {
    resetURL();

    // Mock window.history
    let pushedURL = null;
    global.window.history = {
        pushState: (state, title, url) => {
            pushedURL = url;
        }
    };

    updateURLWithFilters({ search: 'docker' });

    assert.ok(pushedURL);
    assert.ok(pushedURL.toString().includes('search=docker'));
});

runner.test('modal.js: updateURLWithFilters sets multiple keywords', () => {
    resetURL();

    let pushedURL = null;
    global.window.history = {
        pushState: (state, title, url) => {
            pushedURL = url;
        }
    };

    updateURLWithFilters({ keywords: ['linux', 'docker', 'kubernetes'] });

    assert.ok(pushedURL);
    assert.ok(pushedURL.toString().includes('keywords=linux%2Cdocker%2Ckubernetes'));
});

runner.test('modal.js: updateURLWithFilters sets category parameter', () => {
    resetURL();

    let pushedURL = null;
    global.window.history = {
        pushState: (state, title, url) => {
            pushedURL = url;
        }
    };

    updateURLWithFilters({ category: 'container' });

    assert.ok(pushedURL);
    assert.ok(pushedURL.toString().includes('category=container'));
});

runner.test('modal.js: updateURLWithFilters deletes empty search', () => {
    setMockURL('?search=docker');

    let pushedURL = null;
    global.window.history = {
        pushState: (state, title, url) => {
            pushedURL = url;
        }
    };

    updateURLWithFilters({ search: '' });

    assert.ok(pushedURL);
    assert.ok(!pushedURL.toString().includes('search='));
});

runner.test('modal.js: updateURLWithFilters deletes empty keywords', () => {
    setMockURL('?keywords=linux,docker');

    let pushedURL = null;
    global.window.history = {
        pushState: (state, title, url) => {
            pushedURL = url;
        }
    };

    updateURLWithFilters({ keywords: [] });

    assert.ok(pushedURL);
    assert.ok(!pushedURL.toString().includes('keywords='));
});

runner.test('modal.js: updateURLWithFilters deletes null category', () => {
    setMockURL('?category=container');

    let pushedURL = null;
    global.window.history = {
        pushState: (state, title, url) => {
            pushedURL = url;
        }
    };

    updateURLWithFilters({ category: null });

    assert.ok(pushedURL);
    assert.ok(!pushedURL.toString().includes('category='));
});

runner.test('modal.js: updateURLWithFilters includes official=false when not both true', () => {
    resetURL();

    let pushedURL = null;
    global.window.history = {
        pushState: (state, title, url) => {
            pushedURL = url;
        }
    };

    updateURLWithFilters({ official: false, community: true });

    assert.ok(pushedURL);
    assert.ok(pushedURL.toString().includes('official=false'));
    assert.ok(pushedURL.toString().includes('community=true'));
});

runner.test('modal.js: updateURLWithFilters excludes type filters when both true', () => {
    setMockURL('?official=false&community=true');

    let pushedURL = null;
    global.window.history = {
        pushState: (state, title, url) => {
            pushedURL = url;
        }
    };

    updateURLWithFilters({ official: true, community: true });

    assert.ok(pushedURL);
    assert.ok(!pushedURL.toString().includes('official='));
    assert.ok(!pushedURL.toString().includes('community='));
});

runner.test('modal.js: updateURLWithFilters includes duplicates=true', () => {
    resetURL();

    let pushedURL = null;
    global.window.history = {
        pushState: (state, title, url) => {
            pushedURL = url;
        }
    };

    updateURLWithFilters({ duplicates: true });

    assert.ok(pushedURL);
    assert.ok(pushedURL.toString().includes('duplicates=true'));
});

runner.test('modal.js: updateURLWithFilters excludes duplicates when false', () => {
    setMockURL('?duplicates=true');

    let pushedURL = null;
    global.window.history = {
        pushState: (state, title, url) => {
            pushedURL = url;
        }
    };

    updateURLWithFilters({ duplicates: false });

    assert.ok(pushedURL);
    assert.ok(!pushedURL.toString().includes('duplicates='));
});

runner.test('modal.js: updateURLWithFilters includes sort when not default', () => {
    resetURL();

    let pushedURL = null;
    global.window.history = {
        pushState: (state, title, url) => {
            pushedURL = url;
        }
    };

    updateURLWithFilters({ sort: 'stars' });

    assert.ok(pushedURL);
    assert.ok(pushedURL.toString().includes('sort=stars'));
});

runner.test('modal.js: updateURLWithFilters excludes sort when default (name)', () => {
    setMockURL('?sort=stars');

    let pushedURL = null;
    global.window.history = {
        pushState: (state, title, url) => {
            pushedURL = url;
        }
    };

    updateURLWithFilters({ sort: 'name' });

    assert.ok(pushedURL);
    assert.ok(!pushedURL.toString().includes('sort='));
});

runner.test('modal.js: updateURLWithFilters uses replaceState when replace=true', () => {
    resetURL();

    let usedReplaceState = false;
    global.window.history = {
        replaceState: (state, title, url) => {
            usedReplaceState = true;
        },
        pushState: (state, title, url) => {
            usedReplaceState = false;
        }
    };

    updateURLWithFilters({ search: 'docker' }, true);

    assert.ok(usedReplaceState, 'Should use replaceState when replace=true');
});

runner.test('modal.js: updateURLWithFilters uses pushState by default', () => {
    resetURL();

    let usedPushState = false;
    global.window.history = {
        replaceState: (state, title, url) => {
            usedPushState = false;
        },
        pushState: (state, title, url) => {
            usedPushState = true;
        }
    };

    updateURLWithFilters({ search: 'docker' });

    assert.ok(usedPushState, 'Should use pushState by default');
});

runner.test('modal.js: updateURLWithFilters URL-encodes special characters in keywords', () => {
    resetURL();

    let pushedURL = null;
    global.window.history = {
        pushState: (state, title, url) => {
            pushedURL = url;
        }
    };

    updateURLWithFilters({ keywords: ['lima-vm/lima', 'org:lima-vm'] });

    assert.ok(pushedURL);
    const urlStr = pushedURL.toString();
    assert.ok(urlStr.includes('keywords='), 'Should have keywords parameter');

    // Parse the URL back to verify encoding worked
    const params = new URLSearchParams(pushedURL.search);
    const keywords = params.get('keywords').split(',').map(k => decodeURIComponent(k));
    assert.deepEqual(keywords, ['lima-vm/lima', 'org:lima-vm'], 'Keywords should round-trip correctly');
});

// =============================================================================
// SIMILARITY BADGE TESTS
// =============================================================================

runner.test('modal.js: getSimilarityBadge returns original badge for 100% with isOriginal=true', () => {
    const badge = getSimilarityBadge(100, true);
    assert.ok(badge.includes('original'));
    assert.ok(badge.includes('duplicate-badge'));
});

runner.test('modal.js: getSimilarityBadge returns exact badge for 100% with isOriginal=false', () => {
    const badge = getSimilarityBadge(100, false);
    assert.ok(badge.includes('exact'));
    assert.ok(!badge.includes('original'));
});

runner.test('modal.js: getSimilarityBadge returns near badge for 90-99%', () => {
    const badge90 = getSimilarityBadge(90, false);
    const badge95 = getSimilarityBadge(95, false);
    const badge99 = getSimilarityBadge(99, false);

    assert.ok(badge90.includes('near'));
    assert.ok(badge95.includes('near'));
    assert.ok(badge99.includes('near'));
});

runner.test('modal.js: getSimilarityBadge returns similar badge for <90%', () => {
    const badge50 = getSimilarityBadge(50, false);
    const badge89 = getSimilarityBadge(89, false);

    assert.ok(badge50.includes('similar'));
    assert.ok(badge89.includes('similar'));
    assert.ok(!badge89.includes('near'));
});

// =============================================================================
// DIFF GENERATION TESTS
// =============================================================================

runner.test('modal.js: computeLCS finds longest common subsequence for identical arrays', () => {
    const a = ['line1', 'line2', 'line3'];
    const b = ['line1', 'line2', 'line3'];
    const lcs = computeLCS(a, b);

    assert.equal(lcs.length, 3);
    assert.equal(lcs[0].aIndex, 0);
    assert.equal(lcs[0].bIndex, 0);
    assert.equal(lcs[1].aIndex, 1);
    assert.equal(lcs[1].bIndex, 1);
    assert.equal(lcs[2].aIndex, 2);
    assert.equal(lcs[2].bIndex, 2);
});

runner.test('modal.js: computeLCS finds LCS for arrays with additions', () => {
    const a = ['line1', 'line3'];
    const b = ['line1', 'line2', 'line3'];
    const lcs = computeLCS(a, b);

    assert.equal(lcs.length, 2);
    assert.equal(lcs[0].aIndex, 0); // line1
    assert.equal(lcs[0].bIndex, 0);
    assert.equal(lcs[1].aIndex, 1); // line3
    assert.equal(lcs[1].bIndex, 2);
});

runner.test('modal.js: computeLCS finds LCS for arrays with deletions', () => {
    const a = ['line1', 'line2', 'line3'];
    const b = ['line1', 'line3'];
    const lcs = computeLCS(a, b);

    assert.equal(lcs.length, 2);
    assert.equal(lcs[0].aIndex, 0); // line1
    assert.equal(lcs[0].bIndex, 0);
    assert.equal(lcs[1].aIndex, 2); // line3
    assert.equal(lcs[1].bIndex, 1);
});

runner.test('modal.js: computeLCS finds LCS for completely different arrays', () => {
    const a = ['line1', 'line2'];
    const b = ['line3', 'line4'];
    const lcs = computeLCS(a, b);

    assert.equal(lcs.length, 0);
});

runner.test('modal.js: computeLCS handles empty arrays', () => {
    const lcs1 = computeLCS([], ['line1']);
    const lcs2 = computeLCS(['line1'], []);
    const lcs3 = computeLCS([], []);

    assert.equal(lcs1.length, 0);
    assert.equal(lcs2.length, 0);
    assert.equal(lcs3.length, 0);
});

runner.test('modal.js: generateUnifiedDiff returns no differences for identical text', () => {
    const text = 'line1\nline2\nline3';
    const diff = generateUnifiedDiff(text, text);

    assert.equal(diff.text, '# No differences found');
    assert.equal(diff.additions, 0);
    assert.equal(diff.deletions, 0);
});

runner.test('modal.js: generateUnifiedDiff shows additions with + prefix', () => {
    const original = 'line1\nline3';
    const modified = 'line1\nline2\nline3';
    const diff = generateUnifiedDiff(original, modified);

    assert.ok(diff.text.includes('+line2'));
    assert.equal(diff.additions, 1);
    assert.equal(diff.deletions, 0);
});

runner.test('modal.js: generateUnifiedDiff shows deletions with - prefix', () => {
    const original = 'line1\nline2\nline3';
    const modified = 'line1\nline3';
    const diff = generateUnifiedDiff(original, modified);

    assert.ok(diff.text.includes('-line2'));
    assert.equal(diff.additions, 0);
    assert.equal(diff.deletions, 1);
});

runner.test('modal.js: generateUnifiedDiff shows modifications', () => {
    const original = 'line1\nline2\nline3';
    const modified = 'line1\nline2-modified\nline3';
    const diff = generateUnifiedDiff(original, modified);

    assert.ok(diff.text.includes('-line2'));
    assert.ok(diff.text.includes('+line2-modified'));
    assert.equal(diff.additions, 1);
    assert.equal(diff.deletions, 1);
});

runner.test('modal.js: generateUnifiedDiff includes file names in header', () => {
    const original = 'line1';
    const modified = 'line2';
    const diff = generateUnifiedDiff(original, modified, 'original.txt', 'modified.txt');

    assert.ok(diff.text.includes('--- original.txt'));
    assert.ok(diff.text.includes('+++ modified.txt'));
});

runner.test('modal.js: generateUnifiedDiff uses default file names', () => {
    const original = 'line1';
    const modified = 'line2';
    const diff = generateUnifiedDiff(original, modified);

    assert.ok(diff.text.includes('--- original'));
    assert.ok(diff.text.includes('+++ similar'));
});

runner.test('modal.js: generateUnifiedDiff includes hunk headers', () => {
    const original = 'line1\nline2\nline3';
    const modified = 'line1\nline2-modified\nline3';
    const diff = generateUnifiedDiff(original, modified);

    // Should have hunk header with @@ format
    assert.ok(diff.text.includes('@@'));
});

runner.test('modal.js: generateUnifiedDiff handles empty original', () => {
    const original = '';
    const modified = 'line1\nline2';
    const diff = generateUnifiedDiff(original, modified);

    assert.ok(diff.text.includes('+line1'));
    assert.ok(diff.text.includes('+line2'));
    assert.equal(diff.additions, 2);
    // Empty string splits to [''] which counts as 1 line to delete
    assert.equal(diff.deletions, 1);
});

runner.test('modal.js: generateUnifiedDiff handles empty modified', () => {
    const original = 'line1\nline2';
    const modified = '';
    const diff = generateUnifiedDiff(original, modified);

    assert.ok(diff.text.includes('-line1'));
    assert.ok(diff.text.includes('-line2'));
    // Empty string splits to [''] which counts as 1 line to add
    assert.equal(diff.additions, 1);
    assert.equal(diff.deletions, 2);
});

runner.test('modal.js: generateUnifiedDiff counts stats correctly with multiple changes', () => {
    const original = 'line1\nline2\nline3\nline4\nline5';
    const modified = 'line1\nline2-mod\nline3\nnewline\nline5';
    const diff = generateUnifiedDiff(original, modified);

    // Should have 2 additions (line2-mod, newline) and 2 deletions (line2, line4)
    assert.equal(diff.additions, 2);
    assert.equal(diff.deletions, 2);
});

// =============================================================================
// HTML ESCAPING TESTS (XSS Prevention)
// =============================================================================

// Note: escapeHtml is already thoroughly tested in templateCard.test.js
// These tests verify modal.js uses the same implementation correctly

runner.test('modal.js: escapeHtml prevents XSS with script tags', () => {
    const malicious = '<script>alert("XSS")</script>';
    const escaped = escapeHtml(malicious);

    // Should not contain raw tags
    assert.ok(!escaped.includes('<script>'), 'Should not contain unescaped <script> tag');
    assert.ok(!escaped.includes('</script>'), 'Should not contain unescaped </script> tag');
});

runner.test('modal.js: escapeHtml handles empty/null values', () => {
    assert.equal(escapeHtml(''), '');
    assert.equal(escapeHtml(null), '');
    assert.equal(escapeHtml(undefined), '');
});
