/**
 * Main application orchestration
 */

import { loadAllData } from './data.js';
import * as State from './state.js';
import { applyFilters, sortTemplates } from './filters.js';
import { updateSidebar } from './sidebar.js';
import { renderTemplateGrid } from './templateCard.js';
import { openPreviewModal, setupModalEventListeners } from './modal.js';
import { debounce } from './utils.js';
import { initializeTheme } from './theme.js';

/**
 * Update statistics display
 */
function updateStats() {
    const templates = State.getTemplates();
    const filteredTemplates = State.getFilteredTemplates();

    document.getElementById('total-count').textContent = templates.length;
    document.getElementById('visible-count').textContent = filteredTemplates.length;
}

/**
 * Update clear button visibility
 */
function updateClearButtons() {
    // Clear search button
    const searchInput = document.getElementById('search');
    const clearSearchBtn = document.getElementById('clear-search');
    if (searchInput.value) {
        clearSearchBtn.style.display = 'block';
    } else {
        clearSearchBtn.style.display = 'none';
    }

    // Clear keywords button
    const selectedKeywords = State.getSelectedKeywords();
    const clearKeywordsBtn = document.getElementById('clear-keywords');
    if (selectedKeywords.size > 0) {
        clearKeywordsBtn.style.display = 'block';
    } else {
        clearKeywordsBtn.style.display = 'none';
    }
}

/**
 * Filter and render templates based on current state
 */
function filterAndRender(options = {}) {
    const templates = State.getTemplates();
    const repositories = State.getRepositories();
    const selectedKeywords = State.getSelectedKeywords();
    const selectedCategory = State.getSelectedCategory();

    // Get filter values from UI
    const searchTerm = document.getElementById('search').value;
    const showOfficial = document.getElementById('show-official').checked;
    const showCommunity = document.getElementById('show-community').checked;
    const sortBy = document.getElementById('sort').value;

    // Determine type filter based on checkboxes
    let typeFilter = '';
    if (showOfficial && !showCommunity) {
        typeFilter = 'official';
    } else if (!showOfficial && showCommunity) {
        typeFilter = 'community';
    }
    // If both or neither checked, typeFilter remains '' (show all)

    // Apply filters
    let filtered = applyFilters(templates, {
        searchTerm,
        typeFilter,
        selectedCategory,
        selectedKeywords
    });

    // Sort templates
    filtered = sortTemplates(filtered, sortBy, repositories);

    // Update state
    State.setFilteredTemplates(filtered);

    // Update UI
    updateStats();
    updateSidebar({
        filteredTemplates: filtered,
        selectedKeywords,
        selectedCategory
    }, handleKeywordToggle, handleCategoryToggle, options);
    updateClearButtons();

    // Render templates
    const gridElement = document.getElementById('templates-grid');
    renderTemplateGrid(filtered, repositories, gridElement, handleTemplateClick);
}

/**
 * Handle keyword toggle
 */
function handleKeywordToggle(keyword) {
    const wasSelected = State.getSelectedKeywords().has(keyword);
    const wasLastSelected = wasSelected && State.getSelectedKeywords().size === 1;
    State.toggleKeywordSelection(keyword);
    // If we just added a keyword, focus should move to first keyword in cloud
    // If we just removed the last selected keyword, focus should move to first unselected keyword
    filterAndRender({
        focusFirstKeyword: !wasSelected,
        focusFirstUnselected: wasLastSelected
    });
}

/**
 * Handle category toggle
 */
function handleCategoryToggle(category) {
    State.toggleCategorySelection(category);
    filterAndRender();
}

/**
 * Handle template card click
 */
function handleTemplateClick(template) {
    const repositories = State.getRepositories();
    const repo = repositories.get(template.repo);
    openPreviewModal(template, repo);
}

/**
 * Clear search field
 */
function clearSearch() {
    const searchInput = document.getElementById('search');
    searchInput.value = '';
    filterAndRender();
    // Restore focus to search input for continued keyboard navigation
    searchInput.focus();
}

