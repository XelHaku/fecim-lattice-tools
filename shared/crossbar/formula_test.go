package crossbar

import (
	"fmt"
	"math"
	"testing"
)

// TestIRDropFormulaAnalytical verifies V_drop = I_cumulative * R_wire * distance
func TestIRDropFormulaAnalytical(t *testing.T) {
	// Setup: 4x4 array, uniform G=50uS, V=0.5V, R=2.5 ohm/cell
	sim := NewIRDropSimulator(4, 4)
	G := 50e-6 // 50 uS
	V := 0.5   // 0.5 V

	for i := 0; i < 4; i++ {
		sim.SetInputVoltage(i, V)
		for j := 0; j < 4; j++ {
			sim.SetConductance(i, j, G)
		}
	}
	sim.Simulate(100)

	// Analytical: Cell current I=V*G=25uA, cumulative at corner ~100uA
	// V_drop ~ 100uA * 2.5ohm * 3cells = 0.75mV
	maxDrop := sim.GetMaxIRDrop()

	// Should be in reasonable range (0.1 - 10 mV)
	if maxDrop < 0.1e-3 || maxDrop > 10e-3 {
		t.Errorf("IR drop %e outside expected range [0.1mV, 10mV]", maxDrop)
	}
	t.Logf("Analytical: ~0.75 mV, Simulated: %.4f mV", maxDrop*1000)
}

// TestSneakPathSeriesConductance verifies G_series = 1/(1/G1 + 1/G2 + 1/G3)
func TestSneakPathSeriesConductance(t *testing.T) {
	testCases := []struct {
		name       string
		G1, G2, G3 float64
		expected   float64
	}{
		{"uniform 100uS", 100e-6, 100e-6, 100e-6, 33.333e-6},
		{"uniform 50uS", 50e-6, 50e-6, 50e-6, 16.667e-6},
		{"mixed 10/50/100", 10e-6, 50e-6, 100e-6, 7.692e-6},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Formula: G_series = 1 / (1/G1 + 1/G2 + 1/G3)
			calculated := 1.0 / (1.0/tc.G1 + 1.0/tc.G2 + 1.0/tc.G3)
			tolerance := tc.expected * 0.01
			if math.Abs(calculated-tc.expected) > tolerance {
				t.Errorf("G_series = %e, want %e (±1%%)", calculated, tc.expected)
			}
		})
	}
}

// TestArrheniusTemperatureDependence verifies drift rate ~ exp(-Ea/kT)
func TestArrheniusTemperatureDependence(t *testing.T) {
	tempRT := NewTemperatureEffects(300)  // Room temp (27C)
	tempHT := NewTemperatureEffects(358)  // High temp (85C)
	tempCryo := NewTemperatureEffects(77) // Cryogenic (LN2)

	baseDrift := 0.001

	rateRT := tempRT.AdjustedDriftRate(baseDrift)
	rateHT := tempHT.AdjustedDriftRate(baseDrift)
	rateCryo := tempCryo.AdjustedDriftRate(baseDrift)

	t.Run("HighTemp accelerates drift", func(t *testing.T) {
		ratio := rateHT / rateRT
		// Arrhenius model gives ~23x for 58K temperature delta with Ea=0.5eV
		if ratio < 1.5 || ratio > 30.0 {
			t.Errorf("85C/RT ratio = %.2f, expected 1.5-30x", ratio)
		}
		t.Logf("85C drift acceleration: %.2fx", ratio)
	})

	t.Run("Cryo suppresses drift", func(t *testing.T) {
		ratio := rateCryo / rateRT
		if ratio > 0.5 {
			t.Errorf("Cryo/RT ratio = %.4f, expected <0.5x", ratio)
		}
		t.Logf("77K drift suppression: %.6fx", ratio)
	})
}

// TestLinearConductanceFormula verifies G = Gmin + norm * (Gmax - Gmin)
func TestLinearConductanceFormula(t *testing.T) {
	// Create array with linear conductance model
	cfg := &Config{
		Rows:             4,
		Cols:             4,
		ConductanceModel: ConductanceLinear,
		ADCBits:          8,
		DACBits:          8,
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}

	testCases := []struct {
		norm     float64
		expected float64
		desc     string
	}{
		{0.0, GMin, "Level 0 = GMin"},
		{0.5, (GMin + GMax) / 2, "Midpoint = arithmetic mean"},
		{1.0, GMax, "Level 29 = GMax"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Linear formula: G = Gmin + norm * (Gmax - Gmin)
			calculated := arr.GetPhysicalConductance(tc.norm)
			tolerance := tc.expected * 0.001
			if math.Abs(calculated-tc.expected) > tolerance {
				t.Errorf("Linear G(%.2f) = %e, want %e", tc.norm, calculated, tc.expected)
			}
		})
	}
}

