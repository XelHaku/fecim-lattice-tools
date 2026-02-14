#!/usr/bin/env bash
set -euo pipefail

# RG-PAR-05 parity policy (lint gate)
#
# Goal: prevent headless-only physics branches from silently diverging from GUI.
#
# Current implementation is intentionally conservative (heuristic):
# - Detect headless-only conditionals via env flags in non-test physics code.
# - Require a GUI/headless parity test to exist for the owning module.
#
# This gate is cheap, headless-safe, and runs before go test.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

fail=0

require_parity_test() {
  local module="$1"
  local required_path="$2"

  if [[ ! -f "$required_path" ]]; then
    echo "[parity-policy] FAIL: detected headless physics branch in module=$module but missing parity test: $required_path" >&2
    fail=1
  fi
}

# Patterns that indicate headless-only branches.
PATTERN='FECIM_HEADLESS|FECIM_HEADLESS_FAST|FECIM_FYNE_TEST'

# Scan non-test Go files in physics-ish directories.
# (Keep list explicit so we don't accidentally crawl vendor/.)
scan_dirs=(
  "shared/physics"
  "module1-hysteresis/pkg"
  "module4-circuits/pkg"
  "cmd/fecim-lattice-tools"
)

hits=""
for d in "${scan_dirs[@]}"; do
  if [[ -d "$d" ]]; then
    # Exclude *_test.go; we only care about production code paths.
    h=$(grep -R -n -E "$PATTERN" "$d" --include='*.go' --exclude='*_test.go' || true)
    if [[ -n "$h" ]]; then
      hits+="$h"$'\n'
    fi
  fi
done

if [[ -n "$hits" ]]; then
  echo "[parity-policy] detected headless-branch markers:" >&2
  echo "$hits" >&2

  # Required parity tests for known modules.
  require_parity_test "module1-hysteresis" "module1-hysteresis/tests/gui_headless_parity_test.go"
  require_parity_test "module4-circuits" "module4-circuits/pkg/gui/headless_gui_physics_parity_test.go"
fi

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi

echo "[parity-policy] OK"
