/**
 * Unit Tests: Utility Functions (utils.js)
 *
 * High-level overview of what's being tested:
 * - debounce() function:
 *   - Delaying function execution until after wait period
 *   - Canceling previous calls on rapid invocations
 *   - Executing with latest arguments after settling
 *   - Handling custom and default wait times
 *   - Passing multiple arguments correctly
 *
 * - trapFocus() function:
 *   - Focusing first element on initialization
 *   - Wrapping Tab navigation at last element
 *   - Wrapping Shift+Tab at first element
 *   - Allowing normal Tab in middle of elements
 *   - Ignoring non-Tab keys
 *   - Handling single or no focusable elements
 *   - Cleanup and event listener removal
 */

import { runner, assert } from './test-framework.js';
import { debounce, trapFocus } from './utils.js';
import { TimerMock, createMockElement } from './test-utils.js';

// ============================================================================
// debounce() tests
// ============================================================================

runner.test('utils.js: debounce delays function execution', () => {
    const timerMock = new TimerMock();
    timerMock.install();

    let callCount = 0;
    const fn = () => { callCount++; };
    const debounced = debounce(fn, 300);

    // Call immediately
    debounced();
    assert.equal(callCount, 0, 'Function should not execute immediately');

    // Advance time but not enough
    timerMock.tick(200);
    assert.equal(callCount, 0, 'Function should not execute before wait time');

    // Advance past wait time
    timerMock.tick(100);
    assert.equal(callCount, 1, 'Function should execute after wait time');

    timerMock.uninstall();
});

runner.test('utils.js: debounce cancels previous call on rapid invocations', () => {
    const timerMock = new TimerMock();
    timerMock.install();

    let callCount = 0;
    const fn = () => { callCount++; };
    const debounced = debounce(fn, 300);

    // Call multiple times rapidly
    debounced();
    timerMock.tick(100);
    debounced();
    timerMock.tick(100);
    debounced();

    assert.equal(callCount, 0, 'Function should not execute during rapid calls');

    // Wait for final debounce to complete
    timerMock.tick(300);
    assert.equal(callCount, 1, 'Function should execute only once after settling');

    timerMock.uninstall();
});

runner.test('utils.js: debounce executes with latest arguments', () => {
    const timerMock = new TimerMock();
    timerMock.install();

    let lastArg = null;
    const fn = (arg) => { lastArg = arg; };
    const debounced = debounce(fn, 300);

    debounced('first');
    timerMock.tick(100);
    debounced('second');
    timerMock.tick(100);
    debounced('third');
    timerMock.tick(300);

    assert.equal(lastArg, 'third', 'Should execute with most recent argument');

    timerMock.uninstall();
});

runner.test('utils.js: debounce handles multiple sequential calls', () => {
    const timerMock = new TimerMock();
    timerMock.install();

    let callCount = 0;
    const fn = () => { callCount++; };
    const debounced = debounce(fn, 200);

    // First call
    debounced();
    timerMock.tick(200);
    assert.equal(callCount, 1, 'First call should execute');

    // Second call (after first completes)
    debounced();
    timerMock.tick(200);
    assert.equal(callCount, 2, 'Second call should execute');

    timerMock.uninstall();
});

runner.test('utils.js: debounce uses default wait time', () => {
    const timerMock = new TimerMock();
    timerMock.install();

    let callCount = 0;
    const fn = () => { callCount++; };
    const debounced = debounce(fn); // No wait specified, should default to 300

    debounced();
    timerMock.tick(299);
    assert.equal(callCount, 0, 'Should not execute before default 300ms');

    timerMock.tick(1);
    assert.equal(callCount, 1, 'Should execute after default 300ms');

    timerMock.uninstall();
});

runner.test('utils.js: debounce handles different wait times', () => {
    const timerMock = new TimerMock();
    timerMock.install();

    let callCount = 0;
    const fn = () => { callCount++; };
    const debounced = debounce(fn, 500);

    debounced();
    timerMock.tick(300);
    assert.equal(callCount, 0, 'Should wait for custom delay');

    timerMock.tick(200);
    assert.equal(callCount, 1, 'Should execute after custom delay');

    timerMock.uninstall();
});

runner.test('utils.js: debounce passes multiple arguments', () => {
    const timerMock = new TimerMock();
    timerMock.install();

    let result = null;
    const fn = (a, b, c) => { result = `${a}-${b}-${c}`; };
    const debounced = debounce(fn, 100);

    debounced('foo', 'bar', 'baz');
    timerMock.tick(100);

    assert.equal(result, 'foo-bar-baz', 'Should pass all arguments');

    timerMock.uninstall();
});

// ============================================================================
// trapFocus() tests
// ============================================================================

runner.test('utils.js: trapFocus focuses first element on initialization', () => {
    const button1 = createMockElement('button');
    const button2 = createMockElement('button');
    const container = createMockElement('div');
    container.children = [button1, button2];
    container.querySelectorAll = () => [button1, button2];

    // Save and set activeElement
    const originalActiveElement = global.document.activeElement;
    global.document.activeElement = null;

    const cleanup = trapFocus(container);

    assert.ok(button1.focused, 'Should focus first element on init');

    cleanup();
    global.document.activeElement = originalActiveElement;
});

