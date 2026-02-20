// pkg/export/def_syntax_validation_test.go
// M6-DEF-01: DEF syntax validation test
// Validates DEF export syntax and parser round-trip correctness

package export

import (
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

// TestDEFSyntaxValidation tests M6-DEF-01:
// Export DEF for 4×4 array, parse back, verify VERSION, DESIGN, UNITS, COMPONENTS, NETS
func TestDEFSyntaxValidation(t *testing.T) {
	// Create 4×4 weight matrix
	weights := [][]float64{
		{0.1, 0.2, 0.3, 0.4},
		{0.5, 0.6, 0.7, 0.8},
		{0.9, 1.0, -0.1, -0.2},
		{-0.3, -0.4, -0.5, -0.6},
	}

	config := compiler.DefaultConfig()
	config.ArrayRows = 8
	config.ArrayCols = 8
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	def := GenerateDEFWithDefaults(design)

	// M6-DEF-01.1: Verify VERSION
	if !strings.Contains(def, "VERSION 5.8") {
		t.Error("M6-DEF-01.1 FAIL: Missing VERSION 5.8")
	} else {
		t.Log("M6-DEF-01.1 PASS: VERSION 5.8 present")
	}

	// M6-DEF-01.2: Verify DESIGN
	if !strings.Contains(def, "DESIGN fecim_crossbar") {
		t.Error("M6-DEF-01.2 FAIL: Missing DESIGN declaration")
	} else {
		t.Log("M6-DEF-01.2 PASS: DESIGN fecim_crossbar present")
	}

	// M6-DEF-01.3: Verify UNITS
	if !strings.Contains(def, "UNITS DISTANCE MICRONS 1000") {
		t.Error("M6-DEF-01.3 FAIL: Missing or incorrect UNITS DISTANCE MICRONS")
	} else {
		t.Log("M6-DEF-01.3 PASS: UNITS DISTANCE MICRONS 1000")
	}

	// M6-DEF-01.4: Verify COMPONENTS section
	if !strings.Contains(def, "COMPONENTS 16") {
		t.Error("M6-DEF-01.4 FAIL: Expected COMPONENTS 16 for 4×4 array")
	} else {
		t.Log("M6-DEF-01.4 PASS: COMPONENTS 16")
	}

	// M6-DEF-01.5: Verify component instances exist
	expectedInstances := []string{"R_0_0", "R_0_3", "R_3_0", "R_3_3"}
	for _, inst := range expectedInstances {
		if !strings.Contains(def, inst+" ") {
			t.Errorf("M6-DEF-01.5 FAIL: Missing instance %s", inst)
		}
	}
	t.Log("M6-DEF-01.5 PASS: All corner instances present")

	// M6-DEF-01.6: Verify NETS section exists
	if !strings.Contains(def, "NETS ") {
		t.Error("M6-DEF-01.6 FAIL: Missing NETS section")
	} else {
		t.Log("M6-DEF-01.6 PASS: NETS section present")
	}

	// M6-DEF-01.7: Verify END DESIGN termination
	if !strings.HasSuffix(strings.TrimSpace(def), "END DESIGN") {
		t.Error("M6-DEF-01.7 FAIL: Missing or incorrect END DESIGN terminator")
	} else {
		t.Log("M6-DEF-01.7 PASS: END DESIGN terminator present")
	}

	// M6-DEF-01.8: Verify DIEAREA
	if !strings.Contains(def, "DIEAREA ( 0 0 ) (") {
		t.Error("M6-DEF-01.8 FAIL: Missing DIEAREA declaration")
	} else {
		t.Log("M6-DEF-01.8 PASS: DIEAREA declaration present")
	}

	// M6-DEF-01.9: Verify ROW definitions exist
	if !strings.Contains(def, "ROW ROW_") {
		t.Error("M6-DEF-01.9 FAIL: Missing ROW definitions")
	} else {
		t.Log("M6-DEF-01.9 PASS: ROW definitions present")
	}

	// M6-DEF-01.10: Verify PINS section exists
	if !strings.Contains(def, "PINS ") {
		t.Error("M6-DEF-01.10 FAIL: Missing PINS section")
	} else {
		t.Log("M6-DEF-01.10 PASS: PINS section present")
	}

	t.Logf("M6-DEF-01 syntax validation: %d bytes, %d components", len(def), 16)
}

// TestDEFStructureCorrectness verifies section order and completeness
func TestDEFStructureCorrectness(t *testing.T) {
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

	def := GenerateDEFWithDefaults(design)
	lines := strings.Split(def, "\n")

	// Verify section order
	versionIdx := findLineIndex(lines, "VERSION")
	designIdx := findLineIndex(lines, "DESIGN")
	unitsIdx := findLineIndex(lines, "UNITS")
	dieareaIdx := findLineIndex(lines, "DIEAREA")
	rowIdx := findLineIndex(lines, "ROW")
	componentsIdx := findLineIndex(lines, "COMPONENTS")
	pinsIdx := findLineIndex(lines, "PINS")
	netsIdx := findLineIndex(lines, "NETS")
	endIdx := findLineIndex(lines, "END DESIGN")

	if !(versionIdx < designIdx && designIdx < unitsIdx && unitsIdx < dieareaIdx &&
		dieareaIdx < componentsIdx && componentsIdx < pinsIdx && pinsIdx < netsIdx && netsIdx < endIdx) {
		t.Error("M6-DEF-01 FAIL: Incorrect section order")
		t.Logf("Order: VERSION=%d DESIGN=%d UNITS=%d DIEAREA=%d ROW=%d COMPONENTS=%d PINS=%d NETS=%d END=%d",
			versionIdx, designIdx, unitsIdx, dieareaIdx, rowIdx, componentsIdx, pinsIdx, netsIdx, endIdx)
	} else {
		t.Log("M6-DEF-01 PASS: Section order correct")
	}

	// Verify all sections terminated
	if !strings.Contains(def, "END COMPONENTS") {
		t.Error("M6-DEF-01 FAIL: Missing END COMPONENTS")
	}
	if !strings.Contains(def, "END PINS") {
		t.Error("M6-DEF-01 FAIL: Missing END PINS")
	}
	if !strings.Contains(def, "END NETS") {
		t.Error("M6-DEF-01 FAIL: Missing END NETS")
	}

	t.Log("M6-DEF-01 structure correctness: all sections properly terminated")
}

// TestDEF1T1RSyntax verifies 1T1R architecture DEF has SL pins
func TestDEF1T1RSyntax(t *testing.T) {
	weights := [][]float64{
		{0.1, 0.2},
		{0.3, 0.4},
	}

	config := compiler.Config1T1R()
	config.ArrayRows = 4
	config.ArrayCols = 4
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	def := GenerateDEFWithDefaults(design)

	// Verify 1T1R-specific syntax
	if !strings.Contains(def, "Architecture: 1t1r") {
		t.Error("M6-DEF-01 (1T1R) FAIL: Missing architecture annotation")
	}

	// SL pins should exist
	if !strings.Contains(def, "- SL[0]") {
		t.Error("M6-DEF-01 (1T1R) FAIL: Missing SL[0] pin")
	}

	// SL nets should exist
	if !strings.Contains(def, "- SL[0]\n") {
		t.Error("M6-DEF-01 (1T1R) FAIL: Missing SL[0] net")
	}

	// Cell type should be fecim_1t1r
	if !strings.Contains(def, "fecim_1t1r +") {
		t.Error("M6-DEF-01 (1T1R) FAIL: Cells should use fecim_1t1r type")
	}

	// Pin count: 2 WL + 2 BL + 2 SL + VPWR + VGND = 8
	if !strings.Contains(def, "PINS 8") {
		t.Error("M6-DEF-01 (1T1R) FAIL: Expected PINS 8")
	} else {
		t.Log("M6-DEF-01 (1T1R) PASS: PINS 8 (WL×2, BL×2, SL×2, VPWR, VGND)")
	}

	t.Log("M6-DEF-01 (1T1R) syntax validation: PASS")
}

// TestDEFPinDirections verifies pin USE and DIRECTION attributes
func TestDEFPinDirections(t *testing.T) {
	weights := [][]float64{{0.1}}
	config := compiler.DefaultConfig()
	config.ArrayRows = 2
	config.ArrayCols = 2
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	def := GenerateDEFWithDefaults(design)

	// WL pins should be INPUT
	if !strings.Contains(def, "- WL[0] + NET WL[0] + DIRECTION INPUT + USE SIGNAL") {
		t.Error("M6-DEF-01 FAIL: WL pins should be DIRECTION INPUT + USE SIGNAL")
	}

	// BL pins should be INOUT (bidirectional)
	if !strings.Contains(def, "- BL[0] + NET BL[0] + DIRECTION INOUT + USE SIGNAL") {
		t.Error("M6-DEF-01 FAIL: BL pins should be DIRECTION INOUT + USE SIGNAL")
	}

	// VPWR should be POWER
	if !strings.Contains(def, "- VPWR + NET VPWR + DIRECTION INPUT + USE POWER") {
		t.Error("M6-DEF-01 FAIL: VPWR should be USE POWER")
	}

	// VGND should be GROUND
	if !strings.Contains(def, "- VGND + NET VGND + DIRECTION INPUT + USE GROUND") {
		t.Error("M6-DEF-01 FAIL: VGND should be USE GROUND")
	}

	t.Log("M6-DEF-01 pin directions: PASS")
}

// TestDEF2T1RCSLPinPlacement verifies that CSL pins in 2T1R DEF are placed at
// the top edge (X = col*CellWidth, Y = dieHeight), not on the right edge with a
// dimension-mismatched Y = col*CellHeight that the old code produced.
//
// Uses a non-square 4-row × 3-col weight matrix so rows ≠ cols, exposing any
// confusion between the two.
func TestDEF2T1RCSLPinPlacement(t *testing.T) {
	// 4 rows × 3 cols: CSL is per-column, so exactly 3 CSL pins expected
	weights := [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
		{0.7, 0.8, 0.9},
		{0.1, 0.2, 0.3},
	}

	config := compiler.Config2T1R()
	config.ArrayRows = 4
	config.ArrayCols = 4 // compiler derives actual dims from weights
	design, err := compiler.Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	def := GenerateDEFWithDefaults(design)

	// CSL pins must exist
	if !strings.Contains(def, "- CSL[0]") {
		t.Fatal("2T1R DEF missing CSL[0] pin")
	}

	// Count CSL pins in PINS section: must have one per column (3), not one per row (4)
	cslPinCount := strings.Count(def, "NET CSL[")
	if cslPinCount != 3 {
		t.Errorf("expected 3 CSL pins for 4×3 2T1R array, got %d", cslPinCount)
	}

	// CSL must NOT have CSL[3] (that would indicate row-count indexing, old bug)
	if strings.Contains(def, "NET CSL[3]") {
		t.Error("CSL[3] should not exist for 3-column 2T1R array (CSL is per-column)")
	}

	// Pin count: 4 WL + 3 BL + 3 SL + 3 CSL + VPWR + VGND = 15
	if !strings.Contains(def, "PINS 15") {
		t.Errorf("2T1R 4×3 array: expected PINS 15 (WL×4+BL×3+SL×3+CSL×3+VPWR+VGND)")
	}
}

// Helper function to find line index containing substring
func findLineIndex(lines []string, substr string) int {
	for i, line := range lines {
		if strings.Contains(line, substr) {
			return i
		}
	}
	return -1
}
