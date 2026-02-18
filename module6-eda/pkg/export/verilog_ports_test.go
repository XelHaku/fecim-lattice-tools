// Phase 3: Verilog & Digital Export
// M6-VER-04: Verilog Port Consistency Validation
//
// Tests:
// - Input port width = M (columns)
// - Output port width = N (rows) × output_bits
// - Verify no width mismatches

package export

import (
	"fmt"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

// TestM6VER04_InputPortWidths verifies WL input port width matches rows
func TestM6VER04_InputPortWidths(t *testing.T) {
	testCases := []struct {
		name        string
		rows        int
		cols        int
		expectedWL  string // Expected WL port declaration
		expectedSL  string // Expected SL port (for 1T1R/2T1R)
		expectedCSL string // Expected CSL port (for 2T1R only)
	}{
		{
			name:        "4×4 array",
			rows:        4,
			cols:        4,
			expectedWL:  "input  wire [3:0] WL",
			expectedSL:  "input  wire [3:0] SL",
			expectedCSL: "input  wire [3:0] CSL",
		},
		{
			name:        "2×2 array",
			rows:        2,
			cols:        2,
			expectedWL:  "input  wire [1:0] WL",
			expectedSL:  "input  wire [1:0] SL",
			expectedCSL: "input  wire [1:0] CSL",
		},
		{
			name:        "8×4 array",
			rows:        8,
			cols:        4,
			expectedWL:  "input  wire [7:0] WL",
			expectedSL:  "input  wire [3:0] SL",
			expectedCSL: "input  wire [3:0] CSL",
		},
		{
			name:        "1×1 array (edge case)",
			rows:        1,
			cols:        1,
			expectedWL:  "input  wire [0:0] WL",
			expectedSL:  "input  wire [0:0] SL",
			expectedCSL: "input  wire [0:0] CSL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name+" (passive)", func(t *testing.T) {
			weights := make([][]float64, tc.rows)
			for i := range weights {
				weights[i] = make([]float64, tc.cols)
				for j := range weights[i] {
					weights[i][j] = float64(i*tc.cols+j) * 0.1
				}
			}

			config := compiler.DefaultConfig()
			config.ArrayRows = tc.rows * 2
			config.ArrayCols = tc.cols * 2
			config.Architecture = compiler.ArchPassive
			design, err := compiler.Compile(weights, config)
			if err != nil {
				t.Fatalf("Compile failed: %v", err)
			}

			verilog := GenerateVerilogWithDefaults(design)

			// M6-VER-04a: Check WL port width (rows)
			if !strings.Contains(verilog, tc.expectedWL) {
				t.Errorf("M6-VER-04a FAIL: Expected '%s' for %d rows", tc.expectedWL, tc.rows)
				// Show what we got
				lines := strings.Split(verilog, "\n")
				for _, line := range lines {
					if strings.Contains(line, "WL") && strings.Contains(line, "input") {
						t.Logf("  Found: %s", strings.TrimSpace(line))
					}
				}
			} else {
				t.Logf("M6-VER-04a PASS: WL port width = [%d:0] for %d rows", tc.rows-1, tc.rows)
			}

			// M6-VER-04b: Passive should NOT have SL ports
			if strings.Contains(verilog, "SL") {
				t.Error("M6-VER-04b FAIL: Passive architecture should NOT have SL ports")
			} else {
				t.Log("M6-VER-04b PASS: Passive architecture has no SL ports")
			}
		})

		t.Run(tc.name+" (1T1R)", func(t *testing.T) {
			weights := make([][]float64, tc.rows)
			for i := range weights {
				weights[i] = make([]float64, tc.cols)
				for j := range weights[i] {
					weights[i][j] = float64(i*tc.cols+j) * 0.1
				}
			}

			config := compiler.Config1T1R()
			config.ArrayRows = tc.rows * 2
			config.ArrayCols = tc.cols * 2
			design, err := compiler.Compile(weights, config)
			if err != nil {
				t.Fatalf("Compile failed: %v", err)
			}

			verilog := GenerateVerilogWithDefaults(design)

			// M6-VER-04c: Check SL port width (columns) for 1T1R
			if !strings.Contains(verilog, tc.expectedSL) {
				t.Errorf("M6-VER-04c FAIL: Expected '%s' for %d columns (1T1R)", tc.expectedSL, tc.cols)
			} else {
				t.Logf("M6-VER-04c PASS: SL port width = [%d:0] for %d columns (1T1R)", tc.cols-1, tc.cols)
			}

			// M6-VER-04d: 1T1R should NOT have CSL
			if strings.Contains(verilog, "CSL") {
				t.Error("M6-VER-04d FAIL: 1T1R should NOT have CSL ports")
			} else {
				t.Log("M6-VER-04d PASS: 1T1R has no CSL ports")
			}
		})

		t.Run(tc.name+" (2T1R)", func(t *testing.T) {
			weights := make([][]float64, tc.rows)
			for i := range weights {
				weights[i] = make([]float64, tc.cols)
				for j := range weights[i] {
					weights[i][j] = float64(i*tc.cols+j) * 0.1
				}
			}

			config := compiler.Config2T1R()
			config.ArrayRows = tc.rows * 2
			config.ArrayCols = tc.cols * 2
			design, err := compiler.Compile(weights, config)
			if err != nil {
				t.Fatalf("Compile failed: %v", err)
			}

			verilog := GenerateVerilogWithDefaults(design)

			// M6-VER-04e: Check SL port width (columns) for 2T1R
			if !strings.Contains(verilog, tc.expectedSL) {
				t.Errorf("M6-VER-04e FAIL: Expected '%s' for %d columns (2T1R)", tc.expectedSL, tc.cols)
			} else {
				t.Logf("M6-VER-04e PASS: SL port width = [%d:0] for %d columns (2T1R)", tc.cols-1, tc.cols)
			}

			// M6-VER-04f: Check CSL port width (columns) for 2T1R
			if !strings.Contains(verilog, tc.expectedCSL) {
				t.Errorf("M6-VER-04f FAIL: Expected '%s' for %d columns (2T1R)", tc.expectedCSL, tc.cols)
			} else {
				t.Logf("M6-VER-04f PASS: CSL port width = [%d:0] for %d columns (2T1R)", tc.cols-1, tc.cols)
			}
		})
	}
}

