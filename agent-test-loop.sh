#!/bin/bash
# agent-test-loop.sh
# Continuous testing loop for fecim-lattice-tools
# Run this, then point your LLM agent at the output

set -e

ITERATION=0
LOG_DIR="./test_logs"
PROJECT_DIR="<local-path>"
mkdir -p "$LOG_DIR"

cd "$PROJECT_DIR"

echo "========================================="
echo "FECIM AUTOMATED TEST LOOP"
echo "========================================="
echo "Log directory: $LOG_DIR"
echo "Press Ctrl+C to stop"
echo ""

while true; do
  ITERATION=$((ITERATION + 1))
  echo ""
  echo "========================================="
  echo "ITERATION $ITERATION - $(date)"
  echo "========================================="

  # Step 1: Run unit tests
  echo ""
  echo "[TEST] Running unit tests..."
  if GOCACHE=/tmp/go-build go test ./... 2>&1 | tee "$LOG_DIR/test_$ITERATION.log"; then
    TEST_EXIT=0
    echo "[TEST] ✓ PASS"
  else
    TEST_EXIT=1
    echo "[TEST] ✗ FAIL"
  fi

  # Step 2: Race detection (critical for concurrency)
  echo ""
  echo "[RACE] Checking for race conditions..."
  if GOCACHE=/tmp/go-build go test -race ./shared/physics/... ./module2-crossbar/pkg/crossbar/... 2>&1 | tee "$LOG_DIR/race_$ITERATION.log"; then
    RACE_EXIT=0
    echo "[RACE] ✓ PASS"
  else
    RACE_EXIT=1
    echo "[RACE] ✗ FAIL"
  fi

  # Step 3: Physics regression
  echo ""
  echo "[PHYSICS] Running physics validation..."
  if GOCACHE=/tmp/go-build go test -v ./validation/... 2>&1 | tee "$LOG_DIR/physics_$ITERATION.log"; then
    PHYSICS_EXIT=0
    echo "[PHYSICS] ✓ PASS"
  else
    PHYSICS_EXIT=1
    echo "[PHYSICS] ✗ FAIL (or no validation tests)"
  fi

  # Step 4: Coverage (periodic, every 5 iterations)
  if [ $((ITERATION % 5)) -eq 0 ]; then
    echo ""
    echo "[COVERAGE] Generating coverage report..."
    GOCACHE=/tmp/go-build go test -coverprofile="$LOG_DIR/coverage_$ITERATION.out" ./... 2>&1 | tail -5
    COVERAGE=$(go tool cover -func="$LOG_DIR/coverage_$ITERATION.out" 2>/dev/null | tail -1 | awk '{print $3}')
    echo "[COVERAGE] Total: $COVERAGE"
  fi

  # Step 5: Generate summary
  SUMMARY_FILE="$LOG_DIR/summary_$ITERATION.txt"
  {
    echo "========================================="
    echo "ITERATION $ITERATION SUMMARY"
    echo "========================================="
    echo "Timestamp: $(date -Iseconds)"
    echo ""
    echo "Unit Tests: $([ $TEST_EXIT -eq 0 ] && echo '✓ PASS' || echo '✗ FAIL')"
    echo "Race Check: $([ $RACE_EXIT -eq 0 ] && echo '✓ PASS' || echo '✗ FAIL')"
    echo "Physics:    $([ $PHYSICS_EXIT -eq 0 ] && echo '✓ PASS' || echo '✗ FAIL')"
    echo ""
    
    if [ $TEST_EXIT -ne 0 ]; then
      echo "--- TEST FAILURES ---"
      grep -E "FAIL|Error|panic" "$LOG_DIR/test_$ITERATION.log" 2>/dev/null | head -20
      echo ""
    fi
    
    if [ $RACE_EXIT -ne 0 ]; then
      echo "--- RACE CONDITIONS ---"
      grep -E "DATA RACE|WARNING" "$LOG_DIR/race_$ITERATION.log" 2>/dev/null | head -20
      echo ""
    fi
  } > "$SUMMARY_FILE"

  echo ""
  echo "[SUMMARY] Saved to $SUMMARY_FILE"

  # Step 6: Determine agent action
  echo ""
  if [ $TEST_EXIT -eq 0 ] && [ $RACE_EXIT -eq 0 ] && [ $PHYSICS_EXIT -eq 0 ]; then
    echo "==========================================="
    echo "  ✓ ALL CHECKS PASS"
    echo "==========================================="
    echo "AGENT_ACTION: All checks pass. Consider improvements:"
    echo "  - Increase test coverage for untested packages"
    echo "  - Add edge case tests for physics functions"
    echo "  - Profile hot paths (go test -bench=.)"
    echo "  - Review TODO.md for next priority item"
  else
    echo "==========================================="
    echo "  ✗ ISSUES FOUND"
    echo "==========================================="
    echo "AGENT_ACTION_REQUIRED: Fix the following:"
    [ $TEST_EXIT -ne 0 ] && echo "  - Unit test failures (see $LOG_DIR/test_$ITERATION.log)"
    [ $RACE_EXIT -ne 0 ] && echo "  - Race conditions (see $LOG_DIR/race_$ITERATION.log)"
    [ $PHYSICS_EXIT -ne 0 ] && echo "  - Physics regression (see $LOG_DIR/physics_$ITERATION.log)"
  fi

  # Pause between iterations
  echo ""
  echo "Waiting 30 seconds before next iteration..."
  sleep 30
done
