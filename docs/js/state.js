/**
 * Centralized application state management
 */

// State
let templates = [];
let repositories = new Map();
let filteredTemplates = [];
let selectedKeywords = new Set();
let selectedCategory = null;

// Getters
export function getTemplates() {
    return templates;
}

export function getRepositories() {
    return repositories;
}

export function getFilteredTemplates() {
    return filteredTemplates;
}

export function getSelectedKeywords() {
    return selectedKeywords;
}

export function getSelectedCategory() {
    return selectedCategory;
}

// Setters
export function setTemplates(newTemplates) {
    templates = newTemplates;
}

export function setRepositories(newRepositories) {
    repositories = newRepositories;
}

export function setFilteredTemplates(newFilteredTemplates) {
    filteredTemplates = newFilteredTemplates;
}

export function toggleKeywordSelection(keyword) {
    if (selectedKeywords.has(keyword)) {
        selectedKeywords.delete(keyword);
    } else {
        selectedKeywords.add(keyword);
    }
}

export function clearKeywordSelection() {
    selectedKeywords.clear();
}

export function setCategorySelection(category) {
    selectedCategory = category;
}

export function toggleCategorySelection(category) {
    selectedCategory = selectedCategory === category ? null : category;
}

export function clearCategorySelection() {
    selectedCategory = null;
}

export function clearAllSelections() {
    selectedKeywords.clear();
    selectedCategory = null;
}
