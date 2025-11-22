# Lima Catalog - Architecture Overview

**Quick Links**: [ARCHITECTURE](../../ARCHITECTURE.md) | [Backend Design](backend-design.md) | [Frontend Design](frontend-design.md) | [Data Pipeline](data-pipeline.md)

---

## Current State (November 2025)

**Production:**
- ✅ Templates cataloged from official and community sources
- ✅ Daily automated updates via GitHub Actions
- ✅ Smart categorization with keyword extraction
- ✅ Rich web interface with preview and Lima 2.0 URLs
- ✅ Incremental updates (Stages 1-5 complete)
- ✅ Duplicate detection with MinHash + LSH

**Data Pipeline Status:**
- ✅ **Stage 1**: Incremental discovery with timestamp filtering
- ✅ **Stage 2**: Content validation (verify `images:` key)
- ✅ **Stage 3**: Template analysis (keywords, categories)
- ✅ **Stage 4**: Metadata refresh cycle (5% per run, oldest-first)
- ✅ **Stage 5**: Frontend data generation (`catalog.jsonl`)
- ⏳ **Stage 6**: LLM descriptions (planned) - See [Future Work](future-work.md)
- ⏳ **Stage 7**: Template cleanup (planned) - See [Future Work](future-work.md)

---

## System Components

### Backend (Go CLI Tool)

**Purpose**: Discover templates, collect metadata, analyze content

**Key features:**
- GitHub Code Search with incremental updates
- Content-based validation to eliminate false positives
- YAML parsing and technology detection
- Automatic categorization and keyword extraction
- Duplicate detection (MinHash + LSH)
- Efficient metadata refresh (oldest-first, 5% per run)
- Blocklist filtering for false positives

**Data storage:**
- JSON Lines format (one object per line)
- Separate files: `templates.jsonl`, `repos.jsonl`, `orgs.jsonl`
- Frontend-optimized: `catalog.jsonl`
- Stored in `data` branch for isolation

**Schedule**: Runs daily via GitHub Actions

→ [Backend Design Details](backend-design.md)

### Frontend (Static GitHub Pages)

**Purpose**: Browse and search templates

**Key features:**
- Multi-keyword filtering with AND logic
- Dynamic ORG/REPO keyword filters
- Category browsing with dynamic counts
- Template preview modal with YAML syntax highlighting
- Duplicate/similar template detection with visual badges
- Lima 2.0 `github:` URL generation and copy
- About/Help modal with tabbed interface
- Responsive design for mobile/tablet/desktop
- Fetches `catalog.jsonl` directly from `data` branch

**Tech stack**: Modular ES6 JavaScript + highlight.js (no build step)

→ [Frontend Design Details](frontend-design.md)

---

## Data Schema

### Template

```json
{
  "id": "owner/repo/path/template.yaml",
  "repo": "owner/repo",
  "org": "owner",
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
  "analyzed_at": "2025-01-15T10:05:00Z",
  "notability": {
    "message_length": 100,
    "provision_count": 2,
    "provision_total_lines": 50,
    "probe_count": 1,
    "probe_total_lines": 10,
    "param_count": 3,
    "env_count": 5,
    "comment_line_count": 15,
    "unusual_images": ["nixos.org"]
  },
  "minhash_signature": [12345, 67890, ...],
  "similar_templates": [
    {
      "id": "other/repo/similar.yaml",
      "similarity": 0.85,
      "shared_bands": 28
    }
  ]
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

### Catalog (Frontend-Optimized)

```json
{
  "id": "owner/repo/path/template.yaml",
  "name": "Display Name",
  "description": "Short description",
  "keywords": ["docker", "kubernetes"],
  "category": "containers",
  "repo": "owner/repo",
  "org": "owner",
  "path": "path/template.yaml",
  "stars": 123,
  "updated_at": "2024-03-20",
  "official": true,
  "url": "https://github.com/...",
  "raw_url": "https://raw.githubusercontent.com/...",
  "notability_score": 285.5,
  "notability_score_breakdown": {
    "message": 100.0,
    "provision": 25.0,
    "parameters": 60.0,
    "env_vars": 50.0,
    "probes": 10.5,
    "unusual_images": 30.0,
    "comments": 8.0,
    "stars": 12.3,
    "total": 285.5
  },
  "similar_templates": [
    {
      "id": "other/repo/similar.yaml",
      "similarity": 0.85,
      "type": "near"
    }
  ]
}
```

---

## Technical Decisions

### Why Incremental Updates?

**Problem:** Full scan takes 20+ minutes, will break at 1000+ templates

**Solution:**
- Timestamp-based discovery (only new/changed templates)
- Metadata refresh cycle (5% per run, oldest-first)
- SHA-based change detection (only re-analyze when changed)

**Benefits:**
- Runtime: 20min → <5min
- Scalability: Handles >1000 templates
- API efficiency: Spreads load over time

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

### Why catalog.jsonl?

- **Frontend optimization**: Single file with only needed fields
- **Reduced complexity**: Frontend doesn't need to join 3 files
- **Faster page load**: One network request instead of three
- **Smaller payload**: Only includes displayed fields

---

## Related Documentation

- **[Backend Design](backend-design.md)** - Notability scoring, duplicate detection
- **[Frontend Design](frontend-design.md)** - JavaScript modules, UI patterns
- **[Data Pipeline](data-pipeline.md)** - Discovery → Analysis → Frontend (Stages 1-5)
- **[Future Work](future-work.md)** - Planned features (Stages 6-7)
- **[Source Index](source-index.md)** - Find any source file
- **[Research](../research/)** - Decision rationale (GitHub search, duplicate detection)
