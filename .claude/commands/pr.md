# Create Pull Request

Follow these steps exactly when creating a PR:

## 1. Gather Information
First, fetch the base branch:
```bash
git fetch origin main
```

Then gather info (can run in parallel):
- `git status` - check for uncommitted changes
- `git log --oneline origin/main..HEAD` - see all commits to include
- `git diff origin/main...HEAD --stat` - summary of all changes

## 2. Check Test Coverage and Test Comments

**IMPORTANT: Do this BEFORE running tests.**

### 2.1 Verify Test Coverage for Code Changes

For any code changes (non-test files), verify that corresponding tests exist:

```bash
git diff origin/main...HEAD --name-only --diff-filter=AM | grep -E '\.(js|go)$' | grep -v '\.test\.' | grep -v '\.spec\.' | grep -v '_test\.go'
```

For each changed code file, ask yourself:
- Does this change require new test cases?
- Are there existing tests that need to be updated?
- If this is a new file, does it have corresponding test coverage?

Common patterns:
- `web/js/foo.js` → Should have tests in `web/js/foo.test.js`
- `pkg/bar/baz.go` → Should have tests in `pkg/bar/baz_test.go`
- New features → Need new test cases
- Bug fixes → Should add regression tests
- Refactoring → Existing tests should still pass and may need updates

### 2.2 Verify Test Comment Headers

For any changed test files, verify the overview comment at the top is up-to-date:

```bash
git diff origin/main...HEAD --name-only --diff-filter=AM | grep -E '(\.test\.|\.spec\.|_test\.go)'
```

For each changed test file:
- Check if the overview comment block at the top accurately describes what's tested
- E2E tests should have format: `E2E Tests: <Feature Name>`
- Unit tests should have format: `Unit Tests: <Module Name>`
- The bullet points should list all major test categories
- Update the comment if tests were added, removed, or significantly changed

Example patterns:
- Added new test cases → Add bullet point to overview
- Removed test cases → Remove or update bullet points
- Changed test behavior → Update description to match
- New test file → Must have comprehensive overview comment

## 3. Run Tests, Linting, and Rebase
Run the automated PR preparation workflow:

```bash
make pr
```

This will:
- Check for uncommitted changes (fails if any)
- Fetch origin/main
- Show changed files
- Run appropriate tests based on file types:
  - Only JS files changed → `make test-js`
  - Only Go files changed → `make vet test-go` + `make lint`
  - Both changed → `make test` (vet + test-go + test-js) + `make lint`
  - No code files → Skip tests (CI will run them)
- Rebase on origin/main
- Exit with error if conflicts detected

If tests fail or conflicts occur, fix them and commit before running `make pr` again.

## 4. Push Changes
If `make pr` succeeded, push your rebased branch:

```bash
git push -f
```

## 5. Check Documentation
Review if any documentation needs updates based on the changes:
- `docs/architecture/frontend-design.md` - for frontend changes
- `docs/architecture/backend-design.md` - for backend changes
- `docs/guides/ui-ux-guidelines.md` - for UI/UX changes
- Other relevant docs in `docs/`

If docs need updates, make the changes, commit, and run `make pr` again.

## 6. Create PR Command
Provide a copyable `gh pr create` command using HEREDOC format:

```bash
gh pr create --title "Title here" --body "$(cat <<'EOF'
## Summary
- Bullet points describing changes

## Test plan
- [ ] Test step 1
- [ ] Test step 2
EOF
)"
```

### ⚠️ WARNING: Do NOT Use Triple Backticks in PR Body

The entire `gh pr create` command is wrapped in a code block. If you use triple backticks (```) inside the HEREDOC body, it will terminate the outer code block prematurely and break the command.

**WRONG:**
```bash
gh pr create --body "$(cat <<'EOF'
Summary here

```javascript
// This breaks the outer code block!
const foo = 'bar';
```
EOF
)"
```

**CORRECT - Use 4-space indentation instead:**
```bash
gh pr create --body "$(cat <<'EOF'
Summary here

    // 4-space indentation for code
    const foo = 'bar';
EOF
)"
```

**CORRECT - Use inline backticks for short snippets:**
```bash
gh pr create --body "$(cat <<'EOF'
Summary here

Updated the `foo()` function to handle edge cases.
EOF
)"
```

## Important
- Always provide the `gh pr create` command for the user to copy/paste
- Use HEREDOC format for the body to preserve formatting
- Include a test plan with checkboxes
- Keep summary concise (3-5 bullet points)
- **NEVER use triple backticks (```) inside the PR body** - use 4-space indentation or inline backticks instead
