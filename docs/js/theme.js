/**
 * Theme management module
 * Handles dark/light/auto theme switching
 */

const THEME_STORAGE_KEY = 'lima-catalog-theme';
const HIGHLIGHT_LIGHT = 'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/atom-one-light.min.css';
const HIGHLIGHT_DARK = 'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/atom-one-dark.min.css';

/**
 * Get system color scheme preference
 * @returns {string} 'light' or 'dark'
 */
function getSystemTheme() {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

/**
 * Get current theme preference
 * @returns {string} 'light', 'dark', or 'auto'
 */
export function getCurrentTheme() {
    return localStorage.getItem(THEME_STORAGE_KEY) || 'auto';
}

/**
 * Get effective theme (resolves 'auto' to actual theme)
 * @returns {string} 'light' or 'dark'
 */
export function getEffectiveTheme() {
    const theme = getCurrentTheme();
    return theme === 'auto' ? getSystemTheme() : theme;
}

/**
 * Apply theme to document
 * @param {string} theme - 'light' or 'dark'
 */
function applyTheme(theme) {
    if (theme === 'dark') {
        document.documentElement.setAttribute('data-theme', 'dark');
    } else {
        document.documentElement.removeAttribute('data-theme');
    }

    // Update highlight.js theme
    const highlightLink = document.querySelector('link[href*="highlight.js"]');
    if (highlightLink) {
        highlightLink.href = theme === 'dark' ? HIGHLIGHT_DARK : HIGHLIGHT_LIGHT;
    }
}

/**
 * Set theme preference
 * @param {string} theme - 'light', 'dark', or 'auto'
 */
export function setTheme(theme) {
    localStorage.setItem(THEME_STORAGE_KEY, theme);

    // Apply the effective theme
    const effectiveTheme = theme === 'auto' ? getSystemTheme() : theme;
    applyTheme(effectiveTheme);

    // Update button states
    updateThemeButtons(theme);
}

/**
 * Update active state of theme buttons
 * @param {string} currentTheme - 'light', 'dark', or 'auto'
 */
function updateThemeButtons(currentTheme) {
    document.querySelectorAll('.theme-option').forEach(button => {
        const buttonTheme = button.dataset.theme;
        if (buttonTheme === currentTheme) {
            button.classList.add('active');
            button.setAttribute('aria-pressed', 'true');
        } else {
            button.classList.remove('active');
            button.setAttribute('aria-pressed', 'false');
        }
    });
}

/**
 * Initialize theme system
 */
export function initializeTheme() {
    // Apply initial theme
    const currentTheme = getCurrentTheme();
    const effectiveTheme = getEffectiveTheme();
    applyTheme(effectiveTheme);
    updateThemeButtons(currentTheme);

    // Listen for system theme changes (when in auto mode)
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    mediaQuery.addEventListener('change', (e) => {
        if (getCurrentTheme() === 'auto') {
            applyTheme(e.matches ? 'dark' : 'light');
        }
    });

    // Setup button click handlers
    document.querySelectorAll('.theme-option').forEach(button => {
        button.addEventListener('click', () => {
            const theme = button.dataset.theme;
            setTheme(theme);
        });
    });
}
