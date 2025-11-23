/**
 * Preview modal functionality
 */

import { getDefaultBranchURL, getRawContentURL } from './urlHelpers.js';
import { deriveDisplayName } from './templateCard.js';
import { trapFocus } from './utils.js';
import * as State from './state.js';
import { getTemplates, getFilteredTemplates, getSelectedKeywords, getSelectedCategory } from './state.js';
import { applyFilters } from './filters.js';

// Modal state
let currentTemplate = null;
let currentYamlContent = null; // Store original YAML for diff comparison
let releaseFocusTrap = null;
let previouslyFocusedElement = null;
let isHandlingPopState = false; // Flag to prevent duplicate popstate handling
let onYamlLoadedCallback = null; // Callback for when YAML content is loaded
let isDebugMode = false; // Debug mode for showing template object as JSON

/**
 * Get similarity badge HTML based on similarity percentage
 * Thresholds: 100% = exact/original, 90-99% = near, <90% = similar
 * @param {number} similarityPercent - Similarity percentage (0-100)
 * @param {boolean} isOriginal - Whether this is the original template
 * @returns {string} Badge HTML
 */
export function getSimilarityBadge(similarityPercent, isOriginal) {
    if (similarityPercent === 100) {
        if (isOriginal) {
            return '<span class="duplicate-badge original">original</span>';
        }
        return '<span class="duplicate-badge exact">exact</span>';
    } else if (similarityPercent >= 90) {
        return '<span class="duplicate-badge near">near</span>';
    }
    return '<span class="duplicate-badge similar">similar</span>';
}

/**
 * Generate a unified diff between two texts
 * @param {string} originalText - Original text
 * @param {string} newText - New text to compare
 * @param {string} originalName - Name for original file
 * @param {string} newName - Name for new file
 * @returns {{text: string, additions: number, deletions: number}} Diff with stats
 */
export function generateUnifiedDiff(originalText, newText, originalName = 'original', newName = 'similar') {
    const originalLines = originalText.split('\n');
    const newLines = newText.split('\n');

    // Simple LCS-based diff algorithm
    const lcs = computeLCS(originalLines, newLines);
    return buildUnifiedDiff(originalLines, newLines, lcs, originalName, newName);
}

/**
 * Compute Longest Common Subsequence for diff
 * Exported for testing purposes
 */
export function computeLCS(a, b) {
    const m = a.length;
    const n = b.length;
    const dp = Array(m + 1).fill(null).map(() => Array(n + 1).fill(0));

    for (let i = 1; i <= m; i++) {
        for (let j = 1; j <= n; j++) {
            if (a[i - 1] === b[j - 1]) {
                dp[i][j] = dp[i - 1][j - 1] + 1;
            } else {
                dp[i][j] = Math.max(dp[i - 1][j], dp[i][j - 1]);
            }
        }
    }

    // Backtrack to find LCS indices
    const result = [];
    let i = m, j = n;
    while (i > 0 && j > 0) {
        if (a[i - 1] === b[j - 1]) {
            result.unshift({ aIndex: i - 1, bIndex: j - 1 });
            i--;
            j--;
        } else if (dp[i - 1][j] > dp[i][j - 1]) {
            i--;
        } else {
            j--;
        }
    }

    return result;
}

/**
 * Build unified diff output from LCS
 */
function buildUnifiedDiff(originalLines, newLines, lcs, originalName, newName) {
    const hunks = [];
    let currentHunk = null;
    let aPos = 0, bPos = 0;
    let lcsIndex = 0;

    const contextLines = 3;

    while (aPos < originalLines.length || bPos < newLines.length) {
        // Find next LCS match
        const nextMatch = lcs[lcsIndex];

        if (nextMatch && aPos === nextMatch.aIndex && bPos === nextMatch.bIndex) {
            // Matching line
            if (currentHunk) {
                currentHunk.lines.push({ type: ' ', text: originalLines[aPos] });
            }
            aPos++;
            bPos++;
            lcsIndex++;
        } else {
            // Difference found - start or continue a hunk
            if (!currentHunk) {
                // Start new hunk with context
                const contextStart = Math.max(0, aPos - contextLines);
                currentHunk = {
                    aStart: contextStart + 1,
                    bStart: Math.max(0, bPos - contextLines) + 1,
                    lines: []
                };
                // Add leading context
                for (let c = contextStart; c < aPos; c++) {
                    currentHunk.lines.push({ type: ' ', text: originalLines[c] });
                }
            }

            // Add removed lines from original
            const aEnd = nextMatch ? nextMatch.aIndex : originalLines.length;
            while (aPos < aEnd) {
                currentHunk.lines.push({ type: '-', text: originalLines[aPos] });
                aPos++;
            }

            // Add added lines from new
            const bEnd = nextMatch ? nextMatch.bIndex : newLines.length;
            while (bPos < bEnd) {
                currentHunk.lines.push({ type: '+', text: newLines[bPos] });
                bPos++;
            }
        }

        // Check if we should close the hunk (enough context after changes)
        if (currentHunk) {
            const lastChangeIndex = currentHunk.lines.length - 1 -
                currentHunk.lines.slice().reverse().findIndex(l => l.type !== ' ');
            const contextAfter = currentHunk.lines.length - 1 - lastChangeIndex;

            if (contextAfter >= contextLines || (aPos >= originalLines.length && bPos >= newLines.length)) {
                // Trim excess context
                while (currentHunk.lines.length > 0 &&
                       currentHunk.lines[currentHunk.lines.length - 1].type === ' ' &&
                       currentHunk.lines.length - 1 - lastChangeIndex > contextLines) {
                    currentHunk.lines.pop();
                }
                hunks.push(currentHunk);
                currentHunk = null;
            }
        }
    }

    if (currentHunk && currentHunk.lines.some(l => l.type !== ' ')) {
        hunks.push(currentHunk);
    }

    // Build output and count stats
    if (hunks.length === 0) {
        return { text: '# No differences found', additions: 0, deletions: 0 };
    }

    let output = `--- ${originalName}\n+++ ${newName}\n`;
    let additions = 0;
    let deletions = 0;

    for (const hunk of hunks) {
        const aCount = hunk.lines.filter(l => l.type !== '+').length;
        const bCount = hunk.lines.filter(l => l.type !== '-').length;
        output += `@@ -${hunk.aStart},${aCount} +${hunk.bStart},${bCount} @@\n`;

        for (const line of hunk.lines) {
            output += `${line.type}${line.text}\n`;
            if (line.type === '+') additions++;
            else if (line.type === '-') deletions++;
        }
    }

    return { text: output, additions, deletions };
}

