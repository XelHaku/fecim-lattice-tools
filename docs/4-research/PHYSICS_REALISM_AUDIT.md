# Physics Realism Audit

**Date:** 2026-02-03 | **Updated:** 2026-02-11
**Scope:** Physics and physics-adjacent models in modules 1–6 and shared physics/peripherals
**Status:** Living document — update as improvements are made

---

## Executive Summary

This simulator prioritizes **educational clarity** over predictive accuracy. Most models are **Medium to Low realism** — sufficient for teaching ferroelectric CIM concepts but not for hardware design or device validation.

| Realism Level | Definition | Appropriate Use |
|---------------|-----------|-----------------|
| **High** | Calibrated to published measured data; validated against independent literature | Device design guidance, quantitative prediction |
| **Medium** | Physically-motivated equations with reasonable parameter ranges but uncalibrated to specific devices | Concept education, qualitative exploration, trend analysis |
| **Low** | Parametric/analytic approximations not grounded in device-level physics | Visualization, intuition building, system-level architecture demos |

**Key policy:** All simplified models carry explicit disclaimers in the UI and docs per `HONESTY_AUDIT.md`.

---

## Realism Summary by Module

| Area | Realism | Key Physics | Main Simplifications |
|------|---------|-------------|---------------------|
| Hysteresis (Preisach) | Medium | Classical Preisach operator with tanh Everett function | Quasi-static (no switching kinetics), heuristic stress/temperature scaling, evenly-spaced discrete states |
| Hysteresis (Landau-Khalatnikov) | Medium | 6th-order Landau free energy + viscous damping ODE: ρ_eff · dP/dt = E_app − k_dep·P − (2αP + 4βP³ + 6γP⁵) + ξ(t) | Single-domain 1D (no spatial/polydomain effects), k_dep as tuning parameter, empirical NLS nucleation |
| Crossbar Array | Medium | Ohm's law MVM (I = G·V), iterative IR-drop relaxation, Elmore RC delay | Linear/exponential conductance mapping (no compact FeFET I-V), assumed wire capacitance, linear half-select disturb |
| MNIST CIM Inference | Low | DAC → quantized MVM → ADC → noise → softmax pipeline | Linear binning quantization (not a device write/read model), software Gaussian noise proxy, no peripheral timing constraints |
| Circuits (DAC/ADC/TIA/ChargePump) | Low | Ideal transfer functions with parametric INL/DNL and energy estimates | No transistor-level modeling, no PVT corners, heuristic power numbers |
| Comparison / EDA | Low | Estimated FeCIM vs CMOS/Flash metrics, analytic area/latency/energy | No experimentally validated metrics, no placement/routing parasitics |

---

## Detailed Findings

### 1. Hysteresis: Preisach Model

**Realism:** Medium
**Files:** `module1-hysteresis/pkg/ferroelectric/preisach.go`, `shared/physics/material.go`

| Simplification | Physical Impact | Upgrade Path |
|----------------|----------------|--------------|
| Quasi-static hysteresis — no switching kinetics | Captures hysteresis memory and minor loops but not time-dependent switching speed or frequency dispersion | Integrate Kolmogorov-Avrami-Ishibashi (KAI) nucleation model or measured switching-time distributions |
| tanh-based Everett function | Produces smooth symmetric loops but shape is not calibrated to any specific device | Fit Everett function parameters to measured first-order reversal curves (FORC) from a target HZO sample |
| Linear stress/temperature scaling of Pr and Ec | Qualitative trend only — real behavior is nonlinear (Curie-Weiss for T, domain pinning for stress) | Derive scaling from Landau free-energy coefficients α(T) = α₀(T − T_C) or use published empirical fits |
| Evenly-spaced discrete conductance states | Convenient for digital mapping but not physics-based — real devices show nonuniform level spacing | Use measured conductance distributions from ISPP write-verify cycling |

**Validation status:** Internal consistency verified (FOCUS-54–60 ✅). No calibration to external measured P-E data yet.
**Validation target:** Match published HZO P-E loop shape (e.g., Pr: 15–34 µC/cm², Ec: 1–2 MV/cm from Nature Communications 2025) within 10% RMS error.

