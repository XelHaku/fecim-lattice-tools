package export

import (
	"strings"
	"testing"
)

func TestFECapSubcircuit_ValidSyntax(t *testing.T) {
	params := DefaultMaterlikFECapParams()
	netlist := GenerateFECapSubcircuit(params)

	if !strings.Contains(netlist, ".subckt FECAP_HZO pos neg PARAMS:") {
		t.Fatalf("missing/invalid .subckt header: %s", netlist)
	}
	if !strings.Contains(netlist, "Cfe pos n1 {eps0*25*area/thick}") {
		t.Fatal("missing capacitor instance with expected LK expression")
	}
	if !strings.Contains(netlist, "Rvisc n1 neg {rho*thick/area}") {
		t.Fatal("missing resistor instance with expected LK expression")
	}
	if !strings.Contains(netlist, "B_landau n1 neg V =") {
		t.Fatal("missing behavioral source with expected nodes")
	}
	if !strings.Contains(netlist, ".ends FECAP_HZO") {
		t.Fatal("missing .ends statement")
	}
}

func TestFECapSubcircuit_UsesMaterlikDefaults(t *testing.T) {
	netlist := GenerateFECapSubcircuit(FECapParams{})

	if !strings.Contains(netlist, "beta=-6.720000e+08") {
		t.Fatal("missing Materlik beta default")
	}
	if !strings.Contains(netlist, "gamma=1.950000e+10") {
		t.Fatal("missing Materlik gamma default")
	}
	if !strings.Contains(netlist, "rho=5.000000e-02") {
		t.Fatal("missing default viscosity rho")
	}
}

func Test1T1RSubcircuit_IncludesMOSFET(t *testing.T) {
	params := DefaultMaterlikFECapParams()
	netlist := Generate1T1RSubcircuit(params, "")

	if !strings.Contains(netlist, ".subckt FECAP_HZO_1T1R bl wl sl") {
		t.Fatal("missing 1T1R subcircuit header")
	}
	if !strings.Contains(netlist, "M_sel nmid wl sl sl") {
		t.Fatal("missing MOSFET selector instance")
	}
	if !strings.Contains(netlist, "X_fecap bl nmid FECAP_HZO") {
		t.Fatal("missing FeCap instance inside 1T1R")
	}
	if !strings.Contains(netlist, ".model SKY130NMOS NMOS") {
		t.Fatal("missing default SKY130 NMOS model card")
	}
}
