// pkg/export/generators_extended_test.go
package export

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/module6-eda/pkg/config"
)

// ============================================================================
// SPICE Generator Tests
// ============================================================================

func TestGenerateSPICE_NonEmpty(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Name:      "test_array",
			Mode:      compiler.ModeStorage,
			ArrayRows: 2,
			ArrayCols: 2,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Level: 5, Conductance: 50.0, Resistance: 20000.0},
			{Row: 0, Col: 1, Level: 10, Conductance: 60.0, Resistance: 16666.67},
			{Row: 1, Col: 0, Level: 15, Conductance: 70.0, Resistance: 14285.71},
			{Row: 1, Col: 1, Level: 20, Conductance: 80.0, Resistance: 12500.0},
		},
		Stats: compiler.DesignStats{
			TotalCells:  4,
			ActiveCells: 4,
		},
	}

	result := GenerateSPICE(design, 1.8)

	if result == "" {
		t.Fatal("GenerateSPICE returned empty string")
	}

	if len(result) < 100 {
		t.Errorf("GenerateSPICE output too short: %d bytes", len(result))
	}
}

func TestGenerateSPICE_ContainsKeywords(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Name:      "test_array",
			Mode:      compiler.ModeCompute,
			ArrayRows: 2,
			ArrayCols: 2,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Level: 5, Conductance: 50.0, Resistance: 20000.0},
		},
		Stats: compiler.DesignStats{
			TotalCells:  4,
			ActiveCells: 1,
		},
	}

	result := GenerateSPICE(design, 3.3)

	keywords := []string{
		".param",
		"VDD",
		"* FeFET Cells",
		"* Word Line Drivers",
		"* Bit Line Loads",
		".control",
		".end",
	}

	for _, keyword := range keywords {
		if !strings.Contains(result, keyword) {
			t.Errorf("GenerateSPICE output missing keyword: %s", keyword)
		}
	}

	// Verify voltage parameter
	if !strings.Contains(result, "3.3") && !strings.Contains(result, "3.30") {
		t.Error("GenerateSPICE does not contain specified VDD value")
	}
}

func TestGenerateSPICE_DifferentModes(t *testing.T) {
	modes := []compiler.OperationMode{
		compiler.ModeStorage,
		compiler.ModeMemory,
		compiler.ModeCompute,
	}

	for _, mode := range modes {
		t.Run(mode.String(), func(t *testing.T) {
			design := &compiler.ArrayDesign{
				Config: &compiler.ArrayConfig{
					Mode:      mode,
					ArrayRows: 2,
					ArrayCols: 2,
				},
				Cells: []compiler.CellAssignment{
					{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0},
				},
				Stats: compiler.DesignStats{
					TotalCells:  4,
					ActiveCells: 1,
				},
			}

			result := GenerateSPICE(design, 1.8)

			if result == "" {
				t.Errorf("GenerateSPICE returned empty string for mode %s", mode.String())
			}

			if !strings.Contains(result, "* Operation Mode: "+mode.String()) {
				t.Errorf("GenerateSPICE output does not contain mode: %s", mode.String())
			}
		})
	}
}

// ============================================================================
// Liberty Generator Tests
// ============================================================================

func TestGenerateLiberty_Passive(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.CellType = "passive"

	result := GenerateLiberty(cfg)

	if result == "" {
		t.Fatal("GenerateLiberty returned empty string")
	}

	keywords := []string{
		"library(fecim_cells)",
		"technology (cmos)",
		"delay_model : table_lookup",
		"cell(fecim_bitcell)",
		"pin(WL)",
		"pin(BL)",
		"pin(VPWR)",
		"pin(VGND)",
		"operating_conditions(typical)",
	}

	for _, keyword := range keywords {
		if !strings.Contains(result, keyword) {
			t.Errorf("GenerateLiberty passive output missing keyword: %s", keyword)
		}
	}

	// Passive should NOT have SL or CSL pins
	if strings.Contains(result, "pin(SL)") {
		t.Error("Passive cell should not have SL pin")
	}
	if strings.Contains(result, "pin(CSL)") {
		t.Error("Passive cell should not have CSL pin")
	}
}

