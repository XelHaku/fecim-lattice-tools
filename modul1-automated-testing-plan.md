# Module 1 Automated Testing Plan - Research-Grade Edition

## Purpose

Achieve research-grade headless automated validation for Module 1 (`module1-hysteresis`) that meets publication standards for ferroelectric physics simulation. This plan transitions from "headless smoke/regression only" to traceable, physics-validated testing with quantitative acceptance criteria aligned with 2024-2026 literature standards.

**Target:** Numerical stability, physics correctness, and reproducibility sufficient for scientific peer review. Every physics output must be testable, traceable, and regression-safe.

Note: filename intentionally follows requested spelling `modul1`.

---

## Current State Assessment (2026-02-13)

| Metric | Current | Target |
|--------|---------|--------|
| Requirements coverage | 51% (17/33) | 100% |
| Test functions | 23 across 11 files (controller + headless) | ~60 across ~20 files |
| Machine-readable JSON artifacts | 2/23 tests | All regression tests |
| Materials in deep regression | 1/9 (literature_superlattice only) | 9/9 |
| Cross-engine Pr agreement | ~30% relative error | <10% |
| Timestep convergence study | None | O(dt^4) verified |
| Golden P-E regression (6+ sig figs) | 1 material (default_hzo) | All 9 materials |
| WriteVerifyStats JSON export | Tracked but NOT exported | Full export |
| Minor loop / FORC validation | None | vs published data |
| Uncertainty propagation | None | Monte Carlo on P-E |
| LGD coefficients vs Materlik 2015 | Exact match confirmed | Maintain |
| TanhEverett product form | Mathematically correct | Maintain |
| LK solver numerical guards | 14 mechanisms present | All tested |
| Fuzz test robustness | 10,000+ steps, zero NaN/Inf | Maintain |

### Implementation Snapshot

- Fully headless gate now enforced in CI and regression scripts (`scripts/ci/go-test-all.sh`, `scripts/ci/go-test-race.sh`, `scripts/run_headless_ispp_regressions.sh`).
- Remaining high-priority gap for Module 1: add required GUI/headless parity test lane for WRD/ISPP phase and target progression (`RG-PAR-04`).

---

## Critical Gaps for Research Grade

| # | Gap | Impact | Tier | Priority |
|---|-----|--------|------|----------|
| 1 | **No timestep convergence study** | Cannot prove numerical accuracy of LK RK4 solver | T1 | CRITICAL |
| 2 | **No baseline comparator** | Silent degradation undetectable | T2 | HIGH |
| 3 | **Cross-engine Pr agreement ~30%** | Cannot claim engine equivalence | T2 | HIGH |
| 4 | **No minor loop validation** against published FORC data | Preisach model unvalidated for sub-coercive regime | T2 | HIGH |
| 5 | **No uncertainty propagation** to output P-E curves | Cannot report confidence intervals | T2 | HIGH |
| 6 | **No golden P-E loop regression** with 6+ significant figures | Physics drift undetectable at research precision | T1 | CRITICAL |
| 7 | **WriteVerifyStats tracked but NOT exported** to regression JSON | ISPP statistics not regression-tested | T2 | HIGH |
| 8 | **8 of 9 materials missing** from deep regression suite | Only literature_superlattice has deep tests | T1 | CRITICAL |
| 9 | **No PUND switching symmetry test** | Core measurement modality untested | T1 | HIGH |
| 10 | **No loop area energy density validation** | Thermodynamic consistency unverified | T2 | MEDIUM |

---

## Research-Grade Target

- Move Module 1 from regression-only confidence to physically defensible validation with traceable evidence.
- Require quantitative acceptance thresholds for switching dynamics, level convergence, retention, and variability.
- Ensure outputs are reproducible and auditable across engines, materials, and seeds.
- Report uncertainty and bounded behavior per engine/model class.
- Keep dual acceptance profiles where physics model maturity differs (Preisach vs LK baseline).

---

## Material-Selected Awareness (Required)

- Every required test must set an explicit material (no implicit default-only runs).
- Material identity must be part of test IDs and regression artifacts (for example: `m1/preisach/fecim_hzo/seed7`).
- Verdicts must be emitted per material and per engine (`Preisach`, `LK`) before aggregate pass/fail.
- Artifacts must include material physics snapshot fields used in that run:
  - at minimum: `Ec`, `Ps`, `Pr`, thickness, `Gmin`, `Gmax`, `TargetRangeFrac`.
- Per-material verdict is mandatory; aggregate pass cannot mask a single failing material.
- Metrics and thresholds must be material-normalized where possible:
  - field normalized by `Ec`
  - polarization normalized by `Ps`
  - level mapping validated against material-specific conductance bounds.

---

## Fully Headless Requirement (Mandatory)

- Required lanes must run with no display stack:
  - `DISPLAY` unset.
  - `WAYLAND_DISPLAY` unset.
  - no `xvfb-run` in mandatory gates.
- Required lanes must execute controller/physics logic only; GUI rendering tests remain optional and non-gating.

## GUI/Headless Physics Parity (Mandatory)

- GUI and headless entry points must call the same write/verify physics controllers and conductance/level mapping.
- No headless-only duplicate physics equations for phase transitions, ISPP stepping, or target sequencing.
- Parity tests must compare GUI-dispatched trajectory and headless trajectory for:
  - phase transitions (`PREP`, `WRITE`, `VERIFY`),
  - target-level sequencing,
  - convergence/termination outcomes.

---

## Source Documentation

### Project Documentation

- `docs/testing/TEST_GUIDE.md` - Testing infrastructure and conventions
- `docs/development/PHYSICS_ACCEPTANCE_CRITERIA.md` - Cross-module physics thresholds
- `docs/development/CI.md` - Continuous integration configuration
- `docs/development/HEADLESS.md` - Headless execution requirements
- `docs/development/scriptReference.md` - Function lookups and error resolution
- `docs/development/TESTING.md` - Test execution reference
- `docs/development/evidence/G08-mid-stability-evidence-2026-02-11.md` - ISPP regression evidence
- `docs/development/GUI/FYNE_NOTES.md` - GUI-specific testing (out of scope for headless)

### Codebase References

- Physics engines: `module1-hysteresis/pkg/ferroelectric/preisach.go`, `shared/physics/landau.go`
- ISPP controllers: `module1-hysteresis/pkg/controller/writer.go`, `shared/physics/ispp_write.go`
- Material presets: `shared/physics/material.go`
- Conductance models: `shared/physics/transfer.go`
- Existing tests: 43 files, 277 functions across module1 and shared/
- Golden data: `validation/testdata/ispp_regression/`, `validation/testdata/physics_regression/`

### Literature Standards (2024-2026)

- PUND (Positive-Up Negative-Down) methodology for P-E loop measurement
- Merz law validation for switching dynamics (tau proportional to exp(-1/E))
- Nucleation-Limited Switching (NLS) kinetics
- Arrhenius retention model: tau(T) = tau_0 * exp(Ea/kT)
- Endurance: 10^9 cycles for HZO research-grade validation
- Multi-level capability: 7+ levels for credibility (state-of-art: 16 levels)
- D2D variability: Monte Carlo with 1000+ virtual devices
- Temperature range: -40C to 125C (automotive) + cryogenic (77K)

---

## Validation Tiers

All tests are organized into three tiers reflecting research maturity requirements.

### Tier 1 -- Critical (Must Pass for Research Release)

These are hard gates. A failure in any Tier 1 test blocks release.

| ID | Requirement | Pass Criterion | Source File(s) |
|----|-------------|----------------|----------------|
| T1.1 | Hysteresis loop Pr vs literature | Relative error <= 5% of published Pr | `shared/physics/literature_validation_test.go` |
| T1.2 | Hysteresis loop Ps vs literature | Relative error <= 5% of published Ps | `shared/physics/literature_validation_test.go` |
| T1.3 | Hysteresis loop Ec vs literature | Relative error <= 5% of published Ec | `shared/physics/literature_validation_test.go` |
| T1.4 | 30-level quantization uniformity | Max level spacing deviation <= 5% of mean spacing | `shared/physics/quantization_test.go` |
| T1.5 | ISPP convergence success rate | >= 95% across all 30 levels per material | `module1-hysteresis/pkg/controller/headless_regression_test.go` |
| T1.6 | PUND switching charge symmetry | \|QP-QU\| approx \|QN-QD\| within 10% | `shared/physics/worldclass_pund_test.go` |
| T1.7 | Golden P-E loop regression (6+ sig figs) | RMS deviation <= 0.01% full-scale per material | `module1-hysteresis/pkg/ferroelectric/golden_regression_test.go` |
| T1.8 | Timestep convergence (RK4) | Richardson extrapolation confirms O(dt^4), error ratio >= 14 for dt halving | NEW: `shared/physics/timestep_convergence_test.go` |
| T1.9 | Zero NaN/Inf in all outputs | All fields finite across 9 materials x 2 engines | `cmd/fecim-lattice-tools/mode_engine_matrix_test.go` |
| T1.10 | Deterministic reproducibility | Identical seeds produce bitwise-identical results | `cmd/fecim-lattice-tools/mode_lk_ispp_convergence_20targets_test.go` |
| T1.11 | Unit consistency | V/m <-> MV/cm automatic validation on all P-E outputs | `shared/physics/electric_field_units_test.go` |
| T1.12 | LGD coefficients match Materlik 2015 | Exact match (alpha, beta, gamma for HZO) | `shared/physics/landau_materlik_test.go` |
| T1.13 | 9-material deep regression | All 9 materials produce valid Pr/Ps/Ec within tolerance bands | `cmd/fecim-lattice-tools/mode_engine_matrix_test.go` (extend) |
| T1.14 | Energy conservation (major loop) | W = integral(E dP) error <= 1% of theoretical work | NEW: thermodynamics test |

### Tier 2 -- Important (Should Pass for Publication)

Failures produce warnings and require documented justification to proceed.

| ID | Requirement | Pass Criterion | Source File(s) |
|----|-------------|----------------|----------------|
| T2.1 | Cross-engine Pr agreement | Preisach vs LK Pr within 10% for same material | `shared/physics/cross_engine_consistency_test.go` |
| T2.2 | FORC Preisach density positivity | rho(Ea,Eb) >= 0 for all grid points within Ea<=Eb | `shared/physics/worldclass_forc_test.go` |
| T2.3 | Retention self-consistency | P(t) monotonically decreasing; P(0) = Pr; P(inf) -> asymptotic | `shared/physics/worldclass_retention_test.go` |
| T2.4 | Wake-up/fatigue model consistency | Pr increases during wake-up, decreases after fatigue onset | `shared/physics/worldclass_wakeup_test.go` |
| T2.5 | AgingEngine coupled model | Wake-up + fatigue + retention compose correctly | `shared/physics/aging_engine_test.go` |
| T2.6 | Minor loop validation | Sub-coercive loops produce non-zero P change (TanhEverett product form) | `module1-hysteresis/pkg/ferroelectric/preisach_test.go` |
| T2.7 | Uncertainty propagation (Monte Carlo) | 100-run ensemble with 5% parameter jitter; output CI covers literature values | `shared/physics/montecarlo_validation_test.go` |
| T2.8 | WriteVerifyStats export to JSON | All stats fields present and parseable in regression artifacts | NEW |
| T2.9 | ISPP voltage trajectory monotonicity | Voltage steps monotonically in the commanded direction (no direction flip) | `module1-hysteresis/pkg/controller/writer_test.go` |
| T2.10 | Baseline comparator (no silent degradation) | KS test p > 0.05 between current run and frozen baseline | NEW: `shared/physics/baseline_comparator_test.go` |
| T2.11 | PUND current transient shape | I(t) has single peak per pulse, decays to baseline | `shared/physics/worldclass_pund_test.go` |
| T2.12 | Loop area energy density | Integral of P-E loop > 0; consistent with 4*Pr*Ec order-of-magnitude | NEW: `shared/physics/loop_area_test.go` |
| T2.13 | Arrhenius retention model | tau(T) = tau_0 * exp(Ea/kT); Ea within 0.6-1.2 eV for HZO | NEW |
| T2.14 | Preisach congruency (quantitative) | Major loop RMS overlap <= 1% Ps after stabilization | NEW |
| T2.15 | Preisach return-point memory | Reversibility error <= 0.5% Ps for sub-coercive excursions | NEW |
| T2.16 | Clausius-Clapeyron relation | dEc/dT slope within +/-10% of theoretical prediction | NEW |

### Tier 3 -- Enhanced (Comprehensive Characterization)

Informational tests that expand coverage breadth. No release gate.

| ID | Requirement | Pass Criterion | Source File(s) |
|----|-------------|----------------|----------------|
| T3.1 | Frequency dispersion log scaling | Ec increases with ln(f); Pr decreases with ln(f) | `shared/physics/worldclass_frequency_dispersion_test.go` |
| T3.2 | C-V butterfly shape | Two capacitance peaks near +/- Ec; C_max/C_min > 2 | `shared/physics/worldclass_cv_test.go` |
| T3.3 | Full 30-level x 9-material characterization | All levels reachable with <= 80 pulses per material | NEW |
| T3.4 | Process variation Monte Carlo | 1000 device instances; Pr spread matches literature sigma | `shared/physics/device_variation_test.go` |
| T3.5 | ResearchTrace signal chain | DAC->Array->TIA->ADC->Classifier uncertainty propagates correctly | `shared/physics/research_trace_test.go` |
| T3.6 | ConfidenceLedger coverage | All output parameters tagged with provenance; no "placeholder" tags in Tier 1 outputs | `shared/physics/confidence_ledger_test.go` |
| T3.7 | CalibrationStudio round-trip | Import CSV -> fit -> export JSON -> reimport matches | `shared/physics/calibration_studio_test.go` |
| T3.8 | Endurance projection | Failure rate vs cycles follows stretched exponential | `shared/physics/endurance_test.go` |
| T3.9 | Fuzz robustness | 10,000+ random steps with zero NaN/Inf | `shared/physics/fuzz_test.go`, `shared/physics/fuzz_property_test.go` |
| T3.10 | Preisach Everett function properties | Product form non-negative for all (alpha, beta); major loop matches tanh model | `module1-hysteresis/pkg/ferroelectric/preisach_equation_test.go` |
| T3.11 | Temperature range (cryogenic to automotive) | Pr(77K)/Pr(300K) approx 1.5; monotonic Pr(T), Ec(T) decrease | `shared/physics/temperature_test.go` |
| T3.12 | L-K Merz law switching | tau proportional to exp(-1/E); R^2 > 0.95 | NEW |
| T3.13 | NLS kinetics (KAI model) | Dimensionality exponent n in [0.8, 2.2] | NEW |
| T3.14 | Multi-level depolarization stability | 8 intermediate P states stable >= 10 us at E=0 | NEW |
| T3.15 | Noise models (RTN, 1/f) | 1/f exponent alpha in [0.8, 1.2] | NEW |
| T3.16 | BER vs level (Monte Carlo) | Adjacent level BER <= 10^-3 for 7-level operation | NEW |
| T3.17 | Conductance model validation | G(P) strictly monotonic; subthreshold n in [1.0, 1.5] | NEW |
| T3.18 | CSV/JSON schema versioning | Column headers, units, version tag in all exports | NEW |

---

## Full Physics Output Catalog

Every output listed here must be (a) computable headlessly, (b) exported to JSON artifacts, and (c) regression-tested.

### 1. Hysteresis Loop Characterization

| Output | Units | Source API | Test Coverage |
|--------|-------|------------|---------------|
| P-E loop time series | E in V/m, P in C/m^2 | `PreisachStack.Update()`, LK solver | T1.7, T1.9 |
| Remanent polarization Pr | C/m^2 | Loop extraction (zero-field intercept) | T1.1, T2.1 |
| Saturation polarization Ps | C/m^2 | Loop extraction (high-field plateau) | T1.2 |
| Coercive field Ec | V/m | Loop extraction (zero-crossing) | T1.3, T2.1 |
| Loop area (energy density) | J/m^3 | Numerical integration of P dE | T1.14, T2.12 |
| 30-level discrete states | Dimensionless (0-29) | `QuantizeTo30Levels()` | T1.4 |
| Level spacing uniformity | Fraction | Level bin analysis via `level_bins.go` | T1.4 |

### 2. PUND Measurement Simulation

| Output | Units | Source API | Test Coverage |
|--------|-------|------------|---------------|
| QP, QU, QN, QD charges | Coulombs | `AnalyzePUND()` in `worldclass_pund.go` | T1.6 |
| Switching charge QP-QU | Coulombs | `PUNDResult.SwitchingPositive_C` | T1.6 |
| Switching charge QN-QD | Coulombs | `PUNDResult.SwitchingNegative_C` | T1.6 |
| Current transient I(t) | Amperes vs seconds | `PulseSample` arrays | T2.11 |
| Symmetry ratio | Dimensionless | \|SwitchingPositive\| / \|SwitchingNegative\| | T1.6 |

### 3. FORC (First-Order Reversal Curves)

| Output | Units | Source API | Test Coverage |
|--------|-------|------------|---------------|
| FORC curve families | P(E) per reversal field | `RunFORCSweep()` in `worldclass_forc.go` | T2.2 |
| Preisach density rho(Ea,Eb) | C/m^2 / (V/m)^2 | `ComputeFORCDensity()` | T2.2 |
| Reversal field pairs | V/m | `FORCResult.ReversalPairs` | T2.2 |
| Density positivity check | Boolean | All rho >= 0 in valid Ea<=Eb region | T2.2 |

### 4. Retention and Aging

