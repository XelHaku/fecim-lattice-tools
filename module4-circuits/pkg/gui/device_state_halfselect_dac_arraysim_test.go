//go:build legacy_fyne

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

	// DAC-Only Column Drive: all WLs grounded, selected BL driven to -V_write.
	// All cells in the selected column see full V_write (column write).
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

	if got := ds.GetEffectiveCellVoltage(row, col); math.Abs(got-wantTarget) > eps {
		t.Fatalf("target Vcell: got %.9f, want %.9f", got, wantTarget)
	}

	// Same row, different columns: WL=0, BL=0 → ΔV=0 (safe, no disturb)
	for c := 0; c < 3; c++ {
		if c == col {
			continue
		}
		if got := ds.GetEffectiveCellVoltage(row, c); math.Abs(got) > eps {
			t.Fatalf("same-row cell (%d,%d): got %.9f, want 0 (safe — DAC-Only drive)", row, c, got)
		}
	}
	// Same column, different rows: WL=0, BL=-V_write → ΔV=+V_write (full disturb — column write!)
	for r := 0; r < 3; r++ {
		if r == row {
			continue
		}
		if got := ds.GetEffectiveCellVoltage(r, col); math.Abs(got-wantTarget) > eps {
			t.Fatalf("same-col cell (%d,%d): got %.9f, want %.9f (full disturb — column write)", r, col, got, wantTarget)
		}
	}
	// Diagonal/unselected: WL=0, BL=0 → ΔV=0
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