func TestGenerateLiberty_1T1R(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.CellType = "1t1r"

	result := GenerateLiberty(cfg)

	if result == "" {
		t.Fatal("GenerateLiberty returned empty string for 1T1R")
	}

	keywords := []string{
		"library(fecim_1t1r_cells)",
		"cell(fecim_1t1r_bitcell)",
		"pin(WL)",
		"pin(SL)",
		"pin(BL)",
		"function : \"(WL & SL)\"",
	}

	for _, keyword := range keywords {
		if !strings.Contains(result, keyword) {
			t.Errorf("GenerateLiberty 1T1R output missing keyword: %s", keyword)
		}
	}

	// 1T1R should NOT have CSL pin
	if strings.Contains(result, "pin(CSL)") {
		t.Error("1T1R cell should not have CSL pin")
	}
}

func TestGenerateLiberty_2T1R(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.CellType = "2t1r"

	result := GenerateLiberty(cfg)

	if result == "" {
		t.Fatal("GenerateLiberty returned empty string for 2T1R")
	}

	keywords := []string{
		"library(fecim_2t1r_cells)",
		"cell(fecim_2t1r_bitcell)",
		"pin(WL)",
		"pin(CSL)",
		"pin(SL)",
		"pin(BL)",
		"function : \"(WL & CSL & SL)\"",
	}

	for _, keyword := range keywords {
		if !strings.Contains(result, keyword) {
			t.Errorf("GenerateLiberty 2T1R output missing keyword: %s", keyword)
		}
	}
}

func TestGenerateLiberty_DifferentConfigs(t *testing.T) {
	tests := []struct {
		name     string
		voltage  float64
		temp     float64
		process  float64
		riseTime float64
		fallTime float64
	}{
		{"default", 1.8, 25.0, 1.0, 10.0, 10.0},
		{"high_voltage", 3.3, 25.0, 1.0, 8.0, 8.0},
		{"high_temp", 1.8, 85.0, 1.0, 12.0, 12.0},
		{"fast_corner", 1.8, 25.0, 0.9, 9.0, 9.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultCellConfig()
			cfg.Voltage = tt.voltage
			cfg.Temperature = tt.temp
			cfg.Process = tt.process
			cfg.RiseTime = tt.riseTime
			cfg.FallTime = tt.fallTime

			result := GenerateLiberty(cfg)

			if result == "" {
				t.Fatal("GenerateLiberty returned empty string")
			}

			// Should contain voltage and temperature values
			if !strings.Contains(result, "voltage :") {
				t.Error("Liberty output missing voltage specification")
			}
			if !strings.Contains(result, "temperature :") {
				t.Error("Liberty output missing temperature specification")
			}
		})
	}
}

// ============================================================================
// LEF Generator Tests
// ============================================================================

func TestGenerateLEF_Passive(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.CellType = "passive"

	result := GenerateLEF(cfg)

	if result == "" {
		t.Fatal("GenerateLEF returned empty string")
	}

	keywords := []string{
		"VERSION 5.8",
		"MACRO fecim_bitcell",
		"PIN WL",
		"PIN BL",
		"PIN VPWR",
		"PIN VGND",
		"LAYER met1",
		"END LIBRARY",
	}

	for _, keyword := range keywords {
		if !strings.Contains(result, keyword) {
			t.Errorf("GenerateLEF passive output missing keyword: %s", keyword)
		}
	}

	// Passive should NOT have SL or CSL pins
	if strings.Contains(result, "PIN SL") {
		t.Error("Passive cell LEF should not have SL pin")
	}
	if strings.Contains(result, "PIN CSL") {
		t.Error("Passive cell LEF should not have CSL pin")
	}
}

