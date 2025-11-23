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
3. **`keyboard.js`** - Keyboard shortcuts and navigation (extracted for maintainability)

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
- development
- general
- orchestration
- security
- testing

**Dynamic counts** update based on current filter state

**Code**: `sidebar.js:renderCategories()`

### Debug Mode Features

**Activation**: Press `@` to toggle debug mode globally

**Notability Score Breakdown Popup**:
- Appears when hovering over template badge in debug mode
- Shows detailed breakdown of notability score components:
  - Message, Provision, Parameters, Env Vars, Probes
  - Image Name, Comments, Stars
  - Total score
- Each metric displays score value and rank in catalog
- Ranks handle ties (templates with same score get same rank)
- Ranks calculated against full catalog (not filtered templates)
- Grid layout with three columns: Label | Value | Rank

**Debug Sort Options**:
When debug mode is enabled, sort dropdown includes breakdown components:
- [Debug] Message
- [Debug] Provision Scripts
- [Debug] Parameters
- [Debug] Environment Variables
- [Debug] Probes
- [Debug] Image Name
- [Debug] YAML Comments
- [Debug] Repository Stars

**Badge Behavior in Debug Mode**:
- Template cards show notability score instead of "Official"/"Community"
- When sorting by specific breakdown component, badge shows that component's score
- Badge tooltip indicates which score component is displayed
- Click template to open modal and see full breakdown

**Code**: `templateCard.js:createDebugScorePopup()`, `templateCard.js:calculateRank()`, `appActions.js:updateSortDropdown()`

### Template Preview Modal

**Features**:
- Full YAML content with syntax highlighting (highlight.js)
- Template metadata (org, repo, stars, last updated)
- Keywords and category badges
- Lima 2.0 `github:` URL with copy button
- Similar templates section (if duplicates detected)
- Full keyboard navigation (Escape to close, Tab for focus)
- Ctrl+Arrow navigation to adjacent templates without closing modal

**Accessibility**:
- `role="dialog"`, `aria-modal="true"`
- Focus trap within modal
- Escape key to close
- ARIA labels on all interactive elements

**Debug Mode** (`@` key):
- Toggle between YAML and JSON view of internal template object
- Displays pretty-printed JSON with syntax highlighting
- Persists when navigating between templates with Ctrl+Arrow
- Resets when opening new modal (not through Ctrl+Arrow navigation)
- Disabled during diff view (similar templates comparison)

**Code**: `modal.js:showTemplatePreview()`, `modal.js:toggleDebugMode()`

### Similar Templates Detection

**Display**:
- Compact single-line format: `ORG/REPO/TEMPLATEPATH` with monospace font
- Diff stats (+N -N) shown inline with green/red coloring
- Color-coded badges (fixed width for alignment): Original (green), Exact (red), Near (orange), Similar (blue)
- Badge labels derived from similarity percentage: 100% = Exact/Original, 90-99% = Near, <90% = Similar
- Similarity percentage shown
- Sorted by similarity (highest first), originals first within same similarity, then alphabetically
- Scrollable listbox showing up to 4 items
- Filtered by current filters (duplicates checkbox, keywords, search, type, category)
- Hidden when no similar templates match current filters

**Keyboard Navigation**:
- Tab into listbox to focus and select first item
- Arrow Up/Down to navigate between similar templates
- Enter/Space to open the selected similar template
- Tab out to return to YAML view and focus copy button

**Diff View**:
- When listbox is focused, YAML display shows unified diff
- Compares current template with selected similar template
- Syntax highlighted (green additions, red deletions)
- Copy button hidden while showing diff
- Original YAML restored when listbox loses focus

**Implementation**:
- Backend generates MinHash signatures
- LSH finds similar templates (50%+ similarity)
- Frontend displays in "Similar Templates" section of modal
- LCS-based diff algorithm with 3 lines of context

**Code**: `modal.js:populateSimilarTemplates()`, `modal.js:generateUnifiedDiff()`

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

**Purpose**: Share direct links to specific templates with full filter state preserved

