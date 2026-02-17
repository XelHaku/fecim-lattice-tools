# Code Quality Report

> **Note:** This file was previously located at `docs/CODE_QUALITY.md`. It has moved to `docs/3-develop/code-quality.md`.

**Generated:** 2026-02-07  
**Codebase:** FeCIM Lattice Tools  
**Go Version:** 1.22.2

---

## Executive Summary

| Metric | Value | Assessment |
|--------|-------|------------|
| Total Go Files | 424 | — |
| Source Lines (non-test) | 101,767 | Large codebase |
| Test Lines | 45,748 | Good test investment |
| Test/Source Ratio | 0.45 | Healthy |
| Average Coverage | 32.9% | ⚠️ Needs improvement |
| Packages with 0% Coverage | 27 | 🔴 Critical gaps |
| Functions with CC > 15 | 169 | ⚠️ High complexity |
| Duplicate Code Clusters | 241 | ⚠️ Refactoring needed |
| Average Cyclomatic Complexity | 4.36 | ✅ Good overall |

---

## 1. Cyclomatic Complexity Analysis

### Critical Complexity (CC > 50) — Immediate Attention Required

| Complexity | Function | File |
|------------|----------|------|
| **185** | `(*App).updatePhysics` | `module1-hysteresis/pkg/gui/simulation.go:846` |
| **152** | `(*CircuitsApp).drawUnifiedArray` | `module4-circuits/pkg/gui/tab_unified_drawing.go:19` |
| **95** | `MakeBuilderValidationTab` | `module6-eda/pkg/gui/tabs/builder_validation_tab.go:37` |
| **84** | `runHysteresisMode` | `cmd/fecim-lattice-tools/mode.go:40` |
| **79** | `(*WriteController).Update` | `module1-hysteresis/pkg/controller/writer.go:211` |
| **76** | `(*App).refreshGUI` | `module1-hysteresis/pkg/gui/simulation.go:1770` |
| **74** | `main` | `cmd/fecim-lattice-tools/main.go:463` |
| **65** | `(*App).createControlsPanel` | `module1-hysteresis/pkg/gui/controls.go:69` |
| **63** | `(*DigitCanvas).generateImage` | `module3-mnist/pkg/gui/canvas.go:165` |

### High Complexity (CC 30-50) — Refactoring Recommended

| Complexity | Function | File |
|------------|----------|------|
| 49 | `(*WriteController).calculateNextField` | `module1-hysteresis/pkg/controller/writer.go:539` |
| 45 | `(*WriteController).writeTarget` | `shared/physics/ispp_write.go:80` |
| 43 | `createSampleDigit` | `module3-mnist/cmd/mnist/main.go:940` |
| 43 | `inlineSVGUses` | `cmd/latex-svg/main.go:312` |
| 41 | `GenerateDEF` | `module6-eda/pkg/export/def.go:81` |
| 41 | `(*CircuitsApp).drawColTransistors` | `module4-circuits/pkg/gui/tab_unified_drawing.go:834` |
| 39 | `(*peplotRenderer).layoutWithSize` | `module1-hysteresis/pkg/gui/widgets/peplot.go:130` |
| 39 | `(*App).calibrateLevelsLK` | `module1-hysteresis/pkg/gui/simulation.go:2423` |
| 39 | `(*App).calibrateLevels` | `module1-hysteresis/pkg/gui/simulation.go:2191` |
| 38 | `(*CircuitsApp).drawRowTransistors` | `module4-circuits/pkg/gui/tab_unified_drawing.go:742` |
| 37 | `(*TierASolver).Solve` | `module4-circuits/pkg/arraysim/tier_a.go:22` |
| 35 | `(*Array).AnalyzeIRDropIterative` | `module2-crossbar/pkg/crossbar/nonidealities.go:257` |
| 34 | `(*CircuitsApp).drawTimingCompute` | `module4-circuits/pkg/gui/tab_reference_timing.go:344` |
| 34 | `(*ConfusionMatrix).generateImage` | `module3-mnist/pkg/gui/metrics.go:168` |
| 34 | `(*ComparisonCard).generateImage` | `module3-mnist/pkg/gui/comparison_card.go:116` |
| 34 | `(*ParasiticSolver).SolveMVM` | `module2-crossbar/pkg/crossbar/solver.go:111` |
| 33 | `(*SearchIndex).Query` | `module7-docs/pkg/gui/search.go:354` |
| 33 | `(*OptimizedParasiticSolver).SolveMVMFast` | `module2-crossbar/pkg/crossbar/solver_optimized.go:257` |
| 33 | `(*OptimizedParasiticSolver).SolveMVM` | `module2-crossbar/pkg/crossbar/solver_optimized.go:96` |
| 33 | `(*App).simulationLoop` | `module1-hysteresis/pkg/gui/simulation.go:625` |
| 32 | `PreprocessDigit` | `module3-mnist/pkg/gui/preprocess.go:29` |
| 31 | `DetectAudioDevices` | `shared/recording/audio.go:60` |
| 31 | `Run` | `module6-eda/cmd/eda-cli/main.go:36` |
| 31 | `(*DeviceState).computeWithArraysimLocked` | `module4-circuits/pkg/gui/device_state.go:1017` |
| 31 | `referenceSolveDense` | `module4-circuits/pkg/arraysim/refsolve_dense.go:21` |
| 31 | `(*CrossbarHeatmap).generateImage` | `module2-crossbar/pkg/gui/heatmap.go:376` |

