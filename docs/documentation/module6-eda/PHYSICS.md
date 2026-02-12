# Module 6: EDA - Physics

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Tiling/mapping/export behavior documented here aligns with implemented compiler/export paths.
- **Modeled:** Electrical behavior represented through simplified array abstractions unless backed by measured device data.
- **Aspirational:** Silicon-accurate FeFET behavior, full PVT closure, and foundry-calibrated reliability are not yet demonstrated.

## Scope

Module 6 primarily handles **design representation and handoff** (netlist + placement artifacts). It is not a signoff-grade device simulator by itself.

## Core Mapping Model (Implemented)

A weight matrix is partitioned into array-sized tiles and mapped to physical coordinates for export.

```text
TilesRows = ceil(R / ArrayRows)
TilesCols = ceil(C / ArrayCols)
TileIndex = (rowTile, colTile)
```

### Parameters and Units

| Symbol | Meaning | Units |
|---|---|---|
| R | Weight matrix rows | count |
| C | Weight matrix columns | count |
| ArrayRows | Crossbar rows per tile | count |
| ArrayCols | Crossbar columns per tile | count |
| TilesRows | Number of row tiles | count |
| TilesCols | Number of column tiles | count |

## Physics Simplification Audit by Flow Stage

The same physics assumptions must be tracked consistently as we move from architecture intent to chip flow.

| Flow Stage | Dominant Model | Simplifications Kept | What must be upgraded next |
|---|---|---|---|
| **A. Algorithm / quantized mapping (Module 6 input stage)** | Numeric matrix + quantization | Discrete levels, ideal write/read transfer, no transient polarization dynamics | Add calibration hooks for device-level nonlinearity/drift in compute studies |
| **B. Array-to-layout mapping (Module 6 compiler/export)** | Geometric placement + connectivity template | No parasitic RC extraction, no coupling/noise, no variation across cells unless externally injected | Couple with extracted/intermediate RC estimates or bounded parasitic assumptions |
| **C. Functional RTL/netlist handoff (Verilog/DEF)** | Structural connectivity | Logic-level abstraction of analog behavior; FeFET state represented as abstracted behavior/macros | Replace stubs with characterized macro models (`.lib` + behavioral consistency checks) |
| **D. Circuit simulation (SPICE/compact model stage)** | Device/cell electrical simulation | Often reduced compact models, limited temperature/process corners, constrained endurance/retention windows | Calibrate to measured data and sweep PVT + endurance/retention envelopes |
| **E. OpenLane/OpenROAD physical flow** | Digital physical implementation with custom cells/macros | Timing/power closure depends on Liberty quality; analog non-idealities mostly outside digital signoff | Use characterized cell views and annotate what analog effects remain out-of-band |
| **F. Signoff interpretation / pre-tapeout claims** | DRC/LVS/STA + reported checks | Signoff covers design-rule/connectivity/timing constraints, not full FeFET material reliability | Explicitly separate "EDA signoff clean" from "device reliability validated" |

## Consistency Rules (Required for Reporting)

1. **Do not claim silicon-equivalent behavior from Module 6 export alone.**
2. **Keep abstraction labels explicit** in reports: architectural, circuit-modeled, or signoff-verified.
3. **When moving stages, carry forward known simplifications** (e.g., variation ignored, limited PVT) instead of silently dropping them.
4. **If a downstream result depends on placeholder libraries/models, label outcome as modeled.**

## Current Limits

- No built-in parasitic extraction in Module 6 export path.
- No automatic foundry-calibrated FeFET compact model generation.
- No automatic closure of endurance/retention under full PVT in this module.

## Where It Lives in Code

- `module6-eda/pkg/compiler/compiler.go`
- `module6-eda/pkg/compiler/types.go`
- `module6-eda/pkg/export/spice.go`
- `module6-eda/pkg/export/verilog.go`
- `module6-eda/pkg/export/def.go`

## References

- `docs/eda/README.md`
- `docs/eda/guides/integration.md`
- `docs/eda/guides/fecim-to-wafer.md`
- `docs/development/PHYSICS_ACCEPTANCE_CRITERIA.md`
