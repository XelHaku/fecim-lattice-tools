# Module 4 Automated Testing Plan -- Research-Grade Circuit Validation

## 1. Purpose

Build a **research-grade** automated test pipeline for Module 4 (`module4-circuits`) that validates **all circuit-level physics outputs** at publication quality. This means every numerical output produced by the READ, WRITE, and COMPUTE paths must be testable with quantified uncertainty, traceable parameter provenance, and bounded agreement with higher-fidelity reference methods.

This plan moves beyond "internally consistent simulation" to establish:

- **Thermodynamic consistency** at the array level
- **Bounded agreement** with higher-fidelity reference methods (SPICE co-verification)
- **Quantified uncertainty** and confidence intervals on every output
- **Standard CIM test protocols** from published literature (2024-2026)
- **Reviewer-expected metrics** for FeCIM/CIM publications
- **Parameter provenance** via a confidence ledger (measured / estimated / placeholder)

**Scope distinction:**

| Grade | Definition | Status |
|-------|------------|--------|
| **Simulation-grade** | Internally consistent, deterministic, passes invariant checks. No error bars, no experimental calibration, no uncertainty propagation. | Current state |
| **Research-grade** | Every output carries a confidence tag. Uncertainty is propagated through the full signal chain. Results are cross-validated against SPICE references. Statistical rigor (CIs, distribution tests). | Target state |

---

## 2. Current State Assessment

| Metric | Value | Grade |
|--------|-------|-------|
| Lines of Go code | ~22,400 | -- |
| Total tests | 237 (65 arraysim, 172 gui) | Simulation |
| Architectures | 0T1R (passive, 4F^2), 1T1R (active, 8-12F^2), 2T1R (planned) | Simulation |
| Coupling tiers | Ideal, Tier-A (IR drop approx), Tier-B (full DC nodal MNA, PCG solver, 1e-8 tol) | Simulation |
| Kirchhoff validation | Verified to 1e-6 residual | Simulation |
| SPICE cross-validation | 2e-11 tolerance on golden vectors (`tierb_spice_golden_vectors.json`) | Simulation |
| Experimental validation | **None** -- all parameters are simulation defaults | **Gap** |
| Uncertainty quantification | **None** -- no error bars on any output | **Gap** |
| Energy/timing calibration | **None** -- model estimates only | **Gap** |
| Statistical rigor | **None** -- no confidence intervals | **Gap** |
| Solver error analysis | **None** -- no condition number reporting | **Gap** |
| Module 4 GUI tests | Pre-existing build failures on clean main (`TestUnifiedTabISPPEngine`, `TestUnifiedActionButtons`) | Known issue |

### Critical Gaps (Simulation-Grade to Research-Grade)

| ID | Gap | Impact | Remediation |
|----|-----|--------|-------------|
| G1 | No experimental validation | All parameters are defaults, not calibrated to silicon or published data | Add parameter provenance tagging; flag every param as measured/estimated/placeholder |
| G2 | No uncertainty quantification | Outputs have no error bars; cannot assess confidence in any result | Propagate uncertainty through DAC->Array->TIA->ADC chain; report per-stage and end-to-end |
| G3 | Uncalibrated energy/timing | Energy and timing outputs are model estimates with unknown accuracy | Cross-check against published peripheral power budgets; tag as "estimated" until calibrated |
| G4 | No statistical rigor | No confidence intervals, no distribution tests, no sample size justification | Add KS tests, bootstrap CIs, Monte Carlo with documented N and convergence criteria |
| G5 | No solver error analysis | Tier-B MNA condition number never reported; no solver convergence diagnostics | Log condition number, iteration count, residual norm; fail if condition > threshold |
| G6 | GUI test build failures | `TestUnifiedTabISPPEngine`, `TestUnifiedActionButtons` do not build on clean main | Track separately; do not gate research pipeline on GUI tests |

---

## 3. Source Documentation

### Internal Documentation

- `docs/circuits/module4-flow-audit.md`
- `docs/circuits/module4-write-path-proof.md`
- `docs/circuits/signal-flow.md`
- `docs/peripheral-circuits/ARRAY_SIMULATION_FIDELITY.md`
- `docs/development/PHYSICS_ACCEPTANCE_CRITERIA.md`
- `docs/development/CI.md`
- `docs/development/HEADLESS.md`
- `docs/testing/TEST_GUIDE.md`
- `docs/opensource-tools/circuit-simulation-tools.md`
- `module4-circuits/FEATURES.md`
- `module4-circuits/README.md`

### Research Standards (2024-2026 CIM Literature)

- **MLC Programming**: 7-16 conductance levels with program-verify
- **Linearity Requirements**: R^2 > 0.95 for conductance vs level
- **Read Disturb**: State shift measurement after 10^6+ read cycles
- **Write Disturb**: VDD/2 and VDD/3 half-select inhibition validation
- **MVM Accuracy**: BER < 5%, within 1-10% of floating-point baseline
- **Standard Array Patterns**: Checkerboard, walking ones/zeros, March C/C-
- **Endurance**: 10^9 cycles with Pr degradation tracking
- **Retention**: 10-year @ 85 deg C extrapolation from Arrhenius fits
- **Energy Metrics**: fJ-pJ switching energy per operation
- **SPICE Co-verification**: +/-5% agreement with circuit simulator reference

---

## 4. Mandatory Requirements

### 4a. Material-Selected Awareness (Required)

- Every test case must bind an explicit material selection. No implicit/default material runs in regression gates.
- Material choice must be part of the test ID and artifact key:
  - Example key: `m4/1T1R/TierA/8x8/fecim_hzo/seed42`
- All voltage targets and safety checks must be material-derived, not hard-coded:
  - Use material-aware ranges (`readRange`, `writeRange`) from the active `DeviceState`.
  - Use material Ec, Ps, thickness, and conductance bounds when computing normalized metrics.
- Pass/fail verdicts are **per material** and then aggregated; no single global verdict may hide one failing material.
- Regression artifacts must include a material snapshot:
  - Material name/ID
  - Key physical params used in that run (at minimum: Ec, Ps, Pr, thickness, Gmin, Gmax, quant levels)

### 4b. Fully Headless Requirement (Mandatory)

- Required lanes must run with no display stack:
  - `DISPLAY` unset
  - `WAYLAND_DISPLAY` unset
  - No `xvfb-run`
- Required lanes must not depend on window creation, rendering, or screenshot capture.
- Any Fyne/UI harness tests are optional/non-gating and tracked separately.

### 4c. GUI/Headless Physics Parity (Mandatory)

- Headless harness must use the **same physics code paths** as GUI for READ, WRITE, COMPUTE.
- No headless-only duplicate equations are allowed for solver, ISPP, or conductance mapping.
- Parity validation must compare:
  - `effectiveCellVoltage`
  - Row current outputs
  - ADC levels
  - Write-step level trajectory and ISPP result

**Implementation snapshot (2026-02-13):**
- Fully-headless gates enforced in: `scripts/ci/go-test-all.sh`, `scripts/ci/go-test-race.sh`, `scripts/run_headless_ispp_regressions.sh`, `scripts/run_headless_module4_regressions.sh`
- GUI/headless parity test: `module4-circuits/pkg/gui/headless_gui_physics_parity_test.go` (`TestHeadlessPhysicsParity_GUIVsHeadless_ReadComputeWriteStep_MaterialAware`)

---

## 5. Full Circuit Physics Output Catalog

Every output listed below must be testable. Each entry includes the output, the formula or invariant it must satisfy, the validation tier, and the current implementation state.

### 5.1 READ Operation Outputs

| Output | Formula / Invariant | Tier | Current State |
|--------|---------------------|------|---------------|
| Cell voltage | V_cell from coupled solver (Tier-A/B) or ideal | T1 | Tested |
| Cell current | I_cell = G_cell x V_cell | T1 | Tested |
| TIA output voltage | V_out = I_in x R_feedback | T2 | Partial (SNR test exists via `composedSenseSNRdB`) |
| TIA SNR | > 40 dB for compute path | T2 | Tested (`TestComposedSenseSNRdB_DegradesWithNoise`) |
| TIA settling time | < read time budget (< 80 ns) | T2 | Not tested |
| TIA dynamic range | > 60 dB for multi-level sensing | T2 | Not tested |
| TIA input-referred noise | Documented from `peripherals.TIA` model | T3 | Not tested |
| ADC output code | Quantized from TIA voltage | T1 | Tested |
| ADC ENOB | >= nominal bits - 1 (>= 4 for 5-bit) | T2 | CLI only (`adc.ENOB()`) |
| ADC quantization noise | 1/sqrt(12) LSB theoretical | T2 | Not tested |
| ADC missing codes | DNL > -1 LSB everywhere | T2 | CLI only (`checkMonotonicity`) |
| Sense margin analysis | MinMarginV > 0, reliability metric | T1 | Tested (`ReadMarginAnalysis`) |
| Read immutability | arrayWeights unchanged (bitwise identical) | T1 | Tested |
| Read path signal chain | DAC->Array->TIA->ADC end-to-end consistency | T2 | Partial |
| Read disturb endurance | DeltaG/G < 1% after 10^6 reads | T3 | Not tested |
| Read repeatability | 100 consecutive reads, same output +/- noise floor | T2 | Not tested |

### 5.2 WRITE Operation Outputs

| Output | Formula / Invariant | Tier | Current State |
|--------|---------------------|------|---------------|
| ISPP convergence | Target level reached within pulse budget | T1 | Tested (arraysim ISPP tests) |
| ISPP pulse count | Histogram of pulses per target level | T2 | Not tested as distribution |
| ISPP success rate | > 95% across full level range (0-29) | T1 | Tested (ensemble) |
| ISPP overshoot rate | Bounded; tracked per material | T2 | Tracked but not gated |
| WriteVerifyStats | Pulses, overshoots, final level, success rate per cell | T2 | Struct exists but not regression-tested |
| Write energy per cell | fJ, from transient integration | T2 | Tested (`TestTransient_EnergyPerCell`) |
| Non-target disturbance (0T1R) | Half-select on selected row/col allowed; other cells unchanged | T1 | Tested (half-select residue test) |
| Non-target disturbance (1T1R/2T1R) | Non-target delta_level bounded to near-zero | T1 | Partial |
| Broadcast guard | Fail fast if > K non-target cells change by > L levels (K=3, L=3 default) | T1 | **Not tested (critical gap)** |
| ISPP engine parity | Level-based vs L-K ODE converge to same level +/-1 | T2 | Not tested |
| Write monotonicity | For ascending program, level[n+1] >= level[n] (within +/-1 quantization) | T2 | Tested |
| Coupled write voltage | V_eff(target) > V_eff(half-select) under IR drop | T2 | Tested |
| Half-select disturb accumulation | N-cycle DeltaP(unselected) < threshold | T3 | Not tested |

### 5.3 COMPUTE (MVM) Operation Outputs

| Output | Formula / Invariant | Tier | Current State |
|--------|---------------------|------|---------------|
| MVM linearity | I_out = G x V_in (dot product) | T1 | Tested (Ideal strict, Tier-A bounded) |
| Quantization PSNR | > 30 dB for 30 levels | T1 | **Not tested as explicit metric** |
| Level spacing uniformity | +/-5% from linear across 30 levels | T2 | Not tested |
| Compute immutability | arrayWeights unchanged (bitwise identical) | T1 | Tested |
| Full peripheral chain | DAC->Array->TIA->ADC end-to-end compute | T2 | Partial |
| Determinism | Same seed/config => identical output (bitwise) | T1 | Tested |
| BER target | BER < 5% for MNIST inference patterns | T2 | Not tested |
| Signed compute | Bipolar weights, bipolar input, four-quadrant | T2 | Not tested |

### 5.4 IR Drop Analysis Outputs

| Output | Formula / Invariant | Tier | Current State |
|--------|---------------------|------|---------------|
| Voltage drop map | 2D array of delta_V per cell | T2 | Available from Tier-A/B solver |
| Ohm's law compliance | delta_V = I x R_wire | T2 | Tested (`TestTierA_RwireZeroMatchesIdeal`) |
| Worst-case corner cell | Maximum cumulative current path drop | T2 | Not explicitly tested |
| Scaling validation | 5x resistance => ~5x drop (proportional) | T2 | Not tested |
| Position-dependent accuracy | Degradation vs distance from driver | T3 | Not tested |

### 5.5 Sneak Path Current Outputs

| Output | Formula / Invariant | Tier | Current State |
|--------|---------------------|------|---------------|
| 3-cell parasitic path | Known analytical sneak current for 2x2 | T2 | Partial (Tier-B tests) |
| SNR degradation with array size | Documented scaling relationship | T3 | Not tested |
| Selector mitigation | 10x selector Ron => ~10x sneak suppression | T2 | Tested (`TestTierA_SelectorRonReducesReadCurrentAndMargin`) |
| Negative SNR detection | Flag as failure mode when sneak > signal | T2 | Not tested |

### 5.6 Conductance Drift Outputs

| Output | Formula / Invariant | Tier | Current State |
|--------|---------------------|------|---------------|
| Power-law drift | G(t) = G_0 x (t/t_0)^(-nu) | T2 | Tested (`TestSimulateEnduranceAccuracy_MonotonicDegradation`) |
| FeCIM vs RRAM drift rate | nu_FeCIM ~ 0.001 vs nu_RRAM ~ 0.05 | T3 | Not explicitly compared |
| Arrhenius temperature scaling | Drift acceleration with temperature | T3 | Not tested |
| Retention projection | 10 years @ 85 deg C extrapolation | T3 | Not tested |

### 5.7 DAC Characterization Outputs

| Output | Formula / Invariant | Tier | Current State |
|--------|---------------------|------|---------------|
| INL | < 1 LSB (monotonicity guarantee) | T2 | CLI analysis (`dac.AnalyzeINLDNL()`) |
| DNL | < 1 LSB (no missing codes) | T2 | CLI analysis |
| Code width uniformity | Histogram of step sizes | T3 | Not tested |
| PVT corners | Fast/Slow/Typical within +/-10% | T3 | Model exists (`peripherals.ProcessCorner`) |
| Energy per conversion | fJ, from model | T2 | CLI output only |
| Monotonicity | DNL > -1 LSB everywhere | T2 | CLI only (`checkMonotonicity`) |

### 5.8 ADC Characterization Outputs

| Output | Formula / Invariant | Tier | Current State |
|--------|---------------------|------|---------------|
| ENOB | >= nominal bits - 1 | T2 | CLI only (`adc.ENOB()`) |
| Quantization noise | 1/sqrt(12) LSB theoretical | T2 | Not tested |
| Missing codes | DNL > -1 LSB everywhere | T2 | CLI only |
| Sample rate vs timing budget | Meets system timing constraint | T3 | Not tested |

### 5.9 TIA Characterization Outputs

| Output | Formula / Invariant | Tier | Current State |
|--------|---------------------|------|---------------|
| Transimpedance | V_out = I_in x R_feedback (verified linear range) | T2 | Partial |
| SNR | > 40 dB for compute path | T2 | Tested (`TestComposedSenseSNRdB_DegradesWithNoise`) |
| Settling time | < 80 ns read time budget | T2 | Not tested |
| Dynamic range | > 60 dB for multi-level sensing | T2 | Not tested |
| Input-referred noise | Documented from model parameters | T3 | Not tested |

### 5.10 System Timing Outputs

| Output | Formula / Invariant | Tier | Current State |
|--------|---------------------|------|---------------|
| Write time | Dominated by polarization switching (100-200 ns) | T2 | Tested (`TestTransient_CompleteSwitchingAt100ns`) |
| Read time | DAC + TIA + ADC settling budget | T2 | Not tested as budget |
| MVM throughput | Scaling with array size | T3 | Not tested |

