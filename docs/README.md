# Lima Catalog Documentation

**Quick Links**: [ARCHITECTURE](../ARCHITECTURE.md) | [DEVELOPMENT](../DEVELOPMENT.md) | [.claude/instructions](.../.claude/instructions.md)

---

## Documentation by Role

### 🤖 AI Agents
**Start here**: [.claude/instructions.md](../.claude/instructions.md)
- PR workflow (critical!)
- Quick reference links
- Code quality standards
- Common workflows

### 👤 Developers
**Start here**: [DEVELOPMENT.md](../DEVELOPMENT.md)
- Setup instructions
- Running tests
- Contributing policy (AI-only development)

### 📖 Understanding the System
**Start here**: [ARCHITECTURE.md](../ARCHITECTURE.md)
- High-level overview
- System components
- Data flow diagram

---

## Documentation by Topic

### Architecture
- **[Overview](architecture/overview.md)** - Detailed current architecture
- **[Backend Design](architecture/backend-design.md)** - Go patterns, testing, quality standards
- **[Frontend Design](architecture/frontend-design.md)** - JavaScript modules, UI patterns
- **[Data Pipeline](architecture/data-pipeline.md)** - Discovery → Analysis → Frontend (Stages 1-5)
- **[Future Work](architecture/future-work.md)** - Planned features (Stages 6-7)
- **[Source Index](architecture/source-index.md)** - Find any source file quickly

### Guides
- **[Getting Started](guides/getting-started.md)** - First-time setup
- **[UI/UX Guidelines](guides/ui-ux-guidelines.md)** - Complete design system

### Reference
- **[Code Standards](reference/code-standards.md)** - Backend quality requirements
- **[LLM Prompts](reference/llm-prompts.md)** - Prompt generation system documentation

---

## Historical Documentation

### Research (Decision Rationale)
Research findings that inform current design - useful for future decisions:

- **[Architecture Analysis (Nov 2025)](research/architecture-analysis-2025-11.md)** - Comprehensive architectural review with improvement suggestions and new feature recommendations
- **[GitHub Search Behavior](research/github-search-behavior.md)** - Why we search the way we do
- **[Duplicate Detection Algorithms](research/duplicate-detection-algorithms.md)** - Why MinHash + LSH was chosen

### History (Implementation Archive)
Detailed implementation logs - useful for understanding evolution:

- **[Backend Refactoring](history/backend-refactoring/)** - 6-phase refactoring (complete)
  - [Refactoring Plan](history/backend-refactoring/plan.md)
  - [Code Review 2025-01](history/backend-refactoring/code-review-2025-01.md)

- **[Incremental Updates](history/incremental-updates/)** - Stages 1-5 implementation
- **[UI Redesign](history/ui-redesign/)** - Frontend improvements
- **[Migrations](history/migrations/)** - Data migrations and refactoring

---

## Quick Navigation

**Looking for...**

| Need | Go To |
|------|-------|
| How templates are discovered | [Data Pipeline](architecture/data-pipeline.md#stage-1-discovery) |
| Backend patterns and standards | [Backend Design](architecture/backend-design.md) |
| UI component guidelines | [UI/UX Guidelines](guides/ui-ux-guidelines.md) |
| Where a file lives | [Source Index](architecture/source-index.md) |
| Architectural review & suggestions | [Architecture Analysis](research/architecture-analysis-2025-11.md) |
| Why we chose X over Y | [Research](research/) |
| How feature X was built | [History](history/) |
| Future plans | [Future Work](architecture/future-work.md) |

---

## Documentation Standards

When updating documentation:
- **Architecture changes** → Update relevant docs in `architecture/`
- **New features** → Add to `architecture/future-work.md` or mark complete in `architecture/overview.md`
- **Implementation details** → Add to `history/` for reference
- **UI/UX patterns** → Update `guides/ui-ux-guidelines.md`
- **Code patterns** → Update `reference/code-standards.md`

---

**For AI agents**: See [.claude/instructions.md](../.claude/instructions.md) for complete workflow
