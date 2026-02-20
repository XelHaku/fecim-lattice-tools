<!-- Category: Physics | Module: module6-eda | Reading time: ~5 min -->
# Module 6 Physics: Design Representation and Mapping

> Module 6 handles design representation and handoff -- from weight
> matrices to industry-standard export files. It is not a signoff-grade
> device simulator.

---

## Core Mapping Model

A weight matrix is partitioned into array-sized tiles and mapped to
physical coordinates for export.

```
TilesRows = ceil(R / ArrayRows)
TilesCols = ceil(C / ArrayCols)
TileIndex = (rowTile, colTile)
```

| Parameter | Meaning | Units |
|-----------|---------|-------|
| R | Weight matrix rows | count |
| C | Weight matrix columns | count |
| ArrayRows | Crossbar rows per tile | count |
| ArrayCols | Crossbar columns per tile | count |
| TilesRows | Number of row tiles | count |
| TilesCols | Number of column tiles | count |

Each weight value is quantized to a discrete conductance level and
assigned to a specific cell coordinate within its tile.

---

## Abstraction Levels

The design passes through six levels of abstraction from algorithm
to silicon. Each level has its own simplifications.

```
Level A: Algorithm
  - Numeric weight matrix + quantization
  - Simplification: ideal write/read, no transient dynamics

Level B: Array-to-layout mapping (Module 6 compiler)
  - Geometric placement + connectivity template
  - Simplification: no parasitic RC, no coupling/noise

Level C: RTL / netlist handoff (Verilog / DEF)
  - Structural connectivity
  - Simplification: analog behavior represented as macros

Level D: Circuit simulation (SPICE)
  - Device-level electrical simulation
  - Simplification: reduced compact models, limited corners

Level E: Physical implementation (OpenLane / OpenROAD)
  - Digital P&R with custom cells
  - Simplification: analog effects outside digital signoff

Level F: Signoff
  - DRC / LVS / STA checks
  - Simplification: covers design rules, not material reliability
```

Module 6 operates at Levels A through C. Levels D through F require
external tools (ngspice, OpenLane, foundry DRC).

---

## Consistency Rules

When moving between abstraction levels, simplifications must be
carried forward explicitly:

1. Do not claim silicon-equivalent behavior from Module 6 export alone.
2. Keep abstraction labels explicit: architectural, circuit-modeled,
   or signoff-verified.
3. When moving stages, carry forward known simplifications instead of
   silently dropping them.
4. If a downstream result depends on placeholder libraries, label the
   outcome as modeled.

---

## Export File Physics

### SPICE Netlist

Each cell is represented as a resistor with conductance derived from
the quantized weight:

```
R_cell = 1 / G_cell
```

For FeFET physics, the resistor should be replaced with a Verilog-A
compact model in downstream simulation.

### Verilog RTL

Structural module hierarchy: top-level array instantiates cell
modules, with wordline and bitline port connections matching the
physical topology.

### DEF Placement

Cells are placed at grid-aligned coordinates derived from the tile
mapping. Placement uses FIXED status for pre-placed crossbar cells.

```
cell coordinate = (tile_row * ArrayRows + local_row) * pitch
```

---

## Current Limits

- No built-in parasitic extraction in the export path.
- No automatic foundry-calibrated FeFET compact model generation.
- No automatic PVT (process/voltage/temperature) corner analysis.
- Liberty timing values are placeholders, not characterized.

---

## Where It Lives in Code

| Path | Purpose |
|------|---------|
| `module6-eda/pkg/compiler/compiler.go` | Array-to-layout compiler |
| `module6-eda/pkg/compiler/types.go` | IR definitions |
| `module6-eda/pkg/compiler/tiling.go` | Weight matrix partitioning |
| `module6-eda/pkg/export/spice.go` | SPICE netlist export |
| `module6-eda/pkg/export/verilog.go` | Verilog RTL export |
| `module6-eda/pkg/export/def.go` | DEF placement export |
| `module6-eda/pkg/export/lef.go` | LEF macro export |
| `module6-eda/pkg/validation/drc.go` | Design rule checks |

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
