# Module 2 Automated Testing Plan — Research-Grade Crossbar Validation

## 1. Purpose

Build a **research-grade** automated test pipeline for Module 2 (`module2-crossbar`) that validates all crossbar array physics at publication quality. Module 2 is the core compute engine: it implements the NxN resistive crossbar array where matrix-vector multiplication (MVM) happens in the analog domain, with IR drop, sneak path, drift, temperature, and write-disturb non-ideality models.

### Scope

| Grade | Definition | Status |
|-------|------------|--------|
| **Simulation-grade** | Internally consistent, deterministic, passes invariant checks | Current (83% coverage, 564 tests) |
| **Research-grade** | Kirchhoff-verified, literature-calibrated, uncertainty-quantified, SPICE co-verified | Target |

### Current State

| Metric | Value |
|--------|-------|
| Source files | 50 (.go) |
| Test files | 49 (*_test.go) |
| Test functions | 564 pass / 0 fail / 0 skip |
| Coverage | 83.0% |
| Key packages | `pkg/crossbar/` (core), `pkg/gui/` (UI), `pkg/network/`, `pkg/training/`, `pkg/weights/` |

---

## 2. Architecture Overview

### Physics Engine Stack

```
Input Vector (V_in)
    │
    ▼
DAC Quantization (DACBits)
    │
    ▼
Word Line Driver → Row Metal (IR drop) → Cell (G_ij) → Column Metal (IR drop) → Sense Amp
    │                                        │
    ▼                                        ▼
Sneak Path Currents                    Column Output (I_out = Σ V_i × G_ij)
    │                                        │
    ▼                                        ▼
Half-Select Disturb                    ADC Quantization (ADCBits)
    │                                        │
    ▼                                        ▼
Drift/Retention                        MVM Result
```

### Key Source Files

| File | Domain | Lines |
|------|--------|-------|
| `array.go` | Core array, MVM, VMM, programming, conductance models | ~830 |
| `solver.go` | SOR parasitic solver (Kirchhoff nodal analysis) | ~380 |
| `solver_optimized.go` | Optimized SOR solver (cache-friendly) | ~400 |
| `irdrop.go` | IR drop simulation with temperature | ~340 |
| `sneakpath.go` | Sneak path analysis and mitigation | ~230 |
| `drift.go` | Conductance drift/retention modeling | ~470 |
| `drift_calibration.go` | Retention data → drift model calibration | ~90 |
| `temperature.go` | Arrhenius temperature effects on conductance | ~200 |
| `temperature_profile.go` | Spatial temperature profiles (hotspot, gradient) | ~120 |
| `write_disturb.go` | Half-select / write disturb modeling | ~400 |
| `nonidealities.go` | RC delay, wire parasitic modeling | ~160 |
| `device_errors.go` | Programming noise, read noise, variation | ~400 |
| `enhanced.go` | Combined non-ideality MVM (all effects) | ~210 |
| `gpu_mvm.go` | GPU-accelerated MVM (Vulkan placeholder) | ~50 |

---

## 3. Test Plan — 7 Phases

### Phase 1: Kirchhoff & Solver Validation (P0 — must pass for any claim)

**Goal**: Verify that the SOR parasitic solver satisfies KCL/KVL to machine precision.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M2-KCH-01 | `solver_kirchhoff_test.go` | KCL residual on 2×2, 4×4, 8×8, 16×16, 32×32 | Sum of currents at every node must be < 1e-10 |
| M2-KCH-02 | `solver_kirchhoff_test.go` | KVL loop consistency | V_applied = V_wire_drop + V_cell for every path |
| M2-KCH-03 | `solver_convergence_test.go` | SOR convergence vs array size | Iteration count, residual trajectory, omega sensitivity |
| M2-KCH-04 | `solver_convergence_test.go` | Optimized vs standard solver agreement | Max element-wise difference < 1e-8 across all outputs |
| M2-KCH-05 | `solver_condition_test.go` | Condition number reporting | Log cond(A) for varying Rp/Rcell ratios; flag ill-conditioned |

**Acceptance**: KCL residual < 1e-10, KVL residual < 1e-6, solvers agree to < 1e-8.

### Phase 2: MVM Accuracy & Quantization (P0)

