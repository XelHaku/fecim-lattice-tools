# OpenLane Integration Guide

This guide explains how to integrate FeCIM Design Suite outputs into the OpenLane digital design flow.

## Table of Contents

1. [Overview](#overview)
2. [Integration Strategies](#integration-strategies)
3. [Configuration Variables Reference](#configuration-variables-reference)
4. [DEF File Integration](#def-file-integration)
5. [Custom Cell Integration](#custom-cell-integration)
6. [Step-by-Step Workflow](#step-by-step-workflow)
7. [Interactive Mode Commands](#interactive-mode-commands)
8. [Troubleshooting](#troubleshooting)

---

## Overview

OpenLane is an automated RTL-to-GDSII flow built on open-source tools (Yosys, OpenROAD, Magic, KLayout, netgen). The FeCIM Design Suite generates files that can be injected at various stages of this flow.

### OpenLane Flow Stages

```
┌─────────────────────────────────────────────────────────────────────┐
│  1. SYNTHESIS (Yosys)                                               │
│     RTL → Gate-level netlist                                        │
│     Key var: SYNTH_ELABORATE_ONLY, VERILOG_FILES_BLACKBOX           │
├─────────────────────────────────────────────────────────────────────┤
│  2. FLOORPLAN (OpenROAD)                                            │
│     Die area, IO placement, power grid                              │
│     Key var: FP_DEF_TEMPLATE, DIE_AREA, FP_SIZING                   │
├─────────────────────────────────────────────────────────────────────┤
│  3. PLACEMENT (OpenROAD)                                            │
│     Global + detailed placement                                     │
│     Key var: PLACEMENT_CURRENT_DEF, PL_SKIP_INITIAL_PLACEMENT  ◄────┤
├─────────────────────────────────────────────────────────────────────┤ FeCIM
│  4. CTS (OpenROAD)                                                  │ injection
│     Clock tree synthesis                                            │ points
│     Key var: CTS_CURRENT_DEF                                        │
├─────────────────────────────────────────────────────────────────────┤
│  5. ROUTING (OpenROAD)                                              │
│     Global + detailed routing                                       │
│     Key var: ROUTING_CURRENT_DEF                                    │
├─────────────────────────────────────────────────────────────────────┤
│  6. SIGNOFF (Magic, netgen, KLayout)                                │
│     DRC, LVS, antenna checks, GDSII generation                      │
│     Key var: EXTRA_GDS_FILES, EXTRA_LEFS                            │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Integration Strategies

### Strategy 1: DEF Template (Recommended for Pin Alignment)

Use `FP_DEF_TEMPLATE` to copy pin locations and die area from a pre-defined DEF.

**When to use:** When you need exact pin placement matching a parent design.

```json
{
  "FP_DEF_TEMPLATE": "/path/to/fecim_crossbar.def"
}
```

**What happens:**
- Die area is copied from template
- Pin names and locations are replicated (excluding power/ground)
- Standard cells are still placed by OpenROAD

### Strategy 2: Placement DEF Injection (Recommended for Fixed Cells)

Use `PLACEMENT_CURRENT_DEF` to inject a pre-placed DEF at the placement stage.

**When to use:** When FeCIM cells must remain at exact positions (FIXED placement).

```json
{
  "PLACEMENT_CURRENT_DEF": "/path/to/fecim_crossbar.def",
  "PL_SKIP_INITIAL_PLACEMENT": 1
}
```

**What happens:**
- OpenLane skips its own placement algorithm
- Your pre-placed cells (marked FIXED) remain in position
- Routing proceeds with your placement

### Strategy 3: Custom Macro Integration

Treat the FeCIM crossbar as a pre-hardened macro.

**When to use:** When integrating FeCIM as a black box within a larger design.

```json
{
  "VERILOG_FILES_BLACKBOX": "/path/to/fecim_crossbar.v",
  "EXTRA_LEFS": "/path/to/fecim_crossbar.lef",
  "EXTRA_GDS_FILES": "/path/to/fecim_crossbar.gds",
  "EXTRA_LIBS": "/path/to/fecim_crossbar.lib"
}
```

---

## Configuration Variables Reference

### Custom Cell/Macro Files

| Variable | Format | Description |
|----------|--------|-------------|
| `EXTRA_LEFS` | String (paths) | LEF files for custom macros. Used in placement and routing. |
| `EXTRA_GDS_FILES` | String (paths) | GDS files for custom macros. Used during tape-out/GDSII merge. |
| `EXTRA_LIBS` | String (paths) | Liberty files for timing analysis. Optional but recommended. |
| `VERILOG_FILES_BLACKBOX` | String (paths) | Verilog files to black-box (not synthesize). |

**Example (config.json):**
```json
{
  "EXTRA_LEFS": "/home/user/fecim/cells/fecim_bit.lef",
  "EXTRA_GDS_FILES": "/home/user/fecim/cells/fecim_bit.gds",
  "EXTRA_LIBS": "/home/user/fecim/cells/fecim_bit.lib",
  "VERILOG_FILES_BLACKBOX": "/home/user/fecim/cells/fecim_bit.v"
}
```

**Example (config.tcl):**
```tcl
set ::env(EXTRA_LEFS) "/home/user/fecim/cells/fecim_bit.lef"
set ::env(EXTRA_GDS_FILES) "/home/user/fecim/cells/fecim_bit.gds"
set ::env(EXTRA_LIBS) "/home/user/fecim/cells/fecim_bit.lib"
set ::env(VERILOG_FILES_BLACKBOX) "/home/user/fecim/cells/fecim_bit.v"
```

### Stage-Specific DEF Variables

OpenLane tracks DEF files through stages. You can inject custom DEFs at any stage:

| Variable | Stage | Description |
|----------|-------|-------------|
| `CURRENT_DEF` | Global | Currently active DEF file |
| `PLACEMENT_CURRENT_DEF` | Placement | DEF used at start of placement |
| `CTS_CURRENT_DEF` | CTS | DEF used at start of clock tree synthesis |
| `ROUTING_CURRENT_DEF` | Routing | DEF used at start of routing |
| `PARSITICS_CURRENT_DEF` | Extraction | DEF used for parasitic extraction |
| `LVS_CURRENT_DEF` | LVS | DEF used for layout vs schematic |
| `DRC_CURRENT_DEF` | DRC | DEF used for design rule checking |

### Placement Control

| Variable | Default | Description |
|----------|---------|-------------|
| `PL_SKIP_INITIAL_PLACEMENT` | 0 | Skip initial placement (use with pre-placed DEF) |
| `PL_BASIC_PLACEMENT` | 0 | Use basic placement (for tiny designs <100 cells) |
| `PL_RANDOM_GLB_PLACEMENT` | 0 | Use random global placement |
| `PL_RANDOM_INITIAL_PLACEMENT` | 0 | Random initial + RePlAce refinement |

### Floorplan Control

| Variable | Default | Description |
|----------|---------|-------------|
| `FP_DEF_TEMPLATE` | None | DEF file to use as template for die area and pins |
| `FP_SIZING` | "relative" | Die sizing mode: "relative" or "absolute" |
| `DIE_AREA` | None | Absolute die area: "x0 y0 x1 y1" (microns) |
| `CORE_AREA` | None | Absolute core area: "x0 y0 x1 y1" (microns) |

### Synthesis Control

| Variable | Default | Description |
|----------|---------|-------------|
| `SYNTH_ELABORATE_ONLY` | 0 | Only elaborate, don't map to gates (for structural netlists) |
| `SYNTH_NO_FLAT` | 0 | Don't flatten hierarchy during synthesis |

---

## DEF File Integration

### DEF Format Basics

The DEF (Design Exchange Format) describes physical layout. Key sections:

```def
VERSION 5.8 ;
DIVIDERCHAR "/" ;
BUSBITCHARS "[]" ;

DESIGN fecim_crossbar_8x8 ;
UNITS DISTANCE MICRONS 1000 ;

DIEAREA ( 0 0 ) ( 50000 50000 ) ;

COMPONENTS 64 ;
  - cell_0_0 fecim_bit + FIXED ( 1000 1000 ) N ;
  - cell_0_1 fecim_bit + FIXED ( 1460 1000 ) N ;
  - cell_0_2 fecim_bit + PLACED ( 1920 1000 ) N ;
  ...
END COMPONENTS

PINS 16 ;
  - WL[0] + NET WL[0] + DIRECTION INPUT + USE SIGNAL
    + LAYER met3 ( -140 0 ) ( 140 400 )
    + FIXED ( 0 5000 ) N ;
  ...
END PINS

END DESIGN
```

### FIXED vs PLACED Keywords

| Keyword | Meaning | Use Case |
|---------|---------|----------|
| `FIXED` | Cell position is locked; tools must not move it | FeCIM cells that must stay in exact grid positions |
| `PLACED` | Cell has a position but tools may adjust it | Cells that can be optimized during routing |
| `UNPLACED` | Cell has no position yet | Let tools decide placement |

**FeCIM cells should use FIXED** to maintain the crossbar grid structure.

### Coordinate System

- Origin (0, 0) is bottom-left
- Units are in database units (typically 1000 units = 1 micron)
- Specify `UNITS DISTANCE MICRONS 1000` for clarity

### Orientation Codes

| Code | Rotation | Mirror |
|------|----------|--------|
| N | 0° | No |
| S | 180° | No |
| E | 90° | No |
| W | 270° | No |
| FN | 0° | Yes (X-axis) |
| FS | 180° | Yes (X-axis) |
| FE | 90° | Yes (X-axis) |
| FW | 270° | Yes (X-axis) |

---

## Custom Cell Integration

### Required Files for Custom Cells

To add FeCIM cells to OpenLane, you need:

| File | Purpose | Required? |
|------|---------|-----------|
| `fecim_bit.lef` | Abstract view (pins, blockages, size) | Yes |
| `fecim_bit.gds` | Physical layout for GDSII merge | Yes |
| `fecim_bit.lib` | Liberty timing model | Recommended |
| `fecim_bit.v` | Verilog behavioral model | For simulation |
| `fecim_bit.mag` | Magic layout source | For DRC/LVS |
| `fecim_bit.spice` | SPICE netlist | For LVS |

### LEF File Structure

```lef
VERSION 5.8 ;
BUSBITCHARS "[]" ;
DIVIDERCHAR "/" ;

MACRO fecim_bit
  CLASS CORE ;
  FOREIGN fecim_bit ;
  ORIGIN 0 0 ;
  SIZE 0.46 BY 2.72 ;
  SYMMETRY X Y ;
  SITE unithd ;

  PIN WL
    DIRECTION INPUT ;
    USE SIGNAL ;
    PORT
      LAYER met1 ;
        RECT 0.0 0.0 0.1 2.72 ;
    END
  END WL

  PIN BL
    DIRECTION OUTPUT ;
    USE SIGNAL ;
    PORT
      LAYER met2 ;
        RECT 0.36 0.0 0.46 2.72 ;
    END
  END BL

  PIN VPWR
    DIRECTION INOUT ;
    USE POWER ;
    PORT
      LAYER met1 ;
        RECT 0.0 2.62 0.46 2.72 ;
    END
  END VPWR

  PIN VGND
    DIRECTION INOUT ;
    USE GROUND ;
    PORT
      LAYER met1 ;
        RECT 0.0 0.0 0.46 0.1 ;
    END
  END VGND

  OBS
    LAYER li1 ;
      RECT 0.05 0.1 0.41 2.62 ;
  END

END fecim_bit

END LIBRARY
```

### Liberty File Structure

```liberty
library(fecim_bit) {
  delay_model : table_lookup;
  time_unit : "1ns";
  voltage_unit : "1V";
  current_unit : "1mA";
  capacitive_load_unit(1, pf);

  cell(fecim_bit) {
    area : 1.2512;  /* 0.46 * 2.72 */

    pin(WL) {
      direction : input;
      capacitance : 0.001;
    }

    pin(BL) {
      direction : output;
      function : "WL";  /* Simplified: output follows input through resistance */
      max_capacitance : 0.1;

      timing() {
        related_pin : "WL";
        timing_type : combinational;
        cell_rise(scalar) { values("0.1"); }
        cell_fall(scalar) { values("0.1"); }
        rise_transition(scalar) { values("0.05"); }
        fall_transition(scalar) { values("0.05"); }
      }
    }

    pin(VPWR) {
      direction : inout;
      pg_type : primary_power;
    }

    pin(VGND) {
      direction : inout;
      pg_type : primary_ground;
    }
  }
}
```

---

## Step-by-Step Workflow

### 1. Generate FeCIM Files

Launch the unified app and use the EDA tab to generate Verilog and DEF files from your design configuration.

```bash
./launch.sh
# Select "EDA" tab → Configure → Export
```

### 2. Create OpenLane Design Directory

```bash
mkdir -p ~/OpenLane/designs/fecim_crossbar
cd ~/OpenLane/designs/fecim_crossbar

# Copy generated files
cp /path/to/output/crossbar.v src/
cp /path/to/output/crossbar.def .
cp /path/to/fecim_cells/*.lef cells/
cp /path/to/fecim_cells/*.gds cells/
```

### 3. Create Configuration File

**config.json:**
```json
{
  "DESIGN_NAME": "fecim_crossbar_32x32",
  "VERILOG_FILES": "dir::src/*.v",
  "CLOCK_PERIOD": 10,
  "CLOCK_PORT": "CLK",
  "CLOCK_NET": "CLK",

  "EXTRA_LEFS": "dir::cells/fecim_bit.lef",
  "EXTRA_GDS_FILES": "dir::cells/fecim_bit.gds",
  "VERILOG_FILES_BLACKBOX": "dir::cells/fecim_bit.v",

  "SYNTH_ELABORATE_ONLY": 1,

  "FP_SIZING": "absolute",
  "DIE_AREA": "0 0 200 200",

  "PLACEMENT_CURRENT_DEF": "dir::crossbar.def",
  "PL_SKIP_INITIAL_PLACEMENT": 1,

  "FP_PDN_ENABLE_RAILS": 0,
  "DESIGN_IS_CORE": 0
}
```

### 4. Run OpenLane Flow

```bash
cd ~/OpenLane

# Full automated flow
./flow.tcl -design fecim_crossbar

# Or interactive mode for debugging
./flow.tcl -design fecim_crossbar -interactive
```

### 5. Interactive Mode Session

```tcl
% package require openlane
% prep -design fecim_crossbar -tag test_run

# Run synthesis (elaborate only)
% run_synthesis

# Run floorplan
% run_floorplan

# Inject our DEF and skip placement
% set_def $::env(PLACEMENT_CURRENT_DEF)
% run_cts

# Continue with routing
% run_routing

# Signoff
% run_magic
% run_klayout
% run_lvs
```

---

## Interactive Mode Commands

### Useful Commands

| Command | Description |
|---------|-------------|
| `prep -design <name>` | Initialize design |
| `set_def <def_file>` | Set current DEF |
| `run_synthesis` | Run Yosys synthesis |
| `run_floorplan` | Run floorplanning |
| `run_placement` | Run placement |
| `run_cts` | Run clock tree synthesis |
| `run_routing` | Run routing |
| `run_magic` | Generate GDSII via Magic |
| `run_lvs` | Run layout vs schematic |
| `run_magic_drc` | Run DRC via Magic |

### Manual Macro Placement

```tcl
# Add macro at specific location
% add_macro_placement fecim_array_0 100 100 N

# Apply placements with FIXED flag
% manual_macro_placement -f
```

### Checking Results

```tcl
# View current DEF in KLayout
% open_in_klayout

# Generate summary report
% generate_final_summary_report
```

---

## Troubleshooting

### Common Errors

#### "EXTRA_LEFS not set correctly"

**Symptom:** `check_floorplan_missing_lef` fails

**Solution:** Verify paths are absolute and files exist:
```bash
ls -la /path/to/fecim_bit.lef
```

#### "Unplaced instances remain"

**Symptom:** Placement fails with unplaced cells

**Solution:** Ensure all cells in Verilog have corresponding LEF definitions:
```tcl
% check_floorplan_missing_pins
```

#### "DRC violations in custom cells"

**Symptom:** Magic DRC reports errors in FeCIM cells

**Solution:**
1. Check cell LEF matches GDS dimensions
2. Verify metal layers match PDK rules
3. Run standalone Magic DRC on cell GDS

#### "LVS mismatch"

**Symptom:** netgen reports port/net mismatches

**Solution:**
1. Verify Verilog port names match LEF pin names exactly
2. Check power/ground pin names (tool outputs VPWR/VGND; ensure your cells match)
3. Ensure `LVS_INSERT_POWER_PINS=1`

### Debug Tips

1. **Enable verbose logging:**
   ```tcl
   % prep -design fecim_crossbar -verbose 2
   ```

2. **Check intermediate DEFs:**
   ```tcl
   % puts $::env(CURRENT_DEF)
   ```

3. **Visualize in KLayout:**
   ```tcl
   % open_in_klayout -layout $::env(CURRENT_DEF)
   ```

4. **Run individual stages:**
   ```tcl
   % global_placement_or
   % detailed_placement_or
   ```

---

## References

- [OpenLane Documentation](https://openlane.readthedocs.io/)
- [OpenROAD Documentation](https://openroad.readthedocs.io/)
- [SKY130 PDK Documentation](https://skywater-pdk.readthedocs.io/)
- [DEF/LEF Reference Manual](https://si2.org/)

---

## Related Documentation

| Document | Description |
|----------|-------------|
| [eda.demo.md](./demo.md) | FeCIM Design Suite interface guide |
| [eda.opensource.md](../ecosystem/opensource-eda.md) | Open-source EDA ecosystem overview |
| [SKY130.md](../pdk/sky130.md) | SKY130 PDK integration specifics |
| [README.md](../README.md) | Module 6 overview and disclaimers |

---

**Part of the FeCIM Lattice Tools educational suite**
