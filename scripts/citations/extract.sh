#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
    echo "Usage: $0 <paper-key>" >&2
    exit 2
fi

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

source scripts/citations/_paper_records.sh

key="$1"
if ! paper_file="$(citation_paper_record_path "$key")"; then
    echo "Paper record not found: citations/papers/**/${key}.md" >&2
    exit 2
fi

bash scripts/citations/_run_agent.sh tools/prompts/citations/03_extractor.md "$key"
