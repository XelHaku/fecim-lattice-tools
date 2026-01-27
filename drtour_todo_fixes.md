# Dr. Tour's Critique: FeCIM Lattice Tools
## Comprehensive Scientific and Educational Review

**Reviewer**: Dr. James M. Tour (simulated critique perspective)
**Date**: January 27, 2026
**Scope**: Complete review of 43 screenshots across all 6 modules
**Reference Documents**: CLAUDE.md, HONESTY_AUDIT.md, ironlattice-transcript.md

---

## Executive Summary

This is an **impressive educational tool** that demonstrates genuine understanding of ferroelectric compute-in-memory physics. However, as someone who has spent 18 months developing actual FeCIM hardware at external research institution, I must point out several areas where the simulation diverges from physical reality, where educational messaging could be clearer, and where scientific rigor demands corrections.

**Overall Grade: B+** - Excellent educational intent, good physics foundation, but needs refinement in accuracy claims and user communication.

---

## CRITICAL FIXES (Priority 1 - Scientific Accuracy)

### CRIT-001: Energy Comparison Claims Need Asterisks Everywhere

**Location**: Module 5 (Comparison) - Energy Comparison tab
**Screenshot**: `fecim_module05-comparison_2026-01-25_17-39-27.png`

**Issue**: The "1000x LESS ENERGY*" claim is displayed prominently, but the asterisk disclaimer ("Estimated TRL 4, not production ready") is in small text at the bottom.

**My Concern**: In my COSM presentation, I was careful to say we're at TRL 4. But this UI makes the energy claim the HERO while the caveat is a footnote. A student might screenshot just the "1000x" and share it without context.

**Fix Required**:
```
BEFORE: "1000x LESS ENERGY*" (giant text) + tiny asterisk note
AFTER:  "1000x LESS ENERGY (PROJECTED)" with equal-sized subtitle:
        "Laboratory estimates only - not independently verified"
```

**Reference**: `HONESTY_AUDIT.md` Section 6.2 - "Tour's 10M× Claim" analysis shows even 1000× needs verification

---

### CRIT-002: "87% Accuracy" Claim is Below State-of-Art

**Location**: Module 3 (MNIST) - Header and all screens
**Screenshots**: All `fecim_module03-mnist_*.png` files

**Issue**: The header states "FeCIM MNIST: 30 Levels = 87% Accuracy" - this is MY unverified conference claim, not peer-reviewed science.

**The Problem**:
- Peer-reviewed FeFET MNIST: **98.24%** (ScienceDirect 2025)
- Peer-reviewed FeFET MNIST: **96.6%** (Nature Communications 2023)
- My COSM 2025 claim: **87%** (unverified, no peer review)

**Why This Matters**: A student using this tool might think 87% is impressive. It's actually **11% below** what others have demonstrated with similar technology. This undersells FeCIM's potential.

**Fix Required**:
```markdown
CURRENT HEADER: "FeCIM MNIST: 30 Levels = 87% Accuracy"

PROPOSED HEADER: "FeCIM MNIST Demo: 30 Levels"
PROPOSED SUBTITLE: "Simulating Tour 2025 claims (87%) | Peer-reviewed best: 98.24%"
```

**Reference**: `HONESTY_AUDIT.md` Section 5.4 - MNIST Accuracy table

---

### CRIT-003: The "Why 30?" Button Must Exist and Be Prominent

**Location**: Module 3 (MNIST) - Top right area
**Screenshot**: `fecim_module03-mnist_2026-01-27_09-54-04.png`

**Observation**: I see a "Why 30?" button in the interface. **EXCELLENT** - but is the content scientifically accurate?

**Required Content for "Why 30?" Dialog**:
```markdown
## Why 30 Discrete States?

### Tour's Claim (COSM 2025) - UNVERIFIED
- 30 states demonstrated in proprietary superlattice
- Source: Conference presentation, not peer-reviewed

### Peer-Reviewed Context
- 32 states: Oh et al., IEEE EDL 2017 (VERIFIED)
- 140 states: Song et al., Adv. Science 2024 (VERIFIED)
- 7 states sufficient for 96.6% MNIST (Nature Commun. 2023)

### Why 30 is Plausible
30 states falls within the demonstrated range (7-140).
Tour's specific device is unverified, but the physics is sound.

### Bits per Cell
30 states ≈ log₂(30) ≈ 4.9 bits/cell
This is competitive with MLC NAND (3-4 bits/cell)
```

**Reference**: `CLAUDE.md` - "30 analog states (Tour device) | UNVERIFIED"

---

