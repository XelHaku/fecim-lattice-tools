package crossbar

import (
	"math"
	"math/rand"
	"testing"
)

// M2-WRD-01: Half-select accumulation: apply N pulses of V/2, verify DisturbShift increases monotonically, always >= 0.
func TestM2WRD01_HalfSelectDisturb_MonotonicAndNonNegative(t *testing.T) {
	rand.Seed(4)

	cfg := &Config{
		Rows:             4,
		Cols:             4,
		NoiseLevel:       0,
		ADCBits:          8,
		DACBits:          8,
		ConductanceModel: ConductanceLinear,
		HalfSelect: &HalfSelectConfig{
			Enabled:          true,
			DisturbThreshold: 0.3,
			DisturbRate:      0.01, // exaggerated for test visibility
		},
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	targetRow, targetCol := 1, 1
	probeRow, probeCol := targetRow, 3 // same row, half-selected

	prev := arr.cells[probeRow][probeCol].DisturbShift
	if prev < 0 {
		t.Fatalf("initial DisturbShift should be >=0, got %g", prev)
	}

	N := 50
	for k := 0; k < N; k++ {
		if err := arr.ProgramWeightWithDisturb(targetRow, targetCol, 1.0, true); err != nil {
			t.Fatalf("ProgramWeightWithDisturb: %v", err)
		}
		cur := arr.cells[probeRow][probeCol].DisturbShift
		if cur < 0 {
			t.Fatalf("DisturbShift became negative at pulse=%d: %g", k, cur)
		}
		if cur < prev {
			t.Fatalf("DisturbShift should be monotonic non-decreasing: pulse=%d prev=%g cur=%g", k, prev, cur)
		}
		prev = cur
	}

	t.Logf("M2-WRD-01 DisturbShift after N=%d half-select pulses: %g", N, prev)
}

// M2-WRD-02: Architecture comparison: 0T1R disturb must be > 1T1R disturb by at least 2x.
func TestM2WRD02_WriteDisturb_0T1RGreaterThan1T1R(t *testing.T) {
	rows, cols := 8, 8
	targetRow, targetCol := 3, 4
	probeRow, probeCol := targetRow, 0
	N := 100

	cfgPassive := &WriteDisturbConfig{Enable: true, StressAccumulationRate: 1e-3, StressThreshold: 1.0, Architecture1T1R: false, Architecture1T1RReduction: 0.1}
	cfgActive := &WriteDisturbConfig{Enable: true, StressAccumulationRate: 1e-3, StressThreshold: 1.0, Architecture1T1R: true, Architecture1T1RReduction: 0.1}

	passive := NewWriteDisturbEngine(rows, cols, cfgPassive)
	active := NewWriteDisturbEngine(rows, cols, cfgActive)

	for i := 0; i < N; i++ {
		passive.RecordWrite(targetRow, targetCol)
		active.RecordWrite(targetRow, targetCol)
	}

	s0 := passive.GetCellStress(probeRow, probeCol)
	s1 := active.GetCellStress(probeRow, probeCol)
	if s0 <= 0 || s1 <= 0 {
		t.Fatalf("expected positive stress accumulation: s0=%g s1=%g", s0, s1)
	}
	ratio := s0 / s1
	if math.IsNaN(ratio) || math.IsInf(ratio, 0) {
		t.Fatalf("invalid ratio: s0=%g s1=%g ratio=%g", s0, s1, ratio)
	}
	if ratio < 2.0 {
		t.Fatalf("expected 0T1R disturb >= 2x 1T1R: s0=%g s1=%g ratio=%0.3f", s0, s1, ratio)
	}

	t.Logf("M2-WRD-02 stress ratio (0T1R/1T1R) = %0.3f (s0=%g s1=%g)", ratio, s0, s1)
}
