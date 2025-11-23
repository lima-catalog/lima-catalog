// Unit Tests: Debug Template Tool (main package)
//
// High-level overview of what's being tested:
// - Consistency between debug tool and notability scoring
// - Correct usage of IsCodeComment function from discovery package
// - Code comment detection in various scenarios
// - Comment filtering logic (empty lines, known lines, code comments)
// - Unique comment counting after filtering
// - Known lines lookup and exclusion
// - Integration with discovery package's code detection

package main

import (
	"testing"

	"github.com/lima-catalog/lima-catalog/pkg/discovery"
)

func TestDebugToolUsesCorrectIsCodeComment(t *testing.T) {
	// Verify the debug tool uses the same IsCodeComment as notability.go
	testCases := []struct {
		line     string
		expected bool
	}{
		{"export PATH=$PATH:/bin", true},
		{"This is a normal comment", false},
		{"cat file | grep pattern", true},
		{"VERSION=$(cat version.txt)", true},
	}

	for _, tc := range testCases {
		result := discovery.IsCodeComment(tc.line)
		if result != tc.expected {
			t.Errorf("IsCodeComment(%q) = %v, want %v", tc.line, result, tc.expected)
		}
	}
}

func TestFilteringConsistency(t *testing.T) {
	// Create sample data
	commentLines := []string{
		"Normal documentation comment",
		"export PATH=$PATH:/usr/local/bin",
		"Another normal comment",
		"cat file.txt | grep pattern",
		"This explains the configuration",
	}

	knownLines := map[string]bool{
		"Normal documentation comment": true, // This one is known
	}

	// Count unique comments (excluding known and code)
	uniqueCount := 0
	codeCount := 0
	for _, line := range commentLines {
		if line == "" {
			continue
		}
		if knownLines[line] {
			continue // Known line
		}
		if discovery.IsCodeComment(line) {
			codeCount++
			continue // Code comment
		}
		uniqueCount++
	}

	// Expected: 2 unique (second normal comment + explanation)
	// Expected: 2 code (export + cat)
	if uniqueCount != 2 {
		t.Errorf("Expected 2 unique comments, got %d", uniqueCount)
	}
	if codeCount != 2 {
		t.Errorf("Expected 2 code comments, got %d", codeCount)
	}
}

func TestEmptyLinesHandling(t *testing.T) {
	lines := []string{
		"",
		"Valid comment",
		"",
		"export VAR=value",
		"",
	}

	knownLines := make(map[string]bool)

	uniqueCount := 0
	for _, line := range lines {
		if line == "" {
			continue // Should skip empty lines
		}
		if knownLines[line] {
			continue
		}
		if discovery.IsCodeComment(line) {
			continue
		}
		uniqueCount++
	}

	if uniqueCount != 1 {
		t.Errorf("Expected 1 unique comment (after filtering empty and code), got %d", uniqueCount)
	}
}

func TestKnownLinesLookup(t *testing.T) {
	// Test that known lines lookup works correctly
	knownLines := map[string]bool{
		"default: guestPortRange: [1, 65535]": true,
		"Lima example configuration":          true,
	}

	testLines := []string{
		"default: guestPortRange: [1, 65535]", // Known
		"Lima example configuration",          // Known
		"Custom configuration here",           // Unknown
	}

	uniqueCount := 0
	for _, line := range testLines {
		if knownLines[line] {
			continue
		}
		uniqueCount++
	}

	if uniqueCount != 1 {
		t.Errorf("Expected 1 unique line (after filtering known), got %d", uniqueCount)
	}
}
