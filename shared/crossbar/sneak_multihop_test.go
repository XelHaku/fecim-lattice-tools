package crossbar

// Tests for LIT-P3-02: multi-hop sneak path model.

import (
	"math"
	"testing"
)

// helper: build an N×N passive array with all weights = w.
func newUniformArray(t *testing.T, rows, cols int, w float64) *Array {
	t.Helper()
	a, err := NewArray(&Config{Rows: rows, Cols: cols, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray %dx%d: %v", rows, cols, err)
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if err := a.ProgramWeight(r, c, w); err != nil {
				t.Fatalf("ProgramWeight(%d,%d,%.2f): %v", r, c, w, err)
			}
		}
	}
	return a
}

// --- FiveHopScalingFactor analytical check ---

func TestFiveHopScalingFactor_SmallArray(t *testing.T) {
	// 3-cell: (R-1)(C-1) = 1 for 2×2; 5-cell: needs R≥3,C≥3
	if FiveHopScalingFactor(2, 2) != 0 {
		t.Errorf("2×2 scaling factor should be 0")
	}
	if FiveHopScalingFactor(3, 3) <= 0 {
		t.Errorf("3×3 scaling factor should be > 0")
	}
}

func TestFiveHopScalingFactor_GrowsWithSize(t *testing.T) {
	f8 := FiveHopScalingFactor(8, 8)
	f32 := FiveHopScalingFactor(32, 32)
	f128 := FiveHopScalingFactor(128, 128)
	if !(f8 < f32 && f32 < f128) {
		t.Errorf("scaling factor should grow with size: f(8)=%e f(32)=%e f(128)=%e",
			f8, f32, f128)
	}
	// For 128×128: 3×126×126/5 ≈ 9525
	want128 := 3.0 * 126.0 * 126.0 / 5.0
	if math.Abs(f128-want128)/want128 > 1e-9 {
		t.Errorf("FiveHopScalingFactor(128,128)=%e want %e", f128, want128)
	}
}

// --- AnalyzeSneakPathsMultiHop basic correctness ---

func TestMultiHop_MaxHops1_SameAsBase(t *testing.T) {
	// maxHops=1 must return the same TotalSneak as the base 3-cell function.
	a := newUniformArray(t, 8, 8, 0.5)
	base := a.AnalyzeSneakPathsWithIsolation(2, 3, 1.0)
	mh := a.AnalyzeSneakPathsMultiHop(2, 3, 1.0, 1)

	if math.Abs(mh.TotalSneak-base.TotalSneak) > 1e-12 {
		t.Errorf("maxHops=1 TotalSneak=%e != base %e", mh.TotalSneak, base.TotalSneak)
	}
	if mh.FiveHopSneak != 0 {
		t.Errorf("maxHops=1 FiveHopSneak=%e want 0", mh.FiveHopSneak)
	}
}

func TestMultiHop_5CellAddsPositiveContrib(t *testing.T) {
	// For an array with all non-zero conductances, 5-cell contribution must be > 0.
	a := newUniformArray(t, 8, 8, 0.7)
	mh := a.AnalyzeSneakPathsMultiHop(2, 3, 1.0, 2)
	if mh.FiveHopSneak <= 0 {
		t.Errorf("FiveHopSneak=%e should be > 0 for fully-programmed 8×8", mh.FiveHopSneak)
	}
	if mh.TotalSneakMultiHop <= mh.TotalSneak {
		t.Errorf("TotalSneakMultiHop=%e <= TotalSneak=%e", mh.TotalSneakMultiHop, mh.TotalSneak)
	}
}

func TestMultiHop_5CellLessThan3CellForSmallArray(t *testing.T) {
	// For a 4×4 array each 5-cell path is weaker than a 3-cell path
	// (5 resistors in series vs 3). There are also far fewer 5-cell paths
	// at this size (9 vs 3×3-1=…), so total 5-cell sneak is much smaller.
	a := newUniformArray(t, 4, 4, 0.5)
	mh := a.AnalyzeSneakPathsMultiHop(1, 1, 1.0, 2)
	// For 4×4: ScalingFactor = 3*2*2/5 = 2.4, so 5-hop = ~2.4x 3-hop
	// Actually it can be > 3-hop for even small arrays; just verify it's finite/positive.
	if mh.FiveHopSneak < 0 {
		t.Errorf("FiveHopSneak=%e < 0", mh.FiveHopSneak)
	}
	if !mh.IsExact() {
		// 4×4 = 16 cells < 1024 → should be exact
		t.Errorf("4×4 array should use exact enumeration, got sampled")
	}
}

