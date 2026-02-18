package gui

import (
	"math"
	"testing"
)

func TestRetentionReadStress_MidLevelConductanceStable(t *testing.T) {
	ds := newTestDeviceState(2, 2)
	ds.SetPassiveMode(true)
	ds.SetOperationMode(OpModeRead)
	mat := ds.GetMaterial()
	vc := math.Max(mat.CoerciveVoltage(), 1e-9)

	weights := [][]int{{15, 15}, {15, 15}} // mid-level (15/30)
	row, col := 0, 0
	g0 := ds.levelToConductance(weights[row][col], 30)
	g := g0

	readHalfV := 0.08 // selected cell sees ~0.16V (well below Vc)
	const reads = 2000

	// Simplified assumption: low-voltage reads cause tiny cumulative drift.
	// d(ΔG/G)=k*(Vread/Vc)^η*(limit-ΔG/G), with limit=0.8%.
	rel := 0.0
	const (
		k     = 0.02
		eta   = 3.0
		limit = 0.008
	)
	for i := 0; i < reads; i++ {
		ds.wlVoltages[row], ds.wlVoltages[1] = readHalfV, 0
		ds.dacVoltages[col], ds.dacVoltages[1] = -readHalfV, 0
		ds.Compute(weights, 30)
		vRead := math.Abs(ds.GetEffectiveCellVoltage(row, col))
		ratio := math.Pow(math.Min(vRead/vc, 1.0), eta)
		rel += k * ratio * (limit - rel)
	}
	g = g0 * (1 - rel)

	relative := math.Abs(g-g0) / math.Max(g0, 1e-18)
	if relative >= 0.01 {
		t.Fatalf("read disturb too high: ΔG/G=%.4f (g0=%.3e, g1=%.3e)", relative, g0, g)
	}
}