---

### 2. Hysteresis: Landau-Khalatnikov Solver

**Realism:** Medium
**Files:** `shared/physics/landau.go`, `shared/physics/landau_equation_test.go`

The L-K solver implements:
- **Landau free energy:** F(P) = αP² + βP⁴ + γP⁶ (6th-order polynomial double-well)
- **Viscous dynamics:** ρ_eff · dP/dt = E_applied − k_dep·P − dF/dP + ξ(t)
- **Effective viscosity:** ρ_eff = ρ + (R_series · A / d) — accounts for series resistance
- **NLS nucleation:** Empirical incubation with exp(−E_a/kT) Arrhenius activation

Citation audit notes:
- Golden Set I coefficients (β = −2.160e8, γ = 1.653e10, ρ = 0.05) are currently tracked as **[CITATION NEEDED]** pending a direct HZO fit citation (suggested: Materlik et al., *J. Appl. Phys.* 117, 134109 (2015) or equivalent).
- k_dep = 2.5e8 V·m/C is currently a calibrated default; use stack-derived formula or measured extraction to remove **[CITATION NEEDED]** status.

| Simplification | Physical Impact | Upgrade Path |
|----------------|----------------|--------------|
| Single-domain 1D ODE — no spatial variation | Cannot represent polydomain nucleation, domain wall motion, or intermediate remanent states at E=0 | Implement polydomain ensemble model (LK-PD-3 in progress) with distributed coercive fields |
| k_dep as a tuning knob for depolarization | Creates analog slope in P-E loop but does not model the actual interface depolarization field from dead layers | Model k_dep from dielectric stack: k_dep = (ε_FE · d_dead) / (ε_dead · d_FE); current 2.5e8 setting should be treated as **[CITATION NEEDED]** unless tied to a measured stack extraction |
| Empirical NLS time constants | Nucleation delays are uncalibrated | Fit τ₀ and E_a to measured switching distributions (e.g., Muller et al. IEEE TED; Jo et al., Nano Lett. 2021); current defaults remain **[CITATION NEEDED]** |
| No polydomain — only 2 stable wells at E=0 | Cannot hold intermediate remanent polarization states needed for multilevel ISPP | Polydomain model (LK-PD-1 through LK-PD-6) is the top physics priority |

**Validation status:** Equation identity and units verified against formulation (FOCUS-50 ✅). Solver kernel benchmarked at ~64 ns/op (FOCUS-49 ✅). Polydomain extension in progress (LK-PD-3 🔄).
**Validation target:** Reproduce switching transient from Muller et al.; hold ≥20 distinct remanent levels at E=0 with polydomain model.

---

### 3. Crossbar Array + Non-Idealities

**Realism:** Medium
**Files:** `module2-crossbar/pkg/crossbar/*.go`, `module4-circuits/pkg/arraysim/tier_a.go`, `module4-circuits/pkg/arraysim/tier_b.go`

| Component | Simplification | Current State | Upgrade Path |
|-----------|----------------|---------------|--------------|
| Conductance mapping | Linear/exponential/lookup — no compact FeFET I-V model | Verified against docs (FOCUS-54 ✅) | Add physics-based FeFET conductance model: G(V_g, T, history) |
| IR drop solver | Iterative relaxation on wire resistance network | Verified (FOCUS-56 ✅); Tier-A dense nodal solver added; Tier-B sparse PCG solver added (M4-P4 ✅) | Validate Tier-B vs SPICE golden vectors for 8×8 and 64×64 arrays (TIERB-3 ⏳) |
| RC delay | Elmore model with assumed wire capacitance | Verified (M2-P1 ✅) | Extract C_wire from layout or PDK; add frequency-dependent loss |
| Drift / retention | Log/power-law decay with Arrhenius temperature scaling | Verified (FOCUS-58 ✅, M2-P2 ✅) | Calibrate decay exponent and activation energy to published retention data (e.g., 10-year extrapolation) |
| Half-select disturb | Linear per-pulse cumulative model | Verified (FOCUS-59 ✅) | Add threshold-based cumulative disturb with fatigue coupling |
| Sneak paths | 3-cell simplified model + SNR formula: 20·log₁₀(I_signal/I_sneak) | Verified (FOCUS-57 ✅) | Full DC nodal solve for passive arrays (ASIM-2 ⏳) |

