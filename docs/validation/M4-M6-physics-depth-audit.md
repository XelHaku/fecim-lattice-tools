# Module 4 + Module 6: Physics & EDA Depth Audit

**Author:** Riju (Sage of Lightning)  
**Date:** 2026-02-12  
**Status:** Observations + actionable gaps

---

## Module 4: What's physically modeled vs what's missing

### ✅ Already modeled (real physics)
| Feature | Implementation | Physical basis |
|---------|---------------|----------------|
| Wire parasitics | `arraysim/types.go` CellGeometry | ρ·L/A for Cu, real SKY130 pitch (0.46×2.72 µm) |
| IR drop | `nonidealities.go` iterative solver | KCL nodal with segment currents + contact resistance |
| Cell conductance | `shared/physics/transfer.go` | G = Gmin + (Gmax−Gmin)·(P/Ps+1)/2, calibrated to HZO |
| Sneak paths | `enhanced.go` | Architecture-dependent: 0T1R (present), 1T1R (suppressed), 2T1R (eliminated) |
| Half-select disturb | `tab_unified_voltage.go` | Cumulative sub-threshold stress accumulation |
| DAC/ADC/TIA | `shared/peripherals/` | DNL/INL, gain, noise models |
| Noise | `shared/peripherals/noise.go` | Thermal (4kTR·BW), 1/f, shot (2qI·BW), quantization (LSB²/12) |
| Film geometry | `shared/physics/cell_geometry.go` | Thickness=10nm, Area=100nm², E=V/t, Q=P·A, G=σ·A/t |

### ❌ Missing: CMOS selector transistor model
- **Current state:** 1T1R/2T1R selectors are conductance masks (on/off boolean), not sized MOSFETs
- **What's needed:** W/L sizing, Vth, subthreshold leakage (Ioff), on-current (Ion), gate capacitance
- **Impact:** Without this, read current is idealized, write disturb through selector leakage isn't physical, and area estimates ignore transistor footprint

### ❌ Missing: Cell footprint calculation
- **Current state:** `CellGeometry` has film area (100 nm²) but not layout footprint
- **What's needed:** Total cell area = FeFET area + selector area + routing overhead
  - 0T1R: 4F² (just crosspoint)
  - 1T1R: ~6-12F² (FeFET + 1 transistor + contacts)
  - 2T1R: ~12-20F² (FeFET + 2 transistors + routing)
- **Impact:** Can't compute real array density or compare to SRAM (120-150F²)

### ❌ Missing: Technology node scaling in Module 4
- **Current state:** Hardcoded SKY130 values
- **What's needed:** Node selector (130nm, 65nm, 28nm, 14nm) that scales wire R, transistor parasitics, leakage

---

## Module 6: What exists vs what's needed for research-grade EDA

### ✅ Already implemented
| Feature | File | Quality |
|---------|------|---------|
| LEF generation (passive/1T1R/2T1R) | `export/lef.go` | Good — proper pin placement, OBS, site defs |
| Liberty timing (.lib) | `export/liberty.go` | ⚠️ PLACEHOLDER timing — marked honestly |
| Verilog netlists (cell + array) | `export/verilog.go`, `cell_verilog.go`, `array_verilog.go` | Good structural Verilog |
| SPICE netlist | `export/spice.go` | Basic — FeFET as resistor model only |
| DEF placement | `export/def.go`, `layout/def_generator.go` | Functional |
| SVG visualization | `export/svg.go` | Functional |
| OpenLane config | `export/openlane_config.go` | Config generation for flow |
| CSV export | `export/csv.go` | Basic data export |
| JSON export | `export/json.go` | Design serialization |
| Builder/Validation GUI tab | `gui/tabs/builder_validation_tab.go` | Functional |
| Learn tab (cell/array/transistor visuals) | `gui/tabs/learn_tab.go` + visuals | Educational |
| Technology support | CLI: SKY130, GF180MCU, IHP_SG13G2 | 3 PDKs |

### ❌ Critical gaps

#### 1. SPICE model is too simple
- FeFET modeled as fixed resistor (R = 1/G)
- No ferroelectric capacitance, no switching dynamics, no compact model
- Should at minimum use a voltage-dependent conductance or piecewise I-V

#### 2. Liberty timing is all placeholders
- Honestly marked, but still: rise/fall times are guesses, not characterized
- No NLDM (Non-Linear Delay Model) lookup tables
- No multi-corner support (only "typical")
- **Action:** Generate timing from SPICE characterization flow, or at minimum use published FeFET data

#### 3. No DRC/LVS validation
- LEF is abstract view only
- No Magic layout (.mag) generation
- No design rule checks against actual PDK rules
- No LVS (Layout vs Schematic) verification

#### 4. No power analysis
- Liberty has placeholder leakage (0.0003 nW)
- No dynamic power model (C·V²·f per cell)
- No array-level power estimation

#### 5. GUI only has 2 of planned tabs
- Builder & Validation + Learn
- Missing: Export viewer, Layout visualizer, Timing analysis, Corner analysis

#### 6. No back-annotation from Module 4
- Module 4's simulation results don't feed back into Module 6's characterization
- The wire parasitics computed in Module 4 should inform Module 6's timing/power

---

## Cross-module gap: Module 4 ↔ Module 6 integration

| What should flow | Direction | Current state |
|-----------------|-----------|---------------|
| Cell geometry (pitch, area) | M6 → M4 | Both have defaults but not linked |
| Wire resistance model | M4 → M6 | M4 computes from geometry; M6 uses hardcoded |
| Timing characterization | M4 sim → M6 Liberty | Not connected |
| Power numbers | M4 energy model → M6 Liberty | Not connected |
| Technology node | Shared | M4 hardcoded SKY130; M6 supports 3 PDKs |

---

## Observations (Riju's assessment)

1. **Module 4 is electrically honest** for the analog signal chain — real wire R, real noise, real ADC/DAC nonlinearity. The gap is the selector transistor and layout area.

2. **Module 6 has the right EDA skeleton** — LEF/Liberty/Verilog/SPICE/DEF covers the standard ASIC flow. But everything after Verilog is placeholder-quality. The Liberty file says so honestly, which is good.

3. **The SPICE model is the weakest link** — a fixed resistor doesn't capture what makes FeFET interesting (multi-state, voltage-dependent, hysteretic). Even a simple piecewise model would be a massive improvement.

4. **Module 4 and Module 6 should share a technology config** — right now they both hardcode SKY130 independently. A shared `TechnologyNode` type would unify cell dimensions, wire params, and transistor models.

5. **For a conference demo, the current state is defensible** — the physics where it matters (Preisach, crossbar, noise) is validated. For a paper submission, the SPICE and Liberty gaps need addressing.
