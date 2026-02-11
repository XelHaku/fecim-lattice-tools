# FeCIM Lattice Tools - Comprehensive TODO

**Mission**: Educational FeCIM visualization and simulation tool based on HfO₂-ZrO₂ superlattice research.

**Last Updated**: 2026-02-11 (Refocused priorities)

**Source Documents**: `CRITIQUE_MASTER_LIST.md`, `docs/neural-network/mnist.fixes.todo.md`, `docs/ACCESSIBILITY_AUDIT.md`, `docs/peripheral-circuits/ARRAY_SIMULATION_FIDELITY.md`, `docs/development/ARCHITECTURE.md`, code comments

---

## Current Focus & Direction

### 1. Module 4 Circuits: Physics Correction (HIGH PRIORITY)

| ID | Task | Status |
|----|------|--------|
| FOCUS-01 | Make READ behavior physically consistent (array-level, not independent cells) | ⏳ |
| FOCUS-02 | Include material-dependent behavior in READ path | ⏳ |
| FOCUS-03 | Include geometry scaling (area/thickness) into resistance/conductance path | ✅ |
| FOCUS-04 | Treat crossbar as full resistor network (not per-cell ideal) | ⏳ |
| FOCUS-05 | Reconcile input voltages and TIA conversion with correct math/signs/end-to-end consistency | ⏳ |

### 2. Module Linkage: Module 1 → Module 4

| ID | Task | Status |
|----|------|--------|
| FOCUS-06 | Ensure hysteresis outputs from Module 1 feed Module 4 correctly | ✅ |
| FOCUS-07 | Keep cell-size/access/conductance dependencies consistent across both modules | ✅ |

**Evidence (2026-02-11):**
- Added cross-module integration tests in `module4-circuits/pkg/gui/module1_module4_integration_test.go` validating Module 1 material outputs (Vc/levels/conductance) propagate into Module 4.
- Fixed `module4-circuits/pkg/gui/device_state.go` ideal compute path to use `levelToConductance(...)`, aligning geometry scaling with coupled path.

### 3. UI Fixes

| ID | Task | Status |
|----|------|--------|
| FOCUS-08 | Improve UI where percentages are too small / poorly ranged | ⏳ |
| FOCUS-09 | Re-range values and layout so output is readable and meaningful | ⏳ |

### 4. Scope Control

- **Skip/defer Module 5** for now to reduce complexity.
- **Focus on Module 4 + integration path** with Module 1.

### 5. CMOS / OpenLane-OpenROAD Path (Module 6 Direction)

| ID | Task | Status |
|----|------|--------|
| FOCUS-10 | Push integration framing for chip-design flow using open-source EDA tools | ⏳ |
| FOCUS-11 | Keep physics assumptions consistent when moving toward schematic/chip flow | ⏳ |

---

## Priority Legend

| Priority | Meaning |
|----------|---------|
| 🔴 **Critical** | Must fix before any public release; blocks core functionality |
| 🟠 **High** | Fix before academic/educational use; significant issues |
| 🟡 **Medium** | Polish and enhancement; improves quality |
| 🟢 **Low** | Nice to have; future enhancements |

## Status Legend

| Symbol | Meaning |
|--------|---------|
| ⏳ | Pending |
| 🔄 | In Progress |
| ✅ | Complete |

---

## 🔴 Critical Priority

### Physics Engine Issues

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| LK-C01 | Verify LK equation terms/signs match compendium (E_eff = E_applied - k_dep·P) | `shared/physics/landau.go` | ✅ | 2hr |
| LK-C02 | Verify effective-viscosity wiring `rho_eff = rho + (R_series·A/d)` | `shared/physics/landau.go` | ✅ | 1hr |
| LK-C03 | Headless LK run: E-field units, 5-target ISPP without NaN/Inf | `cmd/fecim-lattice-tools` | ✅ | 2hr |