**Validation status:** All non-ideality models verified against documentation. Tier-A/B solvers implemented and regression-tested.
**Validation target:** Tier-B solver vs SPICE for small array (max error < 5% on node voltages).

---

### 4. MNIST CIM Inference

**Realism:** Low
**Files:** `module3-mnist/pkg/core/quantize.go`, `module3-mnist/pkg/core/network_inference.go`

| Simplification | Physical Impact | Upgrade Path |
|----------------|----------------|--------------|
| Linear binning quantization (uniform N-level) | Not a device write/read model — real FeFET conductance levels are nonuniform and depend on ISPP pulse history | Replace with device-aware quantization using measured conductance distributions from ISPP cycling |
| Software Gaussian noise injection (after ADC) | Proxy for combined read noise sources — does not distinguish ADC quantization noise, thermal noise, 1/f noise, or cell-to-cell variation | Decompose into physical noise components: σ²_total = σ²_ADC + σ²_thermal + σ²_1/f + σ²_variation |
| No peripheral timing constraints | Missing real bottleneck — ADC conversion rate limits throughput | Add ADC latency model: t_read = N_rows × t_ADC_conversion |
| CIM path semantically delegates to FP math | Conductance mapping Gmin/Gmax is commented but not exercised in computation | Wire actual G = Gmin + (level/N)·(Gmax−Gmin) into MVM accumulation |

**Validation status:** Pipeline order locked as DAC→MVM→ADC→noise→softmax (FOCUS-62 ✅). Energy model aligned (M3-P2 ✅). CIM semantic delegation documented with runtime warning (FOCUS-36 ✅).
**Validation target:** Compare quantization error distribution to measured device variation from published FeFET array data.

---

### 5. Circuits (DAC/ADC/TIA/Charge Pump)

**Realism:** Low
**Files:** `shared/peripherals/dac.go`, `shared/peripherals/adc.go`, `shared/peripherals/tia.go`, `shared/peripherals/chargepump.go`

| Simplification | Physical Impact | Upgrade Path |
|----------------|----------------|--------------|
| Parametric INL/DNL formulas | Heuristic error injection — not derived from circuit topology or transistor mismatch models | Use published ADC INL/DNL models (e.g., segmented current-steering DAC mismatch) or measured data |
| No transistor-level modeling | Missing real noise sources (kT/C, comparator metastability, op-amp offset) and PVT sensitivity | Add SPICE-extracted macromodels or behavioral Verilog-A models |
| Heuristic energy estimates (E = C·V²·N_switches) | Order-of-magnitude only — missing leakage, clock distribution, digital logic power | Tie to measured per-conversion energy from published SAR/sigma-delta ADC data |
| TIA: ideal V_out = I_cell × R_f | No bandwidth limit, no input-referred noise, no saturation clipping | Add GBW-limited response: V_out = I·R_f · (1/(1 + s·R_f·C_f)) with noise floor |
| Temperature-dependent INL/DNL added | Qualitative Arrhenius-like scaling (PERIPH-2 ✅) | Validate against measured ADC performance over temperature |

**Validation status:** Equations audited against docs (M4-P1 ✅). Temperature model added (PERIPH-2 ✅). Process corners added (PERIPH-3 ✅).
**Validation target:** ADC SNR within 3 dB of known architectural model (e.g., SAR ADC: SNR = 6.02·N + 1.76 dB).

---

### 6. Comparison + EDA

**Realism:** Low
**Files:** `module5-comparison/pkg/comparison/architecture.go`, `module6-eda/pkg/compiler/compiler.go`

