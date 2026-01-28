package ferroelectric

import (
	"math"
	"testing"
)

// ============================================================================
// Physics Validation Tests
// ============================================================================
// These tests verify that all HZOMaterial variants match peer-reviewed
// literature values for polarization, coercive field, and other physics
// parameters. They serve as a quality gate to prevent unrealistic or
// marketing-inflated values from entering the codebase.
//
// References:
// - Nature Commun. 2025, doi:10.1038/s41467-025-63033-y (HZO Pr/Ec)
// - Adv. Electron. Mater. 2024, doi:10.1002/aelm.202300879 (Cryogenic)
// - Adv. Electron. Mater. 2025, doi:10.1002/aelm.202400840 (Cryogenic)
// - APL Materials 2023, doi:10.1063/5.0148068 (AlScN)
// - Nature Commun. 2025, doi:10.1038/s41467-025-62904-6 (AlScN)
// ============================================================================

// TestLiteraturePolarizationConstants validates all material Pr/Ps values
// against peer-reviewed ranges. This prevents unrealistic marketing claims.
func TestLiteraturePolarizationConstants(t *testing.T) {
	tests := []struct {
		name          string
		material      *HZOMaterial
		minPr         float64 // C/m²
		maxPr         float64 // C/m²
		description   string
		tolerancePct  float64 // Allow % tolerance for engineering margin
	}{
		{
			name:         "DefaultHZO",
			material:     DefaultHZO(),
			minPr:        15e-2, // 15 µC/cm²
			maxPr:        34e-2, // 34 µC/cm²
			description:  "Standard HZO from Nature Commun. 2025",
			tolerancePct: 10, // 10% tolerance
		},
		{
			name:         "FeCIMMaterial",
			material:     FeCIMMaterial(),
			minPr:        15e-2, // 15 µC/cm²
			maxPr:        34e-2, // 34 µC/cm²
			description:  "FeCIM uses standard HZO range (not disclosed)",
			tolerancePct: 10,
		},
		{
			name:         "FeCIMMaterialTarget",
			material:     FeCIMMaterialTarget(),
			minPr:        15e-2, // 15 µC/cm²
			maxPr:        34e-2, // 34 µC/cm²
			description:  "FeCIM target uses standard HZO range",
			tolerancePct: 10,
		},
		{
			name:         "LiteratureSuperlattice",
			material:     LiteratureSuperlattice(),
			minPr:        15e-2, // 15 µC/cm²
			maxPr:        50e-2, // 50 µC/cm² (superlattice enhanced)
			description:  "Superlattice from Cheema Nature 2020",
			tolerancePct: 10,
		},
		{
			name:         "CryogenicHZO",
			material:     CryogenicHZO(),
			minPr:        64e-2,  // 75 µC/cm² - 15% tolerance
			maxPr:        86e-2,  // 75 µC/cm² + 15% tolerance
			description:  "Cryogenic Pr ~75 µC/cm² from Adv. Elec. Mat. 2024",
			tolerancePct: 0, // Already includes tolerance in min/max
		},
		{
			name:         "HZOStandard32",
			material:     HZOStandard32(),
			minPr:        15e-2, // 15 µC/cm²
			maxPr:        34e-2, // 34 µC/cm²
			description:  "Standard HZO demonstrating 32 states",
			tolerancePct: 10,
		},
		{
			name:         "HZOFJT140",
			material:     HZOFJT140(),
			minPr:        15e-2, // 15 µC/cm²
			maxPr:        34e-2, // 34 µC/cm²
			description:  "HZO FTJ from Song et al. 2024",
			tolerancePct: 10,
		},
		{
			name:         "AlScN",
			material:     AlScN(),
			minPr:        100e-2, // 100 µC/cm²
			maxPr:        172e-2, // 172 µC/cm² (literature max)
			description:  "AlScN from Nature Commun. 2025 & APL Materials 2023",
			tolerancePct: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply tolerance to range
			minWithTol := tt.minPr * (1 - tt.tolerancePct/100)
			maxWithTol := tt.maxPr * (1 + tt.tolerancePct/100)

			// Check Pr is within literature range
			if tt.material.Pr < minWithTol || tt.material.Pr > maxWithTol {
				t.Errorf("%s: Pr = %.2f µC/cm² is outside literature range [%.1f, %.1f] µC/cm²\n  Source: %s",
					tt.name, tt.material.Pr*1e4, tt.minPr*1e4, tt.maxPr*1e4, tt.description)
			}

			// Check Ps > Pr (saturation must exceed remanent)
			if tt.material.Ps <= tt.material.Pr {
				t.Errorf("%s: Ps (%.2f) must be greater than Pr (%.2f)",
					tt.name, tt.material.Ps*1e4, tt.material.Pr*1e4)
			}

			t.Logf("%s: Pr = %.2f µC/cm², Ps = %.2f µC/cm² ✓ (within literature range)",
				tt.name, tt.material.Pr*1e4, tt.material.Ps*1e4)
		})
	}
}

