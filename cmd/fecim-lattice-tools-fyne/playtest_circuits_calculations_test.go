//go:build !ci
// +build !ci

package main

import (
	"math"
	"math/rand"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	"fecim-lattice-tools/shared/peripherals"
)

// TestPlaytestCircuits_DACQuantization verifies DAC linear interpolation and level count.
func TestPlaytestCircuits_DACQuantization(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	dac := peripherals.DefaultDAC() // 4-bit, VrefLow=-1.5, VrefHigh=1.5

	tests := []struct {
		code  int
		wantV float64
	}{
		{0, -1.5},                       // code 0 → VrefLow
		{15, 1.5},                       // code 15 → VrefHigh
		{7, -1.5 + 7.0/15.0*3.0},       // linear interpolation: ~-0.1
		{8, -1.5 + 8.0/15.0*3.0},       // ~+0.1
	}
	for _, tc := range tests {
		got := dac.Convert(tc.code)
		if math.Abs(got-tc.wantV) > 1e-9 {
			t.Errorf("DAC.Convert(%d) = %v, want %v", tc.code, got, tc.wantV)
		}
	}

	if dac.Levels() != 16 {
		t.Errorf("DAC.Levels() = %d, want 16", dac.Levels())
	}
}

// TestPlaytestCircuits_TIASenseChain verifies TIA voltage conversion and saturation flags.
func TestPlaytestCircuits_TIASenseChain(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	sense := arraysim.SenseChain{
		TIA: arraysim.TIAConfig{Rf: 10e3, Vref: 0.5, Vmin: 0, Vmax: 1.0},
		ADC: arraysim.ADCConfig{Bits: 4, Vmin: 0, Vmax: 1.0},
	}
	tests := []struct {
		name     string
		currentA float64
		wantVout float64
		wantSat  bool
	}{
		{"10µA", 10e-6, 0.6, false},          // 0.5 + 10e-6*10e3 = 0.6
		{"-10µA", -10e-6, 0.4, false},        // 0.5 - 0.1 = 0.4
		{"100µA_saturated", 100e-6, 1.0, true}, // 0.5 + 1.0 = 1.5 → clamped to 1.0
		{"-60µA_saturated", -60e-6, 0.0, true}, // 0.5 - 0.6 = -0.1 → clamped to 0
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := sense.ConvertCurrent(tc.currentA)
			if math.Abs(r.Vout-tc.wantVout) > 1e-9 {
				t.Errorf("Vout = %v, want %v", r.Vout, tc.wantVout)
			}
			if r.TIASaturated != tc.wantSat {
				t.Errorf("TIASaturated = %v, want %v", r.TIASaturated, tc.wantSat)
			}
		})
	}
}

// TestPlaytestCircuits_ADCQuantization verifies ADC round-to-nearest quantization and clamping.
func TestPlaytestCircuits_ADCQuantization(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	adc := peripherals.DefaultADC() // 4-bit, VrefLow=0, VrefHigh=1.0
	tests := []struct {
		voltage  float64
		wantCode int
	}{
		{0.0, 0},
		{1.0, 15},
		{0.5, 8},   // round(0.5*15+0.5) = round(8.0) = 8
		{0.1, 2},   // round(0.1*15+0.5) = round(2.0) = 2
		{-0.1, 0},  // clamped to VrefLow
		{1.5, 15},  // clamped to VrefHigh
	}
	for _, tc := range tests {
		got := adc.Convert(tc.voltage)
		if got != tc.wantCode {
			t.Errorf("ADC.Convert(%v) = %d, want %d", tc.voltage, got, tc.wantCode)
		}
	}
}

