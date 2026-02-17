# World-Class Gap Analysis: Module 1 (Hysteresis) & Module 4 (Circuits)

Date: 2026-02-13  
Scope reviewed: `module1-hysteresis/**`, `module4-circuits/**`, related `shared/**`

## Executive Summary

- **Module 1 strength:** strong core hysteresis engine, temperature-aware calibration, retention/endurance/wake-up physics primitives, JSON/CSV export.
- **Module 1 main gap:** lacks measurement workflows expected in ferroelectric labs (PUND, FORC UI/tools, C(V)/I(V), batch recipes, small-signal measurement mode).
- **Module 4 strength:** good circuit-chain educational model (DAC→array→TIA→ADC), write-verify timing, coupled array solver, export pipeline.
- **Module 4 main gap:** no algorithm-level co-sim loop (accuracy vs hardware), no systematic DSE/Pareto mode, no validated peripheral calibration track, no benchmark automation.

---

## Deliverable 1 — Gap Matrix

### Module 1 gaps vs Radiant Vision + FerroX

- [ ] **PUND measurement mode** — **Missing**. No PUND pulse protocol/charge separation workflow found.
- [ ] **Retention measurement mode** — **Partial**. Physics model exists (`RetentionAtTime`), but no dedicated measurement workflow panel/report.
- [x] **Fatigue cycling** — **Present (model-level)**. `EnduranceAtCycles` + tests/log evidence.
- [ ] **Imprint measurement** — **Partial**. Imprint parameter exists in material model; no dedicated extraction flow after bias-stress protocol.
- [ ] **C(V) butterfly curves** — **Missing**. No voltage-dependent capacitance sweep/derivative workflow.
- [ ] **I-V leakage characterization** — **Missing**. No Schottky/Poole-Frenkel/Fowler-Nordheim mode.
- [ ] **Small-signal capacitance measurement** — **Missing** (only static capacitance formula exposed).
- [x] **Data export: CSV** — **Present** (`pkg/gui/export.go`), plus quick-win column configurability + clipboard copy added.
- [ ] **Batch/recipe mode** — **Missing**. No multi-step measurement recipe executor.
- [x] **Wake-up effect modeling** — **Present (model/UI labels)**.
- [ ] **Frequency-dependent hysteresis** — **Partial**. Frequency control exists; no dispersive parameter extraction/fit workflow.
- [ ] **FORC visualization** — **Partial**. FORC references in tests; no dedicated UI/data product.
- [x] **Temperature sweep mode** — **Partial-to-Present**. Temperature control and calibration cache exist; no one-click automated sweep export report.
- [ ] **Literature comparison overlay** — **Partial**. sim-vs-exp widget exists, but not generic file-loader overlay pipeline.

### Module 4 gaps vs DNN+NeuroSim

- [ ] **Algorithm-level integration** — **Missing**. No direct inference-accuracy-in-loop with module3 network workload in module4 workflow.
- [ ] **Design space exploration mode** — **Missing/Partial**. No DSE UI; quick-win backend sweep helper added (`BuildDesignSpaceSweep`).
- [ ] **Validated peripheral models** — **Missing**. No SPICE/post-layout calibration workflow with error bars.
- [ ] **MLC programming characterization panel** — **Partial**. Write-verify exists, but no dedicated MLC characterization dashboard.
- [ ] **Tiled architecture support** — **Missing**. Mentioned in tooltips, no implemented multi-array/global-buffer model.
- [ ] **Process variation Monte Carlo** — **Missing/Partial**. quick-win backend sampler added (`RunProcessVariationMonteCarlo`), no UI/report integration yet.
- [ ] **Write verify with actual device model in loop** — **Partial**. Program-verify loop exists with current model, but not technology-calibrated device programming model variants.
- [ ] **Endurance-aware accuracy degradation** — **Missing**.
- [ ] **Batch benchmark mode** — **Missing**.
- [ ] **Device-technology comparison** — **Partial**. comparison tab exists; no rigorous side-by-side technology model suite.

---

## Deliverable 2 — Priority Implementation Plan

Scoring: **Priority = Impact (1-5) × Feasibility (1-5)**

