# Module 6: EDA Tools and Layout Export

**Navigation:** [← Back to Learn](../README.md) | [ELI5](./eli5.md) | [Physics](./physics.md) | [Features](./features.md)

---

## Overview

Module 6 provides Electronic Design Automation (EDA) tools for compiling high-level array specifications into physical layouts. It exports netlists (SPICE, Verilog) and placement files (DEF, LEF) for integration with standard semiconductor design flows including OpenLane/OpenROAD.

**Key Concept:** Bridge the gap between architectural simulation and silicon implementation. Export crossbar designs in industry-standard formats for fabrication.

---

## Quick Links

### For Beginners
- **[ELI5 Explanation](./eli5.md)** - What is EDA?
- **[Integration Guide](../../eda/guides/integration.md)** - OpenLane workflow

### For Developers
- **[Physics Reference](./physics.md)** - Mapping model, abstraction levels
- **[Features](./features.md)** - Compiler, export formats
- **[CLI Reference](../../eda/references/cli-reference.md)** - Command-line tools

### For Researchers
- **[OpenLane Integration](../../eda/guides/integration.md)** - From FeCIM to GDS
- **[PDK Setup](./pdk/README.md)** - Sky130, other PDKs

---

## Module Contents

```
module6-eda/
├── cmd/
│   ├── eda-cli/              # CLI entry point
│   ├── eda-gui/              # Standalone GUI entry point
│   └── lattice-gen/          # Lattice generator tool
├── pkg/
│   ├── compiler/
│   │   ├── compiler.go       # Weight quantization + design generation
│   │   └── types.go          # IR type definitions
│   ├── config/
│   │   └── types.go          # ArrayConfig (rows, cols, arch, tech, …)
│   ├── export/               # 20+ export generators (see Export Formats)
│   │   ├── spice.go          # SPICE netlist
│   │   ├── array_verilog.go  # Array-level Verilog RTL
│   │   ├── cell_verilog.go   # Cell-level Verilog module
│   │   ├── def.go            # DEF placement
│   │   ├── lef.go            # LEF macro abstract
│   │   ├── liberty.go        # Liberty timing/power
│   │   ├── sdc.go            # SDC timing constraints
│   │   ├── openlane_config.go # LibreLane JSON config
│   │   ├── openlane_tcl.go   # OpenLane v1 TCL config
│   │   ├── crosssim.go       # CrossSim YAML + Python
│   │   ├── pyspice.go        # PySpice Python
│   │   ├── magic_drc.go      # Magic DRC script
│   │   ├── netgen.go         # Netgen LVS script
│   │   ├── svg.go            # SVG layout diagram
│   │   ├── summary.go        # Design summary text
│   │   ├── scripts.go        # Shell runner + Yosys/OpenROAD/KLayout scripts
│   │   ├── csv.go            # CSV cell conductance table
│   │   ├── json.go           # JSON design config
│   │   └── lattice_generator.go # Generate All + Export Package
│   ├── gui/
│   │   ├── app.go            # Main GUI application
│   │   ├── embedded.go       # Embedded file assets
│   │   ├── keyboard.go       # Keyboard shortcuts
│   │   ├── tabs/
│   │   │   ├── learn_tab.go              # Learn tab (EDA education)
│   │   │   ├── builder_validation_tab.go # Array builder + validation
│   │   │   ├── export_viewer_tab.go      # Export file viewer
│   │   │   ├── layout_visualizer_tab.go  # SVG layout visualizer
│   │   │   ├── flow_scripts_tab.go       # EDA flow scripts (16 formats)
│   │   │   ├── conductance_heatmap.go    # Conductance heatmap panel
│   │   │   ├── learn_visuals.go          # OpenLane flow diagram
│   │   │   ├── learn_visuals_array.go    # Isometric crossbar diagrams
│   │   │   ├── learn_visuals_cell.go     # Cell cross-section diagrams
│   │   │   ├── learn_visuals_transistor.go # Transistor diagrams
│   │   │   └── data/                     # Bundled 2×2 passive example design
│   │   └── widgets/
│   │       └── layout_canvas.go          # Interactive layout canvas widget
│   ├── layout/
│   │   ├── def_generator.go      # DEF from layout IR
│   │   ├── placement_routing.go  # Force-directed placement + BFS routing
│   │   └── verilog_generator.go  # Verilog from layout IR
│   ├── openlane/
│   │   ├── config.go             # OpenLane config structures
│   │   ├── manager.go            # Tool detection (Docker / native)
│   │   └── runner.go             # OpenROAD / Yosys / KLayout execution
│   ├── validate/                 # Design rule and netlist checks
│   │   ├── drc.go                # Design rule checks
│   │   ├── lvs.go                # Layout vs Schematic
│   │   ├── pdk_bridge.go         # PDK bridge
│   │   └── yosys.go              # Yosys synthesis check
│   └── validation/               # Export validation + cross-checks
│       ├── def_validator.go      # DEF parser / validator
│       ├── cross_check.go        # LEF / Liberty / Verilog consistency
│       ├── openlane.go           # OpenROAD placement check
│       ├── yosys.go              # Yosys hierarchy validation
│       ├── circuit_image.go      # Yosys schematic + OpenROAD layout image
│       └── layout_image.go       # KLayout layout image export
```

