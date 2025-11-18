# Lima Template Catalog - Architecture & Plan

## Overview

A searchable catalog of 700+ Lima VM templates from across GitHub, with automated discovery, smart categorization, and a beautiful web interface.

**Live Site**: [lima-catalog.github.io/lima-catalog](https://lima-catalog.github.io/lima-catalog/)

## Current State

**Production:**
- ✅ 716 templates cataloged (51 official + 665 community)
- ✅ Daily automated updates via GitHub Actions
- ✅ Smart categorization with keyword extraction
- ✅ Rich web interface with preview and Lima 2.0 URLs
- ✅ Incremental updates (Stages 1-5 complete)

**Data Pipeline Status:**
- ✅ **Stage 1**: Incremental discovery with timestamp filtering
- ✅ **Stage 2**: Content validation (verify `images:` key)
- ✅ **Stage 3**: Template analysis (keywords, categories)
- ✅ **Stage 4**: Metadata refresh cycle (5% per run, oldest-first)
- ✅ **Stage 5**: Frontend data generation (`catalog.jsonl`)
- ⏳ **Stage 6**: LLM descriptions (planned)
- ⏳ **Stage 7**: Template cleanup (planned)

## System Architecture

### Backend (Go CLI Tool)

**Purpose**: Discover templates, collect metadata, analyze content

**Key features:**
- GitHub Code Search with incremental updates
- Content-based validation to eliminate false positives
- YAML parsing and technology detection
- Automatic categorization and keyword extraction
- Efficient metadata refresh (oldest-first, 5% per run)
- Blocklist filtering for false positives

**Data storage:**
- JSON Lines format (one object per line)
- Separate files: `templates.jsonl`, `repos.jsonl`, `orgs.jsonl`
- Frontend-optimized: `catalog.jsonl`
- Stored in `data` branch for isolation

**Schedule**: Runs daily via GitHub Actions

### Frontend (Static GitHub Pages)

**Purpose**: Browse and search templates

**Key features:**
- Multi-keyword filtering with AND logic
- Category browsing with dynamic counts
- Template preview modal with YAML syntax highlighting
- Lima 2.0 `github:` URL generation and copy
- About/Help modal with tabbed interface (catalog info + keyboard shortcuts)
- Responsive design for mobile/tablet/desktop
- Fetches `catalog.jsonl` directly from `data` branch

**Tech stack**: Modular ES6 JavaScript + highlight.js (no build step)

See [INTERFACE_GUIDELINES.md](INTERFACE_GUIDELINES.md) for complete design system documentation.

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
  }
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
  "notability_score": 285.5
}
```

## Notability Scoring System

**Purpose**: Identify and prioritize the most "interesting" templates for LLM analysis and user discovery.

**Raw Metrics** (stored in `notability` field):
- `message_length` - Length of user-facing message (>0 indicates template meant for reuse)
- `provision_count` - Number of provision scripts
- `provision_total_lines` - Total lines across all provision scripts
- `probe_count` - Number of probe scripts
- `probe_total_lines` - Total lines across all probe scripts
- `param_count` - Number of configurable parameters
- `env_count` - Number of environment variables
- `comment_line_count` - Number of YAML comment lines (indicates documentation quality)
- `unusual_images` - List of unusual image **domains** (not in official templates)

**Official Images Detection** (Domain-Based):
- Dynamically fetched from `lima-vm/lima/templates/_images/` directory via GitHub API
- Extracts **domains** from image URLs (e.g., `cloud-images.ubuntu.com`)
- Domain-based matching handles version updates automatically
- No hardcoded data - fully autonomous system
- Fetched once per analyzer run, cached for all template analyses

**Unusual Images** (what gets stored):
- Template image URLs are parsed to extract their domains
- Domains not found in official images list are stored in `unusual_images`
- Example: Template using `https://nixos.org/channels/...` would store `nixos.org`
- Deduplicates domains (multiple images from same domain counted once)
- Skips `template://` references (internal template references)

**Score Calculation** (weighted sum):
1. **Message**: 100 points (strong signal for reusability)
2. **Provision scripts**: 10 points per script + 1 point per 10 lines
3. **Parameters**: 20 points per param (indicates configurability)
4. **Environment vars**: 10 points per var (shows configuration effort)
5. **Probes**: 5 points per probe + 1 point per 10 lines
6. **Unusual images**: 30 points if any unusual domains present (Lima uses first available)
7. **Comment lines**: 2 points per comment line (documentation quality)
8. **Repository stars**: 1 point per 10 stars (capped at 50 points)

**Usage**:
- Frontend: Sort templates by `notability_score` to show most interesting first
- LLM Analysis: Process templates in notability order (highest first) to prioritize valuable templates
- Rate limiting: With ~20-30 LLM requests/day, notability ensures we analyze the best templates first

**Design Decision**: Store raw metrics, calculate score on-demand. This allows weight tuning without re-analyzing all templates.

## Remaining Work

### Stage 6: LLM Descriptions (Optional Enhancement)

**Goal**: Generate quality descriptions for templates

**New file**: `descriptions.jsonl`

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

**Generation logic:**
- Compute `source_hash` from template + repo + org data (exclude timestamps)
- Generate description only if:
  1. No description exists, OR
  2. source_hash doesn't match (data changed), AND
  3. Template not in blocklist, AND
  4. Template does not have `meta.description` (author provided), AND
  5. Template does not have `meta.noindex: true` (author opted out)
- Rate limit: Start with 1 description/run (configurable)
- Use cheapest/fastest LLM (Claude Haiku, GPT-3.5-turbo, etc.)
- Fallback to analysis-based keywords if LLM unavailable

**Prompt Builder Architecture:**

