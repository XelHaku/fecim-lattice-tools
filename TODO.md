# FeCIM Lattice Tools - TODO

**Mission**: Educational FeCIM visualization and simulation tool based on reported in literature HfO₂-ZrO₂ superlattice research.

**Last updated**: 2026-02-04

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
| L01 | Hysteresis cycle labels (wake-up, stable, fatigue phases) | ✅ | Done |
| L02 | Screenshot metadata embedding (PNG EXIF with parameters) | ✅ | Done |
| L03 | Add GitHub URL to glossary widget TODO | ✅ | Done (existing) |
| L04 | Hysteresis polarization bar indicator - increase to 16px with pulsing | ✅ | Done |
| L11 | Add [LK] indicators to material_picker.go for Landau-Khalatnikov parameters (requires LK model implementation first - see docs/hysteresis/hysteresis-glm.md Phase 4: P1 Landau-Khalatnikov model) | ⏳ | 1hr |

### P4-D2: Medium-Effort Low-Priority Fixes

| ID | Task | Status | Est. |
|----|------|--------|------|
| L05 | "About the Science" unified Learn More section | ✅ | Done |
| L06 | Accessibility audit (keyboard nav, ARIA labels, high-contrast mode) | ✅ | Done |
| L07 | Demo video creation (2-3 min walkthrough) | ⏳ | 4hr |

### P4-D3: Hard Low-Priority Fixes

| ID | Task | Status | Est. |
|----|------|--------|------|
| L08 | Web deployment (WASM) for browser-based demos | ⏳ | 16hr |
| L09 | Vulkan rendering implementation for large arrays | ⏳ | 20hr |
| L10 | 3D multi-layer visualization (512-layer roadmap) | ⏳ | 24hr |

---

## 7. Documentation Work

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

### Open Issues

| ID | Issue | Priority | Difficulty |
|----|-------|----------|------------|
| LK04 | L-K coefficients not calibrated to Ec/Pr | P2 | D3 |
| LK05 | ISPP controller not optimized for L-K dynamics | P2 | D3 |
| LK06 | Missing Q12 in some materials | P3 | D1 |
| LK07 | Need longer WAIT phases for L-K settling | P2 | D2 |

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
- [ ] DefaultHZO + L-K + ISPP → all 30 levels
- [ ] LiteratureSuperlattice + L-K + ISPP → all 30 levels
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
| P2 High | 16 | 16 | 0 | 100% |
| P3 Medium | 16 | 16 | 0 | 100% |
| P4 Low | 11 | 7 | 4 | 64% |
| **TOTAL** | **56** | **52** | **4** | **93%** |

**Estimated Remaining Effort**: ~28 hours

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
