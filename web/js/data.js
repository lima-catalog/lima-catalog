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
 * Load combined catalog data from GitHub
 * This file contains templates with all necessary repo/org data pre-joined
 * @returns {Promise<Array>} Array of template objects with embedded metadata
 */
export async function loadCatalog() {
    const response = await fetch(`${DATA_BASE_URL}/catalog.jsonl`);
    if (!response.ok) {
        throw new Error(`Failed to load catalog: HTTP ${response.status}`);
    }
    const text = await response.text();
    return parseJsonLines(text);
}

/**
 * Load all data (templates with embedded repo/org metadata)
 * @returns {Promise<Object>} Object with templates array
 */
export async function loadAllData() {
    const templates = await loadCatalog();
    return { templates };
}