New package `pkg/prompt` provides comprehensive context gathering for LLM analysis:

Context included in prompts:
1. **Template YAML content** - Full template with provisioning scripts, probes, messages, env vars, params
2. **Template comments** - Extracted YAML comments for author intent
3. **Repository metadata** - Description, topics/keywords, language, stars
4. **Organization info** - Owner name, type, description, location
5. **README content** - Repository README (truncated to ~5000 chars)
6. **Template references** - Shallow clone + grep for template filename with 15 lines context before/after

Standalone CLI tool `cmd/prompt-generator`:
- Purpose: Test prompts with different LLM models before backend integration
- Usage: `prompt-generator owner/repo/path/template.yaml`
- Output: Formatted prompt ready for LLM testing (stdout or file)
- Configuration: Context lines, max README length, max references, enable/disable sections
- Token estimation: Shows approximate token count for prompt size planning

Backend integration:
- Same `pkg/prompt.Builder` will be used by analyzer
- Configurable via `PromptConfig` (context lines, included sections, limits)
- Shallow clones are temporary (cleaned up after reference gathering)

**Important caveats in prompts:**
- Warns LLM that repo purpose may differ from template purpose
- Repository might be CI scaffolding, docs, or template collection
- Template itself may make the repo's project available in VM
- Instructs LLM to prioritize template content over repo metadata

**Environment variables:**
```bash
LLM_API_KEY=<api-key>
LLM_MODEL=claude-3-haiku-20240307
LLM_MAX_DESCRIPTIONS_PER_RUN=1  # Start conservative
LLM_PROVIDER=anthropic  # or openai, etc.
```

**Integration with Stage 5:**

After implementing Stage 6, update combiner to use description priority:
1. `meta.description` (author-provided, if available)
2. LLM-generated description
3. Analysis-based keywords (current fallback)

**Error handling:**
- Log failures but don't block pipeline
- Continue with analysis-based data if LLM fails
- Retry failed descriptions next run

---

### Stage 7: Template Cleanup (Future)

**Goal**: Remove templates that no longer exist

**Deletion detection:**
- Check templates not updated in 14+ days
- Fetch template URL (HEAD request for efficiency)
- Mark as failed if 404/403/500 received
- Retry logic: Check again after 7 days, then 14 days
- Delete after 3 consecutive failures (total 35 days)

**New fields:**
```yaml
template:
  last_check_failed: "2024-03-01T00:00:00Z"  # First failure timestamp
  check_failures: 2                           # Consecutive failure count
  pending_deletion: false                     # Flagged for removal
```

**Orphan cleanup:**
```
# After template deletion, clean up orphaned metadata
active_repos = set(t.repo for t in templates)
active_orgs = set(t.org for t in templates)

repos = [r for r in repos if r.id in active_repos]
orgs = [o for o in orgs if o.id in active_orgs]
descriptions = [d for d in descriptions if d.template_id in active_templates]
```

**Note on meta.noindex templates:**
- Templates with `meta.noindex: true` kept in database for now
- Excluded from frontend (like blocklist) but not deleted
- Future enhancement: Treat long-standing noindex as deletion candidates
- Allows authors to toggle noindex without losing historical data

---

## Template Meta Field Support (Future)

Lima templates may add a `meta` field for user-defined metadata. Our pipeline will respect these conventions:

**Meta field conventions:**
```yaml
# In template YAML:
meta:
  description: "Authoritative description from template author"
  keywords: ["user", "defined", "keywords"]
  noindex: true  # Exclude from catalog
```

**Priority order** (for final output):
1. `meta.description` (if present) - authoritative
2. LLM-generated description (if available)
3. Analysis-based keywords (fallback)

**Noindex handling:**
- Treat `meta.noindex: true` exactly like blocklist at ALL stages
- Skip LLM generation (save tokens)
- Exclude from `catalog.jsonl` entirely
- Keep in `templates.jsonl` (deletion is future work)

---

## Discovery Strategy

### Search Queries

Four queries to maximize template discovery:

1. `minimumLimaVersion extension:yaml -repo:lima-vm/lima`
2. `minimumLimaVersion extension:yml -repo:lima-vm/lima`
3. `images: provision: extension:yaml -repo:lima-vm/lima`
4. `images: provision: extension:yml -repo:lima-vm/lima`

**Incremental mode:** Add `pushed:>DATE` qualifier to each query

### Content Validation

Downloads and validates each file to eliminate false positives:
- Must contain `images:` as top-level YAML key
- Filters out Kubernetes ConfigMaps, GitHub Actions, etc.
- ~31% of initial results are false positives

### Blocklist Filtering

**File:** `config/blocklist.yaml`

Path patterns (matched against file path within repo):
- `^\.github/workflows/` - GitHub Actions
- `/lima\.REJECTED\.yaml$` - Rejected templates
- `/rancher-desktop/lima/0/lima\.yaml$` - Old Rancher Desktop config

Repo patterns (matched against full org/repo/path):
- Add patterns as needed for spam orgs or specific repos

### Official Templates

Separately enumerates `lima-vm/lima/templates/` directory to get 51 official templates.

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

## Research Findings

See [FINDINGS.md](FINDINGS.md) for detailed research on:
- GitHub Code Search behavior with forks
- Lima template structure investigation
- Scale assessment and strategy decisions

**Key insight:** GitHub doesn't index forks unless they have more stars than the parent. Since lima-vm/lima has 18,903 stars, all 750 forks are invisible to search. We focus on independent community templates instead.

---

## Implementation Notes

For detailed implementation notes on completed stages, testing strategy, rollout history, and migration notes, see [IMPLEMENTATION_NOTES.md](IMPLEMENTATION_NOTES.md).