---

## Quick Start

### Launch the GUI

```bash
# From the repo root — Module 6 is integrated into the main app
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools
# Navigate to the "EDA" module in the sidebar
```

### Generate via CLI

```bash
# Build the CLI tool
go build -o fecim-eda-cli ./module6-eda/cmd/eda-cli

# Export a 128×128 compute array (SKY130, passive 0T1R)
./fecim-eda-cli --mode compute --rows 128 --cols 128 \
  --arch passive --tech SKY130 \
  --spice --verilog --def --export-json \
  --output ./my_design

# Export 64×64 1T1R array for GF180MCU with flow scripts
./fecim-eda-cli --mode storage --rows 64 --cols 64 \
  --arch 1t1r --tech GF180MCU \
  --scripts --output ./my_design

# Compute mode with pre-trained weights
./fecim-eda-cli --mode compute --input weights.json \
  --rows 64 --cols 64 --output ./my_design

# JSON output for scripting
./fecim-eda-cli --mode memory --rows 32 --cols 32 --json-output
```

### CLI Options Reference

| Flag | Default | Description |
|------|---------|-------------|
| `--mode` | `compute` | Operation mode: `storage`, `memory`, `compute` |
| `--arch` | `passive` | Architecture: `passive` (0T1R), `1t1r`, `2t1r` |
| `--tech` | `SKY130` | Technology: `SKY130`, `GF180MCU`, `IHP_SG13G2` |
| `--rows` | `128` | Array rows |
| `--cols` | `128` | Array columns |
| `--levels` | `30` | Conductance levels (2–30) |
| `--vdd` | `1.8` | Supply voltage (V) |
| `--gmin` | `10.0` | Min conductance (µS) |
| `--gmax` | `100.0` | Max conductance (µS) |
| `--input` | — | Weights JSON (compute mode, optional) |
| `--output` | `data` | Output directory |
| `--name` | `fecim_array` | Design name prefix |
| `--spice` | on | Export SPICE netlist |
| `--verilog` | on | Export Verilog RTL |
| `--def` | on | Export DEF placement |
| `--export-json` | on | Export JSON config |
| `--csv` | on | Export CSV cell table |
| `--scripts` | off | Export EDA flow scripts |
| `--json-output` | off | Print results as JSON to stdout |
| `--quiet` | off | Suppress informational output |
| `--config` | — | Load config from YAML/JSON file |

### Run OpenLane (after export)

```bash
# Generated example in pkg/gui/tabs/data/fecim_crossbar_2x2/
cd module6-eda/pkg/gui/tabs/data/fecim_crossbar_2x2/
bash run_flow.sh  # Requires yosys, klayout, openroad or Docker
```

---

## What You'll Learn

1. **Design Abstraction Levels**
   - Algorithm (weight matrices)
   - Architecture (tiled arrays)
   - RTL (structural Verilog)
   - Layout (DEF/LEF)
   - GDS (final silicon)

2. **Array Architectures**
   - Passive (0T1R): resistor crossbar, column-write only
   - 1T1R: row-select transistor + ferroelectric resistor
   - 2T1R: row + column transistors for half-select suppression

3. **EDA File Formats**
   - SPICE: Circuit simulation
   - Verilog: RTL representation
   - DEF: Placement and routing
   - LEF: Cell abstracts
   - Liberty: Timing and power
   - GDS: Final layout

4. **Design Flow Integration**
   - OpenLane/OpenROAD (LibreLane v2+ and legacy v1)
   - Sky130, GF180MCU, IHP SG13G2 PDKs
   - Custom cell libraries

---

## Design Flow

