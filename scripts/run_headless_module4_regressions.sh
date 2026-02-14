#!/usr/bin/env bash
set -euo pipefail

# Required: this regression lane is fully headless (no display stack).
if [[ -n "${DISPLAY:-}" || -n "${WAYLAND_DISPLAY:-}" ]]; then
  echo "[m4-regression] ERROR: DISPLAY/WAYLAND_DISPLAY detected; run this lane fully headless." >&2
  echo "[m4-regression] Hint: unset DISPLAY WAYLAND_DISPLAY" >&2
  exit 1
fi

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${FECIM_M4_REGRESSION_JSON_DIR:-$REPO_ROOT/output/regression/module4}"

mkdir -p "$OUT_DIR"
export FECIM_M4_REGRESSION_JSON_DIR="$OUT_DIR"

echo "[m4-regression] output dir: $FECIM_M4_REGRESSION_JSON_DIR"
echo "[m4-regression] running arraysim invariants + GUI/headless physics parity"

go test ./module4-circuits/pkg/arraysim \
  -run 'TestCurrentValidation_(SingleCell_Exact50uA|2x2_AnalyticParallelRows|4x4_KCLAndBounds|OperationInvariants)$' \
  -count=1 -v

go test ./module4-circuits/pkg/gui \
  -run 'Test(HeadlessPhysicsParity_GUIVsHeadless_ReadComputeWriteStep_MaterialAware|ReadCoupling_DefaultsToTierA|ReadChain_EndToEndKnownConductanceToADCCode|DeviceState_ProgramLevelFromCoupledVoltage_UsesActualCoupledVoltage)$' \
  -count=1 -v | tee "$OUT_DIR/gui_test.log"

echo "[m4-regression] per-material verdicts (if present):"
grep -E 'VERDICT material=' -n "$OUT_DIR/gui_test.log" || true

echo "[m4-regression] complete"
