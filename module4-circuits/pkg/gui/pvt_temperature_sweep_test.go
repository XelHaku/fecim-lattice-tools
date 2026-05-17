//go:build legacy_fyne

package gui

import (
	"math"
	"testing"

	"fecim-lattice-tools/shared/peripherals"
)

func TestPVTTemperatureSweep_EcPrMonotonicAndFunctional(t *testing.T) {
	ds := newTestDeviceState(2, 2)
	mat := ds.GetMaterial()
	if mat == nil {
		t.Fatal("expected material")
	}

	tempsC := []float64{-40, 25, 85, 125}
	var prevEc, prevPr float64
	for i, tc := range tempsC {
		tk := tc + 273.15
		ec := mat.CoerciveFieldAtTemp(tk)
		pr := mat.PolarizationAtTemp(tk)
		if ec <= 0 || pr <= 0 {
			t.Fatalf("non-physical temperature scaling at %.1f°C: Ec=%.3e Pr=%.3e", tc, ec, pr)
		}
		if i > 0 {
			if ec > prevEc {
				t.Fatalf("Ec should not increase with temperature: %.3e -> %.3e", prevEc, ec)
			}
			if pr > prevPr {
				t.Fatalf("Pr should not increase with temperature: %.3e -> %.3e", prevPr, pr)
			}
		}
		prevEc, prevPr = ec, pr

		ds.SetPeripheralPVT(tk, peripherals.CornerTypical)
		applied, code := ds.DACWriteVoltage(0.7)
		if !isFinite(applied) || code < 0 {
			t.Fatalf("invalid DAC write result at %.1f°C: V=%.6g code=%d", tc, applied, code)
		}
		weights := [][]int{{15, 15}, {15, 15}}
		ds.Compute(weights, 30)
		if !isFinite(ds.GetRowCurrent(0)) {
			t.Fatalf("invalid read current at %.1f°C", tc)
		}
	}
}

func isFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