### Documentation Accuracy

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| DOC-CITE-1 | Add DOI citations for ELI5 energy, HZO property, data-center projections | `docs/ELI5.md` | ✅ | 1-2hr |
| DOC-CITE-2 | Verify/replace literature DOIs in crossbar voltage/physics references | `docs/crossbar/reference/` | ✅ | 2-4hr |
| DOC-CITE-3 | Cite peripheral timing/energy assumptions or label as placeholders | `docs/peripheral-circuits/PHYSICS.md` | ✅ | 1-2hr |
| DOC-CITE-4 | Cite hysteresis parameter values or label as placeholders | `docs/hysteresis/hysteresis.physics.md` | ✅ | 1-2hr |
| DOC-LINK-1 | Fix broken internal markdown links in docs/ (112 links fixed; docs/README.md prioritized) | `docs/` | ✅ | 2-4hr |

---

## 🟠 High Priority

### Polydomain Landau (Top Priority)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| LK-PD-1 | Define polydomain LK target behavior: verify-at-E=0 must yield 30 stable remanent levels (quantized by level mapping), not just 2 wells | Spec (Juan) | ⏳ | 30-60m |
| LK-PD-2 | Add “remanent staircase sweep” diagnostic: pulse magnitude → (P_rem, level) distribution; require >=20 distinct levels for multilevel claim | `module1-hysteresis/pkg/controller` + `shared/physics` | ⏳ | 1-2hr |
| LK-PD-3 | Implement polydomain LK model (domain ensemble with distributed thresholds/parameters, not just additive bias). Must hold intermediate remanent states at E=0 | `shared/physics/landau.go` | 🔄 | 4-12hr |
| LK-PD-4 | Wire GUI ISPP (Write/Read waveform) to use polydomain LK when engine=LandauK (toggle), keep single-domain for baseline hysteresis unless enabled | `module1-hysteresis/pkg/gui` | ⏳ | 2-4hr |
| LK-PD-5 | ISPP convergence test for polydomain LK: targets {5,10,15,20,25} within <=25 pulses (verify-at-0) | `module1-hysteresis/pkg/controller` | ⏳ | 1-3hr |
| LK-PD-6 | Literature grounding: cite hafnia/HZO polydomain/partial switching or “intermediate state retention” sources; mark any claim as CITATION NEEDED until done | `docs/hysteresis/*` + HONESTY | 🔄 | 2-6hr |

### Engineering Guardrails

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| G01 | Calibration drift guard: `scripts/calib-guard.sh` fails CI on uncommitted calibration JSON changes | `cmd/.../calibrations/` | ✅ | 1-2hr |
| G02 | Intentional calibration update policy: require evidence log links in commits | Process | ✅ | 30m |
| G03 | Provide optional pre-commit hook template that warns on calibration JSON changes | Process | ✅ | 30m |
| G04 | Headless WRD/ISPP regression suite: Preisach HI/MID/LO targets + JSON summary | Shared | ⏳ | 2-4hr |
| G05 | Headless LK regression suite: same targets + overshoot/pulse stats | Shared | ⏳ | 2-4hr |
| G06 | Normalize/verify CLI engine selector (`--engine {preisach,lk}`) | CLI | ⏳ | 30-60m |
| G06b | Verification matrix: Preisach + LK for each material → HI/MID/LO | Testing | ⏳ | 1-2hr |
| G04b | One-source-of-truth ISPP write engine: refactor duplicates to `shared/physics` | `shared/physics` | ⏳ | 4-12hr |
| G04c | Shared ISPP migration plan: define API, adapters, deprecation plan | Architecture | ⏳ | 1-2hr |

### LK Stabilization

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| G07 | LK ISPP overshoot accounting: overshoots/target, max Δ, stuck-breaker count | `shared/physics` | ⏳ | 1-2hr |
| G08 | Define acceptance criteria for Literature Superlattice MID stability | `hysteresis-prompt.md` | ⏳ | 30-60m |
| LK05 | ISPP controller not optimized for L-K dynamics (overshoots near MID) | `module1-hysteresis` | ⏳ | 4-8hr |
| LK07 | Need longer WAIT phases for L-K settling | `module1-hysteresis` | ⏳ | 2-4hr |

### Performance Diagnosis

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| G09 | LK perf evidence script: 3 targets → steps, dt stats, solverMs | `scripts/` | ⏳ | 1-2hr |
| G10 | Add `pprof` toggle for headless hysteresis runs (`FECIM_PPROF=1`) | Debug | ⏳ | 1-2hr |

