# Create Pull Request

Follow these steps exactly when creating a PR:

## 1. Gather Information (run in parallel)
- `git status` - check for uncommitted changes
- `git diff` - see staged and unstaged changes
- `git log --oneline <base>..HEAD` - see all commits to include
- `git diff <base>...HEAD --stat` - summary of all changes

## 2. Run Tests and Linting
Before creating a PR, ensure all tests and linting checks pass.

**Smart test selection based on changed files:**

```bash
# Detect which files changed
CHANGED_FILES=$(git diff <base>...HEAD --name-only)
HAS_GO_FILES=$(echo "$CHANGED_FILES" | grep -E '\.(go|mod|sum)$' || true)
HAS_JS_FILES=$(echo "$CHANGED_FILES" | grep -E '\.(js|html|css)$' || true)

# Run appropriate tests
if [ -n "$HAS_GO_FILES" ] && [ -n "$HAS_JS_FILES" ]; then
    echo "Running all tests (Go + JS changed)..."
    make test
elif [ -n "$HAS_GO_FILES" ]; then
    echo "Running Go tests only..."
    make test-go
elif [ -n "$HAS_JS_FILES" ]; then
    echo "Running JS tests only..."
    make test-js
else
    echo "No Go or JS files changed, running all tests to be safe..."
    make test
fi

# Run linting (if golangci-lint is installed and Go files changed)
if [ -n "$HAS_GO_FILES" ]; then
    make lint
fi
```

If any tests fail or linting issues are found, fix them and commit before proceeding.

## 3. Check Documentation
Review if any documentation needs updates based on the changes:
- `docs/architecture/frontend-design.md` - for frontend changes
- `docs/architecture/backend-design.md` - for backend changes
- `docs/guides/ui-ux-guidelines.md` - for UI/UX changes
- Other relevant docs in `docs/`

If docs need updates, make the changes and commit before proceeding.

## 4. Push Changes
```bash
git push -u origin <branch-name>
```

## 5. Create PR Command
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

## Important
- Always provide the `gh pr create` command for the user to copy/paste
- Use HEREDOC format for the body to preserve formatting
- Include a test plan with checkboxes
- Keep summary concise (3-5 bullet points)
