package crossbar

import (
	"math/rand"
	"testing"
)

// M2-WRD-03: Spatial disturb pattern: same-row cells accumulate more than different-row cells.
func TestM2WRD03_WriteDisturb_SpatialSameRowGreaterThanDifferentRow(t *testing.T) {
	rand.Seed(5)

	cfg := &Config{
		Rows:             6,
		Cols:             6,
		NoiseLevel:       0,
		ADCBits:          8,
		DACBits:          8,
		ConductanceModel: ConductanceLinear,
		HalfSelect:       &HalfSelectConfig{Enabled: true, DisturbThreshold: 0.3, DisturbRate: 0.02},
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	targetRow, targetCol := 2, 2
	N := 40
	for i := 0; i < N; i++ {
		if err := arr.ProgramWeightWithDisturb(targetRow, targetCol, 0.9, true); err != nil {
			t.Fatalf("ProgramWeightWithDisturb: %v", err)
		}
	}

	sameRowSum := 0.0
	sameRowN := 0
	diffSum := 0.0
	diffN := 0

	for r := 0; r < cfg.Rows; r++ {
		for c := 0; c < cfg.Cols; c++ {
			if r == targetRow && c == targetCol {
				continue
			}
			shift := arr.cells[r][c].DisturbShift
			if r == targetRow {
				sameRowSum += shift
				sameRowN++
			} else if r != targetRow && c != targetCol {
				diffSum += shift
				diffN++
			}
		}
	}

	avgSameRow := sameRowSum / float64(sameRowN)
	avgDiff := diffSum / float64(diffN)
	if !(avgSameRow > avgDiff) {
		t.Fatalf("expected same-row disturb > different-row disturb: avgSameRow=%g avgDiff=%g", avgSameRow, avgDiff)
	}

	t.Logf("M2-WRD-03 avg disturb shift: same-row=%g different-row=%g", avgSameRow, avgDiff)
}
