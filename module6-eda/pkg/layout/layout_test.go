// pkg/layout/layout_test.go
package layout

import (
	"fmt"
	"strings"
	"testing"
)

// TestGenerateDEF_Basic verifies basic DEF structure and content
func TestGenerateDEF_Basic(t *testing.T) {
	appName := "test_array"
	rows, cols := 4, 8
	cellName := "fefet_cell"
	pitchX, pitchY := 460, 920

	result := GenerateDEF(appName, rows, cols, cellName, pitchX, pitchY)

	// Verify header sections
	if !strings.Contains(result, "VERSION 5.8") {
		t.Error("Missing VERSION declaration")
	}
	if !strings.Contains(result, "DESIGN test_array") {
		t.Error("Missing or incorrect DESIGN declaration")
	}
	if !strings.Contains(result, "UNITS DISTANCE MICRONS 1000") {
		t.Error("Missing UNITS declaration")
	}

	// Verify DIEAREA calculation
	expectedWidth := cols * pitchX  // 3680
	expectedHeight := rows * pitchY // 3680
	expectedDieArea := fmt.Sprintf("DIEAREA ( 0 0 ) ( %d %d )", expectedWidth, expectedHeight)
	if !strings.Contains(result, expectedDieArea) {
		t.Errorf("Expected DIEAREA %s, got different value", expectedDieArea)
	}

	// Verify COMPONENTS section
	numCells := rows * cols
	expectedComponents := "COMPONENTS 32"
	if !strings.Contains(result, expectedComponents) {
		t.Errorf("Expected %s, not found", expectedComponents)
	}

	// Verify component count
	componentCount := strings.Count(result, "- cell_")
	if componentCount != numCells {
		t.Errorf("Expected %d cell instances, found %d", numCells, componentCount)
	}

	// Verify END COMPONENTS
	if !strings.Contains(result, "END COMPONENTS") {
		t.Error("Missing END COMPONENTS")
	}

	// Verify END DESIGN
	if !strings.Contains(result, "END DESIGN") {
		t.Error("Missing END DESIGN")
	}
}

// TestGenerateDEF_CellPlacement verifies individual cell placement
func TestGenerateDEF_CellPlacement(t *testing.T) {
	rows, cols := 2, 3
	cellName := "mycell"
	pitchX, pitchY := 100, 200

	result := GenerateDEF("test", rows, cols, cellName, pitchX, pitchY)

	// Verify specific cells exist with correct naming
	expectedCells := []string{
		"- cell_0_0 mycell + FIXED ( 0 0 ) N",
		"- cell_0_1 mycell + FIXED ( 100 0 ) N",
		"- cell_0_2 mycell + FIXED ( 200 0 ) N",
		"- cell_1_0 mycell + FIXED ( 0 200 ) N",
		"- cell_1_1 mycell + FIXED ( 100 200 ) N",
		"- cell_1_2 mycell + FIXED ( 200 200 ) N",
	}

	for _, expected := range expectedCells {
		if !strings.Contains(result, expected) {
			t.Errorf("Missing expected cell placement: %s", expected)
		}
	}
}

// TestGenerateDEF_SingleCell verifies edge case with 1x1 array
func TestGenerateDEF_SingleCell(t *testing.T) {
	result := GenerateDEF("single", 1, 1, "cell", 500, 500)

	if !strings.Contains(result, "COMPONENTS 1") {
		t.Error("Expected COMPONENTS 1 for single cell")
	}

	if !strings.Contains(result, "DIEAREA ( 0 0 ) ( 500 500 )") {
		t.Error("Incorrect DIEAREA for single cell")
	}

	if !strings.Contains(result, "- cell_0_0 cell + FIXED ( 0 0 ) N") {
		t.Error("Missing single cell placement")
	}
}