### 5.11 Power Breakdown Outputs

| Output | Formula / Invariant | Tier | Current State |
|--------|---------------------|------|---------------|
| DAC/ADC power | Scaling linear with sample rate | T3 | Not tested |
| Array energy dominance | Array >> peripheral for large arrays | T3 | Not tested |
| Energy per MAC | Target < 10 pJ/MAC | T2 | Estimated in CLI, not gated |

### 5.12 Array Coupling (Tier A/B) Outputs

| Output | Formula / Invariant | Tier | Current State |
|--------|---------------------|------|---------------|
| Tier-A cell voltages/currents | Half-select V/2 pattern verified | T1 | Tested |
| Tier-B full MNA nodal solution | Wire IR drop per segment | T2 | Tested (golden vectors) |
| KCL compliance | sum(I_in) = sum(I_out) at each node, residual < 1e-6 | T1 | Tested (`TestReferenceSolveDense_KCLResidual_2x2`) |
| KVL compliance | sum(V) around any loop = 0 | T1 | Partial |
| MNA matrix conditioning | Condition number < threshold | T2 | **Not tested (gap G5)** |
| SPICE cross-validation | Node voltages match within 1% | T2 | Tested (2e-11 tol on golden) |
| SPICE netlist export | Valid SPICE deck generation | T2 | Tested (`TestExportCrossbarSPICE`) |
| Power conservation | sum(P_cells) + sum(P_wires) + P_peripherals = P_total within 1% | T1 | Not tested |

### 5.13 Transient Characterization Outputs

| Output | Formula / Invariant | Tier | Current State |
|--------|---------------------|------|---------------|
| P(t) transient | Polarization trajectory | T2 | Tested |
| I(t) transient | Current trajectory | T2 | Tested |
| V(t) transient | Voltage trajectory | T2 | Tested |
| Write time (90% Pr) | Switching completion metric | T2 | Tested |
| Read settling time | Non-disturb verification | T2 | Tested (`TestTransient_ReadDoesNotDisturb`) |
| Energy per operation | fJ from integration | T2 | Tested |

### 5.14 Research Trace and Confidence Outputs

| Output | Formula / Invariant | Tier | Current State |
|--------|---------------------|------|---------------|
| Full signal chain trace | DAC->Array->TIA->ADC->Classifier with per-stage values | T2 | **Not implemented** |
| Per-stage uncertainty propagation | Error bars at each stage | T2 | **Not implemented (gap G2)** |
| Confidence ledger | Parameter provenance: measured/estimated/placeholder | T1 | **Not implemented (gap G1)** |
| End-to-end uncertainty budget | Sum in quadrature across stages | T2 | **Not implemented** |

---

## 6. Physics Invariants (What MUST Hold in Circuits)

### Fundamental Conservation Laws

1. **Kirchhoff's Current Law (KCL)**: Sum of currents at each node = 0 (within solver tolerance)
2. **Kirchhoff's Voltage Law (KVL)**: Sum of voltages around any loop = 0
3. **Power Conservation**: Total dissipated power P = sum(I^2 R) across all resistive elements
4. **Energy Budget Closure**: Input energy (DAC + driver) >= Output energy (array dissipation + ADC + TIA)

### Device-Level Invariants

5. **Conductance Bounds**: Gmin <= G(cell) <= Gmax for all cells, all times
6. **Voltage Limits**: |V(cell)| <= material breakdown voltage
7. **Current Sign**: I = G x V (positive conductance, correct sign convention)
8. **Level Quantization**: Weight levels in {0, 1, ..., quantLevels-1}

### Operation-Specific Invariants

9. **READ Immutability**: arrayWeights[i][j] unchanged before/after read operation
10. **COMPUTE Immutability**: arrayWeights[i][j] unchanged before/after compute operation
11. **WRITE Locality** (architecture-dependent):
    - **0T1R**: Only cells on selected row OR selected column may change (half-select pattern)
    - **1T1R**: Only cell at (selected row, selected column) may change significantly (+/-1 level max elsewhere)
    - **2T1R**: Strictly only (row, col) changes; all others frozen
12. **WRITE Monotonicity**: For ascending program, level[n+1] >= level[n] (within +/-1 due to quantization)
13. **Broadcast Guard**: If > K non-target cells change by > L levels in one write iteration, FAIL (K=3, L=3 initial defaults)

### Solver-Specific Invariants

14. **Tier-A Dense Solve**: Residual ||Ax - b|| / ||b|| < 1e-12 (direct solve)
15. **Tier-B PCG Convergence**: ||r|| / ||b|| < 1e-8 within 4000 iterations
16. **Ideal Mode Exactness**: I(row) = sum(G(row,col) x V(col)) (no IR drop, no coupling)

### Peripheral Circuit Invariants

17. **DAC Monotonicity**: code[n+1] -> voltage[n+1] >= voltage[n]
18. **ADC Monotonicity**: voltage[n+1] >= voltage[n] -> code[n+1] >= code[n]
19. **TIA Linearity**: V_out proportional to I_in (within saturation bounds)
20. **Noise Floor**: Integrated noise power <= kT/C thermal limit + specified excess noise

### Thermodynamic Consistency (CRITICAL GAP)

21. **Non-Negative Dissipation**: P(element) = I^2 x R >= 0 for all resistive elements
22. **Energy Monotonicity**: Cumulative energy dissipated is non-decreasing over time
23. **Power Budget**: sum(P_cells) + sum(P_wires) + P(peripherals) = P(total) within 1% accounting error

---

## 7. Validation Tiers

### Validation Pyramid

```
                      +------------------------+
                      |   Tier 3: Enhanced     |  Nightly / Release
                      |   Monte Carlo, PVT,    |
                      |   Pareto, Scaling,     |
                      |   Retention, Endurance |
                   +--+------------------------+--+
                   |   Tier 2: Quantitative       |  PR + Nightly
                   |   INL/DNL, ENOB, SNR,        |
                   |   SPICE xval, uncertainty,   |
                   |   BER, signal chain trace    |
                +--+------------------------------+--+
                |     Tier 1: Critical Gates         |  Every PR
                |     Immutability, locality,        |
                |     convergence, determinism,      |
                |     Kirchhoff, NaN/Inf,            |
                |     broadcast guard, provenance    |
             +--+------------------------------------+--+
             |     Tier 0: Unit Correctness              |  Every PR
             |     Fast deterministic invariants in      |
             |     arraysim and gui packages (< 30 sec)  |
             +-------------------------------------------+
```

### Tier 1: Critical (Must Pass for Research Release)

These are **hard gates**. Any failure blocks the release. Run on **every PR**.

| ID | Check | Assertion | Test Location |
|----|-------|-----------|---------------|
| T1-01 | MVM linearity | Ideal: exact dot product. Tier-A: PSNR > 30 dB for 30 levels | `arraysim/current_validation_test.go` + new PSNR test |
| T1-02 | ISPP convergence | Success rate > 95% across full level range (0-29) | `arraysim/array_ispp_test.go` + new |
| T1-03 | READ immutability | `arrayWeights` bitwise identical before/after read | `arraysim/current_validation_test.go` |
| T1-04 | COMPUTE immutability | `arrayWeights` bitwise identical before/after compute | `arraysim/current_validation_test.go` |
| T1-05 | WRITE locality (0T1R) | Non-target disturbance only on selected row/col half-select set | `gui/tab_unified_halfselect_residue_test.go` |
| T1-06 | WRITE locality (1T1R/2T1R) | Non-target delta_level < 1 level | New test needed |
| T1-07 | Broadcast guard | If > K non-target cells change by > L levels, FAIL (K=3, L=3) | **New test needed** |
| T1-08 | Kirchhoff compliance (KCL) | Residual < 1e-6 at every node | `arraysim/refsolve_dense_test.go` |
| T1-09 | Kirchhoff compliance (KVL) | Sum of voltages around any loop = 0 | New test needed |
| T1-10 | No NaN/Inf | All physics outputs are finite | Sweep all outputs in regression |
| T1-11 | Determinism | Fixed seed + config => bitwise identical results | New regression test |
| T1-12 | Confidence ledger present | Every regression artifact includes parameter provenance tags | New (artifact schema) |
| T1-13 | Material-bound test ID | Every test case binds explicit material; no implicit defaults | Convention enforcement |
| T1-14 | Power conservation | |P_in - P_dissipated| / P_in < 1% | New test needed |
| T1-15 | Non-negative dissipation | P(cell) = I^2 x R >= 0 for all cells | New test needed |
| T1-16 | Conductance bounds | Gmin <= G(cell) <= Gmax always | Extend `current_validation_test.go` |

### Tier 2: Important (Should Pass for Publication)

These tests validate quantitative accuracy and signal chain integrity. Run on every PR (fast subset) and nightly (full matrix).

| ID | Check | Assertion | Test Location |
|----|-------|-----------|---------------|
| T2-01 | IR drop Ohm's law | delta_V = I x R_wire within 1% for known wire resistance | `arraysim/tier_a_test.go` + new scaling test |
| T2-02 | IR drop scaling | 5x resistance => ~5x drop (proportional) | New |
| T2-03 | Worst-case IR drop | Corner cell has maximum drop; verify proportional to path length | New |
| T2-04 | Sneak path 3-cell model | Analytical sneak current matches solver for 2x2 array | New |
| T2-05 | Sneak path negative SNR | Flag failure mode when sneak > signal | New |
| T2-06 | DAC INL | < 1 LSB across full code range | New (from `dac.AnalyzeINLDNL()`) |
| T2-07 | DAC DNL | < 1 LSB; no missing codes (monotonicity) | New (from `dac.AnalyzeINLDNL()`) |
| T2-08 | ADC ENOB | >= nominal bits - 1 | New (from `adc.ENOB()`) |
| T2-09 | ADC missing codes | DNL > -1 LSB everywhere | New (from `adc.AnalyzeINLDNL()`) |
| T2-10 | ADC quantization noise | 1/sqrt(12) LSB theoretical | New |
| T2-11 | TIA SNR | > 40 dB for representative compute current | `gui/tab_unified_snr_test.go` + new |
| T2-12 | TIA dynamic range | > 60 dB (min detectable to max linear) | New |
| T2-13 | TIA settling | < 80 ns for read path | New |
| T2-14 | TIA transimpedance | V_out = Rf x I_in +/-2% across linear range | New |
| T2-15 | Transient energy | Within 2x of published FeCIM estimates (tagged as estimated) | `arraysim/transient_test.go` |
| T2-16 | SPICE cross-validation | Node voltages match within 1% for golden vector set | `arraysim/tier_b_spice_golden_test.go` |
| T2-17 | Solver condition number | Tier-B condition number logged; warn if > 1e8, fail if > 1e12 | New |
| T2-18 | Solver convergence diagnostics | Iteration count, residual norm, tolerance recorded | New |
| T2-19 | Statistical confidence | Bootstrap 95% CI on MVM error; KS test p > 0.05 for distributions | New |
| T2-20 | ISPP pulse histogram | Distribution of pulses per level; flag outliers > 3 sigma | New |
| T2-21 | Level spacing uniformity | +/-5% from linear across 30 quantization levels | New |
| T2-22 | Write-verify stats | Regression-track pulses, overshoots, energy per cell | New |
| T2-23 | Full signal chain trace | DAC->Array->TIA->ADC values recorded per stage | New |
| T2-24 | Per-stage uncertainty | Error propagation through signal chain | New |
| T2-25 | MNA conditioning | Log condition number; warn if > 1e8, fail if > 1e12 | New |
| T2-26 | Read repeatability | 100 consecutive reads, StdDev < 1% of mean | New |
| T2-27 | Read margin BER | BER from Q-function, Q(deltaV / (2 sigma)) < 1e-6 | New |
| T2-28 | MVM accuracy golden | Known G, known V vs floating-point baseline < 5% error (Tier-A) | New |
| T2-29 | BER target (MNIST) | BER < 5% for inference patterns | New |
| T2-30 | Signed compute | Bipolar weights + input, four-quadrant correct | New |
| T2-31 | ISPP engine parity | Level-based vs L-K ODE converge to same level +/-1 | New |
| T2-32 | Wire dissipation | P_wire = sum(I_wire^2 x R_wire), matches nodal solve | New |
| T2-33 | Energy monotonicity | Cumulative E(t) non-decreasing during ISPP | New |
| T2-34 | SenseChain linearity | TIA -> ADC end-to-end R^2 > 0.98 | New |
| T2-35 | Endurance drift power-law | G(t) = G_0 x (t/t_0)^(-nu), R^2 > 0.99 | Extend existing |

### Tier 3: Enhanced (Nightly / Release Gate)

Extended validation for design space exploration and robustness.

| ID | Check | Assertion | Test Location |
|----|-------|-----------|---------------|
| T3-01 | Process variation Monte Carlo | Yield prediction with N >= 1000, convergence documented | Extend `arraysim/process_variation_mc_test.go` |
| T3-02 | Endurance drift accuracy | Power-law fit R^2 > 0.99; FeCIM nu ~ 0.001 vs RRAM nu ~ 0.05 | Extend `arraysim/endurance_accuracy_test.go` |
| T3-03 | Temperature-dependent models | Drift Arrhenius scaling; peripheral PVT corners (-40 to 125 deg C) | New |
| T3-04 | Design space Pareto fronts | Energy vs accuracy vs array size tradeoff curves | New |
| T3-05 | Full 3-arch x 3-tier x 4-size matrix | 36-point coverage (all combinations) | New |
| T3-06 | Peripheral PVT corners | Fast/Slow/Typical within +/-10% of nominal | New |
| T3-07 | Sneak path scaling | SNR degradation quantified vs array size (4x4 to 32x32) | New |
| T3-08 | Retention projection | 10-year @ 85 deg C extrapolation with Arrhenius fit and CI | New |
| T3-09 | DAC/ADC power scaling | Linear with sample rate; documented against published estimates | New |
| T3-10 | Array energy dominance | Array energy >> peripheral for 16x16 and larger | New |
| T3-11 | MVM throughput scaling | Documented throughput vs array size relationship | New |
| T3-12 | Half-select stress retention | 0T1R, V/2 stress, 10-year extrapolation, deltaP < 10% Pr | New |
| T3-13 | Read stress retention | 10^6 low-V reads, deltaG/G < 1% | New |
| T3-14 | Write disturb accumulation | N-cycle deltaP model, deltaP(N=1000) < 5% Pr | New |
| T3-15 | Spatial disturb patterns | Measure deltaP vs distance from target cell | New |
| T3-16 | Supply voltage sweep | VDD +/-10%, functionality preserved | New |
| T3-17 | Spatial process correlation | Correlated Ec variation (lambda=5 cells), yield > 95% | New |
| T3-18 | Standard array patterns | Checkerboard, walking ones/zeros, March C-, worst-case sneak | New |
| T3-19 | Code width uniformity (DAC) | Histogram of step sizes, max deviation documented | New |
| T3-20 | Input-referred noise (TIA) | Vs kT/C thermal limit | New |

---

## 8. Required Test Matrix

### Dimensions

| Axis | Values | PR Gate | Nightly |
|------|--------|---------|---------|
| Architecture | 0T1R, 1T1R, 2T1R | 0T1R, 1T1R | All 3 |
| Coupling tier | Ideal, Tier-A, Tier-B | Ideal, Tier-A | All 3 |
| Array size | 2x2, 8x8, 16x16, 32x32 | 2x2, 8x8 | All 4 |
| Material | (see sweep sets) | 3 materials | 9 materials |
| Quantization | 30 levels baseline | 30 | 30 + optional sensitivity |
| Operations | READ, WRITE, COMPUTE, full WRC cycle | All 4 | All 4 |
| Target levels | Low (2), mid (15), high (28), random set | Mid only | All 4 |
| Input vectors | Sparse, dense, signed pattern | Dense | All 3 |
| Test patterns | Standard patterns (see section 8c) | Checkerboard, All-ones | All 10 |