| Output | Units | Source API | Test Coverage |
|--------|-------|------------|---------------|
| P(t) retention decay | C/m^2 vs seconds | `SimulateRetentionExponential()` in `worldclass_retention.go` | T2.3 |
| Log-time sweep generation | Seconds array | `GenerateLogTimeSweep()` | T2.3 |
| Wake-up Pr(cycles) | C/m^2 vs cycle count | `WakeUpPolarization()` in `worldclass_wakeup.go` | T2.4 |
| Fatigue onset tracking | Cycle count | `WakeUpModelConfig.FatigueOnsetCycles` | T2.4 |
| Coupled aging model | C/m^2 vs (cycle, hold_time) | `AgingEngine.ApplyCycle()` in `aging_engine.go` | T2.5 |
| Aging history | Cycle list | `AgingEngine.CycleHistory` | T2.5 |

### 5. Frequency Dispersion

| Output | Units | Source API | Test Coverage |
|--------|-------|------------|---------------|
| Pr vs frequency | C/m^2 vs Hz | `ApplyFrequencyDispersion()` in `worldclass_frequency_dispersion.go` | T3.1 |
| Ec vs frequency | V/m vs Hz | `ApplyFrequencyDispersion()` | T3.1 |
| Loop area vs frequency | J/m^3 vs Hz | `ApplyFrequencyDispersion()` | T3.1 |
| Log-scaling coefficients | Dimensionless | `FrequencyDispersionConfig` slopes (EcLogSlope, PrLogSlope, LoopAreaLogSlope) | T3.1 |
| HysteresisMetrics at target freq | Composite struct | `HysteresisMetrics{FrequencyHz, Pr_Cm2, Ec_Vm, LoopArea_Jm3}` | T3.1 |

### 6. C-V Characteristics

| Output | Units | Source API | Test Coverage |
|--------|-------|------------|---------------|
| Butterfly C-V curve | F vs V | `ExtractButterflyCV()` in `worldclass_cv.go` | T3.2 |
| C-V data points | `CVPoint{Voltage_V, Capacitance_F}` array | `ExtractButterflyCV()` | T3.2 |
| Capacitance peaks near +/- Ec | V | Peak detection on `CVPoint` array | T3.2 |
| C_max / C_min ratio | Dimensionless | Max/min of capacitance array | T3.2 |

### 7. ISPP Convergence Statistics

| Output | Units | Source API | Test Coverage |
|--------|-------|------------|---------------|
| Pulses histogram (per target) | Count[10] | `WriteVerifyStats.PulsesHistogram` | T1.5, T2.8 |
| Overall success rate | Fraction [0,1] | `WriteVerifyStats.GetSuccessRate()` | T1.5 |
| Overall failure rate | Fraction [0,1] | `WriteVerifyStats.GetFailureRate()` | T2.8 |
| Overshoot rate | Fraction [0,1] | `WriteVerifyStats.GetOvershootRate()` | T2.8 |
| Per-level success rates | Fraction[256] | `WriteVerifyStats.GetLevelSuccessRates()` | T1.5 |
| Hardest levels identification | Level indices | `WriteVerifyStats.GetHardestLevels(n)` | T2.8 |
| Failure rate vs cycles | Fraction[] | `WriteVerifyStats.GetFailureRateVsCycles()` (every 100 cycles) | T3.8 |
| Average pulses per write | Float | `WriteVerifyStats.GetAveragePulses()` | T2.8 |
| Reset count (overshoot recovery) | Integer | `WriteVerifyStats.ResetCount` | T2.8 |
| Total write time | Microseconds | `WriteVerifyStats.TotalWriteTimeUs` | T2.8 |
| Cycle count | Integer | `WriteVerifyStats.CycleCount` | T3.8 |
| Voltage trajectory | V/m sequence per target | ISPP field log | T2.9 |
| Endurance fatigue projection | Fraction vs cycles | `SimulateFailureRateProgression()` (stretched exponential) | T3.8 |

### 8. Research Trace and Confidence

| Output | Units | Source API | Test Coverage |
|--------|-------|------------|---------------|
| Full inference path trace | Multi-stage JSON | `BuildResearchTrace()` in `research_trace.go` | T3.5 |
| DAC stage | `DACTrace{InputCode, ReferenceVoltage, OutputVoltage, SettlingTime, QuantizationError}` | `BuildResearchTrace()` | T3.5 |
| Array stage | `ArrayTrace{WordlineVoltage, BitlineCurrent, CellConductance, IRDrop, ArrayOutput}` | `BuildResearchTrace()` | T3.5 |
| TIA stage | `TIATrace{InputCurrent, FeedbackOhms, OutputVoltage, InputReferred}` | `BuildResearchTrace()` | T3.5 |
| ADC stage | `ADCTrace{InputVoltage, ResolutionBits, SampleRate, OutputCode, QuantizationNoise}` | `BuildResearchTrace()` | T3.5 |
| Classifier stage | `ClassifierTrace{Logit, Probability, PredictedClass, ConfidenceInterval}` | `BuildResearchTrace()` | T3.5 |
| Per-stage uncertainty (1-sigma) | Various | `TraceValue.Uncertainty` field on all values | T3.5 |
| Parameter provenance tags | Enum + confidence | `ConfidenceLedger.TagOutput()` in `confidence_ledger.go` | T3.6 |
| Provenance categories | `measured` / `calibrated` / `estimated` / `placeholder` | `Provenance` enum | T3.6 |
| Tagged physics value | `TaggedPhysicsValue{Parameter, Value, Tag{Provenance, Confidence}}` | `ConfidenceLedger.TagOutput()` | T3.6 |
| Provenance coverage ratio | Fraction | Non-placeholder / total parameters | T3.6 |

### 9. Calibration

| Output | Units | Source API | Test Coverage |
|--------|-------|------------|---------------|
| Calibration fit result | `CalibrationBundle{Model, Parameters, RMSE, RelativeRMSE}` | `FitCalibration()` in `calibration_studio.go` | T3.7 |
| CSV import | `CalibrationPoint{E, P}` array | `ImportCalibrationCSV()` | T3.7 |
| JSON export | Full bundle | `ExportCalibrationBundle()` | T3.7 |
| Model types | `preisach` (tanh) or `lk` (polynomial) | `CalibrationModel` enum | T3.7 |

---

## Physics Invariants (What MUST Hold)

### Fundamental Thermodynamics

1. **Energy Conservation**: W = integral(E dP) (work per cycle must match enclosed loop area)
2. **Second Law**: Entropy cannot decrease (dS >= 0 for spontaneous switching)
3. **Clausius-Clapeyron**: dEc/dT relation for temperature-dependent coercivity

### Preisach Model Properties

4. **Congruency**: Major loops must overlap within <= 1% after stabilization
5. **Wiping-out**: Return to major loop after minor loop excursion
6. **Return-Point Memory**: Reversibility along nested loop paths
7. **Everett Non-negativity**: Everett function >= 0 for all (alpha, beta) pairs (guaranteed by product form)

### Landau-Khalatnikov Dynamics

8. **Gibbs Free Energy Landscape**: dG/dP = 2*alpha*P + 4*beta*P^3 + 6*gamma*P^5
9. **Switching Dynamics**: Sub-nanosecond capability (research: 360 ps)
10. **Multi-well Stability**: Local minima at +/-Pr under E=0 (single-domain)
11. **Depolarization Field**: K_dep enables intermediate polarization states

### ISPP Programming

12. **Monotonicity**: Field magnitude increases or decreases monotonically per direction
13. **Convergence**: Must reach target level or exceed pulse/overshoot budget
14. **Bounds Integrity**: VMin <= VMax throughout bisection search
15. **Level Accuracy**: +/-1 level for research-grade (strict profile)

### Conductance Mapping

16. **Monotonicity**: G(P) must be monotonic with P for linear/subthreshold models
17. **Range Constraints**: Gmin <= G <= Gmax at all times
18. **Quantization**: Discrete levels match material.TargetLevels specification

### Physical Bounds

19. **Polarization**: |P| <= Ps (spontaneous polarization)
20. **Coercive Field**: Ec > 0 (must require non-zero field to switch)
21. **Remanence**: 0 < Pr < Ps (remanent polarization at E=0)
22. **Temperature**: Ec(T), Pr(T) must decrease monotonically with T up to Curie point

---

## Test Layers (Organized by Physics Domain)

### Layer 1: Thermodynamic Consistency

**Gap:** CRITICAL - No current validation of energy conservation

**Tests:**
- `TestEnergyConservation_MajorLoop`: W = integral(E dP) for full loop
  - Numerical integration error <= 1% of theoretical work
  - Test across all 9 materials x 2 engines
- `TestEnergyConservation_MinorLoops`: Nested loop area conservation
  - Minor loop work must be recoverable (wiping-out property)
- `TestEntropyIncrease_Switching`: dS >= 0 for spontaneous switching events
  - Track Gibbs free energy before/after domain nucleation
- `TestClausius_Clapeyron_Relation`: dEc/dT validation
  - Sweep T from 77K to 400K, verify Ec(T) slope matches theory

**Acceptance:**
- Energy error <= 1% of integral(E dP)
- Entropy dS >= -1e-6 (J/K, numerical tolerance)
- Ec(T) slope within +/-10% of Clausius-Clapeyron prediction

**Implementation Phase:** Phase 5 (New Research-Grade Tests)

---

### Layer 2: Retention and Time-Dependent Physics

**Gap:** CRITICAL - No Arrhenius tau(T) validation

**Tests:**
- `TestRetention_Arrhenius_Model`: tau(T) = tau_0 * exp(Ea/kT)
  - Fit retention data to Arrhenius, extract activation energy Ea
  - Extrapolate to 10 years at 85C (industry standard)
  - Compare Ea against literature: 0.6-1.2 eV for HZO
  - **Uses:** `SimulateRetentionExponential()`, `GenerateLogTimeSweep()`
- `TestRetention_Temperature_Sweep`: 25C, 85C, 125C
  - Verify exponential decay tau proportional to exp(1/T)
- `TestImprint_UnipolarStress`: Horizontal loop shift under DC stress
  - Track coercive field asymmetry over 10^6 pulses
- `TestDrift_LongTerm`: Conductance drift over simulated 10^4 seconds
  - Power-law or logarithmic drift model validation

**Acceptance:**
- Arrhenius fit R^2 > 0.95
- Ea within 0.6-1.2 eV (HZO literature range)
- 10-year retention extrapolation >= 90% for Pr
- Imprint dEc/Ec <= 5% after 10^6 unipolar cycles

**Implementation Phase:** Phase 5 (New Research-Grade Tests)

---

### Layer 3: Preisach Congruency and Memory Properties

**Gap:** HIGH - Qualitative testing only, no quantitative thresholds

**Tests:**
- `TestPreisach_Congruency_Quantitative`: Major loop overlap
  - Run 5 consecutive major loops
  - RMS error between loops 4 and 5 <= 1% of Ps
  - Max point-wise error <= 2% of Ps
- `TestPreisach_WipingOut_Nested`: Minor loop recovery
  - Execute nested minor loops at 0.5Ec, 0.7Ec, 0.9Ec
  - Verify return to major loop within <= 1% after full excursion
  - **Critical for TanhEverett product form:** Sub-coercive loops must produce non-zero P change
- `TestPreisach_ReturnPointMemory_Reversibility`: Reversibility metric
  - Traverse path: E_0 -> E_1 -> E_0, measure P(E_0)_before vs P(E_0)_after
  - Reversibility error <= 0.5% of Ps for |E_1| < Ec
- `TestPreisach_Everett_Positivity`: Everett(alpha,beta) >= 0
  - Sample 1000 random (alpha,beta) pairs within [-3Ec, +3Ec]
  - Assert Everett >= -1e-12 (numerical tolerance)
  - **Product form guarantees this:** W = [1+tanh((a-Ec)/D)] * [1-tanh((b+Ec)/D)] * Ps/4

**Acceptance:**
- Congruency RMS <= 1% Ps, max error <= 2% Ps
- Wiping-out error <= 1% Ps
- Return-point reversibility <= 0.5% Ps
- Everett non-negativity: 100% of samples >= 0

**Implementation Phase:** Phase 5 (New Research-Grade Tests)

---

### Layer 4: Full Temperature Range Validation

**Gap:** HIGH - Missing cryogenic (77K) and automotive (-40C to 125C)

**Tests:**
- `TestTemperature_Cryogenic_77K`: HZO behavior at liquid nitrogen temp
  - Verify Pr(77K) approx 1.5 x Pr(300K) (literature trend)
  - Ec(77K) approx 1.3 x Ec(300K)
- `TestTemperature_Automotive_Range`: -40C to 125C sweep
  - 20 temperature points
  - Verify monotonic Pr(T), Ec(T) decrease with T
  - Linearity check: dPr/dT, dEc/dT within +/-15% of mean slope
- `TestTemperature_CuriePoint_Approach`: Near-Tc behavior
  - For PZT (Tc approx 650K), verify Pr -> 0 as T -> Tc
  - Second-order phase transition signature

**Acceptance:**
- Cryogenic Pr/Ec ratios within +/-20% of literature
- Automotive monotonicity: 100% of points satisfy Pr(T+dT) < Pr(T)
- Curie point: Pr(0.95 x Tc) <= 10% of Pr(300K)

**Implementation Phase:** Phase 6 (Extended Research Features)

---

### Layer 5: Endurance, Wake-up, Fatigue

**Gap:** HIGH - Not integrated with ISPP convergence

**Tests:**
- `TestEndurance_ISPP_Coupling`: 10^6 cycle endurance with periodic ISPP
  - Program 8 levels, cycle 10^5 times, reprogram, repeat 10x
  - Track Pr degradation, ISPP convergence time, pulse count drift
  - **Uses:** `WriteVerifyStats` for per-cycle tracking, `SimulateFailureRateProgression()` for projection
- `TestWakeup_InitialCycles`: O-vacancy redistribution (first 10^3 to 10^6 cycles)
  - Measure Pr(N) for N = 1, 10, 100, 1000, 10^4, 10^5, 10^6
  - Expect Pr increase 20-50% from N=1 to N=10^4 (HZO literature)
  - **Uses:** `WakeUpPolarization()` with `WakeUpModelConfig`
- `TestFatigue_Degradation`: Pr(N) power-law fit
  - Pr(N) = Pr_0 - k * N^beta
  - Verify beta in [0.05, 0.15] (typical ferroelectric fatigue exponent)
- `TestFatigue_ISPP_Pulse_Budget_Drift`: Pulse count vs cycle number
  - At N=10^6, pulse count <= 2x initial pulse count
- `TestAgingEngine_CoupledModel`: Wake-up + fatigue + retention compose correctly
  - **Uses:** `AgingEngine.ApplyCycle(cycle, holdTimeSec)` -- verify wakeup, fatigue, retention factors combine multiplicatively

**Acceptance:**
- 10^6 cycle survival: Pr(10^6) >= 70% of Pr(10^3)
- Wake-up signature: Pr(10^4)/Pr(1) >= 1.2
- Fatigue exponent beta in [0.05, 0.15]
- ISPP pulse drift <= 2x over endurance
- AgingEngine: wakeup * fatigue * retention = observed Pr / PrFresh

**Implementation Phase:** Phase 6 (Extended Research Features)

---

### Layer 6: Spatial Variability and Monte Carlo

**Gap:** HIGH - No clustering, wafer gradient models

**Tests:**
- `TestMonteCarlo_D2D_Variability_1000Devices`: Device-to-device variation
  - 1000 virtual devices with Gaussian parameter scatter
  - sigma(Ec) in [5%, 10%], sigma(Pr) in [3%, 8%] (literature ranges)
  - Report mean, std, P5, P95 for all metrics
- `TestMonteCarlo_SpatialCorrelation_Clustering`: Wafer-level clustering
  - 2D spatial grid (100x100 devices)
  - Gaussian random field with correlation length lambda = 10-50 device pitches
  - Verify spatial autocorrelation function matches model
- `TestMonteCarlo_WaferGradient`: Radial thickness/composition gradient
  - Linear gradient: center to edge 10% variation in thickness
  - Track Ec(r), Pr(r) vs radius
- `TestMonteCarlo_BER_vs_Level`: Bit error rate for MLC
  - 7+ levels, 1000 devices, measure state overlap
  - BER <= 10^-3 for adjacent level separation

**Acceptance:**
- MC 1000 devices: sigma_observed within +/-20% of sigma_input
- Spatial correlation: autocorrelation(lambda) >= 0.6
- BER <= 10^-3 for 7-level operation

**Implementation Phase:** Phase 7 (Advanced Research Features)

---

### Layer 7: Intrinsic Noise Models

**Gap:** MEDIUM - No RTN, 1/f, telegraph noise

**Tests:**
- `TestNoise_RTN_DomainSwitching`: Random Telegraph Noise
  - Single-domain switching events create RTN
  - Measure dwell time distribution: exponential tau_high, tau_low
- `TestNoise_1_f_Spectrum`: 1/f noise power spectral density
  - FFT of conductance time series
  - Verify PSD proportional to 1/f^alpha with alpha in [0.8, 1.2]
- `TestNoise_Telegraph_TwoState`: Two-level fluctuators
  - Bistable trap model with Arrhenius capture/emission
  - Verify Lorentzian spectrum in frequency domain

**Acceptance:**
- RTN: exponential dwell time fit R^2 > 0.9
- 1/f: power-law exponent alpha in [0.8, 1.2]
- Telegraph: Lorentzian corner frequency fc within +/-30% of theory

**Implementation Phase:** Phase 7 (Advanced Research Features)

---

### Layer 8: L-K Switching Dynamics Validation

