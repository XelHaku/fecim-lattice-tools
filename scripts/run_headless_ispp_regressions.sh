#!/usr/bin/env bash
set -euo pipefail

# Required: this regression lane is fully headless (no display stack).
if [[ -n "${DISPLAY:-}" || -n "${WAYLAND_DISPLAY:-}" ]]; then
  echo "[regression] ERROR: DISPLAY/WAYLAND_DISPLAY detected; run this lane fully headless." >&2
  echo "[regression] Hint: unset DISPLAY WAYLAND_DISPLAY" >&2
  exit 1
fi

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${FECIM_REGRESSION_JSON_DIR:-$REPO_ROOT/output/regression}"

mkdir -p "$OUT_DIR"
export FECIM_REGRESSION_JSON_DIR="$OUT_DIR"

echo "[regression] output dir: $FECIM_REGRESSION_JSON_DIR"
echo "[regression] running Preisach + LK headless WRD/ISPP suites"

go test ./module1-hysteresis/pkg/controller \
  -run 'TestHeadlessRegression_WRD_ISPP_(Preisach|LK)$' \
  -count=1 -v | tee "$OUT_DIR/test.log"

echo "[regression] per-material verdicts:" 
grep -E 'VERDICT material=' -n "$OUT_DIR/test.log" || true

echo "[regression] summaries:"
ls -1 "$OUT_DIR"/*.json