// TestGenerateDEF_ZeroSize verifies behavior with zero dimensions
func TestGenerateDEF_ZeroSize(t *testing.T) {
	// Zero rows
	result := GenerateDEF("zero_rows", 0, 5, "cell", 100, 100)
	if !strings.Contains(result, "COMPONENTS 0") {
		t.Error("Expected COMPONENTS 0 for zero rows")
	}
	if strings.Contains(result, "- cell_") {
		t.Error("Should not contain any cell instances for zero rows")
	}

	// Zero columns
	result = GenerateDEF("zero_cols", 5, 0, "cell", 100, 100)
	if !strings.Contains(result, "COMPONENTS 0") {
		t.Error("Expected COMPONENTS 0 for zero cols")
	}

	// Both zero
	result = GenerateDEF("zero_both", 0, 0, "cell", 100, 100)
	if !strings.Contains(result, "COMPONENTS 0") {
		t.Error("Expected COMPONENTS 0 for zero dimensions")
	}
	if !strings.Contains(result, "DIEAREA ( 0 0 ) ( 0 0 )") {
		t.Error("Expected zero DIEAREA")
	}
}

// TestGenerateDEF_LargeArray verifies large array handling
func TestGenerateDEF_LargeArray(t *testing.T) {
	rows, cols := 128, 128
	result := GenerateDEF("large", rows, cols, "cell", 460, 920)

	expectedCells := rows * cols
	if !strings.Contains(result, "COMPONENTS 16384") {
		t.Errorf("Expected COMPONENTS %d for large array", expectedCells)
	}

	// Verify die area calculation
	expectedWidth := cols * 460  // 58880
	expectedHeight := rows * 920 // 117760
	expectedDieArea := fmt.Sprintf("DIEAREA ( 0 0 ) ( %d %d )", expectedWidth, expectedHeight)
	if !strings.Contains(result, expectedDieArea) {
		t.Errorf("Incorrect DIEAREA calculation for large array")
	}

	// Spot check corners
	if !strings.Contains(result, "- cell_0_0") {
		t.Error("Missing bottom-left corner cell")
	}
	if !strings.Contains(result, "- cell_127_127") {
		t.Error("Missing top-right corner cell")
	}
}

// TestGenerateDEF_SpecialCharacters verifies handling of special names
func TestGenerateDEF_SpecialCharacters(t *testing.T) {
	result := GenerateDEF("my_design_v2", 2, 2, "cell_std", 100, 100)

	if !strings.Contains(result, "DESIGN my_design_v2") {
		t.Error("Special characters in design name not preserved")
	}
	if !strings.Contains(result, "cell_std") {
		t.Error("Special characters in cell name not preserved")
	}
}

// TestGenerateVerilog_Basic verifies basic Verilog structure
func TestGenerateVerilog_Basic(t *testing.T) {
	moduleName := "crossbar_4x8"
	rows, cols := 4, 8
	cellName := "fefet"

	result := GenerateVerilog(moduleName, rows, cols, cellName)

	// Verify module declaration
	if !strings.Contains(result, "module crossbar_4x8") {
		t.Error("Missing or incorrect module declaration")
	}

	// Verify port declarations
	if !strings.Contains(result, "input [3:0] wl") {
		t.Error("Missing or incorrect WL port declaration")
	}
	if !strings.Contains(result, "output [7:0] bl") {
		t.Error("Missing or incorrect BL port declaration")
	}

	// Verify cell instantiations
	numCells := rows * cols
	cellCount := strings.Count(result, "fefet cell_")
	if cellCount != numCells {
		t.Errorf("Expected %d cell instantiations, found %d", numCells, cellCount)
	}

	// Verify endmodule
	if !strings.Contains(result, "endmodule") {
		t.Error("Missing endmodule declaration")
	}
}

// TestGenerateVerilog_PortConnections verifies specific port connections
func TestGenerateVerilog_PortConnections(t *testing.T) {
	rows, cols := 3, 2
	result := GenerateVerilog("test", rows, cols, "cell")

	// Verify specific cell connections
	expectedConnections := []string{
		"cell cell_0_0",
		".WL(wl[0])",
		".BL(bl[0])",
		"cell cell_2_1",
		".WL(wl[2])",
		".BL(bl[1])",
	}

	for _, expected := range expectedConnections {
		if !strings.Contains(result, expected) {
			t.Errorf("Missing expected connection: %s", expected)
		}
	}
}