**Gap:** MEDIUM - Not validated against Merz law or NLS kinetics

**Tests:**
- `TestLK_Merz_Law`: Switching time tau proportional to exp(-1/E)
  - Sweep E from 1.5Ec to 5Ec
  - Fit log(tau) vs 1/E, extract activation field E_a
  - Verify sub-nanosecond capability at high fields
- `TestLK_NLS_Kinetics`: Nucleation-Limited Switching
  - Kolmogorov-Avrami-Ishibashi (KAI) model: P(t) = P_s[1 - exp(-(t/tau)^n)]
  - Fit n (dimensionality parameter), expect n approx 1-2 for thin films
- `TestLK_Switching_Spectroscopy`: tau(E) over 3 decades of field
  - E from 0.5 MV/cm to 5 MV/cm
  - Verify smooth transition from thermally-activated to field-dominated
- `TestLK_Depolarization_MultiLevel`: K_dep stability
  - Verify intermediate P states stable for >= 10 us at E=0
  - Test 8 equally-spaced levels between -Pr and +Pr
- `TestLK_TimestepConvergence`: Richardson extrapolation for RK4 solver
  - Run at dt = {1e-10, 5e-11, 2.5e-11, 1.25e-11} seconds
  - Compute error ratio: must be >= 14 (theoretical 16 for RK4)
  - **This is Tier 1 critical (T1.8)**

**Acceptance:**
- Merz fit R^2 > 0.95, E_a within +/-20% of literature
- NLS: KAI exponent n in [0.8, 2.2]
- Switching: 360 ps achievable at 5xEc
- Multi-level: 8 stable states with lifetime >= 10 us
- Timestep convergence: error ratio >= 14, confirming O(dt^4)

**Implementation Phase:** Phase 2 (timestep convergence) + Phase 6 (remaining dynamics)

---

### Layer 9: ISPP Convergence and Controller Robustness

**Current:** Strong coverage in 9 controller test files
**Gap:** Extend to full material x engine matrix; export WriteVerifyStats to JSON

**Tests (Enhanced):**
- `TestISPP_Convergence_Matrix_9Materials_2Engines`: Existing + extended
  - All 9 materials x 2 engines x 3 target sets
  - Assert: level error <= +/-1 (strict profile) OR bounded completion (relaxed profile)
  - **NEW: Export `WriteVerifyStats` to JSON artifact per run**
- `TestISPP_Bounds_Integrity`: VMin <= VMax throughout execution
  - Inject overshoot scenarios, stuck detection triggers
  - Verify bounds never collapse to [0,0] or [MaxField, MaxField]
- `TestISPP_GuardPulse_Interaction`: Guard-band + ACCEPT+/-1 logic
  - Verify guard pulses <= 2 (current fix)
  - Verify ACCEPT+/-1 threshold = 8 overshoots (current fix)
- `TestISPP_Temperature_Coupling`: ISPP across temperature sweep
  - Program 8 levels at 25C, 85C, 125C
  - Verify convergence at all temperatures
- `TestISPP_WriteVerifyStats_Export`: Validate all stats fields in JSON
  - `PulsesHistogram` has exactly 10 entries
  - `GetLevelSuccessRates()` returns rates for levels 0-29
  - `GetHardestLevels(5)` returns the 5 most difficult levels
  - `GetFailureRateVsCycles()` returns monotonically non-decreasing rates
  - Success rate is a float in [0, 1]

**Acceptance:**
- Strict profile (Preisach): 100% convergence, level error <= +/-1, pulses <= 30
- Relaxed profile (LK baseline): 100% bounded completion, pulses <= 80, overshoots <= 20
- Bounds: 0 collapses across all runs
- Temperature: convergence at all 3 temps
- WriteVerifyStats: all fields present and parseable in JSON

**Implementation Phase:** Phase 1-2 (Consolidate Existing + Extend)

---

### Layer 10: Conductance Model Validation

**Tests:**
- `TestConductance_Linear_Monotonicity`: G(P) strictly increasing
  - Sweep P from -Ps to +Ps in 100 steps
  - Verify G(P_i+1) > G(P_i) for all i
- `TestConductance_Subthreshold_Exponential`: IDS proportional to exp(dVt/(n*Vth))
  - Fit log(IDS) vs dVt, extract subthreshold slope n
  - Verify n in [1.0, 1.5] (typical MOSFET range)
- `TestConductance_Saturation_Quadratic`: IDS proportional to (VGS-VT)^2
  - Fit IDS vs VGS, verify R^2 > 0.98 in saturation region
- `TestConductance_Quantization_Levels`: Discrete level mapping
  - For 30-level material, verify 30 unique G values
  - Verify dG between levels >= 1% of (Gmax-Gmin)

**Acceptance:**
- Linear: 100% monotonicity
- Subthreshold: n in [1.0, 1.5], R^2 > 0.95
- Saturation: R^2 > 0.98
- Quantization: 30 levels resolvable, dG >= 1%

**Implementation Phase:** Phase 4 (Research-Grade Acceptance)

---

### Layer 11: Data Export and Schema Validation

**Current:** CSV/JSON export tested for finite values
**Gap:** Schema versioning, metadata completeness

**Tests (Enhanced):**
- `TestExport_CSV_Schema_Versioning`: Column headers + units
  - Verify columns: `e_field_mv_cm`, `polarization_uc_cm2`, `conductance_s`, `time_s`
  - Verify units in header comments
  - Version schema (e.g., v1.2) in metadata row
- `TestExport_JSON_Metadata_Completeness`: Required fields
  - Material parameters: Ec, Ps, Pr, thickness, Gmin, Gmax, TargetLevels
  - Engine: preisach or lk
  - Controller: waveform or lk_solver
  - Environment: Go version, commit hash, timestamp
- `TestExport_Clipboard_Roundtrip`: Copy -> paste -> parse
  - Export to clipboard, parse as CSV, verify numerical match
- `TestExport_DownstreamCompatibility`: Module2 crossbar ingestion
  - Export conductance array, import into crossbar model
  - Verify MVM operation produces finite results

**Acceptance:**
- CSV: 100% schema compliance, zero missing units
- JSON: 100% required metadata fields present
- Clipboard: roundtrip error <= 1e-12 (floating-point precision)
- Downstream: zero NaN/Inf in crossbar MVM

**Implementation Phase:** Phase 2 (Artifact-Driven Regression)

---

### Layer 12: Determinism and Reproducibility

**Current:** Seed-based determinism tested
**Gap:** Multi-run statistical validation

**Tests (Enhanced):**
- `TestDeterminism_Seed_Reproducibility`: Same seed -> identical output
  - Run ISPP with seed=42, repeat 10 times
  - Verify byte-for-byte identical output (P, E, G, pulse count)
- `TestDeterminism_MultiRun_Statistics`: Statistical confidence intervals
  - 100 runs with different seeds
  - Report mean, std, P5, P95 for pulse count, overshoot count, level error
  - Verify std <= 10% of mean (for well-behaved materials)
- `TestDeterminism_Platform_Independence`: Linux vs macOS vs Windows
  - Run same seed on 3 platforms (CI matrix)
  - Verify results identical to <= 1e-9 (floating-point tolerance)

**Acceptance:**
- Seed reproducibility: 100% byte-for-byte match
- Multi-run: std <= 10% of mean
- Platform: cross-platform error <= 1e-9

**Implementation Phase:** Phase 4 (Research-Grade Acceptance)

---

## ISPP Convergence Physics Properties

Beyond basic "did it converge" validation, these tests validate the physics properties of the ISPP convergence process itself.

**Implementation files:**
- `module1-hysteresis/pkg/controller/ispp_physics_properties_test.go`
- `shared/physics/ispp_physics_properties_test.go`

### P1: Monotonicity

For a sweep of target levels 0..N-1, the achieved conductance level must be monotonically non-decreasing.

**Test:** `TestISPP_P1_Monotonicity_TargetSweep`
- For each material x engine combination, program all N target levels in ascending order.
- Verify achieved_G(target_i) <= achieved_G(target_i+1) for all i.
- Statistical measure: Spearman rank correlation rho between target_level and achieved_G must equal 1.0 exactly (perfect monotone).
- Tolerance: rho = 1.0 (exact). Any rho < 1.0 is a failure indicating level inversion.

### P2: Convergence Rate vs Material Coercivity

Materials with higher Ec require more ISPP iterations to converge because the bisection search covers a wider voltage range.

**Test:** `TestISPP_P2_ConvergenceRate_vs_Coercivity`
- Sweep all 9 materials, record Ec and mean_attempts for mid-range target levels.
- Compute Pearson correlation r between Ec and mean_attempts.
- Acceptance: r > 0.5 (moderate positive correlation).
- Report: r value, 95% confidence interval (Fisher z-transform), p-value, n=9.

### P3: Overshoot Statistics vs Switching Distribution

Materials with sharper switching (lower NLSSigma in NLS model, or steeper tanh in Preisach) exhibit more overshoots because the P-vs-E transfer is more nonlinear near Ec.

**Test:** `TestISPP_P3_Overshoot_vs_SwitchingDistribution`
- Group materials by switching sharpness: sharp (NLSSigma < 0.3), medium (0.3-0.6), broad (> 0.6).
- Record overshoot counts per group.
- Kruskal-Wallis H-test across groups.
- Acceptance: p < 0.05 (groups differ significantly).
- Report: H statistic, p-value, eta-squared (effect size), group medians.

### P4: Minor Loop Closure (Remanent Stability)

After programming a specific level via ISPP, the polarization at E=0 must be stable. Re-applying the same final voltage must trace the same minor-loop branch.

**Test:** `TestISPP_P4_MinorLoopClosure`
- Program target level L. Record P_remanent_1 at E=0.
- Re-apply the same final programming voltage. Record P_remanent_2 at E=0.
- Acceptance: |P_remanent_1 - P_remanent_2| / Ps < 1% for all levels and materials.

### P5: Multi-Level Retention at Zero Field

All N programmed levels must remain distinguishable at zero applied field. The conductance gap between adjacent programmed levels must exceed device-level noise.

**Test:** `TestISPP_P5_MultiLevelRetention_ZeroField`
- Program all N levels. At E=0, measure conductance G for each level.
- Compute min gap: min(G_i+1 - G_i) for adjacent levels.
- Acceptance: min gap > 2 * sigma_device_variation (where sigma is the expected read noise at each level).
- For simulation (no noise model): min gap > 0.5% of (Gmax - Gmin).

### ISPP Convergence CSV Artifact

Every ISPP sweep test must emit a CSV artifact for post-hoc analysis.

**File:** `ispp_sweep_{material}_{engine}.csv`

| Column | Unit | Description |
|--------|------|-------------|
| target_level | - | Integer target level index |
| target_G | S | Target conductance from transfer function |
| achieved_G | S | Actually achieved conductance |
| attempts | - | Number of ISPP pulses used |
| overshoots | - | Number of overshoot events |
| final_P | C/m^2 | Final polarization state |
| pulse_voltages | V | Semicolon-separated list of applied voltages |

---

## Preisach Model Validation

Classical Preisach model properties that MUST hold for any correct implementation. These are mathematical identities of the Preisach formalism, not empirical approximations.

**Implementation files:**
- `module1-hysteresis/pkg/ferroelectric/preisach_validation_test.go`

### PR1: Saturation Symmetry

The major hysteresis loop must be point-symmetric about the origin: P(E) = -P(-E).

**Test:** `TestPreisach_PR1_SaturationSymmetry`
- Trace a full major loop with 1000 field points from -E_sat to +E_sat and back.
- For each point E_i, compare P(E_i) with -P(-E_i).
- Acceptance: max|P(E) + P(-E)| / Ps < 1e-6.

### PR2: Saturation Values

At saturation field, polarization must reach the configured Ps.

**Test:** `TestPreisach_PR2_SaturationValues`
- Apply E = 3*Ec (well into saturation).
- Acceptance: |P(E_sat) - Ps| / Ps < 0.01 (1% tolerance).
- Test for all 9 materials.

### PR3: Coercive Field Accuracy

The measured coercive field (P=0 crossing on descending branch) must match the configured Ec.

**Test:** `TestPreisach_PR3_CoerciveFieldAccuracy`
- Trace major loop, interpolate E where P crosses zero on descending branch.
- Acceptance: |Ec_measured - Ec_configured| / Ec_configured < 1%.
- Test for all 9 materials.

### PR4: Remanent Polarization

Pr measured from P(E=0) on the descending branch of the major loop.

**Test:** `TestPreisach_PR4_RemanentPolarization`
- Trace major loop from +E_sat descending through E=0.
- Record P at E=0.
- Acceptance: |Pr_measured - Pr_configured| / Pr_configured < 5%.
- Test for all 9 materials.

### PR5: Return-Point Memory (Madelung Rules)

The Preisach model must exhibit exact return-point memory: if the field traverses E1 -> E2 -> E1, the polarization must return exactly to its value at E1 before the excursion.

**Test:** `TestPreisach_PR5_ReturnPointMemory`
- Start from major loop at field E1.
- Excurse to E2 (|E2| < |E1|), creating a minor loop.
- Return to E1.
- Acceptance: |P_return - P_initial| / Ps < 1e-10 (machine precision for exact property).
- Test with multiple (E1, E2) pairs across the loop.

### PR6: Wipe-Out Property

When the field exceeds a previous outer reversal point, all inner minor loops must be erased (the Preisach stack shortens).

**Test:** `TestPreisach_PR6_WipeOut`
- Create nested minor loops: E1 -> E2 -> E1 (inner), then exceed E1 to E3 (outer).
- Verify the inner loop history is erased: subsequent traversal from E3 follows the major loop branch, not the minor loop.
- Verification method: check Preisach model internal stack length decreases after wipe-out.
- Acceptance: stack length after wipe-out < stack length before wipe-out.

### PR7: Congruent Minor Loops

Minor loops of the same field amplitude delta_E but at different bias points should produce the same polarization swing delta_P (congruency property).

**Test:** `TestPreisach_PR7_CongruentMinorLoops`
- At bias points E_bias = -0.5Ec, 0, +0.5Ec, apply delta_E = 0.3*Ec minor loops.
- Record delta_P for each.
- Acceptance: coefficient of variation (CV) of delta_P across bias points < 5%.

### Everett Function Properties

The Everett function E(alpha, beta) is the fundamental building block of the Preisach model. These properties are mathematical requirements, guaranteed by the product form `W = [1+tanh((a-Ec)/D)] * [1-tanh((b+Ec)/D)] * Ps/4`.

#### EV1: Non-Negativity

**Test:** `TestPreisach_EV1_EverettNonNegativity`
- Grid sweep: alpha from -3Ec to +3Ec, beta from -3Ec to alpha, in steps of 0.1*Ec.
- Acceptance: Everett(alpha, beta) >= 0 for ALL grid points.
- Total grid points: approximately 1800 per material.
- Test for all 9 materials.

#### EV2: Monotonicity

**Test:** `TestPreisach_EV2_EverettMonotonicity`
- For fixed beta, Everett must be non-decreasing in alpha.
- For fixed alpha, Everett must be non-increasing in beta.
- Grid sweep with same resolution as EV1.
- Acceptance: zero monotonicity violations.

#### EV3: Boundary Conditions

**Test:** `TestPreisach_EV3_EverettBoundary`
- Everett(alpha, alpha) = 0 for all alpha (diagonal is zero).
- Everett(E_sat, -E_sat) = Ps (full switching gives saturation polarization).
- Acceptance: |Everett(a,a)| < 1e-12; |Everett(E_sat, -E_sat) - Ps| / Ps < 1%.

---

## Landau-Khalatnikov Solver Validation

Validation of the L-K ODE solver and its free-energy landscape against analytical predictions.

**Implementation files:**
- `shared/physics/landau_validation_test.go`

### LK1: Remanent Polarization from Free Energy

The remanent polarization Pr corresponds to the minima of the Gibbs free energy G(P) at E=0. Solve dG/dP = 0 analytically or numerically and compare.

**Test:** `TestLK_LK1_RemanentFromFreeEnergy`
- For each material, after `ConfigureFromMaterial()`, find P that minimizes G(P) at E=0.
- Acceptance: |P_solution - Pr_advertised| / Pr < 5%.
- Test for all 9 materials.

### LK2: Coercive Field from Landau Polynomial

The estimated coercive field from the Landau polynomial (via `estimateLandauEc()`) must match the material's configured Ec.

**Test:** `TestLK_LK2_CoerciveFieldFromLandau`
- Call `estimateLandauEc()` after `ConfigureFromMaterial()`.
- Acceptance: |Ec_estimated - Ec_configured| / Ec_configured < 5%.
- Test for all 9 materials.

### LK3: Free Energy Double-Well Structure

The Gibbs free energy G(P) at E=0 must exhibit exactly two minima at approximately +/-Pr and one maximum at P=0. The energy barrier must be thermally significant.

**Test:** `TestLK_LK3_FreeEnergyDoubleWell`
- Evaluate G(P) for P in [-1.5*Ps, +1.5*Ps] at 10000 points.
- Find all local minima and maxima.
- Acceptance:
  - Exactly 2 minima at positions within 10% of +/-Pr.
  - Exactly 1 maximum at P near 0 (|P_max| < 0.1*Pr).
  - Barrier height G(0) - G(Pr) > 100*kT at 300K (thermally stable).

### LK4: Energy Conservation Per Cycle

In the zero-damping limit (gamma_viscosity -> 0), the total energy should be conserved. For finite damping, energy should decrease monotonically (dissipation).