func TestGenerateLEF_1T1R(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.CellType = "1t1r"

	result := GenerateLEF(cfg)

	if result == "" {
		t.Fatal("GenerateLEF returned empty string for 1T1R")
	}

	keywords := []string{
		"VERSION 5.8",
		"MACRO fecim_1t1r_bitcell",
		"PIN WL",
		"PIN SL",
		"PIN BL",
		"SITE fecim_1t1r_site",
		"END LIBRARY",
	}

	for _, keyword := range keywords {
		if !strings.Contains(result, keyword) {
			t.Errorf("GenerateLEF 1T1R output missing keyword: %s", keyword)
		}
	}

	// 1T1R should NOT have CSL pin
	if strings.Contains(result, "PIN CSL") {
		t.Error("1T1R cell LEF should not have CSL pin")
	}
}

func TestGenerateLEF_2T1R(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.CellType = "2t1r"

	result := GenerateLEF(cfg)

	if result == "" {
		t.Fatal("GenerateLEF returned empty string for 2T1R")
	}

	keywords := []string{
		"VERSION 5.8",
		"MACRO fecim_2t1r_bitcell",
		"PIN WL",
		"PIN CSL",
		"PIN SL",
		"PIN BL",
		"SITE fecim_2t1r_site",
		"END LIBRARY",
	}

	for _, keyword := range keywords {
		if !strings.Contains(result, keyword) {
			t.Errorf("GenerateLEF 2T1R output missing keyword: %s", keyword)
		}
	}
}

func TestGenerateLEF_MetalParameters(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.MetalPitch = 0.50
	cfg.MetalWidth = 0.15

	result := GenerateLEF(cfg)

	if !strings.Contains(result, "PITCH 0.50") {
		t.Error("LEF output does not contain custom metal pitch")
	}
	if !strings.Contains(result, "WIDTH 0.15") {
		t.Error("LEF output does not contain custom metal width")
	}
}

// ============================================================================
// CSV Export Tests
// ============================================================================

func TestExportCSV_StorageMode(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:      compiler.ModeStorage,
			ArrayRows: 2,
			ArrayCols: 2,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Level: 5, Conductance: 50.0, Resistance: 20000.0, ProgramV: 3.5},
			{Row: 1, Col: 1, Level: 10, Conductance: 60.0, Resistance: 16666.67, ProgramV: 4.0},
		},
		Stats: compiler.DesignStats{
			TotalCells:  4,
			ActiveCells: 2,
		},
	}

	tmpFile := t.TempDir() + "/test.csv"
	err := ExportCSV(design, tmpFile)

	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	// Read the file to verify contents
	content := readFile(t, tmpFile)

	// Storage mode should NOT have weight column
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		t.Fatal("CSV file has fewer than 2 lines")
	}

	header := lines[0]
	if strings.Contains(header, "weight") {
		t.Error("Storage mode CSV should not have weight column")
	}

	expectedHeaders := []string{"row", "col", "level", "conductance_uS", "resistance_ohm", "program_V"}
	for _, h := range expectedHeaders {
		if !strings.Contains(header, h) {
			t.Errorf("CSV header missing: %s", h)
		}
	}
}

func TestExportCSV_ComputeModeWithWeights(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:      compiler.ModeCompute,
			ArrayRows: 2,
			ArrayCols: 2,
			ComputeConfig: &compiler.ComputeArrayConfig{
				InitialWeights: [][]float64{{0.5, 0.8}, {0.3, 0.9}},
			},
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Level: 5, Conductance: 50.0, Resistance: 20000.0, ProgramV: 3.5, InitialWeight: 0.5},
			{Row: 0, Col: 1, Level: 8, Conductance: 60.0, Resistance: 16666.67, ProgramV: 4.0, InitialWeight: 0.8},
		},
		Stats: compiler.DesignStats{
			TotalCells:  4,
			ActiveCells: 2,
		},
	}

	tmpFile := t.TempDir() + "/test_compute.csv"
	err := ExportCSV(design, tmpFile)

	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	content := readFile(t, tmpFile)
	lines := strings.Split(content, "\n")

	header := lines[0]
	if !strings.Contains(header, "weight") {
		t.Error("Compute mode CSV should have weight column")
	}

	// Check data row contains weight value
	if len(lines) > 1 && lines[1] != "" {
		if !strings.Contains(lines[1], "0.5") {
			t.Error("CSV data row should contain weight value 0.5")
		}
	}
}

