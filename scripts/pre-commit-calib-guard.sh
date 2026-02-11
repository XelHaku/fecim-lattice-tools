#!/usr/bin/env bash
set -euo pipefail

# Optional pre-commit hook template.
#
# Install (manual):
#   cp scripts/pre-commit-calib-guard.sh .git/hooks/pre-commit
#   chmod +x .git/hooks/pre-commit
#
# Bypass intentionally:
#   git commit --no-verify
# or (less recommended):
#   ALLOW_CALIBRATION_UPDATES=1 git commit -m "..."

CALIB_DIR="cmd/fecim-lattice-tools/data/calibrations"

if [[ "${ALLOW_CALIBRATION_UPDATES:-}" == "1" ]]; then
  exit 0
fi

changed="$(git diff --cached --name-only --diff-filter=ACMRTUXB -- "$CALIB_DIR" | grep -E '\.json$' || true)"

if [[ -n "$changed" ]]; then
  echo "pre-commit: calibration JSON changes detected under $CALIB_DIR:" >&2
  echo "$changed" | sed 's/^/  - /' >&2
  echo >&2
  echo "pre-commit: Blocking commit by default to avoid unintentional calibration drift." >&2
  echo "pre-commit: If intentional, use 'git commit --no-verify' or set ALLOW_CALIBRATION_UPDATES=1." >&2
  exit 1
fi
