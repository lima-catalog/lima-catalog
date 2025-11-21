# Getting Started

**Quick Links**: [DEVELOPMENT](../../DEVELOPMENT.md) | [Architecture](../architecture/overview.md) | [Source Index](../architecture/source-index.md)

---

## For New Users

**What is Lima Catalog?**

A searchable catalog of 700+ Lima VM templates from across GitHub. Browse templates by category, filter by keywords, and preview full YAML content before using.

**Live Site**: [lima-catalog.github.io/lima-catalog](https://lima-catalog.github.io/lima-catalog/)

---

## For Contributors

### Contributing Policy

This project is developed by AI agents under human guidance.

- **Feature Requests**: Open an issue
- **Bug Reports**: Open an issue with reproduction steps
- **Pull Requests**: Discuss in an issue first (AI-only development)

→ See [DEVELOPMENT.md](../../DEVELOPMENT.md) for details

---

## For Developers

### Quick Setup

**Prerequisites:**
- Go 1.24+ (backend)
- Node.js 18+ (frontend testing)
- GitHub Personal Access Token

**Clone repository:**
```bash
git clone https://github.com/lima-catalog/lima-catalog.git
cd lima-catalog
```

**Install dependencies:**
```bash
# Go dependencies
go mod download

# Node.js test dependencies
npm install
```

**Set environment variables:**
```bash
export GITHUB_TOKEN=your_personal_access_token
export ANALYZE=true  # Enable template analysis
```

### Running Tests

**All tests:**
```bash
make test
```

**Backend only:**
```bash
go test ./pkg/...
```

**Frontend only:**
```bash
npm test
```

### Testing Frontend Locally

**Start local server:**
```bash
cd web
python3 -m http.server 8000
# Visit: http://localhost:8000
```

**Important**: Never open `index.html` directly - ES6 modules require a web server.

### Running Discovery

**Full discovery:**
```bash
export GITHUB_TOKEN=your_token
export ANALYZE=true
./lima-catalog
```

**Incremental (faster):**
```bash
export GITHUB_TOKEN=your_token
export ANALYZE=true
export INCREMENTAL=true
./lima-catalog
```

### Building

**Backend:**
```bash
go build -o lima-catalog ./cmd/lima-catalog
```

**Frontend:**
No build step - uses ES6 modules directly.

---

## Understanding the Codebase

### Architecture Overview

**Three main components:**
1. **Backend (Go)** - Discovers, validates, and analyzes templates
2. **Data Storage** - JSON Lines on `data` branch
3. **Frontend (Web)** - Static GitHub Pages site

→ See [Architecture Overview](../architecture/overview.md)

### Project Structure

```
lima-catalog/
├── cmd/                  # CLI tools
│   ├── lima-catalog/     # Main discovery tool
│   └── prompt-generator/ # LLM prompt generator
├── pkg/                  # Go packages
│   ├── discovery/        # Template discovery & analysis
│   ├── storage/          # JSON Lines storage
│   ├── github/           # GitHub API wrapper
│   └── ...
├── web/                  # GitHub Pages site
│   ├── index.html
│   ├── js/               # ES6 modules
│   └── style.css
├── docs/                 # Project documentation
└── config/               # Configuration files
```

→ See [Source Index](../architecture/source-index.md) for complete file reference

### Documentation Map

**By Role:**
- **AI Agent**: [.claude/instructions.md](../../.claude/instructions.md)
- **Developer**: [DEVELOPMENT.md](../../DEVELOPMENT.md)
- **Understanding System**: [ARCHITECTURE.md](../../ARCHITECTURE.md)

**By Topic:**
- **Architecture**: [docs/architecture/](../architecture/)
- **Guides**: [docs/guides/](../guides/)
- **Reference**: [docs/reference/](../reference/)

---

## Common Development Tasks

### Adding a New Feature

1. Review [Future Work](../architecture/future-work.md) for planned features
2. Check relevant design doc ([Backend](../architecture/backend-design.md) or [Frontend](../architecture/frontend-design.md))
3. Follow [Code Standards](../reference/code-standards.md)
4. Write tests (required!)
5. Update documentation

### Fixing a Bug

1. Write a failing test that reproduces the bug
2. Fix the bug
3. Verify test passes
4. Update relevant docs if needed

### Modifying UI

1. Edit `web/` files (HTML/CSS/JS)
2. Follow [UI/UX Guidelines](ui-ux-guidelines.md)
3. Include accessibility features (ARIA labels, roles)
4. Test in browser with local server
5. Remember browser cache (hard refresh may be needed)

### Modifying Backend

1. Edit `pkg/` or `cmd/` files
2. Follow dependency injection patterns
3. Add tests in `*_test.go` files
4. Build to verify: `go build`
5. Run tests: `make test`

---

## Next Steps

- **Read [ARCHITECTURE.md](../../ARCHITECTURE.md)** for system overview
- **Review [Code Standards](../reference/code-standards.md)** for quality requirements
- **Explore [Source Index](../architecture/source-index.md)** to find files
- **Check [Future Work](../architecture/future-work.md)** for contribution ideas

---

## Getting Help

- **Documentation**: Browse [docs/](../)
- **Issues**: Check existing issues or create new one
- **Architecture**: See [docs/architecture/](../architecture/)
- **Code Questions**: See [Source Index](../architecture/source-index.md)
