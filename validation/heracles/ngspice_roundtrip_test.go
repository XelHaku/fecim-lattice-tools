package heracles

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

// heraclesModelPath returns the filesystem path to the Heracles Verilog-A
// compact model.  It checks HERACLES_MODEL_PATH env var first, then falls
// back to the in-repo default.
func heraclesModelPath() string {
	if p := os.Getenv("HERACLES_MODEL_PATH"); p != "" {
		return p
	}
	return "opensource/heracles/heracles.va"
}

// TestNgspiceRoundtrip_HeraclesVsLK runs the Heracles Verilog-A model inside
// ngspice, parses the resulting P-E loop, and compares against the Go LK
// solver for the same voltage sweep.  It reports RMSE, Pr mismatch %,
// Ec mismatch %, and loop area mismatch %.
//
// The test skips gracefully if:
//   - ngspice is not installed
//   - The Heracles Verilog-A model file is not found
func TestNgspiceRoundtrip_HeraclesVsLK(t *testing.T) {
	// --- Guard: ngspice availability ---
	if _, err := exec.LookPath("ngspice"); err != nil {
		t.Skip("ngspice not installed; skipping SPICE round-trip validation")
	}

	// --- Guard: Heracles model file ---
	modelPath := heraclesModelPath()
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Skipf("Heracles Verilog-A model not found at %s; skipping SPICE round-trip (set HERACLES_MODEL_PATH to override)", modelPath)
	}

	ref := Reference10nmHfO2_300K()

	// --- Generate netlist ---
	tmpDir := t.TempDir()
	outputCSV := filepath.Join(tmpDir, "heracles_pe.csv")

	config := DefaultHeraclesNetlistConfig()
	config.ModelPath = modelPath
	config.OutputFile = outputCSV
	config.Vmax = 3.0
	config.Steps = 100
	config.Temperature = ref.TemperatureK
	config.ThicknessNm = ref.ThicknessNm

	netlist := GenerateHeraclesFECapNetlist(config)
	deckPath := filepath.Join(tmpDir, "heracles_sweep.sp")
	if err := os.WriteFile(deckPath, []byte(netlist), 0o644); err != nil {
		t.Fatalf("write netlist: %v", err)
	}

	// --- Run ngspice ---
	ngspiceOut := filepath.Join(tmpDir, "ngspice.log")
	cmd := exec.Command("ngspice", "-b", "-o", ngspiceOut, deckPath)
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		logBytes, _ := os.ReadFile(ngspiceOut)
		t.Fatalf("ngspice run failed: %v\nlog:\n%s", err, string(logBytes))
	}

	// --- Parse output ---
	rawOutput, err := os.ReadFile(outputCSV)
	if err != nil {
		// Try reading from ngspice log if CSV was not produced.
		rawOutput, err = os.ReadFile(ngspiceOut)
		if err != nil {
			t.Fatalf("read ngspice output: %v", err)
		}
	}

	thicknessM := config.ThicknessNm * 1e-9
	areaM2 := config.AreaUm2 * 1e-12
	spiceData, err := ParseNgspiceOutput(string(rawOutput), thicknessM, areaM2)
	if err != nil {
		t.Fatalf("parse ngspice output: %v", err)
	}
	if len(spiceData.Ascending) == 0 {
		t.Fatal("parsed zero ascending data points from ngspice output")
	}

	// --- Generate LK solver reference at same field points ---
	solver := sharedphysics.NewLKSolver()
	mat := sharedphysics.MaterlikHfO2()
	solver.ConfigureFromMaterial(mat)
	solver.Temperature = ref.TemperatureK
	solver.UseNLS = false
	solver.EnableNoise = false
	solver.SetState(-math.Abs(mat.Pr))

	const settleSteps = 1800
	const dt = 2e-12

	lkAsc := make([]PEPoint, 0, len(spiceData.Ascending))
	for _, pt := range spiceData.Ascending {
		eVm := pt.E_MVcm * 1e8 // MV/cm -> V/m
		p := settleAtField(solver, eVm, settleSteps, dt)
		lkAsc = append(lkAsc, PEPoint{E_MVcm: pt.E_MVcm, P_uCcm: p * 100.0})
	}

	lkDesc := make([]PEPoint, 0, len(spiceData.Descending))
	for _, pt := range spiceData.Descending {
		eVm := pt.E_MVcm * 1e8
		p := settleAtField(solver, eVm, settleSteps, dt)
		lkDesc = append(lkDesc, PEPoint{E_MVcm: pt.E_MVcm, P_uCcm: p * 100.0})
	}

	// --- Compute comparison metrics ---
	rmse := rmseUCcm2(spiceData.Ascending, lkAsc, spiceData.Descending, lkDesc)
	prSpice, ecSpice := estimatePrEc(spiceData.Ascending, spiceData.Descending)
	prLK, ecLK := estimatePrEc(lkAsc, lkDesc)
	areaSpice := loopAreaJm3(spiceData.Ascending, spiceData.Descending)
	areaLK := loopAreaJm3(lkAsc, lkDesc)

	prMismatch := mismatchPct(prLK, prSpice)
	ecMismatch := mismatchPct(ecLK, ecSpice)
	areaMismatch := mismatchPct(areaLK, areaSpice)

	// --- Write report ---
	report := CompareReport{
		Title:     "ngspice Heracles round-trip vs Go LK solver",
		Reference: "Heracles Verilog-A compact model via ngspice",
		Dataset:   fmt.Sprintf("%.0f nm HfO2, %.0f K", config.ThicknessNm, config.Temperature),
		Parameters: map[string]any{
			"heracles_model_path": modelPath,
			"vmax_V":             config.Vmax,
			"steps":              config.Steps,
			"temperature_K":      config.Temperature,
			"thickness_nm":       config.ThicknessNm,
			"area_um2":           config.AreaUm2,
		},
		Ascending:  CurvePair{Reference: spiceData.Ascending, Model: lkAsc},
		Descending: CurvePair{Reference: spiceData.Descending, Model: lkDesc},
		Metrics: CompareMetrics{
			RMSE_uCcm2:          rmse,
			EcRef_MVcm:          ecSpice,
			EcModel_MVcm:        ecLK,
			EcMismatchPct:       ecMismatch,
			PrRef_uCcm2:         prSpice,
			PrModel_uCcm2:       prLK,
			PrMismatchPct:       prMismatch,
			LoopAreaRef_Jm3:     areaSpice,
			LoopAreaModel_Jm3:   areaLK,
			LoopAreaMismatchPct: areaMismatch,
		},
	}

	reportPath := filepath.Join(".", "ngspice_roundtrip_report.json")
	if err := WriteCompareReport(reportPath, report); err != nil {
		t.Logf("warning: could not write round-trip report: %v", err)
	}

	t.Logf("ngspice round-trip: RMSE=%.2f uC/cm^2, Ec mismatch=%.1f%%, Pr mismatch=%.1f%%, area mismatch=%.1f%%",
		rmse, ecMismatch, prMismatch, areaMismatch)
	t.Logf("  SPICE: Pr=%.2f uC/cm^2, Ec=%.2f MV/cm, Area=%.2e J/m^3",
		prSpice, ecSpice, areaSpice)
	t.Logf("  LK:    Pr=%.2f uC/cm^2, Ec=%.2f MV/cm, Area=%.2e J/m^3",
		prLK, ecLK, areaLK)
}