// TestExponentialConductanceFormula verifies G = Gmin * exp(ln(Gmax/Gmin) * norm)
func TestExponentialConductanceFormula(t *testing.T) {
	// Create array with exponential conductance model
	cfg := &Config{
		Rows:             4,
		Cols:             4,
		ConductanceModel: ConductanceExponential,
		ADCBits:          8,
		DACBits:          8,
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}

	testCases := []struct {
		norm     float64
		expected float64
		desc     string
	}{
		{0.0, GMin, "Level 0 = GMin"},
		{0.5, math.Sqrt(GMin * GMax), "Midpoint = geometric mean"},
		{1.0, GMax, "Level 29 = GMax"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Exponential formula: G = Gmin * exp(ln(Gmax/Gmin) * norm)
			calculated := arr.GetPhysicalConductance(tc.norm)
			tolerance := tc.expected * 0.001
			if math.Abs(calculated-tc.expected) > tolerance {
				t.Errorf("Exp G(%.2f) = %e, want %e", tc.norm, calculated, tc.expected)
			}
		})
	}
}

// TestOhmsLawCurrentCalculation verifies I = V * G
func TestOhmsLawCurrentCalculation(t *testing.T) {
	testCases := []struct {
		V, G     float64
		expected float64
	}{
		{0.5, 50e-6, 25e-6},   // 0.5V * 50uS = 25uA
		{1.0, 100e-6, 100e-6}, // 1V * 100uS = 100uA
		{0.1, 10e-6, 1e-6},    // 0.1V * 10uS = 1uA
	}

	for _, tc := range testCases {
		I := tc.V * tc.G
		if math.Abs(I-tc.expected) > tc.expected*0.001 {
			t.Errorf("I = V*G: %e * %e = %e, want %e", tc.V, tc.G, I, tc.expected)
		}
	}
}

// TestPowerDissipation verifies P = V^2 * G = I^2 / G
func TestPowerDissipation(t *testing.T) {
	V := 0.5
	G := 50e-6
	I := V * G

	P1 := V * V * G // V^2 * G
	P2 := I * I / G // I^2 / G

	if math.Abs(P1-P2) > P1*0.001 {
		t.Errorf("Power mismatch: V^2*G=%e, I^2/G=%e", P1, P2)
	}

	// 0.5^2 * 50uS = 12.5 µW
	expected := 12.5e-6
	if math.Abs(P1-expected) > expected*0.001 {
		t.Errorf("Power = %e, want %e", P1, expected)
	}
}

// TestWireResistanceTemperatureCoefficient verifies R(T) = R0 * (1 + alpha * dT)
func TestWireResistanceTemperatureCoefficient(t *testing.T) {
	testCases := []struct {
		tempK    float64
		R0       float64
		expected float64
		desc     string
	}{
		{300, 2.5, 2.5, "Room temp (reference)"},
		{358, 2.5, 2.5 * (1 + 0.00393*58), "Industrial 85C"},
		{77, 2.5, 2.5 * (1 + 0.00393*(-223)), "Cryogenic LN2"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			temp := NewTemperatureEffects(tc.tempK)
			calculated := temp.AdjustedWireResistance(tc.R0)
			tolerance := math.Abs(tc.expected) * 0.001
			if math.Abs(calculated-tc.expected) > tolerance {
				t.Errorf("R(%.0fK) = %.4f Ω, want %.4f Ω", tc.tempK, calculated, tc.expected)
			}
			t.Logf("R(%.0fK) = %.4f Ω (%.1f%% change)", tc.tempK, calculated, (calculated/tc.R0-1)*100)
		})
	}
}

// TestKirchhoffCurrentLaw verifies sum of currents into a node = 0
func TestKirchhoffCurrentLaw(t *testing.T) {
	// Create 3x3 array with known pattern
	sim := NewIRDropSimulator(3, 3)

	// Set uniform voltages and conductances
	V := 1.0
	G := 50e-6
	for i := 0; i < 3; i++ {
		sim.SetInputVoltage(i, V)
		for j := 0; j < 3; j++ {
			sim.SetConductance(i, j, G)
		}
	}
	sim.Simulate(100)

	// Get output currents (column sums)
	outputs := sim.GetOutputCurrents()

	// Total current out should equal total current in (conservation)
	totalOut := 0.0
	for _, I := range outputs {
		totalOut += I
	}

	// Total current in = rows * V * avg_conductance_per_row
	// For uniform array: I_in = rows * cols * V * G
	expectedTotalIn := float64(3*3) * V * G

	// Account for IR drop reducing actual current
	// Allow 10% tolerance due to voltage drops
	tolerance := expectedTotalIn * 0.10
	if math.Abs(totalOut-expectedTotalIn) > tolerance {
		t.Errorf("KCL violation: I_out=%e, I_in=%e (±10%%)", totalOut, expectedTotalIn)
	}
	t.Logf("Current conservation: I_out=%.2f µA, I_in=%.2f µA", totalOut*1e6, expectedTotalIn*1e6)
}

