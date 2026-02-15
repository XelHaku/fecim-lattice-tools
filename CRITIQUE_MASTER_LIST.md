# FeCIM Lattice Tools - Master Critique Consolidation

**Consolidated from:** drtour_todo_fixes.md, drtour-conversation.md, drtour-conversation-v2.md (a.md archived / not in repo)
**Date:** 2026-02-03
**Purpose:** Unified priority and difficulty ranking for all critique items
**Status Note:** This file is a **snapshot**. For current progress and phase info, see `docs/project/STATUS.md` and `TODO.md`.

---

## Summary Statistics (Snapshot: 2026-02-03)

| Metric | Value |
|--------|-------|
| Total unique items | 58 |
| Completed | 50 |
| Pending | 8 |
| Pending IDs | H03, H04, H13, L05, L07, L08, L09, L10 |
| Source note | a.md not present in repo; items tracked by ID in this file |

---

## Priority × Difficulty Matrix

### Legend
- **Priority**: P1 (Critical), P2 (High), P3 (Medium), P4 (Low)
- **Difficulty**: D1 (Easy, <1hr), D2 (Medium, 1-4hr), D3 (Hard, 4-16hr), D4 (Very Hard, 16+hr)
- **Status**: ✅ Done, 🔄 In Progress, ⏳ Pending

---

## TIER 1: CRITICAL (Must Fix Before Release)

### P1-D1: Easy Critical Fixes (Do First)

| ID | Item | Source | Status | Est. Time |
|----|------|--------|--------|-----------|
| C01 | Add "SIMULATION ONLY" banners to Module 5 | academic review (archived) | ✅ | Done |
| C02 | Change "30 states" from fact to hypothesis language | academic review (archived) | ✅ | Done |
| C03 | Add simulation-only labeling to energy comparison charts | drtour_todo_fixes | ✅ | Done |
| C04 | Remove 87% MNIST claim; label external benchmarks as reported | drtour_todo_fixes | ✅ | Done |
| C05 | Add "Why 30?" dialog with baseline explanation | drtour_todo_fixes | ✅ | Done |

### P1-D2: Medium-Effort Critical Fixes

| ID | Item | Source | Status | Est. Time |
|----|------|--------|--------|-----------|
| C06 | Add error bars to all physics parameters in UI | academic review (archived) | ✅ | Done |
| C07 | Fix temperature-dependent retention (Arrhenius scaling) | academic review (archived) | ✅ | Done |
| C08 | Accuracy degradation chart - label as projected; cite sources if used | drtour_todo_fixes | ✅ | Done |
| C09 | Label all extrapolated accuracy as "projected" | academic review (archived) | ✅ | Done |
| C10 | Add total system power breakdown (array + ADC/DAC + peripherals) | academic review (archived) | ✅ | Done |

### P1-D3: Hard Critical Fixes

| ID | Item | Source | Status | Est. Time |
|----|------|--------|--------|-----------|
| C11 | Implement device-to-device variation (Gaussian Ec/Pr distribution) | academic review (archived) | ✅ | Done |
| C12 | Add write-verify statistics visualization | academic review (archived) | ✅ | Done |
| C13 | Validate Preisach model against experimental hysteresis loops | academic review (archived) | ✅ | Done |

---

## TIER 2: HIGH PRIORITY (Fix Before Academic Use)

### P2-D1: Easy High-Priority Fixes

| ID | Item | Source | Status | Est. Time |
|----|------|--------|--------|-----------|
| H01 | Home screen module descriptions - add simulation caveats | drtour_todo_fixes | ✅ | Done |
| H02 | MAC count parallelism explanation | drtour_todo_fixes | ✅ | Done |
| H03 | Voltage range citations (thickness-dependent) | drtour_todo_fixes | ✅ | Done |
| H04 | Read parameter sources - mark as empirical | drtour_todo_fixes | ✅ | Done |
| H05 | Market chart disclaimers - simulation-only and projection warnings | drtour_todo_fixes | ✅ | Done |
| H06 | Cite strain coefficients (replace magic -0.15) | academic review (archived) | ✅ | Done |
| H07 | Add Preisach grid size convergence study reference | academic review (archived) | ✅ | Done |

### P2-D2: Medium-Effort High-Priority Fixes

| ID | Item | Source | Status | Est. Time |
|----|------|--------|--------|-----------|
| H08 | Add "Fabrication Reality" section (18-month timeline, $2M cost) | academic review (archived) | ✅ | Done |
| H09 | Module 4 - Add SAR ADC noise modeling | academic review (archived) | ✅ | Done |
| H10 | Module 2 - Add write disturb (half-select stress) model | academic review (archived) | ✅ | Done |
| H11 | Module 2 - Add parasitic capacitance for RC delay | academic review (archived) | ✅ | Done |
| H12 | Weight error context - add % of range explanation | drtour_todo_fixes | ✅ | Done |
| H13 | GPU comparison nuance - add batched operation context | drtour_todo_fixes | ✅ | Done |

### P2-D3: Hard High-Priority Fixes

| ID | Item | Source | Status | Est. Time |
|----|------|--------|--------|-----------|
| H14 | Add ISPP (Incremental Step Pulse Programming) visualization | academic review (archived) | ✅ | Done |
| H15 | Implement thermal physics (retention vs temperature curves) | academic review (archived) | ✅ | Done |
| H16 | Add "Simulation vs Experiment" comparison tab | academic review (archived) | ✅ | Done |