**Test:** `TestLK_LK4_EnergyConservationPerCycle`
- Run one full P-E cycle with very low damping (gamma = 0.01 * nominal).
- Track total energy E_total = G(P) + kinetic at each timestep.
- Acceptance: |delta_E_total| / E_total < 1e-6 per cycle.
- Also verify: for nominal damping, E_total is non-increasing.

### LK5: Polydomain Ensemble Averaging

When using ensemble mode (multiple domains), the average Pr should match the configured Pr. Individual domains should spread around Pr according to the configured distribution.

**Test:** `TestLK_LK5_PolydomainEnsemble`
- Run ensemble with N=100 domains.
- Compute mean Pr across domains.
- Acceptance: |mean(Pr_domains) - Pr_configured| / Pr_configured < 5%.
- Verify: std(Pr_domains) > 0 (domains are NOT identical).
- Verify: individual Pr values span a range of at least 5% of mean Pr.

---

## NLS Stochastic Switching Validation

Validation of the Nucleation-Limited Switching (NLS) model, which governs domain nucleation statistics.

**Implementation files:**
- `module1-hysteresis/pkg/ferroelectric/nls_validation_test.go`

### NLS-1: CDF Shape (Log-Normal Distribution)

The switching time distribution f(t) must follow a log-normal distribution, which is the hallmark of NLS kinetics in polycrystalline ferroelectrics.

**Test:** `TestNLS_1_CDFShape_LogNormal`
- At fixed field E = 2*Ec, collect 10000 switching events.
- Compute the CDF of switching times.
- Create probit plot: probit(f) vs ln(t).
- Fit linear regression to probit plot.
- Acceptance: R^2 > 0.99 (high linearity confirms log-normal).
- Extract mu (mean of ln(t)) and sigma (std of ln(t)).
- Reference: Guo et al., APL 112, 262903 (2018).

### NLS-2: Field Dependence (Merz's Law)

The mean switching time must follow Merz's exponential law: ln(tau_mean) is linear in 1/E.

**Test:** `TestNLS_2_FieldDependence_MerzLaw`
- Sweep E from 1.5*Ec to 5*Ec in 10 steps.
- At each field, collect 1000 switching events, compute tau_mean.
- Fit: ln(tau_mean) = a + b/E.
- Extract activation field E_activation = b.
- Acceptance: |E_activation - E_activation_configured| / E_activation_configured < 10%.
- Report: fitted slope b, R^2, 95% CI for b.
- Reference: Merz, Phys. Rev. 95, 690 (1954).

### NLS-3: Sigma Consistency

The fitted log-normal sigma from NLS-1 must be consistent with the configured NLSSigma material parameter.

**Test:** `TestNLS_3_SigmaConsistency`
- From NLS-1 probit fit, extract sigma.
- Compute 95% CI for sigma (from regression standard error).
- Acceptance: configured NLSSigma falls within the 95% CI of fitted sigma.
- Test for all materials that have NLS parameters defined.

---

## Conductance Transfer Function Validation

Three conductance models (Linear, Subthreshold, Saturation) each must satisfy specific mathematical properties.

**Implementation files:**
- `shared/physics/transfer_validation_test.go`

### R1: Boundary Conditions

**Test:** `TestTransfer_R1_BoundaryConditions`
- For each model, verify:
  - G(-Ps) = Gmin exactly (negative saturation maps to minimum conductance).
  - G(+Ps) = Gmax exactly (positive saturation maps to maximum conductance).
  - G(0) depends on model:
    - Linear: G(0) = (Gmin + Gmax) / 2 (arithmetic mean).
    - Subthreshold: G(0) = sqrt(Gmin * Gmax) (geometric mean).
    - Saturation: G(0) = Gmin + (Gmax - Gmin) / 4 (quarter-point for quadratic).
- Acceptance: exact match for boundary values (|error| < 1e-12). Midpoint within 1% for analytical predictions.
- Test for all 9 materials x 3 models.

### R2: Monotonicity

**Test:** `TestTransfer_R2_Monotonicity`
- For each model, sweep P from -Ps to +Ps in 10000 steps.
- Verify dG/dP > 0 for all P in (-Ps, +Ps).
- Acceptance: zero monotonicity violations.
- Report: min(dG/dP) and location of minimum.
- Test for all 9 materials x 3 models.

### R3: Level Separability (Spacing Pattern)

**Test:** `TestTransfer_R3_LevelSeparability`
- For each model, compute G at all N target levels.
- Verify spacing pattern matches analytical prediction:
  - Linear: delta_G_i = constant (uniform spacing). Max deviation from mean < 1%.
  - Subthreshold: delta_G_i grows exponentially. Ratio delta_G_i+1/delta_G_i = constant. Max deviation from constant ratio < 1%.
  - Saturation: delta_G_i grows linearly (quadratic curve). Max deviation from linear growth < 1%.
- Acceptance: max deviation < 1% for each model's predicted spacing pattern.

---

## Read Disturb Quantification

Quantify the impact of sub-coercive read operations on programmed states.

**Implementation files:**
- `module1-hysteresis/pkg/controller/read_disturb_test.go`

### RD1: Sub-Coercive Field Response

**Test:** `TestReadDisturb_RD1_SubCoerciveResponse`
- For each material x engine, program each target level.
- From each programmed level, apply E_read = 0.1*Ec for 10,000 consecutive steps (simulating 10,000 read operations).
- Track polarization P after each read pulse.
- Acceptance:
  - |delta_P| / Ps per individual read pulse < 1e-6.
  - Cumulative |delta_P_total| after 10,000 reads < 1 level spacing (min gap between adjacent programmed levels).
- Report per level: P_initial, P_final, delta_P, delta_P_per_read.

**CSV artifact:** `read_disturb_{material}.csv`

| Column | Unit | Description |
|--------|------|-------------|
| level | - | Programmed level index |
| read_count | - | Number of read pulses applied |
| P_initial | C/m^2 | Polarization before reads |
| P_final | C/m^2 | Polarization after all reads |
| delta_P | C/m^2 | Total polarization drift |
| delta_P_per_read | C/m^2 | Average drift per read pulse |

---

## Temperature Dependence Validation

Validate the temperature scaling of key ferroelectric parameters against mean-field theory predictions.

**Implementation files:**
- `shared/physics/temperature_validation_test.go`

### T1: Coercive Field Temperature Scaling

Near the Curie temperature Tc, the coercive field follows a power law: Ec(T) = Ec0 * (1 - T/Tc)^beta.

**Test:** `TestTemp_T1_EcTemperatureScaling`
- Sweep temperature from 250K to 400K in 10K steps.
- Measure Ec at each temperature from major loop P=0 crossing.
- Fit: log(Ec) vs log(1 - T/Tc) to extract critical exponent beta.
- Acceptance: beta in [0.4, 0.6] (mean-field prediction: beta = 0.5).
- Report: fitted beta, 95% CI, R^2 of log-log fit.

### T2: Remanent Polarization Temperature Scaling

Same power-law validation for Pr(T).

**Test:** `TestTemp_T2_PrTemperatureScaling`
- Same temperature sweep as T1.
- Measure Pr at each temperature from P(E=0) on descending branch.
- Fit power law: Pr(T) = Pr0 * (1 - T/Tc)^beta.
- Acceptance: beta in [0.4, 0.6] (mean-field).
- Report: fitted beta, 95% CI, R^2.

### T3: Level Distinguishability vs Temperature

As temperature increases, thermal fluctuations reduce the separation between programmed levels.

**Test:** `TestTemp_T3_LevelDistinguishability`
- At each temperature (250K, 300K, 350K, 400K), program all N levels.
- Compute min level spacing as fraction of room-temperature (300K) spacing.
- Acceptance: identify the temperature at which min_spacing drops below 50% of room-temp spacing (operational limit).
- Report: temperature vs min_spacing curve, operational limit temperature.

**CSV artifact:** `temp_dependence_{material}.csv`

| Column | Unit | Description |
|--------|------|-------------|
| temperature | K | Temperature |
| Ec | V/m | Measured coercive field |
| Pr | C/m^2 | Measured remanent polarization |
| min_level_spacing | - | Minimum spacing as fraction of 300K value |

---

## Cross-Engine Consistency

Compare Preisach and Landau-Khalatnikov engines for the same material to quantify systematic differences. Extends NT-6 with additional metrics.

**Implementation files:**
- `module1-hysteresis/pkg/ferroelectric/cross_engine_test.go`

### CE1: Major Loop Pr Agreement

**Test:** `TestCrossEngine_CE1_PrAgreement`
- For each material, run major loop on both Preisach and L-K engines.
- Extract Pr from P(E=0) on descending branch.
- Acceptance: |Pr_LK - Pr_Preisach| / Pr_configured < 15%.
- Test for all 9 materials.

### CE2: Major Loop Ec Agreement

**Test:** `TestCrossEngine_CE2_EcAgreement`
- Same as CE1 but for coercive field.
- Acceptance: |Ec_LK - Ec_Preisach| / Ec_configured < 15%.
- Test for all 9 materials.

### CE3: Loop Area Agreement

**Test:** `TestCrossEngine_CE3_LoopAreaAgreement`
- Compute enclosed area of major P-E loop for both engines (trapezoidal integration).
- Acceptance: |Area_LK - Area_Preisach| / Area_LK < 20%.
- Test for all 9 materials.

### CE4: ISPP Level Reachability

**Test:** `TestCrossEngine_CE4_ISPPLevelReachability`
- For each material, attempt to program all N target levels with both engines.
- Count how many levels successfully converge (within +/-1 level).
- Acceptance: both engines reach >= 90% of configured levels.
- Report: per-engine reachability fraction, list of unreachable levels.

**CSV artifact:** `cross_engine_{material}.csv`

| Column | Unit | Description |
|--------|------|-------------|
| field | V/m | Applied electric field |
| P_landau | C/m^2 | Polarization from L-K engine |
| P_preisach | C/m^2 | Polarization from Preisach engine |
| delta_P | C/m^2 | Absolute difference |

---

## Material Validation Suite

Parameter consistency checks for ALL 9 material presets. These are sanity checks on the material definition itself, not the physics engine.

**Implementation files:**
- `shared/physics/material_validation_test.go`

### M1: Fundamental Parameter Bounds

**Test:** `TestMaterial_M1_ParameterBounds`
- For all 9 materials, verify:
  - Pr < Ps (remanence cannot exceed saturation).
  - Ec * Thickness in [0.1V, 10V] (switching voltage in practical range).
  - Gmin < Gmax (conductance ordering).
  - Gmin > 0 (positive conductance).
  - Gmax > 0 (positive conductance).
- Acceptance: all constraints satisfied for all materials.

### M2: Landau Coefficient Consistency

**Test:** `TestMaterial_M2_LandauCoefficients`
- For all materials with Landau coefficients defined:
  - Beta < 0 (negative for first-order phase transition).
  - Gamma > 0 (positive for stability of free energy at large P).
- Acceptance: all constraints satisfied.

### M3: Depolarization and Analog Levels

**Test:** `TestMaterial_M3_DepolarizationAnalog`
- For all materials:
  - K_dep > 0 (depolarization field enables intermediate states).
  - TargetLevels >= 2 (at least binary operation).
- Acceptance: all constraints satisfied.

### M4: NLS Parameter Consistency

**Test:** `TestMaterial_M4_NLSParameters`
- For materials with NLS parameters:
  - Tau0NLS < TauInf (initial time constant less than saturation).
  - NLSSigma > 0 (positive distribution width).
  - NLSSigma < 2.0 (physically reasonable; typical range 0.1-1.0).
- Acceptance: all constraints satisfied.

---

## Statistical Validation Framework

All quantitative claims in this test plan must be backed by appropriate statistical tests with standardized reporting.

### Test Selection Guide

| Claim Type | Statistical Test | Report Format |
|------------|------------------|---------------|
| "X increases with Y" | Pearson r (if linear) or Spearman rho (if monotone) | r or rho, 95% CI, p-value, n |
| "X is log-normal" | Probit plot linear regression | R^2, fitted mu, fitted sigma |
| "Groups differ" | Kruskal-Wallis H-test (non-parametric) | H statistic, p-value, eta-squared (effect size) |
| "X scales as N^b" | Log-log linear regression | fitted b, 95% CI for b, R^2 |
| "X matches reference value" | Relative error | |X - ref| / ref, tolerance, n |
| "Distribution matches model" | Kolmogorov-Smirnov test (using `validation.KolmogorovSmirnovTest`) | D statistic, p-value |
| "Two distributions identical" | Mann-Whitney U test | U, p-value, effect size r |

### Confidence Interval Methods

- **Normal data** (verified by Shapiro-Wilk): parametric CI using t-distribution.
- **Skewed or count data**: bootstrap CI with 10,000 resamples (BCa method).
- **Proportions** (e.g., convergence rate): Wilson score interval.

### Effect Size Reporting

All statistical tests must report effect size in addition to p-value:
- Correlation: r or rho (small: 0.1, medium: 0.3, large: 0.5).
- Group comparison: eta-squared (small: 0.01, medium: 0.06, large: 0.14).
- Two-sample: Cohen's d or rank-biserial r.

### Multiple Comparison Correction

When running the same statistical test across multiple materials (n=9), apply Bonferroni correction: adjusted alpha = 0.05 / 9 = 0.0056.

---

## Research Data Artifacts

CSV artifacts the test suite must produce for post-hoc analysis, publication figures, and reproducibility.

### Required Artifacts

| Artifact Filename | Columns | Producer Test(s) |
|--------------------|---------|-------------------|
| `ispp_sweep_{material}_{engine}.csv` | target_level, target_G(S), achieved_G(S), attempts, overshoots, final_P(C/m^2) | P1-P5, Layer 9 |
| `pe_loop_{material}_{engine}.csv` | E(V/m), P(C/m^2), direction, branch | PR1-PR4, CE1-CE3 |
| `material_params_validated.csv` | material, Ec(V/m), Pr(C/m^2), Ps, Ec_theory, Pr_theory, Ec_error%, Pr_error% | M1-M4, LK1-LK2 |
| `nls_switching_times_{material}.csv` | field(V/m), time(s), switched_fraction, ln_time, probit_f | NLS-1, NLS-2, NLS-3 |
| `cross_engine_{material}.csv` | field(V/m), P_landau(C/m^2), P_preisach(C/m^2), delta_P | CE1-CE3 |
| `level_retention_{material}.csv` | level, G_programmed(S), G_after_hold(S), drift(S), hold_time(s), temperature(K) | Layer 2, T3 |
| `read_disturb_{material}.csv` | level, read_count, P_initial, P_final, delta_P, delta_P_per_read | RD1 |
| `temp_dependence_{material}.csv` | temperature(K), Ec(V/m), Pr(C/m^2), min_level_spacing | T1, T2, T3 |

### Artifact Storage

- Output directory: `output/research_artifacts/module1/`
- Naming convention: `{test_id}_{material}_{engine}_{timestamp}.csv`
- Header row: column names with units in parentheses
- Metadata comment line: `# material={name} engine={engine} seed={seed} commit={hash} date={ISO8601}`

### Artifact Generation Control

- Default: artifacts are NOT generated during CI (fast/nightly lanes).
- Enable: `FECIM_EMIT_ARTIFACTS=1` environment variable.
- Always enabled: during release gate and research suite runs.

---

## Publication-Quality Figures

Figures the research suite must be capable of generating for publication or internal review.

### Figure 1: P-E Hysteresis Loop Gallery

- All 9 materials overlaid on a single normalized (P/Ps vs E/Ec) plot.
- Ec and Pr annotated with arrows for each material.
- Both Preisach and L-K traces (solid vs dashed).
- Output: `output/figures/pe_loop_gallery.png` (300 DPI, 8x6 inches).

### Figure 2: ISPP Convergence Staircase

- G (conductance) vs target level index for a representative material.
- Error bars showing +/-1 sigma across 100 seeds.
- Ideal staircase overlaid for comparison.
- Output: `output/figures/ispp_staircase_{material}.png`.

### Figure 3: ISPP Attempt Distribution

- Histogram of attempt counts across all materials and levels.
- Separate panels (or colors) for Preisach vs L-K engines.
- Vertical lines at median and 95th percentile.
- Output: `output/figures/ispp_attempts_histogram.png`.

### Figure 4: NLS Probit Plot

- ln(time) vs probit(switched fraction) at 3-5 different field values.
- Linear fit lines overlaid with R^2 annotation.
- Reference: matches format of Guo et al., APL 112, 262903 (2018), Fig. 3.
- Output: `output/figures/nls_probit_{material}.png`.

### Figure 5: Cross-Engine P-E Comparison

- For each material: Preisach P-E loop and L-K P-E loop overlaid.
- Residual (P_LK - P_Preisach) plotted below.
- Pr and Ec mismatch annotated.
- Output: `output/figures/cross_engine_{material}.png`.

### Figure 6: Temperature Sensitivity

- Ec(T) and Pr(T) vs temperature with power-law fit overlay.
- Fitted exponent beta annotated with 95% CI.
- Tc marked with vertical dashed line.
- Output: `output/figures/temp_sensitivity_{material}.png`.

---

## Test Classification Tiers

Tests are classified into tiers by strictness and execution frequency. This extends the Validation Tiers above with the new test IDs from the physics properties sections.

