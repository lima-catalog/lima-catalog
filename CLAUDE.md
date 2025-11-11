# Instructions for Claude

This file contains persistent instructions for Claude when working on this project.

## Before Creating Pull Requests

**⚠️ CRITICAL: ALWAYS follow these steps IN ORDER before suggesting a PR:**

### Step 1: Update PLAN.md (DO THIS FIRST!)

**READ PLAN.md and UPDATE IT if you made significant changes:**
- New features → Add to appropriate section (e.g., "Template Preview Modal")
- UI changes → Update "Recent: GitHub Pages UI Redesign" section
- Backend changes → Add to relevant section
- Bug fixes → May not need PLAN.md update (use judgment)

**Then commit PLAN.md updates:**
```bash
git add PLAN.md
git commit -m "Update PLAN.md to document [feature name]"
```

### Step 2: Check and rebase on main

**Check if main has been updated:**
```bash
git fetch origin main
git log HEAD..origin/main --oneline
```

**If there are new commits, rebase your branch:**
```bash
git rebase origin/main
```

### Step 3: Build/test if applicable

- Go changes: `go build -o lima-catalog ./cmd/lima-catalog`
- JavaScript changes: Manual testing on GitHub Pages may be needed

**REMINDER: If you skip Step 1 (PLAN.md), the user will notice and ask why you forgot!**

## Project Context

See [PLAN.md](PLAN.md) for full project architecture, implementation details, and progress tracking.

## Key Reminders

- **Analysis is incremental:** Templates are only re-analyzed if their SHA changes (see `analyzer.go:170`)
- **Browser caching:** GitHub Pages changes may need hard refresh (Cmd+Shift+R / Ctrl+Shift+R)
- **Branch naming:** Must start with `claude/` and end with session ID for push permissions
- **No PR creation:** Cannot run `gh pr create` directly - provide command for user to run

## Common Workflows

### Making UI Changes
1. Edit `docs/` files (HTML/CSS/JS)
2. Test changes will be visible after GitHub Pages deploys
3. Remind user about browser cache (hard refresh)

### Making Backend Changes
1. Edit `pkg/` or `cmd/` files
2. Build with `go build` to verify
3. Changes take effect on next workflow run
4. New templates are discovered daily; existing templates only re-analyzed if SHA changes

### Updating Keywords/Analysis Logic
- Changes to `parser.go` or `analyzer.go` only affect NEW templates or templates with updated files
- Existing analyzed templates keep their current keywords/categories until the template file changes
- To force re-analysis of all templates, would need to clear AnalyzedAt timestamps (generally not needed)
