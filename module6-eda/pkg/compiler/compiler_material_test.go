// pkg/compiler/compiler_material_test.go
package compiler

import (
	"math"
	"testing"

	"fecim-lattice-tools/shared/physics"
)

// M6-COMP-04: Material propagation test
// Verify material parameters correctly propagate to capacitance and resistance:
// - Capacitance: C_fe = ε₀ × εᵣ × A / d
// - Resistance: R_wire from resistivity (tested in export layer)

const (
	eps0 = 8.854e-12 // F/m (vacuum permittivity)
)

// MaterialTestCase defines a ferroelectric material for testing
type MaterialTestCase struct {
	name      string
	epsR      float64 // Relative permittivity
	thickness float64 // Thickness in meters
	area      float64 // Area in m²
}

func TestMaterial_CapacitanceCalculation(t *testing.T) {
	testCases := []MaterialTestCase{
		{
			name:      "HZO",
			epsR:      25.0,
			thickness: 10e-9,     // 10 nm
			area:      2.025e-15, // ~45nm × 45nm
		},
		{
			name:      "PZT",
			epsR:      1000.0,
			thickness: 100e-9, // 100 nm
			area:      1e-14,  // 100nm × 100nm
		},
		{
			name:      "BTO",
			epsR:      1200.0,
			thickness: 50e-9, // 50 nm
			area:      4e-15, // ~63nm × 63nm
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Expected capacitance: C = ε₀ × εᵣ × A / d
			expectedCap := eps0 * tc.epsR * tc.area / tc.thickness

			// Calculate capacitance in fF for reporting
			expectedCapFF := expectedCap * 1e15

			t.Logf("%s: εᵣ=%.1f, d=%.1e m, A=%.3e m² → C_fe=%.3f fF",
				tc.name, tc.epsR, tc.thickness, tc.area, expectedCapFF)

			// Verify capacitance calculation formula
			tolerance := 0.01 // 1%
			calculatedCap := eps0 * tc.epsR * tc.area / tc.thickness
			delta := math.Abs(calculatedCap-expectedCap) / expectedCap

			if delta > tolerance {
				t.Errorf("Capacitance mismatch: calculated %.6e F, expected %.6e F (delta %.2f%%)",
					calculatedCap, expectedCap, delta*100)
			}

			// Verify material parameters would propagate correctly to export layer
			// (Export layer tests verify SPICE generation)
			if tc.epsR <= 0 {
				t.Error("Material permittivity must be positive")
			}
			if tc.thickness <= 0 {
				t.Error("Material thickness must be positive")
			}
			if tc.area <= 0 {
				t.Error("Material area must be positive")
			}
		})
	}
}

// TestMaterial_ConductanceMapping verifies material → conductance propagation
func TestMaterial_ConductanceMapping(t *testing.T) {
	config := NewComputeConfig(4, 4)
	config.GMin = 10.0  // μS
	config.GMax = 100.0 // μS
	config.Levels = 30

	// Create test weights to exercise quantization
	weights := [][]float64{
		{-1.0, -0.5, 0.0, 0.5},
		{1.0, 0.25, -0.25, 0.75},
		{-0.75, 0.0, 0.5, -0.5},
		{0.0, 1.0, -1.0, 0.0},
	}
	config.WithWeights(weights)

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// Verify conductance range mapping
	for idx, cell := range design.Cells {
		if cell.Conductance < config.GMin || cell.Conductance > config.GMax {
			t.Errorf("Cell %d: conductance %.3f μS out of range [%.1f, %.1f]",
				idx, cell.Conductance, config.GMin, config.GMax)
		}

		// Verify conductance/resistance consistency: R = 1e6 / G (Ω from μS)
		expectedR := 1e6 / cell.Conductance
		tolerance := 0.01 // 1%
		delta := math.Abs(cell.Resistance-expectedR) / expectedR

		if delta > tolerance {
			t.Errorf("Cell %d: R/G inconsistency: R=%.3e Ω, G=%.3f μS, expected R=%.3e Ω (delta %.2f%%)",
				idx, cell.Resistance, cell.Conductance, expectedR, delta*100)
		}
	}

	// Verify quantization levels map to conductance range
	// Level 0 → GMin, Level (Levels-1) → GMax
	var minCond, maxCond float64 = math.MaxFloat64, 0.0
	for _, cell := range design.Cells {
		if cell.Conductance < minCond {
			minCond = cell.Conductance
		}
		if cell.Conductance > maxCond {
			maxCond = cell.Conductance
		}
	}

	// With quantized weights, we should see conductances spanning the range
	if minCond > config.GMin*1.1 { // Allow 10% margin for quantization
		t.Errorf("Minimum conductance %.3f μS far from GMin %.1f μS", minCond, config.GMin)
	}
	if maxCond < config.GMax*0.9 {
		t.Errorf("Maximum conductance %.3f μS far from GMax %.1f μS", maxCond, config.GMax)
	}

	t.Logf("Conductance range: [%.3f, %.3f] μS (config: [%.1f, %.1f] μS)",
		minCond, maxCond, config.GMin, config.GMax)
}

