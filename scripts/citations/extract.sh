#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
    echo "Usage: $0 <paper-key>" >&2
    exit 2
fi

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

key="$1"
paper_file="citations/papers/${key}.md"

if [[ ! -f "$paper_file" ]]; then
    echo "Paper record not found: $paper_file" >&2
    exit 2
fi

bash scripts/citations/_run_agent.sh prompts/citations/03_extractor.md "$key"
