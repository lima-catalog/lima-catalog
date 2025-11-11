# Daily Updates and Template Categorization

## Summary

Implements daily catalog updates and comprehensive template analysis with automatic categorization. Templates now have human-readable names, descriptions, categories, and searchable keywords.

## What's New

### 1. Daily Updates
- Changed workflow from weekly to **daily** (runs at 00:00 UTC)
- Ensures catalog stays fresh with latest templates

### 2. Smart Template Naming
Handles generic filenames like `lima.yaml` intelligently:
- **Generic filename** (`lima.yaml`) → Uses repo name ("container-security")
- **Path context** (`lima/box.yml`) → Uses parent directory ("box")
- **Descriptive** (`aistor_ubuntu.yaml`) → Uses filename ("aistor-ubuntu")
- **Display names**: Auto-generates human-readable versions ("Container Security", "Aistor Ubuntu")

### 3. Template Analysis & Categorization
Automatic analysis of template content:

**Parsing:**
- Extracts OS images (Ubuntu, Alpine, Debian, Kali, etc.)
- Detects architectures (x86_64, aarch64)
- Analyzes provisioning scripts

**Technology Detection:**
- Container runtimes: Docker, Kubernetes, Podman
- Databases: PostgreSQL, MySQL, MongoDB, Redis
- Dev tools: Git, Node, Python, Go, Rust
- And more...

**Auto-Categorization:**
- `containers` - Docker/Podman environments
- `orchestration` - Kubernetes setups
- `development` - Development environments
- `database` - Database servers
- `security` - Security/pentest tools
- `testing` - CI/CD environments
- `general` - Other use cases

**Generated Data:**
- Short names and display names
- Categories and use cases
- Descriptions and summaries
- Searchable keywords
- OS and architecture info

### 4. Extended Data Schema
Added analysis fields to Template type:
```go
Name             string    // "container-security"
DisplayName      string    // "Container Security"
ShortDescription string    // "Ubuntu-based containers with Docker"
Description      string    // Full description
Category         string    // "containers"
UseCase          string    // "container-runtime"
Keywords         []string  // ["ubuntu", "docker", "git"]
Images           []string  // ["ubuntu"]
Arch             []string  // ["x86_64", "aarch64"]
AnalyzedAt       time.Time // Analysis timestamp
```

## Usage

### Run Analysis
```bash
export ANALYZE=true
export GITHUB_TOKEN=your_token
./lima-catalog
```

Analysis automatically:
- Derives meaningful names from paths
- Parses YAML templates
- Detects technologies from scripts
- Assigns categories and keywords
- Generates descriptions

### LLM Enhancement (Optional)
```bash
export ANALYZE=true
export LLM_API_KEY=your_api_key
./lima-catalog
```
*(LLM integration is a placeholder for future enhancement)*

## Testing

Tested on real templates with excellent results:
- ✅ **container-security/lima.yaml** → Detected Ubuntu + Docker → "containers" category
- ✅ **pentest-lab/box.yml** → Detected Kali ARM → "box" (from path)
- ✅ **aistor_ubuntu.yaml** → Detected Ubuntu + Git → "development" category

## Files Changed

**New Files:**
- `pkg/discovery/naming.go` - Smart name derivation logic
- `pkg/discovery/parser.go` - YAML parsing and technology detection
- `pkg/discovery/analyzer.go` - Analysis coordination

**Modified Files:**
- `.github/workflows/update-catalog.yml` - Daily schedule
- `pkg/types/types.go` - Extended Template struct
- `cmd/lima-catalog/main.go` - Added ANALYZE mode
- `README.md` - Analysis documentation
- `PLAN.md` - Phase 2 implementation + Phase 3 next steps

## Next Steps (After Merge)

**Phase 3: GitHub Pages Website**
- Build static site to browse catalog
- Display templates with categories, descriptions, keywords
- Add search/filter functionality
- Deploy to GitHub Pages for easy review
- Enables visual validation of analyzed templates

This will allow reviewing the auto-generated names and categories before implementing LLM enhancements.

## Implementation Details

**Smart Naming:**
- Avoids generic terms ("lima", "template", "config")
- Uses repo context for disambiguation
- Generates clean, readable identifiers

**Parsing:**
- Downloads templates via raw.githubusercontent.com
- Parses YAML structure safely
- Extracts key fields (images, scripts, mounts, ports)

**Technology Detection:**
- Pattern matching on provisioning scripts
- Detects install commands and tool names
- Builds keyword list automatically

**Categorization:**
- Priority-based category assignment
- Uses detected technologies as primary signal
- Falls back to repo topics/description
- Assigns default "general" category if unknown

## Benefits

1. **Searchable**: Keywords enable finding templates by technology
2. **Browsable**: Categories organize templates logically
3. **Discoverable**: Human-readable names aid navigation
4. **Informative**: Descriptions explain template purposes
5. **Validated**: Analysis results can be reviewed via website

## Backwards Compatible

All existing fields preserved. New fields use `omitempty` JSON tags, so they're optional and won't break existing data readers.