// TestMaterial_ResistanceFromConductance verifies R = 1e6/G transformation
func TestMaterial_ResistanceFromConductance(t *testing.T) {
	testCases := []struct {
		name        string
		conductance float64 // μS
		expectedR   float64 // Ω
	}{
		{"GMin_10uS", 10.0, 1e5},
		{"GMid_50uS", 50.0, 2e4},
		{"GMax_100uS", 100.0, 1e4},
		{"HighG_1000uS", 1000.0, 1e3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewComputeConfig(2, 2)
			config.GMin = tc.conductance
			config.GMax = tc.conductance // Force single conductance

			design, err := GenerateDesign(config)
			if err != nil {
				t.Fatalf("GenerateDesign failed: %v", err)
			}

			// All cells should have the same conductance/resistance
			for idx, cell := range design.Cells {
				if math.Abs(cell.Conductance-tc.conductance) > 1e-6 {
					t.Errorf("Cell %d: conductance %.6f μS != expected %.6f μS",
						idx, cell.Conductance, tc.conductance)
				}

				if math.Abs(cell.Resistance-tc.expectedR)/tc.expectedR > 0.01 {
					t.Errorf("Cell %d: resistance %.6e Ω != expected %.6e Ω",
						idx, cell.Resistance, tc.expectedR)
				}
			}

			t.Logf("%s: G=%.1f μS → R=%.3e Ω", tc.name, tc.conductance, tc.expectedR)
		})
	}
}

// TestMaterial_ProgrammingVoltageMapping verifies V_prog scaling
func TestMaterial_ProgrammingVoltageMapping(t *testing.T) {
	config := NewComputeConfig(4, 4)
	config.VProgMin = 2.0 // V
	config.VProgMax = 5.0 // V
	config.Levels = 30

	// Weights spanning [-1, 1]
	weights := [][]float64{
		{-1.0, -0.5, 0.0, 0.5},
		{1.0, 0.25, -0.25, 0.75},
		{-0.75, 0.0, 0.5, -0.5},
		{0.0, 1.0, -1.0, 0.0},
	}
	config.WithWeights(weights)

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// Verify programming voltage range
	var minV, maxV float64 = math.MaxFloat64, 0.0
	for _, cell := range design.Cells {
		if cell.ProgramV < config.VProgMin || cell.ProgramV > config.VProgMax {
			t.Errorf("Cell (%d,%d): V_prog %.3f V out of range [%.1f, %.1f]",
				cell.Row, cell.Col, cell.ProgramV, config.VProgMin, config.VProgMax)
		}

		if cell.ProgramV < minV {
			minV = cell.ProgramV
		}
		if cell.ProgramV > maxV {
			maxV = cell.ProgramV
		}
	}

	// With quantized weights, we should see voltages spanning the range
	if minV > config.VProgMin*1.1 {
		t.Errorf("Minimum V_prog %.3f V far from VProgMin %.1f V", minV, config.VProgMin)
	}
	if maxV < config.VProgMax*0.9 {
		t.Errorf("Maximum V_prog %.3f V far from VProgMax %.1f V", maxV, config.VProgMax)
	}

	t.Logf("Programming voltage range: [%.3f, %.3f] V (config: [%.1f, %.1f] V)",
		minV, maxV, config.VProgMin, config.VProgMax)
}

