# Instructions for Claude

This file contains persistent instructions for Claude when working on this project.

---

## ⚠️ CRITICAL: Before Creating Pull Requests

**ALWAYS follow these steps IN ORDER before suggesting a PR:**

### Step 1: Update Documentation (DO THIS FIRST!)

**READ and UPDATE architecture docs if you made significant changes:**
- New features → Update [docs/architecture/overview.md](../docs/architecture/overview.md) or [future-work.md](../docs/architecture/future-work.md)
- UI changes → Update [docs/guides/ui-ux-guidelines.md](../docs/guides/ui-ux-guidelines.md)
- Backend changes → Update [docs/architecture/backend-design.md](../docs/architecture/backend-design.md)
- Data pipeline changes → Update [docs/architecture/data-pipeline.md](../docs/architecture/data-pipeline.md)
- Bug fixes → May not need doc updates (use judgment)

**⚠️ CRITICAL: Keep docs focused on current state**
- Active documentation (everything except `docs/history/`) should describe **what is**, not **what was**
- When implementing planned features from `future-work.md`:
  1. Update current docs (overview.md, backend-design.md, etc.) with the new implementation
  2. Remove completed items from future-work.md
  3. Add implementation details to `docs/history/implementation-details.md` if noteworthy
- Avoid phrases like "after refactoring", "recently added", "Phase X" outside `docs/history/`
- Historical context belongs in `docs/history/` - keep active docs evergreen

**Then commit documentation updates:**
```bash
git add docs/
git commit -m "Update documentation for [feature name]"
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

### Step 3: Run tests and linting (REQUIRED!)

**⚠️ CRITICAL: ALL tests and linting must pass before creating a PR:**

```bash
make test    # Runs go vet + all tests (Go + JavaScript)
make lint    # Runs golangci-lint (if installed)
```

**The test command will run:**
- `go vet` to catch common errors
- Go backend tests (83 tests)
- JavaScript frontend tests (76 tests)

**All tests must pass (exit code 0). If any tests fail, fix them before proceeding.**

**⚠️ NEVER IGNORE FAILING TESTS:**
- It is NEVER acceptable to ignore failing tests, even if they appear unrelated to your changes
- If tests fail due to missing dependencies (e.g., GITHUB_TOKEN), make them skip gracefully with a warning, not fail
- If tests are genuinely failing, investigate and fix them
- Tests that require external resources should use `t.Skip()` when resources are unavailable
- Always run `make test` before committing to ensure all tests pass

### Step 4: Build/test additional components if applicable

- Go changes: `make build` (builds lima-catalog binary)
- JavaScript changes: Tests are automated, but manual testing on GitHub Pages may be helpful

**REMINDERS:**
- **If you skip Step 1 (docs), the user will notice and ask why you forgot!**
- **If you skip Step 3 (tests), the PR will be rejected!**

---

## Quick Reference for New Sessions

**Starting a task?** Check these first:
- 🏗️ [Architecture Overview](../ARCHITECTURE.md) - System design at a glance
- 📋 [Source Index](../docs/architecture/source-index.md) - Find any file quickly
- ✅ [Code Standards](../docs/reference/code-standards.md) - Quality checklist
- 🎨 [UI Guidelines](../docs/guides/ui-ux-guidelines.md) - Frontend patterns

**Common Tasks:**
- Backend work → [Backend Design](../docs/architecture/backend-design.md)
- Frontend work → [Frontend Design](../docs/architecture/frontend-design.md)
- New feature → [Future Work](../docs/architecture/future-work.md)
- Understanding flow → [Data Pipeline](../docs/architecture/data-pipeline.md)

**Historical Reference:**
- Research findings → [docs/research/](../docs/research/)
- Implementation details → [docs/history/](../docs/history/)

---

## Using Makefile Targets

**⚠️ ALWAYS use `make` targets instead of running commands directly:**

This project uses a Makefile with helpful targets for common operations. **ALWAYS prefer these over raw commands:**

```bash
# ✅ CORRECT - Use make targets
make test          # Runs go vet + all tests (Go + JavaScript)
make test-go       # Run only Go tests
make test-js       # Run only JavaScript tests
make build         # Build the lima-catalog binary
make vet           # Run go vet
make lint          # Run golangci-lint
make clean         # Clean build artifacts

# ❌ INCORRECT - Don't run raw commands
go test ./...      # Missing vet step
go vet ./...       # Should use make vet
go build ./...     # Should use make build
```

**Why use make targets?**
- Ensures consistent workflow across sessions
- Runs checks in the correct order (e.g., vet before test)
- Provides helpful output formatting
- Documents the development workflow

**When you can use raw commands:**
- One-off git operations (git status, git log, etc.)
- Specialized commands not in Makefile
- Debugging specific issues

---

## Key Reminders

- **Analysis is incremental:** Templates only re-analyzed if SHA changes
- **Browser caching:** GitHub Pages changes may need hard refresh (Cmd+Shift+R / Ctrl+Shift+R)
- **Branch naming:** Must start with `claude/` and end with session ID for push permissions
- **No PR creation:** Cannot run `gh pr create` directly - provide command for user to run
- **⚠️ CRITICAL: NO CODE BLOCKS IN PR DESCRIPTIONS!**
  - NEVER use ``` code blocks in `gh pr create --body` text
  - Code blocks will break the heredoc (<<'EOF') and make the command uncopyable
  - Use indentation or plain text for code examples instead
  - If you need to show code, use 4-space indentation without backticks