### GUI Correctness

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| G11 | Throttled WRD phase-boundary logging spec | `module1-hysteresis` | ⏳ | 1hr |
| G11b | Refactor target/phase snapshot wiring: single snapshot struct for widgets | `module1-hysteresis` | ⏳ | 1-2hr |
| G11c | Write Cell ISPP + circuit-coupled updates: DAC→array, neighbor polarization | `module4-circuits` | ⏳ | 4-12hr |
| G12 | GUI parity smoke test checklist: log lines + screenshots | Testing | ⏳ | 30-60m |

### Module-Specific High Priority

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| M1-D1 | Document run modes (GUI/TUI/headless/Vulkan), L-K vs Preisach defaults | `docs/.../module1-hysteresis/` | ✅ | 30-60m |
| M1-U1 | Fix WRD target marker parity (single snapshot for target/marker/logs) | `module1-hysteresis` | ⏳ | 1-2hr |
| M1-U2 | Equation widget perf: cold <1s, warm <200ms, no freeze | `module1-hysteresis` | ⏳ | 30-60m |
| M1-P1 | L-K performance accounting + ISPP stabilization evidence | `module1-hysteresis` | ⏳ | 2-4hr |
| M2-U1 | Align `crossbar-gui -help` with implemented features | `cmd/crossbar-gui` | ⏳ | 30-60m |
| M2-P1 | Full physics audit vs PHYSICS.md (IR drop, sneak, drift, temp) | `module2-crossbar` | ⏳ | 2-4hr |
| M2-P2 | Temperature scalings beyond wire resistance | `module2-crossbar` | ⏳ | 1-2hr |
| M3-D1 | Sync docs with file paths and core vs training split | `docs/.../module3-mnist/` | ✅ | Done (docs/documentation/module3-mnist/FEATURES.md updated with runtime vs training map) |
| M3-D2 | Align noise bounds (docs/UI 0.20 max vs code clamp 0.50) | `module3-mnist` | ✅ | Done (core clamp now 0.20 in `pkg/core/network_config.go`, tests updated) |
| M3-U1 | Audit GUI labels: accuracy/energy labeled as modeled (not verified) | `module3-mnist` | ✅ | Done (`dualmode.go`, `app.go`, `metrics.go` labels switched to modeled wording) |
| M3-P1 | Verify FP vs CIM inference pipeline + quantization/noise injection | `module3-mnist` | ⏳ | 2-4hr |
| M3-P2 | Align energy model between core and GUI widgets | `module3-mnist` | ⏳ | 1-2hr |
| M3-U2 | Decide dual-mode confusion matrix/metrics exposure | `module3-mnist` | ⏳ | 1-2hr |
| M4-D1 | Update docs to reference `shared/peripherals` everywhere | `docs/.../module4-circuits/` | ✅ | Done (`docs/documentation/module4-circuits/FEATURES.md` explicitly marks `shared/peripherals` as canonical, adds `chargepump.go`) |
| M4-U1 | Validate ISPP engine toggle wiring (Fast vs L-K) | `module4-circuits` | ✅ | Done (`tab_unified_voltage.go` routes by `GetISPPEngine()` and selector writes via `SetISPPEngine`; `tab_unified_extended_test.go` now asserts selector->state sync) |
| M4-U3 | Sense-chain UI: TIA output, ADC code/saturation, measurement presets | `module4-circuits` | ✅ | 1-2hr |
| M4-P1 | Audit DAC/ADC/TIA/ChargePump equations vs docs | `module4-circuits` | ✅ | Done (`docs/documentation/module4-circuits/PHYSICS.md` equations aligned to `shared/peripherals/{dac,adc,tia,chargepump}.go`) |
| M4-P3 | Define/centralize cell geometry (area, thickness, stack) | `module4-circuits` | ✅ | 1-2hr |
| M4-P4 | **Tier B DC solver** (full resistive network) + regression tests | `module4-circuits/pkg/arraysim` | ⏳ | 4-12hr |
| M4-U2d | Tests/visual checks for half-select disturb + DAC voltage display | `module4-circuits` | ⏳ | 1-2hr |

