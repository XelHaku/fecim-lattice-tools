# FeCIM Lattice Tools - Comprehensive TODO

**Mission**: Educational FeCIM visualization and simulation tool based on HfO₂-ZrO₂ superlattice research.

**Last Updated**: 2026-02-11 (Refocused priorities)

**Source Documents**: `CRITIQUE_MASTER_LIST.md`, `docs/neural-network/mnist.fixes.todo.md`, `docs/ACCESSIBILITY_AUDIT.md`, `docs/peripheral-circuits/ARRAY_SIMULATION_FIDELITY.md`, `docs/development/ARCHITECTURE.md`, code comments

---

## Current Focus & Direction

### 1. Module 4 Circuits: Physics Correction (HIGH PRIORITY)

| ID | Task | Status |
|----|------|--------|
| FOCUS-01 | Make READ behavior physically consistent (array-level, not independent cells) | ✅ |
| FOCUS-02 | Include material-dependent behavior in READ path | ✅ |
| FOCUS-03 | Include geometry scaling (area/thickness) into resistance/conductance path | ✅ |
| FOCUS-04 | Treat crossbar as full resistor network (not per-cell ideal) | ✅ |
| FOCUS-05 | Reconcile input voltages and TIA conversion with correct math/signs/end-to-end consistency | ✅ |

### 2. Module Linkage: Module 1 → Module 4

| ID | Task | Status |
|----|------|--------|
| FOCUS-06 | Ensure hysteresis outputs from Module 1 feed Module 4 correctly | ✅ |
| FOCUS-07 | Keep cell-size/access/conductance dependencies consistent across both modules | ✅ |

**Evidence (2026-02-11):**
- Added cross-module integration tests in `module4-circuits/pkg/gui/module1_module4_integration_test.go` validating Module 1 material outputs (Vc/levels/conductance) propagate into Module 4.
- Fixed `module4-circuits/pkg/gui/device_state.go` ideal compute path to use `levelToConductance(...)`, aligning geometry scaling with coupled path.
- FOCUS-01: `NewDeviceState(...)` now defaults coupling mode to `CouplingTierA`, so READ path uses coupled array-level simulation by default instead of independent-cell ideal math.
- FOCUS-02: READ conductance mapping now resolves quantization via material-native levels (`resolveConductanceLevels`), and READ current changes with material selection are covered by tests.
- FOCUS-04: `module4-circuits/pkg/arraysim/tier_a.go` now solves READ coupling through the full WL/BL resistive network via dense DC nodal solve (`referenceSolveDense`), eliminating Tier-A per-cell ideal approximation.
- Added/strengthened Tier-A network tests in `module4-circuits/pkg/arraysim/tier_a_test.go`:
  - `TestTierA_MatchesDenseReferenceSolve` (Tier-A result equality vs full nodal reference)
  - Updated passive half-select + active-row masking assertions for coupled-network behavior
- Added/strengthened tests in `module4-circuits/pkg/gui/device_state_read_coupling_test.go`:
  - `TestReadCoupling_DefaultsToTierA`
  - `TestReadCoupling_MaterialSelectionChangesReadCurrent`
  - Existing signed per-cell READ coupling test retained.
  - New `TestReadChain_EndToEndKnownConductanceToADCCode` (1x1 known conductance, ±DAC voltage polarity, checks DAC→array current→TIA output→ADC code exact consistency).
- Reconciled sign math in ideal compute path (`module4-circuits/pkg/gui/device_state.go`): row current now uses `I = G × V` (signed), matching coupled solver conventions and sense-chain polarity.
- Verification commands:
  - `go test ./module4-circuits/pkg/gui -run "Test(ReadCoupling_SignedPerCellVI|ReadCoupling_DefaultsToTierA|ReadCoupling_MaterialSelectionChangesReadCurrent|ReadChain_EndToEndKnownConductanceToADCCode)" -count=1 -v` (PASS)
  - `go test -race ./module4-circuits/pkg/gui -run "Test(ReadCoupling_SignedPerCellVI|ReadCoupling_DefaultsToTierA|ReadCoupling_MaterialSelectionChangesReadCurrent|ReadChain_EndToEndKnownConductanceToADCCode)" -count=1` (PASS)
  - `go test -race ./...` currently blocked by pre-existing unrelated compile failure in `module1-hysteresis/pkg/gui/equation_dialog_test.go` (`ShowPhysicsEquationsDialog` vs `showPhysicsEquationsDialog`).
