# Data Pipeline

**Quick Links**: [Overview](overview.md) | [Backend Design](backend-design.md) | [Future Work](future-work.md)

---

## Overview

The data pipeline consists of 5 completed stages that run daily via GitHub Actions, transforming GitHub search results into a browsable catalog.

**Pipeline Flow:**
```
Stage 1: Discovery → Stage 2: Validation → Stage 3: Analysis → Stage 4: Metadata → Stage 5: Frontend
```

**Status**: ✅ All stages (1-5) complete and production-ready

→ For planned future stages, see [Future Work](future-work.md)

---

## Stage 1: Incremental Discovery

**Goal**: Find new and changed Lima templates across GitHub

### Search Queries

Four queries to maximize template discovery:

1. `minimumLimaVersion extension:yaml -repo:lima-vm/lima`
2. `minimumLimaVersion extension:yml -repo:lima-vm/lima`
3. `images: provision: extension:yaml -repo:lima-vm/lima`
4. `images: provision: extension:yml -repo:lima-vm/lima`

**Incremental mode**: Add `pushed:>DATE` qualifier to each query (uses last discovery timestamp)

### Search Result Deduplication

The four search queries often return overlapping results (the same template matching multiple queries). A simple map-based deduplication prevents adding the same template ID twice:

```go
templateMap := make(map[string]types.Template)
for _, t := range templates {
    if _, exists := templateMap[t.ID]; !exists {
        templateMap[t.ID] = t
    }
}
```

**Important**: This is **not** content-based duplicate detection. It only prevents the same search result (same `owner/repo/path`) from being processed multiple times within a single discovery run. Content-based similarity detection (finding templates with similar content across different repositories) happens in Stage 3 using MinHash + LSH.

### Official Templates

Separately enumerates `lima-vm/lima/templates/` directory to get official templates.

### Implementation

**File**: `pkg/discovery/discovery.go`

**Key functions:**
- `DiscoverAll()` - Main orchestrator
- `DiscoverCommunityTemplates()` - GitHub Code Search queries
- `DiscoverOfficialTemplates()` - Enumerate lima-vm/lima/templates/

**Performance**:
- Full scan: ~20 minutes
- Incremental: ~2-5 minutes (only new/changed)

---

## Stage 2: Content Validation

**Goal**: Eliminate false positives from GitHub Code Search

### Why Needed?

GitHub Code Search returns files that match keywords but aren't Lima templates:
- Kubernetes ConfigMaps with `minimumLimaVersion` annotations
- GitHub Actions workflows with `provision:` keys
- Documentation files with code examples

**False positive rate**: ~31% without validation

### Validation Logic

Downloads each file and checks for `images:` as a top-level YAML key:

```go
// Simplified pseudocode - actual implementation includes error handling
func (d *Discoverer) isLimaTemplate(owner, repo, path string) bool {
    content, err := d.client.GetRepositoryContent(owner, repo, path)
    if err != nil {
        return false
    }

    // Decode base64 content and check for "images:" line
    contentStr, err := content.GetContent()
    if err != nil {
        return false
    }

    lines := strings.Split(contentStr, "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "images:") {
            return true
        }
    }
    return false
}
```

### Blocklist Filtering

**File**: `config/blocklist.yaml`

**Path patterns** (matched against file path within repo):
- `^\.github/workflows/` - GitHub Actions
- `/lima\.REJECTED\.yaml$` - Rejected templates
- `/rancher-desktop/lima/0/lima\.yaml$` - Old Rancher Desktop config

**Repo patterns** (matched against full org/repo/path):
- Add patterns as needed for spam orgs or specific repos

**Implementation**: `pkg/discovery/blocklist.go`

---

## Stage 3: Template Analysis

**Goal**: Extract keywords, categories, and detect duplicates

### YAML Parsing with Lima Integration

Uses Lima's native template parsing to properly handle:
- **Base template references**: `base: template:_images/ubuntu`
- **Script file embedding**: External provisioning scripts
- **Template composition**: Multi-layer template inheritance

**Environment Requirements**:
- `LIMA_TEMPLATES_PATH`: Path to Lima's templates directory (required for `template:` reference resolution)
- `GITHUB_TOKEN`: Required for API access

**URL Strategy**:
- When default branch known: `https://raw.githubusercontent.com/owner/repo/branch/path` (most efficient)
- When branch unknown: `github:owner/repo/path` (Lima resolves via API call)

