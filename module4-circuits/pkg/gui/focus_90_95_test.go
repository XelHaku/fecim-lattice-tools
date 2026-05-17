//go:build legacy_fyne

package gui

import (
	"math"
	"testing"
)

func TestFocus91_WriteRangeIsBipolarFromMaterialVc(t *testing.T) {
	ds := NewDeviceState(4, 4, nil, nil)
	wr := ds.GetWriteRange()
	vc := ds.GetMaterial().CoerciveVoltage()

	if wr.Min >= 0 || wr.Max <= 0 {
		t.Fatalf("write range must be bipolar, got [%.6f, %.6f]", wr.Min, wr.Max)
	}
	if math.Abs(wr.Max+wr.Min) > 1e-9 {
		t.Fatalf("write range must be symmetric around 0, got [%.6f, %.6f]", wr.Min, wr.Max)
	}
	if wr.Max < vc {
		t.Fatalf("write max must be derived from material coercive voltage (Vc=%.6f), got %.6f", vc, wr.Max)
	}

	applied, code := ds.DACWriteVoltage(wr.Min)
	if code != 0 {
		t.Fatalf("DAC code 0 should map to erase polarity (negative range edge): got code=%d applied=%.6f", code, applied)
	}
}

func TestFocus95_RandomInputVectorAppliesAfterResize(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	ca.resizeArray(4, 4)
	ca.setOperationMode(OpModeCompute)
	ca.randomizeInputVectorEntries()

	if len(ca.inputVector) != 4 {
		t.Fatalf("input vector length after resize: got %d, want 4", len(ca.inputVector))
	}

	readMax := ca.deviceState.GetReadRange().Max
	for c, code := range ca.inputVector {
		gotV := ca.deviceState.GetDACVoltage(c)
		wantV := (float64(code) / 255.0) * readMax
		if math.Abs(gotV-wantV) > 1e-9 {
			t.Fatalf("col %d DAC voltage mismatch after randomize+resize: got %.9f V, want %.9f V (code=%d)", c, gotV, wantV, code)
		}
	}
}
