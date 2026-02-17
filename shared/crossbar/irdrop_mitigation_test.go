package crossbar

import "testing"

func TestIRDropMitigation_AllStrategiesReduceMaxDrop(t *testing.T) {
	if testing.Short() {
		t.Skip("IR-drop mitigation sweep is noisy/slow; skip in -short")
	}
	const n = 32

	baseline := NewIRDropSimulator(n, n)
	for i := 0; i < n; i++ {
		baseline.SetInputVoltage(i, 1.0)
	}
	baseline.Simulate(800)
	baseDrop := baseline.GetMaxIRDrop()
	if baseDrop <= 0 {
		t.Fatalf("expected baseline max IR drop > 0; got %.12g", baseDrop)
	}
	baseRowR := baseline.RowResist
	baseColR := baseline.ColResist

	cases := []struct {
		name string
		mit  IRDropMitigation
	}{
		{
			name: "WidenedLines_2x",
			mit:  IRDropMitigation{UseWidenedLines: true, LineWidthIncrease: 2.0},
		},
		{
			name: "HierarchicalBus",
			mit:  IRDropMitigation{UseHierarchical: true},
		},
		{
			name: "TiledArray_8",
			mit:  IRDropMitigation{UseTiledArray: true, TileSize: 8},
		},
		{
			name: "Combined",
			mit:  IRDropMitigation{UseWidenedLines: true, LineWidthIncrease: 2.0, UseHierarchical: true, UseTiledArray: true, TileSize: 8},
		},
	}

	for _, tc := range cases {
		sim := NewIRDropSimulator(n, n)
		for i := 0; i < n; i++ {
			sim.SetInputVoltage(i, 1.0)
		}
		sim.RowResist = baseRowR
		sim.ColResist = baseColR
		sim.Simulate(800)

		before := sim.GetMaxIRDrop()
		sim.ApplyMitigation(tc.mit)
		after := sim.GetMaxIRDrop()

		if after >= before {
			t.Fatalf("%s: expected mitigation to reduce max drop; before=%.12g after=%.12g", tc.name, before, after)
		}
		improvePct := (before - after) / before * 100
		t.Logf("%s: maxDrop %.6g -> %.6g (improvement %.3f%%)", tc.name, before, after, improvePct)

		if after >= baseDrop {
			t.Fatalf("%s: expected mitigation drop < baseline; baseline=%.12g after=%.12g", tc.name, baseDrop, after)
		}
	}
}
