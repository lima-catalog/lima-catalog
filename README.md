# Lima Template Catalog

Discover and browse 700+ [Lima VM](https://lima-vm.io/) templates from across GitHub.

## 🌐 Browse the Catalog

**[lima-catalog.github.io/lima-catalog](https://lima-catalog.github.io/lima-catalog/)**

Search, filter, and preview Lima templates with:
- **Multi-keyword filtering** - Find templates by technology (e.g., "alpine" + "docker" + "k8s")
- **Category browsing** - Browse by containers, development, orchestration, security, etc.
- **Template preview** - View YAML content with syntax highlighting without leaving the page
- **Lima 2.0 URLs** - One-click copy of shortest `github:` URLs for instant use

## What's Inside

- **700+ templates** from across GitHub
  - 51 official templates from lima-vm/lima
  - 650+ community templates from independent repositories
- **Daily automated updates** to discover new templates
- **Smart categorization** with automatic keyword extraction
- **Rich metadata** for each template and repository

## For Template Authors

Want your Lima template included in the catalog? Just create a valid Lima template (with `images:` field) in a public GitHub repository - it will be automatically discovered within 24 hours!

## Project Architecture

This project consists of two main components:

### 1. Web Catalog (GitHub Pages)

A static website (`docs/`) that fetches data from the `data` branch and provides an interactive interface to browse templates.

### 2. Backend Tool (Go)

A CLI tool that:
- Discovers templates via GitHub Code Search
- Collects metadata for repositories and organizations
- Analyzes templates to extract categories and keywords
- Stores data in JSON Lines format for minimal git diffs
- Runs daily via GitHub Actions to keep catalog fresh

## For Developers

### Building the Backend Tool

```bash
go build -o lima-catalog ./cmd/lima-catalog
```

### Running Template Collection

```bash
export GITHUB_TOKEN=your_token_here
export ANALYZE=true  # Enable template analysis
./lima-catalog
```

The tool will discover templates, collect metadata, analyze content, and save to `./data/`.

### Options

- `INCREMENTAL=true` - Merge with existing data (faster, preserves history)
- `ANALYZE=true` - Parse templates and extract metadata
- `DATA_DIR=/path` - Use custom output directory

See [PLAN.md](PLAN.md) for detailed architecture and implementation notes.

## Data Access

Catalog data is stored in the `data` branch:

```bash
# Clone data branch
git clone -b data https://github.com/lima-catalog/lima-catalog.git lima-catalog-data
cd lima-catalog-data/data

# View templates
cat templates.jsonl | jq .
```

### Data Files

- `templates.jsonl` - Template metadata (700+ templates)
- `repos.jsonl` - Repository information
- `orgs.jsonl` - Organization/user information
- `progress.json` - Collection state

Each file uses JSON Lines format (one JSON object per line) for minimal diffs and easy processing.

## Documentation

- **[PLAN.md](PLAN.md)** - Project architecture and implementation details
- **[FINDINGS.md](FINDINGS.md)** - Research findings on GitHub search behavior
- **[CLAUDE.md](CLAUDE.md)** - Development workflow and session instructions
- **[experiments/](experiments/)** - Historical research scripts

## How It Works

1. **Discovery**: GitHub Code Search finds templates with `minimumLimaVersion` or Lima structure
2. **Validation**: Content-based filtering ensures files are valid Lima templates
3. **Analysis**: YAML parsing extracts OS images, technologies, and patterns
4. **Categorization**: Auto-assigns categories based on detected technologies
5. **Storage**: JSON Lines format enables efficient git-based storage
6. **UI**: Static GitHub Pages site fetches data and provides rich browsing experience

## License

MIT
