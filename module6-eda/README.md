# Module 6: EDA Tools

Educational Electronic Design Automation (EDA) toolchain for FeCIM array design. Generates **illustrative** artifacts for learning and early exploration. **Simulation-only; not tape-out ready.**

## Trust Boundary (Read First)

**Module 6 exports are educational/research prototypes and are NOT suitable for signoff, tapeout, or production characterization.**

Missing for production:
- Measured NLDM timing tables
- Multi-corner silicon-backed characterization
- DRC/LVS closure versus real foundry PDK decks
- CCS timing/noise models for signoff STA

What is provided:
- Analytical models with literature-calibrated parameters
- Structurally valid export files (syntax/format correctness)
- Research-grade approximations for exploration

If you need production quality, use Module 6 outputs only as scaffolding and re-run full foundry signoff characterization/verification flows.

See: `docs/validation/eda-trust-boundary.md`

---

## Overview

Module 6 provides an RTL-to-layout **exploration flow** for FeCIM arrays. It supports three operation modes (Storage, Memory, Compute) and exports multiple file formats suitable for OpenLane-style flows. It is designed for education and planning, not fabrication signoff.

---

## Features

- **Three Operation Modes** - Storage, Memory, Compute
- **CLI + GUI** - Scriptable CLI and an interactive GUI
- **Export Coverage (module-wide)** - JSON, CSV, SPICE, Verilog, DEF, LEF, Liberty, SVG
- **CLI Exports** - JSON, CSV, SPICE, Verilog, DEF (flags)
- **GUI/API Exports** - LEF, Liberty, SVG (via `pkg/export`; GUI uses LEF/Liberty directly)
- **Architecture Support** - Passive/1T1R in CLI; 2T1R available via GUI/API (`With2T1R` path)
- **PDK Support** - SKY130, GF180MCU, IHP_SG13G2
- **OpenLane Helpers** - DEF + OpenLane config generation and validation

---

## Quick Start

### Basic Usage (CLI)

From the repo root:

```bash
go run ./cmd/fecim-lattice-tools eda cli -mode compute -rows 64 -cols 64 -output ./output
```

With optional weights (compute mode only):

```bash
go run ./cmd/fecim-lattice-tools eda cli -mode compute -rows 64 -cols 64 \
  -input weights.json -output ./output -name my_design
```

### GUI Application

```bash
go run ./cmd/fecim-lattice-tools eda gui
```

The current GUI exposes two views:
- **Builder & Validation** (array config + validation)
- **Learn** (educational visuals)

---

## CLI Options

```
Usage: fecim-lattice-tools eda cli [flags]

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
  -gmin float           Min conductance uS (default 10.0)
  -gmax float           Max conductance uS (default 100.0)

Output Configuration:
  -name string          Design name for output files (default "fecim_array")
  -output string        Output directory (default "data")

Export Formats:
  -json bool            Export JSON mapping (default true)
  -csv bool             Export CSV cell assignments (default true)
  -spice bool           Export SPICE netlist (default true)
  -verilog bool         Export Verilog netlist (default true)
  -def bool             Export DEF placement (default true)
```

**CLI output file names:**
- `*_design.json`
- `*_cells.csv`
- `*.sp`, `*.v`, `*.def`

## GUI/CLI Parity Notes

Current defaults are intentionally different:

- **CLI defaults:** `mode=compute`, `rows=128`, `cols=128`, `arch=passive`
- **GUI defaults:** `mode=storage`, `rows=4`, `cols=4`, `arch=passive`

Current export parity:

- **Both GUI and CLI:** Verilog, DEF (core flow artifacts)
- **CLI only (today):** JSON, CSV, SPICE flag-driven batch exports
- **GUI/API only (today):** LEF + Liberty generation in Builder flow; SVG available via `pkg/export`

This split is documented to avoid confusion when comparing headless automation (CLI) versus interactive learning flow (GUI).

---

## Package Structure

### pkg/compiler

- `NewArrayConfig(mode, rows, cols)`
- `GenerateDesign(config)`
- `GenerateBlank(config)`
- `With1T1R()`, `With2T1R()`

### pkg/export

- `ExportJSON`, `ExportCSV`, `ExportSPICE`, `ExportVerilog`, `ExportDEF`
- `GenerateLEF`, `GenerateLiberty`, `GenerateLayoutSVG`

### pkg/openlane

OpenLane config generation and runner helpers.

### pkg/validation

- Yosys checks
- DEF validation
- Cross-file consistency checks

### pkg/gui

Fyne GUI for Builder & Validation + Learn views.

---

## Limitations

- Liberty timing values are placeholders (need SPICE characterization).
- Exports are educational artifacts; not tape-out ready.
- No validated FeFET compact models included.
