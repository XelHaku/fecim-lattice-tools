# FeCIM EDA Workflow: RTL to GDSII

Complete end-to-end workflow for transforming neural network weights into fabrication-ready GDSII files using the FeCIM compiler and OpenLane.

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Step 1: Prepare Weights](#step-1-prepare-weights)
4. [Step 2: Generate Design](#step-2-generate-design)
5. [Step 3: Review Outputs](#step-3-review-outputs)
6. [Step 4: Validate Design](#step-4-validate-design)
7. [Step 5: OpenLane Integration](#step-5-openlane-integration)
8. [Step 6: Run Synthesis and P&R](#step-6-run-synthesis-and-pr)
9. [Step 7: Generate GDSII](#step-7-generate-gdsii)
10. [Example Workflows](#example-workflows)
11. [Troubleshooting](#troubleshooting)

## Overview

The FeCIM workflow transforms trained neural network weights into a complete physical chip design:

```
Weights (weights.json)
         ↓
    [EDA CLI]
         ↓
Design Files (Verilog, DEF, SPICE)
         ↓
    [OpenLane]
         ↓
GDSII (fabrication-ready)
```

### Key Concepts

- **Quantization**: Neural network weights are quantized to 30 discrete conductance levels (demo baseline; conference claim)
- **30-Level System**: Demo baseline uses 30 analog states (~4.9 bits/cell; conference claim)
- **Three Modes**: Storage (NAND replacement), Memory (DRAM replacement), Compute (AI accelerator)
- **Passive Architecture**: Default crossbar with wordlines and bitlines (sneak path effects possible)
- **1T1R Architecture**: Transistor-gated array (optional, mitigates sneak paths)

## Prerequisites

### Required Tools

```bash
# Go build system
go version  # 1.18 or later

# SPICE simulator (for validation)
ngspice --version

# Verilog syntax checker
iverilog --version

# OpenLane (for GDSII generation)
# Installation: https://github.com/The-OpenROAD-Project/OpenLane
~/OpenLane/flow.tcl --version
```

### Required Files

- Neural network weights in JSON format
- Technology PDK (SKY130, GF180MCU, or IHP_SG13G2)
- OpenLane installation (if proceeding to GDSII)

### Environment Setup

```bash
# Add OpenLane to PATH
export PATH=$PATH:~/OpenLane

# Set PDK root
export PDK_ROOT=~/.volare
export PDK=sky130A
```

## Step 1: Prepare Weights

### Weight File Format

Weights must be provided as a JSON file with the following structure:

```json
{
  "name": "mnist_layer1",
  "rows": 32,
  "cols": 32,
  "weights": [
    [0.15, -0.23, 0.45, -0.12, ...],
    [-0.34, 0.27, -0.11, 0.56, ...],
    ...
  ]
}
```

**Fields:**
- `name` (string): Human-readable design name
- `rows` (int): Number of rows in weight matrix
- `cols` (int): Number of columns in weight matrix
- `weights` (2D array): Row-major float64 values

### Weight Constraints

- **Value Range**: Weights can be any real number. The compiler automatically finds min/max and normalizes.
- **Array Size**: Weight dimensions must fit within physical array (check `-rows` and `-cols` flags)
- **Positive/Negative**: The system supports both positive and negative weights through 30-level baseline quantization (conference claim)
- **PSNR Target**: Aim for 40+ dB quantization PSNR for minimal accuracy loss

### Quantization Process

The compiler automatically:

1. Finds min and max weight values
2. Normalizes weights to [-1, +1] range
3. Quantizes to integer level (0-29)
4. Maps level to conductance value (G_min to G_max)

No manual quantization needed—the CLI handles it automatically.

### Example: Creating Weight Files

**From PyTorch:**

```python
import json
import numpy as np
import torch

# After training...
model.eval()
weights = model.fc1.weight.detach().numpy()  # Shape: (output, input)

# Save first 32x32 subset
subset = weights[:32, :32]

data = {
    "name": "mnist_fc1_subset",
    "rows": 32,
    "cols": 32,
    "weights": subset.tolist()
}

with open("weights.json", "w") as f:
    json.dump(data, f, indent=2)
```

**From NumPy:**

```python
import numpy as np
import json

# Load from checkpoint
weights = np.load("model_weights.npy")

data = {
    "name": "my_layer",
    "rows": weights.shape[0],
    "cols": weights.shape[1],
    "weights": weights.tolist()
}

with open("weights.json", "w") as f:
    json.dump(data, f)
```

## Step 2: Generate Design

### Building the CLI Tool

```bash
# From repository root
cd module6-eda
go build -o eda-cli ./cmd/eda-cli
```

### Basic Usage

**With weights (compute mode):**

```bash
./eda-cli \
  -mode compute \
  -input weights.json \
  -output ./output \
  -rows 64 \
  -cols 64 \
  -levels 30 \
  -name my_design
```

**Without weights (unprogrammed array):**

```bash
./eda-cli \
  -mode compute \
  -output ./output \
  -rows 64 \
  -cols 64
```

**Storage mode (no weights applicable):**

```bash
./eda-cli \
  -mode storage \
  -output ./output \
  -rows 256 \
  -cols 256
```

### Command-Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-mode` | compute | Operation mode: `storage`, `memory`, or `compute` |
| `-input` | (empty) | JSON weights file (optional in compute mode) |
| `-output` | data | Output directory for generated files |
| `-name` | fecim_array | Design name (used in filenames) |
| `-rows` | 128 | Array rows |
| `-cols` | 128 | Array columns |
| `-levels` | 30 | Quantization levels (2-30, typically 30) |
| `-tech` | SKY130 | Technology node: `SKY130`, `GF180MCU`, `IHP_SG13G2` |
| `-arch` | passive | Architecture: `passive` or `1t1r` |
| `-vdd` | 1.8 | Supply voltage (V) |
| `-gmin` | 10.0 | Minimum conductance (μS) |
| `-gmax` | 100.0 | Maximum conductance (μS) |
| `-json` | true | Export JSON mapping |
| `-csv` | true | Export CSV cells |
| `-spice` | true | Export SPICE netlist |
| `-verilog` | true | Export Verilog netlist |
| `-def` | true | Export DEF placement |

### Example: Complete Compute Design

```bash
./eda-cli \
  -mode compute \
  -input examples/01-basic-8x8/weights.json \
  -output ./my_output \
  -rows 8 \
  -cols 8 \
  -levels 30 \
  -name fecim_8x8 \
  -tech SKY130 \
  -arch passive \
  -vdd 1.8 \
  -gmin 10.0 \
  -gmax 100.0 \
  -json \
  -csv \
  -spice \
  -verilog \
  -def
```

### Expected Output

```
FeCIM Array Generator - Compute Mode
========================================

Loaded weights: mnist_layer1 (32x32 = 1024 weights)

Configuration:
  Mode:         Compute
  Array Size:   32 × 32 (1024 cells)
  Technology:   SKY130
  Architecture: passive
  Levels:       30 (4.90 bits/cell)
  Conductance:  10.0 - 100.0 μS

Design Statistics:
  Total Cells:  1024
  Active Cells: 1024
  Area:         0.1234 mm²
  Est. Power:   0.10 mW
  Throughput:   8.23 GOPS
  Weight Range: [-0.9000, +0.9000]
  Quant PSNR:   42.35 dB

Exporting files to ./output/
  ✓ my_design_design.json
  ✓ my_design_cells.csv
  ✓ my_design.sp
  ✓ my_design.v
  ✓ my_design.def

Done!
```

## Step 3: Review Outputs

The compiler generates five key files:

### 1. mapping.json (fecim_array_design.json)

Complete design specification with all parameters and cell assignments.

**Contents:**
- Configuration (array size, technology, architecture)
- Cell assignments (row, col, level, conductance, resistance)
- Design statistics (area, power, quantization quality)

**Use case:** Version control, data analysis, programmatic consumption

**Example snippet:**

```json
{
  "config": {
    "mode": "compute",
    "array_rows": 8,
    "array_cols": 8,
    "architecture": "passive",
    "levels": 30,
    "g_min": 10.0,
    "g_max": 100.0
  },
  "cells": [
    {
      "row": 0,
      "col": 0,
      "level": 16,
      "conductance": 55.17,
      "resistance": 18125.5,
      "weight": 0.1,
      "program_v": 1.2
    },
    ...
  ],
  "stats": {
    "total_cells": 64,
    "active_cells": 64,
    "area_mm2": 0.00167,
    "power_mw": 0.01,
    "quant_psnr": 42.35
  }
}
```

### 2. cells.csv (fecim_array_cells.csv)

Tabular format for spreadsheet tools and data analysis.

**Columns:** row, col, weight, level, conductance_us, resistance_ohm

**Use case:** Import into Excel/pandas for analysis

**Example:**

```csv
row,col,weight,level,conductance_us,resistance_ohm
0,0,0.100,16,55.17,18125
0,1,-0.200,13,44.83,22319
0,2,0.300,19,64.65,15466
...
7,7,0.900,29,100.00,10000
```

### 3. crossbar.v (fecim_array.v)

Structural Verilog netlist instantiating FeCIM cells.

**Architecture:**
- Passive: `fecim_bit` cells with WL[], BL[] busses
- 1T1R: `fecim_1t1r` cells with WL[], BL[], SL[] busses

**Use case:** Synthesis with Yosys, simulation with iverilog, OpenLane integration

**Example (passive):**

```verilog
module fecim_crossbar_8x8 (
    input  wire [7:0]  WL,
    inout  wire [7:0]  BL,
    inout  wire        VPWR,
    inout  wire        VGND
);
    // Row 0
    fecim_bit cell_0_0 (.WL(WL[0]), .BL(BL[0]), .VPWR(VPWR), .VGND(VGND));
    fecim_bit cell_0_1 (.WL(WL[0]), .BL(BL[1]), .VPWR(VPWR), .VGND(VGND));
    // ... 64 cells total
endmodule
```

**Example (1T1R):**

```verilog
module fecim_crossbar_8x8 (
    input  wire [7:0]  WL,
    inout  wire [7:0]  BL,
    input  wire [7:0]  SL,
    inout  wire        VPWR,
    inout  wire        VGND
);
    // Row 0
    fecim_1t1r cell_0_0 (.WL(WL[0]), .BL(BL[0]), .SL(SL[0]), .VPWR(VPWR), .VGND(VGND));
    // ... cells
endmodule
```

### 4. crossbar.def (fecim_array.def)

Design Exchange Format file with physical placement of all cells.

**Key Feature:** All cells marked `FIXED` to preserve pre-calculated positions.

**Use case:** OpenLane integration with `PLACEMENT_CURRENT_DEF` or `FP_DEF_TEMPLATE`

**Structure:**

```def
VERSION 5.8 ;

DIVIDERCHAR "/" ;
BUSBITCHARS "[]" ;

DESIGN fecim_8x8 ;

UNITS DISTANCE MICRONS 1000 ;

DIEAREA 0 0 10000 10000 ;

COMPONENTS 64 ;
  - cell_0_0 fecim_bit + FIXED ( 10000 10000 ) N ;
  - cell_0_1 fecim_bit + FIXED ( 10460 10000 ) N ;
  - cell_0_2 fecim_bit + FIXED ( 10920 10000 ) N ;
  ...
  - cell_7_7 fecim_bit + FIXED ( 13680 12720 ) N ;
END COMPONENTS

NETS 0 ;
END NETS

END DESIGN
```

**Fields:**
- `FIXED`: Locks cell position (cannot be moved by router)
- Coordinates: (X, Y) in database units (default: 1000 units = 1 μm)
- Orientation: `N` (normal), `FN` (flipped), `W` (90° rotate), etc.

### 5. crossbar.sp (fecim_array.sp)

SPICE netlist for analog simulation and verification.

**Use case:** ngspice/HSPICE simulation, power analysis, timing verification

**Structure:** Resistor network representing crossbar with conductance values

```spice
* FeCIM Crossbar Netlist
* Array: 8x8
* Generated: 2025-01-27

.title fecim_crossbar

* Supply voltages
.param vdd=1.8

* Conductance values (in units of 1/R where R is resistance)
* Level 0: 10.0 uS
* Level 29: 100.0 uS
* Array cells mapped to resistors

* Row 0
R_0_0 BL_0 VGND 18125.5
R_0_1 BL_1 VGND 22319.2
...

* Simulation commands (if included)
.dc VDD 0 1.8 0.01
.print dc v(BL_*)
.end
```

## Step 4: Validate Design

### 4.1 Verilog Syntax Check

```bash
# Verify Verilog syntax without building
iverilog -o /dev/null output/my_design.v
echo $?  # 0 = success

# Or with more verbose output
iverilog -v output/my_design.v 2>&1 | head -20
```

### 4.2 SPICE Netlist Check

```bash
# Verify SPICE syntax
ngspice -b -c "source output/my_design.sp; quit" 2>&1

# Expected output: No errors
# Errors indicate bad conductance values or syntax
```

### 4.3 JSON Schema Validation

```bash
# Verify JSON is well-formed
python3 -m json.tool output/my_design_design.json > /dev/null
echo "JSON valid" || echo "JSON invalid"
```

### 4.4 CSV Integrity Check

```bash
# Count cells
wc -l output/my_design_cells.csv  # Should be (rows*cols + 1)

# Verify level range
awk -F, '{print $4}' output/my_design_cells.csv | sort -n | uniq
# Should show 0-29 (no gaps okay)

# Check for negative conductance or resistance
awk -F, '$5 < 0 {print "Bad conductance: " $0}' output/my_design_cells.csv
awk -F, '$6 < 0 {print "Bad resistance: " $0}' output/my_design_cells.csv
```

### 4.5 DEF Validation

```bash
# Check DEF format
grep "^COMPONENTS" output/my_design.def
# Should show: COMPONENTS 64 ;

# Count cells in DEF
grep -c "fecim_bit" output/my_design.def
# Should match array size (64 for 8x8)

# Verify FIXED placements
grep -c "FIXED" output/my_design.def
# Should equal number of components
```

### 4.6 Cross-Check: LEF/Liberty/Verilog Consistency

```bash
# List all referenced cell types in Verilog
grep -o "fecim_[a-z0-9]*" output/my_design.v | sort | uniq

# These must match:
# 1. Cell names in DEF (second field after -)
# 2. Cell names in your custom cell library (LEF/Liberty files)
# 3. Module definitions in behavioral Verilog

# For example, if Verilog uses fecim_bit:
# ✓ DEF must have: - cell_0_0 fecim_bit + ...
# ✓ LEF must define: MACRO fecim_bit
# ✓ Liberty must define: cell(fecim_bit)
```

## Step 5: OpenLane Integration

### 5.1 Design Directory Structure

Create OpenLane design directory:

```bash
export OPENLANE_ROOT=~/OpenLane
export DESIGN_NAME=my_fecim
mkdir -p $OPENLANE_ROOT/designs/$DESIGN_NAME/{src,cells}
```

### 5.2 Copy Generated Files

```bash
# Copy design files from EDA CLI output
cp output/my_design.v $OPENLANE_ROOT/designs/$DESIGN_NAME/src/crossbar.v
cp output/my_design.def $OPENLANE_ROOT/designs/$DESIGN_NAME/src/crossbar.def

# Copy or create custom cell files
cp cells/fecim_bit.lef $OPENLANE_ROOT/designs/$DESIGN_NAME/cells/
cp cells/fecim_bit.lib $OPENLANE_ROOT/designs/$DESIGN_NAME/cells/
cp cells/fecim_bit.v $OPENLANE_ROOT/designs/$DESIGN_NAME/cells/
cp cells/fecim_bit.gds $OPENLANE_ROOT/designs/$DESIGN_NAME/cells/
```

### 5.3 OpenLane Configuration

Create `config.json` with critical settings for crossbar design:

```json
{
  "DESIGN_NAME": "my_fecim_8x8",
  "VERILOG_FILES": "dir::src/crossbar.v",
  "CLOCK_PERIOD": 10,
  "CLOCK_PORT": "CLK",

  "EXTRA_LEFS": "dir::cells/fecim_bit.lef",
  "EXTRA_GDS_FILES": "dir::cells/fecim_bit.gds",
  "EXTRA_LIBS": "dir::cells/fecim_bit.lib",
  "VERILOG_FILES_BLACKBOX": "dir::cells/fecim_bit.v",

  "SYNTH_ELABORATE_ONLY": 1,
  "SYNTH_FLATTEN_BEFORE_ABC": 1,

  "FP_SIZING": "absolute",
  "DIE_AREA": "0 0 100 100",

  "FP_DEF_TEMPLATE": "dir::src/crossbar.def",
  "PL_SKIP_INITIAL_PLACEMENT": 1,

  "DESIGN_IS_CORE": 0,
  "FP_PDN_ENABLE_RAILS": 0,
  "RUN_CTS": 0,

  "QUIT_ON_MAGIC_DRC": 0,
  "QUIT_ON_LVS_ERROR": 0
}
```

**Critical Settings Explained:**

| Setting | Value | Purpose |
|---------|-------|---------|
| `SYNTH_ELABORATE_ONLY` | 1 | Skip logic synthesis (netlist is already structural) |
| `FP_DEF_TEMPLATE` | src/crossbar.def | Use pre-placed cell positions |
| `PL_SKIP_INITIAL_PLACEMENT` | 1 | Skip placement tool (cells already placed in DEF) |
| `RUN_CTS` | 0 | Skip clock tree synthesis (no clock) |
| `DESIGN_IS_CORE` | 0 | Treat as macro, not top-level design |

### 5.4 Custom Cell Library

For development, you can use stub cells. For tape-out, design real cells in Magic VLSI.

**Minimum fecim_bit.lef (stub):**

```lef
MACRO fecim_bit
  CLASS CORE ;
  SIZE 0.46 BY 2.72 ;
  SYMMETRY X Y ;
  SITE unithd ;

  PIN WL
    DIRECTION INPUT ;
    PORT LAYER met1 ;
      RECT 0.0 0.0 0.1 2.72 ;
    END
  END WL

  PIN BL
    DIRECTION OUTPUT ;
    PORT LAYER met2 ;
      RECT 0.36 0.0 0.46 2.72 ;
    END
  END BL

  PIN VPWR
    DIRECTION INOUT ; USE POWER ;
    PORT LAYER met1 ;
      RECT 0.0 2.62 0.46 2.72 ;
    END
  END VPWR

  PIN VGND
    DIRECTION INOUT ; USE GROUND ;
    PORT LAYER met1 ;
      RECT 0.0 0.0 0.46 0.1 ;
    END
  END VGND
END fecim_bit
```

**Minimum fecim_bit.v (behavioral):**

```verilog
module fecim_bit (
    input  wire WL,
    output wire BL,
    inout  wire VPWR,
    inout  wire VGND
);
    // For simulation: BL tracks WL
    assign BL = WL;
endmodule
```

**Minimum fecim_bit.lib (timing):**

```liberty
library(fecim_bit) {
  cell(fecim_bit) {
    area : 1.2512 ;
    pin(WL) {
      direction : input ;
      capacitance : 0.001 ;
    }
    pin(BL) {
      direction : output ;
      function : "WL" ;
      timing() {
        related_pin : "WL" ;
        cell_rise(scalar) { values("0.1"); }
        cell_fall(scalar) { values("0.1"); }
      }
    }
    pin(VPWR) {
      direction : inout ;
      pg_type : primary_power ;
    }
    pin(VGND) {
      direction : inout ;
      pg_type : primary_ground ;
    }
  }
}
```

## Step 6: Run Synthesis and P&R

### 6.1 Run OpenLane

From OpenLane root directory:

```bash
cd ~/OpenLane

# Mount and run (Docker)
make mount
./flow.tcl -design my_fecim -tag v1 2>&1 | tee run.log
```

Or native execution:

```bash
./flow.tcl -design my_fecim -tag v1
```

### 6.2 Monitor Progress

```bash
# Watch the run in real-time
tail -f runs/my_fecim/v1/logs/*/latest.log

# Or check current step
cat runs/my_fecim/v1/.flow_status
```

### 6.3 Expected Output

```
[INFO]: OpenLane v1.0
[INFO]: Starting design my_fecim...

[STEP 1/13] Running Synthesis...
[INFO]: Elaborating design using Yosys...
[INFO]: Flattened design...

[STEP 2/13] Running Floorplan...
[INFO]: Using absolute die area: 0 0 100 100...

[STEP 3/13] Running Placement...
[INFO]: Skipping initial placement (PL_SKIP_INITIAL_PLACEMENT=1)...
[INFO]: Running legalizer...

[STEP 4/13] Skipping CTS (RUN_CTS=0)...

[STEP 5/13] Running Routing...
[INFO]: Running FastRoute...
[INFO]: Running TritonRoute...

[STEP 6/13] Running Signoff...
[INFO]: Running Magic GDS...
[INFO]: Running DRC...

[SUCCESS]: Flow completed!
[INFO]: Design area: 12.34 um^2
[INFO]: Final die area: 100.00 um^2
[INFO]: Peak memory: 1234 MB
```

### 6.4 Check Results

```bash
# View summary metrics
cat runs/my_fecim/v1/reports/metrics.csv | head -20

# Check DRC violations
cat runs/my_fecim/v1/reports/signoff/drc.rpt | grep -i violation

# View timing (not applicable for crossbar)
cat runs/my_fecim/v1/reports/signoff/timing_summary.txt
```

## Step 7: Generate GDSII

The final GDSII file is generated by OpenLane during the flow. Location:

```
~/OpenLane/designs/my_fecim/runs/v1/results/final/gds/my_fecim.gds
```

### 7.1 Verify GDSII

```bash
# Check file exists and size
ls -lh ~/OpenLane/designs/my_fecim/runs/v1/results/final/gds/

# Inspect with gdspy
python3 << 'EOF'
import gdspy
gds = gdspy.GdsLibrary(infile="my_fecim.gds")
print(f"Libraries: {gds.libraries}")
print(f"Cells: {list(gds.cells.keys())}")
for cell_name, cell in gds.cells.items():
    print(f"  {cell_name}: {len(cell.polygons)} polygons, {len(cell.labels)} labels")
EOF
```

### 7.2 View Layout in KLayout

```bash
# Open GDSII in KLayout
klayout ~/OpenLane/designs/my_fecim/runs/v1/results/final/gds/my_fecim.gds
```

Navigation tips:
- Zoom: Mouse wheel or `+`/`-`
- Pan: Middle mouse or spacebar + drag
- Measure: Click ruler icon, then two points
- Layer view: Right panel shows all layers

### 7.3 Extract Additional Formats

**DEF (final placement with routed nets):**

```bash
cp ~/OpenLane/designs/my_fecim/runs/v1/results/final/def/my_fecim.def .
```

**LEF (abstract for reuse):**

```bash
cp ~/OpenLane/designs/my_fecim/runs/v1/results/final/lef/my_fecim.lef .
```

**Spice netlist (post-layout):**

```bash
cp ~/OpenLane/designs/my_fecim/runs/v1/results/final/sdc/*.sdc .
```

## Example Workflows

### Example 1: Basic 8×8 Crossbar

See `<local-path>`

Quick start:

```bash
cd module6-eda/examples/01-basic-8x8
bash run.sh
```

**Outputs:** Fully validated 8×8 design with reference outputs

---

### Example 2: MNIST Network Layer

See `<local-path>`

Compile realistic neural network weights:

```bash
cd module6-eda/examples/02-mnist-layer
bash run.sh

# Simulate with ngspice
cd output
ngspice -b ../testbench.sp -o sim_results.log
```

**Outputs:** 32×32 MNIST layer with simulation validation

---

### Example 3: Full OpenLane Integration

See `<local-path>`

Complete workflow from weights to GDSII:

```bash
cd module6-eda/examples/03-openlane-integration

# Step 1: Compile weights
bash run_compile.sh

# Step 2: Run OpenLane
bash run_openlane.sh

# Step 3: View results
klayout output/gds/fecim_crossbar.gds
```

**Outputs:** GDSII-ready design with full OpenLane integration

---

## Troubleshooting

### Weights File Issues

**Problem:** "Error parsing weights JSON"

**Solution:**
```bash
# Validate JSON
python3 -m json.tool weights.json

# Check required fields
python3 << 'EOF'
import json
with open("weights.json") as f:
    data = json.load(f)
    assert "rows" in data and "cols" in data and "weights" in data
    assert len(data["weights"]) == data["rows"]
    assert all(len(row) == data["cols"] for row in data["weights"])
    print("✓ Valid weights file")
EOF
```

**Problem:** "weights exceed array dimensions"

**Solution:** Increase array size or trim weights:
```bash
# Use larger array
./eda-cli -input weights.json -rows 64 -cols 64

# Or subset weights in Python
import json
with open("weights.json") as f:
    data = json.load(f)
data["weights"] = [row[:32] for row in data["weights"][:32]]
data["rows"] = data["cols"] = 32
with open("weights_subset.json", "w") as f:
    json.dump(data, f)
```

---

### Design Generation Issues

**Problem:** "Error creating output directory"

**Solution:**
```bash
# Check permissions
ls -ld ./output
chmod 755 ./output

# Or use absolute path
./eda-cli -output /tmp/fecim_output -input weights.json
```

**Problem:** High quantization error (PSNR < 30 dB)

**Solution:**
```bash
# Increase conductance range
./eda-cli -input weights.json -gmin 0.1 -gmax 1000

# Or pre-quantize weights to [-1, 1]
# (30-level system assumes normalized weights)
```

---

### Verilog/SPICE Validation Failures

**Problem:** "iverilog: syntax error"

**Solution:**
```bash
# Check for non-standard Verilog
grep -E "real|logic|wire\[" output/my_design.v

# Use specific Verilog standard
iverilog -g2012 -o /dev/null output/my_design.v

# Or regenerate with explicit cell models
./eda-cli -verilog < verify cell templates >
```

**Problem:** "ngspice: unrecognized parameter"

**Solution:**
```bash
# Check conductance format
grep "^R_" output/my_design.sp | head -5

# Verify all values are positive numeric
awk -F' ' '{
    if ($3 !~ /^[0-9.eE+-]+$/) {
        print "Invalid: " $0
    }
}' output/my_design.sp

# Regenerate SPICE with validation
./eda-cli -spice -output ./output
```

---

### OpenLane Integration Issues

**Problem:** "EXTRA_LEFS not found"

**Solution:**
```bash
# Verify file exists
ls -la ~/OpenLane/designs/my_fecim/cells/fecim_bit.lef

# Check config.json path
cat ~/OpenLane/designs/my_fecim/config.json | grep EXTRA_LEFS

# Use absolute path in config
"EXTRA_LEFS": "/full/path/to/fecim_bit.lef"
```

**Problem:** "Unplaced cells after initial placement"

**Solution:**
```bash
# Verify DEF has FIXED keyword
grep "FIXED" ~/OpenLane/designs/my_fecim/src/crossbar.def | wc -l
# Should equal number of cells

# Check DEF syntax
grep "^  -" ~/OpenLane/designs/my_fecim/src/crossbar.def | head -5
# Format: - cell_name cell_type + FIXED ( X Y ) orientation ;

# Regenerate with validation
./eda-cli -def -output ./output
```

**Problem:** "DRC violations"

**Solution:**
```bash
# For development (acceptable):
echo "QUIT_ON_MAGIC_DRC=0" >> config.json
echo "QUIT_ON_LVS_ERROR=0" >> config.json

# For tape-out (required fix):
# 1. Design cell layout in Magic VLSI
# 2. Run Magic DRC (Tools > DRC)
# 3. Fix violations (check Magic documentation)
# 4. Export corrected GDS

# View violations
cat ~/OpenLane/designs/my_fecim/runs/v1/reports/signoff/drc.rpt
```

**Problem:** "Flow timeout"

**Solution:**
```bash
# Increase timeouts in config.json
"PLACEMENT_TIMEOUT": 3600,  # 1 hour
"ROUTING_TIMEOUT": 3600,

# Or disable relaxed settings
"QUIT_ON_MAGIC_DRC": 1,
"QUIT_ON_LVS_ERROR": 1
```

---

### Performance and Scaling

**Problem:** CLI is slow for large arrays

**Solution:**
```bash
# Profile the CLI
time ./eda-cli -rows 128 -cols 128 -output /tmp/large

# For 128×128: ~5-10 seconds expected
# For 256×256: ~30-60 seconds expected
# If slower, check disk I/O and memory
```

**Problem:** GDSII file is huge (> 100 MB)

**Solution:**
```bash
# Check for polygon bloat
python3 << 'EOF'
import gdspy
gds = gdspy.GdsLibrary(infile="my_fecim.gds")
for name, cell in gds.cells.items():
    print(f"{name}: {len(cell.polygons)} polygons, {len(cell.references)} refs")
EOF

# Reduce detail in cell layout (fewer polygons per cell)
# Or compress GDSII
gzip my_fecim.gds
```

---

### Architecture-Specific Issues

**Problem:** Using 1T1R but cells are passive

**Solution:**
```bash
# Regenerate with correct architecture
./eda-cli -arch 1t1r -input weights.json -output ./output

# Verify in Verilog
grep "instantiate" output/my_design.v | head -1
# Should show: fecim_1t1r, not fecim_bit
```

**Problem:** Peripheral circuits not including in design

**Solution:**
```bash
# Peripherals (DAC/ADC) are configured but separate
# Module 4 (circuits) provides peripheral designs
# Integrate manually into top-level design:

# 1. Generate crossbar with EDA CLI
./eda-cli -mode compute -input weights.json

# 2. Design peripherals in Module 4 GUI or separately
# 3. Wrap in top-level (top.v):
#    module top(V_in[bits], I_out[bits], clk, reset, ...);
#        crossbar (.WL(WL), .BL(BL), ...);
#        dac (.input(V_in), .output(V_to_WL), ...);
#        adc (.input(I_from_BL), .output(I_out), ...);
#    endmodule
```

---

## Next Steps

1. **Validate Design:** Run examples (01, 02) to verify workflow
2. **Integrate with OpenLane:** Complete Example 3
3. **Design Real Cells:** Use Magic VLSI for production layouts
4. **Characterize Timing:** Extract parasitic RC, generate accurate Liberty
5. **Scale Up:** Test 64×64, 128×128, 256×256 arrays
6. **Top-Level Integration:** Create SoC with crossbar + peripherals + control logic
7. **DRC/LVS Cleanup:** Fix all violations for tape-out

---

## Further Reading

- **Example Designs:** `module6-eda/examples/`
- **EDA CLI Reference:** `module6-eda/cmd/eda-cli/main.go`
- **Export Formats:** `module6-eda/pkg/export/doc.go`
- **OpenLane Guide:** https://github.com/The-OpenROAD-Project/OpenLane
- **SKY130 PDK:** https://github.com/google/skywater-pdk
- **Magic VLSI:** http://opencircuitdesign.com/magic/