// TestPlaytestCircuits_FullPipeline tests the DAC → 2×2 crossbar MVM → TIA → ADC signal chain end-to-end.
func TestPlaytestCircuits_FullPipeline(t *testing.T) {
	playtestSkipUnlessEnabled(t)

	// DAC generates input voltages from codes.
	dac := peripherals.DefaultDAC()
	inputCodes := []int{8, 12} // 2 columns
	inputV := make([]float64, len(inputCodes))
	for i, code := range inputCodes {
		inputV[i] = dac.Convert(code)
	}

	// 2×2 conductance matrix (Siemens).
	G := [2][2]float64{
		{10e-6, 20e-6},
		{30e-6, 40e-6},
	}

	// MVM: I[r] = Σ_c V[c] * G[r][c].
	rowCurrents := make([]float64, 2)
	for r := 0; r < 2; r++ {
		for c := 0; c < 2; c++ {
			rowCurrents[r] += inputV[c] * G[r][c]
		}
	}

	// Sense chain.
	sense := arraysim.SenseChain{
		TIA: arraysim.TIAConfig{Rf: 10e3, Vref: 0.5, Vmin: 0, Vmax: 1.0},
		ADC: arraysim.ADCConfig{Bits: 4, Vmin: 0, Vmax: 1.0},
	}

	for r, current := range rowCurrents {
		result := sense.ConvertCurrent(current)
		expectedVout := 0.5 + current*10e3
		if expectedVout < 0 {
			expectedVout = 0
		}
		if expectedVout > 1.0 {
			expectedVout = 1.0
		}
		if math.Abs(result.Vout-expectedVout) > 1e-9 {
			t.Errorf("row %d: Vout = %v, want %v", r, result.Vout, expectedVout)
		}
		t.Logf("row %d: I=%.3e A, Vout=%.6f V, code=%d, sat=%v",
			r, current, result.Vout, result.Code, result.TIASaturated)
	}
}

// TestPlaytestCircuits_ENOBCalculation verifies ENOB and EffectiveSNR formulas.
func TestPlaytestCircuits_ENOBCalculation(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	adc := peripherals.DefaultADC() // Bits=6, INL=0.5, DNL=0.25

	// ENOB = N - log2(1 + max(|INL|, |DNL|))
	expectedENOB := float64(adc.Bits) - math.Log2(1.0+math.Max(math.Abs(adc.INL), math.Abs(adc.DNL)))
	gotENOB := adc.ENOB()
	if math.Abs(gotENOB-expectedENOB) > 1e-9 {
		t.Errorf("ENOB = %v, want %v", gotENOB, expectedENOB)
	}

	// EffectiveSNR = 6.02*ENOB + 1.76
	expectedSNR := 6.02*expectedENOB + 1.76
	gotSNR := adc.EffectiveSNR()
	if math.Abs(gotSNR-expectedSNR) > 1e-9 {
		t.Errorf("EffectiveSNR = %v, want %v", gotSNR, expectedSNR)
	}

	t.Logf("ENOB: %.4f bits, EffectiveSNR: %.2f dB", gotENOB, gotSNR)
}

// TestPlaytestCircuits_ThermalNoise verifies that ConvertCurrentWithNoise produces
// output noise consistent with the Johnson–Nyquist (4kTBW/Rf) model.
func TestPlaytestCircuits_ThermalNoise(t *testing.T) {
	playtestSkipUnlessEnabled(t)
	sense := arraysim.SenseChain{
		TIA: arraysim.TIAConfig{Rf: 10e3, Vref: 0.5, Vmin: 0, Vmax: 3.0}, // Wide range to avoid saturation
		ADC: arraysim.ADCConfig{Bits: 8, Vmin: 0, Vmax: 3.0},
	}

	baseI := 50e-6 // 50 µA base current

	N := 10000
	rng := rand.New(rand.NewSource(42))
	vouts := make([]float64, N)
	for i := 0; i < N; i++ {
		r := sense.ConvertCurrentWithNoise(baseI, rng)
		vouts[i] = r.Vout
	}

	mean := 0.0
	for _, v := range vouts {
		mean += v
	}
	mean /= float64(N)

	variance := 0.0
	for _, v := range vouts {
		d := v - mean
		variance += d * d
	}
	stddev := math.Sqrt(variance / float64(N-1))

	// Expected Vout stddev = noise_sigma_I * Rf
	// sigma_I = sqrt(4*kT*BW/Rf), kT=1.38e-23*300, BW=100e6
	kT := 1.38e-23 * 300.0
	bw := 100e6
	expectedSigmaI := math.Sqrt(4 * kT * bw / 10e3)
	expectedSigmaV := expectedSigmaI * 10e3

	// Allow 15% tolerance for statistical sampling.
	if math.Abs(stddev-expectedSigmaV)/expectedSigmaV > 0.15 {
		t.Errorf("noise stddev = %e V, expected ~%e V (>15%% off)", stddev, expectedSigmaV)
	}
	t.Logf("Noise: mean=%.6f V, stddev=%.6e V, expected_sigma=%.6e V", mean, stddev, expectedSigmaV)
}
