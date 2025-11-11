// Configuration
const DATA_BASE_URL = 'https://raw.githubusercontent.com/lima-catalog/lima-catalog/data/data';

// State
let templates = [];
let repositories = new Map();
let filteredTemplates = [];

// Initialize
document.addEventListener('DOMContentLoaded', async () => {
    await loadData();
    setupEventListeners();
    renderTemplates();
});

// Load data from GitHub
async function loadData() {
    const loading = document.getElementById('loading');
    const error = document.getElementById('error');

    try {
        loading.textContent = 'Loading templates...';

        // Load templates
        const templatesResponse = await fetch(`${DATA_BASE_URL}/templates.jsonl`);
        if (!templatesResponse.ok) throw new Error('Failed to load templates');
        const templatesText = await templatesResponse.text();
        templates = parseJsonLines(templatesText);

        loading.textContent = 'Loading repository data...';

        // Load repositories
        const reposResponse = await fetch(`${DATA_BASE_URL}/repos.jsonl`);
        if (!reposResponse.ok) throw new Error('Failed to load repositories');
        const reposText = await reposResponse.text();
        const repos = parseJsonLines(reposText);
        repos.forEach(repo => repositories.set(repo.id, repo));

        filteredTemplates = [...templates];

        // Populate category filter
        populateCategoryFilter();

        // Update stats
        updateStats();

        loading.style.display = 'none';
    } catch (err) {
        console.error('Error loading data:', err);
        loading.style.display = 'none';
        error.style.display = 'block';
        error.textContent = `Error loading catalog data: ${err.message}`;
    }
}

// Parse JSON Lines format
function parseJsonLines(text) {
    return text
        .trim()
        .split('\n')
        .filter(line => line.trim())
        .map(line => JSON.parse(line));
}

// Populate category filter dropdown
function populateCategoryFilter() {
    const categories = new Set();
    templates.forEach(t => {
        if (t.category) categories.add(t.category);
    });

    const select = document.getElementById('category-filter');
    Array.from(categories)
        .sort()
        .forEach(cat => {
            const option = document.createElement('option');
            option.value = cat;
            option.textContent = cat.charAt(0).toUpperCase() + cat.slice(1);
            select.appendChild(option);
        });
}

// Setup event listeners
function setupEventListeners() {
    document.getElementById('search').addEventListener('input', filterAndRender);
    document.getElementById('category-filter').addEventListener('change', filterAndRender);
    document.getElementById('type-filter').addEventListener('change', filterAndRender);
    document.getElementById('sort').addEventListener('change', filterAndRender);
}

// Filter and render templates
function filterAndRender() {
    const searchTerm = document.getElementById('search').value.toLowerCase();
    const categoryFilter = document.getElementById('category-filter').value;
    const typeFilter = document.getElementById('type-filter').value;

    // Filter templates
    filteredTemplates = templates.filter(template => {
        // Search filter
        if (searchTerm) {
            const searchText = [
                template.name,
                template.display_name,
                template.short_description,
                template.category,
                ...(template.keywords || []),
                ...(template.images || [])
            ].join(' ').toLowerCase();

            if (!searchText.includes(searchTerm)) return false;
        }

        // Category filter
        if (categoryFilter && template.category !== categoryFilter) return false;

        // Type filter
        if (typeFilter === 'official' && !template.is_official) return false;
        if (typeFilter === 'community' && template.is_official) return false;

        return true;
    });

    // Sort templates
    const sortBy = document.getElementById('sort').value;
    filteredTemplates.sort((a, b) => {
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

    updateStats();
    renderTemplates();
}

// Update statistics
function updateStats() {
    const official = templates.filter(t => t.is_official).length;
    const community = templates.filter(t => !t.is_official).length;

    document.getElementById('total-count').textContent = templates.length;
    document.getElementById('official-count').textContent = official;
    document.getElementById('community-count').textContent = community;
    document.getElementById('visible-count').textContent = filteredTemplates.length;
}

// Render templates
function renderTemplates() {
    const grid = document.getElementById('templates-grid');
    grid.innerHTML = '';

    if (filteredTemplates.length === 0) {
        grid.innerHTML = '<p style="grid-column: 1/-1; text-align: center; padding: 3rem; color: var(--text-light);">No templates found matching your criteria.</p>';
        return;
    }

    filteredTemplates.forEach(template => {
        const card = createTemplateCard(template);
        grid.appendChild(card);
    });
}

// Create template card
function createTemplateCard(template) {
    const card = document.createElement('div');
    card.className = 'template-card';

    const repo = repositories.get(template.repo);
    const displayName = template.display_name || template.name || template.path;
    const description = template.short_description || (repo?.description || 'No description available');

    card.innerHTML = `
        <div class="template-header">
            <div class="template-title">
                <h3 class="template-name">${escapeHtml(displayName)}</h3>
                <div class="template-id">${escapeHtml(template.id)}</div>
            </div>
            <span class="template-badge ${template.is_official ? 'official' : 'community'}">
                ${template.is_official ? 'Official' : 'Community'}
            </span>
        </div>

        <p class="template-description">${escapeHtml(description)}</p>

        <div class="template-meta">
            ${template.category ? `
                <span class="template-category">
                    📦 ${escapeHtml(template.category)}
                </span>
            ` : ''}
            ${template.images && template.images.length > 0 ? `
                <span class="template-os">
                    💿 ${escapeHtml(template.images[0])}
                </span>
            ` : ''}
        </div>

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

    // Make card clickable
    card.style.cursor = 'pointer';
    card.addEventListener('click', (e) => {
        // Don't navigate if clicking on a link
        if (e.target.tagName === 'A') return;
        window.open(template.url, '_blank');
    });

    return card;
}

// Escape HTML to prevent XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
