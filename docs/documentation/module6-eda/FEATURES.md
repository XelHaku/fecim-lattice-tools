# Module 6: EDA - Features

## What This Module Does

- Compiles arrays for storage, memory, or compute modes.
- Exports mappings to JSON, CSV, SPICE, Verilog, and DEF.
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

## Extension Points

- Add export formats or mapping constraints.
- Improve placement/tiling strategies.
- Integrate additional layout workflows.

## Known Limitations

- Liberty timing values are placeholders (need characterization).
- Exports are educational artifacts, not tape-out ready.
- CLI supports passive/1T1R; 2T1R is available via API only.