// TestM6VER04_OutputPortWidths verifies BL output port width matches cols
func TestM6VER04_OutputPortWidths(t *testing.T) {
	testCases := []struct {
		name       string
		rows       int
		cols       int
		expectedBL string
	}{
		{"4×4 array", 4, 4, "inout  wire [3:0] BL"},
		{"2×8 array", 2, 8, "inout  wire [7:0] BL"},
		{"8×2 array", 8, 2, "inout  wire [1:0] BL"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			weights := make([][]float64, tc.rows)
			for i := range weights {
				weights[i] = make([]float64, tc.cols)
				for j := range weights[i] {
					weights[i][j] = float64(i*tc.cols+j) * 0.1
				}
			}

			config := compiler.DefaultConfig()
			config.ArrayRows = tc.rows * 2
			config.ArrayCols = tc.cols * 2
			design, err := compiler.Compile(weights, config)
			if err != nil {
				t.Fatalf("Compile failed: %v", err)
			}

			verilog := GenerateVerilogWithDefaults(design)

			// M6-VER-04g: Check BL port width (cols)
			if !strings.Contains(verilog, tc.expectedBL) {
				t.Errorf("M6-VER-04g FAIL: Expected '%s' for %d columns", tc.expectedBL, tc.cols)
				// Show what we got
				lines := strings.Split(verilog, "\n")
				for _, line := range lines {
					if strings.Contains(line, "BL") && strings.Contains(line, "inout") {
						t.Logf("  Found: %s", strings.TrimSpace(line))
					}
				}
			} else {
				t.Logf("M6-VER-04g PASS: BL port width = [%d:0] for %d columns", tc.cols-1, tc.cols)
			}
		})
	}
}