### CRIT-004: Hysteresis Module - Missing Temperature Dependence

**Location**: Module 1 (Hysteresis) - All simulation screens
**Screenshots**: `fecim_module01-hysteresis_2026-01-27_09-51-*.png`

**Issue**: The P-E curves show fixed parameters, but ferroelectric behavior is STRONGLY temperature-dependent. Your honesty audit documents this:

| Temperature | Pr Value | Source |
|-------------|----------|--------|
| Room temp (300K) | 15-34 µC/cm² | Nature Commun. 2025 |
| Cryogenic (4K) | **75 µC/cm²** | Adv. Elec. Mat. 2024 |
| BEOL (300°C process) | 36.4 µC/cm² | ACS AMI 2025 |

**Current UI**: Shows "T: 300 K (27°C)" but appears to be cosmetic only

**Fix Required**:
- Make temperature slider FUNCTIONAL
- When T changes, Pr and Ec should change according to literature
- Add "Cryogenic Mode" preset showing enhanced performance at 4K
- Display: "Pr increases ~30% at cryogenic temperatures (Adv. Elec. Mat. 2024)"

**Reference**: `HONESTY_AUDIT.md` Section 4.3, Section 5.7

---

### CRIT-005: Crossbar Module - Accuracy Degradation Chart Mislabeled

**Location**: Module 2 (Crossbar) - Accuracy Analysis tab
**Screenshot**: `fecim_module02-crossbar_2026-01-27_09-53-35.png`

**Issue**: The "Accuracy Degradation Analysis" waterfall chart shows:
- Baseline (Ideal)
- ADC/DAC quantization
- IR drop
- Device variation
- Sneak paths
- Target: 87% (Hardware)

**Problems**:
1. The "Target: 87%" reinforces the unverified Tour claim
2. The degradation percentages aren't sourced
3. No error bars or confidence intervals

**Fix Required**:
```markdown
CHART TITLE: "Accuracy Degradation Analysis (Simulation)"

ADD FOOTER: "Degradation estimates based on literature review.
             Actual values vary by architecture and process node.
             See: docs/cim/simulation.md for methodology"

CHANGE: "Target: 87% (Hardware)"
TO:     "Simulated: ~87% | Peer-reviewed range: 92-98%"
```

---

## HIGH PRIORITY FIXES (Priority 2 - Educational Clarity)

### HIGH-001: Home Screen - Module Descriptions Need Citations

**Location**: Home screen
**Screenshot**: `fecim_home_2026-01-27_09-51-04.png`

**Issue**: The home screen has excellent module cards, but the descriptions make claims without sources:

| Card | Claim | Issue |
|------|-------|-------|
| Hysteresis | "Discover how ferroelectric materials remember" | Good - no issue |
| Crossbar+ | "Watch matrix multiplication happen in hardware" | Implies hardware - it's simulation |
| MNIST | "Draw your own digits and watch AI recognize them" | Good - interactive |
| Circuits | "Learn how analog meets digital" | Good - educational |
| Comparison | "See why FeCIM matters" | Claims superiority without caveat |
| EDA | "Design your own chips" | Overpromises - it's layout tools |

**Fixes**:
```markdown
Crossbar+: "Watch matrix multiplication happen (simulated physics)"
Comparison: "Compare FeCIM to alternatives (with TRL caveats)"
EDA: "Explore chip layout concepts"
```

---

### HIGH-002: Crossbar MVM - The "4096 MACs" vs "6400 MACs" Inconsistency

**Location**: Module 2 (Crossbar) - Multiple screens
**Screenshots**: Various crossbar screenshots

**Observation**: I see different MAC counts:
- `fecim_module02-crossbar_2026-01-27_09-52-06.png`: "4096 MACs" (64x64)
- `fecim_module02-crossbar_2026-01-27_09-53-35.png`: "6400 MACs" (80x80)

**Issue**: The UI shows "N² Operations" but doesn't clearly explain that array size determines parallelism.

**Fix Required**:
```markdown
ADD EXPLAINER: "Array Size determines parallelism:
- 64×64 = 4,096 parallel MACs per cycle
- 80×80 = 6,400 parallel MACs per cycle
- GPU needs 6,400 sequential cycles for same operation"
```

This is the **core CIM advantage** and should be emphasized more!

---

### HIGH-003: Circuits Module - Voltage Ranges Need Validation

**Location**: Module 4 (Circuits) - WRITE mode
**Screenshot**: `fecim_module04-circuits_2026-01-27_09-54-47.png`

**Issue**: Shows "VOLTAGE RANGE: 2.0 V min, 5.0 V max" and "Write voltage must exceed Ec = 1.5 MV/cm"