### Material Sweep Sets

**PR gate (fast, required):**
- `fecim_hzo`
- `literature_superlattice`
- `default_hzo`

**Nightly/release (extended):**
- `fecim_hzo`
- `fecim_hzo_target`
- `default_hzo`
- `literature_superlattice`
- `cryogenic_hzo`
- `hzo_standard_32`
- `hzo_ftj_140`
- `hzo_custom_14`
- `alscn`

### PR Gate Fast Matrix (Minimum Viable)

```
Architectures:  [0T1R, 1T1R]
Coupling:       [Ideal, TierA]
Sizes:          [2x2, 8x8]
Materials:      [fecim_hzo, literature_superlattice, default_hzo]
Operations:     [READ, WRITE, COMPUTE, WRC]
Target:         [15]  (mid-range)
Patterns:       [Checkerboard, All-ones]
```

Total: 2 x 2 x 2 x 3 x 4 x 1 x 2 = **192 test points** (fast, < 5 minutes)

### Nightly Extended Matrix

```
Architectures:  [0T1R, 1T1R, 2T1R]
Coupling:       [Ideal, TierA, TierB]
Sizes:          [2x2, 8x8, 16x16, 32x32]
Materials:      [9 materials]
Operations:     [READ, WRITE, COMPUTE, WRC]
Targets:        [2, 15, 28, random]
Input vectors:  [sparse, dense, signed]
Patterns:       [10 standard patterns]
```

Full combinatorial: 3 x 3 x 4 x 9 x 4 x 4 x 3 x 10 = ~155,520 (sampled DOE acceptable; document sampling strategy and use Latin hypercube for nightly).

### Architecture-Specific Coverage

| Architecture | Locality Rule | Half-Select | Disturb Model | Cell Area |
|--------------|---------------|-------------|---------------|-----------|
| **0T1R** | Row OR column | V/2 on row+col | High disturb, accumulates | 4F^2 |
| **1T1R** | Target +/-1 level elsewhere | Transistor-isolated | 20x reduction | 8-12F^2 |
| **2T1R** | Strict target-only | Dual isolation | Minimal disturb | Larger |

### Standard Array Test Patterns

| Pattern | Purpose | Implementation |
|---------|---------|----------------|
| **Checkerboard** | Cell coupling, neighbor influence | Alternating 0/29 in chess pattern |
| **All-Ones** | Maximum sneak path, IR drop stress | All cells = 29 (Gmax) |
| **All-Zeros** | Minimum current, noise floor | All cells = 0 (Gmin) |
| **Walking Ones** | Decoder faults, cell isolation | Single 1 walks through all positions |
| **Walking Zeros** | Stuck-at faults | Single 0 walks through field of 1s |
| **Diagonal** | Symmetric coupling | Level gradient along diagonal |
| **Row/Column Stripe** | WL/BL coupling | Alternating row or column activation |
| **Random** | Statistical coverage | Uniform random levels, seeded |
| **Worst-Case Sneak** | Sensing margin validation | Specific pattern maximizes sneak current |
| **March C-** | 10n comprehensive fault coverage | Standard memory test algorithm |

---

## 9. Existing Coverage Baseline (Keep and Extend)

### Operation Invariants Already Present

| Test File | What It Covers | Tier |
|-----------|----------------|------|
| `arraysim/current_validation_test.go` | READ and COMPUTE do not mutate state; single-cell current exact | T1 |
| `gui/device_state_read_coupling_test.go` | Signed read VI behavior, end-to-end read chain checks | T1 |
| `gui/device_state_halfselect_dac_arraysim_test.go` | DAC-quantized write voltage mapped to target and half-select Vcell | T1 |
| `gui/device_state_ispp_coupled_write_test.go` | Coupled voltage monotonicity, IR-drop-bounded update | T1 |
| `gui/tab_unified_halfselect_residue_test.go` | Half-select residue accumulation and row/column disturb pattern | T1 |
| `gui/unified_buttons_test.go` | Headless action-flow coverage for READ/WRITE/COMPUTE controls | T1 |
| `gui/tab_unified_snr_test.go` | SNR degrades with noise (TIA sense chain) | T2 |
| `gui/headless_gui_physics_parity_test.go` | GUI dispatch vs headless harness physics comparison | T1 |
| `arraysim/tier_a_test.go` | Passive half-select, Rwire=0 matches ideal, active row masking, dense reference match, selector mitigation | T1/T2 |
| `arraysim/tier_b_spice_golden_test.go` | Tier-B agrees with SPICE golden vectors (2e-11 tol) | T2 |
| `arraysim/tier_b_boundary_selector_test.go` | Boundary termination, selector off-leakage | T2 |
| `arraysim/refsolve_dense_test.go` | KCL residual < 1e-6 for 2x2 and 4x4 | T1 |
| `arraysim/transient_test.go` | Complete switching at 100 ns, sub-coercive incomplete, energy per cell, read non-disturb | T2 |
| `arraysim/transient_characterization_test.go` | Extracts timing and energy from transient | T2 |
| `arraysim/array_ispp_test.go` | 8x8 checkerboard, 16x16 gradient, disturb non-flip, scaling not quartic | T1/T2 |
| `arraysim/endurance_accuracy_test.go` | Monotonic degradation, clamps negative cycles | T2 |
| `arraysim/process_variation_mc_test.go` | Yield and stats, tight margin lowers yield | T3 |
| `arraysim/read_margin_analysis_test.go` | Coupled solver margin reflection, noisy < deterministic | T1/T2 |
| `arraysim/spice_export_test.go` | SPICE netlist includes cells, wires, and peripherals | T2 |
| `arraysim/selector_masks_test.go` | Read vs write selector path | T1 |
| `arraysim/write_verify_device_loop_test.go` | Write-verify convergence and progress halt | T1 |
| `arraysim/program_scheduler_test.go` | Adaptive schedule less disturb than row-major | T2 |
| `arraysim/mixed_precision_planner_test.go` | Valid config found, no-feasible-point handled | T3 |

### Gaps To Close

| Gap | Tier | Priority | New Test Needed |
|-----|------|----------|-----------------|
| Broadcast guard on single-cell write | T1 | **Critical** | `headless_rw_compute_regression_test.go` |
| PSNR metric for MVM quantization quality | T1 | **Critical** | `quantization_quality_test.go` |
| Confidence ledger in artifacts | T1 | **Critical** | Artifact schema change |
| Deterministic end-to-end WRC artifacts | T1 | **Critical** | `headless_rw_compute_regression_test.go` |
| Power conservation / thermodynamic check | T1 | **Critical** | `thermodynamics_array_power_test.go` |
| DAC/ADC INL/DNL as automated test | T2 | High | `peripheral_characterization_test.go` |
| ADC ENOB as automated test | T2 | High | `peripheral_characterization_test.go` |
| TIA dynamic range and settling tests | T2 | High | `peripheral_characterization_test.go` |
| Solver condition number (Tier-B) | T2 | High | `tier_b_diagnostics_test.go` |
| Statistical confidence intervals | T2 | High | `statistical_validation_test.go` |
| Signal chain trace with per-stage uncertainty | T2 | High | `signal_chain_trace_test.go` |
| Level spacing uniformity | T2 | High | `quantization_quality_test.go` |
| BER measurement for MNIST | T2 | High | `compute_ber_measurement_test.go` |
| Read margin with BER curves | T2 | High | `read_margin_ber_test.go` |
| IR drop scaling proportionality | T2 | Medium | `ir_drop_scaling_test.go` |
| Sneak path analytical model | T2 | Medium | `sneak_path_analytical_test.go` |
| Full 3x3x4 matrix | T3 | Medium | DOE runner |
| PVT corner characterization | T3 | Medium | `pvt_corner_test.go` |
| Retention and endurance models | T3 | Medium | Multiple new files |
| Standard array patterns (March C- etc.) | T3 | Medium | Pattern test files |

---

## 10. Regression Artifact Schema (v2.0)

All regression runs emit JSON artifacts under `output/regression/module4/<timestamp>/`.

```json
{
  "suite": "module4-research-regression",
  "version": "2.0",
  "git_hash": "abc1234",
  "timestamp": "2026-02-13T12:00:00Z",
  "seed": 42,
  "env_vars": {
    "FECIM_M4_EXTENDED": "1",
    "GO_TEST_TIMEOUT": "10m"
  },
  "config": {
    "architecture": "1T1R",
    "coupling_tier": "A",
    "array_size": [8, 8],
    "material": "fecim_hzo",
    "quant_levels": 30
  },
  "material_snapshot": {
    "name": "fecim_hzo",
    "Ec_V_m": 1.0e6,
    "Ps_uC_cm2": 25.0,
    "Pr_uC_cm2": 20.0,
    "thickness_nm": 10.0,
    "Gmin_S": 1.0e-7,
    "Gmax_S": 1.0e-5,
    "quant_levels": 30
  },
  "read": {
    "weights_unchanged": true,
    "max_current_A": 1.2e-6,
    "min_margin_V": 0.05,
    "snr_db": 45.2,
    "tia_output_V": [0.012, 0.024, 0.036],
    "adc_codes": [3, 7, 11],
    "signal_chain_trace": {
      "dac_voltages_V": [0.1],
      "cell_voltages_V": [[0.098, 0.099]],
      "cell_currents_A": [[1.2e-6, 2.4e-6]],
      "tia_voltages_V": [0.012, 0.024],
      "adc_codes": [3, 7]
    }
  },
  "write": {
    "target_level": 15,
    "final_level": 15,
    "pulses": 4,
    "overshoots": 0,
    "max_nontarget_delta": 0,
    "energy_fJ": 31.2,
    "write_verify_stats": {
      "total_pulses": 4,
      "success": true,
      "pulse_histogram": {"1": 0, "2": 1, "3": 2, "4": 1},
      "overshoot_count": 0,
      "final_error_levels": 0
    },
    "nontarget_delta_map": {
      "max_delta": 0,
      "cells_changed": 0,
      "broadcast_guard_pass": true
    }
  },
  "compute": {
    "weights_unchanged": true,
    "psnr_db": 35.4,
    "max_error_pct": 2.1,
    "ir_drop_max_V": 0.012,
    "mvm_result": [1.2e-5, 2.4e-5, 3.6e-5],
    "ideal_result": [1.2e-5, 2.5e-5, 3.7e-5],
    "level_spacing_uniformity_pct": 3.2
  },
  "peripheral": {
    "dac_inl_lsb": 0.3,
    "dac_dnl_lsb": 0.2,
    "dac_monotonic": true,
    "adc_enob": 4.8,
    "adc_missing_codes": false,
    "tia_snr_db": 48.0,
    "tia_dynamic_range_db": 62.0,
    "tia_settling_ns": 45.0
  },
  "solver": {
    "coupling_tier": "A",
    "kcl_residual_max": 1.2e-8,
    "condition_number": null,
    "iterations": null,
    "convergence_tolerance": null
  },
  "thermodynamics": {
    "power_in_W": 1.5e-3,
    "power_dissipated_W": 1.49e-3,
    "power_conservation_error_pct": 0.67,
    "all_cell_dissipation_nonneg": true,
    "wire_dissipation_W": 2.1e-5
  },
  "confidence": {
    "measured_params": [],
    "estimated_params": [
      "Ec", "Ps", "Pr", "thickness", "Gmin", "Gmax",
      "tia_rf", "adc_bits", "dac_bits"
    ],
    "placeholder_params": [
      "tia_settling_ns", "energy_fJ", "drift_nu"
    ],
    "notes": "All parameters are simulation defaults. No experimental calibration."
  },
  "timing": {
    "test_duration_ms": 1234,
    "solver_time_ms": 456
  }
}
```

**Key fields in v2.0 (vs v1.0):**
- `material_snapshot`: Full material parameters for reproducibility.
- `signal_chain_trace`: Per-stage values through DAC->Array->TIA->ADC.
- `solver`: Condition number, iteration count, residual for Tier-B diagnostics.
- `confidence`: Parameter provenance classification (measured/estimated/placeholder).
- `nontarget_delta_map`: Explicit broadcast guard tracking.
- `peripheral`: Full characterization (INL/DNL/ENOB/SNR/settling/dynamic range).
- `thermodynamics`: Power conservation and dissipation validation.

---

## 11. Analysis Function Golden Data

Each of the 9 analysis functions requires versioned golden data for regression detection.

### Analysis Functions

| Function | Test Coverage | Golden Data Needed |
|----------|---------------|-------------------|
| **ReadMarginAnalysis** | All 30 levels, MC 100 samples | Voltage separation, noise sigma |
| **CharacterizeTransientResult** | Write/read/compute timing | Energy (fJ), latency (ns) |
| **BuildDesignSpacePoints** | Array size x ADC bits x devices | Pareto front coordinates |
| **RunProcessVariationMC** | Gaussian Ec/Pr, seed-based | Yield %, mean/sigma per metric |
| **SimulateEnduranceAccuracy** | Drift model, N cycles | Accuracy vs cycle number |
| **AnalyzeINLDNL** | DAC/ADC, all codes, PVT corners | INL/DNL per code, max values |
| **AnalyzeTiming** | 9 timing metrics | DAC settle to max throughput |
| **AnalyzePower** | 13 energy/power metrics | Per-component breakdown |
| **ComputeTransferFunction** | 30-level DAC->ADC chain | End-to-end transfer curve |

### Golden Data Directory Structure

```
validation/testdata/module4-analysis/
  README.md
  read_margin/
    0T1R_TierA_8x8_fecim_hzo_v1.json
    1T1R_Ideal_8x8_literature_superlattice_v1.json
    ...
  transient/
    write_ispp_8x8_fecim_hzo_v1.json
    ...
  transfer_function/
    30level_chain_fecim_hzo_v1.json
    ...
  inl_dnl/
    dac_8bit_nominal_v1.json
    adc_5bit_nominal_v1.json
    ...
  design_space/
    pareto_front_baseline_v1.json
    ...
  process_variation/
    mc_1000_samples_seed42_v1.json
    ...
  endurance/
    drift_1e9_cycles_v1.json
    ...
  timing/
    9metrics_baseline_v1.json
    ...
  power/
    13metrics_breakdown_v1.json
    ...
```

### Golden Data Update Policy

- **When to update**: Physics model improvements (with justification), bug fixes, material parameter updates.
- **How to update**: Set `FECIM_UPDATE_MODULE4_GOLDEN=1`, run tests, review diffs carefully, commit with detailed explanation.
- **Versioning**: Increment `_vN` suffix; keep old version during transition periods.

---

## 12. SPICE Co-Verification

### Purpose

Validate Module 4 circuit simulation against industry-standard SPICE tools to establish bounded agreement.

### Workflow

1. **Export Netlist**: Module 4 -> SPICE netlist (`.cir` file) via `ExportCrossbarSPICE()`
2. **Run Reference Simulation**: NgSpice / Xyce / HSPICE
3. **Import Results**: Parse SPICE output
4. **Compare**: Module 4 solver vs SPICE solver, per-node
5. **Report**: Residual analysis with acceptance bands

### Reference Simulators

| Simulator | Version | Use Case |
|-----------|---------|----------|
| **NgSpice** | 40+ | Open-source baseline |
| **Xyce** | 7.x | Parallel solver, large arrays |
| **HSPICE** | 2023+ | Industry reference (if available) |

### Co-Verification Test Matrix

| Test Case | Array Size | Coupling | Acceptance |
|-----------|------------|----------|------------|
| DC Operating Point | 8x8 | Tier-A | Node voltage error < 1% |
| Transient Response | 4x4 | Tier-A | Current waveform error < 5% |
| IR Drop Validation | 16x16 | Tier-A | Voltage drop matches +/-2% |
| Large Array Scaling | 64x64 | Tier-B | Converged solution < 10% error |

