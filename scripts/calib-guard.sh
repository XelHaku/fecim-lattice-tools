#!/usr/bin/env bash
set -euo pipefail

# scripts/calib-guard.sh
#
# CI guard for calibration data drift.
# Fails if there are uncommitted JSON changes under:
#   cmd/fecim-lattice-tools/data/calibrations/
#
# Intended use (CI): run after tests/codegen/simulations to ensure jobs did not
# mutate protected calibration files.

CALIB_DIR="cmd/fecim-lattice-tools/data/calibrations"

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "calib-guard: not in a git repository" >&2
  exit 2
fi

collect_changes() {
  {
    # Unstaged changes
    git diff --name-only -- "$CALIB_DIR"
    # Staged changes not yet committed
    git diff --cached --name-only -- "$CALIB_DIR"
    # Untracked files
    git ls-files --others --exclude-standard -- "$CALIB_DIR"
  } | grep -E '\.json$' | sort -u || true
}

changed="$(collect_changes)"

if [[ -n "$changed" ]]; then
  echo "calib-guard: FAIL - uncommitted calibration JSON changes detected under $CALIB_DIR:" >&2
  echo "$changed" | sed 's/^/  - /' >&2
  echo >&2
  echo "calib-guard: Commit/revert these files before CI finishes." >&2
  echo "calib-guard: Intentional calibration updates must include evidence links in the commit message" >&2
  echo "calib-guard: (see CONTRIBUTING.md: Calibration update policy)." >&2
  exit 1
fi

echo "calib-guard: OK (no uncommitted calibration JSON changes)"