```
┌─────────────────────────────────────────────┐
│  Algorithm Level (NumPy/PyTorch)            │
│  weights.json (M×N matrix)                  │
└────────────┬────────────────────────────────┘
             │ fecim-eda-cli --mode compute --input
             ▼
┌─────────────────────────────────────────────┐
│  Architecture Level (JSON config)           │
│  - Array dimensions                         │
│  - Conductance mapping (30 levels)          │
│  - Technology parameters                    │
└────────────┬────────────────────────────────┘
             │ export generators
             ▼
┌─────────────────────────────────────────────┐
│  RTL Level (Verilog)                        │
│  - Module hierarchy                         │
│  - Interconnect                             │
│  - Cell instances                           │
└────────────┬────────────────────────────────┘
             │ Synthesis (Yosys)
             ▼
┌─────────────────────────────────────────────┐
│  Netlist Level (Gate-level)                 │
│  - Standard cells + macros                  │
│  - Timing constraints (SDC)                 │
└────────────┬────────────────────────────────┘
             │ Place & Route (OpenROAD)
             ▼
┌─────────────────────────────────────────────┐
│  Layout Level (DEF)                         │
│  - Cell placement (FIXED)                   │
│  - Routing topology                         │
└────────────┬────────────────────────────────┘
             │ Final assembly (KLayout / Magic)
             ▼
┌─────────────────────────────────────────────┐
│  Silicon Level (GDS)                        │
│  - Mask layers                              │
│  - Ready for fabrication                    │
└─────────────────────────────────────────────┘
```

---

## Operating Modes

### Storage Mode
```
Optimize for: Density, retention
Use case: Memory arrays
Export: SRAM-like interface
```

### Memory Mode
```
Optimize for: Read speed, low power
Use case: Cache replacement
Export: Standard memory controller
```

### Compute Mode
```
Optimize for: Throughput, energy efficiency
Use case: Neural network accelerator
Export: MVM interface (optional weight pre-programming)
```

---

## Export Formats

Module 6 generates **21 output formats** from a single `ArrayConfig`:

| Format | File | Tool |
|--------|------|------|
| SPICE netlist | `.sp` | ngspice, HSPICE |
| Verilog RTL (cell) | `_cell.v` | Yosys, Icarus |
| Verilog RTL (array) | `.v` | Yosys, Icarus |
| DEF placement | `.def` | OpenROAD |
| LEF macro abstract | `.lef` | OpenROAD |
| Liberty timing | `.lib` | OpenSTA |
| SDC constraints | `.sdc` | OpenSTA |
| LibreLane JSON | `config.json` | LibreLane (OpenLane v2+) |
| OpenLane v1 TCL | `config.tcl` | OpenLane v1 flow.tcl |
| Macro placement | `macros.cfg` | OpenLane v1 |
| SVG layout diagram | `.svg` | Browser, Inkscape |
| Design summary | `design_summary.txt` | Human-readable |
| JSON config | `.json` | Scripting |
| CSV cell table | `.csv` | Spreadsheet, Python |
| CrossSim YAML | `.yaml` | CrossSim |
| CrossSim Python | `.py` | CrossSim |
| PySpice Python | `.py` | PySpice |
| OpenVAF Verilog-A | `.va` | OpenVAF, ngspice (OSDI) |
| Magic DRC script | `.sh` | Magic VLSI |
| Netgen LVS script | `.sh` | Netgen |
| Shell runner | `run_flow.sh` | bash |

### SPICE Netlist (.sp)

```spice
* FeCIM Crossbar Array
.subckt fecim_bitcell wl bl gnd
R_ferro wl bl {1/conductance}
.ends fecim_bitcell

.subckt fecim_crossbar_4x4 WL[3:0] BL[3:0]
X_0_0 WL[0] BL[0] 0 fecim_bitcell
...
.ends
```

### Verilog RTL (.v)

```verilog
module fecim_crossbar_4x4 (
  input  [3:0] WL,
  inout  [3:0] BL
);
  fecim_bitcell c00 (.WL(WL[0]), .BL(BL[0]));
  // ... 16 cells total
endmodule
```

### DEF Placement (.def)

```def
DESIGN fecim_crossbar_4x4 ;
UNITS DISTANCE MICRONS 1000 ;
DIEAREA ( 0 0 ) ( 42400 12800 ) ;

COMPONENTS 16 ;
  - cell_0_0 fecim_bitcell + FIXED ( 10000 10000 ) N ;
  - cell_0_1 fecim_bitcell + FIXED ( 11840 10000 ) N ;
  ...
END COMPONENTS
```

