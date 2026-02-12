package physics

import (
	"math"
	"testing"
)

func TestDefaultHZO(t *testing.T) {
	mat := DefaultHZO()

	if mat == nil {
		t.Fatal("DefaultHZO should return non-nil")
	}
	if mat.Name == "" {
		t.Error("Name should not be empty")
	}
	if mat.Pr <= 0 {
		t.Error("Pr should be positive")
	}
	if mat.Ps <= 0 {
		t.Error("Ps should be positive")
	}
	if mat.Ec <= 0 {
		t.Error("Ec should be positive")
	}
	if mat.NumLevels != 30 {
		t.Errorf("expected NumLevels=30, got %d", mat.NumLevels)
	}
}

func TestFeCIMMaterial(t *testing.T) {
	mat := FeCIMMaterial()

	if mat == nil {
		t.Fatal("FeCIMMaterial should return non-nil")
	}
	if mat.NumLevels != 30 {
		t.Errorf("FeCIM should have 30 levels, got %d", mat.NumLevels)
	}
	// Verified values from Dr. Tour's presentation
	if mat.EnduranceCycles != 1e9 {
		t.Errorf("FeCIM demonstrated endurance should be 1e9, got %e", mat.EnduranceCycles)
	}
}

func TestFeCIMMaterialTarget(t *testing.T) {
	mat := FeCIMMaterialTarget()

	if mat == nil {
		t.Fatal("FeCIMMaterialTarget should return non-nil")
	}
	// Target values
	if mat.EnduranceCycles != 1e12 {
		t.Errorf("FeCIM target endurance should be 1e12, got %e", mat.EnduranceCycles)
	}
}

func TestLiteratureSuperlattice(t *testing.T) {
	mat := LiteratureSuperlattice()

	if mat == nil {
		t.Fatal("LiteratureSuperlattice should return non-nil")
	}
	// Literature reports 64 states for superlattice
	if mat.NumLevels != 64 {
		t.Errorf("expected 64 levels, got %d", mat.NumLevels)
	}
}

func TestCryogenicHZO(t *testing.T) {
	mat := CryogenicHZO()

	if mat == nil {
		t.Fatal("CryogenicHZO should return non-nil")
	}
	// Cryogenic HZO has enhanced Pr (75 µC/cm²)
	if mat.Pr < 0.7 {
		t.Errorf("expected Pr >= 0.7 C/m², got %f", mat.Pr)
	}
}

func TestHZOStandard32(t *testing.T) {
	mat := HZOStandard32()

	if mat == nil {
		t.Fatal("HZOStandard32 should return non-nil")
	}
	if mat.NumLevels != 32 {
		t.Errorf("expected 32 levels, got %d", mat.NumLevels)
	}
}

func TestHZOFJT140(t *testing.T) {
	mat := HZOFJT140()

	if mat == nil {
		t.Fatal("HZOFJT140 should return non-nil")
	}
	if mat.NumLevels != 140 {
		t.Errorf("expected 140 levels, got %d", mat.NumLevels)
	}
	// FTJ has ultrathin thickness
	if mat.Thickness > 5e-9 {
		t.Errorf("FTJ should have ultrathin barrier, got %e", mat.Thickness)
	}
}

func TestPZT(t *testing.T) {
	mat := PZT()

	if mat == nil {
		t.Fatal("PZT should return non-nil")
	}
	if math.Abs(mat.Pr-0.30) > 1e-9 {
		t.Fatalf("PZT Pr: got %f, want 0.30 C/m²", mat.Pr)
	}
	if math.Abs(mat.Ps-0.40) > 1e-9 {
		t.Fatalf("PZT Ps: got %f, want 0.40 C/m²", mat.Ps)
	}
	if math.Abs(mat.Ec-6.0e6) > 1 {
		t.Fatalf("PZT Ec: got %e, want 6.0e6 V/m", mat.Ec)
	}
}

func TestAlScN(t *testing.T) {
	mat := AlScN()

	if mat == nil {
		t.Fatal("AlScN should return non-nil")
	}
	// AlScN has very high Pr
	if mat.Pr < 1.0 {
		t.Errorf("AlScN should have Pr >= 100 µC/cm², got %f C/m²", mat.Pr)
	}
	// AlScN has very high Ec which limits states
	if mat.NumLevels > 20 {
		t.Errorf("AlScN high Ec should limit states, got %d", mat.NumLevels)
	}
}

