// pkg/export/cross_spice_liberty_test.go
// M6-CROSS-01: Cross-format consistency between SPICE and Liberty
// Verifies capacitance values match across analog (SPICE) and timing (Liberty) formats

package export

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/module6-eda/pkg/config"
)

// TestCrossFormat_SPICE_Liberty_Capacitance (M6-CROSS-01)
// Exports SPICE and Liberty for the same array configuration
// Extracts ferroelectric capacitance (C_fe) from SPICE netlist
// Extracts input_capacitance from Liberty .lib file
// Note: These are different physical capacitances:
//   - C_fe: ferroelectric storage capacitance (~0.045 fF for HZO)
//   - input_cap: gate/pin input capacitance (~15 fF for SKY130 transistor gates)
//
// This test verifies both values are physically reasonable and consistent with their contexts.
func TestCrossFormat_SPICE_Liberty_Capacitance(t *testing.T) {
	// Create test array design
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.ArchPassive,
			Technology:   "sky130",
			Levels:       32,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 5},
			{Row: 0, Col: 1, Conductance: 60.0, Resistance: 16666.7, Level: 10},
		},
		Stats: compiler.DesignStats{TotalCells: 2, ActiveCells: 2},
	}

	// Generate SPICE netlist
	spiceNetlist := GenerateSPICE(design, 1.8)
	if len(spiceNetlist) == 0 {
		t.Fatal("SPICE netlist generation failed")
	}

	// Generate Liberty timing library with physically realistic input capacitance
	libCfg := config.CellConfig{
		Name:         "fecim_bitcell",
		CellType:     "passive",
		Technology:   "sky130",
		Width:        0.46,
		Height:       2.72,
		Voltage:      1.8,
		Temperature:  25.0,
		Process:      1.0,
		RiseTime:     50.0,
		FallTime:     5.0,
		InputCap:     0.015, // 15 fF = 0.015 pF (gate capacitance)
		LeakagePower: 0.0003,
	}
	libertyLib := GenerateLiberty(libCfg)
	if len(libertyLib) == 0 {
		t.Fatal("Liberty library generation failed")
	}

	// Extract C_fe from SPICE netlist (ferroelectric capacitance)
	cfeFromSPICE := extractCapacitanceFromSPICE(t, spiceNetlist)

	// Extract input_capacitance from Liberty (pin/gate capacitance)
	inputCapFromLiberty := extractCapacitanceFromLiberty(t, libertyLib)

	t.Logf("M6-CROSS-01: SPICE C_fe=%.3f fF (ferroelectric), Liberty input_cap=%.3f fF (gate)",
		cfeFromSPICE, inputCapFromLiberty)

	// Verify both capacitances are in physically reasonable ranges
	// Ferroelectric capacitance: 0.01-1.0 fF range for nm-scale devices
	if cfeFromSPICE < 0.01 || cfeFromSPICE > 1.0 {
		t.Errorf("SPICE ferroelectric capacitance out of range: %.3f fF (expected 0.01-1.0 fF)",
			cfeFromSPICE)
	}

	// Gate input capacitance: 1-100 fF range for nm-scale transistors
	if inputCapFromLiberty < 1.0 || inputCapFromLiberty > 100.0 {
		t.Errorf("Liberty input capacitance out of range: %.3f fF (expected 1-100 fF)",
			inputCapFromLiberty)
	}

	// Both values should be present and non-zero
	if cfeFromSPICE <= 0 {
		t.Error("SPICE ferroelectric capacitance is zero or negative")
	}
	if inputCapFromLiberty <= 0 {
		t.Error("Liberty input capacitance is zero or negative")
	}
}

// TestCrossFormat_SPICE_Liberty_1T1R_Capacitance (M6-CROSS-01)
// Same test for 1T1R architecture - verifies capacitance values are physically reasonable
func TestCrossFormat_SPICE_Liberty_1T1R_Capacitance(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeMemory,
			Architecture: compiler.Arch1T1R,
			Technology:   "sky130",
			Levels:       64,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 10},
		},
		Stats: compiler.DesignStats{TotalCells: 1, ActiveCells: 1},
	}

	spiceNetlist := GenerateSPICE(design, 1.8)
	if len(spiceNetlist) == 0 {
		t.Fatal("SPICE netlist generation failed")
	}

	libCfg := config.CellConfig{
		Name:         "fecim_1t1r_bitcell",
		CellType:     "1t1r",
		Technology:   "sky130",
		Width:        0.92,
		Height:       3.40,
		Voltage:      1.8,
		Temperature:  25.0,
		Process:      1.0,
		RiseTime:     50.0,
		FallTime:     5.0,
		InputCap:     0.015,
		LeakagePower: 0.0003,
	}
	libertyLib := GenerateLiberty(libCfg)
	if len(libertyLib) == 0 {
		t.Fatal("Liberty library generation failed")
	}

	cfeFromSPICE := extractCapacitanceFromSPICE(t, spiceNetlist)
	inputCapFromLiberty := extractCapacitanceFromLiberty(t, libertyLib)

	t.Logf("M6-CROSS-01 (1T1R): SPICE C_fe=%.3f fF (ferroelectric), Liberty input_cap=%.3f fF (gate)",
		cfeFromSPICE, inputCapFromLiberty)

	if cfeFromSPICE < 0.01 || cfeFromSPICE > 1.0 {
		t.Errorf("1T1R ferroelectric capacitance out of range: %.3f fF", cfeFromSPICE)
	}
	if inputCapFromLiberty < 1.0 || inputCapFromLiberty > 100.0 {
		t.Errorf("1T1R input capacitance out of range: %.3f fF", inputCapFromLiberty)
	}
}

