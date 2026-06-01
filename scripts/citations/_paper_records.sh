#!/usr/bin/env bash
# Shared helpers for locating citation paper records.
# Source this file from scripts/citations/*.sh after changing to repo root.

citation_paper_records() {
    find citations/papers -type f -name '*.md' | LC_ALL=C sort
}

citation_paper_record_path() {
    if [[ $# -ne 1 ]]; then
        echo "usage: citation_paper_record_path <paper-key>" >&2
        return 2
    fi

    local key="$1"
    local path
    path="$(find citations/papers -type f -name "${key}.md" | LC_ALL=C sort | head -n 1)"
    if [[ -z "$path" ]]; then
        return 1
    fi
    printf '%s\n' "$path"
}
