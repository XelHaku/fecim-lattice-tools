#!/bin/bash
set -euo pipefail

MODE="full"
JSON=false

for arg in "$@"; do
  case $arg in
    --fast) MODE="fast" ;;
    --full) MODE="full" ;;
    --json) JSON=true ;;
  esac
done

echo "=== Module 1 Hysteresis Automation Suite (mode: $MODE) ==="
echo "Commit: $(git rev-parse HEAD)"
echo "Date: $(date -Iseconds)"

PASS=0
FAIL=0

run_tests() {
  local label="$1"
  local pattern="$2"
  local pkg="$3"

  echo "--- $label ---"
  if go test -short -count=1 -run "$pattern" "$pkg" -timeout 120s 2>&1 | tail -3; then
    PASS=$((PASS+1))
  else
    FAIL=$((FAIL+1))
    echo "FAILED: $label"
  fi
}

# Fast gate: core physics + regression + basic ISPP
run_tests "preisach-core" "^TestPreisach" "./module1-hysteresis/pkg/ferroelectric/"
run_tests "physics-regression" "^TestPhysicsRegressionCurves$" "./validation/..."
run_tests "ispp-convergence" "^TestISPPConverges_" "./module1-hysteresis/pkg/controller/"
run_tests "determinism" "^TestDeterminism_" "./validation/..."

if [ "$MODE" = "full" ]; then
  # Full: LK solver + broad M1 packages + discontinuity + literature + headless ISPP
  run_tests "lk-solver" "TestLandau|TestLKSolver" "./shared/physics/..."
  run_tests "m1-algo" "^Test" "./module1-hysteresis/pkg/algo/"
  run_tests "m1-simulation" "^Test" "./module1-hysteresis/pkg/simulation/"
  run_tests "m1-gui" "^Test" "./module1-hysteresis/pkg/gui/..."
  run_tests "m1-render" "^Test" "./module1-hysteresis/pkg/render/"
  run_tests "m1-tui" "^Test" "./module1-hysteresis/pkg/tui/"

  run_tests "discontinuity" "^TestHeadlessISPPContinuityValidation_" "./cmd/fecim-lattice-tools/"
  run_tests "experimental-data" "^TestExperimentalDataValidation$" "./validation/..."
  run_tests "headless-ispp" "^TestHeadless(LK|Preisach)|^TestHeadlessHysteresis_" "./cmd/fecim-lattice-tools/"
fi

echo ""
echo "=== SUMMARY ==="
echo "Pass: $PASS"
echo "Fail: $FAIL"
echo "Mode: $MODE"

if $JSON; then
  echo "{\"pass\": $PASS, \"fail\": $FAIL, \"mode\": \"$MODE\", \"commit\": \"$(git rev-parse --short HEAD)\"}"
fi

if [ $FAIL -gt 0 ]; then
  echo "RESULT: FAIL"
  exit 1
else
  echo "RESULT: PASS"
  exit 0
fi
