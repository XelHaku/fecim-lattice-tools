package export

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

func TestGenerateFeFETSubcircuit_BalancedBlocks(t *testing.T) {
	s := GenerateFeFETSubcircuit(DefaultHzoFeFETMaterial())
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

func TestGenerateFeFETSubcircuit_CFeRange(t *testing.T) {
	s := GenerateFeFETSubcircuit(DefaultHzoFeFETMaterial())
	re := regexp.MustCompile(`(?m)^\.param C_fe=([0-9.eE+-]+)$`)
	m := re.FindStringSubmatch(s)
	if len(m) != 2 {
		t.Fatalf("missing C_fe param in subcircuit:\n%s", s)
	}
	cfe, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		t.Fatalf("invalid C_fe value %q: %v", m[1], err)
	}
	cfeFF := cfe * 1e15
	if cfeFF < 1.0 || cfeFF > 10.0 {
		t.Fatalf("C_fe out of expected range: %.3f fF (want 1..10 fF)", cfeFF)
	}
}

func TestGenerate1T1RSubcircuit_ContainsMOSAndFeFET(t *testing.T) {
	s := Generate1T1RSubcircuit()
	if !strings.Contains(s, "MSEL") {
		t.Fatal("1T1R subcircuit missing MOSFET instance")
	}
	if !strings.Contains(s, "XFE") || !strings.Contains(s, "fefet_cell") {
		t.Fatal("1T1R subcircuit missing FeFET instance")
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
