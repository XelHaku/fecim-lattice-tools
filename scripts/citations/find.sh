#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <topic>" >&2
    exit 2
fi

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

bash scripts/citations/_run_agent.sh prompts/citations/01_fetcher.md "Discover candidate sources for topic only; do not create files unless explicitly requested: $*"
