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
function filterAndRender() {
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
    }, handleKeywordToggle, handleCategoryToggle);
    updateClearButtons();

    // Render templates
    const gridElement = document.getElementById('templates-grid');
    renderTemplateGrid(filtered, repositories, gridElement, handleTemplateClick);
}

/**
 * Handle keyword toggle
 */
function handleKeywordToggle(keyword) {
    State.toggleKeywordSelection(keyword);
    filterAndRender();
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
    document.getElementById('search').value = '';
    filterAndRender();
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

        // "/" hotkey to focus search box (like Gmail, GitHub)
        if (e.key === '/' && !isTyping) {
            e.preventDefault();
            searchInput.focus();
            searchInput.select();
            return;
        }

        // "?" hotkey to show keyboard help
        if (e.key === '?' && !isTyping) {
            e.preventDefault();
            showKeyboardHelp();
            return;
        }

        // K to focus first keyword
        if (e.key === 'k' && !isTyping) {
            e.preventDefault();
            const firstKeyword = document.querySelector('.keyword-tag');
            if (firstKeyword) firstKeyword.focus();
            return;
        }

        // C to focus first category
        if (e.key === 'c' && !isTyping) {
            e.preventDefault();
            const firstCategory = document.querySelector('.category-item');
            if (firstCategory) firstCategory.focus();
            return;
        }

        // T to focus first template card
        if (e.key === 't' && !isTyping) {
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
}

/**
 * Show keyboard help overlay
 */
function showKeyboardHelp() {
    const existingOverlay = document.getElementById('keyboard-help-overlay');
    if (existingOverlay) {
        existingOverlay.remove();
        return; // Toggle off if already shown
    }

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
                <button class="keyboard-help-close" aria-label="Close keyboard help">×</button>
            </div>
            <div class="keyboard-help-body">
                <div class="keyboard-help-section">
                    <h3>Navigation</h3>
                    <dl class="keyboard-shortcuts">
                        <dt><kbd>/</kbd></dt>
                        <dd>Focus search box</dd>
                        <dt><kbd>Esc</kbd></dt>
                        <dd>Clear search box</dd>
                        <dt><kbd>K</kbd></dt>
                        <dd>Jump to keywords</dd>
                        <dt><kbd>C</kbd></dt>
                        <dd>Jump to categories</dd>
                        <dt><kbd>T</kbd></dt>
                        <dd>Jump to templates</dd>
                        <dt><kbd>↑</kbd> <kbd>↓</kbd> <kbd>←</kbd> <kbd>→</kbd></dt>
                        <dd>Navigate within sections</dd>
                        <dt><kbd>Tab</kbd></dt>
                        <dd>Navigate between elements</dd>
                    </dl>
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

    closeBtn.addEventListener('click', () => overlay.remove());
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) overlay.remove();
    });

    // Close on Escape key
    overlay.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' || e.key === '?') {
            e.preventDefault();
            overlay.remove();
        }
    });

    // Focus the close button for accessibility
    closeBtn.focus();
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