- FOCUS-31: `shared/widgets/notification.go` toast renderer now derives layout spacing/sizes from Fyne theme metrics (`SizeNameInnerPadding`, `SizeNameInlineIcon`, `SizeNamePadding`) instead of fixed `12/20/24`, making toast layout DPI/theme-scale aware.
- FOCUS-32: `shared/theme/theme.go` now honors `variant` in `FeCIMTheme.Color()` with distinct light/dark palette outputs. Added regression test `TestFeCIMTheme_VariantAwareColors` in `shared/theme/theme_test.go`.
- Verification (FOCUS-31/32): `go test ./shared/theme`; `go test -race ./shared/theme ./shared/widgets -run TestFeCIMTheme_VariantAwareColors -count=1`; `go test -race ./shared/widgets -run TestNotificationType_String -count=1` (PASS).
- FOCUS-34: `shared/widgets/debug.go` now bounds layout debug maps with `maxTrackedLayoutWidgets=1024` and periodic cleanup (`layoutCleanupInterval=256`) to prevent unbounded growth of `layoutCallCounts`/`lastLayoutTime`.
- FOCUS-35: `shared/widgets/debug.go` debug prints (`[LAYOUT]`, `[RESIZE]`, `[RESIZE-BUG]`, `[INTERACTION]`) were migrated from `fmt.Printf` to `shared/logging.Printf` so debug output flows through the project logging system.

### 3. UI Fixes

| ID | Task | Status |
|----|------|--------|
| FOCUS-08 | Improve UI where percentages are too small / poorly ranged | ⏳ |
| FOCUS-09 | Re-range values and layout so output is readable and meaningful | ⏳ |
| FOCUS-31 | Toast/notification layout uses magic numbers (padding=12, icon=20, close=24) — not DPI-aware | ✅ |
| FOCUS-32 | Theme has no dark/light mode variants — `FeCIMTheme.Color()` ignores variant parameter | ✅ |
| FOCUS-33 | Screen reader `Announce()` and `SetAccessibleLabel()` are no-ops — placeholder only | ⏳ |
| FOCUS-34 | Debug layout tracker uses unbounded maps (`layoutCallCounts`, `lastLayoutTime`) — memory leak risk | ✅ |
| FOCUS-35 | Debug output goes to `fmt.Printf` (stdout) instead of logging system | ✅ |

### 3b. Module 3 MNIST Consistency

| ID | Task | Status |
|----|------|--------|
| FOCUS-36 | CIM forward pass is purely semantic (delegates to FP) — conductance mapping Gmin/Gmax only in comments | ⏳ |
| FOCUS-37 | DAC quantization assumes input [0,1] but never validates — silent clamp | ✅ (2026-02-11: added invalid-range validation + clamp warning in `quantizeDAC`) |
| FOCUS-38 | Silent fallback to CPU on GPU error with no user notification | ✅ (2026-02-11: emit user notice on GPU→CPU fallback in `forwardFP`) |
| FOCUS-39 | Silent fallback to default weights if level-specific file missing — user not warned | ✅ (2026-02-11: controller now warns when loading default weights due to missing level-specific file) |
| FOCUS-40 | ADC dialog says "6-bit (64 levels)" but code defaults to 8-bit — mismatch | ✅ (2026-02-11: dialog text reconciled to 8-bit default / finite-resolution wording) |
| FOCUS-41 | `SetNumLevels` silently clamps values — user sets 50, gets 31 with no feedback | ✅ (2026-02-11: emit user notice with actual clamped level) |

### 3c. CLI & Configuration

| ID | Task | Status |
|----|------|--------|
| FOCUS-42 | Recent Files menu TODO — clicking doesn't load file (`main.go:1228`) | ⏳ |
| FOCUS-43 | 9 undocumented env vars (FECIM_MATERIAL, FECIM_RANGE_FRAC, etc.) — add to `--help` output | ✅ (2026-02-11: `cmd/fecim-lattice-tools --help` now prints dedicated headless env var section listing all 9 vars) |
| FOCUS-44 | Screenshots/recordings dirs hardcoded to `screenshots/` and `recordings/` — no CLI override | ✅ (2026-02-11: added `--screenshot-dir` and `--recording-dir` flags; capture paths now configurable) |
| FOCUS-45 | Config search only uses relative paths — no XDG_CONFIG_HOME or `~/.config/fecim/` support | ✅ (2026-02-11: `shared/cli.ConfigLoader` now resolves via `$XDG_CONFIG_HOME/fecim` then `~/.config/fecim`) |