**URL Parameters**:
- `template=<id>` - Focus on template (required for template selection)
- `modal=open` - Open template modal overlay
- `search=<term>` - Search term
- `keywords=<k1>,<k2>` - Comma-separated keywords (URL-encoded)
- `category=<name>` - Selected category
- `official=<bool>` - Show official templates (default: true)
- `community=<bool>` - Show community templates (default: true)
- `duplicates=true` - Show duplicates (only included when true)
- `sort=<field>` - Sort order (only included when not "name")

**Example URLs**:
- `?template=lima-vm/lima/templates/alpine.yaml` - Focus template
- `?template=...&modal=open` - Open modal
- `?search=alpine&keywords=docker%2Clinux&category=containers&sort=stars` - Filters only
- `?search=alpine&template=...&modal=open` - Full state

**Features**:
- URL updates automatically during keyboard navigation
- Filter changes update URL via `replaceState` (no history pollution)
- Pasting URL focuses and scrolls to template
- Browser back/forward buttons work correctly
- Template stays selected when modal closes
- Retry logic handles DOM timing issues

**Behavior**:
- **Filter Changes**: URL updated via `replaceState` to avoid history spam
- **Keyboard Navigation**: URL updates as you navigate with arrow keys
- **Opening Modal**: Adds `&modal=open` to URL
- **Closing Modal**: Removes `&modal=open`, keeps template selected
- **Browser Back**: From open modal → closes modal, keeps template selected
- **Browser Back**: From selected template → clears selection
- **Page Load**: Parses URL and restores full state (filters + focus + modal)

**Implementation**:
- `getFiltersFromURL()` - Parse filter state from URL parameters
- `updateURLWithFilters()` - Sync current filter state to URL
- `applyFiltersFromURL()` - Apply URL filter state to UI and app state
- `updateURLForTemplateSelection()` - Called on template card focus events
- `focusTemplateCard()` - Focus and scroll with retry logic (up to 5 attempts)
- `closeModalKeepTemplate()` - Remove modal parameter but keep template
- `handlePopState()` - Sync state with browser history (filters + templates)
- `data-template-id` attribute on cards for DOM queries

**Code**: `modal.js` (URL functions), `appActions.js` (filter sync), `templateCard.js` (focus handler)

### Theme Switching

**Modes**:
- Light mode
- Dark mode
- Auto mode (matches system preference)
- Remembers preference in localStorage

**Implementation**:
- CSS custom properties for colors
- Smooth transitions
- Theme switcher button in header (no keyboard shortcut)

**Code**: `theme.js`, `style.css`

---

## Keyboard Navigation

Keyboard navigation code extracted into `keyboard.js` for maintainability.

### Jump to Section Shortcuts

| Key | Action |
|-----|--------|
| `/` | Focus search box |
| `k` or `K` | Focus first keyword |
| `c` or `C` | Focus first category |
| `s` or `S` | Focus sort dropdown |
| `t` or `T` | Focus first template card |
| `Ctrl+↑` | Focus header (theme switcher) |

**Note**: Uppercase shortcuts (`K`, `C`, `S`, `T`) work even while typing in the search box.

### Template Grid Navigation

| Key | Action |
|-----|--------|
| `↑` / `↓` / `←` / `→` | Navigate between template cards (grid-aware) |
| `Enter` / `Space` | Open template preview modal |
| `PageUp` / `PageDown` | Scroll one page + focus visible template |
| `Home` | Focus first template + scroll to top |
| `End` | Focus last template + scroll to bottom |

**Auto-focus behavior**: Arrow keys (`↑` / `↓`) automatically focus the first visible template when scrolling without focus.

### Section Navigation

| Key | Action |
|-----|--------|
| `Ctrl+←` | Templates → Search box (sidebar) |
| `Ctrl+→` | Sidebar → First template |
| `Ctrl+↓` | Header → First template |
| `Tab` | Navigate between interactive elements (native) |

### Filter & Keyword Shortcuts

