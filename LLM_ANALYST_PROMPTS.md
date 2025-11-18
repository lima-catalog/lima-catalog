# LLM Analyst Prompt System

## Overview

The LLM analyst prompt system is designed to generate comprehensive, context-rich prompts for analyzing Lima templates. This enables testing different LLM models to find the best approach for automatically generating template descriptions and keywords.

## Architecture

### Package: `pkg/prompt`

Core package for gathering context and building LLM prompts.

**Key Components:**

1. **`types.go`** - Data structures
   - `TemplateContext`: Contains all context needed for analysis
   - `TemplateReference`: References to template files in the repo
   - `PromptConfig`: Configuration for prompt generation

2. **`builder.go`** - Main logic
   - `Builder`: Main builder with GitHub API integration
   - `GatherContext()`: Collects all necessary context
   - `FormatPrompt()`: Formats context into structured prompt
   - Helper functions for fetching data and parsing YAML

3. **`builder_test.go`** - Unit tests
   - Tests for comment extraction
   - Tests for prompt formatting
   - Tests for configuration

### CLI Tool: `cmd/prompt-generator`

Standalone command-line tool for generating prompts.

**Features:**
- Takes template path (owner/repo/path format)
- Fetches all context via GitHub API
- Performs shallow clone to find template references
- Outputs formatted prompt ready for LLM testing
- Configurable context lines, size limits, and section inclusion

**Usage:**
```bash
export GITHUB_TOKEN=your_token
prompt-generator lima-vm/lima/templates/ubuntu.yaml
```

See `cmd/prompt-generator/README.md` for complete documentation.

## Prompt Context

The generated prompts include rich context from multiple sources:

### 1. Template YAML Content

**What's included:**
- Full YAML template
- Provisioning scripts
- Probe scripts
- Message field
- Environment variables
- Parameters
- Mount points
- Port forwards
- Containerd configuration

**Why:** The template content is the primary source of truth about what the template does.

### 2. Template Comments

**What's included:**
- All YAML comments (excluding section dividers)
- Extracted and listed separately for emphasis

**Why:** Comments often contain author intent, usage notes, and important caveats.

### 3. Repository Metadata

**What's included:**
- Repository name and description
- Topics/keywords
- Primary language
- Star count
- Homepage

**Why:** Provides context about the project, though may not directly relate to template purpose.

**Important caveat:** The prompt explicitly warns the LLM that repository purpose may differ from template purpose. Examples:
- Repo might be CI scaffolding, but template provisions a specific environment
- Repo might be documentation, but template makes a project available in VM
- Repo might be a template collection

### 4. Organization Information

**What's included:**
- Owner login and type (User/Organization)
- Display name
- Description/bio
- Location

**Why:** Helps understand the context and authority of the template source.

### 5. README Content

**What's included:**
- Repository README (if found)
- Truncated to max length (default: 5000 chars ≈ 1000 tokens)

**Why:** READMEs often provide valuable context about project purpose, usage, and features.

### 6. Template File References

**What's included:**
- Shallow clone of repository
- Grep search for template filename
- 15 lines of context before/after each reference (configurable)
- Up to 10 reference files (configurable)

**Why:** Shows how the template is used in practice:
- Documentation examples
- Test configurations
- CI/CD workflows
- Usage instructions

**Implementation:** Uses `git clone --depth 1` for efficiency, then `grep -r -n -B15 -A15` to find references.

## Prompt Format

The generated prompt has a clear structure:

```
# Lima Template Analysis Request

[Instructions for what to provide]

## Context

### Template File
- Repository
- Path

### Repository Information
- Description, topics, stars, language
- CAVEAT about repo vs template purpose

### Organization/Owner Information
- Type, name, description, location

### README Content
```
[README text]
```

### Template YAML Content
```yaml
[Full template]
```

### Template Comments
- Comment 1
- Comment 2
...

### References to Template in Repository
#### path/to/file.md (line 42)
```
[Before context]
>>> [Match line]
[After context]
```

## Analysis Instructions

[Detailed instructions for the LLM]
[Expected JSON output format]
```

## Configuration

### PromptConfig Options

