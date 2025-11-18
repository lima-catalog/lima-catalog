/**
 * Application actions that can be called from other modules
 * This breaks circular dependencies between app.js and keyboard.js
 */

import * as State from './state.js';
import { applyFilters, sortTemplates } from './filters.js';
import { updateSidebar } from './sidebar.js';
import { renderTemplateGrid } from './templateCard.js';

/**
 * Filter and render templates based on current state
 */
export function filterAndRender(options = {}) {
    const templates = State.getTemplates();
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
    filtered = sortTemplates(filtered, sortBy);

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
    renderTemplateGrid(filtered, gridElement, handleTemplateClick, sortBy);
}

/**
 * Clear search field
 */
export function clearSearch() {
    const searchInput = document.getElementById('search');
    searchInput.value = '';
    filterAndRender();
    // Restore focus to search input for continued keyboard navigation
    searchInput.focus();
}

/**
 * Update sort dropdown options based on debug mode
 */
export function updateSortDropdown() {
    const sortDropdown = document.getElementById('sort');
    if (!sortDropdown) return;

    const currentValue = sortDropdown.value;
    const debugMode = State.getDebugMode();

    // Base options (always available)
    const baseOptions = [
        { value: 'name', label: 'Name (A-Z)' },
        { value: 'stars', label: 'Stars (High to Low)' },
        { value: 'updated', label: 'Recently Updated' },
        { value: 'notability', label: 'Notability Score' }
    ];

    // Debug options (only in debug mode)
    const debugOptions = [
        { value: 'breakdown-message', label: '[Debug] Message Length' },
        { value: 'breakdown-provision', label: '[Debug] Provision Scripts' },
        { value: 'breakdown-parameters', label: '[Debug] Parameters + Env' },
        { value: 'breakdown-comments', label: '[Debug] YAML Comments' },
        { value: 'breakdown-unusual_images', label: '[Debug] Unusual Images' }
    ];

    // Combine options based on debug mode
    const allOptions = debugMode ? [...baseOptions, ...debugOptions] : baseOptions;

    // Rebuild dropdown
    sortDropdown.innerHTML = '';
    allOptions.forEach(opt => {
        const option = document.createElement('option');
        option.value = opt.value;
        option.textContent = opt.label;
        sortDropdown.appendChild(option);
    });

    // Restore previous value if it's still valid
    const validValues = allOptions.map(o => o.value);
    if (validValues.includes(currentValue)) {
        sortDropdown.value = currentValue;
    } else {
        sortDropdown.value = 'name'; // Default fallback
    }
}

/**
 * Show debug mode notification
 */
export function showDebugModeNotification(enabled) {
    // Remove any existing notification
    const existing = document.getElementById('debug-mode-notification');
    if (existing) existing.remove();

    // Create new notification
    const notification = document.createElement('div');
    notification.id = 'debug-mode-notification';
    notification.className = 'debug-mode-notification';
    notification.textContent = enabled ?
        'Debug mode enabled - showing notability breakdowns in sort dropdown' :
        'Debug mode disabled';

    document.body.appendChild(notification);

    // Auto-remove after 3 seconds
    setTimeout(() => {
        notification.remove();
    }, 3000);
}

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
 * Update clear keywords button visibility
 */
function updateClearButtons() {
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
    // Import dynamically to avoid circular dependency
    import('./modal.js').then(({ openPreviewModal }) => {
        openPreviewModal(template);
    });
}
