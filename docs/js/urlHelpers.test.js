/**
 * Tests for urlHelpers.js
 */

import { runner, assert } from './test-framework.js';
import { getDefaultBranchURL, getGitHubSchemeURL, getRawContentURL } from './urlHelpers.js';

// Test getDefaultBranchURL
runner.test('getDefaultBranchURL: converts commit SHA URL to default branch', () => {
    const template = {
        url: 'https://github.com/owner/repo/blob/abc123def456abc123def456abc123def456abc1/path/to/template.yaml'
    };
    const repo = {
        default_branch: 'main'
    };
    const result = getDefaultBranchURL(template, repo);
    assert.equal(result, 'https://github.com/owner/repo/blob/main/path/to/template.yaml');
});

runner.test('getDefaultBranchURL: returns original URL if no repo info', () => {
    const template = {
        url: 'https://github.com/owner/repo/blob/abc123def456abc123def456abc123def456abc1/template.yaml'
    };
    const result = getDefaultBranchURL(template, null);
    assert.equal(result, template.url);
});

runner.test('getDefaultBranchURL: returns original URL if not commit URL pattern', () => {
    const template = {
        url: 'https://github.com/owner/repo/blob/main/template.yaml'
    };
    const repo = {
        default_branch: 'main'
    };
    const result = getDefaultBranchURL(template, repo);
    assert.equal(result, template.url);
});

runner.test('getDefaultBranchURL: handles deep nested paths', () => {
    const template = {
        url: 'https://github.com/lima-vm/lima/blob/1234567890abcdef1234567890abcdef12345678/examples/alpine.yaml'
    };
    const repo = {
        default_branch: 'master'
    };
    const result = getDefaultBranchURL(template, repo);
    assert.equal(result, 'https://github.com/lima-vm/lima/blob/master/examples/alpine.yaml');
});

// Test getGitHubSchemeURL
runner.test('getGitHubSchemeURL: generates standard format', () => {
    const template = {
        repo: 'owner/repo',
        path: 'examples/alpine.yaml'
    };
    const result = getGitHubSchemeURL(template);
    assert.equal(result, 'github:owner/repo/examples/alpine');
});

runner.test('getGitHubSchemeURL: removes .yaml extension', () => {
    const template = {
        repo: 'owner/repo',
        path: 'template.yaml'
    };
    const result = getGitHubSchemeURL(template);
    assert.equal(result, 'github:owner/repo/template');
});

runner.test('getGitHubSchemeURL: removes .lima from path', () => {
    const template = {
        repo: 'owner/repo',
        path: 'examples/.lima.yaml'
    };
    const result = getGitHubSchemeURL(template);
    assert.equal(result, 'github:owner/repo/examples');
});

runner.test('getGitHubSchemeURL: handles root .lima.yaml', () => {
    const template = {
        repo: 'owner/repo',
        path: '.lima.yaml'
    };
    const result = getGitHubSchemeURL(template);
    assert.equal(result, 'github:owner/repo');
});

runner.test('getGitHubSchemeURL: handles org repos (owner == repo)', () => {
    const template = {
        repo: 'lima-vm/lima-vm',
        path: '.lima.yaml'
    };
    const result = getGitHubSchemeURL(template);
    assert.equal(result, 'github:lima-vm');
});

runner.test('getGitHubSchemeURL: handles org repos with path', () => {
    const template = {
        repo: 'myorg/myorg',
        path: 'templates/alpine.yaml'
    };
    const result = getGitHubSchemeURL(template);
    assert.equal(result, 'github:myorg//templates/alpine');
});

runner.test('getGitHubSchemeURL: handles empty path', () => {
    const template = {
        repo: 'owner/repo',
        path: ''
    };
    const result = getGitHubSchemeURL(template);
    assert.equal(result, 'github:owner/repo');
});

// Test getRawContentURL
runner.test('getRawContentURL: converts blob URL to raw URL', () => {
    const url = 'https://github.com/owner/repo/blob/main/template.yaml';
    const result = getRawContentURL(url);
    assert.equal(result, 'https://raw.githubusercontent.com/owner/repo/main/template.yaml');
});

runner.test('getRawContentURL: handles commit SHA URLs', () => {
    const url = 'https://github.com/owner/repo/blob/abc123def456abc123def456abc123def456abc1/template.yaml';
    const result = getRawContentURL(url);
    assert.equal(result, 'https://raw.githubusercontent.com/owner/repo/abc123def456abc123def456abc123def456abc1/template.yaml');
});

runner.test('getRawContentURL: handles nested paths', () => {
    const url = 'https://github.com/lima-vm/lima/blob/master/examples/alpine.yaml';
    const result = getRawContentURL(url);
    assert.equal(result, 'https://raw.githubusercontent.com/lima-vm/lima/master/examples/alpine.yaml');
});
