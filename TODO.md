# FeCIM Lattice Tools - TODO

**Mission**: Educational FeCIM visualization and simulation tool based on reported in literature HfO₂-ZrO₂ superlattice research.

**Last updated**: 2026-02-05

---

## 0. Focus Areas

This TODO prioritizes (1) **physics accuracy**, (2) **UI/UX correctness**, and (3) **documentation alignment** across all modules.

### This Week (Top 5)

1. M4-D1: Update Module 4 docs to reference `shared/peripherals` everywhere (fix stale `module4-circuits/pkg/peripherals` paths).
2. ✅ M4-U3: Sense-chain UI: expose TIA output, ADC code/saturation, and measurement-path toggles (Tier A arraysim).
3. ✅ M4-P3: Define/centralize cell geometry (area, thickness) and use it consistently in arraysim current/charge equations.
4. M4-P4: Implement Tier B DC solver for arraysim (full resistive network solve) + regression tests.
5. M3-D2: Align noise bounds (docs/UI 0.20 max vs code clamp 0.50) and document rationale.

### Module 1: Hysteresis (Physics + UI/UX + Docs)

| ID | Area | Task | Status | Est. |
|----|------|------|--------|------|
| M1-D1 | Docs | Document run modes (GUI/TUI/headless/Vulkan) and clarify L-K vs Preisach defaults in `docs/documentation/module1-hysteresis/*.md` | ⏳ | 30-60m |
| M1-U1 | UI/UX | Fix WRD target marker parity (single snapshot for target/marker/logs) | ✅ | 1-2hr |
| M1-U2 | UI/UX | Equation widget perf acceptance: cold <1s, warm <200ms, no UI freeze (async load + SVG cache reuse; measure via `FECIM_EQUATION_PERF=1` logs or benchmarks). | ✅ | 30-60m |
| M1-P1 | Physics | L-K performance accounting + ISPP stabilization evidence (HI/MID/LO) vs `docs/hysteresis/hysteresis-gemini.md` | ⏳ | 2-4hr |

### Module 2: Crossbar (Physics + UI/UX + Docs)

| ID | Area | Task | Status | Est. |
|----|------|------|--------|------|
| M2-U1 | UI/UX | Align `cmd/crossbar-gui -help` feature list with implemented features (remove write-verify/differential claims or implement). AC: `-help` list matches actual tabs/features; no write-verify/differential claims unless implemented. Cmd: `go run ./cmd/crossbar-gui -help` | ✅ | 30-60m |
| M2-U2 | UI/UX | Replace 87% “measured hardware” target in accuracy waterfall with HONESTY_AUDIT-aligned labeling | ✅ | 30-60m |
| M2-D1 | Docs | Update `docs/crossbar/reference/PHYSICS.md` to reflect actual ADC/DAC defaults (6/8 bits) or change code to match docs | ✅ | 1-2hr |
| M2-D2 | Docs | Update `docs/crossbar/reference/VOLTAGE_RULES.md` peripheral file references to `shared/peripherals` and re-verify V/2 references | ✅ | 1-2hr |
| M2-P2 | Physics | Apply temperature scalings beyond wire resistance (conductance window, noise, drift) in MVM-with-non-idealities or gate behind options; add tests | ✅ | 1-2hr |
| M2-P1 | Physics | Full physics audit vs `docs/crossbar/reference/PHYSICS.md` (IR drop, sneak paths, drift, variation, temperature) | ⏳ | 2-4hr |

**M2 Validation Notes (2026-02-05)**
- GUI temperature slider uses Kelvin with Celsius in the label (`formatTemperatureLabel`).
- GUI temperature slider range is 77K–450K; docs list a 4K preset that is not currently selectable via slider.
- MVM temperature path now applies wire resistance + conductance window scaling + variation/noise scaling; drift remains time-based via the drift simulator.
- Wire resistance factor is clamped at deep cryo to avoid negative/NaN resistances (see `TempColdSpace` case).
- Acceptance check: `GOCACHE=/tmp/go-build go test ./module2-crossbar/...`
- Acceptance check (temperature-specific): `go test ./module2-crossbar/pkg/crossbar -run TestTemperature`

### Module 3: MNIST (Physics + UI/UX + Docs)

