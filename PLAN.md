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
- Smooth fade-in display (prevents content flicker during loading)
- Natural height expansion (up to 90vh) to show full content
- Copy button for entire template
- Keyboard scrolling support:
  - Vertical: Arrow Up/Down, Page Up/Down, Home/End
  - Horizontal: Arrow Left/Right (scrolls code element for long lines)
- Custom scrollbar styling for dark mode consistency

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

**Dark/Light/Auto Theme Switcher** ✅
- Segmented control in header with icon buttons (☀️ 🌓 🌙)
- Three modes: Light, Auto (follows system), Dark (default: Auto)
- localStorage persistence across sessions
- System preference detection via prefers-color-scheme
- Automatic theme switching when system preference changes
- Complete dark theme with proper color hierarchy
- Elevation-based modal colors (lighter surfaces = higher elevation)
- All UI elements adapt: buttons, tags, badges, code blocks
- Consistent interaction patterns following Material Design / Apple HIG
- Accessible with ARIA labels and keyboard support

**Keyboard Navigation** ✅
- **Search shortcuts**:
  - Auto-focus search box on page load for immediate typing
  - "/" hotkey to focus search (Gmail/GitHub pattern)
  - ESC key to clear search and restore focus to search input
  - "?" works everywhere, even in search field (opens keyboard help)
  - Uppercase letters blocked in search field (reserved for shortcuts)
- **Section jump shortcuts** (single-letter, Gmail-style):
  - K (or Shift+K) to jump to keywords (selected first, then unselected)
  - C (or Shift+C) to jump to categories
  - S (or Shift+S) to jump to sort dropdown
  - T (or Shift+T) to jump to first template card
  - Uppercase variants work even when typing in search field (and on checkboxes)
  - Unassigned uppercase letters trigger shake animation feedback
- **Unified keyword navigation** (seamless between selected/unselected):
  - ArrowRight from last selected keyword → jumps to first unselected keyword
  - ArrowLeft from first unselected keyword → jumps to last selected keyword
  - ArrowUp from first row of unselected → jumps to last row of selected keywords
  - ArrowDown from last row of selected → jumps to first row of unselected keywords
  - Keywords and selected keywords feel like one continuous list
- **Continuous sidebar navigation** (Arrow Up/Down moves between all groups):
  - Search input → Official checkbox → Community checkbox → Sort dropdown
  - Sort dropdown → Selected keywords → Unselected keywords → Categories
  - Within keywords: preserves row-based navigation (Up/Down by row, Left/Right within row)
  - Within categories: Up/Down navigates items sequentially
  - Transitions happen at boundaries (first/last row or item)
  - Dropdown Arrow Down opens menu via SPACE (no longer opens with Arrow Down)
- **Row-based arrow key navigation**:
  - Keywords: Left/Right for adjacent tags, Up/Down to jump to first tag on previous/next row
  - Selected keywords: Same row-based navigation, Delete/Backspace to remove
  - Categories: Up/Down to navigate between categories (vertical list)
  - Template cards: Arrow keys navigate by grid rows/columns (calculates column count dynamically)
- **Interactive element navigation**:
  - Keywords: Tab to focus, Enter/Space to select, fully keyboard accessible
  - Selected keywords: Tab to focus, Enter/Space/Delete/Backspace to remove
  - Categories: Tab to focus, Enter/Space to select/deselect, aria-pressed state
  - Template cards: Tab to focus, Enter/Space to open preview modal
  - Sort dropdown: O shortcut to focus, Arrow keys to navigate options
- **Smart focus management**:
  - Focus preservation when toggling keywords/categories (doesn't lose position)
  - Focus jumps to first keyword in cloud after selecting one
  - Focus jumps to next selected keyword when deselecting (or last if deselected was last)
  - Focus jumps to first unselected keyword only when all selected are removed
  - Focus returns to search after clearing with ESC or clear button
- **Keyboard help modal**:
  - "?" key works everywhere (even in search field) to show/hide help
  - Discoverable question mark icon (?) in header for easy access
  - Focus trap: TAB stays within modal, cycles between close button and content
  - ESC or "?" to close the help overlay
  - Shortcuts (K/C/S/T/?) close modal and execute action
  - Smart focus restoration: returns to search if opened from search, otherwise to previous element
  - Lists all available shortcuts organized by category (Navigation & Actions)
  - Documents both lowercase and uppercase (Shift+) shortcut variants
  - Compact 2-column layout on screens ≥600px (all shortcuts visible without scrolling)
- **Template preview modal focus management**:
  - Focus trap when modal is open for accessibility
  - ESC to close modal and restore focus to the template card that opened it
  - Seamless navigation flow with focus restoration
- **Accessibility features**:
  - All interactive elements have tabindex="0" for keyboard focus
  - Proper ARIA labels (aria-label, role="button", aria-pressed, aria-modal)
  - Visible focus indicators with 2px primary color outlines
  - Complete keyboard-only navigation without mouse requirement
  - Full WCAG 2.1 AA keyboard navigation compliance

## Design System

**📖 Complete design system documentation**: See [INTERFACE_GUIDELINES.md](INTERFACE_GUIDELINES.md)

All UI/UX guidelines are documented in INTERFACE_GUIDELINES.md, including:
- Color palette (light & dark themes)
- Button interaction patterns
- Badge & tag styles
- Interactive element feedback
- Accessibility requirements (WCAG AA)
- Dark mode principles
- Animation timing guidelines

Refer to that document for all design decisions and UI implementation details.

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
