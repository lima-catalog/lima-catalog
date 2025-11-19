# Lima Template Prompt Generator

A standalone CLI tool for generating LLM prompts to analyze Lima VM templates.

## Purpose

This tool helps you test and select the best LLM model for generating template descriptions and keywords. It gathers comprehensive context about a template and formats it into a prompt ready for LLM analysis.

## Installation

```bash
cd cmd/prompt-generator
go build -o prompt-generator
```

Or use directly with `go run`:

```bash
go run ./cmd/prompt-generator/main.go [args]
```

## Usage

### Basic Usage

```bash
# Set your GitHub token
export GITHUB_TOKEN=your_github_token_here

# Generate prompt for a template
./prompt-generator lima-vm/lima/templates/ubuntu.yaml

# Or using separate flags
./prompt-generator -owner=lima-vm -repo=lima -path=templates/ubuntu.yaml
```

### Save to File

```bash
# Save prompt to file for later use
./prompt-generator lima-vm/lima/templates/ubuntu.yaml -output=ubuntu-prompt.txt

# Then use with your favorite LLM
cat ubuntu-prompt.txt | llm  # Using Simon Willison's llm tool
# Or copy-paste into Claude, ChatGPT, etc.
```

### Customize Context

```bash
# More context lines around references (default: 15)
./prompt-generator lima-vm/lima/templates/k3s.yaml -context=20

# Minimal prompt (no references or README)
./prompt-generator lima-vm/lima/templates/docker.yaml -no-references -no-readme

# Just template content and comments
./prompt-generator lima-vm/lima/templates/alpine.yaml -no-references -no-readme

# Limit README length (default: 5000 chars)
./prompt-generator lima-vm/lima/templates/ubuntu.yaml -max-readme=3000

# Limit number of reference files (default: 10)
./prompt-generator lima-vm/lima/templates/kubernetes.yaml -max-refs=5
```

## What's Included in the Prompt

The generated prompt includes comprehensive context:

1. **Template YAML Content**
   - Full YAML with all configurations
   - Provisioning scripts and probes
   - Environment variables and parameters
   - Mount points and port forwards

2. **Template Comments**
   - Extracted YAML comments (if enabled)
   - Helps understand author's intent

3. **Repository Context**
   - Repository description
   - Topics/keywords
   - Primary language
   - Star count
   - **Important caveat**: Notes that repo purpose may differ from template purpose

4. **Organization Info**
   - Owner name and type
   - Organization description
   - Location

5. **README Content**
   - Repository README (if found)
   - Truncated to max length if needed
   - Provides broader project context

6. **Template References**
   - Files that reference this template
   - 15 lines of context before/after each reference (configurable)
   - Helps understand how template is used
   - Found via shallow clone + grep

## Output Format

The prompt is formatted as a structured request asking the LLM to provide:

1. **Short Description** (max 100 chars) - Brief one-liner for search results
2. **Long Description** (max 500 chars) - Detailed explanation
3. **Keywords** (5-10 keywords) - Technologies, use cases, features

The LLM is asked to return JSON:

```json
{
  "short_description": "Ubuntu-based development environment with Docker and Kubernetes",
  "long_description": "Full-featured Ubuntu VM with Docker, kubectl, k3s, and common development tools pre-installed. Ideal for container development and Kubernetes testing on macOS or Linux hosts.",
  "keywords": ["ubuntu", "docker", "kubernetes", "k3s", "development", "containers"]
}
```

## Environment Variables

- `GITHUB_TOKEN` - **Required**. Your GitHub personal access token
  - Needed to fetch repository metadata, README, and template content
  - Create at: https://github.com/settings/tokens
  - Needs `public_repo` scope for public repos
