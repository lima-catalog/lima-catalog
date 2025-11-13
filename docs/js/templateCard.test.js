/**
 * Tests for templateCard.js
 */

import { runner, assert } from './test-framework.js';
import { escapeHtml, formatName, formatCategoryName, deriveDisplayName } from './templateCard.js';

// Test escapeHtml
runner.test('escapeHtml: escapes HTML special characters', () => {
    const result = escapeHtml('<script>alert("xss")</script>');
    assert.equal(result, '&lt;script&gt;alert("xss")&lt;/script&gt;');
});

runner.test('escapeHtml: escapes ampersands', () => {
    const result = escapeHtml('Tom & Jerry');
    assert.equal(result, 'Tom &amp; Jerry');
});

runner.test('escapeHtml: escapes quotes', () => {
    const result = escapeHtml('He said "hello"');
    assert.includes(result, 'hello');
    // Different browsers may escape quotes differently
});

runner.test('escapeHtml: handles empty string', () => {
    const result = escapeHtml('');
    assert.equal(result, '');
});

runner.test('escapeHtml: preserves safe text', () => {
    const result = escapeHtml('Hello World 123');
    assert.equal(result, 'Hello World 123');
});

// Test formatName
runner.test('formatName: formats hyphenated names', () => {
    const result = formatName('alpine-docker');
    assert.equal(result, 'Alpine Docker');
});

runner.test('formatName: formats underscored names', () => {
    const result = formatName('ubuntu_dev');
    assert.equal(result, 'Ubuntu Dev');
});

runner.test('formatName: formats mixed separators', () => {
    const result = formatName('alpine-dev_server');
    assert.equal(result, 'Alpine Dev Server');
});

runner.test('formatName: capitalizes each word', () => {
    const result = formatName('one-two-three');
    assert.equal(result, 'One Two Three');
});

runner.test('formatName: handles single word', () => {
    const result = formatName('alpine');
    assert.equal(result, 'Alpine');
});

runner.test('formatName: handles already capitalized', () => {
    const result = formatName('Alpine');
    assert.equal(result, 'Alpine');
});

// Test formatCategoryName
runner.test('formatCategoryName: formats category with hyphens', () => {
    const result = formatCategoryName('operating-systems');
    assert.equal(result, 'Operating Systems');
});

runner.test('formatCategoryName: formats single word category', () => {
    const result = formatCategoryName('containers');
    assert.equal(result, 'Containers');
});

runner.test('formatCategoryName: formats multi-word category', () => {
    const result = formatCategoryName('web-development-tools');
    assert.equal(result, 'Web Development Tools');
});

// Test deriveDisplayName
runner.test('deriveDisplayName: uses display_name if available', () => {
    const template = {
        display_name: 'My Custom Name',
        name: 'ignored',
        path: 'ignored.yaml'
    };
    const result = deriveDisplayName(template);
    assert.equal(result, 'My Custom Name');
});

runner.test('deriveDisplayName: uses name if no display_name', () => {
    const template = {
        name: 'alpine',
        path: 'alpine.yaml'
    };
    const result = deriveDisplayName(template);
    assert.equal(result, 'alpine');
});

runner.test('deriveDisplayName: derives from path filename', () => {
    const template = {
        path: 'examples/alpine-docker.yaml',
        repo: 'owner/repo'
    };
    const result = deriveDisplayName(template);
    assert.equal(result, 'Alpine Docker');
});

runner.test('deriveDisplayName: removes .yaml extension', () => {
    const template = {
        path: 'ubuntu-dev.yaml',
        repo: 'owner/repo'
    };
    const result = deriveDisplayName(template);
    assert.equal(result, 'Ubuntu Dev');
});

runner.test('deriveDisplayName: removes .yml extension', () => {
    const template = {
        path: 'debian.yml',
        repo: 'owner/repo'
    };
    const result = deriveDisplayName(template);
    assert.equal(result, 'Debian');
});

runner.test('deriveDisplayName: handles generic filename (lima) with parent directory', () => {
    const template = {
        path: 'alpine/lima.yaml',
        repo: 'owner/repo'
    };
    const result = deriveDisplayName(template);
    assert.equal(result, 'Alpine');
});

runner.test('deriveDisplayName: handles generic filename (template) with parent directory', () => {
    const template = {
        path: 'ubuntu/template.yaml',
        repo: 'owner/repo'
    };
    const result = deriveDisplayName(template);
    assert.equal(result, 'Ubuntu');
});

runner.test('deriveDisplayName: falls back to repo name for generic filename at root', () => {
    const template = {
        path: 'lima.yaml',
        repo: 'owner/alpine-docker'
    };
    const result = deriveDisplayName(template);
    assert.equal(result, 'Alpine Docker');
});

runner.test('deriveDisplayName: handles deep nested paths', () => {
    const template = {
        path: 'templates/linux/alpine.yaml',
        repo: 'owner/repo'
    };
    const result = deriveDisplayName(template);
    assert.equal(result, 'Alpine');
});

runner.test('deriveDisplayName: handles generic config filename', () => {
    const template = {
        path: 'k8s/config.yaml',
        repo: 'owner/repo'
    };
    const result = deriveDisplayName(template);
    assert.equal(result, 'K8s');
});

runner.test('deriveDisplayName: handles generic default filename', () => {
    const template = {
        path: 'docker/default.yaml',
        repo: 'owner/repo'
    };
    const result = deriveDisplayName(template);
    assert.equal(result, 'Docker');
});
