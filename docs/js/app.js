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
 * Filter and render templates based on current state
 */
function filterAndRender() {
    const templates = State.getTemplates();
    const repositories = State.getRepositories();
    const selectedKeywords = State.getSelectedKeywords();
    const selectedCategory = State.getSelectedCategory();

    // Get filter values from UI
    const searchTerm = document.getElementById('search').value;
    const typeFilter = document.getElementById('type-filter').value;
    const sortBy = document.getElementById('sort').value;

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
 * Clear all filters
 */
function clearAllFilters() {
    document.getElementById('search').value = '';
    document.getElementById('type-filter').value = '';
    State.clearAllSelections();
    filterAndRender();
}

/**
 * Setup UI event listeners
 */
function setupEventListeners() {
    // Debounce search input to avoid filtering on every keystroke
    const debouncedFilter = debounce(filterAndRender, 300);
    document.getElementById('search').addEventListener('input', debouncedFilter);

    // Immediate filtering for dropdowns and buttons
    document.getElementById('type-filter').addEventListener('change', filterAndRender);
    document.getElementById('sort').addEventListener('change', filterAndRender);
    document.getElementById('clear-filters').addEventListener('click', clearAllFilters);
}

/**
 * Load data and initialize application
 */
async function initialize() {
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
