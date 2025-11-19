# Debug Template Tool

This tool helps debug the notability scoring system by showing what lines remain in a template after removing known lines from official templates.

## Getting official.json

The official knowledge database is stored in the `data` branch. To use it with this tool on the `main` branch:

```bash
# Check out official.json from the data branch
git show data:data/official.json > data/official.json

# Or if you want to use a temporary copy
git show data:data/official.json > /tmp/official.json
```

## Usage

```bash
go run ./cmd/debug-template -url <template-url> [-official <path-to-official.json>]
```

Or build it first:

```bash
go build -o debug-template ./cmd/debug-template
./debug-template -url <template-url>
```

## Examples

```bash
# Debug a template from lima-vm/lima
./debug-template -url "https://github.com/lima-vm/lima/blob/master/templates/ubuntu.yaml"

# Debug a community template
./debug-template -url "https://github.com/user/repo/blob/main/lima.yaml"

# Use a different official.json file
./debug-template -url "https://github.com/user/repo/blob/main/lima.yaml" -official /path/to/official.json
```

## Output

The tool shows:
1. **Statistics**: How many known lines are loaded from official.json
2. **Unique comment lines**: Comments that don't exist in official templates (excludes code-like comments)
3. **Unique provision lines**: Provision script lines not in official templates
4. **Unique probe lines**: Probe script lines not in official templates
5. **Unique message lines**: Message lines not in official templates
6. **Summary**: Total counts of unique lines

Lines marked with `[CODE]` are filtered out as commented-out code (e.g., `# apt-get install foo`).

## Purpose

This tool helps verify that:
- Official knowledge extraction is working correctly
- Template normalization is consistent
- Filtering logic correctly identifies unique vs known lines
- Code comment detection works as expected

Use this when debugging why a template has a high or low notability score.
