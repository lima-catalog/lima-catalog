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

**⚠️ IMPORTANT: Keep PLAN.md concise!**

When updating PLAN.md, check if implementation details should go to IMPLEMENTATION_NOTES.md instead:
- **PLAN.md**: Current architecture, remaining work, high-level design decisions
- **IMPLEMENTATION_NOTES.md**: Detailed "how we did it" notes, completed stage details, migration notes

If you're adding detailed implementation notes for a completed feature, put them in IMPLEMENTATION_NOTES.md and keep only a summary in PLAN.md.

**Then commit PLAN.md updates:**
```bash
git add PLAN.md IMPLEMENTATION_NOTES.md  # If both changed
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

### Step 3: Run tests (REQUIRED!)

**⚠️ CRITICAL: ALL tests must pass before creating a PR:**

```bash
make test
```

**This will run all tests (Go + JavaScript):**
- Go backend tests (blocklist, metadata, combiner, etc.)
- JavaScript frontend tests (URL helpers, data parsing, filters, etc.)

**All tests must pass (exit code 0). If any tests fail, fix them before proceeding.**

**⚠️ NEVER IGNORE FAILING TESTS:**
- It is NEVER acceptable to ignore failing tests, even if they appear unrelated to your changes
- If tests fail due to missing dependencies (e.g., GITHUB_TOKEN), make them skip gracefully with a warning, not fail
- If tests are genuinely failing, investigate and fix them
- Tests that require external resources should use `t.Skip()` when resources are unavailable
- Always run `make test` before committing to ensure all tests pass

### Step 4: Build/test additional components if applicable

- Go changes: `go build -o lima-catalog ./cmd/lima-catalog`
- JavaScript changes: Tests are automated, but manual testing on GitHub Pages may be helpful

**REMINDERS:**
- **If you skip Step 1 (PLAN.md), the user will notice and ask why you forgot!**
- **If you skip Step 3 (tests), the PR will be rejected!**

## Project Context

- **[PLAN.md](PLAN.md)** - Current architecture, remaining work, design decisions
- **[IMPLEMENTATION_NOTES.md](IMPLEMENTATION_NOTES.md)** - Detailed implementation notes for completed features
- **[INTERFACE_GUIDELINES.md](INTERFACE_GUIDELINES.md)** - Complete UI/UX design system
- **[SOURCE_INDEX.md](SOURCE_INDEX.md)** - Quick reference for all source files and their purposes
- **[FINDINGS.md](FINDINGS.md)** - Research findings on GitHub search behavior
- **[LLM_ANALYST_PROMPTS.md](LLM_ANALYST_PROMPTS.md)** - Documentation for LLM prompt generation system

## Key Reminders

- **Analysis is incremental:** Templates are only re-analyzed if their SHA changes (see `analyzer.go:170`)
- **Browser caching:** GitHub Pages changes may need hard refresh (Cmd+Shift+R / Ctrl+Shift+R)
- **Branch naming:** Must start with `claude/` and end with session ID for push permissions
- **No PR creation:** Cannot run `gh pr create` directly - provide command for user to run
- **⚠️ CRITICAL: NO CODE BLOCKS IN PR DESCRIPTIONS!**
  - NEVER use ``` code blocks in `gh pr create --body` text
  - Code blocks will break the heredoc (<<'EOF') and make the command uncopyable
  - Use indentation or plain text for code examples instead
  - If you need to show code, use 4-space indentation without backticks

## Common Workflows

### Making UI Changes
1. Edit `docs/` files (HTML/CSS/JS)
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
2. Build with `go build` to verify
3. Changes take effect on next workflow run
4. New templates are discovered daily; existing templates only re-analyzed if SHA changes

### Updating Keywords/Analysis Logic
- Changes to `parser.go` or `analyzer.go` only affect NEW templates or templates with updated files
- Existing analyzed templates keep their current keywords/categories until the template file changes
- To force re-analysis of all templates, would need to clear AnalyzedAt timestamps (generally not needed)

## Writing Tests

**⚠️ CRITICAL: All new code MUST include tests!**

### When to Write Tests

**Always write tests when:**
- Adding new JavaScript functions or modules
- Modifying existing business logic
- Adding new data processing or filtering logic
- Creating new URL helpers or utility functions

**Tests may not be needed for:**
- Pure CSS changes
- HTML structure changes (unless affecting functionality)
- Documentation updates
- Configuration changes

