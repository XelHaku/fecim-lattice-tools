package physics

import (
	"testing"
)

func TestWriteVerifyStats_Basic(t *testing.T) {
	stats := NewWriteVerifyStats()

	// Record some successful writes
	stats.RecordWrite(15, 2, true, false)
	stats.RecordWrite(10, 1, true, false)
	stats.RecordWrite(20, 3, true, false)

	if stats.TotalWrites != 3 {
		t.Errorf("Expected 3 total writes, got %d", stats.TotalWrites)
	}
	if stats.SuccessfulWrites != 3 {
		t.Errorf("Expected 3 successful writes, got %d", stats.SuccessfulWrites)
	}
	if stats.GetSuccessRate() != 1.0 {
		t.Errorf("Expected 100%% success rate, got %.2f", stats.GetSuccessRate())
	}
}

func TestWriteVerifyStats_Failures(t *testing.T) {
	stats := NewWriteVerifyStats()

	// Record mix of success and failure
	stats.RecordWrite(15, 2, true, false)
	stats.RecordWrite(10, 5, false, true) // Failed with overshoot
	stats.RecordWrite(20, 3, true, false)
	stats.RecordWrite(5, 5, false, false) // Failed, no overshoot

	if stats.TotalWrites != 4 {
		t.Errorf("Expected 4 total writes, got %d", stats.TotalWrites)
	}
	if stats.SuccessfulWrites != 2 {
		t.Errorf("Expected 2 successful writes, got %d", stats.SuccessfulWrites)
	}
	if stats.FailedWrites != 2 {
		t.Errorf("Expected 2 failed writes, got %d", stats.FailedWrites)
	}
	if stats.GetSuccessRate() != 0.5 {
		t.Errorf("Expected 50%% success rate, got %.2f", stats.GetSuccessRate())
	}
	if stats.OvershootCount != 1 {
		t.Errorf("Expected 1 overshoot, got %d", stats.OvershootCount)
	}
}

func TestWriteVerifyStats_PulsesHistogram(t *testing.T) {
	stats := NewWriteVerifyStats()

	// Record writes with different pulse counts
	stats.RecordWrite(10, 1, true, false) // 1 pulse
	stats.RecordWrite(11, 1, true, false) // 1 pulse
	stats.RecordWrite(12, 2, true, false) // 2 pulses
	stats.RecordWrite(13, 3, true, false) // 3 pulses
	stats.RecordWrite(14, 3, true, false) // 3 pulses
	stats.RecordWrite(15, 3, true, false) // 3 pulses

	histogram := stats.GetPulsesHistogram()

	if histogram[0] != 2 {
		t.Errorf("Expected 2 writes at 1 pulse, got %d", histogram[0])
	}
	if histogram[1] != 1 {
		t.Errorf("Expected 1 write at 2 pulses, got %d", histogram[1])
	}
	if histogram[2] != 3 {
		t.Errorf("Expected 3 writes at 3 pulses, got %d", histogram[2])
	}

	avgPulses := stats.GetAveragePulses()
	expected := float64(1*2+2*1+3*3) / 6.0 // (2+2+9)/6 = 2.17
	if avgPulses < 2.0 || avgPulses > 2.5 {
		t.Errorf("Expected avg pulses ~%.2f, got %.2f", expected, avgPulses)
	}
}

func TestWriteVerifyStats_LevelTracking(t *testing.T) {
	stats := NewWriteVerifyStats()

	// Level 15: 3 attempts, 2 success
	stats.RecordWrite(15, 2, true, false)
	stats.RecordWrite(15, 3, true, false)
	stats.RecordWrite(15, 5, false, false)

	// Level 5: 2 attempts, 2 success
	stats.RecordWrite(5, 1, true, false)
	stats.RecordWrite(5, 1, true, false)

	rates := stats.GetLevelSuccessRates()

	if rates[15] < 0.66 || rates[15] > 0.67 {
		t.Errorf("Expected level 15 success rate ~0.67, got %.2f", rates[15])
	}
	if rates[5] != 1.0 {
		t.Errorf("Expected level 5 success rate 1.0, got %.2f", rates[5])
	}

	hardest := stats.GetHardestLevels(1)
	if len(hardest) != 1 || hardest[0] != 15 {
		t.Errorf("Expected hardest level to be 15, got %v", hardest)
	}
}

func TestWriteVerifyStats_Reset(t *testing.T) {
	stats := NewWriteVerifyStats()

	stats.RecordWrite(10, 2, true, false)
	stats.RecordWrite(15, 3, true, true)

	if stats.TotalWrites != 2 {
		t.Errorf("Expected 2 writes before reset")
	}

	stats.Reset()

	if stats.TotalWrites != 0 {
		t.Errorf("Expected 0 writes after reset, got %d", stats.TotalWrites)
	}
	if stats.OvershootCount != 0 {
		t.Errorf("Expected 0 overshoots after reset, got %d", stats.OvershootCount)
	}
}

func TestSimulateFailureRateProgression(t *testing.T) {
	endurance := 1e9 // 10^9 cycles

	// Early cycles - very low failure rate
	earlyRate := SimulateFailureRateProgression(1000, endurance)
	if earlyRate > 0.01 {
		t.Errorf("Expected <1%% failure rate at 1000 cycles, got %.2f%%", earlyRate*100)
	}

	// Mid-life - slightly higher
	midRate := SimulateFailureRateProgression(int(endurance/10), endurance)
	if midRate < earlyRate {
		t.Errorf("Mid-life failure rate should be > early rate")
	}

	// Near endurance - higher still
	lateRate := SimulateFailureRateProgression(int(endurance/2), endurance)
	if lateRate < midRate {
		t.Errorf("Late failure rate should be > mid-life rate")
	}

	// At endurance limit - 100%
	endRate := SimulateFailureRateProgression(int(endurance), endurance)
	if endRate != 1.0 {
		t.Errorf("Expected 100%% failure at endurance limit, got %.2f%%", endRate*100)
	}
}

func TestWriteVerifyStats_Summary(t *testing.T) {
	stats := NewWriteVerifyStats()

	// Empty stats
	summary := stats.GetSummary()
	if summary != "No writes recorded" {
		t.Errorf("Expected 'No writes recorded', got '%s'", summary)
	}

	// With some data
	stats.RecordWrite(10, 2, true, false)
	stats.RecordWrite(15, 3, true, true)

	summary = stats.GetSummary()
	if len(summary) < 20 {
		t.Errorf("Expected longer summary, got '%s'", summary)
	}
}