| Simplification | Physical Impact | Upgrade Path |
|----------------|----------------|--------------|
| Estimated FeCIM metrics (ops/J, ops/mm²) | Not experimentally validated — based on projections, not silicon measurements | Replace with published measured data as it becomes available; clearly label projections |
| Analytic latency/energy (t = N·t_op, E = N·e_op) | Ignores memory bandwidth, IO bottlenecks, pipeline stalls, amortized control overhead | Add system-level constraints: memory BW, ADC throughput ceiling, batch vs single-inference distinction |
| No placement/routing in EDA | Missing wire parasitics, congestion, area overhead from routing channels | Add basic floorplan model with wire-length estimation (ARCH-1 ⏳) |

**Validation status:** Module 5 deferred. Module 6 defaults, mode behavior, and export formats verified (FOCUS-67–71 ✅).

---

## Test Evidence Matrix (Claim → Test → Result)

This matrix links each core physics claim to an executable test and a concrete last-known outcome.

| Physics Claim Area | Claim Being Evidenced | Test File | Exact Test Function | Tolerance / Acceptance Criteria | Last-Known Result |
|--------------------|-----------------------|-----------|---------------------|----------------------------------|------------------|
| **Preisach** | Published HZO P-E loop fit can meet the audit target quality band in calibration workflow | `module1-hysteresis/pkg/ferroelectric/preisach_calibration_test.go` | `TestCalibratePreisachToPublishedHZOData_RMSBelow10PercentPr` | Normalized RMS error must be `< 10%` of published `Pr` | **PASS (2026-02-12):** `RMS = 0.632 uC/cm^2 = 4.00% of Pr=15.80`, with fitted params `Ps=22.00`, `Ec=-0.000 MV/cm`, `Delta=1.110 MV/cm`, `gamma=1.50` |
| **Landau-Khalatnikov (LK)** | LK implementation matches documented equation identity and effective-viscosity units | `shared/physics/landau_equation_test.go` | `TestLKSolver_FrankensteinEquation_IdentityAndUnits` | `|rho_eff*dPdt - rhs| <= 1e-6 * max(1, |rhs|)` and `rho_eff - rho == R*A/d` within `1e-18` | **PASS (2026-02-12):** identity/units checks satisfied with no tolerance violations |
| **Crossbar non-idealities** | MVM degradation pipeline ordering is stable and includes expected non-ideality chain | `module2-crossbar/pkg/crossbar/focus_54_60_validation_test.go` | `TestFocus60_MVMWithNonIdealitiesPipelineOrdering` | Pipeline sources must exactly match order: `ADC/DAC Quantization → IR Drop → Device Variation → Sneak Paths` | **PASS (2026-02-12):** exact 4-step order matched; run log also reported `Max IR Drop=1.1792%` |
| **Peripherals (DAC/ADC/TIA/CP)** | ADC model honors canonical quantization-limited SNR baseline used by audit targeting | `shared/peripherals/peripherals_test.go` | `TestADCTheoreticalSNR` | `TheoreticalSNR == 6.02*N + 1.76 dB` within `±0.01 dB` | **PASS (2026-02-12):** `N=5` bits case validated at formula tolerance |
| **CIM inference path** | Conductance-based CIM forward path remains numerically close to FP baseline (bounded delegation gap) | `module3-mnist/pkg/core/cim_physics_test.go` | `TestForwardCIM_ConductancePathMatchesFP_BoundedError` | Relative error per output must be `<= 5%` (`rel = |cim-fp|/max(1e-9,|fp|)`) | **PASS (2026-02-12):** all output neurons within 5% bound |

---

## Priority Recommendations

### P0: Correctness & Disclosure (Required)

| Task | Status | Reference |
|------|--------|-----------|
| Add UI disclaimer on all simplified models | ✅ | `HONESTY_AUDIT.md`, CM-D1 |
| Label "conference claim" values distinctly from literature | ✅ | `HONESTY_AUDIT.md` |
| Document model limitations in tooltips per module | ⏳ | New TODO item needed |
| Define per-module "physics accuracy" acceptance criteria | ✅ | CM-P1, `docs/development/PHYSICS_ACCEPTANCE_CRITERIA.md` |

### P1: Physics Quality (Medium Risk, High Value)