```go
type PromptConfig struct {
    ContextLines      int  // Lines before/after references (default: 15)
    IncludeComments   bool // Include YAML comments (default: true)
    IncludeReferences bool // Include file references (default: true)
    IncludeReadme     bool // Include README (default: true)
    MaxReadmeLength   int  // Max README chars (default: 5000)
    MaxReferenceFiles int  // Max reference files (default: 10)
}
```

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-context` | 15 | Context lines around references |
| `-no-comments` | false | Exclude YAML comments |
| `-no-references` | false | Exclude template references |
| `-no-readme` | false | Exclude README |
| `-max-readme` | 5000 | Max README length |
| `-max-refs` | 10 | Max reference files |
| `-output` | stdout | Output file path |

## Expected Output

The prompt asks the LLM to return JSON:

```json
{
  "short_description": "Brief one-liner (max 100 chars)",
  "long_description": "Detailed explanation (max 500 chars)",
  "keywords": ["keyword1", "keyword2", "..."]
}
```

**Short description:** Suitable for search result snippets and list views

**Long description:** Provides comprehensive overview including:
- What the template provides
- Technologies/tools included
- Use cases
- Notable features

**Keywords:** 5-10 relevant tags for:
- Operating systems
- Technologies and frameworks
- Use cases (development, testing, security, etc.)
- Key features (kubernetes, docker, etc.)

## Testing Workflow

1. **Generate prompts** for sample templates:
   ```bash
   prompt-generator lima-vm/lima/templates/ubuntu.yaml -output=ubuntu.txt
   prompt-generator lima-vm/lima/templates/k3s.yaml -output=k3s.txt
   prompt-generator lima-vm/lima/templates/docker.yaml -output=docker.txt
   ```

2. **Test with different models:**
   ```bash
   cat ubuntu.txt | llm -m claude-3-haiku-20240307
   cat ubuntu.txt | llm -m claude-3-sonnet-20240229
   cat ubuntu.txt | llm -m gpt-4
   cat ubuntu.txt | llm -m gpt-3.5-turbo
   ```

3. **Compare results:**
   - Accuracy of descriptions
   - Relevance of keywords
   - Consistency across templates
   - Cost per analysis
   - Speed

4. **Select optimal model** based on:
   - Quality of output
   - Cost efficiency
   - Processing speed
   - Rate limits

5. **Integrate into backend** using same `pkg/prompt.Builder`

## Backend Integration (Future)

Once optimal model is selected, integrate into `pkg/discovery/analyzer.go`:

```go
import "github.com/lima-catalog/lima-catalog/pkg/prompt"

// In analyzer
promptBuilder := prompt.NewBuilder(ctx, githubToken, config)
ctx, err := promptBuilder.GatherContext(owner, repo, path)
if err != nil {
    // fallback to analysis-based keywords
}

// Call LLM API with formatted prompt
promptText := promptBuilder.FormatPrompt(ctx)
response := callLLM(promptText) // Implement LLM API call

// Parse JSON response and update template
template.ShortDescription = response.ShortDescription
template.Description = response.LongDescription
template.Keywords = response.Keywords
```

## Token Usage Estimation

The tool provides approximate token estimates:

- Template content: 200-1000 tokens (varies widely)
- README: ~1000 tokens (truncated to 5000 chars)
- References: ~50-100 tokens per reference × 10 = 500-1000 tokens
- Repository metadata: ~100 tokens
- Organization info: ~50 tokens
- Instructions: ~300 tokens

**Total estimate:** 2000-3500 tokens per prompt

**Cost examples** (Claude Haiku):
- Input: $0.25 per million tokens
- Output: $1.25 per million tokens
- Per analysis: ~$0.001 input + ~$0.0002 output ≈ **$0.0012**

For 700 templates: ~$0.84 for initial generation (very affordable!)

## Design Decisions

### Why shallow clone instead of GitHub API?

GitHub Code Search API:
- Rate limited (10 queries per minute)
- Complex query syntax
- May not return all results
- Harder to get context lines

Shallow clone + grep:
- Fast (only fetches latest commit)
- Reliable (finds all references)
- Easy to get context (`-B` and `-A` flags)
- No API rate limits
- Clean up is easy (temp directory)

### Why include repository metadata with caveats?

Repository context can be valuable:
- Sometimes template makes repo project available
- Repo topics provide useful keywords
- Repo description may explain template purpose

But can also be misleading:
- Repo might be unrelated (CI scaffolding)
- Repo might be template collection
- Repo might be documentation

**Solution:** Include it but warn the LLM to use judgment and prioritize template content.

### Why extract comments separately?

YAML comments are often buried in the template:
- Easy to overlook in full YAML
- Contain important context about author intent
- May include usage instructions or caveats

Listing them separately gives them prominence in the prompt.

### Why limit README and references?

Token costs and prompt length:
- Very long prompts cost more
- May exceed model context windows
- Can include noise

**Defaults chosen:**
- README: 5000 chars (≈1000 tokens) - enough for most READMEs
- References: 10 files - covers common usage patterns
- Context lines: 15 - enough to understand usage

## Future Enhancements

1. **Caching:** Cache repo clones for batch processing
2. **Parallel processing:** Generate multiple prompts concurrently
3. **Private repos:** Add SSH key support for private repositories
4. **Direct LLM integration:** Optional API calls from CLI tool
5. **Prompt templates:** Multiple prompt styles for different models
6. **Response validation:** Check JSON format and content constraints
7. **Batch mode:** Process multiple templates in one run
8. **Diff mode:** Compare prompts before/after changes

## Files

- `pkg/prompt/types.go` - Data structures and configuration
- `pkg/prompt/builder.go` - Core prompt building logic
- `pkg/prompt/builder_test.go` - Unit tests
- `cmd/prompt-generator/main.go` - CLI tool
- `cmd/prompt-generator/README.md` - CLI documentation
- `cmd/prompt-generator/example.sh` - Usage examples
- `LLM_ANALYST_PROMPTS.md` - This document

## Related Documentation

- `PLAN.md` - Stage 6: LLM Descriptions section
- `pkg/discovery/analyzer.go` - Current analysis logic (will integrate with this)
- `pkg/discovery/parser.go` - Template parsing (used by prompt builder)
