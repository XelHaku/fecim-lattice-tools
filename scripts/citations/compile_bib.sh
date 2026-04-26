#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

out="citations/refs.bib"
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

shopt -s nullglob
papers=(citations/papers/*.md)

{
    printf '%% FeCIM Lattice Tools bibliography\n'
    printf '%% Source records live in citations/papers/*.md.\n'
    printf '%% Regenerate with: bash scripts/citations/compile_bib.sh\n\n'
} > "$tmp"

entry_count=0
for paper in "${papers[@]}"; do
    if awk '
        /^```bibtex[[:space:]]*$/ { inside = 1; next }
        /^```[[:space:]]*$/ && inside { inside = 0; next }
        inside { print }
    ' "$paper" | grep -q '^@'; then
        printf '%% Source: %s\n' "$paper" >> "$tmp"
        awk '
            /^```bibtex[[:space:]]*$/ { inside = 1; next }
            /^```[[:space:]]*$/ && inside { inside = 0; next }
            inside { print }
        ' "$paper" >> "$tmp"
        printf '\n' >> "$tmp"
        entry_count=$((entry_count + 1))
    fi
done

if [[ "$entry_count" -eq 0 ]]; then
    printf '%% No verified entries have been added yet.\n' >> "$tmp"
fi

mv "$tmp" "$out"
trap - EXIT

printf 'Compiled %s (%d entries)\n' "$out" "$entry_count"
