/**
 * Data fetching and parsing utilities
 */

import { DATA_BASE_URL } from './config.js';

/**
 * Parse JSON Lines format (one JSON object per line)
 * @param {string} text - JSONL text
 * @returns {Array} Parsed objects
 */
export function parseJsonLines(text) {
    return text
        .trim()
        .split('\n')
        .filter(line => line.trim())
        .map(line => JSON.parse(line));
}

/**
 * Load templates data from GitHub
 * @returns {Promise<Array>} Array of template objects
 */
export async function loadTemplates() {
    const response = await fetch(`${DATA_BASE_URL}/templates.jsonl`);
    if (!response.ok) {
        throw new Error('Failed to load templates');
    }
    const text = await response.text();
    return parseJsonLines(text);
}

/**
 * Load repositories data from GitHub
 * @returns {Promise<Map>} Map of repo ID to repo object
 */
export async function loadRepositories() {
    const response = await fetch(`${DATA_BASE_URL}/repos.jsonl`);
    if (!response.ok) {
        throw new Error('Failed to load repositories');
    }
    const text = await response.text();
    const repos = parseJsonLines(text);

    const repoMap = new Map();
    repos.forEach(repo => repoMap.set(repo.id, repo));
    return repoMap;
}

/**
 * Load all data (templates and repositories)
 * @returns {Promise<Object>} Object with templates and repositories
 */
export async function loadAllData() {
    const [templates, repositories] = await Promise.all([
        loadTemplates(),
        loadRepositories()
    ]);

    return { templates, repositories };
}