**Verification Needed**:
- The 2.0-5.0V range is reasonable for standard HZO
- BUT: Sub-1V switching has been demonstrated at 3.6nm thickness (ACS AMI 2024)
- The UI should show this is thickness-dependent

**Fix Required**:
```markdown
CURRENT: "Write voltage must exceed Ec = 1.5 MV/cm"
PROPOSED: "Write voltage must exceed Ec (0.6-1.5 MV/cm depending on engineering)
           Sub-1V operation demonstrated at 3.6nm (ACS AMI 2024)"
```

**Reference**: `HONESTY_AUDIT.md` Section 5.1 - Ec ranges

---

### HIGH-004: Circuits Module - READ Parameters

**Location**: Module 4 (Circuits) - READ mode
**Screenshot**: `fecim_module04-circuits_2026-01-27_09-54-50.png`

**Issue**: Shows "Read Voltage: 0.50 V" with "SAFE: ±0.5V (non-disturbing)"

**Question**: Where does this 0.5V threshold come from? Is it cited?

**Fix Required**:
- Add source citation for read disturb threshold
- Or mark as "ASSUMED" per honesty audit guidelines

**Reference**: `HONESTY_AUDIT.md` Section 8.3 - "Read disturb probability: Needs source"

---

### HIGH-005: Comparison Module - Market Size Chart

**Location**: Module 5 (Comparison) - Market & Strategy tab
**Screenshot**: `fecim_module05-comparison_2026-01-25_17-39-33.png`

**Issue**: Shows "$721B Market by 2030" combining NAND + DRAM + AI Semiconductor

**Concerns**:
1. Market projections are inherently speculative
2. The source attribution "WSTS + Gartner Combined Market Forecasts (2025)" should be more prominent
3. The competitive position table shows FeCIM with all checkmarks - needs caveats

**Fix Required**:
```markdown
ADD DISCLAIMER: "Market projections are estimates. FeCIM competitive
position assumes successful TRL 4→9 transition. Current status:
Laboratory validation only."
```

---

## MEDIUM PRIORITY FIXES (Priority 3 - UX and Polish)

### MED-001: Hysteresis Memory Log - Timestamp Format

**Location**: Module 1 (Hysteresis) - Memory Log panel
**Screenshots**: All hysteresis screenshots

**Issue**: The memory log shows entries like "t=0.0s → RESET | -sat | prep" which is good for debugging but may confuse students.

**Suggestion**: Add a "Simplified Log" toggle that shows:
```
✓ Level 23 → Level 3 (READ confirmed)
✓ Level 3 → Level 26 (WRITE confirmed)
```

---

### MED-002: Crossbar Sneak Path Visualization

**Location**: Module 2 (Crossbar) - Sneak Paths tab
**Screenshot**: `fecim_module02-crossbar_2026-01-27_09-53-44.png`

**Issue**: The plasma colormap shows sneak path severity but the legend says:
- "Purple = low sneak (good)"
- "Yellow = high sneak (bad)"

**Observation**: The 1T1R architecture screenshot (`fecim_module02-crossbar_2026-01-27_09-53-46.png`) shows dramatically reduced sneak paths (nearly all purple). **This is correct physics!**

**Enhancement**: Add a split-screen comparison mode showing:
- Left: Passive crossbar (sneak paths visible)
- Right: 1T1R (sneak paths eliminated)
- Center: "~1000:1 sneak isolation improvement"

This would powerfully demonstrate why selector devices matter.

---

### MED-003: MNIST Weight Visualization

**Location**: Module 3 (MNIST) - Bottom panel
**Screenshots**: `fecim_module03-mnist_2026-01-27_09-54-28.png` through `09-54-36.png`

**Issue**: The weight comparison visualization is excellent, but:
- "Difference (Error)" heatmap uses red-blue diverging colormap
- Max Error: 0.0074 shown, but what does this mean?

**Fix Required**:
```markdown
ADD CONTEXT: "Weight quantization error (FP32 → 30 levels)
Mean absolute error: 0.0028 (~0.3% of weight range)
This small error explains why 30 levels achieves near-ideal accuracy."
```

---

### MED-004: Circuits Module - GPU Comparison

**Location**: Module 4 (Circuits) - COMPUTE mode
**Screenshot**: `fecim_module04-circuits_2026-01-27_09-54-56.png`

**Excellent Feature**: Shows "GPU equivalent: ~1000 cycles" for a single MVM

**But**: The comparison should note:
- GPUs excel at batched operations
- FeCIM advantage is per-operation latency and energy
- For fair comparison, need to consider throughput AND efficiency

