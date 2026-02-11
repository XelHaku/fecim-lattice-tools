#!/usr/bin/env bash
set -euo pipefail

# scripts/calib-guard.sh
#
# Fail if any calibration JSON under:
#   cmd/fecim-lattice-tools/data/calibrations/*.json
# changes relative to the PR base (or push "before" SHA), unless explicitly allowed.
#
# Bypass intentionally:
#   ALLOW_CALIBRATION_UPDATES=1 scripts/calib-guard.sh
#
# Optional override:
#   CALIB_GUARD_BASE=<git-ref> scripts/calib-guard.sh

CALIB_DIR="cmd/fecim-lattice-tools/data/calibrations"

if [[ "${ALLOW_CALIBRATION_UPDATES:-}" == "1" ]]; then
  echo "calib-guard: ALLOW_CALIBRATION_UPDATES=1 set; skipping calibration drift check"
  exit 0
fi

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "calib-guard: not in a git repository" >&2
  exit 2
fi

base_ref="${CALIB_GUARD_BASE:-}"

# GitHub Actions PRs expose GITHUB_BASE_REF.
if [[ -z "$base_ref" && -n "${GITHUB_BASE_REF:-}" ]]; then
  base_ref="origin/${GITHUB_BASE_REF}"
fi

# GitHub Actions push events expose a 'before' SHA in the event payload.
if [[ -z "$base_ref" && "${GITHUB_EVENT_NAME:-}" == "push" && -n "${GITHUB_EVENT_PATH:-}" && -f "${GITHUB_EVENT_PATH:-}" ]]; then
  before_sha="$(python3 - <<'PY'
import json, os
p = os.environ.get('GITHUB_EVENT_PATH')
with open(p,'r',encoding='utf-8') as f:
  evt=json.load(f)
print(evt.get('before',''))
PY
  )"
  if [[ -n "$before_sha" && "$before_sha" != "0000000000000000000000000000000000000000" ]]; then
    base_ref="$before_sha"
  fi
fi

# Local/dev fallback: compare against HEAD~1 (best-effort).
if [[ -z "$base_ref" ]]; then
  if git rev-parse --verify HEAD~1 >/dev/null 2>&1; then
    base_ref="HEAD~1"
  else
    # Single-commit repo; nothing meaningful to diff.
    echo "calib-guard: no base ref available; skipping (single-commit history?)"
    exit 0
  fi
fi

if ! git rev-parse --verify "$base_ref" >/dev/null 2>&1; then
  echo "calib-guard: base ref '$base_ref' not found (did you fetch full history?)" >&2
  echo "calib-guard: hint: in CI set actions/checkout fetch-depth: 0" >&2
  exit 2
fi

# Check for any .json changes in the calibration directory.
changed="$(git diff --name-only --diff-filter=ACMRTUXB "$base_ref...HEAD" -- "$CALIB_DIR" | grep -E '\.json$' || true)"

if [[ -n "$changed" ]]; then
  echo "calib-guard: calibration JSON changes detected under $CALIB_DIR:" >&2
  echo "$changed" | sed 's/^/  - /' >&2
  echo >&2
  echo "calib-guard: This repo treats calibration JSON as drift-sensitive." >&2
  echo "calib-guard: If this update is intentional, re-run with:" >&2
  echo "  ALLOW_CALIBRATION_UPDATES=1 scripts/calib-guard.sh" >&2
  echo >&2
  echo "calib-guard: In GitHub Actions, set a repository variable 'ALLOW_CALIBRATION_UPDATES' to '1'" >&2
  echo "calib-guard: (or set the env var for the job) to allow intentional updates." >&2
  exit 1
fi

echo "calib-guard: OK (no calibration JSON changes detected)"
