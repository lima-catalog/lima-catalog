// Configuration
const DATA_BASE_URL = 'https://raw.githubusercontent.com/lima-catalog/lima-catalog/data/data';
const MAX_KEYWORDS_DISPLAY = 50; // Show top 50 keywords in cloud

// State
let templates = [];
let repositories = new Map();
let filteredTemplates = [];
let selectedKeywords = new Set();
let selectedCategory = null;

// Initialize
document.addEventListener('DOMContentLoaded', async () => {
    await loadData();
    setupEventListeners();
    renderKeywordCloud();
    renderCategoryList();
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

// Get all keywords with counts from a specific template list
function getKeywordCounts(templateList) {
    const counts = new Map();
    templateList.forEach(template => {
        if (template.keywords && Array.isArray(template.keywords)) {
            template.keywords.forEach(kw => {
                // Don't include already selected keywords in the cloud
                if (!selectedKeywords.has(kw)) {
                    counts.set(kw, (counts.get(kw) || 0) + 1);
                }
            });
        }
    });
    return Array.from(counts.entries())
        .sort((a, b) => b[1] - a[1]) // Sort by count descending
        .slice(0, MAX_KEYWORDS_DISPLAY);
}

// Get category counts from a specific template list
function getCategoryCounts(templateList) {
    const counts = new Map();
    templateList.forEach(template => {
        if (template.category) {
            counts.set(template.category, (counts.get(template.category) || 0) + 1);
        }
    });
    return Array.from(counts.entries())
        .sort((a, b) => a[0].localeCompare(b[0])); // Sort alphabetically
}

// Render keyword cloud based on current filter state
function renderKeywordCloud() {
    const cloud = document.getElementById('keyword-cloud');

    // Get keywords from currently filtered templates
    // This shows what keywords are available in the current filtered selection
    const keywords = getKeywordCounts(filteredTemplates);

    if (keywords.length === 0) {
        cloud.innerHTML = '<p style="color: var(--text-light); font-size: 0.875rem; padding: 0.5rem 0;">No additional keywords available</p>';
        return;
    }

    cloud.innerHTML = keywords.map(([keyword, count]) => `
        <div class="keyword-tag" data-keyword="${escapeHtml(keyword)}">
            <span>${escapeHtml(keyword)}</span>
            <span class="keyword-count">${count}</span>
        </div>
    `).join('');

    // Add click handlers
    cloud.querySelectorAll('.keyword-tag').forEach(tag => {
        tag.addEventListener('click', () => {
            const keyword = tag.dataset.keyword;
            toggleKeyword(keyword);
        });
    });
}

// Render category list based on current filter state
function renderCategoryList() {
    const list = document.getElementById('category-list');

    // Get categories from currently filtered templates
    const categories = getCategoryCounts(filteredTemplates);

    if (categories.length === 0) {
        list.innerHTML = '<p style="color: var(--text-light); font-size: 0.875rem; padding: 0.5rem 0;">No categories available</p>';
        return;
    }

    list.innerHTML = categories.map(([category, count]) => `
        <div class="category-item" data-category="${escapeHtml(category)}">
            <span class="category-name">${formatCategoryName(category)}</span>
            <span class="category-count">${count}</span>
        </div>
    `).join('');

    // Add click handlers
    list.querySelectorAll('.category-item').forEach(item => {
        item.addEventListener('click', () => {
            const category = item.dataset.category;
            toggleCategory(category);
        });
    });
}

// Toggle keyword selection
function toggleKeyword(keyword) {
    if (selectedKeywords.has(keyword)) {
        selectedKeywords.delete(keyword);
    } else {
        selectedKeywords.add(keyword);
    }
    updateSelectedKeywords();
    filterAndRender();
}

// Toggle category selection
function toggleCategory(category) {
    if (selectedCategory === category) {
        selectedCategory = null;
    } else {
        selectedCategory = category;
    }
    updateCategorySelection();
    filterAndRender();
}

// Update selected keywords display
function updateSelectedKeywords() {
    const container = document.getElementById('selected-keywords');

    if (selectedKeywords.size === 0) {
        container.innerHTML = '';
        return;
    }

    container.innerHTML = Array.from(selectedKeywords).map(keyword => `
        <div class="selected-keyword" data-keyword="${escapeHtml(keyword)}">
            <span>${escapeHtml(keyword)}</span>
            <span class="remove">×</span>
        </div>
    `).join('');

    // Add click handlers for removal
    container.querySelectorAll('.selected-keyword').forEach(tag => {
        tag.addEventListener('click', () => {
            const keyword = tag.dataset.keyword;
            toggleKeyword(keyword);
        });
    });

    // Update keyword cloud selected state
    document.querySelectorAll('.keyword-tag').forEach(tag => {
        if (selectedKeywords.has(tag.dataset.keyword)) {
            tag.classList.add('selected');
        } else {
            tag.classList.remove('selected');
        }
    });
}

// Update category selection display
function updateCategorySelection() {
    document.querySelectorAll('.category-item').forEach(item => {
        if (item.dataset.category === selectedCategory) {
            item.classList.add('selected');
        } else {
            item.classList.remove('selected');
        }
    });
}

// Setup event listeners
function setupEventListeners() {
    document.getElementById('search').addEventListener('input', filterAndRender);
    document.getElementById('type-filter').addEventListener('change', filterAndRender);
    document.getElementById('sort').addEventListener('change', filterAndRender);
    document.getElementById('clear-filters').addEventListener('click', clearAllFilters);
}

// Clear all filters
function clearAllFilters() {
    document.getElementById('search').value = '';
    document.getElementById('type-filter').value = '';
    selectedKeywords.clear();
    selectedCategory = null;
    updateSelectedKeywords();
    updateCategorySelection();
    filterAndRender();
}

// Filter and render templates
function filterAndRender() {
    const searchTerm = document.getElementById('search').value.toLowerCase();
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
    renderKeywordCloud(); // Update keyword cloud based on filtered templates
    renderCategoryList(); // Update category list based on filtered templates
    updateCategorySelection(); // Restore category selection state
    renderTemplates();
}

// Update statistics
function updateStats() {
    const official = templates.filter(t => t.is_official).length;
    const community = templates.filter(t => !t.is_official).length;

    document.getElementById('total-count').textContent = templates.length;
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

// Derive a nice display name from path if needed
function deriveDisplayName(template) {
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

// Format a name to be more readable
function formatName(name) {
    return name
        .replace(/[-_]/g, ' ')
        .split(' ')
        .map(word => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ');
}

// Format category name
function formatCategoryName(category) {
    return category
        .split('-')
        .map(word => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ');
}

// Create template card
function createTemplateCard(template) {
    const card = document.createElement('div');
    card.className = 'template-card';

    const repo = repositories.get(template.repo);
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
