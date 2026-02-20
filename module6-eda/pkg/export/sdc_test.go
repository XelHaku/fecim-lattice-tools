package export

import (
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

func TestGenerateSDC_NonEmpty(t *testing.T) {
	s := GenerateSDC(defaultTestArrayConfig())
	if s == "" {
		t.Fatal("expected non-empty SDC output")
	}
}

func TestGenerateSDC_ClocklessConstraints(t *testing.T) {
	s := GenerateSDC(defaultTestArrayConfig())
	for _, kw := range []string{
		"set_input_delay",
		"set_output_delay",
		"set_max_transition",
		"set_load",
	} {
		if !strings.Contains(s, kw) {
			t.Errorf("SDC missing constraint: %s", kw)
		}
	}
}

func TestGenerateSDC_ZeroDelays(t *testing.T) {
	s := GenerateSDC(defaultTestArrayConfig())
	if !strings.Contains(s, "0.0 [all_inputs]") {
		t.Error("SDC must set input delay = 0 for clockless array")
	}
	if !strings.Contains(s, "0.0 [all_outputs]") {
		t.Error("SDC must set output delay = 0 for clockless array")
	}
}

func TestGenerateSDC_ClockSectionCommentedOut(t *testing.T) {
	s := GenerateSDC(defaultTestArrayConfig())
	// Clock section must be commented out (FeCIM array has no clock)
	if strings.Contains(s, "\ncreate_clock") {
		t.Error("SDC must have create_clock commented out for clockless design")
	}
	// But the commented-out example must exist as guidance
	if !strings.Contains(s, "create_clock") {
		t.Error("SDC must include commented-out clock example for digital wrapper use case")
	}
}

func TestGenerateSDC_ContainsDesignInfo(t *testing.T) {
	s := GenerateSDC(defaultTestArrayConfig())
	if !strings.Contains(s, "fecim_crossbar_8x8") {
		t.Error("SDC must reference the design name")
	}
}

func TestGenerateSDC_VariousTechnologies(t *testing.T) {
	for _, tech := range []string{"SKY130", "GF180MCU", "IHP_SG13G2", ""} {
		cfg := config.ArrayConfig{
			Rows: 4, Cols: 4, Architecture: "passive",
			Technology: tech, CellWidth: 0.46, CellHeight: 2.72,
		}
		s := GenerateSDC(cfg)
		if s == "" {
			t.Errorf("SDC empty for technology %q", tech)
		}
	}
}

func TestGenerateOpenSTAScript_NonEmpty(t *testing.T) {
	s := GenerateOpenSTAScript(defaultTestArrayConfig())
	if s == "" {
		t.Fatal("expected non-empty OpenSTA script")
	}
}

func TestGenerateOpenSTAScript_ContainsSTACommands(t *testing.T) {
	s := GenerateOpenSTAScript(defaultTestArrayConfig())
	for _, kw := range []string{
		"read_liberty",
		"read_verilog",
		"link_design",
		"read_sdc",
		"report_checks",
	} {
		if !strings.Contains(s, kw) {
			t.Errorf("OpenSTA script missing command: %s", kw)
		}
	}
}

func TestGenerateOpenSTAScript_ReferencesSDCFile(t *testing.T) {
	s := GenerateOpenSTAScript(defaultTestArrayConfig())
	if !strings.Contains(s, "constraints.sdc") {
		t.Error("OpenSTA script must read constraints.sdc")
	}
}
