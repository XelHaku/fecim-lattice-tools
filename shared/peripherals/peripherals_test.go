package peripherals

import (
	"math"
	"testing"
)

// TestDACConversion verifies DAC converts levels to correct voltages.
func TestDACConversion(t *testing.T) {
	dac := DefaultDAC()

	// Test endpoints
	minV := dac.Convert(0)
	maxV := dac.Convert(dac.Levels() - 1)

	if math.Abs(minV-dac.VrefLow) > 0.001 {
		t.Errorf("Level 0 should be VrefLow (%.3f), got %.3f", dac.VrefLow, minV)
	}
	if math.Abs(maxV-dac.VrefHigh) > 0.001 {
		t.Errorf("Max level should be VrefHigh (%.3f), got %.3f", dac.VrefHigh, maxV)
	}

	// Test monotonicity
	prevV := dac.Convert(0)
	for level := 1; level < dac.Levels(); level++ {
		v := dac.Convert(level)
		if v <= prevV {
			t.Errorf("DAC not monotonic: level %d (%.4fV) <= level %d (%.4fV)",
				level, v, level-1, prevV)
		}
		prevV = v
	}
}

// TestDACLevels verifies 30 FeCIM levels.
func TestDACLevels(t *testing.T) {
	dac := DefaultDAC()

	if dac.Levels() < 30 {
		t.Errorf("DAC should support at least 30 levels, has %d", dac.Levels())
	}
}

// TestADCConversion verifies ADC converts voltages to correct levels.
func TestADCConversion(t *testing.T) {
	adc := DefaultADC()

	// Test endpoints
	minLevel := adc.Convert(adc.VrefLow)
	maxLevel := adc.Convert(adc.VrefHigh)

	if minLevel != 0 {
		t.Errorf("VrefLow should convert to 0, got %d", minLevel)
	}
	if maxLevel != adc.Levels()-1 {
		t.Errorf("VrefHigh should convert to %d, got %d", adc.Levels()-1, maxLevel)
	}

	// Test clamping
	belowMin := adc.Convert(adc.VrefLow - 1.0)
	aboveMax := adc.Convert(adc.VrefHigh + 1.0)

	if belowMin != 0 {
		t.Errorf("Below VrefLow should clamp to 0, got %d", belowMin)
	}
	if aboveMax != adc.Levels()-1 {
		t.Errorf("Above VrefHigh should clamp to %d, got %d", adc.Levels()-1, aboveMax)
	}
}

// TestADCENOB verifies ENOB calculation.
func TestADCENOB(t *testing.T) {
	adc := DefaultADC()

	enob := adc.ENOB()
	if enob <= 0 || enob > float64(adc.Bits) {
		t.Errorf("ENOB %.2f should be between 0 and %d bits", enob, adc.Bits)
	}

	// With ideal (0 INL/DNL), ENOB should equal bits
	idealADC := DefaultADC()
	idealADC.INL = 0
	idealADC.DNL = 0
	idealENOB := idealADC.ENOB()

	if math.Abs(idealENOB-float64(idealADC.Bits)) > 0.01 {
		t.Errorf("Ideal ADC ENOB should be %d, got %.2f", idealADC.Bits, idealENOB)
	}
}

// TestTIAConversion verifies current-to-voltage conversion.
func TestTIAConversion(t *testing.T) {
	tia := DefaultTIA()

	// Test zero current (should equal offset)
	v0 := tia.Convert(0)
	if math.Abs(v0-tia.OutputOffset) > 0.001 {
		t.Errorf("Zero current should give offset (%.3fV), got %.3fV", tia.OutputOffset, v0)
	}

	// Test linearity
	i1 := 10e-6
	i2 := 20e-6
	v1 := tia.Convert(i1)
	v2 := tia.Convert(i2)

	expectedRatio := 2.0
	actualRatio := (v2 - tia.OutputOffset) / (v1 - tia.OutputOffset)

	if math.Abs(actualRatio-expectedRatio) > 0.01 {
		t.Errorf("TIA should be linear: 2x current should give 2x voltage (got %.2fx)", actualRatio)
	}
}

