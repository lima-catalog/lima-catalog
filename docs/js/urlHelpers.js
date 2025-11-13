/**
 * URL generation utilities for GitHub and Lima URLs
 */

/**
 * Get URL with default branch instead of commit SHA
 * @param {Object} template - Template object
 * @param {Object} repo - Repository object
 * @returns {string} URL with default branch
 */
export function getDefaultBranchURL(template, repo) {
    if (!repo || !repo.default_branch) {
        return template.url; // Fallback to original URL
    }

    // Pattern: https://github.com/owner/repo/blob/COMMIT_SHA/path
    const urlPattern = /^https:\/\/github\.com\/([^\/]+\/[^\/]+)\/blob\/([a-f0-9]{40})\/(.+)$/;
    const match = template.url.match(urlPattern);

    if (!match) {
        return template.url; // Not a commit URL, return as-is
    }

    const [, repoPath, , filePath] = match;
    return `https://github.com/${repoPath}/blob/${repo.default_branch}/${filePath}`;
}

/**
 * Generate shortest possible github: URL for Lima
 * @param {Object} template - Template object
 * @returns {string} Lima github: scheme URL
 */
export function getGitHubSchemeURL(template) {
    // Parse the template.repo (format: "owner/repo")
    const [owner, repo] = template.repo.split('/');
    let path = template.path;

    // Remove .yaml extension (Lima adds .yaml automatically, not .yml)
    path = path.replace(/\.yaml$/, '');

    // If path ends with .lima, remove it (default filename)
    path = path.replace(/\/\.lima$/, '');

    // If path is just .lima (root), can omit entirely
    if (path === '.lima' || path === '') {
        // For org repos (owner == repo), use shortest format
        if (owner === repo) {
            return `github:${owner}`;
        }
        return `github:${owner}/${repo}`;
    }

    // For org repos (owner == repo), use double slash shorthand
    if (owner === repo) {
        return `github:${owner}//${path}`;
    }

    // Standard format
    return `github:${owner}/${repo}/${path}`;
}

/**
 * Convert GitHub blob URL to raw content URL
 * @param {string} url - GitHub blob URL
 * @returns {string} Raw content URL
 */
export function getRawContentURL(url) {
    let rawURL = url.replace('github.com', 'raw.githubusercontent.com');
    rawURL = rawURL.replace('/blob/', '/');
    return rawURL;
}