| Tier | Description | Example Tests | Tolerance | When to Run |
|------|-------------|---------------|-----------|-------------|
| **T0: Physics Identity** | Mathematical identities that must hold exactly | P-E symmetry (PR1), Everett >= 0 (EV1), conductance monotonicity (R2), seed determinism, return-point memory (PR5) | Exact or < 1e-10 | Every PR |
| **T1: Model Validation** | Physics model correctness within calibration tolerance | Landau Ec/Pr (LK1-LK2), NLS log-normal shape (NLS-1), ISPP monotonicity (P1), boundary conditions (R1), material sanity (M1-M4) | < 5% | Every PR |
| **T2: Cross-Engine** | Engine-to-engine agreement (inherently looser due to different formalisms) | Pr agreement (CE1), Ec agreement (CE2), loop area (CE3), level reachability (CE4), temperature scaling (T1-T3) | < 15-20% | Nightly |
| **T3: Research Artifacts** | Publication-quality data generation and figure export | Figure generation, full material sweep, CSV artifacts, publication data | No hard pass/fail | Release / On-demand |

### Tier Assignment for New Test IDs

| Test ID | Tier | Rationale |
|---------|------|-----------|
| PR1, PR5, EV1, EV3 | T0 | Mathematical identities of Preisach formalism |
| EV2, PR6 | T0 | Structural properties (monotonicity, stack behavior) |
| P1 (monotonicity) | T0 | Achieved levels must be monotone (physics identity) |
| R2 (conductance monotonicity) | T0 | Transfer function must be monotone |
| PR2, PR3, PR4, PR7 | T1 | Calibration-dependent Preisach validation |
| LK1, LK2, LK3, LK4, LK5 | T1 | Landau model validation |
| NLS-1, NLS-2, NLS-3 | T1 | NLS statistical model validation |
| R1, R3 | T1 | Transfer function model validation |
| P2, P3, P4, P5 | T1 | ISPP convergence physics properties |
| RD1 | T1 | Read disturb quantification |
| M1, M2, M3, M4 | T1 | Material parameter sanity |
| CE1, CE2, CE3, CE4 | T2 | Cross-engine comparison |
| T1, T2, T3 (temp tests) | T2 | Temperature-dependent level spacing |
| Figures 1-6 | T3 | Publication artifacts |
| Full CSV artifact generation | T3 | Research data output |

---

## Literature References

| Reference ID | Citation | DOI / Source | Used For |
|--------------|----------|--------------|----------|
| Park2015 | Park et al., Adv. Mater. 27, 1811 (2015) | 10.1002/adma.201404531 | HZO Pr, Ec bounds; foundational HZO ferroelectricity |
| Materlik2015 | Materlik et al., J. Appl. Phys. 117, 134109 (2015) | 10.1063/1.4916707 | LGD coefficients for HfO2 polymorphs; used in T1.12 |
| Cheema2020 | Cheema et al., Nature 580, 478 (2020) | 10.1038/s41586-020-2208-x | Superlattice Pr values, ultrathin ferroelectrics |
| Guo2018 | Guo et al., APL 112, 262903 (2018) | 10.1063/1.5030178 | NLS sigma, switching time distributions; used in NLS-1 |
| Oh2017 | Oh et al., IEEE EDL 38(6), 732 (2017) | 10.1109/LED.2017.2698083 | 32-level FeFET multi-level cell |
| Song2024 | Song et al., Adv. Science (2024) | -- | 140-level FTJ analog programming |
| Merz1954 | Merz, Phys. Rev. 95, 690 (1954) | 10.1103/PhysRev.95.690 | Merz's exponential switching law; used in NLS-2, Layer 8 |

---

## New Tests to Implement (Priority Order)

### NT-1: Timestep Convergence Study (Tier 1, CRITICAL)

**File:** `shared/physics/timestep_convergence_test.go` (new)

**Purpose:** Prove the LK RK4 solver achieves its theoretical O(dt^4) convergence rate.

**Method:**
1. Run LK hysteresis loop at dt = {1e-10, 5e-11, 2.5e-11, 1.25e-11} seconds.
2. Extract Pr at each dt.
3. Compute Richardson error ratios: `(Pr[i-1] - Pr[i]) / (Pr[i] - Pr[i+1])`.
4. For RK4, expected ratio = 2^4 = 16. Accept >= 14 (allowing for nonlinearity).
5. Emit JSON artifact with full convergence table.

**Pass criteria:**
- Error ratio >= 14 for at least one pair of consecutive halvings.
- Final Pr within 0.1% of Richardson-extrapolated value.
- All runs produce identical Pr for same dt (determinism).

**JSON output block:** See "Timestep Convergence Block" in JSON Schema section.

---

### NT-2: Golden P-E Loop Regression for All 9 Materials (Tier 1, CRITICAL)

**File:** `module1-hysteresis/pkg/ferroelectric/golden_regression_test.go` (extend existing)

**Purpose:** Detect physics drift at 6+ significant figure precision across all materials.

**Method:**
1. For each of 9 materials, generate a complete P-E hysteresis loop with fixed parameters.
2. Compare point-by-point against frozen golden JSON in `validation/testdata/physics_regression/`.
3. RMSE must be <= 0.01% of Ps for that material.

**Golden data regeneration:** `FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ...`

**Currently:** Only `preisach_loop_default_hzo.json` exists. Need 8 more golden files:
- `preisach_loop_fecim_hzo.json`
- `preisach_loop_literature_superlattice.json`
- `preisach_loop_fecim_hzo_target.json`
- `preisach_loop_cryogenic_hzo.json`
- `preisach_loop_hzo_standard_32.json`
- `preisach_loop_hzo_ftj_140.json`
- `preisach_loop_hzo_custom_14.json`
- `preisach_loop_alscn.json`

---

### NT-3: Baseline Comparator (Tier 2, HIGH)

**File:** `shared/physics/baseline_comparator_test.go` (new)

**Purpose:** Detect silent degradation by comparing current outputs against a frozen baseline.

**Method:**
1. Run standard P-E loop for each PR-gate material.
2. Load frozen baseline from `validation/testdata/physics_regression/`.
3. Apply KS test (`validation.KolmogorovSmirnovTest`) between current and baseline P arrays.
4. Apply RMSE check (`validation.RootMeanSquaredError`) normalized by Ps.
5. Compute correlation coefficient between current and baseline loop shapes.

**Pass criteria:**
- KS test p > 0.05 (distributions not significantly different).
- RMSE < 5% of Ps.
- Correlation r > 0.95.

---

### NT-4: WriteVerifyStats JSON Export (Tier 2, HIGH)

**File:** `module1-hysteresis/pkg/controller/headless_regression_test.go` (extend)

**Purpose:** Export all `WriteVerifyStats` fields to regression JSON artifacts.

**Method:**
1. After each ISPP run, serialize `WriteVerifyStats` to the `write_verify_stats` JSON block.
2. Include: `PulsesHistogram`, success rate, overshoot rate, per-level success rates (30 levels), hardest levels, reset count, average pulses, total write time, cycle count.
3. Verify all fields are present and parseable in the emitted JSON.

**Pass criteria:**
- JSON artifact contains `write_verify_stats` key with all required fields.
- `success_rate` is a float in [0, 1].
- `pulses_histogram` has exactly 10 entries (matching `WriteVerifyStats.PulsesHistogram[10]`).
- `level_success_rates_30` has entries for levels 0-29.
- `hardest_levels` contains valid level indices.

---

### NT-5: PUND Switching Symmetry Test (Tier 1, HIGH)

**File:** `shared/physics/worldclass_pund_test.go` (extend)

**Purpose:** Validate PUND measurement produces physically correct switching charges.

**Method:**
1. Generate P/U/N/D current pulse traces from simulated hysteresis switching.
2. Call `AnalyzePUND()` with `PulseSample` traces (each with TimeS, CurrentA fields).
3. Check `|SwitchingPositive_C|` approximately equals `|SwitchingNegative_C|` within 10%.
4. Check QP_C > QU_C (program pulse has more charge than non-switching up pulse).
5. Check all charges are finite and have correct signs.

**Pass criteria:**
- `||SwitchingPositive_C| - |SwitchingNegative_C|| / max(|SP|, |SN|) < 0.10`
- QP_C > QU_C, |QN_C| > |QD_C|
- All charge values finite.

---

### NT-6: Cross-Engine Pr Agreement Improvement (Tier 2, HIGH)

**File:** `shared/physics/cross_engine_consistency_test.go` (extend)

**Purpose:** Track and enforce Preisach vs LK Pr agreement.

**Method:**
1. For each material, run both Preisach and LK engines with identical parameters.
2. Extract Pr from each.
3. Compute relative error: `|Pr_preisach - Pr_lk| / Pr_literature`.
4. Use `validation.RelativeError()` and `validation.WithinTolerance()`.

**Pass criteria (target, phased enforcement):**
- Phase 1: Track relative error, emit to artifact (informational).
- Phase 2: Warn if relative error > 20%.
- Phase 3: Fail if relative error > 10% for PR-gate materials.

---

### NT-7: FORC Density Positivity and Shape (Tier 2)

**File:** `shared/physics/worldclass_forc_test.go` (extend)

**Purpose:** Validate Preisach density rho(Ea,Eb) from FORC analysis.

**Method:**
1. Run `RunFORCSweep(model, Emax, numReversals)` with numReversals >= 20.
2. Call `ComputeFORCDensity()` on result.
3. Check all `rho[i][j] >= 0` where `Ea <= Eb` (valid Preisach half-plane).
4. Check at least some `rho > 0` (density is not trivially zero).
5. Verify `ReversalPairs` contains valid (Ea, Eb) coordinate pairs.

**Pass criteria:**
- Zero negative density values in the valid Ea <= Eb region.
- At least 50% of valid grid points have rho > 0.
- All values finite.

---

### NT-8: Loop Area Energy Density (Tier 2)

**File:** `shared/physics/loop_area_test.go` (new)

**Purpose:** Validate P-E loop area is physically meaningful (energy conservation).

**Method:**
1. Generate saturating P-E loop using `PEPoint{Field_Vm, Polarization_Cm}` samples.
2. Numerically integrate P dE around the loop (trapezoidal rule).
3. Compare against `4 * Pr * Ec` order-of-magnitude estimate.

**Pass criteria:**
- Loop area > 0 for saturating field.
- Loop area within 0.5x to 2.0x of `4 * Pr * Ec`.
- Loop area finite.

---

### NT-9: 9-Material Deep Regression Suite (Tier 1, CRITICAL)

**File:** `cmd/fecim-lattice-tools/mode_engine_matrix_test.go` (extend)

**Purpose:** Extend existing NaN/crash checks to include deep physics validation for all 9 materials.

**Method:**
1. For each of 9 materials x 2 engines:
   - Run hysteresis loop.
   - Extract Pr, Ps, Ec using loop analysis.
   - Compare against material-specific literature ranges (use `validation.WithinTolerance()`).
   - Run ISPP with lo/mid/hi targets.
   - Record `WriteVerifyStats`.
2. Emit per-material JSON artifact with `physics_validation` block.
3. Emit per-material JSON artifact with `write_verify_stats` block.

**Pass criteria:**
- All 18 combinations (9 x 2) produce finite results.
- Pr, Ps, Ec within material-specific tolerance bands (5-10% depending on material).
- ISPP success rate >= 95% for Preisach (Profile A), >= 80% for LK (Profile B).

---

## JSON Schema for Regression Artifacts

Current schema has 11 metrics/case. The following extended schema is required for research-grade artifacts.

### Source Metadata Block (required in every artifact)

```json
{
  "source_metadata": {
    "git_hash": "abc1234",
    "git_dirty": false,
    "seed": 42,
    "timestamp_utc": "2026-02-13T10:30:00Z",
    "go_version": "1.23.0",
    "env_vars": {
      "FECIM_MATERIAL": "literature_superlattice",
      "FECIM_RANGE_FRAC": "0.8",
      "FECIM_ISPP_STEPS_PER_PULSE": "200"
    },
    "material_snapshot": {
      "Ec_Vm": 1e8,
      "Ps_Cm2": 0.54,
      "Pr_Cm2": 0.40,
      "thickness_m": 1e-8,
      "Gmin_S": 1e-6,
      "Gmax_S": 1e-4,
      "TargetRangeFrac": 0.8
    }
  }
}
```

### Write-Verify Statistics Block (new -- maps to `WriteVerifyStats` struct)

```json
{
  "write_verify_stats": {
    "total_writes": 90,
    "successful_writes": 87,
    "failed_writes": 3,
    "success_rate": 0.9667,
    "avg_pulses_per_write": 3.2,
    "overshoot_count": 5,
    "overshoot_rate": 0.0556,
    "reset_count": 2,
    "cycle_count": 90,
    "total_write_time_us": 450.0,
    "pulses_histogram": [12, 25, 18, 15, 8, 5, 2, 1, 0, 1],
    "level_success_rates_30": [1.0, 1.0, 0.95, 0.90, 1.0, "...30 entries total..."],
    "hardest_levels": [14, 15, 16],
    "failure_rate_history": [0.001, 0.002, 0.003]
  }
}
```

**Field mapping from `WriteVerifyStats`:**
- `total_writes` -> `TotalWrites`
- `successful_writes` -> `SuccessfulWrites`
- `failed_writes` -> `FailedWrites`
- `success_rate` -> `GetSuccessRate()`
- `avg_pulses_per_write` -> `GetAveragePulses()`
- `overshoot_count` -> `OvershootCount`
- `overshoot_rate` -> `GetOvershootRate()`
- `reset_count` -> `ResetCount`
- `cycle_count` -> `CycleCount`
- `total_write_time_us` -> `TotalWriteTimeUs`
- `pulses_histogram` -> `GetPulsesHistogram()` (10 entries)
- `level_success_rates_30` -> `GetLevelSuccessRates()[:30]`
- `hardest_levels` -> `GetHardestLevels(5)`
- `failure_rate_history` -> `GetFailureRateVsCycles()`

### Physics Validation Block (new)

```json
{
  "physics_validation": {
    "Pr_extracted_Cm2": 0.398,
    "Pr_literature_Cm2": 0.40,
    "Pr_relative_error": 0.005,
    "Ps_extracted_Cm2": 0.535,
    "Ps_literature_Cm2": 0.54,
    "Ps_relative_error": 0.009,
    "Ec_extracted_Vm": 9.8e7,
    "Ec_literature_Vm": 1.0e8,
    "Ec_relative_error": 0.02,
    "loop_area_Jm3": 2.5e7,
    "quantization_uniformity": 0.97,
    "cross_engine_Pr_error": 0.08
  }
}
```

### Confidence Tags Block (new -- maps to `ConfidenceLedger`)

```json
{
  "confidence_tags": {
    "Pr": {"provenance": "measured", "confidence": 0.95},
    "Ps": {"provenance": "measured", "confidence": 0.93},
    "Ec": {"provenance": "measured", "confidence": 0.92},
    "beta_landau": {"provenance": "calibrated", "confidence": 0.86},
    "gamma_landau": {"provenance": "calibrated", "confidence": 0.86},
    "rho_viscosity": {"provenance": "estimated", "confidence": 0.72}
  }
}
```

**Provenance values** (from `confidence_ledger.go`):
- `measured` -- experimentally measured parameter
- `calibrated` -- fitted to experimental data
- `estimated` -- derived from theory or analogy
- `placeholder` -- default value, not validated

### Timestep Convergence Block (new)

```json
{
  "timestep_convergence": {
    "method": "richardson_extrapolation",
    "solver": "rk4",
    "dt_sequence": [1e-10, 5e-11, 2.5e-11, 1.25e-11],
    "Pr_at_dt": [0.3980, 0.3992, 0.3995, 0.3996],
    "error_ratio_sequence": [null, null, 15.2, 15.8],
    "expected_ratio_rk4": 16.0,
    "converged": true,
    "extrapolated_Pr": 0.3996
  }
}
```

---

## Existing Infrastructure Reference

### Physics Simulation APIs

| File | Key APIs | Used By Tests |
|------|----------|---------------|
| `shared/physics/worldclass_pund.go` | `AnalyzePUND(P, U, N, D []PulseSample)`, `IntegrateCurrent([]PulseSample)` | T1.6, T2.11 |
| `shared/physics/worldclass_forc.go` | `RunFORCSweep(*PreisachStack, Emax, numReversals)`, `ComputeFORCDensity(FORCResult)` | T2.2 |
| `shared/physics/worldclass_retention.go` | `SimulateRetentionExponential(P0, Pinf, tau, times)`, `GenerateLogTimeSweep(tMin, tMax, points)` | T2.3 |
| `shared/physics/worldclass_wakeup.go` | `WakeUpPolarization(cycles, WakeUpModelConfig)` | T2.4 |
| `shared/physics/worldclass_frequency_dispersion.go` | `ApplyFrequencyDispersion(base HysteresisMetrics, targetHz, FrequencyDispersionConfig)` | T3.1 |
| `shared/physics/worldclass_cv.go` | `ExtractButterflyCV([]PEPoint, areaM2, thicknessM)` | T3.2 |
| `shared/physics/aging_engine.go` | `NewAgingEngine(prFresh)`, `AgingEngine.ApplyCycle(cycle, holdTimeSec)` | T2.5 |
| `shared/physics/write_verify_stats.go` | `WriteVerifyStats` struct, `RecordWrite()`, `GetSuccessRate()`, `GetLevelSuccessRates()`, `GetHardestLevels(n)`, `GetFailureRateVsCycles()`, `GetOvershootRate()`, `GetAveragePulses()`, `GetPulsesHistogram()`, `SimulateFailureRateProgression()` | T1.5, T2.8, T3.8 |
| `shared/physics/research_trace.go` | `BuildResearchTrace(inputCode, gCell, tiaFeedback, adcBits)`, `ResearchTrace` struct, `TraceValue{Value, Unit, Uncertainty}` | T3.5 |
| `shared/physics/confidence_ledger.go` | `NewConfidenceLedger()`, `ConfidenceLedger.Register()`, `ConfidenceLedger.TagOutput()`, `TaggedPhysicsValue`, `ConfidenceTag{Provenance, Confidence}` | T3.6 |
| `shared/physics/calibration_studio.go` | `FitCalibration(points, model, iterations, seed)`, `ImportCalibrationCSV(path)`, `ExportCalibrationBundle(path, bundle)`, `CalibrationBundle` struct | T3.7 |
| `shared/physics/characterization.go` | `CharacterizationResult{WriteTimeNs, ReadTimeNs, WriteEnergy_fJ, ReadEnergy_fJ}` | T3.18 |