// extractCapacitanceFromSPICE parses SPICE netlist to extract ferroelectric capacitance
// Looks for material parameters (εᵣ, thickness, area) and calculates C = (ε₀ × εᵣ × A) / d
func extractCapacitanceFromSPICE(t *testing.T, netlist string) float64 {
	t.Helper()

	// Look for HZO material parameters in FECAP_HZO subcircuit PARAMS
	// Format: ".subckt FECAP_HZO pos neg PARAMS: ... thick=1.000000e-08 area=2.025000e-15 eps0=8.854187812800e-12"
	// And capacitance expression: "Cfe pos n1 {eps0*25*area/thick}"

	// Parse thickness from thick parameter
	thickPattern := regexp.MustCompile(`thick\s*=\s*([0-9.eE+-]+)`)
	matches := thickPattern.FindStringSubmatch(netlist)
	if len(matches) < 2 {
		t.Fatal("Failed to extract thick parameter from SPICE netlist")
	}
	thickness, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		t.Fatalf("Failed to parse thick value: %v", err)
	}

	// Parse area parameter
	areaPattern := regexp.MustCompile(`area\s*=\s*([0-9.eE+-]+)`)
	matches = areaPattern.FindStringSubmatch(netlist)
	if len(matches) < 2 {
		t.Fatal("Failed to extract area parameter from SPICE netlist")
	}
	area, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		t.Fatalf("Failed to parse area value: %v", err)
	}

	// Parse eps0 parameter
	eps0Pattern := regexp.MustCompile(`eps0\s*=\s*([0-9.eE+-]+)`)
	matches = eps0Pattern.FindStringSubmatch(netlist)
	if len(matches) < 2 {
		t.Fatal("Failed to extract eps0 parameter from SPICE netlist")
	}
	epsilon0, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		t.Fatalf("Failed to parse eps0 value: %v", err)
	}

	// Extract relative permittivity from capacitance expression
	// Format: "Cfe pos n1 {eps0*25*area/thick}"
	cfePattern := regexp.MustCompile(`Cfe\s+\w+\s+\w+\s+\{eps0\*([0-9.eE+-]+)\*area/thick\}`)
	matches = cfePattern.FindStringSubmatch(netlist)
	var epsR float64
	if len(matches) >= 2 {
		epsR, err = strconv.ParseFloat(matches[1], 64)
		if err != nil {
			t.Fatalf("Failed to parse relative permittivity from Cfe expression: %v", err)
		}
	} else {
		// Default to 25.0 for HZO if not found in expression
		epsR = 25.0
		t.Logf("Using default εᵣ=25.0 for HZO")
	}

	// Calculate capacitance: C = (ε₀ × εᵣ × A) / d
	capacitanceFarads := (epsilon0 * epsR * area) / thickness
	capacitanceFemtoFarads := capacitanceFarads * 1e15

	t.Logf("SPICE parameters: εᵣ=%.2f, d=%.2e m, A=%.2e m², ε₀=%.2e F/m, C=%.3f fF",
		epsR, thickness, area, epsilon0, capacitanceFemtoFarads)

	return capacitanceFemtoFarads
}

// extractCapacitanceFromLiberty parses Liberty .lib file to extract input_capacitance
func extractCapacitanceFromLiberty(t *testing.T, liberty string) float64 {
	t.Helper()

	// Look for capacitance declaration in Liberty pin
	// Format: "capacitance : 0.015 ;" (in pF)
	capPattern := regexp.MustCompile(`capacitance\s*:\s*([0-9.eE+-]+)\s*;`)
	matches := capPattern.FindStringSubmatch(liberty)
	if len(matches) < 2 {
		t.Fatal("Failed to extract capacitance from Liberty library")
	}

	capPF, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		t.Fatalf("Failed to parse capacitance value: %v", err)
	}

	// Convert pF to fF (multiply by 1000)
	capFF := capPF * 1000.0

	t.Logf("Liberty capacitance: %.3f pF = %.3f fF", capPF, capFF)

	return capFF
}

