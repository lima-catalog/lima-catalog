# macOS Scrollbar Visibility in WebKit Browsers

## Problem

On macOS, Safari and Chrome use **overlay scrollbars** by default, which automatically hide when the user is not actively scrolling. This creates a poor user experience for UI elements where the scrollbar provides important visual feedback about:
- Whether content is scrollable
- How much content is available
- Current scroll position

## The Challenge

CSS cannot fully override the macOS system-level overlay scrollbar behavior. The operating system preference "Show scroll bars: Automatically based on mouse or trackpad" takes precedence over most web styling attempts.

## Solutions That Don't Work

### ❌ Using Standard CSS Properties Alone

```css
/* These don't force visibility on macOS Safari/Chrome */
.scrollable {
    overflow-y: scroll;
    scrollbar-color: #999 #f0f0f0;  /* Firefox only */
    scrollbar-width: thin;           /* Firefox only */
    scrollbar-gutter: stable;        /* No effect with overlay scrollbars */
}
```

**Why this fails:**
- `scrollbar-color` and `scrollbar-width` are Firefox-only (Safari doesn't support them)
- `scrollbar-gutter` has zero effect when browsers use overlay scrollbars
- `overflow-y: scroll` alone doesn't prevent auto-hiding on macOS

### ❌ Using `::-webkit-scrollbar` WITH Standard Properties

```css
/* This won't work! */
.scrollable {
    scrollbar-color: #999 #f0f0f0;  /* Blocks WebKit styling! */
    scrollbar-width: thin;           /* Blocks WebKit styling! */
}

.scrollable::-webkit-scrollbar {
    width: 14px;  /* Won't apply! */
}
```

**Critical Issue (Safari 18.2+):**
> **"If you've already implemented `scrollbar-color` or `scrollbar-width` properties, setting `::-webkit-scrollbar` properties won't work!"**

The standard CSS properties take precedence over WebKit pseudo-elements, preventing custom styling from being applied.

## ✅ Solution That Works

### Step 1: Remove Conflicting Properties

**Do NOT use these properties** if you want to style scrollbars with `::-webkit-scrollbar`:
- ❌ `scrollbar-color`
- ❌ `scrollbar-width`
- ❌ `scrollbar-gutter`

### Step 2: Apply Complete WebKit Scrollbar Styling

The key is setting `background-color` directly on `::-webkit-scrollbar`:

```css
.scrollable {
    overflow-y: scroll;
}

/* Force non-overlay scrollbar on macOS Safari/Chrome */
.scrollable::-webkit-scrollbar {
    -webkit-appearance: none;           /* Disable default appearance */
    background-color: #f0f0f0;          /* KEY: Forces always-visible scrollbar */
    width: 14px;
}

.scrollable::-webkit-scrollbar-track {
    background-color: #f0f0f0;
    border-left: 1px solid #e0e0e0;     /* Visual definition */
}

.scrollable::-webkit-scrollbar-thumb {
    background-color: #999;
    background-clip: padding-box;        /* Better rendering */
    border-radius: 7px;
    border: 3px solid #f0f0f0;          /* Padding around thumb */
    min-height: 40px;
}

.scrollable::-webkit-scrollbar-thumb:hover {
    background-color: #666;
}

.scrollable::-webkit-scrollbar-button {
    display: none;
}

.scrollable::-webkit-scrollbar-corner {
    background: transparent;
}
```

### Step 3: Conditionally Hide When Not Needed

If your element has a fixed height that shows N items, hide the scrollbar when there are N or fewer items:

```javascript
// Example: Max height shows 4 items
if (items.length <= 4) {
    element.classList.add('no-scroll');
} else {
    element.classList.remove('no-scroll');
}
```

```css
.scrollable.no-scroll {
    overflow-y: hidden;
}
```

## Implementation in lima-catalog

### Files Modified

**web/style.css:**
```css
.similar-templates-list {
    overflow-y: scroll;
    /* No scrollbar-color or scrollbar-width! */
}

.similar-templates-list::-webkit-scrollbar {
    -webkit-appearance: none;
    background-color: var(--scrollbar-track);
    width: 14px;
}

/* Full styling as shown above... */

.similar-templates-list.no-scroll {
    overflow-y: hidden;
}
```

**web/js/modal.js:**
```javascript
// Hide scrollbar if 4 or fewer items (max-height shows 4)
if (sortedSimilarTemplates.length <= 4) {
    similarList.classList.add('no-scroll');
} else {
    similarList.classList.remove('no-scroll');
}
```

## Browser Support

| Browser | Support | Notes |
|---------|---------|-------|
| Safari (macOS) | ✅ Yes | Requires removing `scrollbar-color`/`scrollbar-width` |
| Chrome (macOS) | ✅ Yes | Requires removing `scrollbar-color`/`scrollbar-width` |
| Edge (macOS) | ✅ Yes | Chromium-based, same as Chrome |
| Firefox (macOS) | ⚠️ Partial | Use `scrollbar-color`/`scrollbar-width` instead |
| Safari (iOS) | ❌ No | `::-webkit-scrollbar` not supported |

## Key Takeaways

1. **Setting `background-color` on `::-webkit-scrollbar`** is what forces the scrollbar to be always visible (converts overlay to classic)
2. **Remove `scrollbar-color` and `scrollbar-width`** - they block WebKit pseudo-elements in WebKit browsers
3. **Use `-webkit-appearance: none`** to disable default overlay behavior
4. **Conditionally hide** the scrollbar when there's no overflow to avoid wasted space

## References

- [WebKit Features in Safari 18.2](https://webkit.org/blog/16301/webkit-features-in-safari-18-2/)
- [Force scrollbar visibility on macOS (Gist)](https://gist.github.com/IceCreamYou/cd517596e5847a88e2bb0a091da43fb4)
- [::-webkit-scrollbar - MDN](https://developer.mozilla.org/en-US/docs/Web/CSS/::-webkit-scrollbar)
- [CSS Overflow Scroll on macOS](https://stackoverflow.com/questions/7492062/css-overflow-scroll-always-show-vertical-scroll-bar)

## Last Updated

2025-11-23