/**
 * URL parameter utilities for deep linking
 */

/**
 * Get template ID from URL parameters
 * @returns {string|null} Template ID if present in URL
 */
export function getTemplateFromURL() {
    const params = new URLSearchParams(window.location.search);
    return params.get('template');
}

/**
 * Get filter state from URL parameters
 * @returns {Object} Filter state object
 */
export function getFiltersFromURL() {
    const params = new URLSearchParams(window.location.search);

    const filters = {
        search: params.get('search') || '',
        keywords: [],
        category: params.get('category') || null,
        official: true,
        community: true,
        duplicates: false,
        similars: false,
        sort: params.get('sort') || 'name'
    };

    // Parse keywords (comma-separated)
    const keywordsParam = params.get('keywords');
    if (keywordsParam) {
        filters.keywords = keywordsParam.split(',').map(k => decodeURIComponent(k)).filter(k => k);
    }

    // Parse boolean filters - only set if explicitly present in URL
    if (params.has('official')) {
        filters.official = params.get('official') !== 'false';
    }
    if (params.has('community')) {
        filters.community = params.get('community') !== 'false';
    }
    if (params.has('duplicates')) {
        filters.duplicates = params.get('duplicates') === 'true';
    }
    if (params.has('similars')) {
        filters.similars = params.get('similars') === 'true';
    }

    return filters;
}

/**
 * Update URL with current filter state
 * @param {Object} filterState - Current filter state
 * @param {boolean} replace - Use replaceState instead of pushState (default: false)
 */
export function updateURLWithFilters(filterState, replace = false) {
    const url = new URL(window.location);

    // Search term
    if (filterState.search) {
        url.searchParams.set('search', filterState.search);
    } else {
        url.searchParams.delete('search');
    }

    // Keywords (comma-separated)
    if (filterState.keywords && filterState.keywords.length > 0) {
        url.searchParams.set('keywords', filterState.keywords.map(k => encodeURIComponent(k)).join(','));
    } else {
        url.searchParams.delete('keywords');
    }

    // Category
    if (filterState.category) {
        url.searchParams.set('category', filterState.category);
    } else {
        url.searchParams.delete('category');
    }

    // Type filters - only include if not default (both true)
    if (!filterState.official || !filterState.community) {
        url.searchParams.set('official', String(filterState.official));
        url.searchParams.set('community', String(filterState.community));
    } else {
        url.searchParams.delete('official');
        url.searchParams.delete('community');
    }

    // Duplicates - only include if true (non-default)
    if (filterState.duplicates) {
        url.searchParams.set('duplicates', 'true');
    } else {
        url.searchParams.delete('duplicates');
    }

    // Similars - only include if true (non-default)
    if (filterState.similars) {
        url.searchParams.set('similars', 'true');
    } else {
        url.searchParams.delete('similars');
    }

    // Sort - only include if not default
    if (filterState.sort && filterState.sort !== 'name') {
        url.searchParams.set('sort', filterState.sort);
    } else {
        url.searchParams.delete('sort');
    }

    // Use replaceState for filter changes to avoid polluting history
    if (replace) {
        window.history.replaceState({ filters: filterState }, '', url);
    } else {
        window.history.pushState({ filters: filterState }, '', url);
    }
}

/**
 * Check if modal should be open based on URL parameters
 * @returns {boolean} True if modal parameter is set to 'open'
 */
function isModalOpenInURL() {
    const params = new URLSearchParams(window.location.search);
    return params.get('modal') === 'open';
}

/**
 * Update URL with template ID and optional modal state
 * @param {string} templateId - Template ID to add to URL
 * @param {boolean} modalOpen - Whether modal is open (default: false)
 */
