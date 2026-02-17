# FeCIM Test Suite Guide

This guide documents the test suite itself: what categories exist, how to run them, what key files validate, physics tolerances, environment limitations, and how to add new tests safely.

## 1) Test Categories

The repo uses seven practical categories:

1. **Unit tests**
   - Fast, package-local behavior checks (math, parsing, helpers, model invariants).
2. **Integration tests**
   - Multi-component checks inside a module or across modules.
3. **Physics regression tests**
   - Golden/reference comparison for hysteresis and LK curves with numeric tolerances.
4. **Fuzz tests (deterministic randomized stress)**
   - Randomized numeric trajectories to catch NaN/Inf, bounds violations, and unstable updates.
5. **Property-based tests**
   - Invariant-focused tests over random material/solver spaces.
6. **Performance tests**
   - Benchmarks and stress tests for throughput/runtime behavior.
7. **End-to-end (E2E) tests**
   - Workflow-level tests (headless core workflows and optional GUI/display workflows).

---

## 2) How to Run Each Category

> Run from repo root: `<local-path>`

### A. Baseline / all tests

### Headless, material-aware parity lanes (required gates)

These lanes are designed to be **fully headless** (no `DISPLAY`, no `WAYLAND_DISPLAY`, no `xvfb-run`) while still exercising the same physics paths the GUI uses.

```bash
# Required headless regression lanes (material-aware)
./scripts/run_headless_ispp_regressions.sh
./scripts/run_headless_module4_regressions.sh

# CI-like wrappers (also enforce headless requirements)
./scripts/ci/go-test-all.sh
./scripts/ci/go-test-race.sh
```

Notes:
- These runners emit per-material verdicts; aggregate PASS must not hide a single failing material.
- GUI lifecycle and screenshot/visual tests are optional and explicitly gated behind xvfb env vars.


```bash
go test ./...
```

Recommended pre-merge safety gate:

```bash
go test -race ./...
```

### B. Unit tests

Quick unit-oriented pass (many packages):

```bash
go test -short ./...
```

Target examples:

```bash
go test ./shared/physics/...
go test ./module2-crossbar/pkg/crossbar/...
go test ./module6-eda/pkg/export/...
```

### C. Integration tests

Run integration-named tests explicitly:

```bash
go test -v ./... -run Integration
```

Important integration packages:

```bash
go test -v ./cmd/fecim-lattice-tools/... -run Integration
go test -v ./validation/... -run Integration
go test -v ./module3-mnist/pkg/core/... -run Integration
```

### D. Physics regression tests

Primary physics regression suite:

```bash
go test -v ./validation/... -run TestPhysicsRegressionCurves
```

Related golden/stability suites:

```bash
go test -v ./module1-hysteresis/pkg/ferroelectric/... -run Golden
go test -v ./module1-hysteresis/pkg/controller/... -run HeadlessRegression
```

Update curve goldens (intentional updates only):

```bash
FECIM_UPDATE_PHYSICS_GOLDEN=1 go test -v ./validation/... -run TestPhysicsRegressionCurves
```

### E. Fuzz tests (deterministic randomized)

```bash
go test -v ./shared/physics/... -run Fuzz
```

Optional expansion knobs:

```bash
FECIM_FUZZ_ITERS=200 FECIM_FUZZ_STEPS=500 go test -v ./shared/physics/... -run Fuzz
```

### F. Property-based tests

```bash
go test -v ./shared/physics/... -run Property
```

### G. Performance tests

All benchmarks:

```bash
make bench
# or
go test ./... -run '^$' -bench . -benchmem
```

Targeted benchmarks (physics/crossbar hot paths):

```bash
go test -run '^$' -bench . -benchmem ./shared/physics/... ./module2-crossbar/pkg/crossbar/...
```

### H. E2E tests

Headless E2E workflows:

```bash
go test -v ./cmd/fecim-lattice-tools/... -run E2E
```

GUI E2E (display required, excluded from CI with `!ci` build tag):

```bash
go test -v ./cmd/fecim-lattice-tools/... -run E2EGUI
```

Xvfb-assisted visual/UI crawler tests:

```bash
FECIM_RUN_XVFB=1 xvfb-run -a go test -v ./cmd/fecim-lattice-tools/... -run 'VisualXvfb'
FECIM_UI_CRAWL=1 xvfb-run -a go test -v ./cmd/fecim-lattice-tools/... -run 'UICrawler'
FECIM_LAYOUT_AUDIT=1 xvfb-run -a go test -v ./cmd/fecim-lattice-tools/... -run 'LayoutAudit'
```

### I. CI-like deterministic local run

```bash
make ci
go test -race -short ./shared/... ./validation/...
go test -short ./... -coverprofile=coverage.out
```

---

## 3) Key Test Files and What They Validate

### Cross-cutting / top-level workflows

- `cmd/fecim-lattice-tools/integration_test.go`
  - Cross-module consistency and high-level integration checks.
- `cmd/fecim-lattice-tools/e2e_test.go`
  - Headless workflow E2E: MVM flows, MNIST dual-path inference, quantization workflows, concurrency.
- `cmd/fecim-lattice-tools/e2e_gui_test.go` (`!ci`)
  - GUI lifecycle/switching/concurrency tests under Fyne test harness.
- `cmd/fecim-lattice-tools/e2e_visual_xvfb_test.go`
  - Display-driver visual tests behind explicit env gating.

### Physics core and numerical stability

- `shared/physics/fuzz_test.go`
  - Deterministic randomized stress of Preisach/LK/ISPP paths; checks finite outputs and state bounds.