| ID | Area | Task | Status | Est. |
|----|------|------|--------|------|
| M3-D2 | Docs | Align noise bounds (docs/UI 0.20 max vs code clamp 0.50) and document rationale. AC: UI slider max, core clamp, presets, and docs agree on one max; tests updated. Cmd: `rg -n "Noise" module3-mnist/pkg/core/network_config.go module3-mnist/pkg/gui/dualmode_controls.go docs/documentation/module3-mnist/FEATURES.md` | ⏳ | 15-30m |
| M3-D1 | Docs | Sync `docs/documentation/module3-mnist/*.md` with current file paths and core vs training split | ⏳ | 30-60m |
| M3-U1 | UI/UX | Audit GUI labels/metrics to ensure accuracy/energy values are labeled as modeled (not verified) | ⏳ | 30-60m |
| M3-P2 | Physics | Align energy model between core (bit-scaled µJ) and GUI energy widget/comparison card (fixed 50 fJ/MAC, static ratios); choose SSOT + update UI/docs. AC: GUI energy uses same formula or explicitly labeled separate model; docs updated; tests pass. Cmd: `rg -n -e "EnergyPerMAC" -e "EnergyUsed" module3-mnist/pkg/core module3-mnist/pkg/gui` | ✅ | 1-2hr |
| M3-U2 | UI/UX | Decide whether dual-mode should expose confusion matrix/metrics or label them as single-mode-only in UI | ⏳ | 1-2hr |
| M3-P1 | Physics | Verify FP vs CIM inference pipeline ordering + quantization/noise injection vs docs; add evidence logs/tests | ⏳ | 2-4hr |

### Module 4: Circuits (Physics + UI/UX + Docs)

| ID | Area | Task | Status | Est. |
|----|------|------|--------|------|
| M4-D1 | Docs | Update Module 4 docs to reference `shared/peripherals` (`docs/documentation/module4-circuits/*.md`, `docs/peripheral-circuits/*`). AC: no `module4-circuits/pkg/peripherals` references remain; sources point to `shared/peripherals/analysis.go`. Cmd: `grep -RIn "module4-circuits/pkg/peripherals" docs/documentation/module4-circuits docs/peripheral-circuits` | ⏳ | 30-60m |
| M4-P2 | Physics | **Arraysim Tier A**: couple array state to peripheral models (DAC→array, array→sense chain via TIA/ADC) with deterministic testable semantics | ✅ | 1-2hr |
| M4-U1 | UI/UX | Validate ISPP engine toggle wiring (Fast vs L-K) and ensure GUI uses shared/physics with thin adapter | ⏳ | 1-2hr |
| M4-U3 | UI/UX | Sense-chain UI: show TIA output (V/I), ADC code + saturation/clip indicators, and clearly labeled measurement-path presets for Tier A arraysim | ✅ | 1-2hr |
| M4-P3 | Physics | Define/centralize cell geometry (area, thickness, stack) and use it consistently for current/charge/voltage conversions (arraysim + docs) | ✅ | 1-2hr |
| M4-P1 | Physics | Audit DAC/ADC/TIA/ChargePump equations vs `docs/peripheral-circuits/PHYSICS.md` | ⏳ | 2-4hr |
| M4-P4 | Physics | **Arraysim Tier B**: DC solver (full resistive network solve for line/selector models) + regression tests; ensure Tier A vs Tier B fidelity is documented | ⏳ | 4-12hr |
| M4-U2 | UI/UX | Write/Write Cell UX + circuit-coupled updates during ISPP (selected cell + neighbors update each pulse/verify) | ⏳ | 4-12hr |

#### M4-U2 Subtasks

| ID | Task | Status | Est. |
|----|------|--------|------|
| M4-U2a | Route write voltage through DAC output (quantization + INL/DNL) before applying to array | ✅ | 1-2hr |
| M4-U2b | Apply V/2 (0T1R) or pass-through (1T1R/2T1R) voltages and update neighbor polarization live | ✅ | 2-4hr |
| M4-U2c | Throttle ISPP UI refresh and enforce Fyne-safe updates from goroutines | ✅ | 1-2hr |
| M4-U2d | Add targeted tests/visual checks for half-select disturb + applied DAC voltage display | ⏳ | 1-2hr |
| M4-U2e | Disambiguate WRITE mode vs program action (rename Program Cell, tooltips, tests) | ✅ | 30-60m |

**Decision (2026-02-05):** Keep WRITE as the mode selector; make `Program Cell` the only write action with explicit tooltips. Remaining work: M4-U2d.

### Module 5: Comparison (Physics + UI/UX + Docs)

| ID | Area | Task | Status | Est. |
|----|------|------|--------|------|
| M5-U1 | UI/UX | Replace CLI/GUI comparison banners with HONESTY_AUDIT-aligned copy (remove TRL/verified claims; fix path) | ✅ | 30-60m |
| M5-P1 | Physics | Verify energy/ROI equations vs code defaults; ensure values are labeled as model inputs | ✅ | 1-2hr |
| M5-D1 | Docs | Update `docs/documentation/module5-comparison/*.md` to match workload defaults and honesty phrasing | ✅ | 1-2hr |

### Module 6: EDA (Physics + UI/UX + Docs)

