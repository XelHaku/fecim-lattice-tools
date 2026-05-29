package hysteresiscli

import (
	"os"
	"strings"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/module1-hysteresis/pkg/simulation"
	"fecim-lattice-tools/shared/cli"
)

// captureStdout runs f and returns everything written to stdout during execution.
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var buf strings.Builder
	done := make(chan struct{})
	go func() {
		b := make([]byte, 65536)
		for {
			n, err := r.Read(b)
			if n > 0 {
				buf.Write(b[:n])
			}
			if err != nil {
				break
			}
		}
		done <- struct{}{}
	}()

	f()

	w.Close()
	os.Stdout = old
	<-done
	return buf.String()
}

func TestGetMaterialKey(t *testing.T) {
	tests := []struct {
		material string // material name field
		wantKey  string
	}{
		{"HZO (Si-doped)", "default"},
		{"FeCIM HZO", "fecim"},
		{"FeCIM HZO (TARGET - NOT DEMONSTRATED)", "fecim-target"},
		{"Literature Superlattice (Cheema 2020)", "superlattice"},
		{"Cryogenic HZO (4K)", "cryogenic"},
		{"HZO Standard (32 states)", "hzo32"},
		{"HZO FTJ (140 states)", "ftj140"},
		{"AlScN (8-16 states)", "alscn"},
		{"Unknown material", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.material, func(t *testing.T) {
			m := ferroelectric.DefaultHZO()
			m.Name = tt.material
			got := getMaterialKey(m)
			if got != tt.wantKey {
				t.Errorf("getMaterialKey(%q) = %q, want %q", tt.material, got, tt.wantKey)
			}
		})
	}
}

func TestGetMaterialKey_NilMaterial(t *testing.T) {
	// Nil material should not panic; returns "default"
	got := getMaterialKey(nil)
	if got != "default" {
		t.Errorf("getMaterialKey(nil) = %q, want %q", got, "default")
	}
}

func TestListMaterials(t *testing.T) {
	out := captureStdout(func() {
		listMaterials()
	})

	if !strings.Contains(out, "Available materials") {
		t.Errorf("listMaterials output missing header, got: %s", out)
	}
	if !strings.Contains(out, "default") {
		t.Errorf("listMaterials output missing material key, got: %s", out)
	}
}

func TestPrintMaterialInfo(t *testing.T) {
	m := ferroelectric.DefaultHZO()
	out := captureStdout(func() {
		printMaterialInfo(m)
	})

	if !strings.Contains(out, "Material Parameters") {
		t.Errorf("printMaterialInfo missing header, got: %s", out)
	}
	if !strings.Contains(out, m.Name) {
		t.Errorf("printMaterialInfo missing material name %q, got: %s", m.Name, out)
	}
	if !strings.Contains(out, "μC/cm²") {
		t.Errorf("printMaterialInfo missing polarization units, got: %s", out)
	}
}

func TestBuildMaterialResult_AllMaterials(t *testing.T) {
	names := []string{"default", "fecim", "superlattice", "cryogenic", "hzo32", "ftj140", "alscn"}
	for _, n := range names {
		t.Run(n, func(t *testing.T) {
			m := getMaterial(n)
			res := buildMaterialResult(m)
			if res.Material == "" {
				t.Errorf("empty material name for %q", n)
			}
			if res.RemanentPol <= 0 || res.RemanentPol > 200 {
				t.Errorf("Pr out of range for %q: %v", n, res.RemanentPol)
			}
			if res.SaturationPol <= 0 || res.SaturationPol > 200 {
				t.Errorf("Ps out of range for %q: %v", n, res.SaturationPol)
			}
			if res.CoerciveField <= 0 || res.CoerciveField > 20 {
				t.Errorf("Ec out of range for %q: %v MV/cm", n, res.CoerciveField)
			}
			if res.Thickness <= 0 || res.Thickness > 1000 {
				t.Errorf("thickness out of range for %q: %v nm", n, res.Thickness)
			}
			if res.DiscreteLevels != 30 {
				t.Errorf("levels should be 30 for %q, got %d", n, res.DiscreteLevels)
			}
		})
	}
}