**Goal**: Validate MVM/VMM correctness against ideal floating-point baseline.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M2-MVM-01 | `mvm_accuracy_test.go` | Ideal MVM vs numpy-style reference | 4×4, 8×8, 16×16 with known W and x; compare to W·x |
| M2-MVM-02 | `mvm_accuracy_test.go` | DAC/ADC quantization error budget | Sweep 4-8 bit DAC/ADC, measure SNR degradation |
| M2-MVM-03 | `mvm_accuracy_test.go` | MVM with all non-idealities | IR drop + noise + drift combined; BER < 5% |
| M2-MVM-04 | `mvm_vmm_symmetry_test.go` | MVM vs VMM transpose relationship | MVM(W,x) ≈ VMM(W^T,x) to quantization tolerance |
| M2-MVM-05 | `conductance_model_test.go` | Linear vs exponential vs lookup | Each model produces valid G in [GMin, GMax] |

**Acceptance**: Ideal MVM matches to machine epsilon; quantized MVM within ADC/DAC budget; BER < 5% with non-idealities.

### Phase 3: IR Drop & Sneak Path Physics (P0)

**Goal**: Verify IR drop and sneak path models match analytic solutions for small arrays.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M2-IRD-01 | `irdrop_analytic_test.go` | 1×1 cell: V_cell = V_in - I×(R_row + R_col) | Exact analytic check |
| M2-IRD-02 | `irdrop_analytic_test.go` | 2×2 uniform array analytic solution | Compare to hand-solved KCL equations |
| M2-IRD-03 | `irdrop_scaling_test.go` | IR drop scales with array size | Max drop increases monotonically with N; corner cells worst |
| M2-IRD-04 | `irdrop_scaling_test.go` | Temperature coefficient on wire resistance | R(T) = R(300K) × [1 + α(T-300)]; validate α range |
| M2-IRD-05 | `irdrop_mitigation_test.go` | Mitigation strategies reduce error | Each IRDropMitigation enum value reduces max drop |
| M2-SNK-01 | `sneak_analytic_test.go` | 2×2 worst-case sneak: one target, three parasitic | Analytic sneak current ratio |
| M2-SNK-02 | `sneak_scaling_test.go` | Sneak current ratio vs array size | Worst-case ratio increases with N (quantify) |
| M2-SNK-03 | `sneak_mitigation_test.go` | 1T1R vs 0T1R sneak comparison | 1T1R eliminates sneak; 0T1R has measurable ratio |

**Acceptance**: Analytic match < 1e-6 for small arrays; monotonic scaling trends; mitigation measurably reduces error.

### Phase 4: Drift, Retention & Write Disturb (P1)

**Goal**: Validate time-dependent physics: conductance drift, retention loss, half-select disturb.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M2-DFT-01 | `drift_physics_test.go` | Drift direction: conductance relaxes toward mean | High-G cells decrease, low-G cells increase |
| M2-DFT-02 | `drift_physics_test.go` | Power-law time dependence | ΔG ∝ t^ν, verify ν ∈ [0.01, 0.1] for HZO |
| M2-DFT-03 | `drift_retention_test.go` | Retention extrapolation: 10yr @ 85°C Arrhenius | Ea ∈ [0.5, 1.5] eV; validate extrapolation bounds |
| M2-DFT-04 | `drift_calibration_physics_test.go` | CalibrateDriftToRetention round-trip | Feed known retention data → recover drift coefficients |
| M2-DFT-05 | `drift_level_integrity_test.go` | Multi-level drift: levels don't cross | After 10^4s drift, level ordering preserved |
| M2-WRD-01 | `write_disturb_physics_test.go` | Half-select accumulation is monotonic | More V/2 pulses → more shift; shift ≥ 0 always |
| M2-WRD-02 | `write_disturb_physics_test.go` | 0T1R vs 1T1R disturb ratio | 0T1R disturb >> 1T1R (quantify ratio) |
| M2-WRD-03 | `write_disturb_spatial_test.go` | Spatial disturb pattern: same-row > different-row | Cells sharing a word line with target accumulate more |
| M2-WRD-04 | `write_disturb_endurance_test.go` | Endurance degradation curve | Pr degrades with cycles; validate power-law or stretched-exp |

**Acceptance**: Drift physics directionally correct; levels don't cross; 0T1R disturb > 1T1R by ≥ 2×; endurance trend matches published HZO data.

### Phase 4a: WriteDisturbEngine API Contract Tests (P0)

