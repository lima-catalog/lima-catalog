# Create Pull Request

Follow these steps exactly when creating a PR:

## 1. Gather Information (run in parallel)
- `git status` - check for uncommitted changes
- `git diff` - see staged and unstaged changes
- `git log --oneline <base>..HEAD` - see all commits to include
- `git diff <base>...HEAD --stat` - summary of all changes

## 2. Run Tests, Linting, and Rebase
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

## 3. Push Changes
If `make pr` succeeded, push your rebased branch:

```bash
git push -f
```

## 4. Check Documentation
Review if any documentation needs updates based on the changes:
- `docs/architecture/frontend-design.md` - for frontend changes
- `docs/architecture/backend-design.md` - for backend changes
- `docs/guides/ui-ux-guidelines.md` - for UI/UX changes
- Other relevant docs in `docs/`

If docs need updates, make the changes, commit, and run `make pr` again.

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
