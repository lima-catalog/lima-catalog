#!/usr/bin/env node

/**
 * Test Fixture Extraction Tool for E2E Tests
 *
 * This script extracts a representative subset of templates from the catalog
 * and downloads their YAML files to use as test fixtures in E2E tests.
 *
 * Usage:
 *   node scripts/create-test-fixtures.js
 *
 * Requirements:
 *   - catalog.jsonl must be available (from data branch or web/)
 *   - Internet connection to download YAML files from GitHub
 *
 * Output:
 *   - tests/e2e/fixtures/templates/ - Downloaded YAML files
 *   - tests/e2e/fixtures/manifest.json - Mapping of template IDs to files
 */

const fs = require('fs');
const path = require('path');
const https = require('https');
const { promisify } = require('util');

const writeFile = promisify(fs.writeFile);
const mkdir = promisify(fs.mkdir);

// Configuration
const CATALOG_PATH = process.argv[2] || 'web/catalog.jsonl';
const FIXTURES_DIR = 'tests/e2e/fixtures';
const TEMPLATES_DIR = path.join(FIXTURES_DIR, 'templates');
const MANIFEST_PATH = path.join(FIXTURES_DIR, 'manifest.json');
const NUM_TEMPLATES = 10; // Number of templates to extract

/**
 * Read and parse catalog.jsonl
 */
function readCatalog(catalogPath) {
  console.log(`Reading catalog from: ${catalogPath}`);
  const content = fs.readFileSync(catalogPath, 'utf8');
  const lines = content.trim().split('\n');
  const templates = lines.map(line => JSON.parse(line));
  console.log(`Found ${templates.length} templates in catalog`);
  return templates;
}

/**
 * Select diverse, representative templates for testing
 *
 * Selection criteria:
 * - At least 1 official template from lima-vm/lima
 * - Mix of official and community templates
 * - Different categories (development, containers, etc.)
 * - Templates with similar_templates (for diff testing)
 * - Templates without similar_templates
 * - Various keyword counts
 * - Range of notability scores
 */
function selectTemplates(templates, count = NUM_TEMPLATES) {
  console.log('\nSelecting representative templates...');

  const selected = [];

  // 1. Select official templates (2-3 templates)
  const official = templates.filter(t => t.official === true);
  if (official.length > 0) {
    // Get the most notable official template
    const topOfficial = official.sort((a, b) => b.notability_score - a.notability_score)[0];
    selected.push(topOfficial);
    console.log(`  ✓ Official: ${topOfficial.id} (${topOfficial.category})`);

    // Get another official from a different category if available
    const otherOfficial = official.find(t =>
      t.id !== topOfficial.id &&
      t.category !== topOfficial.category
    );
    if (otherOfficial) {
      selected.push(otherOfficial);
      console.log(`  ✓ Official: ${otherOfficial.id} (${otherOfficial.category})`);
    }
  }

  // 2. Select templates with similar_templates (2-3 templates)
  // These are important for testing the modal's diff feature
  const withSimilar = templates.filter(t =>
    t.similar_templates &&
    t.similar_templates.length > 0 &&
    !selected.find(s => s.id === t.id)
  );

  if (withSimilar.length > 0) {
    // Pick templates with the most similar templates
    const bySimilarCount = withSimilar.sort((a, b) =>
      b.similar_templates.length - a.similar_templates.length
    );

    for (let i = 0; i < Math.min(2, bySimilarCount.length); i++) {
      selected.push(bySimilarCount[i]);
      console.log(`  ✓ With similars: ${bySimilarCount[i].id} (${bySimilarCount[i].similar_templates.length} similar)`);
    }
  }

  // 3. Select templates from different categories
  const categories = [...new Set(templates.map(t => t.category).filter(Boolean))];
  const categoriesNeeded = categories.filter(cat =>
    !selected.find(s => s.category === cat)
  );

  for (const category of categoriesNeeded.slice(0, 2)) {
    const inCategory = templates.filter(t =>
      t.category === category &&
      !selected.find(s => s.id === t.id)
    );

    if (inCategory.length > 0) {
      // Pick the most notable one in this category
      const best = inCategory.sort((a, b) => b.notability_score - a.notability_score)[0];
      selected.push(best);
      console.log(`  ✓ Category ${category}: ${best.id}`);
    }
  }

  // 4. Fill remaining slots with diverse templates
  const remaining = count - selected.length;
  if (remaining > 0) {
    const candidates = templates.filter(t => !selected.find(s => s.id === t.id));

    // Sort by notability to get quality templates
    const byNotability = candidates.sort((a, b) => b.notability_score - a.notability_score);

    // Pick templates with varying characteristics
    for (let i = 0; i < Math.min(remaining, byNotability.length); i++) {
      selected.push(byNotability[i]);
      console.log(`  ✓ Diverse: ${byNotability[i].id} (score: ${byNotability[i].notability_score.toFixed(1)})`);
    }
  }

  console.log(`\nSelected ${selected.length} templates`);
  return selected.slice(0, count);
}

