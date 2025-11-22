# Branch Protection Setup

This document explains how to configure branch protection rules to prevent merging PRs with failing tests.

## GitHub Actions CI Workflow

The repository now includes a CI workflow (`.github/workflows/ci.yml`) that runs on all pull requests to the `main` branch. This workflow:

1. **Tests**: Runs `go test` with race detection and coverage
2. **Build**: Compiles the `lima-catalog` binary to ensure it builds successfully
3. **Lint**: Runs `go vet` and `golangci-lint` for code quality checks

## Setting Up Branch Protection

To enable branch protection and require CI checks to pass before merging:

### Step 1: Access Branch Protection Settings

1. Go to the GitHub repository: https://github.com/lima-catalog/lima-catalog
2. Click **Settings** (requires admin access)
3. Navigate to **Branches** in the left sidebar
4. Under "Branch protection rules", click **Add rule** or edit the existing rule for `main`

### Step 2: Configure Protection Rules

Configure the following settings:

#### Required Settings:
- **Branch name pattern**: `main`
- ✅ **Require a pull request before merging**
  - Require approvals: 1 (optional, but recommended)
  - Dismiss stale pull request approvals when new commits are pushed: ✅ (recommended)
- ✅ **Require status checks to pass before merging**
  - ✅ Require branches to be up to date before merging
  - **Status checks that are required**:
    - `Test` (from CI workflow)
    - `Build` (from CI workflow)
    - `Lint` (from CI workflow)

#### Optional but Recommended Settings:
- ✅ **Require conversation resolution before merging**
- ✅ **Do not allow bypassing the above settings** (prevents admins from bypassing protections)

### Step 3: Save the Rule

Click **Create** or **Save changes** to apply the branch protection rule.

## What This Prevents

With these settings enabled:

1. ❌ **Cannot merge PRs with failing tests** - The compilation error from PR #121 would have been caught
2. ❌ **Cannot push directly to `main`** - All changes must go through a pull request
3. ❌ **Cannot merge without required checks** - The CI workflow must complete successfully
4. ✅ **Forces re-running tests** - When new commits are pushed to a PR, tests automatically re-run

## Testing the Setup

After configuring branch protection:

1. Create a test PR that intentionally breaks compilation
2. Verify that the CI checks fail
3. Confirm that the PR shows "Merge blocked" status
4. Fix the issue and push a new commit
5. Verify that CI re-runs and passes
6. Confirm that the PR can now be merged

## Troubleshooting

### "No status checks found" when setting up required checks

This happens when the CI workflow hasn't run yet. To fix:

1. Merge the CI workflow into `main` first (without branch protection)
2. Create a test PR to trigger the workflow
3. Once the workflow runs, the status checks will appear in the branch protection settings
4. Add the checks as required and enable branch protection

### Existing PRs blocked after enabling protection

Existing PRs may need to be updated to trigger the CI workflow:

1. Make a trivial commit (e.g., add a comment or whitespace)
2. Push to the PR branch
3. Wait for CI to complete
4. The PR should now be mergeable (if tests pass)

## Best Practices

1. **Always re-run `/pr` after pushing new commits** to a PR branch
2. **Review CI logs** if tests fail to understand what broke
3. **Keep PRs focused** to make testing and review easier
4. **Run tests locally** before pushing (`go test ./...` and `go build ./...`)

## References

- [GitHub Branch Protection Documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches)
- [GitHub Actions Status Checks](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/collaborating-on-repositories-with-code-quality-features/about-status-checks)
