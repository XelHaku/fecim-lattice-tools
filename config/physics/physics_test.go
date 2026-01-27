package physics

import (
	"testing"
)

func TestLoad(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify constants
	if cfg.Constants.FeCIMLevels != 30 {
		t.Errorf("Expected 30 FeCIM levels, got %d", cfg.Constants.FeCIMLevels)
	}

	if cfg.Constants.BitsPerCell < 4.9 || cfg.Constants.BitsPerCell > 5.0 {
		t.Errorf("Expected ~4.91 bits/cell, got %f", cfg.Constants.BitsPerCell)
	}
}

func TestMaterials(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test default material exists
	defaultMat := cfg.DefaultMaterial()
	if defaultMat == nil {
		t.Fatal("Default material not found")
	}
	if defaultMat.Name == "" {
		t.Error("Default material has no name")
	}

	// Test FeCIM material
	fecim := cfg.FeCIMMaterial()
	if fecim == nil {
		t.Fatal("FeCIM material not found")
	}

	// Verify Ec is in valid range (0.5 - 2.0 MV/cm)
	ecMVcm := fecim.EcMVcm()
	if ecMVcm < 0.5 || ecMVcm > 2.0 {
		t.Errorf("Ec out of range: %f MV/cm", ecMVcm)
	}

	// Verify Pr is positive
	if fecim.PrMicroCcm2() <= 0 {
		t.Errorf("Pr should be positive, got %f", fecim.PrMicroCcm2())
	}
}

func TestMaterialNames(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	names := cfg.MaterialNames()
	if len(names) < 8 {
		t.Errorf("Expected at least 8 materials (CMOS compatible), got %d", len(names))
	}

	// Check all CMOS-compatible materials exist
	expected := []string{
		"default_hzo",
		"fecim_hzo",
		"fecim_hzo_target",
		"literature_superlattice",
		"cryogenic_hzo",
		"hzo_standard_32",
		"hzo_ftj_140",
		"alscn",
	}
	for _, name := range expected {
		found := false
		for _, n := range names {
			if n == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected material %s not found", name)
		}
	}

	// Log all materials for visibility
	t.Logf("Found %d materials:", len(names))
	for _, name := range names {
		mat := cfg.GetMaterial(name)
		t.Logf("  %-25s | %s", name, mat.Name)
	}
}

func TestCrossbarConfig(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	xbar := cfg.Crossbar
	if xbar.QuantizationLevels != 30 {
		t.Errorf("Expected 30 quantization levels, got %d", xbar.QuantizationLevels)
	}

	// Verify conductance range
	if xbar.ConductanceMinS >= xbar.ConductanceMaxS {
		t.Error("Gmin should be less than Gmax")
	}

	ratio := xbar.ConductanceMaxS / xbar.ConductanceMinS
	if ratio < 10 {
		t.Errorf("Conductance ratio should be >= 10, got %f", ratio)
	}
}

func TestCalibrationConfig(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cal := cfg.Calibration
	if cal.Iterations < 10 {
		t.Errorf("Calibration iterations should be >= 10, got %d", cal.Iterations)
	}

	if cal.AdjustmentRate <= 0 || cal.AdjustmentRate > 0.2 {
		t.Errorf("Adjustment rate should be 0 < r <= 0.2, got %f", cal.AdjustmentRate)
	}
}

func TestMaterialConvenienceMethods(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	mat := cfg.FeCIMMaterial()

	// Test CoerciveVoltage (should be ~1V for 10nm film at 1 MV/cm)
	vCoercive := mat.CoerciveVoltage()
	if vCoercive < 0.5 || vCoercive > 2.0 {
		t.Errorf("Coercive voltage out of expected range: %f V", vCoercive)
	}

	// Test thickness conversion
	thicknessNm := mat.ThicknessNm()
	if thicknessNm < 5 || thicknessNm > 50 {
		t.Errorf("Thickness out of expected range: %f nm", thicknessNm)
	}
}

func TestPreisachConfig(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	preisach := cfg.Preisach
	if preisach.GridSize < 10 || preisach.GridSize > 100 {
		t.Errorf("Grid size out of range: %d", preisach.GridSize)
	}

	if preisach.AlphaSigmaRatio <= 0 || preisach.AlphaSigmaRatio > 0.5 {
		t.Errorf("Alpha sigma ratio out of range: %f", preisach.AlphaSigmaRatio)
	}
}
