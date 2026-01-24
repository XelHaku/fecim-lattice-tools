// pkg/export/generators_test.go
package export

import (
	"strings"
	"testing"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
)

func TestGenerateLEF(t *testing.T) {
	cfg := config.DefaultCellConfig()
	lef := GenerateLEF(cfg)

	// Test LEF version
	if !strings.Contains(lef, "VERSION 5.8") {
		t.Error("LEF should contain VERSION 5.8")
	}

	// Test MACRO declaration
	if !strings.Contains(lef, "MACRO fecim_bitcell") {
		t.Error("LEF should contain MACRO declaration with cell name")
	}

	// Test cell size
	if !strings.Contains(lef, "SIZE 0.460 BY 2.720") {
		t.Error("LEF should contain correct SIZE declaration")
	}

	// Test required pins
	requiredPins := []string{"PIN WL", "PIN BL", "PIN VPWR", "PIN VGND"}
	for _, pin := range requiredPins {
		if !strings.Contains(lef, pin) {
			t.Errorf("LEF should contain %s declaration", pin)
		}
	}

	// Test pin directions
	if !strings.Contains(lef, "DIRECTION INPUT") {
		t.Error("LEF should contain INPUT pin direction")
	}
	if !strings.Contains(lef, "DIRECTION OUTPUT") {
		t.Error("LEF should contain OUTPUT pin direction")
	}

	// Test power/ground usage
	if !strings.Contains(lef, "USE POWER") {
		t.Error("LEF should contain power pin usage")
	}
	if !strings.Contains(lef, "USE GROUND") {
		t.Error("LEF should contain ground pin usage")
	}
}

func TestGenerateLiberty(t *testing.T) {
	cfg := config.DefaultCellConfig()
	lib := GenerateLiberty(cfg)

	// Test library declaration
	if !strings.Contains(lib, "library(fecim_cells)") {
		t.Error("Liberty should contain library declaration")
	}

	// Test units
	requiredUnits := []string{"time_unit", "voltage_unit", "current_unit", "capacitive_load_unit"}
	for _, unit := range requiredUnits {
		if !strings.Contains(lib, unit) {
			t.Errorf("Liberty should contain %s declaration", unit)
		}
	}

	// Test cell declaration
	if !strings.Contains(lib, "cell(fecim_bitcell)") {
		t.Error("Liberty should contain cell declaration")
	}

	// Test pins
	if !strings.Contains(lib, "pin(WL)") || !strings.Contains(lib, "pin(BL)") {
		t.Error("Liberty should contain pin declarations")
	}

	// Test timing arc
	if !strings.Contains(lib, "timing()") {
		t.Error("Liberty should contain timing arc")
	}
	if !strings.Contains(lib, "cell_rise") || !strings.Contains(lib, "cell_fall") {
		t.Error("Liberty should contain cell_rise and cell_fall timing")
	}
}

func TestGenerateCellVerilog(t *testing.T) {
	cfg := config.DefaultCellConfig()
	verilog := GenerateCellVerilog(cfg)

	// Test module declaration
	if !strings.Contains(verilog, "module fecim_bitcell") {
		t.Error("Verilog should contain module declaration")
	}

	// Test ports
	requiredPorts := []string{"WL", "BL", "VPWR", "VGND"}
	for _, port := range requiredPorts {
		if !strings.Contains(verilog, port) {
			t.Errorf("Verilog should contain port %s", port)
		}
	}

	// Test input/output declarations
	if !strings.Contains(verilog, "input") {
		t.Error("Verilog should contain input declaration")
	}
	if !strings.Contains(verilog, "output") {
		t.Error("Verilog should contain output declaration")
	}

	// Test endmodule
	if !strings.Contains(verilog, "endmodule") {
		t.Error("Verilog should contain endmodule")
	}
}

func TestGenerateArrayVerilog(t *testing.T) {
	testCases := []struct {
		rows int
		cols int
		name string
	}{
		{4, 4, "4x4"},
		{8, 8, "8x8"},
		{16, 16, "16x16"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.ArrayConfig{
				Rows:         tc.rows,
				Cols:         tc.cols,
				Mode:         "storage",
				Architecture: "passive",
				Technology:   "sky130",
				CellWidth:    0.46,
				CellHeight:   2.72,
			}

			verilog := GenerateArrayVerilog(cfg)

			// Test module name
			expectedModule := "fecim_crossbar_" + tc.name
			if !strings.Contains(verilog, "module "+expectedModule) {
				t.Errorf("Verilog should contain module %s", expectedModule)
			}

			// Test that WL and BL ports exist
			if !strings.Contains(verilog, "WL") || !strings.Contains(verilog, "BL") {
				t.Error("Array Verilog should contain WL and BL ports")
			}

			// Test instance count (should have rows*cols instances)
			expectedInstances := tc.rows * tc.cols
			instanceCount := strings.Count(verilog, "fecim_bitcell cell_")
			if instanceCount != expectedInstances {
				t.Errorf("Expected %d cell instances, got %d", expectedInstances, instanceCount)
			}
		})
	}
}

func TestGenerateOpenLaneConfig(t *testing.T) {
	cfg := config.ArrayConfig{
		Rows:         16,
		Cols:         16,
		Mode:         "compute",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	configJSON := GenerateOpenLaneConfig(cfg)

	// Test critical variables
	criticalVars := []string{
		"DESIGN_NAME",
		"FP_DEF_TEMPLATE",  // CRITICAL: must be this, not PLACEMENT_CURRENT_DEF
		"RUN_CTS",
		"DESIGN_IS_CORE",
		"SYNTH_ELABORATE_ONLY",
		"EXTRA_LEFS",
		"EXTRA_LIBS",
	}

	for _, varName := range criticalVars {
		if !strings.Contains(configJSON, varName) {
			t.Errorf("OpenLane config should contain %s", varName)
		}
	}

	// CRITICAL TEST: Ensure we're NOT using the wrong variable
	if strings.Contains(configJSON, "PLACEMENT_CURRENT_DEF") {
		t.Error("CRITICAL: OpenLane config must NOT contain PLACEMENT_CURRENT_DEF (use FP_DEF_TEMPLATE)")
	}

	// Test design name
	if !strings.Contains(configJSON, "fecim_crossbar_16x16") {
		t.Error("OpenLane config should contain correct design name")
	}

	// Test JSON format
	if !strings.Contains(configJSON, "{") || !strings.Contains(configJSON, "}") {
		t.Error("OpenLane config should be valid JSON")
	}
}

func TestLEFWithCustomCell(t *testing.T) {
	cfg := config.CellConfig{
		Name:       "custom_cell",
		Width:      1.0,
		Height:     3.0,
		CellType:   "1t1r",
		Technology: "sky130",
	}

	lef := GenerateLEF(cfg)

	if !strings.Contains(lef, "MACRO custom_cell") {
		t.Error("LEF should use custom cell name")
	}
	if !strings.Contains(lef, "SIZE 1.000 BY 3.000") {
		t.Error("LEF should use custom dimensions")
	}
}

func TestVerilogArrayNaming(t *testing.T) {
	cfg := config.ArrayConfig{
		Rows:         32,
		Cols:         32,
		Mode:         "compute",
		Architecture: "1t1r",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	verilog := GenerateArrayVerilog(cfg)

	// Test that mode and architecture are documented in comments
	if !strings.Contains(verilog, "Mode: compute") {
		t.Error("Verilog should document mode in header")
	}
	if !strings.Contains(verilog, "Architecture: 1t1r") {
		t.Error("Verilog should document architecture in header")
	}
}
