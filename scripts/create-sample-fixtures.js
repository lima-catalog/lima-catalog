#!/usr/bin/env node

/**
 * Create Sample Test Fixtures for E2E Tests
 *
 * This script creates realistic sample YAML files based on catalog data
 * for use in E2E tests when actual downloads aren't possible.
 */

const fs = require('fs');
const path = require('path');

const CATALOG_PATH = process.argv[2] || 'web/catalog.jsonl';
const FIXTURES_DIR = 'tests/e2e/fixtures';
const TEMPLATES_DIR = path.join(FIXTURES_DIR, 'templates');
const MANIFEST_PATH = path.join(FIXTURES_DIR, 'manifest.json');

/**
 * Generate realistic Lima template YAML based on catalog metadata
 */
function generateSampleYAML(template) {
  const hasDocker = template.keywords.includes('docker');
  const hasContainerd = template.keywords.includes('containerd');
  const hasK8s = template.keywords.includes('kubernetes');

  const yaml = `# Lima template: ${template.name}
# ${template.description}
# Category: ${template.category}
# Official: ${template.official}
# Source: ${template.github_url}

images:
  - location: "https://cloud-images.ubuntu.com/releases/24.04/release/ubuntu-24.04-server-cloudimg-amd64.img"
    arch: "x86_64"
    digest: "sha256:example-digest-for-testing"

cpus: 4
memory: "4GiB"
disk: "100GiB"

mounts:
  - location: "~"
    writable: true
  - location: "/tmp/lima"
    writable: true

${hasContainerd ? `containerd:
  system: false
  user: true
` : ''}
${hasDocker ? `provision:
  - mode: system
    script: |
      #!/bin/bash
      set -eux -o pipefail
      # Install Docker
      curl -fsSL https://get.docker.com | sh
      systemctl enable --now docker
` : ''}
${hasK8s ? `  - mode: system
    script: |
      #!/bin/bash
      # Install Kubernetes tools
      apt-get update
      apt-get install -y kubectl
` : ''}
${template.keywords.length > 0 ? `# Keywords: ${template.keywords.join(', ')}` : ''}

# Notability Score: ${template.notability_score.toFixed(1)}
`;

  return yaml;
}

/**
 * Sanitize filename
 */
function sanitizeFilename(str) {
  return str
    .replace(/\//g, '_')
    .replace(/:/g, '_')
    .replace(/\s+/g, '_')
    .replace(/[^a-zA-Z0-9._-]/g, '');
}

/**
 * Select templates for test fixtures
 * Prioritizes first templates in catalog AND alphabetically by name
 */
function selectTemplates(templates, count = 20) {
  const selected = [];

  // IMPORTANT: Include first few templates from catalog
  for (let i = 0; i < Math.min(3, templates.length); i++) {
    selected.push(templates[i]);
  }

  // IMPORTANT: Include first templates sorted by name (default UI sort)
  const byName = [...templates].sort((a, b) => a.name.localeCompare(b.name));
  for (let i = 0; i < Math.min(5, byName.length) && selected.length < count; i++) {
    if (!selected.find(s => s.id === byName[i].id)) {
      selected.push(byName[i]);
    }
  }

  // Official templates
  const official = templates.filter(t =>
    t.official === true && !selected.find(s => s.id === t.id)
  );
  if (official.length > 0) {
    const topOfficial = official.sort((a, b) => b.notability_score - a.notability_score)[0];
    selected.push(topOfficial);

    const otherOfficial = official.find(t =>
      t.id !== topOfficial.id && t.category !== topOfficial.category
    );
    if (otherOfficial) {
      selected.push(otherOfficial);
    }
  }

  // Templates with similars
  const withSimilar = templates.filter(t =>
    t.similar_templates &&
    t.similar_templates.length > 0 &&
    !selected.find(s => s.id === t.id)
  );

  if (withSimilar.length > 0) {
    const bySimilarCount = withSimilar.sort((a, b) =>
      b.similar_templates.length - a.similar_templates.length
    );

    for (let i = 0; i < Math.min(2, bySimilarCount.length) && selected.length < count; i++) {
      selected.push(bySimilarCount[i]);
    }
  }

  // Fill remaining with diverse templates
  const remaining = count - selected.length;
  if (remaining > 0) {
    const candidates = templates.filter(t => !selected.find(s => s.id === t.id));
    const byNotability = candidates.sort((a, b) => b.notability_score - a.notability_score);

    for (let i = 0; i < Math.min(remaining, byNotability.length); i++) {
      selected.push(byNotability[i]);
    }
  }

  return selected.slice(0, count);
}

async function main() {
  console.log('Creating sample test fixtures...\n');

  // Read catalog
  const content = fs.readFileSync(CATALOG_PATH, 'utf8');
  const templates = content.trim().split('\n').map(line => JSON.parse(line));

  // Select templates (increase to 20 to cover more edge cases)
  const selected = selectTemplates(templates, 20);

  // Create directories
  fs.mkdirSync(FIXTURES_DIR, { recursive: true });
  fs.mkdirSync(TEMPLATES_DIR, { recursive: true });

  const manifest = {
    generated_at: new Date().toISOString(),
    template_count: selected.length,
    templates: {}
  };

  // Generate sample YAML files
  for (const template of selected) {
    const filename = sanitizeFilename(template.id) + '.yaml';
    const filepath = path.join(TEMPLATES_DIR, filename);
    const yaml = generateSampleYAML(template);

    fs.writeFileSync(filepath, yaml, 'utf8');
    console.log(`✓ Created: ${filename}`);

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
  }

  // Save manifest
  fs.writeFileSync(MANIFEST_PATH, JSON.stringify(manifest, null, 2), 'utf8');
  console.log(`\n✓ Manifest saved: ${MANIFEST_PATH}`);

  // Print summary
  console.log('\n' + '='.repeat(60));
  console.log('Test Fixtures Summary');
  console.log('='.repeat(60));
  console.log(`Total templates: ${manifest.template_count}`);

  const values = Object.values(manifest.templates);
  const official = values.filter(t => t.official);
  const withSimilars = values.filter(t => t.has_similars);

  console.log(`Official: ${official.length}, Community: ${values.length - official.length}`);
  console.log(`With similars: ${withSimilars.length}`);
  console.log();

  for (const [id, info] of Object.entries(manifest.templates)) {
    const badge = info.official ? '[OFFICIAL]' : '[COMMUNITY]';
    const similars = info.has_similars ? ` (+${info.similar_count} similar)` : '';
    console.log(`${badge} ${info.name}${similars}`);
  }

  console.log('\n✓ Sample fixtures ready for E2E tests');
}

main().catch(console.error);
