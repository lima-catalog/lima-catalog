# Lima Catalog Source Code Index

Quick reference for all source files and their purposes.

## Backend (Go)

### Command-line Tools

| File | Purpose |
|------|---------|
| `cmd/lima-catalog/main.go` | Main CLI entry point for template discovery and analysis pipeline |
| `cmd/debug-template/main.go` | Debug CLI tool for analyzing templates against official knowledge baseline |
| `cmd/debug-template/main_test.go` | Tests for debug tool |
| `cmd/prompt-generator/main.go` | Standalone CLI for generating LLM prompts for template analysis |
| `cmd/prompt-generator/README.md` | Documentation for prompt generator tool |
| `cmd/prompt-generator/example.sh` | Usage examples for prompt generator |

### Core Packages

#### `pkg/types/`
| File | Purpose |
|------|---------|
| `types.go` | Core data structures: Template, Repository, Organization, Blocklist, Progress |

#### `pkg/github/`
| File | Purpose |
|------|---------|
| `client.go` | GitHub API wrapper with rate limit management and caching |
| `client_test.go` | Tests for GitHub client (caching, rate limiting, error handling) |

#### `pkg/storage/`
| File | Purpose |
|------|---------|
| `storage.go` | JSON Lines storage for templates, repos, orgs (data branch) |
| `storage_test.go` | Tests for storage (save/load, JSON Lines format, error handling) |

#### `pkg/discovery/`
| File | Purpose |
|------|---------|
| `discovery.go` | Main discovery orchestration |
| `parser.go` | YAML template parsing and technology detection |
| `parser_test.go` | Tests for parser |
| `analyzer.go` | Template analysis and categorization |
| `analyzer_test.go` | Tests for analyzer (category inference, description generation) |
| `metadata.go` | Repository and organization metadata fetching with concurrent batching |
| `metadata_test.go` | Tests for metadata collection |
| `naming.go` | Template name derivation from path |
| `naming_test.go` | Tests for naming (derivation, sanitization, display names) |
| `blocklist.go` | Blocklist filtering logic |
| `blocklist_test.go` | Tests for blocklist |
| `notability.go` | Template notability scoring |
| `notability_test.go` | Tests for notability scoring |
| `notability_codecomment_test.go` | Additional tests for notability code comment scoring |
| `official.go` | Official template knowledge management (scans lima-vm/lima git history) |
| `update.go` | Incremental update logic with timestamp-based discovery |
| `update_test.go` | Tests for incremental update logic |

#### `pkg/combiner/`
| File | Purpose |
|------|---------|
| `combiner.go` | Combines templates/repos/orgs into frontend-optimized catalog.jsonl |
| `combiner_test.go` | Tests for combiner |

#### `pkg/minhash/`
| File | Purpose |
|------|---------|
| `minhash.go` | MinHash signature generation for document similarity detection |
| `minhash_test.go` | Tests for MinHash algorithm |
| `lsh.go` | Locality-Sensitive Hashing for efficient similarity search |
| `lsh_test.go` | Tests for LSH implementation |
| `duplicates.go` | Duplicate template detection using LSH and MinHash |
| `duplicates_test.go` | Tests for duplicate detection |
| `example_test.go` | Example usage tests demonstrating MinHash/LSH workflow |

#### `pkg/prompt/`
| File | Purpose |
|------|---------|
| `types.go` | Data structures for LLM prompt generation (TemplateContext, PromptConfig) |
| `builder.go` | Core prompt builder: gathers context, formats prompts for LLM analysis |
| `builder_test.go` | Tests for prompt builder |

#### `pkg/cache/`
| File | Purpose |
|------|---------|
| `cache.go` | Thread-safe in-memory cache with TTL support for API response caching |
| `cache_test.go` | Tests for cache (TTL, cleanup, concurrency) |

#### `pkg/interfaces/`
| File | Purpose |
|------|---------|
| `interfaces.go` | Dependency injection interfaces (HTTPClient, FileSystem, Clock) for testability |

#### `pkg/validation/`
| File | Purpose |
|------|---------|
| `validation.go` | Input validation functions (tokens, paths, template IDs) |
| `validation_test.go` | Tests for validation |

#### `pkg/retry/`
| File | Purpose |
|------|---------|
| `retry.go` | Exponential backoff retry logic for resilient API calls |
| `retry_test.go` | Tests for retry logic |