// ============================================================================
// JSON Export Tests
// ============================================================================

func TestExportJSON_ValidFormat(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:      compiler.ModeMemory,
			ArrayRows: 2,
			ArrayCols: 2,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Level: 5, Conductance: 50.0},
		},
		Stats: compiler.DesignStats{
			TotalCells:  4,
			ActiveCells: 1,
		},
	}

	tmpFile := t.TempDir() + "/test.json"
	err := ExportJSON(design, tmpFile)

	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	content := readFile(t, tmpFile)

	// Verify it's valid JSON
	var parsed compiler.ArrayDesign
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		t.Fatalf("ExportJSON produced invalid JSON: %v", err)
	}

	// Verify structure
	if parsed.Config.Mode != compiler.ModeMemory {
		t.Errorf("JSON config mode mismatch: got %v, want %v", parsed.Config.Mode, compiler.ModeMemory)
	}
	if parsed.Stats.TotalCells != 4 {
		t.Errorf("JSON stats mismatch: got %d total cells, want 4", parsed.Stats.TotalCells)
	}
}

// ============================================================================
// Cell Verilog Generator Tests
// ============================================================================

func TestGenerateCellVerilog_Passive(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.CellType = "passive"

	result := GenerateCellVerilog(cfg)

	if result == "" {
		t.Fatal("GenerateCellVerilog returned empty string")
	}

	keywords := []string{
		"module fecim_bitcell",
		"input  wire WL",
		"output wire BL",
		"inout  wire VPWR",
		"inout  wire VGND",
		"endmodule",
		"assign BL = WL",
	}

	for _, keyword := range keywords {
		if !strings.Contains(result, keyword) {
			t.Errorf("GenerateCellVerilog passive output missing keyword: %s", keyword)
		}
	}

	// Passive should NOT have SL or CSL
	if strings.Contains(result, "input  wire SL") {
		t.Error("Passive cell should not have SL input")
	}
	if strings.Contains(result, "input  wire CSL") {
		t.Error("Passive cell should not have CSL input")
	}
}

func TestGenerateCellVerilog_1T1R(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.CellType = "1t1r"

	result := GenerateCellVerilog(cfg)

	if result == "" {
		t.Fatal("GenerateCellVerilog returned empty string for 1T1R")
	}

	keywords := []string{
		"module fecim_1t1r_bitcell",
		"input  wire WL",
		"output wire BL",
		"input  wire SL",
		"endmodule",
		"assign BL = WL ? SL : 1'bz",
	}

	for _, keyword := range keywords {
		if !strings.Contains(result, keyword) {
			t.Errorf("GenerateCellVerilog 1T1R output missing keyword: %s", keyword)
		}
	}

	// 1T1R should NOT have CSL
	if strings.Contains(result, "input  wire CSL") {
		t.Error("1T1R cell should not have CSL input")
	}
}

func TestGenerateCellVerilog_2T1R(t *testing.T) {
	cfg := config.DefaultCellConfig()
	cfg.CellType = "2t1r"

	result := GenerateCellVerilog(cfg)

	if result == "" {
		t.Fatal("GenerateCellVerilog returned empty string for 2T1R")
	}

	keywords := []string{
		"module fecim_2t1r_bitcell",
		"input  wire WL",
		"input  wire CSL",
		"output wire BL",
		"input  wire SL",
		"endmodule",
		"assign BL = (WL && CSL) ? SL : 1'bz",
	}

	for _, keyword := range keywords {
		if !strings.Contains(result, keyword) {
			t.Errorf("GenerateCellVerilog 2T1R output missing keyword: %s", keyword)
		}
	}
}