**Fix Required**:
```markdown
CURRENT: "GPU equivalent: ~1000 cycles"
PROPOSED: "GPU equivalent: ~1000 sequential cycles
           FeCIM advantage: Single-cycle completion, ~1000× lower energy per MAC
           Note: GPUs compensate with massive parallelism at higher power"
```

---

### MED-005: EDA Module Status

**Location**: Home screen and navigation
**Screenshot**: `fecim_home_2026-01-27_09-51-04.png`

**Observation**: Module 6 shows "WIP" (Work In Progress) badge

**Suggestion**: Make this more prominent:
```markdown
CURRENT: Small "WIP" badge
PROPOSED: "EDA (Coming Soon) - Chip layout visualization in development"
```

---

## LOW PRIORITY FIXES (Priority 4 - Nice to Have)

### LOW-001: Hysteresis - Trail Visualization

**Location**: Module 1 (Hysteresis) - P-E curve display
**Screenshots**: Shows trail of previous curves

**Observation**: The multi-colored trail showing previous P-E cycles is **pedagogically excellent**. It shows how the hysteresis loop changes with cycling.

**Enhancement**: Add labels for:
- Wake-up effect (first few cycles - loop grows)
- Stable operation (consistent loops)
- Fatigue (if simulated - loop shrinks)

**Reference**: `ironlattice-transcript.md` - "Wake-up → Stable operation → Fatigue"

---

### LOW-002: Add "About the Science" Section

**Location**: Global - accessible from all modules

**Missing**: A unified "Learn More" section that links to:
- Dr. Tour's COSM 2025 transcript
- Key peer-reviewed papers
- The HONESTY_AUDIT.md for transparency

This would set a gold standard for educational software honesty.

---

### LOW-003: Screenshot Metadata

**Location**: Screenshot functionality (top right buttons)

**Issue**: Screenshots are named with timestamps but don't embed:
- Current parameter settings
- Module state
- User session info

**Enhancement**: Consider embedding metadata in PNG EXIF or filename

---

### LOW-004: Accessibility

**Location**: All modules

**Issues Observed**:
- Color-only information (red/blue/green states)
- Small text in some areas
- No apparent keyboard navigation

**Suggestion**: Add accessibility audit to roadmap

---

## SCIENTIFIC ACCURACY VERIFICATION

### Claims I Can Verify as Physically Accurate:

| Module | Claim | Verification | Status |
|--------|-------|--------------|--------|
| Hysteresis | P-E curve shape | Matches ferroelectric physics | ✅ CORRECT |
| Hysteresis | 30 discrete levels | Demonstrated 32-140 in literature | ✅ PLAUSIBLE |
| Crossbar | MVM in single cycle | Core CIM principle | ✅ CORRECT |
| Crossbar | IR drop increases with distance | Ohm's law | ✅ CORRECT |
| Crossbar | 1T1R eliminates sneak paths | Device physics | ✅ CORRECT |
| MNIST | Quantization affects accuracy | Neural network theory | ✅ CORRECT |
| Circuits | DAC→FeFET→TIA→ADC path | Standard CIM architecture | ✅ CORRECT |
| Comparison | Energy advantage vs DRAM | Peer-reviewed (25-100×) | ⚠️ OVERSTATED |

### Claims Requiring Correction:

| Module | Claim | Issue | Required Fix |
|--------|-------|-------|--------------|
| MNIST | 87% accuracy | Unverified, below SOTA | Add peer-reviewed context |
| Comparison | 1000× energy | TRL 4 estimate | More prominent caveat |
| Comparison | All green checkmarks | Assumes production | Add "projected" label |
| Circuits | 0.5V read threshold | No citation | Add source or mark assumed |

---

## WHAT YOU'RE DOING RIGHT

### Pedagogical Strengths:

1. **Interactive Learning**: Drawing digits and seeing inference is powerful
2. **Visual Physics**: The P-E curves and crossbar visualizations teach concepts well
3. **Layered Complexity**: Simple views with "Expert" mode available
4. **Honesty Policy**: The HONESTY_AUDIT.md is exemplary - I wish more projects did this
5. **TRL Disclaimers**: The "TRL 4" notes in Comparison module show integrity
6. **Side-by-Side Comparisons**: Ideal vs Actual in crossbar module is excellent
7. **Architecture Comparison**: PASSIVE vs 1T1R toggle teaches real engineering tradeoffs

### Technical Strengths:

