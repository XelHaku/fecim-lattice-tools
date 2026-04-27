#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <prompt-file> [input...]" >&2
    exit 2
fi

prompt_file="$1"
shift || true

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

if [[ ! -f "$prompt_file" ]]; then
    echo "Prompt file not found: $prompt_file" >&2
    exit 2
fi

if [[ -z "${CITATION_AGENT_CMD:-}" ]]; then
    cat >&2 <<'EOF'
CITATION_AGENT_CMD is not set.

This command intentionally fails closed. Set CITATION_AGENT_CMD to a
reviewed local or remote agent command that reads a complete prompt from
standard input before using agent-backed citation scripts.
EOF
    exit 2
fi

{
    printf 'Repository: %s\n' "$repo_root"
    printf 'Date: %s\n\n' "$(date +%F)"
    cat "$prompt_file"
    printf '\n\nInput:\n'
    printf '%s\n' "$*"
} | ${CITATION_AGENT_CMD}
