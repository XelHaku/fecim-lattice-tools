# Module 6: EDA - Features

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## What This Module Does

- Compiles arrays for storage, memory, or compute modes.
- Supports module export formats: JSON, CSV, SPICE, Verilog, DEF, LEF, Liberty, and SVG.
- Provides a GUI for configuration, validation, and learning visuals.

## Primary Components

- `module6-eda/pkg/compiler/compiler.go`
- `module6-eda/pkg/export/*.go`
- `module6-eda/pkg/validation/*.go`
- `module6-eda/pkg/gui/app.go`

## Key Workflows

- Configure array settings and generate a design.
- Export mapping to files for downstream tools.
- Validate outputs (Yosys, DEF checks, cross-file consistency).

## GUI/CLI Parity (Validated)

- Shared: Verilog + DEF generation, validation pipeline entry points.
- CLI defaults: compute mode, 128x128 array, passive architecture.
- GUI defaults: storage mode, 4x4 array, passive architecture.
- CLI export flags currently cover JSON/CSV/SPICE/Verilog/DEF.
- GUI builder flow also generates LEF/Liberty; SVG is available through `pkg/export` APIs.

## Extension Points

- Add export formats or mapping constraints.
- Improve placement/tiling strategies.
- Integrate additional layout workflows.

## Known Limitations

- Liberty timing values are placeholders (need characterization).
- Exports are educational artifacts, not tape-out ready.
- CLI supports passive/1T1R; 2T1R is available via API only.
- PDK support refers to preset dimensions/labels; no foundry PDK files are bundled (DOI: (add)).