**Evidence (FOCUS-43/44/45, 2026-02-11):**
- `cmd/fecim-lattice-tools/main.go`: added custom `flag.Usage` section documenting 9 headless env vars (`FECIM_MATERIAL`, `FECIM_RANGE_FRAC`, `FECIM_ISPP_STEPS_PER_PULSE`, `FECIM_HEADLESS_FAST`, `FECIM_ISPP_TARGETS`, `FECIM_ISPP_TARGET_SEED`, `FECIM_ISPP_TARGET_LEVELS`, `FECIM_ISPP_MAX_PULSES`, `FECIM_HEADLESS_ALLOW_TIMEOUT`).
- `cmd/fecim-lattice-tools/main.go`: added `--screenshot-dir` and `--recording-dir`; replaced hardcoded `screenshots/` and `recordings/` outputs with flag-driven directories.
- `shared/cli/cli.go`: added config path resolution with XDG/home search roots (`$XDG_CONFIG_HOME/fecim`, `$HOME/.config/fecim`) plus `~/` expansion.
- `shared/cli/cli_test.go`: added path-resolution tests for XDG and home config fallback.

### 3d. Error Handling (panic → graceful)

| ID | Task | Status |
|----|------|--------|
| FOCUS-46 | GPU peripherals `structToBytes` panics on unknown type — should return error (`gpu_peripherals.go:382`) | ✅ |
| FOCUS-47 | GPU peripherals size mismatch panics — should return error (`gpu_peripherals.go:506`) | ✅ |
| FOCUS-48 | Physics config init panics on missing YAML — should use `log.Fatal` or return error (`physics.go:432`) | ✅ |

**Evidence (FOCUS-46/47/48, 2026-02-11):**
- `module4-circuits/pkg/gpuperiph/gpu_peripherals.go`: `structToBytes` now returns `([]byte, error)`; unknown struct types return `error` (no panic).
- `module4-circuits/pkg/gpuperiph/gpu_peripherals.go`: runtime layout check moved to `validateGPUPeripheralStructLayout() error`; `NewGPUPeripherals()` now returns wrapped error on mismatch instead of panicking.
- `config/physics/physics.go`: `MustLoad()` now uses `log.Fatalf(...)` (no panic path).
- Added tests in `module4-circuits/pkg/gpuperiph/gpu_peripherals_test.go` for unsupported type error + supported type success + layout validation.

### 3e. Module 1 Hysteresis (from hysteresis-prompt.md)

| ID | Task | Status |
|----|------|--------|
| FOCUS-49 | L-K performance: quantify why slow — dtNominal too small, 21k-221k solver steps/target, math-bound | ⏳ |
| FOCUS-50 | Frankenstein equation fidelity: verify all terms/signs/units match `hysteresis-gemini.md` formulation | ⏳ |
| FOCUS-51 | Target/marker parity: GUI yellow target must match active controller target (no early jump to next) | ⏳ |
| FOCUS-52 | Headless Preisach WRD/ISPP parity with GUI — run headless to debug target/marker mismatches | ⏳ |
| FOCUS-53 | Physics equations UI: keep labels/links coherent across L-K, Preisach, and ISPP tabs | ⏳ |

### 3f. Module 2 Crossbar (from module2-prompt.md)

| ID | Task | Status |
|----|------|--------|
| FOCUS-54 | Verify conductance models (linear, exponential, lookup) and quantization to 30 levels match docs | ⏳ |
| FOCUS-55 | Validate MVM/VMM equations, Ohm's law, DAC/ADC quantization, output normalization vs PHYSICS.md | ⏳ |
| FOCUS-56 | Confirm IR drop solver (wire params, iterative relaxation, effective voltage) matches docs | ⏳ |
| FOCUS-57 | Confirm sneak path modeling (3-cell paths, simplified vs full) and SNR math | ⏳ |
| FOCUS-58 | Validate drift models (log/power-law), temperature effects (Arrhenius), and variation | ⏳ |
| FOCUS-59 | Verify endurance/fatigue and half-select disturb behavior if enabled | ⏳ |
| FOCUS-60 | Ensure MVMWithNonIdealities pipeline ordering matches documented signal flow | ⏳ |

### 3g. Module 3 MNIST (from module3-prompt.md)

