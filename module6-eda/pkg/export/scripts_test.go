package export

import (
	"encoding/json"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// defaultTestArrayConfig returns a small array config for testing.
func defaultTestArrayConfig() config.ArrayConfig {
	return config.ArrayConfig{
		Rows:         8,
		Cols:         8,
		Mode:         "compute",
		Architecture: "passive",
		Technology:   "SKY130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}
}

// ── GenerateSynthesisScript ────────────────────────────────────────────────

func TestGenerateSynthesisScript_NonEmpty(t *testing.T) {
	s := GenerateSynthesisScript(defaultTestArrayConfig())
	if s == "" {
		t.Fatal("expected non-empty synthesis script")
	}
}

func TestGenerateSynthesisScript_ContainsYosysCommands(t *testing.T) {
	s := GenerateSynthesisScript(defaultTestArrayConfig())
	for _, kw := range []string{
		"read_verilog",
		"hierarchy",
		"-check",
		"check",
		"stat",
		"fecim_crossbar_8x8",
	} {
		if !strings.Contains(s, kw) {
			t.Errorf("synthesis script missing keyword %q", kw)
		}
	}
}

func TestGenerateSynthesisScript_BlackboxCell(t *testing.T) {
	s := GenerateSynthesisScript(defaultTestArrayConfig())
	if !strings.Contains(s, "-lib") {
		t.Error("synthesis script must use read_verilog -lib for blackbox cell")
	}
	if !strings.Contains(s, "fecim_bitcell") {
		t.Error("synthesis script must reference fecim_bitcell")
	}
}

func TestGenerateSynthesisScript_1T1RCellName(t *testing.T) {
	cfg := defaultTestArrayConfig()
	cfg.Architecture = "1t1r"
	s := GenerateSynthesisScript(cfg)
	if !strings.Contains(s, "fecim_1t1r_bitcell") {
		t.Error("1T1R architecture must use fecim_1t1r_bitcell in synthesis script")
	}
}

// ── GenerateKLayoutGDSScript ───────────────────────────────────────────────

func TestGenerateKLayoutGDSScript_NonEmpty(t *testing.T) {
	s := GenerateKLayoutGDSScript(defaultTestArrayConfig())
	if s == "" {
		t.Fatal("expected non-empty KLayout GDS script")
	}
}

func TestGenerateKLayoutGDSScript_ContainsPya(t *testing.T) {
	s := GenerateKLayoutGDSScript(defaultTestArrayConfig())
	if !strings.Contains(s, "import pya") {
		t.Error("KLayout script must import pya")
	}
	if !strings.Contains(s, "GDS2") {
		t.Error("KLayout script must reference GDS2 format")
	}
	if !strings.Contains(s, "pya.Layout()") {
		t.Error("KLayout script must create pya.Layout()")
	}
}

func TestGenerateKLayoutGDSScript_ContainsDesignName(t *testing.T) {
	s := GenerateKLayoutGDSScript(defaultTestArrayConfig())
	if !strings.Contains(s, "fecim_crossbar_8x8") {
		t.Error("KLayout script must reference design name fecim_crossbar_8x8")
	}
}

func TestGenerateKLayoutGDSScript_UsageInstructions(t *testing.T) {
	s := GenerateKLayoutGDSScript(defaultTestArrayConfig())
	if !strings.Contains(s, "klayout -z -r gen_gds.py") {
		t.Error("KLayout script must include usage instructions")
	}
}

// ── GenerateOpenROADFlowScript ─────────────────────────────────────────────

func TestGenerateOpenROADFlowScript_NonEmpty(t *testing.T) {
	s := GenerateOpenROADFlowScript(defaultTestArrayConfig())
	if s == "" {
		t.Fatal("expected non-empty OpenROAD flow script")
	}
}

func TestGenerateOpenROADFlowScript_ContainsTCLCommands(t *testing.T) {
	s := GenerateOpenROADFlowScript(defaultTestArrayConfig())
	for _, kw := range []string{
		"read_lef",
		"read_def",
		"check_placement",
		"report_checks",
		"write_def",
	} {
		if !strings.Contains(s, kw) {
			t.Errorf("OpenROAD flow script missing TCL command %q", kw)
		}
	}
}

func TestGenerateOpenROADFlowScript_ClocklessFeCIM(t *testing.T) {
	s := GenerateOpenROADFlowScript(defaultTestArrayConfig())
	if !strings.Contains(s, "clockless") {
		t.Error("OpenROAD flow script must document that FeCIM array is clockless")
	}
}

func TestGenerateOpenROADFlowScript_EnvVars(t *testing.T) {
	s := GenerateOpenROADFlowScript(defaultTestArrayConfig())
	if !strings.Contains(s, "CELL_LEF") {
		t.Error("OpenROAD flow script must use CELL_LEF env var")
	}
	if !strings.Contains(s, "DEF_FILE") {
		t.Error("OpenROAD flow script must use DEF_FILE env var")
	}
}

// ── GenerateFlowRunner ─────────────────────────────────────────────────────

func TestGenerateFlowRunner_NonEmpty(t *testing.T) {
	s := GenerateFlowRunner(defaultTestArrayConfig())
	if s == "" {
		t.Fatal("expected non-empty flow runner script")
	}
}

func TestGenerateFlowRunner_IsBashScript(t *testing.T) {
	s := GenerateFlowRunner(defaultTestArrayConfig())
	if !strings.HasPrefix(s, "#!/usr/bin/env bash") {
		t.Error("flow runner must start with bash shebang")
	}
}

func TestGenerateFlowRunner_MentionsLibreLane(t *testing.T) {
	s := GenerateFlowRunner(defaultTestArrayConfig())
	if !strings.Contains(s, "librelane") {
		t.Error("flow runner must mention LibreLane as preferred tool")
	}
}

func TestGenerateFlowRunner_HasAllSteps(t *testing.T) {
	s := GenerateFlowRunner(defaultTestArrayConfig())
	for _, step := range []string{"yosys", "klayout", "openroad"} {
		if !strings.Contains(s, step) {
			t.Errorf("flow runner must include step: %s", step)
		}
	}
}

// ── GenerateLibreLaneConfig ────────────────────────────────────────────────

func TestGenerateLibreLaneConfig_NonEmpty(t *testing.T) {
	s := GenerateLibreLaneConfig(defaultTestArrayConfig())
	if s == "" {
		t.Fatal("expected non-empty LibreLane config")
	}
}

func TestGenerateLibreLaneConfig_IsValidJSON(t *testing.T) {
	s := GenerateLibreLaneConfig(defaultTestArrayConfig())
	// Check basic JSON structure
	if !strings.HasPrefix(strings.TrimSpace(s), "{") {
		t.Error("LibreLane config must be valid JSON (start with {)")
	}
	if !strings.HasSuffix(strings.TrimSpace(s), "}") {
		t.Error("LibreLane config must be valid JSON (end with })")
	}
}

func TestGenerateLibreLaneConfig_RequiredFields(t *testing.T) {
	s := GenerateLibreLaneConfig(defaultTestArrayConfig())
	for _, field := range []string{
		"DESIGN_NAME",
		"VERILOG_FILES",
		"BASE_SDC_FILE",
		"RUN_CTS",
		"EXTRA_LEFS",
		"EXTRA_LIBS",
		"EXTRA_GDS_FILES",
		"SYNTH_ELABORATE_ONLY",
		"FP_DEF_TEMPLATE",
	} {
		if !strings.Contains(s, field) {
			t.Errorf("LibreLane config missing required field: %s", field)
		}
	}
}

func TestGenerateLibreLaneConfig_NoV2Label(t *testing.T) {
	s := GenerateLibreLaneConfig(defaultTestArrayConfig())
	// Must NOT claim to be "OpenLane v2.0" (that is incorrect labeling)
	if strings.Contains(s, "v2.0") {
		t.Error("LibreLane config must not label itself as 'OpenLane v2.0' (OpenLane is v1.x; LibreLane is the v2 successor)")
	}
}

func TestGenerateLibreLaneConfig_GDSNote(t *testing.T) {
	s := GenerateLibreLaneConfig(defaultTestArrayConfig())
	// Must document that GDS needs to be generated via KLayout
	if !strings.Contains(s, "gen_gds.py") {
		t.Error("LibreLane config must document KLayout GDS generation step")
	}
}

// ── cellNameFromArch ───────────────────────────────────────────────────────

func TestCellNameFromArch(t *testing.T) {
	cases := []struct {
		arch string
		want string
	}{
		{"passive", "fecim_bitcell"},
		{"1t1r", "fecim_1t1r_bitcell"},
		{"2t1r", "fecim_2t1r_bitcell"},
		{"", "fecim_bitcell"},
		{"PASSIVE", "fecim_bitcell"},
	}
	for _, c := range cases {
		got := cellNameFromArch(c.arch)
		if got != c.want {
			t.Errorf("cellNameFromArch(%q) = %q, want %q", c.arch, got, c.want)
		}
	}
}

// ── Architecture variation ─────────────────────────────────────────────────

func TestAllScriptGenerators_2T1RArch(t *testing.T) {
	cfg := defaultTestArrayConfig()
	cfg.Architecture = "2t1r"

	if !strings.Contains(GenerateSynthesisScript(cfg), "fecim_2t1r_bitcell") {
		t.Error("synthesis script: 2T1R must use fecim_2t1r_bitcell")
	}
	if !strings.Contains(GenerateKLayoutGDSScript(cfg), "fecim_2t1r_bitcell") {
		t.Error("KLayout script: 2T1R must use fecim_2t1r_bitcell")
	}
	if !strings.Contains(GenerateLibreLaneConfig(cfg), "fecim_2t1r_bitcell") {
		t.Error("LibreLane config: 2T1R must use fecim_2t1r_bitcell")
	}
}

// ── Large array ────────────────────────────────────────────────────────────

func TestAllScriptGenerators_LargeArray(t *testing.T) {
	cfg := config.ArrayConfig{
		Rows: 128, Cols: 128,
		Mode: "compute", Architecture: "1t1r",
		Technology: "GF180MCU",
		CellWidth: 0.46, CellHeight: 2.72,
	}

	// All generators must handle large arrays without panicking
	_ = GenerateSynthesisScript(cfg)
	_ = GenerateKLayoutGDSScript(cfg)
	_ = GenerateOpenROADFlowScript(cfg)
	_ = GenerateFlowRunner(cfg)
	_ = GenerateLibreLaneConfig(cfg)
}

// TestGenerateLibreLaneConfig_ValidJSONParse verifies the config is parseable JSON
// and contains the MACROS dict required by OpenLane 2 / LibreLane.
func TestGenerateLibreLaneConfig_ValidJSONParse(t *testing.T) {
	s := GenerateLibreLaneConfig(defaultTestArrayConfig())
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("LibreLane config is not valid JSON: %v\n%s", err, s)
	}

	// MACROS dict must be present (OpenLane 2 idiomatic format)
	macros, ok := m["MACROS"]
	if !ok {
		t.Fatal("LibreLane config missing MACROS dict (required for OpenLane 2 custom cell integration)")
	}
	macrosMap, ok := macros.(map[string]interface{})
	if !ok {
		t.Fatal("MACROS must be a JSON object")
	}

	// Must have exactly one entry: the cell name
	if len(macrosMap) != 1 {
		t.Fatalf("MACROS must have exactly 1 cell entry, got %d", len(macrosMap))
	}

	// The cell entry must have gds, lef, lib, nl fields
	for cellName, cellVal := range macrosMap {
		cellMap, ok := cellVal.(map[string]interface{})
		if !ok {
			t.Fatalf("MACROS[%q] must be an object", cellName)
		}
		for _, required := range []string{"gds", "lef", "lib", "nl"} {
			if _, exists := cellMap[required]; !exists {
				t.Errorf("MACROS[%q] missing field %q", cellName, required)
			}
		}
		t.Logf("MACROS[%q]: gds/lef/lib/nl all present", cellName)
	}
}

// TestGenerateLibreLaneConfig_CornerNames verifies Liberty corner names match
// SKY130 PDK convention so OpenLane 2 MACROS lib glob patterns resolve correctly.
func TestGenerateLibreLaneConfig_CornerNames(t *testing.T) {
	// Liberty multi-corner file uses SKY130-convention names like tt_025C_1v80
	lib := GenerateMultiCornerLiberty(config.DefaultCellConfig())
	for _, corner := range []string{"tt_025C_1v80", "ff_n40C_1v95", "ss_125C_1v60"} {
		if !strings.Contains(lib, corner) {
			t.Errorf("Liberty missing SKY130-convention corner name: %s", corner)
		}
	}
}
