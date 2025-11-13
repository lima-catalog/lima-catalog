/**
 * Template card rendering utilities
 */

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
 * @param {Object} repo - Repository object
 * @param {Function} onCardClick - Click handler for card
 * @returns {HTMLElement} Card element
 */
export function createTemplateCard(template, repo, onCardClick) {
    const card = document.createElement('div');
    card.className = 'template-card';
    card.setAttribute('tabindex', '0');
    card.setAttribute('role', 'article');
    card.setAttribute('aria-label', `Template: ${deriveDisplayName(template)}`);

    const displayName = deriveDisplayName(template);
    const description = template.short_description || (repo?.description || 'No description available');

    card.innerHTML = `
        <div class="template-header">
            <div class="template-title">
                <h3 class="template-name">${escapeHtml(displayName)}</h3>
                <div class="template-id">${escapeHtml(template.path)}</div>
            </div>
            <span class="template-badge ${template.is_official ? 'official' : 'community'}">
                ${template.is_official ? 'Official' : 'Community'}
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
            ${repo && repo.stars > 0 ? `
                <span class="template-stars">
                    ⭐ ${repo.stars}
                </span>
            ` : ''}
        </div>
    `;

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
        } else if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
            e.preventDefault();
            const grid = card.parentElement;
            const cards = Array.from(grid.querySelectorAll('.template-card'));
            const currentIndex = cards.indexOf(card);
            const nextCard = cards[currentIndex + 1];
            if (nextCard) nextCard.focus();
        } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
            e.preventDefault();
            const grid = card.parentElement;
            const cards = Array.from(grid.querySelectorAll('.template-card'));
            const currentIndex = cards.indexOf(card);
            const prevCard = cards[currentIndex - 1];
            if (prevCard) prevCard.focus();
        }
    });

    return card;
}

/**
 * Render templates to grid
 * @param {Array} templates - Templates to render
 * @param {Map} repositories - Repository data
 * @param {HTMLElement} gridElement - Grid container element
 * @param {Function} onCardClick - Click handler for cards
 */
export function renderTemplateGrid(templates, repositories, gridElement, onCardClick) {
    gridElement.innerHTML = '';

    if (templates.length === 0) {
        gridElement.innerHTML = '<p style="grid-column: 1/-1; text-align: center; padding: 3rem; color: var(--text-light);">No templates found matching your criteria.</p>';
        return;
    }

    templates.forEach(template => {
        const repo = repositories.get(template.repo);
        const card = createTemplateCard(template, repo, onCardClick);
        gridElement.appendChild(card);
    });
}
