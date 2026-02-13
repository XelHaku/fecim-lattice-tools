package export

import (
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

func TestGenerateFeFETSubcircuit_BalancedBlocks(t *testing.T) {
	s := GenerateFeFETSubcircuit(DefaultHzoFeFETMaterial())
	if !strings.Contains(s, ".subckt FECAP_HZO pos neg") {
		t.Fatal("missing LK FECAP_HZO subcircuit")
	}
	if !strings.Contains(s, ".subckt fefet_cell") {
		t.Fatal("missing .subckt fefet_cell")
	}
	if !strings.Contains(s, ".ends fefet_cell") {
		t.Fatal("missing .ends fefet_cell")
	}
	if strings.Count(s, ".subckt") != strings.Count(s, ".ends") {
		t.Fatalf("unbalanced subcircuit delimiters: subckt=%d ends=%d", strings.Count(s, ".subckt"), strings.Count(s, ".ends"))
	}
}

func TestGenerateFeFETSubcircuit_UsesMaterlikDefaults(t *testing.T) {
	s := GenerateFeFETSubcircuit(DefaultHzoFeFETMaterial())
	if !strings.Contains(s, "beta=-6.720000e+08") {
		t.Fatalf("missing Materlik beta default in generated subckt:\n%s", s)
	}
	if !strings.Contains(s, "gamma=1.950000e+10") {
		t.Fatalf("missing Materlik gamma default in generated subckt:\n%s", s)
	}
	if !strings.Contains(s, "full nonlinear LK production models generally require Verilog-A") {
		t.Fatalf("missing LK limitation note in generated subckt:\n%s", s)
	}
}

func TestGenerate1T1RSubcircuit_ContainsMOSAndFeFET(t *testing.T) {
	s := Generate1T1RSubcircuit()
	if !strings.Contains(s, ".subckt fefet_1t1r WL BL SL PARAMS: R_level=1e4") {
		t.Fatal("1T1R subcircuit missing ngspice PARAMS header")
	}
	if !strings.Contains(s, "MSEL") || !strings.Contains(s, "SKY130NMOS") {
		t.Fatal("1T1R subcircuit missing SKY130 MOSFET selector")
	}
	if !strings.Contains(s, "XFE") || !strings.Contains(s, "fefet_cell") {
		t.Fatal("1T1R subcircuit missing FeFET instance")
	}
}

func TestGenerate2T1RSubcircuit_ContainsDualMOSAndFeFET(t *testing.T) {
	s := Generate2T1RSubcircuit()
	if !strings.Contains(s, ".subckt fefet_2t1r WL CSL BL SL PARAMS: R_level=1e4") {
		t.Fatal("2T1R subcircuit missing ngspice PARAMS header")
	}
	if !strings.Contains(s, "MROW") || !strings.Contains(s, "MCOL") {
		t.Fatal("2T1R subcircuit missing dual NMOS selectors")
	}
	if !strings.Contains(s, "XFE n1 BL fefet_cell") {
		t.Fatal("2T1R subcircuit missing FeFET instance/node mapping")
	}
}

func TestGenerateSPICE_PassiveArrayInstanceCount(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.ArchPassive,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 5},
			{Row: 0, Col: 1, Conductance: 60.0, Resistance: 16666.7, Level: 10},
			{Row: 1, Col: 0, Conductance: 70.0, Resistance: 14285.7, Level: 15},
			{Row: 1, Col: 1, Conductance: 80.0, Resistance: 12500.0, Level: 20},
		},
		Stats: compiler.DesignStats{TotalCells: 4, ActiveCells: 4},
	}

	netlist := GenerateSPICE(design, 1.8)
	if count := strings.Count(netlist, "\nX_"); count != 4 {
		t.Fatalf("unexpected passive FeFET instance count: got %d want 4", count)
	}
	if !strings.Contains(netlist, ".subckt fefet_cell") || !strings.Contains(netlist, ".ends fefet_cell") {
		t.Fatal("netlist missing FeFET subcircuit definition")
	}
}

func TestGenerateSPICE_2T1RIncludesSelectorsAndPins(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{Mode: compiler.ModeCompute, Architecture: compiler.Arch2T1R},
		Cells:  []compiler.CellAssignment{{Row: 0, Col: 1, Conductance: 50.0, Resistance: 20000.0, Level: 7}},
		Stats:  compiler.DesignStats{TotalCells: 1, ActiveCells: 1},
	}
	netlist := GenerateSPICE(design, 1.8)
	if !strings.Contains(netlist, ".model SKY130NMOS NMOS") {
		t.Fatal("missing SKY130 model card")
	}
	if !strings.Contains(netlist, ".subckt fefet_2t1r") {
		t.Fatal("missing 2T1R subcircuit")
	}
	if !strings.Contains(netlist, "X_0_1 wl0 csl1 bl1 sl0 fefet_2t1r") {
		t.Fatal("2T1R instance connection does not match expected node map WL/CSL/BL/SL")
	}
}