func TestAllMaterials(t *testing.T) {
	materials := AllMaterials()

	if len(materials) == 0 {
		t.Fatal("AllMaterials should return at least one material")
	}

	// Check all materials have valid names
	for i, mat := range materials {
		if mat.Name == "" {
			t.Errorf("material %d has empty name", i)
		}
		if mat.Pr <= 0 {
			t.Errorf("material %q has invalid Pr", mat.Name)
		}
	}
}

func TestCorePresetCount(t *testing.T) {
	presets := []*HZOMaterial{
		DefaultHZO(),
		FeCIMMaterial(),
		FeCIMMaterialTarget(),
		LiteratureSuperlattice(),
		CryogenicHZO(),
		HZOStandard32(),
		HZOFJT140(),
		PZT(),
		AlScN(),
	}
	if len(presets) != 9 {
		t.Fatalf("expected 9 core presets, got %d", len(presets))
	}
}

func TestHZOMaterial_GetNumLevels(t *testing.T) {
	tests := []struct {
		name      string
		numLevels int
		expect    int
	}{
		{"default value", 0, DefaultLevels},
		{"negative value", -1, DefaultLevels},
		{"explicit 30", 30, 30},
		{"explicit 64", 64, 64},
		{"explicit 140", 140, 140},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mat := &HZOMaterial{NumLevels: tt.numLevels}
			result := mat.GetNumLevels()
			if result != tt.expect {
				t.Errorf("GetNumLevels() = %d, expected %d", result, tt.expect)
			}
		})
	}
}

func TestHZOMaterial_CoerciveVoltage(t *testing.T) {
	mat := DefaultHZO()
	Vc := mat.CoerciveVoltage()

	// Vc = Ec * thickness
	expected := mat.Ec * mat.Thickness
	if math.Abs(Vc-expected) > 1e-15 {
		t.Errorf("CoerciveVoltage() = %e, expected %e", Vc, expected)
	}
}

func TestHZOMaterial_RemanentCharge(t *testing.T) {
	mat := DefaultHZO()
	Q := mat.RemanentCharge()

	if Q != mat.Pr {
		t.Errorf("RemanentCharge() = %e, expected %e", Q, mat.Pr)
	}
}

func TestHZOMaterial_Capacitance(t *testing.T) {
	mat := DefaultHZO()
	C := mat.Capacitance()

	if C <= 0 {
		t.Error("Capacitance should be positive")
	}

	// C = epsilon0 * epsilon * A / d
	epsilon0 := 8.854e-12
	expected := epsilon0 * mat.Epsilon * mat.Area / mat.Thickness
	if math.Abs(C-expected)/expected > 1e-10 {
		t.Errorf("Capacitance() = %e, expected %e", C, expected)
	}
}

func TestHZOMaterial_SwitchingEnergy(t *testing.T) {
	mat := DefaultHZO()
	E := mat.SwitchingEnergy()

	if E <= 0 {
		t.Error("SwitchingEnergy should be positive")
	}

	// E = 2 * Pr * Ec * Volume
	volume := mat.Area * mat.Thickness
	expected := 2 * mat.Pr * mat.Ec * volume
	if math.Abs(E-expected)/expected > 1e-10 {
		t.Errorf("SwitchingEnergy() = %e, expected %e", E, expected)
	}
}

func TestHZOMaterial_SwitchingTime(t *testing.T) {
	mat := DefaultHZO()

	// At room temperature (300K)
	tau300 := mat.SwitchingTime(300)
	if tau300 <= 0 {
		t.Error("SwitchingTime should be positive")
	}

	// At higher temperature, switching should be faster
	tau400 := mat.SwitchingTime(400)
	if tau400 >= tau300 {
		t.Error("SwitchingTime should decrease with temperature")
	}
}

func TestHZOMaterial_CoerciveFieldAtTemp(t *testing.T) {
	mat := DefaultHZO()

	// At room temperature
	Ec300 := mat.CoerciveFieldAtTemp(300)
	if Ec300 <= 0 {
		t.Error("CoerciveField should be positive below Curie temperature")
	}

	// At Curie temperature
	EcCurie := mat.CoerciveFieldAtTemp(mat.CurieTemp)
	if EcCurie != 0 {
		t.Error("CoerciveField should be zero at Curie temperature")
	}

	// Above Curie temperature
	EcAbove := mat.CoerciveFieldAtTemp(mat.CurieTemp + 100)
	if EcAbove != 0 {
		t.Error("CoerciveField should be zero above Curie temperature")
	}
}