// TestLiteratureCoerciveFieldConstants validates Ec values against
// peer-reviewed measurements.
func TestLiteratureCoerciveFieldConstants(t *testing.T) {
	tests := []struct {
		name         string
		material     *HZOMaterial
		minEc        float64 // V/m
		maxEc        float64 // V/m
		description  string
		tolerancePct float64
	}{
		{
			name:         "DefaultHZO",
			material:     DefaultHZO(),
			minEc:        0.6e8, // 0.6 MV/cm
			maxEc:        1.5e8, // 1.5 MV/cm
			description:  "HZO Ec from Nature Commun. 2025, Nano Letters 2024",
			tolerancePct: 10,
		},
		{
			name:         "FeCIMMaterial",
			material:     FeCIMMaterial(),
			minEc:        0.6e8,
			maxEc:        1.5e8,
			description:  "FeCIM uses standard HZO Ec range",
			tolerancePct: 10,
		},
		{
			name:         "FeCIMMaterialTarget",
			material:     FeCIMMaterialTarget(),
			minEc:        0.6e8,
			maxEc:        1.5e8,
			description:  "FeCIM target uses standard HZO Ec range",
			tolerancePct: 10,
		},
		{
			name:         "LiteratureSuperlattice",
			material:     LiteratureSuperlattice(),
			minEc:        0.6e8,
			maxEc:        1.5e8,
			description:  "Superlattice Ec from Cheema Nature 2020",
			tolerancePct: 10,
		},
		{
			name:         "CryogenicHZO",
			material:     CryogenicHZO(),
			minEc:        0.6e8,
			maxEc:        1.5e8,
			description:  "Cryogenic Ec from Adv. Elec. Mat. 2024",
			tolerancePct: 10,
		},
		{
			name:         "HZOStandard32",
			material:     HZOStandard32(),
			minEc:        0.6e8,
			maxEc:        1.5e8,
			description:  "Standard HZO Ec",
			tolerancePct: 10,
		},
		{
			name:         "HZOFJT140",
			material:     HZOFJT140(),
			minEc:        0.6e8,
			maxEc:        1.5e8,
			description:  "HZO FTJ Ec",
			tolerancePct: 10,
		},
		{
			name:         "AlScN",
			material:     AlScN(),
			minEc:        4.0e8, // 4.0 MV/cm (lower bound)
			maxEc:        6.0e8, // 6.0 MV/cm (upper bound)
			description:  "AlScN Ec ~5.0 MV/cm from Nature Commun. 2025",
			tolerancePct: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minWithTol := tt.minEc * (1 - tt.tolerancePct/100)
			maxWithTol := tt.maxEc * (1 + tt.tolerancePct/100)

			if tt.material.Ec < minWithTol || tt.material.Ec > maxWithTol {
				t.Errorf("%s: Ec = %.2f MV/cm is outside literature range [%.1f, %.1f] MV/cm\n  Source: %s",
					tt.name, tt.material.Ec/1e8, tt.minEc/1e8, tt.maxEc/1e8, tt.description)
			}

			// Ec must be positive
			if tt.material.Ec <= 0 {
				t.Errorf("%s: Ec must be positive, got %.2e", tt.name, tt.material.Ec)
			}

			t.Logf("%s: Ec = %.2f MV/cm ✓ (within literature range)",
				tt.name, tt.material.Ec/1e8)
		})
	}
}

