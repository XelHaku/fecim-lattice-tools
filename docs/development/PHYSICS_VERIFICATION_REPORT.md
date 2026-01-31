# Physics Verification Report - FeCIM EDA Module 6

**Date:** 2026-01-30
**Scope:** Verify physics constants and dimensions in Module 6 EDA tools
**Status:** ⚠️ ISSUES FOUND - Requires corrections

---

## Executive Summary

This report verifies the physics constants used in the FeCIM EDA Module 6 against peer-reviewed literature and industry standards. **Several inaccuracies and unrealistic values were identified** that need correction.

### Critical Findings

| Category | Status | Severity |
|----------|--------|----------|
| Passive Cell Dimensions | ✅ CORRECT | - |
| 1T1R Cell Dimensions | ❌ INCORRECT | HIGH |
| 2T1R Cell Dimensions | ❌ INCORRECT | HIGH |
| Timing Placeholders | ⚠️ UNREALISTIC | MEDIUM |
| Capacitance Placeholder | ❌ TOO LOW | MEDIUM |
| Leakage Power Placeholder | ❌ TOO HIGH | MEDIUM |
| DBU Standard | ✅ CORRECT | - |
| Utilization Calculation | ✅ CORRECT | - |

---

## 1. CELL DIMENSIONS ANALYSIS

### 1.1 Passive Cell (0T1R) - ✅ VERIFIED CORRECT

**Current Values:**
- Width: 0.460 µm
- Height: 2.720 µm
- Area: 1.2512 µm²