// TestTIAClamping verifies output clamping.
func TestTIAClamping(t *testing.T) {
	tia := DefaultTIA()

	// High current should clamp
	vMax := tia.Convert(1.0) // 1 A would saturate any TIA
	if vMax > tia.MaxOutputVoltage {
		t.Errorf("TIA should clamp at %.2fV, got %.2fV", tia.MaxOutputVoltage, vMax)
	}
}

// TestChargePumpBoost verifies voltage boost.
func TestChargePumpBoost(t *testing.T) {
	pump := DefaultChargePump()

	idealV := pump.IdealOutputVoltage()
	actualV := pump.ActualOutputVoltage()

	// Ideal should be (N+1) * Vin
	expectedIdeal := float64(pump.Stages+1) * pump.InputVoltage
	if math.Abs(idealV-expectedIdeal) > 0.01 {
		t.Errorf("Ideal output should be %.2fV, got %.2fV", expectedIdeal, idealV)
	}

	// Actual should be less than ideal
	if actualV >= idealV {
		t.Errorf("Actual output (%.2fV) should be less than ideal (%.2fV)", actualV, idealV)
	}

	// Should still boost
	if actualV <= pump.InputVoltage {
		t.Errorf("Output (%.2fV) should exceed input (%.2fV)", actualV, pump.InputVoltage)
	}
}

// TestChargePumpEfficiency verifies energy calculations.
func TestChargePumpEfficiency(t *testing.T) {
	pump := DefaultChargePump()

	pIn := pump.PowerInput()
	pOut := pump.PowerOutput()
	pLoss := pump.PowerLoss()

	// Power balance
	if math.Abs(pIn-pOut-pLoss) > 1e-15 {
		t.Errorf("Power balance: Pin (%.2e) != Pout (%.2e) + Ploss (%.2e)",
			pIn, pOut, pLoss)
	}

	// Efficiency check
	calculatedEff := pOut / pIn
	if math.Abs(calculatedEff-pump.Efficiency) > 0.01 {
		t.Errorf("Efficiency mismatch: calculated %.2f, specified %.2f",
			calculatedEff, pump.Efficiency)
	}
}

func TestChargePumpEnergyMonotonicWithPulseDuration(t *testing.T) {
	pump := DefaultChargePump()

	d1 := 10e-9
	d2 := 20e-9

	e1 := pump.EnergyPerOperation(d1)
	e2 := pump.EnergyPerOperation(d2)

	if e2 <= e1 {
		t.Fatalf("Energy should increase with pulse duration: E(%.0fns)=%.3e, E(%.0fns)=%.3e",
			d1*1e9, e1, d2*1e9, e2)
	}
}

func TestChargePumpPowerOutputUsesActualVoltage(t *testing.T) {
	pump := DefaultChargePump()

	// Force a case where target rail is unattainable due to huge diode drops.
	pump.OutputVoltage = 1.5
	pump.DiodeDrop = 2.0 // unrealistically large; makes unregulated headroom small

	vActual := math.Abs(pump.ActualOutputVoltage())
	pOut := pump.PowerOutput()

	expected := vActual * math.Abs(pump.LoadCurrent)
	if math.Abs(pOut-expected) > 1e-18 {
		t.Fatalf("PowerOutput should use actual voltage: got %.3e, expected %.3e", pOut, expected)
	}
}

// TestDACToADCRoundTrip verifies end-to-end conversion.
func TestDACToADCRoundTrip(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()

	// Set ADC range to match DAC output
	adc.VrefLow = dac.VrefLow
	adc.VrefHigh = dac.VrefHigh

	// Test round-trip for several levels
	testLevels := []int{0, 7, 15, 22, 29}
	for _, level := range testLevels {
		if level >= dac.Levels() || level >= adc.Levels() {
			continue
		}

		voltage := dac.Convert(level)
		recoveredLevel := adc.Convert(voltage)

		// Allow ±1 level due to quantization
		if abs(recoveredLevel-level) > 1 {
			t.Errorf("Round-trip for level %d: got %d (Δ=%d)",
				level, recoveredLevel, abs(recoveredLevel-level))
		}
	}
}