function updateURLWithTemplate(templateId, modalOpen = false) {
    const url = new URL(window.location);
    url.searchParams.set('template', templateId);
    if (modalOpen) {
        url.searchParams.set('modal', 'open');
    } else {
        url.searchParams.delete('modal');
    }
    window.history.pushState({ templateId, modalOpen }, '', url);
}

/**
 * Update URL to set template selection (exposed for use by template cards)
 * @param {string} templateId - Template ID to add to URL
 */
export function updateURLForTemplateSelection(templateId) {
    // Only update if not already the current template in URL (avoid spam)
    const currentTemplateId = getTemplateFromURL();
    if (currentTemplateId !== templateId) {
        updateURLWithTemplate(templateId, false);
    }
}

/**
 * Clear template ID from URL
 */
function clearTemplateFromURL() {
    const url = new URL(window.location);
    url.searchParams.delete('template');
    url.searchParams.delete('modal');
    window.history.pushState({}, '', url);
}

/**
 * Close modal but keep template selected in URL
 */
function closeModalKeepTemplate() {
    const templateId = getTemplateFromURL();
    if (templateId) {
        // Keep template in URL, just remove modal=open
        updateURLWithTemplate(templateId, false);
    } else {
        // No template in URL, just clear everything
        clearTemplateFromURL();
    }
}

/**
 * Open preview modal for a template
 * @param {Object} template - Template object
 * @param {boolean} preserveDebugMode - Whether to preserve debug mode state (default: false)
 */
export function openPreviewModal(template, preserveDebugMode = false) {
    currentTemplate = template;

    // Reset debug mode when opening a new modal (unless navigating with Ctrl+Arrow)
    if (!preserveDebugMode) {
        isDebugMode = false;
    }

    // Store the currently focused element to restore later
    previouslyFocusedElement = document.activeElement;

    const modal = document.getElementById('preview-modal');
    const modalTitle = document.getElementById('modal-title');
    const modalLoading = document.getElementById('modal-loading');
    const modalCode = document.getElementById('modal-code');
    const modalGithubLink = document.getElementById('modal-github-link');
    const modalGithubScheme = document.getElementById('modal-github-scheme');
    const copyYamlButton = document.getElementById('copy-yaml');

    // Set title
    modalTitle.textContent = deriveDisplayName(template);

    // Set github: scheme URL (pre-generated by backend)
    modalGithubScheme.textContent = template.github_url;

    // Use default branch URL for display
    const displayURL = getDefaultBranchURL(template);
    modalGithubLink.href = displayURL;
    modalGithubLink.textContent = displayURL;

    // Show modal and loading state
    modal.style.display = 'flex';
    modalLoading.classList.remove('hidden');
    modalCode.classList.add('hidden');
    copyYamlButton.style.display = 'none';
    document.body.style.overflow = 'hidden';

    // Remove ready class to hide modal-content during loading
    const modalContent = modal.querySelector('.modal-content');
    modalContent.classList.remove('ready');

    // Populate similar templates section
    populateSimilarTemplates(template);

    // Trap focus within modal for accessibility
    setTimeout(() => {
        releaseFocusTrap = trapFocus(modal.querySelector('.modal-content'));
    }, 100);

    // Update URL with template ID and modal state for deep linking (only if not handling popstate)
    if (!isHandlingPopState) {
        updateURLWithTemplate(template.id, true);
    }

    // Fetch and display template content
    fetchTemplateContent(template);
}

/**
 * Close preview modal
 */
export function closePreviewModal() {
    const modal = document.getElementById('preview-modal');
    modal.style.display = 'none';
    document.body.style.overflow = 'auto';

    // Store the template before clearing it
    const closedTemplate = currentTemplate;
    currentTemplate = null;

    // Release focus trap
    if (releaseFocusTrap) {
        releaseFocusTrap();
        releaseFocusTrap = null;
    }

    // Keep template selected in URL, just remove modal=open (only if not handling popstate)
    if (!isHandlingPopState) {
        closeModalKeepTemplate();
    }

    // Focus the template card if there's a template in the URL
    const templateId = getTemplateFromURL();
    if (templateId) {
        // Use requestAnimationFrame to ensure modal is fully closed before focusing
        requestAnimationFrame(() => {
            focusTemplateCard(templateId);
        });
    } else if (previouslyFocusedElement && previouslyFocusedElement.focus) {
        // Otherwise restore focus to the element that opened the modal
        previouslyFocusedElement.focus();
        previouslyFocusedElement = null;
    }
}

/**
 * Fetch and display template content
 * @param {Object} template - Template object
 */
