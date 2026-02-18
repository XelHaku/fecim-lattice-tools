package export

import (
	"strings"
	"testing"

	"fecim-lattice-tools/shared/physics"
)

// TestM6_SPICE_02_Subcircuit_1T1R_Topology validates 1T1R subcircuit structure:
// - M1 (FET) with correct model reference (SKY130NMOS)
// - C1 (FeCap) subcircuit instance
// - Port order: WL, BL, SL
// - FET params: W, L from shared/physics SKY130NMOS preset
func TestM6_SPICE_02_Subcircuit_1T1R_Topology(t *testing.T) {
	subckt := Generate1T1RSubcircuit()

	// Verify subcircuit header with correct port order
	if !strings.Contains(subckt, ".subckt fefet_1t1r WL BL SL") {
		t.Fatal("M6-SPICE-02: 1T1R subcircuit missing or has incorrect port order (expected: WL BL SL)")
	}

	// Verify MOSFET selector (MSEL) with SKY130NMOS model
	if !strings.Contains(subckt, "MSEL") {
		t.Error("M6-SPICE-02: 1T1R subcircuit missing MSEL (selector transistor)")
	}
	if !strings.Contains(subckt, "SKY130NMOS") {
		t.Error("M6-SPICE-02: 1T1R subcircuit missing SKY130NMOS model reference")
	}

	// Verify FeFET instance (XFE)
	if !strings.Contains(subckt, "XFE") {
		t.Error("M6-SPICE-02: 1T1R subcircuit missing XFE (ferroelectric capacitor instance)")
	}
	if !strings.Contains(subckt, "fefet_cell") {
		t.Error("M6-SPICE-02: 1T1R subcircuit missing fefet_cell subcircuit reference")
	}

	// Verify FET parameters match shared/physics preset
	sel := physics.SKY130NMOS()
	// SPICE uses format: W=4.200000000000e-07 L=1.500000000000e-07
	if !strings.Contains(subckt, "W=") || !strings.Contains(subckt, "L=") {
		t.Error("M6-SPICE-02: 1T1R MSEL missing W/L parameters")
	}

	// Verify values are present (allow for formatting variations)
	if !strings.Contains(subckt, "4.2") && !strings.Contains(subckt, "4.20000") {
		t.Errorf("M6-SPICE-02: 1T1R MSEL width value incorrect (expected W≈4.2e-07 from SKY130NMOS preset W=%.12e)",
			sel.W)
	}
	if !strings.Contains(subckt, "1.5") && !strings.Contains(subckt, "1.50000") {
		t.Errorf("M6-SPICE-02: 1T1R MSEL length value incorrect (expected L≈1.5e-07 from SKY130NMOS preset L=%.12e)",
			sel.L)
	}

	// Verify parameterized resistance
	if !strings.Contains(subckt, "PARAMS: R_level=1e4") {
		t.Error("M6-SPICE-02: 1T1R subcircuit missing R_level parameter")
	}

	// Verify connections: MSEL drain (n1) connects to XFE top electrode
	if !strings.Contains(subckt, "MSEL n1 WL SL SL") {
		t.Error("M6-SPICE-02: 1T1R MSEL connections incorrect (expected: drain=n1, gate=WL, source=SL, bulk=SL)")
	}
	if !strings.Contains(subckt, "XFE n1 BL") {
		t.Error("M6-SPICE-02: 1T1R XFE connections incorrect (expected: top=n1, bottom=BL)")
	}

	t.Logf("M6-SPICE-02 PASS: 1T1R subcircuit topology validated — ports: WL BL SL, W=%.3e m, L=%.3e m",
		sel.W, sel.L)
}

