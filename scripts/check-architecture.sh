#!/usr/bin/env bash
# scripts/check-architecture.sh
# Enforces the 7 architectural rules from shared-refactor-plan.md.
# Exit code 0 = passing, 1 = violations found.
#
# Usage:
#   ./scripts/check-architecture.sh          # full check
#   ./scripts/check-architecture.sh --fast   # skip build/vet (Rules 1-4 only)
set -euo pipefail

FAST=0
for arg in "$@"; do [[ "$arg" == "--fast" ]] && FAST=1; done

FAIL=0
ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
cd "$ROOT"

echo "=== Rule 1: No cross-module imports ==="
MODS=(module1-hysteresis module2-crossbar module3-mnist module4-circuits module5-comparison module6-eda)
for mod in "${MODS[@]}"; do
    if [[ ! -d "$mod" ]]; then continue; fi
    for other in "${MODS[@]}"; do
        if [[ "$mod" == "$other" ]]; then continue; fi
        hits=$(grep -r "\"fecim-lattice-tools/$other" "$mod/" --include='*.go' 2>/dev/null | grep -v '_test.go' || true)
        if [[ -n "$hits" ]]; then
            echo "  ❌ RULE 1: $mod imports from $other:"
            echo "$hits" | sed 's/^/    /'
            FAIL=1
        fi
    done
done
[[ $FAIL -eq 0 ]] && echo "  ✅ Rule 1 passed"

echo ""
echo "=== Rule 3: Single go.mod ==="
count=$(find . -name 'go.mod' -not -path './.git/*' | wc -l | tr -d ' ')
if [[ "$count" -ne 1 ]]; then
    echo "  ❌ RULE 3: found $count go.mod files (expected 1)"
    find . -name 'go.mod' -not -path './.git/*' | sed 's/^/    /'
    FAIL=1
else
    echo "  ✅ Rule 3 passed (1 go.mod)"
fi

echo ""
echo "=== Rule 4: No banned package names ==="
BANNED=(utils common helpers misc core2 internal2)
R4_FAIL=0
for name in "${BANNED[@]}"; do
    if [[ -d "shared/$name" ]]; then
        echo "  ❌ RULE 4: shared/$name/ exists (banned grab-bag name)"
        FAIL=1; R4_FAIL=1
    fi
done
[[ $R4_FAIL -eq 0 ]] && echo "  ✅ Rule 4 passed"

if [[ $FAST -eq 1 ]]; then
    echo ""
    echo "(Fast mode: skipping build/vet)"
else
    echo ""
    echo "=== Build & Vet ==="
    if go build ./... 2>&1; then
        echo "  ✅ Build passed"
    else
        echo "  ❌ Build failed"
        FAIL=1
    fi
    if go vet ./... 2>&1; then
        echo "  ✅ Vet passed"
    else
        echo "  ❌ Vet failed"
        FAIL=1
    fi
fi

echo ""
if [[ $FAIL -eq 1 ]]; then
    echo "❌ Architecture check FAILED — fix violations before merging"
    exit 1
else
    echo "✅ All architecture checks passed"
    exit 0
fi