### Complexity Distribution

- **CC ≤ 5:** 89% of functions ✅ (excellent)
- **CC 6-10:** 7% of functions
- **CC 11-15:** 2% of functions
- **CC 16-30:** 1.5% of functions
- **CC > 30:** 0.5% of functions (35 functions)

---

## 2. Test Coverage Analysis

### High Coverage Packages (≥80%) ✅

| Package | Coverage |
|---------|----------|
| `module6-eda/pkg/config` | 100.0% |
| `module2-crossbar/pkg/training` | 95.6% |
| `shared/errors` | 94.6% |
| `module2-crossbar/pkg/network` | 94.4% |
| `module6-eda/pkg/compiler` | 92.4% |
| `module3-mnist/pkg/core` | 87.4% |
| `shared/io` | 85.7% |
| `module1-hysteresis/pkg/simulation` | 84.3% |
| `module6-eda/pkg/export` | 81.3% |
| `module4-circuits/pkg/arraysim` | 80.9% |

### Moderate Coverage (40-80%)

| Package | Coverage |
|---------|----------|
| `module2-crossbar/pkg/weights` | 78.6% |
| `shared/logging` | 75.2% |
| `module4-circuits/pkg/gpuperiph` | 74.5% |
| `module2-crossbar/pkg/crossbar` | 69.3% |
| `shared/gpu` | 68.8% |
| `shared/peripherals` | 65.5% |
| `shared/physics` | 64.0% |
| `config/physics` | 63.5% |
| `module5-comparison/pkg/comparison` | 62.8% |
| `shared/compute` | 59.1% |
| `module3-mnist/pkg/training` | 45.2% |
| `cmd/latex-svg` | 43.1% |
| `module4-circuits/pkg/gui` | 41.3% |
| `module1-hysteresis/pkg/ferroelectric` | 40.5% |

### Critical Coverage Gaps (0%) 🔴

**27 packages with no test coverage:**

| Module | Package |
|--------|---------|
| **module1-hysteresis** | `cmd/hysteresis`, `pkg/algo`, `pkg/controller`, `pkg/render`, `pkg/tui` |
| **module2-crossbar** | `pkg/gui/tabs`, `pkg/visualization` |
| **module3-mnist** | `cmd/mnist`, `cmd/mnist-gui`, `cmd/train-network`, `cmd/train-ptq`, `cmd/train-single-layer` |
| **module4-circuits** | `cmd/circuits`, `cmd/circuits-gui` |
| **module5-comparison** | `cmd/comparison`, `cmd/comparison-gui` |
| **module6-eda** | `cmd/eda-cli`, `cmd/eda-gui`, `cmd/hello`, `cmd/lattice-gen`, `pkg/gui`, `pkg/gui/tabs`, `pkg/gui/widgets`, `pkg/layout`, `pkg/openlane`, `pkg/validate` |
| **module7-docs** | `pkg/gui` |

### Low Coverage (< 10%)

| Package | Coverage | Note |
|---------|----------|------|
| `module1-hysteresis/pkg/gui` | 0.6% | 2,991 LOC |
| `module5-comparison/pkg/gui` | 1.7% | GUI code |
| `module6-eda/pkg/validation` | 1.7% | Validation logic |
| `module2-crossbar/cmd/crossbar-gui` | 2.7% | Entry point |
| `module2-crossbar/pkg/gui` | 4.1% | GUI code |
| `module3-mnist/pkg/gui` | 4.8% | GUI code |

---

## 3. Duplicate Code Analysis

**241 duplicate code clusters detected** (threshold: 50 tokens)

### Cross-Module Duplication (High Priority)

