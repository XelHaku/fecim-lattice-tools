# M4-INV-07 Results — SPICE Netlist Export from M4 State

Source tests:
- `module4-circuits/pkg/arraysim/m4_investigations_test.go::TestM4INV07_SPICEExportFromArrayState`
- Existing coverage: `module4-circuits/pkg/arraysim/spice_export_test.go`

Method: construct a representative 4×4 M4 state (`WL`, `BL`, conductance matrix) and export behavioral ngspice deck via `ExportCrossbarSPICE`.

## Results

- Netlist generated successfully: `3196 bytes`
- Cell elements emitted: `16 RCELL_*` instances (matches 4×4)
- Required ngspice/control tokens present: `.control`, `op`, `.end`
- Peripheral subcircuits present (e.g., `XADC_0`)

Conclusion: module supports exporting current M4 electrical state as an ngspice-ready behavioral deck for external validation.