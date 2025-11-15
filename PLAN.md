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
- Keywords clear button (× symbol) appears when keywords selected
- ESC key clears search field (no button needed, simplifies navigation)
- Smaller, tighter checkbox styling matching category list
- Improved sidebar scroll positioning (no gap at top)
- No empty space reserved for keywords when none selected

**Header Design with Icon** ✅
- Lima catalog logo icon (96×96px) displayed in header
- Teal/blue-gray color scheme (#5a7a8c) matching icon background
- Icon aligns with sidebar controls (32px left margin)
- 30px title font with optimized spacing for visual balance
- Compact header height (20px top/bottom padding)
- Icon from original 888×888px PNG, scales crisply at any size

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
  - Home/End/PageUp/PageDown transfer focus from search to templates (like other sidebar fields)
  - "?" works everywhere, even in search field (opens keyboard help)
  - Uppercase letters blocked in search field (reserved for shortcuts)
- **Section jump shortcuts** (single-letter, Gmail-style):
  - K (or Shift+K) to jump to keywords (selected first, then unselected)
  - C (or Shift+C) to jump to categories
  - S (or Shift+S) to jump to sort dropdown
  - T (or Shift+T) to jump to first template card
  - Uppercase variants work even when typing in search field (and on checkboxes)
  - Unassigned uppercase letters trigger shake animation feedback
- **Section navigation shortcuts** (Ctrl+Arrow for major sections):
  - Ctrl+← to move from templates to sidebar (search box)
  - Ctrl+→ to move from sidebar to first template
  - Ctrl+↑ to move to header (theme switcher + help button)
  - Ctrl+↓ to move from header to templates (first template)
  - Use `/` shortcut to jump directly to search box from anywhere
  - Enables quick navigation between the three main page areas
- **Unified keyword navigation** (seamless between selected/unselected):
  - ArrowRight from last selected keyword → jumps to first unselected keyword
  - ArrowLeft from first unselected keyword → jumps to last selected keyword
  - ArrowUp from first row of unselected → jumps to last row of selected keywords
  - ArrowDown from last row of selected → jumps to first row of unselected keywords
  - Keywords and selected keywords feel like one continuous list
- **Header navigation** (Arrow keys between header buttons and templates):
  - ArrowLeft/Right to navigate between: Light ☀️ / Auto 🌓 / Dark 🌙 / Help ?
  - Navigation wraps (left from first → last, right from last → first)
  - ArrowUp/Down transfers focus to first template card (leaves header)
  - Includes keyboard help button (?) in header navigation flow
  - Provides quick keyboard access to theme switching and help
- **Continuous sidebar navigation** (Arrow Up/Down moves between all groups):
  - Search input → Official checkbox → Community checkbox → Sort dropdown
  - Sort dropdown → Selected keywords → Unselected keywords → Categories
  - Within keywords: preserves row-based navigation (Up/Down by row, Left/Right within row)
  - Within categories: Up/Down navigates items sequentially
  - Transitions happen at boundaries (first/last row or item)
  - Dropdown Arrow Down opens menu via SPACE (no longer opens with Arrow Down)
- **Viewport-aware template navigation**:
  - ArrowUp/ArrowDown: Auto-focus first visible template when scrolling from body/html (not when in header)
  - PageUp/PageDown: Scroll page normally, then focus first visible template card in viewport
  - Home: Focus very first template card and scroll to top
  - End: Focus very last template card and scroll to bottom
  - Intelligent viewport detection ensures focus follows scroll position
  - Template cards have scroll-margin (half the gap) for better visibility when navigating
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
  - Focus returns to search after clearing with ESC key
- **Keyboard help modal**:
  - "?" key works everywhere (even in search field) to show/hide help
  - Discoverable question mark icon (?) in header for easy access
  - Focus trap: TAB stays within modal, cycles between close button and content
  - Background scroll locked (body overflow hidden) while modal is open
  - ESC or "?" to close the help overlay
  - Shortcuts (K/C/S/T/?) close modal and execute action
  - Smart focus restoration: returns to search if opened from search, otherwise to previous element
  - Lists all available shortcuts organized by category (Jump to Section & Navigate & Scroll)
  - Documents both lowercase and uppercase (Shift+) shortcut variants
  - Balanced 2-column layout (7 items / 12 items, no scrolling needed)
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

## Incremental Update Redesign (Planned)

### Current Limitations

The existing data collection process has scalability issues:
- **Long runtime**: Full scan takes 20+ minutes
- **Code search limit**: Will break when results exceed 1000 items
- **Inefficient metadata refresh**: Re-fetches all repo/org data every run
- **No LLM descriptions**: Missing human-readable summaries
- **No cleanup**: Deleted templates never removed from database

### Redesign Goals

1. **Incremental discovery**: Only process changed/new templates
2. **Scalability**: Handle >1000 templates without code search limits
3. **Efficient metadata**: Refresh only stale data (spread over 20+ days)
4. **LLM descriptions**: Generate quality summaries for better UX
5. **Automatic cleanup**: Remove templates that no longer exist
6. **Path-based filtering**: Skip known false positives before content fetch

### Architecture Overview

**Data Pipeline Stages:**

```
1. DISCOVER  → Find new/changed templates (incremental code search + path filter)
2. VALIDATE  → Content-based testing (verify images: key exists)
3. ANALYZE   → Extract keywords, categories, detect technologies
4. METADATA  → Fetch repo/org info (new templates + refresh cycle)
5. DESCRIBE  → Generate LLM descriptions (rate-limited)
6. COMBINE   → Create frontend data file (only needed fields)
7. CLEANUP   → Remove deleted templates and orphaned metadata
```

### Implementation Plan

#### Stage 1: Incremental Discovery

**Goal**: Find only new and recently changed templates

**Approach**:
- Find the newest template in our existing data (by `DiscoveredAt` timestamp)
- Query GitHub for templates pushed since 24 hours before that newest template
- Query: `minimumLimaVersion extension:yaml pushed:>YYYY-MM-DD`
- This ensures we always refetch the last template and get all new ones
- No need to track "last check" timestamp separately
- Built-in sanity check: if we get 0 results, something is likely wrong

**Blocklist Filter**:
- Maintain blocklist file: `docs/blocklist.yaml` (application config, not generated data)
- Two separate filter lists (both regex-based):
  1. **Path patterns** - regex matched against file path within repo (e.g., `.github/workflows/`)
  2. **Repo patterns** - regex matched against full `org/repo/path` (e.g., `^spamorg/`)
- Check before downloading content (saves API calls)
- Support comments for documentation

**Examples**:
```yaml
# Path patterns (regex) - matched against file path within repo
# Use this for patterns that apply across all repos
paths:
  - '^\.github/workflows/'      # GitHub Actions (any repo)
  - '^\.gitlab-ci\.ya?ml$'      # GitLab CI (any repo)
  - '^kubernetes/'              # K8s configs (any repo)
  - '/tests?/'                  # Test directories (any repo)
  - '/examples?/'               # Example directories (any repo)
  - '^docs?/'                   # Documentation (any repo)
  - '^\.circleci/'              # CircleCI (any repo)

# Repo patterns (regex) - matched against full org/repo/path
# Provides fine-grained control: block entire repos, orgs, or specific templates
repos:
  - '^spamorg/'                           # Block entire org
  - '^someorg/spam-repo$'                 # Block specific repo
  - '^someorg/repo/bad-template\.yaml$'  # Block specific template
  - '^someorg/repo/subdir/'              # Block directory in specific repo
  # Add more as needed
```

**Filter Logic**:
```go
func isBlocklisted(owner, repo, path string, blocklist Blocklist) bool {
    fullPath := owner + "/" + repo + "/" + path

    // Check repo patterns (matches against full org/repo/path)
    for _, pattern := range blocklist.Repos {
        if matched, _ := regexp.MatchString(pattern, fullPath); matched {
            return true
        }
    }

    // Check path patterns (matches against path within repo)
    for _, pattern := range blocklist.Paths {
        if matched, _ := regexp.MatchString(pattern, path); matched {
            return true
        }
    }

    return false
}
```

**New Fields**:
```yaml
template:
  first_seen: "2024-01-15T10:30:00Z"    # When first discovered
  last_updated: "2024-03-20T15:45:00Z"  # When file SHA changed
  last_analyzed: "2024-03-20T16:00:00Z" # When analysis last ran
```

#### Stage 2: Content Validation

**No changes** - existing logic works well:
- Download template content
- Parse YAML
- Verify `images:` key exists at top level
- ~31% false positive rate remains acceptable

#### Stage 3: Template Analysis

**No changes** - existing keyword/category extraction works well:
- Technology detection
- Keyword extraction
- Category assignment
- Name derivation

**Future: Template Meta Field Support**

Lima templates may add a `meta` field for user-defined metadata. Our pipeline will respect these conventions:

**Meta Field Conventions** (future):
```yaml
# In template YAML:
meta:
  description: "Authoritative description from template author"
  keywords: ["user", "defined", "keywords"]
  noindex: true  # Exclude from catalog
```

**Extraction** (Stage 3 - Analysis):
- Parse `meta` field if present during YAML parsing
- Store in templates.jsonl:
  ```json
  {
    "meta_description": "Author's description",
    "meta_keywords": ["user", "keywords"],
    "meta_noindex": true
  }
  ```
- Falls back to null/empty if meta fields not present

**Priority Order** (for final output):
1. `meta.description` (if present) - authoritative
2. LLM-generated description (if available)
3. Analysis-based keywords (fallback)

**Keyword Merging**:
- Combine `meta.keywords` + analysis keywords
- Remove duplicates
- Optionally boost meta keywords (they appear first)
- LLM can augment but not replace meta keywords

**Noindex Handling**:
- Treat `meta.noindex: true` exactly like path filter blocklist at ALL stages
- **Stage 5 (LLM)**: Skip LLM generation (save tokens)
- **Stage 6 (Combine)**: Exclude from frontend data file entirely
- **Stage 7 (Cleanup)**: Keep in templates.jsonl for now (deletion is future work)
- **Transition handling**: If author adds noindex later:
  - Next analysis detects it and sets meta_noindex: true
  - Template immediately excluded from frontend on next run
  - Remains in database until deletion feature implemented
  - Eventually becomes candidate for deletion (future enhancement)
- Important: noindex can be added/removed at any time, so check on every run

**Implementation Notes**:
- Graceful degradation: Works with or without meta field
- No breaking changes to existing templates
- Forward-compatible with Lima upstream changes
- Meta field takes precedence over all automated systems

**Example Priority**:
```go
// In combine stage
func getDescription(template, llmDesc) string {
    if template.MetaDescription != "" {
        return template.MetaDescription  // Authoritative
    }
    if llmDesc != nil {
        return llmDesc.ShortDescription  // LLM-generated
    }
    return template.AnalysisKeywords[0]  // Fallback
}

func getKeywords(template, llmDesc) []string {
    keywords := template.MetaKeywords  // Start with meta
    if llmDesc != nil && len(keywords) == 0 {
        keywords = llmDesc.Keywords  // LLM if no meta
    }
    if len(keywords) == 0 {
        keywords = template.AnalysisKeywords  // Fallback
    }
    // Optionally merge meta + analysis for richer data
    return deduplicate(append(template.MetaKeywords, template.AnalysisKeywords...))
}

func shouldIndex(template) bool {
    // Treat meta.noindex exactly like blocklist
    return !template.MetaNoindex
}

func shouldGenerateLLM(template) bool {
    // Skip LLM if: blocklisted, noindex, or already has meta.description
    if template.MetaNoindex {
        return false  // Treat like blocklist
    }
    if template.MetaDescription != "" {
        return false  // Author provided
    }
    return true
}
```

**Migration Path**:
1. Add meta field parsing to YAML parser (no-op if field missing)
2. Update templates.jsonl schema with meta_* fields
3. Update LLM stage to skip templates with meta.description
4. Update combine stage to use priority order
5. No changes needed to existing templates without meta

#### Stage 4: Metadata Management

**Goal**: Fetch repo/org data efficiently

**New Templates**:
- Fetch repo metadata for all newly discovered templates
- Fetch org metadata for any new organizations
- No rate limit concerns (typically <10 new templates/day)

**Existing Templates** (Refresh Cycle):
- Track last fetch time for each repo/org
- Identify entries >30 days old
- Refresh max 5% of total entries per run
- Spreads load over ~20 days (100% / 5%)
- Prevents thundering herd on daily runs

**New Fields**:
```yaml
repository:
  last_fetched: "2024-03-15T12:00:00Z"  # When metadata refreshed

organization:
  last_fetched: "2024-03-15T12:00:00Z"  # When metadata refreshed
```

**Algorithm**:
```
new_repos = templates_added_today.repos
refresh_candidates = repos where last_fetched > 30 days ago
refresh_count = min(len(refresh_candidates), total_repos * 0.05)
refresh_list = random.sample(refresh_candidates, refresh_count)

fetch_metadata(new_repos + refresh_list)
```

#### Stage 5: LLM Descriptions

**Goal**: Generate quality descriptions for templates

**New File**: `descriptions.jsonl`

**Schema**:
```json
{
  "template_id": "owner/repo/path/to/template.yaml",
  "short_description": "Brief one-liner (max 100 chars)",
  "long_description": "Detailed explanation (max 500 chars)",
  "keywords": ["keyword1", "keyword2"],
  "generated_at": "2024-03-20T10:00:00Z",
  "source_hash": "abc123...",
  "llm_model": "claude-3-haiku-20240307"
}
```

**Generation Logic**:
- Compute `source_hash` from template + repo + org data (exclude timestamps)
- Generate description only if:
  1. No description exists, OR
  2. source_hash doesn't match (data changed), AND
  3. Template not in path filter blocklist, AND
  4. Template does not have `meta.description` (author provided), AND
  5. Template does not have `meta.noindex: true` (author opted out)
- Rate limit: Start with 1 description/run (configurable via env var)
- Skip templates with author-provided metadata to save tokens
- Use cheapest/fastest LLM (Claude Haiku, GPT-3.5-turbo, etc.)
- Include template YAML, repo description, topics in prompt
- Fallback to analysis-based keywords if no LLM description

**Environment Variables**:
```bash
LLM_API_KEY=<api-key>
LLM_MODEL=claude-3-haiku-20240307
LLM_MAX_DESCRIPTIONS_PER_RUN=1  # Start conservative
LLM_PROVIDER=anthropic  # or openai, etc.
```

**Error Handling**:
- Log failures but don't block pipeline
- Continue with analysis-based data if LLM fails
- Retry failed descriptions next run

#### Stage 6: Frontend Data Combination

**Goal**: Create optimized file for web interface

**Output**: `templates-combined.jsonl` (already exists)

**Fields** (only include what frontend needs):
```json
{
  "id": "owner/repo/path/template.yaml",
  "name": "Display Name",
  "description": "From LLM or analysis",
  "keywords": ["from", "llm", "or", "analysis"],
  "categories": ["containers"],
  "repo": "owner/repo",
  "org": "owner",
  "path": "path/template.yaml",
  "stars": 123,
  "updated_at": "2024-03-20",
  "official": true,
  "url": "https://github.com/...",
  "raw_url": "https://raw.githubusercontent.com/..."
}
```

**Logic**:
- Skip templates in path filter blocklist
- Skip templates with `meta.noindex: true` (treat exactly like blocklist)
- Use description priority: meta.description > LLM > analysis
- Use keyword priority: meta.keywords (merged with analysis) > LLM > analysis
- Include only templates with valid repo/org data
- Keep file size minimal (path filtering and data reduction should be sufficient)

**Priority Example**:
```go
// Combine stage logic
if template.MetaNoindex {
    continue  // Skip this template entirely
}

description := template.MetaDescription
if description == "" && llmDesc != nil {
    description = llmDesc.ShortDescription
}
if description == "" {
    description = strings.Join(template.Keywords, ", ")
}

keywords := template.MetaKeywords
if len(keywords) == 0 && llmDesc != nil {
    keywords = llmDesc.Keywords
}
if len(keywords) == 0 {
    keywords = template.Keywords
}
```

#### Stage 7: Template Cleanup

**Goal**: Remove templates that no longer exist

**Deletion Detection**:
- Check templates not updated in 14+ days
- Fetch template URL (HEAD request for efficiency)
- Mark as failed if 404/403/500 received
- Retry logic: Check again after 7 days, then 14 days
- Delete after 3 consecutive failures (total 35 days)

**Note on meta.noindex templates**:
- Templates with `meta.noindex: true` are kept in database for now
- They are excluded from frontend (like blocklist) but not deleted
- Future enhancement: Treat long-standing noindex templates as deletion candidates
- This allows authors to toggle noindex without losing historical data immediately

**New Fields**:
```yaml
template:
  last_check_failed: "2024-03-01T00:00:00Z"  # First failure timestamp
  check_failures: 2                           # Consecutive failure count
  pending_deletion: false                     # Flagged for removal
```

**Algorithm**:
```
stale_templates = templates where last_updated < 14 days ago

for template in stale_templates:
    response = HEAD(template.raw_url)

    if response.status >= 400:
        if not template.last_check_failed:
            template.last_check_failed = now()
            template.check_failures = 1
        else:
            days_since_failure = (now() - template.last_check_failed).days

            if days_since_failure >= 7 and template.check_failures == 1:
                template.check_failures = 2
            elif days_since_failure >= 21 and template.check_failures == 2:
                template.check_failures = 3
                template.pending_deletion = true
    else:
        # Template exists, clear failure tracking
        template.last_check_failed = null
        template.check_failures = 0
        template.pending_deletion = false

# Remove templates marked for deletion
templates = [t for t in templates if not t.pending_deletion]
```

**Orphan Cleanup**:
```
# After template deletion, clean up orphaned metadata
active_repos = set(t.repo for t in templates)
active_orgs = set(t.org for t in templates)

repos = [r for r in repos if r.id in active_repos]
orgs = [o for o in orgs if o.id in active_orgs]
descriptions = [d for d in descriptions if d.template_id in active_templates]
```

### Data File Schema Updates

**blocklist.yaml** (new):
```yaml
# Blocklist for templates that should be excluded from catalog
# Both lists use regex patterns for maximum flexibility

# Path patterns (regex) - matched against file path within repo
# Use this for patterns that apply across all repositories
paths:
  - '^\.github/workflows/'      # GitHub Actions
  - '^\.gitlab-ci\.ya?ml$'      # GitLab CI
  - '^kubernetes/'              # K8s configs
  - '/tests?/'                  # Test directories
  - '/examples?/'               # Example directories
  - '^docs?/'                   # Documentation
  - '^\.circleci/'              # CircleCI
  # Add more as we discover false positives

# Repo patterns (regex) - matched against full org/repo/path
# Provides fine-grained control for specific repos, orgs, or templates
repos:
  - '^spamorg/'                           # Block entire org
  - '^someorg/spam-repo$'                 # Block specific repo
  - '^someorg/repo/bad-template\.yaml$'  # Block specific template
  - '^someorg/repo/subdir/'              # Block directory in specific repo
  # Add more as needed
```

**templates.jsonl** (updated):
```json
{
  "id": "owner/repo/path/template.yaml",
  "file_sha": "abc123...",
  "first_seen": "2024-01-15T10:30:00Z",
  "last_updated": "2024-03-20T15:45:00Z",
  "last_analyzed": "2024-03-20T16:00:00Z",
  "last_check_failed": null,
  "check_failures": 0,
  "pending_deletion": false,
  "name": "Derived Name",
  "keywords": ["from", "analysis"],
  "categories": ["containers"],
  "meta_description": null,
  "meta_keywords": null,
  "meta_noindex": false,
  "official": false,
  "repo": "owner/repo",
  "org": "owner",
  "path": "path/template.yaml",
  "url": "https://github.com/...",
  "raw_url": "https://raw.githubusercontent.com/..."
}
```

**repos.jsonl** (updated):
```json
{
  "id": "owner/repo",
  "last_fetched": "2024-03-15T12:00:00Z",
  "name": "repo",
  "description": "Repo description",
  "stars": 123,
  "updated_at": "2024-03-20",
  "topics": ["lima", "vm"],
  "language": "Go"
}
```

**orgs.jsonl** (updated):
```json
{
  "id": "owner",
  "last_fetched": "2024-03-15T12:00:00Z",
  "name": "Organization Name",
  "avatar_url": "https://...",
  "type": "Organization"
}
```

**descriptions.jsonl** (new):
```json
{
  "template_id": "owner/repo/path/template.yaml",
  "short_description": "Brief summary",
  "long_description": "Detailed explanation",
  "keywords": ["llm", "generated"],
  "generated_at": "2024-03-20T10:00:00Z",
  "source_hash": "abc123...",
  "llm_model": "claude-3-haiku-20240307"
}
```

### Data File Sorting

**Goal**: Maintain stable, human-readable file ordering

**Rationale**:
- **Human browsing**: Grouped by org/repo makes files easier to navigate
- **Stable diffs**: New templates from existing repos appear near related templates
- **Pattern detection**: Issues and trends easier to spot when related entries are adjacent
- **Minimal overhead**: Sorting 700-1000 items is negligible (<1ms)

**Sort Orders**:

**templates.jsonl**:
- Primary: `org` (alphabetical)
- Secondary: `repo` (alphabetical)
- Tertiary: `path` (alphabetical)
- Example order:
  ```
  acme-corp/backend/lima.yaml
  acme-corp/frontend/lima.yaml
  acme-corp/tools/dev.yaml
  beta-org/app/template.yaml
  ```

**repos.jsonl**:
- Primary: `org` (alphabetical)
- Secondary: `repo` (alphabetical)
- Groups repos by organization

**orgs.jsonl**:
- Single key: `id` (alphabetical)

**descriptions.jsonl**:
- Match `templates.jsonl` order: `template_id` (alphabetical by org/repo/path)
- Ensures descriptions align with templates for easier cross-reference

**Implementation**:
```go
// Sort templates by org, repo, path
sort.Slice(templates, func(i, j int) bool {
    if templates[i].Org != templates[j].Org {
        return templates[i].Org < templates[j].Org
    }
    if templates[i].Repo != templates[j].Repo {
        return templates[i].Repo < templates[j].Repo
    }
    return templates[i].Path < templates[j].Path
})

// Sort repos by org, name
sort.Slice(repos, func(i, j int) bool {
    if repos[i].Org != repos[j].Org {
        return repos[i].Org < repos[j].Org
    }
    return repos[i].Name < repos[j].Name
})

// Sort orgs by id
sort.Slice(orgs, func(i, j int) bool {
    return orgs[i].ID < orgs[j].ID
})

// Sort descriptions by template_id
sort.Slice(descriptions, func(i, j int) bool {
    return descriptions[i].TemplateID < descriptions[j].TemplateID
})
```

**Migration Impact**:
- Initial sort will create one large diff (entire file reordered)
- Subsequent updates will have minimal, localized diffs
- Consider separate commit: "Sort data files for human readability"
- Future changes will be much cleaner

**Benefits for Git**:
```diff
# Before sorting (templates scattered):
+ owner1/repo-a/template.yaml
+ owner2/repo-b/template.yaml
+ owner1/repo-c/template.yaml   # Same owner, far apart

# After sorting (templates grouped):
+ owner1/repo-a/template.yaml
+ owner1/repo-c/template.yaml   # Same owner, adjacent
+ owner2/repo-b/template.yaml
```

### Workflow Configuration

**Environment Variables**:
```bash
# Required
GITHUB_TOKEN=<token>

# LLM Configuration (optional, degrades gracefully if missing)
LLM_API_KEY=<api-key>
LLM_PROVIDER=anthropic  # anthropic, openai, etc.
LLM_MODEL=claude-3-haiku-20240307
LLM_MAX_PER_RUN=1

# Configuration
DISCOVERY_LOOKBACK_HOURS=48
METADATA_REFRESH_AGE_DAYS=30
METADATA_REFRESH_PERCENT=0.05
DELETION_CHECK_AGE_DAYS=14
DELETION_RETRY_DAYS=7
```

**GitHub Actions Steps**:
```yaml
steps:
  - name: Checkout
  - name: Setup Go
  - name: Load Last Run Timestamp
  - name: Discover Templates (incremental)
  - name: Validate Content
  - name: Analyze Templates
  - name: Fetch Metadata (new + refresh cycle)
  - name: Generate Descriptions (rate-limited)
  - name: Combine Frontend Data
  - name: Check Deleted Templates
  - name: Cleanup Orphans
  - name: Sort Data Files (by org/repo/path)
  - name: Save Run Timestamp
  - name: Commit & Push Data
  - name: Report Statistics
```

### Observability & Monitoring

**Statistics to Track**:
- Templates discovered: X
- Templates validated: Y
- New templates: Z
- Repos fetched: A (B new + C refreshed)
- Orgs fetched: D (E new + F refreshed)
- Descriptions generated: G
- Templates deleted: H
- Orphans cleaned: I repos, J orgs
- Run duration: MM:SS
- Errors encountered: K

**Logging**:
- Log level: INFO for normal ops, DEBUG for troubleshooting
- Include timestamps and component names
- Log rate limit warnings
- Log LLM API failures separately

### Testing Strategy

**Local Development & Testing**:

The pipeline should be fully testable locally without deploying to GitHub Actions. This requires:

1. **Command-line interface**:
   ```bash
   # Run full pipeline
   go run ./cmd/lima-catalog --github-token=$GITHUB_TOKEN

   # Run specific stages
   go run ./cmd/lima-catalog discover --since=48h
   go run ./cmd/lima-catalog analyze
   go run ./cmd/lima-catalog metadata --refresh-percent=0.05
   go run ./cmd/lima-catalog describe --max=1
   go run ./cmd/lima-catalog combine
   go run ./cmd/lima-catalog cleanup

   # Test mode (smaller dataset, verbose output)
   go run ./cmd/lima-catalog --test-mode --limit=10
   ```

2. **GitHub Token Configuration**:
   - Use personal access token (PAT) for local testing
   - Same token works for both Code Search and REST API
   - Required scopes: `public_repo` (read-only)
   - Configure via environment variable or command flag:
     ```bash
     export GITHUB_TOKEN=ghp_xxxxx
     # or
     go run ./cmd/lima-catalog --github-token=ghp_xxxxx
     ```

3. **Test Data**:
   - Small test dataset (10-20 templates) for quick iteration
   - Fixture data in `testdata/` directory for unit tests
   - Mock HTTP responses for unit tests (no real API calls)

4. **Incremental Testing**:
   - Each pipeline stage should be independently testable
   - Output intermediate data to verify correctness
   - Verbose logging mode for debugging
   - Dry-run mode (no file writes, just show what would change)

**Unit Tests** (no network calls):
- Blocklist filter matching (path patterns and repo names)
- Source hash computation
- Deletion retry logic
- Metadata refresh selection
- Date parsing and timestamp handling
- Template YAML parsing
- Data file sorting algorithms

**Integration Tests** (with real GitHub API):
- Use test GitHub token
- Run discovery on small query (limit=10)
- Verify blocklist filtering works
- Test incremental discovery (48-hour window)
- Validate data file formats
- Check error handling (rate limits, network failures)
- Run full pipeline on test dataset

**Continuous Integration**:
- Unit tests run on every PR (no token needed)
- Integration tests run with secrets in GitHub Actions
- Test mode runs quickly (<2 minutes)
- Full pipeline tested in staging before production

**Manual Testing**:
- Run locally before deploying to GitHub Actions
- Verify incremental updates work correctly
- Check LLM description quality
- Validate deletion of removed templates
- Test edge cases (network failures, invalid YAML, etc.)

**Example Test Workflow**:
```bash
# 1. Unit tests (fast, no network)
go test ./...

# 2. Integration test (small dataset)
export GITHUB_TOKEN=ghp_xxxxx
go run ./cmd/lima-catalog --test-mode --limit=10 --verbose

# 3. Dry-run on real data
go run ./cmd/lima-catalog --dry-run --verbose

# 4. Real run (after verification)
go run ./cmd/lima-catalog

# 5. Verify output
cat data/templates.jsonl | wc -l
cat data/repos.jsonl | wc -l
```

### Rollout Plan

**Phase 1: Data File Sorting & Blocklist** (Week 1)
- **Sort existing data files** (one-time large diff)
  - Sort templates by org/repo/path
  - Sort repos by org/name
  - Sort orgs by id
  - Commit as: "Sort data files for human readability"
- **Add blocklist.yaml**
  - Initial path patterns (workflows, CI configs, tests, docs)
  - Initial repo blocklist (if any known spam repos)
  - Implement filter checking in discovery
  - Unit tests for filter matching
- Test locally with small dataset
- Monitor false positive reduction
- Verify sorted files improve browsability

**Phase 2: Incremental Discovery** (Week 2)
- Implement date-based code search
- Add timestamp tracking
- Implement CLI for local testing
- Test locally with 48-hour lookback
- Integration tests with real GitHub API
- Verify no templates missed
- Test dry-run mode

**Phase 3: Metadata Refresh Cycle** (Week 3)
- Add last_fetched timestamps
- Implement 5% refresh logic
- Monitor API rate limits
- Verify coverage over 20 days

**Phase 4: LLM Descriptions** (Week 4)
- Add descriptions.jsonl
- Implement hash-based change detection
- Start with 1 description/run
- Monitor quality and costs
- Gradually increase limit

**Phase 5: Template Cleanup** (Week 5)
- Add deletion tracking fields
- Implement check and retry logic
- Test with known-deleted templates
- Verify orphan cleanup

**Phase 6: Frontend Integration** (Week 6)
- Update frontend to use LLM descriptions
- Add fallback to analysis keywords
- Test with mixed data (some LLM, some not)
- Deploy to production

### Success Criteria

- ✅ Runtime reduced from 20min to <5min (incremental updates)
- ✅ Handles >1000 templates (no code search limit)
- ✅ Metadata refresh spreads over 20+ days (5% per run)
- ✅ LLM descriptions generated for active templates (rate-limited)
- ✅ Deleted templates removed automatically (35-day cycle)
- ✅ Zero manual intervention required
- ✅ No data integrity issues
- ✅ Graceful degradation when LLM unavailable

### Future Enhancements

Beyond this redesign:
- **Template validation**: YAML structure and Lima compatibility checks
- **Quality scoring**: Rank by stars, recency, completeness, description quality
- **CLI search tool**: Command-line template discovery
- **Template detail pages**: Dedicated pages with full metadata
- **Advanced LLM features**: Generate installation instructions, detect use cases
- **Community feedback**: Allow users to suggest description improvements
- **A/B testing**: Compare LLM vs analysis-based discovery

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