**Implementation**:
- `pkg/discovery/parser.go:ParseTemplate()` - Uses Lima's `Read()` and `Embed()`
- `pkg/discovery/parser.go:ParseTemplateContent()` - Validates with Lima's `limayaml.Validate()` and extracts structured information from Lima's parsed `LimaYAML` struct

**Extracts**:
- Images and architectures (after base template embedding)
- Provision scripts and probes
- Parameters and environment variables
- Message field
- Comments

### Keyword Extraction

**Sources**:
- Image names (ubuntu, alpine, debian, etc.)
- Provisioning scripts (docker, kubernetes, git, etc.)
- Parameters and env vars
- Repository topics

**Deduplication**: Lowercase, deduplicate, sort

**Implementation**: `pkg/discovery/analyzer.go:AnalyzeTemplate()`

### Category Inference

**Categories**:
- containers (docker, podman)
- kubernetes (k3s, k8s, kubectl)
- development (python, node, rust, go)
- databases (postgres, mysql, mongodb, redis)
- security (vault, secrets, encryption)
- machine-learning (jupyter, tensorflow, pytorch)
- networking (vpn, proxy, dns)
- gaming (steam, wine, proton)
- other (default)

**Logic**: Keyword-based classification with priority order

**Implementation**: `pkg/discovery/analyzer.go:inferCategory()`

### Duplicate Detection (Content-Based)

**Algorithm**: MinHash + LSH