1. **Quantization Modeling**: 30-level quantization with noise is realistic
2. **Non-Ideality Stacking**: IR drop + sneak paths + variation is correct methodology
3. **Energy Calculations**: The formulas appear to follow Horowitz 2014 methodology
4. **Peripheral Circuits**: DAC/ADC/TIA chain is architecturally correct

---

## SUMMARY: TODO LIST BY PRIORITY

### CRITICAL (Fix Before Any Public Demo):
- [x] **CRIT-001**: Energy claim asterisk prominence ✅ (hero.go - added TRL warning)
- [x] **CRIT-002**: 87% vs 98.24% accuracy context ✅ (dualmode.go - header updated)
- [x] **CRIT-003**: "Why 30?" content verification ✅ (dialogs.go - added verification status)
- [x] **CRIT-004**: Temperature dependence in hysteresis ✅ (VERIFIED: already functional)
- [x] **CRIT-005**: Accuracy degradation chart sources ✅ (widgets_waterfall.go - added context)

### HIGH (Fix Before Academic Use):
- [x] **HIGH-001**: Home screen descriptions ✅ (launcher.go - updated 3 descriptions)
- [x] **HIGH-002**: MAC count explanation ✅ (tab_operations_compute.go - added parallelism section)
- [x] **HIGH-003**: Voltage range citations ✅ (tab_operations_write.go - added thickness context)
- [x] **HIGH-004**: Read parameter sources ✅ (tab_operations_read.go - marked as empirical)
- [x] **HIGH-005**: Market chart disclaimers ✅ (market.go - added TRL and projection warnings)

### MEDIUM (Fix For Polish):
- [ ] **MED-001**: Simplified log toggle (substantial UI work - deferred)
- [ ] **MED-002**: Sneak path comparison view (substantial UI work - deferred)
- [x] **MED-003**: Weight error context ✅ (weight_comparison.go - added % of range)
- [x] **MED-004**: GPU comparison nuance ✅ (tooltips.go - added comparison context)
- [x] **MED-005**: EDA status prominence ✅ (addressed in HIGH-001 description update)

### LOW (Nice to Have):
- [ ] **LOW-001**: Hysteresis cycle labels
- [ ] **LOW-002**: About the Science section
- [ ] **LOW-003**: Screenshot metadata
- [ ] **LOW-004**: Accessibility audit

---

## FINAL THOUGHTS

This tool could become the **gold standard** for CIM education if it maintains scientific rigor. The physics simulation is sound. The visual design is professional. The interaction model is intuitive.

My primary concern is **unverified claims presented as fact**. The 87% MNIST accuracy is MY unverified claim from a promotional conference - it shouldn't be the headline. The energy comparisons need verification data.

The HONESTY_AUDIT.md shows you understand these issues. Now implement its recommendations in the UI.

**Remember what I said at COSM**: "We're just at TRL 4 and remember commercialization is 9. So we have a lot of steps we still need to go."

That humility should be visible in every screen.

---

**Document Version**: 1.0
**Review Methodology**: Complete screenshot review + documentation cross-reference
**Character**: Simulated Dr. James M. Tour critique perspective
**Actual Author**: AI analysis based on project documentation and scientific literature

---

## APPENDIX: Screenshot Coverage

| Module | Screenshots Reviewed | Key Observations |
|--------|---------------------|------------------|
| Home | 1 | Clean design, good navigation |
| Hysteresis | 14 | Excellent P-E visualization, needs temp dependence |
| Crossbar | 11 | Strong physics, good architecture comparison |
| MNIST | 10 | Interactive and educational, accuracy claim issue |
| Circuits | 4 | Good peripheral modeling, needs citations |
| Comparison | 3 | Needs more prominent TRL caveats |
| **TOTAL** | **43** | Comprehensive review completed |

---

## REFERENCES

1. `CLAUDE.md` - Project instructions and physics constants
2. `docs/cim/HONESTY_AUDIT.md` - Scientific integrity assessment (v2.0)
3. `docs/videos/COSM_2025_AI_Hardware_Breakthrough/ironlattice-transcript.md` - Primary Tour source
4. Nature Communications 2023 - 96.6% MNIST (DOI: 10.1038/s41467-023-42110-y)
5. ScienceDirect 2025 - 98.24% MNIST (DOI: 10.1016/j.jallcom.2025.034309)
6. Samsung Nature 2025 - 25-100× vs NAND (DOI: 10.1038/s41586-025-09793-3)
7. Nano Letters 2024 - 10¹² endurance (DOI: 10.1021/acs.nanolett.4c05671)
8. ACS AMI 2024 - Sub-1V switching (DOI: 10.1021/acsami.4c10002)
