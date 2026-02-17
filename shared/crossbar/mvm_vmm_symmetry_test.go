package crossbar

import (
	"math"
	"math/rand"
	"testing"
)

func transpose(W [][]float64) [][]float64 {
	if len(W) == 0 {
		return nil
	}
	rows := len(W)
	cols := len(W[0])
	WT := make([][]float64, cols)
	for j := range WT {
		WT[j] = make([]float64, rows)
		for i := 0; i < rows; i++ {
			WT[j][i] = W[i][j]
		}
	}
	return WT
}

func TestMVMVMMSymmetry_M2MVM04(t *testing.T) {
	rand.Seed(3)
	// Use 8-bit quantization (typical), and compare within 1 LSB plus epsilon.
	const (
		n    = 16
		dac  = 8
		adc  = 8
		eps  = 1e-12
	)
	lsb := 1.0 / float64((1<<adc)-1)
	tol := lsb + eps

	W := make([][]float64, n)
	for i := range W {
		W[i] = make([]float64, n)
		for j := range W[i] {
			W[i][j] = 0.1 + 0.8*rand.Float64()
		}
	}
	WT := transpose(W)

	x := make([]float64, n)
	for i := range x {
		x[i] = 0.1 + 0.8*rand.Float64()
	}

	arrM, err := NewArray(&Config{Rows: n, Cols: n, ADCBits: adc, DACBits: dac, NoiseLevel: 0})
	if err != nil {
		t.Fatalf("NewArray(MVM): %v", err)
	}
	arrV, err := NewArray(&Config{Rows: n, Cols: n, ADCBits: adc, DACBits: dac, NoiseLevel: 0})
	if err != nil {
		t.Fatalf("NewArray(VMM): %v", err)
	}

	// Directly set conductances to avoid 30-level programming quantization.
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			arrM.cells[i][j].Conductance = W[i][j]
			arrM.cells[i][j].NoiseFactor = 1
			arrV.cells[i][j].Conductance = WT[i][j]
			arrV.cells[i][j].NoiseFactor = 1
		}
	}

	yM, err := arrM.MVM(x)
	if err != nil {
		t.Fatalf("MVM: %v", err)
	}
	yV, err := arrV.VMM(x)
	if err != nil {
		t.Fatalf("VMM: %v", err)
	}
	if len(yM) != len(yV) {
		t.Fatalf("length mismatch: MVM=%d VMM=%d", len(yM), len(yV))
	}

	maxDelta := 0.0
	for i := range yM {
		d := math.Abs(yM[i] - yV[i])
		if d > maxDelta {
			maxDelta = d
		}
		if d > tol {
			t.Fatalf("symmetry mismatch at idx=%d: MVM=%.6g VMM=%.6g |diff|=%.3g > tol=%.3g", i, yM[i], yV[i], d, tol)
		}
	}
	t.Logf("max symmetry |delta|=%.3g (tol=%.3g)", maxDelta, tol)
}
