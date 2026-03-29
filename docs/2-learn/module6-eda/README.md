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

### For Developers
- **[Physics Reference](./physics.md)** - Mapping model, abstraction levels
- **[Features](./features.md)** - Compiler, export formats
- **[Open-Source Tools](./tools.md)** - EDA tool ecosystem

---

## Module Contents

```
module6-eda/
├── pkg/compiler/
│   ├── compiler.go           # Array-to-layout compiler
│   ├── types.go              # IR definitions
│   └── tiling.go             # Weight matrix partitioning
├── pkg/export/
│   ├── spice.go              # SPICE netlist export
│   ├── verilog.go            # Verilog RTL export
│   ├── def.go                # DEF placement export
│   └── lef.go                # LEF macro export
├── pkg/validation/
│   └── drc.go                # Design rule checks
└── cmd/eda-cli/
    └── main.go               # CLI entry point
```

---

## Quick Start

### Export SPICE Netlist
```bash
fecim-lattice-tools eda --mode storage \
  --rows 128 --cols 128 \
  --export-spice crossbar.sp
```

### Export for OpenLane
```bash
fecim-lattice-tools eda --mode compute \
  --config design.json \
  --export-def layout.def \
  --export-lef macros.lef
```

### Full Flow
```bash
# 1. Compile high-level spec
fecim-eda compile --input weights.npy --output ir.json

# 2. Export for fabrication
fecim-eda export --input ir.json \
  --spice netlist.sp \
  --verilog array.v \
  --def placement.def

# 3. Run OpenLane
cd openlane && make mount
./flow.tcl -design fecim_array
```

---

## What You'll Learn

1. **Design Abstraction Levels**
   - Algorithm (weight matrices)
   - Architecture (tiled arrays)
   - RTL (structural Verilog)
   - Layout (DEF/LEF)
   - GDS (final silicon)

2. **Tiling and Mapping**
   - Partition large matrices into tiles
   - Map tiles to physical coordinates
   - Generate interconnect

3. **EDA File Formats**
   - SPICE: Circuit simulation
   - Verilog: RTL representation
   - DEF: Placement and routing
   - LEF: Cell abstracts
   - GDS: Final layout

4. **Design Flow Integration**
   - OpenLane/OpenROAD
   - Sky130 PDK
   - Custom cell libraries

---

## Design Flow

```
┌─────────────────────────────────────────────┐
│  Algorithm Level (NumPy/PyTorch)            │
│  weights.npy (M×N matrix)                   │
└────────────┬────────────────────────────────┘
             │ fecim-eda compile
             ▼
┌─────────────────────────────────────────────┐
│  Architecture Level (JSON IR)               │
│  - Tiling specification                     │
│  - Array dimensions                         │
│  - Conductance mapping                      │
└────────────┬────────────────────────────────┘
             │ fecim-eda export
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
│  - Timing constraints                       │
└────────────┬────────────────────────────────┘
             │ Place & Route (OpenROAD)
             ▼
┌─────────────────────────────────────────────┐
│  Layout Level (DEF)                         │
│  - Cell placement                           │
│  - Routing topology                         │
└────────────┬────────────────────────────────┘
             │ Final assembly
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
Export: MVM interface
```

---

## Tiling Example

```
Input: 512×512 weight matrix
Tile size: 128×128 array

Tiling:
  TilesRows = ceil(512/128) = 4
  TilesCols = ceil(512/128) = 4
  Total tiles = 16

Layout:
  ┌─────┬─────┬─────┬─────┐
  │ T00 │ T01 │ T02 │ T03 │
  ├─────┼─────┼─────┼─────┤
  │ T10 │ T11 │ T12 │ T13 │
  ├─────┼─────┼─────┼─────┤
  │ T20 │ T21 │ T22 │ T23 │
  ├─────┼─────┼─────┼─────┤
  │ T30 │ T31 │ T32 │ T33 │
  └─────┴─────┴─────┴─────┘

Each tile: 128×128 crossbar array
```

---

## Export Formats

### SPICE Netlist (.sp)

```spice
* FeCIM Crossbar Array
.subckt fecim_cell wl bl gnd
R_cell wl bl {1/conductance}
.ends

.subckt fecim_array_128x128
X_0_0 wl[0] bl[0] gnd fecim_cell conductance=50u
...
.ends
```

### Verilog RTL (.v)

```verilog
module fecim_array_128x128 (
  input [127:0] wordlines,
  output [127:0] bitlines,
  input vdd,
  input gnd
);
  // Cell instances
  fecim_cell cell_0_0 (.wl(wordlines[0]), .bl(bitlines[0]), ...);
  ...
endmodule
```

### DEF Placement (.def)

```def
DESIGN fecim_array_128x128 ;
UNITS DISTANCE MICRONS 1000 ;
DIEAREA ( 0 0 ) ( 1000000 1000000 ) ;

COMPONENTS 16384 ;
  - cell_0_0 fecim_cell + PLACED ( 0 0 ) N ;
  - cell_0_1 fecim_cell + PLACED ( 10000 0 ) N ;
  ...
END COMPONENTS
```

---

## Key Features

- **Automatic tiling:** Partition large arrays
- **Multi-format export:** SPICE, Verilog, DEF, LEF
- **OpenLane integration:** Compatible with Sky130 flow
- **Design rule checking:** Basic DRC validation
- **Parameterized cells:** Configurable array sizes
- **Hierarchical design:** Modular cell libraries

---

## PDK Integration

### Sky130 (Open-Source)

```bash
# Setup Sky130 PDK
export PDK_ROOT=/path/to/skywater-pdk
export PDK=sky130A

# Export with Sky130 constraints
fecim-eda export --pdk sky130 --tech-file sky130A.tech
```

### Custom PDKs

```bash
# Define custom technology
fecim-eda --tech-config my_pdk.json
```

See [tools.md](./tools.md) for PDK and EDA tool information.

---

## Validation

### Design Rule Checks (DRC)

```bash
# Run built-in DRC
fecim-eda validate --input layout.def --rules sky130

# Check:
# - Minimum spacing
# - Width constraints
# - Via stacking rules
```

### LVS (Layout vs Schematic)

```bash
# Export both netlist and layout
fecim-eda export --spice netlist.sp --def layout.def

# Run LVS with Magic/Netgen
magic -T sky130A -rcfile .magicrc layout.def
netgen -batch lvs netlist.sp layout.extracted.sp
```

---

## Documentation Index

| Document | Purpose | Audience |
|----------|---------|----------|
| [eli5.md](./eli5.md) | EDA basics | Beginners |
| [physics.md](./physics.md) | Abstraction levels | Developers |
| [features.md](./features.md) | Compiler, export | Developers |
| [tools.md](./tools.md) | EDA tool ecosystem | Researchers |

---

## Evidence Status

- **Demonstrated:** Export formats, tiling algorithm, OpenLane compatibility
- **Modeled:** Default cell parameters, timing estimates
- **Aspirational:** Full foundry signoff, production tapeout

---

## Related Modules

- **[Module 2: Crossbar](../module2-crossbar/README.md)** - Array source for layout
- **[Module 4: Circuits](../module4-circuits/README.md)** - Peripheral circuits for export

---

## Testing

```bash
go test ./module6-eda/pkg/compiler
go test ./module6-eda/pkg/export
go test ./module6-eda/pkg/validation
```

---

**Last Updated:** 2026-02-16
