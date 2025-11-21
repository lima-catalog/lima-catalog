/**
 * URL generation utilities for GitHub and Lima URLs
 */

/**
 * Get URL with default branch instead of commit SHA
 * Uses the raw_url field to construct the display URL with default branch
 * @param {Object} template - Template object with raw_url field
 * @returns {string} URL with default branch
 */
export function getDefaultBranchURL(template) {
    // If we have raw_url, convert it to display URL
    // From: https://raw.githubusercontent.com/owner/repo/branch/path
    // To: https://github.com/owner/repo/blob/branch/path
    if (template.raw_url) {
        // Parse the raw URL
        // Format: https://raw.githubusercontent.com/OWNER/REPO/BRANCH/PATH
        const rawPattern = /^https:\/\/raw\.githubusercontent\.com\/([^\/]+)\/([^\/]+)\/([^\/]+)\/(.+)$/;
        const match = template.raw_url.match(rawPattern);

        if (match) {
            const [, owner, repo, branch, path] = match;
            return `https://github.com/${owner}/${repo}/blob/${branch}/${path}`;
        }
    }

    // Fallback to original URL
    return template.url;
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