| Files | Pattern |
|-------|---------|
| 8 modules | Main initialization pattern in `main.go` files |
| `module3-mnist` ↔ `module5-comparison` | `embedded.go` font loading (lines 79-92 / 55-68) |
| `module1-hysteresis/pkg/algo` ↔ `pkg/gui` | Calibration logic duplicated |
| `module3-mnist/cmd/train-*` | Training setup duplicated across 3 commands |

### Within-Module Duplication

| File | Pattern |
|------|---------|
| `module6-eda/pkg/validation/cross_check.go` | 3 nearly identical validation blocks |
| `module1-hysteresis/pkg/ferroelectric/physics_validation_test.go` | 2 large duplicate test blocks |
| `shared/physics/units.go` | Unit conversion functions repeated |
| `shared/widgets/material_picker_test.go` | 3 duplicate test setup blocks |

### Test Code Duplication

Multiple test files contain duplicated setup/teardown patterns:
- `logging_test.go`: overlapping test blocks
- `recording_test.go`: repeated fixture setup
- `temperature_thermal_test.go` ↔ `resize_detector_test.go`: identical patterns

---

## 4. Largest Files (Complexity Hotspots)

| Lines | File | Concern |
|-------|------|---------|
| 2,991 | `module1-hysteresis/pkg/gui/simulation.go` | 🔴 God file, contains 5 CC>30 functions |
| 2,079 | `module4-circuits/pkg/gui/device_state.go` | ⚠️ Complex state management |
| 1,725 | `module4-circuits/pkg/gui/tab_unified.go` | ⚠️ Large UI component |
| 1,472 | `module1-hysteresis/pkg/render/vulkan.go` | 0% coverage |
| 1,394 | `module6-eda/pkg/gui/tabs/builder_validation_tab.go` | CC=95 function |
| 1,108 | `module3-mnist/cmd/mnist/main.go` | 0% coverage, CC=43 function |
| 1,094 | `cmd/fecim-lattice-tools/main.go` | CC=74 main function |

---

## 5. Recommendations

### 🔴 Critical (Address Immediately)

1. **Refactor `simulation.go` (module1-hysteresis)**
   - Split `updatePhysics` (CC=185) into smaller functions
   - Extract physics calculations, state updates, and rendering into separate methods
   - Target: CC < 20 per function

2. **Refactor `tab_unified_drawing.go` (module4-circuits)**
   - `drawUnifiedArray` (CC=152) should be decomposed
   - Consider a strategy pattern for different drawing modes

3. **Add tests for core algorithms**
   - `module1-hysteresis/pkg/algo` (0% coverage)
   - `module1-hysteresis/pkg/controller` (0% coverage)
   - These contain critical physics simulation logic

### ⚠️ High Priority

4. **Reduce main() complexity**
   - `cmd/fecim-lattice-tools/main.go` (CC=74)
   - Extract subcommand handlers into separate functions
   - Consider a command pattern or CLI framework

5. **Eliminate cross-module duplication**
   - Create `shared/init` package for common initialization
   - Extract font loading to `shared/assets`
   - Consolidate training setup into `shared/training`

6. **Improve GUI test coverage**
   - All GUI packages have < 5% coverage
   - Add unit tests for non-rendering logic
   - Consider snapshot testing for visual components

### 📋 Medium Priority

7. **Standardize test patterns**
   - Create test helpers in `shared/testutil`
   - Reduce test code duplication
   - Add table-driven tests where appropriate

8. **Address validation gap**
   - `module6-eda/pkg/validate` has 0% coverage
   - `module6-eda/pkg/validation` has 1.7% coverage
   - Critical for EDA correctness

9. **Split large files**
   - Files > 1000 LOC should be reviewed for separation
   - Consider feature-based file organization

### ✅ Good Practices to Maintain

- Core packages (`crossbar`, `core`, `training`) have good coverage
- Shared utilities (`errors`, `io`, `logging`) are well-tested
- Average complexity is healthy at 4.36
- Test-to-source ratio of 0.45 shows investment in testing

---

## 6. Suggested Refactoring Order

1. **Week 1-2:** Split `simulation.go` functions
2. **Week 3:** Add tests for `pkg/algo` and `pkg/controller`
3. **Week 4:** Refactor main.go entry points
4. **Week 5-6:** Address cross-module duplication
5. **Ongoing:** Improve GUI test coverage incrementally

---

## Appendix: Tool Commands

```bash
# Cyclomatic complexity
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
gocyclo -top 30 -avg .

# Duplicate detection
go install github.com/mibk/dupl@latest
dupl -t 50 .

# Test coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Coverage by package
go test -cover ./... 2>&1 | grep -E "coverage:"
```

---

*Report generated by automated analysis. Manual review recommended for architectural decisions.*