// TestDACConvertWithNonlinearity verifies DAC nonlinearity application.
func TestDACConvertWithNonlinearity(t *testing.T) {
	dac := DefaultDAC()
	dac.INL = 0.5
	dac.DNL = 0.5

	level := 15
	idealV := dac.Convert(level)
	noisyV := dac.ConvertWithNonlinearity(level)

	if math.Abs(noisyV-idealV) < 1e-9 {
		t.Errorf("DAC nonlinearity should change output, but got same as ideal: %.4fV", idealV)
	}

	lsb := dac.Resolution()
	if math.Abs(noisyV-idealV) > 2.0*lsb {
		t.Errorf("DAC nonlinearity deviation too large: ideal %.4fV, noisy %.4fV (Δ=%.2f LSB)",
			idealV, noisyV, math.Abs(noisyV-idealV)/lsb)
	}
}

// TestADCConvertWithNonlinearity verifies ADC nonlinearity effects.
func TestADCConvertWithNonlinearity(t *testing.T) {
	adc := DefaultADC()
	adc.INL = 1.0
	adc.DNL = 1.0

	v := (adc.VrefHigh + adc.VrefLow) / 2.0
	idealL := adc.Convert(v)
	noisyL := adc.ConvertWithNonlinearity(v)

	if abs(noisyL-idealL) > 2 {
		t.Errorf("ADC nonlinearity caused too large deviation: ideal %d, noisy %d", idealL, noisyL)
	}
}

// TestADCTheoreticalSNR verifies SNR calculation.
func TestADCTheoreticalSNR(t *testing.T) {
	adc := DefaultADC()
	adc.Bits = 5
	snr := adc.TheoreticalSNR()
	expected := 6.02*5.0 + 1.76
	if math.Abs(snr-expected) > 0.01 {
		t.Errorf("Expected TheoreticalSNR %.2f, got %.2f", expected, snr)
	}
}

// TestADCEffectiveSNR verifies ENOB-based SNR.
func TestADCEffectiveSNR(t *testing.T) {
	adc := DefaultADC()
	adc.INL = 0.5
	adc.DNL = 0.5

	tSNR := adc.TheoreticalSNR()
	eSNR := adc.EffectiveSNR()

	if eSNR >= tSNR {
		t.Errorf("Effective SNR (%.2f) should be less than Theoretical SNR (%.2f) when nonlinearity exists", eSNR, tSNR)
	}

	adc.INL = 0
	adc.DNL = 0
	if math.Abs(adc.EffectiveSNR()-adc.TheoreticalSNR()) > 0.01 {
		t.Errorf("Ideal ADC should have EffectiveSNR == TheoreticalSNR")
	}
}

// TestTIAConvertWithNoise verifies TIA noise injection.
func TestTIAConvertWithNoise(t *testing.T) {
	tia := DefaultTIA()

	current := 10e-6
	vIdeal := tia.Convert(current)

	const iterations = 10
	for i := 0; i < iterations; i++ {
		v := tia.ConvertWithNoise(current)
		if math.Abs(v-vIdeal) > 0.1 {
			t.Errorf("TIA mean voltage (%.4fV) deviated too far from ideal (%.4fV)", v, vIdeal)
		}
	}
}

// TestTIASNR verifies TIA signal-to-noise ratio.
func TestTIASNR(t *testing.T) {
	tia := DefaultTIA()

	snrLow := tia.SNR(1e-7)
	snrHigh := tia.SNR(1e-5)

	if snrHigh <= snrLow {
		t.Errorf("Higher current should have better SNR: SNR(10uA)=%.2fdB, SNR(0.1uA)=%.2fdB", snrHigh, snrLow)
	}
}

// TestTIAMinDetectableCurrent verifies sensitivity.
func TestTIAMinDetectableCurrent(t *testing.T) {
	tia := DefaultTIA()
	minI := tia.MinDetectableCurrent()
	if minI <= 0 {
		t.Errorf("Min detectable current should be positive, got %.2e", minI)
	}
}

