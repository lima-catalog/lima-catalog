/**
 * Keyboard shortcuts and navigation
 */

import { clearSearch, filterAndRender, updateSortDropdown, showDebugModeNotification } from './appActions.js';
import * as State from './state.js';
import { trapFocus } from './utils.js';

function getFirstVisibleTemplateCard() {
    const cards = Array.from(document.querySelectorAll('.template-card'));
    if (cards.length === 0) return null;

    const viewportTop = window.scrollY;
    const viewportBottom = viewportTop + window.innerHeight;

    // Find first card that's at least partially in viewport
    for (const card of cards) {
        const rect = card.getBoundingClientRect();
        const cardTop = rect.top + window.scrollY;
        const cardBottom = cardTop + rect.height;

        // Card is visible if it overlaps with viewport
        if (cardBottom > viewportTop && cardTop < viewportBottom) {
            return card;
        }
    }

    // Fallback to first card
    return cards[0];
}

/**
 * Get the last card that's at least partially visible in viewport
 * @returns {HTMLElement|null} The last visible card or null
 */
function getLastVisibleTemplateCard() {
    const cards = Array.from(document.querySelectorAll('.template-card'));
    if (cards.length === 0) return null;

    const viewportTop = window.scrollY;
    const viewportBottom = viewportTop + window.innerHeight;

    let lastVisible = null;

    for (const card of cards) {
        const rect = card.getBoundingClientRect();
        const cardTop = rect.top + window.scrollY;
        const cardBottom = cardTop + rect.height;

        // Card is visible if it overlaps with viewport
        if (cardBottom > viewportTop && cardTop < viewportBottom) {
            lastVisible = card;
        } else if (cardTop >= viewportBottom) {
            // We've passed the viewport, stop searching
            break;
        }
    }

    return lastVisible || cards[cards.length - 1];
}

/**
 * Calculate how many rows of cards approximately fit in the viewport
 * @returns {number} Number of rows that fit in viewport
 */
function getRowsPerViewport() {
    const cards = Array.from(document.querySelectorAll('.template-card'));
    if (cards.length === 0) return 3; // Default fallback

    // Get first card to measure height
    const firstCard = cards[0];
    const cardRect = firstCard.getBoundingClientRect();
    const cardHeight = cardRect.height;

    // Get grid to measure gap
    const grid = firstCard.parentElement;
    const gridStyle = window.getComputedStyle(grid);
    const gap = parseInt(gridStyle.rowGap) || 0;

    // Calculate rows per viewport (leave a bit of overlap for context)
    const effectiveCardHeight = cardHeight + gap;
    const rowsPerViewport = Math.max(1, Math.floor((window.innerHeight * 0.9) / effectiveCardHeight));

    return rowsPerViewport;
}

/**
 * Get number of columns in the grid
 * @returns {number} Number of columns
 */
function getGridColumnCount() {
    const grid = document.querySelector('.templates-grid');
    if (!grid) return 3; // Default fallback

    const gridStyle = window.getComputedStyle(grid);
    const gridTemplateColumns = gridStyle.gridTemplateColumns;
    return gridTemplateColumns.split(' ').length;
}

/**
 * Keyboard shortcut configuration
 * Each entry defines a keyboard shortcut with its behavior
 */
