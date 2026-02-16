#!/bin/bash
# scripts/module4_automation.sh - Module 4 automated testing and validation.
# Usage: ./scripts/module4_automation.sh [--fast|--full] [--json]
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

echo "=== Module 4 Automation Suite (mode: $MODE) ==="
echo "Commit: $(git rev-parse HEAD)"
echo "Date: $(date -Iseconds)"

PASS=0
FAIL=0
SKIP=0

run_tests() {
  local label="$1"
  local pattern="$2"
  local pkg="$3"
  echo "--- $label ---"

  # Note: `-run` is a RE2 regex. Keep patterns conservative to avoid accidental matches.
  if go test -short -count=1 -run "$pattern" "$pkg" -timeout 120s 2>&1 | tee "/tmp/m4_auto_${label}.log" | tail -3; then
    PASS=$((PASS+1))
  else
    FAIL=$((FAIL+1))
    echo "FAILED: $label"
  fi
}

# Phase 1 P0: Thermodynamics
run_tests "thermo" "TestThermodynamics" "./module4-circuits/pkg/arraysim/"

# Phase 1 P0: Patterns (March C-, walking ones/zeros, sneak)
run_tests "patterns" "TestPattern" "./module4-circuits/pkg/arraysim/"

# Phase 1 P0: Kirchhoff identities + KCL/KVL/Ohm
run_tests "kirchhoff" "TestKirchhoff" "./module4-circuits/pkg/arraysim/"

if [ "$MODE" = "full" ]; then
  # Phase 2: Retention + Disturb
  run_tests "retention" "TestRetention" "./module4-circuits/pkg/gui/"
  run_tests "disturb" "TestWriteDisturb" "./module4-circuits/pkg/gui/"

  # Phase 2: PVT (temperature sweep, process corners, peripheral coupling)
  run_tests "pvt" "TestPVT|TestDeviceState_PeripheralPVT" "./module4-circuits/pkg/gui/"

  # Phase 3: MVM + BER + Margin
  run_tests "mvm" "TestComputeMVM" "./module4-circuits/pkg/gui/"
  run_tests "ber" "TestComputeBER|TestReadMarginBER" "./module4-circuits/pkg/gui/"

  # Phase 3: Peripherals
  run_tests "inl_dnl" "TestPeripheralsINLDNL|TestINLDNL" "./shared/peripherals/"
  run_tests "noise" "TestNoise|TestTIA|TestPeripheralsNoise" "./shared/peripherals/"
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