This is **content-based similarity detection** that identifies templates with similar content across different repositories. This is distinct from the [search result deduplication](#search-result-deduplication) in Stage 1, which only prevents processing the same search result twice.

1. Generate 128-hash MinHash signature from 5-word shingles
2. Build LSH index (32 bands × 4 rows)
3. Find similar templates (50%+ similarity)
4. Classify: exact (>90%), near (70-90%), similar (50-70%)

**Storage**: Signatures and similar template IDs stored in `templates.jsonl`

→ See [Backend Design](backend-design.md#duplicate-detection-system) for details

### Notability Scoring

Calculate template "interestingness" score:
- Message: 50 base + 1 per line (capped at 100 total)
- Provision scripts: 10 points per substantial script (>10 lines, capped at 3, min 1) + 1 point/10 total lines
- Parameters: 20 points/param
- Environment vars: 10 points/var
- Probes: 5 points per substantial probe (>10 lines, capped at 3, min 1) + 1 point/10 total lines
- Unusual images: 30 points if present
- Custom images: 0-70 points (org/repo name matches)
- Comment lines: 2 points/line (filtered)
- Repository stars: 1 point/10 stars (max 50)

→ See [Backend Design](backend-design.md#notability-scoring-system) for details

### Change Detection

**SHA-based**: Only re-analyze templates when SHA changes

Templates with same SHA skip analysis, preserving existing data.

---

## Stage 4: Metadata Refresh

**Goal**: Keep repository and organization metadata fresh

### Metadata Types

**Repository metadata** (`repos.jsonl`):
- Description, topics, stars
- Language, default branch
- Last fetched timestamp

**Organization metadata** (`orgs.jsonl`):
- Display name, location
- Type (User/Organization)
- Last fetched timestamp

### Incremental Refresh Strategy

**5% per run, oldest-first**:
- Sort metadata by `last_fetched` timestamp
- Refresh oldest 5% each run
- Spreads API load over time (20 runs = complete refresh)

**Rationale**:
- Most metadata doesn't change frequently
- Spreads GitHub API load
- Prioritizes stalest data

### Concurrent Fetching

**Semaphore pattern** with MaxMetadataConcurrency = 5:
- Fetch up to 5 repos/orgs concurrently
- ~5x performance improvement vs. sequential

**Implementation**: `pkg/discovery/metadata.go:fetchRepositoriesConcurrent()`

---

## Stage 5: Frontend Data Generation

**Goal**: Generate optimized `catalog.jsonl` for frontend

### Data Merging

Combines three data sources:
1. `templates.jsonl` - Template analysis results
2. `repos.jsonl` - Repository metadata
3. `orgs.jsonl` - Organization metadata

### Field Selection

**Included fields** (frontend-optimized):
- `id`, `name`, `description`
- `keywords`, `category`
- `repo`, `org`, `path`
- `stars`, `updated_at`
- `official`, `url`, `github_url`, `raw_url`
- `notability_score`, `notability_score_breakdown`
- `similar_templates`

**Excluded fields** (backend-only):
- Raw notability metrics
- MinHash signatures (512 bytes/template)
- Full SHA hashes
- Internal timestamps

### URL Generation

The combiner generates two types of URLs using Lima v2.0.1:

**`github_url`** (github: scheme URL):
- Constructed using `getGitHubSchemeURL()` helper
- Format: `github:owner/repo/path` (Lima-compatible)
- Removes `.yaml` extension and `/.lima` suffix
- Handles org shorthand: `github:lima-vm` for `lima-vm/lima-vm`
- Used by Lima CLI to fetch templates

**`raw_url`** (https: URL):
- Generated using Lima's `TransformCustomURL()`
- Automatically resolves symlinks (e.g., `.lima.yaml` → `ubuntu.yaml`)
- Follows GitHub redirects transparently
- Uses repository default branch
- Format: `https://raw.githubusercontent.com/owner/repo/branch/path.yaml`
- Used by frontend to fetch template content

**Benefits:**
- Single source of truth for URL generation (backend)
- Consistent with Lima CLI behavior
- Handles edge cases (symlinks, redirects) automatically
- No network calls required in tests (mockable `URLTransformer` interface)

**Implementation:** See [Lima Integration](backend-design.md#lima-integration--url-handling)

### Blocklist Application

Templates matching blocklist patterns excluded from `catalog.jsonl` but kept in `templates.jsonl`.

### Implementation

**File**: `pkg/combiner/combiner.go`

**Function**: `CombineData()`

---

## Progress Tracking

**File**: `data/progress.json`

**Schema**:
```json
{
  "last_discovery": "2025-01-20T12:00:00Z",
  "templates_discovered": N,
  "repositories_fetched": M,
  "organizations_fetched": K
}
```

*Example values shown - actual counts vary*

**Usage**:
- Last discovery time for incremental queries
- Stats for monitoring
- Debugging aid

---

## Error Handling

### Retry Logic

**Exponential backoff** for transient failures:
- Initial delay: 1s
- Max delay: 30s
- Multiplier: 2.0
- Max retries: 3

**Implementation**: `pkg/retry/retry.go`

### Rate Limiting

**GitHub API limits**:
- Core: 5000 requests/hour
- Search: 30 requests/minute

**Handling**:
- Check rate limits before operations
- Wait when threshold reached (100 core, 5 search)
- Retry with backoff on rate limit errors

**Implementation**: `pkg/github/client.go:HandleRateLimitError()`

### Graceful Degradation

**Template analysis failures**:
- Log warning
- Skip template (don't block pipeline)
- Retry next run

**Metadata fetch failures**:
- Use cached/stale data
- Retry in next refresh cycle

---

## Data Storage

### JSON Lines Format

**Why JSON Lines?**
- Minimal git diffs (one line per object)
- Easy streaming (process incrementally)
- Simple merging (line-by-line deduplication)
- Human readable (plain JSON)

### Files

```
data/
├── templates.jsonl      # Template analysis results
├── repos.jsonl          # Repository metadata
├── orgs.jsonl           # Organization metadata
├── catalog.jsonl        # Frontend-optimized
└── progress.json        # Progress tracking
```

### Data Branch

**Separate branch** for data:
- Clean separation (code vs. data)
- Independent updates
- Easy access (clone just data)
- GitHub Pages fetches from here

---

## Performance

### Pipeline Runtime

**Full scan**: ~20 minutes

**Incremental (typical)**: ~2-5 minutes
- Discovery: ~30 seconds (timestamp-filtered)
- Validation: ~1 minute (only new templates)
- Analysis: ~1 minute (only changed SHAs)
- Metadata: ~1 minute (5% refresh = ~30 items)
- Combining: ~5 seconds

### Optimization Strategies

1. **Incremental discovery**: Timestamp filtering
2. **SHA-based analysis**: Skip unchanged templates
3. **Metadata refresh cycle**: 5% per run
4. **Concurrent fetching**: 5 parallel requests
5. **LSH duplicate detection**: Sub-linear search
6. **Caching**: In-memory cache for API responses

---

## Related Documentation

- **[Overview](overview.md)** - System architecture
- **[Backend Design](backend-design.md)** - Detailed backend features
- **[Frontend Design](frontend-design.md)** - Data consumption
- **[Future Work](future-work.md)** - Stages 6-7 planned
- **[Source Index](source-index.md)** - Find implementation files

**Research**:
- **[GitHub Search Behavior](../research/github-search-behavior.md)** - Why we search this way
- **[Duplicate Detection](../research/duplicate-detection-algorithms.md)** - Why MinHash + LSH