async function fetchTemplateContent(template) {
    const modalLoading = document.getElementById('modal-loading');
    const modalCode = document.getElementById('modal-code');
    const modalCodeContent = document.getElementById('modal-code-content');
    const copyYamlButton = document.getElementById('copy-yaml');

    try {
        // Use raw_url for fetching content (already has default branch)
        const rawURL = template.raw_url;

        const response = await fetch(rawURL);
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        const content = await response.text();

        // Store original YAML for diff comparison
        currentYamlContent = content;

        // Display either debug JSON or YAML based on debug mode
        if (isDebugMode) {
            showDebugJson();
        } else {
            modalCodeContent.textContent = content;
            modalCodeContent.className = 'language-yaml';
            modalCodeContent.removeAttribute('data-highlighted');
            hljs.highlightElement(modalCodeContent);
        }

        // Show code and copy button, hide loading
        modalLoading.classList.add('hidden');
        modalCode.classList.remove('hidden');
        copyYamlButton.style.display = 'block';

        // Reset scroll position to top when content is displayed
        modalCode.scrollTo(0, 0);

        // Show modal-content now that content is loaded (fade in)
        const modal = document.getElementById('preview-modal');
        const modalContent = modal.querySelector('.modal-content');
        modalContent.classList.add('ready');

        // Trigger diff stats fetching now that YAML is available
        if (onYamlLoadedCallback) {
            onYamlLoadedCallback();
            onYamlLoadedCallback = null;
        }
    } catch (error) {
        console.error('Error fetching template:', error);
        modalLoading.textContent = `Error loading template: ${error.message}`;
        copyYamlButton.style.display = 'none';

        // Show modal-content even with error (fade in)
        const modal = document.getElementById('preview-modal');
        const modalContent = modal.querySelector('.modal-content');
        modalContent.classList.add('ready');
    }
}

/**
 * Display the template object as pretty-printed JSON
 */
function showDebugJson() {
    if (!currentTemplate) return;

    const modalCodeContent = document.getElementById('modal-code-content');

    // Serialize template object as pretty-printed JSON
    const jsonContent = JSON.stringify(currentTemplate, null, 2);

    modalCodeContent.textContent = jsonContent;
    modalCodeContent.className = 'language-json';
    modalCodeContent.removeAttribute('data-highlighted');
    hljs.highlightElement(modalCodeContent);
}

/**
 * Restore YAML display from debug mode
 */
function showYamlContent() {
    if (!currentYamlContent) return;

    const modalCodeContent = document.getElementById('modal-code-content');

    modalCodeContent.textContent = currentYamlContent;
    modalCodeContent.className = 'language-yaml';
    modalCodeContent.removeAttribute('data-highlighted');
    hljs.highlightElement(modalCodeContent);
}

/**
 * Toggle debug mode
 */
function toggleDebugMode() {
    // Only toggle if we're not showing a diff (similar template comparison)
    const similarList = document.getElementById('similar-templates-list');
    const hasFocusOnSimilarList = similarList && similarList.contains(document.activeElement);

    if (hasFocusOnSimilarList) {
        // Don't toggle debug mode when user is viewing a diff
        return;
    }

    isDebugMode = !isDebugMode;

    if (isDebugMode) {
        showDebugJson();
    } else {
        showYamlContent();
    }
}

/**
 * Copy text to clipboard with visual feedback
 * @param {string} text - Text to copy
 * @param {HTMLElement} button - Button element for feedback
 */
async function copyToClipboard(text, button) {
    try {
        await navigator.clipboard.writeText(text);

        // Visual feedback
        const originalText = button.textContent;
        button.textContent = 'Copied!';
        button.classList.add('copied');

        setTimeout(() => {
            button.textContent = originalText;
            button.classList.remove('copied');
        }, 2000);
    } catch (err) {
        console.error('Failed to copy:', err);
        button.textContent = 'Failed';
        setTimeout(() => {
            button.textContent = 'Copy';
        }, 2000);
    }
}

/**
 * Populate similar templates section
 * @param {Object} template - Template object
 */
