# E2E Test Fixtures

This directory contains test fixtures for E2E (end-to-end) tests using Playwright.

## Overview

The E2E tests need access to template YAML files to test the modal content loading functionality. Rather than making real network requests to GitHub during tests (which would be slow, flaky, and require internet access), we extract a representative subset of templates from the catalog and serve them locally during tests.

## Directory Structure

```
tests/e2e/fixtures/
├── README.md           # This file
├── manifest.json       # Maps template IDs to local fixture files
└── templates/          # Directory containing YAML fixture files
    ├── template1.yaml
    ├── template2.yaml
    └── ...
```

## How It Works

1. **Generation**: The `scripts/create-sample-fixtures.js` script analyzes the catalog and selects representative templates
2. **Storage**: Sample YAML files are generated and stored in `tests/e2e/fixtures/templates/`
3. **Manifest**: A `manifest.json` file maps template raw URLs to local fixture filenames
4. **Serving**: During tests, `tests/e2e/fixtures.js` intercepts GitHub requests and serves local fixtures instead

## Generating Fixtures

To regenerate the test fixtures:

```bash
# Option 1: Use sample fixtures (recommended for CI/local development)
node scripts/create-sample-fixtures.js /path/to/catalog.jsonl

# Option 2: Download real YAML files (requires internet access)
node scripts/create-test-fixtures.js /path/to/catalog.jsonl
```

### Template Selection Criteria

The sample fixture generation script selects templates to ensure comprehensive test coverage:

- **First 3 templates** from the catalog (for index-based tests)
- **First 5 templates** alphabetically by name (for default UI sorting)
- **2 official templates** (from lima-vm/lima)
- **2 templates with similar templates** (for diff testing)
- **Diverse categories** (development, containers, orchestration, etc.)
- **High notability scores** (well-maintained, complex templates)

Total: 20 representative templates covering various edge cases

## Manifest Format

The `manifest.json` file contains metadata about each fixture:

```json
{
  "generated_at": "2025-11-23T12:34:56.789Z",
  "template_count": 20,
  "templates": {
    "org/repo/path/to/template.yaml": {
      "filename": "org_repo_path_to_template.yaml.yaml",
      "raw_url": "https://raw.githubusercontent.com/org/repo/branch/path/to/template.yaml",
      "github_url": "github:org/repo/path/to/template",
      "name": "Template Name",
      "category": "development",
      "official": false,
      "has_similars": true,
      "similar_count": 3,
      "keywords": ["ubuntu", "docker"],
      "notability_score": 123.4
    }
  }
}
```

## Maintenance

### When to Regenerate

Regenerate fixtures when:
- The catalog schema changes
- Tests fail due to missing fixtures
- Adding new test cases that require specific templates
- Catalog data format changes significantly

### CI/CD Integration

The fixtures are checked into version control, so they're available in CI without regeneration. To update them in CI:

```yaml
- name: Update test fixtures
  run: |
    git fetch origin data:refs/remotes/origin/data
    git show origin/data:data/catalog.jsonl > /tmp/catalog.jsonl
    node scripts/create-sample-fixtures.js /tmp/catalog.jsonl
```

## Testing

Verify fixtures work correctly:

```bash
# Run all E2E tests
npm run test:e2e

# Run only modal tests (which use fixtures)
npx playwright test tests/e2e/modal.spec.js
```

## Troubleshooting

### Test fails with "No matching template"

The test is trying to load a template that isn't in the fixtures. Either:
1. Regenerate fixtures with more templates
2. Update the test to use a template that's in the fixtures
3. Add the specific template manually to the fixtures

### Fixture file not found

Ensure the `manifest.json` and `templates/` directory are in sync. Regenerate fixtures if needed.

### Tests timeout waiting for content

The fixture may not be served correctly. Check:
1. The `raw_url` in the manifest matches the catalog
2. The filename exists in `tests/e2e/fixtures/templates/`
3. The `tests/e2e/fixtures.js` is correctly loading the manifest