### Tier B Array Simulation (from code TODOs)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| TIERB-1 | Replace dense reference solver with scalable sparse/iterative solver | `module4-circuits/pkg/arraysim/tier_b.go:11` | ⏳ | 4-8hr |
| TIERB-2 | Add realistic boundary conditions and selector devices | `module4-circuits/pkg/arraysim/tier_b.go:12` | ⏳ | 2-4hr |
| TIERB-3 | Validate against SPICE golden vectors | `module4-circuits/pkg/arraysim/tier_b.go:13` | ⏳ | 4-8hr |
| TIERB-4 | Revisit boundary conditions to match SPICE conventions | `module4-circuits/pkg/arraysim/refsolve_dense.go:20` | ⏳ | 2-4hr |

### Citations Pending

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| H03 | Voltage range citations (thickness-dependent) | `drtour_todo_fixes.md` | ⏳ | 1hr |
| H04 | Read parameter sources - mark as empirical | `drtour_todo_fixes.md` | ⏳ | 1hr |
| H13 | GPU comparison nuance - add batched operation context | `drtour_todo_fixes.md` | ⏳ | 1hr |

---

## 🟡 Medium Priority

### Module 6 & 7

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| M6-D1 | Sync docs with actual exports (JSON/CSV/SPICE/Verilog/DEF/LEF/Liberty/SVG) | `docs/.../module6-eda/` | ⏳ | 1-2hr |
| M6-U1 | Check GUI/CLI parity (Start/Stop, defaults) | `module6-eda` | ⏳ | 1-2hr |
| M6-P1 | Audit mapping/quantization/topology vs docs | `module6-eda` | ⏳ | 2-4hr |
| M7-D1 | Confirm curriculum tree order + shortcuts match docs | `module7-docs` | ⏳ | 30-60m |
| M7-U1 | Validate layout breakpoints + click targets | `module7-docs` | ⏳ | 1-2hr |
| M7-U2 | Add colored category badges in tree rows | `module7-docs` | ⏳ | 30-60m |
| M7-U3 | Hide "On This Page" sidebar when ToC < 3 headings | `module7-docs` | ⏳ | 30-60m |
| M7-P1 | Verify search ranking + reading time math | `module7-docs` | ⏳ | 1-2hr |

### Cross-Module

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| CM-P1 | Define "physics accuracy" acceptance criteria per module | Shared | ⏳ | 30-60m |
| CM-D1 | Keep HONESTY_AUDIT.md as SSOT; ensure UI labels match | Shared | ⏳ | 30-60m |
| CM-U1 | Ensure UI values/plots never desync from engine state | Shared | ⏳ | 1-2hr |
| CM-D2 | Equation widgets pipeline: LaTeX→SVG SSOT, hotspot alignment | Shared | ⏳ | 1-2hr |
| CM-P2 | Minimal headless regression suite per engine with JSON summary | Shared | ⏳ | 2-4hr |
| CM-D3 | Tighten module docs: equations, assumptions, units, validated labels | Shared | ⏳ | 2-4hr |

### UX Polish

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| G13 | Define minimum supported GUI size (1024×768) | UX | ⏳ | 30-60m |
| G14 | GUI overlap audit: fix widget overlap/clipping on resize | UX | ⏳ | 1-2hr |
| G15 | Update GUI layout docs to match current code | `docs/development/GUI/` | ⏳ | 1-2hr |
| G16 | Documentation mapping sweep: audit docs for drift vs code | `docs/development/GUI/` | ⏳ | 2-4hr |

### Array Simulation Fidelity (from docs)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| ASIM-1 | Add explicit "fidelity tier" selector to GUI | `docs/peripheral-circuits/ARRAY_SIMULATION_FIDELITY.md` | ⏳ | 2-4hr |
| ASIM-2 | Add DC nodal solver for passive sneak paths | `docs/peripheral-circuits/ARRAY_SIMULATION_FIDELITY.md` | ⏳ | 4-8hr |
| ASIM-3 | Implement 2T1R masks | `docs/peripheral-circuits/ARRAY_SIMULATION_FIDELITY.md` | ⏳ | 2-4hr |
| ASIM-4 | Add headless test suite for coupling + tiers | `docs/peripheral-circuits/ARRAY_SIMULATION_FIDELITY.md` | ⏳ | 2-4hr |

