#!/bin/bash
# Test different GitHub search queries to find Lima templates without minimumLimaVersion

echo "Testing GitHub search queries for Lima templates"
echo "================================================================================"

# Function to search and display results
test_search() {
    local description="$1"
    local query="$2"

    echo ""
    echo "$description"
    echo "Query: $query"

    # Use gh CLI to search
    count=$(gh api search/code -X GET -f q="$query" --jq '.total_count' 2>/dev/null)

    if [ $? -eq 0 ]; then
        echo "Results: $count"
    else
        echo "Results: Error querying GitHub"
    fi

    # Rate limiting
    sleep 2
}

# Original search
test_search \
    "Current search (minimumLimaVersion)" \
    "minimumLimaVersion extension:yml -repo:lima-vm/lima"

# Basic field combinations
test_search \
    "images + mounts fields" \
    "images: mounts: extension:yml -repo:lima-vm/lima"

test_search \
    "images + provision fields" \
    "images: provision: extension:yml -repo:lima-vm/lima"

test_search \
    "images + mounts + provision" \
    "images: mounts: provision: extension:yml -repo:lima-vm/lima"

# Lima-specific fields
test_search \
    "vmType field (Lima-specific)" \
    "vmType extension:yml -repo:lima-vm/lima"

test_search \
    "probes field (Lima-specific)" \
    "probes: script: extension:yml -repo:lima-vm/lima"

test_search \
    "copyToHost field (Lima-specific)" \
    "copyToHost: extension:yml -repo:lima-vm/lima"

# Path-based searches
test_search \
    "Files in 'lima' directory" \
    "path:lima filename:.yaml -repo:lima-vm/lima"

test_search \
    "'lima' path + images field" \
    "path:lima images: extension:yml -repo:lima-vm/lima"

# Multi-arch pattern
test_search \
    "Multi-arch images" \
    "arch: aarch64 x86_64 images: extension:yml -repo:lima-vm/lima"

# Container runtime
test_search \
    "containerd + images + mounts" \
    "containerd: images: mounts: extension:yml -repo:lima-vm/lima"

echo ""
echo "================================================================================"
echo ""
echo "Recommendation: Use queries with specific Lima fields (vmType, probes)"
echo "or combine common fields (images + mounts + provision) to reduce false positives"
