// pkg/export/verilog_test.go
package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

func TestGenerateVerilog(t *testing.T) {
	mapping := getTestMapping()
	verilog := GenerateVerilogWithDefaults(mapping)

	// Check module declaration
	if !strings.Contains(verilog, "module fecim_crossbar") {
		t.Error("Verilog should contain module declaration")
	}

	// Check port declarations
	if !strings.Contains(verilog, "input  wire [1:0] WL") {
		t.Error("Verilog should have WL port declaration for 2 rows")
	}
	if !strings.Contains(verilog, "inout  wire [2:0] BL") {
		t.Error("Verilog should have BL port declaration for 3 cols")
	}

	// Check cell instances - should have 6 cells (2x3 matrix)
	instanceCount := strings.Count(verilog, "fecim_bit #(")
	if instanceCount != 6 {
		t.Errorf("Expected 6 cell instances, got %d", instanceCount)
	}

	// Check naming convention matches spice.go: R_{row}_{col}
	if !strings.Contains(verilog, "R_0_0") {
		t.Error("Verilog should use R_{row}_{col} naming convention")
	}
	if !strings.Contains(verilog, "R_1_2") {
		t.Error("Verilog should have R_1_2 instance for cell [1,2]")
	}

	// Check endmodule
	if !strings.Contains(verilog, "endmodule") {
		t.Error("Verilog should end with endmodule")
	}

	t.Logf("Verilog generation: %d bytes, %d instances", len(verilog), instanceCount)
}

func TestExportVerilog(t *testing.T) {
	mapping := getTestMapping()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.v")

	err := ExportVerilog(mapping, path)
	if err != nil {
		t.Fatalf("ExportVerilog failed: %v", err)
	}

	// Verify file exists and has content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Exported Verilog file is empty")
	}

	content := string(data)

	// Verify structure
	if !strings.Contains(content, "module fecim_crossbar") {
		t.Error("Verilog should contain module declaration")
	}
	if !strings.Contains(content, "endmodule") {
		t.Error("Verilog should contain endmodule")
	}

	t.Logf("Verilog export: %d bytes", len(data))
}

func TestVerilogWithCustomConfig(t *testing.T) {
	mapping := getTestMapping()

	config := VerilogConfig{
		ModuleName: "my_crossbar",
		CellName:   "custom_cell",
	}

	verilog := GenerateVerilog(mapping, config)

	if !strings.Contains(verilog, "module my_crossbar") {
		t.Error("Verilog should use custom module name")
	}
	if !strings.Contains(verilog, "custom_cell #(") {
		t.Error("Verilog should use custom cell name")
	}
}

func TestVerilog4x4Array(t *testing.T) {
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

	verilog := GenerateVerilogWithDefaults(mapping)

	// Check 4 rows, 4 cols
	if !strings.Contains(verilog, "input  wire [3:0] WL") {
		t.Error("4x4 array should have WL[3:0]")
	}
	if !strings.Contains(verilog, "inout  wire [3:0] BL") {
		t.Error("4x4 array should have BL[3:0]")
	}

	// Check 16 instances
	instanceCount := strings.Count(verilog, "fecim_bit #(")
	if instanceCount != 16 {
		t.Errorf("4x4 array should have 16 instances, got %d", instanceCount)
	}

}

// Phase 2 Tests: Architecture Configuration

func TestVerilogPassiveArchitecture(t *testing.T) {
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

	verilog := GenerateVerilogWithDefaults(mapping)

	// Passive should have WL, BL but NO SL
	if !strings.Contains(verilog, "Architecture: passive") {
		t.Error("Should indicate passive architecture")
	}
	if !strings.Contains(verilog, "input  wire [1:0] WL") {
		t.Error("Passive should have WL ports")
	}
	if !strings.Contains(verilog, "inout  wire [1:0] BL") {
		t.Error("Passive should have BL ports")
	}
	if strings.Contains(verilog, "SL") {
		t.Error("Passive architecture should NOT have SL ports")
	}

	// Should use fecim_bit cell
	if !strings.Contains(verilog, "fecim_bit #(") {
		t.Error("Passive should use fecim_bit cell")
	}

	t.Log("Passive architecture Verilog generated successfully")
}

func TestVerilog1T1RArchitecture(t *testing.T) {
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

	verilog := GenerateVerilogWithDefaults(mapping)

	// 1T1R should have WL, BL, AND SL
	if !strings.Contains(verilog, "Architecture: 1t1r") {
		t.Error("Should indicate 1T1R architecture")
	}
	if !strings.Contains(verilog, "input  wire [1:0] WL") {
		t.Error("1T1R should have WL ports")
	}
	if !strings.Contains(verilog, "inout  wire [1:0] BL") {
		t.Error("1T1R should have BL ports")
	}
	if !strings.Contains(verilog, "input  wire [1:0] SL") {
		t.Error("1T1R architecture MUST have SL ports")
	}

	// Should use fecim_1t1r cell
	if !strings.Contains(verilog, "fecim_1t1r #(") {
		t.Error("1T1R should use fecim_1t1r cell")
	}

	// Cell instances should have SL pin connections
	if !strings.Contains(verilog, ".SL  (SL[") {
		t.Error("1T1R cell instances should connect SL pins")
	}

	t.Log("1T1R architecture Verilog generated successfully with SL[] ports")
}

func TestVerilog2T1RArchitecture(t *testing.T) {
	weights := [][]float64{
		{0.1, 0.2},
		{0.3, 0.4},
	}

	config := compiler.Config2T1R()
	config.ArrayRows = 4
	config.ArrayCols = 4
	mapping, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	verilog := GenerateVerilogWithDefaults(mapping)

	// 2T1R should have WL, BL, SL, AND CSL
	if !strings.Contains(verilog, "Architecture: 2t1r") {
		t.Error("Should indicate 2T1R architecture")
	}
	if !strings.Contains(verilog, "input  wire [1:0] WL") {
		t.Error("2T1R should have WL ports")
	}
	if !strings.Contains(verilog, "inout  wire [1:0] BL") {
		t.Error("2T1R should have BL ports")
	}
	if !strings.Contains(verilog, "input  wire [1:0] SL") {
		t.Error("2T1R architecture MUST have SL ports")
	}
	if !strings.Contains(verilog, "input  wire [1:0] CSL") {
		t.Error("2T1R architecture MUST have CSL ports (Column Select Lines)")
	}

	// Should use fecim_2t1r cell
	if !strings.Contains(verilog, "fecim_2t1r #(") {
		t.Error("2T1R should use fecim_2t1r cell")
	}

	// Cell instances should have SL and CSL pin connections
	if !strings.Contains(verilog, ".SL  (SL[") {
		t.Error("2T1R cell instances should connect SL pins")
	}
	if !strings.Contains(verilog, ".CSL (CSL[") {
		t.Error("2T1R cell instances should connect CSL pins")
	}

	// Should have the 2T1R note in comments
	if !strings.Contains(verilog, "2T1R architecture includes Source Lines (SL) and Column Select Lines (CSL)") {
		t.Error("2T1R should have descriptive comment about architecture")
	}

	t.Log("2T1R architecture Verilog generated successfully with SL[] and CSL[] ports")
}
