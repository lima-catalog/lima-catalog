# Lima Template Catalog - Architecture

## Overview

A searchable catalog of 700+ Lima VM templates from across GitHub, with automated discovery, smart categorization, and a beautiful web interface.

**Live Site**: [lima-catalog.github.io/lima-catalog](https://lima-catalog.github.io/lima-catalog/)

## System Components

### 1. Backend (Go CLI Tool)

**Purpose**: Discover templates, collect metadata, analyze content

**Key Features**:
- GitHub Code Search with multiple query strategies
- Content-based validation to eliminate false positives
- Smart name derivation from paths
- YAML parsing and technology detection
- Automatic categorization and keyword extraction
- Incremental updates with SHA-based change detection

**Data Storage**:
- JSON Lines format (one object per line)
- Separate files for templates, repos, orgs
- Stored in `data` branch for isolation from code

**Schedule**: Runs daily via GitHub Actions

### 2. Frontend (Static GitHub Pages)

**Purpose**: Browse and search templates

**Key Features**:
- Multi-keyword filtering with AND logic
- Category browsing with dynamic counts
- Template preview modal with YAML syntax highlighting
- Lima 2.0 `github:` URL generation and copy
- Responsive design for mobile/tablet/desktop
- Fetches data directly from `data` branch

**Tech Stack**: Modular ES6 JavaScript + highlight.js (no build step)

**Local Development**:
- Uses native ES6 modules (requires web server)
- Cannot be opened directly as `file://` URLs
- Test locally with: `cd docs && python3 -m http.server 8000`
- Or use: `npx serve docs`

## Discovery Strategy

### Search Queries

Four queries to maximize template discovery:

1. `minimumLimaVersion extension:yaml -repo:lima-vm/lima`
2. `minimumLimaVersion extension:yml -repo:lima-vm/lima`
3. `images: provision: extension:yaml -repo:lima-vm/lima`
4. `images: provision: extension:yml -repo:lima-vm/lima`

### Content Validation

Downloads and validates each file to eliminate false positives:
- Must contain `images:` as top-level YAML key
- Filters out Kubernetes ConfigMaps, GitHub Actions, etc.
- ~31% of initial results are false positives

### Official Templates

Separately enumerates `lima-vm/lima/templates/` directory to get 51 official templates.

## Template Analysis

### Smart Naming

Handles generic filenames like `lima.yaml`:
- Uses repository name for context
- Falls back to parent directory path
- Generates human-readable display names

### Technology Detection

Scans provisioning scripts for:
- **Container runtimes**: Docker, Podman, containerd, nerdctl
- **Orchestration**: Kubernetes, K3s, K0s, K8sd, MicroK8s, Minikube
- **Databases**: PostgreSQL, MySQL, MongoDB, Redis, Elasticsearch
- **Languages**: Go, Rust, Python, Node.js, Java, Ruby
- **Tools**: Git, Make, curl, wget, jq, yq

### Categorization

Auto-assigns categories based on detected technologies:
- `containers` - Docker/Podman environments
- `orchestration` - Kubernetes setups
- `development` - Dev tools and languages
- `database` - Database servers
- `security` - Security/pentest tools
- `testing` - CI/CD and testing environments
- `general` - Everything else

## Data Schema

### Template

```json
{
  "id": "owner/repo/path/template.yaml",
  "repo": "owner/repo",
  "path": "path/template.yaml",
  "sha": "abc123...",
  "url": "https://github.com/.../blob/sha/path",
  "is_official": false,
  "discovered_at": "2025-01-15T10:00:00Z",
  "last_checked": "2025-01-20T12:00:00Z",
  "name": "template-name",
  "display_name": "Template Name",
  "short_description": "Ubuntu-based container runtime",
  "category": "containers",
  "keywords": ["ubuntu", "docker", "git"],
  "images": ["ubuntu"],
  "arch": ["x86_64", "aarch64"],
  "analyzed_at": "2025-01-15T10:05:00Z"
}
```

### Repository

```json
{
  "id": "owner/repo",
  "owner": "owner",
  "name": "repo",
  "description": "Repository description",
  "topics": ["lima", "kubernetes"],
  "stars": 42,
  "language": "Go",
  "default_branch": "main",
  "last_fetched": "2025-01-20T12:00:00Z"
}
```

### Organization

```json
{
  "id": "owner",
  "login": "owner",
  "type": "Organization",
  "name": "Display Name",
  "location": "San Francisco",
  "last_fetched": "2025-01-20T12:00:00Z"
}
```

## Web Interface Features

### Recent Improvements

**Compact Sidebar Layout** ✅
- Sticky filters remain visible while scrolling
- Multi-keyword filtering with AND logic
- Dynamic counts update based on current selection
- Keyword tag cloud shows top 50 technologies
- Clear visual feedback for active filters

**Template Preview Modal** ✅
- Click any card to preview YAML content
- Professional syntax highlighting (highlight.js)
- Multiple close methods (Escape, click outside, buttons)
- Instant display with no animation delay
- Copy button for entire template

**Lima 2.0 GitHub URL Scheme** ✅
- Generates shortest valid `github:` URLs
- Handles org repo shorthand (`github:owner//path`)
- Auto-omits `.yaml` extension
- One-click copy to clipboard
- Inline copy button for clean layout

**Default Branch URLs** ✅
- Shows latest template content from default branch
- Backend stores commit SHA for change detection
- Frontend converts SHA URLs to branch URLs on-the-fly

**Modular ES6 Architecture** ✅
- Refactored 667-line monolith into 10 focused ES6 modules
- Clear separation of concerns (config, state, data, filters, UI)
- Pure functions for easier testing and maintainability
- JSDoc documentation throughout all modules
- Zero build step required (native ES modules)

**Accessibility & UX Enhancements** ✅
- Comprehensive ARIA labels for screen reader support
- Semantic HTML roles (main, complementary, dialog)
- Keyboard focus trap for modal navigation
- Debounced search input (300ms) for better performance
- Improved error handling and display

**Automated Testing** ✅
- Comprehensive unit test suite (67+ tests)
- Node.js test runner for command-line execution
- Browser-based test runner with visual output (optional)
- Custom lightweight test framework (no external dependencies)
- Tests cover: URL helpers, data parsing, filters, formatters
- All tests must pass before PR creation
- Minimal DOM mocking for Node.js compatibility

**UI Simplification & Consistency** ✅
- Removed top filter bar - all controls consolidated in sidebar
- Replaced Type dropdown with checkboxes (Official/Community)
- Moved search, sort, and clear filters to sidebar
- Contextual clear buttons (× symbol) that appear only when needed
- Search field clear button (visible when text entered)
- Keywords clear button (visible when keywords selected)
- Smaller, tighter checkbox styling matching category list
- Improved sidebar scroll positioning (no gap at top)
- No empty space reserved for keywords when none selected

## Design System

**📖 Full interface guidelines**: See [INTERFACE_GUIDELINES.md](INTERFACE_GUIDELINES.md) for complete interaction patterns, button behaviors, accessibility requirements, and animation guidelines.

### Color Palette & Guidelines

Based on Material Design and Apple Human Interface Guidelines for proper dark mode implementation.

**Light Theme Colors**:
```css
--primary: #2563eb       /* Primary blue for interactive elements */
--primary-dark: #1e40af  /* Darker blue for hover states */
--primary-light: #3b82f6 /* Lighter blue for selected states */
--bg: #f8fafc            /* Page background (light gray) */
--surface: #ffffff       /* Cards, tags, inputs (white) */
--surface-elevated: #ffffff  /* Modals (same as surface in light mode) */
--surface-code: #f8fafc  /* Code blocks (matches page background) */
--text: #1e293b          /* Primary text (dark gray) */
--text-light: #64748b    /* Secondary text (medium gray) */
--border: #e2e8f0        /* Standard borders */
--border-elevated: #cbd5e1  /* Modal borders (slightly darker) */
```

**Dark Theme Colors**:
```css
--primary: #3b82f6       /* Primary blue (brighter for visibility) */
--primary-dark: #2563eb  /* Darker blue for hover */
--primary-light: #60a5fa /* Lighter blue for selected states */
--bg: #0f172a            /* Page background (very dark blue-gray) */
--surface: #1e293b       /* Cards, tags, inputs (dark blue-gray) */
--surface-elevated: #2d3748  /* Modals (lighter, shows elevation) */
--surface-code: #1a202c  /* Code blocks (distinct from surface) */
--text: #f1f5f9          /* Primary text (off-white) */
--text-light: #94a3b8    /* Secondary text (light gray) */
--border: #334155        /* Standard borders */
--border-elevated: #475569  /* Modal borders (lighter) */
```

### Design Principles

**1. Elevation & Hierarchy** (Material Design)
- Modals use `--surface-elevated` to appear "on top" of regular content
- In dark mode: elevation = lighter background (adds ~16% white overlay effect)
- Regular surfaces: `#1e293b` → Elevated surfaces: `#2d3748`
- Provides visual separation between UI layers

**2. Contrast Requirements** (WCAG/Apple HIG)
- **Minimum 4.5:1 contrast ratio** for text and interactive elements
- **Avoid pure white (#ffffff)** on dark backgrounds (causes blurring/distortion)
- Use **light gray (#f1f5f9)** for primary text in dark mode
- Reduce saturation for colors in dark mode to avoid visual intensity

**3. Surface Differentiation**
- `--bg`: Page/container backgrounds (lowest level)
- `--surface`: Cards, tags, buttons, inputs (mid level)
- `--surface-elevated`: Modals, dialogs, popovers (highest level)
- `--surface-code`: Code blocks (specialized, distinct from other surfaces)

**4. Button & Interactive Elements**
- **Primary buttons**: Use `--primary` background with white text
- **Secondary buttons**: Transparent background with `--border-elevated` outline
- **Hover states**: Border color changes, subtle lift effect (translateY + shadow)
- Never use `--bg` for button backgrounds (blends with page)

**5. Color Saturation**
- Light mode: Full saturation for vibrant feel
- Dark mode: Reduced saturation to prevent eye strain
- Primary blue: #2563eb (light) → #3b82f6 (dark, 10% lighter and less saturated)

### References
- [Material Design Dark Theme](https://m3.material.io/styles/color/dark-theme/overview)
- [Apple HIG Dark Mode](https://developer.apple.com/design/human-interface-guidelines/foundations/dark-mode/)
- [Dark Mode UI Design Best Practices](https://blog.logrocket.com/ux-design/dark-mode-ui-design-best-practices-and-examples/)

## Implementation Milestones

### Phase 1: Data Collection ✅

1. Template discovery via GitHub Code Search
2. Content-based validation to eliminate false positives
3. Metadata collection for repos and organizations
4. JSON Lines storage with minimal diffs
5. Incremental updates with SHA-based change detection
6. GitHub Actions workflow for daily automation

### Phase 2: Content Analysis ✅

1. Smart name derivation for generic filenames
2. YAML parsing to extract images and provisioning
3. Technology detection from scripts
4. Automatic categorization
5. Keyword extraction for searching

### Phase 3: Web Interface ✅

1. Static GitHub Pages site
2. Search and filtering
3. Category browsing
4. Template cards with rich metadata
5. Sidebar layout with tag cloud
6. Multi-keyword filtering
7. Template preview modal
8. Lima 2.0 `github:` URL support

### Current State

- **716 templates** cataloged (51 official + 665 community)
- **Daily automated updates** via GitHub Actions
- **Smart categorization** with keyword extraction
- **Rich web interface** with preview and Lima 2.0 URLs
- **Incremental analysis**: Only re-analyzes changed templates

## Future Enhancements

- **LLM-based descriptions**: Optional AI enhancement for better summaries
- **Template validation**: YAML structure and Lima compatibility checks
- **Quality scoring**: Rank by stars, recency, completeness
- **CLI search tool**: Command-line template discovery
- **Template detail pages**: Dedicated pages with full metadata

## Technical Details

### Why JSON Lines?

- **Minimal git diffs**: Adding one item = one line change
- **Easy streaming**: Process large files incrementally
- **Simple merging**: Line-by-line deduplication
- **Human readable**: Plain JSON, one per line

### Why Separate Data Branch?

- **Clean separation**: Code changes don't trigger data rebuilds
- **Independent updates**: Data updates don't clutter main branch history
- **Easy access**: Users can clone just data without code
- **GitHub Pages**: Fetches data from separate branch

### Why Content Validation?

GitHub Code Search returns false positives:
- Kubernetes ConfigMaps with `minimumLimaVersion` annotations
- GitHub Actions workflows with `provision:` keys
- Documentation files with code examples

Content-based filtering (checking for `images:` key) eliminates ~31% false positives.

### Why Incremental Analysis?

Templates are only re-analyzed if their SHA changes:
- **Efficiency**: Avoids re-parsing unchanged files
- **Consistency**: Templates keep same metadata until updated
- **Speed**: Daily updates complete in minutes, not hours

See `pkg/discovery/analyzer.go:170` for implementation.

## Research Findings

See [FINDINGS.md](FINDINGS.md) for detailed research on:
- GitHub Code Search behavior with forks
- Lima template structure investigation
- Scale assessment and strategy decisions

Key insight: GitHub doesn't index forks unless they have more stars than the parent. Since lima-vm/lima has 18,903 stars, all 750 forks are invisible to search. We focus on independent community templates instead.
