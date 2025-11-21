/**
 * Centralized application state management
 */

// State
let templates = [];
let filteredTemplates = [];
let selectedKeywords = new Set();
let selectedCategory = null;
let debugMode = false;
let focusedTemplate = null;

// Getters
export function getTemplates() {
    return templates;
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

export function getFocusedTemplate() {
    return focusedTemplate;
}

// Setters
export function setTemplates(newTemplates) {
    templates = newTemplates;
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

export function setFocusedTemplate(template) {
    focusedTemplate = template;
}

export function clearAllSelections() {
    selectedKeywords.clear();
    selectedCategory = null;
}

// Debug mode
export function isDebugMode() {
    return debugMode;
}

export function toggleDebugMode() {
    debugMode = !debugMode;
    return debugMode;
}