const KEYBOARD_SHORTCUTS = {
    // Help and search
    '/': {
        description: 'Focus search box',
        skipIfTyping: true,
        preventDefault: true,
        action: (e, ctx) => {
            ctx.searchInput.focus();
            ctx.searchInput.select();
        }
    },
    '?': {
        description: 'Show keyboard help',
        skipIfTyping: false, // Works everywhere, even in search
        preventDefault: true,
        action: (e, ctx) => showKeyboardHelp(ctx.isTypingInSearch, 'help')
    },

    // Ctrl+Arrow navigation between major sections
    'Ctrl+ArrowLeft': {
        description: 'Focus sidebar from templates',
        skipIfTyping: false,
        preventDefault: true,
        action: (e, ctx) => {
            ctx.searchInput.focus();
            ctx.searchInput.select();
        }
    },
    'Ctrl+ArrowRight': {
        description: 'Focus first template from sidebar',
        skipIfTyping: false,
        preventDefault: true,
        action: (e, ctx) => {
            const firstTemplate = document.querySelector('.template-card');
            if (firstTemplate) firstTemplate.focus();
        }
    },
    'Ctrl+ArrowUp': {
        description: 'Focus header (theme switcher)',
        skipIfTyping: false,
        preventDefault: true,
        action: (e, ctx) => {
            const themeButton = document.querySelector('.theme-switcher button');
            if (themeButton) themeButton.focus();
        }
    },
    'Ctrl+ArrowDown': {
        description: 'Focus first template from header',
        skipIfTyping: false,
        preventDefault: true,
        action: (e, ctx) => {
            const firstTemplate = document.querySelector('.template-card');
            if (firstTemplate) firstTemplate.focus();
        }
    },

    // Vertical scrolling keys
    'PageUp': {
        description: 'Scroll up one page of templates',
        skipIfTyping: true,
        preventDefault: true,
        action: (e, ctx) => {
            const cards = Array.from(document.querySelectorAll('.template-card'));
            if (cards.length === 0) return;

            // Get current card (focused or first visible)
            const currentCard = cards.includes(document.activeElement) ?
                document.activeElement : getFirstVisibleTemplateCard();
            if (!currentCard) return;

            const currentIndex = cards.indexOf(currentCard);
            const columnCount = getGridColumnCount();
            const currentColumn = currentIndex % columnCount;

            // If already on the top row, just scroll to show header
            if (currentIndex < columnCount) {
                window.scrollTo({ top: 0, behavior: 'smooth' });
                return;
            }

            const viewportTop = window.scrollY;
            const viewportBottom = viewportTop + window.innerHeight;

            // Find the first card in this column that's at or near the top of viewport
            let firstVisibleInColumn = null;
            for (let i = currentColumn; i < cards.length; i += columnCount) {
                const card = cards[i];
                const rect = card.getBoundingClientRect();
                const cardTop = rect.top + window.scrollY;
                const cardBottom = cardTop + rect.height;

                // Is this card visible?
                if (cardBottom > viewportTop && cardTop < viewportBottom) {
                    firstVisibleInColumn = card;
                    break;
                }
            }

            if (!firstVisibleInColumn) return;

            const firstVisibleIndex = cards.indexOf(firstVisibleInColumn);
            const firstVisibleRect = firstVisibleInColumn.getBoundingClientRect();
            const firstVisibleTop = firstVisibleRect.top + window.scrollY;

            // Check if first visible card in column is partially clipped at top
            const isClipped = firstVisibleTop < viewportTop + 10;

            let targetCard;
            if (isClipped && firstVisibleIndex >= columnCount) {
                // Use the clipped card
                targetCard = firstVisibleInColumn;
            } else {
                // Move up by one viewport worth of rows
                const rowsPerPage = getRowsPerViewport();
                const cardsToMove = columnCount * rowsPerPage;
                const targetIndex = Math.max(currentColumn, firstVisibleIndex - cardsToMove);
                targetCard = cards[targetIndex];
            }

            const targetIndex = cards.indexOf(targetCard);
            // If we're at the top row, scroll to show header
            if (targetIndex < columnCount) {
                window.scrollTo({ top: 0, behavior: 'smooth' });
            } else {
                // Scroll so target card is at the top of viewport with small offset for border/focus
                const rect = targetCard.getBoundingClientRect();
                const cardTop = rect.top + window.scrollY;
                const offset = 20; // Space for border and focus outline
                window.scrollTo({ top: Math.max(0, cardTop - offset), behavior: 'smooth' });
            }
            targetCard.focus({ preventScroll: true });
        }
    },
    'PageDown': {
        description: 'Scroll down one page of templates',
        skipIfTyping: true,
        preventDefault: true,
        action: (e, ctx) => {
            const cards = Array.from(document.querySelectorAll('.template-card'));
            if (cards.length === 0) return;

            // Get current card (focused or first visible)
            const currentCard = cards.includes(document.activeElement) ?
                document.activeElement : getFirstVisibleTemplateCard();
            if (!currentCard) return;

            const currentIndex = cards.indexOf(currentCard);
            const columnCount = getGridColumnCount();
            const currentColumn = currentIndex % columnCount;

            // If already on the last row, just scroll to show footer
            if (currentIndex >= cards.length - columnCount) {
                window.scrollTo({ top: document.documentElement.scrollHeight, behavior: 'smooth' });
                return;
            }

            const viewportTop = window.scrollY;
            const viewportBottom = viewportTop + window.innerHeight;

            // Find the LAST card in this column that's currently visible
            let lastVisibleInColumn = null;
            for (let i = currentColumn; i < cards.length; i += columnCount) {
                const card = cards[i];
                const rect = card.getBoundingClientRect();
                const cardTop = rect.top + window.scrollY;
                const cardBottom = cardTop + rect.height;

                // Is this card visible?
                if (cardBottom > viewportTop && cardTop < viewportBottom) {
                    lastVisibleInColumn = card;
                    // Keep going to find the LAST visible card
                } else if (cardTop >= viewportBottom) {
                    // We've gone past the viewport
                    break;
                }
            }

            if (!lastVisibleInColumn) return;

            const lastVisibleIndex = cards.indexOf(lastVisibleInColumn);
            const lastVisibleRect = lastVisibleInColumn.getBoundingClientRect();
            const lastVisibleBottom = lastVisibleRect.top + window.scrollY + lastVisibleRect.height;

            // Check if last visible card in column is partially clipped at bottom
            const isClipped = lastVisibleBottom > viewportBottom - 10;

            let targetCard;
            if (isClipped) {
                // Use the clipped card as target
                targetCard = lastVisibleInColumn;
            } else {
                // All visible cards are fully visible, move to next card in column
                const nextIndex = lastVisibleIndex + columnCount;
                if (nextIndex < cards.length && nextIndex % columnCount === currentColumn) {
                    targetCard = cards[nextIndex];
                } else {
                    // No more cards below, use last card in column
                    targetCard = lastVisibleInColumn;
                }
            }

            const targetIndex = cards.indexOf(targetCard);
            // If we're at or near the last row, scroll to show footer
            if (targetIndex >= cards.length - columnCount) {
                window.scrollTo({ top: document.documentElement.scrollHeight, behavior: 'smooth' });
            } else {
                // Scroll so target card is at the top of viewport with small offset for border/focus
                const rect = targetCard.getBoundingClientRect();
                const cardTop = rect.top + window.scrollY;
                const offset = 20; // Space for border and focus outline
                window.scrollTo({ top: Math.max(0, cardTop - offset), behavior: 'smooth' });
            }
            targetCard.focus({ preventScroll: true });
        }
    },
    'Home': {
        description: 'Focus first template card',
        skipIfTyping: true,
        preventDefault: true,
        action: (e, ctx) => {
            const firstCard = document.querySelector('.template-card');
            if (firstCard) {
                // Scroll to top of page to show header
                window.scrollTo({ top: 0, behavior: 'smooth' });
                firstCard.focus();
            }
        }
    },
    'End': {
        description: 'Focus last template card',
        skipIfTyping: true,
        preventDefault: true,
        action: (e, ctx) => {
            const cards = document.querySelectorAll('.template-card');
            if (cards.length > 0) {
                const lastCard = cards[cards.length - 1];
                // Scroll to bottom of page to show footer
                window.scrollTo({ top: document.documentElement.scrollHeight, behavior: 'smooth' });
                lastCard.focus();
            }
        }
    },

    // Auto-focus on scroll (without Ctrl modifier)
    'ArrowUp': {
        description: 'Auto-focus template when scrolling up',
        skipIfTyping: true,
        preventDefault: false,
        requiresNoModifiers: true, // Only plain ArrowUp
        condition: (e, ctx) => {
            const activeElement = document.activeElement;
            const isInHeader = activeElement && (
                activeElement.closest('.theme-switcher') ||
                activeElement.id === 'keyboard-help-btn'
            );
            // Only auto-focus if not in an interactive element or header
            return !isInHeader && (!activeElement || activeElement === document.body || activeElement.tagName === 'HTML');
        },
        action: (e, ctx) => {
            setTimeout(() => {
                // Don't auto-focus if we're at the very top (header visible)
                if (window.scrollY <= 50) return;
                const visibleCard = getFirstVisibleTemplateCard();
                if (visibleCard) visibleCard.focus();
            }, 100);
        }
    },
    'ArrowDown': {
        description: 'Auto-focus template when scrolling down',
        skipIfTyping: true,
        preventDefault: false,
        requiresNoModifiers: true, // Only plain ArrowDown
        condition: (e, ctx) => {
            const activeElement = document.activeElement;
            const isInHeader = activeElement && (
                activeElement.closest('.theme-switcher') ||
                activeElement.id === 'keyboard-help-btn'
            );
            // Only auto-focus if not in an interactive element or header
            return !isInHeader && (!activeElement || activeElement === document.body || activeElement.tagName === 'HTML');
        },
        action: (e, ctx) => {
            setTimeout(() => {
                // Don't auto-focus if we're at the very bottom (footer visible)
                const scrolledToBottom = window.scrollY + window.innerHeight >= document.documentElement.scrollHeight - 50;
                if (scrolledToBottom) return;
                const visibleCard = getFirstVisibleTemplateCard();
                if (visibleCard) visibleCard.focus();
            }, 100);
        }
    },

    // Letter shortcuts (work with Shift, even when typing)
    'k': {
        description: 'Focus first keyword',
        skipIfTyping: true,
        allowsShift: true, // K works even when typing
        preventDefault: true,
        action: (e, ctx) => {
            const firstSelected = document.querySelector('.selected-keyword');
            const firstKeyword = document.querySelector('.keyword-tag');
            if (firstSelected) {
                firstSelected.focus();
            } else if (firstKeyword) {
                firstKeyword.focus();
            }
        }
    },
    'c': {
        description: 'Focus first category',
        skipIfTyping: true,
        allowsShift: true, // C works even when typing
        preventDefault: true,
        action: (e, ctx) => {
            const firstCategory = document.querySelector('.category-item');
            if (firstCategory) firstCategory.focus();
        }
    },
    's': {
        description: 'Focus sort dropdown',
        skipIfTyping: true,
        allowsShift: true, // S works even when typing
        preventDefault: true,
        action: (e, ctx) => {
            const sortDropdown = document.getElementById('sort');
            if (sortDropdown) sortDropdown.focus();
        }
    },
    't': {
        description: 'Focus first template card',
        skipIfTyping: true,
        allowsShift: true, // T works even when typing
        preventDefault: true,
        action: (e, ctx) => {
            const firstTemplate = document.querySelector('.template-card');
            if (firstTemplate) firstTemplate.focus();
        }
    },
    '@': {
        description: 'Toggle debug mode',
        skipIfTyping: false, // Works everywhere, even in search
        preventDefault: true,
        action: (e, ctx) => {
            const newDebugMode = State.toggleDebugMode();
            // Update sort dropdown to show/hide debug options
            updateSortDropdown();
            // Re-render templates to show/hide debug info
            filterAndRender();
            // Show a subtle notification
            showDebugModeNotification(newDebugMode);
        }
    },
    'o': {
        description: 'Toggle ORG keyword filter',
        skipIfTyping: true,
        preventDefault: true,
        action: (e, ctx) => {
            const focusedTemplate = State.getFocusedTemplate();
            if (focusedTemplate && focusedTemplate.org) {
                const orgKeyword = focusedTemplate.org;
                State.toggleKeywordSelection(orgKeyword);
                filterAndRender();
            }
        }
    },
    'r': {
        description: 'Toggle ORG/REPO keyword filter',
        skipIfTyping: true,
        preventDefault: true,
        action: (e, ctx) => {
            const focusedTemplate = State.getFocusedTemplate();
            if (focusedTemplate && focusedTemplate.repo) {
                const repoKeyword = focusedTemplate.repo;
                State.toggleKeywordSelection(repoKeyword);
                filterAndRender();
            }
        }
    }
};