These tests validate the WriteDisturbEngine as a shared component used by Module 4.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M2-WDE-01 | `write_disturb_api_test.go` | Config validation | DefaultWriteDisturbConfig returns correct defaults (rate=1e-4, threshold=1.0, 1T1R reduction=0.1) |
| M2-WDE-02 | `write_disturb_api_test.go` | RecordWrite topology | After RecordWrite(r,c), only cells in same row or same column (excluding target) have non-zero stress |
| M2-WDE-03 | `write_disturb_api_test.go` | Stress accumulation rate | N RecordWrite calls produce stress = N × rate on each half-selected cell |
| M2-WDE-04 | `write_disturb_api_test.go` | Architecture reduction | 1T1R stress = 0T1R stress × 0.1 for same number of writes |
| M2-WDE-05 | `write_disturb_api_test.go` | Threshold level shift | After stress >= threshold, ApplyDisturbEffects shifts conductance by 1 quantum |
| M2-WDE-06 | `write_disturb_api_test.go` | Stress reset after shift | After ApplyDisturbEffects triggers a shift, residual stress = stress - threshold |
| M2-WDE-07 | `write_disturb_api_test.go` | Midpoint bias | Shifts bias toward center conductance (high G decreases, low G increases) |
| M2-WDE-08 | `write_disturb_api_test.go` | RecordBatchWrite equivalence | RecordBatchWrite([cells]) == sequential RecordWrite for each cell |
| M2-WDE-09 | `write_disturb_api_test.go` | GetStressMatrix snapshot | Returns deep copy (modifying returned matrix doesn't affect engine) |
| M2-WDE-10 | `write_disturb_api_test.go` | GetCellStress boundary | Out-of-bounds (r,c) returns 0, doesn't panic |
| M2-WDE-11 | `write_disturb_api_test.go` | Reset clears all state | After Reset(), all stress=0, TotalWriteOps=0, TotalHalfSelects=0 |
| M2-WDE-12 | `write_disturb_api_test.go` | Resize preserves nothing | After Resize(newR, newC), all stress=0, dimensions updated |
| M2-WDE-13 | `write_disturb_api_test.go` | Disabled engine is no-op | When Enable=false, RecordWrite does nothing, ApplyDisturbEffects returns 0 |
| M2-WDE-14 | `write_disturb_api_test.go` | Thread safety | Concurrent RecordWrite from 10 goroutines doesn't race (use -race flag) |
| M2-WDE-15 | `write_disturb_api_test.go` | IsDisturbCritical voltage check | V/2 > Vc × safety_margin returns true; V/2 < margin returns false |
| M2-WDE-16 | `write_disturb_api_test.go` | GetStressStats accuracy | TotalWriteOps, TotalHalfSelects, MaxStress, AvgStress all correct after known sequence |

Acceptance: All API contracts hold; thread-safe under -race; stress arithmetic exact to float64 precision.

### Phase 4b: WriteDisturbEngine Physics Validation (P1)

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M2-WDP-01 | `write_disturb_physics_validation_test.go` | 10k-pulse convergence | After 10,000 writes to same cell, half-selected cells have shifted exactly floor(10000 × rate / threshold) levels |
| M2-WDP-02 | `write_disturb_physics_validation_test.go` | Disturb probability matches analytic | PredictDisturbProbability(avgHalfSelects, config) matches Monte Carlo simulation within 5% |
| M2-WDP-03 | `write_disturb_physics_validation_test.go` | Spatial decay pattern | Cells sharing both row AND column with target get 2× stress vs cells sharing only one |
| M2-WDP-04 | `write_disturb_physics_validation_test.go` | Array-size independence | Stress per half-selected cell is independent of array dimensions (4×4 vs 64×64) |
| M2-WDP-05 | `write_disturb_physics_validation_test.go` | Multi-target write interaction | Writing to (0,0) then (0,1): column-0 cells get stress from first write, row-0 cells get stress from both |
| M2-WDP-06 | `write_disturb_physics_validation_test.go` | Worst-case disturb pattern | Identify the (row,col) combination that maximizes stress on any single cell for NxN array |

### Phase 5: Temperature & Process Variation (P1)

**Goal**: Validate temperature and statistical variation models.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M2-TMP-01 | `temperature_arrhenius_test.go` | Arrhenius conductance shift | G(T) follows Ea activation; validate Ea range |
| M2-TMP-02 | `temperature_profile_test.go` | Spatial temperature profiles | Hotspot model: center hotter than edge; gradient model: linear |
| M2-TMP-03 | `temperature_mvm_test.go` | MVM accuracy degrades with temperature | Error at 400K > error at 300K (quantify) |
| M2-PV-01 | `process_variation_test.go` | Gaussian variation statistics | σ matches configured NoiseLevel; mean ≈ 0 |
| M2-PV-02 | `process_variation_mvm_test.go` | Monte Carlo MVM spread | N=100 runs, compute μ ± 3σ; BER distribution |
| M2-PV-03 | `process_variation_yield_test.go` | Yield vs variation level | Yield (BER < threshold) decreases with σ |

**Acceptance**: Arrhenius physically correct (Ea > 0); hotspot > edge by > 5K; variation σ matches config to 10%.

### Phase 6: Scaling & Performance (P1)

**Goal**: Benchmark MVM throughput and verify scaling behavior.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M2-SCL-01 | `scaling_accuracy_test.go` | Accuracy vs array size (4×4 to 256×256) | Error increases monotonically with size |
| M2-SCL-02 | `scaling_performance_test.go` | MVM throughput benchmark | ns/op for 8×8, 32×32, 64×64, 128×128 |
| M2-SCL-03 | `scaling_memory_test.go` | Memory footprint vs array size | Allocs scale as O(N²) |
| M2-SCL-04 | `scaling_solver_test.go` | Solver iterations vs Rp/Rcell ratio | More parasitic → more iterations |

**Acceptance**: Monotonic accuracy degradation; throughput regression < 5% vs baseline; memory O(N²).

### Phase 7: GUI & Integration (P2)

**Goal**: Headless GUI smoke tests and end-to-end integration.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M2-GUI-01 | `gui_lifecycle_test.go` | Tab lifecycle: Ideal, IRDrop, Sneak, Drift | BuildContent/Start/Stop without panic |
| M2-GUI-02 | `gui_heatmap_test.go` | Heatmap rendering with known data | Cell colors match conductance values |
| M2-GUI-03 | `gui_export_test.go` | CSV/JSON export from GUI state | Export contains correct dimensions and values |
| M2-INT-01 | `integration_m2_m3_test.go` | M2 array feeds M3 MNIST inference | Program weights → MVM → classify digit → accuracy check |
| M2-INT-02 | `integration_m2_m4_test.go` | M2 physics matches M4 circuit model | Same conductance matrix → same MVM output (within tolerance) |

**Acceptance**: No panics in lifecycle; exports valid; cross-module MVM agreement < 1%.

### Phase 8: Enhanced Non-Ideality Interaction Tests (P1)

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M2-ENH-01 | `enhanced_interaction_test.go` | IR drop + write disturb | Write disturb doesn't violate Kirchhoff with IR drop enabled |
| M2-ENH-02 | `enhanced_interaction_test.go` | Drift + disturb ordering | Drift then disturb vs disturb then drift: order matters but both are bounded |
| M2-ENH-03 | `enhanced_interaction_test.go` | Temperature + disturb | Disturb rate at 400K should be higher than at 300K (Arrhenius acceleration) |
| M2-ENH-04 | `enhanced_interaction_test.go` | All non-idealities combined MVM | MVM with IR drop + sneak + drift + disturb + noise + temperature: result is finite, bounded, no NaN |
| M2-ENH-05 | `enhanced_interaction_test.go` | Noise floor vs disturb floor | Quantify: at what point does write disturb dominate over read noise |
| M2-ENH-06 | `enhanced_interaction_test.go` | Endurance interaction | After N write cycles, disturb sensitivity increases (Pr degradation amplifies disturb) |

### Phase 9: Cross-Module Integration (P1)

Expand existing Phase 7 M2-INT-02 with detailed Module 4 integration tests:

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M2-XM4-01 | `cross_module4_integration_test.go` | WriteDisturbEngine config parity | Module 4's lazy-init produces identical config to Module 2's DefaultWriteDisturbConfig |
| M2-XM4-02 | `cross_module4_integration_test.go` | Stress matrix parity | Same write sequence produces identical stress matrices in both modules |
| M2-XM4-03 | `cross_module4_integration_test.go` | Level shift parity | Same stressed array produces identical level shifts from both modules |
| M2-XM4-04 | `cross_module4_integration_test.go` | V/2 safety check integration | Module 4's 2×Vc voltage cap ensures IsDisturbCritical never returns true during normal ISPP |

### Phase 10: Statistical Disturb Analysis (P2)

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M2-SDA-01 | `disturb_statistical_test.go` | Monte Carlo disturb distribution | N=1000 random write patterns → histogram of disturb counts; verify normal-like distribution |
| M2-SDA-02 | `disturb_statistical_test.go` | Worst-case vs average disturb | Ratio of max/mean disturb for 8×8, 16×16, 32×32 arrays |
| M2-SDA-03 | `disturb_statistical_test.go` | Disturb-free write budget | Maximum number of writes before any cell shifts, as function of array size |
| M2-SDA-04 | `disturb_statistical_test.go` | Confidence interval on disturb probability | Bootstrap CI for PredictDisturbProbability across material parameter variations |

---

## 4. Automation Scripts

### Fast Gate (CI — < 30s)

```bash
#!/bin/bash
# scripts/run_module2_fast_gate.sh
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
echo "=== M2 Fast Gate ==="
go build ./module2-crossbar/...
go vet ./module2-crossbar/...
go test -short -count=1 ./module2-crossbar/...
echo "=== M2 Fast Gate PASS ==="
```

### Full Gate (Nightly — < 5min)

```bash
#!/bin/bash
# scripts/run_module2_full_gate.sh
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
echo "=== M2 Full Gate ==="
go build ./module2-crossbar/...
go vet ./module2-crossbar/...
go test -race -count=1 ./module2-crossbar/...
echo "=== M2 Full Gate PASS ==="
```

### Unified Runner

```bash
#!/bin/bash
# scripts/module2_automation.sh [--fast|--full|--json]
```

---

## 5. Evidence Requirements

Every test must produce:
1. **Exact numerical output** with units (not "it works")
2. **Comparison to expected** with tolerance and source citation
3. **Deterministic** — same seed → same result
4. **Material-aware** — tests bind explicit material parameters

### Reporting Format

```
M2-KCH-01: KCL residual 2×2=3.2e-15, 8×8=1.7e-12, 32×32=4.1e-10 (limit: 1e-10) — PASS
M2-MVM-01: Ideal MVM max error 8×8=2.2e-16 (machine epsilon) — PASS
M2-IRD-01: 1×1 analytic V_cell=0.9975V (expected=0.9975V, delta=0.0%) — PASS
```

---

## 6. Priority & Timeline

| Phase | Priority | Est. Tests | Description |
|-------|----------|-----------|-------------|
| Phase 1 | P0 | 10 | Kirchhoff & Solver — foundational correctness |
| Phase 2 | P0 | 10 | MVM Accuracy — core functionality |
| Phase 3 | P0 | 16 | IR Drop & Sneak — primary non-idealities |
| Phase 4 | P1 | 18 | Drift & Disturb — time-dependent physics |
| Phase 4a | P0 | 16 | WriteDisturbEngine API Contracts |
| Phase 4b | P1 | 6 | WriteDisturbEngine Physics |
| Phase 5 | P1 | 12 | Temperature & Variation — statistical |
| Phase 6 | P1 | 8 | Scaling & Performance — benchmarks |
| Phase 7 | P2 | 10 | GUI & Integration — end-to-end |
| Phase 8 | P1 | 6 | Non-Ideality Interactions |
| Phase 9 | P1 | 4 | Cross-Module 4 Integration |
| Phase 10 | P2 | 4 | Statistical Disturb Analysis |
| **New Total** | | **120** | (was 84) |

Phase 1-3 (P0) are mandatory for any research claim.
Phase 4-6 (P1) are required for publication.
Phase 7 (P2) is for completeness and CI.

---

## 7. Cross-Module Validation Chain

```
M1 (Hysteresis) → Material params → M2 (Crossbar) → MVM output → M3 (MNIST)
                                        ↓
                                   M4 (Circuits) imports WriteDisturbEngine
                                   M4 uses M2 stress model for V/2 disturb
                                        ↓
                                   M6 (EDA) exports M2 array specs
```

Tests M2-INT-01, M2-INT-02, and M2-XM4-01 through M2-XM4-04 verify this chain.

---

## 8. Literature References

- **IR Drop**: Chen et al., IEEE JSSC 2018 — wire resistance scaling for crossbar arrays
- **Sneak Path**: Linn et al., Nanotechnology 2012 — sneak path analysis in passive crossbar
- **Drift**: Ielmini, IEEE TED 2011 — conductance drift in resistive memories
- **Temperature**: Fantini et al., IEDM 2012 — Arrhenius temperature acceleration
- **Write Disturb**: Grossi et al., IEEE TED 2018 — half-select disturb in HfOx crossbar
- **MVM Accuracy**: Hu et al., Nature Comm. 2018 — DNN+NeuroSim accuracy benchmarks
- **Process Variation**: Chen & Yu, IEEE JETCAS 2019 — device-to-device variation in RRAM crossbar
- **V/2 Half-Select**: Luo et al., IEEE TED 2019 — V/2 scheme analysis for passive RRAM crossbar
- **Cumulative Stress**: Ambrogio et al., Nature 2018 — phase-change memory cumulative programming stress model
- **FeCIM Disturb**: Reis et al., IEDM 2024 — half-select disturb in ferroelectric CIM arrays
