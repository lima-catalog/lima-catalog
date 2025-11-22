# Development Guide

**Quick Links**: [Architecture](ARCHITECTURE.md) | [Code Standards](docs/reference/code-standards.md) | [Getting Started](docs/guides/getting-started.md)

---

## Contributing

This project is developed by AI agents under human guidance.

### Feature Requests & Bug Reports
- **Feature Requests**: Open an issue describing your idea
- **Bug Reports**: Open an issue with reproduction steps
- **Pull Requests**: Please discuss in an issue first

**Note**: Implementation is done by AI. PRs from humans are welcome but should be discussed and agreed upon in advance.

### For AI Agents

Detailed development workflow and instructions are in [`.claude/instructions.md`](.claude/instructions.md).

---

## Development Setup

### Prerequisites

**Backend (Go)**:
```bash
# Requires Go 1.24+
go version

# Install dependencies
go mod download
```

**Frontend (JavaScript)**:
```bash
# Requires Node.js for testing
node --version  # 18+ recommended
```

### Environment Variables

**Required for backend**:
```bash
export GITHUB_TOKEN=your_personal_access_token
export ANALYZE=true  # Enable template analysis
```

**Required for Lima template embedding**:
```bash
# Clone Lima repository for template reference resolution
git clone https://github.com/lima-vm/lima.git /tmp/lima

# Point to Lima templates directory
export LIMA_TEMPLATES_PATH=/tmp/lima/templates
```

This enables proper parsing of templates with base template references like `base: template:_images/ubuntu`.

**Optional**:
```bash
export DATA_DIR=./data        # Custom data directory
export INCREMENTAL=true       # Incremental updates (faster)
```

---

## CI/CD

### Continuous Integration

The repository uses GitHub Actions to run automated tests on all pull requests:

- **Test**: Runs `go test` with race detection and coverage
- **Build**: Compiles `lima-catalog` to verify successful builds
- **Lint**: Runs `go vet` and `golangci-lint` for code quality

See `.github/workflows/ci.yml` for the full workflow configuration.

**Branch Protection**: The `main` branch should be protected to require all CI checks to pass before merging. This prevents compilation errors and test failures from being merged.

---

## Running Tests

### All Tests (Recommended)
```bash
make test
```

This runs:
- Go backend tests (83 tests)
- JavaScript frontend tests (220 tests)

### Backend Tests Only
```bash
go test ./pkg/...
```

### Frontend Tests Only
```bash
make test-js
# or
node test.js
```

### Browser Tests (Optional)
```bash
# Start local server
cd web
python3 -m http.server 8000

# Visit http://localhost:8000/tests.html
```

### Test Coverage

Current coverage:
- **Backend:** 49.7% (83 tests) | Target: 70%+
- **Frontend:** ~82% (220 tests, 7 fully + 4 partially) | Target: 95%+

For detailed test coverage plans and progress tracking, see:
- **[Backend Test Coverage Plan](docs/testing/test-coverage-plan.md)** - Backend improvement roadmap
- **[Frontend Test Coverage Plan](docs/testing/test-coverage-plan-frontend.md)** - Frontend improvement roadmap

---

## Building

### Backend
```bash
# Main data collection tool
go build -o lima-catalog ./cmd/lima-catalog

# Prompt generator (for LLM testing)
go build -o prompt-generator ./cmd/prompt-generator
```

### Frontend
No build step required - uses ES6 modules directly in browser.

---

## Running Locally

### Testing Web App
```bash
# Option 1: Python
cd web
python3 -m http.server 8000
# Visit: http://localhost:8000

# Option 2: Node.js
npx serve web
# Visit: http://localhost:3000
```

**Important**: Never open `index.html` directly - ES6 modules require a web server.

### Running Discovery
```bash
export GITHUB_TOKEN=your_token
export ANALYZE=true
./lima-catalog
```

The tool will:
1. Discover templates via GitHub Code Search
2. Validate content (check for `images:` field)
3. Analyze templates (extract keywords, categories)
4. Fetch metadata (repos, orgs)
5. Detect duplicates (MinHash + LSH)
6. Save to `./data/` (JSON Lines format)

---

## Project Structure

```
lima-catalog/
├── cmd/                   # Command-line tools
│   ├── lima-catalog/      # Main discovery tool
│   └── prompt-generator/  # LLM prompt generator
├── pkg/                   # Go packages
│   ├── discovery/         # Template discovery & analysis
│   ├── storage/           # JSON Lines storage
│   ├── github/            # GitHub API wrapper
│   ├── combiner/          # Frontend data generation
│   └── ...
├── web/                   # GitHub Pages site
│   ├── index.html
│   ├── js/                # ES6 modules
│   └── style.css
├── docs/                  # Project documentation
│   ├── guides/
│   ├── architecture/
│   ├── reference/
│   └── ...
├── config/                # Configuration files
│   └── blocklist.yaml     # Filter patterns
└── .claude/               # AI agent instructions
    └── instructions.md
```

For detailed file purposes, see [Source Index](docs/architecture/source-index.md).

---

## Common Tasks

### Adding a New Feature
1. Check [Future Work](docs/architecture/future-work.md) for planned features
2. Review [Backend Design](docs/architecture/backend-design.md) or [Frontend Design](docs/architecture/frontend-design.md)
3. Follow [Code Standards](docs/reference/code-standards.md)
4. Write tests (see [Testing Guide](#running-tests))

### Modifying UI
1. Edit `web/` files (HTML/CSS/JS)
2. Always include accessibility features (aria-label, role, etc.)
3. Follow [UI/UX Guidelines](docs/guides/ui-ux-guidelines.md)
4. Test in browser with local server
5. Remember: Changes visible after GitHub Pages deploys

### Modifying Backend
1. Edit `pkg/` or `cmd/` files
2. Follow dependency injection patterns (see [Backend Design](docs/architecture/backend-design.md))
3. Add tests in `*_test.go` files
4. Build to verify: `go build -o lima-catalog ./cmd/lima-catalog`
5. Run tests: `make test`

### Updating Keywords/Categories
- Changes to `parser.go` or `analyzer.go` only affect NEW templates
- Existing templates keep current analysis until SHA changes
- To force re-analysis, clear `AnalyzedAt` timestamps (rarely needed)

---

## Documentation

### For Developers
- [Architecture Overview](ARCHITECTURE.md) - System design
- [Getting Started](docs/guides/getting-started.md) - First-time setup
- [Backend Design](docs/architecture/backend-design.md) - Go patterns
- [Frontend Design](docs/architecture/frontend-design.md) - JavaScript modules
- [Code Standards](docs/reference/code-standards.md) - Quality requirements

### For AI Agents
- [.claude/instructions.md](.claude/instructions.md) - Workflow and reminders
- [Source Index](docs/architecture/source-index.md) - Find any file quickly

### Historical
- [Research](docs/research/) - Decision rationale
- [History](docs/history/) - Implementation archive

---

## Getting Help

- **Documentation**: Browse [docs/](docs/)
- **Issues**: Check existing issues or create new one
- **Architecture Questions**: See [ARCHITECTURE.md](ARCHITECTURE.md)
- **Code Questions**: See [Source Index](docs/architecture/source-index.md)

---

## Code Quality

All code must meet these standards:

- **Tests**: >60% coverage, all tests passing
- **Go**: Follow [Backend Code Standards](docs/reference/code-standards.md)
- **JavaScript**: Follow [UI/UX Guidelines](docs/guides/ui-ux-guidelines.md)
- **Accessibility**: All UI changes must include ARIA attributes

For detailed standards, see [docs/reference/code-standards.md](docs/reference/code-standards.md).