/**
 * Clear all selected keywords
 */
function clearKeywords() {
    State.clearAllSelections();
    filterAndRender();
}

/**
 * Setup UI event listeners
 */
function setupEventListeners() {
    // Debounce search input to avoid filtering on every keystroke
    const debouncedFilter = debounce(filterAndRender, 300);
    const searchInput = document.getElementById('search');
    searchInput.addEventListener('input', debouncedFilter);

    // Update clear search button visibility on input
    searchInput.addEventListener('input', updateClearButtons);

    // Immediate filtering for checkboxes and dropdown
    document.getElementById('show-official').addEventListener('change', filterAndRender);
    document.getElementById('show-community').addEventListener('change', filterAndRender);
    document.getElementById('sort').addEventListener('change', filterAndRender);

    // Clear buttons
    document.getElementById('clear-search').addEventListener('click', clearSearch);
    document.getElementById('clear-keywords').addEventListener('click', clearKeywords);

    // Keyboard help button
    document.getElementById('keyboard-help-btn').addEventListener('click', showKeyboardHelp);
}

/**
 * Setup global keyboard shortcuts
 */
function setupKeyboardShortcuts() {
    const searchInput = document.getElementById('search');

    // Global keyboard shortcuts
    document.addEventListener('keydown', (e) => {
        // Skip if user is typing in an input/textarea
        const isTyping = document.activeElement.tagName === 'INPUT' ||
                        document.activeElement.tagName === 'TEXTAREA' ||
                        document.activeElement.isContentEditable;

        // Check if typing in search specifically
        const isTypingInSearch = document.activeElement === searchInput;

        // "/" hotkey to focus search box (like Gmail, GitHub)
        if (e.key === '/' && !isTyping) {
            e.preventDefault();
            searchInput.focus();
            searchInput.select();
            return;
        }

        // "?" hotkey to show keyboard help - works everywhere, even in search
        if (e.key === '?') {
            e.preventDefault();
            showKeyboardHelp(isTypingInSearch);
            return;
        }

        // K/k to focus first keyword
        // Uppercase works even when typing (e.g., Shift+K from search box)
        if ((e.key === 'k' && !isTyping) || e.key === 'K') {
            e.preventDefault();
            const firstKeyword = document.querySelector('.keyword-tag');
            if (firstKeyword) firstKeyword.focus();
            return;
        }

        // C/c to focus first category
        // Uppercase works even when typing
        if ((e.key === 'c' && !isTyping) || e.key === 'C') {
            e.preventDefault();
            const firstCategory = document.querySelector('.category-item');
            if (firstCategory) firstCategory.focus();
            return;
        }

        // S/s to focus first selected keyword
        // Uppercase works even when typing
        if ((e.key === 's' && !isTyping) || e.key === 'S') {
            e.preventDefault();
            const firstSelected = document.querySelector('.selected-keyword');
            if (firstSelected) firstSelected.focus();
            return;
        }

        // O/o to focus sort dropdown
        // Uppercase works even when typing
        if ((e.key === 'o' && !isTyping) || e.key === 'O') {
            e.preventDefault();
            const sortDropdown = document.getElementById('sort');
            if (sortDropdown) sortDropdown.focus();
            return;
        }

        // T/t to focus first template card
        // Uppercase works even when typing
        if ((e.key === 't' && !isTyping) || e.key === 'T') {
            e.preventDefault();
            const firstTemplate = document.querySelector('.template-card');
            if (firstTemplate) firstTemplate.focus();
            return;
        }
    });

    // ESC key to clear search box
    searchInput.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            clearSearch();
        }
    });

    // Prevent uppercase letters in search box (reserved for shortcuts)
    searchInput.addEventListener('keydown', (e) => {
        // Check if it's an uppercase letter
        if (e.key.length === 1 && e.key >= 'A' && e.key <= 'Z') {
            // Always prevent uppercase letters from being typed
            e.preventDefault();

            // If it's an assigned shortcut (K, C, S, T, O), the global handler will handle navigation
            const assignedShortcuts = ['K', 'C', 'S', 'T', 'O'];
            if (!assignedShortcuts.includes(e.key)) {
                // For unassigned uppercase letters, give visual feedback
                searchInput.classList.add('shake');
                setTimeout(() => searchInput.classList.remove('shake'), 300);
            }
            // Note: assigned shortcuts will trigger navigation via the global handler
        }
        // Note: '?' is handled by the global handler and works in search field
    });
}

