package physics

import "testing"

func TestNLS_CumulativeSwitching(t *testing.T) {
	s := NewLKSolver()
	E := 1.8e8
	prev := 0.0
	for i := 1; i <= 40; i++ {
		f := s.nlsSwitchedFraction(E, float64(i)*1e-9)
		if f < prev {
			t.Fatalf("fraction decreased at step %d: prev=%g now=%g", i, prev, f)
		}
		prev = f
	}
}

func TestNLS_HigherFieldFasterSwitching(t *testing.T) {
	s := NewLKSolver()
	time := 5e-9
	fLow := s.nlsSwitchedFraction(0.8e8, time)
	fHigh := s.nlsSwitchedFraction(2.0e8, time)
	if fHigh <= fLow {
		t.Fatalf("expected higher field to switch faster: low=%g high=%g", fLow, fHigh)
	}
}

func TestNLS_DeterministicForGivenParams(t *testing.T) {
	s := NewLKSolver()
	E := 1.5e8
	time := 7e-9
	f1 := s.nlsSwitchedFraction(E, time)
	f2 := s.nlsSwitchedFraction(E, time)
	if f1 != f2 {
		t.Fatalf("expected deterministic result, got f1=%g f2=%g", f1, f2)
	}
}

func TestNLS_LogNormalShape(t *testing.T) {
	s := NewLKSolver()
	E := 1.5e8
	times := []float64{1e-10, 5e-10, 1e-9, 2e-9, 5e-9, 1e-8, 2e-8, 5e-8, 1e-7, 5e-7, 1e-6}
	vals := make([]float64, len(times))
	for i, tt := range times {
		vals[i] = s.nlsSwitchedFraction(E, tt)
	}

	slopes := make([]float64, len(vals)-1)
	for i := 0; i < len(vals)-1; i++ {
		slopes[i] = vals[i+1] - vals[i]
	}

	peakIdx := 0
	for i := 1; i < len(slopes); i++ {
		if slopes[i] > slopes[peakIdx] {
			peakIdx = i
		}
	}
	if peakIdx == 0 || peakIdx == len(slopes)-1 {
		t.Fatalf("expected interior slope peak for S-shaped curve, peakIdx=%d slopes=%v vals=%v", peakIdx, slopes, vals)
	}
}
