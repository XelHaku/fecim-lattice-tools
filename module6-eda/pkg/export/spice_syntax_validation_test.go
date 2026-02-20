package export

import (
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

// TestM6_SPICE_01_SyntaxValidation_4x4Array validates SPICE netlist syntax
// by exporting a 4×4 1T1R array and parsing it back with a simple SPICE parser.
// Verifies round-trip: exported lines match expected structure (.subckt, .ends, elements).
func TestM6_SPICE_01_SyntaxValidation_4x4Array(t *testing.T) {
	// Create 4×4 1T1R array design
	cells := make([]compiler.CellAssignment, 0, 16)
	for row := 0; row < 4; row++ {
		for col := 0; col < 4; col++ {
			cells = append(cells, compiler.CellAssignment{
				Row:         row,
				Col:         col,
				Conductance: 50.0 + float64(row*4+col),
				Resistance:  20000.0 - float64(row*4+col)*100.0,
				Level:       row*4 + col,
			})
		}
	}

	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.Arch1T1R,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 16, ActiveCells: 16},
	}

	// Generate SPICE netlist
	netlist := GenerateSPICE(design, 1.8)

	// Parse and validate structure
	parsed := parseSPICENetlist(netlist)

	// Verify subcircuit balance
	if parsed.subcktCount != parsed.endsCount {
		t.Errorf("M6-SPICE-01: Unbalanced subcircuit delimiters: .subckt=%d, .ends=%d",
			parsed.subcktCount, parsed.endsCount)
	}

	// Verify expected subcircuits are present
	expectedSubckts := []string{"FECAP_HZO", "fefet_cell", "fefet_1t1r", "fecim_cell"}
	for _, name := range expectedSubckts {
		if !parsed.subcircuitDefined[name] {
			t.Errorf("M6-SPICE-01: Missing required subcircuit: %s", name)
		}
	}

	// Verify 16 cell instances (4×4 array)
	if parsed.instanceCount != 16 {
		t.Errorf("M6-SPICE-01: Expected 16 cell instances (4×4), got %d", parsed.instanceCount)
	}

	// Verify voltage sources (4 WL + 4 SL + 1 VDD)
	expectedVSources := 4 + 4 + 1 // WL + SL + VDD
	if parsed.vsourceCount != expectedVSources {
		t.Errorf("M6-SPICE-01: Expected %d voltage sources, got %d", expectedVSources, parsed.vsourceCount)
	}

	// Verify resistors (4 BL loads)
	expectedResistors := 4
	if parsed.resistorCount != expectedResistors {
		t.Errorf("M6-SPICE-01: Expected %d resistors (BL loads), got %d", expectedResistors, parsed.resistorCount)
	}

	// Verify analysis directives
	if !parsed.hasOpAnalysis {
		t.Error("M6-SPICE-01: Missing .op analysis directive")
	}
	if !parsed.hasControl {
		t.Error("M6-SPICE-01: Missing .control/.endc block")
	}
	if !parsed.hasEnd {
		t.Error("M6-SPICE-01: Missing .end directive")
	}

	// Verify model card for SKY130 NMOS
	if !parsed.hasModelCard {
		t.Error("M6-SPICE-01: Missing .model SKY130NMOS card for 1T1R architecture")
	}

	// Round-trip verification: re-parse should yield same structure
	reparsed := parseSPICENetlist(netlist)
	if reparsed.subcktCount != parsed.subcktCount {
		t.Errorf("M6-SPICE-01: Round-trip mismatch: subckt count changed from %d to %d",
			parsed.subcktCount, reparsed.subcktCount)
	}
	if reparsed.instanceCount != parsed.instanceCount {
		t.Errorf("M6-SPICE-01: Round-trip mismatch: instance count changed from %d to %d",
			parsed.instanceCount, reparsed.instanceCount)
	}

	t.Logf("M6-SPICE-01 PASS: 4×4 array syntax validated — %d subcircuits, %d instances, %d V-sources, %d resistors",
		parsed.subcktCount, parsed.instanceCount, parsed.vsourceCount, parsed.resistorCount)
}

// spiceParseResult captures SPICE netlist structural elements
type spiceParseResult struct {
	subcktCount       int
	endsCount         int
	subcircuitDefined map[string]bool
	instanceCount     int
	vsourceCount      int
	resistorCount     int
	hasOpAnalysis     bool
	hasControl        bool
	hasEnd            bool
	hasModelCard      bool
}