// ============================================================================
// Array Verilog Generator Tests
// ============================================================================

func TestGenerateArrayVerilog_Passive(t *testing.T) {
	cfg := config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Architecture: "passive",
		Mode:         "storage",
	}

	result := GenerateArrayVerilog(cfg)

	if result == "" {
		t.Fatal("GenerateArrayVerilog returned empty string")
	}

	keywords := []string{
		"module fecim_crossbar_4x4",
		"input  wire [3:0] WL",
		"inout  wire [3:0] BL",
		"inout  wire VPWR",
		"inout  wire VGND",
		"fecim_bitcell cell_0_0",
		"fecim_bitcell cell_3_3",
		"endmodule",
	}

	for _, keyword := range keywords {
		if !strings.Contains(result, keyword) {
			t.Errorf("GenerateArrayVerilog passive output missing keyword: %s", keyword)
		}
	}

	// Passive should NOT have SL or CSL
	if strings.Contains(result, "input  wire [3:0] SL") {
		t.Error("Passive array should not have SL port")
	}
	if strings.Contains(result, "input  wire [3:0] CSL") {
		t.Error("Passive array should not have CSL port")
	}
}

func TestGenerateArrayVerilog_1T1R(t *testing.T) {
	cfg := config.ArrayConfig{
		Rows:         2,
		Cols:         2,
		Architecture: "1t1r",
		Mode:         "compute",
	}

	result := GenerateArrayVerilog(cfg)

	if result == "" {
		t.Fatal("GenerateArrayVerilog returned empty string for 1T1R")
	}

	keywords := []string{
		"module fecim_crossbar_2x2",
		"input  wire [1:0] WL",
		"inout  wire [1:0] BL",
		"input  wire [1:0] SL",
		"fecim_1t1r_bitcell cell_0_0",
		".SL(SL[0])",
	}

	for _, keyword := range keywords {
		if !strings.Contains(result, keyword) {
			t.Errorf("GenerateArrayVerilog 1T1R output missing keyword: %s", keyword)
		}
	}

	// 1T1R should NOT have CSL
	if strings.Contains(result, "input  wire [1:0] CSL") {
		t.Error("1T1R array should not have CSL port")
	}
}

func TestGenerateArrayVerilog_2T1R(t *testing.T) {
	cfg := config.ArrayConfig{
		Rows:         2,
		Cols:         2,
		Architecture: "2t1r",
		Mode:         "compute",
	}

	result := GenerateArrayVerilog(cfg)

	if result == "" {
		t.Fatal("GenerateArrayVerilog returned empty string for 2T1R")
	}

	keywords := []string{
		"module fecim_crossbar_2x2",
		"input  wire [1:0] WL",
		"input  wire [1:0] CSL",
		"inout  wire [1:0] BL",
		"input  wire [1:0] SL",
		"fecim_2t1r_bitcell cell_0_0",
		".CSL(CSL[0])",
		".SL(SL[0])",
	}

	for _, keyword := range keywords {
		if !strings.Contains(result, keyword) {
			t.Errorf("GenerateArrayVerilog 2T1R output missing keyword: %s", keyword)
		}
	}
}

func TestGenerateArrayVerilog_DifferentSizes(t *testing.T) {
	tests := []struct {
		rows int
		cols int
	}{
		{2, 2},
		{4, 4},
		{8, 4},
		{16, 16},
	}

	for _, tt := range tests {
		t.Run(strings.ReplaceAll(t.Name(), "TestGenerateArrayVerilog_DifferentSizes/", ""), func(t *testing.T) {
			cfg := config.ArrayConfig{
				Rows:         tt.rows,
				Cols:         tt.cols,
				Architecture: "passive",
				Mode:         "memory",
			}

			result := GenerateArrayVerilog(cfg)

			if result == "" {
				t.Fatalf("GenerateArrayVerilog returned empty string for %dx%d", tt.rows, tt.cols)
			}

			// Verify cell instantiations
			expectedCells := tt.rows * tt.cols
			actualCells := strings.Count(result, "fecim_bitcell cell_")
			if actualCells != expectedCells {
				t.Errorf("Expected %d cell instantiations, got %d", expectedCells, actualCells)
			}
		})
	}
}

