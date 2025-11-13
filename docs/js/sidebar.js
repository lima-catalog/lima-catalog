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
export function renderKeywordCloud(filteredTemplates, selectedKeywords, cloudElement, onKeywordClick, shouldFocusFirst = false) {
    // Store currently focused keyword before re-rendering
    const focusedKeyword = document.activeElement?.dataset?.keyword;
    const isFocusedInCloud = document.activeElement?.classList?.contains('keyword-tag');

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
    let firstTag = null;
    cloudElement.querySelectorAll('.keyword-tag').forEach((tag, index) => {
        if (index === 0) firstTag = tag;
        const keyword = tag.dataset.keyword;

        tag.addEventListener('click', () => {
            onKeywordClick(keyword);
        });

        tag.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                onKeywordClick(keyword);
            } else if (e.key === 'ArrowRight') {
                e.preventDefault();
                const tags = Array.from(cloudElement.querySelectorAll('.keyword-tag'));
                const currentIndex = tags.indexOf(tag);
                const nextTag = tags[currentIndex + 1];
                if (nextTag) nextTag.focus();
            } else if (e.key === 'ArrowLeft') {
                e.preventDefault();
                const tags = Array.from(cloudElement.querySelectorAll('.keyword-tag'));
                const currentIndex = tags.indexOf(tag);
                const prevTag = tags[currentIndex - 1];
                if (prevTag) {
                    prevTag.focus();
                } else {
                    // We're at the first unselected keyword, jump to last selected keyword
                    const selectedTags = document.querySelectorAll('.selected-keyword');
                    if (selectedTags.length > 0) {
                        selectedTags[selectedTags.length - 1].focus();
                    }
                }
            } else if (e.key === 'ArrowDown') {
                e.preventDefault();
                const tags = Array.from(cloudElement.querySelectorAll('.keyword-tag'));
                const currentTop = tag.offsetTop;

                // Find the first tag on the next row (offsetTop > currentTop)
                const nextRowTag = tags.find(t => t.offsetTop > currentTop);
                if (nextRowTag) nextRowTag.focus();
            } else if (e.key === 'ArrowUp') {
                e.preventDefault();
                const tags = Array.from(cloudElement.querySelectorAll('.keyword-tag'));
                const currentTop = tag.offsetTop;

                // Find all tags on previous rows (offsetTop < currentTop)
                const tagsAbove = tags.filter(t => t.offsetTop < currentTop);
                if (tagsAbove.length > 0) {
                    // Find the closest row above (maximum offsetTop of tags above)
                    const closestRowTop = Math.max(...tagsAbove.map(t => t.offsetTop));
                    // Find the first tag on that row
                    const prevRowTag = tags.find(t => t.offsetTop === closestRowTop);
                    if (prevRowTag) prevRowTag.focus();
                } else {
                    // No previous row in unselected, try to go to last row of selected keywords
                    const selectedTags = Array.from(document.querySelectorAll('.selected-keyword'));
                    if (selectedTags.length > 0) {
                        // Find the last row of selected keywords
                        const lastSelectedTop = Math.max(...selectedTags.map(t => t.offsetTop));
                        const lastRowTags = selectedTags.filter(t => t.offsetTop === lastSelectedTop);
                        if (lastRowTags.length > 0) {
                            lastRowTags[0].focus(); // Focus first tag on last row
                        }
                    }
                }
            }
        });

        // Restore focus if this was the focused keyword (and it was in the cloud, not selected keywords)
        if (focusedKeyword && isFocusedInCloud && keyword === focusedKeyword) {
            // Use setTimeout to ensure DOM is ready
            setTimeout(() => tag.focus(), 0);
        }
    });

    // Focus first keyword if requested (e.g., after selecting a keyword)
    if (shouldFocusFirst && firstTag) {
        setTimeout(() => firstTag.focus(), 0);
    }
}

/**
 * Render selected keywords
 * @param {Set} selectedKeywords - Currently selected keywords
 * @param {HTMLElement} containerElement - Container element
 * @param {Function} onRemoveClick - Click handler for removal
 * @param {boolean} focusFirstUnselected - Whether to focus first unselected keyword after render
 * @param {number} focusIndex - Index of selected keyword to focus (-1 for last)
 */