/**
 * Download a file from a URL using HTTPS
 */
function downloadFile(url) {
  return new Promise((resolve, reject) => {
    console.log(`  Downloading: ${url}`);

    https.get(url, (res) => {
      if (res.statusCode === 301 || res.statusCode === 302) {
        // Follow redirect
        return downloadFile(res.headers.location).then(resolve).catch(reject);
      }

      if (res.statusCode !== 200) {
        reject(new Error(`Failed to download ${url}: HTTP ${res.statusCode}`));
        return;
      }

      let data = '';
      res.on('data', chunk => { data += chunk; });
      res.on('end', () => resolve(data));
    }).on('error', reject);
  });
}

/**
 * Sanitize filename to be safe for filesystem
 */
function sanitizeFilename(str) {
  return str
    .replace(/\//g, '_')
    .replace(/:/g, '_')
    .replace(/\s+/g, '_')
    .replace(/[^a-zA-Z0-9._-]/g, '');
}

/**
 * Download YAML files for selected templates
 */
async function downloadTemplates(templates) {
  console.log('\nDownloading YAML files...');

  // Create directories if they don't exist
  await mkdir(FIXTURES_DIR, { recursive: true });
  await mkdir(TEMPLATES_DIR, { recursive: true });

  const manifest = {
    generated_at: new Date().toISOString(),
    template_count: templates.length,
    templates: {}
  };

  for (const template of templates) {
    try {
      // Create a safe filename based on the template ID
      const filename = sanitizeFilename(template.id) + '.yaml';
      const filepath = path.join(TEMPLATES_DIR, filename);

      // Download the YAML file
      const content = await downloadFile(template.raw_url);

      // Save to disk
      await writeFile(filepath, content, 'utf8');
      console.log(`  ✓ Saved: ${filename}`);

      // Add to manifest
      manifest.templates[template.id] = {
        filename: filename,
        raw_url: template.raw_url,
        github_url: template.github_url,
        name: template.name,
        category: template.category,
        official: template.official,
        has_similars: !!(template.similar_templates && template.similar_templates.length > 0),
        similar_count: template.similar_templates ? template.similar_templates.length : 0,
        keywords: template.keywords || [],
        notability_score: template.notability_score
      };

    } catch (error) {
      console.error(`  ✗ Failed to download ${template.id}: ${error.message}`);
    }
  }

  // Save manifest
  await writeFile(MANIFEST_PATH, JSON.stringify(manifest, null, 2), 'utf8');
  console.log(`\n✓ Manifest saved to: ${MANIFEST_PATH}`);

  return manifest;
}

/**
 * Generate a summary of the fixtures
 */
function printSummary(manifest) {
  console.log('\n' + '='.repeat(60));
  console.log('Test Fixtures Summary');
  console.log('='.repeat(60));
  console.log(`Total templates: ${manifest.template_count}`);
  console.log(`Generated at: ${manifest.generated_at}`);
  console.log();

  const templates = Object.values(manifest.templates);
  const official = templates.filter(t => t.official);
  const community = templates.filter(t => !t.official);
  const withSimilars = templates.filter(t => t.has_similars);
  const categories = [...new Set(templates.map(t => t.category))];

  console.log(`Official templates: ${official.length}`);
  console.log(`Community templates: ${community.length}`);
  console.log(`Templates with similars: ${withSimilars.length}`);
  console.log(`Categories covered: ${categories.join(', ')}`);
  console.log();

  console.log('Templates:');
  for (const [id, info] of Object.entries(manifest.templates)) {
    const badge = info.official ? '[OFFICIAL]' : '[COMMUNITY]';
    const similars = info.has_similars ? `(+${info.similar_count} similar)` : '';
    console.log(`  ${badge} ${info.name} - ${info.category} ${similars}`);
    console.log(`    ${id}`);
  }

  console.log();
  console.log('✓ Fixtures ready for E2E tests');
  console.log('  Location: ' + TEMPLATES_DIR);
  console.log('  Manifest: ' + MANIFEST_PATH);
  console.log();
  console.log('Next steps:');
  console.log('  1. Update tests/e2e/fixtures.js to serve these local files');
  console.log('  2. Run E2E tests: npm run test:e2e');
  console.log('='.repeat(60));
}

/**
 * Main execution
 */
async function main() {
  try {
    console.log('Lima Catalog - Test Fixture Extraction Tool');
    console.log('='.repeat(60));

    // Read catalog
    const templates = readCatalog(CATALOG_PATH);

    // Select representative templates
    const selected = selectTemplates(templates, NUM_TEMPLATES);

    // Download templates
    const manifest = await downloadTemplates(selected);

    // Print summary
    printSummary(manifest);

    process.exit(0);
  } catch (error) {
    console.error('\n✗ Error:', error.message);
    console.error(error.stack);
    process.exit(1);
  }
}

// Run if called directly
if (require.main === module) {
  main();
}

module.exports = { selectTemplates, downloadFile, sanitizeFilename };