// TestTIADynamicRange verifies dynamic range.
func TestTIADynamicRange(t *testing.T) {
	tia := DefaultTIA()
	dr := tia.DynamicRange()
	if dr < 50 || dr > 120 {
		t.Errorf("Reasonable Dynamic Range expected, got %.2fdB", dr)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ============================================================================
// Additional Comprehensive Tests
// ============================================================================

// TestADCEnergyPerConversion_AllTypes verifies energy varies by ADC type.
func TestADCEnergyPerConversion_AllTypes(t *testing.T) {
	testCases := []struct {
		name    string
		adcType ADCType
	}{
		{"SAR", ADCTypeSAR},
		{"Flash", ADCTypeFlash},
		{"SigmaDelta", ADCTypeSigmaDelta},
	}

	energies := make(map[ADCType]float64)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adc := DefaultADC()
			adc.Type = tc.adcType
			energy := adc.EnergyPerConversion()

			if energy <= 0 {
				t.Errorf("Expected positive energy, got %e", energy)
			}
			energies[tc.adcType] = energy
		})
	}

	// Verify different types have different energies
	if energies[ADCTypeSAR] == energies[ADCTypeFlash] {
		t.Error("SAR and Flash should have different energy characteristics")
	}
}

// TestADC_SARNoise_Enable verifies SAR noise enable/disable.
func TestADC_SARNoise_Enable(t *testing.T) {
	adc := DefaultADC()

	if adc.NoiseConfig != nil {
		t.Error("Expected NoiseConfig to be nil initially")
	}

	adc.EnableSARNoise()
	if adc.NoiseConfig == nil {
		t.Error("Expected NoiseConfig to be set after EnableSARNoise")
	}

	adc.DisableSARNoise()
	if adc.NoiseConfig != nil {
		t.Error("Expected NoiseConfig to be nil after DisableSARNoise")
	}
}

// TestADC_SARNoise_ThermalNoise verifies thermal noise calculation.
func TestADC_SARNoise_ThermalNoise(t *testing.T) {
	adc := DefaultADC()
	adc.EnableSARNoise()

	thermalNoise := adc.GetThermalNoiseVoltage()
	if thermalNoise <= 0 {
		t.Errorf("Expected positive thermal noise, got %e", thermalNoise)
	}

	// Verify formula: sqrt(kT/C)
	const kB = 1.380649e-23
	T := adc.NoiseConfig.TemperatureK
	C := adc.NoiseConfig.SamplingCapacitorF
	expected := math.Sqrt(kB * T / C)

	if math.Abs(thermalNoise-expected) > 1e-9 {
		t.Errorf("Expected thermal noise %e, got %e", expected, thermalNoise)
	}
}

// TestADC_SARNoise_EffectiveVref verifies reference drift with temperature.
func TestADC_SARNoise_EffectiveVref(t *testing.T) {
	adc := DefaultADC()
	adc.EnableSARNoise()

	// At reference temperature, drift should be minimal (aging only)
	vrefLow1, vrefHigh1 := adc.GetEffectiveVref()

	// Set different temperature (50K above reference)
	adc.SetTemperature(350.0)
	vrefLow2, vrefHigh2 := adc.GetEffectiveVref()

	// Effective refs should drift with temperature change
	if vrefLow1 == vrefLow2 && vrefHigh1 == vrefHigh2 {
		t.Error("Expected effective Vref to drift with temperature change")
	}
}

// TestADC_SARNoise_MetastabilityRate verifies metastability probability.
func TestADC_SARNoise_MetastabilityRate(t *testing.T) {
	adc := DefaultADC()
	adc.EnableSARNoise()

	threshold := 0.5
	nearThreshold := 0.5001
	farThreshold := 0.8

	rateNear := adc.GetMetastabilityErrorRate(nearThreshold, threshold)
	rateFar := adc.GetMetastabilityErrorRate(farThreshold, threshold)

	if rateFar >= rateNear {
		t.Errorf("Expected metastability rate higher near threshold: near=%e far=%e", rateNear, rateFar)
	}

	// Rate should be clamped to reasonable range
	if rateNear > 0.5 || rateNear < 0 {
		t.Errorf("Metastability rate should be in [0, 0.5], got %e", rateNear)
	}
}

// TestADC_ConvertWithSARNoise verifies noisy conversion produces valid levels.
func TestADC_ConvertWithSARNoise(t *testing.T) {
	adc := DefaultADC()
	adc.EnableSARNoise()

	testCases := []struct {
		voltage float64
		seed    int64
	}{
		{0.0, 123},
		{0.5, 456},
		{1.0, 789},
		{0.25, 111},
		{0.75, 222},
	}

	for _, tc := range testCases {
		level := adc.ConvertWithSARNoise(tc.voltage, tc.seed)

		if level < 0 || level >= adc.Levels() {
			t.Errorf("ConvertWithSARNoise(%.2f, %d) returned invalid level %d", tc.voltage, tc.seed, level)
		}
	}
}