| ID | Task | Status |
|----|------|--------|
| FOCUS-61 | Verify FP path math: linear layers, ReLU, softmax, normalization, output probabilities | ⏳ |
| FOCUS-62 | Validate CIM path: weight quantization to N levels, DAC/ADC quantization, noise injection order | ⏳ |
| FOCUS-63 | Confirm disagreement metrics (KL divergence), accuracy tracking, confusion matrix logic | ⏳ |
| FOCUS-64 | Verify energy/performance models in GUI match documented formulas and defaults | ⏳ |
| FOCUS-65 | Validate MNIST IDX parsing, bounds checks, and sanity limits for dataset sizes | ⏳ |
| FOCUS-66 | Verify weight file loading, QAT level selection, and fallback behavior — document silent fallbacks | ⏳ |

### 3h. Module 6 EDA (from module6-prompt.md)

| ID | Task | Status |
|----|------|--------|
| FOCUS-67 | Verify ArrayConfig/CellConfig defaults (rows, cols, levels, gmin/gmax, vdd, tech, architecture) | ⏳ |
| FOCUS-68 | Validate storage/memory/compute mode behavior and mode-specific parameters | ⏳ |
| FOCUS-69 | Confirm weight mapping and quantization including sign handling | ⏳ |
| FOCUS-70 | Validate export format correctness: JSON/CSV/SPICE/Verilog/DEF contents and indexing | ⏳ |
| FOCUS-71 | Ensure CLI and GUI flows produce equivalent outputs given same configuration | ⏳ |

### 3i. Documentation Curriculum (from documentation-prompt.md)

| ID | Task | Status |
|----|------|--------|
| FOCUS-72 | Ensure `docs/documentation/` has complete curriculum: ELI5/PHYSICS/FEATURES/OPENSOURCE-TOOLS per module | ⏳ |
| FOCUS-73 | Module 7 sidebar order: module folders first, then research-papers, then README/MODULES | ⏳ |
| FOCUS-74 | Content standards: distinguish demonstrated vs modeled vs aspirational in all docs | ⏳ |

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
| G06 | Normalize/verify CLI engine selector (`--engine {preisach,lk}`) | CLI | ✅ | 30-60m |
| G06b | Verification matrix: Preisach + LK for each material → HI/MID/LO | Testing | ✅ | 1-2hr |
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
| M2-U1 | Align `crossbar-gui -help` with implemented features | `cmd/crossbar-gui` | ✅ | 30-60m |
| M2-P1 | Full physics audit vs PHYSICS.md (IR drop, sneak, drift, temp) | `module2-crossbar` | ✅ | 2-4hr |
| M2-P2 | Temperature scalings beyond wire resistance | `module2-crossbar` | ✅ | 1-2hr |
| M3-D1 | Sync docs with file paths and core vs training split | `docs/.../module3-mnist/` | ✅ | Done (docs/documentation/module3-mnist/FEATURES.md updated with runtime vs training map) |
| M3-D2 | Align noise bounds (docs/UI 0.20 max vs code clamp 0.50) | `module3-mnist` | ✅ | Done (core clamp now 0.20 in `pkg/core/network_config.go`, tests updated) |
| M3-U1 | Audit GUI labels: accuracy/energy labeled as modeled (not verified) | `module3-mnist` | ✅ | Done (`dualmode.go`, `app.go`, `metrics.go` labels switched to modeled wording) |
| M3-P1 | Verify FP vs CIM inference pipeline + quantization/noise injection | `module3-mnist` | ✅ | Done (`pkg/core/dualmode_metrics_test.go::TestInfer_CIMOrder_ADCBeforeNoise` locks CIM order as DAC→MVM→ADC→noise→softmax) |
| M3-P2 | Align energy model between core and GUI widgets | `module3-mnist` | ✅ | Done (`pkg/gui/energy_widget_test.go` verifies GUI widget uses `core.EstimateInferenceEnergyJ` + shared MAC counts, incl. single-layer mode) |
| M3-U2 | Decide dual-mode confusion matrix/metrics exposure | `module3-mnist` | ✅ | Done (exposed FP+CIM confusion matrices and per-class metrics in core eval; CLI now prints both modes) |
| M4-D1 | Update docs to reference `shared/peripherals` everywhere | `docs/.../module4-circuits/` | ✅ | Done (`docs/documentation/module4-circuits/FEATURES.md` explicitly marks `shared/peripherals` as canonical, adds `chargepump.go`) |
| M4-U1 | Validate ISPP engine toggle wiring (Fast vs L-K) | `module4-circuits` | ✅ | Done (`tab_unified_voltage.go` routes by `GetISPPEngine()` and selector writes via `SetISPPEngine`; `tab_unified_extended_test.go` now asserts selector->state sync) |
| M4-U3 | Sense-chain UI: TIA output, ADC code/saturation, measurement presets | `module4-circuits` | ✅ | 1-2hr |
| M4-P1 | Audit DAC/ADC/TIA/ChargePump equations vs docs | `module4-circuits` | ✅ | Done (`docs/documentation/module4-circuits/PHYSICS.md` equations aligned to `shared/peripherals/{dac,adc,tia,chargepump}.go`) |
| M4-P3 | Define/centralize cell geometry (area, thickness, stack) | `module4-circuits` | ✅ | 1-2hr |
| M4-P4 | **Tier B DC solver** (full resistive network) + regression tests | `module4-circuits/pkg/arraysim` | ⏳ | 4-12hr |
| M4-U2d | Tests/visual checks for half-select disturb + DAC voltage display | `module4-circuits` | ✅ | Done (`tab_unified_halfselect_voltage_test.go`: verifies V/2 indicator text + overlay colors, disturb change reporting count, and ISPP status DAC voltage/code display) |

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
| H03 | Voltage range citations (thickness-dependent) | `drtour_todo_fixes.md` | ✅ | Done (`module4-circuits/pkg/gui/tab_reference_voltage.go`: added thickness-dependent Ec note + sub-1V@~3.6nm citation context) |
| H04 | Read parameter sources - mark as empirical | `drtour_todo_fixes.md` | ✅ | Done (`module4-circuits/pkg/gui/tab_reference_voltage.go`: read thresholds labeled empirical/assumed simulator guardrails) |
| H13 | GPU comparison nuance - add batched operation context | `drtour_todo_fixes.md` | ✅ | Done (`module4-circuits/pkg/gui/tab_comparison.go`: per-op vs batched throughput caveats in header/table/status) |

