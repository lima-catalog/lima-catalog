/**
 * Utility functions
 */

/**
 * Debounce function - delays execution until after wait time has elapsed
 * @param {Function} func - Function to debounce
 * @param {number} wait - Wait time in milliseconds
 * @returns {Function} Debounced function
 */
export function debounce(func, wait = 300) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

/**
 * Trap focus within an element for keyboard accessibility
 * @param {HTMLElement} element - Element to trap focus within
 * @returns {Function} Cleanup function
 */
export function trapFocus(element) {
    const selector = 'a[href]:not([tabindex="-1"]), button:not([disabled]):not([tabindex="-1"]), textarea:not([disabled]):not([tabindex="-1"]), input:not([disabled]):not([tabindex="-1"]), select:not([disabled]):not([tabindex="-1"]), [tabindex]:not([tabindex="-1"])';

    // Get visible focusable elements (recomputed on each call to handle dynamic content)
    const getVisibleFocusable = () => {
        const all = element.querySelectorAll(selector);
        return Array.from(all).filter(el => el.offsetParent !== null);
    };

    const handleTabKey = (e) => {
        if (e.key !== 'Tab') return;

        // Recompute on each Tab press to handle tab switching, dynamic content
        const focusableElements = getVisibleFocusable();
        if (focusableElements.length === 0) return;

        const firstFocusable = focusableElements[0];
        const lastFocusable = focusableElements[focusableElements.length - 1];

        if (e.shiftKey && document.activeElement === firstFocusable) {
            e.preventDefault();
            lastFocusable.focus();
        } else if (!e.shiftKey && document.activeElement === lastFocusable) {
            e.preventDefault();
            firstFocusable.focus();
        }
    };

    element.addEventListener('keydown', handleTabKey);

    // Initial focus
    const initialFocusable = getVisibleFocusable();
    if (initialFocusable.length > 0) {
        initialFocusable[0].focus();
    }

    return () => {
        element.removeEventListener('keydown', handleTabKey);
    };
}