/**
 * Show keyboard help overlay
 */
let keyboardHelpPreviousFocus = null;
let shouldRestoreFocus = true;

function showKeyboardHelp(returnFocusToSearch = false) {
    const existingOverlay = document.getElementById('keyboard-help-overlay');
    if (existingOverlay) {
        closeKeyboardHelp(returnFocusToSearch);
        return; // Toggle off if already shown
    }

    // Store the currently focused element to restore later
    keyboardHelpPreviousFocus = document.activeElement;
    shouldRestoreFocus = true;

    const overlay = document.createElement('div');
    overlay.id = 'keyboard-help-overlay';
    overlay.className = 'keyboard-help-overlay';
    overlay.setAttribute('role', 'dialog');
    overlay.setAttribute('aria-labelledby', 'keyboard-help-title');
    overlay.setAttribute('aria-modal', 'true');

    overlay.innerHTML = `
        <div class="keyboard-help-content">
            <div class="keyboard-help-header">
                <h2 id="keyboard-help-title">Keyboard Shortcuts</h2>
                <button class="keyboard-help-close" tabindex="0" aria-label="Close keyboard help">×</button>
            </div>
            <div class="keyboard-help-body">
                <div class="keyboard-help-section">
                    <h3>Navigation</h3>
                    <dl class="keyboard-shortcuts">
                        <dt><kbd>/</kbd></dt>
                        <dd>Focus search box</dd>
                        <dt><kbd>Esc</kbd></dt>
                        <dd>Clear search box</dd>
                        <dt><kbd>K</kbd> or <kbd>Shift+K</kbd></dt>
                        <dd>Jump to keywords</dd>
                        <dt><kbd>S</kbd> or <kbd>Shift+S</kbd></dt>
                        <dd>Jump to selected keywords</dd>
                        <dt><kbd>C</kbd> or <kbd>Shift+C</kbd></dt>
                        <dd>Jump to categories</dd>
                        <dt><kbd>O</kbd> or <kbd>Shift+O</kbd></dt>
                        <dd>Jump to sort dropdown</dd>
                        <dt><kbd>T</kbd> or <kbd>Shift+T</kbd></dt>
                        <dd>Jump to templates</dd>
                        <dt><kbd>↑</kbd> <kbd>↓</kbd> <kbd>←</kbd> <kbd>→</kbd></dt>
                        <dd>Navigate within sections</dd>
                        <dt><kbd>Tab</kbd></dt>
                        <dd>Navigate between elements</dd>
                    </dl>
                    <p style="font-size: 0.875rem; color: var(--text-light); margin-top: 1rem; font-style: italic;">
                        Tip: Uppercase shortcuts (Shift+K/C/S/T/O) work even when typing in the search box
                    </p>
                </div>
                <div class="keyboard-help-section">
                    <h3>Actions</h3>
                    <dl class="keyboard-shortcuts">
                        <dt><kbd>Enter</kbd> or <kbd>Space</kbd></dt>
                        <dd>Select keyword/category/template</dd>
                        <dt><kbd>Delete</kbd> or <kbd>Backspace</kbd></dt>
                        <dd>Remove selected keyword</dd>
                        <dt><kbd>?</kbd></dt>
                        <dd>Show/hide this help</dd>
                    </dl>
                </div>
            </div>
        </div>
    `;

    document.body.appendChild(overlay);

    // Close on click outside or close button
    const closeBtn = overlay.querySelector('.keyboard-help-close');
    const content = overlay.querySelector('.keyboard-help-content');

    closeBtn.addEventListener('click', () => closeKeyboardHelp(returnFocusToSearch));
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) closeKeyboardHelp(returnFocusToSearch);
    });

    // Get all focusable elements in the modal for focus trap
    const focusableElements = content.querySelectorAll('button, [tabindex="0"]');
    const firstFocusable = focusableElements[0];
    const lastFocusable = focusableElements[focusableElements.length - 1];

    // Handle keyboard events - ESC, ?, and focus trap
    overlay.addEventListener('keydown', (e) => {
        // Close on Escape or ?
        if (e.key === 'Escape' || e.key === '?') {
            e.preventDefault();
            closeKeyboardHelp(returnFocusToSearch);
            return;
        }

        // Focus trap - keep Tab within modal
        if (e.key === 'Tab') {
            if (e.shiftKey) {
                // Shift+Tab
                if (document.activeElement === firstFocusable) {
                    e.preventDefault();
                    lastFocusable.focus();
                }
            } else {
                // Tab
                if (document.activeElement === lastFocusable) {
                    e.preventDefault();
                    firstFocusable.focus();
                }
            }
        }

        // Handle shortcuts - close modal and execute shortcut
        const isShortcutKey = (e.key === 'k' || e.key === 'K' ||
                              e.key === 'c' || e.key === 'C' ||
                              e.key === 's' || e.key === 'S' ||
                              e.key === 't' || e.key === 'T' ||
                              e.key === 'o' || e.key === 'O' ||
                              e.key === '/');

        if (isShortcutKey) {
            e.preventDefault();
            shouldRestoreFocus = false; // Don't restore focus, let the shortcut handle it
            overlay.remove();
            // Re-dispatch the event to trigger the global handler
            const newEvent = new KeyboardEvent('keydown', {
                key: e.key,
                code: e.code,
                shiftKey: e.shiftKey,
                ctrlKey: e.ctrlKey,
                altKey: e.altKey,
                metaKey: e.metaKey,
                bubbles: true
            });
            document.dispatchEvent(newEvent);
        }
    });

    // Focus the close button for accessibility
    closeBtn.focus();
}

