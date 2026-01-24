# Plan-Demo6 Improvement Recommendations

**Focus:** Technical accuracy + Proper attribution  
**Date:** 2026-01-24

---

## 🔴 Critical Corrections Required

### 1. OpenLane Variable Names (Line 2946)

**Current (INCORRECT):**
```go
"PLACEMENT_CURRENT_DEF": fmt.Sprintf("dir::output/%s.def", designName),
```

**Fix:**
```go
"FP_DEF_TEMPLATE": fmt.Sprintf("dir::output/%s.def", designName),
```

**Add citation:**
```go
// Reference: OpenLane v2.0 Configuration Variables
// https://openlane.readthedocs.io/en/latest/reference/configuration.html#floorplanning
```

---

### 2. Add Missing Citation: SKY130 PDK

**Where:** Line 78 (Technology: "sky130")

**Add:**
```markdown
// Technology reference: SkyWater SKY130 PDK
// https://skywater-pdk.readthedocs.io/
// Open-source 130nm process, released 2020
```

---

### 3. Cell Dimensions Need Source

**Current:** Lines 75-76
```go
Width:  0.46 μm
Height: 2.72 μm
```

**Problem:** No justification for these specific dimensions

**Fix:** Add one of:
```markdown
Option A: "Based on SKY130 unithd site: 0.46 × 2.72 μm (1× standard cell height)"
Option B: "Placeholder dimensions - requires actual FeFET cell layout in Magic"
Option C: Cite if based on specific FeFET paper
```

**Citation format:**
```
// Cell dimensions: [Source]
// - Width: 0.46 μm (SKY130 site width)
// - Height: 2.72 μm (SKY130 standard cell height)
// Note: Actual FeFET cell requires custom layout verification
```

---

### 4. Liberty Timing Values Are Placeholders

**Current:** Lines 80-84 have placeholder values without disclaimer

**Fix:** Add clear "PLACEHOLDER" markers:
```go
// PLACEHOLDER TIMING VALUES
// Real FeFET characterization required via:
// 1. SPICE simulation (ngspice + FeFET model)
// 2. Liberty characterization tools
// 3. Measurement from fabricated test structures
RiseTime:     0.1,  // ns (PLACEHOLDER)
FallTime:     0.1,  // ns (PLACEHOLDER)
InputCap:     0.002, // pF (PLACEHOLDER)
LeakagePower: 0.001, // nW (PLACEHOLDER)
```

---

### 5. FeFET Behavioral Model Disclaimer

**Current:** Cell Verilog has comment but not prominent enough

**Fix:** Add to top of file and in plan:
```verilog
// ⚠️ CRITICAL LIMITATION ⚠️
// This is a PLACEHOLDER behavioral model
// Real FeFET requires:
// - Polarization state modeling (P+ / P-)
// - Threshold voltage hysteresis
// - Retention time effects
// - SPICE-level compact model
//
// References needed:
// - FeFET compact models: [cite papers]
// - Verilog-A implementations: [cite if available]
```

---

## 📚 Required Citations Framework

### Section 1: OpenLane Integration

**Add to plan header:**
```markdown
## References

### OpenLane Documentation
- **Configuration Variables:** https://openlane.readthedocs.io/en/latest/reference/configuration.html
- **Version:** OpenLane v2.0.0a47+ (SYNTH_ELABORATE_ONLY functional)
- **Pre-placed Macros Guide:** https://openlane.readthedocs.io/en/latest/usage/hardening_macros.html

### PDK References
- **SkyWater SKY130:** https://skywater-pdk.readthedocs.io/
  - Standard cell height: 2.72 μm
  - Site definition: unithd (0.46 × 2.72 μm)
  - Metal layers: M1-M5 available
```

### Section 2: File Format Specifications

**Add:**
```markdown
### EDA File Format Specifications
- **LEF (Library Exchange Format):** 
  - LEF/DEF 5.8 Language Reference
  - Si2/OpenAccess Coalition
  - Used: PIN definitions, MACRO geometry, OBS layers

- **Liberty (.lib):**
  - Liberty Timing Format Reference Manual
  - Synopsys (industry standard)
  - Used: cell_rise, cell_fall, timing arcs

- **DEF (Design Exchange Format):**
  - DEF 5.8 Specification
  - COMPONENTS, PINS, NETS sections
  - FIXED vs PLACED placement status
```

### Section 3: FeFET Technology

