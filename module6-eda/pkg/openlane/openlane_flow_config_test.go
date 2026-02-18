// pkg/openlane/openlane_flow_config_test.go
// M6-OL-02: OpenLane Flow Configuration Validation
//
// Tests:
// - Generate config.json (OpenLane v2.0 format)
// - Verify all required variables set: DESIGN_NAME, VERILOG_FILES, etc.
// - Check that paths reference correct files

package openlane

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

// TestFlowConfigRequiredVariables_M6_OL_02 tests all required OpenLane variables are set
func TestFlowConfigRequiredVariables_M6_OL_02(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	cfg.Rows = 4
	cfg.Cols = 4

	configContent := export.GenerateOpenLaneConfig(cfg)

	// Parse JSON
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(configContent), &configMap); err != nil {
		t.Fatalf("Failed to parse config as JSON: %v", err)
	}

	// Required OpenLane v2.0 variables
	// Reference: https://openlane.readthedocs.io/en/latest/reference/configuration.html
	requiredVars := []string{
		"DESIGN_NAME",          // Design identification
		"VERILOG_FILES",        // Source Verilog files
		"FP_SIZING",            // Floorplan sizing mode
		"DIE_AREA",             // Die dimensions
		"FP_DEF_TEMPLATE",      // Pre-placed DEF template
		"EXTRA_LEFS",           // Custom cell LEF
		"EXTRA_LIBS",           // Custom cell Liberty
		"RUN_CTS",              // Clock tree synthesis flag
		"SYNTH_ELABORATE_ONLY", // Synthesis mode
	}

	for _, varName := range requiredVars {
		if _, exists := configMap[varName]; !exists {
			t.Errorf("Required variable missing: %s", varName)
		} else {
			t.Logf("M6-OL-02: %s = %v — PASS", varName, configMap[varName])
		}
	}
}

// TestFlowConfigDesignName_M6_OL_02 tests DESIGN_NAME format
func TestFlowConfigDesignName_M6_OL_02(t *testing.T) {
	testCases := []struct {
		rows     int
		cols     int
		expected string
	}{
		{4, 4, "fecim_crossbar_4x4"},
		{8, 8, "fecim_crossbar_8x8"},
		{16, 16, "fecim_crossbar_16x16"},
		{32, 32, "fecim_crossbar_32x32"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%dx%d", tc.rows, tc.cols), func(t *testing.T) {
			cfg := config.DefaultArrayConfig()
			cfg.Rows = tc.rows
			cfg.Cols = tc.cols

			configContent := export.GenerateOpenLaneConfig(cfg)

			var configMap map[string]interface{}
			if err := json.Unmarshal([]byte(configContent), &configMap); err != nil {
				t.Fatalf("JSON parse error: %v", err)
			}

			designName, ok := configMap["DESIGN_NAME"].(string)
			if !ok {
				t.Fatalf("DESIGN_NAME not a string: %v", configMap["DESIGN_NAME"])
			}

			if designName != tc.expected {
				t.Errorf("DESIGN_NAME = %q, want %q", designName, tc.expected)
			}

			t.Logf("M6-OL-02: DESIGN_NAME = %s — PASS", designName)
		})
	}
}

// TestFlowConfigVerilogFiles_M6_OL_02 tests VERILOG_FILES path
func TestFlowConfigVerilogFiles_M6_OL_02(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	cfg.Rows = 4
	cfg.Cols = 4

	configContent := export.GenerateOpenLaneConfig(cfg)

	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(configContent), &configMap); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}

	verilogFiles, ok := configMap["VERILOG_FILES"].(string)
	if !ok {
		t.Fatalf("VERILOG_FILES not a string: %v", configMap["VERILOG_FILES"])
	}

	// Check path format (OpenLane v2.0 uses dir:: prefix)
	if !strings.Contains(verilogFiles, "dir::") {
		t.Errorf("VERILOG_FILES missing 'dir::' prefix: %s", verilogFiles)
	}

	// Check .v extension
	if !strings.HasSuffix(verilogFiles, ".v") {
		t.Errorf("VERILOG_FILES missing .v extension: %s", verilogFiles)
	}

	// Check contains design name
	if !strings.Contains(verilogFiles, "fecim_crossbar_4x4") {
		t.Errorf("VERILOG_FILES missing design name: %s", verilogFiles)
	}

	t.Logf("M6-OL-02: VERILOG_FILES = %s — PASS", verilogFiles)
}