// TestADC_SARNoiseReport verifies noise report structure.
func TestADC_SARNoiseReport(t *testing.T) {
	t.Run("Disabled", func(t *testing.T) {
		adc := DefaultADC()
		report := adc.GetSARNoiseReport()

		if report["enabled"] != 0 {
			t.Error("Expected enabled=0 when noise disabled")
		}
	})

	t.Run("Enabled", func(t *testing.T) {
		adc := DefaultADC()
		adc.EnableSARNoise()
		report := adc.GetSARNoiseReport()

		if report["enabled"] != 1 {
			t.Error("Expected enabled=1 when noise enabled")
		}

		expectedKeys := []string{
			"vref_low_effective",
			"vref_high_effective",
			"vref_drift_ppm",
			"temperature_k",
			"thermal_noise_uv",
			"thermal_noise_lsb",
			"metastability_base_prob",
			"effective_bits_with_noise",
		}

		for _, key := range expectedKeys {
			if _, ok := report[key]; !ok {
				t.Errorf("Expected report to contain key %s", key)
			}
		}
	})
}

// TestDAC_VoltageRange verifies voltage range reporting.
func TestDAC_VoltageRange(t *testing.T) {
	dac := DefaultDAC()
	min, max := dac.VoltageRange()

	if min != dac.VrefLow {
		t.Errorf("Expected min=%f, got %f", dac.VrefLow, min)
	}
	if max != dac.VrefHigh {
		t.Errorf("Expected max=%f, got %f", dac.VrefHigh, max)
	}
}

// TestDAC_Resolution verifies resolution calculation.
func TestDAC_Resolution(t *testing.T) {
	dac := DefaultDAC()
	resolution := dac.Resolution()

	expected := (dac.VrefHigh - dac.VrefLow) / float64(dac.Levels()-1)
	if math.Abs(resolution-expected) > 1e-9 {
		t.Errorf("Expected resolution %f, got %f", expected, resolution)
	}
}

// TestDAC_EnergyPerConversion verifies energy calculation.
func TestDAC_EnergyPerConversion(t *testing.T) {
	dac := DefaultDAC()
	energy := dac.EnergyPerConversion()

	if energy <= 0 || math.IsInf(energy, 0) {
		t.Errorf("Expected positive finite energy, got %e", energy)
	}

	// Verify formula: C * (Vspan/2)^2 * Levels
	capacitance := 0.2e-15
	levels := float64(dac.Levels())
	vrefSpan := dac.VrefHigh - dac.VrefLow
	vref := vrefSpan / 2.0
	expected := capacitance * vref * vref * levels

	if math.Abs(energy-expected) > 1e-20 {
		t.Errorf("Expected energy %e, got %e", expected, energy)
	}
}

// TestTIA_Convert_Linear verifies linear I-V conversion.
func TestTIA_Convert_Linear(t *testing.T) {
	tia := DefaultTIA()

	testCases := []struct {
		current  float64
		expected float64
	}{
		{0, tia.OutputOffset},
		{10e-6, 10e-6*tia.Gain + tia.OutputOffset},
		{50e-6, 50e-6*tia.Gain + tia.OutputOffset},
	}

	for _, tc := range testCases {
		output := tia.Convert(tc.current)

		// Clamp expected to output range
		expectedClamped := tc.expected
		if expectedClamped < 0 {
			expectedClamped = 0
		}
		if expectedClamped > tia.MaxOutputVoltage {
			expectedClamped = tia.MaxOutputVoltage
		}

		if math.Abs(output-expectedClamped) > 1e-9 {
			t.Errorf("Convert(%e) expected %f, got %f", tc.current, expectedClamped, output)
		}
	}
}

// TestTIA_SettlingTime verifies settling time calculation.
func TestTIA_SettlingTime(t *testing.T) {
	tia := DefaultTIA()
	settleTime := tia.SettlingTime()

	if settleTime <= 0 {
		t.Errorf("Expected positive settling time, got %e", settleTime)
	}

	// Verify formula: ln(1/0.001) / (2*pi*BW)
	accuracy := 0.001
	expected := math.Log(1/accuracy) / (2 * math.Pi * tia.Bandwidth)

	if math.Abs(settleTime-expected) > 1e-12 {
		t.Errorf("Expected settling time %e, got %e", expected, settleTime)
	}
}