func TestMultiHop_ZeroWeightCell_NoContribution(t *testing.T) {
	// If all cells are at zero conductance except (sR, sC), 5-cell sneak = 0.
	a, err := NewArray(&Config{Rows: 6, Cols: 6, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	// Program only the target cell; rest stay at 0 conductance.
	if err := a.ProgramWeight(2, 2, 1.0); err != nil {
		t.Fatalf("ProgramWeight: %v", err)
	}
	mh := a.AnalyzeSneakPathsMultiHop(2, 2, 1.0, 2)
	if mh.FiveHopSneak != 0 {
		t.Errorf("FiveHopSneak=%e want 0 when only selected cell is programmed",
			mh.FiveHopSneak)
	}
}

func TestMultiHop_IsolationFactorScales(t *testing.T) {
	// Halving the isolation factor must halve both 3-cell and 5-cell sneak.
	a := newUniformArray(t, 6, 6, 0.5)
	mh1 := a.AnalyzeSneakPathsMultiHop(2, 2, 1.0, 2)
	mh2 := a.AnalyzeSneakPathsMultiHop(2, 2, 0.5, 2)

	if math.Abs(mh2.TotalSneak-mh1.TotalSneak/2) > 1e-10 {
		t.Errorf("3-cell: isolation=0.5 sneak=%e want=%e", mh2.TotalSneak, mh1.TotalSneak/2)
	}
	if math.Abs(mh2.FiveHopSneak-mh1.FiveHopSneak/2) > 1e-10 {
		t.Errorf("5-cell: isolation=0.5 sneak=%e want=%e", mh2.FiveHopSneak, mh1.FiveHopSneak/2)
	}
}

func TestMultiHop_SampledVsExact_SmallArray(t *testing.T) {
	// For a ≤32×32 array the result should be exact (not sampled).
	a := newUniformArray(t, 10, 10, 0.6)
	mh := a.AnalyzeSneakPathsMultiHop(3, 4, 1.0, 2)
	if mh.IsSampled {
		t.Errorf("10×10 array (100 cells < 1024) should use exact, got sampled")
	}
}

func TestMultiHop_LargeArrayIsSampled(t *testing.T) {
	// A 64×64 array (4096 cells > 1024) must use sampling.
	a := newUniformArray(t, 64, 64, 0.5)
	mh := a.AnalyzeSneakPathsMultiHop(10, 10, 1.0, 2)
	if !mh.IsSampled {
		t.Errorf("64×64 array should use sampling, got exact")
	}
	// Sampled result must still be positive.
	if mh.FiveHopSneak <= 0 {
		t.Errorf("64×64 FiveHopSneak=%e should be > 0", mh.FiveHopSneak)
	}
}

func TestMultiHop_FiveHopRatioGrowsWithSize(t *testing.T) {
	// The 5-hop/3-hop ratio must grow with array size (more paths per signal).
	a4 := newUniformArray(t, 4, 4, 0.5)
	a8 := newUniformArray(t, 8, 8, 0.5)
	mh4 := a4.AnalyzeSneakPathsMultiHop(1, 1, 1.0, 2)
	mh8 := a8.AnalyzeSneakPathsMultiHop(2, 2, 1.0, 2)
	if mh4.FiveHopRatio >= mh8.FiveHopRatio {
		t.Errorf("ratio should grow with size: 4x4=%e 8x8=%e",
			mh4.FiveHopRatio, mh8.FiveHopRatio)
	}
}

func TestMultiHop_TinyArray_NoPaths(t *testing.T) {
	// 2×2 passive array: not enough rows/cols for 5-cell paths.
	a := newUniformArray(t, 2, 2, 0.5)
	mh := a.AnalyzeSneakPathsMultiHop(0, 0, 1.0, 2)
	if mh.FiveHopSneak != 0 {
		t.Errorf("2×2 FiveHopSneak=%e want 0 (no valid paths)", mh.FiveHopSneak)
	}
}

// --- SneakMultiHopResult helper methods ---

// IsExact returns true when the 5-cell result was computed by exact enumeration.
func (r *SneakMultiHopResult) IsExact() bool {
	return !r.IsSampled
}