### Peripheral Circuits Enhancements

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| PERIPH-1 | Export functionality (diagrams/data) | `docs/peripheral-circuits/circuits.operations.md` | ⏳ | 2-4hr |
| PERIPH-2 | Temperature-dependent INL/DNL model | `docs/peripheral-circuits/circuits.operations.md` | ⏳ | 2-4hr |
| PERIPH-3 | Fast/slow/typical process corner analysis | `docs/peripheral-circuits/circuits.operations.md` | ⏳ | 4-8hr |
| PERIPH-4 | Write-verify animation (iterative cycle) | `docs/peripheral-circuits/circuits.operations.md` | ⏳ | 2-4hr |
| PERIPH-5 | Sneak path quantification display | `docs/peripheral-circuits/circuits.operations.md` | ⏳ | 1-2hr |

### Accessibility (from audit)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| A11Y-1 | Increase font sizes below 14px to minimum | `docs/ACCESSIBILITY_AUDIT.md` | ✅ | 1-2hr |
| A11Y-2 | Wire up FocusIndicator to interactive widgets | `shared/widgets/accessibility.go` | ⏳ | 2-4hr |
| A11Y-3 | Expose HighContrastTheme via settings menu | Settings | ⏳ | 1-2hr |
| A11Y-4 | Show KeyboardNavigationHelp via F1 key | Settings | ✅ | 30-60m |
| A11Y-5 | Add Tab order to launcher demo cards | Launcher | ⏳ | 1-2hr |
| A11Y-6 | Arrow key navigation in data widgets | Widgets | ⏳ | 2-4hr |

---

## 🟢 Low Priority

### Vulkan Rendering (from code TODOs)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| VK-1 | Implement actual Vulkan calls using go-vk or vgpu | `module1-hysteresis/pkg/render/render.go:303` | ⏳ | 16-24hr |
| VK-2 | Implement actual Vulkan initialization | `module1-hysteresis/pkg/render/render.go:351` | ⏳ | 4-8hr |
| VK-3 | Implement actual render loop | `module1-hysteresis/pkg/render/render.go:365` | ⏳ | 8-12hr |
| VK-4 | Release Vulkan resources properly | `module1-hysteresis/pkg/render/render.go:388` | ⏳ | 1-2hr |

### Platform Extensions

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| L07 | Demo video creation (2-3 min walkthrough) | TODO.md | ⏳ | 4hr |
| L08 | Web deployment (WASM) for browser-based demos | TODO.md | ⏳ | 16hr |
| L09 | Vulkan rendering implementation for large arrays | TODO.md | ⏳ | 20hr |
| L10 | 3D multi-layer visualization (512-layer roadmap) | TODO.md | ⏳ | 24hr |
| L11 | Add [LK] indicators to material_picker.go | `module1-hysteresis` | ⏳ | 1hr |
| L05 | "About the Science" unified Learn More section | `drtour_todo_fixes.md` | ⏳ | 2hr |

### Architecture Improvements (from ARCHITECTURE.md)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| ARCH-1 | Module 6 (EDA): Complete placement algorithm | `docs/development/ARCHITECTURE.md` | ⏳ | 8-16hr |
| ARCH-2 | Multi-cell arrays in Module 1 | `docs/development/ARCHITECTURE.md` | ⏳ | 4-8hr |
| ARCH-3 | Advanced MVM sneak path current tracing visualization | `docs/development/ARCHITECTURE.md` | ⏳ | 4-8hr |
| ARCH-4 | Custom neural network training in Module 3 | `docs/development/ARCHITECTURE.md` | ⏳ | 8-16hr |
| ARCH-5 | More chip peripheral types in Module 4 | `docs/development/ARCHITECTURE.md` | ⏳ | 4-8hr |
| ARCH-6 | Behavioral model export (SPICE) | `docs/development/ARCHITECTURE.md` | ⏳ | 8-16hr |
| ARCH-7 | EDA routing algorithm completion | `docs/development/ARCHITECTURE.md` | ⏳ | 8-16hr |