/**
 * Build a key string from a keyboard event for matching against shortcuts
 */
function getKeyString(e) {
    const modifiers = [];
    if (e.ctrlKey) modifiers.push('Ctrl');
    if (e.altKey) modifiers.push('Alt');
    if (e.metaKey) modifiers.push('Meta');

    if (modifiers.length > 0) {
        return `${modifiers.join('+')}+${e.key}`;
    }
    return e.key;
}

function setupKeyboardShortcuts() {
    const searchInput = document.getElementById('search');

    // Global keyboard shortcuts
    document.addEventListener('keydown', (e) => {
        // Skip if any modal/overlay is open - let them handle their own keyboard navigation
        const previewModal = document.getElementById('preview-modal');
        const keyboardHelpOverlay = document.getElementById('keyboard-help-overlay');
        const isModalOpen = (previewModal && previewModal.style.display !== 'none') || keyboardHelpOverlay;
        if (isModalOpen) {
            return;
        }

        // Skip if user is typing in a text input/textarea (but not checkboxes, radio, etc.)
        const isTyping = (document.activeElement.tagName === 'INPUT' &&
                         ['text', 'search', 'password', 'email', 'tel', 'url', 'number'].includes(document.activeElement.type)) ||
                        document.activeElement.tagName === 'TEXTAREA' ||
                        document.activeElement.isContentEditable;

        const isTypingInSearch = document.activeElement === searchInput;
        const context = { searchInput, isTyping, isTypingInSearch };

        // Build key string for matching
        const keyString = getKeyString(e);

        // Check for uppercase letter shortcuts (k/c/s/t with Shift)
        // These work even when typing in search box
        const upperKey = e.key.toUpperCase();
        if (e.shiftKey && e.key.length === 1 && upperKey >= 'A' && upperKey <= 'Z') {
            const lowerKey = e.key.toLowerCase();
            const shortcut = KEYBOARD_SHORTCUTS[lowerKey];
            if (shortcut && shortcut.allowsShift) {
                if (shortcut.preventDefault) e.preventDefault();
                shortcut.action(e, context);
                return;
            }
        }

        // Match against configured shortcuts
        const shortcut = KEYBOARD_SHORTCUTS[keyString];
        if (shortcut) {
            // Check if we should skip when typing
            if (shortcut.skipIfTyping && isTyping) {
                return;
            }

            // Check if shortcut requires no modifiers (for ArrowUp/Down auto-focus)
            if (shortcut.requiresNoModifiers && (e.ctrlKey || e.altKey || e.metaKey || e.shiftKey)) {
                return;
            }

            // Check custom condition if present
            if (shortcut.condition && !shortcut.condition(e, context)) {
                return;
            }

            // Execute the shortcut
            if (shortcut.preventDefault) {
                e.preventDefault();
            }
            shortcut.action(e, context);
        }
    });

    // ESC key to clear search box
    searchInput.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            clearSearch();
        }
    });

    // Home/End/PageUp/PageDown to transfer focus to templates (like other sidebar fields)
    searchInput.addEventListener('keydown', (e) => {
        if (e.key === 'Home') {
            e.preventDefault();
            const firstCard = document.querySelector('.template-card');
            if (firstCard) {
                // Scroll to top of page to show header
                window.scrollTo({ top: 0, behavior: 'smooth' });
                firstCard.focus();
            }
        } else if (e.key === 'End') {
            e.preventDefault();
            const cards = document.querySelectorAll('.template-card');
            if (cards.length > 0) {
                const lastCard = cards[cards.length - 1];
                // Scroll to bottom of page to show footer
                window.scrollTo({ top: document.documentElement.scrollHeight, behavior: 'smooth' });
                lastCard.focus();
            }
        } else if (e.key === 'PageUp') {
            e.preventDefault();
            const cards = Array.from(document.querySelectorAll('.template-card'));
            if (cards.length === 0) return;

            const currentCard = getFirstVisibleTemplateCard();
            if (!currentCard) return;

            const currentIndex = cards.indexOf(currentCard);
            const columnCount = getGridColumnCount();
            const currentColumn = currentIndex % columnCount;

            // If already on the top row, just scroll to show header
            if (currentIndex < columnCount) {
                window.scrollTo({ top: 0, behavior: 'smooth' });
                return;
            }

            const viewportTop = window.scrollY;
            const viewportBottom = viewportTop + window.innerHeight;

            // Find the first card in this column that's at or near the top of viewport
            let firstVisibleInColumn = null;
            for (let i = currentColumn; i < cards.length; i += columnCount) {
                const card = cards[i];
                const rect = card.getBoundingClientRect();
                const cardTop = rect.top + window.scrollY;
                const cardBottom = cardTop + rect.height;

                if (cardBottom > viewportTop && cardTop < viewportBottom) {
                    firstVisibleInColumn = card;
                    break;
                }
            }

            if (!firstVisibleInColumn) return;

            const firstVisibleIndex = cards.indexOf(firstVisibleInColumn);
            const firstVisibleRect = firstVisibleInColumn.getBoundingClientRect();
            const firstVisibleTop = firstVisibleRect.top + window.scrollY;

            const isClipped = firstVisibleTop < viewportTop + 10;

            let targetCard;
            if (isClipped && firstVisibleIndex >= columnCount) {
                targetCard = firstVisibleInColumn;
            } else {
                const rowsPerPage = getRowsPerViewport();
                const cardsToMove = columnCount * rowsPerPage;
                const targetIndex = Math.max(currentColumn, firstVisibleIndex - cardsToMove);
                targetCard = cards[targetIndex];
            }

            const targetIndex = cards.indexOf(targetCard);
            if (targetIndex < columnCount) {
                window.scrollTo({ top: 0, behavior: 'smooth' });
            } else {
                const rect = targetCard.getBoundingClientRect();
                const cardTop = rect.top + window.scrollY;
                const offset = 20; // Space for border and focus outline
                window.scrollTo({ top: Math.max(0, cardTop - offset), behavior: 'smooth' });
            }
            targetCard.focus({ preventScroll: true });
        } else if (e.key === 'PageDown') {
            e.preventDefault();
            const cards = Array.from(document.querySelectorAll('.template-card'));
            if (cards.length === 0) return;

            const currentCard = getFirstVisibleTemplateCard();
            if (!currentCard) return;

            const currentIndex = cards.indexOf(currentCard);
            const columnCount = getGridColumnCount();
            const currentColumn = currentIndex % columnCount;

            // If already on the last row, just scroll to show footer
            if (currentIndex >= cards.length - columnCount) {
                window.scrollTo({ top: document.documentElement.scrollHeight, behavior: 'smooth' });
                return;
            }

            const viewportTop = window.scrollY;
            const viewportBottom = viewportTop + window.innerHeight;

            // Find the LAST card in this column that's currently visible
            let lastVisibleInColumn = null;
            for (let i = currentColumn; i < cards.length; i += columnCount) {
                const card = cards[i];
                const rect = card.getBoundingClientRect();
                const cardTop = rect.top + window.scrollY;
                const cardBottom = cardTop + rect.height;

                if (cardBottom > viewportTop && cardTop < viewportBottom) {
                    lastVisibleInColumn = card;
                } else if (cardTop >= viewportBottom) {
                    break;
                }
            }

            if (!lastVisibleInColumn) return;

            const lastVisibleIndex = cards.indexOf(lastVisibleInColumn);
            const lastVisibleRect = lastVisibleInColumn.getBoundingClientRect();
            const lastVisibleBottom = lastVisibleRect.top + window.scrollY + lastVisibleRect.height;

            const isClipped = lastVisibleBottom > viewportBottom - 10;

            let targetCard;
            if (isClipped) {
                targetCard = lastVisibleInColumn;
            } else {
                const nextIndex = lastVisibleIndex + columnCount;
                if (nextIndex < cards.length && nextIndex % columnCount === currentColumn) {
                    targetCard = cards[nextIndex];
                } else {
                    targetCard = lastVisibleInColumn;
                }
            }

            const targetIndex = cards.indexOf(targetCard);
            if (targetIndex >= cards.length - columnCount) {
                window.scrollTo({ top: document.documentElement.scrollHeight, behavior: 'smooth' });
            } else {
                const rect = targetCard.getBoundingClientRect();
                const cardTop = rect.top + window.scrollY;
                const offset = 20; // Space for border and focus outline
                window.scrollTo({ top: Math.max(0, cardTop - offset), behavior: 'smooth' });
            }
            targetCard.focus({ preventScroll: true });
        }
    });

    // Prevent uppercase letters in search box (reserved for shortcuts)
    searchInput.addEventListener('keydown', (e) => {
        // Check if it's an uppercase letter
        if (e.key.length === 1 && e.key >= 'A' && e.key <= 'Z') {
            // Always prevent uppercase letters from being typed
            e.preventDefault();

            // If it's an assigned shortcut (K, C, S, T), the global handler will handle navigation
            const assignedShortcuts = ['K', 'C', 'S', 'T'];
            if (!assignedShortcuts.includes(e.key)) {
                // For unassigned uppercase letters, give visual feedback
                searchInput.classList.add('shake');
                setTimeout(() => searchInput.classList.remove('shake'), 300);
            }
            // Note: assigned shortcuts will trigger navigation via the global handler
        }
        // Note: '?' is handled by the global handler and works in search field
    });
}

