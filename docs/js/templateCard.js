/**
 * Template card rendering utilities
 */

import { isDebugMode } from './state.js';
import * as State from './state.js';
import { filterAndRender } from './appActions.js';

/**
 * Escape HTML to prevent XSS
 * @param {string} text - Text to escape
 * @returns {string} Escaped HTML
 */
export function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

/**
 * Format a name to be more readable
 * @param {string} name - Name to format
 * @returns {string} Formatted name
 */
export function formatName(name) {
    return name
        .replace(/[-_]/g, ' ')
        .split(' ')
        .map(word => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ');
}

/**
 * Format category name
 * @param {string} category - Category name
 * @returns {string} Formatted category name
 */
export function formatCategoryName(category) {
    return category
        .split('-')
        .map(word => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ');
}

/**
 * Get badge text for debug mode
 * @param {Object} template - Template object
 * @param {string} sortBy - Current sort field
 * @returns {Object} Badge info with text and label, or null
 */
function getDebugBadgeText(template, sortBy) {
    if (!template.notability_score_breakdown) {
        return null;
    }

    // Map sort field to breakdown component
    const breakdownMap = {
        'breakdown-message': { value: template.notability_score_breakdown.message, label: 'Message' },
        'breakdown-provision': { value: template.notability_score_breakdown.provision, label: 'Provision' },
        'breakdown-parameters': { value: template.notability_score_breakdown.parameters, label: 'Parameters' },
        'breakdown-env_vars': { value: template.notability_score_breakdown.env_vars, label: 'Env Vars' },
        'breakdown-probes': { value: template.notability_score_breakdown.probes, label: 'Probes' },
        'breakdown-unusual_images': { value: template.notability_score_breakdown.unusual_images, label: 'Unusual Images' },
        'breakdown-comments': { value: template.notability_score_breakdown.comments, label: 'Comments' },
        'breakdown-stars': { value: template.notability_score_breakdown.stars, label: 'Stars' }
    };

    // If sorting by a specific breakdown component, show that score
    if (breakdownMap[sortBy]) {
        const { value, label } = breakdownMap[sortBy];
        return {
            text: `${value.toFixed(1)}`,
            label: label
        };
    }

    // Default: show total score
    const score = template.notability_score || 0;
    return {
        text: `${score.toFixed(1)}`,
        label: 'Total'
    };
}

/**
 * Create debug score popup element
 * @param {Object} template - Template object
 * @returns {HTMLElement|null} Popup element or null
 */
function createDebugScorePopup(template) {
    if (!template.notability_score_breakdown) {
        return null;
    }

    const breakdown = template.notability_score_breakdown;

    const popup = document.createElement('div');
    popup.className = 'debug-score-popup';
    popup.innerHTML = `
        <div class="debug-popup-title">Notability Score Breakdown</div>
        <div class="debug-popup-items">
            <div class="debug-popup-item">
                <span class="debug-popup-label">Message:</span>
                <span class="debug-popup-value">${breakdown.message.toFixed(1)}</span>
            </div>
            <div class="debug-popup-item">
                <span class="debug-popup-label">Provision:</span>
                <span class="debug-popup-value">${breakdown.provision.toFixed(1)}</span>
            </div>
            <div class="debug-popup-item">
                <span class="debug-popup-label">Parameters:</span>
                <span class="debug-popup-value">${breakdown.parameters.toFixed(1)}</span>
            </div>
            <div class="debug-popup-item">
                <span class="debug-popup-label">Env Vars:</span>
                <span class="debug-popup-value">${breakdown.env_vars.toFixed(1)}</span>
            </div>
            <div class="debug-popup-item">
                <span class="debug-popup-label">Probes:</span>
                <span class="debug-popup-value">${breakdown.probes.toFixed(1)}</span>
            </div>
            <div class="debug-popup-item">
                <span class="debug-popup-label">Unusual Images:</span>
                <span class="debug-popup-value">${breakdown.unusual_images.toFixed(1)}</span>
            </div>
            <div class="debug-popup-item">
                <span class="debug-popup-label">Comments:</span>
                <span class="debug-popup-value">${breakdown.comments.toFixed(1)}</span>
            </div>
            <div class="debug-popup-item">
                <span class="debug-popup-label">Stars:</span>
                <span class="debug-popup-value">${breakdown.stars.toFixed(1)}</span>
            </div>
            <div class="debug-popup-divider"></div>
            <div class="debug-popup-item debug-popup-total">
                <span class="debug-popup-label">Total:</span>
                <span class="debug-popup-value">${breakdown.total.toFixed(1)}</span>
            </div>
        </div>
    `;

    return popup;
}

/**
 * Derive a nice display name from template
 * @param {Object} template - Template object
 * @returns {string} Display name
 */
export function deriveDisplayName(template) {
    // If we have a proper display_name, use it
    if (template.display_name) return template.display_name;
    if (template.name) return template.name;

    // Otherwise, derive from path
    const path = template.path || '';
    const filename = path.split('/').pop() || '';
    const nameWithoutExt = filename.replace(/\.(yaml|yml)$/, '');

    // Check if filename is generic
    const genericNames = ['lima', 'template', 'config', 'default'];
    if (genericNames.includes(nameWithoutExt.toLowerCase())) {
        // Use parent directory name
        const parts = path.split('/');
        if (parts.length > 1) {
            const parent = parts[parts.length - 2];
            return formatName(parent);
        }
        // Fall back to repo name
        const repoName = template.repo.split('/').pop();
        return formatName(repoName);
    }

    return formatName(nameWithoutExt);
}

/**
 * Create template card DOM element
 * @param {Object} template - Template object
 * @param {Function} onCardClick - Click handler for card
 * @param {string} sortBy - Current sort field (optional)
 * @returns {HTMLElement} Card element
 */
export function createTemplateCard(template, onCardClick, sortBy = 'name') {
    const card = document.createElement('div');
    card.className = 'template-card';
    card.setAttribute('tabindex', '0');
    card.setAttribute('role', 'article');
    card.setAttribute('aria-label', `Template: ${deriveDisplayName(template)}`);

    const displayName = deriveDisplayName(template);
    const description = template.description || 'No description available';

    // In debug mode, show score instead of Official/Community
    const debugBadgeInfo = isDebugMode() ? getDebugBadgeText(template, sortBy) : null;
    const badgeText = debugBadgeInfo ? debugBadgeInfo.text : (template.official ? 'Official' : 'Community');
    const badgeClass = template.official ? 'official' : 'community';
    const badgeTitle = debugBadgeInfo ? `${debugBadgeInfo.label} Score (hover for breakdown)` : '';

    card.innerHTML = `
        <div class="template-header">
            <div class="template-title">
                <h3 class="template-name">${escapeHtml(displayName)}</h3>
                <div class="template-id">${escapeHtml(template.path)}</div>
            </div>
            <span class="template-badge ${badgeClass}" ${badgeTitle ? `title="${badgeTitle}"` : ''}>
                ${escapeHtml(badgeText)}
            </span>
        </div>

        <p class="template-description">${escapeHtml(description)}</p>

        ${template.category || (template.images && template.images.length > 0) ? `
        <div class="template-meta">
            ${template.category ? `
                <span class="template-category">
                    📦 ${escapeHtml(formatCategoryName(template.category))}
                </span>
            ` : ''}
            ${template.images && template.images.length > 0 ? `
                <span class="template-os">
                    💿 ${escapeHtml(template.images[0])}
                </span>
            ` : ''}
        </div>
        ` : ''}

        ${template.keywords && template.keywords.length > 0 ? `
            <div class="template-keywords">
                ${template.keywords.slice(0, 6).map(kw =>
                    `<span class="keyword">${escapeHtml(kw)}</span>`
                ).join('')}
            </div>
        ` : ''}

        <div class="template-footer">
            <a href="https://github.com/${escapeHtml(template.repo)}"
               target="_blank"
               class="template-repo">
                📁 ${escapeHtml(template.repo)}
            </a>
            ${template.stars > 0 ? `
                <span class="template-stars">
                    ⭐ ${template.stars}
                </span>
            ` : ''}
        </div>
    `;

    // Attach debug popup to badge if in debug mode
    if (debugBadgeInfo) {
        const popup = createDebugScorePopup(template);
        if (popup) {
            const badge = card.querySelector('.template-badge');
            badge.style.position = 'relative';
            badge.style.cursor = 'help';
            badge.appendChild(popup);

            // Show/hide popup on hover
            badge.addEventListener('mouseenter', () => {
                popup.style.display = 'block';
            });
            badge.addEventListener('mouseleave', () => {
                popup.style.display = 'none';
            });
        }
    }

    // Make card clickable - open preview modal
    card.style.cursor = 'pointer';

    const handleOpen = (e) => {
        // Don't open modal if clicking on a link (repo link should open GitHub)
        if (e.target.tagName === 'A' || e.target.closest('a')) return;
        onCardClick(template);
    };

    card.addEventListener('click', handleOpen);

    // Add keyboard support
    card.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' || e.key === ' ') {
            // Only prevent default if not on a link
            if (e.target.tagName !== 'A' && !e.target.closest('a')) {
                e.preventDefault();
                onCardClick(template);
            }
        } else if (e.key === 'ArrowRight') {
            e.preventDefault();
            const grid = card.parentElement;
            const cards = Array.from(grid.querySelectorAll('.template-card'));
            const currentIndex = cards.indexOf(card);
            const nextCard = cards[currentIndex + 1];
            if (nextCard) nextCard.focus();
        } else if (e.key === 'ArrowLeft') {
            e.preventDefault();
            const grid = card.parentElement;
            const cards = Array.from(grid.querySelectorAll('.template-card'));
            const currentIndex = cards.indexOf(card);
            const prevCard = cards[currentIndex - 1];
            if (prevCard) prevCard.focus();
        } else if (e.key === 'ArrowDown') {
            e.preventDefault();
            const grid = card.parentElement;
            const cards = Array.from(grid.querySelectorAll('.template-card'));
            const currentIndex = cards.indexOf(card);

            // Calculate number of columns in the grid
            const gridStyle = window.getComputedStyle(grid);
            const gridTemplateColumns = gridStyle.gridTemplateColumns;
            const columnCount = gridTemplateColumns.split(' ').length;

            // Move down by one row (columnCount cards)
            const nextCard = cards[currentIndex + columnCount];
            if (nextCard) {
                nextCard.focus({ preventScroll: true });

                // Check if next card is fully visible
                const rect = nextCard.getBoundingClientRect();
                const margin = 20; // Space for border and focus outline
                const isFullyVisible = rect.top >= margin && rect.bottom <= window.innerHeight - margin;

                // Only scroll if not fully visible
                if (!isFullyVisible) {
                    const cardTop = rect.top + window.scrollY;
                    window.scrollTo({ top: cardTop - margin, behavior: 'smooth' });
                }
            } else {
                // At bottom row - scroll to bottom to show footer
                window.scrollTo({ top: document.documentElement.scrollHeight, behavior: 'smooth' });
            }
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            const grid = card.parentElement;
            const cards = Array.from(grid.querySelectorAll('.template-card'));
            const currentIndex = cards.indexOf(card);

            // Calculate number of columns in the grid
            const gridStyle = window.getComputedStyle(grid);
            const gridTemplateColumns = gridStyle.gridTemplateColumns;
            const columnCount = gridTemplateColumns.split(' ').length;

            // Move up by one row (columnCount cards)
            const prevCard = cards[currentIndex - columnCount];
            if (prevCard) {
                prevCard.focus({ preventScroll: true });

                // Check if prev card is fully visible
                const rect = prevCard.getBoundingClientRect();
                const margin = 20; // Space for border and focus outline
                const isFullyVisible = rect.top >= margin && rect.bottom <= window.innerHeight - margin;

                // Only scroll if not fully visible
                if (!isFullyVisible) {
                    const cardTop = rect.top + window.scrollY;
                    const cardHeight = rect.height;
                    // Position card so its bottom is at (viewport bottom - margin)
                    const targetScrollY = cardTop + cardHeight - window.innerHeight + margin;
                    window.scrollTo({ top: Math.max(0, targetScrollY), behavior: 'smooth' });
                }
            } else {
                // At top row - scroll to top to show header
                window.scrollTo({ top: 0, behavior: 'smooth' });
            }
        }
    });

    // Update focused template on hover/focus to show dynamic keywords
    const updateFocusedTemplate = () => {
        console.log('[templateCard] Setting focused template:', template.id);
        State.setFocusedTemplate(template);
        filterAndRender();
    };

    const clearFocusedTemplate = () => {
        console.log('[templateCard] Clearing focused template');
        State.setFocusedTemplate(null);
        filterAndRender();
    };

    // On mouse hover
    card.addEventListener('mouseenter', updateFocusedTemplate);
    card.addEventListener('mouseleave', clearFocusedTemplate);

    // On keyboard focus
    card.addEventListener('focus', updateFocusedTemplate);
    card.addEventListener('blur', clearFocusedTemplate);

    return card;
}

/**
 * Render templates to grid
 * @param {Array} templates - Templates to render
 * @param {HTMLElement} gridElement - Grid container element
 * @param {Function} onCardClick - Click handler for cards
 * @param {string} sortBy - Current sort field (optional)
 */
export function renderTemplateGrid(templates, gridElement, onCardClick, sortBy = 'name') {
    gridElement.innerHTML = '';

    if (templates.length === 0) {
        gridElement.innerHTML = '<p style="grid-column: 1/-1; text-align: center; padding: 3rem; color: var(--text-light);">No templates found matching your criteria.</p>';
        return;
    }

    templates.forEach(template => {
        const card = createTemplateCard(template, onCardClick, sortBy);
        gridElement.appendChild(card);
    });
}
