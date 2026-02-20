// Package export provides multi-format output generation for FeCIM crossbar mappings.
//
// This package converts the compiled CrossbarMapping into various industry-standard
// formats for different downstream tools and workflows.
//
// # Supported Formats
//
// JSON (.json):
// Complete mapping data including configuration, all cell assignments, and
// compilation statistics. Suitable for version control, data analysis, and
// programmatic consumption.
//
// CSV (.csv):
// Tabular format with one row per cell containing row, col, weight, level,
// conductance, and resistance. Compatible with spreadsheet tools and pandas.
//
// SPICE (.sp):
// ngspice/HSPICE-compatible netlist for analog simulation. Includes resistor
// network representing the crossbar with conductance values mapped to resistance.
//
// Verilog (.v):
// Structural netlist instantiating fecim_bit cells. Compatible with Yosys
// synthesis (use SYNTH_ELABORATE_ONLY=1) and digital simulation with iverilog.
//
// DEF (.def):
// Design Exchange Format file with cell placements. Uses FIXED keyword for
// locked positions compatible with OpenLane's FP_DEF_TEMPLATE injection.
//
// # OpenLane Integration
//
// The generated Verilog and DEF files are designed for OpenLane integration:
//
//	// Generate files
//	mapping := compiler.Compile(weights, config)
//	export.ExportVerilog(mapping, "crossbar.v")
//	export.ExportDEF(mapping, "crossbar.def")
//
//	// OpenLane config.json:
//	// "VERILOG_FILES": "dir::crossbar.v",
//	// "FP_DEF_TEMPLATE": "dir::crossbar.def",
//	// "SYNTH_ELABORATE_ONLY": 1,
//	// "PL_SKIP_INITIAL_PLACEMENT": 1
//
// # Architecture Support
//
// Both passive crossbar and 1T1R architectures are supported. The architecture
// affects cell instantiation in Verilog and placement pitch in DEF:
//
//   - Passive: fecim_bit cells with WL, BL, VPWR, VGND pins
//   - 1T1R: fecim_1t1r cells with WL, BL, SL, VPWR, VGND pins
//
// # Usage Example
//
//	mapping, _ := compiler.Compile(weights, config)
//
//	// Export all formats
//	export.ExportJSON(mapping, "output/mapping.json")
//	export.ExportCSV(mapping, "output/cells.csv")
//	export.ExportSPICE(mapping, "output/crossbar.sp")
//	export.ExportVerilog(mapping, "output/crossbar.v")
//	export.ExportDEF(mapping, "output/crossbar.def")
package export