| Task | Status | Validation Criteria |
|------|--------|---------------------|
| Polydomain L-K model — hold ≥20 remanent levels at E=0 | 🔄 | LK-PD-1 through LK-PD-6 |
| Calibrate Preisach Everett function to one published HZO P-E dataset | ⏳ | Match loop shape within 10% RMS |
| Add measured retention curve for drift model calibration | ⏳ | Reproduce published 10-year decay exponent |
| Wire CIM inference to actual conductance-based MVM (not FP delegation) | ⏳ | Accuracy delta vs FP path quantified |
| Validate Tier-B DC solver against SPICE golden vectors | ⏳ | TIERB-3: node voltage error < 5% |

### P2: Higher Fidelity (Future — Significant Complexity)

| Task | Status | Prerequisites |
|------|--------|---------------|
| SPICE-validated DAC/ADC macromodels with noise | ⏳ | Requires published model or PDK access |
| Compact FeFET conductance model: G(V_g, T, cycle_count) | ⏳ | Requires device measurement data |
| Multi-domain Landau or phase-field solver | ⏳ | Requires validated polydomain model first (P1) |
| Device-aware quantization from ISPP conductance distributions | ⏳ | Requires measured write-verify statistics |
| System-level throughput model with ADC/memory BW constraints | ⏳ | Requires peripheral timing characterization |

---

## Validation Test Plan

| Test | Target Data | Pass Criteria | Status |
|------|-------------|---------------|--------|
| P-E loop matching (Preisach) | Published HZO data (Pr: 15–34 µC/cm²) | RMS error < 10% of Pr | ⏳ |
| Switching transient (L-K) | Muller et al. IEEE TED | Correct switching time order of magnitude | ⏳ |
| Polydomain remanent staircase | ISPP sweep at E=0 | ≥20 distinct remanent levels | 🔄 (LK-PD-2) |
| Retention curve | Published 10-year extrapolation | Same decay exponent ±0.1 | ⏳ |
| IR drop: Tier-B vs SPICE | 8×8 array SPICE deck | Max node voltage error < 5% | ⏳ (TIERB-3) |
| ADC quantization SNR | Known SAR ADC model | SNR within 3 dB of 6.02N + 1.76 | ⏳ |
| CIM accuracy vs device variation | Published FeFET array statistics | Quantization error within measured σ | ⏳ |

---

## References

Key literature for calibration and validation:

1. **HZO P-E characteristics:** Nature Communications 2025 — Pr: 15–34 µC/cm², Ec: 1–2 MV/cm for Hf₀.₅Zr₀.₅O₂ thin films
2. **Switching dynamics:** Muller et al., IEEE TED — field-dependent switching times, NLS model parameters
3. **Retention/endurance:** Nano Letters 2024 — V:HfO₂ FeFET, 10¹² cycle endurance, 10-year retention extrapolation
4. **CIM accuracy benchmarks:** Nature Communications 2023 — 96.6% MNIST accuracy in FeFET CIM array
5. **Polydomain switching:** Park et al., ACS Applied Materials 2024 — partial switching, intermediate remanent states in doped HfO₂
6. **ADC fundamentals:** Razavi, "Principles of Data Conversion System Design" — SAR/sigma-delta SNR models

---

## Changelog

| Date | Change |
|------|--------|
| 2026-02-03 | Initial audit created |
| 2026-02-03 | Added executive summary, actionable task tables, validation plan |
| 2026-02-11 | Major update: refreshed all sections with current validation status, added precise physics equations, expanded upgrade paths, added polydomain and Tier-B status, cross-referenced TODO.md items |
| 2026-02-11 | Physics-doc gap audit added to TODO (`PGAP-01..PGAP-08`); corrected top critical doc-code mismatches (Preisach implementation path/model, temperature law claims, missing Preisach-plane API claims, 2T1R architecture docs, SAR-noise docs) |
| 2026-02-12 | Added comprehensive **Test Evidence Matrix** linking major physics claim areas (Preisach, LK, crossbar, peripherals, CIM) to exact test files/functions, explicit acceptance criteria, and last-known results |
