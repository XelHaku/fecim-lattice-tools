// pkg/export/def_test.go
package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

func TestGenerateDEF(t *testing.T) {
	mapping := getTestMapping()
	def := GenerateDEFWithDefaults(mapping)

	// Check VERSION
	if !strings.Contains(def, "VERSION 5.8") {
		t.Error("DEF should contain VERSION 5.8")
	}

	// Check DESIGN name
	if !strings.Contains(def, "DESIGN fecim_crossbar") {
		t.Error("DEF should contain DESIGN declaration")
	}

	// Check UNITS
	if !strings.Contains(def, "UNITS DISTANCE MICRONS 1000") {
		t.Error("DEF should have UNITS DISTANCE MICRONS 1000")
	}

	// Check DIEAREA
	if !strings.Contains(def, "DIEAREA") {
		t.Error("DEF should have DIEAREA section")
	}

	// Check COMPONENTS section - should have 6 cells (2x3 matrix)
	if !strings.Contains(def, "COMPONENTS 6") {
		t.Error("DEF should have COMPONENTS 6 for 2x3 matrix")
	}

	// Check component naming matches spice.go: R_{row}_{col}
	if !strings.Contains(def, "R_0_0 fecim_bit") {
		t.Error("DEF should use R_{row}_{col} naming convention")
	}
	if !strings.Contains(def, "R_1_2 fecim_bit") {
		t.Error("DEF should have R_1_2 instance for cell [1,2]")
	}

	// Check FIXED placement
	fixedCount := strings.Count(def, "+ FIXED (")
	if fixedCount < 6 {
		t.Errorf("All 6 components should have FIXED placement, found %d", fixedCount)
	}

	// Check PINS section
	if !strings.Contains(def, "PINS") {
		t.Error("DEF should have PINS section")
	}

	// Check NETS section
	if !strings.Contains(def, "NETS") {
		t.Error("DEF should have NETS section")
	}

	// Check END DESIGN
	if !strings.Contains(def, "END DESIGN") {
		t.Error("DEF should end with END DESIGN")
	}

	t.Logf("DEF generation: %d bytes", len(def))
}

func TestExportDEF(t *testing.T) {
	mapping := getTestMapping()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.def")

	err := ExportDEF(mapping, path)
	if err != nil {
		t.Fatalf("ExportDEF failed: %v", err)
	}

	// Verify file exists and has content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Exported DEF file is empty")
	}

	content := string(data)

	// Verify structure
	if !strings.Contains(content, "VERSION 5.8") {
		t.Error("DEF should contain VERSION")
	}
	if !strings.Contains(content, "END DESIGN") {
		t.Error("DEF should contain END DESIGN")
	}

	t.Logf("DEF export: %d bytes", len(data))
}

func TestDEFWithCustomConfig(t *testing.T) {
	mapping := getTestMapping()

	config := DEFConfig{
		DesignName:   "my_array",
		CellName:     "custom_cell",
		CellWidth:    1.0,
		CellHeight:   3.0,
		OriginX:      5.0,
		OriginY:      5.0,
		MarginX:      10.0,
		MarginY:      10.0,
		DatabaseUnit: 1000,
	}

	def := GenerateDEF(mapping, config)

	if !strings.Contains(def, "DESIGN my_array") {
		t.Error("DEF should use custom design name")
	}
	if !strings.Contains(def, "R_0_0 custom_cell") {
		t.Error("DEF should use custom cell name")
	}
}

func TestDEF4x4Array(t *testing.T) {
	// Create 4x4 weight matrix for Phase 1 validation
	weights := [][]float64{
		{0.1, 0.2, 0.3, 0.4},
		{0.5, 0.6, 0.7, 0.8},
		{0.9, 1.0, -0.1, -0.2},
		{-0.3, -0.4, -0.5, -0.6},
	}

	config := compiler.DefaultConfig()
	config.ArrayRows = 8
	config.ArrayCols = 8
	mapping, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	def := GenerateDEFWithDefaults(mapping)

	// Check 16 components
	if !strings.Contains(def, "COMPONENTS 16") {
		t.Error("4x4 array should have COMPONENTS 16")
	}

	// Check for corner cells
	if !strings.Contains(def, "R_0_0 fecim_bit") {
		t.Error("4x4 array should have R_0_0")
	}
	if !strings.Contains(def, "R_3_3 fecim_bit") {
		t.Error("4x4 array should have R_3_3")
	}

	// Check WL and BL pins
	if !strings.Contains(def, "- WL[0]") {
		t.Error("DEF should have WL[0] pin")
	}
	if !strings.Contains(def, "- WL[3]") {
		t.Error("DEF should have WL[3] pin")
	}
	if !strings.Contains(def, "- BL[0]") {
		t.Error("DEF should have BL[0] pin")
	}
	if !strings.Contains(def, "- BL[3]") {
		t.Error("DEF should have BL[3] pin")
	}

}

func TestDEFCoordinates(t *testing.T) {
	mapping := getTestMapping()
	config := DefaultDEFConfig()
	def := GenerateDEF(mapping, config)

	// Cell at [0,0] should be at origin (10um, 10um) = (10000, 10000) DBU
	if !strings.Contains(def, "R_0_0 fecim_bit + FIXED ( 10000 10000 )") {
		t.Error("R_0_0 should be at ( 10000 10000 )")
	}

	// Cell at [0,1] should be at (10um + 0.46um, 10um) = (10460, 10000) DBU
	if !strings.Contains(def, "R_0_1 fecim_bit + FIXED ( 10460 10000 )") {
		t.Error("R_0_1 should be at ( 10460 10000 )")
	}

	// Cell at [1,0] should be at (10um, 10um + 2.72um) = (10000, 12720) DBU
	if !strings.Contains(def, "R_1_0 fecim_bit + FIXED ( 10000 12720 )") {
		t.Error("R_1_0 should be at ( 10000 12720 )")
	}
}