function populateSimilarTemplates(template) {
    const similarSection = document.getElementById('similar-templates-section');
    const similarList = document.getElementById('similar-templates-list');
    const modalCodeContent = document.getElementById('modal-code-content');
    const copyYamlButton = document.getElementById('copy-yaml');

    // Always clear previous state
    onYamlLoadedCallback = null;
    similarList.innerHTML = '';

    // Check if template has similar templates
    if (!template.similar_templates || template.similar_templates.length === 0) {
        similarSection.classList.add('hidden');
        return;
    }

    // Show section
    similarSection.classList.remove('hidden');

    // Get all templates for looking up data
    const allTemplates = getTemplates();
    const templateMap = new Map(allTemplates.map(t => [t.id, t]));

    // Track selected index for keyboard navigation
    let selectedIndex = -1;
    const items = [];
    const yamlCache = new Map(); // Cache fetched YAML content
    let isShowingDiff = false;

    // Make the list itself focusable as a listbox
    similarList.setAttribute('role', 'listbox');
    similarList.setAttribute('tabindex', '0');
    similarList.setAttribute('aria-label', 'Similar templates');

    // Restore original YAML display (or debug JSON if in debug mode)
    const restoreOriginalYaml = () => {
        if (currentYamlContent && isShowingDiff) {
            if (isDebugMode) {
                showDebugJson();
            } else {
                modalCodeContent.textContent = currentYamlContent;
                modalCodeContent.className = 'language-yaml';
                modalCodeContent.removeAttribute('data-highlighted');
                hljs.highlightElement(modalCodeContent);
            }
            copyYamlButton.style.display = 'block';
            isShowingDiff = false;
        }
    };

    // Show diff for selected similar template
    const showDiffForTemplate = async (similarTemplate) => {
        if (!similarTemplate || !currentYamlContent) return;

        // Hide copy button when showing diff
        copyYamlButton.style.display = 'none';

        // Check cache first
        let similarYaml = yamlCache.get(similarTemplate.id);
        if (!similarYaml) {
            try {
                const response = await fetch(similarTemplate.raw_url);
                if (response.ok) {
                    similarYaml = await response.text();
                    yamlCache.set(similarTemplate.id, similarYaml);
                } else {
                    modalCodeContent.textContent = `# Error loading ${similarTemplate.id}`;
                    isShowingDiff = true;
                    return;
                }
            } catch (error) {
                modalCodeContent.textContent = `# Error loading ${similarTemplate.id}: ${error.message}`;
                isShowingDiff = true;
                return;
            }
        }

        // Generate and display diff
        const diff = generateUnifiedDiff(
            currentYamlContent,
            similarYaml,
            currentTemplate.id,
            similarTemplate.id
        );

        modalCodeContent.textContent = diff.text;
        modalCodeContent.className = 'language-diff';
        modalCodeContent.removeAttribute('data-highlighted');
        hljs.highlightElement(modalCodeContent);
        isShowingDiff = true;
    };

    // Get current filter state from UI
    const showDuplicates = document.getElementById('show-duplicates')?.checked ?? false;
    const searchTerm = document.getElementById('search')?.value ?? '';
    const showOfficial = document.getElementById('show-official')?.checked ?? true;
    const showCommunity = document.getElementById('show-community')?.checked ?? true;
    const typeFilter = showOfficial && showCommunity ? '' : (showOfficial ? 'official' : 'community');

    // Get full template objects for similar templates
    const similarTemplateObjects = template.similar_templates
        .map(similar => templateMap.get(similar.id))
        .filter(t => t != null);

    // Apply the same filters used for the main template list
    const filteredTemplateObjects = applyFilters(similarTemplateObjects, {
        searchTerm,
        typeFilter,
        selectedCategory: getSelectedCategory(),
        selectedKeywords: getSelectedKeywords(),
        showDuplicates
    });

    // Create a set of filtered IDs for quick lookup
    const filteredIds = new Set(filteredTemplateObjects.map(t => t.id));

    // Filter similar_templates to only include those that passed the filter
    // Also filter out 100% matches when showDuplicates is false
    const filteredSimilarTemplates = template.similar_templates.filter(similar => {
        if (!filteredIds.has(similar.id)) return false;
        // Additional check: hide 100% matches when duplicates unchecked
        // (applyFilters checks original_id, but we also want to hide by similarity)
        if (!showDuplicates && Math.round(similar.similarity * 100) === 100) {
            return false;
        }
        return true;
    });

    // If no similar templates match filters, hide section
    if (filteredSimilarTemplates.length === 0) {
        similarSection.classList.add('hidden');
        return;
    }

    // Sort by similarity (descending), then originals first, then by id (ascending)
    const sortedSimilarTemplates = [...filteredSimilarTemplates].sort((a, b) => {
        if (b.similarity !== a.similarity) {
            return b.similarity - a.similarity;
        }
        // Within same similarity, originals come first
        if (a.is_original !== b.is_original) {
            return a.is_original ? -1 : 1;
        }
        return a.id.localeCompare(b.id);
    });

    // Update title with count
    const similarTitle = document.querySelector('.similar-templates-title');
    if (similarTitle) {
        similarTitle.textContent = `Similar Templates (${sortedSimilarTemplates.length})`;
    }

    // Hide scrollbar if 4 or fewer items (nothing to scroll)
    // Max height shows 4 items, so only need scrollbar for 5+
    if (sortedSimilarTemplates.length <= 4) {
        similarList.classList.add('no-scroll');
    } else {
        similarList.classList.remove('no-scroll');
    }

    sortedSimilarTemplates.forEach((similar, index) => {
        const similarTemplate = templateMap.get(similar.id);
        const similarityPercent = Math.round(similar.similarity * 100);

        const item = document.createElement('div');
        item.className = 'similar-template-item';
        item.setAttribute('role', 'option');
        item.setAttribute('aria-selected', 'false');
        item.dataset.index = index;

        // Single line format: ORG/REPO/TEMPLATEPATH [stats] [badge] percent
        // Badge derived purely from similarity score
        const badgeHtml = getSimilarityBadge(similarityPercent, similar.is_original);
        item.innerHTML = `
            <span class="similar-template-path">${escapeHtml(similar.id)}</span>
            <span class="item-diff-stats"></span>
            ${badgeHtml}
            <span class="similarity-percentage">${similarityPercent}%</span>
        `;

        // Single click: select and show diff
        item.addEventListener('click', () => {
            updateSelection(index);
        });

        // Double click: switch to this template
        item.addEventListener('dblclick', () => {
            if (similarTemplate) {
                closePreviewModal();
                setTimeout(() => openPreviewModal(similarTemplate), 100);
            }
        });

        // Hover tooltip to explain interaction
        let hoverTimeout = null;
        let tooltip = null;
        let lastMouseX = 0;
        let lastMouseY = 0;

        const updateTooltipPosition = () => {
            if (tooltip) {
                tooltip.style.left = `${lastMouseX + 15}px`;
                tooltip.style.top = `${lastMouseY + 15}px`;
            }
        };

        const showTooltip = () => {
            hoverTimeout = setTimeout(() => {
                tooltip = document.createElement('div');
                tooltip.className = 'similar-template-tooltip';
                tooltip.textContent = 'Click to compare • Double-click to open';
                document.body.appendChild(tooltip);

                tooltip.style.position = 'fixed';
                updateTooltipPosition();
            }, 800); // 800ms delay
        };

        const hideTooltip = () => {
            if (hoverTimeout) {
                clearTimeout(hoverTimeout);
                hoverTimeout = null;
            }
            if (tooltip) {
                tooltip.remove();
                tooltip = null;
            }
        };

        const trackMouse = (e) => {
            lastMouseX = e.clientX;
            lastMouseY = e.clientY;
            updateTooltipPosition();
        };

        item.addEventListener('mouseenter', showTooltip);
        item.addEventListener('mouseleave', hideTooltip);
        item.addEventListener('mousemove', trackMouse);

        items.push({ element: item, template: similarTemplate, id: similar.id });
        similarList.appendChild(item);
    });

    // Fetch diff stats for all similar templates in parallel
    const fetchDiffStats = async (item, similarTemplate) => {
        if (!similarTemplate || !currentYamlContent) return;

        let similarYaml = yamlCache.get(similarTemplate.id);
        if (!similarYaml) {
            try {
                const response = await fetch(similarTemplate.raw_url);
                if (response.ok) {
                    similarYaml = await response.text();
                    yamlCache.set(similarTemplate.id, similarYaml);
                } else {
                    return;
                }
            } catch {
                return;
            }
        }

        const diff = generateUnifiedDiff(currentYamlContent, similarYaml, '', '');
        const statsEl = item.element.querySelector('.item-diff-stats');
        if (statsEl && (diff.additions > 0 || diff.deletions > 0)) {
            statsEl.innerHTML = `<span class="diff-additions">+${diff.additions}</span><span class="diff-deletions">-${diff.deletions}</span>`;
        }
    };

    // Set callback to fetch diff stats once YAML is loaded
    onYamlLoadedCallback = () => {
        items.forEach(item => fetchDiffStats(item, item.template));
    };

    // Update selection styling and show diff
    const updateSelection = async (newIndex) => {
        if (selectedIndex >= 0 && selectedIndex < items.length) {
            items[selectedIndex].element.classList.remove('selected');
            items[selectedIndex].element.setAttribute('aria-selected', 'false');
        }
        selectedIndex = newIndex;
        if (selectedIndex >= 0 && selectedIndex < items.length) {
            items[selectedIndex].element.classList.add('selected');
            items[selectedIndex].element.setAttribute('aria-selected', 'true');
            // Scroll into view if needed
            items[selectedIndex].element.scrollIntoView({ block: 'nearest' });
            // Show diff for selected template
            await showDiffForTemplate(items[selectedIndex].template);
        }
    };

    // Keyboard navigation on the list
    similarList.addEventListener('keydown', (e) => {
        if (e.key === 'ArrowDown') {
            e.preventDefault();
            e.stopPropagation();
            // Stop at the end instead of wrapping to the beginning
            const newIndex = Math.min(selectedIndex + 1, items.length - 1);
            updateSelection(newIndex);
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            e.stopPropagation();
            // Stop at the beginning instead of wrapping to the end
            const newIndex = Math.max(selectedIndex - 1, 0);
            updateSelection(newIndex);
        } else if ((e.key === 'Enter' || e.key === ' ') && selectedIndex >= 0) {
            e.preventDefault();
            e.stopPropagation();
            const selected = items[selectedIndex];
            if (selected && selected.template) {
                closePreviewModal();
                setTimeout(() => openPreviewModal(selected.template), 100);
            }
        } else if (e.key === 'Tab' && !e.shiftKey) {
            // Tab forward: restore YAML and focus copy button
            e.preventDefault();
            e.stopPropagation();
            if (selectedIndex >= 0) {
                items[selectedIndex].element.classList.remove('selected');
                items[selectedIndex].element.setAttribute('aria-selected', 'false');
            }
            selectedIndex = -1;
            restoreOriginalYaml();
            copyYamlButton.focus();
        }
    });

    // Select first item when list gets focus and show diff
    similarList.addEventListener('focus', () => {
        if (selectedIndex < 0 && items.length > 0) {
            updateSelection(0);
        } else if (selectedIndex >= 0) {
            // Re-show diff if already had selection
            showDiffForTemplate(items[selectedIndex].template);
        }
    });

    // Restore original YAML when list loses focus
    similarList.addEventListener('blur', () => {
        if (selectedIndex >= 0) {
            items[selectedIndex].element.classList.remove('selected');
            items[selectedIndex].element.setAttribute('aria-selected', 'false');
        }
        selectedIndex = -1;
        restoreOriginalYaml();
    });
}