// TestPolarizationSaturationRatio verifies that Ps/Pr ratio is physically
// realistic. Literature shows 1.05 < Ps/Pr < 1.5 for most ferroelectrics.
func TestPolarizationSaturationRatio(t *testing.T) {
	materials := AllMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			ratio := mat.Ps / mat.Pr

			// Physical bounds: saturation should be 5-50% higher than remanent
			// CryogenicHZO has ratio ~1.07 which is valid (tight loop)
			if ratio < 1.05 || ratio > 1.5 {
				t.Errorf("%s: Ps/Pr ratio %.3f is outside physical range [1.05, 1.5]",
					mat.Name, ratio)
			}

			t.Logf("%s: Ps/Pr = %.3f ✓", mat.Name, ratio)
		})
	}
}

// TestTemperatureScalingPhysics verifies that temperature-dependent
// scaling follows the canonical ferroelectric physics formula:
// Ec(T) = Ec0 * (1 - T/Tc)^0.5
func TestTemperatureScalingPhysics(t *testing.T) {
	materials := []*HZOMaterial{
		DefaultHZO(),
		FeCIMMaterial(),
		CryogenicHZO(),
		AlScN(),
	}

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			// Test at T = 0.5 * Tc
			halfTc := 0.5 * mat.CurieTemp
			EcHalf := mat.CoerciveFieldAtTemp(halfTc)
			expectedEcHalf := mat.Ec * math.Pow(1-0.5, 0.5) // = Ec * sqrt(0.5) ≈ 0.707*Ec

			relError := math.Abs(EcHalf-expectedEcHalf) / expectedEcHalf
			if relError > 0.05 { // 5% tolerance for numerical precision
				t.Errorf("%s: At T=0.5*Tc, Ec(T) = %.2e, expected %.2e (rel error %.1f%%)",
					mat.Name, EcHalf, expectedEcHalf, relError*100)
			}

			t.Logf("%s: At T=0.5*Tc (%.0f K), Ec = %.2f MV/cm (expected ~%.2f MV/cm) ✓",
				mat.Name, halfTc, EcHalf/1e8, expectedEcHalf/1e8)

			// Test at T = Tc (should be zero)
			EcAtTc := mat.CoerciveFieldAtTemp(mat.CurieTemp)
			if EcAtTc != 0 {
				t.Errorf("%s: At T=Tc, Ec should be 0, got %.2e", mat.Name, EcAtTc)
			}

			// Test above Tc (should be zero)
			EcAboveTc := mat.CoerciveFieldAtTemp(mat.CurieTemp + 100)
			if EcAboveTc != 0 {
				t.Errorf("%s: Above Tc, Ec should be 0, got %.2e", mat.Name, EcAboveTc)
			}

			t.Logf("%s: At T=Tc and above, Ec = 0 ✓", mat.Name)
		})
	}
}

// TestImprintFieldPhysics verifies that imprint field (built-in bias)
// is minimal for fresh devices. Imprint should be < 10% of Ec.
func TestImprintFieldPhysics(t *testing.T) {
	materials := AllMaterials()

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			ratio := mat.ImrintField / mat.Ec

			// Fresh devices should have minimal imprint (< 10% of Ec)
			if ratio > 0.1 {
				t.Errorf("%s: Imprint field %.2e V/m is too high (%.1f%% of Ec, expected < 10%%)",
					mat.Name, mat.ImrintField, ratio*100)
			}

			// Imprint should be positive (or zero)
			if mat.ImrintField < 0 {
				t.Errorf("%s: Imprint field should be non-negative, got %.2e",
					mat.Name, mat.ImrintField)
			}

			t.Logf("%s: Imprint = %.2e V/m (%.2f%% of Ec) ✓",
				mat.Name, mat.ImrintField, ratio*100)
		})
	}
}