func TestDEFInstanceNamesMatchVerilog(t *testing.T) {
	mapping := getTestMapping()

	verilog := GenerateVerilogWithDefaults(mapping)
	def := GenerateDEFWithDefaults(mapping)

	// Filter to only check active cells (those that get exported)
	cellsToExport := filterActiveCells(mapping)

	// Extract instance names from both
	// In Verilog: R_0_0, R_0_1, etc. in "fecim_bit #(...) R_0_0 ("
	// In DEF: R_0_0 in "- R_0_0 fecim_bit"

	for _, cell := range cellsToExport {
		instanceName := "R_" + itoa(cell.Row) + "_" + itoa(cell.Col)

		// Check Verilog has this instance
		if !strings.Contains(verilog, instanceName) {
			t.Errorf("Verilog missing instance %s", instanceName)
		}

		// Check DEF has this instance
		if !strings.Contains(def, instanceName) {
			t.Errorf("DEF missing instance %s", instanceName)
		}
	}

	// Verify instance counts match
	verilogCount := strings.Count(verilog, "fecim_bit #(")
	defCount := strings.Count(def, "- R_")
	if verilogCount != defCount {
		t.Errorf("Instance count mismatch: Verilog has %d, DEF has %d", verilogCount, defCount)
	}
	if verilogCount != len(cellsToExport) {
		t.Errorf("Expected %d instances, but Verilog has %d", len(cellsToExport), verilogCount)
	}
}

// Simple int to string without importing strconv
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string('0'+byte(n%10)) + s
		n /= 10
	}
	return s
}

// Phase 2 Tests: Architecture Configuration

func TestDEFPassiveArchitecture(t *testing.T) {
	weights := [][]float64{
		{0.1, 0.2},
		{0.3, 0.4},
	}

	config := compiler.DefaultConfig()
	config.ArrayRows = 4
	config.ArrayCols = 4
	config.Architecture = compiler.ArchPassive
	mapping, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	def := GenerateDEFWithDefaults(mapping)

	// Passive should have WL, BL but NO SL pins
	if !strings.Contains(def, "Architecture: passive") {
		t.Error("Should indicate passive architecture")
	}
	if !strings.Contains(def, "- WL[0]") {
		t.Error("Passive should have WL pins")
	}
	if !strings.Contains(def, "- BL[0]") {
		t.Error("Passive should have BL pins")
	}
	if strings.Contains(def, "- SL[") {
		t.Error("Passive architecture should NOT have SL pins")
	}

	// Should use fecim_bit cell
	if !strings.Contains(def, "fecim_bit +") {
		t.Error("Passive should use fecim_bit cell")
	}

	// Count pins: 2 WL + 2 BL + VPWR + VGND = 6
	if !strings.Contains(def, "PINS 6") {
		t.Error("Passive 2x2 array should have 6 pins")
	}

	t.Log("Passive architecture DEF generated successfully")
}

func TestDEF1T1RArchitecture(t *testing.T) {
	weights := [][]float64{
		{0.1, 0.2},
		{0.3, 0.4},
	}

	config := compiler.Config1T1R()
	config.ArrayRows = 4
	config.ArrayCols = 4
	mapping, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	def := GenerateDEFWithDefaults(mapping)

	// 1T1R should have WL, BL, AND SL pins
	if !strings.Contains(def, "Architecture: 1t1r") {
		t.Error("Should indicate 1T1R architecture")
	}
	if !strings.Contains(def, "- WL[0]") {
		t.Error("1T1R should have WL pins")
	}
	if !strings.Contains(def, "- BL[0]") {
		t.Error("1T1R should have BL pins")
	}
	if !strings.Contains(def, "- SL[0]") {
		t.Error("1T1R architecture MUST have SL pins")
	}
	if !strings.Contains(def, "- SL[1]") {
		t.Error("1T1R architecture MUST have SL[1] pin for 2 columns")
	}

	// Should use fecim_1t1r cell
	if !strings.Contains(def, "fecim_1t1r +") {
		t.Error("1T1R should use fecim_1t1r cell")
	}

	// Count pins: 2 WL + 2 BL + 2 SL + VPWR + VGND = 8
	if !strings.Contains(def, "PINS 8") {
		t.Error("1T1R 2x2 array should have 8 pins")
	}

	// SL nets should exist
	if !strings.Contains(def, "- SL[0]\n") {
		t.Error("1T1R should have SL[0] net")
	}

	t.Log("1T1R architecture DEF generated successfully with SL[] pins")
}

func TestDEFCellDimensionsFromConfig(t *testing.T) {
	weights := [][]float64{
		{0.1, 0.2},
		{0.3, 0.4},
	}

	// Use 1T1R config which has larger cell pitch (0.92um vs 0.46um)
	config := compiler.Config1T1R()
	config.ArrayRows = 4
	config.ArrayCols = 4
	mapping, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	def := GenerateDEFWithDefaults(mapping)

	// Cell at [0,1] should be at (10um + 0.92um, 10um) = (10920, 10000) DBU
	// (Passive was 10460, 1T1R should be 10920 due to larger cell pitch)
	if !strings.Contains(def, "R_0_1 fecim_1t1r + FIXED ( 10920 10000 )") {
		t.Error("1T1R R_0_1 should be at ( 10920 10000 ) due to larger cell pitch")
	}

	t.Log("1T1R cell dimensions correctly applied from CompileConfig")
}