/**
 * Escape HTML to prevent XSS
 * @param {string} str - String to escape
 * @returns {string} Escaped string
 */
export function escapeHtml(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

/**
 * Focus a template card by template ID
 * @param {string} templateId - Template ID to focus
 * @param {number} retryCount - Number of retries (for timing)
 */
function focusTemplateCard(templateId, retryCount = 0) {
    const card = document.querySelector(`.template-card[data-template-id="${CSS.escape(templateId)}"]`);
    if (card) {
        card.focus();
        card.scrollIntoView({ behavior: 'smooth', block: 'center' });
        return true;
    } else if (retryCount < 5) {
        // Card not found yet, retry after a short delay (DOM might still be rendering)
        setTimeout(() => focusTemplateCard(templateId, retryCount + 1), 50);
        return false;
    } else {
        console.warn(`Template card not found in DOM: ${templateId}`);
        return false;
    }
}

/**
 * Open template from URL if template parameter exists
 * Called on page load and popstate events
 */
export function openTemplateFromURL() {
    const templateId = getTemplateFromURL();
    const shouldOpenModal = isModalOpenInURL();

    if (templateId) {
        // Find template by ID
        const templates = getTemplates();
        const template = templates.find(t => t.id === templateId);

        if (template) {
            if (shouldOpenModal) {
                // Open modal if modal=open in URL
                openPreviewModal(template);
            } else {
                // Just set focus on the template card without opening modal
                // Use requestAnimationFrame to ensure DOM is ready
                requestAnimationFrame(() => {
                    focusTemplateCard(templateId);
                });
            }
        } else {
            console.warn(`Template not found: ${templateId}`);
            // Clear invalid template from URL
            clearTemplateFromURL();
        }
    }
}

/**
 * Handle browser back/forward navigation
 */
async function handlePopState() {
    isHandlingPopState = true;

    // Import appActions dynamically to avoid circular dependency
    const { applyFiltersFromURL, filterAndRender, setHandlingPopState } = await import('./appActions.js');
    setHandlingPopState(true);

    // Restore filter state from URL
    applyFiltersFromURL();
    filterAndRender();

    const templateId = getTemplateFromURL();
    const shouldOpenModal = isModalOpenInURL();

    if (templateId) {
        if (shouldOpenModal) {
            // URL has template with modal=open - open modal if not already open
            if (!currentTemplate || currentTemplate.id !== templateId) {
                openTemplateFromURL();
            }
        } else {
            // URL has template without modal=open - close modal if open, focus card
            if (currentTemplate) {
                closePreviewModal();
            }
            // Focus the template card
            requestAnimationFrame(() => {
                focusTemplateCard(templateId);
            });
        }
    } else {
        // URL has no template - close modal if open and clear focus
        if (currentTemplate) {
            closePreviewModal();
        }
    }

    // Reset flags after a short delay
    setTimeout(() => {
        isHandlingPopState = false;
        setHandlingPopState(false);
    }, 100);
}

/**
 * Setup modal event listeners
 */
export function setupModalEventListeners() {
    const modal = document.getElementById('preview-modal');
    const modalOverlay = modal.querySelector('.modal-overlay');
    const modalCloseButton = document.getElementById('modal-close-button');
    const copyGithubUrlButton = document.getElementById('copy-github-url');
    const copyYamlButton = document.getElementById('copy-yaml');

    // Close on overlay click
    modalOverlay.addEventListener('click', closePreviewModal);

    // Close on Close button click
    modalCloseButton.addEventListener('click', closePreviewModal);

    // Copy github: URL to clipboard
    copyGithubUrlButton.addEventListener('click', async () => {
        const githubSchemeURL = document.getElementById('modal-github-scheme').textContent;
        await copyToClipboard(githubSchemeURL, copyGithubUrlButton);
    });

    // Copy YAML template to clipboard
    copyYamlButton.addEventListener('click', async () => {
        const yamlContent = document.getElementById('modal-code-content').textContent;
        await copyToClipboard(yamlContent, copyYamlButton);
    });

    // Handle keyboard navigation in modal
    document.addEventListener('keydown', (e) => {
        if (!currentTemplate) return; // Modal not open

        // Close on Escape key
        if (e.key === 'Escape') {
            closePreviewModal();
            return;
        }

        // Toggle debug mode with '@' key
        if (e.key === '@') {
            e.preventDefault();
            toggleDebugMode();
            return;
        }

        // Toggle ORG keyword filter with 'o' key
        if (e.key === 'o') {
            e.preventDefault();
            // Use the current modal template's org directly for toggling
            if (currentTemplate && currentTemplate.org) {
                const orgKeyword = currentTemplate.org;
                State.toggleKeywordSelection(orgKeyword);
                // Re-render to apply filter
                import('./appActions.js').then(({ filterAndRender }) => {
                    filterAndRender();
                });
            }
            return;
        }

        // Toggle ORG/REPO keyword filter with 'r' key
        if (e.key === 'r') {
            e.preventDefault();
            // Use the current modal template's repo directly for toggling
            if (currentTemplate && currentTemplate.repo) {
                const repoKeyword = currentTemplate.repo;
                State.toggleKeywordSelection(repoKeyword);
                // Re-render to apply filter
                import('./appActions.js').then(({ filterAndRender }) => {
                    filterAndRender();
                });
            }
            return;
        }

        // Ctrl+Arrow: Navigate to adjacent template in the list
        if (e.ctrlKey && ['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'].includes(e.key)) {
            e.preventDefault();
            const filteredTemplates = getFilteredTemplates();
            const currentIndex = filteredTemplates.findIndex(t => t.id === currentTemplate.id);
            if (currentIndex === -1) return;

            let nextIndex = currentIndex;

            if (e.key === 'ArrowRight') {
                nextIndex = currentIndex + 1;
            } else if (e.key === 'ArrowLeft') {
                nextIndex = currentIndex - 1;
            } else if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
                // Calculate grid column count from the template grid
                const grid = document.getElementById('template-grid');
                if (grid) {
                    const gridStyle = window.getComputedStyle(grid);
                    const columnCount = gridStyle.gridTemplateColumns.split(' ').length;
                    if (e.key === 'ArrowDown') {
                        nextIndex = currentIndex + columnCount;
                    } else {
                        nextIndex = currentIndex - columnCount;
                    }
                }
            }

            // Navigate to adjacent template if within bounds
            if (nextIndex >= 0 && nextIndex < filteredTemplates.length) {
                const nextTemplate = filteredTemplates[nextIndex];
                openPreviewModal(nextTemplate, true); // Preserve debug mode when navigating
            }
            return;
        }

        // Scroll the YAML content with keyboard
        // IMPORTANT: The CSS structure uses flex layout to keep scrollbars at viewport bottom:
        // - .modal-body: flex container with overflow: hidden (not scrollable)
        // - .modal-code-wrapper: flex item that grows (flex: 1)
        // - .modal-code: scrollable element (overflow: auto) with flex: 1 1 0
        //
        // This structure ensures:
        // - .modal-code is constrained by flex and scrolls both vertically and horizontally
        // - Both scrollbars stay at the bottom of the viewport (not the document)
        // - Modal-content grows naturally based on content (not always 90vh)
        //
        // SOLUTION:
        // - Scroll .modal-code for BOTH vertical and horizontal navigation
        const modalCode = document.querySelector('#preview-modal #modal-code');
        if (!modalCode) return;

        const scrollAmount = 40; // pixels per arrow key press
        const pageScrollAmount = modalCode.clientHeight * 0.9; // 90% of visible height

        let shouldScrollVertical = false;
        let shouldScrollHorizontal = false;
        let verticalScrollTo = null;
        let horizontalScrollTo = null;

        switch(e.key) {
            case 'Home':
                e.preventDefault();
                verticalScrollTo = 0;
                shouldScrollVertical = true;
                break;

            case 'End':
                e.preventDefault();
                verticalScrollTo = modalCode.scrollHeight;
                shouldScrollVertical = true;
                break;

            case 'PageUp':
                e.preventDefault();
                verticalScrollTo = Math.max(0, modalCode.scrollTop - pageScrollAmount);
                shouldScrollVertical = true;
                break;

            case 'PageDown':
                e.preventDefault();
                verticalScrollTo = Math.min(modalCode.scrollHeight, modalCode.scrollTop + pageScrollAmount);
                shouldScrollVertical = true;
                break;

            case 'ArrowUp':
                e.preventDefault();
                verticalScrollTo = Math.max(0, modalCode.scrollTop - scrollAmount);
                shouldScrollVertical = true;
                break;

            case 'ArrowDown':
                e.preventDefault();
                verticalScrollTo = Math.min(modalCode.scrollHeight, modalCode.scrollTop + scrollAmount);
                shouldScrollVertical = true;
                break;

            case 'ArrowLeft':
                // Only scroll if modalCode has horizontal overflow
                if (modalCode.scrollWidth > modalCode.clientWidth) {
                    e.preventDefault();
                    horizontalScrollTo = Math.max(0, modalCode.scrollLeft - scrollAmount);
                    shouldScrollHorizontal = true;
                }
                break;

            case 'ArrowRight':
                // Only scroll if modalCode has horizontal overflow
                if (modalCode.scrollWidth > modalCode.clientWidth) {
                    e.preventDefault();
                    horizontalScrollTo = Math.min(modalCode.scrollWidth, modalCode.scrollLeft + scrollAmount);
                    shouldScrollHorizontal = true;
                }
                break;
        }

        // Handle vertical scrolling
        if (shouldScrollVertical && verticalScrollTo !== null) {
            modalCode.scrollTo({
                top: verticalScrollTo,
                behavior: e.key.startsWith('Arrow') ? 'auto' : 'smooth'
            });
        }

        // Handle horizontal scrolling
        if (shouldScrollHorizontal && horizontalScrollTo !== null) {
            modalCode.scrollTo({
                left: horizontalScrollTo,
                behavior: 'auto'
            });
        }
    });

    // Handle browser back/forward navigation
    window.addEventListener('popstate', handlePopState);
}