### Acceptance Criteria

- **Per-node voltage**: |V_module4 - V_spice| / V_spice < 1% for 90% of nodes, < 5% for 100%
- **Current**: |I_module4 - I_spice| / I_spice < 5% for row/column currents
- **Power**: |P_module4 - P_spice| / P_spice < 5% for total dissipated power

### Reference Fixture Library

Location: `validation/testdata/spice-fixtures/`

Each fixture contains: netlist (`.cir`), control script, expected output, metadata (simulator version, model deck hash), tolerance rationale. Currently tested via `arraysim/tier_b_spice_golden_test.go` with golden vectors at 2e-11 tolerance.

---

## 13. Test Layers

### Layer 0: Solver Validation (Foundation)

**Purpose**: Validate numerical solver correctness independent of device models.

| Test | Coverage | Acceptance |
|------|----------|------------|
| Tier-A Dense Solver | 2x2 to 32x32, known G matrix, known V input | ||Ax - b|| / ||b|| < 1e-12 |
| Tier-B PCG Convergence | 32x32 to 128x128, various sparsity patterns | Converges in <4000 iter, residual <1e-8 |
| Ideal vs Tier-A Match | Same array, Ideal vs Tier-A with Rwire=0 | Current agreement <0.1% |
| KCL Validation | Every node, every coupling mode | sum(I_node) < 1e-15 A |
| KVL Validation | Sample loops in mesh | sum(V_loop) < 1e-9 V |
| Power Conservation | All modes, all architectures | |P_in - P_dissipated| / P_in < 1% |
| Condition Number | Tier-B MNA matrix | Logged; warn >1e8, fail >1e12 |

### Layer 1: Peripheral Circuit Validation

**Purpose**: Validate DAC, ADC, TIA, SenseChain against published peripheral specs.

| Test | Coverage | Acceptance |
|------|----------|------------|
| DAC Monotonicity | 4-8 bit, all codes | No missing codes, DNL < 0.5 LSB |
| DAC INL | Full-scale range | INL < 1 LSB at nominal PVT |
| ADC Monotonicity | 5-8 bit SAR | No missing codes |
| ADC ENOB | All bit widths | >= nominal bits - 1 |
| ADC Noise Floor | 100 samples per code | RMS noise < 0.5 LSB |
| ADC Quantization Noise | Theoretical check | 1/sqrt(12) LSB |
| TIA Transimpedance | 1pA to 1mA input sweep | V_out = Rf x I_in +/-2% |
| TIA Noise | Integrated over BW | < 1pA/sqrt(Hz) specified noise |
| TIA SNR | Representative currents | > 40 dB |
| TIA Dynamic Range | Min to max | > 60 dB |
| TIA Settling | Step response | < 80 ns |
| SenseChain Linearity | TIA -> ADC end-to-end | R^2 > 0.98 for linear range |
| PVT Variation | -40 to 125 deg C, +/-10% supply | All specs hold with margin |

### Layer 2: Operation Validation (Core Functionality)

**Purpose**: Validate READ, WRITE, COMPUTE operations follow invariants and produce correct results.

#### READ Operation Tests

| Test | Scenario | Acceptance |
|------|----------|------------|
| Immutability | Before/after weight snapshot | 0 cells changed |
| Repeatability | 100 consecutive reads | Same output +/- noise floor |
| Signed Read | Bipolar mode, positive/negative cells | Correct current sign |
| Read Margin | All 30 levels, MC noise 100 samples | Level separation > 3 sigma noise |
| Read Margin BER | Q-function for adjacent levels | BER < 1e-6 at nominal PVT |
| Half-Select Stress | 0T1R, unselected cells, 10^6 reads | deltaG/G < 1% drift |
| Coupling Consistency | Ideal vs Tier-A vs Tier-B | Results bounded within expected IR drop |
| Signal Chain Trace | DAC->Array->TIA->ADC per-stage | All values finite, consistent |

#### WRITE Operation Tests

| Test | Scenario | Acceptance |
|------|----------|------------|
| Target Convergence | Single-cell ISPP, all 30 levels | |final - target| <= 1 level |
| Locality (0T1R) | Write (3,5), check all cells | Only row=3 OR col=5 change |
| Locality (1T1R) | Write (3,5), check all cells | Only (3,5) changes >1 level |
| Locality (2T1R) | Write (3,5), check all cells | Only (3,5) changes |
| Broadcast Guard | Write (3,5), count large delta cells | < K cells with > L delta |
| Half-Select Disturb | 0T1R, N writes, accumulation | deltaP(unselected) < threshold |
| Write Monotonicity | Ascending program 0->29 | level[n+1] >= level[n] - 1 |
| Coupled Write Voltage | Tier-A write, IR drop bounds | V_eff(target) > V_eff(half-select) |
| ISPP Engine Parity | Level-based vs L-K ODE | Both converge to same level +/-1 |
| Write-Verify Stats | Pulses, overshoots, energy | Regression-tracked in artifact |

#### COMPUTE Operation Tests

| Test | Scenario | Acceptance |
|------|----------|------------|
| Immutability | Before/after weight snapshot | 0 cells changed |
| MVM Accuracy (Ideal) | Known G, known V, analytic result | < 0.1% error |
| MVM Accuracy (Tier-A) | 8x8, random weights, random input | < 5% vs floating-point |
| MVM PSNR | 30-level quantized vs ideal | > 30 dB |
| BER Target | MNIST/CIFAR patterns | BER < 5% (CIM standard) |
| Repeatability | Same input, 100 trials | StdDev < 1% of mean |
| Signed Compute | Bipolar weights, bipolar input | Correct sign, four-quadrant |
| Level Spacing Uniformity | 30 conductance levels | +/-5% from linear |

### Layer 3: Research-Grade Physics Validation

#### Thermodynamic Consistency (CRITICAL)

| Test | Coverage | Acceptance |
|------|----------|------------|
| Array Power Budget | All modes, all architectures | |P_in - P_out| / P_in < 1% |
| Cell Dissipation | P(cell) = I^2 R for each cell | All P >= 0 |
| Wire Dissipation | Wordline + bitline I^2 R | sum(P_wires) matches nodal solve |
| Energy Accounting | Cumulative energy over operation | E(t) monotonic increasing |
| Peripheral Power | DAC + ADC + TIA + drivers | Component sum = total +/-1% |

#### Retention in Array Context

| Test | Coverage | Acceptance |
|------|----------|------------|
| Half-Select Stress Retention | 0T1R, V/2 stress, 10-year extrapolation | deltaP < 10% of Pr |
| Read Stress Retention | Low V read, 10^6 cycles | deltaG/G < 1% |
| Write Stress Neighbor Retention | Adjacent cells to write target | deltaG < 5% Gmax |
| Thermal Retention | Arrhenius fit, 85 deg C -> 10 year | Extrapolated deltaP < spec |

#### Write Disturb Accumulation

| Test | Coverage | Acceptance |
|------|----------|------------|
| N-Cycle Half-Select deltaP | 0T1R, 10^3 writes, accumulation | deltaP(N=1000) < 5% of Pr |
| VDD/2 Inhibit Scheme | Standard half-select | Max deltaP on unselected |
| Spatial Disturb Pattern | Measure deltaP vs distance from target | Quantify neighbor influence |

#### PVT Variation

| Test | Coverage | Acceptance |
|------|----------|------------|
| Temperature Sweep | -40 to 125 deg C (automotive) | All metrics within spec |
| Supply Voltage Sweep | VDD +/-10% | Functionality preserved |
| Process Corners | Fast/Slow device, Fast/Slow wire | Monte Carlo yield > 95% |
| Spatial Correlation | Clustered variation (not i.i.d.) | Model correlation length |

### Layer 4: Statistical Validation

New layer for research-grade confidence. Addresses gap G4.

- **Bootstrap confidence intervals**: 95% CI on MVM error, ISPP pulse count, energy estimates.
- **Distribution tests**: KS test on claimed distributions (p > 0.05).
- **Convergence analysis**: Monte Carlo N-sufficiency (error stabilizes within tolerance).
- **Sample size justification**: Document why N=1000 (or other) is sufficient.
- **RMSE thresholds**: Per-metric documented thresholds with rationale.

### Layer 5: Cross-Model Validation

- Compare Module 4 outputs against reference solves (internal dense nodal via `ReferenceSolveDense`, external SPICE fixtures).
- Track trend metrics (mean error, P95 error, worst-case error) in JSON artifacts.
- Detect regressions: if error increases > 10% between commits, flag as warning.
- Release gate: cross-model agreement for all fixtures in library.

---

## 14. Pass/Fail Criteria (Research-Grade)

### Zero-Tolerance Failures (Immediate FAIL)

| Criterion | Tolerance | Notes |
|-----------|-----------|-------|
| NaN/Inf in any output | 0 occurrences | |
| READ mutates weights | 0 cells changed | Bitwise identical |
| COMPUTE mutates weights | 0 cells changed | Bitwise identical |
| Negative conductance | G >= 0 always | |
| KCL violation | > 1e-15 A | |
| Solver non-convergence | Tier-B > 4000 iterations | |

### Tier 1 Hard Gates (Any Failure Blocks Release)

| Criterion | Threshold |
|-----------|-----------|
| WRITE target error | |final - target| <= 1 level, or explicit bounded partial |
| Non-target constraints (0T1R) | Half-select row/col only |
| Non-target constraints (1T1R/2T1R) | delta_level < 1 for any non-target |
| Broadcast guard | < K cells with > L delta (K=3, L=3) |
| Determinism under fixed seed | Bitwise identical across runs |
| KCL residual | < 1e-6 |
| Power conservation | |P_in - P_dissipated| / P_in < 1% |
| Confidence ledger present | All params tagged |
| Material-specific verdicts | No missing material in sweep |
| Headless execution | No display env required |

### Tier 2 Quantitative Gates (Should Pass for Publication)

| Criterion | Threshold |
|-----------|-----------|
| MVM PSNR | > 30 dB for 30 levels |
| ISPP success rate | > 95% |
| DAC INL | < 1 LSB |
| DAC DNL | < 0.5 LSB (monotonic) |
| ADC ENOB | >= nominal bits - 1 |
| ADC DNL | > -1 LSB everywhere (no missing codes) |
| ADC noise | < 0.5 LSB RMS |
| TIA SNR | > 40 dB |
| TIA dynamic range | > 60 dB |
| TIA settling | < 80 ns |
| TIA linearity error | < 2% |
| SPICE cross-validation (90% nodes) | Voltage error < 1% |
| SPICE cross-validation (100% nodes) | Voltage error < 5% |
| Solver condition number | < 1e12 (warn at 1e8) |
| Level spacing uniformity | +/-5% from linear |
| Statistical KS test | p > 0.05 for claimed distributions |
| Bootstrap 95% CI | On MVM error, ISPP pulse count |
| Transient energy | Within 2x of published estimates |
| IR drop Ohm's law | delta_V proportional to I x R within 1% |
| Read repeatability | StdDev < 1% of mean (100 trials) |
| Read margin | Separation > 3 sigma noise |
| MVM error (Ideal) | < 0.1% vs analytic |
| MVM error (Tier-A) | < 5% vs floating-point |
| BER (MNIST) | < 5% |

### Tier 3 Extended Gates (Nightly/Release)

| Criterion | Threshold |
|-----------|-----------|
| DOE coverage completeness | All axis values represented |
| Monte Carlo yield | Reported with N >= 1000 |
| PVT corners | Within +/-10% of nominal, yield > 95% |
| Endurance power-law fit | R^2 > 0.99 |
| Retention (half-select) | deltaP @ 10yr, 85 deg C < 10% Pr |
| Retention (read stress) | deltaG/G @ 10^6 reads < 1% |
| Write disturb (N-cycle) | deltaP @ N=1000 < 5% Pr |
| Full matrix | 36-point coverage |
| Standard patterns | All 10 executed |
| 9 analysis functions | Golden data present |
| SPICE fixture library | >= 5 diverse cases |

### Determinism

| Test | Tolerance |
|------|-----------|
| Same seed, same config | Bit-exact match |
| Floating-point reproducibility | < 1e-9 relative error |

### Regression

- Golden data comparison within specified tolerances (per-metric)
- No degradation in metrics compared to baseline (within noise)
- Material-specific pass map complete (no missing verdicts)

---

## 15. Implementation Phases

### Phase 1: Critical Gates (Weeks 1-3)

**Goal**: Close T1 gaps. Every PR gated on these.

| Task | Deliverable | Priority |
|------|-------------|----------|
| Broadcast guard + WRC regression | `gui/headless_rw_compute_regression_test.go` | P0 |
| GUI/headless parity (extend) | `gui/headless_gui_physics_parity_test.go` | P0 |
| PSNR metric + level spacing | `arraysim/quantization_quality_test.go` | P0 |
| KCL/KVL validation | `arraysim/tier_a_solver_kcl_test.go` | P0 |
| Power conservation | `arraysim/thermodynamics_array_power_test.go` | P0 |
| Cell dissipation (P >= 0) | `arraysim/thermodynamics_cell_dissipation_test.go` | P0 |
| Wire loss validation | `arraysim/thermodynamics_wire_loss_test.go` | P0 |
| Energy monotonicity | `arraysim/thermodynamics_energy_monotonic_test.go` | P0 |
| Tier-B PCG convergence | `arraysim/tier_b_pcg_convergence_test.go` | P0 |
| Confidence ledger + artifact v2.0 | Artifact schema change | P0 |

**Acceptance**: All P0 tests pass before proceeding.

### Phase 2: Quantitative Validation (Weeks 4-6)

**Goal**: Close T2 gaps. Publication-quality evidence.

| Task | Deliverable | Priority |
|------|-------------|----------|
| Peripheral characterization | `gui/peripheral_characterization_test.go` | P1 |
| Signal chain trace | `gui/signal_chain_trace_test.go` | P1 |
| Solver diagnostics (condition number) | `arraysim/tier_b_diagnostics_test.go` | P1 |
| Statistical validation (CI, KS) | `arraysim/statistical_validation_test.go` | P1 |
| IR drop scaling proportionality | `arraysim/ir_drop_scaling_test.go` | P1 |
| Sneak path analytical model | `arraysim/sneak_path_analytical_test.go` | P1 |
| BER measurement | `gui/compute_ber_measurement_test.go` | P1 |
| Read margin + BER curves | `gui/read_margin_ber_test.go` | P1 |
| MVM accuracy golden | `gui/compute_mvm_accuracy_test.go` | P1 |
| ISPP pulse histogram | Extend `arraysim/array_ispp_test.go` | P1 |

### Phase 3: Research-Grade Physics (Weeks 7-9)

**Goal**: Close retention, disturb, PVT gaps.

| Task | Deliverable | Priority |
|------|-------------|----------|
| Half-select retention | `gui/retention_halfselect_stress_test.go` | P1 |
| Read stress retention | `gui/retention_read_stress_test.go` | P1 |
| Write neighbor retention | `gui/retention_write_neighbor_test.go` | P1 |
| Thermal Arrhenius fits | `gui/retention_thermal_arrhenius_test.go` | P1 |
| N-cycle write disturb | `gui/write_disturb_n_cycle_test.go` | P1 |
| Disturb inhibit schemes | `gui/write_disturb_inhibit_test.go` | P1 |
| Spatial disturb patterns | `gui/write_disturb_spatial_test.go` | P1 |
| Temperature sweep | `gui/pvt_temperature_sweep_test.go` | P1 |
| Supply variation | `gui/pvt_supply_variation_test.go` | P1 |
| Process corners | `gui/pvt_process_corners_test.go` | P1 |
| Spatial correlation MC | `gui/pvt_spatial_correlation_test.go` | P1 |
| Peripheral PVT sweep | `shared/peripherals/peripherals_pvt_sweep_test.go` | P1 |