| Key | Action |
|-----|--------|
| `o` | Toggle ORG filter for focused template |
| `r` | Toggle REPO filter for focused template |
| `Delete` / `Backspace` | Remove selected keyword (when focused) |
| `Escape` | Clear search box (when focused in search) |

### Debug & Help

| Key | Action |
|-----|--------|
| `@` | Toggle debug mode globally |
| `?` | Show/hide keyboard help modal |

### Modal Shortcuts (when preview modal is open)

| Key | Action |
|-----|--------|
| `Ctrl+←` / `Ctrl+→` | Previous / next template |
| `Ctrl+↑` / `Ctrl+↓` | Previous / next row |
| `↑` / `↓` / `←` / `→` | Scroll YAML content |
| `@` | Toggle debug mode (YAML ↔ JSON) |
| `Escape` | Close modal |

**Similar Templates Navigation** (within modal):
- `Tab` into similar templates list
- `↑` / `↓` to navigate, shows diff preview
- `Enter` / `Space` to open selected template
- `Tab` out to restore original YAML

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

#### Focus Trap (`utils.js:trapFocus`)

Modal dialogs use `trapFocus()` to keep Tab navigation within the modal. Implementation details:

**Selector for tabbable elements**:
```javascript
'a[href]:not([tabindex="-1"]), button:not([disabled]):not([tabindex="-1"]), ...'
```

**Key considerations**:
1. **Exclude `tabindex="-1"`**: Elements with `tabindex="-1"` are focusable via JavaScript but NOT in the Tab order. The selector must exclude them from all element types (not just `[tabindex]`).

2. **Filter hidden elements**: Use `el.offsetParent !== null` to exclude elements hidden via `display: none`. Hidden elements are still returned by `querySelectorAll` but aren't tabbable.

3. **Dynamic content**: Recompute focusable elements on each Tab press, not just at initialization. This handles:
   - Tab panels that show/hide content
   - Elements whose `tabindex` changes dynamically
   - Any DOM modifications after the trap is set up

4. **Scrollable containers**: Add `tabindex="-1"` to scrollable containers (`overflow: auto/scroll`) to prevent them from being tab stops. Browsers may make scrollable regions implicitly focusable.

**Code**: `utils.js:trapFocus()`

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

**Frontend Unit Tests** (226 tests) covering:
- URL helpers (`urlHelpers.test.js` - 7 tests)
- Data parsing (`data.test.js` - 7 tests)
- Filters and sorting (`filters.test.js` - 50 tests)
- Template card rendering (`templateCard.test.js` - 25 tests)
- Theme management (`theme.test.js` - 9 tests)
- State management (`state.test.js` - 29 tests)
- Utilities (`utils.test.js` - 15 tests)
- App actions (`appActions.test.js` - 11 tests)
- App initialization (`app.test.js` - 11 tests)
- Sidebar rendering (`sidebar.test.js` - 12 tests)
- Modal functionality (`modal.test.js` - 50 tests)

**Frontend E2E Tests** (125 tests, 100 active) using Playwright:
- Search and filtering (`search.spec.js` - 17 tests)
- Categories and keywords (`categories.spec.js` - 24 tests)
- Modal interactions (`modal.spec.js` - 22 tests)
- Keyboard navigation (`keyboard.spec.js` - 22 tests)
- Theme switching (`theme.spec.js` - 15 tests)
- Visual regression (`visual.spec.js` - 25 tests, currently skipped)

**Running tests**:
```bash
# Unit tests - Node.js (with DOM mocking)
node test.js

# Unit tests - Browser (visual)
# Open web/tests.html in browser

# E2E tests - Playwright
npm run test:e2e

# E2E tests - Headed mode (see browser)
npm run test:e2e:headed
```

**Test Documentation:**
- [Frontend Test Coverage Plan](../testing/test-coverage-plan-frontend.md)
- [E2E Testing with Playwright](../testing/debugging-with-playwright.md)
- [E2E Testing Options Analysis](../testing/e2e-integration-testing-options.md)

---

## Performance

### Efficient Rendering

**Virtual scrolling**: Not needed for current catalog size

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