---

## TIER 3: MEDIUM PRIORITY (Polish)

### P3-D1: Easy Medium-Priority Fixes

| ID | Item | Source | Status | Est. Time |
|----|------|--------|--------|-----------|
| M01 | EDA status prominence - "Coming Soon" label | drtour_todo_fixes | ✅ | Done |
| M02 | Hysteresis Ec threshold visualization (dashed lines) | drtour_todo_fixes | ✅ | Done |
| M03 | Voltage zone legend (green/yellow/red) | drtour_todo_fixes | ✅ | Done |
| M04 | Energy breakdown annotation (peripheral percentages) | drtour_todo_fixes | ✅ | Done |
| M05 | Glossary widget integration | drtour_todo_fixes | ✅ | Done |
| M06 | References widget with DOI links | drtour_todo_fixes | ✅ | Done |

### P3-D2: Medium-Effort Medium-Priority Fixes

| ID | Item | Source | Status | Est. Time |
|----|------|--------|--------|-----------|
| M07 | Simplified log toggle for hysteresis | drtour_todo_fixes | ✅ | Done |
| M08 | Sneak path side-by-side comparison view | drtour_todo_fixes | ✅ | Done |
| M09 | Crossbar cell-level inspection (hover tooltips) | drtour_todo_fixes | ✅ | Done |
| M10 | Architecture comparison split-screen mode | drtour_todo_fixes | ✅ | Done |
| M11 | Error attribution breakdown in accuracy analysis | drtour_todo_fixes | ✅ | Done |
| M12 | Hysteresis state stability warnings (color-coded levels) | drtour_todo_fixes | ✅ | Done |

### P3-D3: Hard Medium-Priority Fixes

| ID | Item | Source | Status | Est. Time |
|----|------|--------|--------|-----------|
| M13 | Responsive layout breakpoints (1600px/1024px/768px) | drtour_todo_fixes | ✅ | Done |
| M14 | Conductance drift time-dependent visualization | TODO.md | ✅ | Done |
| M15 | Device-to-device variation modeling GUI | TODO.md | ✅ | Done |

---

## TIER 4: LOW PRIORITY (Nice to Have)

### P4-D1: Easy Low-Priority Fixes

| ID | Item | Source | Status | Est. Time |
|----|------|--------|--------|-----------|
| L01 | Hysteresis cycle labels (wake-up, stable, fatigue) | drtour_todo_fixes | ✅ | Done |
| L02 | Screenshot metadata (PNG EXIF) | drtour_todo_fixes | ✅ | Done |
| L03 | Add GitHub URL to glossary widget | TODO.md | ✅ | Done |
| L04 | Hysteresis polarization bar indicator size increase | drtour_todo_fixes | ✅ | Done |

### P4-D2: Medium-Effort Low-Priority Fixes

| ID | Item | Source | Status | Est. Time |
|----|------|--------|--------|-----------|
| L05 | About the Science section (unified Learn More) | drtour_todo_fixes | ⏳ | 2hr |
| L06 | Accessibility audit (keyboard nav, high-contrast) | drtour_todo_fixes | ✅ | Done |
| L07 | Demo video creation (2-3 min walkthrough) | TODO.md | ⏳ | 4hr |

### P4-D3: Hard Low-Priority Fixes

| ID | Item | Source | Status | Est. Time |
|----|------|--------|--------|-----------|
| L08 | Web deployment (WASM) | TODO.md | ⏳ | 16hr |
| L09 | Vulkan rendering implementation | TODO.md | ⏳ | 20hr |
| L10 | 3D multi-layer visualization | TODO.md | ⏳ | 24hr |

---

## DEFERRED / OUT OF SCOPE

| Item | Reason |
|------|--------|
| Production chip design tools | Educational tool, not EDA replacement |
| Hardware-accurate SPICE models | Requires proprietary foundry PDKs |
| Real-time OS integration | Beyond educational scope |
| Web-based collaboration | Single-user educational tool |
| Investor pitch decks | Scientific tool, not marketing material |

---

## Recommended Execution Order

### Sprint 1: Citations + Transparency (1 day)
1. H03 - Voltage range citations (Circuits reference voltage)
2. H04 - Read parameter sources / mark as assumed
3. H13 - GPU comparison nuance (batched throughput caveat)

### Sprint 2: About + Media (2-4 days)
1. L05 - About the Science section (global entry point)
2. L07 - Demo video creation (2-3 min walkthrough)

### Sprint 3: Platform Extensions (2-4 weeks)
1. L08 - Web deployment (WASM)
2. L09 - Vulkan rendering implementation
3. L10 - 3D multi-layer visualization

---

## Progress Tracking

**Completed:** 50/58 (86%)
**Critical Remaining:** 0
**High Remaining:** 3
**Medium Remaining:** 0
**Low Remaining:** 5

**Estimated Total Remaining Effort:** ~50-80 hours (dominated by WASM/3D work)

---

## Cross-Reference: Document Reconciliation

- **a.md is not present in this repo**; items from the academic review are preserved by ID and labeled `academic review (archived)` in the Source column.
- **drtour_todo_fixes.md** remains the primary status log for Tour critique and UI/UX items.
- **TODO.md** remains the source for longer-term platform work (WASM/Vulkan/3D).

---

*Document updated 2026-02-03*
*Next update: After H03/H04/H13 and L05 are complete*
