package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/lima-catalog/lima-catalog/pkg/discovery"
	"github.com/lima-catalog/lima-catalog/pkg/interfaces"
)

func main() {
	templateURL := flag.String("url", "", "Template URL to analyze")
	officialPath := flag.String("official", "data/official.json", "Path to official.json")
	flag.Parse()

	if *templateURL == "" {
		fmt.Println("Usage: debug-template -url <template-url> [-official <path-to-official.json>]")
		fmt.Println()
		fmt.Println("This tool shows what lines remain in a template after removing known lines from official templates.")
		os.Exit(1)
	}

	// Load official knowledge
	ok, err := discovery.LoadOfficialKnowledge(*officialPath)
	if err != nil {
		fmt.Printf("Error loading official knowledge: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded official knowledge:\n")
	fmt.Printf("  - %d known comment lines\n", len(ok.KnownLines.Comments))
	fmt.Printf("  - %d known provision lines\n", len(ok.KnownLines.Provision))
	fmt.Printf("  - %d known probe lines\n", len(ok.KnownLines.Probes))
	fmt.Printf("  - %d known message lines\n", len(ok.KnownLines.Messages))
	fmt.Printf("  - %d known image domains\n\n", len(ok.Images))

	// Parse template
	fmt.Printf("Fetching template from: %s\n", *templateURL)
	httpClient := interfaces.NewDefaultHTTPClient()
	info, err := discovery.ParseTemplate(*templateURL, httpClient)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nTemplate parsed successfully:\n")
	fmt.Printf("  - %d comment lines\n", len(info.CommentLines))
	fmt.Printf("  - %d provision lines\n", len(info.ProvisionLines))
	fmt.Printf("  - %d probe lines\n", len(info.ProbeLines))
	fmt.Printf("  - %d message lines\n", len(info.MessageLines))

	// Build lookup maps
	knownComments := make(map[string]bool)
	for _, line := range ok.KnownLines.Comments {
		knownComments[line] = true
	}

	knownProvision := make(map[string]bool)
	for _, line := range ok.KnownLines.Provision {
		knownProvision[line] = true
	}

	knownProbes := make(map[string]bool)
	for _, line := range ok.KnownLines.Probes {
		knownProbes[line] = true
	}

	knownMessages := make(map[string]bool)
	for _, line := range ok.KnownLines.Messages {
		knownMessages[line] = true
	}

	// Filter and show unique lines
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("UNIQUE COMMENT LINES (not in official templates):")
	fmt.Println(strings.Repeat("=", 80))
	uniqueComments := 0
	codeComments := 0
	for _, line := range info.CommentLines {
		if line == "" {
			continue
		}
		if knownComments[line] {
			continue
		}
		if isCodeComment(line) {
			codeComments++
			fmt.Printf("[CODE] %s\n", line)
			continue
		}
		uniqueComments++
		fmt.Printf("       %s\n", line)
	}
	fmt.Printf("\nUnique comments: %d (+ %d filtered as code)\n", uniqueComments, codeComments)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("UNIQUE PROVISION LINES (not in official templates):")
	fmt.Println(strings.Repeat("=", 80))
	uniqueProvision := 0
	for _, line := range info.ProvisionLines {
		if line == "" {
			continue
		}
		if knownProvision[line] {
			continue
		}
		uniqueProvision++
		fmt.Printf("       %s\n", line)
	}
	fmt.Printf("\nUnique provision lines: %d\n", uniqueProvision)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("UNIQUE PROBE LINES (not in official templates):")
	fmt.Println(strings.Repeat("=", 80))
	uniqueProbes := 0
	for _, line := range info.ProbeLines {
		if line == "" {
			continue
		}
		if knownProbes[line] {
			continue
		}
		uniqueProbes++
		fmt.Printf("       %s\n", line)
	}
	fmt.Printf("\nUnique probe lines: %d\n", uniqueProbes)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("UNIQUE MESSAGE LINES (not in official templates):")
	fmt.Println(strings.Repeat("=", 80))
	uniqueMessages := 0
	for _, line := range info.MessageLines {
		if line == "" {
			continue
		}
		if knownMessages[line] {
			continue
		}
		uniqueMessages++
		fmt.Printf("       %s\n", line)
	}
	fmt.Printf("\nUnique message lines: %d\n", uniqueMessages)

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("SUMMARY:")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Unique comment lines:   %d (+ %d filtered as code)\n", uniqueComments, codeComments)
	fmt.Printf("Unique provision lines: %d\n", uniqueProvision)
	fmt.Printf("Unique probe lines:     %d\n", uniqueProbes)
	fmt.Printf("Unique message lines:   %d\n", uniqueMessages)
}

// isCodeComment checks if a comment line looks like commented-out code
// Uses conservative heuristics to avoid false positives
func isCodeComment(line string) bool {
	// Check for shell variable expansion ($VAR, ${VAR}, $(...))
	if strings.Contains(line, "$") {
		dollarIdx := strings.Index(line, "$")
		if dollarIdx+1 < len(line) {
			nextChar := line[dollarIdx+1]
			if nextChar == '{' || nextChar == '(' || (nextChar >= 'A' && nextChar <= 'Z') || (nextChar >= 'a' && nextChar <= 'z') || nextChar == '_' {
				return true
			}
		}
	}

	// Check for backticks (command substitution)
	if strings.Contains(line, "`") {
		return true
	}

	// Check for pipes (excluding markdown table separators)
	if strings.Contains(line, "|") {
		// Allow markdown tables (like "| col1 | col2 |")
		// But reject shell pipes (like "grep foo | sort")
		pipeCount := strings.Count(line, "|")
		if pipeCount == 1 || !strings.Contains(line, " | ") {
			return true
		}
	}

	// Check for command chaining
	if strings.Contains(line, "&&") || strings.Contains(line, "||") {
		return true
	}

	// Check for shell redirects (>, >>, <, 2>&1)
	if strings.Contains(line, ">>") || strings.Contains(line, "2>&1") {
		return true
	}
	if strings.Contains(line, ">") || strings.Contains(line, "<") {
		// Make sure it's not just a comparison operator or arrow
		for i, char := range line {
			if char == '>' || char == '<' {
				// Check context - is this a redirect or part of text?
				before := ""
				after := ""
				if i > 0 {
					before = string(line[i-1])
				}
				if i+1 < len(line) {
					after = string(line[i+1])
				}
				// If surrounded by spaces, likely a redirect
				if (before == " " || before == "") && (after == " " || after == "") {
					return true
				}
			}
		}
	}

	// Check for variable assignment (VAR=value, excluding URLs like http://...)
	if strings.Contains(line, "=") && !strings.Contains(line, "://") {
		eqIdx := strings.Index(line, "=")
		if eqIdx > 0 {
			// Get word before =
			before := line[:eqIdx]
			// Check if it looks like a variable name (alphanumeric + underscore)
			if len(before) > 0 && !strings.Contains(before, " ") {
				if (before[0] >= 'A' && before[0] <= 'Z') ||
					(before[0] >= 'a' && before[0] <= 'z') ||
					before[0] == '_' {
					return true
				}
			}
		}
	}

	// Check if line starts with absolute file path
	if strings.HasPrefix(line, "/etc/") || strings.HasPrefix(line, "/usr/") ||
		strings.HasPrefix(line, "/var/") || strings.HasPrefix(line, "/opt/") ||
		strings.HasPrefix(line, "/tmp/") || strings.HasPrefix(line, "/home/") ||
		strings.HasPrefix(line, "~/") {
		return true
	}

	return false
}