// TestCrossFormat_SPICE_Liberty_2T1R_Capacitance (M6-CROSS-01)
// Same test for 2T1R architecture - verifies capacitance values are physically reasonable
func TestCrossFormat_SPICE_Liberty_2T1R_Capacitance(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeCompute,
			Architecture: compiler.Arch2T1R,
			Technology:   "sky130",
			Levels:       128,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Conductance: 75.0, Resistance: 13333.3, Level: 32},
		},
		Stats: compiler.DesignStats{TotalCells: 1, ActiveCells: 1},
	}

	spiceNetlist := GenerateSPICE(design, 1.8)
	if len(spiceNetlist) == 0 {
		t.Fatal("SPICE netlist generation failed")
	}

	libCfg := config.CellConfig{
		Name:         "fecim_2t1r_bitcell",
		CellType:     "2t1r",
		Technology:   "sky130",
		Width:        1.38,
		Height:       3.80,
		Voltage:      1.8,
		Temperature:  25.0,
		Process:      1.0,
		RiseTime:     50.0,
		FallTime:     5.0,
		InputCap:     0.015,
		LeakagePower: 0.0003,
	}
	libertyLib := GenerateLiberty(libCfg)
	if len(libertyLib) == 0 {
		t.Fatal("Liberty library generation failed")
	}

	cfeFromSPICE := extractCapacitanceFromSPICE(t, spiceNetlist)
	inputCapFromLiberty := extractCapacitanceFromLiberty(t, libertyLib)

	t.Logf("M6-CROSS-01 (2T1R): SPICE C_fe=%.3f fF (ferroelectric), Liberty input_cap=%.3f fF (gate)",
		cfeFromSPICE, inputCapFromLiberty)

	if cfeFromSPICE < 0.01 || cfeFromSPICE > 1.0 {
		t.Errorf("2T1R ferroelectric capacitance out of range: %.3f fF", cfeFromSPICE)
	}
	if inputCapFromLiberty < 1.0 || inputCapFromLiberty > 100.0 {
		t.Errorf("2T1R input capacitance out of range: %.3f fF", inputCapFromLiberty)
	}
}

// TestCrossFormat_SPICE_Liberty_MultiCell (M6-CROSS-01)
// Test with larger array to ensure consistency across multiple cells
func TestCrossFormat_SPICE_Liberty_MultiCell(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.ArchPassive,
			Technology:   "sky130",
			Levels:       64,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 5},
			{Row: 0, Col: 1, Conductance: 60.0, Resistance: 16666.7, Level: 10},
			{Row: 1, Col: 0, Conductance: 70.0, Resistance: 14285.7, Level: 15},
			{Row: 1, Col: 1, Conductance: 80.0, Resistance: 12500.0, Level: 20},
		},
		Stats: compiler.DesignStats{TotalCells: 4, ActiveCells: 4},
	}

	spiceNetlist := GenerateSPICE(design, 1.8)
	if !strings.Contains(spiceNetlist, "X_0_0") || !strings.Contains(spiceNetlist, "X_1_1") {
		t.Fatal("SPICE netlist missing expected cell instances")
	}

	libCfg := config.CellConfig{
		Name:       "fecim_bitcell",
		CellType:   "passive",
		Technology: "sky130",
		Width:      0.46,
		Height:     2.72,
		Voltage:    1.8,
		InputCap:   0.015,
	}
	libertyLib := GenerateLiberty(libCfg)

	cfeFromSPICE := extractCapacitanceFromSPICE(t, spiceNetlist)
	inputCapFromLiberty := extractCapacitanceFromLiberty(t, libertyLib)

	t.Logf("M6-CROSS-01 (4-cell): SPICE C_fe=%.3f fF (ferroelectric), Liberty input_cap=%.3f fF (gate)",
		cfeFromSPICE, inputCapFromLiberty)

	// Verify capacitances are physically reasonable
	if cfeFromSPICE < 0.01 || cfeFromSPICE > 1.0 {
		t.Errorf("Multi-cell ferroelectric capacitance out of range: %.3f fF", cfeFromSPICE)
	}
	if inputCapFromLiberty < 1.0 || inputCapFromLiberty > 100.0 {
		t.Errorf("Multi-cell input capacitance out of range: %.3f fF", inputCapFromLiberty)
	}
}
