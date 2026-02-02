# Module 6: EDA Tools

Educational Electronic Design Automation (EDA) toolchain for FeCIM (Ferroelectric Compute-in-Memory) chip design, from behavioral specification to **illustrative** layout artifacts. **Simulation-only; not tape-out or signoff ready.**

## Overview

Module 6 implements an RTL-to-layout **exploration flow** for FeCIM-based systems. It supports three distinct operation modes (Storage, Memory, Compute) and multiple process design kits (PDKs), enabling rapid design iteration **for education and research planning** (not fabrication).

**Key Capability**: Generate **educational** design artifacts with OpenLane-compatible outputs for learning and early exploration. External validation and signoff are required for any real fabrication.

## Features

### Three Operation Modes

- **Storage Mode**: High-density non-volatile memory replacing NAND flash. Optimizes for data retention and write endurance with a 30-level baseline (4.9 bits/cell, conference claim).

- **Memory Mode**: High-speed zero-refresh memory replacing DRAM. Targets 10ns access times with minimal power consumption.

- **Compute Mode**: Analog accelerator for neural network inference. Supports optional pre-loaded weights for AI model deployment and analog matrix-vector multiply operations.

### Export Formats

| Format | Extension | Purpose | Details |
|--------|-----------|---------|---------|
| JSON | .json | Complete design metadata | Configuration, cell assignments, statistics |
| CSV | .csv | Tabular cell data | One row per cell with conductance/resistance |
| SPICE | .sp | Analog simulation | ngspice/HSPICE resistor network netlist |
| Verilog | .v | Digital simulation & synthesis | Structural netlist with cell instantiation |
| DEF | .def | Physical placement | Design Exchange Format with fixed cell positions |

**Additional Formats** (via generate functions):
| Format | Extension | Purpose | Details |
|--------|-----------|---------|---------|
| LEF | .lef | Technology library | Abstract cell definitions for P&R tools |
| Liberty | .lib | Timing characterization | STA and synthesis timing information |
| SVG | .svg | Visual layout | Interactive scalable graphics representation |

### Process Design Kit Support

- **SKY130** (SkyWater 130nm, open-source) - Default, fully supported
- **GF180MCU** (GlobalFoundries 180nm, open-source)
- **IHP_SG13G2** (IHP 130nm SiGe BiCMOS)

### Architecture Support

- **Passive** (0T1R): Crossbar with word lines (WL) and bit lines (BL) only
- **1T1R**: 1 Transistor 1 Resistor with selector (WL, BL, SL)
- **2T1R**: 2 Transistor 1 Resistor with dual select (WL, BL, SL, CSL)

### OpenLane Integration

Generated designs are compatible with OpenLane's automated flow:
- RTL synthesis with Yosys
- Floorplanning and placement with detailed DEF support
- Automatic routing with global and detailed routers
- GDSII generation from routed layout

## Quick Start

### Basic Usage

Generate a 64x64 compute array with the 30-level demo baseline (conference claim):

```bash
go run ./cmd/eda-cli -mode compute -rows 64 -cols 64 -output ./output
```

Export with weights (optional):

```bash
go run ./cmd/eda-cli -mode compute -rows 64 -cols 64 \
  -input weights.json -output ./output -name my_design
```

### CLI Options

```
Usage: eda-cli [flags]

Mode Configuration:
  -mode string          Operation mode: storage, memory, or compute (default "compute")
  -input string         Input weights JSON file (optional, compute mode only)

Array Specifications:
  -rows int             Array rows (default 128)
  -cols int             Array cols (default 128)
  -levels int           Conductance levels 2-30 (default 30)

Technology & Architecture:
  -tech string          Technology: SKY130, GF180MCU, IHP_SG13G2 (default "SKY130")
  -arch string          Architecture: passive or 1T1R (default "passive")

Electrical Parameters:
  -vdd float            Supply voltage V (default 1.8)
  -gmin float           Min conductance μS (default 10.0)
  -gmax float           Max conductance μS (default 100.0)

Output Configuration:
  -name string          Design name for output files (default "fecim_array")
  -output string        Output directory (default "data")

Export Formats:
  -json bool            Export JSON mapping (default true)
  -csv bool             Export CSV cell assignments (default true)
  -spice bool           Export SPICE netlist (default true)
  -verilog bool         Export Verilog netlist (default true)
  -def bool             Export DEF placement (default true)

Note: LEF, Liberty, and SVG formats are available via the export package's
Generate functions for programmatic use.
```

