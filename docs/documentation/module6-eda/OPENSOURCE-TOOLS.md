<!-- Category: Open-Source Tools | Module: module6-eda | Reading time: ~5 min -->
# Module 6 Open-Source Tools: EDA Toolchain

> The open-source EDA ecosystem that takes a design from RTL to silicon.

---

## Tools Used in This Module

| Tool | Role |
|------|------|
| Go toolchain | Compiler, tiling, and export logic |
| Fyne | GUI for configuration and learning visuals |

---

## The Open-Source EDA Stack

These tools form a complete chip design flow. Module 6 generates
artifacts that feed into this pipeline.

### Flow Orchestration

| Tool | Description |
|------|-------------|
| **OpenLane** | End-to-end flow orchestration from RTL to GDS. Wraps Yosys, OpenROAD, Magic, and other tools into an automated pipeline. |
| **OpenLane 2** | Next-generation flow with Python-based configuration and improved modularity. |

### Synthesis

| Tool | Description |
|------|-------------|
| **Yosys** | Open-source synthesis and elaboration. Converts Verilog RTL to a gate-level netlist using standard cell libraries. Also used to verify structural Verilog from Module 6 exports. |

### Place and Route

| Tool | Description |
|------|-------------|
| **OpenROAD** | Physical implementation engine. Handles floorplanning, placement, clock tree synthesis, and routing. Accepts DEF/LEF files exported by Module 6. |
| **OpenSTA** | Static timing analysis. Reads Liberty (.lib) files and reports setup/hold timing. |

### Layout and Verification

| Tool | Description |
|------|-------------|
| **Magic** | VLSI layout editor and DRC/extraction tool. The standard tool for SKY130 design rule checking. |
| **KLayout** | Layout viewer and editor. Supports GDS, DEF, LEF viewing. Good for visual inspection of exported layouts. |
| **netgen** | LVS (Layout vs Schematic) tool. Compares extracted netlist against the original schematic. |

### Circuit Simulation

| Tool | Description |
|------|-------------|
| **ngspice** | Open-source SPICE simulator. Can import SPICE netlists exported by Module 6 for electrical verification. |
| **Xyce** | Parallel SPICE from Sandia. Scales to large crossbar arrays. |

### PDK

| Resource | Description |
|----------|-------------|
| **SKY130 PDK** | First fully open-source PDK (130 nm, SkyWater/Google). Provides standard cells, I/O, device models, and design rules. |
| **GF180MCU PDK** | Open-source 180 nm PDK from GlobalFoundries. Alternative to SKY130 for larger feature sizes. |
| **IHP SG13G2 PDK** | Open-source 130 nm SiGe BiCMOS PDK from IHP. Includes high-speed transistors. |

---

## Integration Path: Module 6 to OpenLane

```
  Step 1: Generate FeCIM artifacts
  $ go run ./cmd/fecim-lattice-tools eda cli \
      -mode compute -rows 8 -cols 8 \
      -name fecim_array_8x8 -output ./output \
      -verilog=true -def=true -spice=true

  Step 2: Copy to OpenLane design folder
  $ mkdir -p $OPENLANE_ROOT/designs/fecim_array/src
  $ cp output/fecim_array_8x8.v $OPENLANE_ROOT/designs/fecim_array/src/
  $ cp output/fecim_array_8x8.def $OPENLANE_ROOT/designs/fecim_array/

  Step 3: Configure OpenLane (config.json)
  - DESIGN_NAME: fecim_array_8x8
  - VERILOG_FILES: dir::src/*.v
  - FP_DEF_TEMPLATE: dir::fecim_array_8x8.def
  - EXTRA_LEFS / EXTRA_GDS / EXTRA_LIBS: custom cell files

  Step 4: Run OpenLane
  $ cd $OPENLANE_ROOT && ./flow.tcl -design fecim_array

  Step 5: Verify
  - DEF import succeeded
  - DRC / LVS / STA results from tool logs
  - Missing LEF/pin checks pass
```

---

## Handoff Boundaries

| Responsibility | Who |
|---------------|-----|
| Array mapping, structural export, placement template | Module 6 |
| Legalization, routing, clocking, signoff | OpenLane / OpenROAD |
| PDK-legal cell views (LEF/GDS/LIB) and characterization | Custom cell authoring |
| Fabrication | Foundry |

---

## Code Locations

| Path | Purpose |
|------|---------|
| `module6-eda/pkg/compiler/compiler.go` | Compiler |
| `module6-eda/pkg/export/spice.go` | SPICE export |
| `module6-eda/pkg/export/verilog.go` | Verilog export |
| `module6-eda/pkg/export/def.go` | DEF export |
| `module6-eda/pkg/export/lef.go` | LEF export |
| `module6-eda/pkg/validation/drc.go` | DRC validation |

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