### Validation Statistics APIs

| File | API | Purpose | Used By |
|------|-----|---------|---------|
| `validation/statistics.go` | `KolmogorovSmirnovTest(s1, s2 []float64) (statistic, pValue)` | Two-sample distribution comparison | NT-3 baseline comparator |
| `validation/statistics.go` | `ChiSquaredTest(observed, expected []float64) (chiSq, df)` | Goodness-of-fit | Pulses histogram validation |
| `validation/statistics.go` | `RootMeanSquaredError(measured, expected []float64) float64` | Loop shape comparison | NT-2 golden regression, NT-3 baseline |
| `validation/statistics.go` | `MeanAbsoluteError(measured, expected []float64) float64` | Trend detection | Multi-run sweeps |
| `validation/statistics.go` | `RelativeError(measured, expected float64) float64` | Parameter comparison | NT-6, NT-9 literature comparison |
| `validation/statistics.go` | `WithinTolerance(measured, expected, tolerancePct float64) bool` | Tolerance checking | All tolerance checks |
| `validation/statistics.go` | `Mean(values []float64) float64` | Arithmetic mean | Multi-run sweeps |
| `validation/statistics.go` | `StandardDeviation(values []float64) float64` | Sample std dev | Multi-run sweeps |

### Existing Test Files (Module 1 Scope)

**Controller tests (9 files):**

| File | Focus | Tier Coverage |
|------|-------|---------------|
| `module1-hysteresis/pkg/controller/headless_regression_test.go` | ISPP regression with JSON | T1.5, T2.8 |
| `module1-hysteresis/pkg/controller/writer_test.go` | Writer algorithm invariants | T2.9 |
| `module1-hysteresis/pkg/controller/writer_stress_test.go` | Stress scenarios | T1.5 |
| `module1-hysteresis/pkg/controller/writer_extended_test.go` | Extended writer tests | T2.9 |
| `module1-hysteresis/pkg/controller/writer_lk_tuning_test.go` | LK tuning parameters | T2.1 |
| `module1-hysteresis/pkg/controller/ispp_full_cycle_test.go` | Full ISPP cycle | T1.5 |
| `module1-hysteresis/pkg/controller/ispp_convergence_test.go` | Convergence behavior | T1.5 |
| `module1-hysteresis/pkg/controller/ispp_landau_ensemble_test.go` | LK ensemble (superlattice) | T1.5, T2.1 |
| `module1-hysteresis/pkg/controller/landau_remanent_sweep_test.go` | Remanent sweep | T1.1 |

**Ferroelectric model tests:**

| File | Focus | Tier Coverage |
|------|-------|---------------|
| `module1-hysteresis/pkg/ferroelectric/golden_regression_test.go` | Golden loop regression | T1.7 |
| `module1-hysteresis/pkg/ferroelectric/preisach_test.go` | Preisach model | T2.6, T3.10 |
| `module1-hysteresis/pkg/ferroelectric/preisach_equation_test.go` | Everett function | T3.10 |
| `module1-hysteresis/pkg/ferroelectric/preisach_calibration_test.go` | Preisach calibration | T3.7 |
| `module1-hysteresis/pkg/ferroelectric/preisach_properties_table_test.go` | Properties table | T2.14 |
| `module1-hysteresis/pkg/ferroelectric/physics_validation_test.go` | Physics bounds | T1.1-T1.3 |
| `module1-hysteresis/pkg/ferroelectric/hysteresis_loop_test.go` | Loop shape | T1.7 |
| `module1-hysteresis/pkg/ferroelectric/level_bins_test.go` | Level binning | T1.4 |

**Shared physics tests:**

| File | Focus | Tier Coverage |
|------|-------|---------------|
| `shared/physics/cross_engine_consistency_test.go` | Cross-engine Pr | T2.1 |
| `shared/physics/montecarlo_validation_test.go` | Monte Carlo | T2.7 |
| `shared/physics/literature_validation_test.go` | Literature comparison | T1.1-T1.3 |
| `shared/physics/landau_materlik_test.go` | LGD coefficients | T1.12 |
| `shared/physics/fuzz_test.go` | Fuzz robustness | T3.9 |
| `shared/physics/fuzz_property_test.go` | Property-based fuzz | T3.9 |
| `shared/physics/solver_stress_test.go` | Solver stress | T1.9 |
| `shared/physics/landau_rk4_table_test.go` | RK4 table test | T1.8 |
| `shared/physics/property_test.go` | Property tests | T1.9 |
| `shared/physics/property_based_test.go` | Property-based tests | T1.9 |
| `shared/physics/edge_cases_test.go` | Edge cases | T1.9 |
| `shared/physics/boundary_value_test.go` | Boundary values | T1.9 |

**Integration tests:**

| File | Focus | Tier Coverage |
|------|-------|---------------|
| `cmd/fecim-lattice-tools/mode_lk_headless_ispp_5targets_test.go` | Headless ISPP 5 targets | T1.5, T1.9 |
| `cmd/fecim-lattice-tools/mode_lk_ispp_test.go` | LK ISPP rows/phases | T1.5 |
| `cmd/fecim-lattice-tools/mode_engine_matrix_test.go` | 9-material NaN check | T1.9, T1.13 |
| `cmd/fecim-lattice-tools/mode_lk_ispp_convergence_20targets_test.go` | 20-target stress | T1.5, T1.10 |
| `cmd/fecim-lattice-tools/mode_preisach_target_progression_test.go` | Target progression | T1.5 |

---

## Required Test Matrix

### Dimensions

| Dimension | Values | Count |
|-----------|--------|-------|
| **Engines** | Preisach, L-K | 2 |
| **Controllers** | Waveform (writer.go), L-K solver (ispp_write.go) | 2 |
| **Materials** | DefaultHZO, FeCIMMaterial, LiteratureSuperlattice, CryogenicHZO, HZOStandard32, HZOFJT140, HZOCustom14, PZT, AlScN | 9 |
| **Temperatures** | 77K, 233K (-40C), 298K (25C), 358K (85C), 398K (125C) | 5 |
| **Target Sets** | lo/mid/hi (3), randomized (10), stress (20) | 3 |
| **Simulation Modes** | Sine, Triangle, Square, ISPP | 4 |
| **Conductance Models** | Linear, Subthreshold, Saturation | 3 |

**Total combinations:** 2 x 2 x 9 x 5 x 3 x 4 x 3 = **6,480 test scenarios**

### Prioritization Strategy

**PR Gate (Fast CI Lane):** ~50 tests, <= 5 minutes
- 3 materials: DefaultHZO, FeCIMMaterial, LiteratureSuperlattice
- 2 engines: Preisach, L-K
- 1 controller each
- 1 temperature: 298K
- 1 target set: lo/mid/hi
- All 4 simulation modes
- 1 conductance model: Linear
- **Count:** 3 x 2 x 4 = 24 core tests + 26 invariant tests = **50 tests**

**Nightly Gate (Extended Lane):** ~500 tests, <= 2 hours
- All 9 materials
- 2 engines x 2 controllers = 4 engine/controller combos
- 3 temperatures: 233K, 298K, 398K
- 2 target sets: lo/mid/hi, randomized
- All 4 simulation modes
- All 3 conductance models
- **Count:** 9 x 4 x 3 x 2 x 4 x 3 = 2,592 combinations -> sample 500 via stratified selection

**Release Gate (Full Suite):** ~2,000 tests, <= 8 hours
- All dimensions, sampled via Latin Hypercube Sampling (LHS)

**Research Suite (On-Demand):** Full 6,480 tests + Monte Carlo
- Full factorial coverage + 1000-device Monte Carlo per material
- **Runtime:** ~48 hours on 16-core system

### Material Gate Policy

| Material | PR Gate | Nightly Gate |
|----------|---------|--------------|
| `default_hzo` | Yes (baseline) | Yes |
| `fecim_hzo` | Yes (primary) | Yes |
| `literature_superlattice` | Yes (high-level) | Yes |
| `fecim_hzo_target` | No | Yes |
| `cryogenic_hzo` | No | Yes |
| `hzo_standard_32` | No | Yes |
| `hzo_ftj_140` | No | Yes |
| `hzo_custom_14` | No | Yes |
| `alscn` | No | Yes |

### Headless Environment Variables To Sweep

- `FECIM_MATERIAL`
- `FECIM_RANGE_FRAC`
- `FECIM_ISPP_STEPS_PER_PULSE`
- `FECIM_HEADLESS_FAST`
- `FECIM_ISPP_TARGETS`
- `FECIM_ISPP_TARGET_SEED`
- `FECIM_ISPP_TARGET_LEVELS`
- `FECIM_ISPP_MAX_PULSES`
- `FECIM_HEADLESS_ALLOW_TIMEOUT`
- `FECIM_UPDATE_PHYSICS_GOLDEN` (for regenerating golden data)
- `FECIM_M1_EXTENDED` (enables extended nightly lane)
- `FECIM_M1_RELEASE_GATE` (enables release gate suite)
- `FECIM_M1_RESEARCH_SUITE` (enables full research suite)
- `FECIM_EMIT_ARTIFACTS` (enables CSV/figure artifact generation; always on for release/research suites)

---

## Research-Grade Test Additions

### Section A: Thermodynamic Validation Suite

**New tests:** 4 (Energy conservation, Entropy, Clausius-Clapeyron)
**Files:**
- `module1-hysteresis/pkg/ferroelectric/thermodynamics_test.go`
- `shared/physics/entropy_test.go`

**Golden data:**
- `validation/testdata/thermodynamics/energy_conservation_*.json`

---

### Section B: Retention and Time Physics Suite

**New tests:** 4 (Arrhenius, Temperature sweep, Imprint, Drift)
**Files:**
- `module1-hysteresis/pkg/ferroelectric/retention_test.go`
- `shared/physics/time_dependent_test.go`

**Golden data:**
- `validation/testdata/retention/arrhenius_fit_*.json`
- `validation/testdata/retention/extrapolation_10yr_85C.json`

---

### Section C: Preisach Advanced Properties Suite

**New tests:** 4 (Congruency, Wiping-out, Return-point, Everett positivity)
**Files:**
- `module1-hysteresis/pkg/ferroelectric/preisach_advanced_test.go`

**Golden data:**
- `validation/testdata/preisach/congruency_loop_overlap.json`
- `validation/testdata/preisach/minor_loop_recovery.json`

---

### Section D: Temperature Range Suite

**New tests:** 3 (Cryogenic, Automotive, Curie point)
**Files:**
- `module1-hysteresis/pkg/ferroelectric/temperature_sweep_test.go`
- `shared/physics/temperature_test.go` (extend)

**Golden data:**
- `validation/testdata/temperature/cryogenic_77K_*.json`
- `validation/testdata/temperature/automotive_range_*.json`

---

### Section E: Endurance and Cycling Suite

**New tests:** 5 (ISPP coupling, Wake-up, Fatigue, Pulse drift, AgingEngine)
**Files:**
- `module1-hysteresis/pkg/controller/endurance_test.go`
- `module1-hysteresis/pkg/ferroelectric/cycling_test.go`

**Golden data:**
- `validation/testdata/endurance/1e6_cycle_degradation.json`
- `validation/testdata/endurance/wakeup_signature.json`

---

### Section F: Monte Carlo Variability Suite

**New tests:** 4 (D2D, Spatial correlation, Wafer gradient, BER)
**Files:**
- `module1-hysteresis/pkg/montecarlo/d2d_test.go`
- `module1-hysteresis/pkg/montecarlo/spatial_test.go`

**Golden data:**
- `validation/testdata/montecarlo/1000dev_statistics.json`
- `validation/testdata/montecarlo/ber_vs_level.json`

---

### Section G: Noise Models Suite

**New tests:** 3 (RTN, 1/f, Telegraph)
**Files:**
- `module1-hysteresis/pkg/noise/rtn_test.go`
- `module1-hysteresis/pkg/noise/spectrum_test.go`

**Golden data:**
- `validation/testdata/noise/1f_spectrum_*.json`

---

### Section H: L-K Dynamics Suite

**New tests:** 5 (Merz law, NLS, Switching spectroscopy, Multi-level, Timestep convergence)
**Files:**
- `shared/physics/landau_dynamics_test.go`
- `shared/physics/switching_test.go`
- `shared/physics/timestep_convergence_test.go`

**Golden data:**
- `validation/testdata/lk/merz_law_fit.json`
- `validation/testdata/lk/kai_kinetics.json`
- `validation/testdata/lk/timestep_convergence.json`

---

## Golden Data Strategy

### Existing Golden Files (Keep + Extend)

- `validation/testdata/ispp_regression/preisach_wrd_ispp_regression.json`
- `validation/testdata/ispp_regression/lk_wrd_ispp_regression.json`
- `validation/testdata/physics_regression/preisach_loop_default_hzo.json`

### New Golden Files (Add)

**By Domain:**

| Domain | Files | Purpose |
|--------|-------|---------|
| Physics regression | 8 | P-E loops for 8 remaining materials |
| Thermodynamics | 12 | Energy/entropy per material, Clausius-Clapeyron fits |
| Retention | 18 | Arrhenius fits (9 materials x 2 engines) |
| Preisach | 36 | Congruency, wiping-out, return-point (9 materials x 4 tests) |
| Temperature | 45 | 5 temps x 9 materials |
| Endurance | 36 | 4 endurance tests x 9 materials |
| Monte Carlo | 36 | 4 MC tests x 9 materials |
| Noise | 27 | 3 noise tests x 9 materials |
| L-K Dynamics | 36 | 4 switching tests x 9 materials + convergence |

**Total new golden files:** ~254

**Update strategy:**
- Environment variable: `FECIM_UPDATE_PHYSICS_GOLDEN=1` regenerates all
- Per-domain: `FECIM_UPDATE_GOLDEN_THERMODYNAMICS=1`, etc.
- Version tagging: each golden file includes schema version

**Storage:**
- Compressed JSON (gzip): ~10-50 KB per file
- Total storage: ~254 files x 30 KB = **~7.6 MB** (acceptable for Git)

---

## Acceptance Profiles

### Profile A: Strict Target-Lock (Research Baseline)

- Used for Preisach and any model proven target-lock capable.
- Criteria:
  - level error `<= +/-1` level (cross-module physics criteria)
  - convergence required
  - pulse and overshoot budgets respected
  - current controller regression defaults:
    - pulses `<= 30` per target (Preisach regression suite)

### Profile B: Bounded Completion (Current LK Baseline)

- Used for current LK single-domain baseline while stabilization work is pending.
- Criteria:
  - must reach terminal done state
  - must satisfy pulse/overshoot/retry bounds
  - current controller regression defaults:
    - pulses `<= 80` per target
    - overshoots `<= 20` per target
  - must remain finite and unit-consistent
  - level error tracked and reported, not hidden

### Engine x Material Profile Map (Required)

| Material | Engine | Profile | Justification |
|----------|--------|---------|---------------|
| DefaultHZO | Preisach | A (strict) | Baseline reference, well-characterized |
| DefaultHZO | L-K | B (relaxed) | Single-domain L-K baseline |
| FeCIMMaterial | Preisach | A (strict) | Project primary, production use |
| FeCIMMaterial | L-K | B (relaxed) | Pending ensemble upgrade |
| LiteratureSuperlattice | Preisach | A (strict) | High-level capability validated |
| LiteratureSuperlattice | L-K | A (strict) | Ensemble mode enabled, stable |
| CryogenicHZO | Preisach | A (strict) | Extreme condition validation |
| CryogenicHZO | L-K | B (relaxed) | Temperature physics under development |
| HZOStandard32 | Preisach | A (strict) | Standard 32-level reference |
| HZOStandard32 | L-K | B (relaxed) | Pending validation |
| HZOFJT140 | Preisach | A (strict) | Ultra-high-level capability |
| HZOFJT140 | L-K | B (relaxed) | Requires ensemble for 140 levels |
| HZOCustom14 | Preisach | A (strict) | Low-level edge case |
| HZOCustom14 | L-K | B (relaxed) | Sharp switching challenge |
| PZT | Preisach | A (strict) | Classical ferroelectric validation |
| PZT | L-K | B (relaxed) | Low Ec requires careful tuning |
| AlScN | Preisach | A (strict) | High-field material |
| AlScN | L-K | B (relaxed) | Extreme parameter regime |

**Promotion criteria (B -> A for L-K):**
- 100 deterministic runs (different seeds) all meet strict criteria
- Level error <= +/-1 for >= 95% of runs
- Pulse count <= 30 for >= 90% of runs
- Documented evidence in `docs/development/evidence/`
- The active profile map must be versioned with artifacts for each release

---

## Quantitative Acceptance Criteria

### Domain-Specific Thresholds

