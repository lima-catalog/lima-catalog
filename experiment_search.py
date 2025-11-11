#!/usr/bin/env python3
"""
Experiment: Search GitHub for Lima templates to estimate scale.

This script searches for YAML files containing 'minimumLimaVersion'
to understand how many templates exist on GitHub.
"""

import os
import requests
import time
from datetime import datetime
from typing import Dict, List, Set

# GitHub API configuration
GITHUB_TOKEN = os.environ.get('GITHUB_TOKEN', '')
HEADERS = {
    'Accept': 'application/vnd.github.v3+json',
}
if GITHUB_TOKEN:
    HEADERS['Authorization'] = f'token {GITHUB_TOKEN}'

BASE_URL = 'https://api.github.com'


def check_rate_limit() -> Dict:
    """Check current GitHub API rate limit status."""
    response = requests.get(f'{BASE_URL}/rate_limit', headers=HEADERS)
    response.raise_for_status()
    return response.json()


def search_code(query: str, max_results: int = 1000) -> List[Dict]:
    """
    Search GitHub code with pagination.

    Note: GitHub Code Search has a rate limit of 30 requests/minute
    and returns max 1000 results per query.
    """
    results = []
    page = 1
    per_page = 100  # Max allowed by GitHub

    while len(results) < max_results:
        url = f'{BASE_URL}/search/code'
        params = {
            'q': query,
            'per_page': per_page,
            'page': page
        }

        print(f"Fetching page {page}...", end=' ')
        response = requests.get(url, headers=HEADERS, params=params)

        if response.status_code == 403:
            print("Rate limit hit!")
            break
        elif response.status_code == 422:
            print("Query too broad or no more results")
            break

        response.raise_for_status()
        data = response.json()

        items = data.get('items', [])
        print(f"Found {len(items)} items (total so far: {len(results) + len(items)})")

        if not items:
            break

        results.extend(items)

        # Check if we've hit the last page
        total_count = data.get('total_count', 0)
        if len(results) >= total_count or len(results) >= max_results:
            break

        page += 1

        # Respect rate limits - GitHub allows 30 code search requests/minute
        # So we wait 2 seconds between requests
        time.sleep(2)

    return results


def analyze_results(results: List[Dict]) -> Dict:
    """Analyze search results to get statistics."""
    repos: Set[str] = set()
    orgs: Set[str] = set()
    paths: Set[str] = set()
    lima_vm_templates = 0
    fork_templates = 0

    for item in results:
        repo_full_name = item['repository']['full_name']
        repos.add(repo_full_name)

        owner = item['repository']['owner']['login']
        orgs.add(owner)

        path = item['path']
        paths.add(f"{repo_full_name}/{path}")

        # Check if from lima-vm/lima
        if repo_full_name.startswith('lima-vm/lima'):
            lima_vm_templates += 1

        # Check if repo is a fork
        if item['repository'].get('fork', False):
            fork_templates += 1

    return {
        'total_files': len(results),
        'unique_repos': len(repos),
        'unique_owners': len(orgs),
        'unique_templates': len(paths),
        'lima_vm_templates': lima_vm_templates,
        'fork_templates': fork_templates,
        'non_lima_vm_templates': len(results) - lima_vm_templates,
        'repos': sorted(repos),
        'owners': sorted(orgs)
    }


def main():
    print("=" * 70)
    print("Lima Template Catalog - GitHub Search Experiment")
    print("=" * 70)
    print()

    # Check if we have a GitHub token
    if not GITHUB_TOKEN:
        print("⚠️  WARNING: No GITHUB_TOKEN found in environment!")
        print("Code search requires authentication.")
        print()
        print("To run this experiment, please set GITHUB_TOKEN:")
        print("  export GITHUB_TOKEN=your_github_token")
        print()
        print("You can create a token at:")
        print("  https://github.com/settings/tokens")
        print("  (No special scopes needed for public repo search)")
        print()
        return

    # Check rate limit first
    print("Checking GitHub API rate limit...")
    try:
        rate_limit = check_rate_limit()
        core_limit = rate_limit['resources']['core']
        search_limit = rate_limit['resources']['search']

        print(f"Core API: {core_limit['remaining']}/{core_limit['limit']} remaining")
        print(f"Search API: {search_limit['remaining']}/{search_limit['limit']} remaining")
        print(f"Reset at: {datetime.fromtimestamp(search_limit['reset'])}")
        print()

        if search_limit['remaining'] < 10:
            print("WARNING: Low rate limit remaining. Consider waiting.")
            return
    except requests.exceptions.HTTPError as e:
        print(f"Error checking rate limit: {e}")
        print("Continuing anyway...")
        print()

    # Search for Lima templates
    # Query: Find YAML files containing 'minimumLimaVersion'
    # Exclude lima-vm/lima repository
    query = 'minimumLimaVersion extension:yml OR extension:yaml -repo:lima-vm/lima'

    print(f"Search query: {query}")
    print()
    print("Searching GitHub (this may take a few minutes)...")
    print()

    results = search_code(query, max_results=1000)

    print()
    print("=" * 70)
    print("Results Analysis")
    print("=" * 70)
    print()

    if not results:
        print("No results found!")

        # Try a simpler query to see if there are any templates at all
        print()
        print("Trying simpler query (including lima-vm/lima)...")
        simple_query = 'minimumLimaVersion extension:yml'
        simple_results = search_code(simple_query, max_results=100)

        if simple_results:
            print(f"Found {len(simple_results)} templates with simpler query")
            stats = analyze_results(simple_results)
            print(f"From {stats['unique_repos']} repositories")
        return

    stats = analyze_results(results)

    print(f"Total files found: {stats['total_files']}")
    print(f"Unique templates: {stats['unique_templates']}")
    print(f"Unique repositories: {stats['unique_repos']}")
    print(f"Unique owners: {stats['unique_owners']}")
    print()
    print(f"Templates from lima-vm/lima: {stats['lima_vm_templates']}")
    print(f"Templates from forks: {stats['fork_templates']}")
    print(f"Non-lima-vm templates: {stats['non_lima_vm_templates']}")
    print()

    # Show some example repositories
    print("Example repositories (first 20):")
    for repo in stats['repos'][:20]:
        print(f"  - {repo}")

    if len(stats['repos']) > 20:
        print(f"  ... and {len(stats['repos']) - 20} more")

    print()
    print("Example owners (first 20):")
    for owner in stats['owners'][:20]:
        print(f"  - {owner}")

    if len(stats['owners']) > 20:
        print(f"  ... and {len(stats['owners']) - 20} more")

    print()
    print("=" * 70)
    print("Conclusion")
    print("=" * 70)
    print()
    print(f"✓ Found approximately {stats['unique_templates']} unique Lima templates")
    print(f"✓ Spread across {stats['unique_repos']} repositories")
    print(f"✓ From {stats['unique_owners']} different owners")
    print()
    print("This scale is very manageable for our JSON Lines approach!")
    print("Each file will be small and diffs will be minimal.")
    print()

    # Check final rate limit
    print("Checking final rate limit...")
    rate_limit = check_rate_limit()
    search_limit = rate_limit['resources']['search']
    print(f"Search API: {search_limit['remaining']}/{search_limit['limit']} remaining")
    print()


if __name__ == '__main__':
    main()
