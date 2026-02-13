#!/bin/bash
set -euo pipefail
echo "=== FeCIM Reproducibility Validation ==="
echo "Commit: $(git rev-parse HEAD)"
echo "Date: $(date -Iseconds)"

echo "--- Build ---"
go build ./...
echo "BUILD: PASS"

echo "--- Vet ---"
go vet ./...
echo "VET: PASS"

echo "--- Core Tests ---"
core_ok_count=$(go test -short -count=1 ./... 2>&1 | grep -cE '^ok')
echo "Core packages passing: ${core_ok_count}"
# Count and report

echo "--- Physics Regression ---"
go test -v ./validation/... -run PhysicsRegression -timeout 60s 2>&1 | tail -5

echo "--- ISPP Continuity ---"
go test -v -run 'TestHeadlessISPPContinuityValidation' ./cmd/fecim-lattice-tools/ -timeout 120s 2>&1 | tail -10

echo "--- Kirchhoff ---"
go test -v ./module4-circuits/pkg/arraysim/ -run 'Kirchhoff|CurrentValidation' -timeout 60s 2>&1 | tail -10

echo "--- ngspice Cross-Validation ---"
go test -v ./validation/external/ -run 'ngspice' -timeout 30s 2>&1 | tail -5

echo "=== COMPLETE ==="