### Where to Put Tests

- JavaScript tests: `docs/js/[module-name].test.js`
- Test framework: `docs/js/test-framework.js` (already exists)
- Main test runner: `test.js` (Node.js runner)

### How to Write Tests

1. Import the test framework and module to test:
   ```javascript
   import { runner, assert } from './test-framework.js';
   import { myFunction } from './myModule.js';
   ```

2. Write tests using `runner.test()`:
   ```javascript
   runner.test('myFunction: does something', () => {
       const result = myFunction(input);
       assert.equal(result, expected);
   });
   ```

3. Run tests locally before committing:
   ```bash
   make test
   ```

4. All tests must pass before creating a PR

### Test Coverage Guidelines

- **Aim for high coverage of pure functions** (functions without side effects)
- Test edge cases: empty inputs, null values, boundary conditions
- Test error cases: invalid inputs should throw appropriate errors
- DOM manipulation functions may need minimal mocking (see `test.js` for examples)


## Backend Code Quality Standards

**⚠️ CRITICAL: The backend has been extensively refactored for quality. Maintain these standards!**

After 6 phases of refactoring (Phases 1-4, 6.1 complete), the backend is in excellent shape with 60%+ test coverage, idiomatic APIs, and solid error handling. **Follow these patterns to keep it that way.**

### Go Code Quality Checklist

When writing or modifying Go backend code:

**1. Use Dependency Injection for Testability**
- ✅ DO: Accept interfaces (HTTPClient, FileSystem, Clock) not concrete types
- ✅ DO: Use functional options pattern for complex constructors
- ❌ DON'T: Call os.Create(), http.Get(), or time.Now() directly in business logic

Example (Good):
    func NewAnalyzer(opts ...AnalyzerOption) *Analyzer {
        a := &Analyzer{
            HTTPClient: interfaces.NewDefaultHTTPClient(),
            Clock:      interfaces.NewDefaultClock(),
        }
        for _, opt := range opts {
            opt(a)
        }
        return a
    }

**2. Add Context Parameters for Cancellation**
- ✅ DO: Add ctx context.Context as first parameter to long-running functions
- ✅ DO: Check ctx.Done() before starting work and in loops
- ✅ DO: Use select with ctx.Done() during sleeps/waits
- ❌ DON'T: Use time.Sleep() in loops without checking context

**3. Return Errors, Not Bools**
- ✅ DO: Return error to communicate what went wrong
- ✅ DO: Use sentinel errors (var ErrNotFound = errors.New(...)) for expected errors
- ✅ DO: Wrap errors with context: fmt.Errorf("failed to X: %w", err)
- ❌ DON'T: Return bool when callers need to know why something failed

**4. Write Tests for New Code**
- ✅ DO: Write tests in *_test.go files alongside implementation
- ✅ DO: Test edge cases (nil, empty, boundary values)
- ✅ DO: Use table-driven tests for multiple scenarios
- ✅ DO: Mock external dependencies using interfaces
- ❌ DON'T: Skip tests because "it's too hard to test"

**5. Follow Existing Patterns**

The codebase uses these patterns consistently:
- Interfaces: HTTPClient, FileSystem, Clock for all external dependencies
- Functional Options: NewX(opts ...Option) for complex constructors
- Context: First parameter in long-running functions
- Sentinel Errors: Named error variables for expected error conditions
- Table-Driven Tests: One test function with []struct{} for multiple cases

When adding new features:
1. Look at similar existing code (e.g., Analyzer, Discoverer, Storage)
2. Follow the same patterns
3. Add tests using the same style
4. Use the same interfaces for dependencies

### Common Anti-Patterns to Avoid

❌ DON'T create God objects (objects that do everything)
   - Keep structs focused on one responsibility
   - Extract helpers when functions get too long (>50 lines)

❌ DON'T ignore errors
   - Always check error returns
   - Log or propagate errors, never silently discard

❌ DON'T use global state
   - Pass dependencies explicitly
   - Use struct fields, not package-level variables

❌ DON'T hard-code external dependencies
   - Use interfaces for anything that touches I/O, network, or time
   - Makes code testable without actual I/O

### When in Doubt

1. Check BACKEND_CODE_REVIEW.md for past issues and solutions
2. Look at recently refactored code (Analyzer, Discoverer, Storage)
3. Run make test frequently during development
4. If test coverage drops below 60%, add more tests
