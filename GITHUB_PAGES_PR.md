# GitHub Pages Catalog Browser

## Summary

Creates a beautiful, responsive static website to browse and search the Lima template catalog. The site fetches data directly from the `data` branch and provides an interactive interface to explore templates with filtering, search, and sorting capabilities.

## Live Demo

Once deployed, the site will be available at:
**https://lima-catalog.github.io/lima-catalog/**

## Features

### 🔍 **Search & Filter**
- **Full-text search**: Find templates by name, keyword, or technology
- **Category filter**: Browse by containers, development, orchestration, security, etc.
- **Type filter**: Show only official templates or community contributions
- **Smart sorting**: Order by name, popularity (stars), or recent updates

### 📊 **Statistics Dashboard**
- Total template count
- Official vs community breakdown
- Live count of filtered results

### 🎨 **Template Cards**
Each template card displays:
- **Name**: Human-readable display name (e.g., "Container Security")
- **Description**: Auto-generated summary of purpose
- **Category**: Primary use case with icon
- **OS/Image**: Detected operating system
- **Keywords**: Searchable technology tags
- **Repository**: Link to source repo with star count
- **Badge**: Official or Community indicator

### ✨ **User Experience**
- **Responsive design**: Works perfectly on mobile, tablet, and desktop
- **Click to view**: Click any card to open the template source on GitHub
- **Fast loading**: Client-side filtering means instant results
- **Modern UI**: Clean, professional design with proper typography
- **No build step**: Pure HTML/CSS/JavaScript

## How It Works

The site is entirely static and client-side:

1. **Data fetching**: Loads `templates.jsonl` and `repos.jsonl` from the `data` branch via `raw.githubusercontent.com`
2. **Parsing**: Parses JSON Lines format client-side
3. **Rendering**: Creates template cards dynamically
4. **Filtering**: All search/filter operations happen in the browser
5. **Navigation**: Links directly to GitHub for template source

## Files

- `docs/index.html` - Main page structure
- `docs/style.css` - Responsive styling
- `docs/app.js` - Data fetching and interactivity

## Deployment Instructions

### Step 1: Merge This PR

### Step 2: Enable GitHub Pages

1. Go to repository **Settings** → **Pages**
2. Under "Source", select:
   - **Branch**: `main`
   - **Folder**: `/docs`
3. Click **Save**

### Step 3: Wait for Deployment

GitHub Actions will build and deploy the site (takes 1-2 minutes).

### Step 4: Visit Your Site

The catalog will be live at:
**https://lima-catalog.github.io/lima-catalog/**

## Benefits

✅ **Visual validation**: Review auto-generated names and categories
✅ **Discoverability**: Users can easily find templates by technology
✅ **Professional**: Clean interface for community showcase
✅ **Zero maintenance**: Auto-updates when data branch changes
✅ **No hosting costs**: GitHub Pages is free
✅ **SEO friendly**: Static HTML is crawlable by search engines

## Future Enhancements

After initial deployment, could add:
- Template detail pages with full YAML preview
- Category landing pages
- Usage statistics and trending templates
- "Copy lima command" buttons
- Dark mode toggle
- Advanced filters (by architecture, OS, etc.)

## Screenshots

The site features:
- **Header**: Gradient blue header with title and subtitle
- **Controls**: Search bar and filter dropdowns
- **Stats**: Four metric cards showing counts
- **Grid**: Responsive card layout (3 columns on desktop, 1 on mobile)
- **Cards**: Clean white cards with hover effects
- **Footer**: Links to source code and Lima project

## Testing

Before merging, you can preview by:
1. Checking out this branch locally
2. Opening `docs/index.html` in a browser
3. *(Note: Due to CORS, you may need a local server for data fetching)*

Or just merge and enable Pages to see it live!

## Next Steps After Deployment

1. **Review categorization**: Check if auto-assigned categories make sense
2. **Validate names**: Ensure smart name derivation works well
3. **Identify improvements**: See which templates need better descriptions
4. **Consider LLM enhancement**: If needed, implement LLM for better descriptions

This provides the visual feedback loop needed before implementing LLM enhancements!
