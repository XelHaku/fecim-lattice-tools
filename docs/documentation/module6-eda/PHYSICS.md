# Module 6: EDA - Physics

## Prerequisites

- Matrix shapes and dimensions
- Basic constraints and tiling
- Understanding of quantization levels

## Core Model

- A weight matrix is partitioned into crossbar-sized tiles.
- Each tile is assigned to a physical array location.
- Exports preserve mapping for downstream tools.

## Key Equations (Simplified)

```
TilesRows = ceil(R / ArrayRows)
TilesCols = ceil(C / ArrayCols)
TileIndex = (rowTile, colTile)
```

## Parameters And Units

| Symbol | Meaning | Units |
|---|---|---|
| R | Weight matrix rows | count |
| C | Weight matrix cols | count |
| ArrayRows | Crossbar rows | count |
| ArrayCols | Crossbar cols | count |
| TilesRows | Number of row tiles (ceil) | count |
| TilesCols | Number of column tiles (ceil) | count |
| rowTile, colTile | Tile indices | count |

## Assumptions And Limits

- No detailed routing or placement optimization.
- Focus is on mapping correctness and export format.
- Quantization levels are assumed to match array configuration.
- Counts are integer quantities; ceilings reflect tiling that may pad edges.

## Where It Lives In Code

- `module6-eda/pkg/compiler/compiler.go`
- `module6-eda/pkg/compiler/types.go`
- `module6-eda/pkg/export/spice.go`

## Sources

- `docs/eda/README.md`
- `docs/development/scriptReference.md#module-6-eda-tools-module6-eda`