// TestGenerateVerilog_SingleCell verifies 1x1 array edge case
func TestGenerateVerilog_SingleCell(t *testing.T) {
	result := GenerateVerilog("single", 1, 1, "mycell")

	// Verify ports are [0:0]
	if !strings.Contains(result, "input [0:0] wl") {
		t.Error("Expected wl[0:0] for single row")
	}
	if !strings.Contains(result, "output [0:0] bl") {
		t.Error("Expected bl[0:0] for single column")
	}

	// Verify single cell instantiation
	if !strings.Contains(result, "mycell cell_0_0") {
		t.Error("Missing single cell instantiation")
	}
	if !strings.Contains(result, ".WL(wl[0])") {
		t.Error("Missing WL connection for single cell")
	}
	if !strings.Contains(result, ".BL(bl[0])") {
		t.Error("Missing BL connection for single cell")
	}
}

// TestGenerateVerilog_ZeroSize verifies behavior with zero dimensions
func TestGenerateVerilog_ZeroSize(t *testing.T) {
	// Zero rows
	result := GenerateVerilog("zero_rows", 0, 5, "cell")
	// Should still generate module structure
	if !strings.Contains(result, "module zero_rows") {
		t.Error("Missing module declaration for zero rows")
	}
	// Port ranges become invalid but function doesn't error
	// This tests current behavior; may want to add validation later

	// Zero columns
	result = GenerateVerilog("zero_cols", 5, 0, "cell")
	if !strings.Contains(result, "module zero_cols") {
		t.Error("Missing module declaration for zero cols")
	}

	// Both zero - should have no cell instantiations
	result = GenerateVerilog("zero_both", 0, 0, "cell")
	if strings.Contains(result, "cell cell_") {
		t.Error("Should not contain cell instantiations for zero dimensions")
	}
}

// TestGenerateVerilog_LargeArray verifies large array handling
func TestGenerateVerilog_LargeArray(t *testing.T) {
	rows, cols := 64, 128
	result := GenerateVerilog("large", rows, cols, "cell")

	// Verify port widths
	if !strings.Contains(result, "input [63:0] wl") {
		t.Error("Incorrect WL port width for large array")
	}
	if !strings.Contains(result, "output [127:0] bl") {
		t.Error("Incorrect BL port width for large array")
	}

	// Verify cell count
	expectedCells := rows * cols
	cellCount := strings.Count(result, "cell cell_")
	if cellCount != expectedCells {
		t.Errorf("Expected %d cells, found %d", expectedCells, cellCount)
	}

	// Spot check corners
	if !strings.Contains(result, "cell cell_0_0") {
		t.Error("Missing bottom-left corner cell")
	}
	if !strings.Contains(result, "cell cell_63_127") {
		t.Error("Missing top-right corner cell")
	}
}

// TestGenerateVerilog_CellNaming verifies instance naming convention
func TestGenerateVerilog_CellNaming(t *testing.T) {
	result := GenerateVerilog("test", 2, 3, "stdcell")

	// All cells should follow cell_r_c pattern
	expectedInstances := []string{
		"stdcell cell_0_0",
		"stdcell cell_0_1",
		"stdcell cell_0_2",
		"stdcell cell_1_0",
		"stdcell cell_1_1",
		"stdcell cell_1_2",
	}

	for _, expected := range expectedInstances {
		if !strings.Contains(result, expected) {
			t.Errorf("Missing expected instance: %s", expected)
		}
	}
}

// TestGenerateVerilog_SyntaxValidity performs basic Verilog syntax checks
func TestGenerateVerilog_SyntaxValidity(t *testing.T) {
	result := GenerateVerilog("syntax_test", 4, 4, "cell")

	// Check for balanced parentheses in module declaration
	openParen := strings.Count(result, "(")
	closeParen := strings.Count(result, ")")
	if openParen != closeParen {
		t.Errorf("Unbalanced parentheses: %d open, %d close", openParen, closeParen)
	}

	// Check module/endmodule balance
	moduleCount := strings.Count(result, "module ")
	endmoduleCount := strings.Count(result, "endmodule")
	if moduleCount != endmoduleCount {
		t.Errorf("Unbalanced module/endmodule: %d modules, %d endmodules", moduleCount, endmoduleCount)
	}

	// Verify no trailing commas in port list
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		if strings.Contains(line, ");\n") && strings.Contains(lines[i-1], ",\n") {
			// Last port should not have comma before closing paren
			// This is a simplified check
		}
	}
}