/**
 * Show keyboard help overlay
 */
let keyboardHelpPreviousFocus = null;
let shouldRestoreFocus = true;

function showKeyboardHelp(returnFocusToSearch = false, initialTab = 'help') {
    const existingOverlay = document.getElementById('keyboard-help-overlay');
    if (existingOverlay) {
        closeKeyboardHelp(returnFocusToSearch);
        return; // Toggle off if already shown
    }

    // Store the currently focused element to restore later
    keyboardHelpPreviousFocus = document.activeElement;
    shouldRestoreFocus = true;

    const overlay = document.createElement('div');
    overlay.id = 'keyboard-help-overlay';
    overlay.className = 'keyboard-help-overlay';
    overlay.setAttribute('role', 'dialog');
    overlay.setAttribute('aria-labelledby', 'keyboard-help-title');
    overlay.setAttribute('aria-modal', 'true');

    overlay.innerHTML = `
        <div class="keyboard-help-content">
            <div class="keyboard-help-header">
                <h2 id="keyboard-help-title"><span class="title-lima">lima-</span><span class="title-catalog">catalog</span></h2>
                <button class="keyboard-help-close" tabindex="0" aria-label="Close modal">×</button>
            </div>
            <div class="keyboard-help-tabs" role="tablist" aria-label="Help sections">
                <button class="help-tab ${initialTab === 'about' ? 'active' : ''}"
                        data-tab="about"
                        role="tab"
                        aria-selected="${initialTab === 'about'}"
                        aria-controls="tab-about"
                        tabindex="${initialTab === 'about' ? '0' : '-1'}">
                    About
                </button>
                <button class="help-tab ${initialTab === 'help' ? 'active' : ''}"
                        data-tab="help"
                        role="tab"
                        aria-selected="${initialTab === 'help'}"
                        aria-controls="tab-help"
                        tabindex="${initialTab === 'help' ? '0' : '-1'}">
                    Keyboard Help
                </button>
            </div>
            <div class="keyboard-help-body" tabindex="-1">
                <div id="tab-about" class="tab-content ${initialTab === 'about' ? 'active' : ''}" role="tabpanel" aria-labelledby="tab-about-label">
                    <h3 id="tab-about-label">About Lima Catalog</h3>
                    <p>
                        Lima Catalog is a searchable directory of Lima VM templates from across GitHub.
                        It helps you discover and explore community-contributed templates for creating Linux virtual machines with Lima.
                    </p>
                    <p style="margin-top: 1rem; padding: 0.75rem; background: rgba(255, 193, 7, 0.15); border-left: 3px solid #ffc107; border-radius: 0.25rem; font-size: 0.875rem;">
                        <strong>Note:</strong> This catalog is under construction. Template descriptions are not yet being generated, limiting the usefulness of the catalog data.
                    </p>
                    <div class="warning-box" role="alert">
                        <h4>⚠️ Important Notice</h4>
                        <p>
                            <strong>Templates in this catalog have not been reviewed or verified.</strong>
                            Exercise caution before using any template. Always review the template contents
                            and source repository before running on your system.
                        </p>
                    </div>
                    <h4>How the Catalog Works</h4>
                    <p>
                        The catalog is automatically maintained through a GitHub Actions workflow that:
                    </p>
                    <ul>
                        <li>Searches GitHub for repositories containing Lima template files (*.yaml files)</li>
                        <li>Analyzes template metadata, including OS images, keywords, and categories</li>
                        <li>Updates the catalog daily with new templates and changes</li>
                        <li>Provides a web interface for browsing and searching templates</li>
                    </ul>
                    <h4>Template Notability</h4>
                    <p>
                        Templates are ranked by a "notability score" that helps surface the most interesting
                        and useful templates. The score considers several factors:
                    </p>
                    <ul>
                        <li><strong>Reusability:</strong> Templates with message text are designed for sharing</li>
                        <li><strong>Customization:</strong> Provision scripts, parameters, and environment variables</li>
                        <li><strong>Documentation:</strong> Comment quality and completeness</li>
                        <li><strong>Specialization:</strong> Use of non-standard OS images</li>
                        <li><strong>Popularity:</strong> Repository star count (with balanced weighting)</li>
                    </ul>
                    <h4>Get Involved</h4>
                    <p>
                        Found a bug or have a suggestion?
                        <a href="https://github.com/lima-catalog/lima-catalog/issues" target="_blank" rel="noopener noreferrer">
                            File an issue on GitHub
                        </a>
                    </p>
                </div>
                <div id="tab-help" class="tab-content ${initialTab === 'help' ? 'active' : ''}" role="tabpanel" aria-labelledby="tab-help-label">
                    <div class="keyboard-help-section">
                        <h3 id="tab-help-label">Jump to Section</h3>
                        <dl class="keyboard-shortcuts">
                            <dt><kbd>/</kbd></dt>
                            <dd>Search box</dd>
                            <dt><kbd>K</kbd> or <kbd>Shift+K</kbd></dt>
                            <dd>Keywords</dd>
                            <dt><kbd>C</kbd> or <kbd>Shift+C</kbd></dt>
                            <dd>Categories</dd>
                            <dt><kbd>S</kbd> or <kbd>Shift+S</kbd></dt>
                            <dd>Sort dropdown</dd>
                            <dt><kbd>T</kbd> or <kbd>Shift+T</kbd></dt>
                            <dd>Templates</dd>
                            <dt><kbd>Ctrl+↑</kbd></dt>
                            <dd>Header (theme + help)</dd>
                        </dl>
                        <p style="font-size: 0.75rem; color: var(--text-light); margin-top: 0.75rem; font-style: italic; line-height: 1.4;">
                            Tip: Uppercase (Shift+K/C/S/T) work while typing
                        </p>
                    </div>
                    <div class="keyboard-help-section">
                        <h3>Navigate & Scroll</h3>
                        <dl class="keyboard-shortcuts">
                            <dt><kbd>↑</kbd> <kbd>↓</kbd> <kbd>←</kbd> <kbd>→</kbd></dt>
                            <dd>Navigate within sections</dd>
                            <dt><kbd>Tab</kbd></dt>
                            <dd>Navigate between elements</dd>
                            <dt><kbd>Ctrl+←</kbd></dt>
                            <dd>Templates → sidebar</dd>
                            <dt><kbd>Ctrl+→</kbd></dt>
                            <dd>Sidebar → templates</dd>
                            <dt><kbd>Ctrl+↓</kbd></dt>
                            <dd>Header → templates</dd>
                            <dt><kbd>Page Up</kbd> <kbd>Page Down</kbd></dt>
                            <dd>Scroll + focus visible template</dd>
                            <dt><kbd>Home</kbd> <kbd>End</kbd></dt>
                            <dd>First / last template</dd>
                            <dt><kbd>Enter</kbd> or <kbd>Space</kbd></dt>
                            <dd>Select / activate</dd>
                            <dt><kbd>Delete</kbd> or <kbd>Backspace</kbd></dt>
                            <dd>Remove selected keyword</dd>
                            <dt><kbd>Esc</kbd></dt>
                            <dd>Clear search</dd>
                            <dt><kbd>?</kbd></dt>
                            <dd>Show/hide this help</dd>
                        </dl>
                    </div>
                    <div class="keyboard-help-section">
                        <h3>Filters & Debug</h3>
                        <dl class="keyboard-shortcuts">
                            <dt><kbd>O</kbd></dt>
                            <dd>Toggle ORG filter (focused template)</dd>
                            <dt><kbd>R</kbd></dt>
                            <dd>Toggle REPO filter (focused template)</dd>
                            <dt><kbd>@</kbd></dt>
                            <dd>Toggle debug mode</dd>
                        </dl>
                    </div>
                    <div class="keyboard-help-section">
                        <h3>Preview Modal</h3>
                        <dl class="keyboard-shortcuts">
                            <dt><kbd>Ctrl+←</kbd> <kbd>Ctrl+→</kbd></dt>
                            <dd>Previous / next template</dd>
                            <dt><kbd>Ctrl+↑</kbd> <kbd>Ctrl+↓</kbd></dt>
                            <dd>Previous / next row</dd>
                            <dt><kbd>↑</kbd> <kbd>↓</kbd> <kbd>←</kbd> <kbd>→</kbd></dt>
                            <dd>Scroll YAML content</dd>
                            <dt><kbd>@</kbd></dt>
                            <dd>Toggle YAML ↔ JSON view</dd>
                            <dt><kbd>Esc</kbd></dt>
                            <dd>Close modal</dd>
                        </dl>
                    </div>
                </div>
            </div>
        </div>
    `;

    document.body.appendChild(overlay);
    document.body.style.overflow = 'hidden'; // Lock scrolling

    // Close on click outside or close button
    const closeBtn = overlay.querySelector('.keyboard-help-close');
    const content = overlay.querySelector('.keyboard-help-content');

    closeBtn.addEventListener('click', () => closeKeyboardHelp(returnFocusToSearch));
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) closeKeyboardHelp(returnFocusToSearch);
    });

    // Tab switching functionality
    const tabs = overlay.querySelectorAll('.help-tab');
    const tabContents = overlay.querySelectorAll('.tab-content');

    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            const targetTab = tab.dataset.tab;

            // Update tabs
            tabs.forEach(t => {
                const isActive = t.dataset.tab === targetTab;
                t.classList.toggle('active', isActive);
                t.setAttribute('aria-selected', isActive);
                t.setAttribute('tabindex', isActive ? '0' : '-1');
            });

            // Update content
            tabContents.forEach(tc => {
                tc.classList.toggle('active', tc.id === `tab-${targetTab}`);
            });
        });

        // Keyboard navigation for tabs
        tab.addEventListener('keydown', (e) => {
            if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
                e.preventDefault();
                const currentIndex = Array.from(tabs).indexOf(tab);
                const nextIndex = e.key === 'ArrowRight'
                    ? (currentIndex + 1) % tabs.length
                    : (currentIndex - 1 + tabs.length) % tabs.length;
                tabs[nextIndex].click();
                tabs[nextIndex].focus();
            }
        });
    });

    // Trap focus within modal for accessibility
    trapFocus(content);

    // Handle keyboard events - ESC, ?, and shortcuts
    overlay.addEventListener('keydown', (e) => {
        // Close on Escape or ?
        if (e.key === 'Escape' || e.key === '?') {
            e.preventDefault();
            closeKeyboardHelp(returnFocusToSearch);
            return;
        }

        // Handle shortcuts - close modal and execute shortcut
        const isShortcutKey = (e.key === 'k' || e.key === 'K' ||
                              e.key === 'c' || e.key === 'C' ||
                              e.key === 's' || e.key === 'S' ||
                              e.key === 't' || e.key === 'T' ||
                              e.key === 'o' || e.key === 'O' ||
                              e.key === 'r' || e.key === 'R' ||
                              e.key === '/');

        if (isShortcutKey) {
            e.preventDefault();
            shouldRestoreFocus = false; // Don't restore focus, let the shortcut handle it
            overlay.remove();
            // Re-dispatch the event to trigger the global handler
            const newEvent = new KeyboardEvent('keydown', {
                key: e.key,
                code: e.code,
                shiftKey: e.shiftKey,
                ctrlKey: e.ctrlKey,
                altKey: e.altKey,
                metaKey: e.metaKey,
                bubbles: true
            });
            document.dispatchEvent(newEvent);
            return;
        }

        // Prevent arrow keys and other navigation keys from scrolling the page
        const navigationKeys = ['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight',
                               'PageUp', 'PageDown', 'Home', 'End', 'Space'];
        if (navigationKeys.includes(e.key)) {
            e.preventDefault();
        }
    });

    // Focus the first tab or close button for accessibility
    const activeTab = overlay.querySelector('.help-tab.active');
    if (activeTab) {
        activeTab.focus();
    } else {
        closeBtn.focus();
    }
}

function closeKeyboardHelp(returnFocusToSearch = false) {
    const overlay = document.getElementById('keyboard-help-overlay');
    if (overlay) {
        overlay.remove();
        document.body.style.overflow = 'auto'; // Unlock scrolling

        // Restore focus
        if (shouldRestoreFocus) {
            if (returnFocusToSearch) {
                const searchInput = document.getElementById('search');
                if (searchInput) searchInput.focus();
            } else if (keyboardHelpPreviousFocus && keyboardHelpPreviousFocus.focus) {
                keyboardHelpPreviousFocus.focus();
            }
        }

        keyboardHelpPreviousFocus = null;
        shouldRestoreFocus = true;
    }
}

export { setupKeyboardShortcuts, showKeyboardHelp };
