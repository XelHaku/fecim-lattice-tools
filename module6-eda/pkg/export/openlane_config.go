// pkg/export/openlane_config.go
// OpenLane v2.0 configuration generator for FeCIM crossbar arrays
//
// References:
// [1] OpenLane v2.0 Documentation:
//     https://openlane.readthedocs.io/en/latest/reference/configuration.html
//     Variables validated: 2026-01-24
// [2] FP_DEF_TEMPLATE: Floorplan template for pre-placed macros
// [3] SYNTH_ELABORATE_ONLY: Structural netlist elaboration without logic mapping
package export

import (
	"encoding/json"
	"fmt"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
)

// GenerateOpenLaneConfig generates an OpenLane v2.0 config.json for FeCIM crossbar
// This configures OpenLane to use pre-placed cells and custom LEF/Liberty files
// Reference: OpenLane v2.0 Configuration [Ref 1]
func GenerateOpenLaneConfig(cfg config.ArrayConfig) string {
	designName := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
	
	// Convert to microns for DIE_AREA
	dieWidth := float64(cfg.Cols)*cfg.CellWidth + 2.0   // Add 2μm margin
	dieHeight := float64(cfg.Rows)*cfg.CellHeight + 2.0 // Add 2μm margin
	
	config := map[string]interface{}{
		// Design identification
		"DESIGN_NAME": designName,
		"DESIGN_IS_CORE": 0, // This is a macro, not a full chip core
		
		// Verilog files
		"VERILOG_FILES": fmt.Sprintf("dir::output/%s.v", designName),
		"VERILOG_FILES_BLACKBOX": "dir::cells/fecim_bitcell/fecim_bitcell.v",
		
		// Clock configuration (crossbar has no clock)
		"RUN_CTS": 0, // Skip Clock Tree Synthesis [Ref 1]
		
		// Floorplanning with pre-placed cells
		"FP_SIZING": "absolute",
		"DIE_AREA": fmt.Sprintf("0 0 %.3f %.3f", dieWidth, dieHeight),
		"FP_DEF_TEMPLATE": fmt.Sprintf("dir::output/%s.def", designName), // Pre-placed DEF template [Ref 2]
		
		// Custom cell library
		"EXTRA_LEFS": "dir::cells/fecim_bitcell/fecim_bitcell.lef",
		"EXTRA_LIBS": "dir::cells/fecim_bitcell/fecim_bitcell.lib",
		"EXTRA_GDS_FILES": "dir::cells/fecim_bitcell/fecim_bitcell.gds",
		
		// Placement strategy for pre-placed macros
		"PL_SKIP_INITIAL_PLACEMENT": 1,  // Skip automatic placement [Ref 1]
		"PL_TARGET_DENSITY": 0.6,
		
		// Synthesis: elaborate structural netlist only
		"SYNTH_ELABORATE_ONLY": 1, // No logic mapping for structural Verilog [Ref 3]
	}
	
	data, _ := json.MarshalIndent(config, "", "  ")
	return string(data)
}
