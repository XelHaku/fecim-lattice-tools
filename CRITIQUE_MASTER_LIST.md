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
| C03 | Add TRL disclaimer to energy comparison charts | drtour_todo_fixes | ✅ | Done |
| C04 | Update 87% MNIST to show peer-reviewed context (96.6-98.24%) | drtour_todo_fixes | ✅ | Done |
| C05 | Add "Why 30?" dialog with verification status | drtour_todo_fixes | ✅ | Done |

### P1-D2: Medium-Effort Critical Fixes

| ID | Item | Source | Status | Est. Time |
|----|------|--------|--------|-----------|
| C06 | Add error bars to all physics parameters in UI | academic review (archived) | ✅ | Done |
| C07 | Fix temperature-dependent retention (Arrhenius scaling) | academic review (archived) | ✅ | Done |
| C08 | Accuracy degradation chart - add sources and confidence intervals | drtour_todo_fixes | ✅ | Done |
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
| H03 | Voltage range citations (thickness-dependent) | drtour_todo_fixes | ⏳ | 1hr |
| H04 | Read parameter sources - mark as empirical | drtour_todo_fixes | ⏳ | 1hr |
| H05 | Market chart disclaimers - TRL and projection warnings | drtour_todo_fixes | ✅ | Done |
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
| H13 | GPU comparison nuance - add batched operation context | drtour_todo_fixes | ⏳ | 1hr |

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

### Sprint 1: Critical Easy Wins (1 day)
1. C01 - Add "SIMULATION ONLY" banners
2. C02 - Change "30 states" language
3. H06 - Cite strain coefficients
4. H07 - Add Preisach grid size reference

### Sprint 2: Critical Medium Effort (2-3 days)
1. C06 - Add error bars to physics parameters
2. C07 - Fix temperature-dependent retention
3. C09 - Label extrapolated accuracy as "projected"
4. C10 - Add total system power breakdown

### Sprint 3: High Priority Physics (1 week)
1. C11 - Device-to-device variation
2. C12 - Write-verify statistics
3. H10 - Write disturb model
4. H15 - Thermal physics implementation

### Sprint 4: Validation (2+ weeks)
1. C13 - Validate Preisach against experimental data
2. H16 - Simulation vs Experiment tab
3. H14 - ISPP visualization

---

## Progress Tracking

**Completed:** 25/58 (43%)
**Critical Remaining:** 10
**High Remaining:** 9
**Medium Remaining:** 9
**Low Remaining:** 5

**Estimated Total Remaining Effort:** ~150 hours

---

## Cross-Reference: Document Reconciliation

### Items Merged from Multiple Sources

| Final ID | drtour_todo_fixes | a.md | Notes |
|----------|-------------------|------|-------|
| C01 | CRIT-001 (done) | Tier 1 #2 | Both require TRL disclaimers |
| C02 | CRIT-002 | Tier 1 #1 | 30 states language |
| C11 | n/a | Tier 2 #1 | New from academic review |
| H15 | CRIT-004 (verified) | Tier 1 #3 | Temperature dependence |

### Items Only in a.md (New)
- All Tier 2 physics items (SAR ADC, write disturb, parasitic C)
- Fabrication Reality section
- ISPP visualization
- Simulation vs Experiment tab

### Items Only in drtour_todo_fixes (UI/UX)
- All UI-### items (typography, spacing, contrast)
- Sneak path comparison view
- Cell-level inspection tooltips

---

*Document auto-generated 2026-01-29*
*Next update: After Sprint 1 completion*