**Add (IF you have sources):**
```markdown
### Ferroelectric Technology References
- **HfO₂-Zr O₂ FeFETs:** [Cite specific papers if claiming 30 levels]
- **Multi-level Programming:** [Papers demonstrating this capability]
- **Retention Characteristics:** [Data on 10+ year retention if claimed]

**If no sources yet:**
```markdown
### FeFET Technology - Preliminary
- ⚠️ 30-level claim: Requires validation with literature search
- ⚠️ 10-year retention: Common industry target, needs specific citations
- ⚠️ Programming voltages: Placeholder values, require characterization
- **Action:** Conduct literature review of HfO₂-based FeFET papers 2020-2024
```

---

## 🎯 Enhancement Recommendations

### 1. Add "Assumptions & Limitations" Section

**Insert after line 66:**
```markdown
═══════════════════════════════════════════════════════════════
ASSUMPTIONS & LIMITATIONS
═══════════════════════════════════════════════════════════════

This plan makes the following assumptions that require validation:

1. **Cell Dimensions:** 
   - Assumes 0.46 × 2.72 μm fits FeFET structure
   - ⚠️ Requires: Actual Magic layout to verify DRC compliance

2. **Timing Values:**
   - All timing parameters are placeholders
   - ⚠️ Requires: SPICE simulation with FeFET compact model

3. **30-Level Operation:**
   - Claims based on literature (citations needed)
   - ⚠️ Requires: Experimental validation or cited sources

4. **OpenLane Compatibility:**
   - Assumes custom cells work with SKY130 PDK
   - ⚠️ Requires: Test flow with simple design first

5. **Behavioral Model:**
   - Verilog model is pass-through only
   - ⚠️ Requires: Actual FeFET behavioral model for simulation
```

### 2. Add Version Control to Generated Files

**In each generator function, add header:**
```go
// Generated by: FeCIM Design Suite Module 6
// Version: 0.1.0-alpha
// Date: %s
// Source: github.com/yourusername/ironlattice-vis
// References: See docs/eda/plan-demo6.md citations
```

### 3. Validation Checklist Addition

**Add after line 3122:**
```markdown
═══════════════════════════════════════════════════════════════
PRE-SUBMISSION VALIDATION CHECKLIST
═══════════════════════════════════════════════════════════════

Before claiming "working OpenLane integration":

□ LEF passes check_lef utility
□ Liberty passes liberty_parser validation
□ Verilog passes yosys syntax check
□ DEF passes check_def utility
□ OpenLane runs without fatal errors (warnings OK)
□ GDS generated and viewable in KLayout
□ DRC violations documented (zero expected for placeholder)
□ LVS clean (if SPICE netlist provided)

Test with minimal design first:
□ 2×2 array successfully completes OpenLane flow
□ Layout visualization confirms cell placement
□ Generated GDS matches expected dimensions
```

### 4. Add "Known Issues" Section

```markdown
═══════════════════════════════════════════════════════════════
KNOWN ISSUES & FUTURE WORK
═══════════════════════════════════════════════════════════════

Current limitations:

1. **No Physical Layout:** 
   - LEF defines abstract view only
   - Missing: Actual Magic .mag file
   - Impact: Cannot generate real GDS

2. **Placeholder Timing:**
   - Liberty uses dummy values
   - Impact: Timing analysis meaningless

3. **Simplified Behavioral Model:**
   - Doesn't model FeFET physics
   - Impact: Functional simulation inaccurate

4. **No SPICE Model:**
   - Cannot run analog verification
   - Impact: No PVT corner analysis

Future work required:
- [ ] Design actual FeFET cell in Magic
- [ ] Extract real SPICE netlist
- [ ] Characterize Liberty timing
- [ ] Implement behavioral Verilog-A model
- [ ] Validate with test chip
```

### 5. Proper OpenLane Flow Documentation

**Replace generic flow description with specific steps:**
```markdown
═══════════════════════════════════════════════════════════════
OPENLANE FLOW - SPECIFIC STEPS
═══════════════════════════════════════════════════════════════

Reference: OpenLane v2.0 Documentation
https://openlane.readthedocs.io/en/latest/

Step-by-step integration:

1. **Preparation:**
   ```bash
   mkdir -p designs/fecim_crossbar_4x4/{src,cells,placement}
   cp output/fecim_crossbar_4x4.v designs/fecim_crossbar_4x4/src/
   cp cells/fecim_bitcell/* designs/fecim_crossbar_4x4/cells/
   cp output/fecim_crossbar_4x4.def designs/fecim_crossbar_4x4/placement/
   cp output/config.json designs/fecim_crossbar_4x4/
   ```

