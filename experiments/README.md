# Experiment Scripts

These scripts were used during the initial research phase to understand GitHub's code search behavior and estimate the scale of the Lima template ecosystem.

## Files

### `experiment_search.py`
**Purpose**: Initial scale estimation for Lima templates on GitHub

Searches GitHub for YAML files containing `minimumLimaVersion` to understand:
- How many community templates exist
- Where they're located
- Repository distribution

**Key Finding**: Found 57 community templates from 35 independent repositories (excluding lima-vm/lima).

### `experiment_fork_search.py`
**Purpose**: Investigate GitHub's fork search behavior

Critical discovery about GitHub Code Search:
- Forks are NOT indexed by default
- Only included if fork has more stars than parent
- lima-vm/lima has 18,903 stars → all 750 forks invisible to search

**Impact**: Shaped the three-tier catalog strategy (community + official + optional fork checking).

### `experiment_lima_structure.py`
**Purpose**: Understand lima-vm/lima repository structure

Investigated why lima-vm/lima templates weren't showing up in initial searches:
- Templates are in `templates/` directory (not `examples/`)
- 52 official template files
- All contain `minimumLimaVersion` field

**Result**: Led to separate enumeration strategy for official templates.

### `experiment_check_template.py`
**Purpose**: Verify actual Lima template structure and format

Downloads and inspects real template files to understand:
- YAML structure
- Required fields
- Provisioning script patterns
- Image specifications

**Use**: Informed the template analysis and parsing implementation.

## Historical Context

These experiments were conducted in November 2025 during the initial planning phase. Their findings are documented in detail in [../FINDINGS.md](../FINDINGS.md).

The research led to the successful catalog architecture that now tracks 700+ templates with daily automated updates.