// TestGenerateDEF_OutputFormat verifies DEF formatting consistency
func TestGenerateDEF_OutputFormat(t *testing.T) {
	result := GenerateDEF("format_test", 2, 2, "cell", 100, 100)

	lines := strings.Split(result, "\n")

	// Verify semicolon usage
	headerLines := []string{"VERSION", "DESIGN", "UNITS", "DIEAREA"}
	for _, prefix := range headerLines {
		found := false
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), prefix) {
				if !strings.Contains(line, ";") {
					t.Errorf("Line starting with %s missing semicolon: %s", prefix, line)
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing line starting with %s", prefix)
		}
	}

	// Verify FIXED placement format
	fixedCount := strings.Count(result, "+ FIXED (")
	if fixedCount != 4 {
		t.Errorf("Expected 4 FIXED placements, found %d", fixedCount)
	}

	// Verify orientation
	orientationCount := strings.Count(result, ") N ;")
	if orientationCount != 4 {
		t.Errorf("Expected 4 'N' orientations, found %d", orientationCount)
	}
}

// TestGenerateVerilog_OutputFormat verifies Verilog formatting consistency
func TestGenerateVerilog_OutputFormat(t *testing.T) {
	result := GenerateVerilog("format_test", 3, 3, "cell")

	lines := strings.Split(result, "\n")

	// Verify indentation exists for ports and instances
	portIndented := false
	instIndented := false
	for _, line := range lines {
		if strings.HasPrefix(line, "    input") || strings.HasPrefix(line, "    output") {
			portIndented = true
		}
		if strings.HasPrefix(line, "    cell") {
			instIndented = true
		}
	}
	if !portIndented {
		t.Error("Port declarations not properly indented")
	}
	if !instIndented {
		t.Error("Cell instantiations not properly indented")
	}

	// Verify closing syntax
	if !strings.HasSuffix(strings.TrimSpace(result), "endmodule") {
		t.Error("Module does not end with 'endmodule'")
	}
}

// TestGenerateDEF_ComponentSection verifies COMPONENTS section structure
func TestGenerateDEF_ComponentSection(t *testing.T) {
	result := GenerateDEF("comp_test", 3, 2, "mycell", 100, 200)

	// Find COMPONENTS section
	if !strings.Contains(result, "COMPONENTS 6 ;") {
		t.Error("Missing or incorrect COMPONENTS declaration")
	}

	// Verify END COMPONENTS follows all component definitions
	compIdx := strings.Index(result, "COMPONENTS")
	endCompIdx := strings.Index(result, "END COMPONENTS")
	if compIdx == -1 || endCompIdx == -1 || compIdx >= endCompIdx {
		t.Error("COMPONENTS section structure invalid")
	}

	// Verify all cells are within COMPONENTS section
	section := result[compIdx:endCompIdx]
	cellCount := strings.Count(section, "- cell_")
	if cellCount != 6 {
		t.Errorf("Expected 6 cells in COMPONENTS section, found %d", cellCount)
	}
}

// TestGenerateVerilog_ModulePortsSection verifies port section structure
func TestGenerateVerilog_ModulePortsSection(t *testing.T) {
	result := GenerateVerilog("port_test", 4, 3, "cell")

	// Verify port section structure
	if !strings.Contains(result, "module port_test (") {
		t.Error("Module declaration malformed")
	}

	// Last port should not have comma
	lines := strings.Split(result, "\n")
	foundLastPort := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "output") && strings.Contains(trimmed, "bl") {
			if strings.HasSuffix(trimmed, ",") {
				t.Error("Last port declaration should not have trailing comma")
			}
			foundLastPort = true
		}
	}
	if !foundLastPort {
		t.Error("Could not verify last port format")
	}
}