| Rank | ID | Item | Impact | Feasibility | Score |
|---:|---|---|---:|---:|---:|
| 1 | M1-WC-01 | PUND measurement mode (full pulse protocol + switching/non-switching separation) | 5 | 4 | 20 |
| 2 | M4-WC-01 | Algorithm-level integration: hardware non-idealities → inference accuracy loop | 5 | 3 | 15 |
| 3 | M1-WC-02 | Retention experiment workflow (stress/hold/read, time sweep, Arrhenius view) | 5 | 3 | 15 |
| 4 | M1-WC-03 | Fatigue+Wake-up experiment runner with cycle scheduling and report | 4 | 4 | 16 |
| 5 | M4-WC-02 | DSE mode (array size × ADC bits × device) + Pareto export | 4 | 4 | 16 |
| 6 | M4-WC-03 | Process variation Monte Carlo integrated in read/compute metrics | 4 | 4 | 16 |
| 7 | M1-WC-04 | C(V) butterfly mode + numerical dP/dV extraction | 4 | 3 | 12 |
| 8 | M1-WC-05 | I-V leakage model panel (Schottky/PF/FN selectable fit) | 4 | 3 | 12 |
| 9 | M4-WC-04 | Endurance-aware inference degradation pipeline | 5 | 2 | 10 |
|10 | M4-WC-05 | Batch benchmark mode (MNIST now; pluggable VGG/ResNet later) | 4 | 2 | 8 |

### Top-10 Full Specs

1. **M1-WC-01 PUND mode**
   - Pulse sequence: P, U, N, D with configurable widths/amplitudes/delays.
   - Outputs: I(t), Q(t), ΔQswitch, derived Pr/Ec.
   - Export: per-pulse CSV + summary JSON.

2. **M4-WC-01 Algorithm integration**
   - Interface module3 weights/activations into module4 non-ideal compute path.
   - Metrics: top-1 accuracy, confusion delta, energy/inference, latency/inference.

3. **M1-WC-02 Retention workflow**
   - Program level → hold at E≈0 for logarithmic time points → verify level/Pr.
   - Multi-temperature batch with Arrhenius projection table.

4. **M1-WC-03 Fatigue/Wake-up runner**
   - Cycle scheduler with wake-up then fatigue regime.
   - Plot/report: Pr(N), Ec(N), window(N), extraction of N10/N50.

5. **M4-WC-02 DSE/Pareto**
   - Sweep dimensions: array size, ADC bits, device technology.
   - Output: CSV with energy/latency/TOPSW and Pareto flag.

6. **M4-WC-03 Variation Monte Carlo**
   - Gaussian/lognormal process knobs on conductance, threshold, noise.
   - Outputs: mean/std/P99 for read current, SNR, ADC-code error.

7. **M1-WC-04 C(V) butterfly**
   - Compute C=dQ/dV using smoothed derivative.
   - Show butterfly branch split with coercive markers.

8. **M1-WC-05 I-V leakage panel**
   - Model selectable fit family with parameter estimator.
   - Residual/error and regime crossover annotation.

9. **M4-WC-04 Endurance-aware accuracy**
   - Map device wear (cycle count) to conductance drift in inference pipeline.
   - Output: accuracy vs cycles with confidence intervals.

10. **M4-WC-05 Batch benchmark mode**
    - One-click benchmark presets + reproducible seed/config artifact.
    - CLI + GUI entrypoint.

---

## Deliverable 4 — Quick Wins Implemented (this pass)

1. **M1 export columns configurability** (`FECIM_EXPORT_COLUMNS`) in `module1-hysteresis/pkg/gui/export.go`.
2. **M1 clipboard export** via `Ctrl+Shift+E` (CSV content copied from current P-E history).
3. **M1 tests** for export-column parsing/content generation (`export_worldclass_test.go`).
4. **M4 backend DSE helper** (`BuildDesignSpaceSweep`) in `module4-circuits/pkg/gui/comparison_metrics.go`.
5. **M4 backend Monte Carlo helper** (`RunProcessVariationMonteCarlo`) + tests (`comparison_metrics_worldclass_test.go`).

Notes from requested checks:
- Fatigue model exists (shared physics endurance model + tests).
- Data export exists (module1 + module4 export paths).
- Temperature sweep capability is partially present (controls + calibration-by-temperature), but not fully productized as a measurement mode.
- No PUND mode found.
- Module4 had no full Monte Carlo/DSE mode prior to this pass (quick backend primitives added now).
