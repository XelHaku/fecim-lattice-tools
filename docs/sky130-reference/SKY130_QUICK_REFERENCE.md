# SKY130 Quick Reference for FeCIM EDA Design

> **Source:** [SkyWater SKY130 Open Source PDK Documentation](https://skywater-pdk.readthedocs.io/)  
> **License:** Apache 2.0  
> **Last Updated:** 2026-01-24

## Overview

The SkyWater SKY130 is a 130nm CMOS process technology made available as an open-source PDK (Process Design Kit). This reference provides key specifications needed for FeCIM (Ferroelectric Compute-in-Memory) custom cell design and integration with OpenLane ASIC design flow.

---

## Standard Cell Library Specifications

### sky130_fd_sc_hd (High Density)

| Parameter | Value | Description |
|-----------|-------|-------------|
| **Cell Height** | 2.72 μm | 8 horizontal routing tracks |
| **Site Width** | 0.46 μm | Unit cell placement grid |
| **Site Name** | `unithd` | For LEF SITE definition |
| **Power Rails** | VPWR, VGND | On met1, top/bottom |
| **Library Type** | Foundry-provided | General-purpose digital cells |

### Other Available Libraries

- **sky130_fd_sc_hdll**: High density, low leakage
- **sky130_fd_sc_hs**: High speed variant
- **sky130_fd_sc_ms**: Medium speed variant
- **sky130_fd_sc_ls**: Low speed variant
- **skysky130_fd_sc_lp**: Low power variant

---

## Metal Layer Stack

| Layer | Name | GDS# | Min Width | Min Space | Typical Use |
|-------|------|------|-----------|-----------|-------------|
| **Local Interconnect** | li1 | 67/20 | 0.17 μm | 0.17 μm | Cell-internal routing |
| **Metal 1** | met1 | 68/20 | 0.14 μm | 0.14 μm | Cell pins, power rails |
| **Metal 2** | met2 | 69/20 | 0.14 μm | 0.14 μm | Horizontal routing |
| **Metal 3** | met3 | 70/20 | 0.30 μm | 0.30 μm | Vertical routing |
| **Metal 4** | met4 | 71/20 | 0.30 μm | 0.30 μm | Horizontal routing |
| **Metal 5** | met5 | 72/20 | 1.60 μm | 1.60 μm | Power distribution |

### Via Stack

| Via | Connects | GDS# | Min Size |
|-----|----------|------|----------|
| **mcon** | li1 ↔ met1 | 67/44 | 0.17×0.17 μm |
| **via1** | met1 ↔ met2 | 68/44 | 0.15×0.15 μm |
| **via2** | met2 ↔ met3 | 69/44 | 0.20×0.20 μm |
| **via3** | met3 ↔ met4 | 70/44 | 0.20×0.20 μm |
| **via4** | met4 ↔ met5 | 71/44 | 0.80×0.80 μm |

---

## Power and Ground Rails

### Standard Cell Power Rails

| Rail | Layer | Width | Position |
|------|-------|-------|----------|
| **VPWR** | met1 | 0.48 μm | Top edge of cell |
| **VGND** | met1 | 0.48 μm | Bottom edge of cell |
| **VPWR** | li1 | 0.17 μm | Internal connection |
| **VGND** | li1 | 0.17 μm | Internal connection |

### Recommended Power Grid (Chip-level)

- **met4**: Vertical power stripes (300-500 μm pitch)
- **met5**: Horizontal power stripes (300-500 μm pitch)
- **met1/met2**: Cell row power distribution

---

## FeCIM Custom Cell Design Guidelines

### Pin Placement Recommendations

| Pin Type | Layer | Placement | Notes |
|----------|-------|-----------|-------|
| **Word Line (WL)** | met1 | Left/right edge, centered | Horizontal routing |
| **Bit Line (BL)** | met2 | Top/bottom edge, centered | Vertical routing |
| **Select Line (SL)** | met1 | Left/right edge | If needed |
| **Power (VPWR)** | met1 | Top edge | 0.48 μm width |
| **Ground (VGND)** | met1 | Bottom edge | 0.48 μm width |

### Cell Sizing Considerations

- **Minimum cell width**: 0.46 μm (1 site)
- **Recommended FeCIM cell width**: Multiple of 0.46 μm (e.g., 1.84 μm = 4 sites)
- **Cell height**: Should match 2.72 μm for standard cell compatibility
- **FeFET transistor sizing**: Depends on retention, endurance, and switching requirements

### Design Rule Constraints

- **Minimum poly width**: 0.15 μm
- **Minimum poly spacing**: 0.21 μm (poly layer 66/20)
- **Minimum diffusion width**: 0.15 μm
- **Minimum gate-to-contact spacing**: 0.055 μm
- **Metal1 minimum area**: 0.083 μm²

---

## OpenLane Integration

### Required Files for Custom Cells

1. **LEF file** (`*.lef`): Physical abstract view
   - Defines OBS (obstruction) layers
   - Pin locations and layers
   - Cell dimensions

2. **Liberty file** (`*.lib`): Timing characterization
   - Setup/hold times
   - Propagation delays
   - Power characteristics

3. **GDS file** (`*.gds`): Full layout
   - For final tapeout integration

4. **Verilog model** (`*.v`): Behavioral/functional model

### Key LEF Definitions

```tcl
# Example LEF snippet for FeCIM cell
MACRO fecim_1t1c_cell
  CLASS CORE ;
  FOREIGN fecim_1t1c_cell 0.0 0.0 ;
  ORIGIN 0.0 0.0 ;
  SIZE 1.84 BY 2.72 ;
  SYMMETRY X Y ;
  SITE unithd ;
  
  PIN WL
    DIRECTION INPUT ;
    USE SIGNAL ;
    PORT
      LAYER met1 ;
        RECT 0.0 1.20 0.14 1.52 ;
    END
  END WL
  
  PIN VPWR
    DIRECTION INOUT ;
    USE POWER ;
    PORT
      LAYER met1 ;
        RECT 0.0 2.48 1.84 2.72 ;
    END
  END VPWR
  
  # ... additional pins ...
END fecim_1t1c_cell
```

---

## Technology File Information

### Layer Purposes (GDS Datatypes)

| Datatype | Purpose | Usage |
|----------|---------|-------|
| 20 | drawing | Actual layer geometry |
| 44 | via/contact | Contact/via cuts |
| 5 | pin | Pin markers |
| 16 | label | Text labels |

### Critical Design Rules (Excerpt)

```
# Metal1 Rules
met1.1  : min width = 0.14 μm
met1.2  : min spacing = 0.14 μm
met1.3  : min area = 0.083 μm²
met1.4  : min width (for length > 0.28μm) = 0.17 μm

# Via1 Rules
via1.1  : exact size = 0.15 × 0.15 μm
via1.2  : min spacing = 0.17 μm
via1.3  : min overlap (met1/met2) = 0.055 μm (horizontal), 0.085 μm (vertical)
```

For complete design rules, consult: https://skywater-pdk.readthedocs.io/en/main/rules/periphery.html

---

## Electrical Characteristics

### Nominal Operating Conditions

| Parameter | Typical | Min | Max | Unit |
|-----------|---------|-----|-----|------|
| **Supply Voltage (VDD)** | 1.8 | 1.65 | 1.95 | V |
| **Temperature** | 25 | -40 | 85 | °C |
| **Junction Temp** | - | - | 125 | °C |

### Transistor Parameters (Typical, TT corner)

| Parameter | NMOS | PMOS | Unit |
|-----------|------|------|------|
| **Vth** | ~0.4 | ~-0.5 | V |
| **Min L** | 0.15 | 0.15 | μm |
| **Min W** | 0.42 | 0.42 | μm |

> **Note:** For FeCIM applications, FeFET characteristics (retention, coercive voltage, endurance) depend on the specific ferroelectric material integration.

---

## Tool Setup for SKY130

### Environment Setup

```bash
# Clone SKY130 PDK (for design tools)
git clone https://github.com/google/skywater-pdk.git
export PDK_ROOT=/path/to/skywater-pdk

# For OpenLane integration
export PDK=sky130A
export STD_CELL_LIBRARY=sky130_fd_sc_hd
```

### OpenLane Configuration Snippet

```tcl
# In your OpenLane config.tcl
set ::env(PDK) "sky130A"
set ::env(STD_CELL_LIBRARY) "sky130_fd_sc_hd"
set ::env(CLOCK_PERIOD) "10.0"
set ::env(CLOCK_PORT) "clk"
set ::env(DESIGN_NAME) "fecim_array"

# Include custom FeCIM cells
set ::env(EXTRA_LEFS) [glob $::env(DESIGN_DIR)/lef/*.lef]
set ::env(EXTRA_LIBS) [glob $::env(DESIGN_DIR)/lib/*.lib]
```

---

## References and Resources

### Official Documentation

1. **SkyWater SKY130 PDK Documentation**  
   https://skywater-pdk.readthedocs.io/en/main/  
   *Comprehensive process specifications and design rules*

2. **SKY130 PDK GitHub Repository**  
   https://github.com/google/skywater-pdk  
   *Full PDK files including tech files, models, and libraries*

3. **Standard Cell Library (sky130_fd_sc_hd)**  
   https://github.com/google/skywater-pdk-libs-sky130_fd_sc_hd  
   *LEF, Liberty, Verilog, and GDS for standard cells*

4. **OpenLane Documentation**  
   https://openlane.readthedocs.io/  
   *Automated RTL-to-GDSII flow using SKY130*

### Academic Citations

For academic papers referencing SKY130 use in FeCIM/memristor research:

```
@misc{skywater130pdk,
  author = {{Google} and {SkyWater Technology Foundry}},
  title = {{SKY130 Process Design Kit}},
  year = {2020},
  howpublished = {\url{https://github.com/google/skywater-pdk}},
  note = {Open source 130nm CMOS process, Apache 2.0 license}
}
```

---

## FeCIM-Specific Considerations

### Integration Challenges

1. **Material Stack**: SKY130 baseline doesn't include ferroelectric materials
   - Requires custom process integration (e.g., HfO₂-based FeFETs)
   - Post-fabrication compatibility depends on thermal budget

2. **Characterization**: Standard Liberty timing models don't capture:
   - Ferroelectric polarization states
   - Retention time dependencies
   - Write endurance degradation
   - → Custom behavioral models needed

3. **Verification**: OpenLane analog verification is limited
   - Use Xschem + ngspice for transistor-level FeCIM cell simulation
   - Use Magic for DRC/LVS verification

### Recommended Workflow

```
1. FeCIM Cell Design → Magic VLSI (layout)
2. Extract SPICE netlist → ngspice (verification)
3. Generate LEF abstract → Custom script
4. Characterize timing → Liberty NCX or custom characterizer
5. Integrate with OpenLane → Standard digital flow
```

---

## License and Attribution

This reference document is derived from the **SkyWater SKY130 PDK** which is licensed under **Apache License 2.0**.

```
Copyright 2020 SkyWater PDK Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
```

**Document Prepared By:** XelHaku  
**Project:** Multilayer Ferroelectric CIM Visualizer  
**Date:** January 24, 2026
