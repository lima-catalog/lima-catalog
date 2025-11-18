#!/bin/bash
# Example usage of the Lima Template Prompt Generator

# Make sure GITHUB_TOKEN is set
if [ -z "$GITHUB_TOKEN" ]; then
    echo "Error: GITHUB_TOKEN environment variable must be set"
    echo "Get a token from: https://github.com/settings/tokens"
    exit 1
fi

# Build the tool
echo "Building prompt-generator..."
go build -o prompt-generator .

# Example 1: Basic usage - Generate prompt for ubuntu template
echo ""
echo "=== Example 1: Basic Ubuntu template ==="
./prompt-generator lima-vm/lima/templates/ubuntu.yaml -output=/tmp/ubuntu-prompt.txt
echo "Prompt saved to /tmp/ubuntu-prompt.txt"
echo ""

# Example 2: Minimal prompt (just template content, no references or README)
echo "=== Example 2: Minimal prompt (no references, no README) ==="
./prompt-generator lima-vm/lima/templates/docker.yaml -no-references -no-readme

# Example 3: Custom context lines
echo ""
echo "=== Example 3: More context around references ==="
./prompt-generator lima-vm/lima/templates/k3s.yaml -context=20 -output=/tmp/k3s-prompt.txt
echo "Prompt with 20 context lines saved to /tmp/k3s-prompt.txt"
echo ""

# Example 4: Using separate flags instead of path notation
echo "=== Example 4: Using separate flags ==="
./prompt-generator -owner=lima-vm -repo=lima -path=templates/alpine.yaml -output=/tmp/alpine-prompt.txt
echo "Prompt saved to /tmp/alpine-prompt.txt"
echo ""

echo "All examples completed!"
echo ""
echo "You can now test these prompts with different LLM models:"
echo "  cat /tmp/ubuntu-prompt.txt | llm -m claude-3-haiku-20240307"
echo "  cat /tmp/ubuntu-prompt.txt | llm -m gpt-4"
echo ""
echo "Or copy-paste them into Claude, ChatGPT, or other LLM interfaces"
