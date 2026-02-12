# Repository Health Dashboard

**Project:** FeCIM Lattice Tools  
**Generated:** 2026-02-12  
**Scope:** Build quality, test posture, coverage, performance, physics validation, and tracked blockers

---

## 1) Build Status

### Build pipeline checks

| Check | Command | Status | Notes |
|---|---|---|---|
| Build | `go build ./...` | ✅ PASS | Clean build across repository packages |
| Vet | `go vet ./...` | ✅ PASS | No vet diagnostics in current run |
| Format | `gofmt -l .` | ⚠️ Needs formatting | 5 files reported |

### Files currently not gofmt-clean

- `module1-hysteresis/pkg/controller/ispp_full_cycle_test.go`
- `module2-crossbar/pkg/crossbar/concurrent_stress_test.go`
- `module6-eda/pkg/export/roundtrip_test.go`
- `module7-docs/pkg/gui/docs_integrity_test.go`
- `shared/presets/presets_comprehensive_test.go`

---

## 2) Test Suite Summary

### Scale and package breadth

- `go list ./...` reports **85 packages** (meets the 65+ package scope requirement).
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

From `docs/development/TESTING.md` and project TODO tracking:

- Display-dependent GUI checks are conditionally skipped or require Xvfb/display server.
- Archived demo code under historical paths is intentionally excluded from active testing.
- Full `go test -race ./...` is currently tracked as blocked by a known unrelated compile mismatch in `module1-hysteresis/pkg/gui/equation_dialog_test.go` (symbol case mismatch).

---

## 3) Coverage Summary by Module

Coverage source: `coverage.out` (`go tool cover -func`).

| Module | Coverage |
|---|---:|
| module1-hysteresis | 39.5% |
| module2-crossbar | 52.2% |
| module3-mnist | 29.5% |
| module4-circuits | 61.3% |
| module5-comparison | 27.2% |
| module6-eda | 78.9% |
| module7-docs | 62.4% |
| validation | 66.8% |
| shared | 65.3% |

**Overall total statement coverage:** **48.0%**

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

This dashboard is a repository-level operational snapshot. Re-run build/vet/format/test/coverage benchmarks before release tagging to confirm drift since this report generation timestamp.