function closeKeyboardHelp(returnFocusToSearch = false) {
    const overlay = document.getElementById('keyboard-help-overlay');
    if (overlay) {
        overlay.remove();

        // Restore focus
        if (shouldRestoreFocus) {
            if (returnFocusToSearch) {
                const searchInput = document.getElementById('search');
                if (searchInput) searchInput.focus();
            } else if (keyboardHelpPreviousFocus && keyboardHelpPreviousFocus.focus) {
                keyboardHelpPreviousFocus.focus();
            }
        }

        keyboardHelpPreviousFocus = null;
        shouldRestoreFocus = true;
    }
}

/**
 * Load data and initialize application
 */
async function initialize() {
    // Initialize theme early to prevent flash
    initializeTheme();

    const loading = document.getElementById('loading');
    const error = document.getElementById('error');

    try {
        loading.textContent = 'Loading templates...';

        // Load data
        const { templates, repositories } = await loadAllData();

        // Update state
        State.setTemplates(templates);
        State.setRepositories(repositories);
        State.setFilteredTemplates([...templates]);

        // Hide loading
        loading.style.display = 'none';

        // Setup UI
        setupEventListeners();
        setupModalEventListeners();
        setupKeyboardShortcuts();

        // Initial render
        filterAndRender();

        // Auto-focus search box for immediate typing
        document.getElementById('search').focus();

    } catch (err) {
        console.error('Error loading data:', err);
        loading.style.display = 'none';
        error.style.display = 'block';
        error.textContent = `Error loading catalog data: ${err.message}`;
    }
}

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', initialize);