### GUI Application

Interactive design builder with visualization and validation:

```bash
go run ./cmd/eda-gui
```

Features:
- Array parameter configuration
- Real-time design statistics
- Layout visualization
- Validation tools
- Design export management

## Package Structure

### pkg/compiler

Core design generation engine supporting all operation modes.

**Key Functions**:
- `NewArrayConfig(mode, rows, cols)` - Create configuration for specified mode
- `GenerateDesign(config)` - Generate complete design with optional weight mapping
- `GenerateBlank(config)` - Create unprogrammed array
- Weight quantization and mapping to 30 conductance levels (demo baseline; conference claim)

**Configuration Options**:
```go
config := compiler.NewArrayConfig(compiler.ModeCompute, 64, 64)
config.Technology = "SKY130"
config.With1T1R()  // Switch to 1T1R architecture
config.WithWeights(weights)  // Load pre-trained weights
design, err := compiler.GenerateDesign(config)
```

### pkg/export

Multi-format output generation for downstream tools.

**Public Export Functions**:
- `ExportJSON(design, path)` - Complete design metadata
- `ExportCSV(design, path)` - Spreadsheet-compatible cell listings
- `ExportSPICE(design, path, vdd)` - Analog simulation netlist
- `ExportVerilog(design, path)` - RTL/structural netlist
- `ExportDEF(design, path)` - Cell placement for P&R

**Generate Functions** (for programmatic use):
- `GenerateLEF(cellConfig)` - LEF format string for bitcell
- `GenerateLiberty(cellConfig)` - Liberty timing library
- `GenerateLayoutSVG(arrayConfig, svgConfig)` - SVG visualization
- `GenerateSPICE(design, vdd)` - SPICE netlist string

**Example**:
```go
// Export to files
export.ExportJSON(design, "design.json")
export.ExportVerilog(design, "design.v")
export.ExportDEF(design, "design.def")

// Generate format strings
lefStr := export.GenerateLEF(cellConfig)
libertyStr := export.GenerateLiberty(cellConfig)
svgStr := export.GenerateLayoutSVG(arrayConfig, export.DefaultSVGConfig())
```

### pkg/openlane

OpenLane flow integration and tool execution.

**Components**:
- `Config` - OpenLane tool configuration (PDK paths, timeouts)
- `Manager` - Manages tool execution mode (Docker or native)
- `Runner` - Executes OpenROAD scripts and synthesis tools
- `openlane_config.go` - Configuration generation for OpenLane flow

**Configuration**:
```go
cfg := openlane.DefaultConfig()
cfg.PDKRoot = os.Getenv("PDK_ROOT")  // e.g., ~/.volare
cfg.PreferredMode = openlane.ModeDocker
```

**PDK Setup**:
```bash
# Install volare for PDK management
pip install volare

# Enable SKY130 PDK
volare enable --pdk sky130 sky130A

# Set environment variable
export PDK_ROOT=~/.volare
```

### pkg/validation

Design verification and cross-checking tools.

**Validators**:
- `yosys.go` - RTL synthesis validation with Yosys
- `def_validator.go` - DEF file format and consistency checking
- `cross_check.go` - Multi-file consistency verification
- `openlane.go` - OpenLane flow validation
- `circuit_image.go` - Circuit diagram generation
- `layout_image.go` - Layout visualization

### pkg/gui

Fyne-based graphical interface for interactive design.

**Components**:
- `app.go` - Main application window and event handling
- `embedded.go` - Embeddable module for unified application
- `tabs/` - User interface tabs:
  - `builder_validation_tab.go` - Design parameter input and validation
  - `learn_tab.go` - Educational visualizations
  - `learn_visuals.go` - Array and cell visualizations
- `widgets/layout_canvas.go` - SVG-based layout rendering

### pkg/config

Type definitions and defaults for design parameters.