// TestTIA_PowerConsumption verifies power calculation.
func TestTIA_PowerConsumption(t *testing.T) {
	tia := DefaultTIA()
	power := tia.PowerConsumption()

	if power <= 0 || math.IsInf(power, 0) {
		t.Errorf("Expected positive finite power, got %e", power)
	}
}

// TestChargePump_IdealOutputVoltage verifies ideal voltage formula.
func TestChargePump_IdealOutputVoltage(t *testing.T) {
	cp := DefaultChargePump()
	ideal := cp.IdealOutputVoltage()

	// For positive pump: (Stages+1) * InputVoltage
	expected := float64(cp.Stages+1) * cp.InputVoltage
	if math.Abs(ideal-expected) > 1e-9 {
		t.Errorf("Expected ideal voltage %f, got %f", expected, ideal)
	}
}

// TestChargePump_ActualOutputVoltage verifies voltage with losses.
func TestChargePump_ActualOutputVoltage(t *testing.T) {
	cp := DefaultChargePump()
	actual := cp.ActualOutputVoltage()
	ideal := cp.IdealOutputVoltage()

	// Actual should be less than ideal
	if actual > ideal {
		t.Errorf("Expected actual <= ideal, got actual=%f ideal=%f", actual, ideal)
	}

	// Should be clamped to OutputVoltage
	if actual > cp.OutputVoltage {
		t.Errorf("Expected actual <= %f, got %f", cp.OutputVoltage, actual)
	}
}

// TestChargePump_OutputRipple verifies ripple calculation.
func TestChargePump_OutputRipple(t *testing.T) {
	cp := DefaultChargePump()
	ripple := cp.OutputRipple()

	if ripple <= 0 {
		t.Errorf("Expected positive ripple, got %e", ripple)
	}

	// Verify formula: |Iload| / (10*C*f)
	cOut := cp.FlyCapacitance * 10
	expected := math.Abs(cp.LoadCurrent) / (cOut * cp.ClockFrequency)

	if math.Abs(ripple-expected) > 1e-15 {
		t.Errorf("Expected ripple %e, got %e", expected, ripple)
	}
}

// TestChargePump_BoostFactor verifies boost calculation.
func TestChargePump_BoostFactor(t *testing.T) {
	cp := DefaultChargePump()
	boostFactor := cp.BoostFactor()

	expected := cp.ActualOutputVoltage() / cp.InputVoltage
	if math.Abs(boostFactor-expected) > 1e-9 {
		t.Errorf("Expected boost factor %f, got %f", expected, boostFactor)
	}

	// Boost factor should be > 1 for positive pump
	if boostFactor <= 1 {
		t.Errorf("Expected boost factor > 1, got %f", boostFactor)
	}
}

// TestChargePump_PowerConservation verifies energy balance.
func TestChargePump_PowerConservation(t *testing.T) {
	cp := DefaultChargePump()

	pIn := cp.PowerInput()
	pOut := cp.PowerOutput()
	pLoss := cp.PowerLoss()

	// Pin should be >= Pout
	if pIn < pOut-1e-15 {
		t.Errorf("Expected Pin >= Pout, got Pin=%e Pout=%e", pIn, pOut)
	}

	// Power balance: Pin = Pout + Ploss
	if math.Abs(pIn-pOut-pLoss) > 1e-15 {
		t.Errorf("Power balance failed: Pin=%e, Pout=%e, Ploss=%e", pIn, pOut, pLoss)
	}
}

// TestChargePump_EnergyPerOperation verifies energy calculation.
func TestChargePump_EnergyPerOperation(t *testing.T) {
	cp := DefaultChargePump()

	duration := 100e-9 // 100 ns
	energy := cp.EnergyPerOperation(duration)

	if energy <= 0 {
		t.Errorf("Expected positive energy, got %e", energy)
	}

	// Verify: E = PowerInput * duration
	expected := cp.PowerInput() * duration
	if math.Abs(energy-expected) > 1e-20 {
		t.Errorf("Expected energy %e, got %e", expected, energy)
	}
}