#### `pkg/config/`
| File | Purpose |
|------|---------|
| `constants.go` | Configuration constants (delays, timeouts, limits) |

### Configuration

| File | Purpose |
|------|---------|
| `config/blocklist.yaml` | Path and repo patterns to exclude from catalog |

### Build & Dependency

| File | Purpose |
|------|---------|
| `go.mod` | Go module definition and dependencies |
| `go.sum` | Dependency checksums |
| `Makefile` | Build and test automation |

## Frontend (JavaScript)

### Main Application

| File | Purpose |
|------|---------|
| `web/index.html` | Main HTML structure with semantic layout and accessibility |
| `web/style.css` | Complete design system (colors, typography, spacing, components) |
| `web/tests.html` | Browser-based test runner page |

### Modular JavaScript (ES6)

| File | Purpose |
|------|---------|
| `web/js/app.js` | Main application initialization and event listener setup |
| `web/js/appActions.js` | Core application actions (filter, render, UI updates) - shared by app.js and keyboard.js |
| `web/js/config.js` | Configuration (data URL, cache keys) |
| `web/js/state.js` | Application state management |
| `web/js/data.js` | Fetch and parse catalog data from data branch (JSONL format) |
| `web/js/filters.js` | Filter and sort templates by keywords, categories, search terms |
| `web/js/sidebar.js` | Sidebar navigation, keyword cloud, category list |
| `web/js/templateCard.js` | Template card rendering with metadata and formatting |
| `web/js/modal.js` | Template preview and help modals with YAML syntax highlighting |
| `web/js/keyboard.js` | Keyboard shortcuts, navigation helpers, and help modal |
| `web/js/theme.js` | Dark/light theme management |
| `web/js/urlHelpers.js` | Lima 2.0 `github:` URL generation and GitHub URL conversions |
| `web/js/utils.js` | Utility functions (debounce, HTML escaping, etc.) |
| `web/js/test-framework.js` | Minimal test framework (assertions, runner) |
| `web/js/*.test.js` | Unit tests for each module |

### Test Infrastructure

| File | Purpose |
|------|---------|
| `test.js` | Node.js test runner with DOM mocking |
| `package.json` | Node.js dependencies for testing (jsdom) |

### Assets

| File | Purpose |
|------|---------|
| `web/favicon.ico` | Favicon |

## Documentation

### Root Documentation

| File | Purpose |
|------|---------|
| `README.md` | Project overview and quick start |
| `ARCHITECTURE.md` | High-level system architecture and design decisions |
| `DEVELOPMENT.md` | Setup, testing, and development workflow |

### Documentation Directory (`docs/`)

| Directory | Purpose |
|-----------|---------|
| `docs/architecture/` | Detailed architecture docs (overview, backend, frontend, data pipeline, future work) |
| `docs/guides/` | How-to guides (getting started, UI/UX guidelines) |
| `docs/reference/` | Reference documentation (code standards, LLM prompts) |
| `docs/research/` | Research findings and decision rationale |
| `docs/history/` | Implementation archive and refactoring history |

### AI Agent Instructions

| File | Purpose |
|------|---------|
| `.claude/instructions.md` | Complete AI workflow, PR checklist, code quality quick reference |

## GitHub Actions

| File | Purpose |
|------|---------|
| `.github/workflows/update-catalog.yml` | Daily automated template discovery and update |
| `.github/workflows/test.yml` | CI testing (Go + JavaScript tests) |

## Data Files (on `data` branch)

| File | Purpose |
|------|---------|
| `templates.jsonl` | All discovered templates with full metadata |
| `repos.jsonl` | Repository metadata |
| `orgs.jsonl` | Organization metadata |
| `catalog.jsonl` | Frontend-optimized combined data |
| `descriptions.jsonl` | LLM-generated descriptions (future, Stage 6) |
| `progress.json` | Pipeline progress tracking for incremental updates |

## Key File Relationships

### Discovery Pipeline Flow
```
discovery.go
  ├─> parser.go (parse YAML)
  ├─> analyzer.go (categorize, keywords)
  ├─> metadata.go (fetch repo/org data)
  ├─> blocklist.go (filter)
  └─> update.go (incremental logic)
      └─> storage.go (save to data branch)
```

### Frontend Data Flow
```
data.js (fetch and parse catalog.jsonl)
  └─> state.js (store in memory)
      └─> app.js (orchestrate UI updates)
          ├─> filters.js (apply filters and sorting)
          ├─> sidebar.js (update keyword cloud and categories)
          └─> templateCard.js (render individual cards)
```