**Types**:
- `ArrayConfig` - FeCIM array design specification
- `CellConfig` - Individual bitcell parameters
- `PeripheralConfig` - DAC/ADC/TIA specifications

## Operation Modes

### Storage Mode

High-density NAND replacement with maximum data retention.

**Configuration**:
```go
cfg := compiler.NewArrayConfig(compiler.ModeStorage, 256, 256)
cfg.StorageConfig.RetentionYears = 10
cfg.StorageConfig.EnduranceCycles = 1000000
cfg.StorageConfig.ErrorCorrection = "SECDED"
```

**Output**: Uninitialized array ready for programming

### Memory Mode

High-speed DRAM replacement with zero refresh.

**Configuration**:
```go
cfg := compiler.NewArrayConfig(compiler.ModeMemory, 128, 128)
cfg.MemoryConfig.AccessTimeNs = 10.0
cfg.MemoryConfig.WriteTimeNs = 50.0
cfg.MemoryConfig.BandwidthGBps = 10.0
```

**Output**: Uninitialized array with fast access requirements

### Compute Mode

AI accelerator for neural network inference with optional weight pre-programming.

**No Weights (Unprogrammed)**:
```go
cfg := compiler.NewArrayConfig(compiler.ModeCompute, 64, 64)
design, err := compiler.GenerateDesign(cfg)
```

**With Weights (Pre-programmed)**:
```go
cfg := compiler.NewArrayConfig(compiler.ModeCompute, 64, 64)
cfg.WithWeights(weights)  // Optional pre-trained weights
design, err := compiler.GenerateDesign(cfg)
// Outputs include weight quantization PSNR and mapping statistics
```

**Weight Format**:
```json
{
  "name": "trained_model",
  "rows": 64,
  "cols": 64,
  "weights": [
    [0.1, -0.2, 0.3],
    [-0.4, 0.5, -0.6]
  ]
}
```

## OpenLane Integration

### Design Flow

1. **Generate Files**:
   ```bash
   go run ./cmd/eda-cli -mode compute -rows 64 -cols 64 \
     -input weights.json -output ./design
   ```
   Produces: `design.v`, `design.def`, `design.sp`

2. **Configure OpenLane**:
   Create or update `config.json`:
   ```json
   {
     "DESIGN_NAME": "fecim_compute",
     "VERILOG_FILES": "dir::design.v",
     "PLACEMENT_CURRENT_DEF": "dir::design.def",
     "SYNTH_ELABORATE_ONLY": 1,
     "PL_SKIP_INITIAL_PLACEMENT": 1,
     "PDK": "sky130A",
     "CLOCK_PERIOD": 10.0,
     "CLOCK_PORT": "clk"
   }
   ```

3. **Run OpenLane**:
   ```bash
   docker run -it -v $(pwd):/work ghcr.io/the-openroad-project/openlane:latest
   # Inside container:
   flow.tcl -design /work/design -config /work/config.json
   ```

4. **Output**:
   - RTL synthesis reports (`reports/`)
   - Placement statistics (`placement/`)
   - Routing results and congestion maps (`routing/`)
   - GDSII layout (`results/gds/design.gds`)

### Key Configuration Parameters

| Parameter | Purpose | Example |
|-----------|---------|---------|
| `PLACEMENT_CURRENT_DEF` | Fixed cell placement | `dir::design.def` |
| `PL_SKIP_INITIAL_PLACEMENT` | Skip auto-placement | 1 |
| `SYNTH_ELABORATE_ONLY` | Structural netlist only | 1 |
| `CLOCK_PERIOD` | Target timing closure | 10 ns |
| `PDK` | Process design kit | sky130A, gf180 |

## Examples

### Minimal Storage Array

```bash
go run ./cmd/eda-cli \
  -mode storage \
  -rows 256 \
  -cols 256 \
  -output ./storage_design
```

### Compute Array with Pre-trained Weights

```bash
go run ./cmd/eda-cli \
  -mode compute \
  -rows 128 \
  -cols 128 \
  -input trained_weights.json \
  -output ./compute_design \
  -name neural_net_v1
```

### 1T1R Memory Design

```bash
go run ./cmd/eda-cli \
  -mode memory \
  -rows 512 \
  -cols 512 \
  -arch 1t1r \
  -output ./memory_1t1r
```