| ID | Area | Task | Status | Est. |
|----|------|------|--------|------|
| M6-D1 | Docs | Sync `docs/documentation/module6-eda/*.md` with actual exports (JSON/CSV/SPICE/Verilog/DEF/LEF/Liberty/SVG) and note GUI-only/placeholder outputs. AC: docs list all export formats and scope; matches module6-eda exporters. Cmd: `rg -n -e "Export" -e "LEF" -e "Liberty" -e "SVG" docs/documentation/module6-eda/*.md module6-eda/pkg/export/*.go` | ⏳ | 1-2hr |
| M6-U1 | UI/UX | Check GUI/CLI parity (Start/Stop, defaults) and document any drift | ⏳ | 1-2hr |
| M6-P1 | Physics | Audit mapping/quantization/topology vs docs; verify export contents | ⏳ | 2-4hr |

### Module 7: Docs (Physics + UI/UX + Docs)

| ID | Area | Task | Status | Est. |
|----|------|------|--------|------|
| M7-D1 | Docs | Confirm curriculum tree order + shortcuts match `docs/documentation/`; update README/MODULES if needed | ⏳ | 30-60m |
| M7-U2 | UI/UX | Add colored category badges in tree rows to match curriculum-first UI spec | ⏳ | 30-60m |
| M7-U3 | UI/UX | Hide the “On This Page” sidebar when ToC has < 3 headings | ✅ | 30-60m |
| M7-U1 | UI/UX | Validate layout breakpoints + click targets vs `docs/development/GUI/GUI.module7.md` | ⏳ | 1-2hr |
| M7-P1 | Physics | Verify search ranking + reading time math vs `docs/documentation/module7-docs/PHYSICS.md` | ⏳ | 1-2hr |

### Cross-Module (Shared Infrastructure)

| ID | Area | Task | Status | Est. |
|----|------|------|--------|------|
| CM-P1 | Physics | Define “physics accuracy” acceptance criteria per module and align wording with `docs/comparison/HONESTY_AUDIT.md` | ⏳ | 30-60m |
| CM-D1 | Docs | Keep `HONESTY_AUDIT.md` as SSOT and ensure UI labels match it everywhere | ⏳ | 30-60m |
| CM-U1 | UI/UX | Ensure UI values/plots never desync from engine state (single source of truth snapshot wiring) | ⏳ | 1-2hr |
| CM-D2 | Docs | Equation widgets pipeline: LaTeX→SVG SSOT, hotspot alignment, and SVG caching | ⏳ | 1-2hr |
| CM-P2 | Physics | Add minimal headless regression suite per engine (Preisach + LK) with compact JSON summary | ⏳ | 2-4hr |
| CM-D3 | Docs | Tighten module docs so each model has equations, assumptions, units, and reported vs validated labels | ⏳ | 2-4hr |

---

## LK Tracking (headless)

| ID | Task | Acceptance | Status |
|----|------|------------|--------|
| LK-C01 | Verify LK equation terms/signs in `shared/physics/landau.go` match the compendium (E_eff = E_applied - k_dep·P; dP/dt = (E_eff - dG/dP + noise)/rho_eff; dG/dP = 2αP + 4βP^3 + 6γP^5). | `go test ./shared/physics -run TestLKSolver_dPdT_Equation` | ✅ |
| LK-C02 | Verify effective-viscosity wiring `rho_eff = rho + (R_series·A/d)` when `UseEffectiveViscosity=true`. | `go test ./shared/physics -run TestLKSolver_effectiveRho` | ✅ |
| LK-C03 | Headless LK run uses E-field units and completes the 5-target ISPP sequence without NaN/Inf states. | `go run ./cmd/fecim-lattice-tools --mode hysteresis --engine lk` | ⏳ |
| LK-C04 | Doc parity: note `UseMaterialAlpha` (Pr-calibrated α) vs dynamic α in compendium. | Check `docs/hysteresis/hysteresis-gemini.md` update. | ✅ |

## Calibration JSON hygiene

- Calibration baselines live in `cmd/fecim-lattice-tools/data/calibrations/*.json`.
- These files are **tracked** (so `.gitignore` will not help). To prevent accidental commits of auto-updated calibration drift, mark them locally as unchanged:
  - `git update-index --assume-unchanged cmd/fecim-lattice-tools/data/calibrations/literature_superlattice.json`
  - (reverse: `git update-index --no-assume-unchanged <file>`)
- Policy: do **not** commit calibration JSON changes unless intentionally updating the baseline + evidence logs.


- LK04: ✅ Ec-normalization implemented (scale Landau coefficients to match material Ec while preserving Pr). Evidence: `logs/2026-02-03_20-16-15-fecim.log` shows HI2/LO5/MID targets converging under Literature Superlattice.
- LK05/LK07: still pending (reduce overshoot/oscillation around MID and tighten bounds/step logic; see overshoots in MID).

**Master Critique Source**: See `CRITIQUE_MASTER_LIST.md` for consolidated items (snapshot). Current status lives in `docs/project/STATUS.md`.

---

## 1. Current Status

