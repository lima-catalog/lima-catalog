/**
 * Preview modal functionality
 */

import { getDefaultBranchURL, getGitHubSchemeURL, getRawContentURL } from './urlHelpers.js';
import { deriveDisplayName } from './templateCard.js';
import { trapFocus } from './utils.js';

// Modal state
let currentTemplate = null;
let releaseFocusTrap = null;
let previouslyFocusedElement = null;

/**
 * Open preview modal for a template
 * @param {Object} template - Template object
 * @param {Object} repo - Repository object
 */
export function openPreviewModal(template, repo) {
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
    const displayURL = getDefaultBranchURL(template, repo);
    modalGithubLink.href = displayURL;
    modalGithubLink.textContent = displayURL;

    // Show modal and loading state
    modal.style.display = 'flex';
    modalLoading.classList.remove('hidden');
    modalCode.classList.add('hidden');
    copyYamlButton.style.display = 'none';
    document.body.style.overflow = 'hidden';

    // Trap focus within modal for accessibility
    setTimeout(() => {
        releaseFocusTrap = trapFocus(modal.querySelector('.modal-content'));
    }, 100);

    // Fetch and display template content
    fetchTemplateContent(template, repo);
}

/**
 * Close preview modal
 */
export function closePreviewModal() {
    const modal = document.getElementById('preview-modal');
    modal.style.display = 'none';
    document.body.style.overflow = 'auto';
    currentTemplate = null;

    // Release focus trap
    if (releaseFocusTrap) {
        releaseFocusTrap();
        releaseFocusTrap = null;
    }

    // Restore focus to the element that opened the modal
    if (previouslyFocusedElement && previouslyFocusedElement.focus) {
        previouslyFocusedElement.focus();
        previouslyFocusedElement = null;
    }
}

/**
 * Fetch and display template content
 * @param {Object} template - Template object
 * @param {Object} repo - Repository object
 */
async function fetchTemplateContent(template, repo) {
    const modalLoading = document.getElementById('modal-loading');
    const modalCode = document.getElementById('modal-code');
    const modalCodeContent = document.getElementById('modal-code-content');
    const copyYamlButton = document.getElementById('copy-yaml');

    try {
        // Use default branch URL for fetching latest content
        const url = getDefaultBranchURL(template, repo);
        const rawURL = getRawContentURL(url);

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
    } catch (error) {
        console.error('Error fetching template:', error);
        modalLoading.textContent = `Error loading template: ${error.message}`;
        copyYamlButton.style.display = 'none';
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
        const modalCode = document.getElementById('modal-code');
        if (!modalCode || modalCode.classList.contains('hidden')) return;

        const scrollAmount = 40; // pixels per arrow key press
        const pageScrollAmount = modalCode.clientHeight * 0.9; // 90% of visible height

        let shouldScroll = false;
        let scrollTo = null;

        switch(e.key) {
            case 'Home':
                e.preventDefault();
                scrollTo = 0;
                shouldScroll = true;
                break;

            case 'End':
                e.preventDefault();
                scrollTo = modalCode.scrollHeight;
                shouldScroll = true;
                break;

            case 'PageUp':
                e.preventDefault();
                scrollTo = Math.max(0, modalCode.scrollTop - pageScrollAmount);
                shouldScroll = true;
                break;

            case 'PageDown':
                e.preventDefault();
                scrollTo = Math.min(modalCode.scrollHeight, modalCode.scrollTop + pageScrollAmount);
                shouldScroll = true;
                break;

            case 'ArrowUp':
                e.preventDefault();
                scrollTo = Math.max(0, modalCode.scrollTop - scrollAmount);
                shouldScroll = true;
                break;

            case 'ArrowDown':
                e.preventDefault();
                scrollTo = Math.min(modalCode.scrollHeight, modalCode.scrollTop + scrollAmount);
                shouldScroll = true;
                break;
        }

        if (shouldScroll && scrollTo !== null) {
            modalCode.scrollTo({
                top: scrollTo,
                behavior: e.key.startsWith('Arrow') ? 'auto' : 'smooth'
            });
        }
    });
}
