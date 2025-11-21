/**
 * Preview modal functionality
 */

import { getDefaultBranchURL, getGitHubSchemeURL, getRawContentURL } from './urlHelpers.js';
import { deriveDisplayName } from './templateCard.js';
import { trapFocus } from './utils.js';
import { getTemplates } from './state.js';

// Modal state
let currentTemplate = null;
let currentYamlContent = null; // Store original YAML for diff comparison
let releaseFocusTrap = null;
let previouslyFocusedElement = null;
let isHandlingPopState = false; // Flag to prevent duplicate popstate handling

/**
 * Generate a unified diff between two texts
 * @param {string} originalText - Original text
 * @param {string} newText - New text to compare
 * @param {string} originalName - Name for original file
 * @param {string} newName - Name for new file
 * @returns {string} Unified diff output
 */
function generateUnifiedDiff(originalText, newText, originalName = 'original', newName = 'similar') {
    const originalLines = originalText.split('\n');
    const newLines = newText.split('\n');

    // Simple LCS-based diff algorithm
    const lcs = computeLCS(originalLines, newLines);
    const diff = buildUnifiedDiff(originalLines, newLines, lcs, originalName, newName);

    return diff;
}

/**
 * Compute Longest Common Subsequence for diff
 */
function computeLCS(a, b) {
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

    // Build output
    if (hunks.length === 0) {
        return '# No differences found';
    }

    let output = `--- ${originalName}\n+++ ${newName}\n`;

    for (const hunk of hunks) {
        const aCount = hunk.lines.filter(l => l.type !== '+').length;
        const bCount = hunk.lines.filter(l => l.type !== '-').length;
        output += `@@ -${hunk.aStart},${aCount} +${hunk.bStart},${bCount} @@\n`;

        for (const line of hunk.lines) {
            output += `${line.type}${line.text}\n`;
        }
    }

    return output;
}

/**
 * URL parameter utilities for deep linking
 */

/**
 * Get template ID from URL parameters
 * @returns {string|null} Template ID if present in URL
 */
function getTemplateFromURL() {
    const params = new URLSearchParams(window.location.search);
    return params.get('template');
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
 */
export function openPreviewModal(template) {
    currentTemplate = template;

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

    // Set github: scheme URL
    const githubSchemeURL = getGitHubSchemeURL(template);
    modalGithubScheme.textContent = githubSchemeURL;

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

        // Apply syntax highlighting with highlight.js
        modalCodeContent.textContent = content;
        modalCodeContent.removeAttribute('data-highlighted');
        hljs.highlightElement(modalCodeContent);

        // Show code and copy button, hide loading
        modalLoading.classList.add('hidden');
        modalCode.classList.remove('hidden');
        copyYamlButton.style.display = 'block';

        // Show modal-content now that content is loaded (fade in)
        const modal = document.getElementById('preview-modal');
        const modalContent = modal.querySelector('.modal-content');
        modalContent.classList.add('ready');
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

    // Check if template has similar templates
    if (!template.similar_templates || template.similar_templates.length === 0) {
        similarSection.classList.add('hidden');
        return;
    }

    // Show section and populate list
    similarSection.classList.remove('hidden');
    similarList.innerHTML = '';

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

    // Restore original YAML display
    const restoreOriginalYaml = () => {
        if (currentYamlContent && isShowingDiff) {
            modalCodeContent.textContent = currentYamlContent;
            modalCodeContent.className = 'language-yaml';
            modalCodeContent.removeAttribute('data-highlighted');
            hljs.highlightElement(modalCodeContent);
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

        modalCodeContent.textContent = diff;
        modalCodeContent.className = 'language-diff';
        modalCodeContent.removeAttribute('data-highlighted');
        hljs.highlightElement(modalCodeContent);
        isShowingDiff = true;
    };

    // Sort by similarity (descending), then by id (ascending)
    const sortedSimilarTemplates = [...template.similar_templates].sort((a, b) => {
        if (b.similarity !== a.similarity) {
            return b.similarity - a.similarity;
        }
        return a.id.localeCompare(b.id);
    });

    sortedSimilarTemplates.forEach((similar, index) => {
        const similarTemplate = templateMap.get(similar.id);
        const similarityPercent = Math.round(similar.similarity * 100);

        const item = document.createElement('div');
        item.className = 'similar-template-item';
        item.setAttribute('role', 'option');
        item.setAttribute('aria-selected', 'false');
        item.dataset.index = index;

        // Single line format: ORG/REPO/TEMPLATEPATH [badge] percent
        item.innerHTML = `
            <span class="similar-template-path">${escapeHtml(similar.id)}</span>
            ${similar.duplicate_type ? `<span class="duplicate-badge ${escapeHtml(similar.duplicate_type)}">${escapeHtml(similar.duplicate_type)}</span>` : ''}
            <span class="similarity-percentage">${similarityPercent}%</span>
        `;

        // Click handler to open the similar template
        item.addEventListener('click', () => {
            if (similarTemplate) {
                closePreviewModal();
                setTimeout(() => openPreviewModal(similarTemplate), 100);
            }
        });

        items.push({ element: item, template: similarTemplate, id: similar.id });
        similarList.appendChild(item);
    });

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
            const newIndex = selectedIndex < items.length - 1 ? selectedIndex + 1 : 0;
            updateSelection(newIndex);
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            e.stopPropagation();
            const newIndex = selectedIndex > 0 ? selectedIndex - 1 : items.length - 1;
            updateSelection(newIndex);
        } else if ((e.key === 'Enter' || e.key === ' ') && selectedIndex >= 0) {
            e.preventDefault();
            e.stopPropagation();
            const selected = items[selectedIndex];
            if (selected && selected.template) {
                closePreviewModal();
                setTimeout(() => openPreviewModal(selected.template), 100);
            }
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
function escapeHtml(str) {
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
function handlePopState() {
    isHandlingPopState = true;

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

    // Reset flag after a short delay
    setTimeout(() => {
        isHandlingPopState = false;
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
