# Code Standards

**Quick Links**: [Backend Design](../architecture/backend-design.md) | [.claude/instructions](../../.claude/instructions.md) | [Getting Started](../guides/getting-started.md)

---

## Overview

This document outlines code quality standards for the Lima Catalog project, with emphasis on the Go backend which has been extensively refactored for quality (60%+ test coverage, idiomatic APIs).

---

## Backend (Go) Standards

### Critical: Dependency Injection for Testability

✅ **DO**: Accept interfaces (HTTPClient, FileSystem, Clock) not concrete types
✅ **DO**: Use functional options pattern for complex constructors
❌ **DON'T**: Call os.Create(), http.Get(), or time.Now() directly in business logic

**Example (Good)**:
```go
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
```

### Add Context Parameters for Cancellation

✅ **DO**: Add ctx context.Context as first parameter to long-running functions
✅ **DO**: Check ctx.Done() before starting work and in loops
✅ **DO**: Use select with ctx.Done() during sleeps/waits
❌ **DON'T**: Use time.Sleep() in loops without checking context

### Return Errors, Not Bools

✅ **DO**: Return error to communicate what went wrong
✅ **DO**: Use sentinel errors (var ErrNotFound = errors.New(...)) for expected errors
✅ **DO**: Wrap errors with context: fmt.Errorf("failed to X: %w", err)
❌ **DON'T**: Return bool when callers need to know why something failed

### Write Tests for New Code

✅ **DO**: Write tests in *_test.go files alongside implementation
✅ **DO**: Test edge cases (nil, empty, boundary values)
✅ **DO**: Use table-driven tests for multiple scenarios
✅ **DO**: Mock external dependencies using interfaces
❌ **DON'T**: Skip tests because "it's too hard to test"

### Follow Existing Patterns

The codebase uses these patterns consistently:

- **Interfaces**: HTTPClient, FileSystem, Clock for all external dependencies
- **Functional Options**: NewX(opts ...Option) for complex constructors
- **Context**: First parameter in long-running functions
- **Sentinel Errors**: Named error variables for expected error conditions
- **Table-Driven Tests**: One test function with []struct{} for multiple cases

**When adding new features:**
1. Look at similar existing code (e.g., Analyzer, Discoverer, Storage)
2. Follow the same patterns
3. Add tests using the same style
4. Use the same interfaces for dependencies

---

## Common Anti-Patterns to Avoid

### ❌ DON'T: Create God Objects

Keep structs focused on one responsibility. Extract helpers when functions get too long (>50 lines).

### ❌ DON'T: Ignore Errors

Always check error returns. Log or propagate errors, never silently discard.

### ❌ DON'T: Use Global State

Pass dependencies explicitly. Use struct fields, not package-level variables.

### ❌ DON'T: Hard-Code External Dependencies

Use interfaces for anything that touches I/O, network, or time. Makes code testable without actual I/O.

---

## Frontend (JavaScript) Standards

### Always Write Tests

**When to write tests:**
- Adding new JavaScript functions or modules
- Modifying existing business logic
- Adding new data processing or filtering logic
- Creating new URL helpers or utility functions

**Tests may not be needed for:**
- Pure CSS changes
- HTML structure changes (unless affecting functionality)
- Documentation updates

### Test Location

- JavaScript tests: `web/js/[module-name].test.js`
- Test framework: `web/js/test-framework.js`
- Main test runner: `test.js` (Node.js)

### Test Example

```javascript
import { runner, assert } from './test-framework.js';
import { myFunction } from './myModule.js';

runner.test('myFunction: does something', () => {
    const result = myFunction(input);
    assert.equal(result, expected);
});
```

### Accessibility Requirements

**ALWAYS include accessibility features:**
- Add `aria-label` attributes to interactive elements (buttons, inputs, links)
- Add `role` attributes for semantic structure (main, complementary, dialog, etc.)
- Add `title` attributes for additional context on hover
- Add `aria-live` regions for dynamic content updates
- Ensure keyboard navigation works properly

→ See [UI/UX Guidelines](../guides/ui-ux-guidelines.md) for complete design system

---

## Testing Requirements

### All Tests Must Pass

**⚠️ CRITICAL**: ALL tests must pass before creating a PR

```bash
make test
```

**This runs:**
- Go backend tests (83 tests)
- JavaScript frontend tests (76 tests)

**⚠️ NEVER IGNORE FAILING TESTS**:
- It is NEVER acceptable to ignore failing tests
- If tests fail due to missing dependencies (e.g., GITHUB_TOKEN), make them skip gracefully with warnings
- If tests are genuinely failing, investigate and fix them
- Tests that require external resources should use `t.Skip()` when resources are unavailable

### Test Coverage Guidelines

- **Aim for high coverage of pure functions** (functions without side effects)
- Test edge cases: empty inputs, null values, boundary conditions
- Test error cases: invalid inputs should throw appropriate errors
- DOM manipulation functions may need minimal mocking

---

## Code Quality Metrics

**Current State (Backend)**:
- ✅ 0 critical issues
- ✅ 60%+ test coverage
- ✅ Idiomatic Go APIs
- ✅ Comprehensive documentation

**Maintain these standards:**
- Test coverage: >60%
- Function length: <50 lines (guideline)
- All errors checked
- No global state
- Interfaces for external dependencies

---

## When in Doubt

1. Check [Backend Refactoring History](../history/backend-refactoring/) for past issues and solutions
2. Look at recently refactored code (Analyzer, Discoverer, Storage)
3. Run `make test` frequently during development
4. If test coverage drops below 60%, add more tests

---

## Documentation Standards

### Before Creating PRs

**⚠️ CRITICAL: Update documentation FIRST**:

1. **New features** → Update [docs/architecture/overview.md](../architecture/overview.md) or [future-work.md](../architecture/future-work.md)
2. **UI changes** → Update [docs/guides/ui-ux-guidelines.md](../guides/ui-ux-guidelines.md)
3. **Backend changes** → Update [docs/architecture/backend-design.md](../architecture/backend-design.md)
4. **Data pipeline changes** → Update [docs/architecture/data-pipeline.md](../architecture/data-pipeline.md)

Then commit docs before creating PR.

### Godoc Standards

All exported types and functions should have godoc comments:

```go
// AnalyzeTemplate extracts keywords, categories, and notability metrics
// from a template. It skips analysis if the template's SHA hasn't changed
// since the last analysis (unless ForceAnalyze is true).
//
// Returns error if template parsing fails or required resources are unavailable.
func (a *Analyzer) AnalyzeTemplate(template *types.Template, repoInfo *types.Repository) error {
    // ...
}
```

---

## Related Documentation

- **[Backend Design](../architecture/backend-design.md)** - Architecture patterns and implementation
- **[.claude/instructions](../../.claude/instructions.md)** - Complete AI workflow (includes this content)
- **[Getting Started](../guides/getting-started.md)** - Setup and development workflow
- **[UI/UX Guidelines](../guides/ui-ux-guidelines.md)** - Frontend design system

**Historical**:
- **[Backend Refactoring History](../history/backend-refactoring/)** - How we got here
