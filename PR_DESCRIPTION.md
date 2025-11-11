# Lima Template Catalog Tool

## Summary

Implements a complete Lima template catalog system that discovers, collects, and maintains metadata about Lima VM templates from GitHub.

## What's Included

### Core Tool (Go)
- **Template Discovery**: Finds 108 templates (51 official + 57 community)
- **Metadata Collection**: Gathers data for 36 repos and 35 organizations
- **Smart Storage**: JSON Lines format for minimal git diffs
- **Incremental Updates**: Detects changes by comparing SHAs
- **Efficient**: Uses only ~140 API calls (2.8% of hourly quota)

### Automation
- **GitHub Actions**: Weekly automated updates (Sundays at 00:00 UTC)
- **Incremental Mode**: Only processes what changed
- **Auto-commit**: Pushes updates to data branch

### Documentation
- `PLAN.md` - Detailed project architecture
- `FINDINGS.md` - GitHub fork search research
- `TEST_RESULTS.md` - Testing validation
- `SUMMARY.md` - Complete implementation summary
- `README.md` - Full usage guide

## Catalog Statistics

- **108 total templates** discovered
  - 51 official from lima-vm/lima
  - 57 community from 35 independent repos
- **36 unique repositories**
- **35 unique owners/organizations**

## Data Storage

Catalog data is stored in a separate `data` branch (currently `claude/data-011CV1boPxnrfUrc3BZHFFpQ`) containing:
- `data/templates.jsonl` - Template metadata
- `data/repos.jsonl` - Repository information
- `data/orgs.jsonl` - Organization/user information
- `data/progress.json` - Collection state

## Usage

```bash
# Build the tool
go build -o lima-catalog ./cmd/lima-catalog

# Full collection
export GITHUB_TOKEN=your_token
./lima-catalog

# Incremental update
export INCREMENTAL=true
./lima-catalog
```

## Testing

✅ All functionality tested end-to-end:
- Template discovery working
- Metadata collection complete
- Incremental updates functional
- GitHub Actions workflow ready

See `TEST_RESULTS.md` for detailed test results.

## Architecture

Two separate branches:

1. **Code branch** (this PR): `claude/lima-template-catalog-tool-011CV1boPxnrfUrc3BZHFFpQ` → `main`
   - Contains the catalog tool, workflows, and documentation
   - Should be merged to `main`

2. **Data branch** (separate): `claude/data-011CV1boPxnrfUrc3BZHFFpQ`
   - Contains the collected catalog data
   - Kept separate from main
   - Auto-updated by GitHub Actions
   - Could be renamed to just `data` for simplicity

## Next Steps After Merge

1. Optionally rename data branch: `claude/data-011CV1boPxnrfUrc3BZHFFpQ` → `data`
2. Update GitHub Actions workflow to reference correct data branch name
3. GitHub Actions will automatically update catalog weekly
4. Consider future enhancements (LLM categorization, web interface, etc.)
