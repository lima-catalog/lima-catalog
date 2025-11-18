/**
 * Main application orchestration
 */

import { loadAllData } from './data.js';
import * as State from './state.js';
import { applyFilters, sortTemplates } from './filters.js';
import { updateSidebar, setupSidebarNavigation } from './sidebar.js';
import { renderTemplateGrid } from './templateCard.js';
import { openPreviewModal, setupModalEventListeners } from './modal.js';
import { debounce } from './utils.js';
import { initializeTheme } from './theme.js';
import { setupKeyboardShortcuts, showKeyboardHelp } from './keyboard.js';

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
    openPreviewModal(template);
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

    // Immediate filtering for checkboxes and dropdown
    document.getElementById('show-official').addEventListener('change', filterAndRender);
    document.getElementById('show-community').addEventListener('change', filterAndRender);
    document.getElementById('sort').addEventListener('change', filterAndRender);

    // Clear keywords button
    document.getElementById('clear-keywords').addEventListener('click', clearKeywords);

    // Keyboard help button - opens help tab
    document.getElementById('keyboard-help-btn').addEventListener('click', () => showKeyboardHelp(false, 'help'));

    // App icon - opens about tab
    const headerIcon = document.querySelector('.header-icon');
    if (headerIcon) {
        headerIcon.style.cursor = 'pointer';
        headerIcon.setAttribute('role', 'button');
        headerIcon.setAttribute('aria-label', 'About Lima Catalog');
        headerIcon.setAttribute('tabindex', '0');
        headerIcon.addEventListener('click', () => showKeyboardHelp(false, 'about'));
        headerIcon.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                showKeyboardHelp(false, 'about');
            }
        });
    }
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
        const { templates } = await loadAllData();

        // Update state
        State.setTemplates(templates);
        State.setFilteredTemplates([...templates]);

        // Hide loading
        loading.style.display = 'none';

        // Setup UI
        setupEventListeners();
        setupModalEventListeners();
        setupKeyboardShortcuts();
        setupSidebarNavigation();

        // Initialize sort dropdown (sets base options)
        updateSortDropdown();

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
