# Backend Design

**Quick Links**: [Overview](overview.md) | [Frontend Design](frontend-design.md) | [Source Index](source-index.md) | [Code Standards](../reference/code-standards.md)

---

## Overview

The backend is a Go CLI tool (`cmd/lima-catalog`) that discovers, validates, and analyzes Lima templates from across GitHub.

**Key Technologies:**
- Go 1.24+
- GitHub API v3 (REST)
- YAML parsing (gopkg.in/yaml.v3)
- MinHash + LSH for duplicate detection

**Architecture Principles:**
- **Dependency Injection**: HTTPClient, FileSystem, Clock interfaces
- **Context Support**: Cancellation for long-running operations
- **Functional Options**: Flexible configuration
- **Idiomatic Error Handling**: Sentinel errors, wrapped errors
- **Comprehensive Test Coverage**: Tests with mocks for all packages

→ For code quality standards, see [Code Standards](../reference/code-standards.md)

---

## Notability Scoring System

**Purpose**: Identify and prioritize the most "interesting" templates for LLM analysis and user discovery.

### Raw Metrics

Stored in `notability` field:

- `message_length` - Length of user-facing message (>0 indicates template meant for reuse)
- `message_line_count` - Number of lines in message (used for scoring)
- `provision_count` - Number of provision scripts
- `provision_substantial` - Number of substantial provision scripts (>10 lines, capped at 3, min 1)
- `provision_total_lines` - Total lines across all provision scripts
- `probe_count` - Number of probe scripts
- `probe_substantial` - Number of substantial probe scripts (>10 lines, capped at 3, min 1)
- `probe_total_lines` - Total lines across all probe scripts
- `param_count` - Number of configurable parameters
- `env_count` - Number of environment variables
- `comment_line_count` - Number of **unique** YAML comment lines (filtered)
- `unusual_images` - List of unusual image **domains** (not in official templates)
- `all_images` - All image domains (stored for org/repo name matching)

### Comment Filtering

Prevents score inflation from inherited comments:

1. Scans entire git history of lima-vm/lima repository (templates/ and examples/ directories)
2. Uses local git clone via `LIMA_REPO_PATH` environment variable
3. Extracts all normalized lines (comments, provision scripts, probes, messages) from official templates across all commits
4. Stores results in `official.json` file
5. Only counts comments/lines that are:
   - Not present in official knowledge (prevents derivative templates from inheriting scores)
   - Not empty (ignores lines with just `#` and whitespace)

**Rationale**: Many templates start from official templates and inherit documentation comments. Without filtering, these templates get artificially inflated scores.

### Official Images Detection

Domain-based, dynamically fetched:

- Fetched from `lima-vm/lima/templates/_images/` directory via GitHub API
- Extracts **domains** from image URLs (e.g., `cloud-images.ubuntu.com`)
- Domain-based matching handles version updates automatically
- No hardcoded data - fully autonomous system
- Fetched once per analyzer run, cached for all template analyses

### Unusual Images

What gets stored:

- Template image URLs are parsed to extract their domains
- **Only `http://` or `https://` URLs are considered** (local references and template expressions are filtered out)
- Domains not found in official images list are stored in `unusual_images`
- Example: Template using `https://nixos.org/channels/...` would store `nixos.org`
- Deduplicates domains (multiple images from same domain counted once)
- Local references like `template://ubuntu` or `/local/path/image.img` are skipped

### Score Calculation

Weighted sum of metrics:

1. **Message**: 50 points base + 1 per line (capped at 100 total)
2. **Provision scripts**: 10 points per substantial script (>10 lines, capped at 3, min 1) + 1 point per 10 total lines
3. **Parameters**: 20 points per param (capped at 100)
4. **Environment vars**: 10 points per var (capped at 100)
5. **Probes**: 5 points per substantial probe (>10 lines, capped at 3, min 1) + 1 point per 10 total lines
6. **Image name**: Combined scoring for image-related metrics (only `http://` or `https://` URLs):
   - -100 points if no remote images (won't work on other computers)
   - 0 points if only official/usual images (neutral)
   - 30 points base if unusual images + optional custom name bonus (0-70):
     - 25 points for one word boundary match (`\bNAME` or `NAME\b`)
     - 35 points for both word boundaries (`\bNAME\b`)
     - Org and repo scored separately and summed (max 70)
     - Custom bonus only available when template has unusual images
7. **Comment lines**: 2 points per comment line (capped at 100)
8. **Repository stars**: 1 point per 10 stars (capped at 50 points)

### Usage

- **Frontend**: Sort templates by `notability_score` to show most interesting first
- **LLM Analysis**: Process templates in notability order (highest first) to prioritize valuable templates
- **Rate limiting**: With ~20-30 LLM requests/day, notability ensures we analyze the best templates first

**Design Decision**: Store raw metrics, calculate score on-demand. This allows weight tuning without re-analyzing all templates.

**Implementation**: See `pkg/discovery/notability.go`

---

## Duplicate Detection System

**Goal**: Identify similar/duplicate templates to help users discover alternatives and understand template relationships.

### Algorithm: MinHash + LSH

- **MinHash**: Generates 128-hash signature from 5-word shingles (YAML content)
- **LSH**: 32 bands × 4 rows for sub-linear similarity search (~42% threshold)
- **Classification**: exact (>90%), near (70-90%), similar (50-70%)

### Performance

- **Signature generation**: ~10-20ms per template
- **Storage**: 512 bytes per template (128 × uint32)
- **Search**: O(n) sub-linear vs O(n²) brute force
- **Accuracy**: ~8.8% error rate with 128 hashes

### Configuration

```go
analyzer := NewAnalyzer(
    WithDetectDuplicates(true),              // Default: enabled
    WithDuplicateSimilarityThreshold(0.5),   // Default: 50%
)
```

### Data Flow

1. Analyzer generates MinHash signature during template analysis
2. `DetectDuplicates()` builds LSH index and finds similar templates
3. Populates `similar_templates` field with IDs, similarity scores, and types
4. Combiner copies to `catalog.jsonl` for frontend
5. Modal displays "Similar Templates" section with badges

### Original Detection

Among exact duplicates (>90% similarity), the system identifies which template is the "original" using these heuristics:

1. **Official templates** (lima-vm/lima) are always considered originals
2. **Oldest repo creation date** - the repo that was created first is likely the original
3. **Higher star count** - tie-breaker for repos created on the same date
4. **Alphabetical order** - final tie-breaker for consistency

Non-original templates get their `original_id` field set, pointing to the original.

**Note**: Ideally we would use the template file's creation date (first git commit), but this would require expensive API calls or cloning repos. Repository creation date is a reasonable approximation.

**Important**: GitHub code search does not return results from forked repositories (unless the fork has more stars than the parent). This means all templates in our catalog are from non-fork repos, simplifying original detection.

### Transitive Similarity Grouping

**The Problem**: Similarity is transitive - if template B is 90% similar to A, and C is 90% similar to B, should C also be considered part of the same duplicate group (even if C is only 85% similar to A)?

**Example**:
```
Template A: ubuntu.yaml
Template B: ubuntu-copy.yaml    (95% similar to A)
Template C: ubuntu-fork.yaml    (92% similar to B, 85% similar to A)
```

Without transitive grouping, we might:
- Show A and hide B (good)
- Show A and C but hide B (confusing - why show C but not B?)

**The Solution**: Union-Find Algorithm

The system uses a [union-find (disjoint-set)](https://en.wikipedia.org/wiki/Disjoint-set_data_structure) data structure to group all transitively connected templates:

1. **Initialize**: Each template starts in its own group
2. **Union**: For each exact duplicate pair (>90%), merge their groups
3. **Find**: Use path compression to efficiently find group representative
4. **Result**: All transitively similar templates end up in the same group

```go
// Simplified pseudocode
func buildExactDuplicateGroups(templates) {
    parent := map[templateID]templateID{}

    // Union all exact duplicates (>90% similarity)
    for each template {
        for each similar in template.SimilarTemplates {
            if similar.Similarity > 0.9 {
                union(template.ID, similar.ID)
            }
        }
    }

    // Group by root
    for each template {
        root := find(template.ID)
        groups[root].append(template.ID)
    }
}
```

**Benefits**:

1. **Consistent filtering**: Only one representative shown per group
2. **Maximizes hiding**: Hides the maximum number of duplicates
3. **Deterministic**: Same result every time (alphabetical tie-breaker)
4. **Efficient**: O(n × α(n)) ≈ O(n) with path compression

**Example Outcome**:
```
Group 1: [A, B, C]
Original: A (oldest repo)
Hidden: B, C (both have original_id = A)
```

When "Duplicates" checkbox is unchecked (default), users see only A. When checked, they see A, B, and C with badges indicating relationships.

**Implementation**: See `buildExactDuplicateGroups()` and `identifyOriginals()` in `pkg/discovery/analyzer.go:539-596`

### Origin vs. Centrality: Design Decision

**The Question**: When picking a representative from a duplicate group, should we choose:
- **Origin-based**: The oldest template (likely the actual original)
- **Centrality-based**: The template most similar to all others (best representative)

**Current Approach**: Origin-based (oldest repo creation date)

**Example where centrality differs from origin**:

```
Group: [A, B, C]

A: ubuntu.yaml (2020)          Origin ← Current choice
   - 85% similar to B
   - 70% similar to C
   - Average: 77.5%

B: ubuntu-enhanced.yaml (2021) Most central!
   - 85% similar to A
   - 95% similar to C
   - Average: 90%

C: ubuntu-fork.yaml (2022)
   - 70% similar to A
   - 95% similar to B
   - Average: 82.5%
```

**Why Origin-Based?**

**Advantages**:
1. **Historical accuracy** - Users see the actual original source
2. **Simplicity** - No additional similarity computations needed
3. **Stability** - Oldest repo never changes, even as new duplicates added
4. **Performance** - Just need repo metadata (already fetched)
5. **Objectivity** - Repo creation date is factual, not derived from error-prone similarity scores (~8.8% error rate)

**Disadvantages**:
1. May not be the best representative of the group
2. Oldest template could be outdated or poorly maintained
3. Doesn't reflect which template users would find most useful

**Why Not Centrality-Based?**

**Advantages**:
1. ✅ Better representative - shows "average" version
2. ✅ Users see template most similar to all variants
3. ✅ Potentially more useful than historical original

**Disadvantages**:
1. ❌ More complex - requires O(n²) pairwise similarity within each group
2. ❌ Less stable - centrality can shift as new templates added
3. ❌ Loses provenance - historical tracking value lost
4. ❌ Error-prone - similarity scores have ~8.8% error rate
5. ❌ Computationally expensive - adds cost to duplicate detection

**Future Consideration**:

If user feedback indicates that centrality-based selection would be more valuable, we could:
- Compute average similarity within each group
- Pick template with highest average similarity to group members
- Store both "original" and "representative" metadata
- Let users toggle between views

For now, the simplicity and historical accuracy of origin-based selection is preferred.

### UI Features

- Color-coded badges: Original (green), Exact (red), Near (orange), Similar (blue)
- Similarity percentage displayed
- Click to navigate between similar templates
- "Duplicates" checkbox to show/hide copies (hidden by default)
- Full keyboard accessibility
- Hidden when no similar templates exist

**Research**: See [Duplicate Detection Research](../research/duplicate-detection-algorithms.md) for algorithm selection rationale and parameter tuning guidelines.

**Implementation**: See `pkg/discovery/analyzer.go`, `pkg/minhash/`

---

## Backend Code Quality

**Current State:**
- ✅ 0 critical issues
- ✅ Comprehensive test coverage across all packages
- ✅ Idiomatic Go APIs with context support
- ✅ Dependency injection for testability
- ✅ Comprehensive documentation

**Key Patterns:**
- **Interfaces**: HTTPClient, FileSystem, Clock for all external dependencies
- **Functional Options**: `NewX(opts ...Option)` for complex constructors
- **Context**: First parameter in long-running functions for cancellation
- **Sentinel Errors**: Named error variables for expected error conditions
- **Table-Driven Tests**: One test function with []struct{} for multiple cases

→ See [Code Standards](../reference/code-standards.md) for detailed quality requirements

---

## Package Structure

```
pkg/
├── discovery/        # Template discovery and analysis
│   ├── discovery.go  # GitHub Code Search integration
│   ├── analyzer.go   # Template analysis (keywords, categories, duplicates)
│   ├── parser.go     # YAML parsing and info extraction
│   ├── naming.go     # Template naming logic
│   ├── metadata.go   # Repository/org metadata collection
│   ├── notability.go # Notability scoring system
│   ├── blocklist.go  # Blocklist pattern matching
│   └── update.go     # Incremental update logic
├── storage/          # JSON Lines storage
│   └── storage.go    # Load/save templates, repos, orgs, progress
├── github/           # GitHub API wrapper
│   └── client.go     # Rate limiting, caching, API wrappers
├── combiner/         # Frontend data generation
│   └── combiner.go   # Merge templates+repos+orgs → catalog.jsonl
├── minhash/          # Duplicate detection
│   ├── minhash.go    # MinHash signature generation
│   └── lsh.go        # LSH index and similarity search
├── prompt/           # LLM prompt generation (future)
│   ├── builder.go    # Context gathering and prompt formatting
│   └── types.go      # Data structures
├── validation/       # Input validation
│   └── validation.go # Token, path, config validation
├── retry/            # Retry logic
│   └── retry.go      # Exponential backoff retry
├── cache/            # In-memory caching
│   └── cache.go      # TTL-based cache
├── config/           # Configuration
│   └── constants.go  # API delays, rate limits, weights
├── interfaces/       # Interfaces for testing
│   └── interfaces.go # HTTPClient, FileSystem, Clock
└── types/            # Core data structures
    └── types.go      # Template, Repository, Organization, Progress, Blocklist, NotabilityMetrics, SimilarTemplate
```

→ For complete file index, see [Source Index](source-index.md)

---

## CLI Tools

### Main Tool: `cmd/lima-catalog`

Daily data collection orchestrator:

```bash
export GITHUB_TOKEN=your_token
export ANALYZE=true
./lima-catalog
```

**Steps:**
1. Load progress and existing templates
2. Discover new/changed templates
3. Validate content (check `images:` field)
4. Analyze templates (keywords, categories, duplicates)
5. Refresh stale metadata (5% per run)
6. Generate frontend data (`catalog.jsonl`)
7. Save progress

### Prompt Generator: `cmd/prompt-generator`

Generate LLM prompts for testing (future Stage 6):

```bash
export GITHUB_TOKEN=your_token
prompt-generator lima-vm/lima/templates/ubuntu.yaml
```

Gathers comprehensive context:
- Template YAML content
- Template comments
- Repository metadata
- Organization info
- README content
- Template references in repo

→ See [LLM Prompts Documentation](../reference/llm-prompts.md)

---

## Related Documentation

- **[Overview](overview.md)** - System architecture and data schema
- **[Frontend Design](frontend-design.md)** - JavaScript modules and UI
- **[Data Pipeline](data-pipeline.md)** - Discovery → Analysis → Frontend
- **[Future Work](future-work.md)** - LLM descriptions, template cleanup
- **[Code Standards](../reference/code-standards.md)** - Quality requirements
- **[Source Index](source-index.md)** - Find any source file

**Related:**
- **[Research](../research/)** - Algorithm selection rationale and decision-making
