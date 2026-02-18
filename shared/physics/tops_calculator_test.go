package physics

import (
	"math"
	"testing"
)

func almostEqual(t *testing.T, got, want, relTol float64) {
	t.Helper()
	if want == 0 {
		if math.Abs(got) > relTol {
			t.Fatalf("got %.12g, want %.12g", got, want)
		}
		return
	}
	if math.Abs(got-want)/math.Abs(want) > relTol {
		t.Fatalf("got %.12g, want %.12g (relTol %.3g)", got, want, relTol)
	}
}

func TestComputeMetrics8x8At1GHz(t *testing.T) {
	m := ComputeMetrics{
		ArrayRows:     8,
		ArrayCols:     8,
		Frequency:     1e9,
		DACBits:       5,
		ADCBits:       5,
		EnergyPerMVM:  2.944e-12, // 64 cells * 46 fJ/cell
		LatencyPerMVM: 1e-9,
	}

	almostEqual(t, m.OpsPerSecond(), 1.28e11, 1e-12)
	almostEqual(t, m.TOPS(), 0.128, 1e-12)
	almostEqual(t, m.PowerW(), 2.944e-3, 1e-12)
	almostEqual(t, m.TOPSPerW(), 43.478260869565, 1e-9)

	if m.TOPSPerW() <= BaselineGPUA100.TOPSPerW {
		t.Fatalf("expected FeCIM TOPS/W %.3f to exceed GPU baseline %.3f", m.TOPSPerW(), BaselineGPUA100.TOPSPerW)
	}
}

func TestComputeMetrics64x64At100MHzScaling(t *testing.T) {
	mSmall := ComputeMetrics{
		ArrayRows:    8,
		ArrayCols:    8,
		Frequency:    1e9,
		EnergyPerMVM: 2.944e-12,
	}
	mLarge := ComputeMetrics{
		ArrayRows:    64,
		ArrayCols:    64,
		Frequency:    1e8,
		EnergyPerMVM: 188.416e-12, // linear scaling from 8x8 baseline
	}

	// Throughput scales as N*M*f.
	almostEqual(t, mLarge.TOPS()/mSmall.TOPS(), 6.4, 1e-12)

	// Energy per MVM scales with array size for fixed per-cell read path energy.
	almostEqual(t, mLarge.EnergyPerMVM/mSmall.EnergyPerMVM, 64.0, 1e-12)

	if mLarge.TOPSPerW() <= BaselineGPUA100.TOPSPerW {
		t.Fatalf("expected scaled FeCIM TOPS/W %.3f to exceed GPU baseline %.3f", mLarge.TOPSPerW(), BaselineGPUA100.TOPSPerW)
	}
}