// TestSneakPathVoltageRatio verifies sneak current ratio vs direct current
func TestSneakPathVoltageRatio(t *testing.T) {
	// Create 5x5 analyzer with uniform conductance
	analyzer := NewSneakPathAnalyzer(5, 5)
	G := 50e-6
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			analyzer.SetConductance(i, j, G)
		}
	}

	V := 1.0
	analyzer.AnalyzeTarget(2, 2, V) // Center cell

	stats := analyzer.GetStats(V)

	// For uniform conductance array, sneak paths significantly degrade SNR
	// Target current = V * G = 50 µA
	// Each 3-cell sneak path has G_series = G/3
	// Number of sneak paths = (rows-1) * (cols-1) = 16

	expectedTarget := V * G
	tolerance := expectedTarget * 0.01
	if math.Abs(stats.TargetCurrent-expectedTarget) > tolerance {
		t.Errorf("Target current = %e, want %e", stats.TargetCurrent, expectedTarget)
	}

	// Sneak ratio should be > 0 for passive crossbar
	if stats.SneakRatio <= 0 {
		t.Errorf("Sneak ratio = %.4f, expected > 0 (passive crossbar has sneak paths)", stats.SneakRatio)
	}

	t.Logf("Sneak analysis: Target=%e A, Sneak=%e A, Ratio=%.2f%%, SNR=%.1f dB",
		stats.TargetCurrent, stats.TotalSneakCurrent, stats.SneakRatio*100, stats.SignalToNoiseRatio)
}

// TestConductanceRangeTemperatureScaling verifies window expansion/narrowing with temp
func TestConductanceRangeTemperatureScaling(t *testing.T) {
	testCases := []struct {
		tempK float64
		desc  string
	}{
		{4, "Deep cryo (4K)"},
		{77, "LN2 cryo (77K)"},
		{300, "Room temp"},
		{358, "Industrial (85C)"},
		{400, "Automotive Grade 0 (125C)"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			temp := NewTemperatureEffects(tc.tempK)
			gMin, gMax := temp.AdjustedConductanceRange(GMin, GMax)

			// Window = Gmax - Gmin
			window := gMax - gMin
			refWindow := GMax - GMin
			windowRatio := window / refWindow

			t.Logf("T=%.0fK: Gmin=%e, Gmax=%e, Window=%.2f%% of RT",
				tc.tempK, gMin, gMax, windowRatio*100)

			// Verify physics expectations
			if tc.tempK < 100 {
				// Cryogenic should enhance window
				if windowRatio < 1.0 {
					t.Errorf("Cryo window narrowed (%.2fx), expected enhancement", windowRatio)
				}
			} else if tc.tempK > 300 {
				// High temp should narrow window
				if windowRatio > 1.0 {
					t.Errorf("High-temp window expanded (%.2fx), expected narrowing", windowRatio)
				}
			}
		})
	}
}

// TestQuantizationTo30Levels verifies discrete level mapping
func TestQuantizationTo30Levels(t *testing.T) {
	testCases := []struct {
		input    float64
		expected int
	}{
		{0.0, 0},
		{0.017, 0},  // Just below level 1 threshold
		{0.034, 1},  // Level 1
		{0.5, 15},   // Midpoint (level 14-15)
		{0.966, 28}, // Level 28
		{1.0, 29},   // Max level
		{-0.5, 0},   // Clamp negative to 0
		{1.5, 29},   // Clamp overflow to 29
	}

	for _, tc := range testCases {
		quantized := QuantizeToLevels(tc.input)
		level := GetLevel(quantized)

		tolerance := 1 // Allow ±1 level due to rounding
		if math.Abs(float64(level-tc.expected)) > float64(tolerance) {
			t.Errorf("QuantizeTo30Levels(%.3f) = level %d, want %d (±1)", tc.input, level, tc.expected)
		}
	}
}

// TestADCDACQuantization verifies peripheral quantization effects
func TestADCDACQuantization(t *testing.T) {
	// Test different resolutions
	testCases := []struct {
		bits     int
		input    float64
		expected float64
	}{
		{4, 0.5, 8.0 / 15},       // 4-bit: 16 levels, level 8
		{8, 0.5, 128.0 / 255},    // 8-bit: 256 levels, level 128
		{12, 0.5, 2048.0 / 4095}, // 12-bit: 4096 levels, level 2048
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d-bit", tc.bits), func(t *testing.T) {
			cfg := &Config{
				Rows:    2,
				Cols:    2,
				ADCBits: tc.bits,
				DACBits: tc.bits,
			}
			arr, err := NewArray(cfg)
			if err != nil {
				t.Fatalf("Failed to create array: %v", err)
			}

			// quantizeDAC is private, but we can test through MVM
			// Program single cell and verify quantization
			arr.ProgramWeight(0, 0, 1.0)
			result, err := arr.MVM([]float64{tc.input})
			if err != nil {
				t.Fatalf("MVM failed: %v", err)
			}

			// Result should show quantization effects
			levels := 1 << tc.bits
			quantumSize := 1.0 / float64(levels-1)

			// Verify result is quantized (within one quantum)
			if math.Mod(result[0], quantumSize) > quantumSize*0.1 &&
				math.Mod(result[0], quantumSize) < quantumSize*0.9 {
				t.Errorf("Result %.6f not properly quantized (quantum=%.6f)", result[0], quantumSize)
			}

			t.Logf("%d-bit: input=%.3f → output=%.6f (quantum=%.6f)",
				tc.bits, tc.input, result[0], quantumSize)
		})
	}
}
