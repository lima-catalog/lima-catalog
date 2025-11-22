# Lima Catalog Architecture

**Quick Links**: [Development Guide](DEVELOPMENT.md) | [Detailed Docs](docs/) | [Source Index](docs/architecture/source-index.md)

---

## Overview

Lima Catalog is a searchable catalog of 700+ Lima VM templates from across GitHub, featuring automated discovery, smart categorization, and a beautiful web interface.

**Live Site**: [lima-catalog.github.io/lima-catalog](https://lima-catalog.github.io/lima-catalog/)

## System Components

### 1. Backend (Go)
**Purpose**: Discover, validate, and analyze Lima templates

- **Discovery**: GitHub Code Search finds templates across repos
- **Validation**: Content-based filtering eliminates false positives
- **Analysis**: Extract keywords, categories, notability scores
- **Metadata**: Fetch repository and organization information
- **Duplicate Detection**: MinHash + LSH identifies similar templates

**Key Technologies**: Go 1.24, GitHub API v3, YAML parsing, MinHash

→ [Backend Design Details](docs/architecture/backend-design.md)

### 2. Data Storage
**Purpose**: Persist catalog data efficiently

- **Format**: JSON Lines (one object per line)
- **Location**: `data` branch (separate from code)
- **Files**: `templates.jsonl`, `repos.jsonl`, `orgs.jsonl`, `catalog.jsonl`
- **Updates**: Incremental (only changed templates re-analyzed)

→ [Data Pipeline Details](docs/architecture/data-pipeline.md)

### 3. Frontend (Static Web)
**Purpose**: Browse and search templates

- **Platform**: GitHub Pages (static HTML/CSS/JS)
- **Architecture**: Modular ES6 JavaScript (no build step)
- **Features**: Multi-keyword filtering, category browsing, template preview
- **UI**: Responsive design with dark/light themes

→ [Frontend Design Details](docs/architecture/frontend-design.md)

## Data Flow

```
┌─────────────────┐
│ GitHub Repos    │
└────────┬────────┘
         │
         ↓ (Daily GitHub Actions)
┌─────────────────┐
│   Discovery     │  ← Search for templates
│   (Go backend)  │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│   Validation    │  ← Check images: field
│   & Analysis    │  ← Extract keywords
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│  Data Storage   │  ← JSON Lines on data branch
│  (data branch)  │
└────────┬────────┘
         │
         ↓ (Fetch at runtime)
┌─────────────────┐
│   Frontend      │  ← Filter & display
│ (GitHub Pages)  │
└─────────────────┘
```

## Design Principles

1. **Incremental Updates**
   - Only re-analyze changed templates (SHA-based detection)
   - Refresh 5% of stale metadata per run
   - Spreads API load over time

2. **Scalability**
   - Sub-linear duplicate detection (LSH)
   - Efficient JSON Lines format
   - Concurrent metadata fetching

3. **No Build Step**
   - Frontend uses ES6 modules directly
   - Backend is single Go binary
   - Simple deployment

4. **Data Isolation**
   - Code on `main` branch
   - Data on `data` branch
   - Clean separation of concerns

## Current State (November 2025)

**Production**:
- ✅ 716 templates cataloged (51 official + 665 community)
- ✅ Daily automated updates
- ✅ Smart categorization with keyword extraction
- ✅ Duplicate detection with MinHash + LSH
- ✅ Responsive web interface with preview modal
- ✅ 60%+ backend test coverage (83 tests passing)

**Data Pipeline**:
- ✅ Stage 1-5 complete (Discovery → Analysis → Frontend)
- ⏳ Stage 6 planned (LLM descriptions)
- ⏳ Stage 7 planned (Template cleanup)

## Learn More

### Development
- **[Development Guide](DEVELOPMENT.md)** - Setup, testing, contributing
- **[Code Standards](docs/reference/code-standards.md)** - Quality requirements
- **[Getting Started](docs/guides/getting-started.md)** - First-time setup

### Architecture
- **[Overview](docs/architecture/overview.md)** - Detailed architecture
- **[Backend Design](docs/architecture/backend-design.md)** - Go patterns, interfaces, testing
- **[Frontend Design](docs/architecture/frontend-design.md)** - JavaScript modules, UI patterns
- **[Data Pipeline](docs/architecture/data-pipeline.md)** - Discovery → Analysis → Frontend flow
- **[Source Index](docs/architecture/source-index.md)** - Find any source file quickly

### Reference
- **[UI/UX Guidelines](docs/guides/ui-ux-guidelines.md)** - Design system
- **[LLM Prompts](docs/reference/llm-prompts.md)** - Prompt generation system
- **[Future Work](docs/architecture/future-work.md)** - Planned features

### Historical
- **[Architecture Analysis (Nov 2025)](docs/research/architecture-analysis-2025-11.md)** - Comprehensive architectural review with improvement suggestions
- **[Research](docs/research/)** - Decision rationale and findings
- **[History](docs/history/)** - Implementation archive

## Technology Stack

**Backend**:
- Go 1.24
- GitHub API (REST v3)
- YAML parsing (gopkg.in/yaml.v3)
- MinHash + LSH for duplicate detection

**Frontend**:
- Vanilla JavaScript (ES6 modules)
- Highlight.js for syntax highlighting
- CSS Grid + Flexbox
- No framework dependencies

**Infrastructure**:
- GitHub Actions (daily automation)
- GitHub Pages (static hosting)
- JSON Lines (data storage)

## Quick Facts

- **716 templates** currently cataloged
- **~1.3KB per template** average size
- **~732KB** total data (including MinHash signatures)
- **~2-5 minutes** typical daily update runtime
- **Sub-linear search** for duplicate detection
- **Zero dependencies** for frontend (pure ES6)

---

**For detailed technical information**, see [docs/architecture/](docs/architecture/)

**For development workflow**, see [DEVELOPMENT.md](DEVELOPMENT.md)

**For AI agents**, see [.claude/instructions.md](.claude/instructions.md)
