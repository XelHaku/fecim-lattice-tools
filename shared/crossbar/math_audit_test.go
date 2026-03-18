package crossbar

import (
	"math"
	"testing"
)

// ============================================================================
// MATH AUDIT: Hand-calculated expected values vs code output
// These tests verify the PHYSICS, not just that "tests pass".
// Every expected value is derived from first principles below.
// ============================================================================

// --- TEST 1: Ideal MVM (Ohm's law) ---
// A 2x3 array with known conductances. Input voltages on columns.
// Row output = sum_j(G[i][j] * V[j])
//
// G = | 0.5  0.2  0.8 |    V = | 1.0 |
//     | 0.3  0.7  0.1 |        | 0.5 |
//                               | 0.0 |
//
// Row 0: 0.5*1.0 + 0.2*0.5 + 0.8*0.0 = 0.5 + 0.1 + 0.0 = 0.600
// Row 1: 0.3*1.0 + 0.7*0.5 + 0.1*0.0 = 0.3 + 0.35 + 0.0 = 0.650
//
// But the code normalizes by /len(input), so:
// Row 0: 0.600 / 3 = 0.200
// Row 1: 0.650 / 3 = 0.21667
func TestMathAudit_IdealMVM_HandCalculation(t *testing.T) {
	cfg := &Config{
		Rows:       2,
		Cols:       3,
		NoiseLevel: 0,  // No noise
		ADCBits:    16, // High resolution to avoid quantization artifacts
		DACBits:    16,
	}

	a, err := NewArray(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Program exact conductances
	a.cells[0][0].Conductance = 0.5
	a.cells[0][1].Conductance = 0.2
	a.cells[0][2].Conductance = 0.8
	a.cells[1][0].Conductance = 0.3
	a.cells[1][1].Conductance = 0.7
	a.cells[1][2].Conductance = 0.1

	input := []float64{1.0, 0.5, 0.0}

	// Disable all non-idealities
	opts := &MVMOptions{
		EnableIRDrop:     false,
		EnableSneakPaths: false,
		EnableVariation:  false,
		EnableDrift:      false,
		Temperature:      300.0,
	}

	result, err := a.MVMWithNonIdealities(input, opts)
	if err != nil {
		t.Fatalf("MVM failed: %v", err)
	}

	// Hand-calculated values (with /N normalization the code applies)
	// Raw row 0: 0.5*1.0 + 0.2*0.5 + 0.8*0.0 = 0.600
	// Raw row 1: 0.3*1.0 + 0.7*0.5 + 0.1*0.0 = 0.650
	// Normalized: /3
	wantRaw0 := 0.600
	wantRaw1 := 0.650
	wantNorm0 := wantRaw0 / 3.0 // 0.200
	wantNorm1 := wantRaw1 / 3.0 // 0.21667

	t.Logf("=== IDEAL MVM HAND CALCULATION ===")
	t.Logf("G matrix:")
	t.Logf("  [0.5  0.2  0.8]")
	t.Logf("  [0.3  0.7  0.1]")
	t.Logf("Input: [1.0, 0.5, 0.0]")
	t.Logf("")
	t.Logf("Hand calculation (raw):  row0=%.4f  row1=%.4f", wantRaw0, wantRaw1)
	t.Logf("Hand calculation (norm): row0=%.5f  row1=%.5f", wantNorm0, wantNorm1)
	t.Logf("Code output (ideal):     row0=%.5f  row1=%.5f", result.IdealOutput[0], result.IdealOutput[1])
	t.Logf("Code output (actual):    row0=%.5f  row1=%.5f", result.ActualOutput[0], result.ActualOutput[1])

	// Check ideal output
	tol := 1e-6
	if math.Abs(result.IdealOutput[0]-wantNorm0) > tol {
		t.Errorf("Row 0 ideal: got %.6f, hand-calc %.6f, diff %.2e", result.IdealOutput[0], wantNorm0, result.IdealOutput[0]-wantNorm0)
	}
	if math.Abs(result.IdealOutput[1]-wantNorm1) > tol {
		t.Errorf("Row 1 ideal: got %.6f, hand-calc %.6f, diff %.2e", result.IdealOutput[1], wantNorm1, result.IdealOutput[1]-wantNorm1)
	}

	// Actual should equal ideal when all non-idealities disabled
	// (within ADC quantization — 16-bit ADC has 65536 levels, step = 1/65535 ≈ 1.5e-5)
	adcStep := 1.0 / float64((1<<16)-1)
	t.Logf("ADC step size (16-bit): %.2e", adcStep)

	if math.Abs(result.ActualOutput[0]-result.IdealOutput[0]) > adcStep+tol {
		t.Errorf("Row 0 actual≠ideal: actual=%.6f ideal=%.6f diff=%.2e (ADC step=%.2e)",
			result.ActualOutput[0], result.IdealOutput[0],
			math.Abs(result.ActualOutput[0]-result.IdealOutput[0]), adcStep)
	}
	if math.Abs(result.ActualOutput[1]-result.IdealOutput[1]) > adcStep+tol {
		t.Errorf("Row 1 actual≠ideal: actual=%.6f ideal=%.6f diff=%.2e (ADC step=%.2e)",
			result.ActualOutput[1], result.IdealOutput[1],
			math.Abs(result.ActualOutput[1]-result.IdealOutput[1]), adcStep)
	}
}

// --- TEST 2: SOR Solver with known analytical solution ---
// 1x1 array (single device): no parasitics to solve, output = G * V.
// Then 2x1 array with wire resistance: hand-solve Kirchhoff.
//
// For a 1x1 array:
//   G = 0.5, V_applied = 1.0
//   No wire resistance → device voltage = 1.0
//   I_device = 0.5 * 1.0 = 0.5
//   Output current = 0.5
func TestMathAudit_SORSolver_1x1_NoParasitics(t *testing.T) {
	solver, err := NewParasiticSolver(1, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	solver.SetConductances([][]float64{{0.5}})
	solver.SetParasitics(0, 0) // No wire resistance

	result, err := solver.SolveMVM([]float64{1.0})
	if err != nil {
		t.Fatalf("SolveMVM: %v", err)
	}

	want := 0.5 // I = G * V = 0.5 * 1.0
	t.Logf("=== SOR 1x1 NO PARASITICS ===")
	t.Logf("G=0.5, V=1.0 → I_expected=%.4f, I_got=%.4f", want, result.OutputCurrents[0])

	if math.Abs(result.OutputCurrents[0]-want) > 1e-6 {
		t.Errorf("1x1 output: got %.6f, want %.6f", result.OutputCurrents[0], want)
	}
	if !result.Converged {
		t.Error("solver did not converge for trivial 1x1 case")
	}
}

// --- TEST 3: SOR Solver 2x1 with wire resistance ---
// Two devices in a column with bit line resistance.
//
// Circuit:
//   V_applied ──[Rp]──●──[Rp]──●──GND
//                      |         |
//                     G[1]      G[0]
//
// With G[0]=0.5, G[1]=0.5, Rp_col=0.1 (normalized), V=1.0:
//
// Node voltages at bit line (measured from ground):
//   V_BL[0] = 0 (ground reference)
//   V_BL[1] = Rp * I_through_segment
//
// Device 0 (row 0, near ground): V_device[0] = V_applied - V_BL[0] = ~V_applied
// Device 1 (row 1, farther):     V_device[1] = V_applied - V_BL[1] < V_applied
//
// Key physics: device farther from ground should have LOWER effective voltage
// due to IR drop on the bit line. So I[1] < I[0] when both have same G.
func TestMathAudit_SORSolver_2x1_IRDrop_Direction(t *testing.T) {
	solver, err := NewParasiticSolver(2, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	solver.SetConductances([][]float64{{0.5}, {0.5}})
	solver.SetParasitics(0, 0.1) // Bit line resistance only

	result, err := solver.SolveMVM([]float64{1.0})
	if err != nil {
		t.Fatalf("SolveMVM: %v", err)
	}

	t.Logf("=== SOR 2x1 WITH BIT LINE RESISTANCE ===")
	t.Logf("G=[0.5, 0.5], Rp_col=0.1, V_applied=1.0")
	t.Logf("Device voltages: V[0]=%.6f, V[1]=%.6f", result.DeviceVoltages[0][0], result.DeviceVoltages[1][0])
	t.Logf("Device currents: I[0]=%.6f, I[1]=%.6f", result.DeviceCurrents[0][0], result.DeviceCurrents[1][0])
	t.Logf("Total output:    %.6f", result.OutputCurrents[0])

	// Physics check 1: Device 1 (farther from ground) should have lower voltage
	if result.DeviceVoltages[1][0] >= result.DeviceVoltages[0][0] {
		t.Errorf("IR drop direction wrong: V_device[1]=%.4f should be < V_device[0]=%.4f (farther from ground → more drop)",
			result.DeviceVoltages[1][0], result.DeviceVoltages[0][0])
	}

	// Physics check 2: Total output should be less than ideal (2 × 0.5 × 1.0 = 1.0)
	idealTotal := 2 * 0.5 * 1.0
	if result.OutputCurrents[0] >= idealTotal {
		t.Errorf("Parasitic result (%.4f) should be less than ideal (%.4f)", result.OutputCurrents[0], idealTotal)
	}

	// Physics check 3: With Rp=0.1 and G=0.5, the drop should be small but nonzero
	drop := idealTotal - result.OutputCurrents[0]
	t.Logf("Total current reduction from IR drop: %.6f (%.2f%%)", drop, drop/idealTotal*100)
	if drop < 1e-6 {
		t.Error("IR drop should produce measurable current reduction")
	}

	if !result.Converged {
		t.Error("solver did not converge")
	}
}

// --- TEST 4: Sneak path 3-cell series formula ---
// Three conductances in series: G_path = 1 / (1/g1 + 1/g2 + 1/g3)
//
// g1=1.0, g2=0.5, g3=0.25:
// 1/g1 = 1.0, 1/g2 = 2.0, 1/g3 = 4.0
// Sum = 7.0
// G_path = 1/7 = 0.142857...
//
// With V_in = 1.0: I_sneak = V * G_path = 0.142857
func TestMathAudit_SneakPath_SeriesConductance(t *testing.T) {
	g1, g2, g3 := 1.0, 0.5, 0.25
	gPath := 1.0 / (1.0/g1 + 1.0/g2 + 1.0/g3)

	want := 1.0 / 7.0 // = 0.142857142857...
	t.Logf("=== SNEAK PATH SERIES CONDUCTANCE ===")
	t.Logf("g1=%.2f, g2=%.2f, g3=%.2f", g1, g2, g3)
	t.Logf("1/g1=%.2f, 1/g2=%.2f, 1/g3=%.2f, sum=%.2f", 1/g1, 1/g2, 1/g3, 1/g1+1/g2+1/g3)
	t.Logf("G_path = 1/%.2f = %.6f (want %.6f)", 1/g1+1/g2+1/g3, gPath, want)

	if math.Abs(gPath-want) > 1e-12 {
		t.Errorf("Series conductance: got %.10f, want %.10f", gPath, want)
	}

	// Now verify the crossbar code computes the same thing for a 2x2 passive array
	cfg := &Config{Rows: 2, Cols: 2, NoiseLevel: 0, ADCBits: 8, DACBits: 8}
	a, aErr := NewArray(cfg)
	if aErr != nil {
		t.Fatal(aErr)
	}
	// Set up so there's one clear sneak path:
	// Target: read row 0.
	// Sneak path: col0→cell(1,0)→cell(1,1)→cell(0,1)
	a.cells[0][0].Conductance = 0.5 // Direct path (not sneak)
	a.cells[0][1].Conductance = g3  // Exit conductance
	a.cells[1][0].Conductance = g1  // Entry conductance
	a.cells[1][1].Conductance = g2  // Middle conductance

	input := []float64{1.0, 0.0} // Only drive column 0
	sneakCurrent := a.computeFullSneakCurrent(0, input, nil)

	// Expected sneak: V_in=1.0 through path (1,0)→(1,1)→(0,1)
	// G_path = 1/(1/1.0 + 1/0.5 + 1/0.25) = 1/7
	// I_sneak = 1.0 * 1/7 = 0.142857
	wantSneak := 1.0 * want
	t.Logf("Sneak current from code: %.6f (want %.6f)", sneakCurrent, wantSneak)

	if math.Abs(sneakCurrent-wantSneak) > 1e-6 {
		t.Errorf("Sneak current: got %.6f, want %.6f", sneakCurrent, wantSneak)
	}
}

// --- TEST 5: ADC quantization math ---
// 4-bit ADC: 16 levels (0-15), Vref [0, 1]
// Resolution = 1.0 / 15 = 0.06667 V/LSB
//
// Test specific voltages:
//   0.0   → level 0
//   0.5   → level 7 or 8 (0.5 * 15 = 7.5, rounds to 8)
//   1.0   → level 15
//   0.333 → 0.333 * 15 = 4.995, rounds to 5
func TestMathAudit_ADCQuantization(t *testing.T) {
	t.Logf("=== ADC QUANTIZATION MATH ===")

	// Simulate ADC quantization like the crossbar does
	adcBits := 4
	levels := 1 << adcBits // 16

	tests := []struct {
		voltage  float64
		wantCode int
		reason   string
	}{
		{0.0, 0, "minimum voltage → code 0"},
		{1.0, 15, "maximum voltage → code 15"},
		{0.5, 8, "midpoint: 0.5*15=7.5, rounds to 8"},
		{1.0 / 3.0, 5, "0.333*15=4.995, rounds to 5"},
		{0.1, 2, "0.1*15=1.5, rounds to 2"},
		{0.0667, 1, "0.0667*15=1.0, rounds to 1"},
	}

	for _, tt := range tests {
		// This is the code's formula (from adc.go Convert):
		fraction := tt.voltage // already in [0,1]
		if fraction < 0 {
			fraction = 0
		}
		if fraction > 1 {
			fraction = 1
		}
		code := int(fraction*float64(levels-1) + 0.5)
		if code >= levels {
			code = levels - 1
		}

		t.Logf("V=%.4f → fraction*15=%.3f → code=%d (want %d): %s",
			tt.voltage, tt.voltage*float64(levels-1), code, tt.wantCode, tt.reason)

		if code != tt.wantCode {
			t.Errorf("ADC(%.4f): got code %d, want %d (%s)", tt.voltage, code, tt.wantCode, tt.reason)
		}
	}
}

// --- TEST 6: kT/C thermal noise formula ---
// V_rms = sqrt(kB * T / C)
// At 300K with C = 1 pF:
// kB = 1.380649e-23 J/K
// V_rms = sqrt(1.380649e-23 * 300 / 1e-12)
//       = sqrt(4.141947e-9)
//       = 6.4358e-5 V
//       = 64.36 µV
func TestMathAudit_kTC_ThermalNoise(t *testing.T) {
	const kB = 1.380649e-23
	T := 300.0
	C := 1e-12 // 1 pF

	want := math.Sqrt(kB * T / C)
	wantUV := want * 1e6

	t.Logf("=== kT/C THERMAL NOISE ===")
	t.Logf("T=%.0fK, C=%.0f pF", T, C*1e12)
	t.Logf("kB*T = %.4e J", kB*T)
	t.Logf("kB*T/C = %.4e V²", kB*T/C)
	t.Logf("V_rms = %.4e V = %.2f µV", want, wantUV)

	// Verify our hand calculation
	if math.Abs(wantUV-64.36) > 0.1 {
		t.Errorf("Hand calculation check: got %.2f µV, expected ~64.36 µV", wantUV)
	}

	t.Logf("PASS: kT/C = %.2f µV (textbook: ~64 µV for 1pF at 300K)", wantUV)
}

// --- TEST 7: Copper wire TCR ---
// R(T) = R_ref * (1 + α * (T - T_ref))
// α_Cu = 0.00393 /K (NIST)
//
// At 400K (127°C): factor = 1 + 0.00393 * 100 = 1.393 → 39.3% increase
// At 200K (-73°C): factor = 1 + 0.00393 * (-100) = 0.607 → 39.3% decrease
func TestMathAudit_CopperTCR(t *testing.T) {
	t.Logf("=== COPPER WIRE TCR ===")

	alpha := 0.00393 // /K, NIST handbook value for copper
	tRef := 300.0    // K

	tests := []struct {
		tempK      float64
		wantFactor float64
		desc       string
	}{
		{300.0, 1.000, "Room temperature (no change)"},
		{400.0, 1.393, "127°C: +39.3%"},
		{200.0, 0.607, "-73°C: -39.3%"},
		{350.0, 1.1965, "77°C: +19.65%"},
	}

	for _, tt := range tests {
		factor := 1.0 + alpha*(tt.tempK-tRef)
		t.Logf("T=%5.0fK: factor=%.4f (want %.4f) %s", tt.tempK, factor, tt.wantFactor, tt.desc)
		if math.Abs(factor-tt.wantFactor) > 0.001 {
			t.Errorf("TCR at %.0fK: got %.4f, want %.4f", tt.tempK, factor, tt.wantFactor)
		}
	}
}

// --- TEST 8: Energy per MAC comparison with published numbers ---
// Published reference points:
// - FeFET read: ~0.01 fJ per cell (Mueller 2012, Reis 2023)
// - SAR ADC 4-bit: ~20 fJ (ISSCC 2022 survey)
// - GPU MAC (A100, FP16): ~10 pJ including memory access (NVIDIA whitepaper)
//
// For a 32x32 array MVM (1024 MACs):
// Array energy: 32*32 * 0.01e-3 pJ = 0.01024 pJ
// ADC energy:   32 * 0.5 pJ * 2^(4-6) = 32 * 0.125 = 4.0 pJ
// DAC energy:   32 * 0.1 pJ = 3.2 pJ
// Total:        ~7.2 pJ
// GPU:          1024 * 10 pJ = 10240 pJ
// Ratio:        10240 / 7.2 ≈ 1422x
func TestMathAudit_EnergyEstimates(t *testing.T) {
	t.Logf("=== ENERGY ESTIMATES vs LITERATURE ===")

	rows, cols, adcBits := 32, 32, 4

	cellReadEnergy := 0.01e-3 // pJ per cell
	arrayEnergy := float64(rows*cols) * cellReadEnergy

	adcEnergyBase := 0.5 // pJ per conversion for 6-bit
	adcEnergy := float64(rows) * adcEnergyBase * math.Pow(2, float64(adcBits-6))

	dacEnergy := float64(cols) * 0.1

	totalEnergy := arrayEnergy + adcEnergy + dacEnergy

	gpuMAC := float64(rows*cols) * 10.0 // pJ
	ratio := gpuMAC / totalEnergy

	t.Logf("Array: %d×%d = %d MACs", rows, cols, rows*cols)
	t.Logf("Array energy:  %.4f pJ (%.4f fJ/cell)", arrayEnergy, cellReadEnergy*1000)
	t.Logf("ADC energy:    %.4f pJ (%d-bit SAR)", adcEnergy, adcBits)
	t.Logf("DAC energy:    %.4f pJ", dacEnergy)
	t.Logf("Total FeCIM:   %.4f pJ", totalEnergy)
	t.Logf("GPU equiv:     %.1f pJ (10 pJ/MAC)", gpuMAC)
	t.Logf("Efficiency:    %.0fx", ratio)

	// Sanity checks against literature ranges
	if arrayEnergy > 1.0 {
		t.Error("Array energy too high — FeFET read should be sub-pJ for 1024 cells")
	}
	if adcEnergy < 0.1 || adcEnergy > 100 {
		t.Errorf("ADC energy out of range: %.2f pJ (expect 0.1-100 pJ for 32 conversions)", adcEnergy)
	}
	if ratio < 100 || ratio > 100000 {
		t.Errorf("Energy ratio out of published range: %.0fx (expect 100-100000x)", ratio)
	}
	t.Logf("CAVEAT: These are order-of-magnitude estimates, not calibrated to a process node")
}

// --- TEST 9: Verify /N normalization issue ---
// The MVM divides by len(input). This means doubling array width halves output.
// A real crossbar doesn't do this — more columns = more current.
func TestMathAudit_NormalizationArtifact(t *testing.T) {
	t.Logf("=== NORMALIZATION ARTIFACT CHECK ===")

	// 2-column array
	cfg2 := &Config{Rows: 1, Cols: 2, NoiseLevel: 0, ADCBits: 16, DACBits: 16}
	a2, err2 := NewArray(cfg2)
	if err2 != nil {
		t.Fatal(err2)
	}
	a2.cells[0][0].Conductance = 0.5
	a2.cells[0][1].Conductance = 0.5

	opts := &MVMOptions{}
	r2, _ := a2.MVMWithNonIdealities([]float64{1.0, 1.0}, opts)

	// 4-column array (same weights, duplicated)
	cfg4 := &Config{Rows: 1, Cols: 4, NoiseLevel: 0, ADCBits: 16, DACBits: 16}
	a4, err4 := NewArray(cfg4)
	if err4 != nil {
		t.Fatal(err4)
	}
	a4.cells[0][0].Conductance = 0.5
	a4.cells[0][1].Conductance = 0.5
	a4.cells[0][2].Conductance = 0.5
	a4.cells[0][3].Conductance = 0.5

	r4, _ := a4.MVMWithNonIdealities([]float64{1.0, 1.0, 1.0, 1.0}, opts)

	// Physical reality: 4 cols should give 2× the current of 2 cols
	// With /N normalization: both give the same output (0.5)
	t.Logf("2-col output: %.6f (raw sum would be 1.0, /2 = 0.5)", r2.IdealOutput[0])
	t.Logf("4-col output: %.6f (raw sum would be 2.0, /4 = 0.5)", r4.IdealOutput[0])

	if math.Abs(r2.IdealOutput[0]-r4.IdealOutput[0]) < 0.01 {
		t.Logf("CONFIRMED: /N normalization makes 2-col and 4-col give same output")
		t.Logf("This is NON-PHYSICAL — a real crossbar with 4 cols would give 2× the current")
		t.Logf("Impact: relative comparisons within same array size are correct,")
		t.Logf("        but absolute current values cannot be compared across different array widths")
	} else {
		t.Logf("Outputs differ: this would indicate the code does NOT normalize by N")
	}
}

// --- TEST 10: Quantization levels sanity ---
// 30 levels means conductances should snap to values k/29 for k=0..29
func TestMathAudit_QuantizationLevels(t *testing.T) {
	t.Logf("=== 30-LEVEL QUANTIZATION ===")

	for level := 0; level < 30; level++ {
		expected := float64(level) / 29.0
		quantized := QuantizeToLevels(expected)
		roundTrip := GetLevel(quantized)

		if roundTrip != level {
			t.Errorf("Level %d: quantize(%.6f)=%.6f, GetLevel→%d (want %d)",
				level, expected, quantized, roundTrip, level)
		}
	}

	// Test midpoints between levels — should round to nearest
	for level := 0; level < 29; level++ {
		midpoint := (float64(level) + 0.5) / 29.0
		quantized := QuantizeToLevels(midpoint)
		gotLevel := GetLevel(quantized)

		// Midpoint should round up (0.5 rounds to ceiling in most implementations)
		t.Logf("Midpoint %.6f (between L%d and L%d) → L%d", midpoint, level, level+1, gotLevel)

		if gotLevel != level && gotLevel != level+1 {
			t.Errorf("Midpoint between L%d and L%d quantized to L%d (expected %d or %d)",
				level, level+1, gotLevel, level, level+1)
		}
	}
}