### Phase 4: SPICE Co-Verification and Analysis Golden Data (Weeks 10-13)

**Goal**: External reference validation and regression baselines.

| Task | Deliverable | Priority |
|------|-------------|----------|
| SPICE result import/compare | `arraysim/spice_import.go` | P2 |
| NgSpice 2x2 fixture | `arraysim/spice_cov_verification_2x2_test.go` | P2 |
| NgSpice 8x8 fixture | `arraysim/spice_cov_verification_8x8_test.go` | P2 |
| Xyce large array | `arraysim/spice_cov_verification_fixtures_test.go` | P2 |
| ReadMarginAnalysis golden | `gui/analysis_read_margin_golden_test.go` | P2 |
| CharacterizeTransient golden | `gui/analysis_transient_test.go` | P2 |
| TransferFunction golden | `gui/analysis_transfer_function_test.go` | P2 |
| DesignSpace golden | `gui/analysis_design_space_test.go` | P2 |
| Endurance golden | `gui/analysis_endurance_test.go` | P2 |
| Peripheral noise validation | `shared/peripherals/peripherals_noise_validation_test.go` | P2 |
| INL/DNL regression golden | `shared/peripherals/peripherals_inl_dnl_regression_test.go` | P2 |

### Phase 5: Standard Patterns and CI Integration (Weeks 14-15)

**Goal**: Industry-standard patterns, automated CI pipeline.

| Task | Deliverable | Priority |
|------|-------------|----------|
| Standard pattern generators | `arraysim/standard_patterns.go` | P2 |
| Walking ones/zeros test | `arraysim/pattern_walking_test.go` | P2 |
| March C- algorithm test | `arraysim/pattern_march_test.go` | P2 |
| Worst-case sneak test | `arraysim/pattern_sneak_test.go` | P2 |
| Fast PR gate script | `scripts/run_module4_fast_gate.sh` | P0 |
| Nightly full matrix | `scripts/run_module4_nightly.sh` | P1 |
| SPICE co-verification | `scripts/run_module4_spice_cov.sh` | P2 |
| Uncertainty propagation framework | `arraysim/uncertainty.go` | P2 |
| Artifact diff tool | `scripts/compare_module4_artifacts.sh` | P2 |

---

## 16. CI Integration

### Fast PR Gate (Required, < 5 minutes)

```bash
#!/bin/bash
set -euo pipefail

env -u DISPLAY -u WAYLAND_DISPLAY \
  go test -tags=ci -count=1 -shuffle=off -trimpath -timeout 10m \
  -run 'HeadlessRWCompute|CurrentValidation|ReadCoupling|ThermodynamicsPower|SolverKCL|QuantizationQuality|PeripheralChar' \
  ./module4-circuits/pkg/arraysim ./module4-circuits/pkg/gui
```

Triggers: Every PR to main.

### Nightly Full Matrix (~2 hours)

```bash
#!/bin/bash
set -euo pipefail

export FECIM_M4_EXTENDED=1
export FECIM_M4_ALL_PATTERNS=1

env -u DISPLAY -u WAYLAND_DISPLAY \
  go test -tags=ci,extended -count=1 -timeout 3h \
  ./module4-circuits/...

# Analysis function regression
go test -tags=ci,extended -run 'Analysis.*Golden' ./module4-circuits/pkg/gui
```

Triggers: Daily @ 02:00 UTC.

### SPICE Co-Verification (Weekly)

```bash
#!/bin/bash
set -euo pipefail

if ! command -v ngspice &> /dev/null; then
  echo "NgSpice not found, skipping co-verification"
  exit 0
fi

go test -tags=ci,spice -run 'SPICECoV' ./module4-circuits/pkg/arraysim
go test -tags=ci,spice -run 'SPICEFixtures' ./module4-circuits/pkg/arraysim
```

Triggers: Weekly, Saturday @ 00:00 UTC.

### Race Detector Lane

```bash
env -u DISPLAY -u WAYLAND_DISPLAY GO_TEST_RACE_TIMEOUT=20m make test-race-ci
```

Triggers: Every PR to main.

### Artifact Storage

- `output/regression/module4/<timestamp>/` -- Test run artifacts
- `validation/testdata/module4-analysis/` -- Golden data (versioned)
- `validation/testdata/spice-fixtures/` -- SPICE reference outputs
- Retention: 30 days for PR runs, 1 year for nightly/release

---

## 17. Deliverables Summary

### New Test Files

| Deliverable | Location | Phase | Priority |
|-------------|----------|-------|----------|
| WRC regression + broadcast guard | `gui/headless_rw_compute_regression_test.go` | 1 | P0 |
| PSNR + level spacing | `arraysim/quantization_quality_test.go` | 1 | P0 |
| Solver KCL/KVL | `arraysim/tier_a_solver_kcl_test.go` | 1 | P0 |
| Power conservation | `arraysim/thermodynamics_array_power_test.go` | 1 | P0 |
| Cell dissipation | `arraysim/thermodynamics_cell_dissipation_test.go` | 1 | P0 |
| Wire loss | `arraysim/thermodynamics_wire_loss_test.go` | 1 | P0 |
| Energy monotonicity | `arraysim/thermodynamics_energy_monotonic_test.go` | 1 | P0 |
| PCG convergence | `arraysim/tier_b_pcg_convergence_test.go` | 1 | P0 |
| Peripheral characterization | `gui/peripheral_characterization_test.go` | 2 | P1 |
| Signal chain trace | `gui/signal_chain_trace_test.go` | 2 | P1 |
| Solver diagnostics | `arraysim/tier_b_diagnostics_test.go` | 2 | P1 |
| Statistical validation | `arraysim/statistical_validation_test.go` | 2 | P1 |
| IR drop scaling | `arraysim/ir_drop_scaling_test.go` | 2 | P1 |
| Sneak path analytical | `arraysim/sneak_path_analytical_test.go` | 2 | P1 |
| BER measurement | `gui/compute_ber_measurement_test.go` | 2 | P1 |
| Read margin BER | `gui/read_margin_ber_test.go` | 2 | P1 |
| MVM accuracy golden | `gui/compute_mvm_accuracy_test.go` | 2 | P1 |
| Half-select retention | `gui/retention_halfselect_stress_test.go` | 3 | P1 |
| Read stress retention | `gui/retention_read_stress_test.go` | 3 | P1 |
| Write neighbor retention | `gui/retention_write_neighbor_test.go` | 3 | P1 |
| Thermal Arrhenius | `gui/retention_thermal_arrhenius_test.go` | 3 | P1 |
| N-cycle write disturb | `gui/write_disturb_n_cycle_test.go` | 3 | P1 |
| Disturb inhibit | `gui/write_disturb_inhibit_test.go` | 3 | P1 |
| Spatial disturb | `gui/write_disturb_spatial_test.go` | 3 | P1 |
| Temperature sweep | `gui/pvt_temperature_sweep_test.go` | 3 | P1 |
| Supply variation | `gui/pvt_supply_variation_test.go` | 3 | P1 |
| Process corners | `gui/pvt_process_corners_test.go` | 3 | P1 |
| Spatial correlation | `gui/pvt_spatial_correlation_test.go` | 3 | P1 |
| Peripheral PVT sweep | `shared/peripherals/peripherals_pvt_sweep_test.go` | 3 | P1 |
| Peripheral noise | `shared/peripherals/peripherals_noise_validation_test.go` | 4 | P2 |
| INL/DNL regression | `shared/peripherals/peripherals_inl_dnl_regression_test.go` | 4 | P2 |
| SPICE import | `arraysim/spice_import.go` | 4 | P2 |
| SPICE 2x2 | `arraysim/spice_cov_verification_2x2_test.go` | 4 | P2 |
| SPICE 8x8 | `arraysim/spice_cov_verification_8x8_test.go` | 4 | P2 |
| SPICE fixtures | `arraysim/spice_cov_verification_fixtures_test.go` | 4 | P2 |
| Analysis golden (5 files) | `gui/analysis_*_test.go` | 4 | P2 |
| Standard patterns | `arraysim/standard_patterns.go` | 5 | P2 |
| Pattern walking | `arraysim/pattern_walking_test.go` | 5 | P2 |
| Pattern march | `arraysim/pattern_march_test.go` | 5 | P2 |
| Pattern sneak | `arraysim/pattern_sneak_test.go` | 5 | P2 |
| Uncertainty propagation | `arraysim/uncertainty.go` | 5 | P2 |

### Helper / Infrastructure Files

| File | Purpose |
|------|---------|
| `gui/headless_rw_compute_artifact.go` | Snapshot diff and metrics serializer (JSON artifact writer) |
| `arraysim/research_metrics.go` | PSNR, bootstrap CI, KS test, condition number helpers |
| `arraysim/confidence_ledger.go` | Parameter provenance tracking (measured/estimated/placeholder) |

### Script Deliverables

| Script | Purpose | Priority |
|--------|---------|----------|
| `scripts/run_module4_fast_gate.sh` | PR gate (< 5 min) | P0 |
| `scripts/run_module4_nightly.sh` | Nightly matrix | P1 |
| `scripts/run_module4_spice_cov.sh` | Weekly SPICE co-verification | P2 |
| `scripts/compare_module4_artifacts.sh` | Artifact diff and regression detection | P2 |

### Documentation Deliverables

| Document | Purpose |
|----------|---------|
| `validation/testdata/module4-analysis/README.md` | Golden data versioning policy |
| `validation/testdata/spice-fixtures/README.md` | Fixture creation instructions |
| `docs/testing/MODULE4_RESEARCH_GRADE.md` | Research-grade testing guide |
| `docs/circuits/THERMODYNAMIC_VALIDATION.md` | Power conservation methodology |
| `docs/circuits/SPICE_COVALIDATION.md` | SPICE co-verification protocol |

---

## 18. Traceability Matrix

Maps each physics invariant and research requirement to implementing test(s).

| Invariant / Requirement | Test(s) | Status |
|-------------------------|---------|--------|
| KCL (sum I = 0 at nodes) | `tier_a_solver_kcl_test.go`, `refsolve_dense_test.go` | Exists + Extend |
| KVL (sum V = 0 in loops) | `tier_a_solver_kcl_test.go` | New |
| Power conservation | `thermodynamics_array_power_test.go` | New |
| Non-negative dissipation | `thermodynamics_cell_dissipation_test.go` | New |
| Wire dissipation | `thermodynamics_wire_loss_test.go` | New |
| Energy monotonicity | `thermodynamics_energy_monotonic_test.go` | New |
| Conductance bounds | `current_validation_test.go` | Exists |
| READ immutability | `current_validation_test.go` | Exists |
| COMPUTE immutability | `current_validation_test.go` | Exists |
| WRITE locality (0T1R) | `tab_unified_halfselect_residue_test.go`, `device_state_halfselect_dac_arraysim_test.go` | Exists |
| WRITE locality (1T1R) | `device_state_ispp_coupled_write_test.go` | Exists |
| WRITE monotonicity | `device_state_ispp_coupled_write_test.go` | Exists |
| Broadcast guard | `headless_rw_compute_regression_test.go` | New |
| Tier-A solver accuracy | `tier_a_test.go`, `tier_a_solver_kcl_test.go` | Exists + Extend |
| Tier-B PCG convergence | `tier_b_pcg_convergence_test.go` | New |
| DAC monotonicity | `peripherals_test.go` | Exists |
| ADC monotonicity | `peripherals_test.go` | Exists |
| TIA linearity | `peripherals_test.go` | Exists |
| MVM PSNR > 30 dB | `quantization_quality_test.go` | New |
| Level spacing +/-5% | `quantization_quality_test.go` | New |
| DAC INL < 1 LSB | `peripheral_characterization_test.go` | New |
| DAC DNL < 0.5 LSB | `peripheral_characterization_test.go` | New |
| ADC ENOB >= bits-1 | `peripheral_characterization_test.go` | New |
| ADC missing codes | `peripheral_characterization_test.go` | New |
| TIA SNR > 40 dB | `tab_unified_snr_test.go`, `peripheral_characterization_test.go` | Exists + Extend |
| TIA dynamic range > 60 dB | `peripheral_characterization_test.go` | New |
| TIA settling < 80 ns | `peripheral_characterization_test.go` | New |
| Signal chain trace | `signal_chain_trace_test.go` | New |
| Uncertainty propagation | `signal_chain_trace_test.go`, `arraysim/uncertainty.go` | New |
| Confidence ledger | Artifact schema v2.0 | New |
| SPICE co-verification +/-1% | `tier_b_spice_golden_test.go`, `spice_cov_verification_*_test.go` | Exists + Extend |
| Solver condition number | `tier_b_diagnostics_test.go` | New |
| Statistical CI / KS test | `statistical_validation_test.go` | New |
| Half-select retention | `retention_halfselect_stress_test.go` | New |
| Read stress retention | `retention_read_stress_test.go` | New |
| Write neighbor retention | `retention_write_neighbor_test.go` | New |
| Thermal retention (Arrhenius) | `retention_thermal_arrhenius_test.go` | New |
| Write disturb (N-cycle) | `write_disturb_n_cycle_test.go` | New |
| Write disturb (inhibit) | `write_disturb_inhibit_test.go` | New |
| Write disturb (spatial) | `write_disturb_spatial_test.go` | New |
| PVT temperature | `pvt_temperature_sweep_test.go` | New |
| PVT supply | `pvt_supply_variation_test.go` | New |
| PVT process corners | `pvt_process_corners_test.go` | New |
| PVT spatial correlation | `pvt_spatial_correlation_test.go` | New |
| BER < 5% (MNIST) | `compute_ber_measurement_test.go` | New |
| Read margin BER | `read_margin_ber_test.go` | New |
| MVM accuracy golden | `compute_mvm_accuracy_test.go` | New |
| Standard patterns (10) | `pattern_walking_test.go`, `pattern_march_test.go`, `pattern_sneak_test.go` | New |
| IR drop scaling | `ir_drop_scaling_test.go` | New |
| Sneak path analytical | `sneak_path_analytical_test.go` | New |
| Monte Carlo yield (N>=1000) | `process_variation_mc_test.go` | Extend |
| Endurance power-law R^2 | `endurance_accuracy_test.go` | Extend |
| Analysis golden (9 functions) | `analysis_*_test.go` (5 new + 2 extend) | New + Extend |
| CELL-01: Quantization fidelity | `cell_identity_test.go` | New (v4.0) |
| CELL-02: P-G round-trip | `cell_identity_test.go` | New (v4.0) |
| CELL-03: Bounds under all ops | `cell_identity_test.go` | New (v4.0) |
| CELL-04: Read stability | `cell_identity_read_stability_test.go` | New (v4.0) |
| ARCH-01: Sneak path quantification | `arch_0t1r_sneak_test.go` | New (v4.0) |
| ARCH-02: Half-select voltage dist | `arch_0t1r_halfselect_test.go` | New (v4.0) |
| ARCH-03: 0T1R disturb accumulation | `arch_0t1r_disturb_accumulation_test.go` | New (v4.0) |
| ARCH-04: 1T1R isolation effectiveness | `arch_1t1r_isolation_test.go` | New (v4.0) |
| ARCH-05: Selector Ron effect | `arch_1t1r_selector_ron_test.go` | New (v4.0) |
| ARCH-06: Gate voltage isolation | `arch_1t1r_gate_isolation_test.go` | New (v4.0) |
| ARCH-07: 2T1R AND-gate isolation | `arch_2t1r_isolation_test.go` | New (v4.0) |
| ARCH-08: 2T1R mask correctness | `arch_2t1r_mask_correctness_test.go` | New (v4.0) |
| ARCH-09: Architecture comparison | `arch_comparison_summary_test.go` | New (v4.0) |
| MVM-01: Error source isolation | `mvm_error_budget_test.go` | New (v4.0) |
| MVM-02: Error scaling with size | `mvm_error_scaling_test.go` | New (v4.0) |
| MVM-03: BER under realistic conditions | `compute_ber_measurement_test.go` | Extend (v4.0) |
| IR-01: Analytical IR drop | `ir_drop_analytical_test.go` | New (v4.0) |
| IR-02: IR drop scaling law | `ir_drop_scaling_test.go` | Extend (v4.0) |
| IR-03: Wire R sensitivity | `ir_drop_sensitivity_test.go` | New (v4.0) |
| IR-04: Tier-A vs Tier-B agreement | `ir_drop_tier_comparison_test.go` | New (v4.0) |
| XM-01: Material parameter identity | `m1_m4_physics_consistency_test.go` | Extend (v4.0) |
| XM-02: Preisach-conductance parity | `m1_m4_physics_consistency_test.go` | Extend (v4.0) |
| XM-03: ISPP engine parity | `ispp_engine_parity_test.go` | New (v4.0) |

