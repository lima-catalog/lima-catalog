#!/usr/bin/env python3
"""
Experiment: Investigate GitHub fork search behavior.

GitHub Code Search has special behavior for forks - by default it doesn't
search in forks unless they have more stars than the parent. This experiment
investigates different approaches to find templates in forks.
"""

import os
import requests
import time
from typing import Dict, List

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


def search_code(query: str, max_results: int = 100) -> List[Dict]:
    """Search GitHub code."""
    results = []
    page = 1
    per_page = 100

    while len(results) < max_results:
        url = f'{BASE_URL}/search/code'
        params = {
            'q': query,
            'per_page': per_page,
            'page': page
        }

        response = requests.get(url, headers=HEADERS, params=params)

        if response.status_code == 403:
            print(f"Rate limit hit after {len(results)} results")
            break
        elif response.status_code == 422:
            print(f"Query error or no more results")
            break

        response.raise_for_status()
        data = response.json()

        items = data.get('items', [])
        if not items:
            break

        results.extend(items)

        total_count = data.get('total_count', 0)
        if len(results) >= total_count or len(results) >= max_results:
            break

        page += 1
        time.sleep(2)  # Respect rate limits

    return results


def get_repo_info(owner: str, repo: str) -> Dict:
    """Get repository information."""
    url = f'{BASE_URL}/repos/{owner}/{repo}'
    response = requests.get(url, headers=HEADERS)
    response.raise_for_status()
    return response.json()


def list_forks(owner: str, repo: str, max_forks: int = 100) -> List[Dict]:
    """List forks of a repository."""
    forks = []
    page = 1
    per_page = 100

    while len(forks) < max_forks:
        url = f'{BASE_URL}/repos/{owner}/{repo}/forks'
        params = {
            'per_page': per_page,
            'page': page,
            'sort': 'stargazers'  # Get most popular forks first
        }

        response = requests.get(url, headers=HEADERS, params=params)
        response.raise_for_status()

        items = response.json()
        if not items:
            break

        forks.extend(items)

        if len(items) < per_page:
            break

        page += 1
        time.sleep(0.5)  # Be nice to the API

    return forks


def main():
    print("=" * 70)
    print("Fork Search Investigation")
    print("=" * 70)
    print()

    if not GITHUB_TOKEN:
        print("ERROR: GITHUB_TOKEN not set")
        return

    # Experiment 1: Search in lima-vm/lima itself
    print("Experiment 1: How many templates in lima-vm/lima?")
    print("-" * 70)
    query = 'minimumLimaVersion extension:yml repo:lima-vm/lima'
    print(f"Query: {query}")

    results = search_code(query, max_results=200)
    print(f"Found {len(results)} template files in lima-vm/lima")

    # Look at paths to understand structure
    if results:
        paths = [r['path'] for r in results[:10]]
        print(f"Example paths:")
        for path in paths:
            print(f"  {path}")
    print()

    # Experiment 2: Try to search a specific fork
    print("Experiment 2: Check lima-vm/lima fork information")
    print("-" * 70)

    print("Getting lima-vm/lima repo info...")
    lima_repo = get_repo_info('lima-vm', 'lima')
    print(f"Forks: {lima_repo['forks_count']}")
    print(f"Stars: {lima_repo['stargazers_count']}")
    print()

    # Get some forks
    print("Fetching first 30 forks (sorted by stars)...")
    forks = list_forks('lima-vm', 'lima', max_forks=30)
    print(f"Retrieved {len(forks)} forks")
    print()

    if forks:
        print("Top 10 forks by stars:")
        for fork in forks[:10]:
            print(f"  {fork['full_name']}: {fork['stargazers_count']} stars, "
                  f"updated {fork['updated_at']}")
    print()

    # Experiment 3: Try searching in a specific fork
    if forks:
        print("Experiment 3: Search in top fork")
        print("-" * 70)
        top_fork = forks[0]
        fork_owner = top_fork['owner']['login']
        fork_name = top_fork['name']

        query = f'minimumLimaVersion extension:yml repo:{fork_owner}/{fork_name}'
        print(f"Query: {query}")

        fork_results = search_code(query, max_results=50)
        print(f"Found {len(fork_results)} templates in {fork_owner}/{fork_name}")
        print()

    # Experiment 4: Test if code search includes forks by default
    print("Experiment 4: Does code search include forks by default?")
    print("-" * 70)
    print("According to GitHub docs:")
    print("  'Forks are only searchable if the fork has more stars than")
    print("   the parent repository.'")
    print()
    print(f"lima-vm/lima has {lima_repo['stargazers_count']} stars")
    if forks:
        forks_with_more_stars = [f for f in forks
                                  if f['stargazers_count'] > lima_repo['stargazers_count']]
        print(f"Forks with MORE stars: {len(forks_with_more_stars)}")
        if forks_with_more_stars:
            for fork in forks_with_more_stars:
                print(f"  {fork['full_name']}: {fork['stargazers_count']} stars")
    print()

    # Conclusion
    print("=" * 70)
    print("Conclusions")
    print("=" * 70)
    print()
    print("✗ GitHub Code Search does NOT include forks by default")
    print("✗ Forks are only searchable if they have more stars than parent")
    print(f"✗ With {lima_repo['stargazers_count']} stars, lima-vm/lima excludes most forks")
    print()
    print("✓ We can enumerate forks using the /repos/:owner/:repo/forks API")
    print("✓ We can then check each fork individually")
    print()
    print("Recommended approach:")
    print("1. Use code search for non-fork repositories")
    print("2. Use /repos/lima-vm/lima/forks API to list all forks")
    print("3. For each fork, check if templates directory exists and differs from parent")
    print("4. Only include modified templates from forks")
    print()

    # Check rate limit
    rate_limit = check_rate_limit()
    core_limit = rate_limit['resources']['core']
    print(f"Rate limit used: {5000 - core_limit['remaining']}/5000")


if __name__ == '__main__':
    main()
