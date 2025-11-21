/**
 * Preview modal functionality
 */

import { getDefaultBranchURL, getGitHubSchemeURL, getRawContentURL } from './urlHelpers.js';
import { deriveDisplayName } from './templateCard.js';
import { trapFocus } from './utils.js';
import { getTemplates } from './state.js';

// Modal state
let currentTemplate = null;
let releaseFocusTrap = null;
let previouslyFocusedElement = null;
let isHandlingPopState = false; // Flag to prevent duplicate popstate handling

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

    // Check if template has similar templates
    if (!template.similar_templates || template.similar_templates.length === 0) {
        similarSection.classList.add('hidden');
        return;
    }

    // Show section and populate list
    similarSection.classList.remove('hidden');
    similarList.innerHTML = '';

    // Get all templates for looking up names
    const allTemplates = getTemplates();
    const templateMap = new Map(allTemplates.map(t => [t.id, t]));

    template.similar_templates.forEach(similar => {
        const similarTemplate = templateMap.get(similar.id);
        const displayName = similarTemplate ? deriveDisplayName(similarTemplate) : similar.id;
        const similarityPercent = Math.round(similar.similarity * 100);

        const item = document.createElement('div');
        item.className = 'similar-template-item';
        item.setAttribute('role', 'listitem');
        item.setAttribute('tabindex', '0');
        item.setAttribute('aria-label', `Similar template: ${displayName}, ${similarityPercent}% similar`);

        item.innerHTML = `
            <div class="similar-template-info">
                <div class="similar-template-name">${escapeHtml(displayName)}</div>
                <div class="similar-template-similarity">From ${escapeHtml(similar.id.split('/').slice(0, 2).join('/'))}</div>
            </div>
            <div class="similar-template-badges">
                ${similar.duplicate_type ? `<span class="duplicate-badge ${escapeHtml(similar.duplicate_type)}">${escapeHtml(similar.duplicate_type)}</span>` : ''}
                <span class="similarity-percentage">${similarityPercent}%</span>
            </div>
        `;

        // Click handler to open the similar template
        const clickHandler = () => {
            if (similarTemplate) {
                closePreviewModal();
                setTimeout(() => openPreviewModal(similarTemplate), 100);
            }
        };
        item.addEventListener('click', clickHandler);

        // Keyboard handler for accessibility
        item.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                clickHandler();
            }
        });

        similarList.appendChild(item);
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
    const modalClose = document.getElementById('modal-close');
    const modalCloseButton = document.getElementById('modal-close-button');
    const copyGithubUrlButton = document.getElementById('copy-github-url');
    const copyYamlButton = document.getElementById('copy-yaml');

    // Close on overlay click
    modalOverlay.addEventListener('click', closePreviewModal);

    // Close on X button click
    modalClose.addEventListener('click', closePreviewModal);

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