// TestM6VER04_PortConnections verifies port-to-instance connections
func TestM6VER04_PortConnections(t *testing.T) {
	// Create 3×3 array
	weights := [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
		{0.7, 0.8, 0.9},
	}

	config := compiler.DefaultConfig()
	config.ArrayRows = 6
	config.ArrayCols = 6
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	verilog := GenerateVerilogWithDefaults(design)

	// M6-VER-04h: Verify WL connections for each row
	for row := 0; row < 3; row++ {
		expectedConnection := fmt.Sprintf(".WL  (WL[%d])", row)
		count := strings.Count(verilog, expectedConnection)
		expectedCount := 3 // 3 columns per row

		if count != expectedCount {
			t.Errorf("M6-VER-04h FAIL: Expected %d connections of '%s', got %d",
				expectedCount, expectedConnection, count)
		} else {
			t.Logf("M6-VER-04h PASS: Row %d has %d WL[%d] connections", row, count, row)
		}
	}

	// M6-VER-04i: Verify BL connections for each column
	for col := 0; col < 3; col++ {
		expectedConnection := fmt.Sprintf(".BL  (BL[%d])", col)
		count := strings.Count(verilog, expectedConnection)
		expectedCount := 3 // 3 rows per column

		if count != expectedCount {
			t.Errorf("M6-VER-04i FAIL: Expected %d connections of '%s', got %d",
				expectedCount, expectedConnection, count)
		} else {
			t.Logf("M6-VER-04i PASS: Column %d has %d BL[%d] connections", col, count, col)
		}
	}

	t.Log("M6-VER-04 Summary: Port widths and connections validated")
}

// TestM6VER04_PowerPortsPresent verifies power/ground ports
func TestM6VER04_PowerPortsPresent(t *testing.T) {
	weights := [][]float64{
		{0.1, 0.2},
		{0.3, 0.4},
	}

	config := compiler.DefaultConfig()
	config.ArrayRows = 4
	config.ArrayCols = 4
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	verilog := GenerateVerilogWithDefaults(design)

	// M6-VER-04j: Check VPWR port
	if !strings.Contains(verilog, "inout  wire       VPWR") {
		t.Error("M6-VER-04j FAIL: Missing 'inout  wire       VPWR' port")
	} else {
		t.Log("M6-VER-04j PASS: VPWR power port present")
	}

	// M6-VER-04k: Check VGND port
	if !strings.Contains(verilog, "inout  wire       VGND") {
		t.Error("M6-VER-04k FAIL: Missing 'inout  wire       VGND' port")
	} else {
		t.Log("M6-VER-04k PASS: VGND ground port present")
	}

	// M6-VER-04l: Verify all instances connect to VPWR/VGND
	instanceCount := strings.Count(verilog, "fecim_bit #(")
	vpwrConnections := strings.Count(verilog, ".VPWR (VPWR)")
	vgndConnections := strings.Count(verilog, ".VGND (VGND)")

	if vpwrConnections != instanceCount {
		t.Errorf("M6-VER-04l FAIL: Expected %d VPWR connections, got %d", instanceCount, vpwrConnections)
	} else {
		t.Logf("M6-VER-04l PASS: All %d instances connect to VPWR", instanceCount)
	}

	if vgndConnections != instanceCount {
		t.Errorf("M6-VER-04l FAIL: Expected %d VGND connections, got %d", instanceCount, vgndConnections)
	} else {
		t.Logf("M6-VER-04l PASS: All %d instances connect to VGND", instanceCount)
	}

	t.Log("M6-VER-04 Summary: All port widths and power connections validated")
}
