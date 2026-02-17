package crossbar

import (
	"math"
	"testing"
)

func TestIRDropScaling_MonotonicWithArraySize(t *testing.T) {
	sizes := []int{4, 8, 16, 32, 64}
	maxDrops := make([]float64, 0, len(sizes))

	for _, n := range sizes {
		sim := NewIRDropSimulator(n, n)
		for i := 0; i < n; i++ {
			sim.SetInputVoltage(i, 1.0)
		}
		for j := 0; j < n; j++ {
			sim.VoltageOut[j] = 0
		}
		sim.Simulate(800)
		md := sim.GetMaxIRDrop()
		maxDrops = append(maxDrops, md)
		t.Logf("N=%d maxIRDrop=%.6g V (RowRes=%.6gΩ ColRes=%.6gΩ)", n, md, sim.RowResist, sim.ColResist)
	}

	for i := 1; i < len(maxDrops); i++ {
		if maxDrops[i] < maxDrops[i-1]-1e-15 {
			t.Fatalf("max IR drop not monotonic: N=%d drop=%.12g < N=%d drop=%.12g",
				sizes[i], maxDrops[i], sizes[i-1], maxDrops[i-1])
		}
	}
}

func TestIRDropScaling_CornerCellsWorstForLargeArrays(t *testing.T) {
	for _, n := range []int{8, 16, 32, 64} {
		sim := NewIRDropSimulator(n, n)
		for i := 0; i < n; i++ {
			sim.SetInputVoltage(i, 1.0)
		}
		sim.Simulate(800)
		st := sim.GetStats()

		isCorner := (st.WorstCellRow == 0 || st.WorstCellRow == n-1) && (st.WorstCellCol == 0 || st.WorstCellCol == n-1)
		if !isCorner {
			t.Fatalf("N=%d expected worst cell to be a corner; got (%d,%d) maxDrop=%.6g",
				n, st.WorstCellRow, st.WorstCellCol, st.MaxIRDrop)
		}
		t.Logf("N=%d worstCell=(%d,%d) maxIRDrop=%.6g", n, st.WorstCellRow, st.WorstCellCol, st.MaxIRDrop)
	}
}

func TestIRDropScaling_TemperatureIncreasesIRDrop(t *testing.T) {
	const n = 16
	temps := []float64{300, 350, 400}
	maxDrops := make([]float64, 0, len(temps))
	rowRes := make([]float64, 0, len(temps))

	for _, T := range temps {
		sim := NewIRDropSimulator(n, n)
		for i := 0; i < n; i++ {
			sim.SetInputVoltage(i, 1.0)
		}
		sim.SetTemperature(T)
		sim.Simulate(800)

		md := sim.GetMaxIRDrop()
		maxDrops = append(maxDrops, md)
		rowRes = append(rowRes, sim.RowResist)
		t.Logf("T=%.1fK RowRes=%.6gΩ ColRes=%.6gΩ maxIRDrop=%.6g V", T, sim.RowResist, sim.ColResist, md)
	}

	for i := 1; i < len(temps); i++ {
		if rowRes[i] <= rowRes[i-1] {
			t.Fatalf("expected RowRes to increase with T: T=%.1fK %.12gΩ <= T=%.1fK %.12gΩ", temps[i], rowRes[i], temps[i-1], rowRes[i-1])
		}
		if !(maxDrops[i] > maxDrops[i-1] || math.Abs(maxDrops[i]-maxDrops[i-1]) < 1e-15) {
			t.Fatalf("expected max IR drop to increase with T: T=%.1fK %.12gV < T=%.1fK %.12gV", temps[i], maxDrops[i], temps[i-1], maxDrops[i-1])
		}
	}
}