// TestFlowConfigDieArea_M6_OL_02 tests DIE_AREA calculation
func TestFlowConfigDieArea_M6_OL_02(t *testing.T) {
	testCases := []struct {
		rows       int
		cols       int
		cellWidth  float64
		cellHeight float64
	}{
		{4, 4, 0.46, 2.72},
		{8, 8, 0.46, 2.72},
		{16, 16, 0.46, 2.72},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%dx%d", tc.rows, tc.cols), func(t *testing.T) {
			cfg := config.DefaultArrayConfig()
			cfg.Rows = tc.rows
			cfg.Cols = tc.cols
			cfg.CellWidth = tc.cellWidth
			cfg.CellHeight = tc.cellHeight

			configContent := export.GenerateOpenLaneConfig(cfg)

			var configMap map[string]interface{}
			if err := json.Unmarshal([]byte(configContent), &configMap); err != nil {
				t.Fatalf("JSON parse error: %v", err)
			}

			dieArea, ok := configMap["DIE_AREA"].(string)
			if !ok {
				t.Fatalf("DIE_AREA not a string: %v", configMap["DIE_AREA"])
			}

			// Parse DIE_AREA format: "0 0 width height"
			var x1, y1, x2, y2 float64
			if _, err := fmt.Sscanf(dieArea, "%f %f %f %f", &x1, &y1, &x2, &y2); err != nil {
				t.Fatalf("Failed to parse DIE_AREA: %s", dieArea)
			}

			// Verify origin
			if x1 != 0 || y1 != 0 {
				t.Errorf("DIE_AREA origin not (0,0): (%f, %f)", x1, y1)
			}

			// Calculate expected die size (with 2μm margin)
			expectedWidth := float64(tc.cols)*tc.cellWidth + 2.0
			expectedHeight := float64(tc.rows)*tc.cellHeight + 2.0

			// Tolerance: 0.01 μm
			tolerance := 0.01
			if abs(x2-expectedWidth) > tolerance {
				t.Errorf("DIE_AREA width = %.3f μm, expected %.3f μm (delta %.3f μm)",
					x2, expectedWidth, abs(x2-expectedWidth))
			}
			if abs(y2-expectedHeight) > tolerance {
				t.Errorf("DIE_AREA height = %.3f μm, expected %.3f μm (delta %.3f μm)",
					y2, expectedHeight, abs(y2-expectedHeight))
			}

			t.Logf("M6-OL-02: DIE_AREA = %s (%.3f×%.3f μm) — PASS",
				dieArea, x2, y2)
		})
	}
}

// TestFlowConfigExtraFiles_M6_OL_02 tests EXTRA_LEFS, EXTRA_LIBS paths
func TestFlowConfigExtraFiles_M6_OL_02(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	cfg.Rows = 4
	cfg.Cols = 4

	configContent := export.GenerateOpenLaneConfig(cfg)

	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(configContent), &configMap); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}

	// Check EXTRA_LEFS
	extraLEFs, ok := configMap["EXTRA_LEFS"].(string)
	if !ok {
		t.Fatalf("EXTRA_LEFS not a string: %v", configMap["EXTRA_LEFS"])
	}
	if !strings.Contains(extraLEFs, "dir::") {
		t.Errorf("EXTRA_LEFS missing 'dir::' prefix: %s", extraLEFs)
	}
	if !strings.HasSuffix(extraLEFs, ".lef") {
		t.Errorf("EXTRA_LEFS missing .lef extension: %s", extraLEFs)
	}
	if !strings.Contains(extraLEFs, "fecim_bitcell") {
		t.Errorf("EXTRA_LEFS missing cell name: %s", extraLEFs)
	}
	t.Logf("M6-OL-02: EXTRA_LEFS = %s — PASS", extraLEFs)

	// Check EXTRA_LIBS
	extraLibs, ok := configMap["EXTRA_LIBS"].(string)
	if !ok {
		t.Fatalf("EXTRA_LIBS not a string: %v", configMap["EXTRA_LIBS"])
	}
	if !strings.Contains(extraLibs, "dir::") {
		t.Errorf("EXTRA_LIBS missing 'dir::' prefix: %s", extraLibs)
	}
	if !strings.HasSuffix(extraLibs, ".lib") {
		t.Errorf("EXTRA_LIBS missing .lib extension: %s", extraLibs)
	}
	if !strings.Contains(extraLibs, "fecim_bitcell") {
		t.Errorf("EXTRA_LIBS missing cell name: %s", extraLibs)
	}
	t.Logf("M6-OL-02: EXTRA_LIBS = %s — PASS", extraLibs)

	// Check EXTRA_GDS_FILES (optional, but should be present)
	if extraGDS, ok := configMap["EXTRA_GDS_FILES"].(string); ok {
		if !strings.HasSuffix(extraGDS, ".gds") {
			t.Errorf("EXTRA_GDS_FILES missing .gds extension: %s", extraGDS)
		}
		t.Logf("M6-OL-02: EXTRA_GDS_FILES = %s — PASS", extraGDS)
	}
}

