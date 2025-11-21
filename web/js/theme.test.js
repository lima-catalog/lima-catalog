/**
 * Tests for theme management
 */

import { runner, assert } from './test-framework.js';
import { getCurrentTheme, getEffectiveTheme, setTheme } from './theme.js';

// Mock localStorage
const localStorageMock = (() => {
    let store = {};
    return {
        getItem: (key) => store[key] || null,
        setItem: (key, value) => { store[key] = value.toString(); },
        clear: () => { store = {}; }
    };
})();

// Mock matchMedia
const createMatchMediaMock = (matches) => {
    return (query) => ({
        matches,
        media: query,
        addEventListener: () => {},
        removeEventListener: () => {}
    });
};

// Setup mocks
if (typeof global !== 'undefined') {
    global.localStorage = localStorageMock;
    global.window = global.window || {};
    global.window.matchMedia = createMatchMediaMock(false); // Default to light mode
}

runner.test('getCurrentTheme: returns auto by default', () => {
    localStorageMock.clear();
    const theme = getCurrentTheme();
    assert.equal(theme, 'auto', 'Default theme should be auto');
});

runner.test('getCurrentTheme: returns stored theme', () => {
    localStorageMock.clear();
    localStorageMock.setItem('lima-catalog-theme', 'dark');
    const theme = getCurrentTheme();
    assert.equal(theme, 'dark', 'Should return stored dark theme');
});

runner.test('getEffectiveTheme: resolves auto to system preference', () => {
    localStorageMock.clear();
    // Mock system preference as light
    if (typeof global !== 'undefined') {
        global.window.matchMedia = createMatchMediaMock(false);
    }
    const theme = getEffectiveTheme();
    assert.equal(theme, 'light', 'Auto should resolve to light when system prefers light');
});

runner.test('getEffectiveTheme: resolves auto to dark when system prefers dark', () => {
    localStorageMock.clear();
    // Mock system preference as dark
    if (typeof global !== 'undefined') {
        global.window.matchMedia = createMatchMediaMock(true);
    }
    const theme = getEffectiveTheme();
    assert.equal(theme, 'dark', 'Auto should resolve to dark when system prefers dark');
});

runner.test('getEffectiveTheme: returns explicit light theme', () => {
    localStorageMock.clear();
    localStorageMock.setItem('lima-catalog-theme', 'light');
    const theme = getEffectiveTheme();
    assert.equal(theme, 'light', 'Should return explicit light theme');
});

runner.test('getEffectiveTheme: returns explicit dark theme', () => {
    localStorageMock.clear();
    localStorageMock.setItem('lima-catalog-theme', 'dark');
    const theme = getEffectiveTheme();
    assert.equal(theme, 'dark', 'Should return explicit dark theme');
});

runner.test('setTheme: stores theme preference', () => {
    localStorageMock.clear();
    setTheme('dark');
    const stored = localStorageMock.getItem('lima-catalog-theme');
    assert.equal(stored, 'dark', 'Should store dark theme preference');
});

runner.test('setTheme: stores light theme preference', () => {
    localStorageMock.clear();
    setTheme('light');
    const stored = localStorageMock.getItem('lima-catalog-theme');
    assert.equal(stored, 'light', 'Should store light theme preference');
});

runner.test('setTheme: stores auto theme preference', () => {
    localStorageMock.clear();
    setTheme('auto');
    const stored = localStorageMock.getItem('lima-catalog-theme');
    assert.equal(stored, 'auto', 'Should store auto theme preference');
});
