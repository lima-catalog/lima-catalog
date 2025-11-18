package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/lima-catalog/lima-catalog/pkg/prompt"
)

const usage = `Lima Template Prompt Generator

Generates LLM prompts for analyzing Lima VM templates.

Usage:
  prompt-generator <owner/repo/path/to/template.yaml> [options]
  prompt-generator -owner=lima-vm -repo=lima -path=templates/ubuntu.yaml

Options:
  -owner string
        GitHub repository owner (alternative to full path)
  -repo string
        GitHub repository name (alternative to full path)
  -path string
        Path to template file in repository (alternative to full path)
  -context int
        Number of context lines to include around references (default 15)
  -no-comments
        Exclude YAML comments from prompt
  -no-references
        Exclude template file references from prompt
  -no-readme
        Exclude README from prompt
  -max-readme int
        Maximum README length in characters (default 5000, 0 = unlimited)
  -max-refs int
        Maximum number of reference files to include (default 10, 0 = unlimited)
  -output string
        Output file path (default: stdout)
  -help
        Show this help message

Environment Variables:
  GITHUB_TOKEN    GitHub API token (required)

Examples:
  # Using full path notation
  prompt-generator lima-vm/lima/templates/ubuntu.yaml

  # Using separate flags
  prompt-generator -owner=lima-vm -repo=lima -path=templates/ubuntu.yaml

  # Save to file
  prompt-generator lima-vm/lima/templates/ubuntu.yaml -output=ubuntu-prompt.txt

  # Minimal prompt (no references, no README)
  prompt-generator lima-vm/lima/templates/ubuntu.yaml -no-references -no-readme

  # Custom context lines
  prompt-generator lima-vm/lima/templates/ubuntu.yaml -context=20
`

func main() {
	var (
		owner           string
		repo            string
		path            string
		contextLines    int
		noComments      bool
		noReferences    bool
		noReadme        bool
		maxReadmeLength int
		maxRefFiles     int
		outputFile      string
		showHelp        bool
	)

	flag.StringVar(&owner, "owner", "", "GitHub repository owner")
	flag.StringVar(&repo, "repo", "", "GitHub repository name")
	flag.StringVar(&path, "path", "", "Path to template file")
	flag.IntVar(&contextLines, "context", 15, "Number of context lines around references")
	flag.BoolVar(&noComments, "no-comments", false, "Exclude YAML comments")
	flag.BoolVar(&noReferences, "no-references", false, "Exclude template references")
	flag.BoolVar(&noReadme, "no-readme", false, "Exclude README")
	flag.IntVar(&maxReadmeLength, "max-readme", 5000, "Maximum README length")
	flag.IntVar(&maxRefFiles, "max-refs", 10, "Maximum reference files")
	flag.StringVar(&outputFile, "output", "", "Output file (default: stdout)")
	flag.BoolVar(&showHelp, "help", false, "Show help")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Parse positional argument if provided (owner/repo/path format)
	if flag.NArg() > 0 {
		fullPath := flag.Arg(0)
		parts := strings.SplitN(fullPath, "/", 3)
		if len(parts) < 3 {
			fmt.Fprintf(os.Stderr, "Error: invalid template path format. Expected: owner/repo/path/to/template.yaml\n")
			fmt.Fprintf(os.Stderr, "Got: %s\n\n", fullPath)
			flag.Usage()
			os.Exit(1)
		}
		owner = parts[0]
		repo = parts[1]
		path = parts[2]
	}

	// Validate required arguments
	if owner == "" || repo == "" || path == "" {
		fmt.Fprintf(os.Stderr, "Error: missing required arguments\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Get GitHub token from environment
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		fmt.Fprintf(os.Stderr, "Error: GITHUB_TOKEN environment variable is required\n")
		fmt.Fprintf(os.Stderr, "Set it with: export GITHUB_TOKEN=your_github_token\n")
		os.Exit(1)
	}

	// Build configuration
	config := &prompt.PromptConfig{
		ContextLines:      contextLines,
		IncludeComments:   !noComments,
		IncludeReferences: !noReferences,
		IncludeReadme:     !noReadme,
		MaxReadmeLength:   maxReadmeLength,
		MaxReferenceFiles: maxRefFiles,
	}

	// Create builder
	ctx := context.Background()
	builder, err := prompt.NewBuilder(ctx, githubToken, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create builder: %v\n", err)
		os.Exit(1)
	}

	// Generate prompt
	fmt.Fprintf(os.Stderr, "Generating prompt for %s/%s/%s...\n", owner, repo, path)
	promptText, err := builder.BuildPrompt(owner, repo, path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to build prompt: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(promptText), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to write output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Prompt written to %s\n", outputFile)
		fmt.Fprintf(os.Stderr, "Characters: %d (~%d tokens)\n", len(promptText), len(promptText)/4)
	} else {
		fmt.Println(promptText)
		fmt.Fprintf(os.Stderr, "\n---\n")
		fmt.Fprintf(os.Stderr, "Prompt generated successfully!\n")
		fmt.Fprintf(os.Stderr, "Characters: %d (~%d tokens)\n", len(promptText), len(promptText)/4)
	}
}