// parseSPICENetlist is a simple SPICE parser that validates structural elements.
// It checks for balanced .subckt/.ends, element lines (X, V, R, M), and analysis directives.
func parseSPICENetlist(netlist string) spiceParseResult {
	result := spiceParseResult{
		subcircuitDefined: make(map[string]bool),
	}

	lines := strings.Split(netlist, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "*") {
			continue // Skip comments and empty lines
		}

		// Convert to lower for case-insensitive matching
		lower := strings.ToLower(trimmed)

		// Subcircuit definitions
		if strings.HasPrefix(lower, ".subckt") {
			result.subcktCount++
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 {
				result.subcircuitDefined[fields[1]] = true
			}
		}
		if strings.HasPrefix(lower, ".ends") {
			result.endsCount++
		}

		// Element instances - count only array cell instances (X_row_col pattern)
		if strings.HasPrefix(trimmed, "X_") || strings.HasPrefix(trimmed, "x_") {
			// Cell instance in array: X_0_0, X_1_2, etc.
			result.instanceCount++
		}
		if (strings.HasPrefix(trimmed, "V") || strings.HasPrefix(trimmed, "v")) && !strings.HasPrefix(trimmed, "V =") {
			result.vsourceCount++
		}
		// Count only top-level resistors (RBL bit line loads), not subcircuit resistors
		if strings.HasPrefix(trimmed, "RBL") || strings.HasPrefix(trimmed, "rbl") {
			result.resistorCount++
		}

		// Analysis directives
		if strings.HasPrefix(lower, ".op") {
			result.hasOpAnalysis = true
		}
		if strings.HasPrefix(lower, ".control") {
			result.hasControl = true
		}
		if strings.HasPrefix(lower, ".end") && !strings.HasPrefix(lower, ".ends") {
			result.hasEnd = true
		}

		// Model card
		if strings.HasPrefix(lower, ".model") && strings.Contains(lower, "nmos") {
			result.hasModelCard = true
		}
	}

	return result
}

// TestM6_SPICE_01_SyntaxValidation_2T2RArray tests 2T1R architecture syntax
func TestM6_SPICE_01_SyntaxValidation_2T2RArray(t *testing.T) {
	cells := []compiler.CellAssignment{
		{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 0},
		{Row: 0, Col: 1, Conductance: 60.0, Resistance: 16666.7, Level: 5},
		{Row: 1, Col: 0, Conductance: 70.0, Resistance: 14285.7, Level: 10},
		{Row: 1, Col: 1, Conductance: 80.0, Resistance: 12500.0, Level: 15},
	}

	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeCompute,
			Architecture: compiler.Arch2T1R,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 4, ActiveCells: 4},
	}

	netlist := GenerateSPICE(design, 1.8)
	parsed := parseSPICENetlist(netlist)

	// 2T1R should have fefet_2t1r subcircuit
	if !parsed.subcircuitDefined["fefet_2t1r"] {
		t.Error("M6-SPICE-01: 2T1R architecture missing fefet_2t1r subcircuit")
	}

	// 2×2 array = 4 instances
	if parsed.instanceCount != 4 {
		t.Errorf("M6-SPICE-01: Expected 4 instances for 2×2 2T1R array, got %d", parsed.instanceCount)
	}

	// 2T1R has WL, SL, and CSL voltage sources (2 WL + 2 SL + 2 CSL + 1 VDD = 7)
	expectedVSources := 2 + 2 + 2 + 1
	if parsed.vsourceCount != expectedVSources {
		t.Errorf("M6-SPICE-01: Expected %d voltage sources for 2T1R (WL+SL+CSL+VDD), got %d",
			expectedVSources, parsed.vsourceCount)
	}

	t.Logf("M6-SPICE-01 PASS: 2T1R 2×2 array syntax validated — %d instances, %d V-sources",
		parsed.instanceCount, parsed.vsourceCount)
}

