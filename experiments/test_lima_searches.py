#!/usr/bin/env python3
"""
Test different GitHub search queries to find Lima templates without minimumLimaVersion
"""

import os
import requests
import time

# Get token from environment
GITHUB_TOKEN = os.environ.get('GITHUB_TOKEN')
if not GITHUB_TOKEN:
    print("Error: GITHUB_TOKEN environment variable not set")
    exit(1)

headers = {
    'Authorization': f'token {GITHUB_TOKEN}',
    'Accept': 'application/vnd.github.v3+json'
}

def search_code(query):
    """Search GitHub code and return count"""
    url = 'https://api.github.com/search/code'
    params = {'q': query, 'per_page': 1}

    try:
        response = requests.get(url, headers=headers, params=params)
        response.raise_for_status()
        data = response.json()
        return data.get('total_count', 0)
    except Exception as e:
        print(f"Error searching: {e}")
        return None

# Test different search strategies
searches = [
    # Original search
    ("minimumLimaVersion extension:yml -repo:lima-vm/lima",
     "Current search (minimumLimaVersion)"),

    # Basic field combinations
    ("images: mounts: extension:yml -repo:lima-vm/lima",
     "images + mounts fields"),

    ("images: provision: extension:yml -repo:lima-vm/lima",
     "images + provision fields"),

    ("images: mounts: provision: extension:yml -repo:lima-vm/lima",
     "images + mounts + provision"),

    # Lima-specific fields
    ("vmType extension:yml -repo:lima-vm/lima",
     "vmType field (Lima-specific)"),

    ("probes: script: extension:yml -repo:lima-vm/lima",
     "probes field (Lima-specific)"),

    ("copyToHost: extension:yml -repo:lima-vm/lima",
     "copyToHost field (Lima-specific)"),

    # Path-based searches
    ("path:lima filename:.yaml -repo:lima-vm/lima",
     "Files in 'lima' directory"),

    ("path:lima images: extension:yml -repo:lima-vm/lima",
     "'lima' path + images field"),

    # Architecture patterns (Lima templates often have arch-specific images)
    ("arch: aarch64 x86_64 images: extension:yml -repo:lima-vm/lima",
     "Multi-arch images"),

    # Container runtime combinations
    ("containerd: images: mounts: extension:yml -repo:lima-vm/lima",
     "containerd + images + mounts"),
]

print("Testing GitHub search queries for Lima templates\n")
print("=" * 80)

for query, description in searches:
    print(f"\n{description}")
    print(f"Query: {query}")

    count = search_code(query)
    if count is not None:
        print(f"Results: {count:,}")

    # Be nice to the API
    time.sleep(2)

print("\n" + "=" * 80)
print("\nRecommendation: Use queries with specific Lima fields (vmType, probes)")
print("or combine common fields (images + mounts + provision) to reduce false positives")
