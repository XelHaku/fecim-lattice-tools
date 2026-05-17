//go:build legacy_fyne

package gui

import (
	"math"
	"testing"
)

func TestWriteDisturbNCycle_HalfSelectedAccumulationBounded(t *testing.T) {
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
	neighbors := [][2]int{{4, 3}, {4, 5}, {3, 4}, {5, 4}}
	fullWriteV := ds.GetWriteRange().Max * 0.9

	const cycles = 80
	delta := make([]float64, 0, cycles)
	curr := 0.0
	dpSatTrue := 0.22 * math.Abs(mat.Ps)
	const eta = 2.2

	for i := 0; i < cycles; i++ {
		ds.ApplyHalfSelectWrite(targetR, targetC, fullWriteV)
		ds.Compute(weights, 30)
		var ratioAvg float64
		for _, rc := range neighbors {
			v := math.Abs(ds.GetEffectiveCellVoltage(rc[0], rc[1]))
			ratioAvg += math.Pow(math.Min(v/vc, 1.0), eta)
		}
		ratioAvg /= float64(len(neighbors))
		curr += 0.03 * ratioAvg * (dpSatTrue - curr)
		delta = append(delta, curr)
		ds.ResetWriteVoltages()
	}

	dpSat, tau := fitAccumulation(delta)
	if tau <= 0 || tau > 10*cycles {
		t.Fatalf("unexpected τ fit: %.3f cycles (ΔP_sat=%.3e)", tau, dpSat)
	}
	if delta[len(delta)-1] > 0.25*math.Abs(mat.Ps) {
		t.Fatalf("write disturb accumulation too high: ΔP_end=%.3e C/m²", delta[len(delta)-1])
	}
}

func fitAccumulation(delta []float64) (dpSat, tau float64) {
	if len(delta) == 0 {
		return 0, 0
	}
	maxD := delta[len(delta)-1]
	if maxD <= 0 {
		return 0, 1
	}
	bestErr := math.Inf(1)
	bestSat, bestTau := maxD, 1.0
	for sat := maxD; sat <= 3*maxD; sat += maxD / 100 {
		for t := 1.0; t <= float64(len(delta))*5; t += 1.0 {
			err := 0.0
			for i, d := range delta {
				pred := sat * (1 - math.Exp(-float64(i+1)/t))
				e := d - pred
				err += e * e
			}
			if err < bestErr {
				bestErr, bestSat, bestTau = err, sat, t
			}
		}
	}
	return bestSat, bestTau
}
