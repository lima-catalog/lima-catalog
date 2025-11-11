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

## Output Files

The tool creates these files in the data directory:

- `templates.jsonl` - One template per line
- `repos.jsonl` - One repository per line
- `orgs.jsonl` - One organization/user per line
- `progress.json` - Progress state for resumability

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