---

## 🟡 Medium Priority

### Module 6 & 7

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| M6-D1 | Sync docs with actual exports (JSON/CSV/SPICE/Verilog/DEF/LEF/Liberty/SVG) | `docs/.../module6-eda/` | ✅ | Done (`module6-eda/README.md`, `docs/documentation/module6-eda/FEATURES.md`: export coverage clarified by surface: CLI vs GUI/API) |
| M6-U1 | Check GUI/CLI parity (Start/Stop, defaults) | `module6-eda` | ✅ | Done (documented parity matrix: CLI defaults `compute 128x128`, GUI defaults `storage 4x4`; Start/Stop no background workers in embedded app) |
| M6-P1 | Audit mapping/quantization/topology vs docs | `module6-eda` | ⏳ | 2-4hr |
| M7-D1 | Confirm curriculum tree order + shortcuts match docs | `module7-docs` | ✅ | Done (`docs_test.go`: `TestEmbeddedDocsApp_SortEntries_*`, `TestModuleShortcutsPanel_MappingAndDisableState`) |
| M7-U1 | Validate layout breakpoints + click targets | `module7-docs` | ✅ | Done (`docs_test.go`: breakpoint coverage + `TestEmbeddedDocsApp_TreeClickTargets` for folder/file row behavior) |
| M7-U2 | Add colored category badges in tree rows | `module7-docs` | ✅ | Done (`embedded.go`: centralized `treeCategory` mapping + tree row badge rendering; `docs_test.go`: `TestEmbeddedDocsApp_TreeCategoryBadges`) |
| M7-U3 | Hide "On This Page" sidebar when ToC < 3 headings | `module7-docs` | ✅ | Done (`layout.go`: `SetTocVisible`; `embedded.go`: auto-toggle after `ParseMarkdown`; `docs_test.go`: `TestEmbeddedDocsApp_LoadDocument_TocVisibility`) |
| M7-P1 | Verify search ranking + reading time math | `module7-docs` | ✅ | Done (`search.go`: IDF floor fix for common terms; `docs_test.go`: `TestRankResults`, `TestExtractMetadata_ReadingTimeMath`) |

Evidence note (2026-02-11): `go test -race ./module6-eda/... ./module7-docs/...` passed after docs sync + new module7 curriculum/layout interaction tests.