// TestNegativePump verifies negative pump configuration.
func TestNegativePump(t *testing.T) {
	cp := NegativePump()

	if cp.OutputVoltage != -1.5 {
		t.Errorf("Expected OutputVoltage=-1.5, got %f", cp.OutputVoltage)
	}

	if cp.Stages != 2 {
		t.Errorf("Expected Stages=2, got %d", cp.Stages)
	}

	// Ideal voltage should be negative
	ideal := cp.IdealOutputVoltage()
	if ideal >= 0 {
		t.Errorf("Expected negative ideal voltage, got %f", ideal)
	}
}

// TestChargePump_SupportsLevel verifies level support checking.
func TestChargePump_SupportsLevel(t *testing.T) {
	cp := DefaultChargePump()

	// Should support all levels up to max
	for level := 0; level <= 30; level++ {
		if !cp.SupportsLevel(level, 30) {
			t.Errorf("Expected to support level %d", level)
		}
	}
}

// TestChargePump_MaxCurrentCapability verifies max current calculation.
func TestChargePump_MaxCurrentCapability(t *testing.T) {
	cp := DefaultChargePump()
	maxCurrent := cp.MaxCurrentCapability()

	if maxCurrent <= 0 {
		t.Errorf("Expected positive max current, got %e", maxCurrent)
	}

	// Verify formula: C * f * (N+1) * Vin / |Vout|
	expected := cp.FlyCapacitance * cp.ClockFrequency * float64(cp.Stages+1) * cp.InputVoltage / math.Abs(cp.OutputVoltage)

	if math.Abs(maxCurrent-expected) > 1e-15 {
		t.Errorf("Expected max current %e, got %e", expected, maxCurrent)
	}
}

// TestChargePump_Area verifies area estimation.
func TestChargePump_Area(t *testing.T) {
	cp := DefaultChargePump()
	area := cp.Area()

	if area <= 0 {
		t.Errorf("Expected positive area, got %e", area)
	}
}

// TestChargePump_ChargeTransferEfficiency verifies efficiency calculation.
func TestChargePump_ChargeTransferEfficiency(t *testing.T) {
	cp := DefaultChargePump()
	efficiency := cp.ChargeTransferEfficiency()

	// Efficiency should be between 0 and 1
	if efficiency < 0 || efficiency > 1 {
		t.Errorf("Expected efficiency in [0,1], got %f", efficiency)
	}

	// Verify: η = |actual/ideal|
	ideal := cp.IdealOutputVoltage()
	actual := cp.ActualOutputVoltage()
	expected := math.Abs(actual / ideal)

	if math.Abs(efficiency-expected) > 1e-9 {
		t.Errorf("Expected efficiency %f, got %f", expected, efficiency)
	}
}

// TestChargePump_RiseTime verifies rise time calculation.
func TestChargePump_RiseTime(t *testing.T) {
	cp := DefaultChargePump()
	riseTime := cp.RiseTime()

	if riseTime <= 0 {
		t.Errorf("Expected positive rise time, got %e", riseTime)
	}

	// Verify: t = Stages * 2.2 / f
	expected := float64(cp.Stages) * 2.2 / cp.ClockFrequency

	if math.Abs(riseTime-expected) > 1e-12 {
		t.Errorf("Expected rise time %e, got %e", expected, riseTime)
	}
}

// TestADC_DefaultSARNoiseConfig verifies default noise parameters.
func TestADC_DefaultSARNoiseConfig(t *testing.T) {
	config := DefaultSARNoiseConfig()

	if config.ComparatorGain <= 0 {
		t.Error("Expected positive comparator gain")
	}
	if config.MetastabilityTau <= 0 {
		t.Error("Expected positive metastability tau")
	}
	if config.MetastabilityProb <= 0 || config.MetastabilityProb >= 1 {
		t.Error("Expected metastability probability in (0,1)")
	}
	if config.VrefDriftPPM <= 0 {
		t.Error("Expected positive Vref drift")
	}
	if config.SamplingCapacitorF <= 0 {
		t.Error("Expected positive sampling capacitor")
	}
	if !config.EnableMetastability {
		t.Error("Expected metastability enabled by default")
	}
	if !config.EnableReferenceDrift {
		t.Error("Expected reference drift enabled by default")
	}
}