---

## Key Features

- **Multi-architecture support:** Passive (0T1R), 1T1R, 2T1R
- **Multi-PDK support:** SKY130, GF180MCU, IHP SG13G2
- **18 export formats:** SPICE, Verilog, DEF, LEF, Liberty, SDC, LibreLane JSON, OpenLane v1 TCL, SVG, CrossSim, PySpice, and more
- **Force-directed placement:** Overlap-free cell layout with BFS Manhattan routing
- **OpenLane integration:** Compatible with LibreLane v2+ and legacy OpenLane v1
- **Cross-validation:** Automatic LEF/Liberty/Verilog pin consistency checks
- **Bundled example:** Complete 2×2 passive design at `pkg/gui/tabs/data/fecim_crossbar_2x2/`
- **GUI tabs:** Learn, Builder/Validation, Export Viewer, Layout Visualizer, Flow Scripts (16 formats)
- **Conductance heatmap:** 7 preset patterns (Gradient, Random, Checkerboard, Uniform Hi/Lo, Neural Weights, Sine Wave) with 30-level quantization histogram

---

## PDK Integration

### Sky130 (Open-Source)

```bash
# Set up Sky130 PDK via volare
export PDK_ROOT=~/.volare
volare enable --pdk sky130 sky130A

# Generate with Sky130 target
./fecim-eda-cli --tech SKY130 --arch passive --rows 64 --cols 64
```

### IHP SG13G2 (Open 130 nm BiCMOS)

```bash
git clone https://github.com/IHP-GmbH/IHP-Open-PDK.git ~/ihp-sg13g2-pdk
export IHP_PDK_ROOT=~/ihp-sg13g2-pdk

./fecim-eda-cli --tech IHP_SG13G2 --arch 1t1r --rows 32 --cols 32
```

### GF180MCU

```bash
volare enable --pdk gf180mcuD gf180mcuD
./fecim-eda-cli --tech GF180MCU --arch 2t1r --rows 16 --cols 16
```

See [pdk/README.md](./pdk/README.md) for PDK setup guide.

---

## Validation

### Design Rule Checks

```bash
# Built-in DRC (via Magic VLSI script export)
./fecim-eda-cli --scripts --output my_design
magic -T sky130A -rcfile .magicrc -batch my_design/fecim_bitcell_drc.tcl
```

### LVS (Layout vs Schematic)

```bash
# Export both netlist and layout
./fecim-eda-cli --spice --def --output my_design

# Run LVS with Netgen
cd my_design && netgen -batch lvs \
  fecim_crossbar.sp fecim_crossbar.def.extracted.sp
```

### Cross-File Consistency

The `validation` package automatically checks that LEF pin names, Liberty pin names, and Verilog port names are mutually consistent whenever files are generated.

---

## Documentation Index

| Document | Purpose | Audience |
|----------|---------|----------|
| [eli5.md](./eli5.md) | EDA basics | Beginners |
| [physics.md](./physics.md) | Abstraction levels | Developers |
| [features.md](./features.md) | Compiler, export | Developers |
| [integration.md](../../eda/guides/integration.md) | OpenLane workflow | Researchers |
| [cli-reference.md](../../eda/references/cli-reference.md) | Command reference | All |
| [pdk/README.md](./pdk/README.md) | PDK setup | Developers |

---

## Evidence Status

- **Demonstrated:** Export formats, placement algorithm, OpenLane compatibility, cross-validation
- **Modeled:** Default cell parameters, timing estimates
- **Aspirational:** Full foundry signoff, production tapeout

---

## Related Modules

- **[Module 2: Crossbar](../module2-crossbar/README.md)** - Array source for layout
- **[Module 4: Circuits](../module4-circuits/README.md)** - Peripheral circuits for export

---

## Testing

```bash
# Run all module6-eda tests (80+ test files, 14 packages)
go test ./module6-eda/... -count=1 -timeout 120s

# Individual packages
go test ./module6-eda/pkg/compiler/...
go test ./module6-eda/pkg/export/...
go test ./module6-eda/pkg/layout/...
go test ./module6-eda/pkg/openlane/...
go test ./module6-eda/pkg/validate/...
go test ./module6-eda/pkg/validation/...
go test ./module6-eda/pkg/gui/...
```

---

**Last Updated:** 2026-02-20