2. **Validate Configuration:**
   ```bash
   cd $OPENLANE_ROOT
   ./flow.tcl -design fecim_crossbar_4x4 -config_file config.json -verbose 2 -synth_explore
   # Check for config errors before proceeding
   ```

3. **Run Flow:**
   ```bash
   ./flow.tcl -design fecim_crossbar_4x4
   # Or interactive:
   ./flow.tcl -design fecim_crossbar_4x4 -interactive
   ```

4. **Expected Warnings (OK):**
   - Timing violations (placeholder timing)
   - No clock tree (RUN_CTS=0)

5. **Expected Errors (NOT OK):**
   - Missing LEF
   - DEF syntax errors
   - Unresolved cell references

Reference: OpenLane troubleshooting guide
```

### 6. Add Traceability

**For each module/function, add:**
```go
// Specification: plan-demo6.md lines 134-198
// Implements: LEF generation for SKY130 compatibility
// Standards: LEF 5.8 (Si2/OpenAccess)
// Status: Placeholder - requires Magic layout validation
```

### 7. Citation for Yosys Validation

**In validation section:**
```go
// Yosys validation reference:
// - Yosys Open SYnthesis Suite: https://yosyshq.net/yosys/
// - Command: yosys -p 'read_verilog -lib <cell> ; read_verilog <design>'
// - Expected: 0 errors for structural netlist
```

### 8. Add Bibliography Section

**At end of plan:**
```markdown
═══════════════════════════════════════════════════════════════
BIBLIOGRAPHY
═══════════════════════════════════════════════════════════════

## EDA Tools & Specifications

[1] OpenLane Documentation. "Flow Configuration Variables."
    https://openlane.readthedocs.io/en/latest/reference/configuration.html
    Accessed: 2026-01-24

[2] SkyWater Technology. "SKY130 Process Design Kit."
    https://skywater-pdk.readthedocs.io/
    Version: 2024.01+

[3] Si2/OpenAccess Coalition. "LEF/DEF 5.8 Language Reference."
    Library Exchange Format & Design Exchange Format Specifications.

[4] Synopsys. "Liberty Timing Format Reference Manual."
    Industry standard for timing libraries.

[5] YosysHQ. "Yosys Open SYnthesis Suite."
    https://yosyshq.net/yosys/
    Used for Verilog syntax validation.

## Ferroelectric Technology

[TODO] Add specific FeFET papers when claiming:
- 30-level programmability
- 10-year retention
- HfO₂-ZrO₂ material system
- Specific conductance ranges

## Project-Specific

[6] This implementation: 
    github.com/[your-username]/ironlattice-vis/tree/main/module6-eda
    
[7] Plan document:
    docs/eda/plan-demo6.md
    Generated: 2026-01-24
```

---

## ✅ Quick Checklist

Before submitting or sharing plan-demo6.md:

- [ ] Fixed PLACEMENT_CURRENT_DEF → FP_DEF_TEMPLATE
- [ ] Added "PLACEHOLDER" markers to all dummy values
- [ ] Cited OpenLane documentation for all config variables
- [ ] Cited SKY130 PDK for cell dimensions
- [ ] Added Assumptions & Limitations section
- [ ] Added Known Issues section
- [ ] Included bibliography
- [ ] Clarified what is implemented vs. theoretical
- [ ] Referenced all external tools (Yosys, Magic, OpenLane)
- [ ] Noted where FeFET-specific citations are needed

---

## 🎓 Academic Integrity Notes

**If this is for research/publication:**
1. Clearly separate "implemented" from "proposed"
2. Mark all placeholder values as "estimated" or "target"
3. Cite papers for any claimed FeFET characteristics
4. Include "Future Work" section for unverified claims
5. Use "preliminary" or "proof-of-concept" qualifiers appropriately

**If showcasing to industry:**
1. Emphasize the tooling/automation (your contribution)
2. Clearly state FeFET values are placeholders needing characterization
3. Focus on the EDA integration methodology
4. Cite all standards (LEF/DEF/Liberty) used

**Good practice:**
- "This work demonstrates OpenLane integration methodology..."
- "Placeholder timing values pending FeFET characterization..."
- "Based on SKY130 PDK site dimensions..."

**Avoid:**
- Claiming validated FeFET performance without data
- Presenting placeholders as real measurements
- Omitting limitations sections
- Using tools without citation

---

**Priority Actions:**
1. Fix PLACEMENT_CURRENT_DEF (Critical)
2. Add PLACEHOLDER markers (High)
3. Add citations section (High)
4. Add limitations section (Medium)
5. Add bibliography (Medium)
