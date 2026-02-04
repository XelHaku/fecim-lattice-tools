# Example 01: Basic 8x8 Crossbar

A minimal example demonstrating FeCIM weight compilation for an 8x8 crossbar array.

## Overview

This example compiles a simple 8x8 weight matrix with mixed positive and negative values, showcasing configurable quantization levels.

## Files

| File | Description |
|---|---|
| `weights.json` | 8x8 weight matrix with values from -0.9 to +0.9 |
| `run.sh` | Script to compile and export all formats |
| `expected_output/` | Reference output for validation |

## Running the Example

```bash
# From repository root
cd module6-eda

# Method 1: Use the provided script
./examples/01-basic-8x8/run.sh

# Method 2: Run CLI directly
go run ../cmd/fecim-lattice-tools eda cli \
  -input examples/01-basic-8x8/weights.json \
  -output examples/01-basic-8x8/output \
  -rows 8 -cols 8 -levels 30
```

## Expected Output (Summary)

You should see a compute-mode run with export confirmations similar to:

```
FeCIM Array Generator - Compute Mode
Configuration:
  Array Size:   8 x 8
  Levels:       <configured>
Exporting files to examples/01-basic-8x8/output/
  OK ..._design.json
  OK ..._cells.csv
  OK ... .sp
  OK ... .v
  OK ... .def
```

## Output Files

Default file names (if `-name` is not provided):

- `fecim_array_design.json`
- `fecim_array_cells.csv`
- `fecim_array.sp`
- `fecim_array.v`
- `fecim_array.def`

### `fecim_array_design.json`

Note: The snippet below shows the **structure** only. Values are determined by your inputs and current defaults.

```json
{
  "config": {
    "array_rows": 8,
    "array_cols": 8,
    "levels": <configured>,
    "g_min": <configured>,
    "g_max": <configured>
  },
  "cells": [
    {"row": 0, "col": 0, "level": <int>, "conductance": <uS>, "program_v": <V>, "initial_weight": <float>},
    ...
  ],
  "stats": { ... }
}
```

### `fecim_array_cells.csv`

```csv
row,col,weight,level,conductance_uS,resistance_ohm,program_V
0,0,<weight>,<level>,<conductance_uS>,<resistance_ohm>,<program_V>
0,1,<weight>,<level>,<conductance_uS>,<resistance_ohm>,<program_V>
...
```

### `fecim_array.v`

```verilog
module fecim_crossbar (
    input wire [7:0] WL,
    inout wire [7:0] BL,
    inout wire VPWR,
    inout wire VGND
);
    fecim_bit cell_0_0 (.WL(WL[0]), .BL(BL[0]), .VPWR(VPWR), .VGND(VGND));
    // ... 64 cells total
endmodule
```

### `fecim_array.def`

```def
COMPONENTS 64 ;
  - cell_0_0 fecim_bit + FIXED ( <x> <y> ) N ;
  - cell_0_1 fecim_bit + FIXED ( <x> <y> ) N ;
  ...
END COMPONENTS
```

## Validation

```bash
# Verify SPICE syntax (optional)
ngspice -b -c 'source output/fecim_array.sp; listing'

# Verify Verilog syntax (optional)
iverilog -o /dev/null output/fecim_array.v
echo "Verilog syntax OK"
```

## Weight Matrix

```
     Col 0   Col 1   Col 2   Col 3   Col 4   Col 5   Col 6   Col 7
Row 0:  0.1   -0.2    0.3   -0.4    0.5   -0.6    0.7   -0.8
Row 1: -0.1    0.2   -0.3    0.4   -0.5    0.6   -0.7    0.8
Row 2:  0.15  -0.25   0.35  -0.45   0.55  -0.65   0.75  -0.85
Row 3: -0.15   0.25  -0.35   0.45  -0.55   0.65  -0.75   0.85
Row 4:  0.05  -0.15   0.25  -0.35   0.45  -0.55   0.65  -0.75
Row 5: -0.05   0.15  -0.25   0.35  -0.45   0.55  -0.65   0.75
Row 6:  0.2   -0.3    0.4   -0.5    0.6   -0.7    0.8   -0.9
Row 7: -0.2    0.3   -0.4    0.5   -0.6    0.7   -0.8    0.9
```

## Next Steps

1. Try larger arrays (16x16, 32x32).
2. Modify `weights.json` with your own values.
3. Run ngspice on the SPICE netlist.
4. See `03-openlane-integration` for OpenLane steps.
