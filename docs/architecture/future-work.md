# Future Work

**Quick Links**: [Overview](overview.md) | [Data Pipeline](data-pipeline.md) | [Backend Design](backend-design.md)

---

## Overview

This document outlines planned enhancements to the Lima Catalog system (Stages 6-7 and beyond).

**Completed**: Stages 1-5 (Discovery → Analysis → Frontend)

**Planned**: Stages 6-7 (LLM Descriptions, Template Cleanup)

→ For completed stages, see [Data Pipeline](data-pipeline.md)

---

## Stage 6: LLM Descriptions (Optional Enhancement)

**Goal**: Generate quality descriptions for templates using AI

**Status**: ⏳ Planned

### Motivation

Current descriptions are keyword-based. LLM-generated descriptions would:
- Provide natural language explanations
- Capture template purpose and use cases
- Improve searchability and discoverability
- Enhance user experience

### Data Schema

**New file**: `descriptions.jsonl`

```json
{
  "template_id": "owner/repo/path/to/template.yaml",
  "short_description": "Brief one-liner (max 100 chars)",
  "long_description": "Detailed explanation (max 500 chars)",
  "keywords": ["keyword1", "keyword2"],
  "generated_at": "2024-03-20T10:00:00Z",
  "source_hash": "abc123...",
  "llm_model": "claude-3-haiku-20240307"
}
```

### Generation Logic

Generate description only if:
1. No description exists, OR
2. source_hash doesn't match (data changed), AND
3. Template not in blocklist, AND
4. Template does not have `meta.description` (author provided), AND
5. Template does not have `meta.noindex: true` (author opted out)

**Rate limit**: Start with 1 description/run (configurable)

**Source hash**: Computed from template + repo + org data (excludes timestamps)

### Prompt Builder Architecture

**Package**: `pkg/prompt`

Context included in prompts:
1. **Template YAML content** - Full template with scripts, probes, messages
2. **Template comments** - Extracted for author intent
3. **Repository metadata** - Description, topics, stars, language
4. **Organization info** - Owner name, type, description, location
5. **README content** - Repository README (truncated to ~5000 chars)
6. **Template references** - Shallow clone + grep with 15 lines context

**Important caveats in prompts:**
- Warns LLM that repo purpose may differ from template purpose
- Repository might be CI scaffolding, docs, or template collection
- Template itself may make repo's project available in VM
- Instructs LLM to prioritize template content over repo metadata

→ See [LLM Prompts Documentation](../reference/llm-prompts.md) for details

### Developer Tools

**CLI tool**: `cmd/prompt-generator`

Purpose: Test prompts with different LLM models before backend integration

Usage:
```bash
export GITHUB_TOKEN=your_token
prompt-generator lima-vm/lima/templates/ubuntu.yaml
```

Features:
- Gathers comprehensive context
- Outputs formatted prompt ready for LLM testing
- Custom templates via `-template` flag or `PROMPT_TEMPLATE` env var
- Token estimation for cost planning
- Configurable context lines, max README length, max references

### Environment Variables

```bash
LLM_API_KEY=<api-key>
LLM_MODEL=claude-3-haiku-20240307
LLM_MAX_DESCRIPTIONS_PER_RUN=1  # Start conservative
LLM_PROVIDER=anthropic  # or openai, etc.
```

### Integration with Stage 5

After implementing Stage 6, update combiner to use description priority:
1. `meta.description` (author-provided, if available)
2. LLM-generated description
3. Analysis-based keywords (current fallback)

### Error Handling

- Log failures but don't block pipeline
- Continue with analysis-based data if LLM fails
- Retry failed descriptions next run

---

## Stage 7: Template Cleanup (Future)

**Goal**: Remove templates that no longer exist

**Status**: ⏳ Planned

### Deletion Detection

**Strategy**: Conservative multi-check approach

1. Check templates not updated in 14+ days
2. Fetch template URL (HEAD request for efficiency)
3. Mark as failed if 404/403/500 received
4. Retry logic: Check again after 7 days, then 14 days
5. Delete after 3 consecutive failures (total 35 days)

