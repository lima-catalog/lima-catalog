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

        // Initial render
        filterAndRender();

    } catch (err) {
        console.error('Error loading data:', err);
        loading.style.display = 'none';
        error.style.display = 'block';
        error.textContent = `Error loading catalog data: ${err.message}`;
    }
}

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', initialize);