func TestHZOMaterial_PolarizationAtTemp(t *testing.T) {
	mat := DefaultHZO()

	// At room temperature
	Pr300 := mat.PolarizationAtTemp(300)
	if Pr300 <= 0 {
		t.Error("Polarization should be positive below Curie temperature")
	}

	// At Curie temperature
	PrCurie := mat.PolarizationAtTemp(mat.CurieTemp)
	if PrCurie != 0 {
		t.Error("Polarization should be zero at Curie temperature")
	}
}

func TestHZOMaterial_EnduranceAtCycles(t *testing.T) {
	mat := DefaultHZO()

	// At zero cycles, should be full Pr
	Pr0 := mat.EnduranceAtCycles(0)
	if math.Abs(Pr0-mat.Pr) > 1e-15 {
		t.Errorf("at 0 cycles, Pr should be %e, got %e", mat.Pr, Pr0)
	}

	// At end of life, should be degraded
	PrEnd := mat.EnduranceAtCycles(mat.EnduranceCycles)
	if PrEnd >= mat.Pr {
		t.Error("Pr should degrade at end of endurance")
	}
	if PrEnd <= 0 {
		t.Error("Pr should not be zero at end of endurance")
	}
}

func TestHZOMaterial_RetentionAtTime(t *testing.T) {
	mat := DefaultHZO()

	// At t=0, should be full Pr
	Pr0 := mat.RetentionAtTime(0, 358) // 85°C reference
	if math.Abs(Pr0-mat.Pr) > 1e-15 {
		t.Errorf("at t=0, Pr should be %e, got %e", mat.Pr, Pr0)
	}

	// At end of retention, should be degraded
	PrEnd := mat.RetentionAtTime(mat.RetentionTime, 358)
	if PrEnd >= mat.Pr {
		t.Error("Pr should degrade at end of retention")
	}

	// Higher temperature should accelerate degradation
	PrHot := mat.RetentionAtTime(mat.RetentionTime/10, 400)
	PrCool := mat.RetentionAtTime(mat.RetentionTime/10, 300)
	if PrHot >= PrCool {
		t.Error("higher temperature should accelerate retention loss")
	}
}

func TestHZOMaterial_DiscreteLevel(t *testing.T) {
	mat := FeCIMMaterial() // Has Gmin and Gmax set

	// Test level 0 (lowest conductance)
	G0 := mat.DiscreteLevel(0, 30)
	if math.Abs(G0-mat.Gmin)/mat.Gmin > 0.01 {
		t.Errorf("level 0 should be Gmin (%e), got %e", mat.Gmin, G0)
	}

	// Test level 29 (highest conductance)
	G29 := mat.DiscreteLevel(29, 30)
	if math.Abs(G29-mat.Gmax)/mat.Gmax > 0.01 {
		t.Errorf("level 29 should be Gmax (%e), got %e", mat.Gmax, G29)
	}

	// Test middle level (should be between Gmin and Gmax)
	G15 := mat.DiscreteLevel(15, 30)
	if G15 <= mat.Gmin || G15 >= mat.Gmax {
		t.Errorf("level 15 should be between Gmin and Gmax, got %e", G15)
	}

	// Test single level (degenerate case)
	G1 := mat.DiscreteLevel(0, 1)
	expected := (mat.Gmin + mat.Gmax) / 2
	if math.Abs(G1-expected)/expected > 0.01 {
		t.Errorf("single level should be midpoint, got %e", G1)
	}
}

func TestHZOMaterial_DiscreteLevel_DefaultConductance(t *testing.T) {
	// Material without Gmin/Gmax set
	mat := &HZOMaterial{
		NumLevels: 30,
		Gmin:      0,
		Gmax:      0,
	}

	// Should use fallback values
	G0 := mat.DiscreteLevel(0, 30)
	G29 := mat.DiscreteLevel(29, 30)

	if G0 <= 0 {
		t.Error("should use fallback Gmin")
	}
	if G29 <= G0 {
		t.Error("Gmax should be greater than Gmin")
	}
}

func BenchmarkDiscreteLevel(b *testing.B) {
	mat := FeCIMMaterial()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mat.DiscreteLevel(i%30, 30)
	}
}

func BenchmarkAllMaterials(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = AllMaterials()
	}
}