### New Fields

```yaml
template:
  last_check_failed: "2024-03-01T00:00:00Z"  # First failure timestamp
  check_failures: 2                           # Consecutive failure count
  pending_deletion: false                     # Flagged for removal
```

### Orphan Cleanup

After template deletion, clean up orphaned metadata:

```python
# After template deletion
active_repos = set(t.repo for t in templates)
active_orgs = set(t.org for t in templates)

repos = [r for r in repos if r.id in active_repos]
orgs = [o for o in orgs if o.id in active_orgs]
descriptions = [d for d in descriptions if d.template_id in active_templates]
```

### Meta.noindex Templates

Templates with `meta.noindex: true`:
- Kept in database for now
- Excluded from frontend (like blocklist)
- Not deleted automatically

**Future enhancement**: Treat long-standing noindex as deletion candidates

**Allows**: Authors to toggle noindex without losing historical data

---

## Template Meta Field Support (Future)

**Goal**: Respect template author metadata

Lima templates may add a `meta` field for user-defined metadata. Our pipeline will respect these conventions:

### Meta Field Conventions

```yaml
# In template YAML:
meta:
  description: "Authoritative description from template author"
  keywords: ["user", "defined", "keywords"]
  noindex: true  # Exclude from catalog
```

### Priority Order

For final output:
1. `meta.description` (if present) - authoritative
2. LLM-generated description (if available)
3. Analysis-based keywords (fallback)

### Noindex Handling

Treat `meta.noindex: true` exactly like blocklist at ALL stages:
- Skip LLM generation (save tokens)
- Exclude from `catalog.jsonl` entirely
- Keep in `templates.jsonl` (deletion is future work)

---

## Additional Future Enhancements

### Template Quality Metrics

Calculate quality scores based on:
- Documentation completeness
- Configuration options
- Provision script complexity
- Community engagement (stars, issues)

**Use cases**:
- Highlight high-quality templates
- Guide new contributors
- Inform LLM prioritization

### Template Testing

Automated validation:
- Syntax validation (YAML parsing)
- Image availability checks
- Provision script linting
- Security scanning

**Benefits**:
- Early problem detection
- Quality assurance
- User confidence

### Template Recommendations

Suggest similar templates based on:
- Duplicate detection results
- Category and keyword overlap
- User behavior (if added later)

**UI**: "You might also like..." section

### Advanced Search

**Features**:
- Fuzzy keyword matching
- Boolean operators (AND/OR/NOT)
- Field-specific search (repo:, org:, keyword:)
- Saved searches

**Implementation**: Client-side with enhanced filters.js

### Template Analytics

Track and display:
- View counts
- Copy URL clicks
- Search terms leading to template
- Popular categories/keywords

**Privacy**: Aggregate only, no user tracking

### API Endpoint

Expose catalog data via JSON API:
- REST endpoints for templates, repos, orgs
- Pagination and filtering
- Rate limiting
- CORS support

**Use cases**:
- Third-party tools
- IDE extensions
- CLI clients

---

## Implementation Priority

**High Priority** (Core Value):
1. Stage 6: LLM Descriptions - Improves discoverability
2. Template meta field support - Respects authors

**Medium Priority** (Quality):
3. Stage 7: Template cleanup - Keeps catalog fresh
4. Template quality metrics - Highlights best templates

**Low Priority** (Nice to Have):
5. Advanced search - Power users
6. Template testing - Quality assurance
7. Analytics - Insights
8. API endpoint - Integrations

---

## Related Documentation

- **[Overview](overview.md)** - Current system state
- **[Data Pipeline](data-pipeline.md)** - Completed stages 1-5
- **[Backend Design](backend-design.md)** - Implementation details
- **[LLM Prompts](../reference/llm-prompts.md)** - Prompt generation system
- **[Research](../research/)** - Decision rationale

---

**Questions or suggestions?** Open an issue on GitHub!
