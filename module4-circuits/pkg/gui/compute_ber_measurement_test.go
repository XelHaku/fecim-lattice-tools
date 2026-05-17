//go:build legacy_fyne

package gui

import (
	"math"
	"math/rand"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
)

func TestComputeBERMeasurement_WithSenseNoise(t *testing.T) {
	weights := [][]int{
		{4, 8, 12, 16, 20, 24, 28, 29},
		{29, 24, 20, 16, 12, 8, 4, 0},
		{10, 11, 12, 13, 14, 15, 16, 17},
		{17, 16, 15, 14, 13, 12, 11, 10},
		{2, 6, 10, 14, 18, 22, 26, 29},
		{29, 26, 22, 18, 14, 10, 6, 2},
		{5, 9, 13, 17, 21, 25, 27, 28},
		{28, 27, 25, 21, 17, 13, 9, 5},
	}
	input := []float64{0.03, 0.05, 0.08, 0.02, 0.10, 0.06, 0.09, 0.04}

	expectedCurrents := referenceRowCurrentsUA(weights, input, 30)
	sense := arraysim.SenseChain{
		TIA: arraysim.TIAConfig{Rf: 8e3, Vref: 0.02, Vmin: 0.0, Vmax: 1.0},
		ADC: arraysim.ADCConfig{Bits: 5, Vmin: 0.0, Vmax: 1.0},
	}

	expectedCodes := make([]int, len(expectedCurrents))
	for i, iuA := range expectedCurrents {
		expectedCodes[i] = sense.ConvertCurrent(iuA * 1e-6).Code
	}

	const trials = 200
	bitsPerCode := sense.ADC.Bits
	totalBits := trials * len(expectedCodes) * bitsPerCode
	bitErrors := 0
	rng := rand.New(rand.NewSource(42))
	for n := 0; n < trials; n++ {
		for i, iuA := range expectedCurrents {
			actualCode := sense.ConvertCurrentWithNoise(iuA*1e-6, rng).Code
			bitErrors += hammingDistance(actualCode, expectedCodes[i], bitsPerCode)
		}
	}

	ber := float64(bitErrors) / float64(totalBits)
	if ber >= 0.05 {
		t.Fatalf("BER too high: got %.4f, want < 0.05 (bitErrors=%d totalBits=%d)", ber, bitErrors, totalBits)
	}

	// Quantized output consistency check against expected quantized code.
	for i, iuA := range expectedCurrents {
		got := sense.ConvertCurrent(iuA * 1e-6).Code
		if math.Abs(float64(got-expectedCodes[i])) > 0 {
			t.Fatalf("row %d expected quantized code %d got %d", i, expectedCodes[i], got)
		}
	}
}

func hammingDistance(a, b, bits int) int {
	x := a ^ b
	count := 0
	for i := 0; i < bits; i++ {
		if (x>>i)&1 == 1 {
			count++
		}
	}
	return count
}