### Cross-Module

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| CM-P1 | Define "physics accuracy" acceptance criteria per module | Shared | ✅ | 30-60m |
| CM-D1 | Keep HONESTY_AUDIT.md as SSOT; ensure UI labels match | Shared | ✅ | 30-60m |
| CM-U1 | Ensure UI values/plots never desync from engine state | Shared | ⏳ | 1-2hr |
| CM-D2 | Equation widgets pipeline: LaTeX→SVG SSOT, hotspot alignment | Shared | ⏳ | 1-2hr |
| CM-P2 | Minimal headless regression suite per engine with JSON summary | Shared | ⏳ | 2-4hr |
| CM-D3 | Tighten module docs: equations, assumptions, units, validated labels | Shared | ⏳ | 2-4hr |

### UX Polish

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| G13 | Define minimum supported GUI size (1024×768) | UX | ✅ | 30-60m |
| G14 | GUI overlap audit: fix widget overlap/clipping on resize | UX | ⏳ | 1-2hr |
| G15 | Update GUI layout docs to match current code | `docs/development/GUI/` | ⏳ | 1-2hr |
| G16 | Documentation mapping sweep: audit docs for drift vs code | `docs/development/GUI/` | ⏳ | 2-4hr |

**Evidence (CM-P1 / CM-D1 / G13, 2026-02-11):**
- Added cross-module acceptance criteria doc: `docs/development/PHYSICS_ACCEPTANCE_CRITERIA.md`.
- UI honesty labels aligned to SSOT language in:
  - `shared/widgets/about_science.go`
  - `shared/widgets/glossary.go`
- Fixed HONESTY audit local-link path to `docs/comparison/HONESTY_AUDIT.md`.
- Defined and enforced minimum supported GUI size in code:
  - `cmd/fecim-lattice-tools/main.go` (`minWindowWidth=1024`, `minWindowHeight=768`).
- Documented GUI minimum in `docs/development/GUI_MINIMUMS.md` and linked from `README.md`.
- Validation commands:
  - `go test ./shared/widgets -run TestQuickTermLookup -count=1` (PASS)
  - `go test ./cmd/fecim-lattice-tools -run TestMainWindow_.* -count=1` (PASS; no tests matched, package build succeeded)

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
| A11Y-2 | Wire up FocusIndicator to interactive widgets | `shared/widgets/accessibility.go` | ✅ | 2-4hr |
| A11Y-3 | Expose HighContrastTheme via settings menu | Settings | ✅ | 1-2hr |
| A11Y-4 | Show KeyboardNavigationHelp via F1 key | Settings | ✅ | 30-60m |
| A11Y-5 | Add Tab order to launcher demo cards | Launcher | ✅ | 1-2hr |
| A11Y-6 | Arrow key navigation in data widgets | Widgets | ✅ | 2-4hr |

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
| **Current Focus** | **74** | **9** | **65** |
| 🔴 Critical | 8 | 3 | 5 |
| 🟠 High | 48 | 15 | 33 |
| 🟡 Medium | 32 | 8 | 24 |
| 🟢 Low | 22 | 0 | 22 |
| **Total** | **184** | **32** | **152** |

*Note: "Current Focus" items (FOCUS-01 through FOCUS-74) are the active work direction. Module 5 is deferred.*

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

**This file is the single source of truth for all tasks.** No separate prompt files.

Any agent tackling a task from this TODO **must**:

1. **Read TODO.md first** — align with current priorities before starting work.
2. **Work fully autonomously** — complete the task end-to-end without stopping for manual intervention. If ambiguity remains, choose the most reasonable default and document the choice.
3. **Validate progress continuously** — run `go test ./...` (headless) or launch the GUI to verify changes work. Never claim "done" without fresh test/build evidence.
4. **Headless-first** — use CLI + tests as primary validation. GUI runs only when explicitly needed.
5. **Minimal changes** — prefer targeted fixes over refactors unless required for correctness. Keep code changes within the smallest possible surface area.
6. **Update this TODO.md** — mark completed items as ✅, add any new tasks discovered during implementation, and update the progress summary.
7. **Never skip validation** — if blocked, report exact error output and last command run.

---

## Contributing

See `CONTRIBUTING.md` and `CLAUDE.md` for development guidelines.

**Scientific accuracy**: All claims must be verified per `HONESTY_AUDIT.md` standards.

---

*This TODO prioritizes scientific rigor and educational honesty over promotional considerations.*
*Document consolidated: 2026-02-07 | Refocused: 2026-02-11*
