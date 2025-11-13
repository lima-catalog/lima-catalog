/**
 * Filtering and counting utilities
 */

import { MAX_KEYWORDS_DISPLAY } from './config.js';

/**
 * Get all keywords with counts from a specific template list
 * @param {Array} templateList - List of templates
 * @param {Set} excludeKeywords - Keywords to exclude from results
 * @returns {Array} Array of [keyword, count] tuples, sorted by count descending
 */
export function getKeywordCounts(templateList, excludeKeywords = new Set()) {
    const counts = new Map();
    templateList.forEach(template => {
        if (template.keywords && Array.isArray(template.keywords)) {
            template.keywords.forEach(kw => {
                // Don't include excluded keywords (e.g., already selected)
                if (!excludeKeywords.has(kw)) {
                    counts.set(kw, (counts.get(kw) || 0) + 1);
                }
            });
        }
    });
    return Array.from(counts.entries())
        .sort((a, b) => b[1] - a[1]) // Sort by count descending
        .slice(0, MAX_KEYWORDS_DISPLAY);
}

/**
 * Get category counts from a specific template list
 * @param {Array} templateList - List of templates
 * @returns {Array} Array of [category, count] tuples, sorted alphabetically
 */
export function getCategoryCounts(templateList) {
    const counts = new Map();
    templateList.forEach(template => {
        if (template.category) {
            counts.set(template.category, (counts.get(template.category) || 0) + 1);
        }
    });
    return Array.from(counts.entries())
        .sort((a, b) => a[0].localeCompare(b[0])); // Sort alphabetically
}

/**
 * Apply filters to templates
 * @param {Array} templates - All templates
 * @param {Object} options - Filter options
 * @param {string} options.searchTerm - Search term
 * @param {string} options.typeFilter - Type filter ('official', 'community', or '')
 * @param {string} options.selectedCategory - Selected category
 * @param {Set} options.selectedKeywords - Selected keywords
 * @returns {Array} Filtered templates
 */
export function applyFilters(templates, { searchTerm = '', typeFilter = '', selectedCategory = null, selectedKeywords = new Set() }) {
    return templates.filter(template => {
        // Search filter
        if (searchTerm) {
            const searchText = [
                template.name,
                template.display_name,
                template.short_description,
                template.category,
                template.repo,
                ...(template.keywords || []),
                ...(template.images || [])
            ].join(' ').toLowerCase();

            if (!searchText.includes(searchTerm.toLowerCase())) return false;
        }

        // Category filter
        if (selectedCategory && template.category !== selectedCategory) return false;

        // Keyword filter (AND logic - template must have ALL selected keywords)
        if (selectedKeywords.size > 0) {
            const templateKeywords = new Set(template.keywords || []);
            for (const keyword of selectedKeywords) {
                if (!templateKeywords.has(keyword)) return false;
            }
        }

        // Type filter
        if (typeFilter === 'official' && !template.is_official) return false;
        if (typeFilter === 'community' && template.is_official) return false;

        return true;
    });
}

/**
 * Sort templates by specified criteria
 * @param {Array} templates - Templates to sort
 * @param {string} sortBy - Sort criteria ('name', 'stars', 'updated')
 * @param {Map} repositories - Repository data map
 * @returns {Array} Sorted templates (mutates original array)
 */
export function sortTemplates(templates, sortBy, repositories) {
    return templates.sort((a, b) => {
        switch (sortBy) {
            case 'name':
                return (a.name || a.path).localeCompare(b.name || b.path);
            case 'stars':
                const repoA = repositories.get(a.repo);
                const repoB = repositories.get(b.repo);
                return (repoB?.stars || 0) - (repoA?.stars || 0);
            case 'updated':
                return new Date(b.last_checked) - new Date(a.last_checked);
            default:
                return 0;
        }
    });
}
