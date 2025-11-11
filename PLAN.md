# Lima Template Catalog - Project Plan

## Overview

Create a searchable catalog of Lima templates found across GitHub, with metadata and AI-generated descriptions to help users discover community templates.

## Architecture

### Phase 1: Data Collection ✅

**Status**: Completed

#### 1.1 Template Discovery
- **Goal**: Find all YAML files on GitHub that are Lima templates
- **Method**: GitHub Code Search API with multiple queries
- **Search Queries**:
  1. `minimumLimaVersion extension:yaml -repo:lima-vm/lima` (original templates)
  2. `minimumLimaVersion extension:yml -repo:lima-vm/lima` (original templates)
  3. `images: provision: extension:yaml -repo:lima-vm/lima` (templates without minimumLimaVersion)
  4. `images: provision: extension:yml -repo:lima-vm/lima` (templates without minimumLimaVersion)
- **Filtering**:
  - Exclude files from `lima-vm/lima` repository (handled separately as official templates)
  - Exclude forks of `lima-vm/lima` (GitHub search doesn't index them anyway)
  - Automatic deduplication across queries
  - **Content-based validation**: Fetch file content and verify it contains `images:` as a top-level YAML key
    - Filters out false positives like Kubernetes ConfigMaps and GitHub Actions workflows
    - Uses regex pattern `^images:` to validate Lima template structure
    - Reports exclusion counts per query for monitoring
- **Detailed Query Logging**:
  - Lists all template IDs found by each search query
  - Shows deduplication statistics (new vs duplicate templates)
  - Helps identify which queries return false positives
- **Results**: 716 unique templates discovered (51 official + 665 community)
  - Initial discovery found 1033 templates, but 317 (31%) were false positives
  - Content-based filtering accurately identifies valid Lima templates

#### 1.2 Metadata Collection
For each discovered template, collect:
- **Template-level data**:
  - Repository owner/name
  - File path within repo
  - File SHA (for change detection)
  - Last modified date
  - File size
  - Raw content URL

- **Repository-level data** (one entry per unique repo):
  - Description
  - Topics/keywords
  - Stars, forks, watchers
  - Primary language
  - License
  - Created/updated dates
  - Homepage URL
  - Is fork? Parent repo if yes

- **Organization-level data** (one entry per unique org/user):
  - Display name
  - Description
  - Location
  - Blog/website
  - Type (user vs org)

### Phase 2: Content Analysis ✅

**Status**: Completed (LLM enhancement deferred to future)

#### 2.1 Template Naming
- **Smart name derivation**:
  - Use filename when descriptive (e.g., `ubuntu.yaml` → "ubuntu")
  - For generic names like `lima.yaml`, derive from:
    - Repository name (e.g., `container-security/lima.yaml` → "container-security")
    - Template path (e.g., `configs/lima/dev.yaml` → "dev")
    - Fallback to full path if needed
- **Display name**: Human-readable name for UI/catalog
- **Unique identifier**: Keep full path as ID for uniqueness

#### 2.2 Template Parsing
- Parse YAML structure
- Extract key fields:
  - Images used (OS distributions, versions)
  - Provisioning scripts (detect tools: Docker, K8s, etc.)
  - Mounts (detect development patterns)
  - Port forwards (detect services)
  - Resource limits (CPU, memory)
  - Architecture (x86_64, aarch64, etc.)

#### 2.3 LLM Analysis
- Use free LLM (e.g., Anthropic, OpenAI, or local models)
- Generate:
  - **Display name**: Descriptive name (e.g., "Ubuntu Development Environment")
  - **Short description**: 1-2 sentences summarizing purpose
  - **Detailed description**: Paragraph explaining use case
  - **Keywords/tags**: Technology stack (docker, kubernetes, python, etc.)
  - **Category**: Primary use case (development, testing, security, ci-cd, etc.)
  - **Use case**: Specific scenario (web development, ML training, etc.)

#### 2.4 Analysis Strategy
- **Primary signal**: Provisioning scripts (strongest indicator of purpose)
- **Secondary signal**: Images used (base OS, pre-built images)
- **Context clues**: Repository description, topics, readme
- **Fallback**: Repository/org metadata if template is minimal

### Phase 3: Catalog Website ✅

**Status**: Completed (with minor layout issues noted for future improvement)

#### 3.1 GitHub Pages Static Site
A static website to browse the catalog, published to GitHub Pages at https://lima-catalog.github.io/lima-catalog/

**Implemented Features:**
- **Browse templates**: List all templates with names, descriptions, categories
- **Category pages**: Group templates by category (containers, development, orchestration, etc.)
- **Search functionality**: Filter by keywords, OS, technologies
- **Template details**: Show full information for each template
- **Statistics**: Display catalog metrics (total templates, categories, etc.)

**Technology Stack:**
- Static site generator (Hugo, Jekyll, or plain HTML/CSS/JS)
- Fetch data from `data` branch via GitHub API or raw.githubusercontent.com
- Deploy to GitHub Pages (gh-pages branch or docs/ folder)
- No backend required - pure static site

**Implementation:**
1. Create static site structure
2. Parse templates.jsonl, repos.jsonl, orgs.jsonl
3. Generate category indexes
4. Add search/filter functionality
5. Deploy to GitHub Pages
6. Update workflow to rebuild site on data changes

**Benefits:**
- Easy review of analyzed templates
- Visual validation of categories and descriptions
- Shareable URL for community discovery
- No hosting costs (GitHub Pages is free)

#### 3.2 CLI Tool (Future)
- Search from command line
- Direct template installation

## Data Storage

### Location
- Store in `data` branch of this repository
- Keeps catalog data separate from code

### Format
All data files use JSON Lines (`.jsonl`) format for minimal diffs:

```
templates.jsonl    - One template per line
repos.jsonl        - One repository per line
orgs.jsonl         - One org/user per line
progress.json      - State tracking for resumability
```

### Schema

**templates.jsonl**:
```json
{
  "id": "owner/repo/path/to/template.yaml",
  "repo": "owner/repo",
  "path": "path/to/template.yaml",
  "sha": "abc123...",
  "size": 1234,
  "last_modified": "2025-01-15T10:30:00Z",
  "url": "https://raw.githubusercontent.com/...",
  "discovered_at": "2025-01-20T12:00:00Z",
  "last_checked": "2025-01-20T12:00:00Z",
  "is_official": false,
  "name": "ubuntu-dev",
  "display_name": "Ubuntu Development Environment",
  "short_description": "Ubuntu-based development environment with Docker and common dev tools",
  "description": "Full-featured Ubuntu development environment with Docker, Git, common build tools...",
  "category": "development",
  "use_case": "web-development",
  "keywords": ["ubuntu", "docker", "nodejs", "python", "git"],
  "images": ["ubuntu:22.04"],
  "arch": ["x86_64", "aarch64"],
  "analyzed_at": "2025-01-21T10:00:00Z"
}
```

**repos.jsonl**:
```json
{
  "id": "owner/repo",
  "owner": "owner",
  "name": "repo",
  "description": "Repository description",
  "topics": ["kubernetes", "lima"],
  "stars": 42,
  "forks": 5,
  "watchers": 10,
  "language": "Go",
  "license": "MIT",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2025-01-15T10:30:00Z",
  "pushed_at": "2025-01-15T10:30:00Z",
  "homepage": "https://example.com",
  "is_fork": false,
  "parent": null,
  "last_fetched": "2025-01-20T12:00:00Z"
}
```

**orgs.jsonl**:
```json
{
  "id": "owner",
  "login": "owner",
  "type": "Organization",
  "name": "Display Name",
  "description": "Org description",
  "location": "San Francisco, CA",
  "blog": "https://example.com",
  "email": "contact@example.com",
  "last_fetched": "2025-01-20T12:00:00Z"
}
```

**progress.json**:
```json
{
  "phase": "discovery",
  "last_search_cursor": "Y3Vyc29yOjE=",
  "templates_discovered": 150,
  "repos_fetched": 45,
  "orgs_fetched": 30,
  "last_updated": "2025-01-20T12:00:00Z",
  "rate_limit_remaining": 4500,
  "rate_limit_reset": "2025-01-20T13:00:00Z"
}
```

## Implementation

### Technology Stack
- **Language**: Go 1.21+
- **Libraries**:
  - `github.com/google/go-github/v57` for GitHub API
  - `gopkg.in/yaml.v3` for YAML parsing
  - Standard library for JSON Lines storage
- **CI/CD**: GitHub Actions (daily workflow)
- **Authentication**: Personal access token (CATALOG_TOKEN) for avoiding GitHub Actions rate limits

### Tool Structure
```
lima-catalog/
├── cmd/lima-catalog/
│   └── main.go          # Main entry point
├── pkg/
│   ├── github/
│   │   └── client.go    # GitHub API client wrapper
│   ├── discovery/
│   │   ├── discovery.go # Template discovery with multi-query search
│   │   ├── metadata.go  # Repository/org metadata collection
│   │   ├── update.go    # Incremental update merging
│   │   ├── naming.go    # Smart name derivation
│   │   ├── parser.go    # YAML template parsing
│   │   └── analyzer.go  # Template categorization
│   ├── storage/
│   │   └── storage.go   # JSON Lines file I/O
│   └── types/
│       └── types.go     # Data structures
├── docs/                # GitHub Pages static site
│   ├── index.html
│   ├── style.css
│   └── app.js
├── data/                # Local data directory (gitignored)
├── experiments/         # Search query testing scripts
├── go.mod
├── README.md
└── PLAN.md
```

### Key Features

#### Rate Limit Management
- **CATALOG_TOKEN**: Uses personal access token instead of GitHub Actions GITHUB_TOKEN to avoid secondary rate limits
- **Pagination delays**: 3-second delays between paginated search requests (max 20 requests/minute)
- **Query delays**: 5-second delays between different search queries
- **Automatic retry**: Detects 403/429 rate limit errors, waits until rate limit reset, then retries
- **Pre-check**: Verifies sufficient quota before starting (core: 100+, search: 5+)
- **Monitoring**: Tracks API usage and displays remaining quota

#### Resumability
- Save progress after each batch (e.g., 10 templates)
- Store cursor position for paginated searches
- Track which repos/orgs have been fetched
- Allow resuming from any interrupted state

#### Incremental Updates
- **Always runs discovery**: In incremental mode, discovery and metadata collection run on every execution
- **SHA-based change detection**: Compares file SHA to detect template modifications
- **Smart merging**: Merges new/updated templates with existing data, preserving historical fields
- **Preserved fields**: `discovered_at` timestamp kept from original discovery
- **Updated fields**: `last_checked`, `sha`, and all metadata refreshed
- **Analysis skipping**: Only analyzes new or modified templates (where `analyzed_at < last_checked`)
- **Deduplication**: Automatically deduplicates templates found by multiple search queries

#### Error Handling
- Retry transient failures (network errors)
- Skip problematic entries and log errors
- Continue processing other items
- Report summary of failures

## Workflow

### Initial Collection
```bash
# Discover all templates
python -m catalog_tool discover --save-progress

# Collect metadata for discovered templates
python -m catalog_tool fetch-metadata --save-progress

# Commit and push to data branch
python -m catalog_tool commit-data
```

### Incremental Updates
```bash
# Check for new/changed templates
python -m catalog_tool update --incremental

# Commit changes
python -m catalog_tool commit-data
```

### GitHub Actions
- **Schedule**: Runs daily at 00:00 UTC (cron: `0 0 * * *`)
- **Manual trigger**: Can be triggered manually via workflow_dispatch
- **Environment**: Uses CATALOG_TOKEN (personal access token) for API access
- **Mode**: Runs in incremental mode with analysis enabled
- **Output**: Commits updated data to `data` branch
- **Workflow file**: `.github/workflows/update-catalog.yml`
- **Commit detection**: Uses `git add` + `git diff --cached` to properly detect both new and modified files
  - Ensures new catalog data is committed even when data branch starts empty
  - Handles incremental updates to existing files

## Improvements & Considerations

### Quality Signals
- **Recency**: Last modified date
- **Popularity**: Stars on repo
- **Completeness**: Template has description, proper structure
- **Maintenance**: Repo is actively maintained

### Deduplication
- Detect forks with unmodified templates
- Group similar templates
- Prefer original over forks when ranking

### Fork Detection
- Use GitHub API `is_fork` field
- Check if template SHA matches parent repo
- Only include fork if template is modified

### Content Hashing
- Store content hash to detect changes without downloading
- Only download full content when hash changes

### False Positive Filtering

#### Problem Discovery
Initial implementation using only GitHub Code Search queries found 1033 templates, but many were false positives:
- **Kubernetes ConfigMaps**: Contained `images:` and `provision:` in discovery configuration contexts
- **GitHub Actions workflows**: Had jobs named "provision" and container image references
- **Other YAML files**: Matched search patterns but weren't Lima templates

Example false positives:
- `wavefrontHQ/observability-for-kubernetes/.../1-default-wavefront-collector-config.yaml` - Kubernetes ConfigMap
- `github-cloudlabsuser-1270/rachid/.github/workflows/contoso-traders-app-deployment.yml` - GitHub Actions workflow

#### Solution: Content-Based Validation
Instead of path-based heuristics, implemented content validation:
1. Fetch file content via GitHub API for each search result
2. Check for `images:` as a top-level YAML key (pattern: `^images:`)
3. Only include files that pass validation
4. Report exclusion counts for monitoring

**Results**: Filtered out 317 false positives (31% reduction), resulting in 716 valid Lima templates.

**Trade-off**: Uses more API quota (one GET request per search result) but significantly improves accuracy.

### Search Optimization

#### Current Implementation
Multiple targeted queries to find templates:
1. `minimumLimaVersion extension:yaml -repo:lima-vm/lima`
2. `minimumLimaVersion extension:yml -repo:lima-vm/lima`
3. `images: provision: extension:yaml -repo:lima-vm/lima`
4. `images: provision: extension:yml -repo:lima-vm/lima`

Automatic pagination with 3-second delays between pages, 5-second delays between queries.

#### GitHub Search API Limits
- **Hard limit**: 1000 results per search query (pages 1-10 at 100 results/page)
- **Current status**: Query 3 returns ~700 results, approaching the limit
- **No workaround**: Cannot retrieve results beyond 1000 for a single query

#### Future: Time-Based Segmentation (when approaching 1000 results)
When any query approaches the 1000 result limit, implement time-based segmentation:

**Strategy**: Split searches by push date to keep each query under 1000 results
```
images: provision: pushed:>2024-12-01 extension:yaml -repo:lima-vm/lima
images: provision: pushed:2024-06-01..2024-11-30 extension:yaml -repo:lima-vm/lima
images: provision: pushed:<2024-06-01 extension:yaml -repo:lima-vm/lima
```

**Benefits**:
- Daily incremental runs only search recent templates (`pushed:>YYYY-MM-DD`)
- Historical templates already in catalog don't need re-discovery
- As long as <1000 new templates per day, system scales indefinitely

**Implementation trigger**: Add monitoring to warn when any query returns >900 results

### Future Enhancements
- **Template Validation**: Parse and validate YAML structure
- **Dependency Analysis**: Track which images/scripts are popular
- **Change Tracking**: Monitor template evolution over time
- **Quality Scoring**: Composite score for ranking
- **Community Contributions**: Allow manual submissions/corrections

## Open Questions

1. **Fork handling**: Should we exclude all forks, or only forks with unmodified templates?
2. **Private templates**: Should we document how users can submit private templates?
3. **Template versioning**: How to handle repos with multiple template versions?
4. **Validation**: Should we validate templates can actually work?
5. **LLM choice**: Which free LLM service for descriptions?

## Success Metrics

- Number of unique templates discovered
- Coverage: percentage of actual Lima templates found
- Freshness: how often data is updated
- Accuracy: quality of generated descriptions
- Usage: number of users finding templates useful

## Implementation Progress

### Completed ✅
1. Initial project plan and architecture
2. Experiments to understand GitHub search behavior and fork indexing
3. Data schema design (JSON Lines format)
4. Go-based discovery tool with multi-query search
5. Metadata collection for repos and organizations
6. Data branch setup and storage implementation
7. Rate limit handling with retry logic and delays
8. GitHub Actions workflow (daily execution)
9. Incremental update mode with SHA-based change detection
10. Template analysis with smart naming and categorization
11. GitHub Pages static website for browsing catalog
12. Authentication with CATALOG_TOKEN to avoid Actions rate limits
13. Content-based filtering to eliminate false positives
14. Detailed query logging for debugging and monitoring
15. Workflow commit detection fix for new files
16. Initial catalog population complete (716 templates: 51 official + 665 community)

### In Progress 🔄
- Monitoring catalog quality and accuracy
- Daily automated updates via GitHub Actions

### Recent: GitHub Pages UI Redesign ✅
With 700+ templates, the original UI needed improvements for better discoverability:

**Problems Addressed:**
- Header took up too much vertical space (1/3 of screen)
- Filters weren't sticky - users lost context when scrolling
- No way to filter by multiple keywords simultaneously
- Categories and keywords were conceptually similar but handled differently

**Solution: Sidebar Layout with Tag Cloud**
- **Compact sticky header**: Reduced from 3rem to 1.5rem padding, moved stats inline
- **Sticky top controls**: Search bar and quick filters remain visible while scrolling
- **Left sidebar (280px)**: Contains keyword cloud and category filters
  - Always visible on desktop, collapsible on mobile
  - Sticky positioning keeps filters accessible
- **Keyword tag cloud**:
  - Displays top 50 keywords by frequency
  - Multi-select with AND logic (e.g., "alpine" + "docker")
  - Shows count for each keyword
  - Selected keywords displayed prominently above cloud
  - Click to toggle selection
- **Category list**:
  - Click to select single category
  - Shows count for each category
  - Visual feedback for selection
- **Clear filters button**: One-click reset of all filters
- **Responsive design**: Sidebar moves to top on tablet/mobile

**User Experience Improvements:**
- More screen space for template cards (reduced header size)
- Filters always accessible (sticky controls)
- Powerful multi-keyword filtering for precise discovery
- Visual feedback for active filters
- Easier navigation with 700+ templates

**Template Preview Modal** ✅
- **In-page preview**: Click any template card to view its YAML content in a modal popup
- **Professional syntax highlighting**: Uses highlight.js with atom-one-light theme
  - Distinct colors for keys, strings, numbers, booleans, comments
  - Much more accurate than regex-based highlighting
- **Instant display**: No animation delay - modal appears immediately at full size
- **Multiple close methods**:
  - Escape key
  - Click outside modal (on overlay)
  - Close button in header (×)
  - Close button in footer
- **Quick navigation**: Preview templates without leaving the catalog page
- **GitHub link preserved**: Clicking org/repo name still opens GitHub in new tab
- **Full URL display**: Shows complete GitHub URL in modal footer instead of generic button text
- **Responsive design**: Modal scales appropriately on mobile devices
- **Loading state**: Shows centered loading indicator while fetching template content

**Default Branch URLs** ✅
- **Latest content**: Template URLs use repository's default branch (main/master) instead of specific commit SHA
- **Backend tracking**: Commit SHA still stored in database for accurate change detection
- **Frontend conversion**: `getDefaultBranchURL()` converts SHA URLs to branch URLs on-the-fly
- **Repository metadata**: Added `default_branch` field to Repository type
- **Automatic updates**: Next catalog run will populate default branch for all existing repositories

**Lima 2.0 GitHub URL Scheme** ✅
- **Shortest URLs**: Generates minimal `github:` scheme URLs for Lima 2.0
  - Standard format: `github:owner/repo/path`
  - Org repos (owner==repo): `github:owner//path`
  - Automatic `.yaml` extension omission
  - Default `.lima` filename handling
- **Inline copy button**: Compact "Copy" button follows github: URL inline
- **YAML copy button**: Top-right copy button in code preview to copy entire template
- **Modal display**: Shows github: URL as second line under template name
- **Smart generation**: `getGitHubSchemeURL()` function creates shortest valid URL
- **Examples**:
  - `github:lima-vm/lima/templates/fedora` → displays templates from lima-vm/lima repo
  - `github:owner//config` → shorthand for org repos where owner==repo

### Future Enhancements 📋
- **LLM-based descriptions**: Optional LLM integration for better template descriptions
- **Time-based search segmentation**: Implement when queries approach 1000 result limit
- **Template validation**: Validate YAML syntax and Lima template structure
- **Quality scoring**: Rank templates by stars, recency, completeness
- **CLI search tool**: Command-line interface for finding and installing templates
- **Template detail pages**: Dedicated pages for each template with full metadata
- **Result count monitoring**: Add warnings when search queries return >900 results