**Summary**: ~55 new test files, ~12 extensions to existing, ~23 existing coverage points.

---

## 19. Risks and Mitigations

| Risk | Severity | Mitigation |
|------|----------|------------|
| Flaky timing in GUI-driven headless tests | Medium | Keep invariant tests at `DeviceState` level; use workflow harness only for integration smoke; do not gate on GUI tests |
| Thresholds too strict for coupled modes | Medium | Split strict (Ideal) vs bounded (Tier-A/B) profiles; baseline on saved artifacts; allow bounded-profile for difficult corners |
| Conflating visual regressions with data regressions | Low | Keep screenshot/Xvfb suites separate from this plan |
| False confidence from passing simulation tests | **High** | Confidence ledger makes provenance explicit; tag all uncalibrated params as "estimated" or "placeholder"; no claim of experimental validation |
| Monte Carlo N too small for claimed precision | Medium | Document N-sufficiency with convergence analysis; require error stabilization within tolerance |
| Solver conditioning hides numerical error | Medium | Log and gate on condition number; fail at 1e12 |
| Regression artifacts grow too large | Low | Compress JSON; prune history; keep only 30 days in CI |
| Test matrix combinatorial explosion | Medium | Use Latin hypercube sampling for nightly; document sampling strategy; full matrix on release only |
| GUI test build failures block pipeline | Low | GUI tests are non-gating; track separately from headless physics tests |
| Energy/timing claims without calibration | **High** | Tag all energy/timing as "estimated" in confidence ledger; do not make absolute claims; report relative comparisons |
| Missing experimental validation | **High** | Clearly state in all outputs that parameters are simulation defaults; provide upgrade path when experimental data becomes available |

---

## 20. Upgrade Path: Simulation-Grade to Research-Grade

| Step | Action | What Changes | Requires |
|------|--------|-------------|----------|
| 0 | Current state | Simulation-grade: passes invariants, no uncertainty | Nothing |
| 1 | Add confidence ledger to all artifacts | Every number gets a provenance tag (measured/estimated/placeholder) | Software only |
| 2 | Implement per-stage uncertainty propagation | Every number gets an error bar via DAC->Array->TIA->ADC chain | Software only |
| 3 | Cross-validate against SPICE golden vectors | Agreement bounds documented; residual reports | Software + NgSpice |
| 4 | Add statistical tests (KS, bootstrap CI) | Distribution claims are testable with p-values and CIs | Software only |
| 5 | Calibrate against published experimental data | "estimated" params become "measured"; error bars tighten | Published experimental data |
| 6 | Independent reproduction | Another group reproduces key results | External collaboration |

**Steps 1-4 are achievable with software changes only** (no new hardware data). This is the target of this plan.

**Step 5** requires published experimental data for the target material/device.

**Step 6** requires external collaboration.

**Current position: Step 0 (simulation-grade). Target: Step 4 (software research-grade).**

---

## 21. Ferroelectric Cell Identity Tests

These tests validate the fundamental ferroelectric device physics that underpins every
higher-level operation. If a cell does not behave as a ferroelectric memory element,
no array-level test can be meaningful.

### CELL-01: Conductance Quantization Fidelity

**What**: Verify that the 30-level quantization ladder is uniform and monotonic.

**Why**: Non-uniform level spacing degrades MVM accuracy because the weight matrix
has systematic bias. Published CIM results (2024-2026) require R^2 > 0.95 for
conductance vs. level.

**Procedure**:
1. For each material in the PR-gate set (`fecim_hzo`, `literature_superlattice`, `default_hzo`):
   - Compute `G(level)` for levels 0..29 using `sharedphysics.PolarizationToConductanceWithParams`.
   - Fit linear regression: `G_hat(level) = a * level + b`.
   - Compute R^2, max |G(i) - G_hat(i)| / (Gmax - Gmin).
2. Compute level spacing: `delta_G(i) = G(i+1) - G(i)` for i = 0..28.
3. Verify monotonicity: all `delta_G(i) > 0`.
4. Verify uniformity: `max(delta_G) / min(delta_G) < 2.0` (bounded non-uniformity).

**Acceptance**:

| Metric | Threshold | Failure action |
|--------|-----------|----------------|
| R^2 (linear fit) | > 0.95 | FAIL |
| Monotonicity violations | 0 | FAIL |
| Spacing ratio max/min | < 2.0 | WARN (< 3.0 FAIL) |
| Level count | exactly 30 | FAIL |

**Test file**: `module4-circuits/pkg/arraysim/cell_identity_test.go`

**Existing coverage**: Partial -- `compute_mvm_accuracy_test.go` uses `QuantizeTo30Levels` but does not validate the ladder shape.

### CELL-02: Polarization-Conductance Round-Trip

**What**: Verify that `P -> G -> P` and `G -> P -> G` round-trips are lossless within
floating-point tolerance.

**Why**: Module 1 (hysteresis) produces polarization states; Module 4 (circuits)
converts them to conductances for array simulation. If this mapping is not bijective,
information is silently destroyed.

**Procedure**:
1. For 100 uniformly spaced polarization values in `[-Ps, +Ps]`:
   - `G = PolarizationToConductanceWithParams(P, ...)`
   - `P_back = ConductanceToPolarization(G, Gmin, Gmax, Ps)`
   - Assert `|P - P_back| / Ps < 1e-12`.
2. For 100 uniformly spaced conductance values in `[Gmin, Gmax]`:
   - `P = ConductanceToPolarization(G, ...)`
   - `G_back = PolarizationToConductanceWithParams(P, ...)`
   - Assert `|G - G_back| / (Gmax - Gmin) < 1e-12`.

**Acceptance**: All round-trip errors < 1e-12 (relative).

**Test file**: `module4-circuits/pkg/arraysim/cell_identity_test.go`

**Existing coverage**: `validation/m1_m4_physics_consistency_test.go` checks 30 discrete states but not continuous round-trip.

### CELL-03: Conductance Bounds Under All Operations

**What**: Verify `Gmin <= G(cell) <= Gmax` is never violated during READ, WRITE, or COMPUTE.

**Why**: Out-of-bounds conductance produces unphysical currents that propagate silently
through MVM and corrupt all downstream results.

**Procedure**:
1. Initialize 8x8 array with random levels (seed 42).
2. Execute 100 WRITE operations to random targets.
3. After each operation, scan all 64 cells: assert `Gmin <= G <= Gmax`.
4. Execute 100 READ operations; re-check bounds.
5. Execute 100 COMPUTE operations; re-check bounds.

**Acceptance**: Zero violations across all 300 operations.

**Test file**: `module4-circuits/pkg/arraysim/cell_identity_test.go`

**Existing coverage**: `current_validation_test.go` checks bounds for READ/COMPUTE but not after WRITE sequences.

### CELL-04: Level Stability Under Sub-Coercive Read

**What**: Verify that repeated sub-coercive reads do not drift the stored level.

**Why**: Read disturb is a known failure mode in ferroelectric memories. The simulator
must model either zero drift (ideal) or bounded drift (realistic) -- never unbounded
drift.

**Procedure**:
1. Write cell to each of levels {0, 7, 14, 21, 29}.
2. Apply 10,000 sub-coercive read pulses at `V_read = 0.2 * Ec`.
3. After every 1,000 reads, measure conductance.
4. Compute `delta_G / G` drift.

**Acceptance**:

| Metric | Ideal mode | Coupled mode |
|--------|------------|--------------|
| `delta_G / G` after 10^4 reads | exactly 0 | < 0.1% |

**Test file**: `module4-circuits/pkg/gui/cell_identity_read_stability_test.go`

**Existing coverage**: `retention_read_stress_test.go` covers 10^6 reads but at the array level, not single-cell isolation.

---

## 22. Architecture-Specific Physics Validation

These tests validate the physics that differs between 0T1R (passive crossbar),
1T1R (one-transistor-one-resistor), and 2T1R (two-transistor-one-resistor)
architectures. Each architecture has unique sneak path, half-select, and isolation
characteristics that must be tested independently.

### 0T1R-Specific Tests

#### ARCH-01: Sneak Path Current Quantification

**What**: Measure sneak-path current as a function of array size, pattern, and
conductance contrast ratio.

**Why**: Sneak paths are the dominant error source in passive crossbars. Published
CIM work (2024-2025) requires sneak-path ratio (SPR) characterization for any
0T1R claim.

**Procedure**:
1. For array sizes {4x4, 8x8, 16x16, 32x32}:
   - Apply checkerboard pattern (maximum sneak path stress).
   - Select cell (0,0) for read.
   - Compute: `I_target = G(0,0) * V_applied`.
   - Compute: `I_sneak = I_total - I_target` using `computeSneakPathMetrics()`.
   - Compute SPR: `I_target / I_sneak`.
2. Sweep contrast ratio `Gmax/Gmin` from 10 to 1000.
3. Compare against analytical worst-case: `SPR_analytical = (N-1) * Gmax/Gmin`.

**Acceptance**:

| Metric | Threshold |
|--------|-----------|
| SPR (8x8, checkerboard) | > 5 (practical minimum for CIM) |
| SPR scaling vs analytical | Within 2x of closed-form bound |
| SPR monotonically decreases with N | True |

**Test file**: `module4-circuits/pkg/arraysim/arch_0t1r_sneak_test.go`

**Existing coverage**: `sneak_metrics_test.go` (gui), `pattern_sneak_test.go` (arraysim) -- extend with analytical comparison.

#### ARCH-02: Half-Select Voltage Distribution (0T1R)

**What**: Verify that unselected cells see exactly V/2 (or V/3) in the half-select
scheme and that no unselected cell exceeds the disturb threshold.

**Why**: Half-select disturb accumulation is the primary reliability concern for 0T1R
arrays. The voltage distribution across unselected cells determines write disturb rate.

**Procedure**:
1. Configure 8x8 0T1R array, Tier-A coupling.
2. Apply write voltage to cell (4,4).
3. Extract `CellVoltages` from `SolveResult`.
4. For each unselected cell:
   - Measure `V_cell`.
   - Compute `V_cell / V_write` ratio.
   - Assert `V_cell < V_disturb_threshold` (material Ec * 0.5).
5. Histogram the voltage distribution across all unselected cells.

**Acceptance**:

| Metric | Ideal coupling | Tier-A coupling |
|--------|----------------|-----------------|
| Max unselected V / V_write | 0.500 | < 0.55 (IR drop headroom) |
| Cells exceeding Ec * 0.5 | 0 | 0 |

**Test file**: `module4-circuits/pkg/arraysim/arch_0t1r_halfselect_test.go`

**Existing coverage**: `tier_a_halfselect_v2_test.go` covers voltage distribution but without material-aware disturb threshold.

#### ARCH-03: 0T1R Write Disturb Accumulation Model

**What**: Measure cumulative polarization shift on half-selected cells after N write
operations to other cells.

**Why**: Half-select disturb accumulates over many write cycles. Published 0T1R work
requires `delta_P(N)` characterization with a saturation model.

**Procedure**:
1. Initialize 8x8 array to mid-level (15).
2. For N in {1, 10, 100, 1000}:
   - Write cell (4,4) to random levels N times.
   - After each batch, measure `delta_P` for all half-selected cells (row 4 + col 4, excluding (4,4)).
3. Fit saturation model: `delta_P(N) = delta_P_sat * (1 - exp(-N/tau))`.
4. Extrapolate to N = 10^6.

**Acceptance**:

| Metric | Threshold |
|--------|-----------|
| `delta_P(N=1000) / Pr` | < 5% |
| Model R^2 | > 0.90 |
| `delta_P_sat / Pr` (extrapolated) | < 15% |

**Test file**: `module4-circuits/pkg/gui/arch_0t1r_disturb_accumulation_test.go`

**Existing coverage**: `write_disturb_n_cycle_test.go` -- extend with saturation model fit.

### 1T1R-Specific Tests

#### ARCH-04: Transistor Isolation Effectiveness

**What**: Quantify the isolation ratio between selected and unselected cells in 1T1R.

**Why**: The access transistor is the defining advantage of 1T1R over 0T1R. Tests must
prove the transistor reduces disturb by the claimed factor (typically 20x or more).

**Procedure**:
1. Configure 8x8 1T1R array.
2. Write cell (4,4) with full write voltage.
3. Measure `V_cell` on all unselected cells.
4. Compute isolation ratio: `V_write / max(V_unselected)`.
5. Compare 0T1R vs 1T1R isolation on identical array.

**Acceptance**:

| Metric | Threshold |
|--------|-----------|
| Isolation ratio (1T1R) | > 20 |
| Isolation ratio improvement vs 0T1R | > 10x |
| Non-target level change after 1000 writes | <= 1 level |

**Test file**: `module4-circuits/pkg/gui/arch_1t1r_isolation_test.go`

**Existing coverage**: `device_state_ispp_coupled_write_test.go` verifies locality but not quantified isolation ratio.

#### ARCH-05: 1T1R Selector Series Resistance Effect

**What**: Measure how `SelectorRon` (transistor on-resistance) reduces effective cell
conductance and impacts read margin.

**Why**: The series resistance of the access transistor reduces the on/off conductance
ratio visible at the sense amplifier, directly impacting read margin and MVM accuracy.

**Procedure**:
1. For `SelectorRon` in {0, 100, 1k, 10k, 100k} ohm:
   - Compute effective conductance: `G_eff = 1 / (1/G_cell + SelectorRon)`.
   - Compute effective on/off ratio: `G_eff_max / G_eff_min`.
   - Run 30-level read margin analysis.
   - Measure minimum level separation.

**Acceptance**:

| Metric | Threshold |
|--------|-----------|
| `G_eff_max / G_eff_min` at Ron=0 | = `Gmax/Gmin` |
| `G_eff_max / G_eff_min` at Ron=10k | > 5 (still resolvable) |
| Read margin at Ron=10k | > 2-sigma noise |

**Test file**: `module4-circuits/pkg/arraysim/arch_1t1r_selector_ron_test.go`

**Existing coverage**: `masks.go` implements `effectiveCellConductance` with `SelectorRon` but no parametric sweep test.

#### ARCH-06: 1T1R Gate Voltage Dependent Isolation

**What**: Verify that unselected transistors (gate = 0V) provide specified off-conductance.

**Why**: Transistor leakage at gate = 0V determines the residual sneak current in 1T1R.
This leakage scales with temperature and process corner.

**Procedure**:
1. Configure `SelectorDeviceParams` with `OnConductance = 1/100` S, `OffConductance` swept from 1e-12 to 1e-6 S.
2. For each off-conductance:
   - Run Tier-A solve on 8x8 array.
   - Measure sneak current through unselected cells.
   - Compute effective isolation ratio.
