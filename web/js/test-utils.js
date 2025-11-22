/**
 * Testing utilities - Mocks and helpers for tests
 */

/**
 * Timer mocking utilities for testing debounce, setTimeout, etc.
 */
export class TimerMock {
    constructor() {
        this.timers = [];
        this.currentTime = 0;
        this.originalSetTimeout = null;
        this.originalClearTimeout = null;
    }

    /**
     * Install timer mocks
     */
    install() {
        this.originalSetTimeout = global.setTimeout;
        this.originalClearTimeout = global.clearTimeout;

        global.setTimeout = (callback, delay) => {
            const id = this.timers.length;
            this.timers.push({
                id,
                callback,
                delay,
                time: this.currentTime + delay,
                cancelled: false
            });
            return id;
        };

        global.clearTimeout = (id) => {
            if (this.timers[id]) {
                this.timers[id].cancelled = true;
            }
        };
    }

    /**
     * Uninstall timer mocks
     */
    uninstall() {
        if (this.originalSetTimeout) {
            global.setTimeout = this.originalSetTimeout;
        }
        if (this.originalClearTimeout) {
            global.clearTimeout = this.originalClearTimeout;
        }
        this.reset();
    }

    /**
     * Advance time by specified milliseconds
     * @param {number} ms - Milliseconds to advance
     */
    tick(ms) {
        this.currentTime += ms;

        // Execute all timers that should have fired
        const toExecute = this.timers.filter(
            timer => !timer.cancelled && timer.time <= this.currentTime
        );

        toExecute.forEach(timer => {
            timer.cancelled = true; // Mark as executed
            timer.callback();
        });
    }

    /**
     * Run all pending timers
     */
    runAll() {
        const maxTime = Math.max(...this.timers.map(t => t.time), this.currentTime);
        this.tick(maxTime - this.currentTime + 1);
    }

    /**
     * Reset timer state
     */
    reset() {
        this.timers = [];
        this.currentTime = 0;
    }

    /**
     * Get count of pending timers
     */
    getPendingCount() {
        return this.timers.filter(t => !t.cancelled && t.time > this.currentTime).length;
    }
}

/**
 * Create a mock DOM element for testing
 * @param {string} tag - Element tag name
 * @param {Object} attributes - Element attributes
 * @returns {Object} Mock element
 */
export function createMockElement(tag = 'div', attributes = {}) {
    const element = {
        tagName: tag.toUpperCase(),
        attributes: { ...attributes },
        children: [],
        style: {},
        dataset: {},
        classList: {
            classes: [],
            add(...classes) {
                this.classes.push(...classes);
            },
            remove(...classes) {
                this.classes = this.classes.filter(c => !classes.includes(c));
            },
            contains(className) {
                return this.classes.includes(className);
            }
        },
        eventListeners: {},

        addEventListener(event, handler) {
            if (!this.eventListeners[event]) {
                this.eventListeners[event] = [];
            }
            this.eventListeners[event].push(handler);
        },

        removeEventListener(event, handler) {
            if (this.eventListeners[event]) {
                this.eventListeners[event] = this.eventListeners[event]
                    .filter(h => h !== handler);
            }
        },

        querySelectorAll(selector) {
            return this.children.filter(child =>
                child.matches && child.matches(selector)
            );
        },

        focus() {
            this.focused = true;
            if (global.document) {
                global.document.activeElement = this;
            }
        },

        matches(selector) {
            // Simple selector matching for tests
            if (selector.startsWith('.')) {
                return this.classList.contains(selector.slice(1));
            }
            if (selector.startsWith('#')) {
                return this.attributes.id === selector.slice(1);
            }
            return this.tagName.toLowerCase() === selector.toLowerCase();
        }
    };

    // Set initial attributes
    Object.keys(attributes).forEach(key => {
        element.attributes[key] = attributes[key];
        if (key === 'id') element.id = attributes[key];
        if (key === 'class') {
            element.classList.classes = attributes[key].split(' ');
        }
    });

    // Mock offsetParent for visibility check (set after element is created)
    element.offsetParent = element;

    return element;
}
