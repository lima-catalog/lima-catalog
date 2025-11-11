# Lima Template Catalog - Project Plan

## Overview

Create a searchable catalog of Lima templates found across GitHub, with metadata and AI-generated descriptions to help users discover community templates.

## Architecture

### Phase 1: Data Collection (Current Focus)

#### 1.1 Template Discovery
- **Goal**: Find all YAML files on GitHub containing `minimumLimaVersion`
- **Method**: GitHub Code Search API
- **Filtering**:
  - Exclude files from `lima-vm/lima` repository (official templates)
  - Exclude forks of `lima-vm/lima` (clones of official templates)
  - Only include files with `.yaml` or `.yml` extensions

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

### Phase 2: Content Analysis (Future)

#### 2.1 Template Parsing
- Parse YAML structure
- Extract key fields:
  - Images used
  - Provisioning scripts
  - Mounts
  - Port forwards
  - Resource limits

#### 2.2 LLM Analysis
- Use free LLM (e.g., GitHub Models, Hugging Face Inference API)
- Generate:
  - Short description (1-2 sentences)
  - Detailed description (paragraph)
  - Keywords/tags
  - Category classification

### Phase 3: Catalog Generation (Future)

#### 3.1 Static Website
- Browse by category
- Sort by: stars, recency, popularity
- Search functionality

#### 3.2 CLI Tool
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
  "last_checked": "2025-01-20T12:00:00Z"
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
- **Language**: Python 3.10+
- **Libraries**:
  - `requests` or `PyGithub` for GitHub API
  - `click` for CLI interface
  - `pyyaml` for template parsing
  - `gitpython` for git operations
- **CI/CD**: GitHub Actions

### CLI Tool Structure
```
lima-catalog/
├── catalog_tool/
│   ├── __init__.py
│   ├── cli.py           # Main CLI interface
│   ├── discovery.py     # Template discovery
│   ├── metadata.py      # Metadata collection
│   ├── storage.py       # Data persistence
│   ├── progress.py      # Progress tracking
│   └── github_api.py    # GitHub API wrapper
├── tests/
├── requirements.txt
├── README.md
└── PLAN.md
```

### Key Features

#### Rate Limit Management
- Monitor rate limit before each API call
- Stop when limit is low (e.g., < 100 requests remaining)
- Save progress and exit gracefully
- Resume from last checkpoint on next run

#### Resumability
- Save progress after each batch (e.g., 10 templates)
- Store cursor position for paginated searches
- Track which repos/orgs have been fetched
- Allow resuming from any interrupted state

#### Incremental Updates
- Check file SHA to detect template changes
- Only re-fetch metadata if template changed
- Update `last_checked` timestamp for unchanged entries
- Remove templates that no longer exist

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
- Run weekly to discover new templates
- Run daily to update metadata for existing templates
- Use GitHub token with appropriate scopes
- Store data branch with results

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

### Search Optimization
For GitHub Code Search:
- Query: `minimumLimaVersion extension:yml OR extension:yaml -repo:lima-vm/lima`
- Use pagination cursors
- Handle rate limits (30 requests/minute for code search)

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

## Next Steps

1. ✅ Create initial plan
2. 🔄 Experiment: Count templates on GitHub
3. ⏳ Design detailed data schemas
4. ⏳ Implement discovery tool
5. ⏳ Implement metadata collection
6. ⏳ Set up data branch and storage
7. ⏳ Test with rate limits
8. ⏳ Create GitHub Action workflow
