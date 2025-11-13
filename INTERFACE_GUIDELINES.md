# Lima Catalog Interface Guidelines

This document defines the interaction patterns and visual design guidelines for the Lima Template Catalog web interface. Following these guidelines ensures consistency across all UI elements.

## Button Interaction Patterns

### Icon-Only Buttons (× close, X dismiss)

**Usage**: Small utility buttons with single icons (clear search, dismiss modals, remove filters)

**Default State**:
- Background: `transparent`
- Color: `var(--text-light)` (#64748b light, #94a3b8 dark)
- No border

**Hover State**:
- Background: `rgba(59, 130, 246, 0.1)` - 10% opacity blue tint
- Color: `var(--primary)` - blue
- Border-radius: `0.25rem`
- Transition: `all 0.2s`

**Rationale**: Modern UI pattern (iOS, macOS, Material Design 3) where icon buttons are subtle by default and reveal a light background on hover. The 10% blue tint provides clear hover feedback without being overwhelming.

**Examples**:
- Search field clear button (×)
- Keywords clear button (×)
- Modal close button (X)

### Primary Action Buttons

**Usage**: Main call-to-action buttons

**Default State**:
- Background: `var(--primary)` (#2563eb light, #3b82f6 dark)
- Color: `white`
- Border: `none`
- Padding: `0.625rem 1.25rem`
- Border-radius: `0.375rem`

**Hover State**:
- Background: `var(--primary-dark)` (#1e40af light, #2563eb dark)

**Examples**:
- "Copy" button in modal
- "Open in GitHub" button

### Secondary Action Buttons

**Usage**: Alternative actions, less emphasis than primary

**Default State**:
- Background: `transparent`
- Color: `var(--text)`
- Border: `1px solid var(--border-elevated)`
- Padding: `0.625rem 1.25rem`
- Border-radius: `0.375rem`

**Hover State**:
- Background: `var(--surface)` - subtle fill
- Border-color: `var(--primary)` - blue accent
- Color: `var(--primary)` - blue text

**Examples**:
- "Close" button in modal
- "View More" secondary actions

### Tertiary/Text Buttons

**Usage**: Low-emphasis actions, often inline with text

**Default State**:
- Background: `transparent`
- Color: `var(--primary)`
- Border: `none`
- Text-decoration: `none`

**Hover State**:
- Text-decoration: `underline`
- Color: `var(--primary-dark)`

**Examples**:
- GitHub repository links
- "Clear all" text buttons

## Badge & Tag Patterns

### Status Badges (Official/Community)

**Purpose**: Indicate template source type

**Official**:
- Background: `var(--primary-light)` (#3b82f6)
- Color: `white`
- Shape: Fully rounded pill (`border-radius: 9999px`)

**Community**:
- Background: `var(--secondary)` (#64748b light, #94a3b8 dark)
- Color: `white`
- Shape: Fully rounded pill

**Rationale**: Uses theme-aware colors that adapt to dark/light mode. Primary blue for official templates emphasizes authority, secondary gray for community templates indicates user-contributed.

### Keyword Tags (in cards)

**Purpose**: Show technologies/keywords associated with template

**Style**:
- Background: `var(--surface)` (#ffffff light, #1e293b dark)
- Color: `var(--text-light)` (#64748b light, #94a3b8 dark)
- Border: `1px solid var(--border)`
- Border-radius: `0.25rem` (small rounding)
- Font-size: `0.6875rem` (11px)

**Rationale**: Subtle styling that groups related technologies without overwhelming the card design. Uses surface color to stand out from card background while remaining neutral.

### Interactive Keyword Tags (sidebar)

**Purpose**: Clickable filters in keyword cloud

**Default State**:
- Background: `var(--surface)`
- Color: `var(--text)`
- Border: `1px solid var(--border)`
- Border-radius: `0.25rem`

**Hover State**:
- Border-color: `var(--primary)`
- Color: `var(--primary)`
- Transform: `translateY(-1px)` - subtle lift
- Box-shadow: `0 2px 4px rgba(0, 0, 0, 0.1)`

**Selected State**:
- Background: `var(--primary-light)` (#3b82f6 light, #60a5fa dark)
- Border-color: `var(--primary)`
- Color: `white`

**Rationale**: Three distinct states provide clear feedback. Hover lift effect indicates interactivity. Selected state uses primary color with white text for maximum visibility.

### Category Items (sidebar)

**Purpose**: Filterable category list

**Style**: Same as Interactive Keyword Tags (above)

**Rationale**: Consistency with keyword tags - both are filters with the same interaction model.

## Color Surface Hierarchy

Understanding the three-level surface system:

### Level 1: Page Background
- Variable: `--bg`
- Light: `#f8fafc` (light gray)
- Dark: `#0f172a` (very dark blue-gray)
- **Usage**: Body background, container backgrounds

### Level 2: Surface
- Variable: `--surface`
- Light: `#ffffff` (white)
- Dark: `#1e293b` (dark blue-gray)
- **Usage**: Cards, tags, inputs, buttons, template cards
- **Elevation**: Sits above page background

### Level 3: Elevated Surface
- Variable: `--surface-elevated`
- Light: `#ffffff` (same as surface)
- Dark: `#2d3748` (lighter than surface - shows elevation)
- **Usage**: Modals, dialogs, popovers, dropdowns
- **Elevation**: Highest level, floats above everything

### Special: Code Surface
- Variable: `--surface-code`
- Light: `#f8fafc` (matches page bg)
- Dark: `#1a202c` (distinct dark tone)
- **Usage**: Code blocks, YAML displays
- **Purpose**: Visually distinct from other surfaces

## Interactive Element Feedback

All interactive elements should provide feedback through multiple channels:

### Hover Feedback (Required)
- **Color change**: Text or border changes to `var(--primary)`
- **Background**: Subtle background tint for icon buttons
- **Cursor**: `cursor: pointer`
- **Transition**: `transition: all 0.2s` for smooth feedback

### Focus Feedback (Required for Accessibility)
- **Outline**: `outline: 2px solid var(--primary)`
- **Outline-offset**: `2px` (breathing room)
- Never use `outline: none` without replacement

### Active/Pressed Feedback (Optional)
- **Transform**: `transform: scale(0.95)` - subtle shrink
- **Background**: Slightly darker than hover state

## Accessibility Requirements

### Contrast Ratios (WCAG AA)
- **Normal text**: Minimum 4.5:1 contrast with background
- **Large text** (18px+): Minimum 3:1 contrast
- **Interactive elements**: Minimum 3:1 contrast with adjacent colors

### Interactive Elements
- All buttons must have:
  - `aria-label` if no visible text
  - Keyboard focus indicators
  - Minimum touch target: 44x44px (mobile)

### Screen Readers
- Use semantic HTML (`<button>`, `<input>`, `<a>`)
- Add `role` attributes for custom components
- Include `aria-live` for dynamic updates

## Dark Mode Principles

### Color Adjustments
- **Reduce saturation**: Colors less vibrant than light mode
- **Increase brightness**: Blues lighter (#2563eb → #3b82f6)
- **Avoid pure black**: Use dark blue-gray (#0f172a) for depth
- **Avoid pure white**: Use off-white (#f1f5f9) for text

### Elevation in Dark Mode
- **Lighter = Higher**: Elevated surfaces are lighter, not darker
- **Material Design overlay**: Simulate ~16% white overlay for modals
- **Subtle differences**: Small color steps for clear hierarchy

### Contrast Requirements
- Maintain 4.5:1 minimum contrast in both themes
- Test on actual displays (OLED shows different contrast than LCD)
- Use semi-transparent overlays for depth, not solid colors

## Animation & Transitions

### Standard Timing
- **Fast**: `0.15s` - Icon state changes
- **Normal**: `0.2s` - Most interactive feedback
- **Slow**: `0.3s` - Page transitions, large movements

### Easing Functions
- **Default**: `ease` - General purpose
- **Ease-out**: Entering the viewport
- **Ease-in**: Leaving the viewport
- **Ease-in-out**: Moving within viewport

### What to Animate
- ✅ Color changes (text, background, border)
- ✅ Opacity (fade in/out)
- ✅ Transform (scale, translate)
- ✅ Box-shadow
- ❌ Width/Height (causes reflow)
- ❌ Position properties (top, left, etc. - use transform instead)

## Design References

This design system is informed by:
- [Material Design 3](https://m3.material.io/)
- [Apple Human Interface Guidelines](https://developer.apple.com/design/human-interface-guidelines/)
- [WCAG 2.1 AA Standards](https://www.w3.org/WAI/WCAG21/quickref/)
- Modern web interface patterns (GitHub, Linear, Notion)

## When to Update This Document

Update these guidelines when:
- Adding new component types (new buttons, new badge styles, etc.)
- Changing existing patterns (hover behaviors, colors, etc.)
- Discovering accessibility issues
- User feedback suggests inconsistency

These guidelines are living documentation - if something doesn't work well in practice, discuss and update!
