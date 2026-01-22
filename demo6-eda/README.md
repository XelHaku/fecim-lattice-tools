# Demo 6: FeCIM Design Suite

EDA bridge tool for ferroelectric compute-in-memory (FeCIM) crossbar arrays.

## Quick Start

```bash
# Build
go build ./...

# Test
go test ./... -v

# Run GUI
go run ./cmd/eda-gui
```

## Features

### Tab 1: Compiler
- Load neural network weights from JSON
- Configure array size, quantization levels, conductance range
- Compile weights to physical cell assignments
- View compilation statistics (utilization, PSNR, etc.)

### Tab 2: Layout
- Visual crossbar grid (color-coded by conductance)
- Blue = low conductance, Red = high conductance
- Click cells to view details (weight, level, conductance, voltage)

### Tab 5: Export
- JSON: Full mapping with stats
- CSV: Cell assignments table
- SPICE: ngspice-compatible netlist

## Sample Data

Load `data/sample_weights_8x8.json` to test.

## Key Formulas

```
Quantize: level = round((weight + maxWeight) / (2 * maxWeight) * (Levels-1))
Conductance: G = GMin + (level / (Levels-1)) * (GMax - GMin)
Resistance: R = 1e6 / G  (ohms from microsiemens)
```

## CLI Tool

For headless/automated compilation:

```bash
go run ./cmd/eda-cli -input data/sample_weights_8x8.json -rows 8 -cols 8

# Options:
#   -input    Input weights JSON (required)
#   -output   Output directory (default: .)
#   -rows     Array rows (default: 128)
#   -cols     Array cols (default: 128)
#   -levels   Quantization levels (default: 30)
#   -vdd      VDD for SPICE (default: 1.8)
#   -json     Export JSON (default: true)
#   -csv      Export CSV (default: true)
#   -spice    Export SPICE (default: true)
```

## Project Structure

```
demo6-eda/
├── cmd/
│   ├── eda-gui/main.go    # GUI application
│   └── eda-cli/main.go    # CLI tool
├── pkg/
│   ├── compiler/          # Weight-to-cell compilation
│   ├── export/            # JSON, CSV, SPICE export
│   └── gui/               # Fyne GUI tabs
├── data/                  # Sample weight files
├── Makefile
├── go.mod
└── go.sum
```