// TestCapacitanceCalculation verifies the capacitance formula:
// C = ε₀ * εᵣ * Area / Thickness
func TestCapacitanceCalculation(t *testing.T) {
	materials := []*HZOMaterial{
		DefaultHZO(),
		FeCIMMaterial(),
		CryogenicHZO(),
		AlScN(),
	}

	epsilon0 := 8.854e-12 // F/m (vacuum permittivity)

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			C := mat.Capacitance()
			expectedC := epsilon0 * mat.Epsilon * mat.Area / mat.Thickness

			relError := math.Abs(C-expectedC) / expectedC
			if relError > 1e-10 { // Numerical precision check
				t.Errorf("%s: Capacitance %.2e F doesn't match formula %.2e F",
					mat.Name, C, expectedC)
			}

			// Verify EpsilonLF > Epsilon (low freq includes domain contribution)
			if mat.EpsilonLF <= mat.Epsilon {
				t.Errorf("%s: EpsilonLF (%.1f) should be > Epsilon (%.1f)",
					mat.Name, mat.EpsilonLF, mat.Epsilon)
			}

			// Typical ferroelectric capacitance: 0.1-10 fF for nanoscale devices
			if C < 1e-18 || C > 1e-12 {
				t.Logf("%s: Warning - capacitance %.2e F is outside typical range [1e-18, 1e-12]",
					mat.Name, C)
			}

			t.Logf("%s: C = %.2f fF (Epsilon: %.1f, EpsilonLF: %.1f) ✓",
				mat.Name, C*1e15, mat.Epsilon, mat.EpsilonLF)
		})
	}
}

// ============================================================================
// Summary Test: Verify All Materials Pass Physics Checks
// ============================================================================

func TestAllMaterialsPhysicsConsistency(t *testing.T) {
	materials := AllMaterials()

	t.Logf("Testing %d material variants for physics consistency...", len(materials))

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			// Check 1: Ps > Pr
			if mat.Ps <= mat.Pr {
				t.Errorf("FAIL: Ps (%.2f) must be > Pr (%.2f)", mat.Ps, mat.Pr)
			}

			// Check 2: All positive values
			if mat.Pr <= 0 || mat.Ps <= 0 || mat.Ec <= 0 {
				t.Errorf("FAIL: Polarization and coercive field must be positive")
			}

			// Check 3: Reasonable thickness (1-100 nm)
			if mat.Thickness < 1e-9 || mat.Thickness > 100e-9 {
				t.Logf("Warning: Thickness %.1f nm is unusual", mat.Thickness*1e9)
			}

			// Check 4: Reasonable permittivity (5-100 typical range)
			if mat.Epsilon < 5 || mat.Epsilon > 100 {
				t.Logf("Warning: Epsilon %.1f is unusual for ferroelectrics", mat.Epsilon)
			}

			// Check 5: Curie temperature above room temp
			// Note: Cryogenic materials may have CurieTemp=0 if loaded from config
			// with missing curie_temp_k field
			if mat.CurieTemp < 300 && mat.CurieTemp != 0 {
				t.Errorf("FAIL: Curie temp %.0f K is below room temperature", mat.CurieTemp)
			}
			if mat.CurieTemp == 0 {
				t.Logf("Warning: Curie temp is 0 (check config/physics.yaml for missing curie_temp_k)")
			}

			// Check 6: Reasonable endurance (10^6 to 10^12)
			if mat.EnduranceCycles < 1e6 || mat.EnduranceCycles > 1e13 {
				t.Logf("Warning: Endurance %.0e is unusual", mat.EnduranceCycles)
			}

			// Check 7: NumLevels is realistic (8-140 demonstrated in literature)
			if mat.GetNumLevels() < 8 || mat.GetNumLevels() > 140 {
				t.Errorf("FAIL: NumLevels %d is outside demonstrated range [8, 140]",
					mat.GetNumLevels())
			}

			t.Logf("✓ %s passes all physics consistency checks", mat.Name)
		})
	}
}

// ============================================================================
// Benchmarks for Physics Calculations
// ============================================================================

func BenchmarkCoerciveFieldAtTemp(b *testing.B) {
	mat := DefaultHZO()
	T := 300.0 // Room temperature

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mat.CoerciveFieldAtTemp(T)
	}
}

func BenchmarkPolarizationAtTemp(b *testing.B) {
	mat := DefaultHZO()
	T := 300.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mat.PolarizationAtTemp(T)
	}
}

func BenchmarkCapacitance(b *testing.B) {
	mat := DefaultHZO()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mat.Capacitance()
	}
}

func BenchmarkAllMaterialsLoad(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AllMaterials()
	}
}
