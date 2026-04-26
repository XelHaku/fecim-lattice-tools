#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <query>" >&2
    exit 2
fi

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

query="$*"
matches=0

echo "Citation matches for: $query"
echo

for path in citations/facts.md citations/disputed.md citations/papers/*.md; do
    [[ -e "$path" ]] || continue
    if grep -qi -- "$query" "$path"; then
        matches=$((matches + 1))
        echo "== $path =="
        grep -in -- "$query" "$path" | head -5 || true
        echo
    fi
done

if [[ "$matches" -eq 0 ]]; then
    echo "No citation records matched."
fi