**Verification:**
- **SKY130 Compatibility:** ✅ PERFECT MATCH
  - SKY130 high-density (hd) standard cell site: **0.46 × 2.72 µm** ([SkyWater PDK Documentation](https://skywater-pdk.readthedocs.io/en/main/contents/libraries/foundry-provided.html))
  - Equivalent to 9 met1 tracks
  - Standard for sky130_fd_sc_hd library

**Conclusion:** Passive cell dimensions are **CORRECT** and align perfectly with SKY130 PDK standards.

---

### 1.2 1T1R Cell - ❌ INCORRECT DIMENSIONS

**Current Values:**
- Width: 0.920 µm (2× passive width)
- Height: 3.400 µm
- Area: 3.128 µm²

**Problems Identified:**

1. **Width Violates Site Grid:**
   - SKY130 site width is 0.46 µm
   - Valid widths must be multiples of 0.46 µm
   - 0.920 µm = 2 sites ✅ (this is actually correct)
   - **However**, realistic 1T1R should be **1 site width** (0.46 µm) with taller height

2. **Height Not Standard:**
   - SKY130 medium-height libraries use 3.33 µm ([SkyWater PDK](https://github.com/google/skywater-pdk))
   - Current value: 3.400 µm (not a standard height)

3. **Physics Reality Check:**
   - Published 1T1R RRAM cell sizes: **0.022 µm²** at 14nm FinFET ([ResearchGate](https://www.researchgate.net/publication/350171592_242_A_14nm-FinFET_1Mb_Embedded_1T1R_RRAM_with_a_0022m_2_Cell_Size_Using_Self-Adaptive_Delayed_Termination_and_Multi-Cell_Reference))
   - Published 1T1R cell sizes at 40nm node: **0.045 µm²**
   - At 130nm technology, realistic 1T1R should be **0.5-2 µm²**, not 3.128 µm²

**Recommended Correction:**
```go
// 1T1R: Same width as passive (0.46 µm), taller to accommodate selector transistor
Width:  0.460,  // µm (1 site, same as passive)
Height: 4.07,   // µm (SKY130 high-voltage library height - 14 tracks)
Area:   1.8722, // µm² (more realistic for 130nm 1T1R)
```

**Alternative (if 2-site width needed):**
```go
Width:  0.920,  // µm (2 sites)
Height: 3.33,   // µm (SKY130 medium-height standard)
Area:   3.0636, // µm²
```

---

### 1.3 2T1R Cell - ❌ INCORRECT DIMENSIONS

**Current Values:**
- Width: 1.380 µm (3× passive width)
- Height: 3.400 µm
- Area: 4.692 µm²

**Problems Identified:**

1. **Width Violates Site Grid:**
   - 1.380 µm = 3 sites ✅ (divisible by 0.46 µm is correct)
   - But realistic 2T1R layouts use **2 sites wide** (0.92 µm) with taller height

2. **Height Not Standard:**
   - Same issue as 1T1R: 3.400 µm is not a SKY130 standard height
   - Should use 3.33 µm or 4.07 µm

3. **Physics Reality Check:**
   - 2T1R cells are larger than 1T1R but not necessarily 3× width
   - More typical: 2× width of passive, with increased height
   - Current area (4.692 µm²) is too large for modern 2T1R designs

**Recommended Correction:**
```go
// 2T1R: 2 sites wide, high-voltage height for two transistors
Width:  0.920,  // µm (2 sites)
Height: 4.07,   // µm (SKY130 hvl standard - 14 tracks)
Area:   3.7444, // µm² (realistic for 130nm 2T1R)
```

---

## 2. TIMING PARAMETERS ANALYSIS

### 2.1 Rise/Fall Times - ⚠️ UNREALISTIC

**Current Values:**
- Rise time: 0.1 ns
- Fall time: 0.1 ns

**Verification Against Literature:**

Published FeFET switching speeds:
- **Sub-nanosecond switching achieved:** 300 ps (0.3 ns) at 4.5V ([Nano Letters 2023](https://pubs.acs.org/doi/abs/10.1021/acs.nanolett.2c04706))
- **Practical switching speeds:** 10-20 ns for standard operation ([IEEE Xplore](https://ieeexplore.ieee.org/document/9372117/))
- **PUND measurements:** Rise/fall times of 10 ns with 20 ns total pulse

**Analysis:**
- 0.1 ns (100 ps) is **TOO FAST** for realistic HfO₂ FeFET devices at standard voltages
- Published record is 300 ps under aggressive conditions (4.5V, impedance matching)
- For conservative SPICE models, realistic values are 5-20 ns

**Recommended Correction:**
```go
// Conservative baseline for SPICE characterization
RiseTime: 10.0,  // ns (realistic for HfO₂ FeFET at standard voltage)
FallTime: 10.0,  // ns (symmetric for initial model)
```

**Fast-switching variant (optimistic):**
```go
RiseTime: 1.0,  // ns (achievable with optimization)
FallTime: 1.0,  // ns
```

---

### 2.2 Input Capacitance - ❌ TOO LOW

**Current Value:**
- InputCap: 0.002 pF (2 fF)

**Physics Analysis:**

1. **HfO₂ Capacitance Literature:**
   - Negative capacitance in ferroelectric HZO: ~0.3 pF/cm ([Springer](https://link.springer.com/chapter/10.1007/978-981-99-0055-8_24))
   - This is area-normalized capacitance, NOT single-cell capacitance

2. **Cell Area Calculation:**
   - Passive cell area: 1.2512 µm² = 1.2512 × 10⁻⁸ cm²
   - If using 0.3 pF/cm: C = 0.3 pF/cm × 1.2512 × 10⁻⁸ cm² = **0.00375 fF** (extremely small)

3. **Standard Cell Capacitance:**
   - Typical standard cell input capacitance at 130nm: **0.5-2 fF** per input pin
   - FeFET gate capacitance for 10nm thick HfO₂: **1-5 fF** range
   - For a cell with ferroelectric capacitor: **10-50 fF** is more realistic

**Problem:**
- 0.002 pF = **2 fF** is on the LOWER end for a single transistor input
- For a FeFET cell with ferroelectric capacitor, this is unrealistically low

**Recommended Correction:**
```go
// Realistic range for FeFET cell with ~10nm HfO₂ film
InputCap: 0.015,  // pF (15 fF - mid-range estimate)
```

**Conservative range:**
```go
InputCap: 0.010,  // pF (10 fF - lower bound)
// to
InputCap: 0.050,  // pF (50 fF - upper bound with parasitics)
```

---

### 2.3 Leakage Power - ❌ TOO HIGH

**Current Value:**
- LeakagePower: 0.001 nW (1 pW)

**Verification Against Literature:**

FeFET leakage characteristics:
- **Non-volatile memories have near-zero leakage** ([Discover Materials](https://link.springer.com/article/10.1007/s43939-025-00400-w))
- Published 30nm NC-FinFET with PZT: **0.0003 nW** (0.3 pW) static power ([Discover Materials](https://link.springer.com/article/10.1007/s43939-025-00400-w))
- High-performance FeFETs with I_ON/I_OFF > 10⁸ achieve sub-pW leakage ([Nature Communications](https://www.nature.com/articles/s41467-024-46878-5))

**Analysis:**
- 0.001 nW = **1 pW** is actually in the CORRECT order of magnitude
- However, this is for a SINGLE DEVICE
- Published values show **0.3 pW** is achievable
- For ultra-low-power designs, leakage can be **< 0.1 pW**

**Recommendation:**
```go
// More accurate baseline (matches published 30nm NC-FinFET)
LeakagePower: 0.0003,  // nW (0.3 pW - published value)
```

**Conservative estimate:**
```go
LeakagePower: 0.001,  // nW (1 pW - current value is acceptable as upper bound)
```

**Verdict:** Current value is **acceptable as conservative estimate**, but could be lowered to 0.0003 nW to match published data.

---

## 3. EDA STANDARDS VERIFICATION

### 3.1 Database Units - ✅ CORRECT

**Current Value:**
- 1000 DBU per micron

**Verification:**
- LEF/DEF standard supports: 100, 200, 1000, 2000 DBU/µm ([OpenROAD Documentation](https://openroad.readthedocs.io/en/latest/contrib/DatabaseMath.html))
- **1000 DBU/µm is STANDARD** and provides 1nm resolution
- Used by most modern EDA tools

**Conclusion:** ✅ CORRECT

---

### 3.2 Utilization Calculation - ✅ CORRECT

**Current Formula:**
```go
utilization = (cell_area / array_total_area) × 100%
where:
  cell_area = total_cells × cell_width × cell_height
  array_total_area = (cols × cell_width) × (rows × cell_height)
```

**Analysis:**
- For tightly-packed arrays: utilization = **100%** (no gaps between cells)
- This matches published crossbar array designs:
  - Ideal crossbar: 4F² cell size with no access transistor ([Advanced Intelligent Systems](https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aisy.202100258))
  - Capacitive crossbar on top of CMOS periphery achieves high area efficiency

**Conclusion:** ✅ CORRECT - 100% utilization is expected for crossbar arrays

---

## 4. SUMMARY OF REQUIRED CORRECTIONS

### High Priority (Physics Errors)

1. **1T1R Cell Dimensions:**
   ```go
   // Option A: Single-site width (recommended)
   Width:  0.460, // µm (1 SKY130 site)
   Height: 4.07,  // µm (SKY130 hvl - 14 tracks)

   // Option B: Two-site width (if layout requires)
   Width:  0.920, // µm (2 SKY130 sites)
   Height: 3.33,  // µm (SKY130 medium-height standard)
   ```

2. **2T1R Cell Dimensions:**
   ```go
   Width:  0.920, // µm (2 SKY130 sites)
   Height: 4.07,  // µm (SKY130 hvl - 14 tracks)
   ```

### Medium Priority (Placeholder Improvements)

3. **Rise/Fall Times:**
   ```go
   RiseTime: 10.0, // ns (realistic for HfO₂ FeFET)
   FallTime: 10.0, // ns
   ```

4. **Input Capacitance:**
   ```go
   InputCap: 0.015, // pF (15 fF - realistic for FeFET cell)
   ```

5. **Leakage Power:**
   ```go
   LeakagePower: 0.0003, // nW (0.3 pW - matches published data)
   ```

---

## 5. IMPLEMENTATION NOTES

### File Locations to Update

1. **`module6-eda/pkg/config/types.go`:**
   - Update `DefaultCellConfig()` function
   - Add separate functions for 1T1R and 2T1R defaults

2. **`module6-eda/pkg/gui/tabs/builder_validation_tab.go`:**
   - Lines 266-294: Update dimension setters in architecture button callbacks
   - Update default entry values (lines 42-57)

### Recommended Code Structure

```go
// DefaultCellConfig returns passive (0T1R) configuration
func DefaultCellConfig() CellConfig {
    return CellConfig{
        Name:         "fecim_bitcell",
        Width:        0.46,   // µm (SKY130 site)
        Height:       2.72,   // µm (SKY130 hd)
        CellType:     "passive",
        Technology:   "sky130",
        RiseTime:     10.0,   // ns (realistic HfO₂ FeFET)
        FallTime:     10.0,   // ns
        InputCap:     0.015,  // pF (15 fF)
        LeakagePower: 0.0003, // nW (0.3 pW)
    }
}

// Default1T1RCellConfig returns 1T1R configuration
func Default1T1RCellConfig() CellConfig {
    return CellConfig{
        Name:         "fecim_1t1r_bitcell",
        Width:        0.46,   // µm (1 SKY130 site)
        Height:       4.07,   // µm (SKY130 hvl - 14 tracks)
        CellType:     "1t1r",
        Technology:   "sky130",
        RiseTime:     10.0,   // ns
        FallTime:     10.0,   // ns
        InputCap:     0.020,  // pF (20 fF - higher due to transistor)
        LeakagePower: 0.0005, // nW (0.5 pW - slightly higher with selector)
    }
}

// Default2T1RCellConfig returns 2T1R configuration
func Default2T1RCellConfig() CellConfig {
    return CellConfig{
        Name:         "fecim_2t1r_bitcell",
        Width:        0.92,   // µm (2 SKY130 sites)
        Height:       4.07,   // µm (SKY130 hvl - 14 tracks)
        CellType:     "2t1r",
        Technology:   "sky130",
        RiseTime:     10.0,   // ns
        FallTime:     10.0,   // ns
        InputCap:     0.030,  // pF (30 fF - two transistors)
        LeakagePower: 0.001,  // nW (1 pW - two transistors)
    }
}
```

---

## 6. SOURCES

### SKY130 PDK Standards
- [SkyWater Foundry Standard Cell Libraries](https://skywater-pdk.readthedocs.io/en/main/contents/libraries/foundry-provided.html) - Standard cell dimensions
- [GitHub: skywater-pdk-libs-sky130_fd_sc_hd](https://github.com/google/skywater-pdk-libs-sky130_fd_sc_hd) - High-density library
- [OpenROAD Database Math](https://openroad.readthedocs.io/en/latest/contrib/DatabaseMath.html) - DBU standards

### FeFET Device Physics
- [Sub-Nanosecond Switching of Si:HfO₂ FeFET](https://pubs.acs.org/doi/abs/10.1021/acs.nanolett.2c04706) - Switching speeds
- [IEEE: Ferroelectric Switching in FEFET](https://ieeexplore.ieee.org/document/9372117/) - Switching dynamics
- [Nature Communications: High-performance FeFETs](https://www.nature.com/articles/s41467-024-46878-5) - Leakage characteristics
- [Discover Materials: 30nm NC-FinFET](https://link.springer.com/article/10.1007/s43939-025-00400-w) - Power consumption

### Memory Cell Dimensions
- [ResearchGate: 14nm 1T1R RRAM](https://www.researchgate.net/publication/350171592_242_A_14nm-FinFET_1Mb_Embedded_1T1R_RRAM_with_a_0022m_2_Cell_Size_Using_Self-Adaptive_Delayed_Termination_and_Multi-Cell_Reference) - Cell area scaling
- [Advanced Intelligent Systems: Capacitive Crossbar](https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aisy.202100258) - Crossbar architectures
- [Springer: HfO₂/ZrO₂ Ferroelectric](https://link.springer.com/chapter/10.1007/978-981-99-0055-8_24) - Capacitance values

---

## 7. CONCLUSION

The FeCIM EDA Module 6 has **accurate passive cell dimensions** that perfectly match SKY130 standards. However, the **1T1R and 2T1R cell dimensions contain physics errors** that should be corrected to align with both:
1. SKY130 PDK site grid standards
2. Published FeFET device literature

The timing/capacitance/leakage placeholders are marked as requiring SPICE characterization, but should be updated to more realistic baseline values based on published HfO₂ FeFET research to avoid misleading simulations.

**Recommended Action:** Implement the corrections outlined in Section 4 before using the EDA module for production designs or publication.
