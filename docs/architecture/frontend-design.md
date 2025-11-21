# Frontend Design

**Quick Links**: [Overview](overview.md) | [Backend Design](backend-design.md) | [UI/UX Guidelines](../guides/ui-ux-guidelines.md)

---

## Overview

The frontend is a static GitHub Pages site that provides an interactive catalog browser with advanced filtering, keyboard navigation, and template preview.

**Live Site**: [lima-catalog.github.io/lima-catalog](https://lima-catalog.github.io/lima-catalog/)

**Tech Stack:**
- Vanilla JavaScript (ES6 modules)
- Highlight.js for syntax highlighting
- CSS Grid + Flexbox
- No framework dependencies
- No build step required

---

## Architecture

### Three-Layer Module Organization

The frontend uses a clean separation of concerns:

1. **`app.js`** - Application initialization and event setup
2. **`appActions.js`** - Shared actions to break circular dependencies (filter, render, UI updates)
3. **`keyboard.js`** - Keyboard shortcuts and navigation (990 lines, extracted for maintainability)

This architecture prevents circular dependencies while maintaining clean code organization.

### Module Structure

```
web/js/
├── app.js              # Entry point, initialization
├── appActions.js       # Core actions (filter, render, update)
├── keyboard.js         # Keyboard navigation (extracted for size)
├── config.js           # Configuration constants
├── state.js            # Application state management
├── data.js             # Fetch and parse catalog data
├── filters.js          # Filter and sort logic
├── sidebar.js          # Sidebar rendering
├── templateCard.js     # Template card rendering
├── modal.js            # Preview and help modals
├── theme.js            # Dark/light theme
├── urlHelpers.js       # Lima 2.0 URL generation
└── utils.js            # Utility functions
```

→ For complete file descriptions, see [Source Index](source-index.md)

---

## Key Features

### Multi-Keyword Filtering

**AND Logic**: All selected keywords must match

**Implementation**:
- Keywords extracted from templates during backend analysis
- Dynamic counts update as filters change
- Clear visual feedback (selected = blue)
- Click keyword badge to add/remove from filter

**Code**: `filters.js:filterTemplates()`

### Dynamic ORG/REPO Keyword Filters

**Purpose**: Filter by organization or repository when viewing a specific template

**Behavior**:
- When focusing on a template card (keyboard navigation), contextual keyword filters appear
- If org has multiple repos: Shows org keyword + repo keywords for each repo
- If org has one repo with multiple templates: Shows repo keyword only
- Visual distinction: Teal/green gradient colors vs. blue for regular keywords
- Filtering: Keywords with `/` filter by `template.repo`, without `/` filter by `template.org`
- **Keyboard-only feature**: Hover caused issues; keywords persist while navigating sidebar

**Code**: `sidebar.js:updateDynamicKeywords()`

### Category Browsing

**Categories**:
- containers
- kubernetes
- development
- databases
- security
- machine-learning
- networking
- gaming
- other

**Dynamic counts** update based on current filter state

**Code**: `sidebar.js:renderCategories()`

### Template Preview Modal

**Features**:
- Full YAML content with syntax highlighting (highlight.js)
- Template metadata (org, repo, stars, last updated)
- Keywords and category badges
- Lima 2.0 `github:` URL with copy button
- Similar templates section (if duplicates detected)
- Full keyboard navigation (Escape to close, Tab for focus)

**Accessibility**:
- `role="dialog"`, `aria-modal="true"`
- Focus trap within modal
- Escape key to close
- ARIA labels on all interactive elements

**Code**: `modal.js:showTemplatePreview()`

### Similar Templates Detection

**Display**:
- Color-coded badges: Exact (red), Near (orange), Similar (blue)
- Similarity percentage shown
- Click to navigate between similar templates
- Hidden when no similar templates exist

**Implementation**:
- Backend generates MinHash signatures
- LSH finds similar templates (50%+ similarity)
- Frontend displays in "Similar Templates" section of modal

**Code**: `modal.js:createSimilarTemplatesSection()`

### Help Modal

**Tabbed Interface**:
- **About**: Catalog stats, update info, GitHub links
- **Shortcuts**: Complete keyboard shortcut reference

**Code**: `modal.js:showHelpModal()`

### Lima 2.0 URL Generation

Converts GitHub URLs to Lima-compatible `github:` format:

```
https://github.com/owner/repo/blob/sha/path/template.yaml
→
github://owner/repo/path/template.yaml?ref=sha
```

**Features**:
- Automatic conversion in preview modal
- Copy to clipboard button
- Handles edge cases (default branches, tags, commits)

**Code**: `urlHelpers.js:generateLimaUrl()`

### URL Deep Linking

**Purpose**: Share direct links to specific templates with or without modal open

**URL Formats**:
- `?template=<id>` - Focus on template, modal closed
- `?template=<id>&modal=open` - Open template modal overlay

**Features**:
- URL updates automatically during keyboard navigation
- Pasting URL focuses and scrolls to template
- Browser back/forward buttons work correctly
- Template stays selected when modal closes
- Retry logic handles DOM timing issues

**Behavior**:
- **Keyboard Navigation**: URL updates as you navigate with arrow keys
- **Opening Modal**: Adds `&modal=open` to URL
- **Closing Modal**: Removes `&modal=open`, keeps template selected
- **Browser Back**: From open modal → closes modal, keeps template selected
- **Browser Back**: From selected template → clears selection
- **Page Load**: Parses URL and restores state (focus + modal if specified)

**Implementation**:
- `updateURLForTemplateSelection()` - Called on template card focus events
- `focusTemplateCard()` - Focus and scroll with retry logic (up to 5 attempts)
- `closeModalKeepTemplate()` - Remove modal parameter but keep template
- `handlePopState()` - Sync state with browser history
- `data-template-id` attribute on cards for DOM queries

**Code**: `modal.js:updateURLForTemplateSelection()`, `modal.js:handlePopState()`, `templateCard.js` (focus handler)

### Theme Switching

**Modes**:
- Light mode
- Dark mode
- Remembers preference in localStorage

**Implementation**:
- CSS custom properties for colors
- Smooth transitions
- Keyboard shortcut: `T`

**Code**: `theme.js`, `style.css`

---

## Keyboard Navigation

**990 lines** of keyboard navigation code extracted into `keyboard.js` for maintainability.

### Navigation Shortcuts

| Key | Action |
|-----|--------|
| `J` / `↓` | Next template |
| `K` / `↑` | Previous template |
| `Enter` | Open template preview |
| `Escape` | Close modal / clear filters |
| `Tab` | Cycle focus areas |
| `1-9` | Jump to category |
| `T` | Toggle theme |
| `S` | Focus search |
| `?` | Show help |

### Focus Management

**Three focus areas**:
1. **Sidebar** - Keywords and categories
2. **Search** - Search input field
3. **Results** - Template cards

**Visual feedback**:
- Orange outline on focused template card
- Blue highlight on sidebar items
- Focus ring on search input

**Code**: `keyboard.js`

### Accessibility

All keyboard features include:
- ARIA labels (`aria-label`, `aria-labelledby`)
- Semantic roles (`role="button"`, `role="dialog"`)
- Focus indicators (visible outlines)
- Screen reader announcements (`aria-live` regions)

→ For complete accessibility guidelines, see [UI/UX Guidelines](../guides/ui-ux-guidelines.md)

---

## Data Loading

### Single Request

Fetches `catalog.jsonl` from data branch:

```javascript
const response = await fetch(DATA_URL);
const text = await response.text();
const lines = text.trim().split('\n');
const templates = lines.map(line => JSON.parse(line));
```

**Benefits**:
- One HTTP request (vs. three for templates/repos/orgs)
- Faster page load
- Simpler code (no joining required)

**Code**: `data.js:loadCatalog()`

### Cache Management

**Browser caching**:
- GitHub Pages: 10-minute cache
- May need hard refresh after updates (Cmd+Shift+R / Ctrl+Shift+R)

---

## Testing

### Test Framework

Custom lightweight framework (`test-framework.js`):

```javascript
import { runner, assert } from './test-framework.js';

runner.test('filters: AND logic', () => {
    const result = filterTemplates(templates, ['docker', 'ubuntu']);
    assert.equal(result.length, 5);
});
```

### Test Coverage

**76 frontend tests** covering:
- URL helpers (`urlHelpers.test.js`)
- Data parsing (`data.test.js`)
- Filters and sorting (`filters.test.js`)
- Template card rendering (`templateCard.test.js`)
- Theme management (`theme.test.js`)

**Running tests**:
```bash
# Node.js (with DOM mocking)
npm test

# Browser (visual)
# Open web/tests.html in browser
```

---

## Performance

### Efficient Rendering

**Virtual scrolling**: Not needed yet (716 templates render instantly)

**Debounced search**: 300ms debounce on search input

**Lazy loading**: Template preview YAML only loaded when modal opens

### Bundle Size

**Zero dependencies** for core functionality:
- No React, Vue, Angular
- No jQuery
- Only highlight.js for syntax highlighting (~50KB)

**Total JS**: ~15KB minified (before highlight.js)

---

## Related Documentation

- **[Overview](overview.md)** - System architecture
- **[Backend Design](backend-design.md)** - Data generation
- **[UI/UX Guidelines](../guides/ui-ux-guidelines.md)** - Complete design system
- **[Source Index](source-index.md)** - Find any file

**Historical:**
- **[UI Redesign History](../history/ui-redesign/)** - Implementation timeline