### Custom Electrical Parameters

```bash
go run ./cmd/eda-cli \
  -mode compute \
  -rows 64 \
  -cols 64 \
  -gmin 0.5 \
  -gmax 200.0 \
  -vdd 2.5 \
  -output ./custom_design
```

## Design Output Files

### Standard Exports

**Design Metadata** (`*_design.json`):
```json
{
  "config": {
    "mode": "compute",
    "array_rows": 64,
    "array_cols": 64,
    "levels": 30,
    "technology": "SKY130"
  },
  "cells": [
    {
      "row": 0, "col": 0,
      "level": 15,
      "conductance": 50.5,
      "resistance": 19801.98,
      "initial_weight": 0.5
    }
  ],
  "stats": {
    "total_cells": 4096,
    "active_cells": 4096,
    "area_mm2": 0.0234,
    "power_mw": 12.5
  }
}
```

**Cell Assignments** (`*_cells.csv`):
```
row,col,level,conductance,resistance,weight
0,0,15,50.5,19801.98,0.5
0,1,8,25.2,39682.54,-0.1
```

**SPICE Netlist** (`*.sp`):
- Complete crossbar resistor network
- Word line driver circuits
- Bit line loading
- Suitable for analog simulation

**Verilog Netlist** (`*.v`):
- Structural instantiation of cells
- Pin-accurate connectivity
- Compatible with Yosys synthesis
- Supports digital simulation with iverilog

**DEF Placement** (`*.def`):
- Cell positions with FIXED keyword
- Exact coordinates from generated layout
- Preserves deterministic placement for OpenLane

**Additional Formats** (available via generate functions):
- **LEF Library** - Abstract cell geometry, pin definitions, technology site definitions for OpenLane P&R
- **Liberty Timing** - CMOS timing library format with cell delay arcs and power characterization (timing values are placeholders)
- **Layout Visualization** - SVG format with interactive crossbar visualization, cell color-coding, and wire routing display

## Testing

Run all tests:

```bash
go test ./module6-eda/...
```

Test specific packages:

```bash
# Compiler tests
go test ./module6-eda/pkg/compiler/...

# Export format tests
go test ./module6-eda/pkg/export/...

# Configuration tests
go test ./module6-eda/pkg/config/...
```

## Important Notes

### Timing Characterization

All timing values in generated Liberty files are placeholders. Production designs require:

1. SPICE simulation with actual FeFET compact model
2. Transient and DC analysis across corners
3. Timing arc extraction (rise/fall delays, slew rates)
4. Temperature and voltage derating
5. Integration with STA tool (OpenROAD)

See `pkg/export/liberty.go` for detailed characterization requirements.

### GDS Generation

Module 6 generates Verilog and DEF for place-and-route. To produce GDSII:

1. Complete layout with OpenLane (automated) or OpenROAD (manual)
2. Run physical verification (DRC/LVS) with Magic or Calibre
3. Generate GDS from final routed layout

The generated LEF file provides abstract cell views for P&R tools but does not contain full physical layout (requires Magic .mag file for real designs).

### PDK Requirements

To use non-SKY130 technologies:

1. Install target PDK via volare or directly
2. Set `PDK_ROOT` environment variable
3. Update OpenLane config with correct PDK variant
4. Adjust cell pitch and timing parameters as needed

### Fabrication Readiness

Module 6 provides:
- Verified RTL and DEF for OpenLane
- Technology-independent topology
- Multi-format exports for different tools

Missing for real tape-out:
- Actual transistor layout (requires custom Magic design)
- Parametric test structures
- Redundancy and repair logic
- Power delivery network (PDN) design
- IO ring and pad frame design

## Architecture References

- **Passive Crossbar**: Simplest form, WL/BL addressing only
- **1T1R**: Single selector transistor per cell, enables read isolation
- **2T1R**: Dual transistors for independent row/column selection, enables row/column parallel reads

## Related Documentation

- `docs/development/scriptReference.md` - Function lookup and API reference
- `docs/development/TESTING.md` - Comprehensive testing guide
- `docs/comparison/HONESTY_AUDIT.md` - Technology accuracy and verification

## License

See LICENSE file in repository root.