- `PROMPT_TEMPLATE` - Optional. Path to custom prompt template file
  - Can be overridden by `-template` flag
  - Allows testing different prompt structures

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-owner` | string | - | GitHub repository owner |
| `-repo` | string | - | GitHub repository name |
| `-path` | string | - | Path to template file in repo |
| `-template` | string | (embedded) | Path to custom prompt template file |
| `-context` | int | 15 | Context lines around references |
| `-no-comments` | bool | false | Exclude YAML comments |
| `-no-references` | bool | false | Exclude template file references |
| `-no-readme` | bool | false | Exclude README |
| `-max-readme` | int | 5000 | Max README length (0 = unlimited) |
| `-max-refs` | int | 10 | Max reference files (0 = unlimited) |
| `-output` | string | stdout | Output file path |
| `-help` | bool | false | Show help |

## Examples

### Compare Different Models

```bash
# Generate prompt
./prompt-generator lima-vm/lima/templates/ubuntu.yaml -output=prompt.txt

# Test with different models
cat prompt.txt | llm -m gpt-4
cat prompt.txt | llm -m claude-3-haiku-20240307
cat prompt.txt | llm -m claude-3-sonnet-20240229

# Compare outputs and select best model
```

### Batch Processing

```bash
# Generate prompts for multiple templates
for template in templates/*.yaml; do
  name=$(basename $template .yaml)
  ./prompt-generator lima-vm/lima/$template -output=prompts/$name.txt
done
```

### Custom Prompt Templates

The tool uses Go templates for flexible prompt customization:

```bash
# Create a custom template file (my-template.tmpl)
cat > my-template.tmpl <<'EOF'
# Analyze This Lima Template

## Template Content
{{.TemplateContent}}

{{- if .Repository}}
Repository: {{.Repository.Name}} (⭐ {{.Repository.Stars}})
{{- end}}

Provide:
- Short description (max 100 chars)
- Keywords (5-10)
EOF

# Use custom template via flag
./prompt-generator lima-vm/lima/templates/ubuntu.yaml -template=my-template.tmpl

# Or via environment variable
export PROMPT_TEMPLATE=my-template.tmpl
./prompt-generator lima-vm/lima/templates/ubuntu.yaml
```

**Template data available:**
- `.Template` - Template metadata (Repo, Path, ID)
- `.Repository` - Repo info (Name, Description, Topics, Stars, etc.)
- `.Organization` - Owner info (Login, Type, Name, etc.)
- `.TemplateContent` - Raw YAML content
- `.Comments` - Extracted YAML comments
- `.ReadmeContent` - README content
- `.References` - File references with context
- `.Config` - Config options

**Template functions:**
- `join` - Join arrays: `{{join .Repository.Topics ", "}}`
- `sub` - Subtract: `{{sub (len .References) .Config.MaxReferenceFiles}}`

See `pkg/prompt/default_template.tmpl` for the default template structure.

## Use in Backend

This tool shares the same `pkg/prompt` package that will be used by the backend analyzer. The package provides:

- `prompt.Builder` - Main builder with `BuildPrompt()` method
- `prompt.TemplateContext` - Full context structure
- `prompt.PromptConfig` - Configuration options

See `pkg/prompt/builder.go` for the API.

## Tips

1. **Start minimal**: Use `-no-references -no-readme` to see if template content alone is sufficient
2. **Increase context**: If references are valuable, increase `-context` to 20-30 lines
3. **Monitor token usage**: Check the output for estimated tokens (chars/4)
4. **Test different models**: Claude Haiku is fast/cheap, Sonnet is more accurate
5. **Save prompts**: Use `-output` to save for comparison and iteration

## Troubleshooting

### "GITHUB_TOKEN environment variable is required"
Set your token: `export GITHUB_TOKEN=ghp_your_token_here`

### "failed to clone repo"
- Check internet connection
- Verify repository exists and is public
- Ensure `git` is installed

### "failed to fetch repository"
- Check token is valid
- Verify repository path is correct
- Check API rate limits

### Prompt too large
- Use `-no-references` to exclude grep results
- Use `-no-readme` to exclude README
- Reduce `-max-readme` or `-max-refs` limits
- Reduce `-context` lines

## Future Enhancements

- Cache repository clones for faster repeated runs
- Support for private repositories
- Parallel processing for batch operations
- Direct LLM API integration (optional)
- Template diffing (compare before/after prompt changes)