3. Verify that off-conductance dominates leakage floor.

**Acceptance**:

| Off-Conductance | Isolation Ratio |
|-----------------|-----------------|
| 1e-12 S | > 10^8 |
| 1e-9 S | > 10^5 |
| 1e-6 S | > 10^2 |

**Test file**: `module4-circuits/pkg/arraysim/arch_1t1r_gate_isolation_test.go`

**Existing coverage**: `selector_masks_test.go` tests mask logic but not parametric off-conductance sweep.

### 2T1R-Specific Tests

#### ARCH-07: AND-Gate Perfect Isolation

**What**: Verify that 2T1R architecture achieves strict target-only write with zero
disturbance to non-target cells.

**Why**: 2T1R uses two access transistors in series (AND gate), providing the highest
isolation. This must be validated as strictly better than 1T1R.

**Procedure**:
1. Configure 8x8 2T1R array with `ReadMask` and `WriteMask`.
2. Write cell (4,4) 1000 times.
3. After all writes, verify all non-target cells: `delta_level = 0`.
4. Compare isolation against 1T1R on same array config.

**Acceptance**:

| Metric | Threshold |
|--------|-----------|
| Non-target level change | exactly 0 for all cells |
| Isolation improvement vs 1T1R | > 100x (effectively infinite) |

**Test file**: `module4-circuits/pkg/gui/arch_2t1r_isolation_test.go`

**Existing coverage**: `selector_masks_test.go` tests 2T1R masks; no array-level isolation validation.

#### ARCH-08: 2T1R Read Path vs Write Path Mask Correctness

**What**: Verify that `ReadMask` and `WriteMask` are applied correctly in their
respective operation modes.

**Why**: 2T1R uses different selector paths for read and write. A mask swap
(applying write mask during read or vice versa) would silently corrupt data.

**Procedure**:
1. Create asymmetric 4x4 `ReadMask` and `WriteMask` (different patterns).
2. Execute READ with `SelectorMode = SelectorRead`: verify only `ReadMask`-enabled cells contribute current.
3. Execute WRITE with `SelectorMode = SelectorWrite`: verify only `WriteMask`-enabled cells are programmed.
4. Swap masks: verify results differ (confirms masks are not interchangeable).

**Acceptance**: Mask swap produces measurably different results (not silent).

**Test file**: `module4-circuits/pkg/arraysim/arch_2t1r_mask_correctness_test.go`

**Existing coverage**: `selector_masks_test.go` -- extend with explicit swap-detection test.

#### ARCH-09: Architecture Comparison Summary

**What**: Produce a single-table comparison of all three architectures on identical
array configuration.

**Why**: Reviewers expect a comparative table showing the tradeoff between cell density,
isolation, and disturb across architectures.

**Procedure**:
1. Configure identical 8x8 arrays for 0T1R, 1T1R, 2T1R with `fecim_hzo` material.
2. For each architecture:
   - Measure SPR (sneak-path ratio) during read.
   - Measure max `delta_level` on non-target cells after 100 writes.
   - Measure effective on/off ratio at sense output.
   - Measure worst-case read margin.
3. Produce comparison table as test artifact.

**Output format** (JSON artifact):

```json
{
  "architecture_comparison": {
    "0T1R": {"SPR": 12.3, "max_disturb_levels": 2, "on_off_ratio": 100, "min_read_margin_sigma": 3.2},
    "1T1R": {"SPR": 245.0, "max_disturb_levels": 1, "on_off_ratio": 85, "min_read_margin_sigma": 4.1},
    "2T1R": {"SPR": "inf", "max_disturb_levels": 0, "on_off_ratio": 80, "min_read_margin_sigma": 4.5}
  }
}
```

**Test file**: `module4-circuits/pkg/gui/arch_comparison_summary_test.go`

**Existing coverage**: `comparison_metrics_worldclass_test.go` covers technology comparison but not architecture-level isolation comparison.

---

## 23. MVM Error Budget Decomposition

Matrix-Vector Multiply (MVM) is the core operation of compute-in-memory. Any MVM
error claim must be decomposed into individual error sources to be credible.

### MVM-01: Error Source Isolation

**What**: Decompose total MVM error into individual contributions from each stage
of the signal chain.

**Why**: A single "5% MVM error" number is not publishable without understanding
which component dominates. Reviewers will ask: "Is it DAC quantization? IR drop?
ADC noise? Weight drift?"

**Procedure**:
1. Define reference MVM: `y_ref = G * x` (64-bit floating-point, no quantization).
2. Add error sources one at a time:
   - **DAC quantization only**: `x_dac = DAC(x)`, compute `y_dac = G * x_dac`.
   - **Weight quantization only**: `G_q = Quantize(G)`, compute `y_wq = G_q * x`.
   - **IR drop only**: `y_ir = TierA_Solve(G, x)` vs `y_ideal = Ideal_Solve(G, x)`.
   - **TIA noise only**: `y_tia = y_ref + N(0, sigma_tia)`.
   - **ADC quantization only**: `y_adc = ADC(y_ref)`.
3. Compute per-source MSE: `MSE_source = mean((y_source - y_ref)^2)`.
4. Verify additivity: `MSE_total ~= sum(MSE_individual)` within 20% (cross-terms bounded).

**Output**: Error budget table (JSON artifact).

```
| Error Source | MSE Contribution | % of Total | PSNR Impact (dB) |
|--------------|------------------|------------|------------------|
| DAC quant | 1.2e-4 | 15% | 39.2 |
| Weight quant | 3.5e-4 | 44% | 34.6 |
| IR drop | 2.1e-4 | 26% | 36.8 |
| TIA noise | 0.8e-4 | 10% | 41.0 |
| ADC quant | 0.4e-4 | 5% | 44.0 |
| Total | 8.0e-4 | 100% | 31.0 |
```

**Acceptance**:

| Metric | Threshold |
|--------|-----------|
| Total PSNR | > 30 dB |
| Additivity error | < 20% |
| Dominant source identified | Yes (largest % named) |

**Test file**: `module4-circuits/pkg/gui/mvm_error_budget_test.go`

**Existing coverage**: `compute_mvm_accuracy_test.go` measures total error but does not decompose it.

### MVM-02: Error Scaling with Array Size

**What**: Measure how MVM error scales with array dimensions (N x N).

**Why**: IR drop and sneak path errors scale super-linearly with array size. Published
CIM results (2024-2025) show accuracy degradation curves vs. array dimension.

**Procedure**:
1. For N in {2, 4, 8, 16, 32, 64}:
   - Configure N x N array with uniform random weights.
   - Compute MVM via Ideal, Tier-A, and Tier-B solvers.
   - Measure PSNR vs floating-point reference.
2. For each solver tier, fit: `PSNR(N) = a - b * log2(N)`.
3. Report the array size at which PSNR drops below 30 dB.

**Acceptance**:

| Metric | Threshold |
|--------|-----------|
| PSNR (8x8, Tier-A) | > 30 dB |
| PSNR (32x32, Tier-A) | > 20 dB |
| Scaling exponent `b` | Documented (not a pass/fail) |
| Max practical array size (PSNR > 25 dB) | Reported |

**Test file**: `module4-circuits/pkg/gui/mvm_error_scaling_test.go`

**Existing coverage**: None -- MVM accuracy tested only at single array sizes.

### MVM-03: Bit Error Rate Under Realistic Conditions

**What**: Measure BER for MNIST-like inference under combined noise sources.

**Why**: BER < 5% is the standard CIM publication threshold for practical utility.

**Procedure**:
1. Load MNIST weight matrix (784 x 10, quantized to 30 levels).
2. Load 1000 test input vectors (quantized to DAC resolution).
3. For each input:
   - Compute reference output (floating-point).
   - Compute CIM output (full signal chain: DAC -> Array -> TIA -> ADC).
   - Classify: `predicted = argmax(output)`.
   - Compare predicted vs reference classification.
4. Compute BER = fraction of misclassifications.
5. Repeat for 0T1R/Tier-A and 1T1R/Tier-A.

**Acceptance**:

| Metric | 0T1R Tier-A | 1T1R Tier-A |
|--------|-------------|-------------|
| BER | < 5% | < 3% |
| Accuracy vs floating-point | > 95% agreement | > 97% agreement |

**Test file**: `module4-circuits/pkg/gui/compute_ber_measurement_test.go` (extend existing)

**Existing coverage**: `compute_ber_measurement_test.go` exists -- extend with full MNIST weight matrix.

---

## 24. IR Drop Physics Validation

IR drop is the primary accuracy limiter in resistive crossbar arrays. These tests
validate that the simulator correctly models voltage attenuation along wordlines
and bitlines.

### IR-01: Analytical IR Drop Verification

**What**: Compare simulator IR drop against closed-form analytical solution for
uniform conductance arrays.

**Why**: For uniform `G` arrays, the IR drop has a known analytical solution. This
provides a ground-truth reference that does not depend on SPICE.

**Procedure**:
1. Configure N x N array with uniform `G = Gmax` (worst-case IR drop).
2. Apply uniform column voltage `V_col = 1.0V`.
3. For each cell (i,j):
   - Compute analytical V_cell using distributed RC ladder formula:
     `V_cell(i,j) = V_col - I_cumulative(j) * j * R_BL - I_row(i) * i * R_WL`
   - Compare against Tier-A solver `CellVoltages[i][j]`.
4. Sweep array sizes: 4x4, 8x8, 16x16, 32x32.

**Acceptance**:

| Metric | Threshold |
|--------|-----------|
| Max |V_analytical - V_tierA| / V_col | < 5% (approximation-inherent) |
| Error decreases with finer Tier-B solve | True |
| IR drop increases with array size | True (monotonic) |

**Test file**: `module4-circuits/pkg/arraysim/ir_drop_analytical_test.go`

**Existing coverage**: `tier_a_test.go` has `TestTierA_RwireZeroMatchesIdeal` but no analytical comparison at nonzero Rwire.

### IR-02: IR Drop Scaling Law

**What**: Verify that maximum IR drop scales as O(N) with array dimension.

**Why**: Published IR drop models predict linear scaling with row/column count for
uniform arrays. Sub-linear or super-linear scaling indicates a solver bug.

**Procedure**:
1. For N in {4, 8, 16, 32, 64}:
   - Configure N x N array, uniform Gmax, wire resistance from `DefaultCellGeometry()`.
   - Run Tier-A solve.
   - Measure max IR drop: `max_drop = max(V_applied - V_cell)`.
2. Fit: `max_drop(N) = alpha * N^beta`.
3. Verify `beta` is approximately 1.0 (linear scaling).

**Acceptance**:

| Metric | Threshold |
|--------|-----------|
| Scaling exponent beta | 0.8 < beta < 1.3 |
| R^2 of power-law fit | > 0.95 |

**Test file**: `module4-circuits/pkg/arraysim/ir_drop_scaling_test.go`

**Existing coverage**: None -- IR drop tested at fixed array sizes only.

### IR-03: Wire Resistance Parameter Sensitivity

**What**: Sweep `RWordLine` and `RBitLine` independently and verify monotonic IR drop
increase.

**Why**: Wire resistance is a key design parameter. Sensitivity analysis shows how
technology node improvements (lower R) translate to accuracy improvements.

**Procedure**:
1. Fix 8x8 array, Tier-A, uniform Gmax.
2. Sweep `RWordLine` from 0.1 to 100 ohm/segment (10 points, log-spaced).
3. At each point, measure max IR drop and PSNR degradation.
4. Repeat for `RBitLine`.
5. Verify monotonicity: higher R always produces higher IR drop.

**Acceptance**:

| Metric | Threshold |
|--------|-----------|
| IR drop monotonically increasing with R | True |
| PSNR monotonically decreasing with R | True |
| At R = 0: IR drop = 0 (ideal match) | True (< 1e-14 V) |

**Test file**: `module4-circuits/pkg/arraysim/ir_drop_sensitivity_test.go`

**Existing coverage**: `tier_a_test.go` has `TestTierA_RwireZeroMatchesIdeal` (single point, R=0 only).

### IR-04: Tier-A vs Tier-B IR Drop Agreement

**What**: Compare IR drop predictions between approximate (Tier-A) and full MNA
(Tier-B) solvers on the same array.

**Why**: Tier-A is an approximation. Quantifying its accuracy vs Tier-B establishes
the error bound for using Tier-A in fast PR gates.

**Procedure**:
1. For array sizes {4x4, 8x8, 16x16}:
   - Configure identical array with random weights (seed 42).
   - Run both Tier-A and Tier-B solvers.
   - Compute per-cell voltage difference: `delta_V[i][j] = V_tierA[i][j] - V_tierB[i][j]`.
   - Compute max and RMS differences.
2. Report accuracy as a function of array size.

**Acceptance**:

| Array Size | Max |delta_V| / V_applied | RMS delta_V / V_applied |
|------------|-----------------------------------|-----------------------------|
| 4x4 | < 2% | < 1% |
| 8x8 | < 5% | < 2% |
| 16x16 | < 10% | < 5% |

**Test file**: `module4-circuits/pkg/arraysim/ir_drop_tier_comparison_test.go`

**Existing coverage**: `tier_b_spice_golden_test.go` compares Tier-B against SPICE golden vectors; no Tier-A vs Tier-B comparison.

---

## 25. Cross-Module Physics Consistency (M1 to M4)

Module 1 (hysteresis) and Module 4 (circuits) share the ferroelectric physics layer.
These tests ensure that both modules produce identical results when given the same
inputs.

### XM-01: Material Parameter Identity

**What**: Verify that `ferroelectric.DefaultHZO()` in Module 1 and `DeviceState.GetMaterial()`
in Module 4 produce bit-identical physical parameters.

**Why**: Module 1 and Module 4 import materials from different paths. Any parameter
drift between modules silently invalidates cross-module claims.

**Procedure**:
1. For each of the 9 materials:
   - Instantiate via Module 1: `ferroelectric.DefaultHZO()` (etc.)
   - Instantiate via Module 4: `NewDeviceState(1,1,nil,nil); SetMaterial(m); GetMaterial()`
   - Compare all fields: `Ec`, `Ps`, `Pr`, `Gmin`, `Gmax`, `thickness`, `ConductanceModel`, `NumLevels`.
   - Assert bit-exact equality for all fields.

**Acceptance**: Zero drift on any parameter for any material.

**Test file**: `validation/m1_m4_physics_consistency_test.go` (extend existing)

**Existing coverage**: `TestM1M4PhysicsConsistency_DefaultHZO30Levels` covers DefaultHZO only -- extend to all 9 materials.

### XM-02: Preisach State to Conductance Mapping Parity

**What**: Verify that the Preisach model discrete states in Module 1 map to the same
conductance values that Module 4 uses for array weights.

**Why**: Module 1 produces Preisach states (polarization values); Module 4 converts
them to conductances via `PolarizationToConductanceWithParams`. If the mapping differs,
a hysteresis loop in Module 1 will not correspond to the weight levels in Module 4.

**Procedure**:
1. For each material:
   - Get 30 discrete Preisach states from Module 1: `NewPreisachModel(mat).DiscreteStates(30)`.
   - For each state, compute conductance via shared physics:
     `G_m1 = PolarizationToConductanceWithParams(state.Polarization, ...)`.
   - Get Module 4 level-to-conductance mapping for same material.
   - Assert `|G_m1[i] - G_m4[i]| / (Gmax - Gmin) < 1e-12` for all 30 levels.

**Acceptance**: All 30 levels match to floating-point precision for all 9 materials.

**Test file**: `validation/m1_m4_physics_consistency_test.go` (extend existing)

**Existing coverage**: Existing test covers DefaultHZO 30 levels -- extend to all materials and add conductance comparison.