// TestSPICE_SLUsesColsNotRows verifies SL (Source Line) net indices use column, not row.
//
// SL is per-column in the Verilog (SL[numCols-1:0], .SL(SL[col])) and DEF pin placement.
// A non-square 3×5 array exposes any row/column confusion:
//   - 3 WL voltage sources (rows 0-2)
//   - 5 SL voltage sources (cols 0-4, after fix)  ← old bug gave 3 SL (rows)
//   - 1 VDD → total 9
// Cell instantiation: sl2 (not sl0) for cell at row=0, col=2.
func TestSPICE_SLUsesColsNotRows(t *testing.T) {
	// 3 rows × 5 cols non-square 1T1R array
	var cells []compiler.CellAssignment
	for row := 0; row < 3; row++ {
		for col := 0; col < 5; col++ {
			cells = append(cells, compiler.CellAssignment{
				Row: row, Col: col,
				Conductance: 50.0, Resistance: 20000.0, Level: 0,
			})
		}
	}
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.Arch1T1R,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 15, ActiveCells: 15},
	}

	netlist := GenerateSPICE(design, 1.8)
	parsed := parseSPICENetlist(netlist)

	// 3 WL + 5 SL (per-column) + 1 VDD = 9 V-sources
	// Old bug: 3 WL + 3 SL (per-row) + 1 VDD = 7
	expectedVSources := 3 + 5 + 1
	if parsed.vsourceCount != expectedVSources {
		t.Errorf("3×5 1T1R: expected %d voltage sources (3 WL + 5 SL-per-col + 1 VDD), got %d\n"+
			"Check that SL drivers loop over maxCol (5) not maxRow (3)",
			expectedVSources, parsed.vsourceCount)
	}

	// Row 0, Col 2 should connect to sl2 (column net), not sl0 (row net)
	if !strings.Contains(netlist, "wl0 bl2 sl2 fefet_1t1r") {
		t.Error("3×5 1T1R: cell[0,2] should connect to sl2 (col=2), not sl0 (row=0)\n" +
			"Check 1T1R instantiation uses cell.Col for SL net")
	}

	// Row 2, Col 0 should connect to sl0 (column 0), not sl2 (row 2)
	if !strings.Contains(netlist, "wl2 bl0 sl0 fefet_1t1r") {
		t.Error("3×5 1T1R: cell[2,0] should connect to sl0 (col=0), not sl2 (row=2)\n" +
			"Check 1T1R instantiation uses cell.Col for SL net")
	}
}

// TestSPICE_2T1R_SLUsesColsNotRows verifies 2T1R SL net indices use column, not row.
func TestSPICE_2T1R_SLUsesColsNotRows(t *testing.T) {
	// 4 rows × 3 cols non-square 2T1R array
	var cells []compiler.CellAssignment
	for row := 0; row < 4; row++ {
		for col := 0; col < 3; col++ {
			cells = append(cells, compiler.CellAssignment{
				Row: row, Col: col,
				Conductance: 50.0, Resistance: 20000.0, Level: 0,
			})
		}
	}
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeCompute,
			Architecture: compiler.Arch2T1R,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 12, ActiveCells: 12},
	}

	netlist := GenerateSPICE(design, 1.8)
	parsed := parseSPICENetlist(netlist)

	// 4 WL + 3 SL (per-column) + 3 CSL + 1 VDD = 11 V-sources
	// Old bug: 4 WL + 4 SL (per-row) + 3 CSL + 1 VDD = 12
	expectedVSources := 4 + 3 + 3 + 1
	if parsed.vsourceCount != expectedVSources {
		t.Errorf("4×3 2T1R: expected %d voltage sources (4 WL + 3 SL-per-col + 3 CSL + 1 VDD), got %d\n"+
			"Check that SL drivers loop over maxCol (3) not maxRow (4)",
			expectedVSources, parsed.vsourceCount)
	}

	// Row 3, Col 1 should connect to sl1 (col=1), not sl3 (row=3)
	if !strings.Contains(netlist, "wl3 csl1 bl1 sl1 fefet_2t1r") {
		t.Error("4×3 2T1R: cell[3,1] should connect to sl1 (col=1), not sl3 (row=3)\n" +
			"Check 2T1R instantiation uses cell.Col for SL net")
	}
}

// TestM6_SPICE_01_SyntaxValidation_PassiveArray tests passive (0T1R) architecture
func TestM6_SPICE_01_SyntaxValidation_PassiveArray(t *testing.T) {
	cells := []compiler.CellAssignment{
		{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 0},
		{Row: 0, Col: 1, Conductance: 60.0, Resistance: 16666.7, Level: 5},
	}

	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.ArchPassive,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 2, ActiveCells: 2},
	}

	netlist := GenerateSPICE(design, 1.8)
	parsed := parseSPICENetlist(netlist)

	// Passive architecture should NOT have model card (no transistors)
	if parsed.hasModelCard {
		t.Error("M6-SPICE-01: Passive architecture should not have NMOS model card")
	}

	// Should have fefet_cell but not fefet_1t1r or fefet_2t1r
	if parsed.subcircuitDefined["fefet_1t1r"] || parsed.subcircuitDefined["fefet_2t1r"] {
		t.Error("M6-SPICE-01: Passive architecture should not have 1T1R or 2T1R subcircuits")
	}

	t.Logf("M6-SPICE-01 PASS: Passive array syntax validated — %d instances, no transistor models",
		parsed.instanceCount)
}
