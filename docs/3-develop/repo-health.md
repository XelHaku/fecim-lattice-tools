# Repository Health Dashboard

> **Note:** This file was previously located at `docs/REPO_HEALTH.md`. It has moved to `docs/3-develop/repo-health.md`.

**Project:** FeCIM Lattice Tools
**Generated:** 2026-03-05
**Scope:** Build quality, test posture, coverage, performance, physics validation, and tracked blockers

---

## 1) Build Status

### Build pipeline checks

| Check | Command | Status | Notes |
|---|---|---|---|
| Build | `go build ./...` | ✅ PASS | Revalidated on 2026-03-05 |
| Vet | `go vet ./...` | ✅ PASS | Revalidated on 2026-03-05 |
| Format | `gofmt -l .` | ⚠️ Needs formatting | 18 files reported on 2026-03-05 |

### Files currently not gofmt-clean

- `cmd/fecim-lattice-tools/playtest_circuits_calculations_test.go`
- `cmd/fecim-lattice-tools/playtest_report_test.go`
- `module1-hysteresis/pkg/ferroelectric/preisach.go`
- `module1-hysteresis/pkg/gui/gui.go`
- `module7-docs/pkg/gui/glossary_integration_test.go`
- `shared/compact/fefet.go`
- `shared/physics/landau.go`
- `shared/physics/material.go`
- `shared/physics/material_calibrated.go`
- `shared/physics/material_config.go`
- `shared/physics/worldclass_c2c.go`
- `shared/physics/worldclass_pund_sim_test.go`
- `shared/render3d/layer_stack.go`
- `shared/system/latency.go`
- `shared/system/power.go`
- `tmp_read_margin_probe.go`
- `validation/literature/module1_pe_loop_test.go`
- `validation/sense_chain_regression_test.go`

---

## 2) Test Suite Summary

### Scale and package breadth

- `go list ./...` reports **107 packages**.
- Test organization spans physics solvers, crossbar non-idealities, MNIST dual-path inference, EDA export/compiler, validation, and shared infrastructure.

### Test categories (high-level)

- **Physics and ferroelectric behavior** (L-K, Preisach, temperature/endurance/retention)
- **Crossbar non-idealities** (IR drop, sneak paths, drift, variation)
- **MNIST inference correctness** (FP vs CIM path, quantization, metrics)
- **Simulation/engine behavior** (state machine, waveform evolution)
- **Integration/E2E paths** (module-to-module workflows)
- **GUI/headless widget behavior** (logic-level coverage in headless mode)
- **EDA compiler/export correctness** (JSON/CSV/SPICE/Verilog/DEF/SVG)

### Known skips / conditional test limitations

From `docs/3-develop/testing/TESTING.md` and project TODO tracking:

- Display-dependent GUI checks are conditionally skipped or require Xvfb/display server.
- Archived demo code under historical paths is intentionally excluded from active testing.
- Full `go test -race ./...` is currently tracked as blocked by a known unrelated compile mismatch in `module1-hysteresis/pkg/gui/equation_dialog_test.go` (symbol case mismatch).

---

## 3) Coverage Status

Coverage is currently **stale / unavailable** in the active dashboard.

- `go tool cover -func=coverage.out` now fails because the checked-in `coverage.out` references missing paths (`module1-hysteresis/pkg/gui/simulation.go: no such file or directory`).
- A fresh coverprofile was **not** generated in this documentation slice, so historical percentages have been removed from the active health summary to avoid overstating confidence.
- Next research-grade docs step: regenerate `coverage.out` from a clean `go test -coverprofile=coverage.out ./...` run and then repopulate per-module numbers.

---

## 4) Performance Baselines

Current dashboard baselines:

- **LK solver step:** **80 ns** baseline
- **MVM (8x8):** **8.4 µs** baseline
- **Inference allocation profile:** **7 allocs/op** baseline
- **Preisach path:** **606 ns** baseline

These are the reference values used for regression/perf tracking in this dashboard.

---

## 5) Physics Validation Summary

Physics validation status is **healthy** based on project validation reports and regression coverage.

### Calibration

- Material calibration paths exist for HZO and related presets.
- Parameter sets are validated against expected ranges and golden references.

### Temperature behavior

- Temperature-dependent behavior is covered (coercive field/polarization trends, Curie-related scaling checks).
- Cold-to-hot sweeps are included in validation/test assets.

### Endurance / retention

- Endurance behavior is validated through cycle-based checks and degradation modeling tests.
- Retention and drift behavior are tested with temperature dependence and model comparisons.

### Precision / numerical robustness

- Physics kernels include deterministic regression checks, property/fuzz tests, and bounded numerical behavior assertions.
- Reported validations indicate stable solver behavior and SI-unit consistency.

---

## 6) Open Issues Count and Blocked Items

### Open issues count (tracked TODO focus items)

- **Tracked FOCUS items:** 106
- **Open (not ✅) FOCUS items:** **0**

### Blocked items currently tracked

1. **Race-suite blocker:** `go test -race ./...` blocked by compile mismatch in `module1-hysteresis/pkg/gui/equation_dialog_test.go` (`ShowPhysicsEquationsDialog` vs `showPhysicsEquationsDialog`).
2. **Tooling/environment blocker:** LaTeX-based SVG regeneration pipeline blocked on missing host `latex` binary (`exec: "latex": executable file not found`).

---

## Notes

This dashboard is a repository-level operational snapshot. Re-run build/vet/format/test/coverage benchmarks before release tagging to confirm drift since 2026-03-05.