---

## Common Workflows

### Making UI Changes
1. Edit `web/` files (HTML/CSS/JS)
2. **ALWAYS include accessibility features:**
   - Add `aria-label` attributes to interactive elements (buttons, inputs, links)
   - Add `role` attributes for semantic structure (main, complementary, dialog, etc.)
   - Add `title` attributes for additional context on hover
   - Add `aria-live` regions for dynamic content updates
   - Ensure keyboard navigation works properly
3. Test changes will be visible after GitHub Pages deploys
4. Remind user about browser cache (hard refresh)

### Making Backend Changes
1. Edit `pkg/` or `cmd/` files
2. Verify with `make test` (runs vet + tests)
3. Build with `make build` to verify binary compilation
4. Changes take effect on next workflow run
5. New templates are discovered daily; existing templates only re-analyzed if SHA changes

### Updating Keywords/Analysis Logic
- Changes to `parser.go` or `analyzer.go` only affect NEW templates or templates with updated files
- Existing analyzed templates keep their current keywords/categories until the template file changes
- To force re-analysis of all templates, would need to clear AnalyzedAt timestamps (generally not needed)

---

## Code Quality Standards

**⚠️ CRITICAL: All new code must follow quality standards and include tests!**

### Quick Checklist

**For Go Backend:**
- ✅ Use dependency injection (accept interfaces, not concrete types)
- ✅ Add context parameters to long-running functions
- ✅ Return errors (not bools) with proper wrapping
- ✅ Write tests in *_test.go files (aim for >60% coverage)
- ✅ Follow existing patterns (Analyzer, Discoverer, Storage)

**For JavaScript Frontend:**
- ✅ Write tests in `web/js/[module-name].test.js`
- ✅ Test edge cases and error conditions
- ✅ Add accessibility features (aria-label, role, keyboard navigation)
- ✅ Run `make test` before committing

**Common Anti-Patterns to Avoid:**
- ❌ God objects (keep code focused on one responsibility)
- ❌ Ignoring errors (always check and propagate)
- ❌ Global state (use dependency injection)
- ❌ Hard-coded dependencies (use interfaces for testability)

**For complete standards and examples, see:**
→ **[Code Standards Reference](../docs/reference/code-standards.md)** - Detailed Go and JavaScript standards
→ **[Backend Design](../docs/architecture/backend-design.md)** - Architecture patterns and best practices
→ **[UI/UX Guidelines](../docs/guides/ui-ux-guidelines.md)** - Complete frontend design system

---

## Creating Pull Requests

Use the gh command via the Bash tool for ALL GitHub-related tasks including working with issues, pull requests, checks, and releases. If given a Github URL use the gh command to get the information needed.

**IMPORTANT: When the user asks you to create a pull request, follow these steps carefully:**

1. You can call multiple tools in a single response. When multiple independent pieces of information are requested and all commands are likely to succeed, run multiple tool calls in parallel using the Bash tool, in order to understand the current state of the branch since it diverged from the main branch:
   - Run a git status command to see all untracked files
   - Run a git diff command to see both staged and unstaged changes that will be committed
   - Check if the current branch tracks a remote branch and is up to date with the remote, so you know if you need to push to the remote
   - Run a git log command and `git diff [base-branch]...HEAD` to understand the full commit history for the current branch (from the time it diverged from the base branch)
2. Analyze all changes that will be included in the pull request, making sure to look at all relevant commits (NOT just the latest commit, but ALL commits that will be included in the pull request!!!)
3. You can call multiple tools in a single response. When multiple independent pieces of information are requested and all commands are likely to succeed, run multiple tool calls in parallel:
   - Create new branch if needed
   - Push to remote with -u flag if needed
   - Create PR using gh pr create with the format below. Use a HEREDOC to pass the body to ensure correct formatting.

Example:
```bash
gh pr create --title "the pr title" --body "$(cat <<'EOF'
## Summary
<1-3 bullet points>

## Test plan
[Bulleted markdown checklist of TODOs for testing the pull request...]
EOF
)"
```

Important:
- DO NOT use the TodoWrite or Task tools
- Return the PR URL when you're done, so the user can see it

---

## Project Documentation

Quick links to essential documentation:

**Core Docs:**
- [ARCHITECTURE.md](../ARCHITECTURE.md) - High-level system design
- [DEVELOPMENT.md](../DEVELOPMENT.md) - Development workflow

**Detailed Docs:**
- [docs/architecture/](../docs/architecture/) - Detailed architecture docs
- [docs/guides/](../docs/guides/) - How-to guides
- [docs/reference/](../docs/reference/) - Reference documentation

**Historical:**
- [docs/research/](../docs/research/) - Research findings and decision rationale
- [docs/history/](../docs/history/) - Implementation archive

---

## Important Notes

- This project uses `web/` for GitHub Pages (not `docs/`)
- Project documentation is in `docs/`
- All development is done by AI agents
- PR workflow must be followed exactly (see top of this file!)
