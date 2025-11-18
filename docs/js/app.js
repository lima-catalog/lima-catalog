/**
 * Main application orchestration
 */

import { loadAllData } from './data.js';
import * as State from './state.js';
import { filterAndRender, updateSortDropdown } from './appActions.js';
import { setupSidebarNavigation } from './sidebar.js';
import { setupModalEventListeners } from './modal.js';
import { initializeTheme } from './theme.js';
import { setupKeyboardShortcuts, showKeyboardHelp } from './keyboard.js';
import { debounce } from './utils.js';

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
