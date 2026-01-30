# FeCIM Lattice Tools - TODO

**Mission**: Educational FeCIM visualization and simulation tool based on peer-reviewed HfO₂-ZrO₂ superlattice research.

**Last updated**: 2026-01-29

**Master Critique Source**: See `CRITIQUE_MASTER_LIST.md` for full consolidation from Dr. Tour reviews.

---

## 1. Current Status

| Module | Purpose | Status | Tests | GUI |
|--------|---------|--------|-------|-----|
| **module1-hysteresis** | P-E hysteresis, Preisach model, 30 analog states | ✅ Complete | ✅ Passing | ✅ Fyne |
| **module2-crossbar** | Matrix-vector multiplication, IR drop, sneak paths | ✅ Complete | ✅ Passing | ✅ Fyne |
| **module3-mnist** | Neural network digit recognition | ✅ Complete | ✅ Passing | ✅ Fyne |
| **module4-circuits** | DAC/ADC/TIA peripheral circuits | ✅ Complete | ✅ Passing | ✅ Fyne |
| **module5-comparison** | Technology comparison framework | ✅ Complete | ✅ Passing | ✅ Fyne |
| **module6-eda** | Verilog/DEF/LEF/Liberty generation | ✅ Complete | ✅ Passing | ✅ Fyne |
| **docs/** | Scientific documentation (78 papers catalogued) | ✅ Complete | N/A | N/A |

**Project Health**:
- Go files: 233
- Test cases: 571 (all passing)
- Generated outputs: Verilog, DEF, LEF, Liberty files (real synthesizable code)
- Scientific Rigor Score: 4.0/5 (HONESTY_AUDIT v3.0)
- Code coverage: ~85%

**Critique Status**: 25/58 items completed (43%)

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
| C02 | Change "30 discrete states" from fact to hypothesis: "30 states (conference presentation, pending peer review)" | ✅ | Done |
| C03 | ~~Add TRL disclaimer to energy comparison charts~~ | ✅ | Done |
| C04 | ~~Update 87% MNIST to show peer-reviewed context (96.6-98.24%)~~ | ✅ | Done |
| C05 | ~~Add "Why 30?" dialog with verification status~~ | ✅ | Done |

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
| H05 | ~~Market chart disclaimers - TRL and projection warnings~~ | ✅ | Done |
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
- Total tests: 571
- Pass rate: 100%
- Coverage: ~85%

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

## 10. Verified Claims Reference

**Use ONLY these verified claims in documentation and demos:**

### Material Properties [VERIFIED]
- **Pr**: 15-34 µC/cm² (RT), 75 µC/cm² (4K)
- **Ec**: 0.6-1.5 MV/cm
- **Min thickness**: 3.6 nm
- **Sub-1V switching**: 0.5V @ 3.6nm

### Multi-Level States [VERIFIED vs UNVERIFIED]
- **32-140 analog states**: ✅ VERIFIED (Oh 2017, Song 2024)
- **30 states (Tour)**: ⚠️ UNVERIFIED - label as "conference claim, pending peer review"

### Endurance [VERIFIED]
- **10^12 cycles**: DEMONSTRATED (V:HfO₂ 2024, Science 2024)
- **10^9 cycles**: Conservative baseline (IEEE IRPS 2022)

### MNIST Accuracy [VERIFIED]
- **98.24%**: HZO-FTJ reservoir (ScienceDirect 2025)
- **96.6%**: 7 VT states (Nature Commun. 2023)
- **87% (Tour)**: ❌ REMOVED - unverified, below peer-reviewed benchmarks

### Energy Efficiency [VERIFIED with CAVEATS]
- **25-100× vs NAND**: Samsung Nature 2025 ✅
- **70,000× vs GPU (LLM)**: Nature Comp. Sci. 2025 ⚠️ (specific workload only)
- **10M× (Tour)**: ❌ REMOVED - no measurement data

### 3D Integration [VERIFIED]
- **22nm BEOL**: CEA-Leti December 2024
- **512 layers roadmap**: Samsung Nature 2025

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

## 12. Progress Summary

| Priority | Total | Done | Remaining | % Complete |
|----------|-------|------|-----------|------------|
| P1 Critical | 13 | 13 | 0 | 100% |
| P2 High | 16 | 16 | 0 | 100% |
| P3 Medium | 16 | 16 | 0 | 100% |
| P4 Low | 10 | 7 | 3 | 70% |
| **TOTAL** | **55** | **52** | **3** | **95%** |

**Estimated Remaining Effort**: ~27 hours

**Session Progress (2026-01-29)**:
- Sprint 2: C06 ✅, C07 ✅, C09 ✅, C10 ✅
- Sprint 3: C11 ✅, C12 ✅, H10 ✅, H15 ✅
- Sprint 4: C13 ✅, H16 ✅, H14 ✅ (COMPLETE)
- P2 Remaining: H08 ✅, H09 ✅, H11 ✅ (ALL P2 COMPLETE)
- P3 Progress: M12 ✅, M07 ✅, M09 ✅, M08 ✅, M11 ✅, M10 ✅, M14 ✅, M15 ✅, M13 ✅ (ALL P3 COMPLETE)
- P4 Progress: L01 ✅, L02 ✅, L03 ✅, L04 ✅, L05 ✅, L06 ✅ (7/10 complete)

---

## Footer

**Next review**: After Sprint 4 completion
**Contributing**: See CLAUDE.md for development guidelines
**Scientific accuracy**: All claims must be verified per HONESTY_AUDIT.md standards

---

*This TODO prioritizes scientific rigor and educational honesty over promotional considerations. The project is an open-source learning tool, not investment material.*
