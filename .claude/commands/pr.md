# Create Pull Request

Follow these steps exactly when creating a PR:

## 1. Gather Information (run in parallel)
- `git status` - check for uncommitted changes
- `git diff` - see staged and unstaged changes
- `git log --oneline <base>..HEAD` - see all commits to include
- `git diff <base>...HEAD --stat` - summary of all changes

## 2. Check Documentation
Review if any documentation needs updates based on the changes:
- `docs/architecture/frontend-design.md` - for frontend changes
- `docs/architecture/backend-design.md` - for backend changes
- `docs/guides/ui-ux-guidelines.md` - for UI/UX changes
- Other relevant docs in `docs/`

If docs need updates, make the changes and commit before proceeding.

## 3. Push Changes
```bash
git push -u origin <branch-name>
```

## 4. Create PR Command
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