- `shared/physics/fuzz_property_test.go`
  - Randomized LK + Preisach stability with configurable iteration depth.
- `shared/physics/property_test.go`
  - Property invariants: loop closure, positive loop area, monotonic branches, Pr/Ps bounds.
- `shared/physics/e2e_pipeline_test.go`
  - Pipeline-level physics checks (with expected skips for known hard convergence scenarios).

### Hysteresis/material validation

- `module1-hysteresis/pkg/ferroelectric/physics_validation_test.go`
  - Literature-bound checks for Pr, Ps, Ec, temperature scaling, imprint, capacitance consistency.
- `module1-hysteresis/pkg/ferroelectric/golden_regression_test.go`
  - Golden loop/temperature-sweep/30-state stability checks and versioned golden metadata.
- `module1-hysteresis/pkg/controller/headless_regression_test.go`
  - Headless WRD/ISPP regressions (Preisach + LK) with pulse/overshoot budgets and JSON summaries.

### Validation package regression

- `validation/physics_regression_test.go`
  - Versioned physics curve regression (`testdata/physics_regression`) with RMS/max error gates.
- `validation/integration_test.go`
  - Validation-layer integration checks.
- `validation/cross_module_test.go`
  - Cross-module consistency assertions.

### EDA and tooling guards

- `module6-eda/pkg/validate/validate_test.go`
  - Yosys invocation validation; skips when Yosys unavailable.
- `module6-eda/pkg/validation/tooling_guards_test.go`
  - Tool availability guardrails and error-path handling.

---

## 4) Physics Tolerance Criteria

Current suites use explicit numerical acceptance bands rather than visual/manual checks.

### A. Validation physics regression (`validation/physics_regression_test.go`)

- **Preisach default HZO loop**
  - `x` (field) max abs error ≤ `1e-12`
  - `y` RMS error ≤ `2% * Ps`
  - `y` max abs error ≤ `3% * Ps`
- **LK loop default**
  - `x` max abs error ≤ `1e-9` V/m
  - `y` RMS error ≤ `3% * PMax`
  - `y` max abs error ≤ `5% * PMax`

### B. Hysteresis golden checks (`module1-hysteresis/pkg/ferroelectric/golden_regression_test.go`)

- Loop RMS tolerance: `2%` of `Emax` for field and `2%` of `Ps` for polarization.
- Temperature sweep max error: `2%` of room-temp Ec/Pr over full range.
- Automotive-range temperature sweep (233–423 K): stricter `1%` limits.
- 30-state quantization check: near-exact (`1e-10`) spacing/state-value tolerance.

### C. Material/literature checks (`physics_validation_test.go`)

- Typical literature envelope checks use ~`10%` tolerance margins around referenced ranges.
- Equation-level checks (e.g., capacitance formula) use tight numerical precision thresholds.

---

## 5) Known Limitations and Environment Constraints

### Fyne harness limitations

- Some GUI tests are skipped under `test.NewApp()` due to theme/font constraints and event-loop behavior.
- Known examples include crossbar/MNIST GUI paths in `e2e_gui_test.go` and visual tests in `e2e_visual_test.go`.
- CI excludes GUI-heavy tests via build tags and headless checks.

### Docker / EDA toolchain limitations

- EDA validation tests may skip when external tools are unavailable (e.g., Yosys).
- Docker/OpenLane-dependent checks are environment-sensitive and not guaranteed on minimal CI/dev hosts.
- Treat these as conditional integration checks, not always-on unit checks.

### Vulkan / GPU limitations

- GPU/Vulkan tests in `shared/compute`, `shared/gpu`, and GPU-adjacent modules skip when compute context is unavailable.
- This is expected behavior on CPU-only CI runners or systems without proper Vulkan drivers/runtime.

### Headless/display constraints

- Tests requiring a display (`DISPLAY`/Wayland) skip in headless environments unless run via Xvfb.
- Some UI crawler/layout audit suites require explicit opt-in env vars to avoid accidental slow/flaky runs.

---

## 6) How to Add New Tests

1. **Choose category first**
   - Unit / integration / physics regression / fuzz / property / performance / E2E.
2. **Place tests near owned code**
   - Use `*_test.go` in same package directory.
3. **Use clear naming**
   - `TestXxx`, `BenchmarkXxx`; for category discoverability include suffixes like `...Integration`, `...Property`, `...Regression` where useful.
4. **Make deterministic by default**
   - Seed random tests explicitly.
   - Avoid clock/time dependence unless required.
5. **Define numeric tolerances explicitly**
   - Include units in failure messages.
   - Prefer relative + absolute bounds for physics comparisons.
6. **Gate environment-sensitive tests**
   - Skip cleanly when missing GPU/display/tooling.
   - Add opt-in env variables for expensive/visual flows.
7. **Use golden updates safely**
   - Keep update flags explicit (e.g., `FECIM_UPDATE_PHYSICS_GOLDEN=1`).
   - Document why golden data changed in PR/commit notes.
8. **Run required validation before merging**

```bash
go test ./...
go test -race ./...
```

9. **If adding a new category/suite, update this guide**
   - Add run command, key files, and tolerance/limitation notes.

---

## Practical Quick Commands

```bash
# Full quick confidence
go test ./... && go test -race ./...

# Physics-focused sweep
go test -v ./shared/physics/... ./module1-hysteresis/pkg/ferroelectric/... ./validation/...

# E2E headless only
go test -v ./cmd/fecim-lattice-tools/... -run E2E

# Bench hot paths
go test -run '^$' -bench . -benchmem ./shared/physics/... ./module2-crossbar/pkg/crossbar/...
```
