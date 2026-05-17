//go:build legacy_fyne

package gui

import (
	"math"
	"testing"
)

func TestWriteDisturbSpatial_DecaysAwayFromSelectedRowCol(t *testing.T) {
	ds := newTestDeviceState(8, 8)
	ds.SetPassiveMode(true)
	ds.SetOperationMode(OpModeWrite)
	mat := ds.GetMaterial()
	vc := math.Max(mat.CoerciveVoltage(), 1e-9)

	weights := make([][]int, 8)
	for r := range weights {
		weights[r] = make([]int, 8)
		for c := range weights[r] {
			weights[r][c] = 15
		}
	}

	targetR, targetC := 4, 4
	ds.ApplyHalfSelectWrite(targetR, targetC, ds.GetWriteRange().Max*0.9)
	ds.Compute(weights, 30)

	var nearSum, farSum float64
	var nearN, farN int
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			v := math.Abs(ds.GetEffectiveCellVoltage(r, c))
			ratio := math.Pow(math.Min(v/vc, 1.0), 2.2)
			dp := 0.2 * math.Abs(mat.Ps) * ratio

			onCross := (r == targetR || c == targetC) && !(r == targetR && c == targetC)
			far := absInt(r-targetR) >= 2 && absInt(c-targetC) >= 2
			if onCross {
				nearSum += dp
				nearN++
			}
			if far {
				farSum += dp
				farN++
			}
		}
	}
	if nearN == 0 || farN == 0 {
		t.Fatal("invalid grouping for spatial disturb test")
	}
	nearAvg := nearSum / float64(nearN)
	farAvg := farSum / float64(farN)
	if nearAvg <= farAvg {
		t.Fatalf("expected cross disturb > far disturb; near=%.3e far=%.3e", nearAvg, farAvg)
	}
	if nearAvg < 2*farAvg {
		t.Fatalf("expected clear spatial decay (near >= 2x far); near=%.3e far=%.3e", nearAvg, farAvg)
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