// TestM6_SPICE_02_Subcircuit_2T1R_Topology validates 2T1R subcircuit structure:
// - MROW and MCOL (dual FETs) with correct model
// - Port order: WL, CSL, BL, SL
// - Series connection: MROW → MCOL → XFE
func TestM6_SPICE_02_Subcircuit_2T1R_Topology(t *testing.T) {
	subckt := Generate2T1RSubcircuit()

	// Verify subcircuit header with correct port order
	if !strings.Contains(subckt, ".subckt fefet_2t1r WL CSL BL SL") {
		t.Fatal("M6-SPICE-02: 2T1R subcircuit missing or has incorrect port order (expected: WL CSL BL SL)")
	}

	// Verify dual MOSFET selectors
	if !strings.Contains(subckt, "MROW") {
		t.Error("M6-SPICE-02: 2T1R subcircuit missing MROW (row selector)")
	}
	if !strings.Contains(subckt, "MCOL") {
		t.Error("M6-SPICE-02: 2T1R subcircuit missing MCOL (column selector)")
	}
	if strings.Count(subckt, "SKY130NMOS") != 2 {
		t.Errorf("M6-SPICE-02: 2T1R subcircuit should have 2 SKY130NMOS references, got %d",
			strings.Count(subckt, "SKY130NMOS"))
	}

	// Verify FET parameters
	sel := physics.SKY130NMOS()

	// Count W= and L= occurrences (should be 2 each for dual transistors)
	wCount := strings.Count(subckt, "W=")
	lCount := strings.Count(subckt, "L=")
	if wCount != 2 {
		t.Errorf("M6-SPICE-02: 2T1R should have 2 FETs with W= parameters, got %d", wCount)
	}
	if lCount != 2 {
		t.Errorf("M6-SPICE-02: 2T1R should have 2 FETs with L= parameters, got %d", lCount)
	}

	// Verify values are approximately correct
	if !strings.Contains(subckt, "4.2") {
		t.Errorf("M6-SPICE-02: 2T1R FET width value incorrect (expected W≈%.3e)", sel.W)
	}
	if !strings.Contains(subckt, "1.5") {
		t.Errorf("M6-SPICE-02: 2T1R FET length value incorrect (expected L≈%.3e)", sel.L)
	}

	// Verify series connection topology: MROW(n1,WL,n2,n2) -> MCOL(n2,CSL,SL,SL) -> XFE(n1,BL)
	if !strings.Contains(subckt, "MROW n1 WL n2 n2") {
		t.Error("M6-SPICE-02: 2T1R MROW connections incorrect (expected: drain=n1, gate=WL, source=n2, bulk=n2)")
	}
	if !strings.Contains(subckt, "MCOL n2 CSL SL SL") {
		t.Error("M6-SPICE-02: 2T1R MCOL connections incorrect (expected: drain=n2, gate=CSL, source=SL, bulk=SL)")
	}
	if !strings.Contains(subckt, "XFE n1 BL fefet_cell") {
		t.Error("M6-SPICE-02: 2T1R XFE connections incorrect (expected: top=n1, bottom=BL)")
	}

	t.Logf("M6-SPICE-02 PASS: 2T1R subcircuit topology validated — ports: WL CSL BL SL, dual selectors in series")
}

// TestM6_SPICE_02_Subcircuit_FeFET_CellStructure validates fefet_cell subcircuit:
// - Contains Rfe (programmable resistance) and XFE (LK FeCap instance)
// - Ports: TE (top electrode), BE (bottom electrode)
// - Parameterized R_level
func TestM6_SPICE_02_Subcircuit_FeFET_CellStructure(t *testing.T) {
	subckt := GenerateFeFETSubcircuit(DefaultHzoFeFETMaterial())

	// Verify fefet_cell subcircuit
	if !strings.Contains(subckt, ".subckt fefet_cell TE BE PARAMS: R_level=1e4") {
		t.Fatal("M6-SPICE-02: fefet_cell subcircuit missing or has incorrect header")
	}

	// Verify programmable resistance
	if !strings.Contains(subckt, "Rfe TE n1 {R_level}") {
		t.Error("M6-SPICE-02: fefet_cell missing Rfe (programmable resistance)")
	}

	// Verify FeCap instance
	if !strings.Contains(subckt, "XFE n1 BE FECAP_HZO") {
		t.Error("M6-SPICE-02: fefet_cell missing XFE instance to FECAP_HZO")
	}

	// Verify FECAP_HZO subcircuit is defined
	if !strings.Contains(subckt, ".subckt FECAP_HZO pos neg") {
		t.Error("M6-SPICE-02: Missing FECAP_HZO subcircuit definition")
	}

	// Verify FECAP_HZO contains LK model elements
	if !strings.Contains(subckt, "beta=") || !strings.Contains(subckt, "gamma=") {
		t.Error("M6-SPICE-02: FECAP_HZO missing Landau-Khalatnikov beta/gamma parameters")
	}

	t.Log("M6-SPICE-02 PASS: fefet_cell subcircuit structure validated — Rfe + FECAP_HZO with LK params")
}

// TestM6_SPICE_02_Subcircuit_FeCap_Connections validates FECAP_HZO port order
func TestM6_SPICE_02_Subcircuit_FeCap_Connections(t *testing.T) {
	subckt := GenerateFeFETSubcircuit(DefaultHzoFeFETMaterial())

	// Extract FECAP_HZO subcircuit definition
	lines := strings.Split(subckt, "\n")
	var fecapDef string
	inFecap := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ".subckt FECAP_HZO") {
			inFecap = true
		}
		if inFecap {
			fecapDef += line + "\n"
			if strings.HasPrefix(trimmed, ".ends") && strings.Contains(strings.ToLower(trimmed), "fecap") {
				break
			}
		}
	}

	// Verify port order: pos neg
	if !strings.Contains(fecapDef, ".subckt FECAP_HZO pos neg") {
		t.Fatal("M6-SPICE-02: FECAP_HZO subcircuit missing or has incorrect port order (expected: pos neg)")
	}

	// Verify it contains capacitor element or behavioral model
	if !strings.Contains(fecapDef, "Cfe") && !strings.Contains(fecapDef, "B") {
		t.Error("M6-SPICE-02: FECAP_HZO missing capacitor or behavioral element")
	}

	t.Log("M6-SPICE-02 PASS: FECAP_HZO port order validated — pos neg")
}