runner.test('utils.js: trapFocus wraps Tab at last element', () => {
    const button1 = createMockElement('button');
    const button2 = createMockElement('button');
    const button3 = createMockElement('button');
    const container = createMockElement('div');

    container.children = [button1, button2, button3];
    container.querySelectorAll = () => [button1, button2, button3];

    const originalActiveElement = global.document.activeElement;

    // trapFocus will focus first element on initialization
    const cleanup = trapFocus(container);

    // Now simulate that user has navigated to last element
    button3.focus(); // This sets global.document.activeElement = button3

    // Simulate Tab at last element
    const event = {
        key: 'Tab',
        shiftKey: false,
        target: button3,
        preventDefault: () => { event.defaultPrevented = true; }
    };

    // Trigger the event handler
    const tabHandler = container.eventListeners['keydown'][0];
    tabHandler(event);

    assert.ok(event.defaultPrevented, 'Should prevent default');
    assert.ok(button1.focused, 'Should wrap to first element');

    cleanup();
    global.document.activeElement = originalActiveElement;
});

runner.test('utils.js: trapFocus wraps Shift+Tab at first element', () => {
    const button1 = createMockElement('button');
    const button2 = createMockElement('button');
    const button3 = createMockElement('button');
    const container = createMockElement('div');

    container.children = [button1, button2, button3];
    container.querySelectorAll = () => [button1, button2, button3];

    const originalActiveElement = global.document.activeElement;
    global.document.activeElement = button1;

    const cleanup = trapFocus(container);

    // Simulate Shift+Tab at first element
    const event = {
        key: 'Tab',
        shiftKey: true,
        target: button1,
        preventDefault: () => { event.defaultPrevented = true; }
    };

    const tabHandler = container.eventListeners['keydown'][0];
    tabHandler(event);

    assert.ok(event.defaultPrevented, 'Should prevent default');
    assert.ok(button3.focused, 'Should wrap to last element');

    cleanup();
    global.document.activeElement = originalActiveElement;
});

runner.test('utils.js: trapFocus allows Tab in middle of elements', () => {
    const button1 = createMockElement('button');
    const button2 = createMockElement('button');
    const button3 = createMockElement('button');
    const container = createMockElement('div');

    container.children = [button1, button2, button3];
    container.querySelectorAll = () => [button1, button2, button3];

    const originalActiveElement = global.document.activeElement;
    global.document.activeElement = button2;

    const cleanup = trapFocus(container);

    // Simulate Tab at middle element (should not prevent default)
    const event = {
        key: 'Tab',
        shiftKey: false,
        target: button2,
        preventDefault: () => { event.defaultPrevented = true; }
    };

    const tabHandler = container.eventListeners['keydown'][0];
    tabHandler(event);

    assert.ok(!event.defaultPrevented, 'Should not prevent default in middle');

    cleanup();
    global.document.activeElement = originalActiveElement;
});

runner.test('utils.js: trapFocus ignores non-Tab keys', () => {
    const button1 = createMockElement('button');
    const container = createMockElement('div');

    container.children = [button1];
    container.querySelectorAll = () => [button1];

    const originalActiveElement = global.document.activeElement;
    global.document.activeElement = button1;

    const cleanup = trapFocus(container);

    // Simulate Enter key
    const event = {
        key: 'Enter',
        shiftKey: false,
        preventDefault: () => { event.defaultPrevented = true; }
    };

    const tabHandler = container.eventListeners['keydown'][0];
    tabHandler(event);

    assert.ok(!event.defaultPrevented, 'Should ignore non-Tab keys');

    cleanup();
    global.document.activeElement = originalActiveElement;
});

runner.test('utils.js: trapFocus handles single focusable element', () => {
    const button = createMockElement('button');
    const container = createMockElement('div');

    container.children = [button];
    container.querySelectorAll = () => [button];

    const originalActiveElement = global.document.activeElement;
    global.document.activeElement = button;

    const cleanup = trapFocus(container);

    // Simulate Tab on single element
    const event = {
        key: 'Tab',
        shiftKey: false,
        target: button,
        preventDefault: () => { event.defaultPrevented = true; }
    };

    const tabHandler = container.eventListeners['keydown'][0];
    tabHandler(event);

    assert.ok(event.defaultPrevented, 'Should prevent default');
    assert.ok(button.focused, 'Should wrap to itself');

    cleanup();
    global.document.activeElement = originalActiveElement;
});

runner.test('utils.js: trapFocus handles no focusable elements', () => {
    const container = createMockElement('div');
    container.children = [];
    container.querySelectorAll = () => [];

    const originalActiveElement = global.document.activeElement;
    global.document.activeElement = null;

    const cleanup = trapFocus(container);

    // Should not crash with no focusable elements
    const event = {
        key: 'Tab',
        shiftKey: false,
        preventDefault: () => { event.defaultPrevented = true; }
    };

    const tabHandler = container.eventListeners['keydown'][0];
    tabHandler(event);

    assert.ok(!event.defaultPrevented, 'Should not prevent default with no elements');

    cleanup();
    global.document.activeElement = originalActiveElement;
});

runner.test('utils.js: trapFocus cleanup removes event listener', () => {
    const button = createMockElement('button');
    const container = createMockElement('div');
    container.children = [button];
    container.querySelectorAll = () => [button];

    const originalActiveElement = global.document.activeElement;
    global.document.activeElement = null;

    const cleanup = trapFocus(container);
    const initialListenerCount = container.eventListeners['keydown']?.length || 0;

    cleanup();

    const afterCleanupCount = container.eventListeners['keydown']?.filter(
        h => h !== undefined
    ).length || 0;

    assert.equal(afterCleanupCount, 0, 'Should remove event listener on cleanup');

    global.document.activeElement = originalActiveElement;
});