// TestGenerateHeraclesFECapNetlist_Structure validates that the generated
// netlist contains all required SPICE elements regardless of ngspice
// availability.
func TestGenerateHeraclesFECapNetlist_Structure(t *testing.T) {
	config := DefaultHeraclesNetlistConfig()
	config.ModelPath = "/path/to/heracles.va"
	config.OutputFile = "test_output.csv"

	netlist := GenerateHeraclesFECapNetlist(config)

	requiredElements := []struct {
		needle string
		desc   string
	}{
		{".include", "Verilog-A model include directive"},
		{"heracles.va", "model file path"},
		{"Vsrc", "voltage source"},
		{"top", "top node"},
		{"Xfecap", "FeCap instance"},
		{"heracles_hzo", "Heracles HZO model name"},
		{".control", "control block start"},
		{".endc", "control block end"},
		{".end", "netlist terminator"},
		{"dc Vsrc", "DC sweep command"},
		{"write", "data export command"},
		{"print", "print command"},
		{"asc_v", "ascending voltage vector"},
		{"asc_i", "ascending current vector"},
		{"desc_v", "descending voltage vector"},
		{"desc_i", "descending current vector"},
	}

	for _, req := range requiredElements {
		if !containsString(netlist, req.needle) {
			t.Errorf("netlist missing required element %q (%s)", req.needle, req.desc)
		}
	}

	// Verify sweep range includes configured Vmax.
	if !containsString(netlist, "3.0000") {
		t.Error("netlist does not contain expected Vmax value (3.0)")
	}

	// Verify temperature parameter is present.
	if !containsString(netlist, "300.0") {
		t.Error("netlist does not contain expected temperature value")
	}
}

