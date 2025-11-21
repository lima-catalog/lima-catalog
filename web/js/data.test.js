/**
 * Tests for data.js
 */

import { runner, assert } from './test-framework.js';
import { parseJsonLines } from './data.js';

// Test parseJsonLines
runner.test('parseJsonLines: parses single line', () => {
    const input = '{"name":"alpine","path":"alpine.yaml"}';
    const result = parseJsonLines(input);
    assert.equal(result.length, 1);
    assert.equal(result[0].name, 'alpine');
    assert.equal(result[0].path, 'alpine.yaml');
});

runner.test('parseJsonLines: parses multiple lines', () => {
    const input = `{"name":"alpine","path":"alpine.yaml"}
{"name":"ubuntu","path":"ubuntu.yaml"}
{"name":"debian","path":"debian.yaml"}`;
    const result = parseJsonLines(input);
    assert.equal(result.length, 3);
    assert.equal(result[0].name, 'alpine');
    assert.equal(result[1].name, 'ubuntu');
    assert.equal(result[2].name, 'debian');
});

runner.test('parseJsonLines: handles empty lines', () => {
    const input = `{"name":"alpine"}

{"name":"ubuntu"}

`;
    const result = parseJsonLines(input);
    assert.equal(result.length, 2);
    assert.equal(result[0].name, 'alpine');
    assert.equal(result[1].name, 'ubuntu');
});

runner.test('parseJsonLines: handles whitespace', () => {
    const input = `  {"name":"alpine"}
  {"name":"ubuntu"}  `;
    const result = parseJsonLines(input);
    assert.equal(result.length, 2);
    assert.equal(result[0].name, 'alpine');
    assert.equal(result[1].name, 'ubuntu');
});

runner.test('parseJsonLines: handles empty string', () => {
    const input = '';
    const result = parseJsonLines(input);
    assert.equal(result.length, 0);
});

runner.test('parseJsonLines: handles complex objects', () => {
    const input = '{"name":"test","keywords":["k8s","docker"],"categories":["containers"]}';
    const result = parseJsonLines(input);
    assert.equal(result.length, 1);
    assert.deepEqual(result[0].keywords, ['k8s', 'docker']);
    assert.deepEqual(result[0].categories, ['containers']);
});

runner.test('parseJsonLines: throws on invalid JSON', () => {
    const input = '{invalid json}';
    assert.throws(() => parseJsonLines(input), 'JSON');
});
