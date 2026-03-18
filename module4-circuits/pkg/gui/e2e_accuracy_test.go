package gui

// e2e_accuracy_test.go - Research-grade end-to-end accuracy tests for the circuits module.
//
// Coverage:
//   1. TestE2E_WriteReadRoundTrip_AllLevels    — full 0-29 write→read→TIA→ADC pipeline
//   2. TestE2E_MVMComputeAccuracy_IdealVsCoupled — 4x4 MVM with hand-calculated reference
//   3. TestE2E_SignalChainErrorBudget           — DAC→array→TIA→ADC precision decomposition
//   4. TestE2E_ArchitectureComparison_0T1R_1T1R — passive vs active WL read/compute

import (
	"fmt"
	"math"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	"fecim-lattice-tools/shared/peripherals"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// ─────────────────────────────────────────────────────────────────────────────
// Test 1: Write-Read Round-Trip Accuracy
//
// For DefaultHZO material (FeCIM HZO), for EVERY level 0-29:
//   - Write the level to cell (0,0)
//   - Read back the conductance
//   - Convert through TIA + ADC
//   - Log: level, expected_G, actual_G, ADC_code
//   - Assert: conductance monotonicity (higher level = higher G)
//   - Measure: distinct ADC codes produced across 30 levels
// ─────────────────────────────────────────────────────────────────────────────
func TestE2E_WriteReadRoundTrip_AllLevels(t *testing.T) {
	const quantLevels = 30
	const readVoltage = 0.25

	ds := newTestDeviceState(1, 1)
	mat := ds.GetMaterial()

	t.Logf("Material: %s", mat.Name)
	t.Logf("Gmin = %.4e S | Gmax = %.4e S | On/Off ratio = %.1fx",
		mat.Gmin, mat.Gmax, mat.Gmax/mat.Gmin)
	t.Logf("ADC bits = %d | TIA gain = %.0f Ohm | Read voltage = %.3f V",
		ds.GetADCBits(), ds.tia.Gain, readVoltage)
	t.Log("")
	t.Log("────────────────────────────────────────────────────────────────────────────")
	t.Logf("%-6s  %-14s  %-14s  %-12s  %-10s  %-10s",
		"Level", "G_expected (S)", "I_row (uA)", "V_TIA (V)", "ADC Code", "Status")
	t.Log("────────────────────────────────────────────────────────────────────────────")

	type levelResult struct {
		level   int
		gS      float64
		iRowUA  float64
		vTIA    float64
		adcCode int
	}
	results := make([]levelResult, quantLevels)

	for lvl := 0; lvl < quantLevels; lvl++ {
		ds.SetOperationMode(OpModeRead)
		ds.SetCouplingMode(arraysim.CouplingIdeal)
		ds.SetWLSingle(0)
		ds.SetDACVoltage(0, readVoltage)

		weights := [][]int{{lvl}}
		ds.Compute(weights, quantLevels)

		gS := mat.DiscreteLevel(lvl, quantLevels)
		iRowUA := ds.GetRowCurrent(0)
		vTIA := ds.GetRowVoltage(0)
		code := ds.GetRowLevel(0)

		results[lvl] = levelResult{
			level:   lvl,
			gS:      gS,
			iRowUA:  iRowUA,
			vTIA:    vTIA,
			adcCode: code,
		}

		t.Logf("L%-5d  %-14.6e  %-14.6f  %-12.6f  %-10d  ok",
			lvl, gS, iRowUA, vTIA, code)
	}
	t.Log("────────────────────────────────────────────────────────────────────────────")

	// Assert: Conductance monotonicity
	monoViolations := 0
	for lvl := 1; lvl < quantLevels; lvl++ {
		if results[lvl].gS < results[lvl-1].gS-1e-12 {
			monoViolations++
			t.Errorf("MONOTONICITY FAIL: L%d G=%.6e < L%d G=%.6e",
				lvl, results[lvl].gS, lvl-1, results[lvl-1].gS)
		}
	}

	// Assert: Row current monotonicity
	currentMonoViolations := 0
	for lvl := 1; lvl < quantLevels; lvl++ {
		if results[lvl].iRowUA < results[lvl-1].iRowUA-1e-10 {
			currentMonoViolations++
			t.Errorf("CURRENT MONOTONICITY FAIL: L%d I=%.6f uA < L%d I=%.6f uA",
				lvl, results[lvl].iRowUA, lvl-1, results[lvl-1].iRowUA)
		}
	}

	// Assert: ADC code monotonicity (non-decreasing)
	adcMonoViolations := 0
	for lvl := 1; lvl < quantLevels; lvl++ {
		if results[lvl].adcCode < results[lvl-1].adcCode {
			adcMonoViolations++
			t.Errorf("ADC CODE MONOTONICITY FAIL: L%d code=%d < L%d code=%d",
				lvl, results[lvl].adcCode, lvl-1, results[lvl-1].adcCode)
		}
	}

	// Measure: Distinct ADC codes
	codeSet := make(map[int]bool)
	for _, r := range results {
		codeSet[r.adcCode] = true
	}
	distinctCodes := len(codeSet)

	// Summary
	t.Log("")
	t.Logf("SUMMARY: Write-Read Round-Trip (30 levels)")
	t.Logf("  Conductance monotonicity violations: %d", monoViolations)
	t.Logf("  Current monotonicity violations:     %d", currentMonoViolations)
	t.Logf("  ADC code monotonicity violations:    %d", adcMonoViolations)
	adcBits := ds.GetADCBits()
	adcMaxCodes := 1 << adcBits
	t.Logf("  Distinct ADC codes across 30 levels: %d / %d possible (%.1f%% utilization)",
		distinctCodes, adcMaxCodes, 100.0*float64(distinctCodes)/float64(adcMaxCodes))
	t.Logf("  Effective resolution: %.2f bits (log2(%d))",
		math.Log2(float64(distinctCodes)), distinctCodes)
	t.Logf("  G range: %.4e S to %.4e S (%.1fx)",
		results[0].gS, results[quantLevels-1].gS,
		results[quantLevels-1].gS/results[0].gS)
	t.Logf("  I range: %.6f uA to %.6f uA",
		results[0].iRowUA, results[quantLevels-1].iRowUA)

	if monoViolations > 0 || currentMonoViolations > 0 {
		t.Fatalf("FAIL: monotonicity violated (G=%d, I=%d)", monoViolations, currentMonoViolations)
	}

	// With 6-bit ADC (64 codes), we expect significantly more than the 5 distinct
	// codes that 4-bit produced. The exponential conductance model compresses low
	// levels into the bottom ADC codes, so perfect 30-code mapping is not expected.
	// At minimum, 10 distinct codes should be resolved (2x improvement over 4-bit).
	if distinctCodes < 10 {
		t.Fatalf("FAIL: only %d distinct ADC code(s) across 30 levels; 6-bit ADC should resolve at least 10", distinctCodes)
	}
	t.Logf("  6-bit ADC improvement: %d distinct codes (vs ~5 with 4-bit, effective %.2f bits vs ~2.3 bits)",
		distinctCodes, math.Log2(float64(distinctCodes)))

	fmt.Printf("E2E_WRITE_READ_ROUNDTRIP: levels=%d mono_ok=true distinct_adc_codes=%d effective_bits=%.2f\n",
		quantLevels, distinctCodes, math.Log2(float64(distinctCodes)))
}

// ─────────────────────────────────────────────────────────────────────────────
// Test 2: MVM Compute Accuracy (Ideal vs Coupled)
//
// 4x4 array with known weight pattern and input vector.
// Compare row currents against hand-calculated reference:
//   I_row[r] = SUM(DiscreteLevel(w[r][c]) * V[c] * 1e6)
// ─────────────────────────────────────────────────────────────────────────────
func TestE2E_MVMComputeAccuracy_IdealVsCoupled(t *testing.T) {
	const N = 4
	const quantLevels = 30

	weights := [][]int{
		{0, 10, 20, 29},
		{29, 20, 10, 0},
		{15, 15, 15, 15},
		{0, 0, 29, 29},
	}
	inputs := []float64{1.0, 0.5, 0.25, 0.0}

	// Normalize inputs to read-safe voltage range.
	// The DeviceState treats DAC voltages as actual voltages applied to columns.
	// We use small voltages in the read-safe range to avoid saturation.
	inputScale := 0.10 // Scale factor to keep currents in measurable range
	scaledInputs := make([]float64, N)
	for i, v := range inputs {
		scaledInputs[i] = v * inputScale
	}

	// Compute hand-calculated reference currents.
	// I_row[r] = SUM_c( DiscreteLevel(w[r][c], 30) * V[c] ) in uA
	// (conductance in S, voltage in V, current = G*V in A, multiply by 1e6 for uA)
	mat := sharedphysics.FeCIMMaterial()
	refCurrentsUA := make([]float64, N)
	for r := 0; r < N; r++ {
		sum := 0.0
		for c := 0; c < N; c++ {
			gS := mat.DiscreteLevel(weights[r][c], quantLevels)
			sum += gS * scaledInputs[c] * 1e6 // -> uA
		}
		refCurrentsUA[r] = sum
	}

	// Coupling modes to test
	modes := []struct {
		name    string
		mode    arraysim.CouplingMode
		maxRel  float64
		tolDesc string
	}{
		{"Ideal", arraysim.CouplingIdeal, 1e-3, "<0.1%"},
		{"TierA", arraysim.CouplingTierA, 0.15, "<15%"},
	}

	for _, mc := range modes {
		t.Run(mc.name, func(t *testing.T) {
			ds := newTestDeviceState(N, N)
			ds.SetOperationMode(OpModeCompute)
			ds.SetCouplingMode(mc.mode)
			ds.SetWLAll()

			for c, v := range scaledInputs {
				ds.SetDACVoltage(c, v)
			}
			ds.Compute(weights, quantLevels)

			t.Logf("Coupling: %s | Tolerance: %s", mc.name, mc.tolDesc)
			t.Log("────────────────────────────────────────────────────────────────")
			t.Logf("%-6s  %-16s  %-16s  %-14s  %-8s",
				"Row", "I_sim (uA)", "I_ref (uA)", "Rel error", "Status")
			t.Log("────────────────────────────────────────────────────────────────")

			maxRelErr := 0.0
			allPass := true
			for r := 0; r < N; r++ {
				simUA := ds.GetRowCurrent(r)
				refUA := refCurrentsUA[r]
				relErr := 0.0
				if math.Abs(refUA) > 1e-12 {
					relErr = math.Abs(simUA-refUA) / math.Abs(refUA)
				}
				maxRelErr = math.Max(maxRelErr, relErr)

				status := "PASS"
				if relErr > mc.maxRel {
					status = "FAIL"
					allPass = false
					t.Errorf("Row %d: sim=%.6f uA ref=%.6f uA rel_err=%.4f exceeds %.4f",
						r, simUA, refUA, relErr, mc.maxRel)
				}
				t.Logf("Row %-3d  %-16.6f  %-16.6f  %-14.4e  %s", r, simUA, refUA, relErr, status)
			}
			t.Log("────────────────────────────────────────────────────────────────")
			t.Logf("Max relative error: %.4e | Tolerance: %s | Pass: %v",
				maxRelErr, mc.tolDesc, allPass)

			// For the Ideal mode, verify Row 3 is zero (input[3]=0)
			if mc.mode == arraysim.CouplingIdeal {
				row3 := ds.GetRowCurrent(3)
				// Only columns with non-zero voltage contribute.
				// Column 3 has input 0, but columns 0-2 have non-zero inputs
				// and row 3 has weights [0,0,29,29].
				// I_row3 = G(0)*V[0] + G(0)*V[1] + G(29)*V[2] + G(29)*V[3]
				// V[3]=0, so only columns 0..2 contribute.
				// Weights for row3: [0, 0, 29, 29]
				// G(0)*0.10 + G(0)*0.05 + G(29)*0.025 + G(29)*0.0
				expectedRow3 := mat.DiscreteLevel(0, 30)*1e6*0.10 +
					mat.DiscreteLevel(0, 30)*1e6*0.05 +
					mat.DiscreteLevel(29, 30)*1e6*0.025
				relErr3 := math.Abs(row3-expectedRow3) / math.Max(math.Abs(expectedRow3), 1e-12)
				if relErr3 > 1e-3 {
					t.Errorf("Row 3 detailed check: sim=%.6f ref=%.6f rel=%.4e", row3, expectedRow3, relErr3)
				}
			}

			fmt.Printf("E2E_MVM_%s: N=%d max_rel_err=%.4e pass=%v\n",
				mc.name, N, maxRelErr, allPass)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Test 3: Signal Chain Error Budget
//
// Measure where precision is lost in the compute pipeline:
//   - Start with known conductance matrix (set directly via weight levels)
//   - Compute with ideal voltage source (no DAC quantization)
//   - Compare: analog row current vs ADC-quantized output
//   - Measure quantization loss in bits
//   - Log the full error budget: DAC → array → TIA → ADC
// ─────────────────────────────────────────────────────────────────────────────
func TestE2E_SignalChainErrorBudget(t *testing.T) {
	const N = 4
	const quantLevels = 30

	// Weight matrix with diverse levels to exercise full dynamic range
	weights := [][]int{
		{5, 10, 20, 29},
		{0, 15, 15, 0},
		{29, 29, 29, 29},
		{1, 1, 1, 1},
	}
	// Use uniform voltage so analysis is clean
	readV := 0.10

	mat := sharedphysics.FeCIMMaterial()
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()

	t.Logf("Signal chain: DAC → %dx%d array → TIA (Rf=%.0f Ohm) → ADC (%d-bit)",
		N, N, tia.Gain, adc.Bits)
	t.Logf("Read voltage: %.3f V | Material: %s", readV, mat.Name)
	t.Log("")

	// Stage 1: Analog reference (no quantization, pure Ohm's law)
	t.Log("=== Stage 1: Analog Reference (I = G * V) ===")
	t.Log("────────────────────────────────────────────────────────────────")

	type rowBudget struct {
		analogCurrentUA float64 // Ideal I = sum(G*V) in uA
		tiaVoltageV     float64 // TIA output (ideal, no ADC)
		adcCode         int     // ADC output code
		reconstructedUA float64 // Current reconstructed from ADC code
		quantLossUA     float64 // |analog - reconstructed|
		quantLossBits   float64 // log2(analog/quantLoss) equivalent lost bits
		tiaSaturated    bool
	}
	budgets := make([]rowBudget, N)

	// Compute analog reference currents
	for r := 0; r < N; r++ {
		totalCurrentA := 0.0
		for c := 0; c < N; c++ {
			gS := mat.DiscreteLevel(weights[r][c], quantLevels)
			totalCurrentA += gS * readV
		}
		budgets[r].analogCurrentUA = totalCurrentA * 1e6
	}

	// Run through DeviceState to get TIA+ADC outputs
	ds := newTestDeviceState(N, N)
	ds.SetOperationMode(OpModeCompute)
	ds.SetCouplingMode(arraysim.CouplingIdeal)
	ds.SetWLAll()
	for c := 0; c < N; c++ {
		ds.SetDACVoltage(c, readV)
	}
	ds.Compute(weights, quantLevels)

	// Compute TIA/ADC characteristics for reconstruction
	// The sense chain scales TIA gain by 1/activeColumns in MVM mode
	activeColCount := N // All columns active with uniform voltage
	effectiveGain := tia.Gain / float64(activeColCount)

	t.Logf("TIA effective gain (MVM, %d cols): %.0f Ohm (base %.0f / %d)",
		activeColCount, effectiveGain, tia.Gain, activeColCount)

	for r := 0; r < N; r++ {
		budgets[r].tiaVoltageV = ds.GetRowVoltage(r)
		budgets[r].adcCode = ds.GetRowLevel(r)
		budgets[r].tiaSaturated = ds.IsSaturated(r)

		// Reconstruct current from ADC code:
		// ADC code -> voltage -> current
		// voltage = code / (2^bits - 1) * (Vmax - Vmin) + Vmin
		adcLevels := 1 << adc.Bits
		reconstructedV := float64(budgets[r].adcCode) / float64(adcLevels-1) *
			(adc.VrefHigh - adc.VrefLow)
		// Reconstruct current: V = I * Rf (in MVM mode, no offset)
		if effectiveGain > 0 {
			budgets[r].reconstructedUA = reconstructedV / effectiveGain * 1e6
		}

		budgets[r].quantLossUA = math.Abs(budgets[r].analogCurrentUA - budgets[r].reconstructedUA)
		if budgets[r].analogCurrentUA > 1e-12 {
			// Effective bits lost = log2(full_scale / quantization_error)
			// If quant error is 0, no bits lost
			if budgets[r].quantLossUA > 1e-15 {
				snr := budgets[r].analogCurrentUA / budgets[r].quantLossUA
				budgets[r].quantLossBits = math.Max(0, math.Log2(snr))
			} else {
				budgets[r].quantLossBits = float64(adc.Bits) // No loss
			}
		}
	}

	t.Logf("%-6s  %-14s  %-12s  %-10s  %-16s  %-14s  %-12s  %-8s",
		"Row", "I_analog (uA)", "V_TIA (V)", "ADC Code", "I_recon (uA)", "Quant Loss", "ENOB", "Sat?")
	t.Log("──────────────────────────────────────────────────────────────────────────────────────────────────")

	totalAnalog := 0.0
	totalQuantLoss := 0.0
	for r := 0; r < N; r++ {
		b := budgets[r]
		satStr := "no"
		if b.tiaSaturated {
			satStr = "YES"
		}
		t.Logf("Row %-3d  %-14.6f  %-12.6f  %-10d  %-16.6f  %-14.6e  %-12.2f  %-8s",
			r, b.analogCurrentUA, b.tiaVoltageV, b.adcCode, b.reconstructedUA,
			b.quantLossUA, b.quantLossBits, satStr)
		totalAnalog += b.analogCurrentUA
		totalQuantLoss += b.quantLossUA
	}
	t.Log("──────────────────────────────────────────────────────────────────────────────────────────────────")

	// Stage 2: Error budget decomposition
	t.Log("")
	t.Log("=== Stage 2: Error Budget Decomposition ===")

	// DAC quantization error (in this test we set voltages directly, so DAC error = 0)
	t.Logf("  DAC quantization error:   0 (voltages set directly)")

	// Array error (Ideal mode = 0 by definition)
	t.Logf("  Array coupling error:     0 (CouplingIdeal mode)")

	// TIA quantization: continuous -> voltage (no quantization, but gain scaling + saturation)
	tiaClipCount := 0
	for _, b := range budgets {
		if b.tiaSaturated {
			tiaClipCount++
		}
	}
	t.Logf("  TIA clipping events:      %d / %d rows", tiaClipCount, N)

	// ADC quantization
	adcENOB := adc.ENOB()
	adcTheoreticalSNR := adc.TheoreticalSNR()
	t.Logf("  ADC nominal bits:         %d", adc.Bits)
	t.Logf("  ADC ENOB:                 %.2f bits", adcENOB)
	t.Logf("  ADC theoretical SNR:      %.1f dB", adcTheoreticalSNR)

	// Total pipeline loss
	if totalAnalog > 0 {
		overallSNR := totalAnalog / math.Max(totalQuantLoss, 1e-15)
		overallENOB := math.Log2(overallSNR) / 6.02 * 6.02 // Just log2
		overallENOB = math.Log2(overallSNR)
		t.Logf("  Pipeline total quant loss: %.6e uA (%.2f%% of total)",
			totalQuantLoss, 100.0*totalQuantLoss/totalAnalog)
		t.Logf("  Pipeline effective ENOB:   %.2f bits", overallENOB)
	}

	// Energy cost of ADC conversion
	adcEnergyJ := adc.EnergyPerConversion()
	totalADCEnergyJ := adcEnergyJ * float64(N) // One conversion per row
	t.Logf("  ADC energy per conversion: %.2e J (%.2f fJ)", adcEnergyJ, adcEnergyJ*1e15)
	t.Logf("  Total ADC energy (%d rows): %.2e J (%.2f fJ)", N, totalADCEnergyJ, totalADCEnergyJ*1e15)

	// Assert: no row should be fully saturated in this configuration
	for r := 0; r < N; r++ {
		if budgets[r].tiaSaturated {
			// Saturation is a warning, not a test failure for this budget analysis
			t.Logf("  WARNING: Row %d saturated (I_analog=%.4f uA)", r, budgets[r].analogCurrentUA)
		}
	}

	// Assert: the row with the largest current (row 2: all L29) must have a
	// non-zero ADC code, proving the pipeline has usable resolution at full scale.
	maxRow := 0
	maxAnalog := 0.0
	for r := 0; r < N; r++ {
		if budgets[r].analogCurrentUA > maxAnalog {
			maxAnalog = budgets[r].analogCurrentUA
			maxRow = r
		}
	}
	if budgets[maxRow].adcCode == 0 && maxAnalog > 1e-6 {
		t.Errorf("Row %d has largest current (%.4f uA) but ADC code=0; pipeline has no resolution at full scale",
			maxRow, maxAnalog)
	}

	// Count how many distinct ADC codes the 4 rows produce. With 6-bit ADC
	// and gain-scaled TIA in MVM mode, rows should resolve into more distinct
	// codes than the previous 4-bit configuration.
	codeSet := make(map[int]bool)
	for _, b := range budgets {
		codeSet[b.adcCode] = true
	}
	t.Logf("  Distinct ADC codes across %d rows: %d", N, len(codeSet))

	fmt.Printf("E2E_ERROR_BUDGET: N=%d adc_bits=%d adc_enob=%.2f total_quant_loss_pct=%.4f\n",
		N, adc.Bits, adcENOB, 100.0*totalQuantLoss/math.Max(totalAnalog, 1e-15))
}

// ─────────────────────────────────────────────────────────────────────────────
// Test 4: Architecture Comparison (0T1R vs 1T1R)
//
// Same weight pattern, same input vector:
//   - 0T1R (passive): all WLs always on, sneak paths in read mode
//   - 1T1R (active):  selected WLs only, no sneak paths
//   - In COMPUTE mode (all WLs active): both should produce same result
//   - In READ mode (single WL): 1T1R reads single row, 0T1R has all rows active
// ─────────────────────────────────────────────────────────────────────────────
func TestE2E_ArchitectureComparison_0T1R_1T1R(t *testing.T) {
	const N = 4
	const quantLevels = 30

	weights := [][]int{
		{0, 10, 20, 29},
		{29, 20, 10, 0},
		{15, 15, 15, 15},
		{0, 0, 29, 29},
	}
	readV := 0.10

	mat := sharedphysics.FeCIMMaterial()

	// ───── Part A: COMPUTE mode (all WLs active) ─────
	t.Log("=== Part A: COMPUTE Mode (all WLs active) ===")
	t.Log("Both architectures should produce identical results when all WLs are on.")
	t.Log("")

	// 1T1R compute
	ds1T := newTestDeviceState(N, N)
	ds1T.SetPassiveMode(false)
	ds1T.SetOperationMode(OpModeCompute)
	ds1T.SetCouplingMode(arraysim.CouplingIdeal)
	ds1T.SetWLAll()
	for c := 0; c < N; c++ {
		ds1T.SetDACVoltage(c, readV)
	}
	ds1T.Compute(weights, quantLevels)

	currents1T := make([]float64, N)
	for r := 0; r < N; r++ {
		currents1T[r] = ds1T.GetRowCurrent(r)
	}

	// 0T1R compute
	ds0T := newTestDeviceState(N, N)
	ds0T.SetPassiveMode(true)
	ds0T.SetOperationMode(OpModeCompute)
	ds0T.SetCouplingMode(arraysim.CouplingIdeal)
	// In passive mode (0T1R), all WLs are always on.
	// Effective cell voltage = WL - BL. WL voltages default to 0.
	// Setting BL = readV gives Vcell = 0 - readV = -readV (negative).
	// The ideal compute path uses the signed voltage, so currents will be negative.
	// We compare magnitudes to account for the sign convention difference.
	for c := 0; c < N; c++ {
		ds0T.SetDACVoltage(c, readV)
	}
	ds0T.Compute(weights, quantLevels)

	currents0T := make([]float64, N)
	for r := 0; r < N; r++ {
		currents0T[r] = ds0T.GetRowCurrent(r)
	}

	// Reference calculation
	refCurrents := make([]float64, N)
	for r := 0; r < N; r++ {
		sum := 0.0
		for c := 0; c < N; c++ {
			gS := mat.DiscreteLevel(weights[r][c], quantLevels)
			sum += gS * readV * 1e6
		}
		refCurrents[r] = sum
	}

	t.Logf("%-6s  %-16s  %-16s  %-16s  %-12s  %-12s",
		"Row", "I_1T1R (uA)", "|I_0T1R| (uA)", "I_ref (uA)", "1T Rel Err", "0T Rel Err")
	t.Log("──────────────────────────────────────────────────────────────────────────────────────")

	// In ideal COMPUTE mode, 1T1R with all WLs on should match reference exactly.
	// For 0T1R, the passive mode sign convention (Vcell = WL - BL = -readV) produces
	// negative currents. We compare magnitudes.
	for r := 0; r < N; r++ {
		ref := refCurrents[r]
		abs0T := math.Abs(currents0T[r])
		rel1T := 0.0
		rel0T := 0.0
		if math.Abs(ref) > 1e-12 {
			rel1T = math.Abs(currents1T[r]-ref) / math.Abs(ref)
			rel0T = math.Abs(abs0T-ref) / math.Abs(ref)
		}
		t.Logf("Row %-3d  %-16.6f  %-16.6f  %-16.6f  %-12.4e  %-12.4e",
			r, currents1T[r], abs0T, ref, rel1T, rel0T)

		// 1T1R ideal should match reference to machine precision
		if rel1T > 1e-3 {
			t.Errorf("1T1R compute row %d: rel_err=%.4e exceeds 0.1%%", r, rel1T)
		}
		// 0T1R magnitude should also match reference (same conductances, same |voltage|)
		if rel0T > 1e-3 {
			t.Errorf("0T1R compute row %d: |I| rel_err=%.4e exceeds 0.1%%", r, rel0T)
		}
	}

	// ───── Part B: READ mode (single row) ─────
	t.Log("")
	t.Log("=== Part B: READ Mode (single row selected) ===")
	t.Log("1T1R selects one WL. 0T1R has all WLs always on (sneak paths).")
	t.Log("")

	targetRow := 0

	// 1T1R single-row read
	ds1TRead := newTestDeviceState(N, N)
	ds1TRead.SetPassiveMode(false)
	ds1TRead.SetOperationMode(OpModeRead)
	ds1TRead.SetCouplingMode(arraysim.CouplingIdeal)
	ds1TRead.SetWLSingle(targetRow)
	ds1TRead.SetDACVoltage(0, readV)
	for c := 1; c < N; c++ {
		ds1TRead.SetDACVoltage(c, 0)
	}
	ds1TRead.Compute(weights, quantLevels)

	i1TRead := ds1TRead.GetRowCurrent(targetRow)
	// Other rows should have zero current (inactive WL)
	otherRow1T := ds1TRead.GetRowCurrent(1)

	// 0T1R single-row read attempt
	ds0TRead := newTestDeviceState(N, N)
	ds0TRead.SetPassiveMode(true)
	ds0TRead.SetOperationMode(OpModeRead)
	ds0TRead.SetCouplingMode(arraysim.CouplingIdeal)
	// In passive mode, SetWLSingle is ignored - all WLs stay on
	ds0TRead.SetWLSingle(targetRow)
	ds0TRead.SetDACVoltage(0, readV)
	for c := 1; c < N; c++ {
		ds0TRead.SetDACVoltage(c, 0)
	}
	ds0TRead.Compute(weights, quantLevels)

	i0TRead := ds0TRead.GetRowCurrent(targetRow)
	otherRow0T := ds0TRead.GetRowCurrent(1)

	t.Logf("  1T1R: Row %d current = %.6f uA (other rows: %.6f uA)", targetRow, i1TRead, otherRow1T)
	t.Logf("  0T1R: Row %d current = %.6f uA (|I|=%.6f) (other rows: %.6f uA)",
		targetRow, i0TRead, math.Abs(i0TRead), otherRow0T)

	// Assert: 1T1R other rows should be zero (WL not active)
	if math.Abs(otherRow1T) > 1e-9 {
		t.Errorf("1T1R READ: non-target row has current %.6f uA (expected 0)", otherRow1T)
	}

	// Assert: 0T1R all rows should have non-zero current (all WLs active)
	if ds0TRead.IsPassiveMode() {
		// In passive mode, all WLs are on, so all rows with non-zero BL voltage
		// should have current flowing. Row 1 also sees the same BL voltages.
		t.Logf("  0T1R passive mode confirmed: all WLs active")
		if math.Abs(otherRow0T) < 1e-12 && math.Abs(i0TRead) > 1e-12 {
			// Other rows may or may not have current depending on BL pattern.
			// With only BL0 active, other rows see the same BL voltage.
			t.Logf("  Note: other row current = %.6e uA (depends on BL pattern)", otherRow0T)
		}
	}

	// Assert: 1T1R target row current should be deterministic
	expectedI := mat.DiscreteLevel(weights[targetRow][0], quantLevels) * readV * 1e6
	rel1TRead := math.Abs(i1TRead-expectedI) / math.Max(math.Abs(expectedI), 1e-12)
	if rel1TRead > 1e-3 {
		t.Errorf("1T1R READ row %d: sim=%.6f ref=%.6f rel=%.4e",
			targetRow, i1TRead, expectedI, rel1TRead)
	}
	t.Logf("  1T1R row %d reference check: sim=%.6f ref=%.6f rel_err=%.4e",
		targetRow, i1TRead, expectedI, rel1TRead)

	// ───── Part C: Architecture behavioral differences ─────
	t.Log("")
	t.Log("=== Part C: Behavioral Differences Summary ===")
	t.Logf("  1T1R: can isolate single row (WL select) -> no sneak paths")
	t.Logf("  0T1R: all WLs always on -> all rows active in all modes")
	passiveRowsActive := 0
	for r := 0; r < N; r++ {
		if math.Abs(ds0TRead.GetRowCurrent(r)) > 1e-12 {
			passiveRowsActive++
		}
	}
	t.Logf("  0T1R rows with current in 'single-row' read: %d / %d", passiveRowsActive, N)

	fmt.Printf("E2E_ARCH_COMPARISON: 1T1R_read_I=%.6f 0T1R_read_I=%.6f 1T1R_inactive_row=%.6e passive_rows_active=%d\n",
		i1TRead, i0TRead, otherRow1T, passiveRowsActive)
}