// TestFlowConfigClockDisabled_M6_OL_02 tests RUN_CTS is disabled for crossbar
func TestFlowConfigClockDisabled_M6_OL_02(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	configContent := export.GenerateOpenLaneConfig(cfg)

	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(configContent), &configMap); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}

	// RUN_CTS should be 0 (crossbar has no clock)
	runCTS, ok := configMap["RUN_CTS"]
	if !ok {
		t.Errorf("RUN_CTS not set")
		return
	}

	// Can be int or float64 after JSON unmarshal
	var ctsValue float64
	switch v := runCTS.(type) {
	case float64:
		ctsValue = v
	case int:
		ctsValue = float64(v)
	default:
		t.Fatalf("RUN_CTS has unexpected type: %T", runCTS)
	}

	if ctsValue != 0 {
		t.Errorf("RUN_CTS = %v, want 0 (crossbar has no clock)", ctsValue)
	}

	t.Logf("M6-OL-02: RUN_CTS = 0 (clock disabled) — PASS")
}

// TestFlowConfigSynthMode_M6_OL_02 tests SYNTH_ELABORATE_ONLY flag
func TestFlowConfigSynthMode_M6_OL_02(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	configContent := export.GenerateOpenLaneConfig(cfg)

	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(configContent), &configMap); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}

	// SYNTH_ELABORATE_ONLY should be 1 (structural netlist, no logic mapping)
	synthElabOnly, ok := configMap["SYNTH_ELABORATE_ONLY"]
	if !ok {
		t.Errorf("SYNTH_ELABORATE_ONLY not set")
		return
	}

	var elabValue float64
	switch v := synthElabOnly.(type) {
	case float64:
		elabValue = v
	case int:
		elabValue = float64(v)
	default:
		t.Fatalf("SYNTH_ELABORATE_ONLY has unexpected type: %T", synthElabOnly)
	}

	if elabValue != 1 {
		t.Errorf("SYNTH_ELABORATE_ONLY = %v, want 1 (structural netlist only)", elabValue)
	}

	t.Logf("M6-OL-02: SYNTH_ELABORATE_ONLY = 1 — PASS")
}

// TestFlowConfigDeterminism_M6_OL_02 tests config generation is deterministic
func TestFlowConfigDeterminism_M6_OL_02(t *testing.T) {
	cfg := config.DefaultArrayConfig()
	cfg.Rows = 8
	cfg.Cols = 8

	// Generate config multiple times
	config1 := export.GenerateOpenLaneConfig(cfg)
	config2 := export.GenerateOpenLaneConfig(cfg)
	config3 := export.GenerateOpenLaneConfig(cfg)

	if config1 != config2 {
		t.Errorf("Config generation non-deterministic: run1 != run2")
	}
	if config2 != config3 {
		t.Errorf("Config generation non-deterministic: run2 != run3")
	}

	t.Logf("M6-OL-02: Config determinism — PASS (3 identical runs)")
}

// abs returns absolute value of float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