| Module | Purpose | Status | Tests | GUI |
|--------|---------|--------|-------|-----|
| **module1-hysteresis** | P-E hysteresis, Preisach model, 30-level baseline (simulation baseline) | ✅ Complete | See CI | ✅ Fyne |
| **module2-crossbar** | Matrix-vector multiplication, IR drop, sneak paths | ✅ Complete | See CI | ✅ Fyne |
| **module3-mnist** | Neural network digit recognition | ✅ Complete | See CI | ✅ Fyne |
| **module4-circuits** | DAC/ADC/TIA peripheral circuits | ✅ Complete | See CI | ✅ Fyne |
| **module5-comparison** | Technology comparison framework | ✅ Complete | See CI | ✅ Fyne |
| **module6-eda** | Verilog/DEF/LEF/Liberty generation | ✅ Complete | See CI | ✅ Fyne |
| **docs/** | Scientific documentation (78 papers catalogued) | ✅ Complete | N/A | N/A |

**Project Status**: See `docs/project/STATUS.md` for phase, validation status, and test/coverage notes.

---

## 2. Priority Rankings

### Priority Legend
- **P1 (Critical)**: Must fix before any public release
- **P2 (High)**: Fix before academic/educational use
- **P3 (Medium)**: Polish and enhancement
- **P4 (Low)**: Nice to have

### Difficulty Legend
- **D1 (Easy)**: <1 hour
- **D2 (Medium)**: 1-4 hours
- **D3 (Hard)**: 4-16 hours
- **D4 (Very Hard)**: 16+ hours

---

## 3. CRITICAL PRIORITY (P1) - Must Fix Before Release

### P1-D1: Easy Critical Fixes (Sprint 1 - Day 1)

| ID | Task | Status | Est. |
|----|------|--------|------|
| C01 | Add "SIMULATION ONLY - NOT VALIDATED" banners to Module 5 comparison screens | ✅ | Done |
| C02 | Change "30 discrete states" from fact to baseline (unverified) | ✅ | Done |
| C03 | ~~Add simulation-only disclaimer to energy comparison charts~~ | ✅ | Done |
| C04 | ~~Remove 87% MNIST claim; label literature benchmarks as reported~~ | ✅ | Done |
| C05 | ~~Add "Why 30?" dialog with baseline explanation~~ | ✅ | Done |

### P1-D2: Medium-Effort Critical Fixes (Sprint 2 - Days 2-4)

| ID | Task | Status | Est. |
|----|------|--------|------|
| C06 | Add error bars/confidence intervals to all physics parameters in UI displays | ✅ | Done |
| C07 | Fix temperature-dependent retention with Arrhenius scaling: `driftRate *= exp((Ea/k) * (1/T_ref - 1/T))` | ✅ | Done |
| C08 | ~~Accuracy degradation chart - add sources and confidence intervals~~ | ✅ | Done |
| C09 | Label ALL extrapolated accuracy as "projected" not "achieved" | ✅ | Done |
| C10 | Add total system power breakdown showing: Array (45%) + ADC/DAC (40%) + Peripherals (15%) | ✅ | Done |

### P1-D3: Hard Critical Fixes (Sprint 3 - Week 1)

| ID | Task | Status | Est. |
|----|------|--------|------|
| C11 | Implement device-to-device variation with Gaussian Ec/Pr distribution (mean=1.0 MV/cm, sigma=0.15) | ✅ | Done |
| C12 | Add write-verify statistics visualization (pulses to converge, failure rate vs cycles) | ✅ | Done |
| C13 | Validate Preisach model against experimental hysteresis loop measurements | ✅ | Done |

---

## 4. HIGH PRIORITY (P2) - Fix Before Academic Use

## Engineering Guardrails (Architecture Review)

### P1: Reproducibility + drift control

| ID | Task | Status | Est. |
|----|------|--------|------|
| G02 | Add “intentional calibration update” policy: require linking the evidence log/CSV paths in the commit message/body when calibrations are updated | ⏳ | 30m |
| G03 | Provide optional `pre-commit` hook template to block calibration JSON commits by default (not auto-installed) | ⏳ | 30m |
| G01 | Calibration drift guard: add `scripts/calib-guard.sh` to fail CI if `cmd/fecim-lattice-tools/data/calibrations/*.json` changes unless `ALLOW_CALIBRATION_UPDATES=1` | ✅ | 1-2hr |

### P1: Engine parity / regression evidence (headless-first)

| ID | Task | Status | Est. |
|----|------|--------|------|
| G06 | Normalize/verify CLI engine selector (`--engine {preisach,lk}` or document actual selector); ensure all docs/runbooks reference the same mechanism | ⏳ | 30-60m |
| G06b | Verification matrix: for each material, verify both Preisach + LK run and hit a small target set (HI/MID/LO) without crashes; record a one-line PASS/FAIL summary | ⏳ | 1-2hr |
| G04c | Shared ISPP migration plan: define the shared API (target=level and/or conductance), mapping adapters per module, and a deprecation plan for old controllers; add parity tests to prove behavior matches across modules | ⏳ | 1-2hr |
| G04 | Headless WRD/ISPP regression suite: Preisach target-hit within N pulses for HI/MID/LO; emits a compact JSON summary artifact | ⏳ | 2-4hr |
| G05 | Headless LK regression suite: same targets + overshoot/pulse stats (looser thresholds OK), emits JSON summary | ⏳ | 2-4hr |
| G04b | One-source-of-truth ISPP write engine: refactor so the ISPP state machine lives in `shared/physics` and is reused by module1-hysteresis (GUI + headless) and module4-circuits (GUI + CLI). Remove duplicate controllers (`module1-hysteresis/pkg/controller/WriteController` vs `shared/physics/WriteController` vs module4 local ISPP state) or make them thin adapters. | ⏳ | 4-12hr |

### P2: LK05/LK07 stabilization accounting (make it measurable)

| ID | Task | Status | Est. |
|----|------|--------|------|
| G08 | Define acceptance criteria for Literature Superlattice MID stability (overshoot rate/variance bounds) and record evidence in `hysteresis-prompt.md` | ⏳ | 30-60m |
| G07 | Add LK ISPP overshoot accounting counters: overshoots/target, max overshoot Δ(level), stuck-breaker invocations, bisection usage rate | ⏳ | 1-2hr |

### P2: Performance diagnosis tooling

| ID | Task | Status | Est. |
|----|------|--------|------|
| G09 | Add a single command/script to generate LK perf evidence for 3 targets (steps, dt min/mean/max, solverMs, %dtMinHits) and save output under `logs/` | ⏳ | 1-2hr |
| G10 | (Optional) Add `pprof` toggle for headless hysteresis runs (`FECIM_PPROF=1`) and document usage | ⏳ | 1-2hr |

### P2: GUI correctness guardrails

| ID | Task | Status | Est. |
|----|------|--------|------|
| G12 | Add “GUI parity smoke test” checklist (5-min manual run) with expected log lines + what screenshots to capture | ⏳ | 30-60m |
| G11 | Throttled WRD phase-boundary logging spec: log `wrdTargetLevel`, `wrdNextTargetLevel`, `controller.TargetLevel`, `controller.State`, `discreteLevel` at transitions (no per-frame spam) | ⏳ | 1hr |
| G11b | Refactor target/phase snapshot wiring: ensure a single snapshot struct is shared across GUI widgets so target/marker/logs cannot desync (eliminate parallel sources of truth) | ⏳ | 1-2hr |
| G11c | Write/Write Cell ISPP UX + circuit-coupled updates: during ISPP, update the selected cell’s polarization/level live on each pulse/verify; drive input voltage from the DAC above via the peripheral circuit (not a direct set); compute neighbor polarization updates due to pass-through/half-select voltages and update them live as well | ⏳ | 4-12hr |

### P3: UX polish standard

| ID | Task | Status | Est. |
|----|------|--------|------|
| G13 | Define minimum supported GUI size (e.g., 1024×768) and ensure key widgets (Physics Equations, Log, Controls) remain usable (scroll/min-size conventions) | ⏳ | 30-60m |
| G14 | GUI overlap audit: reproduce widget overlap/clipping on resize (controls cards, log, plot/level, equations tab) and fix via min-sizes, scrolls, and split offsets (layout-only) | ✅ | 1-2hr |
| G15 | GUI layout doc sync: update `docs/development/GUI/GUI.module1.md` to match current code (post-2026-02-04 layout refactor), including correct min sizes, scroll containers, splits, and component hierarchy | ⏳ | 1-2hr |
| G16 | Documentation mapping sweep: audit `docs/development/GUI/GUI.module{1..7}.md` for drift vs code; add a lightweight checklist + update “Last Updated” stamps when verified | ⏳ | 2-4hr |


### P2-D1: Easy High-Priority Fixes

| ID | Task | Status | Est. |
|----|------|--------|------|
| H01 | ~~Home screen module descriptions - add simulation caveats~~ | ✅ | Done |
| H02 | ~~MAC count parallelism explanation~~ | ✅ | Done |
| H03 | ~~Voltage range citations (thickness-dependent)~~ | ✅ | Done |
| H04 | ~~Read parameter sources - mark as empirical~~ | ✅ | Done |
| H05 | ~~Market chart disclaimers - simulation-only and projection warnings~~ | ✅ | Done |
| H06 | Cite strain coefficients - replace magic `-0.15` with Haun 1987 or DFT reference | ✅ | Done |
| H07 | Add Preisach grid size convergence study reference (why 50×50?) | ✅ | Done |

### P2-D2: Medium-Effort High-Priority Fixes

| ID | Task | Status | Est. |
|----|------|--------|------|
| H08 | Add "Fabrication Reality" section: 18-month tape-out timeline, $2M first silicon cost | ✅ | Done |
| H09 | Module 4 - Add SAR ADC noise modeling (comparator metastability, reference drift) | ✅ | Done |
| H10 | Module 2 - Add write disturb (half-select stress) model | ✅ | Done |
| H11 | Module 2 - Add parasitic capacitance for realistic RC delay | ✅ | Done |
| H12 | ~~Weight error context - add % of range explanation~~ | ✅ | Done |
| H13 | ~~GPU comparison nuance - add batched operation context~~ | ✅ | Done |

### P2-D3: Hard High-Priority Fixes

| ID | Task | Status | Est. |
|----|------|--------|------|
| H14 | Add ISPP (Incremental Step Pulse Programming) visualization with convergence stats | ✅ | Done |
| H15 | Implement full thermal physics (retention vs temperature curves, 25°C-85°C automotive) | ✅ | Done |
| H16 | Add "Simulation vs Experiment" comparison tab with placeholder for real data | ✅ | Done |

---

## 5. MEDIUM PRIORITY (P3) - Polish

### P3-D1: Easy Medium-Priority Fixes

| ID | Task | Status | Est. |
|----|------|--------|------|
| M01 | ~~EDA status prominence - "Coming Soon" label~~ | ✅ | Done |
| M02 | Hysteresis Ec threshold visualization (dashed lines at ±Ec) | ✅ | Done |
| M03 | ~~Voltage zone legend (green/yellow/red)~~ | ✅ | Done |
| M04 | ~~Energy breakdown annotation (peripheral percentages)~~ | ✅ | Done |
| M05 | ~~Glossary widget integration~~ | ✅ | Done |
| M06 | ~~References widget with DOI links~~ | ✅ | Done |
| M16 | Physics equations UI: cover hysteresis ISPP, Preisach, and Landau (labels + links) | ✅ | Done |

### P3-D2: Medium-Effort Medium-Priority Fixes

| ID | Task | Status | Est. |
|----|------|--------|------|
| M07 | Simplified log toggle for hysteresis memory log | ✅ | Done |
| M08 | Sneak path side-by-side comparison view (PASSIVE vs 1T1R) | ✅ | Done |
| M09 | Crossbar cell-level inspection with hover tooltips | ✅ | Done (existing) |
| M10 | Architecture comparison split-screen mode | ✅ | Done (SneakCompareWidget) |
| M11 | Error attribution breakdown in accuracy analysis | ✅ | Done (existing) |
| M12 | Hysteresis state stability warnings (color-coded levels 1-5 green, 6-25 yellow, etc.) | ✅ | Done |

### P3-D3: Hard Medium-Priority Fixes

| ID | Task | Status | Est. |
|----|------|--------|------|
| M13 | Responsive layout breakpoints (>1600px 3-col, 1024-1600px 2-col, <1024px 1-col) | ✅ | Done |
| M14 | Conductance drift time-dependent visualization (1s, 1hr, 1day, 1year) | ✅ | Done |
| M15 | Device-to-device variation modeling GUI with yield prediction | ✅ | Done |

---

## 6. LOW PRIORITY (P4) - Nice to Have

### P4-D1: Easy Low-Priority Fixes

| ID | Task | Status | Est. |
|----|------|--------|------|
| L11 | Add [LK] indicators to material_picker.go for Landau-Khalatnikov parameters (requires LK model implementation first - see docs/hysteresis/hysteresis-glm.md Phase 4: P1 Landau-Khalatnikov model) | ⏳ | 1hr |
| L01 | Hysteresis cycle labels (wake-up, stable, fatigue phases) | ✅ | Done |
| L02 | Screenshot metadata embedding (PNG EXIF with parameters) | ✅ | Done |
| L03 | Add GitHub URL to glossary widget TODO | ✅ | Done (existing) |
| L04 | Hysteresis polarization bar indicator - increase to 16px with pulsing | ✅ | Done |

### P4-D2: Medium-Effort Low-Priority Fixes

| ID | Task | Status | Est. |
|----|------|--------|------|
| L07 | Demo video creation (2-3 min walkthrough) | ⏳ | 4hr |
| L05 | "About the Science" unified Learn More section | ✅ | Done |
| L06 | Accessibility audit (keyboard nav, ARIA labels, high-contrast mode) | ✅ | Done |

### P4-D3: Hard Low-Priority Fixes

| ID | Task | Status | Est. |
|----|------|--------|------|
| L08 | Web deployment (WASM) for browser-based demos | ⏳ | 16hr |
| L09 | Vulkan rendering implementation for large arrays | ⏳ | 20hr |
| L10 | 3D multi-layer visualization (512-layer roadmap) | ⏳ | 24hr |

---

## 7. Documentation Work

### Doc Debt (2026-02-05 Sweep)

| ID | Task | Status | Est. |
|----|------|--------|------|
| DOC-CITE-1 | Add DOI citations (or remove numeric claims) for ELI5 energy, HZO property, and data-center projection numbers (`docs/ELI5.md`) | ⏳ | 1-2hr |
| DOC-CITE-2 | Verify or replace literature DOIs in crossbar voltage/physics references (`docs/crossbar/reference/PHYSICS.md`, `docs/crossbar/reference/VOLTAGE_RULES.md`) | ⏳ | 2-4hr |
| DOC-CITE-3 | Cite peripheral timing/energy assumptions or label as placeholders (`docs/peripheral-circuits/PHYSICS.md`) | ⏳ | 1-2hr |
| DOC-CITE-4 | Cite hysteresis parameter values or label as placeholders (`docs/hysteresis/hysteresis.physics.md`) | ⏳ | 1-2hr |
| DOC-LINK-1 | Fix broken internal markdown links (110 found in docs scan; prioritize `docs/README.md` and `docs/opensource-tools/*`) | ⏳ | 2-4hr |

### Quarterly Literature Review
**Status**: Scheduled | **Due**: April 2026 | **Priority**: P3

**Goal**: Update HONESTY_AUDIT.md with 2026 Q1 publications.

**Search databases**:
- IEEE Xplore (IEDM, ISSCC, VLSI)
- Nature family (Nature Commun., Nature Electronics)
- ACS (Nano Letters, ACS AMI)
- arXiv (cs.ET, cond-mat.mtrl-sci)

### Add Citations for 12 Assumed Parameters
**Status**: P2 | **Priority**: HIGH

From HONESTY_AUDIT Section 6.2:

| Parameter | Location | Current Value | Needed |
|-----------|----------|---------------|--------|
| DRAM energy (32-bit) | physics.md:45 | 640 pJ | Horowitz 2014 citation |
| Drift coefficients | equations.md:407 | Various | FeFET-specific literature |
| FeFET capacitance | equations.md:648 | 10-100 fF | Device physics reference |
| FeFET switching current | equations.md:649 | 1-10 µA | Measurement data |
| Nonlinearity coefficient k | mathematics.md:389 | 5-10 | Crossbar model reference |
| Read disturb probability | mathematics.md:731 | 10^-6 | Endurance study reference |
| Strain factor | materials.go | -0.15 | Haun 1987 or DFT |
| Preisach grid size | preisach.go | 50 | Convergence study |

---

## 8. Testing

### Current Coverage
- Total tests: See CI (`go test ./...`)
- Pass rate: See CI
- Coverage: ~85% (estimated; update when CI coverage is added)

### Needed Tests

| Test Type | Description | Priority |
|-----------|-------------|----------|
| Device variation | Gaussian distribution accuracy | P1 |
| Thermal retention | Temperature-dependent drift | P1 |
| Write-verify convergence | ISPP statistics | P2 |
| Full EDA pipeline | RTL → GDSII validation | P2 |

---

## 9. NOT Planned (Out of Scope)

| Out of Scope | Reason |
|--------------|--------|
| Production chip design tools | Educational tool, not EDA replacement |
| Investor pitch decks | Scientific tool, not marketing material |
| Hardware-accurate SPICE models | Requires foundry PDKs (proprietary) |
| Real-time OS integration | Beyond educational scope |
| Cryptographic accelerators | Specialized application |
| Web-based collaboration | Single-user educational tool |

---

## 10. Claims Reference (Use HONESTY_AUDIT)

All external performance/material claims must be treated as **reported** or **unverified** unless explicitly validated by this codebase. Use `docs/comparison/HONESTY_AUDIT.md` as the single source of truth for what can be cited, and avoid adding “verified” labels in docs or UI.

---

## 11. Sprint Planning

### Sprint 1: Critical Easy Wins (1 day) ✅ COMPLETE
**Goal**: Address all P1-D1 items
- [x] C01: Add "SIMULATION ONLY" banners ✅
- [x] C02: Change "30 states" language to hypothesis ✅
- [x] H06: Cite strain coefficients ✅
- [x] H07: Add Preisach grid size reference ✅

### Sprint 2: Critical Physics (3 days)
**Goal**: Address P1-D2 items
- [ ] C06: Error bars on physics parameters
- [ ] C07: Temperature-dependent retention
- [ ] C09: Label accuracy as "projected"
- [ ] C10: Total system power breakdown

### Sprint 3: Device Variation (1 week)
**Goal**: Address P1-D3 + key P2 items
- [ ] C11: Gaussian Ec/Pr distribution
- [ ] C12: Write-verify statistics
- [ ] H10: Write disturb model
- [ ] H15: Thermal physics

### Sprint 4: Validation (2+ weeks)
**Goal**: Experimental validation groundwork
- [x] C13: Preisach validation ✅
- [x] H16: Simulation vs Experiment tab ✅
- [x] H14: ISPP visualization ✅

---

## 12. Landau-Khalatnikov Physics Engine Issues (NEW)

**Status**: In Progress | **Priority**: P2 | **Added**: 2026-02-03

The L-K dynamic physics engine has issues with ISPP write/read demo, particularly with high-Gamma materials like Literature Superlattice.

### Fixed Issues

| ID | Issue | File | Fix |
|----|-------|------|-----|
| LK01 | UseNLS not disabled for GUI | `physics_engine.go:68` | Added `UseNLS = false` ✅ |
| LK02 | Reverse step cap too low (0.95×Ec) | `writer.go:645` | Increased to 1.5×Ec ✅ |
| LK03 | Numerical overflow in RK4 | `landau.go:250` | Added rate limiter ✅ |  
| LK03b | LK dynamics frozen by overly-aggressive dP/dt clamp | `shared/physics/landau.go` | Clamp now scales with dt (unfreezes switching) ✅ |

### Open Issues

| ID | Issue | Priority | Difficulty |
|----|-------|----------|------------|
| LK06 | Missing Q12 in some materials | P3 | D1 | (headless presets patched; defaults cover remaining)
| LK07 | Need longer WAIT phases for L-K settling | P2 | D2 |
| LK04 | L-K coefficients not calibrated to Ec/Pr (note: LK04 marked ✅ in LK Tracking above; reconcile status) | P2 | D3 |
| LK05 | ISPP controller not optimized for L-K dynamics | P2 | D3 | (Superlattice shows many overshoots near MID; needs LK-aware settling/step tuning)

### LK04: L-K Parameters Don't Match Material Ec/Pr
**Problem**: Landau coefficients (Alpha, Beta, Gamma) don't produce hysteresis matching defined Ec/Pr.
**Symptoms**: Narrow/collapsed hysteresis, wrong switching fields
**Solution Options**:
1. Auto-calibrate from Ec/Pr: `Ec ≈ 2*sqrt(α³/β)`, `Pr ≈ sqrt(-α/β)`
2. Pre-validated parameter sets per material
3. Runtime calibration routine

### LK05: ISPP Controller Assumes Preisach Behavior
**Problem**: ISPP designed for quasi-static switching; L-K has time-dependent dynamics
**Symptoms**: Stuck at intermediate levels, overshoots
**Solution**: Physics-engine-aware step sizes, longer settling times

### Testing Checklist
- [ ] DefaultHZO + L-K + ISPP → all levels
- [ ] FeCIM HZO + L-K + ISPP → baseline targets (✅ headless)
- [ ] LiteratureSuperlattice + L-K + ISPP → all levels (MID ✅; extremes failing → LK04)
- [ ] L-K + Sine Wave → proper hysteresis loop
- [ ] Switch Preisach → L-K mid-sim → smooth transition

### Code Locations
| Component | File |
|-----------|------|
| L-K Solver | `shared/physics/landau.go` |
| Material Params | `shared/physics/material.go` |
| Physics Switch | `module1-hysteresis/pkg/gui/physics_engine.go` |
| ISPP Controller | `module1-hysteresis/pkg/controller/writer.go` |

---

## 13. Progress Summary

| Priority | Total | Done | Remaining | % Complete |
|----------|-------|------|-----------|------------|
| P1 Critical | 13 | 13 | 0 | 100% |
| P2 High (note: Section 4 guardrail items above still ⏳; totals may exclude them) | 16 | 16 | 0 | 100% |
| P3 Medium (note: no remaining P3 item listed above; counts may be stale) | 17 | 16 | 1 | 94% |
| P4 Low (note: Section 6 shows 5 open items: L07-L11) | 11 | 7 | 4 | 64% |
| **TOTAL** | **57** | **52** | **5** | **91%** |

**Estimated Remaining Effort**: ~29 hours

**Session Progress (2026-01-29)**:
- Sprint 2: C06 ✅, C07 ✅, C09 ✅, C10 ✅
- Sprint 3: C11 ✅, C12 ✅, H10 ✅, H15 ✅
- Sprint 4: C13 ✅, H16 ✅, H14 ✅ (COMPLETE)
- P2 Remaining: H08 ✅, H09 ✅, H11 ✅ (ALL P2 COMPLETE)
- P3 Progress: M12 ✅, M07 ✅, M09 ✅, M08 ✅, M11 ✅, M10 ✅, M14 ✅, M15 ✅, M13 ✅ (ALL P3 COMPLETE)
- P4 Progress: L01 ✅, L02 ✅, L03 ✅, L04 ✅, L05 ✅, L06 ✅ (7/11 complete)

---

## Footer

**Next review**: After Sprint 4 completion
**Contributing**: See CLAUDE.md for development guidelines
**Scientific accuracy**: All claims must be verified per HONESTY_AUDIT.md standards

---

*This TODO prioritizes scientific rigor and educational honesty over promotional considerations. The project is an open-source learning tool, not investment material.*
