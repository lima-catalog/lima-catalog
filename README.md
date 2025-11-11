# Lima Template Catalog

A tool to discover and catalog Lima VM templates from GitHub.

## Overview

This tool searches GitHub for Lima template files and collects metadata about them, their repositories, and maintainers. The data is stored in JSON Lines format for easy version control and minimal diffs.

## Features

- **Template Discovery**: Finds community templates across GitHub and official templates from lima-vm/lima
- **Metadata Collection**: Gathers repository and organization information
- **Rate Limit Management**: Respects GitHub API rate limits
- **Resumability**: Can resume after interruptions
- **JSON Lines Storage**: Minimal diffs for git-friendly data storage

## Building

```bash
go build -o lima-catalog ./cmd/lima-catalog
```

## Usage

Set your GitHub token:

```bash
export GITHUB_TOKEN=your_token_here
```

Run the collector:

```bash
./lima-catalog
```

The tool will:
1. Discover community templates (excluding lima-vm/lima)
2. Discover official templates from lima-vm/lima
3. Collect metadata for repositories and organizations
4. Save everything to `./data/` directory

### Custom Data Directory

```bash
export DATA_DIR=/path/to/data
./lima-catalog
```

### Incremental Updates

To perform an incremental update (merge new discoveries with existing data):

```bash
export INCREMENTAL=true
./lima-catalog
```

Incremental mode will:
- Load existing template data
- Compare SHAs to detect changes
- Add newly discovered templates
- Update changed templates (preserving discovery date)
- Update last_checked timestamps for unchanged templates
- Merge repository and organization metadata

This is more efficient than a full collection and preserves historical data.

## Accessing the Catalog Data

The collected catalog data is stored in the `data` branch of this repository. You can access it by:

```bash
# Clone the data branch
git clone -b data https://github.com/lima-catalog/lima-catalog.git lima-catalog-data

# Or checkout the data branch in an existing clone
git fetch origin data:data
git checkout data
```

### Data Files

The tool creates these files in the data directory:

- `templates.jsonl` - One template per line (108 templates)
- `repos.jsonl` - One repository per line (36 repositories)
- `orgs.jsonl` - One organization/user per line (35 orgs/users)
- `progress.json` - Progress state for resumability

## Automated Updates

The catalog is automatically updated weekly via GitHub Actions. The workflow:

1. Runs every Sunday at 00:00 UTC
2. Builds the catalog tool
3. Collects latest template data
4. Commits changes to the `data` branch

You can also trigger manual updates:

```bash
# Via GitHub UI: Actions → Update Lima Template Catalog → Run workflow
# Or via GitHub CLI:
gh workflow run update-catalog.yml
```

## Data Format

### Template

```json
{
  "id": "owner/repo/path/to/template.yaml",
  "repo": "owner/repo",
  "path": "path/to/template.yaml",
  "sha": "abc123...",
  "size": 1234,
  "last_modified": "2025-01-15T10:30:00Z",
  "url": "https://github.com/...",
  "discovered_at": "2025-01-20T12:00:00Z",
  "last_checked": "2025-01-20T12:00:00Z",
  "is_official": false
}
```

### Repository

```json
{
  "id": "owner/repo",
  "owner": "owner",
  "name": "repo",
  "description": "Repository description",
  "topics": ["kubernetes", "lima"],
  "stars": 42,
  "forks": 5,
  "watchers": 10,
  "language": "Go",
  "license": "MIT",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2025-01-15T10:30:00Z",
  "pushed_at": "2025-01-15T10:30:00Z",
  "homepage": "https://example.com",
  "is_fork": false,
  "parent": null,
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
  "description": "Org description",
  "location": "San Francisco, CA",
  "blog": "https://example.com",
  "email": "contact@example.com",
  "last_fetched": "2025-01-20T12:00:00Z"
}
```

## Project Structure

```
lima-catalog/
├── cmd/
│   └── lima-catalog/
│       └── main.go          # CLI entry point
├── pkg/
│   ├── types/
│   │   └── types.go         # Data type definitions
│   ├── github/
│   │   └── client.go        # GitHub API client
│   ├── storage/
│   │   └── storage.go       # JSON Lines storage
│   └── discovery/
│       ├── discovery.go     # Template discovery
│       └── metadata.go      # Metadata collection
├── data/                    # Output directory (gitignored)
├── go.mod
├── go.sum
├── PLAN.md                  # Project plan
├── FINDINGS.md              # Research findings
└── README.md
```

## Development

See [PLAN.md](PLAN.md) for the full project plan and [FINDINGS.md](FINDINGS.md) for research findings about GitHub's code search behavior.

## Future Enhancements

- Incremental updates (detect changed templates)
- LLM-based categorization and descriptions
- Web catalog interface
- Template validation
- Quality scoring