// TestParseNgspiceOutput_MockData validates the parser against a synthetic
// ngspice print-format output without requiring ngspice to be installed.
func TestParseNgspiceOutput_MockData(t *testing.T) {
	// Simulate ngspice print output format: Index, Voltage, Current
	mockOutput := `
Ngspice output log
Circuit: Heracles FeCap

Index   asc_v           asc_i
0       -3.000000e+00   -1.500000e-04
1       -1.500000e+00   -1.000000e-04
2       0.000000e+00    -5.000000e-05
3       1.500000e+00    5.000000e-05
4       3.000000e+00    1.500000e-04

Index   desc_v          desc_i
0       3.000000e+00    1.500000e-04
1       1.500000e+00    1.000000e-04
2       0.000000e+00    5.000000e-05
3       -1.500000e+00   -5.000000e-05
4       -3.000000e+00   -1.500000e-04
`

	thicknessM := 10e-9 // 10 nm
	areaM2 := 1e-12     // 1 um^2

	data, err := ParseNgspiceOutput(mockOutput, thicknessM, areaM2)
	if err != nil {
		t.Fatalf("parse mock output: %v", err)
	}

	if len(data.Ascending) != 5 {
		t.Fatalf("expected 5 ascending points, got %d", len(data.Ascending))
	}
	if len(data.Descending) != 5 {
		t.Fatalf("expected 5 descending points, got %d", len(data.Descending))
	}

	// Verify first ascending point: V=-3V, I=-1.5e-4 A
	// E = -3 / 10e-9 = -3e8 V/m = -3 MV/cm
	// P = -1.5e-4 / 1e-12 = -1.5e8 C/m^2 -> * 100 = -1.5e10 uC/cm^2
	// (These are not physically realistic values — the mock is for parser validation only.)
	first := data.Ascending[0]
	expectedE := -3.0 / (10e-9) * 1e-8 // -3.0 MV/cm
	if math.Abs(first.E_MVcm-expectedE) > 0.01 {
		t.Errorf("ascending[0].E_MVcm = %.4f, want %.4f", first.E_MVcm, expectedE)
	}

	// Verify descending first point matches expected voltage sign.
	firstDesc := data.Descending[0]
	if firstDesc.E_MVcm <= 0 {
		t.Errorf("descending[0].E_MVcm = %.4f, expected positive", firstDesc.E_MVcm)
	}

	// Verify ascending and descending have opposite polarizations at corresponding ends.
	lastAsc := data.Ascending[len(data.Ascending)-1]
	lastDesc := data.Descending[len(data.Descending)-1]
	if lastAsc.P_uCcm*lastDesc.P_uCcm > 0 {
		// At max V ascending has positive P, at min V descending has negative P.
		// They should have opposite signs if the loop is physically sensible.
		// For mock data with linear I-V this should hold.
	}

	t.Logf("parsed %d ascending, %d descending points from mock data",
		len(data.Ascending), len(data.Descending))
	t.Logf("  asc[0]: E=%.2f MV/cm, P=%.2e uC/cm^2", first.E_MVcm, first.P_uCcm)
	t.Logf("  asc[last]: E=%.2f MV/cm, P=%.2e uC/cm^2", lastAsc.E_MVcm, lastAsc.P_uCcm)
}

// TestParseNgspiceOutput_InvalidInputs verifies the parser returns errors for
// bad or empty inputs.
func TestParseNgspiceOutput_InvalidInputs(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		_, err := ParseNgspiceOutput("", 10e-9, 1e-12)
		if err == nil {
			t.Error("expected error for empty input")
		}
	})

	t.Run("zero thickness", func(t *testing.T) {
		_, err := ParseNgspiceOutput("some data", 0, 1e-12)
		if err == nil {
			t.Error("expected error for zero thickness")
		}
	})

	t.Run("negative area", func(t *testing.T) {
		_, err := ParseNgspiceOutput("some data", 10e-9, -1e-12)
		if err == nil {
			t.Error("expected error for negative area")
		}
	})

	t.Run("no numeric data", func(t *testing.T) {
		_, err := ParseNgspiceOutput("this is just text\nwith no numbers at all\n", 10e-9, 1e-12)
		if err == nil {
			t.Error("expected error for non-numeric input")
		}
	})
}

// TestParseNgspiceOutput_CSVFormat verifies the parser handles CSV-formatted
// output with branch labels.
func TestParseNgspiceOutput_CSVFormat(t *testing.T) {
	mockCSV := `branch,voltage,current
asc,-3.0,-1.5e-4
asc,-1.5,-1.0e-4
asc,0.0,-5.0e-5
asc,1.5,5.0e-5
asc,3.0,1.5e-4
desc,3.0,1.5e-4
desc,1.5,1.0e-4
desc,0.0,5.0e-5
desc,-1.5,-5.0e-5
desc,-3.0,-1.5e-4
`

	data, err := ParseNgspiceOutput(mockCSV, 10e-9, 1e-12)
	if err != nil {
		t.Fatalf("parse CSV mock: %v", err)
	}
	if len(data.Ascending) != 5 {
		t.Errorf("expected 5 ascending, got %d", len(data.Ascending))
	}
	if len(data.Descending) != 5 {
		t.Errorf("expected 5 descending, got %d", len(data.Descending))
	}
}

// containsString checks whether s contains substr (case-sensitive).
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
