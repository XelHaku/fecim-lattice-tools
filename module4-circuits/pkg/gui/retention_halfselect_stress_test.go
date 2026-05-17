//go:build legacy_fyne

package gui

import (
	"math"
	"testing"
)

func TestRetentionHalfSelectStress_PowerLawBoundedDrift(t *testing.T) {
	ds := newTestDeviceState(4, 4)
	ds.SetPassiveMode(true)
	ds.SetOperationMode(OpModeWrite)
	mat := ds.GetMaterial()

	weights := [][]int{{15, 15, 15, 15}, {15, 15, 15, 15}, {15, 15, 15, 15}, {15, 15, 15, 15}}
	stressRow, stressCol := 1, 2 // half-selected when writing (1,1)
	fullWriteV := ds.GetWriteRange().Max * 0.9
	vc := math.Max(mat.CoerciveVoltage(), 1e-9)

	// Simplified sub-threshold disturb model for retention stress:
	// ΔP_{n+1}=ΔP_n+k*(Vhalf/Vc)^η*(ΔP_sat-ΔP_n)
	const (
		cycles = 120
		k      = 0.02
		eta    = 2.5
	)
	deltaSat := 0.18 * math.Abs(mat.Ps)
	deltaP := make([]float64, 0, cycles)
	current := 0.0

	for i := 0; i < cycles; i++ {
		ds.ApplyHalfSelectWrite(1, 1, fullWriteV)
		ds.Compute(weights, 30)
		vHalf := math.Abs(ds.GetEffectiveCellVoltage(stressRow, stressCol))
		ratio := math.Pow(math.Min(vHalf/vc, 1.0), eta)
		current += k * ratio * (deltaSat - current)
		deltaP = append(deltaP, current)
		ds.ResetWriteVoltages()
	}

	A, alpha := fitPowerLaw(deltaP)
	if alpha < -0.05 || alpha > 1.0 {
		t.Fatalf("unexpected power-law exponent alpha=%.4f (A=%.3e)", alpha, A)
	}
	if end := deltaP[len(deltaP)-1]; end > 0.20*math.Abs(mat.Ps) {
		t.Fatalf("half-select retention drift too high: ΔP=%.3e C/m²", end)
	}
}

func fitPowerLaw(delta []float64) (A, alpha float64) {
	type pt struct{ x, y float64 }
	pts := make([]pt, 0, len(delta))
	for i, d := range delta {
		if d <= 0 {
			continue
		}
		pts = append(pts, pt{x: math.Log(float64(i + 1)), y: math.Log(d)})
	}
	if len(pts) < 2 {
		return 0, 0
	}
	var sx, sy, sxx, sxy float64
	for _, p := range pts {
		sx += p.x
		sy += p.y
		sxx += p.x * p.x
		sxy += p.x * p.y
	}
	n := float64(len(pts))
	den := n*sxx - sx*sx
	if math.Abs(den) < 1e-18 {
		return 0, 0
	}
	alpha = (n*sxy - sx*sy) / den
	lnA := (sy - alpha*sx) / n
	return math.Exp(lnA), alpha
}
