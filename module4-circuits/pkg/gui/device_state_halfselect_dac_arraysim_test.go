package gui

import (
	"math"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
)

func TestDeviceState_WritePath_AppliesDACQuantizedVoltage_ToTierACellVoltages(t *testing.T) {
	ds := NewDeviceState(3, 3, nil, nil)
	ds.SetPassiveMode(true)
	ds.SetOperationMode(OpModeWrite)
	ds.SetCouplingMode(arraysim.CouplingTierA)
	// Ensure Tier A behaves like an ideal WL-BL difference so we can assert exact values.
	ds.wireParams = arraysim.WireParams{RWordLine: 1e-18, RBitLine: 1e-18}

	// Pick a target voltage that will not land exactly on a DAC code.
	targetVoltage := 1.0
	applied, code := ds.DACWriteVoltage(targetVoltage)
	if code <= 0 {
		t.Fatalf("unexpected DAC code for %.3fV: %d", targetVoltage, code)
	}
	if math.Abs(applied-targetVoltage) < 1e-12 {
		t.Fatalf("test setup: expected quantization to change voltage, target=%.9f applied=%.9f", targetVoltage, applied)
	}

	// Apply the *applied* (post-DAC) voltage through the V/2 half-select scheme.
	row, col := 1, 1
	ds.ApplyHalfSelectWrite(row, col, applied)

	weights := [][]int{
		{15, 15, 15},
		{15, 15, 15},
		{15, 15, 15},
	}
	ds.Compute(weights, 30)

	eps := 1e-6
	wantTarget := applied
	wantHalf := applied / 2.0

	if got := ds.GetEffectiveCellVoltage(row, col); math.Abs(got-wantTarget) > eps {
		t.Fatalf("target Vcell: got %.9f, want %.9f", got, wantTarget)
	}

	// Same row, different columns: +V/2 exposure.
	for c := 0; c < 3; c++ {
		if c == col {
			continue
		}
		if got := ds.GetEffectiveCellVoltage(row, c); math.Abs(got-wantHalf) > eps {
			t.Fatalf("half-select same row cell (%d,%d): got %.9f, want %.9f", row, c, got, wantHalf)
		}
	}
	// Same column, different rows: +V/2 exposure.
	for r := 0; r < 3; r++ {
		if r == row {
			continue
		}
		if got := ds.GetEffectiveCellVoltage(r, col); math.Abs(got-wantHalf) > eps {
			t.Fatalf("half-select same col cell (%d,%d): got %.9f, want %.9f", r, col, got, wantHalf)
		}
	}
	// Diagonal/unselected: 0V.
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if r == row || c == col {
				continue
			}
			if got := ds.GetEffectiveCellVoltage(r, c); math.Abs(got) > eps {
				t.Fatalf("unselected cell (%d,%d): got %.9f, want 0", r, c, got)
			}
		}
	}
}
