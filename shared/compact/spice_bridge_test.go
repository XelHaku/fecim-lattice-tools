package compact

import (
	"strings"
	"testing"
)

func TestGenerateVerilogAHeader_NonEmpty(t *testing.T) {
	params := map[string]float64{
		"Qs":  0.25,
		"Vc":  0.01,
		"a":   10.0,
		"Rho": 0.0,
	}
	out := GenerateVerilogAHeader("fecap_model", params)
	if len(out) == 0 {
		t.Fatal("GenerateVerilogAHeader returned empty string")
	}
	if !strings.Contains(out, "module fecap_model") {
		t.Errorf("output missing module header: %q", out)
	}
	if !strings.Contains(out, "parameter real") {
		t.Errorf("output missing parameter statements: %q", out)
	}
	if !strings.Contains(out, "endmodule") {
		t.Errorf("output missing endmodule: %q", out)
	}
}

func TestGenerateVerilogAHeader_DeterministicOrder(t *testing.T) {
	params := map[string]float64{"z": 3.0, "a": 1.0, "m": 2.0}
	out1 := GenerateVerilogAHeader("m", params)
	out2 := GenerateVerilogAHeader("m", params)
	if out1 != out2 {
		t.Errorf("output is not deterministic:\n%q\n%q", out1, out2)
	}
	// Keys should appear in sorted order: a, m, z
	ia := strings.Index(out1, "parameter real a")
	im := strings.Index(out1, "parameter real m")
	iz := strings.Index(out1, "parameter real z")
	if !(ia < im && im < iz) {
		t.Errorf("parameters not in sorted order in:\n%s", out1)
	}
}

func TestGenerateSPICESubcircuit_NonEmpty(t *testing.T) {
	params := map[string]float64{"Qs": 0.25, "Vc": 0.01}
	nodes := []string{"drain", "gate", "source", "body"}
	out := GenerateSPICESubcircuit("fecap", nodes, params)
	if len(out) == 0 {
		t.Fatal("GenerateSPICESubcircuit returned empty string")
	}
	if !strings.Contains(out, ".SUBCKT fecap") {
		t.Errorf("output missing .SUBCKT header: %q", out)
	}
	if !strings.Contains(out, ".param") {
		t.Errorf("output missing .param statements: %q", out)
	}
	if !strings.Contains(out, ".ENDS fecap") {
		t.Errorf("output missing .ENDS: %q", out)
	}
	for _, n := range nodes {
		if !strings.Contains(out, n) {
			t.Errorf("output missing node %q: %q", n, out)
		}
	}
}

func TestGenerateSPICESubcircuit_DeterministicOrder(t *testing.T) {
	params := map[string]float64{"z": 3.0, "a": 1.0, "m": 2.0}
	nodes := []string{"p", "n"}
	out1 := GenerateSPICESubcircuit("sub", nodes, params)
	out2 := GenerateSPICESubcircuit("sub", nodes, params)
	if out1 != out2 {
		t.Errorf("output is not deterministic:\n%q\n%q", out1, out2)
	}
}