| Domain | Metric | Threshold | Notes |
|--------|--------|-----------|-------|
| **Thermodynamics** | Energy error | <= 1% of integral(E dP) | Numerical integration tolerance |
| | Entropy increase | dS >= -1e-6 J/K | Negative allowed for numerical error only |
| **Retention** | Arrhenius fit | R^2 > 0.95 | Temperature-dependent lifetime |
| | Activation energy | 0.6-1.2 eV | HZO literature range |
| | 10-year retention | >= 90% Pr | Extrapolated at 85C |
| **Preisach** | Congruency RMS | <= 1% Ps | Loop-to-loop stability |
| | Wiping-out error | <= 1% Ps | Major loop recovery |
| | Return-point error | <= 0.5% Ps | Reversibility metric |
| | Everett positivity | 100% samples >= 0 | Non-negativity constraint |
| **Temperature** | Cryogenic Pr ratio | 1.5 +/- 0.3 | Pr(77K)/Pr(300K) |
| | Automotive monotonicity | 100% | dPr/dT < 0, dEc/dT < 0 |
| | Curie approach | Pr(0.95Tc) <= 10% Pr(300K) | Phase transition |
| **Endurance** | 10^6 cycle survival | >= 70% Pr | Fatigue resistance |
| | Wake-up ratio | >= 1.2 | Pr(10^4)/Pr(1) |
| | Fatigue exponent | 0.05-0.15 | Power-law beta |
| | ISPP pulse drift | <= 2x | Convergence stability |
| **Monte Carlo** | D2D std match | +/-20% of input sigma | Variability propagation |
| | Spatial correlation | >= 0.6 at lambda | Autocorrelation function |
| | BER (7-level) | <= 10^-3 | Adjacent state overlap |
| **Noise** | RTN exponential fit | R^2 > 0.9 | Dwell time distribution |
| | 1/f exponent | 0.8-1.2 | Power spectral density |
| | Telegraph Lorentzian | fc +/- 30% | Corner frequency |
| **L-K Dynamics** | Merz law fit | R^2 > 0.95 | tau proportional to exp(-1/E) |
| | NLS exponent | 0.8-2.2 | KAI dimensionality |
| | Switching speed | <= 360 ps at 5Ec | State-of-art capability |
| | Multi-level stability | >= 10 us at E=0 | Intermediate state lifetime |
| | Timestep convergence | Error ratio >= 14 | O(dt^4) for RK4 |
| **ISPP** | Level error (strict) | <= +/-1 level | Preisach, mature L-K |
| | Level error (relaxed) | Bounded completion | L-K baseline |
| | Pulse budget (strict) | <= 30 | Preisach profile |
| | Pulse budget (relaxed) | <= 80 | L-K baseline profile |
| | Overshoot budget | <= 20 | Both profiles |
| | Bounds integrity | 0 collapses | VMin <= VMax always |
| | Success rate | >= 95% | Research-grade |
| **Conductance** | Linear monotonicity | 100% | G(P_i+1) > G(P_i) |
| | Subthreshold slope | 1.0-1.5 | n factor |
| | Saturation fit | R^2 > 0.98 | Quadratic regime |
| | Level separation | >= 1% (Gmax-Gmin) | Resolvability |
| **Export** | CSV schema | 100% compliance | Versioned headers |
| | JSON metadata | 100% completeness | All required fields |
| | Clipboard roundtrip | <= 1e-12 error | Floating-point precision |
| | Downstream NaN/Inf | 0 occurrences | Crossbar compatibility |
| **Determinism** | Seed reproducibility | Byte-for-byte | Identical output |
| | Multi-run std | <= 10% of mean | Well-behaved materials |
| | Platform independence | <= 1e-9 error | Cross-platform |
| **Statistical** | KS test baseline | p > 0.05 | No significant regression |
| | RMSE vs Ps | < 5% | Loop shape stability |
| | Correlation | r > 0.95 | Shape preservation |
| | Chi-squared histogram | p > 0.05 | Pulses distribution stable |
| **ISPP Physics** | Monotonicity (P1) | Spearman rho = 1.0 | Exact rank order |
| | Ec-convergence corr (P2) | Pearson r > 0.5 | Moderate correlation |
| | Overshoot groups (P3) | Kruskal-Wallis p < 0.05 | Groups differ |
| | Minor loop closure (P4) | delta_P/Ps < 1% | Remanent stability |
| | Multi-level retention (P5) | Min gap > 0.5% range | Level distinguishability |
| **Preisach Formal** | Symmetry (PR1) | < 1e-6 | Mathematical identity |
| | Saturation (PR2) | < 1% | Calibration |
| | Ec accuracy (PR3) | < 1% | Calibration |
| | Pr accuracy (PR4) | < 5% | Calibration |
| | Return-point (PR5) | < 1e-10 | Mathematical identity |
| | Wipe-out (PR6) | Stack shortens | Structural property |
| | Congruent loops (PR7) | CV < 5% | Preisach property |
| | Everett >= 0 (EV1) | 100% grid | Mathematical identity |
| | Everett monotone (EV2) | 0 violations | Mathematical identity |
| | Everett boundary (EV3) | < 1% | Boundary condition |
| **L-K Formal** | Pr from G(P) (LK1) | < 5% of Pr | Free energy validation |
| | Ec from Landau (LK2) | < 5% | Polynomial validation |
| | Double-well (LK3) | 2 minima, barrier > 100kT | Structural property |
| | Energy per cycle (LK4) | < 1e-6 | Conservation |
| | Ensemble Pr (LK5) | < 5% | Averaging property |
| **NLS** | CDF shape (NLS-1) | Probit R^2 > 0.99 | Log-normal verification |
| | Field dependence (NLS-2) | E_act within 10% | Merz's law |
| | Sigma (NLS-3) | Config in 95% CI | Parameter consistency |
| **Transfer** | Boundary (R1) | < 1e-12 (exact) | Mathematical identity |
| | Monotonicity (R2) | 0 violations | Mathematical identity |
| | Spacing pattern (R3) | < 1% deviation | Model prediction |
| **Read Disturb** | Per-read drift (RD1) | < 1e-6 Ps | Sub-coercive stability |
| | Cumulative drift (RD1) | < 1 level spacing | Read endurance |
| **Temperature Scaling** | Ec scaling beta (T1) | 0.4-0.6 | Mean-field theory |
| | Pr scaling beta (T2) | 0.4-0.6 | Mean-field theory |
| | Operational limit (T3) | Reported | Characterization |
| **Cross-Engine** | Pr agreement (CE1) | < 15% | Engine consistency |
| | Ec agreement (CE2) | < 15% | Engine consistency |
| | Loop area (CE3) | < 20% | Engine consistency |
| | Level reachability (CE4) | >= 90% | Engine capability |
| **Material** | Parameter bounds (M1) | All pass | Sanity check |
| | Landau coefficients (M2) | All pass | Sanity check |
| | Depolarization (M3) | All pass | Sanity check |
| | NLS params (M4) | All pass | Sanity check |

### Cross-Module Alignment

From `docs/development/PHYSICS_ACCEPTANCE_CRITERIA.md`:
- Loop regression: RMS(E), RMS(P) <= 2% full-scale
- Material parameters: +/-10% engineering tolerance
- ISPP level-hit: +/-1 level (strict profile)

**Consistency check:** All Module 1 thresholds must be <= cross-module thresholds (stricter is OK).

---

## Pass/Fail Criteria (Research Grade)

### Hard Requirements (Must Pass for PR Merge)

1. **Zero NaN/Inf:** No non-finite values in any output (logs, CSV, JSON)
2. **Unit Consistency:** All physical quantities have correct units (V/m, C/m^2, S)
3. **Deterministic Reproduction:** Same seed -> identical output (byte-for-byte)
4. **Bounds Integrity:** No ISPP bounds collapses (VMin <= VMax)
5. **Material Coverage:** 100% of PR gate materials (3) pass
6. **Acceptance Profile:** Each material meets its assigned profile (A or B)
7. **Fast Lane Runtime:** <= 5 minutes on CI hardware
8. **Schema Compliance:** 100% CSV/JSON schema validation
9. **All Tier 1 tests pass**
10. **T0 Physics Identities:** All Tier 0 tests pass (PR1, PR5, EV1, EV2, EV3, P1, R2, seed determinism)
11. **Material Sanity:** All M1-M4 material validation checks pass

### Soft Requirements (Must Pass for Release)

12. **Extended Material Coverage:** All 9 materials pass nightly gate
13. **Statistical Stability:** Multi-run std <= 10% of mean
14. **No Regression:** KS test p > 0.05 vs baseline (using `validation.KolmogorovSmirnovTest`)
15. **Golden Data Match:** Reproducibility 100% (within numerical tolerance, RMSE < 0.01% Ps)
16. **Literature Alignment:** Key metrics within +/-5% of published values (Tier 1) or +/-20% (Tier 2)
17. **Downstream Compatibility:** Module 2 crossbar integration: zero NaN/Inf
18. **Timestep Convergence:** RK4 error ratio >= 14 confirmed
19. **Cross-Engine Agreement:** Tracked, trending toward < 10%
20. **Cross-Engine Consistency:** CE1-CE4 all pass at nightly tolerance
21. **Read Disturb:** RD1 cumulative drift < 1 level spacing

### Provenance Requirements (For Publication)

18. **Every regression artifact contains `source_metadata` block** with git hash and seed
19. **Every physics output tagged** with `confidence_tags` block via ConfidenceLedger
20. **No `placeholder` provenance** on Tier 1 output parameters
21. **Per-material verdict** present in every artifact
22. **Regression comparator** reports no statistically significant degradation

### Research Requirements (For Publication)

23. **Thermodynamic Consistency:** Energy error <= 1%
24. **Retention Validation:** Arrhenius R^2 > 0.95, Ea in [0.6, 1.2] eV
25. **Preisach Congruency:** Loop overlap <= 1% Ps
26. **Temperature Range:** Cryogenic + automotive validated, scaling exponents beta in [0.4, 0.6]
27. **Endurance:** 10^6 cycle survival >= 70% Pr
28. **Monte Carlo:** 1000 devices, D2D std within +/-20%
29. **Switching Dynamics:** Merz law R^2 > 0.95, NLS probit R^2 > 0.99
30. **Multi-Level:** >= 7 stable levels (research credibility)
31. **All CSV Artifacts:** Complete set generated (see Research Data Artifacts section)
32. **All Figures:** Publication-quality figures generated (Figures 1-6)
33. **L-K Solver Validation:** LK1-LK5 all pass
34. **Preisach Formal Properties:** PR1-PR7, EV1-EV3 all pass
35. **NLS Validation:** NLS-1 to NLS-3 all pass

---

## Implementation Phases

### Phase 1: Consolidate Existing Headless Coverage (Week 1)

**Goal:** Audit and strengthen current test files

**Tasks:**
1. Audit all existing test functions for:
   - Deterministic target order assertions
   - Explicit convergence window reporting
   - CSV schema validation for downstream analysis
2. Standardize around `cmd/fecim-lattice-tools/mode_*_test.go` patterns
3. Add missing assertions to `module1-hysteresis/pkg/controller/headless_regression_test.go`
4. Add source metadata block (git hash, seed, env vars) to all existing JSON emitters
5. Document all env vars

**Deliverables:**
- Updated test files with standardized assertions
- Environment variable reference doc
- Baseline coverage report

---

### Phase 2: Timestep Convergence, Golden Regression, and Artifact Schema (Week 2-3)

**Goal:** Close the two CRITICAL Tier 1 gaps + normalize JSON schema

**Tasks:**
1. Implement NT-1 (timestep convergence study) -- **Tier 1 critical**
2. Implement NT-2 (golden P-E regression for all 9 materials) -- **Tier 1 critical**
   - Generate 8 new golden data files
3. Implement NT-5 (PUND switching symmetry) -- **Tier 1 high**
4. Extend JSON schema with source_metadata, write_verify_stats, physics_validation, confidence_tags, timestep_convergence blocks
5. Implement CSV schema versioning and metadata completeness checks

**Deliverables:**
- `shared/physics/timestep_convergence_test.go`
- 8 new golden data files in `validation/testdata/physics_regression/`
- Extended PUND tests
- Schema validation tests

---

### Phase 3: Statistics Export and Baseline Comparator (Week 4)

**Goal:** Export WriteVerifyStats, detect silent degradation

**Tasks:**
1. Implement NT-4 (WriteVerifyStats JSON export)
2. Implement NT-3 (baseline comparator with KS test)
3. Implement NT-8 (loop area energy density)
4. Add automatic comparison against previous baseline

**Deliverables:**
- WriteVerifyStats JSON export in headless regression
- `shared/physics/baseline_comparator_test.go`
- `shared/physics/loop_area_test.go`
- `scripts/compare_regression_baseline.sh`

---

### Phase 4: Cross-Engine and Multi-Material Deep Regression (Week 5-6)

**Goal:** Close remaining Tier 1 and Tier 2 gaps

**Tasks:**
1. Implement NT-9 (9-material deep regression suite) -- **Tier 1 critical**
2. Implement NT-6 (cross-engine Pr agreement tracking) -- **Tier 2 high**
3. Implement NT-7 (FORC density positivity) -- **Tier 2**
4. Implement dual acceptance profiles (Profile A and Profile B)
5. Add per-material verdict reporting

**Deliverables:**
- Extended `mode_engine_matrix_test.go` with physics validation
- Extended `cross_engine_consistency_test.go`
- Extended `worldclass_forc_test.go`
- `module1-hysteresis/pkg/testing/acceptance_profiles.go`

---

### Phase 5: New Research-Grade Tests (Week 7-11)

**Goal:** Add Sections A-D + Preisach Model Validation (PR1-PR7, EV1-EV3) + ISPP Physics Properties (P1-P5) + Conductance Transfer Validation (R1-R3) + Material Validation (M1-M4) + Temperature Dependence (T1-T3)

**Tasks:**
1. **Week 7:** Thermodynamic consistency suite (Section A) - 4 new tests
2. **Week 8:** Retention physics suite (Section B) - 4 new tests + Preisach Model Validation (PR1-PR7, EV1-EV3)
3. **Week 9:** Preisach advanced properties suite (Section C) - 4 new tests + Material Validation (M1-M4)
4. **Week 10:** Temperature range suite (Section D) - 3 new tests + Temperature Dependence Validation (T1-T3)
5. **Week 11:** ISPP Physics Properties (P1-P5) + Conductance Transfer Validation (R1-R3) + Integration across all materials

**Deliverables:**
- 17 new test files (~220 test functions)
- ~111 golden data files
- Research validation report

---

### Phase 6: Extended Research Features (Week 12-16)

**Goal:** Add Sections E, G, H + L-K Solver Validation (LK1-LK5) + NLS Validation (NLS-1 to NLS-3) + Read Disturb (RD1) + Cross-Engine Consistency (CE1-CE4)

**Tasks:**
1. **Week 12-13:** Endurance and cycling suite (Section E) - 5 new tests
2. **Week 14:** L-K dynamics suite (Section H) - 5 new tests + L-K Solver Validation (LK1-LK5) + NLS Validation (NLS-1 to NLS-3)
3. **Week 15:** Noise models suite (Section G) - 3 new tests + Read Disturb (RD1) + Cross-Engine Consistency (CE1-CE4)
4. **Week 16:** Integration, literature comparison, publication-ready figures

**Deliverables:**
- 12 new test files (~180 test functions)
- ~99 golden data files
- Literature comparison report

---

### Phase 7: Advanced Research Features and CI (Week 17-20)

**Goal:** Monte Carlo variability + full CI integration + publication figures + CSV artifacts

**Tasks:**
1. **Week 17-18:** D2D variability, spatial correlation (Section F) - 4 new tests
2. **Week 19:** CI integration (fast lane, nightly, release gate) + Statistical Validation Framework
3. **Week 20:** Publication preparation: generate all Figures 1-6, all CSV artifacts, trend report, supplementary data archive

**Deliverables:**
- 2 new test files (~60 test functions)
- ~36 golden data files
- `.github/workflows/module1-physics.yml`
- Publication draft (methods section)
- All CSV artifacts (72 files, see Research Data Artifacts section)
- All publication-quality figures (Figures 1-6)

---

## CI Integration Details

### Fast Lane (PR Gate)

**Trigger:** Every PR to main
**Runtime:** <= 5 minutes
**Tests:** ~50 (Tier 1 only, PR gate materials)

```bash
env -u DISPLAY -u WAYLAND_DISPLAY \
  FECIM_HEADLESS_FAST=1 \
  go test -v -timeout 5m -count=1 -shuffle=off -trimpath \
  -run 'Headless|ISPP|Physics|Golden|Timestep' \
  ./cmd/fecim-lattice-tools/... \
  ./module1-hysteresis/pkg/controller/... \
  ./module1-hysteresis/pkg/ferroelectric/... \
  ./shared/physics/...
```

### Extended Lane (Nightly)

**Trigger:** Nightly cron (2 AM UTC) + release tags
**Runtime:** <= 2 hours
**Tests:** ~500 (all tiers, all 9 materials)

```bash
env -u DISPLAY -u WAYLAND_DISPLAY \
  FECIM_M1_EXTENDED=1 \
  go test -v -timeout 2h -count=1 -shuffle=off -trimpath \
  ./cmd/fecim-lattice-tools/... \
  ./module1-hysteresis/... \
  ./shared/physics/... \
  ./validation/...
```

### Release Gate (Pre-Release)

**Trigger:** Manual, before tagging release
**Runtime:** <= 8 hours
**Tests:** ~2,000 (Latin Hypercube Sampling)

```bash
env -u DISPLAY -u WAYLAND_DISPLAY \
  FECIM_M1_RELEASE_GATE=1 \
  FECIM_EMIT_ARTIFACTS=1 \
  go test -v -timeout 8h -count=1 -shuffle=off -trimpath \
  ./...
```