export function renderSelectedKeywords(selectedKeywords, containerElement, onRemoveClick, focusFirstUnselected = false, focusIndex = null) {
    // Store currently focused keyword before re-rendering
    const focusedKeyword = document.activeElement?.dataset?.keyword;
    const isFocusedInSelected = document.activeElement?.classList?.contains('selected-keyword');

    // If we're deselecting a keyword, find its index for focus management
    let deselectedIndex = null;
    if (focusedKeyword && isFocusedInSelected && !selectedKeywords.has(focusedKeyword)) {
        // The focused keyword was just deselected, find where it was
        const currentTags = Array.from(containerElement.querySelectorAll('.selected-keyword'));
        deselectedIndex = currentTags.findIndex(tag => tag.dataset.keyword === focusedKeyword);
    }

    if (selectedKeywords.size === 0) {
        containerElement.innerHTML = '';
        // If we should focus first unselected keyword (e.g., after removing last selected)
        if (focusFirstUnselected) {
            setTimeout(() => {
                const firstKeyword = document.querySelector('.keyword-tag');
                if (firstKeyword) firstKeyword.focus();
            }, 0);
        }
        return;
    }

    containerElement.innerHTML = Array.from(selectedKeywords).map(keyword => `
        <div class="selected-keyword" data-keyword="${escapeHtml(keyword)}" tabindex="0" role="button" aria-label="Remove keyword filter: ${escapeHtml(keyword)}">
            <span>${escapeHtml(keyword)}</span>
            <span class="remove">×</span>
        </div>
    `).join('');

    // Add click and keyboard handlers for removal
    const newTags = Array.from(containerElement.querySelectorAll('.selected-keyword'));
    newTags.forEach((tag, index, allTags) => {
        const keyword = tag.dataset.keyword;
        const isLastSelected = index === allTags.length - 1;

        tag.addEventListener('click', () => {
            onRemoveClick(keyword);
        });

        tag.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ' || e.key === 'Delete' || e.key === 'Backspace') {
                e.preventDefault();
                onRemoveClick(keyword);
            } else if (e.key === 'ArrowRight') {
                e.preventDefault();
                const tags = Array.from(containerElement.querySelectorAll('.selected-keyword'));
                const currentIndex = tags.indexOf(tag);
                const nextTag = tags[currentIndex + 1];
                if (nextTag) {
                    nextTag.focus();
                } else {
                    // We're at the last selected keyword, jump to first unselected keyword
                    const firstUnselected = document.querySelector('.keyword-tag');
                    if (firstUnselected) firstUnselected.focus();
                }
            } else if (e.key === 'ArrowLeft') {
                e.preventDefault();
                const tags = Array.from(containerElement.querySelectorAll('.selected-keyword'));
                const currentIndex = tags.indexOf(tag);
                const prevTag = tags[currentIndex - 1];
                if (prevTag) prevTag.focus();
            } else if (e.key === 'ArrowDown') {
                e.preventDefault();
                const tags = Array.from(containerElement.querySelectorAll('.selected-keyword'));
                const currentTop = tag.offsetTop;

                // Find the first tag on the next row (offsetTop > currentTop)
                const nextRowTag = tags.find(t => t.offsetTop > currentTop);
                if (nextRowTag) {
                    nextRowTag.focus();
                } else {
                    // No next row in selected, try to go to unselected keywords
                    const firstUnselected = document.querySelector('.keyword-tag');
                    if (firstUnselected) firstUnselected.focus();
                }
            } else if (e.key === 'ArrowUp') {
                e.preventDefault();
                const tags = Array.from(containerElement.querySelectorAll('.selected-keyword'));
                const currentTop = tag.offsetTop;

                // Find all tags on previous rows (offsetTop < currentTop)
                const tagsAbove = tags.filter(t => t.offsetTop < currentTop);
                if (tagsAbove.length > 0) {
                    // Find the closest row above (maximum offsetTop of tags above)
                    const closestRowTop = Math.max(...tagsAbove.map(t => t.offsetTop));
                    // Find the first tag on that row
                    const prevRowTag = tags.find(t => t.offsetTop === closestRowTop);
                    if (prevRowTag) prevRowTag.focus();
                }
            }
        });

        // Focus management after deselection
        if (deselectedIndex !== null) {
            // A keyword was just deselected, focus the next one or last one
            if (index === deselectedIndex && newTags[index]) {
                // Focus the keyword now at the deselected index (the next one)
                setTimeout(() => tag.focus(), 0);
            } else if (index === newTags.length - 1 && deselectedIndex >= newTags.length) {
                // The last keyword was deselected, focus the new last keyword
                setTimeout(() => tag.focus(), 0);
            }
        }
        // Or restore focus if this was the focused keyword (and it's still selected)
        else if (focusedKeyword && isFocusedInSelected && keyword === focusedKeyword) {
            // Use setTimeout to ensure DOM is ready
            setTimeout(() => tag.focus(), 0);
        }
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
    // Store currently focused category before re-rendering
    const focusedCategory = document.activeElement?.dataset?.category;

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

        // Restore focus if this was the focused category
        if (focusedCategory && category === focusedCategory) {
            // Use setTimeout to ensure DOM is ready
            setTimeout(() => item.focus(), 0);
        }
    });
}

/**
 * Update sidebar with current state
 * @param {Object} state - Current application state
 * @param {Function} onKeywordToggle - Keyword toggle handler
 * @param {Function} onCategoryToggle - Category toggle handler
 */
export function updateSidebar(state, onKeywordToggle, onCategoryToggle, options = {}) {
    const selectedKeywordsEl = document.getElementById('selected-keywords');
    const keywordCloudEl = document.getElementById('keyword-cloud');
    const categoryListEl = document.getElementById('category-list');

    renderSelectedKeywords(state.selectedKeywords, selectedKeywordsEl, onKeywordToggle, options.focusFirstUnselected);
    renderKeywordCloud(state.filteredTemplates, state.selectedKeywords, keywordCloudEl, onKeywordToggle, options.focusFirstKeyword);
    renderCategoryList(state.filteredTemplates, state.selectedCategory, categoryListEl, onCategoryToggle);
}
