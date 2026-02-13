# Status Report — Module 4 Circuits Physics-Correct Flow

Date: 2026-02-13

## Validation gates

- `go build ./...` : PASS (0 errors)
- `go vet ./...` : PASS (0 warnings)
- `go test -short ./module4-circuits/... ./validation/...` : PASS

## Test counts (short mode)
Command used for counting:
`go test -short -json ./module4-circuits/... ./validation/...`

Per-package counts (pass/fail/skip):

- `module4-circuits/cmd/circuits`: 0 / 0 / 0 (no test files)
- `module4-circuits/cmd/circuits-gui`: 0 / 0 / 0 (no test files)
- `module4-circuits/pkg/arraysim`: 76 / 0 / 0
- `module4-circuits/pkg/gpuperiph`: 15 / 0 / 0
- `module4-circuits/pkg/gui`: 195 / 0 / 0
- `module4-circuits/pkg/gui/unified/display`: 0 / 0 / 0 (no test files)
- `module4-circuits/pkg/gui/unified/ispp`: 0 / 0 / 0 (no test files)
- `module4-circuits/pkg/gui/unified/overlay`: 0 / 0 / 0 (no test files)
- `module4-circuits/pkg/gui/unified/sense`: 0 / 0 / 0 (no test files)
- `validation`: 52 / 0 / 0
- `validation/benchmarks`: 2 / 0 / 0
- `validation/calibration`: 1 / 0 / 0
- `validation/comparator`: 1 / 0 / 0
- `validation/configvalidator`: 57 / 0 / 0
- `validation/configvalidator/cmd/validate`: 0 / 0 / 0 (no test files)
- `validation/external`: 1 / 0 / 3 (3 skips are optional external tools)
- `validation/heracles`: 2 / 0 / 0
- `validation/integration`: 24 / 0 / 0

Totals (packages under command scope):
- Tests passed: 426
- Tests failed: 0
- Tests skipped: 3

## ngspice availability
- `which ngspice` returned no path (not installed on this host)
- Structural netlist validation is active; runtime ngspice comparison test auto-skips with message until installed.