### Research Suite (On-Demand)

**Trigger:** Manual, for publication preparation
**Runtime:** ~48 hours (16-core system)
**Tests:** Full 6,480 factorial + MC

```bash
env -u DISPLAY -u WAYLAND_DISPLAY \
  FECIM_M1_RESEARCH_SUITE=1 \
  FECIM_EMIT_ARTIFACTS=1 \
  go test -v -timeout 48h -count=1 -parallel 16 -trimpath \
  ./...
```

### Additional Runner Commands

```bash
# Controller regression only
env -u DISPLAY -u WAYLAND_DISPLAY \
  go test -run HeadlessRegression_WRD_ISPP -count=1 \
  ./module1-hysteresis/pkg/controller

# Golden data regeneration
env -u DISPLAY -u WAYLAND_DISPLAY \
  FECIM_UPDATE_PHYSICS_GOLDEN=1 \
  go test -run TestGoldenRegression -count=1 \
  ./module1-hysteresis/pkg/ferroelectric/...

# Timestep convergence study
env -u DISPLAY -u WAYLAND_DISPLAY \
  go test -run TestTimestepConvergence -count=1 -timeout 10m \
  ./shared/physics/...

# Material-specific run
env -u DISPLAY -u WAYLAND_DISPLAY \
  FECIM_MATERIAL=literature_superlattice \
  go test -v -timeout 10m -count=1 \
  ./cmd/fecim-lattice-tools -run ISPP

# Regression comparison
./scripts/run_headless_ispp_regressions.sh
./scripts/compare_regression_baseline.sh \
  --current output/regression/module1/latest.json \
  --baseline validation/testdata/ispp_regression/preisach_wrd_ispp_regression.json \
  --threshold 0.02
```

---

## Traceability Matrix

Maps each physics claim to validation tests and acceptance criteria.

| Physics Claim | Test(s) | Acceptance Criteria | Tier | Golden Data |
|---------------|---------|---------------------|------|-------------|
| **P-E loop Pr accuracy** | T1.1 | Relative error <= 5% vs literature | 1 | `physics_regression/*.json` |
| **P-E loop Ps accuracy** | T1.2 | Relative error <= 5% vs literature | 1 | `physics_regression/*.json` |
| **P-E loop Ec accuracy** | T1.3 | Relative error <= 5% vs literature | 1 | `physics_regression/*.json` |
| **30-level quantization** | T1.4 | Spacing uniformity <= 5% deviation | 1 | (runtime check) |
| **ISPP convergence** | T1.5 | Success rate >= 95% | 1 | `ispp_regression/*.json` |
| **PUND symmetry** | T1.6 | \|SP-SN\|/max < 10% | 1 | (runtime check) |
| **Golden P-E regression** | T1.7 | RMSE <= 0.01% Ps (6 sig figs) | 1 | `physics_regression/*.json` |
| **RK4 timestep convergence** | T1.8 | Error ratio >= 14 | 1 | `lk/timestep_convergence.json` |
| **Zero NaN/Inf** | T1.9 | 100% finite across 18 combos | 1 | (runtime check) |
| **Deterministic reproducibility** | T1.10 | Bitwise identical | 1 | (runtime check) |
| **Unit consistency** | T1.11 | V/m <-> MV/cm correct | 1 | (runtime check) |
| **LGD coefficients** | T1.12 | Exact match Materlik 2015 | 1 | (runtime check) |
| **9-material deep regression** | T1.13 | All 18 combos within tolerance | 1 | `physics_regression/*.json` |
| **Energy conservation** | T1.14 | Error <= 1% integral(E dP) | 1 | `thermodynamics/*.json` |
| **Cross-engine Pr** | T2.1 | Within 10% for same material | 2 | (runtime check) |
| **FORC density positivity** | T2.2 | rho >= 0 in valid region | 2 | (runtime check) |
| **Retention self-consistency** | T2.3 | P(t) monotonically decreasing | 2 | `retention/*.json` |
| **Wake-up/fatigue** | T2.4, T2.5 | Correct monotonicity per regime | 2 | `endurance/*.json` |
| **Minor loop non-zero** | T2.6 | Sub-coercive dP > 0 (product form) | 2 | `preisach/*.json` |
| **Uncertainty propagation** | T2.7 | CI covers literature values | 2 | (runtime check) |
| **WriteVerifyStats export** | T2.8 | All fields in JSON | 2 | (schema check) |
| **ISPP voltage monotonicity** | T2.9 | No direction flips | 2 | (runtime check) |
| **Baseline comparator** | T2.10 | KS p > 0.05, RMSE < 5% Ps | 2 | `physics_regression/*.json` |
| **Arrhenius retention** | T2.13 | R^2 > 0.95, Ea in [0.6,1.2] eV | 2 | `retention/arrhenius_*.json` |
| **Preisach congruency** | T2.14 | RMS <= 1% Ps | 2 | `preisach/congruency_*.json` |
| **Return-point memory** | T2.15 | Error <= 0.5% Ps | 2 | `preisach/minor_loop_*.json` |
| **Merz law switching** | T3.12 | R^2 > 0.95 | 3 | `lk/merz_law_*.json` |
| **Sub-ns switching** | T3.12 | <= 360 ps at 5Ec | 3 | `lk/switching_*.json` |
| **Multi-level capability** | T3.14 | 8 levels, lifetime >= 10 us | 3 | `lk/multilevel_*.json` |
| **10^6 cycle endurance** | T3.8 | Pr(10^6) >= 70% Pr(10^3) | 3 | `endurance/1e6_*.json` |
| **Cryogenic enhancement** | T3.11 | Pr(77K)/Pr(300K) = 1.5 +/- 0.3 | 3 | `temperature/cryo_*.json` |
| **D2D variability** | T3.4 | sigma_obs within +/-20% sigma_in | 3 | `montecarlo/1000dev_*.json` |
| **7+ level MLC** | T3.17 | 7 resolvable levels | 3 | (runtime check) |
| **BER** | T3.16 | <= 10^-3 for adjacent levels | 3 | `montecarlo/ber_*.json` |
| **1/f noise** | T3.15 | Exponent in [0.8,1.2] | 3 | `noise/1f_*.json` |
| **Linear G(P)** | T3.17, R2 | 100% monotonic, 0 violations | 3, T0 | (runtime check) |
| **Deterministic platform** | T1.10 | Error <= 1e-9 cross-platform | 1 | (CI matrix) |
| **P-E saturation symmetry** | PR1 | max\|P(E)+P(-E)\|/Ps < 1e-6 | T0 | `pe_loop_*.csv` |
| **Preisach Ec accuracy** | PR3 | < 1% relative error | T1 | `pe_loop_*.csv` |
| **Preisach Pr accuracy** | PR4 | < 5% relative error | T1 | `pe_loop_*.csv` |
| **Preisach wipe-out** | PR6 | Stack shortens | T0 | (runtime check) |
| **Congruent minor loops** | PR7 | CV < 5% | T1 | (runtime check) |
| **Everett non-negativity** | EV1 | 100% grid >= 0 | T0 | (runtime check) |
| **Everett monotonicity** | EV2 | 0 violations | T0 | (runtime check) |
| **Everett boundary** | EV3 | diagonal=0, full-range=Ps | T0 | (runtime check) |
| **L-K free energy double-well** | LK3 | 2 minima, barrier > 100kT | T1 | (runtime check) |
| **L-K Pr from free energy** | LK1 | < 5% of Pr | T1 | (runtime check) |
| **L-K Ec from Landau** | LK2 | < 5% of Ec | T1 | (runtime check) |
| **L-K energy conservation** | LK4 | < 1e-6 drift per cycle | T1 | (runtime check) |
| **L-K ensemble averaging** | LK5 | mean Pr within 5% | T1 | (runtime check) |
| **NLS log-normal CDF** | NLS-1 | Probit R^2 > 0.99 | T1 | `nls_switching_*.csv` |
| **NLS field dependence** | NLS-2 | E_act within 10% | T1 | `nls_switching_*.csv` |
| **NLS sigma consistency** | NLS-3 | Config in 95% CI | T1 | `nls_switching_*.csv` |
| **Read disturb stability** | RD1 | < 1 level spacing / 10K reads | T1 | `read_disturb_*.csv` |
| **Transfer function boundaries** | R1 | Exact at +/-Ps | T0 | (runtime check) |
| **Transfer spacing pattern** | R3 | < 1% deviation | T1 | (runtime check) |
| **ISPP monotonicity** | P1 | Spearman rho = 1.0 | T0 | `ispp_sweep_*.csv` |
| **ISPP Ec-convergence corr** | P2 | Pearson r > 0.5 | T1 | `ispp_sweep_*.csv` |
| **ISPP minor loop closure** | P4 | delta_P/Ps < 1% | T1 | (runtime check) |
| **ISPP multi-level retention** | P5 | min gap > 0.5% range | T1 | `ispp_sweep_*.csv` |
| **Temperature Ec scaling** | T1 | beta in [0.4, 0.6] | T2 | `temp_dependence_*.csv` |
| **Temperature Pr scaling** | T2 | beta in [0.4, 0.6] | T2 | `temp_dependence_*.csv` |
| **Cross-engine Pr** | CE1 | < 15% | T2 | `cross_engine_*.csv` |
| **Cross-engine Ec** | CE2 | < 15% | T2 | `cross_engine_*.csv` |
| **Cross-engine loop area** | CE3 | < 20% | T2 | `cross_engine_*.csv` |
| **Cross-engine level reach** | CE4 | >= 90% | T2 | `cross_engine_*.csv` |
| **Material parameter sanity** | M1-M4 | All pass | T1 | `material_params_validated.csv` |

---

## Deliverables

### Code Deliverables

- [ ] NT-1: Timestep convergence test (`shared/physics/timestep_convergence_test.go`)
- [ ] NT-2: 8 new golden data files + extended golden regression test
- [ ] NT-3: Baseline comparator test (`shared/physics/baseline_comparator_test.go`)
- [ ] NT-4: WriteVerifyStats JSON export in headless regression
- [ ] NT-5: PUND switching symmetry assertions
- [ ] NT-6: Cross-engine Pr tracking with 10% target
- [ ] NT-7: FORC density positivity assertions
- [ ] NT-8: Loop area energy density test (`shared/physics/loop_area_test.go`)
- [ ] NT-9: 9-material deep regression with physics_validation block
- [ ] Extended JSON schema (source_metadata, write_verify_stats, physics_validation, confidence_tags, timestep_convergence blocks)

### Documentation Deliverables

- [ ] `docs/testing/TEST_GUIDE.md` - Module 1 regression workflow
- [ ] `docs/development/PHYSICS_ACCEPTANCE_CRITERIA.md` - Updated thresholds, per-material tolerance table
- [ ] `docs/development/TESTING.md` - New test execution commands
- [ ] Tier classification rationale document

### CI Deliverables

- [ ] Fast CI lane configuration (< 5 min, PR gate)
- [ ] Extended nightly lane configuration (< 2 hr, all materials)
- [ ] Release gate configuration (< 8 hr, LHS sampling)
- [ ] Artifact archival pipeline for `output/regression/module1/`

### New Physics Validation Deliverables

- [ ] ISPP Convergence Physics Properties (P1-P5): `module1-hysteresis/pkg/controller/ispp_physics_properties_test.go`
- [ ] Preisach Model Validation (PR1-PR7, EV1-EV3): `module1-hysteresis/pkg/ferroelectric/preisach_validation_test.go`
- [ ] Landau-Khalatnikov Solver Validation (LK1-LK5): `shared/physics/landau_validation_test.go`
- [ ] NLS Stochastic Switching Validation (NLS-1 to NLS-3): `module1-hysteresis/pkg/ferroelectric/nls_validation_test.go`
- [ ] Conductance Transfer Function Validation (R1-R3): `shared/physics/transfer_validation_test.go`
- [ ] Read Disturb Quantification (RD1): `module1-hysteresis/pkg/controller/read_disturb_test.go`
- [ ] Temperature Dependence Validation (T1-T3): `shared/physics/temperature_validation_test.go`
- [ ] Cross-Engine Consistency (CE1-CE4): `module1-hysteresis/pkg/ferroelectric/cross_engine_test.go`
- [ ] Material Validation Suite (M1-M4): `shared/physics/material_validation_test.go`
- [ ] Statistical Validation Framework implementation
- [ ] Research Data Artifacts (8 CSV artifact types x 9 materials)
- [ ] Publication-Quality Figures (Figures 1-6)

### Research Artifact Deliverables

- [ ] Research report artifact containing:
  - Profile verdicts per engine x material
  - Trend metrics across releases
  - Uncertainty summary (Monte Carlo CI)
  - Timestep convergence evidence
  - Provenance coverage report (ConfidenceLedger summary)
  - Cross-engine agreement trend
  - CSV artifact summary (72 files)
  - Publication figure index (6 figure types)

---

## Completion Tracking

| Phase | Description | Duration | Status | Blocking |
|-------|-------------|----------|--------|----------|
| Phase 1 | Consolidate existing coverage | 1 week | Not started | -- |
| Phase 2 | Timestep convergence + golden regression + schema | 2 weeks | Not started | Phase 1 |
| Phase 3 | Statistics export + baseline comparator | 1 week | Not started | Phase 1 |
| Phase 4 | Cross-engine + 9-material deep regression | 2 weeks | Not started | Phase 2, 3 |
| Phase 5 | Core research tests (Sections A-D + PR1-PR7, EV1-EV3, P1-P5, R1-R3, M1-M4, T1-T3) | 5 weeks | Not started | Phase 4 |
| Phase 6 | Extended research (Sections E, G, H + LK1-LK5, NLS-1-3, RD1, CE1-CE4) | 5 weeks | Not started | Phase 5 |
| Phase 7 | Advanced research + CI (Section F + Figures + CSV artifacts) | 4 weeks | Not started | Phase 6 |

**Total:** 20 weeks (5 months)

**Post-Phase 7 Status:**
- **Test count:** ~277 (existing) + ~460 (new) = **~737 total tests**
- **Golden data:** ~11 (existing) + ~254 (new) = **~265 total files**
- **CSV artifacts:** 8 artifact types x 9 materials = **72 research data files**
- **Figures:** 6 publication-quality figure types
- **Coverage:** 50+ physics claims x 9 materials = **450+ claim-material pairs validated**
- **New test IDs:** P1-P5, PR1-PR7, EV1-EV3, LK1-LK5, NLS-1 to NLS-3, R1-R3, RD1, T1-T3, CE1-CE4, M1-M4
- **Publication readiness:** Methods section complete, supplementary data archived

---

## Appendix A: LK Solver Numerical Guard Mechanisms (14 total)

The LK solver contains 14 numerical guard mechanisms that prevent divergence. All must be preserved and tested:

1. Timestep adaptive clipping
2. Field magnitude clamping
3. Polarization saturation clamping
4. dP/dt rate limiting
5. Energy barrier checking
6. NaN/Inf detection and recovery
7. Oscillation detection (sign change counting)
8. Convergence stall detection
9. Maximum iteration limits
10. Minimum timestep enforcement
11. RK4 intermediate stage validation
12. Post-step energy consistency check
13. Bounds bracket collapse recovery
14. Overshoot direction flip prevention

Each guard mechanism should have at least one test that triggers it and verifies correct behavior. Existing coverage in `shared/physics/solver_stress_test.go`, `shared/physics/edge_cases_test.go`, and `shared/physics/boundary_value_test.go`.

## Appendix B: Preisach TanhEverett Product Form

The product-form Everett function (replacing the old factorized-difference form) is:

```
W(alpha, beta) = [1 + tanh((alpha - Ec) / Delta)] * [1 - tanh((beta + Ec) / Delta)] * Ps / 4
```

This form is the mathematically correct integral of the sech^2 Preisach density and is always non-negative. Tests must verify:
- Non-negativity for all (alpha, beta) pairs
- Major loop shape matches tanh model (same Pr/Ps ratio)
- Minor loops produce non-zero P changes (unlike old clamped form)
- Product form golden data matches regression baselines

Existing coverage in `module1-hysteresis/pkg/ferroelectric/preisach_equation_test.go` and `preisach_test.go`. Golden regression data in `validation/testdata/physics_regression/preisach_loop_default_hzo.json`.

## Appendix C: Key Bug Patterns (Regression Guard)

These historically significant bugs must have dedicated regression tests to prevent recurrence:

1. **Guard Sign Direction Flip**: Guard-band logic overriding LastError=0 to +/-1, causing direction flip and catastrophic overshoot. Guard pulses limited to 2 max, calcLevel clamped.

2. **Bounds Collapse**: Binary search [VMin, VMax] collapsing after overshoot recovery. Widened collapsed bracket minimally using direction info.

3. **ACCEPT +/-1 Guard Interaction**: ACCEPT +/-1 firing prematurely when guardActive=true (actual error is 0). Skipped when guard active, threshold raised from 3 to 8 overshoots.

4. **Zero-Field Bounds Reset**: CurrentField=0 during verify after reset shortcut causing VMax=0 collapse. Reset to full [0, MaxField] when absField < 0.01*Ec.

5. **Overshoot Limit as Physics-Limited Convergence**: Sharp-switching materials (fecim_hzo, hzo_custom_14) reaching overshoot limit. OvershootLimit (30) triggers StateSuccess not StateFailed.

6. **Preisach Everett Zero-Clamp Teleportation**: Old factorized-difference form going negative for minor loops, clamped to 0, making sub-coercive ISPP invisible. Fixed by product form.

---

**End of Research-Grade Automated Testing Plan for Module 1**
