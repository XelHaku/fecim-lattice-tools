package arraysim

import "testing"

func TestGenerateProgramScheduleAdaptiveLessDisturbThanRowMajor8x8(t *testing.T) {
	rowMajor, err := GenerateProgramSchedule(8, 8, ProgramOrderRowMajor, 1.0)
	if err != nil {
		t.Fatalf("row-major schedule failed: %v", err)
	}
	adaptive, err := GenerateProgramSchedule(8, 8, ProgramOrderAdaptive, 1.0)
	if err != nil {
		t.Fatalf("adaptive schedule failed: %v", err)
	}

	if adaptive.CumulativeDisturb >= rowMajor.CumulativeDisturb {
		t.Fatalf("expected adaptive disturb < row-major; got adaptive=%.3f row-major=%.3f", adaptive.CumulativeDisturb, rowMajor.CumulativeDisturb)
	}
}

func TestGenerateProgramScheduleLengths(t *testing.T) {
	modes := []ProgramOrderMode{ProgramOrderRowMajor, ProgramOrderCheckerboard, ProgramOrderSerpentine, ProgramOrderAdaptive}
	for _, mode := range modes {
		res, err := GenerateProgramSchedule(4, 5, mode, 1)
		if err != nil {
			t.Fatalf("mode %s failed: %v", mode, err)
		}
		if len(res.Order) != 20 {
			t.Fatalf("mode %s: expected 20 cells, got %d", mode, len(res.Order))
		}
	}
}

func TestGenerateProgramScheduleUnknownMode(t *testing.T) {
	if _, err := GenerateProgramSchedule(2, 2, ProgramOrderMode("unknown"), 1); err == nil {
		t.Fatalf("expected error for unknown mode")
	}
}
