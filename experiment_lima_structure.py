#!/usr/bin/env python3
"""
Experiment: Understand the structure of lima-vm/lima templates.

Investigate why we found 0 templates in lima-vm/lima itself.
"""

import os
import requests
import time

GITHUB_TOKEN = os.environ.get('GITHUB_TOKEN', '')
HEADERS = {
    'Accept': 'application/vnd.github.v3+json',
}
if GITHUB_TOKEN:
    HEADERS['Authorization'] = f'token {GITHUB_TOKEN}'

BASE_URL = 'https://api.github.com'


def get_repo_contents(owner: str, repo: str, path: str = ''):
    """Get contents of a directory in a repo."""
    url = f'{BASE_URL}/repos/{owner}/{repo}/contents/{path}'
    response = requests.get(url, headers=HEADERS)
    response.raise_for_status()
    return response.json()


def get_file_content(owner: str, repo: str, path: str):
    """Get content of a specific file."""
    url = f'{BASE_URL}/repos/{owner}/{repo}/contents/{path}'
    response = requests.get(url, headers=HEADERS)
    response.raise_for_status()
    data = response.json()

    if data.get('encoding') == 'base64':
        import base64
        return base64.b64decode(data['content']).decode('utf-8')

    return data.get('content', '')


def search_code_simple(query: str):
    """Simple code search."""
    url = f'{BASE_URL}/search/code'
    params = {'q': query, 'per_page': 10}
    response = requests.get(url, headers=HEADERS, params=params)
    response.raise_for_status()
    return response.json()


def main():
    print("=" * 70)
    print("Lima Repository Structure Investigation")
    print("=" * 70)
    print()

    if not GITHUB_TOKEN:
        print("ERROR: GITHUB_TOKEN not set")
        return

    # Check root directory
    print("Step 1: Check root directory structure")
    print("-" * 70)
    try:
        contents = get_repo_contents('lima-vm', 'lima')
        dirs = [item['name'] for item in contents if item['type'] == 'dir']
        print(f"Directories in root: {', '.join(dirs)}")
        print()
    except Exception as e:
        print(f"Error: {e}")
        print()

    # Check if there's a templates or examples directory
    print("Step 2: Check for templates/examples directory")
    print("-" * 70)
    for dir_name in ['templates', 'examples', 'docs/templates', '.lima']:
        try:
            contents = get_repo_contents('lima-vm', 'lima', dir_name)
            yaml_files = [item['name'] for item in contents
                         if item['name'].endswith(('.yaml', '.yml'))]
            print(f"{dir_name}: {len(yaml_files)} YAML files")
            if yaml_files:
                print(f"  Files: {', '.join(yaml_files[:5])}")
                if len(yaml_files) > 5:
                    print(f"  ... and {len(yaml_files) - 5} more")
            print()
        except requests.exceptions.HTTPError as e:
            if e.response.status_code == 404:
                print(f"{dir_name}: not found")
            else:
                print(f"{dir_name}: error {e.response.status_code}")
        except Exception as e:
            print(f"{dir_name}: error {e}")

    # Check a specific template file
    print("Step 3: Examine a template file")
    print("-" * 70)
    try:
        # Try to find any yaml file in examples
        contents = get_repo_contents('lima-vm', 'lima', 'examples')
        yaml_files = [item for item in contents
                     if item['name'].endswith(('.yaml', '.yml'))]

        if yaml_files:
            # Get first yaml file
            first_file = yaml_files[0]
            print(f"Examining: examples/{first_file['name']}")
            print()

            content = get_file_content('lima-vm', 'lima', f"examples/{first_file['name']}")

            # Check if it contains minimumLimaVersion
            if 'minimumLimaVersion' in content:
                print("✓ Contains 'minimumLimaVersion'")
                # Extract the line
                for line in content.split('\n'):
                    if 'minimumLimaVersion' in line:
                        print(f"  {line.strip()}")
                        break
            else:
                print("✗ Does NOT contain 'minimumLimaVersion'")

            print()
            print("First 20 lines of file:")
            print("-" * 40)
            for i, line in enumerate(content.split('\n')[:20], 1):
                print(f"{i:3}: {line}")
            print()
    except Exception as e:
        print(f"Error: {e}")
        print()

    # Try different search queries
    print("Step 4: Try different search queries")
    print("-" * 70)

    queries = [
        'minimumLimaVersion repo:lima-vm/lima',
        'lima version repo:lima-vm/lima extension:yaml',
        'images: repo:lima-vm/lima extension:yaml path:examples',
        'location: repo:lima-vm/lima extension:yaml',
    ]

    for query in queries:
        try:
            print(f"Query: {query}")
            results = search_code_simple(query)
            count = results.get('total_count', 0)
            print(f"  Results: {count}")
            if count > 0:
                items = results.get('items', [])
                if items:
                    print(f"  Example: {items[0]['path']}")
            print()
            time.sleep(2)  # Rate limit
        except Exception as e:
            print(f"  Error: {e}")
            print()

    print("=" * 70)
    print("Summary")
    print("=" * 70)
    print()
    print("Investigation complete. Check output above to understand:")
    print("1. Where Lima templates are stored")
    print("2. What keywords to search for")
    print("3. Whether minimumLimaVersion is used in templates")


if __name__ == '__main__':
    main()
