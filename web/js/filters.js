/**
 * Filtering and counting utilities
 */

import { MAX_KEYWORDS_DISPLAY } from './config.js';
import * as State from './state.js';

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
 * Get dynamic ORG/REPO keywords based on focused template
 * @param {Array} allTemplates - All templates in the catalog (to determine org/repo structure)
 * @param {Array} filteredTemplates - Currently filtered templates (to calculate counts)
 * @param {Object} focusedTemplate - Currently focused template
 * @returns {Array} Array of [keyword, count, isDynamic] tuples
 */
export function getDynamicKeywords(allTemplates, filteredTemplates, focusedTemplate) {
    console.log('[getDynamicKeywords] Called with:', {
        focusedTemplate,
        hasOrg: focusedTemplate?.org,
        hasRepo: focusedTemplate?.repo,
        allTemplateCount: allTemplates.length,
        filteredTemplateCount: filteredTemplates.length
    });

    if (!focusedTemplate || !focusedTemplate.org || !focusedTemplate.repo) {
        console.log('[getDynamicKeywords] Returning empty - missing data');
        return [];
    }

    const dynamicKeywords = [];
    const focusedOrg = focusedTemplate.org;
    const focusedRepo = focusedTemplate.repo;

    // Group ALL templates by org and repo (to determine structure)
    const orgRepos = new Map(); // Map<org, Set<repo>>
    allTemplates.forEach(template => {
        if (template.org && template.repo) {
            if (!orgRepos.has(template.org)) {
                orgRepos.set(template.org, new Set());
            }
            orgRepos.get(template.org).add(template.repo);
        }
    });

    // Count FILTERED templates per org/repo
    const orgTemplateCount = new Map(); // Map<org, count>
    const repoTemplateCount = new Map(); // Map<repo, count>
    filteredTemplates.forEach(template => {
        if (template.org) {
            orgTemplateCount.set(template.org, (orgTemplateCount.get(template.org) || 0) + 1);
        }
        if (template.repo) {
            repoTemplateCount.set(template.repo, (repoTemplateCount.get(template.repo) || 0) + 1);
        }
    });

    const reposInFocusedOrg = orgRepos.get(focusedOrg);
    const repoCount = reposInFocusedOrg ? reposInFocusedOrg.size : 0;
    const allTemplatesInFocusedRepo = allTemplates.filter(t => t.repo === focusedRepo).length;

    // Determine which keywords to add based on the rules
    if (repoCount > 1) {
        // Multiple repos from same org: add ORG keyword and REPO keywords for each repo

        // Add ORG keyword if it would filter something out
        const orgCount = orgTemplateCount.get(focusedOrg) || 0;
        if (orgCount > 0 && orgCount < filteredTemplates.length) {
            dynamicKeywords.push([focusedOrg, orgCount, true]);
        }

        // Add REPO keywords for each repo in this org
        reposInFocusedOrg.forEach(repo => {
            const count = repoTemplateCount.get(repo) || 0;
            // Only add if it would filter something out
            if (count > 0 && count < filteredTemplates.length) {
                dynamicKeywords.push([repo, count, true]);
            }
        });
    } else if (repoCount === 1 && allTemplatesInFocusedRepo > 1) {
        // Only one repo from this org, but multiple templates: add REPO keyword
        const count = repoTemplateCount.get(focusedRepo) || 0;
        // Only add if it would filter something out
        if (count > 0 && count < filteredTemplates.length) {
            dynamicKeywords.push([focusedRepo, count, true]);
        }
    }

    console.log('[getDynamicKeywords] Returning:', {
        count: dynamicKeywords.length,
        keywords: dynamicKeywords.map(k => k[0]),
        repoCount,
        filteredTemplateCount: filteredTemplates.length
    });

    return dynamicKeywords;
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
 * @param {boolean} options.showDuplicates - Whether to show duplicate templates (default: false)
 * @returns {Array} Filtered templates
 */
export function applyFilters(templates, { searchTerm = '', typeFilter = '', selectedCategory = null, selectedKeywords = new Set(), showDuplicates = false }) {
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
                // Check if it's a regular keyword first
                if (templateKeywords.has(keyword)) {
                    continue; // Match found, check next keyword
                }

                // Not a regular keyword - check if it's a dynamic org/repo keyword
                // Repo keywords contain '/', org keywords don't
                if (keyword.includes('/')) {
                    // Repo keyword - must match template.repo
                    if (template.repo !== keyword) return false;
                } else {
                    // Org keyword - must match template.org
                    if (template.org !== keyword) return false;
                }
            }
        }

        // Type filter
        if (typeFilter === 'official' && !template.official) return false;
        if (typeFilter === 'community' && template.official) return false;

        // Duplicate filter - hide templates that are copies unless showDuplicates is enabled
        if (!showDuplicates && template.original_id) return false;

        return true;
    });
}

/**
 * Sort templates by specified criteria
 * @param {Array} templates - Templates to sort
 * @param {string} sortBy - Sort criteria ('notability', 'name', 'stars', 'updated', or breakdown components)
 * @returns {Array} Sorted templates (mutates original array)
 */
export function sortTemplates(templates, sortBy) {
    return templates.sort((a, b) => {
        switch (sortBy) {
            case 'notability':
                return (b.notability_score || 0) - (a.notability_score || 0);
            case 'name':
                return (a.name || a.path).localeCompare(b.name || b.path);
            case 'stars':
                return (b.stars || 0) - (a.stars || 0);
            case 'updated':
                return new Date(b.updated_at) - new Date(a.updated_at);

            // Debug mode: sort by individual notability breakdown components
            case 'breakdown-message':
                return (b.notability_score_breakdown?.message || 0) - (a.notability_score_breakdown?.message || 0);
            case 'breakdown-provision':
                return (b.notability_score_breakdown?.provision || 0) - (a.notability_score_breakdown?.provision || 0);
            case 'breakdown-parameters':
                return (b.notability_score_breakdown?.parameters || 0) - (a.notability_score_breakdown?.parameters || 0);
            case 'breakdown-env_vars':
                return (b.notability_score_breakdown?.env_vars || 0) - (a.notability_score_breakdown?.env_vars || 0);
            case 'breakdown-probes':
                return (b.notability_score_breakdown?.probes || 0) - (a.notability_score_breakdown?.probes || 0);
            case 'breakdown-unusual_images':
                return (b.notability_score_breakdown?.unusual_images || 0) - (a.notability_score_breakdown?.unusual_images || 0);
            case 'breakdown-comments':
                return (b.notability_score_breakdown?.comments || 0) - (a.notability_score_breakdown?.comments || 0);
            case 'breakdown-stars':
                return (b.notability_score_breakdown?.stars || 0) - (a.notability_score_breakdown?.stars || 0);

            default:
                return 0;
        }
    });
}
