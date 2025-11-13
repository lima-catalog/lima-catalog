/**
 * Sidebar rendering and interaction
 */

import { escapeHtml, formatCategoryName } from './templateCard.js';
import { getKeywordCounts, getCategoryCounts } from './filters.js';

/**
 * Render keyword cloud
 * @param {Array} filteredTemplates - Currently filtered templates
 * @param {Set} selectedKeywords - Currently selected keywords
 * @param {HTMLElement} cloudElement - Cloud container element
 * @param {Function} onKeywordClick - Click handler for keywords
 */
export function renderKeywordCloud(filteredTemplates, selectedKeywords, cloudElement, onKeywordClick) {
    // Get keywords from currently filtered templates, excluding selected ones
    const keywords = getKeywordCounts(filteredTemplates, selectedKeywords);

    if (keywords.length === 0) {
        cloudElement.innerHTML = '<p style="color: var(--text-light); font-size: 0.875rem; padding: 0.5rem 0;">No additional keywords available</p>';
        return;
    }

    cloudElement.innerHTML = keywords.map(([keyword, count]) => `
        <div class="keyword-tag" data-keyword="${escapeHtml(keyword)}" tabindex="0" role="button" aria-label="Filter by keyword: ${escapeHtml(keyword)}">
            <span>${escapeHtml(keyword)}</span>
            <span class="keyword-count">${count}</span>
        </div>
    `).join('');

    // Add click and keyboard handlers
    cloudElement.querySelectorAll('.keyword-tag').forEach(tag => {
        const keyword = tag.dataset.keyword;

        tag.addEventListener('click', () => {
            onKeywordClick(keyword);
        });

        tag.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                onKeywordClick(keyword);
            } else if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
                e.preventDefault();
                const tags = Array.from(cloudElement.querySelectorAll('.keyword-tag'));
                const currentIndex = tags.indexOf(tag);
                const nextTag = tags[currentIndex + 1];
                if (nextTag) nextTag.focus();
            } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
                e.preventDefault();
                const tags = Array.from(cloudElement.querySelectorAll('.keyword-tag'));
                const currentIndex = tags.indexOf(tag);
                const prevTag = tags[currentIndex - 1];
                if (prevTag) prevTag.focus();
            }
        });
    });
}

/**
 * Render selected keywords
 * @param {Set} selectedKeywords - Currently selected keywords
 * @param {HTMLElement} containerElement - Container element
 * @param {Function} onRemoveClick - Click handler for removal
 */
export function renderSelectedKeywords(selectedKeywords, containerElement, onRemoveClick) {
    if (selectedKeywords.size === 0) {
        containerElement.innerHTML = '';
        return;
    }

    containerElement.innerHTML = Array.from(selectedKeywords).map(keyword => `
        <div class="selected-keyword" data-keyword="${escapeHtml(keyword)}" tabindex="0" role="button" aria-label="Remove keyword filter: ${escapeHtml(keyword)}">
            <span>${escapeHtml(keyword)}</span>
            <span class="remove">×</span>
        </div>
    `).join('');

    // Add click and keyboard handlers for removal
    containerElement.querySelectorAll('.selected-keyword').forEach(tag => {
        const keyword = tag.dataset.keyword;

        tag.addEventListener('click', () => {
            onRemoveClick(keyword);
        });

        tag.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ' || e.key === 'Delete' || e.key === 'Backspace') {
                e.preventDefault();
                onRemoveClick(keyword);
            } else if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
                e.preventDefault();
                const tags = Array.from(containerElement.querySelectorAll('.selected-keyword'));
                const currentIndex = tags.indexOf(tag);
                const nextTag = tags[currentIndex + 1];
                if (nextTag) nextTag.focus();
            } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
                e.preventDefault();
                const tags = Array.from(containerElement.querySelectorAll('.selected-keyword'));
                const currentIndex = tags.indexOf(tag);
                const prevTag = tags[currentIndex - 1];
                if (prevTag) prevTag.focus();
            }
        });
    });
}

/**
 * Render category list
 * @param {Array} filteredTemplates - Currently filtered templates
 * @param {string} selectedCategory - Currently selected category
 * @param {HTMLElement} listElement - List container element
 * @param {Function} onCategoryClick - Click handler for categories
 */
export function renderCategoryList(filteredTemplates, selectedCategory, listElement, onCategoryClick) {
    // Get categories from currently filtered templates
    const categories = getCategoryCounts(filteredTemplates);

    if (categories.length === 0) {
        listElement.innerHTML = '<p style="color: var(--text-light); font-size: 0.875rem; padding: 0.5rem 0;">No categories available</p>';
        return;
    }

    listElement.innerHTML = categories.map(([category, count]) => {
        const isSelected = category === selectedCategory;
        return `
            <div class="category-item ${isSelected ? 'selected' : ''}" data-category="${escapeHtml(category)}" tabindex="0" role="button" aria-label="Filter by category: ${formatCategoryName(category)}" aria-pressed="${isSelected}">
                <span class="category-name">${formatCategoryName(category)}</span>
                <span class="category-count">${count}</span>
            </div>
        `;
    }).join('');

    // Add click and keyboard handlers
    listElement.querySelectorAll('.category-item').forEach(item => {
        const category = item.dataset.category;

        item.addEventListener('click', () => {
            onCategoryClick(category);
        });

        item.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                onCategoryClick(category);
            } else if (e.key === 'ArrowDown') {
                e.preventDefault();
                const items = Array.from(listElement.querySelectorAll('.category-item'));
                const currentIndex = items.indexOf(item);
                const nextItem = items[currentIndex + 1];
                if (nextItem) nextItem.focus();
            } else if (e.key === 'ArrowUp') {
                e.preventDefault();
                const items = Array.from(listElement.querySelectorAll('.category-item'));
                const currentIndex = items.indexOf(item);
                const prevItem = items[currentIndex - 1];
                if (prevItem) prevItem.focus();
            }
        });
    });
}

/**
 * Update sidebar with current state
 * @param {Object} state - Current application state
 * @param {Function} onKeywordToggle - Keyword toggle handler
 * @param {Function} onCategoryToggle - Category toggle handler
 */
export function updateSidebar(state, onKeywordToggle, onCategoryToggle) {
    const selectedKeywordsEl = document.getElementById('selected-keywords');
    const keywordCloudEl = document.getElementById('keyword-cloud');
    const categoryListEl = document.getElementById('category-list');

    renderSelectedKeywords(state.selectedKeywords, selectedKeywordsEl, onKeywordToggle);
    renderKeywordCloud(state.filteredTemplates, state.selectedKeywords, keywordCloudEl, onKeywordToggle);
    renderCategoryList(state.filteredTemplates, state.selectedCategory, categoryListEl, onCategoryToggle);
}