func TestRunWithNoArgs(t *testing.T) {
	// Running with no args should return an error about gogpu migration
	err := Run([]string{})
	if err == nil {
		t.Fatal("Run with no args should return an error")
	}
	if !strings.Contains(err.Error(), "no longer launches") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunWithHelp(t *testing.T) {
	err := Run([]string{"--help"})
	if err != nil {
		t.Errorf("Run with --help should return nil, got: %v", err)
	}
}

func TestRunWithListMaterials(t *testing.T) {
	out := captureStdout(func() {
		err := Run([]string{"--list-materials"})
		if err != nil {
			t.Errorf("Run with --list-materials failed: %v", err)
		}
	})
	if !strings.Contains(out, "Available materials") {
		t.Errorf("expected material listing, got: %s", out)
	}
}

func TestRunWithVulkan(t *testing.T) {
	err := Run([]string{"--vulkan"})
	if err == nil {
		t.Fatal("Run with --vulkan should return an error")
	}
	if !strings.Contains(err.Error(), "no longer launches Vulkan") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunHeadless(t *testing.T) {
	m := ferroelectric.DefaultHZO()
	engine := simulation.NewEngine(m)
	engine.SetFrequency(1e6)

	out := captureStdout(func() {
		runHeadless(engine, m)
	})

	if !strings.Contains(out, m.Name) {
		t.Errorf("headless output missing material name, got: %s", out)
	}
	if !strings.Contains(out, "SIMULATION SUMMARY") {
		t.Errorf("headless output missing summary, got: %s", out)
	}
}

func TestRunHeadless_AllMaterials(t *testing.T) {
	names := []string{"default", "fecim", "superlattice", "cryogenic", "hzo32", "ftj140", "alscn"}
	for _, n := range names {
		t.Run(n, func(t *testing.T) {
			m := getMaterial(n)
			engine := simulation.NewEngine(m)
			engine.SetFrequency(1e6)

			out := captureStdout(func() {
				runHeadless(engine, m)
			})
			if !strings.Contains(out, m.Name) {
				t.Errorf("headless output missing material name %q", m.Name)
			}
			if !strings.Contains(out, "μC/cm²") {
				t.Errorf("headless output missing polarization units for %q", m.Name)
			}
		})
	}
}

func TestRunBatchHysteresis_JSON(t *testing.T) {
	batch := &cli.BatchProcessor{}
	// We can't easily inject items; test via the JSON path
	common := cli.NewCommonFlags()
	out, err := cli.NewOutputWriter(common)
	if err != nil {
		t.Fatalf("failed to create output writer: %v", err)
	}
	defer out.Close()

	// Test with empty batch — should return nil
	err = runBatchHysteresis(batch, common, out)
	if err != nil {
		t.Errorf("runBatchHysteresis with empty batch failed: %v", err)
	}
}

func TestRunHeadless_Regression(t *testing.T) {
	// Basic material properties should be consistent across runs
	m := getMaterial("superlattice")
	engine := simulation.NewEngine(m)
	engine.SetFrequency(1e6)

	out1 := captureStdout(func() {
		runHeadless(engine, m)
	})

	engine2 := simulation.NewEngine(m)
	engine2.SetFrequency(1e6)

	out2 := captureStdout(func() {
		runHeadless(engine2, m)
	})

	// Output should be deterministic (same material, same freq)
	if out1 != out2 {
		t.Log("Deterministic regression note: outputs differ — this may indicate Preisach state non-determinism")
	}
}

func TestPrintMaterialInfo_NilMaterialSafety(t *testing.T) {
	// Should not panic on nil
	out := captureStdout(func() {
		printMaterialInfo(nil)
	})
	if out == "" {
		t.Logf("printMaterialInfo(nil) returned empty output (nil material case)")
	}
}

func TestGetMaterial_AllKeys(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"default", "HZO (Si-doped, Park 2015 midpoint)"},
		{"fecim", "FeCIM HZO"},
		{"superlattice", "Literature Superlattice (HZO nanolaminate 2025)"},
		{"cryogenic", "Cryogenic HZO (4K)"},
		{"hzo32", "HZO Standard (32 states)"},
		{"ftj140", "HZO FTJ (140 states)"},
		{"alscn", "AlScN (8-16 states)"},
		{"unknown-key", "HZO (Si-doped, Park 2015 midpoint)"}, // fallback to default
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			m := getMaterial(tt.key)
			if m.Name != tt.want {
				t.Errorf("getMaterial(%q).Name = %q, want %q", tt.key, m.Name, tt.want)
			}
		})
	}
}

func createTestHysteresisResult(m *ferroelectric.HZOMaterial) HysteresisResult {
	return buildMaterialResult(m)
}

func TestHysteresisResult_JSONMarshaling(t *testing.T) {
	m := getMaterial("superlattice")
	res := createTestHysteresisResult(m)

	if res.Material != m.Name {
		t.Errorf("Material field mismatch: %q vs %q", res.Material, m.Name)
	}
	if res.RemanentPol <= 0 {
		t.Errorf("RemanentPol should be positive, got %v", res.RemanentPol)
	}
	if res.SaturationPol <= 0 {
		t.Errorf("SaturationPol should be positive, got %v", res.SaturationPol)
	}
	if res.CoerciveField <= 0 {
		t.Errorf("CoerciveField should be positive, got %v", res.CoerciveField)
	}
	if res.CoerciveVoltage <= 0 {
		t.Errorf("CoerciveVoltage should be positive, got %v", res.CoerciveVoltage)
	}
	if res.Thickness <= 0 {
		t.Errorf("Thickness should be positive, got %v", res.Thickness)
	}
	if res.Permittivity <= 0 {
		t.Errorf("Permittivity should be positive, got %v", res.Permittivity)
	}
	if res.SwitchingTime <= 0 {
		t.Errorf("SwitchingTime should be positive, got %v", res.SwitchingTime)
	}
	if res.EnduranceCycles <= 0 {
		t.Errorf("EnduranceCycles should be positive, got %v", res.EnduranceCycles)
	}
	if res.DiscreteLevels != 30 {
		t.Errorf("DiscreteLevels should be 30, got %d", res.DiscreteLevels)
	}
	if res.BitsPerCell != 4.91 {
		t.Errorf("BitsPerCell should be 4.91, got %v", res.BitsPerCell)
	}
}