// ============================================================================
// OpenLane Config Generator Tests
// ============================================================================

func TestGenerateOpenLaneConfig_ValidJSON(t *testing.T) {
	cfg := config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Architecture: "passive",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	result := GenerateOpenLaneConfig(cfg)

	if result == "" {
		t.Fatal("GenerateOpenLaneConfig returned empty string")
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("GenerateOpenLaneConfig produced invalid JSON: %v", err)
	}

	// Verify expected keys exist
	expectedKeys := []string{
		"DESIGN_NAME",
		"VERILOG_FILES",
		"EXTRA_LEFS",
		"EXTRA_LIBS",
		"FP_DEF_TEMPLATE",
		"DIE_AREA",
		"RUN_CTS",
		"PL_SKIP_INITIAL_PLACEMENT",
		"SYNTH_ELABORATE_ONLY",
	}

	for _, key := range expectedKeys {
		if _, exists := parsed[key]; !exists {
			t.Errorf("OpenLane config missing key: %s", key)
		}
	}
}

func TestGenerateOpenLaneConfig_DesignName(t *testing.T) {
	cfg := config.ArrayConfig{
		Rows:         8,
		Cols:         16,
		Architecture: "1t1r",
		CellWidth:    0.92,
		CellHeight:   3.40,
	}

	result := GenerateOpenLaneConfig(cfg)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	designName, ok := parsed["DESIGN_NAME"].(string)
	if !ok {
		t.Fatal("DESIGN_NAME is not a string")
	}

	expected := "fecim_crossbar_8x16"
	if designName != expected {
		t.Errorf("DESIGN_NAME mismatch: got %s, want %s", designName, expected)
	}
}

func TestGenerateOpenLaneConfig_DieArea(t *testing.T) {
	cfg := config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Architecture: "passive",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	result := GenerateOpenLaneConfig(cfg)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	dieArea, ok := parsed["DIE_AREA"].(string)
	if !ok {
		t.Fatal("DIE_AREA is not a string")
	}

	// DIE_AREA should be "0 0 width height"
	if !strings.HasPrefix(dieArea, "0 0 ") {
		t.Errorf("DIE_AREA should start with '0 0 ': got %s", dieArea)
	}

	// Width = 4 * 0.46 + 2 = 3.84
	// Height = 4 * 2.72 + 2 = 12.88
	if !strings.Contains(dieArea, "3.840") || !strings.Contains(dieArea, "12.880") {
		t.Errorf("DIE_AREA dimensions incorrect: got %s", dieArea)
	}
}

func TestGenerateOpenLaneConfig_SkipFlags(t *testing.T) {
	cfg := config.ArrayConfig{
		Rows:         2,
		Cols:         2,
		Architecture: "passive",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}

	result := GenerateOpenLaneConfig(cfg)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify skip flags are set correctly
	if runCTS, ok := parsed["RUN_CTS"].(float64); !ok || runCTS != 0 {
		t.Error("RUN_CTS should be 0 (skip clock tree synthesis)")
	}

	if plSkip, ok := parsed["PL_SKIP_INITIAL_PLACEMENT"].(float64); !ok || plSkip != 1 {
		t.Error("PL_SKIP_INITIAL_PLACEMENT should be 1")
	}

	if synthElab, ok := parsed["SYNTH_ELABORATE_ONLY"].(float64); !ok || synthElab != 1 {
		t.Error("SYNTH_ELABORATE_ONLY should be 1")
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	return string(content)
}
