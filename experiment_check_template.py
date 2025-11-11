#!/usr/bin/env python3
"""
Check actual Lima template structure.
"""

import os
import requests

GITHUB_TOKEN = os.environ.get('GITHUB_TOKEN', '')
HEADERS = {
    'Accept': 'application/vnd.github.v3+json',
}
if GITHUB_TOKEN:
    HEADERS['Authorization'] = f'token {GITHUB_TOKEN}'

BASE_URL = 'https://api.github.com'


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


def main():
    print("=" * 70)
    print("Checking Lima Template Content")
    print("=" * 70)
    print()

    # Check a few template files from lima-vm/lima
    template_files = [
        'templates/default.yaml',
        'templates/ubuntu.yaml',
        'templates/k3s.yaml'
    ]

    for template_path in template_files:
        print(f"File: {template_path}")
        print("-" * 70)
        try:
            content = get_file_content('lima-vm', 'lima', template_path)

            # Check for minimumLimaVersion
            if 'minimumLimaVersion' in content:
                print("✓ Contains 'minimumLimaVersion'")
            else:
                print("✗ Does NOT contain 'minimumLimaVersion'")

            # Show first 30 lines
            lines = content.split('\n')
            for i, line in enumerate(lines[:30], 1):
                print(f"{i:3}: {line}")

            print()
            print()

        except Exception as e:
            print(f"Error: {e}")
            print()

    # Now check one of the 57 files from our original search
    print("=" * 70)
    print("Now checking files from original search (outside lima-vm/lima)")
    print("=" * 70)
    print()

    # These are from our original 57 results
    external_files = [
        ('felix-kaestner', 'lima-templates', 'debian.yml'),
        ('annie444', 'utils-util', 'lima-template.yaml'),
    ]

    for owner, repo, path in external_files:
        try:
            print(f"File: {owner}/{repo}/{path}")
            print("-" * 70)
            content = get_file_content(owner, repo, path)

            # Check for minimumLimaVersion
            if 'minimumLimaVersion' in content:
                print("✓ Contains 'minimumLimaVersion'")
                for line in content.split('\n'):
                    if 'minimumLimaVersion' in line:
                        print(f"  Line: {line.strip()}")
            else:
                print("✗ Does NOT contain 'minimumLimaVersion'")

            # Show first 25 lines
            lines = content.split('\n')
            for i, line in enumerate(lines[:25], 1):
                print(f"{i:3}: {line}")

            print()
            print()

        except Exception as e:
            print(f"Error: {e}")
            print()


if __name__ == '__main__':
    main()