### Accessibility Phase 3 (Enhancements)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| A11Y-7 | Text alternatives for all visualizations | `docs/ACCESSIBILITY_AUDIT.md` | ⏳ | 4-8hr |
| A11Y-8 | Accessible data export (CSV, HTML) | `docs/ACCESSIBILITY_AUDIT.md` | ⏳ | 2-4hr |
| A11Y-9 | Large text mode option | `docs/ACCESSIBILITY_AUDIT.md` | ⏳ | 2-4hr |
| A11Y-10 | Reduced motion preference | `docs/ACCESSIBILITY_AUDIT.md` | ⏳ | 1-2hr |

### Sky130 PDK (from docs)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| SKY-1 | Add Apache 2.0 LICENSE.txt for PDK | `docs/eda/pdk/sky130.md:238` | ⏳ | 15m |

---

## Completed Items (Recent)

### MNIST Module (from mnist.fixes.todo.md) ✅

All 46 items complete:
- 3 Critical issues (nil pointer fixes)
- 9 High priority (race conditions, error handling)
- 13 Medium priority (code cleanup, validation)
- 6 Low priority (naming, documentation)
- 2 Security issues (type assertions, bounds checks)
- 5 Architecture items (interfaces, extraction)
- 4 Documentation items
- 4 Test coverage items

### Critical Fixes ✅

- C01-C13: All simulation banners, disclaimers, physics parameters complete

### High Priority Fixes ✅

- H01, H02, H05, H06, H07, H08, H09, H10, H11, H12, H14, H15, H16: Complete

### Medium Priority Fixes ✅

- M01-M16: All polish items complete

### Accessibility ✅

- Color contrast fixes in canvas.go, metrics.go, activations.go
- DigitCanvas keyboard navigation (Arrow keys + Space/Enter)
- Accessibility helpers infrastructure

---

## Progress Summary

| Priority | Total | Complete | Remaining |
|----------|-------|----------|-----------|
| **Current Focus** | **11** | **0** | **11** |
| 🔴 Critical | 8 | 0 | 8 |
| 🟠 High | 48 | 0 | 48 |
| 🟡 Medium | 32 | 0 | 32 |
| 🟢 Low | 22 | 0 | 22 |
| **Total** | **121** | **0** | **121** |

*Note: "Current Focus" items (FOCUS-01 through FOCUS-11) are the active work direction. Module 5 is deferred.*

---

## Quarterly Literature Review

**Status**: Scheduled | **Due**: April 2026 | **Priority**: Medium

**Goal**: Update HONESTY_AUDIT.md with 2026 Q1 publications.

**Search databases**: IEEE Xplore (IEDM, ISSCC, VLSI), Nature family, ACS, arXiv

---

## Calibration JSON Policy

Calibration baselines in `cmd/fecim-lattice-tools/data/calibrations/*.json` are tracked. To prevent accidental commits:
```bash
git update-index --assume-unchanged cmd/fecim-lattice-tools/data/calibrations/literature_superlattice.json
```

**Policy**: Do **not** commit calibration JSON changes unless intentionally updating baseline + evidence logs.

---

## Deferred

| Item | Reason |
|------|--------|
| Module 5 (Comparison) | Deferred to reduce complexity; focus on Module 4 + integration path |

## Out of Scope

| Item | Reason |
|------|--------|
| Production chip design tools | Educational tool, not EDA replacement |
| Hardware-accurate SPICE models | Requires proprietary foundry PDKs |
| Real-time OS integration | Beyond educational scope |
| Web-based collaboration | Single-user educational tool |
| Investor pitch decks | Scientific tool, not marketing material |
| Cryptographic accelerators | Specialized application |

---

## Agent Work Policy

Any agent tackling a task from this TODO **must**:

1. **Work fully autonomously** — complete the task end-to-end without stopping for manual intervention.
2. **Validate progress continuously** — run `go test ./...` (headless) or launch the GUI to verify changes work. Never claim "done" without fresh test/build evidence.
3. **Update this TODO.md** — mark completed items as ✅, add any new tasks discovered during implementation, and update the progress summary.

---

## Contributing

See `CONTRIBUTING.md` and `CLAUDE.md` for development guidelines.

**Scientific accuracy**: All claims must be verified per `HONESTY_AUDIT.md` standards.

---

*This TODO prioritizes scientific rigor and educational honesty over promotional considerations.*
*Document consolidated: 2026-02-07 | Refocused: 2026-02-11*