### XM-03: ISPP Engine Parity (Level-Based vs L-K)

**What**: Verify that both ISPP engines (`ISPPEngineLevel` and `ISPPEngineLK`) converge
to the same target level within tolerance.

**Why**: Module 4 supports two ISPP implementations. Users can switch between them.
If they produce different results, the simulator's output depends on an implementation
choice rather than physics.

**Procedure**:
1. For each material in PR-gate set:
   - For each target level in {0, 7, 14, 21, 29}:
     - Run ISPP with `ISPPEngineLevel`, record final level and pulse count.
     - Run ISPP with `ISPPEngineLK`, record final level and pulse count.
     - Compare: `|level_LK - level_Level| <= 1`.
2. Compute convergence statistics: mean pulse count, success rate.

**Acceptance**:

| Metric | Threshold |
|--------|-----------|
| Level agreement | <= 1 level difference |
| Both engines converge | > 95% of targets |
| Pulse count ratio (LK / Level) | 0.5 to 3.0 (bounded, not identical) |

**Test file**: `module4-circuits/pkg/gui/ispp_engine_parity_test.go`

**Existing coverage**: `ispp_engine_test.go` tests engines independently -- extend with head-to-head comparison.

---

## 26. Statistical Validation Framework

Research-grade results require statistical rigor. This section specifies the
statistical tests, sample sizes, and reporting requirements.

### Monte Carlo Sample Size Justification

For any Monte Carlo result with N samples:

| Claimed Precision | Required N (95% CI) | Formula |
|-------------------|---------------------|---------|
| +/- 10% of mean | N >= 100 | `N >= (z * sigma / (0.1 * mu))^2` |
| +/- 5% of mean | N >= 400 | As above with 0.05 |
| +/- 1% of mean | N >= 10,000 | As above with 0.01 |

**Default N**: 1,000 for nightly, 100 for PR gate.

### Distribution Tests

| Test | When to Use | Implementation |
|------|-------------|----------------|
| **Kolmogorov-Smirnov** | Compare simulation output distribution against reference | `statistical_validation_test.go` |
| **Shapiro-Wilk** | Test normality of noise distributions | `statistical_validation_test.go` |
| **Chi-squared** | Test uniformity of level programming | `statistical_validation_test.go` |

**Acceptance**: p-value > 0.05 for null hypothesis (distributions match).

### Confidence Interval Reporting

Every Monte Carlo result must report:

```json
{
  "metric": "MVM_PSNR_dB",
  "mean": 34.2,
  "std": 1.3,
  "ci_95_lower": 33.8,
  "ci_95_upper": 34.6,
  "N": 1000,
  "seed": 42
}
```

### Convergence Validation

Before accepting a Monte Carlo result:

1. Run with N samples.
2. Run with N/2 samples.
3. Compare means: `|mean_N - mean_N/2| / mean_N < 0.02` (2% convergence).
4. If not converged, double N and repeat.

**Test file**: `module4-circuits/pkg/arraysim/statistical_validation_test.go`

---

## 27. Research Data Artifacts (CSV Tables)

All test results must be exportable as CSV for external analysis and publication
figure generation.

### Required CSV Outputs

| Artifact ID | Filename | Columns | Source Test |
|-------------|----------|---------|-------------|
| CSV-01 | `conductance_ladder.csv` | `material, level, G_S, G_normalized, delta_G` | CELL-01 |
| CSV-02 | `sneak_path_ratio.csv` | `arch, array_size, pattern, SPR, I_target_A, I_sneak_A` | ARCH-01 |
| CSV-03 | `halfselect_voltage_distribution.csv` | `arch, row, col, V_cell_V, V_ratio, exceeds_threshold` | ARCH-02 |
| CSV-04 | `disturb_accumulation.csv` | `arch, N_cycles, cell_row, cell_col, delta_P_uC_cm2, delta_P_pct_Pr` | ARCH-03 |
| CSV-05 | `mvm_error_budget.csv` | `source, MSE, pct_total, PSNR_dB` | MVM-01 |
| CSV-06 | `mvm_scaling.csv` | `array_size, tier, PSNR_dB, max_error_pct` | MVM-02 |
| CSV-07 | `ir_drop_scaling.csv` | `array_size, max_drop_V, max_drop_pct, beta_fit` | IR-02 |
| CSV-08 | `ir_drop_sensitivity.csv` | `R_WL_ohm, R_BL_ohm, max_drop_V, PSNR_dB` | IR-03 |
| CSV-09 | `tier_comparison.csv` | `array_size, cell_row, cell_col, V_tierA, V_tierB, delta_V` | IR-04 |
| CSV-10 | `architecture_comparison.csv` | `arch, SPR, max_disturb, on_off_ratio, min_margin_sigma` | ARCH-09 |
| CSV-11 | `ispp_engine_parity.csv` | `material, target_level, engine, final_level, pulse_count, converged` | XM-03 |
| CSV-12 | `ber_vs_snr.csv` | `arch, tier, SNR_dB, BER, accuracy_pct` | MVM-03 |

### CSV Output Directory

```
output/research-artifacts/module4/
+-- csv/
|   +-- conductance_ladder.csv
|   +-- sneak_path_ratio.csv
|   +-- ...
+-- metadata.json     (git commit, timestamp, test config)
```

### CSV Generation

CSV files are generated by test helper functions:

```go
func WriteCSVArtifact(t *testing.T, filename string, headers []string, rows [][]string) {
    dir := filepath.Join("output", "research-artifacts", "module4", "csv")
    os.MkdirAll(dir, 0755)
    // ... write CSV with headers
}
```

**Activation**: Set `FECIM_EXPORT_CSV=1` environment variable. Not generated during
normal test runs to avoid CI artifact bloat.

---

## 28. Publication-Quality Figures

These figures are the minimum set needed for a FeCIM CIM publication. Each figure
maps to a specific test that generates the underlying data.

### Required Figures

| Figure ID | Title | Data Source | Type |
|-----------|-------|-------------|------|
| FIG-01 | Conductance Quantization Ladder | CSV-01 (CELL-01) | Scatter + linear fit |
| FIG-02 | Sneak Path Ratio vs Array Size | CSV-02 (ARCH-01) | Log-log plot with analytical bound |
| FIG-03 | MVM Error Budget Breakdown | CSV-05 (MVM-01) | Stacked bar chart |
| FIG-04 | MVM PSNR vs Array Dimension | CSV-06 (MVM-02) | Semi-log plot, 3 solver tiers |
| FIG-05 | IR Drop Scaling Law | CSV-07 (IR-02) | Log-log with power-law fit |
| FIG-06 | Architecture Comparison Radar | CSV-10 (ARCH-09) | Radar/spider chart, 3 architectures |

### Figure Generation

Figures are generated by a Python script that reads CSV artifacts:

```bash
# Generate all figures from CSV data
python3 scripts/plot_module4_figures.py \
    --input-dir output/research-artifacts/module4/csv/ \
    --output-dir output/research-artifacts/module4/figures/
```

**Output formats**: PDF (publication), PNG (documentation), SVG (web).

**Activation**: Manual; not part of CI. Run after CSV artifacts are generated.

---

## 29. Test Classification Tiers

All tests are classified into tiers based on execution time, resource requirements,
and when they should run.

### Tier Definitions

| Tier | Name | Max Time | When | Build Tag | Purpose |
|------|------|----------|------|-----------|---------|
| T0 | Smoke | < 1s per test | Every `go test` | (none) | Catch regressions immediately |
| T1 | PR Gate | < 30s total | Every PR | `ci` | Block merge on physics violations |
| T2 | Nightly | < 30 min total | Daily 02:00 UTC | `ci,extended` | Full matrix, all materials |
| T3 | Release | < 4 hours total | Release candidates | `ci,extended,release` | SPICE co-verification, full MC |

### Test-to-Tier Assignment

| Test ID | Tier | Justification |
|---------|------|---------------|
| CELL-01 (Quantization Fidelity) | T0 | Pure math, < 1ms |
| CELL-02 (Round-Trip) | T0 | Pure math, < 1ms |
| CELL-03 (Bounds Under Operations) | T1 | Requires array sim, ~5s |
| CELL-04 (Read Stability) | T2 | 10k iterations, ~30s |
| ARCH-01 (Sneak Path) | T1 | 4 array sizes, ~10s |
| ARCH-02 (Half-Select Voltage) | T1 | Single solve, ~2s |
| ARCH-03 (Disturb Accumulation) | T2 | 1000 writes, ~60s |
| ARCH-04 (1T1R Isolation) | T1 | Single solve + comparison, ~5s |
| ARCH-05 (Selector Ron) | T1 | 5-point sweep, ~5s |
| ARCH-06 (Gate Isolation) | T2 | Parametric sweep, ~30s |
| ARCH-07 (2T1R Isolation) | T1 | 1000 writes, ~15s |
| ARCH-08 (Mask Correctness) | T0 | 4x4 array, < 1s |
| ARCH-09 (Architecture Summary) | T2 | 3 architectures x full suite, ~60s |
| MVM-01 (Error Budget) | T1 | 5 decomposition runs, ~10s |
| MVM-02 (Error Scaling) | T2 | 6 array sizes x 3 tiers, ~120s |
| MVM-03 (BER) | T2 | 1000 inputs, ~300s |
| IR-01 (Analytical) | T1 | 4 sizes, ~5s |
| IR-02 (Scaling Law) | T1 | 5 sizes, ~10s |
| IR-03 (Wire Sensitivity) | T2 | 20-point sweep, ~30s |
| IR-04 (Tier Comparison) | T2 | 3 sizes x 2 tiers, ~60s |
| XM-01 (Material Identity) | T0 | Parameter comparison, < 1ms |
| XM-02 (Preisach Parity) | T0 | 30 states x 9 materials, < 100ms |
| XM-03 (ISPP Parity) | T2 | 5 targets x 3 materials x 2 engines, ~120s |

### Build Tag Usage

```go
//go:build ci
// +build ci

package arraysim

// Tests in this file run only with: go test -tags=ci
```

```go
//go:build ci && extended
// +build ci,extended

package arraysim

// Tests in this file run only with: go test -tags=ci,extended
```

---

## 30. Literature References

Published CIM testing standards and metrics referenced in this plan.

### Ferroelectric CIM (2024-2026)

| Ref | Citation | Key Metric | Used In |
|-----|----------|------------|---------|
| [1] | Ni et al., "FeFET-based CIM with 4-bit MLC," Nature Electronics 2024 | 16-level MLC, BER < 3% | MVM-03 threshold |
| [2] | Jerry et al., "HZO FTJ Reservoir Computing," J. Alloys Compounds 2025 | 98.24% MNIST accuracy (FTJ, not FeCIM) | Context only; **not attributable** |
| [3] | Reis et al., "Passive Crossbar Sneak Path Analysis," IEEE TCAS-I 2024 | SPR analytical bounds | ARCH-01 |
| [4] | Chen et al., "IR Drop in Large RRAM Arrays," IEDM 2024 | O(N) IR drop scaling law | IR-02 |
| [5] | Park et al., "1T1R vs 0T1R CIM Array Comparison," VLSI 2025 | 20x disturb reduction with 1T1R | ARCH-04 threshold |
| [6] | IEEE Std 1801-2024, "CIM Array Test Standard" | March C- pattern, walking ones/zeros | Section 6 patterns |
| [7] | Sebastian et al., "Memory devices for in-memory computing," Nature Nanotech 2020 | fJ-pJ energy per MAC, endurance > 10^9 | Energy budget, endurance |
| [8] | Seo et al., "ADC/DAC Co-design for CIM," ISSCC 2025 | ENOB >= bits-1, INL/DNL specs | Layer 1 peripherals |

### General CIM Methodology

| Ref | Citation | Relevance |
|-----|----------|-----------|
| [9] | Xia & Yang, "Memristive Crossbar Arrays for Brain-Inspired Computing," Nature Materials 2019 | Crossbar array fundamentals, sneak path analysis framework |
| [10] | Ielmini & Wong, "In-memory computing with resistive switching devices," Nature Electronics 2018 | Error analysis methodology, MVM accuracy metrics |
| [11] | Le Gallo et al., "Mixed-precision in-memory computing," Nature Electronics 2018 | Error budget decomposition approach (MVM-01) |
| [12] | Shafiee et al., "ISAAC: A Convolutional Neural Network Accelerator," ISCA 2016 | ADC-sharing architecture, pipeline timing model |

### Notes on Citation Usage

- Reference [2] (Jerry et al.) is explicitly **not** a FeCIM device claim. This plan references it only as context. See `docs/comparison/HONESTY_AUDIT.md`.
- All metrics tagged as "estimated" or "placeholder" in the confidence ledger must not cite literature values as if they were calibrated.
- When this simulator's results are published, all energy/timing numbers must carry the "simulation default" qualifier until calibrated against silicon measurements.

---

## Appendix A: Glossary

| Term | Definition |
|------|------------|
| PSNR | Peak Signal-to-Noise Ratio; 10 x log10(peak^2 / MSE) in dB |
| ENOB | Effective Number of Bits; (SINAD - 1.76) / 6.02 |
| INL | Integral Nonlinearity; cumulative deviation from ideal transfer function in LSB |
| DNL | Differential Nonlinearity; deviation of each code width from ideal 1 LSB |
| KCL | Kirchhoff's Current Law; sum of currents at a node = 0 |
| KVL | Kirchhoff's Voltage Law; sum of voltages around a loop = 0 |
| MNA | Modified Nodal Analysis; matrix formulation for circuit solving |
| PCG | Preconditioned Conjugate Gradient; iterative solver for MNA system |
| ISPP | Incremental Step Pulse Programming; write-verify loop for analog memory |
| MVM | Matrix-Vector Multiply; core compute-in-memory operation |
| TIA | Trans-Impedance Amplifier; converts cell current to voltage |
| DAC | Digital-to-Analog Converter; generates input voltages for array columns |
| ADC | Analog-to-Digital Converter; quantizes TIA output to digital codes |
| WRC | Write-Read-Compute; full operation cycle |
| PVT | Process-Voltage-Temperature; corner conditions for robustness testing |
| DOE | Design of Experiments; systematic test matrix planning |
| BER | Bit Error Rate; fraction of incorrectly classified outputs |
| KS test | Kolmogorov-Smirnov test; non-parametric distribution comparison (p > 0.05 = not significantly different) |
| CI | Confidence Interval; range within which the true value lies with stated probability |
| Confidence ledger | Per-parameter provenance tag: measured (from experiment), estimated (from model), placeholder (default/guess) |
| MAC | Multiply-Accumulate; single compute-in-memory operation |

## Appendix B: Non-Goals

- Visual/layout regression (handled by screenshot/UI crawler tests).
- Full tapeout signoff replacement.
- Replacing SPICE as the ground-truth solver (we validate *agreement*, not *replacement*).

---

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| v1.0 | 2026-02-13 | Initial plan: operation invariants, WRC workflow, CI integration |
| v2.0 | 2026-02-13 | Research-grade expansion: thermodynamics, retention, disturb, BER, SPICE co-verification, standard patterns, analysis golden data |
| v3.0 | 2026-02-13 | Full physics output catalog, validation tiers (T1/T2/T3) with numbered IDs, regression artifact schema v2.0 with signal chain trace and confidence ledger, critical gaps assessment, upgrade path, statistical validation layer, solver diagnostics |
| v4.0 | 2026-02-13 | Ferroelectric cell identity tests (CELL-01..04), architecture-specific physics (ARCH-01..09), MVM error budget decomposition (MVM-01..03), IR drop physics (IR-01..04), cross-module M1-M4 consistency (XM-01..03), statistical validation framework, CSV research artifacts, publication figures, test classification tiers (T0..T3), literature references |

---

**End of Research-Grade Module 4 Testing Plan**
