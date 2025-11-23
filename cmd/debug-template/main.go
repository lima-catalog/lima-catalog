// Command debug-template analyzes a Lima template and shows unique content not found in official templates.
//
// The tool compares a template against official templates from lima-vm/lima to identify
// unique comments, provision scripts, probes, messages, and unusual image domains.
// This helps evaluate template notability and discover interesting customizations.
//
// Usage:
//
//	debug-template -url <template-url> [-official <path-to-official.json>]
//
// Example:
//
//	debug-template -url github:lima-vm/lima/templates/ubuntu.yaml
//	debug-template -url https://raw.githubusercontent.com/owner/repo/main/lima.yaml
//
// The official.json file contains known lines extracted from official templates.
// Generate it using the lima-catalog tool with ANALYZE=1 LIMA_REPO_PATH=/path/to/lima.
package main

import (
	"context"
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
	ctx := context.Background()
	info, err := discovery.ParseTemplate(ctx, *templateURL, httpClient)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nTemplate parsed successfully:\n")
	fmt.Printf("  - %d comment lines\n", len(info.CommentLines))
	fmt.Printf("  - %d provision lines\n", len(info.ProvisionLines))
	fmt.Printf("  - %d probe lines\n", len(info.ProbeLines))
	fmt.Printf("  - %d message lines\n", len(info.MessageLines))
	fmt.Printf("  - %d images\n", len(info.Images))

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

	// Build official domains map
	officialDomains := make(map[string]bool)
	for _, domain := range ok.Images {
		officialDomains[domain] = true
	}

	// Debug: Show all images found
	if len(info.Images) > 0 {
		fmt.Printf("\nAll images found in template:\n")
		for _, img := range info.Images {
			fmt.Printf("  - %s\n", img)
		}
	}

	// Identify unusual images
	unusualImages := discovery.IdentifyUnusualImages(info.Images, officialDomains)

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
		if discovery.IsCodeComment(line) {
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

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("UNUSUAL IMAGES (not in official templates):")
	fmt.Println(strings.Repeat("=", 80))
	if len(unusualImages) == 0 {
		fmt.Println("       (none - all images use official domains)")
	} else {
		for _, domain := range unusualImages {
			fmt.Printf("       %s\n", domain)
		}
	}
	fmt.Printf("\nUnusual image domains: %d\n", len(unusualImages))
	if len(unusualImages) > 0 {
		fmt.Println("Template gets unusual_images bonus: YES (30 points)")
	} else {
		fmt.Println("Template gets unusual_images bonus: NO")
	}

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("SUMMARY:")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Unique comment lines:   %d (+ %d filtered as code)\n", uniqueComments, codeComments)
	fmt.Printf("Unique provision lines: %d\n", uniqueProvision)
	fmt.Printf("Unique probe lines:     %d\n", uniqueProbes)
	fmt.Printf("Unique message lines:   %d\n", uniqueMessages)
	fmt.Printf("Unusual image domains:  %d", len(unusualImages))
	if len(unusualImages) > 0 {
		fmt.Printf(" (bonus: YES)")
	}
	fmt.Println()
}