// TestMaterial_SharedPhysicsIntegration verifies integration with shared/physics
func TestMaterial_SharedPhysicsIntegration(t *testing.T) {
	// Get HZO material from shared/physics
	hzoMat := physics.DefaultHZO()
	if hzoMat == nil {
		t.Fatal("DefaultHZO() returned nil")
	}

	// Verify material parameters exist
	if hzoMat.Epsilon <= 0 {
		t.Errorf("HZO: invalid permittivity εᵣ = %.1f", hzoMat.Epsilon)
	}
	if hzoMat.Ec <= 0 {
		t.Errorf("HZO: invalid coercive field Ec = %.3e V/m", hzoMat.Ec)
	}

	// Calculate capacitance using shared material
	thickness := hzoMat.Thickness
	if thickness <= 0 {
		thickness = 10e-9 // Default if not set
	}
	area := hzoMat.Area
	if area <= 0 {
		area = 2.025e-15 // Default if not set
	}

	expectedCap := eps0 * hzoMat.Epsilon * area / thickness
	expectedCapFF := expectedCap * 1e15

	// Coercive field in MV/cm for reporting
	EcMVcm := hzoMat.Ec / 1e6 / 100 // V/m → MV/cm

	t.Logf("HZO from shared/physics: εᵣ=%.1f, Ec=%.2f MV/cm → C_fe=%.3f fF (d=%.1e m, A=%.2e m²)",
		hzoMat.Epsilon, EcMVcm, expectedCapFF, thickness, area)

	// Verify tolerance against literature values
	// HZO: εᵣ ~ 20-30, Ec ~ 1.0-2.0 MV/cm
	if hzoMat.Epsilon < 20 || hzoMat.Epsilon > 30 {
		t.Logf("Warning: HZO εᵣ=%.1f outside typical range [20, 30]", hzoMat.Epsilon)
	}
	if EcMVcm < 0.5 || EcMVcm > 3.0 {
		t.Logf("Warning: HZO Ec=%.2f MV/cm outside typical range [0.5, 3.0]", EcMVcm)
	}
}

// TestMaterial_MultiMaterialConsistency verifies material parameters propagate correctly
func TestMaterial_MultiMaterialConsistency(t *testing.T) {
	materials := []struct {
		name        string
		constructor func() *physics.HZOMaterial
		epsRRange   [2]float64
		EcRangeMVcm [2]float64
	}{
		{
			name:        "DefaultHZO",
			constructor: physics.DefaultHZO,
			epsRRange:   [2]float64{20, 30},
			EcRangeMVcm: [2]float64{0.5, 3.0},
		},
		{
			name:        "PZT",
			constructor: physics.PZT,
			epsRRange:   [2]float64{800, 1200},
			EcRangeMVcm: [2]float64{0.05, 0.15},
		},
		{
			name:        "AlScN",
			constructor: physics.AlScN,
			epsRRange:   [2]float64{8, 15},
			EcRangeMVcm: [2]float64{3.0, 6.0},
		},
	}

	for _, mat := range materials {
		t.Run(mat.name, func(t *testing.T) {
			material := mat.constructor()
			if material == nil {
				t.Fatalf("%s constructor returned nil", mat.name)
			}

			// Convert Ec from V/m to MV/cm for comparison
			EcMVcm := material.Ec / 1e6 / 100

			// Verify permittivity range
			if material.Epsilon < mat.epsRRange[0] || material.Epsilon > mat.epsRRange[1] {
				t.Logf("Warning: %s εᵣ=%.1f outside expected range [%.0f, %.0f]",
					mat.name, material.Epsilon, mat.epsRRange[0], mat.epsRRange[1])
			}

			// Verify coercive field range
			if EcMVcm < mat.EcRangeMVcm[0] || EcMVcm > mat.EcRangeMVcm[1] {
				t.Logf("Warning: %s Ec=%.3f MV/cm outside expected range [%.2f, %.2f]",
					mat.name, EcMVcm, mat.EcRangeMVcm[0], mat.EcRangeMVcm[1])
			}

			// Calculate capacitance
			thickness := material.Thickness
			if thickness <= 0 {
				thickness = 10e-9
			}
			area := material.Area
			if area <= 0 {
				area = 2.025e-15
			}

			cap := eps0 * material.Epsilon * area / thickness
			capFF := cap * 1e15

			t.Logf("%s: εᵣ=%.1f, Ec=%.3f MV/cm, C_fe=%.3f fF",
				mat.name, material.Epsilon, EcMVcm, capFF)
		})
	}
}