### LLM Prompt Generation Flow *(NEW)*
```
prompt-generator (CLI)
  └─> prompt/builder.go
      ├─> GitHub API (fetch template, repo, org, README)
      ├─> git clone --depth 1 (shallow clone for references)
      ├─> grep -r -B15 -A15 (find template references)
      └─> FormatPrompt() (assemble into structured prompt)
```

## Testing

### Go Tests
```bash
make test          # All tests
go test ./pkg/...  # All package tests
go test ./pkg/discovery/...  # Discovery tests only
go test ./pkg/prompt/...     # Prompt builder tests only
```

### JavaScript Tests
```bash
make test       # Includes JS tests
npm test        # JS tests only
node test.js    # Direct test runner
```

## Building

### Backend
```bash
go build -o lima-catalog ./cmd/lima-catalog
go build -o prompt-generator ./cmd/prompt-generator
```

### Frontend
No build step required - uses ES6 modules directly in browser.

## Module Purpose Summary

### Backend Modules

- **discovery**: Find and validate Lima templates via GitHub Code Search
- **parser**: Parse YAML templates and detect technologies
- **analyzer**: Categorize templates and extract keywords
- **metadata**: Fetch repository and organization metadata from GitHub (with concurrent batching)
- **storage**: Read/write JSON Lines files to data branch
- **combiner**: Merge data into frontend-optimized format
- **github**: Wrap GitHub API with rate limiting and caching
- **cache**: Thread-safe in-memory cache with TTL support
- **interfaces**: Dependency injection interfaces for testability
- **validation**: Input validation functions
- **retry**: Exponential backoff retry logic
- **types**: Shared data structures
- **prompt**: Generate LLM prompts for template analysis

### Frontend Modules

- **app**: Main initialization and event listener setup
- **appActions**: Core application actions (filter, render, UI updates) - shared module to break circular dependencies
- **config**: Configuration constants
- **state**: Centralized application state
- **data**: Fetch and parse catalog data (JSONL)
- **filters**: Template filtering and sorting logic
- **sidebar**: Sidebar navigation, keyword cloud, category list
- **templateCard**: Individual template card rendering
- **modal**: Template preview and help modals with syntax highlighting
- **keyboard**: Keyboard shortcuts, navigation helpers, and help modal
- **theme**: Dark/light theme management
- **urlHelpers**: Lima 2.0 URL generation and GitHub URL conversions
- **utils**: Utility functions (debounce, HTML escaping)

## Recent Additions (This Session)

### New Files
- `pkg/prompt/types.go` - LLM prompt context data structures
- `pkg/prompt/builder.go` - Prompt builder with context gathering
- `pkg/prompt/builder_test.go` - Unit tests for prompt builder
- `cmd/prompt-generator/main.go` - Standalone prompt generation CLI
- `cmd/prompt-generator/README.md` - Prompt generator documentation
- `cmd/prompt-generator/example.sh` - Usage examples
- `docs/reference/llm-prompts.md` - LLM prompt system documentation
- `docs/architecture/source-index.md` - This file

### Modified Files
- `docs/architecture/future-work.md` - Stage 6 prompt builder architecture
- `go.mod` - Downgraded to Go 1.24 for compatibility

## Quick Navigation Tips

**Looking for...**
- Template discovery logic → `pkg/discovery/discovery.go`
- Template parsing → `pkg/discovery/parser.go`
- Categorization → `pkg/discovery/analyzer.go`
- GitHub API calls → `pkg/github/client.go`
- Data storage → `pkg/storage/storage.go`
- Frontend data loading → `web/js/data.js`
- Template rendering → `web/js/templateCard.js`
- Search/filter → `web/js/filters.js` + `web/js/app.js`
- UI design system → `docs/guides/ui-ux-guidelines.md` + `web/style.css`
- LLM prompt generation → `pkg/prompt/builder.go`
- CLI tools → `cmd/*/main.go`

**Working on...**
- New feature → Check `docs/architecture/` for architecture first
- UI changes → Check `docs/guides/ui-ux-guidelines.md` for design system
- Code standards → Check `docs/reference/code-standards.md`
- Testing → See `Makefile` for test commands
- AI workflow → See `.claude/instructions.md`
- LLM integration → See `docs/reference/llm-prompts.md`
