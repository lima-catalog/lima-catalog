# Lima Catalog Source Code Index

Quick reference for all source files and their purposes.

## Backend (Go)

### Command-line Tools

| File | Purpose |
|------|---------|
| `cmd/lima-catalog/main.go` | Main CLI entry point for template discovery and analysis pipeline |
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
| `client.go` | GitHub API wrapper with rate limit management |

#### `pkg/storage/`
| File | Purpose |
|------|---------|
| `storage.go` | JSON Lines storage for templates, repos, orgs (data branch) |

#### `pkg/discovery/`
| File | Purpose |
|------|---------|
| `discovery.go` | Main discovery orchestration |
| `parser.go` | YAML template parsing and technology detection |
| `analyzer.go` | Template analysis and categorization (with LLM hooks) |
| `metadata.go` | Repository and organization metadata fetching |
| `naming.go` | Template name derivation from path |
| `blocklist.go` | Blocklist filtering logic |
| `update.go` | Incremental update logic with timestamp-based discovery |
| `*_test.go` | Tests for each module |

#### `pkg/combiner/`
| File | Purpose |
|------|---------|
| `combiner.go` | Combines templates/repos/orgs into frontend-optimized catalog.jsonl |
| `combiner_test.go` | Tests for combiner |

#### `pkg/prompt/` *(NEW)*
| File | Purpose |
|------|---------|
| `types.go` | Data structures for LLM prompt generation (TemplateContext, PromptConfig) |
| `builder.go` | Core prompt builder: gathers context, formats prompts for LLM analysis |
| `builder_test.go` | Tests for prompt builder |

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
| `docs/index.html` | Main HTML structure with semantic layout and accessibility |
| `docs/style.css` | Complete design system (colors, typography, spacing, components) |
| `docs/tests.html` | Browser-based test runner page |

### Modular JavaScript (ES6)

| File | Purpose |
|------|---------|
| `docs/js/app.js` | Main application orchestration and event handlers |
| `docs/js/config.js` | Configuration (data URL, cache keys) |
| `docs/js/state.js` | Application state management |
| `docs/js/data.js` | Fetch and parse catalog data from data branch (JSONL format) |
| `docs/js/filters.js` | Filter and sort templates by keywords, categories, search terms |
| `docs/js/sidebar.js` | Sidebar navigation, keyword cloud, category list |
| `docs/js/templateCard.js` | Template card rendering with metadata and formatting |
| `docs/js/modal.js` | Template preview and help modals with YAML syntax highlighting |
| `docs/js/keyboard.js` | Keyboard shortcuts, navigation helpers, and help modal |
| `docs/js/theme.js` | Dark/light theme management |
| `docs/js/urlHelpers.js` | Lima 2.0 `github:` URL generation and GitHub URL conversions |
| `docs/js/utils.js` | Utility functions (debounce, HTML escaping, etc.) |
| `docs/js/test-framework.js` | Minimal test framework (assertions, runner) |
| `docs/js/*.test.js` | Unit tests for each module |

### Test Infrastructure

| File | Purpose |
|------|---------|
| `test.js` | Node.js test runner with DOM mocking |
| `package.json` | Node.js dependencies for testing (jsdom) |

### Assets

| File | Purpose |
|------|---------|
| `docs/favicon.ico` | Favicon |

## Documentation

| File | Purpose |
|------|---------|
| `README.md` | Project overview and quick start |
| `PLAN.md` | Architecture, current state, remaining work (Stages 1-7) |
| `IMPLEMENTATION_NOTES.md` | Detailed implementation notes for completed features |
| `INTERFACE_GUIDELINES.md` | Complete UI/UX design system and component specs |
| `FINDINGS.md` | Research findings (GitHub search behavior, scale assessment) |
| `CLAUDE.md` | Instructions for Claude when working on this project |
| `SOURCE_INDEX.md` | This file - quick reference for all source files |
| `LLM_ANALYST_PROMPTS.md` | Documentation for LLM prompt generation system |

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
- **metadata**: Fetch repository and organization metadata from GitHub
- **storage**: Read/write JSON Lines files to data branch
- **combiner**: Merge data into frontend-optimized format
- **github**: Wrap GitHub API with rate limiting
- **types**: Shared data structures
- **prompt**: Generate LLM prompts for template analysis *(NEW)*

### Frontend Modules

- **app**: Main orchestration and event handling
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
- `LLM_ANALYST_PROMPTS.md` - LLM prompt system documentation
- `SOURCE_INDEX.md` - This file

### Modified Files
- `PLAN.md` - Added Stage 6 prompt builder architecture
- `go.mod` - Downgraded to Go 1.24 for compatibility

## Quick Navigation Tips

**Looking for...**
- Template discovery logic → `pkg/discovery/discovery.go`
- Template parsing → `pkg/discovery/parser.go`
- Categorization → `pkg/discovery/analyzer.go`
- GitHub API calls → `pkg/github/client.go`
- Data storage → `pkg/storage/storage.go`
- Frontend data loading → `docs/js/data.js`
- Template rendering → `docs/js/templateCard.js`
- Search/filter → `docs/js/filters.js` + `docs/js/app.js`
- UI design system → `INTERFACE_GUIDELINES.md` + `docs/style.css`
- LLM prompt generation → `pkg/prompt/builder.go`
- CLI tools → `cmd/*/main.go`

**Working on...**
- New feature → Check `PLAN.md` for architecture first
- UI changes → Check `INTERFACE_GUIDELINES.md` for design system
- Testing → See `Makefile` for test commands
- Documentation → Update relevant `.md` files
- LLM integration → See `LLM_ANALYST_PROMPTS.md`
